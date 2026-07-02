package httpdelivery

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/platform/tenant"
)

// tenantCtxRequest builds a request carrying a valid tenant context, mirroring
// the fixture used by the sibling iam delivery/http tests
// (routes_roles_caps_test.go) so 501 memory-mode assertions below exercise the
// same happy-path context shape production requests carry.
func tenantCtxRequest(method, path string) *http.Request {
	ctx := tenant.WithTenantID(context.Background(), tenant.DevTenantID)
	return httptest.NewRequest(method, path, nil).WithContext(ctx)
}

// TestHandler_NilService_Returns501 locks down SEC-12: when the security
// module is constructed without a service (deps.SQLDB == nil / memory mode,
// see apps/api/cmd/metaldocs-api/main.go's `securitydelivery.NewHandler(nil)`
// fallback), every read endpoint must return 501 problem+json instead of a
// fabricated 200 with zero/empty data. Mirrors the sibling guard in
// internal/modules/iam/delivery/http/sessions_handler.go:88 and
// routes_roles_caps_test.go's TestListRoleCapabilitiesNilRepoIs501.
func TestHandler_NilService_Returns501(t *testing.T) {
	h := NewHandler(nil)

	cases := []struct {
		name string
		fn   func(http.ResponseWriter, *http.Request)
		path string
	}{
		{"mfa-coverage", h.handleMfaCoverage, "/api/v1/security/mfa-coverage"},
		{"lockouts", h.handleLockouts, "/api/v1/security/lockouts"},
		{"signals", h.handleSignals, "/api/v1/security/signals"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tc.fn(rec, tenantCtxRequest(http.MethodGet, tc.path))

			if got, want := rec.Code, http.StatusNotImplemented; got != want {
				t.Fatalf("%s: status = %d, want %d, body=%s", tc.name, got, want, rec.Body.String())
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
				t.Fatalf("%s: content-type = %q, want application/problem+json", tc.name, ct)
			}
		})
	}
}

// TestHandler_NilService_RequiresTenantFirst asserts the tenant guard runs
// before the nil-service guard: an unauthenticated caller against a
// misconfigured deployment still gets 401, not a 501 that would leak
// deployment-mode information pre-auth.
func TestHandler_NilService_RequiresTenantFirst(t *testing.T) {
	h := NewHandler(nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/security/mfa-coverage", nil)
	h.handleMfaCoverage(rec, req)

	if got, want := rec.Code, http.StatusUnauthorized; got != want {
		t.Fatalf("status = %d, want %d, body=%s", got, want, rec.Body.String())
	}
}
