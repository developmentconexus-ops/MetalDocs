package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	authapp "metaldocs/internal/modules/auth/application"
	authdelivery "metaldocs/internal/modules/auth/delivery/http"
	"metaldocs/internal/platform/observability"
)

// §10.3 rows 2-3's chain-level assertions: they require driving the REAL
// authn layer (apiChain's "authn" link) in front of the mux, because the
// live behavior change is that authn — not the mux — is what turns an
// unmounted/unmatched health path into a 401. A mux-level test alone (see
// health_routing_test.go's TestHealthzIsGone, which correctly asserts 404
// from the bare mux) cannot exercise this: it never runs authn at all.
//
// The composed handler here is buildChain(mux, apiChain(...)) with authn as
// the one real, non-nil link — the same authdelivery.Middleware main.go
// wires, using the real generated-httpSurface-backed newPermissionResolver
// via WithPublicPathChecker (permissions.go); a nil checker is no longer an
// option (WithPublicPathChecker is mandatory). Every other link is nil
// (buildChain skips nil links), so a request that reaches the mux proves
// authn treated it as public; a 401 proves authn rejected it before the mux
// was ever consulted.
func buildHealthDeltaChain(t *testing.T) http.Handler {
	t.Helper()

	mux := http.NewServeMux()
	observability.NewHealthHandler(nil).Mount(mux)

	surface := httpSurface
	permResolver := newPermissionResolver(mux, &surface)
	authMiddleware := authdelivery.NewMiddleware(nil, authapp.Config{SessionCookieName: "sid"}, true).
		WithPublicPathChecker(newPublicPathChecker(permResolver))

	return buildChain(mux, apiChain(
		nil, // panic_recovery
		nil, // otel
		nil, // http_obs
		nil, // cors
		nil, // origin_protection
		nil, // pre_auth_login_rate_limit
		authMiddleware.Wrap,
		nil, // iam_authz
		nil, // presence_bump
		nil, // rate_limit
		nil, // method_not_allowed
	))
}

// §10.3 row 2: unauthenticated GET /api/v1/health/<unknown> is 404 from the
// bare mux (no such pattern), but the real chain never lets it reach the mux
// unauthenticated — routeRules has no entry for it, so it fails closed to
// VisibilitySessionRequired, and authn returns 401 before the mux runs.
func TestHealthDelta_UnauthenticatedUnknownHealthPath_Returns401(t *testing.T) {
	h := buildHealthDeltaChain(t)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/health/unknown", nil))
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("GET /api/v1/health/unknown (no session) = %d, want 401", rr.Code)
	}
}

// §10.3 row 3: /healthz is deleted, not excepted (operator ruling C). Its
// routeRules entry is gone too (permissions.go), so it now falls through to
// the fail-closed default (VisibilitySessionRequired) instead of being
// classified public — authn rejects it with 401 before the mux, which no
// longer has a pattern for /healthz at all, is ever consulted.
func TestHealthDelta_UnauthenticatedHealthz_Returns401(t *testing.T) {
	h := buildHealthDeltaChain(t)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("GET /healthz (no session) = %d, want 401", rr.Code)
	}
}

// Control: the two real health paths stay public and reach the mux
// unauthenticated, proving the 401s above come from routeRules
// classification, not from authn rejecting every path indiscriminately.
func TestHealthDelta_UnauthenticatedRealHealthPaths_StayPublic(t *testing.T) {
	h := buildHealthDeltaChain(t)
	for _, path := range []string{"/api/v1/health/live", "/api/v1/health/ready"} {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, path, nil))
		if rr.Code != http.StatusOK {
			t.Errorf("GET %s (no session) = %d, want 200 (public)", path, rr.Code)
		}
	}
}
