package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/taxonomy/application"
	"metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/iamtypes"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

// F-QA4-2 — role vocabularies are single-sourced from the contract
// (api/openapi UserRole / AreaRole, mirrored by internal/platform/iamtypes).
// These tests pin the DELIVERY half: a bound-vocabulary violation surfaces as a
// friendly 422 problem+json naming the accepted values, not a 400, a 500, or a
// raw DB constraint error.

func taxonomyWriteRequest(t *testing.T, method, target, body string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "actor-test", []iamdomain.Role{iamdomain.RoleSystemAdmin}))
	req.Header.Set("Idempotency-Key", testIdempotencyKey)
	return req
}

func decodeProblem(t *testing.T, rec *httptest.ResponseRecorder) problem.Problem {
	t.Helper()
	var p problem.Problem
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatalf("response body is not problem+json: %v (body=%s)", err, rec.Body.String())
	}
	return p
}

func TestProfilesHandler_InvalidEditableByRoleReturns422(t *testing.T) {
	// The domain rejects the free-text role; the handler must translate that
	// sentinel into 422 rather than letting it fall through to 500.
	createErr := fmt.Errorf("taxonomy: validate profile create: %w", domain.ErrInvalidEditableByRole)
	handler := NewHandler(fakeProfileService{createErr: createErr}, fakeAreaService{}, fakeFamilyService{}, newIdempotentMockDB(t))
	mux := http.NewServeMux()
	handler.Mount(mux)

	req := taxonomyWriteRequest(t, http.MethodPost, "/api/v1/taxonomy/profiles",
		`{"code":"PO","family_code":"quality","name":"Proc","description":"d","review_interval_days":365,"editable_by_role":"manager"}`)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusUnprocessableEntity, rec.Body.String())
	}
	p := decodeProblem(t, rec)
	if p.Code.String() != "validation.failed" {
		t.Fatalf("problem code = %q, want %q", p.Code, "validation.failed")
	}
	// The message is rendered from the registry, so it must name every canonical
	// role — including system_admin, which IS valid for editable_by_role.
	for _, role := range iamtypes.Roles() {
		if !strings.Contains(p.Title, string(role)) {
			t.Fatalf("message %q does not name canonical role %q", p.Title, role)
		}
	}
}

func TestAreasHandler_InvalidDefaultApproverRoleReturns422(t *testing.T) {
	createErr := fmt.Errorf("taxonomy: validate area create: %w", domain.ErrInvalidDefaultApproverRole)
	handler := NewHandler(fakeProfileService{}, fakeAreaService{createErr: createErr}, fakeFamilyService{}, newIdempotentMockDB(t))
	mux := http.NewServeMux()
	handler.Mount(mux)

	req := taxonomyWriteRequest(t, http.MethodPost, "/api/v1/taxonomy/areas",
		`{"code":"A1","name":"Area","description":"d","default_approver_role":"system_admin"}`)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusUnprocessableEntity, rec.Body.String())
	}
	p := decodeProblem(t, rec)
	if p.Code.String() != "validation.failed" {
		t.Fatalf("problem code = %q, want %q", p.Code, "validation.failed")
	}
	// system_admin is the whole point of the AreaRole subset: it must NOT be
	// offered as an accepted value.
	if strings.Contains(p.Title, string(iamtypes.RoleSystemAdmin)) {
		t.Fatalf("message %q offers system_admin, which is not area-assignable", p.Title)
	}
	for _, role := range iamtypes.AreaRoles() {
		if !strings.Contains(p.Title, string(role)) {
			t.Fatalf("message %q does not name area-assignable role %q", p.Title, role)
		}
	}
}

// ─── UPDATE path ────────────────────────────────────────────────────────────
//
// Create and update are separate seams: ProfileService.Update does NOT run
// NewDocumentProfile (it takes an already-normalized profile), so the create
// test alone would not prove the update path is fail-closed. These pin both
// handlers' error mapping for the update route.

func TestProfilesHandler_InvalidEditableByRoleOnUpdateReturns422(t *testing.T) {
	updateErr := fmt.Errorf("taxonomy: validate profile update: %w", domain.ErrInvalidEditableByRole)
	handler := NewHandler(fakeProfileService{updateErr: updateErr}, fakeAreaService{}, fakeFamilyService{}, newIdempotentMockDB(t))
	mux := http.NewServeMux()
	handler.Mount(mux)

	req := taxonomyWriteRequest(t, http.MethodPatch, "/api/v1/taxonomy/profiles/PO",
		`{"code":"PO","family_code":"quality","name":"Proc","description":"d","review_interval_days":365,"editable_by_role":"manager"}`)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusUnprocessableEntity, rec.Body.String())
	}
	if p := decodeProblem(t, rec); p.Code.String() != "validation.failed" {
		t.Fatalf("problem code = %q, want %q", p.Code, "validation.failed")
	}
}

func TestAreasHandler_SystemAdminDefaultApproverRoleOnUpdateReturns422(t *testing.T) {
	updateErr := fmt.Errorf("taxonomy: validate area update: %w", domain.ErrInvalidDefaultApproverRole)
	handler := NewHandler(fakeProfileService{}, fakeAreaService{updateErr: updateErr}, fakeFamilyService{}, newIdempotentMockDB(t))
	mux := http.NewServeMux()
	handler.Mount(mux)

	req := taxonomyWriteRequest(t, http.MethodPut, "/api/v1/taxonomy/areas/A1",
		`{"code":"A1","name":"Area","description":"d","default_approver_role":"system_admin"}`)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusUnprocessableEntity, rec.Body.String())
	}
	if p := decodeProblem(t, rec); p.Code.String() != "validation.failed" {
		t.Fatalf("problem code = %q, want %q", p.Code, "validation.failed")
	}
}

// ─── end-to-end validation wiring (no mocked sentinel) ──────────────────────
//
// The tests above inject the sentinel, which proves the HTTP mapping but not
// that the real services actually produce it. These drive the REAL application
// services, so the chain under test is:
//
//	HTTP body -> handler decode -> application service -> domain validator
//	          -> sentinel -> handler error mapping -> 422 problem+json
//
// No infrastructure is needed: both services validate BEFORE touching their
// repository, so a nil repository is never dereferenced on this path. That is
// itself part of the contract — validation must be the first thing the write
// path does. If someone moves the check below BeginTx, these tests panic
// instead of passing, which is the correct alarm.

type stubTemplateChecker struct{}

func (stubTemplateChecker) IsPublished(context.Context, string) (bool, string, error) {
	return true, "", nil
}

type stubGovernanceLogger struct{}

func (stubGovernanceLogger) Log(context.Context, domain.GovernanceEvent) error { return nil }
func (stubGovernanceLogger) LogTx(context.Context, db.Tx, domain.GovernanceEvent) error {
	return nil
}

// realUpdateProfileService serves reads from the fake (the update handler loads
// the current profile first, which does need a repository) and delegates the
// write to the real service, whose validation is what we are exercising.
type realUpdateProfileService struct {
	fakeProfileService
	real *application.ProfileService
}

func (s realUpdateProfileService) Update(ctx context.Context, p *domain.DocumentProfile) error {
	return s.real.Update(ctx, p)
}

type realUpdateAreaService struct {
	fakeAreaService
	real *application.AreaService
}

func (s realUpdateAreaService) Update(ctx context.Context, a *domain.ProcessArea) error {
	return s.real.Update(ctx, a)
}

func TestProfilesHandler_RealServiceRejectsNonCanonicalRoleWith422(t *testing.T) {
	svc := realUpdateProfileService{
		real: application.NewProfileService(nil, stubTemplateChecker{}, stubGovernanceLogger{}),
	}
	handler := NewHandler(svc, fakeAreaService{}, fakeFamilyService{}, newIdempotentMockDB(t))
	mux := http.NewServeMux()
	handler.Mount(mux)

	req := taxonomyWriteRequest(t, http.MethodPatch, "/api/v1/taxonomy/profiles/PO",
		`{"code":"PO","family_code":"quality","name":"Proc","description":"d","review_interval_days":365,"editable_by_role":"manager"}`)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusUnprocessableEntity, rec.Body.String())
	}
	p := decodeProblem(t, rec)
	if p.Code.String() != "validation.failed" {
		t.Fatalf("problem code = %q, want %q", p.Code, "validation.failed")
	}
	if !strings.Contains(p.Title, string(iamtypes.RoleEditor)) {
		t.Fatalf("message %q does not name the accepted roles", p.Title)
	}
}

func TestAreasHandler_RealServiceRejectsSystemAdminWith422(t *testing.T) {
	svc := realUpdateAreaService{real: application.NewAreaService(nil, stubGovernanceLogger{})}
	handler := NewHandler(fakeProfileService{}, svc, fakeFamilyService{}, newIdempotentMockDB(t))
	mux := http.NewServeMux()
	handler.Mount(mux)

	req := taxonomyWriteRequest(t, http.MethodPut, "/api/v1/taxonomy/areas/A1",
		`{"code":"A1","name":"Area","description":"d","default_approver_role":"system_admin"}`)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusUnprocessableEntity, rec.Body.String())
	}
	p := decodeProblem(t, rec)
	if strings.Contains(p.Title, string(iamtypes.RoleSystemAdmin)) {
		t.Fatalf("message %q offers system_admin, which is not area-assignable", p.Title)
	}
}

// The two vocabularies must stay distinct: AreaRole is exactly UserRole minus
// system_admin. If someone collapses them, this fails.
func TestRoleVocabulariesAreDistinct(t *testing.T) {
	all := iamtypes.Roles()
	area := iamtypes.AreaRoles()
	if len(all) != 8 {
		t.Fatalf("canonical roles = %d, want 8: %v", len(all), all)
	}
	if len(area) != len(all)-1 {
		t.Fatalf("area roles = %d, want %d: %v", len(area), len(all)-1, area)
	}
	for _, r := range area {
		if r == iamtypes.RoleSystemAdmin {
			t.Fatal("system_admin must not be area-assignable")
		}
		if !iamtypes.IsValidRole(r) {
			t.Fatalf("area role %q is not a canonical role", r)
		}
	}
}
