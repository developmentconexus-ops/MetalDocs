//go:build integration

package migrations_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"metaldocs/tests/integration/testdb"
)

// TestMigration0311_RetiresDocumentPublishCapability proves the ADR 0085
// Stage B capability retirement against a real database: after bootstrap
// (baseline + reference data + the full forward tail, which is exactly what a
// deployed database has), no `document.publish` grant survives anywhere in
// metaldocs.role_capabilities, and the '0311' ledger row is present exactly
// once. `document.supersede` must be untouched — ADR 0085 retains it and
// re-homes it to submit-time cross-document plan authorization, so a migration
// that swept it away would silently disarm that check.
func TestMigration0311_RetiresDocumentPublishCapability(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.OpenFreshDatabase(t)

	if n := capabilityGrantCount(t, ctx, db, "document.publish"); n != 0 {
		t.Fatalf("expected zero role_capabilities grants for 'document.publish' after migration 0311, got %d", n)
	}
	if n := capabilityGrantCount(t, ctx, db, "document.supersede"); n == 0 {
		t.Fatal("migration 0311 removed 'document.supersede' — ADR 0085 RETAINS it (re-homed to submit-time cross-document plan authorization)")
	}
	if n := ledgerCount(t, ctx, db, "0311"); n != 1 {
		t.Fatalf("expected exactly one schema_migrations row for version '0311' after bootstrap, got %d", n)
	}
}

// TestMigration0311_ScheduledPublishCutoverStrandCleanup is the strand-safety
// proof for the River job kind Stage B deleted.
//
// `scheduled_publish_cutover` (was
// internal/modules/approval/jobs/scheduled_publish_args.go) has no worker
// registered anywhere after this deploy. River's producer does not ignore an
// unknown kind — it fetches the row, fails to resolve a worker, and the row
// churns as erroring/retryable noise indefinitely. 0311 therefore removed every
// row still parked in a runnable state.
//
// What is proved here is the post-migration invariant a deployed database must
// satisfy: no non-terminal row of that kind survives.
//
// The former replay half of this test (seed one row per non-terminal state plus
// terminal history, re-execute the committed migration bytes, assert the sweep
// + ledger idempotency) was deleted with the 2026-07-29 fold: 0311 is folded
// into db/baseline/0001_current_schema.sql and archived under
// archive/migrations/post-baseline-2026-07-fold/, so it is never applied to any
// database again and its replay behaviour guards nothing.
func TestMigration0311_ScheduledPublishCutoverStrandCleanup(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.OpenFreshDatabase(t)

	// River provisions its own schema (rivermigrate), so this is a real
	// precondition, not a formality: without it the assertion below would be
	// vacuously green.
	if got := regclass(t, ctx, db, "public.river_job"); !got.Valid {
		t.Fatal("public.river_job does not exist — the strand-cleanup assertion would be vacuous")
	}

	if n := cutoverJobCount(t, ctx, db, nonTerminalStates); n != 0 {
		t.Fatalf("expected zero non-terminal 'scheduled_publish_cutover' river_job rows after migrations, got %d", n)
	}
}

// nonTerminalStates is exactly the set 0311 sweeps: rows River would still hand
// to a producer. `running` is deliberately absent — such a row belongs to a
// live worker's lease and drains when the old binary stops.
var nonTerminalStates = []string{"available", "scheduled", "retryable", "pending"}

func cutoverJobCount(t *testing.T, ctx context.Context, db *sql.DB, states []string) int {
	t.Helper()
	var n int
	if err := db.QueryRowContext(ctx, `
		SELECT count(*) FROM public.river_job
		 WHERE kind = 'scheduled_publish_cutover'
		   AND state::text = ANY($1::text[])`, pgTextArray(states)).Scan(&n); err != nil {
		t.Fatalf("count scheduled_publish_cutover river_job rows: %v", err)
	}
	return n
}

func capabilityGrantCount(t *testing.T, ctx context.Context, db *sql.DB, capability string) int {
	t.Helper()
	var n int
	if err := db.QueryRowContext(ctx,
		`SELECT count(*) FROM metaldocs.role_capabilities WHERE capability = $1`, capability).Scan(&n); err != nil {
		t.Fatalf("count role_capabilities grants for %q: %v", capability, err)
	}
	return n
}

// pgTextArray renders a Go string slice as a Postgres text[] literal. The
// database/sql driver has no native slice encoding, and pulling in a pgx array
// type for a three-element constant would be heavier than the literal.
func pgTextArray(values []string) string {
	return "{" + strings.Join(values, ",") + "}"
}
