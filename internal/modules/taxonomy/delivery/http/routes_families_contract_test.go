package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/modules/taxonomy/domain"
)

type fakeFamilyService struct{}

func (f fakeFamilyService) List(_ context.Context, includeInactive bool) ([]domain.DocumentFamily, error) {
	return nil, nil
}
func (f fakeFamilyService) Get(_ context.Context, code string) (*domain.DocumentFamily, error) {
	return nil, domain.ErrFamilyNotFound
}
func (f fakeFamilyService) Create(_ context.Context, fam *domain.DocumentFamily) error { return nil }
func (f fakeFamilyService) Update(_ context.Context, fam *domain.DocumentFamily) (*domain.DocumentFamily, error) {
	return fam, nil
}
func (f fakeFamilyService) Deactivate(_ context.Context, code string) error            { return nil }

func TestFamiliesHandler_GetMissing_Returns404(t *testing.T) {
	handler := &Handler{families: fakeFamilyService{}}
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v2/taxonomy/families/missing", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("Content-Type = %q, want application/problem+json", ct)
	}
	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["code"] == nil {
		t.Fatalf("expected error code in body: %s", rec.Body.String())
	}
}

func TestFamiliesHandler_ListReturns200(t *testing.T) {
	handler := &Handler{families: fakeFamilyService{}}
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v2/taxonomy/families", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}
