package http

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/tenant"
)

func TestExportPDF_InvalidPaperSize_Returns400(t *testing.T) {
	h := NewExportHandler(nil)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc-1/export/pdf", bytes.NewReader([]byte(`{"paper_size":"Ledger"}`)))
	req.SetPathValue("id", "doc-1")
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "user-1", []iamdomain.Role{}))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}
