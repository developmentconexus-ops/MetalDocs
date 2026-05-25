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
