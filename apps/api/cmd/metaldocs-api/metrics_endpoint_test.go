package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	platformmw "metaldocs/internal/platform/middleware"
	"metaldocs/internal/platform/observability"
)

// denyAllMiddleware simulates authMiddleware.Wrap/iamMiddleware.Wrap failing
// closed for any request that reaches them: every request that is NOT routed
// around the API chain gets a 401. This proves the dispatch-mux fix works by
// construction (the scrape never enters the chain) rather than merely proving
// the tier-1 resolver classifies /metrics as public.
func denyAllMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})
}

// buildComposedServerHandler assembles the same shape of handler main.go
// builds: apiChain(...) + buildChain(...) for the API mux, then (post-fix) a
// root dispatch mux that serves GET /metrics ahead of the chain and falls
// through everything else to the chained handler, all wrapped in
// platformmw.Recovery. Reuses chain_test.go's apiChain/buildChain harness
// (TestAPIChainOrder_REQMW7) rather than standing up the full production
// wiring (DB, auth service, IAM service, rate limiters), which is unnecessary
// to prove the routing/bypass behavior under test here.
func buildComposedServerHandler(httpObs *observability.HTTPObservability) http.Handler {
	apiMux := http.NewServeMux()
	apiMux.Handle("/api/v1/metrics", httpObs.MetricsHandler())
	apiMux.HandleFunc("GET /api/v1/health/live", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	links := apiChain(
		platformmw.Recovery,
		nil, // otel: skipped when unconfigured, same as production (Z-1)
		httpObs.Wrap,
		func(next http.Handler) http.Handler { return next }, // cors stub
		func(next http.Handler) http.Handler { return next }, // origin protection stub
		func(next http.Handler) http.Handler { return next }, // pre-auth login limiter stub
		denyAllMiddleware,                                     // authn: fails closed, like production authMiddleware.Wrap
		denyAllMiddleware,                                     // iam authz: fails closed, like production iamMiddleware.Wrap
		nil,                                                    // presence bump: nil when no SQL DB, same as production
		func(next http.Handler) http.Handler { return next },  // global rate limit stub
		platformmw.MethodNotAllowedJSON,
	)
	chained := buildChain(apiMux, links)

	// Post-fix wiring under test: /metrics is served from a top-level dispatch
	// mux ahead of the chained API handler, so it bypasses authn/iam/httpObs
	// counting/rate-limit entirely. Recovery still wraps the whole thing.
	rootMux := http.NewServeMux()
	rootMux.Handle("GET /metrics", httpObs.PrometheusHandler())
	rootMux.Handle("/", chained)

	return platformmw.Recovery(rootMux)
}

// TestMetricsEndpoint_BypassesAuthChain_REQF83 is the spec-review regression
// test for the M8 F8.3 defect: main.go previously mounted GET /metrics on the
// SAME mux wrapped by authMiddleware.Wrap + iamMiddleware.Wrap, and
// permissions.go's routeRules had no /metrics entry, so it failed closed to
// 401 for an unauthenticated Prometheus scraper. It also passed through
// httpObs.Wrap (self-counting the scrape) and the global rate limiter.
//
// Uses the apiChain/buildChain harness from chain_test.go (see
// buildComposedServerHandler) with real platformmw.Recovery and real
// observability.HTTPObservability, and deny-all stand-ins for
// authn/iam — proving bypass by construction, not by resolver classification.
func TestMetricsEndpoint_BypassesAuthChain_REQF83(t *testing.T) {
	httpObs := observability.NewHTTPObservability(func(*http.Request) string { return "" })
	handler := buildComposedServerHandler(httpObs)

	// Drive one non-metrics request through the full chain first so the
	// metaldocs_http_requests_total / errors_total / duration_seconds
	// CounterVec/HistogramVec families have at least one observed label
	// combination — Prometheus vecs emit nothing for a family with zero
	// samples, so scraping before any traffic would make the "family present"
	// assertion below vacuous.
	warmup := httptest.NewRequest(http.MethodGet, "/api/v1/health/live", nil)
	// health/live has no rule in this stub (only authn/iam stand-ins exist),
	// so it hits denyAllMiddleware and returns 401 — that is fine, httpObs.Wrap
	// runs OUTSIDE authn (REQ-MW-4) and records the request regardless of the
	// downstream status.
	handler.ServeHTTP(httptest.NewRecorder(), warmup)

	// The regression assertion: GET /metrics with NO auth cookie/header must
	// succeed. Pre-fix (mux.Handle("/metrics", ...) inside the chain) this
	// would hit denyAllMiddleware and return 401.
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /metrics with no auth = %d, want 200 (must bypass the API auth chain entirely)", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/plain") {
		t.Fatalf("Content-Type = %q, want Prometheus text exposition (text/plain;...)", ct)
	}

	body := rec.Body.String()
	for _, family := range []string{
		"metaldocs_http_requests_total",
		"metaldocs_http_request_errors_total",
		"metaldocs_http_request_duration_seconds",
	} {
		if !strings.Contains(body, family) {
			t.Errorf("scrape body missing metric family %q", family)
		}
	}
	if !strings.Contains(body, "go_goroutines") && !strings.Contains(body, "process_") {
		t.Error("scrape body missing a go_ or process_ collector line")
	}

	// Self-scrape must not be counted: /metrics never passes through
	// httpObs.Wrap (it is served ahead of the chain), so no
	// metaldocs_http_requests_total series should ever carry route="/metrics".
	if strings.Contains(body, `route="/metrics"`) {
		t.Error("scrape body contains a route=\"/metrics\" label series — the scrape endpoint is self-counting, meaning it did not bypass httpObs.Wrap")
	}

	// A second scrape must not change that either (proves it structurally,
	// not just as an artifact of the first call being the very first scrape).
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rec2.Code != http.StatusOK {
		t.Fatalf("second GET /metrics = %d, want 200", rec2.Code)
	}
	if strings.Contains(rec2.Body.String(), `route="/metrics"`) {
		t.Error("second scrape body contains a route=\"/metrics\" label series — self-scrape pollution")
	}

	// Non-GET methods on /metrics fall through to "/" → the full chain (Go
	// 1.22 ServeMux: "GET /metrics" matches GET only). denyAllMiddleware sits
	// in that chain, so POST must be 401, NOT 200 and NOT routed to the
	// Prometheus handler.
	postReq := httptest.NewRequest(http.MethodPost, "/metrics", nil)
	postRec := httptest.NewRecorder()
	handler.ServeHTTP(postRec, postReq)
	if postRec.Code != http.StatusUnauthorized {
		t.Fatalf("POST /metrics = %d, want 401 (must fall through to the full API chain, not the scrape handler)", postRec.Code)
	}
}
