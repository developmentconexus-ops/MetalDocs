//go:build integration

package authz_test

import (
	"context"
	"testing"

	"metaldocs/internal/modules/iam/authz"
	"metaldocs/tests/integration/testdb"
)

func TestRequire_SystemAdmin_Bypasses(t *testing.T) {
	db, _ := testdb.Open(t)
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	// Set session vars for admin-local user
	_, err = tx.ExecContext(ctx,
		"SELECT set_config('metaldocs.actor_id', 'admin-local', true), set_config('metaldocs.tenant_id', 'ffffffff-ffff-ffff-ffff-ffffffffffff', true)")
	if err != nil {
		t.Fatalf("set session vars: %v", err)
	}

	// Insert admin-local into iam_users so the FK on iam_user_roles is satisfied
	_, err = tx.ExecContext(ctx,
		`INSERT INTO metaldocs.iam_users (user_id, display_name) VALUES ('admin-local', 'Administrator') ON CONFLICT DO NOTHING`)
	if err != nil {
		t.Fatalf("seed iam_users: %v", err)
	}

	// Grant system_admin role to admin-local
	_, err = tx.ExecContext(ctx,
		`INSERT INTO metaldocs.iam_user_roles (user_id, role_code, tenant_id)
		 VALUES ('admin-local', 'system_admin', 'ffffffff-ffff-ffff-ffff-ffffffffffff')
		 ON CONFLICT DO NOTHING`)
	if err != nil {
		t.Fatalf("seed iam_user_roles: %v", err)
	}

	// system_admin should bypass even a non-existent capability
	ctx = authz.WithCapCache(ctx)
	if err := authz.Require(ctx, tx, "does.not.exist", "any-area"); err != nil {
		t.Fatalf("system_admin bypass failed: %v", err)
	}
}
