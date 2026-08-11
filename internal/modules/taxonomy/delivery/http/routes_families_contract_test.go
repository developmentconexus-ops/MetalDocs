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
	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

type fakeFamilyService struct {
	createErr error
	updateErr error
}

func (f fakeFamilyService) List(_ context.Context, tenantID string, includeInactive bool) ([]domain.DocumentFamily, error) {
	return nil, nil
}
func (f fakeFamilyService) Get(_ context.Context, tenantID string, code domain.FamilyCode) (*domain.DocumentFamily, error) {
	return nil, domain.ErrFamilyNotFound
}
func (f fakeFamilyService) Create(_ context.Context, fam *domain.DocumentFamily) error {
	return f.createErr
}
func (f fakeFamilyService) Update(_ context.Context, fam *domain.DocumentFamily) (*domain.DocumentFamily, error) {
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	return fam, nil
}
func (f fakeFamilyService) Deactivate(_ context.Context, code domain.FamilyCode) error { return nil }

// TestFamiliesHandler_UpdateMissingActorReturns401 is the PR #108 review-round-1
// remediation for FINDING 1: FamilyService.Update (A3.3/T1) resolves the actor
// before any mutation work and returns authn.ErrMissingActor when it is absent,
// but writeFamilyError had no case for that sentinel and fell through to the
// default 500 internal.unknown arm — the documented contract is 401
// auth.unauthenticated, the same code every other actor-gated route answers
// with (see internal/platform/authn/context.go and
// internal/modules/tokens/delivery/http/handler.go's writeTokenError).
func TestFamiliesHandler_UpdateMissingActorReturns401(t *testing.T) {
	handler := &Handler{families: fakeFamilyService{updateErr: authn.ErrMissingActor}}
	mux := http.NewServeMux()
	handler.Mount(mux)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/taxonomy/families/F1", strings.NewReader(`{"name":"Family"}`))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
	var prob problem.Problem
	if err := json.Unmarshal(rec.Body.Bytes(), &prob); err != nil {
		t.Fatalf("decode problem body: %v (body=%s)", err, rec.Body.String())
	}
	if prob.Code != problem.CodeAuthUnauthenticated {
		t.Fatalf("problem code = %q, want %q", prob.Code, problem.CodeAuthUnauthenticated)
	}
}

func TestFamiliesHandler_GetMissing_Returns404(t *testing.T) {
	handler := &Handler{families: fakeFamilyService{}}
	mux := http.NewServeMux()
	handler.Mount(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/taxonomy/families/missing", nil)
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
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
	handler.Mount(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/taxonomy/families", nil)
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestFamiliesHandler_CreateUniqueViolationReturns409(t *testing.T) {
	handler := NewHandler(fakeProfileService{}, fakeAreaService{}, fakeFamilyService{createErr: &pgconn.PgError{Code: "23505"}}, newIdempotentMockDB(t))
	mux := http.NewServeMux()
	handler.Mount(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/taxonomy/families", strings.NewReader(`{"code":"F1","name":"Family"}`))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-test", []iamdomain.Role{iamdomain.RoleSystemAdmin}))
	req.Header.Set("Idempotency-Key", testIdempotencyKey)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestFamiliesHandler_CreateConstraintViolationReturnsGenericValidationMessage(t *testing.T) {
	handler := NewHandler(fakeProfileService{}, fakeAreaService{}, fakeFamilyService{createErr: &pgconn.PgError{Code: "23514", Message: "sensitive detail"}}, newIdempotentMockDB(t))
	mux := http.NewServeMux()
	handler.Mount(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/taxonomy/families", strings.NewReader(`{"code":"F1","name":"Family"}`))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-test", []iamdomain.Role{iamdomain.RoleSystemAdmin}))
	req.Header.Set("Idempotency-Key", testIdempotencyKey)
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
