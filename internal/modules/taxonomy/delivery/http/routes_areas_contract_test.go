package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/tenant"
)

type fakeAreaService struct {
	createErr error
	getErr    error
}

func (f fakeAreaService) List(ctx context.Context, tenantID string, includeArchived bool) ([]domain.ProcessArea, error) {
	return nil, nil
}

func (f fakeAreaService) Get(ctx context.Context, tenantID string, code domain.AreaCode) (*domain.ProcessArea, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return &domain.ProcessArea{Code: code, TenantID: tenantID, Name: "Area"}, nil
}

func (f fakeAreaService) Create(ctx context.Context, a *domain.ProcessArea) error {
	return f.createErr
}

func (f fakeAreaService) Update(ctx context.Context, a *domain.ProcessArea) error {
	return nil
}

func (f fakeAreaService) Archive(ctx context.Context, tenantID string, areaCode domain.AreaCode, actorID string) error {
	return nil
}

func TestAreasHandler_CreateUniqueViolationReturns409(t *testing.T) {
	handler := &Handler{areas: fakeAreaService{createErr: &pgconn.PgError{Code: "23505"}}}
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/taxonomy/areas", strings.NewReader(`{"code":"A1","name":"Area"}`))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestAreasHandler_UpdateReturns404WhenAreaMissing(t *testing.T) {
	handler := &Handler{areas: fakeAreaService{getErr: domain.ErrAreaNotFound}}
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/taxonomy/areas/A1", strings.NewReader(`{"name":"Area"}`))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
