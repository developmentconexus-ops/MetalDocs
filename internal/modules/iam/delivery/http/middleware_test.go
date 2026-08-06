package httpdelivery

import (
	"net/http"
	"net/http/httptest"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/tenant"
)

func TestMiddlewareStripsTrustedIdentityHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents", nil)
	req.Header.Set("X-User-ID", "attacker")
	req.Header.Set("X-User-Roles", "system_admin")

	ctx := iamdomain.WithAuthContext(req.Context(), "real-user", []iamdomain.Role{iamdomain.RoleSystemAdmin})
	ctx = tenant.WithTenantID(ctx, tenant.DevTenantID)
	req = req.WithContext(ctx)

	resolver := func(method, path string) (iamdomain.Capability, Visibility) {
		return iamdomain.Capability(""), VisibilitySessionRequired
	}

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if got := iamdomain.UserIDFromContext(r.Context()); got != "real-user" {
			t.Fatalf("UserIDFromContext() = %q, want %q", got, "real-user")
		}
		if got := r.Header.Get("X-User-ID"); got != "" {
			t.Fatalf("X-User-ID header = %q, want empty", got)
		}
		if got := r.Header.Get("X-User-Roles"); got != "" {
			t.Fatalf("X-User-Roles header = %q, want empty", got)
		}
	})

	NewMiddleware(nil, nil, true).WithPermissionResolver(resolver).Wrap(next).ServeHTTP(httptest.NewRecorder(), req)

	if !called {
		t.Fatal("next handler was not called")
	}
}

// §6: VisibilityUnresolved gets its OWN explicit case, returning the same 500
// the default arm returns. Leaving it to fall through default: would work and
// would be wrong — a named, expected value silently relying on an
// unknown-value branch is a guard a future author closes by "completing" the
// switch, at which point an unwired route becomes a pass-through.
func TestWrap_VisibilityUnresolved_Is500(t *testing.T) {
	m := NewMiddleware(nil, nil, true).WithPermissionResolver(
		func(string, string) (iamdomain.Capability, Visibility) { return "", VisibilityUnresolved },
	)
	rr := httptest.NewRecorder()
	m.Wrap(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("handler must not be reached")
	})).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/anything", nil))
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rr.Code)
	}
}

// The default arm STAYS for genuinely unknown values.
func TestWrap_UnknownVisibility_Is500(t *testing.T) {
	m := NewMiddleware(nil, nil, true).WithPermissionResolver(
		func(string, string) (iamdomain.Capability, Visibility) { return "", Visibility(99) },
	)
	rr := httptest.NewRecorder()
	m.Wrap(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("handler must not be reached")
	})).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/anything", nil))
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rr.Code)
	}
}
