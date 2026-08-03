//go:build integration

package migrations_test

import (
	"context"
	"database/sql"
	"testing"

	"metaldocs/tests/integration/testdb"
)

// TestMigration0302_DropTemplatesApprovalConfig proves the bootstrap state
// 0302 (ADR 0082 phase c, unit 3.1a S5) is responsible for: the curated
// bootstrap this test database was cloned from must leave
// public.templates_approval_config gone and the '0302' ledger row present
// exactly once.
//
// The former replay halves of this test (recreate the table, plant a row,
// re-execute the committed migration bytes to prove the pre-drop emptiness
// assert fires and rolls back; then empty and re-run green) were deleted with
// the 2026-07-29 fold: 0302 is folded into db/baseline/0001_current_schema.sql
// and archived under archive/migrations/post-baseline-2026-07-fold/, so it is
// never applied to any database again and its replay behaviour guards nothing.
// The bootstrap-state assertions below remain the live invariant.
func TestMigration0302_DropTemplatesApprovalConfig(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.OpenFreshDatabase(t)

	if got := regclass0302(t, ctx, db); got.Valid {
		t.Fatalf("bootstrap should have dropped public.templates_approval_config (migration 0302), but it exists as %q", got.String)
	}
	if n := ledgerCount0302(t, ctx, db); n != 1 {
		t.Fatalf("expected exactly one schema_migrations row for version '0302' after bootstrap, got %d", n)
	}
}

func regclass0302(t *testing.T, ctx context.Context, db *sql.DB) sql.NullString {
	t.Helper()
	var got sql.NullString
	if err := db.QueryRowContext(ctx, `SELECT to_regclass('public.templates_approval_config')::text`).Scan(&got); err != nil {
		t.Fatalf("to_regclass(public.templates_approval_config): %v", err)
	}
	return got
}

func ledgerCount0302(t *testing.T, ctx context.Context, db *sql.DB) int {
	t.Helper()
	var n int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM public.schema_migrations WHERE version = '0302'`).Scan(&n); err != nil {
		t.Fatalf("count '0302' schema_migrations rows: %v", err)
	}
	return n
}
