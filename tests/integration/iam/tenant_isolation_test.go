//go:build integration

package iam_test

import (
	"context"
	"errors"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"
	iampostgres "metaldocs/internal/modules/iam/infrastructure/postgres"
	"metaldocs/tests/integration/testdb"
)

const tenantA = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
const tenantB = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"

func insertUserRoleForTenant(t *testing.T, userID, role, tenantID string) {
	t.Helper()
	db := openDB(t)
	ctx := context.Background()

	if _, err := db.ExecContext(ctx,
		`INSERT INTO metaldocs.iam_users (user_id, display_name, tenant_id)
		 VALUES ($1, $2, $3::uuid)
		 ON CONFLICT (user_id) DO UPDATE SET tenant_id = EXCLUDED.tenant_id`,
		userID, userID, tenantID,
	); err != nil {
		t.Fatalf("insert iam_users: %v", err)
	}

	if _, err := db.ExecContext(ctx,
		`INSERT INTO metaldocs.iam_user_roles (user_id, role_code, tenant_id, assigned_by)
		 VALUES ($1, $2, $3::uuid, $1)
		 ON CONFLICT (tenant_id, user_id) DO UPDATE SET role_code = EXCLUDED.role_code`,
		userID, role, tenantID,
	); err != nil {
		t.Fatalf("insert iam_user_roles: %v", err)
	}

	t.Cleanup(func() {
		db.ExecContext(context.Background(), `DELETE FROM metaldocs.iam_user_roles WHERE user_id = $1`, userID)   //nolint:errcheck
		db.ExecContext(context.Background(), `DELETE FROM metaldocs.iam_users WHERE user_id = $1`, userID)        //nolint:errcheck
	})
}

// TestRoleProvider_TenantIsolation verifies that RolesByUserID scoped to tenantA
// does not return roles granted under tenantB.
func TestRoleProvider_TenantIsolation(t *testing.T) {
	db := openDB(t)
	ctx := context.Background()
	userID := testdb.DeterministicID(t, "tenant-isolation-user")

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

	tenantC := "33333333-3333-3333-3333-333333333333"
	tenantD := "44444444-4444-4444-4444-444444444444"
	userID := testdb.DeterministicID(t, "alice-admin")

	if _, err := db.ExecContext(ctx, `
INSERT INTO metaldocs.iam_users (user_id, display_name, is_active)
VALUES ($1, 'Alice Admin', TRUE)
ON CONFLICT (user_id) DO NOTHING;
INSERT INTO metaldocs.iam_user_roles (user_id, tenant_id, role_code)
VALUES ($1, $2::uuid, 'system_admin')
ON CONFLICT (tenant_id, user_id) DO UPDATE SET role_code = EXCLUDED.role_code
`, userID, tenantC); err != nil {
		t.Fatalf("seed: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), `DELETE FROM metaldocs.iam_user_roles WHERE user_id = $1`, userID) //nolint:errcheck
		db.ExecContext(context.Background(), `DELETE FROM metaldocs.iam_users WHERE user_id = $1`, userID)     //nolint:errcheck
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
