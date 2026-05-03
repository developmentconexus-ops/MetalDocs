//go:build integration

package iam_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"metaldocs/internal/platform/tenant"
	"metaldocs/tests/integration/testdb"
)

func TestMigration0170_FlipsApproverFromSystemAdminToApprover(t *testing.T) {
	db, _ := testdb.Open(t)
	ctx := context.Background()

	const tenantID = tenant.DevTenantID

	// Seed: approver user with system_admin role (simulates pre-0170 state)
	if _, err := db.ExecContext(ctx, `
DELETE FROM metaldocs.iam_user_roles WHERE user_id = 'approver';
INSERT INTO metaldocs.iam_user_roles (user_id, tenant_id, role_code)
VALUES ('approver', $1::uuid, 'system_admin')
`, tenantID); err != nil {
		t.Fatalf("seed: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(),
			`DELETE FROM metaldocs.iam_user_roles WHERE user_id = 'approver'`) //nolint:errcheck
	})

	sqlBytes, err := os.ReadFile(filepath.Join("..", "..", "..", "migrations", "0170_dev_approver_role_correction.sql"))
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	if _, err := db.ExecContext(ctx, string(sqlBytes)); err != nil {
		t.Fatalf("apply 0170: %v", err)
	}

	var role string
	if err := db.QueryRowContext(ctx,
		`SELECT role_code FROM metaldocs.iam_user_roles WHERE user_id = 'approver'`).
		Scan(&role); err != nil {
		t.Fatalf("query role: %v", err)
	}
	if role != "approver" {
		t.Fatalf("got role %q, want approver", role)
	}
}

func TestMigration0170_Idempotent(t *testing.T) {
	db, _ := testdb.Open(t)
	ctx := context.Background()
	const tenantID = tenant.DevTenantID

	if _, err := db.ExecContext(ctx, `
DELETE FROM metaldocs.iam_user_roles WHERE user_id = 'approver';
INSERT INTO metaldocs.iam_user_roles (user_id, tenant_id, role_code)
VALUES ('approver', $1::uuid, 'system_admin')
`, tenantID); err != nil {
		t.Fatalf("seed: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(),
			`DELETE FROM metaldocs.iam_user_roles WHERE user_id = 'approver'`) //nolint:errcheck
	})

	sqlBytes, _ := os.ReadFile(filepath.Join("..", "..", "..", "migrations", "0170_dev_approver_role_correction.sql"))
	if _, err := db.ExecContext(ctx, string(sqlBytes)); err != nil {
		t.Fatalf("first apply: %v", err)
	}
	if _, err := db.ExecContext(ctx, string(sqlBytes)); err != nil {
		t.Fatalf("second apply (idempotency): %v", err)
	}

	var role string
	db.QueryRowContext(ctx,
		`SELECT role_code FROM metaldocs.iam_user_roles WHERE user_id = 'approver'`).Scan(&role)
	if role != "approver" {
		t.Fatalf("got role %q, want approver", role)
	}
}
