package httpdelivery

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	iamapp "metaldocs/internal/modules/iam/application"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

// noopMembershipLogger satisfies MembershipGovernanceLogger with a no-op so
// contract tests that exercise error-envelope shapes can construct the service
// without a real audit sink.
type noopMembershipLogger struct{}

func (noopMembershipLogger) LogTx(_ context.Context, _ db.Tx, _ string, _ iamdomain.UserProcessArea) error {
	return nil
}

// noopMembershipTx satisfies iamdomain.MembershipTx for contract test repos.
type noopMembershipTx struct{}

func (noopMembershipTx) ExecContext(_ context.Context, _ string, _ ...any) (sql.Result, error) {
	return nil, nil
}
func (noopMembershipTx) QueryContext(_ context.Context, _ string, _ ...any) (*sql.Rows, error) {
	return nil, nil
}
func (noopMembershipTx) QueryRowContext(_ context.Context, _ string, _ ...any) *sql.Row {
	return nil
}
func (noopMembershipTx) Commit() error   { return nil }
func (noopMembershipTx) Rollback() error { return nil }

type fakeUserAreaWriteRepository struct{}

func (f fakeUserAreaWriteRepository) ListActive(ctx context.Context, userID, tenantID string, now time.Time) ([]iamdomain.UserProcessArea, error) {
	return nil, nil
}

func (f fakeUserAreaWriteRepository) ListByTenant(ctx context.Context, tenantID, userID, areaCode, role string, now time.Time) ([]iamdomain.UserProcessArea, error) {
	return nil, nil
}

func (f fakeUserAreaWriteRepository) MembershipDirectoryScope(ctx context.Context, tenantID, actorID, capability string, now time.Time) (bool, bool, error) {
	return false, false, nil
}

func (f fakeUserAreaWriteRepository) ListByTenantInManagedAreas(ctx context.Context, tenantID, userID, areaCode, role, actorID, capability string, now time.Time) ([]iamdomain.UserProcessArea, error) {
	return nil, nil
}

func (f fakeUserAreaWriteRepository) GetActiveByUserAndArea(ctx context.Context, userID, tenantID, areaCode string, now time.Time) (*iamdomain.UserProcessArea, error) {
	return nil, nil
}

func (f fakeUserAreaWriteRepository) BeginTx(_ context.Context) (iamdomain.MembershipTx, error) {
	return noopMembershipTx{}, nil
}

func (f fakeUserAreaWriteRepository) InsertTx(_ context.Context, _ iamdomain.MembershipTx, _ iamdomain.UserProcessArea) error {
	return nil
}

func (f fakeUserAreaWriteRepository) CloseActiveTx(_ context.Context, _ iamdomain.MembershipTx, _, _, _ string, _ time.Time, _ string) error {
	return nil
}

func (f fakeUserAreaWriteRepository) GrantAtomicTx(_ context.Context, _ iamdomain.MembershipTx, _, _ iamdomain.UserProcessArea) error {
	return nil
}

// passThroughVerifier accepts every (tenant, user) combination so the handler
// reaches the service layer in unit tests. Cross-tenant 404 enforcement lives
// in tests/integration/iam where the real PeopleService is exercised.
type passThroughVerifier struct{}

func (passThroughVerifier) VerifyUserInTenant(ctx context.Context, tenantID, userID string) error {
	return nil
}

// capDeniedAreaRepo simulates the real tier-2 outcome of ADR 0022 Phase 3: an
// active membership exists, but the area-scoped authz.Require denies the revoke
// (actor lacks membership.manage in the target area). GetActiveByUserAndArea
// must return a live row first so Revoke reaches CloseActive (else it short-
// circuits to ErrMembershipNotFound/404 before authz runs).
type capDeniedAreaRepo struct{ fakeUserAreaWriteRepository }

func (capDeniedAreaRepo) GetActiveByUserAndArea(ctx context.Context, userID, tenantID, areaCode string, now time.Time) (*iamdomain.UserProcessArea, error) {
	return &iamdomain.UserProcessArea{
		UserID:        userID,
		TenantID:      tenantID,
		AreaCode:      areaCode,
		Role:          iamdomain.RoleAuthor,
		EffectiveFrom: now.Add(-time.Hour),
	}, nil
}

func (capDeniedAreaRepo) CloseActiveTx(_ context.Context, _ iamdomain.MembershipTx, userID, tenantID, areaCode string, _ time.Time, actorID string) error {
	// Returns the bare ErrCapDenied value. The real repository wraps it once in
	// fmt.Errorf("...: %w", err); writeMembershipError uses errors.As, which
	// resolves both forms, so the unwrapped value is sufficient for this contract.
	return authz.ErrCapDenied{Capability: string(iamdomain.CapMembershipManage), AreaCode: areaCode, ActorID: actorID}
}

// TestMembershipsHandler_ErrorEnvelopeContract verifies the problem+json error
// envelope for a tier-2 area denial (ADR 0022 Phase 3). The former trigger was
// the removed RoleSystemAdmin handler gate; the equivalent forbidden outcome is
// now ErrCapDenied bubbling from the repository, mapped to 403 AUTH_FORBIDDEN.
func TestMembershipsHandler_ErrorEnvelopeContract(t *testing.T) {
	svc := iamapp.NewAreaMembershipService(capDeniedAreaRepo{}, noopMembershipLogger{})
	handler := NewMembershipHandler(svc, passThroughVerifier{}, nil)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/iam/area-memberships/user-1/ops", nil)
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "session-user", []iamdomain.Role{iamdomain.RoleAreaAdmin}))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}

	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("Content-Type = %q, want application/problem+json", ct)
	}

	var apiErr problem.Problem
	if err := json.Unmarshal(rec.Body.Bytes(), &apiErr); err != nil {
		t.Fatalf("unmarshal api error: %v body=%s", err, rec.Body.String())
	}
	if apiErr.Code != "AUTH_FORBIDDEN" {
		t.Fatalf("code = %q, want AUTH_FORBIDDEN", apiErr.Code)
	}
}

func TestMembershipsHandler_SystemAdminCanTargetOtherUser(t *testing.T) {
	svc := iamapp.NewAreaMembershipService(fakeUserAreaWriteRepository{}, noopMembershipLogger{})
	handler := NewMembershipHandler(svc, passThroughVerifier{}, nil)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/iam/area-memberships/user-1/ops", nil)
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "admin-1", []iamdomain.Role{iamdomain.RoleSystemAdmin}))
	req = req.WithContext(tenant.WithTenantID(req.Context(), "test-tenant"))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
