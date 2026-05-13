package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	apiv2 "metaldocs/internal/api/v2"
	"metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/tenant"
)

type fakeProfileService struct {
	createErr error
}

func (f fakeProfileService) List(ctx context.Context, tenantID string, includeArchived bool) ([]domain.DocumentProfile, error) {
	return nil, nil
}

func (f fakeProfileService) Get(ctx context.Context, tenantID, code string) (*domain.DocumentProfile, error) {
	return nil, domain.ErrProfileNotFound
}

func (f fakeProfileService) Create(ctx context.Context, p *domain.DocumentProfile) error {
	return f.createErr
}

func (f fakeProfileService) Update(ctx context.Context, p *domain.DocumentProfile) error {
	return nil
}

func (f fakeProfileService) SetDefaultTemplate(ctx context.Context, tenantID, profileCode, templateVersionID, actorID string) error {
	return nil
}

func (f fakeProfileService) Archive(ctx context.Context, tenantID, profileCode, actorID string) error {
	return nil
}

func TestProfilesHandler_ErrorEnvelopeContract(t *testing.T) {
	handler := &Handler{profiles: fakeProfileService{}}
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/taxonomy/profiles/missing", nil)
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	var apiErr apiv2.APIError
	if err := json.Unmarshal(rec.Body.Bytes(), &apiErr); err != nil {
		t.Fatalf("unmarshal api error: %v body=%s", err, rec.Body.String())
	}
	if apiErr.Code == "" {
		t.Fatalf("expected non-empty code in API error: %s", rec.Body.String())
	}
}

func TestProfilesHandler_UpdateUsesPatch(t *testing.T) {
	handler := &Handler{profiles: fakeProfileService{}}
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/taxonomy/profiles/foo", strings.NewReader(`{}`))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestProfilesHandler_CreateUniqueViolationReturns409(t *testing.T) {
	handler := &Handler{profiles: fakeProfileService{createErr: &pgconn.PgError{Code: "23505"}}}
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/taxonomy/profiles", strings.NewReader(`{"code":"P1","familyCode":"F1","name":"Profile"}`))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}
