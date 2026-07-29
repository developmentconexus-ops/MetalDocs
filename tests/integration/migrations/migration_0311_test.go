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

// migration0311Path is the committed migration file both tests below replay
// verbatim. Asserting against the real file bytes (rather than a hand-copied
// statement) is what makes these proofs about the deployed artifact.
const migration0311Path = "../../../db/migrations/0311_retire_document_publish_capability.sql"

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
// churns as erroring/retryable noise indefinitely. 0311 therefore has to remove
// every row still parked in a runnable state, and it has to do so on a database
// whose River schema it does not own.
//
// Two things are proved here:
//
//  1. Post-migration state: a fully bootstrapped database carries no
//     non-terminal row of that kind. (In THIS harness that state is reached
//     trivially — testdb's curated bootstrap provisions River's own schema
//     AFTER the db/migrations tail, mirroring production where
//     bootstrap.MigrateRiverSchema runs at binary startup, so 0311's guarded
//     block legitimately no-ops during bootstrap. The assertion is still the
//     invariant a deployed database must satisfy; step 2 is what proves the
//     statement can actually do the work.)
//  2. The cleanup itself: seed one row per non-terminal state plus terminal
//     history, replay the committed migration, and require the non-terminal
//     rows to be gone, the terminal history to survive, an unrelated kind to be
//     untouched, and the ledger row not to duplicate. Replaying again must be a
//     clean no-op.
func TestMigration0311_ScheduledPublishCutoverStrandCleanup(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.OpenFreshDatabase(t)

	// River provisions its own schema (rivermigrate), so this is a real
	// precondition, not a formality: without it the assertions below would be
	// vacuously green.
	if got := regclass(t, ctx, db, "public.river_job"); !got.Valid {
		t.Fatal("public.river_job does not exist — the strand-cleanup assertions would be vacuous")
	}

	// (1) Post-bootstrap invariant.
	if n := cutoverJobCount(t, ctx, db, nonTerminalStates); n != 0 {
		t.Fatalf("expected zero non-terminal 'scheduled_publish_cutover' river_job rows after migrations, got %d", n)
	}

	// (2) Seed the strand: one row per state 0311 must clear, plus terminal
	// history it must NOT touch, plus a live row of a DIFFERENT kind that must
	// survive untouched (the DELETE is kind-scoped, and a kind-blind sweep
	// would take the release coordinator's own jobs with it).
	for _, state := range nonTerminalStates {
		insertRiverJob(t, ctx, db, "scheduled_publish_cutover", state)
	}
	for _, state := range terminalStates {
		insertRiverJob(t, ctx, db, "scheduled_publish_cutover", state)
	}
	insertRiverJob(t, ctx, db, "release_evaluate", "available")

	if n := cutoverJobCount(t, ctx, db, nonTerminalStates); n != len(nonTerminalStates) {
		t.Fatalf("seed failed: %d non-terminal cutover rows, want %d", n, len(nonTerminalStates))
	}

	replayMigration0311(t, ctx, db)

	if n := cutoverJobCount(t, ctx, db, nonTerminalStates); n != 0 {
		t.Fatalf("migration 0311 left %d non-terminal 'scheduled_publish_cutover' rows behind — those strand as unknown-kind noise after deploy", n)
	}
	if n := cutoverJobCount(t, ctx, db, terminalStates); n != len(terminalStates) {
		t.Fatalf("terminal 'scheduled_publish_cutover' history = %d rows, want %d — finished jobs are audit trail and must survive", n, len(terminalStates))
	}
	if n := jobCountByKind(t, ctx, db, "release_evaluate"); n != 1 {
		t.Fatalf("migration 0311 deleted rows of an unrelated kind: 'release_evaluate' rows = %d, want 1", n)
	}
	if n := ledgerCount(t, ctx, db, "0311"); n != 1 {
		t.Fatalf("replay must not duplicate the ledger row (ON CONFLICT DO NOTHING): got %d rows for '0311'", n)
	}

	// Idempotency: a second replay against the already-clean database is a
	// no-op, not an error (the guarded DO block and the bare DELETEs are both
	// re-runnable).
	replayMigration0311(t, ctx, db)
	if n := cutoverJobCount(t, ctx, db, nonTerminalStates); n != 0 {
		t.Fatalf("second replay resurrected %d non-terminal rows", n)
	}
	if n := ledgerCount(t, ctx, db, "0311"); n != 1 {
		t.Fatalf("second replay duplicated the ledger row: got %d rows for '0311'", n)
	}
}

// nonTerminalStates is exactly the set 0311 sweeps: rows River would still hand
// to a producer. `running` is deliberately absent — such a row belongs to a
// live worker's lease and drains when the old binary stops.
var nonTerminalStates = []string{"available", "scheduled", "retryable", "pending"}

// terminalStates are finished rows: history 0311 must preserve.
var terminalStates = []string{"completed", "cancelled", "discarded"}

func replayMigration0311(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	migrationSQL, err := os.ReadFile(migration0311Path)
	if err != nil {
		t.Fatalf("read committed 0311 migration file: %v", err)
	}
	if _, err := db.ExecContext(ctx, string(migrationSQL)); err != nil {
		t.Fatalf("0311 must re-apply cleanly against an already-migrated database: %v", err)
	}
}

// insertRiverJob writes one river_job row in the requested state. The
// finalized_or_finalized_at_null CHECK ties finalized_at to terminality, so the
// timestamp is derived from the state rather than passed in.
func insertRiverJob(t *testing.T, ctx context.Context, db *sql.DB, kind, state string) {
	t.Helper()
	// finalized_at is NOT NULL exactly for the terminal states and NULL for
	// every other one; supplying the wrong pairing trips the CHECK rather than
	// producing the row this test needs.
	terminal := false
	for _, s := range terminalStates {
		if s == state {
			terminal = true
		}
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO public.river_job (state, kind, max_attempts, args, metadata, queue, finalized_at)
		VALUES ($1::river_job_state, $2, 5, '{}'::jsonb, '{}'::jsonb, 'default',
		        CASE WHEN $3::bool THEN now() ELSE NULL END)`,
		state, kind, terminal,
	); err != nil {
		t.Fatalf("insert river_job (kind=%s state=%s): %v", kind, state, err)
	}
}

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

func jobCountByKind(t *testing.T, ctx context.Context, db *sql.DB, kind string) int {
	t.Helper()
	var n int
	if err := db.QueryRowContext(ctx,
		`SELECT count(*) FROM public.river_job WHERE kind = $1`, kind).Scan(&n); err != nil {
		t.Fatalf("count river_job rows for kind %q: %v", kind, err)
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
