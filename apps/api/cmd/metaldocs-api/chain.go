package main

import (
	"net/http"

	authdelivery "metaldocs/internal/modules/auth/delivery/http"
	"metaldocs/internal/platform/ratelimit"
)

// chainLink names one middleware layer of the API chain so the composed order
// can be asserted by test (REQ-MW-7).
type chainLink struct {
	name string
	wrap func(http.Handler) http.Handler
}

// apiChain returns the canonical middleware chain, outermost first, per
// wiki/architecture/backend-target-architecture.md §2.1 (REQ-MW-1..5).
// A nil wrap is skipped (e.g. presence when no SQL DB is wired, or otel when no
// exporter is configured — Z-1, REQ-OBS-3). The otel link sits directly inside
// panic_recovery: recovery must stay outermost so a panicked span never kills
// the process, but otel wraps everything else so httpObs sees the active span
// for trace-id correlation. chain_test.go asserts both the declared order and
// the composed execution order; reorder here and the build breaks, not production.
func apiChain(recovery, otel, httpObs, cors, origin, preAuthLoginLimit, authn, iamAuthz, presence, rateLimit, methodNotAllowed func(http.Handler) http.Handler) []chainLink {
	return []chainLink{
		{"panic_recovery", recovery},
		{"otel", otel},
		{"http_obs", httpObs},
		{"cors", cors},
		{"origin_protection", origin},
		{"pre_auth_login_rate_limit", preAuthLoginLimit},
		{"authn", authn},
		{"iam_authz", iamAuthz},
		{"presence_bump", presence},
		{"rate_limit", rateLimit},
		// Innermost (nearest the mux): rewrite the stdlib text/plain 404/405 the
		// method-routed ServeMux emits into problem+json, preserving Allow (D-03).
		{"method_not_allowed", methodNotAllowed},
	}
}

// buildChain composes links around final, outermost first: links[0] wraps
// everything else.
func buildChain(final http.Handler, links []chainLink) http.Handler {
	h := final
	for i := len(links) - 1; i >= 0; i-- {
		if links[i].wrap == nil {
			continue
		}
		h = links[i].wrap(h)
	}
	return h
}

// loginRateLimit applies the pre-auth IP-keyed limiter to the login endpoint
// only (REQ-MW-5); every other path passes through untouched. It runs before
// authn, so the user extractor deliberately returns "" — the limiter then
// keys by trusted-proxy-resolved client IP and fails closed.
func loginRateLimit(limiter *ratelimit.Middleware) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		limited := limiter.Limit(ratelimit.RouteAuthLogin, func(*http.Request) string { return "" }, next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost && r.URL.Path == authdelivery.PathLogin {
				limited.ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
