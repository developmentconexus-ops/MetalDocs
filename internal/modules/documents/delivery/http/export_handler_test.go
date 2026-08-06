package http

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/tenant"
)

// Business behavior, called directly (Task 18 step 3): the real mount for
// POST /documents/{id}/export/pdf is documentsapi.HandlerWithOptions via
// routes_generated.go -> h.export.exportPDF, not ExportHandler.RegisterRoutes
// (deleted, orphaned). No mux is a subject here -- this asserts the paper-size
// validation maps to 400, not routing shape.
func TestExportPDF_InvalidPaperSize_Returns400(t *testing.T) {
	h := NewExportHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc-1/export/pdf", bytes.NewReader([]byte(`{"paper_size":"Ledger"}`)))
	req.SetPathValue("id", "doc-1")
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "user-1", []iamdomain.Role{}))
	rec := httptest.NewRecorder()

	h.exportPDF(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}
