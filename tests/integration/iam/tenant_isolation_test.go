//go:build integration

package iam_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"
	iampostgres "metaldocs/internal/modules/iam/infrastructure/postgres"
	"metaldocs/tests/integration/testdb"
)


// SEC-05 / migration 0259: iam_users and iam_user_roles carry
// trg_require_cap_asserted (user.manage) — seed via the scheduler bypass GUC
// (withBypass/withBypassErr, shared with membership_area_scope_test.go in this
// package).
func insertUserRoleForTenant(t *testing.T, userID, role, tenantID string) {
	t.Helper()
	db := openDB(t)
	ctx := context.Background()

	withBypass(t, db, func(tx *sql.Tx) {
		// Shared-DB hardening: this suite runs against the shared `metaldocs`
		// database (testdb.DSN returns the raw DATABASE_URL — no per-test clone)
		// with deterministic user IDs and manual t.Cleanup teardown. A prior run
		// killed mid-test leaves an (user_id, role_code) row behind. The upsert
		// below keys on (tenant_id, user_id), which does NOT cover the table PK
		// (user_id, role_code): with a fresh random tenant per run the ON CONFLICT
		// arm misses the leftover and the INSERT collides on iam_user_roles_pkey
		// (23505). Clear any leftover rows for this deterministic user first so
		// the seed is self-healing regardless of prior-run pollution.
		if _, err := tx.ExecContext(ctx,
			`DELETE FROM metaldocs.iam_user_roles WHERE user_id = $1`, userID,
		); err != nil {
			t.Fatalf("pre-clean iam_user_roles: %v", err)
		}

		if _, err := tx.ExecContext(ctx,
			`INSERT INTO metaldocs.iam_users (user_id, display_name, tenant_id)
			 VALUES ($1, $2, $3::uuid)
			 ON CONFLICT (user_id) DO UPDATE SET tenant_id = EXCLUDED.tenant_id`,
			userID, userID, tenantID,
		); err != nil {
			t.Fatalf("insert iam_users: %v", err)
		}

		if _, err := tx.ExecContext(ctx,
			`INSERT INTO metaldocs.iam_user_roles (user_id, role_code, tenant_id, assigned_by)
			 VALUES ($1, $2, $3::uuid, $1)
			 ON CONFLICT (tenant_id, user_id) DO UPDATE SET role_code = EXCLUDED.role_code`,
			userID, role, tenantID,
		); err != nil {
			t.Fatalf("insert iam_user_roles: %v", err)
		}
	})

	t.Cleanup(func() {
		_ = withBypassErr(db, func(tx *sql.Tx) error {
			_, err := tx.ExecContext(context.Background(), `DELETE FROM metaldocs.iam_user_roles WHERE user_id = $1`, userID)
			return err
		})
		_ = withBypassErr(db, func(tx *sql.Tx) error {
			_, err := tx.ExecContext(context.Background(), `DELETE FROM metaldocs.iam_users WHERE user_id = $1`, userID)
			return err
		})
	})
}

// TestRoleProvider_TenantIsolation verifies that RolesByUserID scoped to tenantA
// does not return roles granted under tenantB.
func TestRoleProvider_TenantIsolation(t *testing.T) {
	db := openDB(t)
	ctx := context.Background()
	userID := testdb.DeterministicID(t, "tenant-isolation-user")

	// Real tenants rows — iam_users/iam_user_roles FK to metaldocs.tenants,
	// so hardcoded literal ids violate the tenant_id FKs.
	tenantA := testdb.NewTenant(t, db).ID
	tenantB := testdb.NewTenant(t, db).ID

	insertUserRoleForTenant(t, userID, "viewer", tenantA)

	provider := iampostgres.NewRoleProvider(db)

	rolesA, err := provider.RolesByUserID(ctx, userID, tenantA)
	if err != nil {
		t.Fatalf("RolesByUserID tenantA: %v", err)
	}
	found := false
	for _, r := range rolesA {
		if r == iamdomain.RoleViewer {
			found = true
		}
	}
	if !found {
		t.Errorf("expected viewer role for tenantA; got %v", rolesA)
	}

	_, err = provider.RolesByUserID(ctx, userID, tenantB)
	if err == nil {
		t.Errorf("tenant isolation breach: RolesByUserID returned no error for tenantB but user only granted in tenantA")
	} else if !errors.Is(err, iamdomain.ErrUserNotFound) {
		t.Logf("RolesByUserID tenantB returned expected not-found error: %v", err)
	}
}

// TestHasAnyRole_TenantIsolation verifies that HasAnyRole scoped to tenantA
// does not match a role granted only in tenantB.
func TestHasAnyRole_TenantIsolation(t *testing.T) {
	db := openDB(t)
	ctx := context.Background()

	// Real tenants rows — iam_user_roles carries an FK to metaldocs.tenants,
	// so hardcoded literal ids violate iam_user_roles_tenant_id_fkey.
	tenantC := testdb.NewTenant(t, db).ID
	tenantD := testdb.NewTenant(t, db).ID
	userID := testdb.DeterministicID(t, "alice-admin")

	// SEC-05 / migration 0259: iam_users/iam_user_roles carry
	// trg_require_cap_asserted (user.manage) — seed via the scheduler bypass GUC.
	withBypass(t, db, func(tx *sql.Tx) {
		// pgx's extended protocol rejects multi-command parameterized
		// statements — one Exec per statement.
		if _, err := tx.ExecContext(ctx, `
INSERT INTO metaldocs.iam_users (user_id, display_name, is_active)
VALUES ($1, 'Alice Admin', TRUE)
ON CONFLICT (user_id) DO NOTHING
`, userID); err != nil {
			t.Fatalf("seed iam_users: %v", err)
		}
		// Shared-DB hardening (see insertUserRoleForTenant): clear leftover
		// deterministic-user rows so a fresh random tenant can't collide on the
		// iam_user_roles_pkey (user_id, role_code) uncovered by ON CONFLICT below.
		if _, err := tx.ExecContext(ctx,
			`DELETE FROM metaldocs.iam_user_roles WHERE user_id = $1`, userID,
		); err != nil {
			t.Fatalf("pre-clean iam_user_roles: %v", err)
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO metaldocs.iam_user_roles (user_id, tenant_id, role_code)
VALUES ($1, $2::uuid, 'system_admin')
ON CONFLICT (tenant_id, user_id) DO UPDATE SET role_code = EXCLUDED.role_code
`, userID, tenantC); err != nil {
			t.Fatalf("seed iam_user_roles: %v", err)
		}
	})
	t.Cleanup(func() {
		_ = withBypassErr(db, func(tx *sql.Tx) error {
			_, err := tx.ExecContext(context.Background(), `DELETE FROM metaldocs.iam_user_roles WHERE user_id = $1`, userID)
			return err
		})
		_ = withBypassErr(db, func(tx *sql.Tx) error {
			_, err := tx.ExecContext(context.Background(), `DELETE FROM metaldocs.iam_users WHERE user_id = $1`, userID)
			return err
		})
	})

	repo := iampostgres.NewRoleAdminRepository(db)

	hasC, err := repo.HasAnyRole(ctx, iamdomain.RoleSystemAdmin, tenantC)
	if err != nil {
		t.Fatalf("tenantC: %v", err)
	}
	if !hasC {
		t.Fatalf("tenantC: expected true, got false")
	}

	hasD, err := repo.HasAnyRole(ctx, iamdomain.RoleSystemAdmin, tenantD)
	if err != nil {
		t.Fatalf("tenantD: %v", err)
	}
	if hasD {
		t.Fatalf("tenantD: expected false (cross-tenant bleed), got true")
	}
}
