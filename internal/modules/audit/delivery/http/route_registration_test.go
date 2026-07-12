package httpdelivery_test

// route_registration_test.go — T-008 (residual of CON-09) regression guard.
// Proves the audit module's 4 spec routes (list + 3 export ops) dispatch
// through the generated auditapi.HandlerWithOptions mount rather than the
// prior hand-written mux.HandleFunc registrations, and that the mount does
// not accidentally serve extra, unspec'd routes. Mirrors the approval
// CON-03 pin test (internal/modules/approval/http/router_test.go
// TestRegisterRoutes_AllRoutesRegistered).

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/modules/audit/application"
	httpdelivery "metaldocs/internal/modules/audit/delivery/http"
	"metaldocs/internal/modules/audit/infrastructure/memory"
)

func TestRegisterRoutes_AllSpecRoutesRegistered(t *testing.T) {
	t.Parallel()

	service := application.NewService(memory.NewWriter())
	handler := httpdelivery.NewHandler(service).WithExporter(service)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	routes := []struct {
		method string
		path   string
	}{
		// GET /audit/events — ListAuditEvents
		{method: http.MethodGet, path: "/api/v1/audit/events"},
		// POST /audit/events/export — ExportAuditEvents
		{method: http.MethodPost, path: "/api/v1/audit/events/export"},
		// GET /audit/events/export/{export_id} — GetAuditExportStatus
		{method: http.MethodGet, path: "/api/v1/audit/events/export/exp-1"},
		// GET /audit/events/export/{export_id}/download — DownloadAuditExport
		{method: http.MethodGet, path: "/api/v1/audit/events/export/exp-1/download?token=t"},
	}

	for _, rt := range routes {
		req := httptest.NewRequest(rt.method, rt.path, nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code == http.StatusNotFound {
			t.Errorf("route %s %s not registered (got 404)", rt.method, rt.path)
		}
	}
}

// TestRegisterRoutes_NoExtraRoutesServed proves the generated mount serves
// exactly the 4 spec'd audit operations and nothing else — a sibling,
// unrelated path (even one sharing the /api/v1/audit/events prefix) must
// still 404, so the mount can't be silently over-broad (e.g. a stray prefix
// registration reintroducing the old hand-written catch-all behavior).
func TestRegisterRoutes_NoExtraRoutesServed(t *testing.T) {
	t.Parallel()

	service := application.NewService(memory.NewWriter())
	handler := httpdelivery.NewHandler(service).WithExporter(service)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	notFoundRoutes := []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/api/v1/audit/events/unknown"},
		{method: http.MethodGet, path: "/api/v1/audit/unknown"},
		{method: http.MethodDelete, path: "/api/v1/audit/events/export/exp-1/download/extra"},
	}

	for _, rt := range notFoundRoutes {
		req := httptest.NewRequest(rt.method, rt.path, nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("route %s %s expected 404 (not part of the audit spec), got %d", rt.method, rt.path, w.Code)
		}
	}
}

// TestRegisterRoutes_WrapperValidationError exercises the generated wrapper's
// own ErrorHandlerFunc path (RegisterRoutes' inline handler): a malformed
// limit query param fails runtime.BindQueryParameterWithOptions inside
// ServerInterfaceWrapper.ListAuditEvents before Handler.ListAuditEvents (and
// therefore handleEvents) is ever reached, so this is answered as plain 400
// VALIDATION_ERROR — not RFC 9457 problem+json, since the wrapper's binding
// error path predates any call into problem.Write. This pins the new
// wrapper-level failure mode introduced by mounting via HandlerWithOptions.
func TestRegisterRoutes_WrapperValidationError(t *testing.T) {
	t.Parallel()

	service := application.NewService(memory.NewWriter())
	handler := httpdelivery.NewHandler(service)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit/events?limit=not-a-number", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d (body=%s)", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

// TestRegisterRoutes_HandlerLevelValidationStillProblemJSON proves
// handleEvents' own validation (reached once the wrapper's param binding
// succeeds) still answers RFC 9457 application/problem+json via
// writeProblem/VALIDATION_ERROR — unchanged from pre-migration behavior (see
// TestAuditHandler_InvalidLimitIsProblemJSON in handler_test.go). Uses
// occurred_after (a string param the wrapper binds successfully but
// handleEvents' parseTime rejects) to reach past the wrapper into the
// delegated handler logic.
func TestRegisterRoutes_HandlerLevelValidationStillProblemJSON(t *testing.T) {
	t.Parallel()

	service := application.NewService(memory.NewWriter())
	handler := httpdelivery.NewHandler(service)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := authenticatedAuditRequest(http.MethodGet, "/api/v1/audit/events?occurred_after=not-a-timestamp", "tenant-a")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d (body=%s)", w.Code, http.StatusBadRequest, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("Content-Type: got %q, want application/problem+json", ct)
	}
}
