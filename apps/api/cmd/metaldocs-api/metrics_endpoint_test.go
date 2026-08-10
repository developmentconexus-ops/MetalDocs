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
// closed for any request that reaches them: every request routed through the
// API chain gets a 401. It lets the public-handler test prove — by construction
// — that /metrics is NOT served on the public listener (it enters the chain and
// fails closed) rather than merely proving a resolver classification.
func denyAllMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})
}

// buildPublicHandler assembles the same shape of PUBLIC handler main.go builds
// post-F-R1: apiChain(...) + buildChain(...) over the API mux, wrapped by the
// chain's own Recovery — and NOTHING mounts /metrics. This mirrors production,
// where the public http.Server.Handler is the chained API handler only; the
// scrape surface lives on a separate listener (buildMetricsHandler). Reuses
// chain_test.go's apiChain/buildChain harness rather than the full production
// wiring (DB, auth, IAM, rate limiters), which is unnecessary to prove the
// routing/isolation behavior under test.
func buildPublicHandler(httpObs *observability.HTTPObservability) http.Handler {
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
		denyAllMiddleware, // authn: fails closed, like production authMiddleware.Wrap
		denyAllMiddleware, // iam authz: fails closed, like production iamMiddleware.Wrap
		nil,               // presence bump: nil when no SQL DB, same as production
		func(next http.Handler) http.Handler { return next }, // global rate limit stub
		nil, // contract validation: this fixture serves a stub route, not the spec surface
		platformmw.MethodNotAllowedJSON,
	)
	return buildChain(apiMux, links)
}

// buildMetricsHandler assembles the DEDICATED metrics listener's handler exactly
// as main.go builds it post-F-R1: platformmw.Recovery over a mux that serves
// GET /metrics from the Prometheus handler. This listener is never part of the
// API chain, so the scrape is credential-less and never self-counts.
func buildMetricsHandler(httpObs *observability.HTTPObservability) http.Handler {
	metricsMux := http.NewServeMux()
	metricsMux.Handle("GET /metrics", httpObs.PrometheusHandler())
	return platformmw.Recovery(metricsMux)
}

// TestMetricsEndpoint_DedicatedListener_REQF83 is the F-R1 regression test for
// the M8 F8.3 → Dim-9 DEBT: /metrics used to be served on a top-level dispatch
// mux fronting the SAME listener as the public API, so it was reachable (unauth)
// on any host-published API port. F-R1 moves it to a dedicated listener. This
// test pins the split structurally:
//
//   - PUBLIC handler: GET /metrics must NOT be an unauthenticated 200 scrape.
//     The public mux has no /metrics route, so the request enters the fail-closed
//     auth chain and returns 401. The scrape surface has left the public port.
//   - METRICS handler: GET /metrics returns 200 Prometheus exposition,
//     unauthenticated, with the binding metric families and a go_/process_ line,
//     and never self-counts (no route="/metrics" series).
func TestMetricsEndpoint_DedicatedListener_REQF83(t *testing.T) {
	httpObs := observability.NewHTTPObservability(func(*http.Request) string { return "" })
	public := buildPublicHandler(httpObs)
	metrics := buildMetricsHandler(httpObs)

	// Warm up one request through the PUBLIC chain so the
	// metaldocs_http_requests_total / errors_total / duration_seconds families
	// have at least one observed label combination — Prometheus vecs emit nothing
	// for a family with zero samples, so scraping before any traffic would make
	// the "family present" assertions below vacuous. health/live hits
	// denyAllMiddleware and returns 401; that is fine — httpObs.Wrap runs OUTSIDE
	// authn (REQ-MW-4) and records the request regardless of downstream status.
	// Deliberately a NON-/metrics path: we scrape before probing /metrics so the
	// self-count assertion is meaningful.
	warmup := httptest.NewRequest(http.MethodGet, "/api/v1/health/live", nil)
	public.ServeHTTP(httptest.NewRecorder(), warmup)

	// (1) Scrape: GET /metrics on the DEDICATED metrics handler succeeds
	// unauthenticated, before any /metrics traffic has touched the public chain.
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	metrics.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /metrics on the metrics listener with no auth = %d, want 200", rec.Code)
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

	// Self-scrape must not be counted: the metrics listener is not part of the
	// API chain, so scraping it never records a metaldocs_http_* sample. At this
	// point no /metrics request has hit the public chain either, so no
	// route="/metrics" series may exist.
	if strings.Contains(body, `route="/metrics"`) {
		t.Error("scrape body contains a route=\"/metrics\" label series before any /metrics traffic — the scrape endpoint is self-counting")
	}
	// A second scrape must not create one either (structural, not an artifact of
	// the first call being the very first scrape).
	rec2 := httptest.NewRecorder()
	metrics.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rec2.Code != http.StatusOK {
		t.Fatalf("second GET /metrics on the metrics listener = %d, want 200", rec2.Code)
	}
	if strings.Contains(rec2.Body.String(), `route="/metrics"`) {
		t.Error("second metrics scrape produced a route=\"/metrics\" series — self-scrape pollution")
	}

	// (2) Isolation: GET /metrics on the PUBLIC handler must NOT be a 200 scrape.
	// The public mux has no /metrics route → the request runs the chain and
	// denyAllMiddleware returns 401. Pre-F-R1 (rootMux dispatch on the same
	// listener) this returned 200 unauthenticated. (This 401 IS counted by
	// httpObs.Wrap with route="/metrics" — legitimate, it is a real request to
	// the public port, not a self-scrape; hence the assertions above run first.)
	pubReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	pubRec := httptest.NewRecorder()
	public.ServeHTTP(pubRec, pubReq)
	if pubRec.Code == http.StatusOK {
		t.Fatalf("GET /metrics on the PUBLIC listener = 200 — the scrape surface must NOT be served on the public port")
	}
	if pubRec.Code != http.StatusUnauthorized {
		t.Fatalf("GET /metrics on the PUBLIC listener = %d, want 401 (fail-closed auth chain; no /metrics route on the public mux)", pubRec.Code)
	}

	// (3) Non-GET on the metrics listener: the mux registers "GET /metrics" only,
	// so POST does not match → non-200 from the mux (NOT routed to the Prometheus
	// handler).
	postReq := httptest.NewRequest(http.MethodPost, "/metrics", nil)
	postRec := httptest.NewRecorder()
	metrics.ServeHTTP(postRec, postReq)
	if postRec.Code == http.StatusOK {
		t.Fatalf("POST /metrics on the metrics listener = 200, want non-200 (only GET is registered)")
	}
}
