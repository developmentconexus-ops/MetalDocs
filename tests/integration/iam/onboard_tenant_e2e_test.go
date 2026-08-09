//go:build integration

// onboard_tenant_e2e_test.go — M7 F7.2 Task D: the spec.md validation-gate row
// "End-to-end: onboard → login as new admin → capability-gated action
// succeeds; cross-tenant → 404" (f7.2-onboarding/spec.md line 54), proven at
// the service/DB layer (the HTTP-level onboard→login→gated-action pass is
// covered separately by the live QA drive per spec.md §6.2 — this test does
// not spin up an HTTP server; e2e_happy_test.go's harness is for the async
// outbox/notifications scenarios and is not reused here).
//
// TestOnboardTenant_EndToEnd_LoginAndAct proves, in sequence:
//
//  1. OnboardTenant (iamapp.OnboardTenantService, same helper as
//     onboard_tenant_test.go) provisions a fresh tenant + admin identity in
//     one tx.
//  2. The new admin authenticates via the auth module's own published
//     Authenticate path (authapp.Service backed by the real Postgres
//     authpg.Repository — the same construction auth's own service tests use
//     against authmemory.Repository, here against Postgres) — proving the
//     Argon2id hash OnboardTenant wrote is a real, verifiable credential. The
//     admin is provisioned with MustChangePassword=true; Authenticate does
//     NOT hard-block on that flag (service.go: MustChangePassword only gates
//     ChangePassword's currentPassword requirement, not Authenticate), so
//     first login is expected to succeed — documented explicitly below.
//  3. A capability-gated action for the NEW tenant succeeds: in a fresh tx,
//     authz.SeedTxIdentity(new tenant, admin_user_id) + authz.Require
//     ("user.manage","tenant") passes, because OnboardTenant granted the
//     admin system_admin in the new tenant (ADR 0022 R1 tier-2 inheritance
//     bypass) — mirroring onboard_tenant_test.go's authz assertion idiom.
//  4. Cross-tenant isolation: the SAME admin identity seeded under a
//     DIFFERENT tenant id is denied by authz.Require (ErrCapDenied) — the
//     admin holds no role at all in the foreign tenant, encoding the
//     "cross-tenant → 404" bar at the service/authz level (the HTTP 404 itself
//     is a routing/handler concern proven by the live QA drive).
//
// Run: go test -tags=integration ./tests/integration/iam/... -run TestOnboardTenant_EndToEnd_LoginAndAct
package iam_test

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"testing"
	"time"

	authapp "metaldocs/internal/modules/auth/application"
	authdomain "metaldocs/internal/modules/auth/domain"
	authpg "metaldocs/internal/modules/auth/infrastructure/postgres"
	iamapp "metaldocs/internal/modules/iam/application"
	iamauthz "metaldocs/internal/modules/iam/authz"
	iampg "metaldocs/internal/modules/iam/infrastructure/postgres"
	"metaldocs/tests/integration/testdb"
)

// noopLoginCtxPortE2E satisfies iamdomain.LoginContextPort with a no-op —
// this test only needs Authenticate to succeed, not the governance hint it
// best-effort records (RecordLoginContext failures are already swallowed by
// the real service, so a real port is unnecessary here).
type noopLoginCtxPortE2E struct{}

func (noopLoginCtxPortE2E) RecordLoginContext(_ context.Context, _, _, _, _, _ string) error {
	return nil
}

func newAuthServiceForE2E(t *testing.T, sqlDB *sql.DB) *authapp.Service {
	t.Helper()
	authRepo := authpg.NewRepository(sqlDB, iampg.NewUserTenantRepository(sqlDB))
	roleProvider := iampg.NewRoleProvider(sqlDB)
	roleAdmin := iampg.NewRoleAdminRepository(sqlDB)
	svc, err := authapp.NewService(authRepo, roleProvider, roleAdmin, noopLoginCtxPortE2E{}, authapp.Config{
		SessionCookieName:      "metaldocs_session",
		SessionTTL:             time.Hour,
		SessionSecret:          "onboard-e2e-test-secret-32-bytes-min",
		PasswordMinLength:      8,
		LoginMaxFailedAttempts: 5,
		LoginLockDuration:      5 * time.Minute,
		// The admin's ONLY role is in the freshly onboarded tenant; without
		// this fallback, resolveLoginTenant would still succeed because the
		// admin has exactly one tenant (len(tenants)==1 short-circuits before
		// the fallback branch) — kept false here regardless, matching prod
		// defaults, since the single-tenant path is what's actually exercised.
		AllowDevTenantFallback: false,
	})
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}
	return svc
}

func loginRequestForE2E() *http.Request {
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	return req
}

// TestOnboardTenant_EndToEnd_LoginAndAct is the spec.md validation-gate proof
// for F7.2's binding end-to-end bar (line 54): onboard → login as the new
// admin → capability-gated action succeeds; cross-tenant → denied.
func TestOnboardTenant_EndToEnd_LoginAndAct(t *testing.T) {
	sqlDB := openDB(t)
	ctx := context.Background()

	// --- Step 1: onboard a fresh tenant + admin -----------------------------
	actor := testdb.DeterministicID(t, "onboard-e2e-actor")
	seedIdentity(t, sqlDB, actor)
	seedSystemAdminRole(t, sqlDB, actor)

	slug := "onboard-e2e-" + testdb.DeterministicID(t, "onboard-e2e-slug")
	adminUserID := testdb.DeterministicID(t, "onboard-e2e-admin")
	adminPassword := "correct-horse-battery-staple-e2e"
	t.Cleanup(func() { cleanupOnboardedTenant(sqlDB, slug, adminUserID) })

	onboardSvc := newOnboardTenantService(sqlDB)
	result, err := onboardSvc.OnboardTenant(ctx, iamapp.OnboardTenantInput{
		Name:             "Onboard E2E Tenant",
		Slug:             slug,
		AdminUserID:      adminUserID,
		AdminDisplayName: "Onboard E2E Admin",
		AdminPassword:    adminPassword,
	}, actor, devTenant) // actorTenantID = the ACTOR's tenant — grants are tenant-scoped
	if err != nil {
		t.Fatalf("OnboardTenant() = %v, want nil", err)
	}
	newTenantID := result.TenantID
	if newTenantID == "" {
		t.Fatalf("OnboardTenant() result.TenantID is empty")
	}
	if result.AdminUserID != adminUserID {
		t.Fatalf("OnboardTenant() result.AdminUserID = %q, want %q", result.AdminUserID, adminUserID)
	}

	// --- Step 2: the new admin logs in with the issued credential ----------
	// Documented behavior: OnboardTenant provisions the identity with
	// MustChangePassword=true, but auth's Authenticate does not gate on that
	// flag — it only affects the subsequent ChangePassword call's
	// currentPassword requirement (service.go ~L475/483). So first login with
	// the admin_password issued at onboarding is expected to SUCCEED here;
	// enforcing a forced password-change redirect is a front-end/handler
	// concern outside OnboardTenant's or Authenticate's contract.
	authSvc := newAuthServiceForE2E(t, sqlDB)
	session, err := authSvc.Authenticate(ctx, adminUserID, adminPassword, loginRequestForE2E())
	if err != nil {
		t.Fatalf("Authenticate(new admin, issued password) = %v, want nil (first login succeeds; MustChangePassword does not block Authenticate)", err)
	}
	if session.CurrentUser.UserID != adminUserID {
		t.Fatalf("Authenticate() CurrentUser.UserID = %q, want %q", session.CurrentUser.UserID, adminUserID)
	}
	if session.CurrentUser.TenantID != newTenantID {
		t.Fatalf("Authenticate() CurrentUser.TenantID = %q, want %q (must resolve to the newly onboarded tenant)", session.CurrentUser.TenantID, newTenantID)
	}
	if session.RawToken == "" {
		t.Fatalf("Authenticate() RawToken is empty, want a session token")
	}

	// A wrong password must still be rejected — the identity is a real,
	// verifiable Argon2id credential, not a stub.
	if _, err := authSvc.Authenticate(ctx, adminUserID, "definitely-wrong-password", loginRequestForE2E()); !errors.Is(err, authdomain.ErrInvalidCredentials) {
		t.Fatalf("Authenticate(new admin, wrong password) = %v, want authdomain.ErrInvalidCredentials", err)
	}

	// --- Step 3: capability-gated action succeeds for the NEW tenant -------
	// The admin holds system_admin in newTenantID (granted atomically by
	// OnboardTenant) — authz.Require("user.manage","tenant") must pass via
	// the ADR 0022 R1 tier-2 inheritance bypass, mirroring the authz
	// assertion idiom onboard_tenant_test.go and iam_users_tripwire_test.go
	// already use for tier-2 proofs.
	func() {
		tx, err := sqlDB.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("begin tx (gated action, new tenant): %v", err)
		}
		defer tx.Rollback() //nolint:errcheck

		if err := iamauthz.SeedTxIdentity(ctx, tx, newTenantID, adminUserID); err != nil {
			t.Fatalf("SeedTxIdentity(new tenant, admin) = %v, want nil", err)
		}
		if err := iamauthz.Require(ctx, tx, "user.manage", "tenant"); err != nil {
			t.Fatalf("Require(user.manage, new tenant) = %v, want nil (admin is system_admin of the new tenant)", err)
		}
	}()

	// --- Step 4: cross-tenant isolation — denied under a foreign tenant ----
	// Same admin identity, seeded (GUC-only, no role granted) under a
	// DIFFERENT tenant id: the admin holds zero roles/capabilities there, so
	// authz.Require must deny — the service/authz-level encoding of the
	// "cross-tenant → 404" bar (the HTTP 404 itself is proven by the live
	// QA drive per spec.md §6.2).
	foreignTenantID := devTenant
	if foreignTenantID == newTenantID {
		t.Fatalf("test setup invariant violated: foreign tenant must differ from the newly onboarded tenant")
	}
	func() {
		tx, err := sqlDB.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("begin tx (cross-tenant probe): %v", err)
		}
		defer tx.Rollback() //nolint:errcheck

		if err := iamauthz.SeedTxIdentity(ctx, tx, foreignTenantID, adminUserID); err != nil {
			t.Fatalf("SeedTxIdentity(foreign tenant, admin) = %v, want nil", err)
		}
		err = iamauthz.Require(ctx, tx, "user.manage", "tenant")
		if !isCapDenied(err) {
			t.Fatalf("Require(user.manage, foreign tenant) = %v, want authz.ErrCapDenied (cross-tenant isolation)", err)
		}
	}()
}
