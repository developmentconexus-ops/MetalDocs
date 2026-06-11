package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/platform/ratelimit"
)

// normativeChainOrder is the §2.1 target order, outermost first (REQ-MW-7).
var normativeChainOrder = []string{
	"panic_recovery",
	"http_obs",
	"cors",
	"origin_protection",
	"pre_auth_login_rate_limit",
	"authn",
	"iam_authz",
	"presence_bump",
	"rate_limit",
}

// TestAPIChainOrder_REQMW7 asserts (a) apiChain declares the normative order
// and (b) buildChain executes the layers in that order around the final
// handler. Reordering apiChain breaks this test, not production.
func TestAPIChainOrder_REQMW7(t *testing.T) {
	var executed []string
	probe := func(name string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				executed = append(executed, name)
				next.ServeHTTP(w, r)
			})
		}
	}

	links := apiChain(
		probe("panic_recovery"),
		probe("http_obs"),
		probe("cors"),
		probe("origin_protection"),
		probe("pre_auth_login_rate_limit"),
		probe("authn"),
		probe("iam_authz"),
		probe("presence_bump"),
		probe("rate_limit"),
	)

	if len(links) != len(normativeChainOrder) {
		t.Fatalf("apiChain has %d links, want %d", len(links), len(normativeChainOrder))
	}
	for i, l := range links {
		if l.name != normativeChainOrder[i] {
			t.Fatalf("apiChain[%d] = %q, want %q (REQ-MW-7: chain order is normative)", i, l.name, normativeChainOrder[i])
		}
	}

	final := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		executed = append(executed, "handler")
		w.WriteHeader(http.StatusOK)
	})
	buildChain(final, links).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	want := append(append([]string{}, normativeChainOrder...), "handler")
	if len(executed) != len(want) {
		t.Fatalf("executed %v, want %v", executed, want)
	}
	for i := range want {
		if executed[i] != want[i] {
			t.Fatalf("execution order[%d] = %q, want %q (full: %v)", i, executed[i], want[i], executed)
		}
	}
}

func TestBuildChain_SkipsNilLinks(t *testing.T) {
	links := []chainLink{
		{"a", func(next http.Handler) http.Handler { return next }},
		{"presence_bump", nil},
	}
	h := buildChain(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}), links)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
}

// TestLoginRateLimit_OnlyThrottlesLoginPath proves the pre-auth limiter
// applies to POST /api/v1/auth/login only and is IP-keyed (REQ-MW-5).
func TestLoginRateLimit_OnlyThrottlesLoginPath(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg, err := ratelimit.NewConfig(map[ratelimit.RouteKey]int{ratelimit.RouteAuthLogin: 1})
	if err != nil {
		t.Fatal(err)
	}
	limiter := ratelimit.New(ctx, cfg)
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	h := loginRateLimit(limiter)(next)

	login := func() int {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
		req.RemoteAddr = "203.0.113.7:51234"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code
	}

	if got := login(); got != http.StatusOK {
		t.Fatalf("first login = %d, want 200", got)
	}
	if got := login(); got != http.StatusTooManyRequests {
		t.Fatalf("second login within window = %d, want 429", got)
	}

	// Non-login paths bypass the limiter entirely.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health/live", nil)
	req.RemoteAddr = "203.0.113.7:51234"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("non-login path = %d, want 200 (must not be throttled)", rec.Code)
	}
}
