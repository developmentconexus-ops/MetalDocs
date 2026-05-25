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

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc_1/export/pdf", bytes.NewReader([]byte(`{"paper_size":"A0"}`)))
	req.SetPathValue("id", "doc_1")
	req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant_1"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "user_1", []iamdomain.Role{iamdomain.Role("document_filler")}))
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
	}
}
