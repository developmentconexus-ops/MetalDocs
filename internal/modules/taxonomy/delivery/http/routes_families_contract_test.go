package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"metaldocs/internal/modules/taxonomy/domain"
)

type fakeFamilyService struct {
	createErr error
}

func (f fakeFamilyService) List(_ context.Context, includeInactive bool) ([]domain.DocumentFamily, error) {
	return nil, nil
}
func (f fakeFamilyService) Get(_ context.Context, code domain.FamilyCode) (*domain.DocumentFamily, error) {
	return nil, domain.ErrFamilyNotFound
}
func (f fakeFamilyService) Create(_ context.Context, fam *domain.DocumentFamily) error {
	return f.createErr
}
func (f fakeFamilyService) Update(_ context.Context, fam *domain.DocumentFamily) (*domain.DocumentFamily, error) {
	return fam, nil
}
func (f fakeFamilyService) Deactivate(_ context.Context, code domain.FamilyCode) error { return nil }

func TestFamiliesHandler_GetMissing_Returns404(t *testing.T) {
	handler := &Handler{families: fakeFamilyService{}}
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/taxonomy/families/missing", nil)
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

	req := httptest.NewRequest(http.MethodGet, "/api/v1/taxonomy/families", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestFamiliesHandler_CreateUniqueViolationReturns409(t *testing.T) {
	handler := &Handler{families: fakeFamilyService{createErr: &pgconn.PgError{Code: "23505"}}}
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/taxonomy/families", strings.NewReader(`{"code":"F1","name":"Family"}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestFamiliesHandler_CreateConstraintViolationReturnsGenericValidationMessage(t *testing.T) {
	handler := &Handler{families: fakeFamilyService{createErr: &pgconn.PgError{Code: "23514", Message: "sensitive detail"}}}
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/taxonomy/families", strings.NewReader(`{"code":"F1","name":"Family"}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if strings.Contains(rec.Body.String(), "sensitive detail") {
		t.Fatalf("response must not include pg constraint detail: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "request violates data constraints") {
		t.Fatalf("response must include generic constraint message: %s", rec.Body.String())
	}
}
