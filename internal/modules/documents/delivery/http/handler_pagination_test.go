package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/modules/documents/application"
	httphandler "metaldocs/internal/modules/documents/delivery/http"
	"metaldocs/internal/modules/documents/domain"
)

// statsCapturingSvc wraps fakeSvc to record the userID the handler resolves and
// passes to DocumentStats, so the admin -> all-scope ("") resolution can be
// asserted (fakeSvc.DocumentStats itself discards the userID arg).
type statsCapturingSvc struct {
	*fakeSvc
	statsUser string
}

func (s *statsCapturingSvc) DocumentStats(ctx context.Context, tenantID, userID string, opts application.ListOptions) (*application.DocumentStats, error) {
	s.statsUser = userID
	return s.fakeSvc.DocumentStats(ctx, tenantID, userID, opts)
}

func TestListDocuments_Paginated_Envelope(t *testing.T) {
	svc := &fakeSvc{
		listPaginatedItems: []*domain.Document{{ID: "doc_1", Name: "Doc 1"}},
		listPaginatedTotal: 7,
	}
	mux := newMux(t, svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents?page=1&pageSize=5", nil)
	withAuthHeaders(req, "editor")
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
	withAuthHeaders(req, "editor")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestListDocuments_InvalidStatus_Returns400(t *testing.T) {
	mux := newMux(t, &fakeSvc{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents?status=not_real", nil)
	withAuthHeaders(req, "editor")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestDocumentStats_OK(t *testing.T) {
	svc := &statsCapturingSvc{fakeSvc: &fakeSvc{
		statsResult: &application.DocumentStats{
			ByStatus: map[string]int64{"draft": 2},
			ByArea:   map[string]int64{"RH": 1},
		},
	}}
	h := httphandler.NewHandler(svc).WithCaps(fakeCaps{admin: true})
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

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
	// admin must resolve to the all-scope: effectiveUserID == "" (mirrors
	// TestListDocuments_AdminSeesAllVsOwnScope's listPaginatedUser assertion).
	if svc.statsUser != "" {
		t.Fatalf("admin must see all: stats userID = %q, want \"\"", svc.statsUser)
	}
}
