package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/modules/documents/application"
	"metaldocs/internal/modules/documents/domain"
)

func TestListDocuments_Paginated_Envelope(t *testing.T) {
	svc := &fakeSvc{
		listPaginatedItems: []*domain.Document{{ID: "doc_1", Name: "Doc 1"}},
		listPaginatedTotal: 7,
	}
	mux := newMux(t, svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents?page=1&pageSize=5", nil)
	withAuthHeaders(req, "document_filler")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var out struct {
		Items    []map[string]any `json:"items"`
		Page     int              `json:"page"`
		PageSize int              `json:"pageSize"`
		Total    int64            `json:"total"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out.Items) != 1 || out.Page != 1 || out.PageSize != 5 || out.Total != 7 {
		t.Fatalf("unexpected envelope: %+v", out)
	}
}

func TestListDocuments_PageSizeCap_Returns400(t *testing.T) {
	mux := newMux(t, &fakeSvc{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents?page=1&pageSize=999", nil)
	withAuthHeaders(req, "document_filler")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestListDocuments_InvalidStatus_Returns400(t *testing.T) {
	mux := newMux(t, &fakeSvc{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents?status=not_real", nil)
	withAuthHeaders(req, "document_filler")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestDocumentStats_OK(t *testing.T) {
	svc := &fakeSvc{
		statsResult: &application.DocumentStats{
			ByStatus: map[string]int64{"draft": 2},
			ByArea:   map[string]int64{"RH": 1},
		},
	}
	mux := newMux(t, svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/stats", nil)
	withAuthHeaders(req, "system_admin")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var out application.DocumentStats
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.ByStatus["draft"] != 2 || out.ByArea["RH"] != 1 {
		t.Fatalf("unexpected stats: %+v", out)
	}
}
