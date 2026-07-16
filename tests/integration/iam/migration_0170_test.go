//go:build integration

package iam_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"metaldocs/internal/platform/tenant"
	"metaldocs/tests/integration/testdb"
)

// SEC-05 / migration 0259 (and pre-existing since migration 0188): iam_user_roles
// carries trg_require_cap_asserted (user.manage) on INSERT/DELETE/UPDATE — both
// the test seed and the archived 0170 migration's UPDATE against iam_user_roles
// must run under the scheduler bypass GUC (withBypass/withBypassErr, shared with
// membership_area_scope_test.go in this package). This test predates migration
// 0188 and was not updated at the time — drive-by repair, not new breakage.
func TestMigration0170_FlipsApproverFromSystemAdminToApprover(t *testing.T) {
	db, _ := testdb.OpenFreshDatabase(t)
	ctx := context.Background()

	const tenantID = tenant.DevTenantID

	// Seed: approver user with system_admin role (simulates pre-0170 state).
	// pgx's extended protocol rejects multi-command parameterized statements —
	// one Exec per statement.
	withBypass(t, db, func(tx *sql.Tx) {
		if _, err := tx.ExecContext(ctx,
			`DELETE FROM metaldocs.iam_user_roles WHERE user_id = 'approver'`); err != nil {
			t.Fatalf("seed delete: %v", err)
		}
		// testdb.Open clones the curated-baseline template, which carries no
		// dev-seed users — the 'approver' iam_users row must be seeded here
		// (iam_user_roles.user_id FKs to iam_users).
		if _, err := tx.ExecContext(ctx, `
INSERT INTO metaldocs.iam_users (user_id, display_name, tenant_id)
VALUES ('approver', 'Approver', $1::uuid)
ON CONFLICT (user_id) DO NOTHING
`, tenantID); err != nil {
			t.Fatalf("seed iam_users: %v", err)
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO metaldocs.iam_user_roles (user_id, tenant_id, role_code)
VALUES ('approver', $1::uuid, 'system_admin')
`, tenantID); err != nil {
			t.Fatalf("seed: %v", err)
		}
	})
	t.Cleanup(func() {
		_ = withBypassErr(db, func(tx *sql.Tx) error {
			_, err := tx.ExecContext(context.Background(),
				`DELETE FROM metaldocs.iam_user_roles WHERE user_id = 'approver'`)
			return err
		})
	})

	sqlBytes, err := os.ReadFile(filepath.Join("..", "..", "..", "archive", "migrations", "0170_dev_approver_role_correction.sql"))
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	// The migration text carries its own BEGIN/COMMIT, so it cannot run inside a
	// Go-level tx (withBypassErr) — pin one physical connection instead and set
	// the scheduler bypass GUC session-locally (is_local=false) on it before
	// applying, matching Postgres session-GUC semantics for connection.ExecContext.
	applyMigrationWithBypass(ctx, t, db, string(sqlBytes))

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
	db, _ := testdb.OpenFreshDatabase(t)
	ctx := context.Background()
	const tenantID = tenant.DevTenantID

	// pgx's extended protocol rejects multi-command parameterized statements —
	// one Exec per statement.
	withBypass(t, db, func(tx *sql.Tx) {
		if _, err := tx.ExecContext(ctx,
			`DELETE FROM metaldocs.iam_user_roles WHERE user_id = 'approver'`); err != nil {
			t.Fatalf("seed delete: %v", err)
		}
		// testdb.Open clones the curated-baseline template, which carries no
		// dev-seed users — the 'approver' iam_users row must be seeded here
		// (iam_user_roles.user_id FKs to iam_users).
		if _, err := tx.ExecContext(ctx, `
INSERT INTO metaldocs.iam_users (user_id, display_name, tenant_id)
VALUES ('approver', 'Approver', $1::uuid)
ON CONFLICT (user_id) DO NOTHING
`, tenantID); err != nil {
			t.Fatalf("seed iam_users: %v", err)
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO metaldocs.iam_user_roles (user_id, tenant_id, role_code)
VALUES ('approver', $1::uuid, 'system_admin')
`, tenantID); err != nil {
			t.Fatalf("seed: %v", err)
		}
	})
	t.Cleanup(func() {
		_ = withBypassErr(db, func(tx *sql.Tx) error {
			_, err := tx.ExecContext(context.Background(),
				`DELETE FROM metaldocs.iam_user_roles WHERE user_id = 'approver'`)
			return err
		})
	})

	sqlBytes, _ := os.ReadFile(filepath.Join("..", "..", "..", "archive", "migrations", "0170_dev_approver_role_correction.sql"))
	applyMigrationWithBypass(ctx, t, db, string(sqlBytes))
	applyMigrationWithBypass(ctx, t, db, string(sqlBytes))

	var role string
	db.QueryRowContext(ctx,
		`SELECT role_code FROM metaldocs.iam_user_roles WHERE user_id = 'approver'`).Scan(&role)
	if role != "approver" {
		t.Fatalf("got role %q, want approver", role)
	}
}

// applyMigrationWithBypass runs a raw multi-statement migration script (carrying
// its own BEGIN/COMMIT) on one pinned physical connection with the scheduler
// authz bypass GUC set session-locally first, so the migration's own UPDATE
// against metaldocs.iam_user_roles (tripwire-guarded since migration 0188) is
// not rejected. Session-local (is_local=false) is required here — the migration
// text opens/closes its own transaction, so a tx-local (is_local=true) set_config
// from a wrapping Go tx would not survive into it.
func applyMigrationWithBypass(ctx context.Context, t *testing.T, db *sql.DB, script string) {
	t.Helper()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("applyMigrationWithBypass: acquire conn: %v", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, `SELECT set_config('metaldocs.bypass_authz', 'scheduler', false)`); err != nil {
		t.Fatalf("applyMigrationWithBypass: set bypass: %v", err)
	}
	if _, err := conn.ExecContext(ctx, script); err != nil {
		t.Fatalf("applyMigrationWithBypass: exec: %v", err)
	}
	if _, err := conn.ExecContext(ctx, `SELECT set_config('metaldocs.bypass_authz', '', false)`); err != nil {
		t.Fatalf("applyMigrationWithBypass: clear bypass: %v", err)
	}
}
