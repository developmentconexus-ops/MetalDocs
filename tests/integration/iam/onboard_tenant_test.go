//go:build integration

// onboard_tenant_test.go — M7 F7.2 Task C TDD (RED-first): drives
// iamapp.OnboardTenantService end to end against a real Postgres instance.
// Three properties proven, mirroring the established iam integration
// patterns (tenants_tripwire_test.go for the tripwire-arm side,
// iam_users_tripwire_test.go / membership_area_scope_test.go for the
// tier-2-authz + seeded-actor idiom):
//
//   - TestOnboardTenant_CreatesTenantAdminAndAudit: a system_admin actor
//     onboards a new tenant; the tenants/iam_users/iam_user_roles/
//     auth_identities rows and the tenant.onboarded audit event all land
//     atomically.
//
//   - TestOnboardTenant_DuplicateSlugConflict: onboarding twice with the same
//     slug returns domain.ErrTenantSlugConflict (409 at the handler layer).
//
//   - TestOnboardTenant_RequiresCapability: an actor holding no tenant.onboard
//     grant is denied via authz.ErrCapDenied before any row is written.
//
//     go test -tags=integration ./tests/integration/iam/... -run TestOnboardTenant
package iam_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	auditpg "metaldocs/internal/modules/audit/infrastructure/postgres"
	authpg "metaldocs/internal/modules/auth/infrastructure/postgres"
	iamapp "metaldocs/internal/modules/iam/application"
	iamdomain "metaldocs/internal/modules/iam/domain"
	iampg "metaldocs/internal/modules/iam/infrastructure/postgres"
	"metaldocs/internal/platform/db"
	"metaldocs/tests/integration/testdb"
)

func isTenantSlugConflict(err error) bool {
	return errors.Is(err, iamdomain.ErrTenantSlugConflict)
}

func newOnboardTenantService(sqlDB *sql.DB) *iamapp.OnboardTenantService {
	authRepo := authpg.NewRepository(sqlDB, iampg.NewUserTenantRepository(sqlDB))
	return iamapp.NewOnboardTenantService(
		iampg.NewTenantRepository(sqlDB),
		authRepo,
		db.NewTxRunner(sqlDB),
		auditpg.NewWriter(sqlDB),
		nil, // NoopTenantKeyProvisioner
		nil, // defaults to authapp.HashPassword (auth's published bcrypt seam)
	)
}

// TestOnboardTenant_CreatesTenantAdminAndAudit proves the full onboarding tx:
// tenant + admin iam_users + system_admin role + auth identity + audit event,
// all committed atomically by a system_admin actor.
func TestOnboardTenant_CreatesTenantAdminAndAudit(t *testing.T) {
	sqlDB := openDB(t)
	ctx := context.Background()

	actor := testdb.DeterministicID(t, "onboard-actor")
	seedIdentity(t, sqlDB, actor)
	seedSystemAdminRole(t, sqlDB, actor)

	slug := "onboard-ok-" + testdb.DeterministicID(t, "onboard-ok-slug")
	adminUserID := testdb.DeterministicID(t, "onboard-ok-admin")
	t.Cleanup(func() { cleanupOnboardedTenant(sqlDB, slug, adminUserID) })

	svc := newOnboardTenantService(sqlDB)
	result, err := svc.OnboardTenant(ctx, iamapp.OnboardTenantInput{
		Name:             "Onboard OK Tenant",
		Slug:             slug,
		AdminUserID:      adminUserID,
		AdminDisplayName: "Onboard OK Admin",
		AdminPassword:    "correct-horse-battery-staple",
	}, actor, devTenant) // actorTenantID = the ACTOR's tenant — grants are tenant-scoped
	if err != nil {
		t.Fatalf("OnboardTenant() = %v, want nil", err)
	}
	if result.TenantID == "" {
		t.Fatalf("OnboardTenant() result.TenantID is empty")
	}
	if result.AdminUserID != adminUserID {
		t.Fatalf("OnboardTenant() result.AdminUserID = %q, want %q", result.AdminUserID, adminUserID)
	}

	var tenantName string
	if err := sqlDB.QueryRowContext(ctx,
		`SELECT name FROM metaldocs.tenants WHERE id = $1::uuid`, result.TenantID,
	).Scan(&tenantName); err != nil {
		t.Fatalf("read tenants row: %v", err)
	}
	if tenantName != "Onboard OK Tenant" {
		t.Fatalf("tenants.name = %q, want %q", tenantName, "Onboard OK Tenant")
	}

	var iamDisplayName, iamTenantID string
	if err := sqlDB.QueryRowContext(ctx,
		`SELECT display_name, tenant_id::text FROM metaldocs.iam_users WHERE user_id = $1`, adminUserID,
	).Scan(&iamDisplayName, &iamTenantID); err != nil {
		t.Fatalf("read iam_users row: %v", err)
	}
	if iamDisplayName != "Onboard OK Admin" {
		t.Fatalf("iam_users.display_name = %q, want %q", iamDisplayName, "Onboard OK Admin")
	}
	if iamTenantID != result.TenantID {
		t.Fatalf("iam_users.tenant_id = %q, want %q", iamTenantID, result.TenantID)
	}

	var roleCode string
	if err := sqlDB.QueryRowContext(ctx,
		`SELECT role_code FROM metaldocs.iam_user_roles WHERE user_id = $1 AND tenant_id = $2::uuid`,
		adminUserID, result.TenantID,
	).Scan(&roleCode); err != nil {
		t.Fatalf("read iam_user_roles row: %v", err)
	}
	if roleCode != "system_admin" {
		t.Fatalf("iam_user_roles.role_code = %q, want %q", roleCode, "system_admin")
	}

	var identityCount int
	if err := sqlDB.QueryRowContext(ctx,
		`SELECT count(*) FROM metaldocs.auth_identities WHERE user_id = $1 AND password_algo = 'bcrypt'`,
		adminUserID,
	).Scan(&identityCount); err != nil {
		t.Fatalf("read auth_identities row: %v", err)
	}
	if identityCount != 1 {
		t.Fatalf("auth_identities rows for %s = %d, want 1", adminUserID, identityCount)
	}

	var auditCount int
	if err := sqlDB.QueryRowContext(ctx,
		`SELECT count(*) FROM metaldocs.audit_events
		  WHERE action = 'tenant.onboarded' AND resource_id = $1 AND tenant_id = $1`,
		result.TenantID,
	).Scan(&auditCount); err != nil {
		t.Fatalf("read audit_events row: %v", err)
	}
	if auditCount != 1 {
		t.Fatalf("audit_events tenant.onboarded rows for %s = %d, want 1", result.TenantID, auditCount)
	}

	var auditPayload string
	if err := sqlDB.QueryRowContext(ctx,
		`SELECT payload::text FROM metaldocs.audit_events
		  WHERE action = 'tenant.onboarded' AND resource_id = $1`,
		result.TenantID,
	).Scan(&auditPayload); err != nil {
		t.Fatalf("read audit_events payload: %v", err)
	}
	if containsSubstring(auditPayload, "admin_password") || containsSubstring(auditPayload, "correct-horse-battery-staple") {
		t.Fatalf("audit_events payload leaked admin_password material: %s", auditPayload)
	}
}

// TestOnboardTenant_DuplicateSlugConflict proves the second onboarding
// attempt with an already-used slug is rejected with
// iamdomain.ErrTenantSlugConflict (409 at the handler layer) and does not
// create a second tenant/admin pair.
func TestOnboardTenant_DuplicateSlugConflict(t *testing.T) {
	sqlDB := openDB(t)
	ctx := context.Background()

	actor := testdb.DeterministicID(t, "onboard-dup-actor")
	seedIdentity(t, sqlDB, actor)
	seedSystemAdminRole(t, sqlDB, actor)

	slug := "onboard-dup-" + testdb.DeterministicID(t, "onboard-dup-slug")
	firstAdmin := testdb.DeterministicID(t, "onboard-dup-admin-1")
	secondAdmin := testdb.DeterministicID(t, "onboard-dup-admin-2")
	t.Cleanup(func() {
		cleanupOnboardedTenant(sqlDB, slug, firstAdmin)
		cleanupOnboardedTenant(sqlDB, slug, secondAdmin)
	})

	svc := newOnboardTenantService(sqlDB)
	if _, err := svc.OnboardTenant(ctx, iamapp.OnboardTenantInput{
		Name:             "Onboard Dup Tenant",
		Slug:             slug,
		AdminUserID:      firstAdmin,
		AdminDisplayName: "Onboard Dup Admin 1",
		AdminPassword:    "correct-horse-battery-staple",
	}, actor, devTenant); err != nil {
		t.Fatalf("first OnboardTenant() = %v, want nil", err)
	}

	_, err := svc.OnboardTenant(ctx, iamapp.OnboardTenantInput{
		Name:             "Onboard Dup Tenant Again",
		Slug:             slug,
		AdminUserID:      secondAdmin,
		AdminDisplayName: "Onboard Dup Admin 2",
		AdminPassword:    "correct-horse-battery-staple",
	}, actor, devTenant)
	if !isTenantSlugConflict(err) {
		t.Fatalf("second OnboardTenant(same slug) = %v, want iamdomain.ErrTenantSlugConflict", err)
	}

	var secondAdminCount int
	if err := sqlDB.QueryRowContext(ctx,
		`SELECT count(*) FROM metaldocs.iam_users WHERE user_id = $1`, secondAdmin,
	).Scan(&secondAdminCount); err != nil {
		t.Fatalf("count iam_users for rejected onboarding: %v", err)
	}
	if secondAdminCount != 0 {
		t.Fatalf("rejected duplicate-slug onboarding still wrote %d iam_users rows for %s, want 0", secondAdminCount, secondAdmin)
	}
}

// TestOnboardTenant_RequiresCapability proves an actor holding no
// tenant.onboard grant (no role at all) is denied via authz.ErrCapDenied
// before any row is written — the BOLA escalation guard, same shape as
// TestIAMUsersTripwire_Tier2DeniedWithoutGrant.
func TestOnboardTenant_RequiresCapability(t *testing.T) {
	sqlDB := openDB(t)
	ctx := context.Background()

	noCapsActor := testdb.DeterministicID(t, "onboard-nocaps-actor")
	seedIdentity(t, sqlDB, noCapsActor) // identity only — zero roles, zero capabilities

	slug := "onboard-denied-" + testdb.DeterministicID(t, "onboard-denied-slug")
	adminUserID := testdb.DeterministicID(t, "onboard-denied-admin")
	t.Cleanup(func() { cleanupOnboardedTenant(sqlDB, slug, adminUserID) })

	svc := newOnboardTenantService(sqlDB)
	_, err := svc.OnboardTenant(ctx, iamapp.OnboardTenantInput{
		Name:             "Onboard Denied Tenant",
		Slug:             slug,
		AdminUserID:      adminUserID,
		AdminDisplayName: "Onboard Denied Admin",
		AdminPassword:    "correct-horse-battery-staple",
	}, noCapsActor, devTenant)
	if !isCapDenied(err) {
		t.Fatalf("OnboardTenant(no-caps actor) = %v, want authz.ErrCapDenied", err)
	}

	var tenantCount int
	if err := sqlDB.QueryRowContext(ctx,
		`SELECT count(*) FROM metaldocs.tenants WHERE slug = $1`, slug,
	).Scan(&tenantCount); err != nil {
		t.Fatalf("count tenants after denial: %v", err)
	}
	if tenantCount != 0 {
		t.Fatalf("denied OnboardTenant still wrote %d tenants rows, want 0 (BOLA escalation guard)", tenantCount)
	}
}

// cleanupOnboardedTenant removes the probe tenant + admin rows created by a
// (possibly partial, possibly successful) onboarding attempt. Error-discarding
// by design, matching cleanupTenant/cleanupIAMUser's established convention.
func cleanupOnboardedTenant(sqlDB *sql.DB, slug, adminUserID string) {
	_ = withBypassErr(sqlDB, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(),
			`DELETE FROM metaldocs.auth_identities WHERE user_id = $1`, adminUserID)
		return err
	})
	_ = withBypassErr(sqlDB, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(),
			`DELETE FROM metaldocs.iam_user_roles WHERE user_id = $1`, adminUserID)
		return err
	})
	_ = withBypassErr(sqlDB, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(),
			`DELETE FROM metaldocs.iam_users WHERE user_id = $1`, adminUserID)
		return err
	})
	cleanupTenant(sqlDB, slug)
}

func containsSubstring(haystack, needle string) bool {
	return len(needle) > 0 && (len(haystack) >= len(needle)) && (indexOf(haystack, needle) >= 0)
}

func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}
