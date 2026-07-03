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

// TestHandleExport_405_Allow asserts the export POST endpoint advertises Allow: POST
// on a 405 (F7.1 drive-by, RFC 7231).
func TestHandleExport_405_Allow(t *testing.T) {
	service := application.NewService(memory.NewWriter())
	handler := httpdelivery.NewHandler(service)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit/events/export", nil)
	ctx := tenant.WithTenantID(req.Context(), "tenant-test")
	ctx = iamdomain.WithAuthContext(ctx, "user-a", []iamdomain.Role{})
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d body=%s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Allow"); got != "POST" {
		t.Errorf("Allow = %q, want POST", got)
	}
}

// TestHandleExportSubresource_405_Allow asserts the export subresource endpoint
// advertises an Allow header including GET on a 405 (F7.1 drive-by, RFC 7231).
// T-008: since the route mount moved to the generated auditapi.HandlerWithOptions
// router (method-qualified net/http 1.22 patterns), the stdlib mux itself produces
// this 405 and folds in HEAD alongside GET (GET registrations implicitly also
// match HEAD per net/http's ServeMux docs) — "GET, HEAD" is the RFC 7231-correct
// value, not a behavior regression from the mount-pattern migration.
func TestHandleExportSubresource_405_Allow(t *testing.T) {
	// Method check precedes the nil-exporter check, so the 405 fires without an exporter.
	service := application.NewService(memory.NewWriter())
	handler := httpdelivery.NewHandler(service)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/audit/events/export/some-id", nil)
	ctx := tenant.WithTenantID(req.Context(), "tenant-test")
	ctx = iamdomain.WithAuthContext(ctx, "user-a", []iamdomain.Role{})
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d body=%s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Allow"); got != "GET, HEAD" {
		t.Errorf("Allow = %q, want \"GET, HEAD\"", got)
	}
}
