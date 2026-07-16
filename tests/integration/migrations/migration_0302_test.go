//go:build integration

package migrations_test

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"

	"metaldocs/tests/integration/testdb"
)

// TestMigration0302_DropTemplatesApprovalConfig proves both directions of the
// 0302 guarded DROP (ADR 0082 phase c, unit 3.1a S5) by executing the EXACT
// committed migration file bytes against a real, per-test-isolated database:
//
//  1. Positive path (bootstrap proof): the curated bootstrap this test
//     database was cloned from applies the baseline (which CREATEs
//     public.templates_approval_config) and then the forward migration tail
//     through 0302 — so before this test touches anything, the table must
//     already be gone and the '0302' ledger row present.
//  2. Negative path (pre-drop emptiness assert): recreate the table with the
//     baseline's columns/PK, plant one row, and re-execute the committed 0302
//     file — the DO-block assert must RAISE (P0001, "pre-drop assert") and
//     the file's own BEGIN/COMMIT must roll back, leaving both the table and
//     the planted row intact.
//  3. Green re-run: empty the table and re-execute the same bytes — the
//     migration must apply cleanly, drop the table, and leave exactly one
//     '0302' ledger row (ON CONFLICT DO NOTHING re-run idempotency).
//
// The recreate omits the baseline's FK to templates_template: the guard under
// test counts rows via information_schema + count(*) and never consults the
// FK, and omitting it lets the planted row use a literal UUID with no parent
// fixture.
func TestMigration0302_DropTemplatesApprovalConfig(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.OpenFreshDatabase(t)

	// (1) Bootstrap already applied 0302: table gone, ledger row present.
	if got := regclass0302(t, ctx, db); got.Valid {
		t.Fatalf("bootstrap should have dropped public.templates_approval_config (migration 0302), but it exists as %q", got.String)
	}
	if n := ledgerCount0302(t, ctx, db); n != 1 {
		t.Fatalf("expected exactly one schema_migrations row for version '0302' after bootstrap, got %d", n)
	}

	migrationSQL, err := os.ReadFile("../../../db/migrations/0302_drop_templates_approval_config.sql")
	if err != nil {
		t.Fatalf("read committed 0302 migration file: %v", err)
	}

	// (2) Plant-before-apply: recreate the table (baseline columns/PK, FK
	// deliberately omitted — see doc comment) and plant a single config row.
	if _, err := db.ExecContext(ctx, `
CREATE TABLE public.templates_approval_config (
    template_id uuid NOT NULL,
    reviewer_role text,
    approver_role text NOT NULL,
    CONSTRAINT templates_approval_config_pkey PRIMARY KEY (template_id)
)`); err != nil {
		t.Fatalf("recreate templates_approval_config: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO public.templates_approval_config (template_id, reviewer_role, approver_role)
VALUES ('00000000-0000-0000-0000-00000000dead', 'reviewer', 'approver')`); err != nil {
		t.Fatalf("plant legacy config row: %v", err)
	}

	_, err = db.ExecContext(ctx, string(migrationSQL))
	if err == nil {
		t.Fatal("0302 must fail its pre-drop emptiness assert while templates_approval_config holds a row; it succeeded")
	}
	if !strings.Contains(err.Error(), "pre-drop assert") {
		t.Fatalf("expected the 0302 pre-drop assert message in the error, got: %v", err)
	}

	// The file's own BEGIN/COMMIT must have rolled back: table + row intact.
	if got := regclass0302(t, ctx, db); !got.Valid {
		t.Fatal("failed 0302 run must roll back: templates_approval_config should still exist after the assert fired")
	}
	var rows int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM public.templates_approval_config`).Scan(&rows); err != nil {
		t.Fatalf("count planted rows after failed run: %v", err)
	}
	if rows != 1 {
		t.Fatalf("failed 0302 run must roll back: planted row count = %d, want 1", rows)
	}

	// (3) Empty the table; the exact same bytes must now apply green.
	if _, err := db.ExecContext(ctx, `DELETE FROM public.templates_approval_config`); err != nil {
		t.Fatalf("empty the recreated table: %v", err)
	}
	if _, err := db.ExecContext(ctx, string(migrationSQL)); err != nil {
		t.Fatalf("0302 must apply cleanly against an empty table: %v", err)
	}
	if got := regclass0302(t, ctx, db); got.Valid {
		t.Fatal("green 0302 run must drop templates_approval_config")
	}
	if n := ledgerCount0302(t, ctx, db); n != 1 {
		t.Fatalf("re-run must not duplicate the ledger row (ON CONFLICT DO NOTHING): got %d rows for '0302'", n)
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
