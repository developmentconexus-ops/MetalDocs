package httpdelivery_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	application "metaldocs/internal/modules/audit/application"
	httpdelivery "metaldocs/internal/modules/audit/delivery/http"
	"metaldocs/internal/modules/audit/infrastructure/memory"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/tenant"
)

// TestHandleEvents_405_Allow asserts that non-GET requests to /api/v1/audit/events
// return 405 with Allow: GET header (RFC 7231). RED before fix.
func TestHandleEvents_405_Allow(t *testing.T) {
	service := application.NewService(memory.NewWriter())
	handler := httpdelivery.NewHandler(service)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/audit/events", nil)
	ctx := tenant.WithTenantID(req.Context(), "tenant-test")
	ctx = iamdomain.WithAuthContext(ctx, "user-a", []iamdomain.Role{})
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d body=%s", rec.Code, rec.Body.String())
	}
	allow := rec.Header().Get("Allow")
	if allow == "" {
		t.Error("expected Allow header in 405 response (RFC 7231), got empty")
	}
}
