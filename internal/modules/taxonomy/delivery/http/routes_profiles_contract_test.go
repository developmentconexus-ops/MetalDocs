package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

type fakeProfileService struct {
	createErr error
	getErr    error
}

func (f fakeProfileService) List(ctx context.Context, tenantID string, includeArchived bool) ([]domain.DocumentProfile, error) {
	return nil, nil
}

func (f fakeProfileService) Get(ctx context.Context, tenantID string, code domain.ProfileCode) (*domain.DocumentProfile, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return &domain.DocumentProfile{Code: code, TenantID: tenantID}, nil
}

func (f fakeProfileService) Create(ctx context.Context, p *domain.DocumentProfile) error {
	return f.createErr
}

func (f fakeProfileService) Update(ctx context.Context, p *domain.DocumentProfile) error {
	return nil
}

func (f fakeProfileService) Reclassify(ctx context.Context, tenantID string, profileCode domain.ProfileCode, newClass domain.GovernanceClass, actorID string) error {
	return nil
}

func (f fakeProfileService) SetDefaultTemplate(ctx context.Context, tenantID string, profileCode domain.ProfileCode, templateVersionID, actorID string) error {
	return nil
}

func (f fakeProfileService) Archive(ctx context.Context, tenantID string, profileCode domain.ProfileCode, actorID string) error {
	return nil
}

func TestProfilesHandler_ErrorEnvelopeContract(t *testing.T) {
	handler := &Handler{profiles: fakeProfileService{getErr: domain.ErrProfileNotFound}}
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/taxonomy/profiles/missing", nil)
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	var apiErr problem.Problem
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

func TestProfilesHandler_UpdateReturns404WhenProfileMissing(t *testing.T) {
	handler := &Handler{profiles: fakeProfileService{getErr: domain.ErrProfileNotFound}}
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/taxonomy/profiles/foo", strings.NewReader(`{}`))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestProfilesHandler_CreateUniqueViolationReturns409(t *testing.T) {
	handler := NewHandler(fakeProfileService{createErr: &pgconn.PgError{Code: "23505"}}, fakeAreaService{}, fakeFamilyService{}, newIdempotentMockDB(t))
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/taxonomy/profiles", strings.NewReader(`{"code":"P1","familyCode":"F1","name":"Profile"}`))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-test", []iamdomain.Role{iamdomain.RoleSystemAdmin}))
	req.Header.Set("Idempotency-Key", testIdempotencyKey)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}
