//go:build integration
// +build integration

package retention

// Real-DB acceptance proofs for the M5 F5.4 staging outbox retention purge
// (spec.md Validation Gate criteria 1 and 2; contract §4.2). Mirrors the
// testdb-factory convention used by
// internal/modules/render/fanout/dispatchjobs/dispatch_integration_test.go:
// a template-cloned database per test, the real StagingOutboxRepository
// (fanout.NewPDFOutboxRepository / fanout.NewMaterializeOutboxRepository), no
// fakes for the system-under-test.
//
// Coverage:
//  1. TestRetentionPurge_Integration_PurgesOnlyOldDispatchedRows — for BOTH
//     pdf and materialize tables: seeds an old-dispatched row (dispatched_at
//     backdated past the 7-day cutoff), a recent-dispatched row, and a
//     dead-lettered row, then runs StagingOutboxRepository.PurgeDispatched
//     directly. Asserts the old row is gone and the recent + dead-lettered
//     rows survive untouched.
//  2. TestRetentionPurge_Integration_PurgeWorkerRunsBothTables — the same
//     shape driven through PurgeWorker.Work (the actual River entrypoint)
//     instead of calling the repo method directly, proving the worker wires
//     both repos correctly end to end.
//  3. TestRetentionPurge_Integration_BatchBounded_CapsAtBatchSizeTimesMaxIterations
//     — seeds 10 eligible old-dispatched pdf rows and calls PurgeDispatched
//     with batchSize=3, maxIterations=2 (i.e. a cap of 6, well under the full
//     eligible set), proving the delete loop is bounded by the batch/iteration
//     product rather than deleting all eligible rows in one unbounded
//     statement. Production always calls this method with the real BatchSize
//     (5000) / MaxIterations (10) constants; this test proves the bounding
//     mechanism generically using smaller numbers, which is equivalent proof
//     of the mechanism per spec.md Validation Gate criterion 2.
//
// Execution is deferred to the M5-close live drive (no DATABASE_URL in this
// authoring session) — see the commit message for the exact -run commands.

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/riverqueue/river"

	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/modules/render/fanout"
	"metaldocs/tests/integration/testdb"
)

// openRetentionDB opens a template-cloned DB pinned to a single connection,
// as dispatch_integration_test.go's openDispatchDB does, so tx-local GUCs
// survive within one connection and the runtime search path resolves the
// metaldocs-schema staging outbox tables.
func openRetentionDB(t *testing.T) *sql.DB {
	t.Helper()
	db, _ := testdb.OpenFreshDatabase(t)
	db.SetMaxOpenConns(1)
	if _, err := db.ExecContext(context.Background(), `SET search_path TO metaldocs, public`); err != nil {
		t.Fatalf("set search_path: %v", err)
	}
	return db
}

// seedTenantTx begins a real business tx and seeds it with tenantID via
// authz.SeedTxTenant, returning the live *sql.Tx for the caller to enqueue
// into and commit. Mirrors dispatch_integration_test.go's helper of the same
// name.
func seedTenantTx(t *testing.T, ctx context.Context, db *sql.DB, tenantID string) *sql.Tx {
	t.Helper()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin business tx: %v", err)
	}
	if err := authz.SeedTxTenant(ctx, tx, tenantID); err != nil {
		_ = tx.Rollback()
		t.Fatalf("seed tx tenant: %v", err)
	}
	return tx
}

// seedDispatchedRow enqueues+commits a pending outbox row for
// (tenantID, revisionID), then marks it dispatched via the real repo method.
// If backdateBy is non-zero, the row's dispatched_at is pushed back by that
// duration with a direct SQL UPDATE (test-only scaffolding — there is no repo
// method to set an arbitrary dispatched_at, since MarkDispatched always uses
// NOW()).
func seedDispatchedRow(t *testing.T, ctx context.Context, db *sql.DB, table string, repo *fanout.StagingOutboxRepository, tenantID, revisionID string, backdateBy time.Duration) {
	t.Helper()

	tx := seedTenantTx(t, ctx, db, tenantID)
	if _, err := repo.Enqueue(ctx, tx, tenantID, revisionID, []byte{0x01}); err != nil {
		_ = tx.Rollback()
		t.Fatalf("enqueue: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit enqueue tx: %v", err)
	}

	var id string
	if err := db.QueryRowContext(ctx,
		`SELECT id::text FROM `+table+` WHERE tenant_id=$1::uuid AND revision_id=$2::uuid`,
		tenantID, revisionID,
	).Scan(&id); err != nil {
		t.Fatalf("load outbox row id: %v", err)
	}

	if err := repo.MarkDispatched(ctx, tenantID, id); err != nil {
		t.Fatalf("mark dispatched: %v", err)
	}

	if backdateBy != 0 {
		if _, err := db.ExecContext(ctx,
			`UPDATE `+table+` SET dispatched_at = now() - $1::interval WHERE tenant_id=$2::uuid AND revision_id=$3::uuid`,
			backdateBy.String(), tenantID, revisionID,
		); err != nil {
			t.Fatalf("backdate dispatched_at: %v", err)
		}
	}
}

// lookupOutboxID resolves the id for a (tenantID, revisionID) row in the
// given fully-qualified table.
func lookupOutboxID(t *testing.T, ctx context.Context, db *sql.DB, table, tenantID, revisionID string) string {
	t.Helper()
	var id string
	if err := db.QueryRowContext(ctx,
		`SELECT id::text FROM `+table+` WHERE tenant_id=$1::uuid AND revision_id=$2::uuid`,
		tenantID, revisionID,
	).Scan(&id); err != nil {
		t.Fatalf("load outbox row id from %s: %v", table, err)
	}
	return id
}

// rowExists reports whether a row for (tenantID, revisionID) still exists in
// the given fully-qualified table.
func rowExists(t *testing.T, ctx context.Context, db *sql.DB, table, tenantID, revisionID string) bool {
	t.Helper()
	var n int
	if err := db.QueryRowContext(ctx,
		`SELECT count(*) FROM `+table+` WHERE tenant_id=$1::uuid AND revision_id=$2::uuid`,
		tenantID, revisionID,
	).Scan(&n); err != nil {
		t.Fatalf("count rows in %s: %v", table, err)
	}
	return n > 0
}

// retentionFixture holds the three seeded rows and their table for one
// pdf-or-materialize equivalence case in
// TestRetentionPurge_Integration_PurgesOnlyOldDispatchedRows.
type retentionFixture struct {
	name  string
	table string
	repo  *fanout.StagingOutboxRepository
}

func pdfAndMaterializeFixtures(db *sql.DB) []retentionFixture {
	return []retentionFixture{
		{name: "pdf", table: "metaldocs.pdf_dispatch_outbox", repo: fanout.NewPDFOutboxRepository(db)},
		{name: "materialize", table: "metaldocs.materialize_dispatch_outbox", repo: fanout.NewMaterializeOutboxRepository(db)},
	}
}

// TestRetentionPurge_Integration_PurgesOnlyOldDispatchedRows is spec.md
// Validation Gate criterion 1: for BOTH pdf and materialize tables, seed an
// old-dispatched row (dispatched_at backdated past the 7-day cutoff), a
// recent-dispatched row, and a dead-lettered row, run PurgeDispatched once,
// and assert only the old-dispatched row is gone.
func TestRetentionPurge_Integration_PurgesOnlyOldDispatchedRows(t *testing.T) {
	ctx := context.Background()
	db := openRetentionDB(t)
	tenant := testdb.NewTenant(t, db)

	for _, fx := range pdfAndMaterializeFixtures(db) {
		t.Run(fx.name, func(t *testing.T) {
			oldRevisionID := testdb.DeterministicID(t, fx.name+"-old-dispatched")
			recentRevisionID := testdb.DeterministicID(t, fx.name+"-recent-dispatched")
			deadRevisionID := testdb.DeterministicID(t, fx.name+"-dead-lettered")

			// (a) old-dispatched: dispatched 8 days ago, past the 7-day cutoff.
			seedDispatchedRow(t, ctx, db, fx.table, fx.repo, tenant.ID, oldRevisionID, 8*24*time.Hour)

			// (b) recent-dispatched: dispatched now, well within the cutoff.
			seedDispatchedRow(t, ctx, db, fx.table, fx.repo, tenant.ID, recentRevisionID, 0)

			// (c) dead-lettered: status='failed', dead_lettered_at set, never
			// eligible regardless of age.
			tx := seedTenantTx(t, ctx, db, tenant.ID)
			if _, err := fx.repo.Enqueue(ctx, tx, tenant.ID, deadRevisionID, []byte{0x03}); err != nil {
				_ = tx.Rollback()
				t.Fatalf("enqueue dead-lettered row: %v", err)
			}
			if err := tx.Commit(); err != nil {
				t.Fatalf("commit enqueue tx: %v", err)
			}
			deadID := lookupOutboxID(t, ctx, db, fx.table, tenant.ID, deadRevisionID)
			if err := fx.repo.MarkFailed(ctx, tenant.ID, deadID, "boom"); err != nil {
				t.Fatalf("mark failed: %v", err)
			}

			cutoff := time.Now().Add(-RetentionPeriod)
			deleted, err := fx.repo.PurgeDispatched(ctx, cutoff, BatchSize, MaxIterations)
			if err != nil {
				t.Fatalf("PurgeDispatched: %v", err)
			}
			if deleted != 1 {
				t.Fatalf("deleted = %d, want 1 (only the old-dispatched row)", deleted)
			}

			if rowExists(t, ctx, db, fx.table, tenant.ID, oldRevisionID) {
				t.Error("old-dispatched row still exists after purge, want gone")
			}

			if !rowExists(t, ctx, db, fx.table, tenant.ID, recentRevisionID) {
				t.Error("recent-dispatched row was purged, want survive")
			}
			recentStatus, err := fx.repo.ReadState(ctx, tenant.ID, recentRevisionID)
			if err != nil {
				t.Fatalf("ReadState recent: %v", err)
			}
			if recentStatus != "dispatched" {
				t.Errorf("recent-dispatched row status = %q, want dispatched (unchanged)", recentStatus)
			}

			if !rowExists(t, ctx, db, fx.table, tenant.ID, deadRevisionID) {
				t.Error("dead-lettered row was purged, want survive (never auto-purged)")
			}
			deadStatus, err := fx.repo.ReadState(ctx, tenant.ID, deadRevisionID)
			if err != nil {
				t.Fatalf("ReadState dead-lettered: %v", err)
			}
			if deadStatus != "failed" {
				t.Errorf("dead-lettered row status = %q, want failed (unchanged)", deadStatus)
			}
		})
	}
}

// TestRetentionPurge_Integration_PurgeWorkerRunsBothTables drives the same
// old/recent/dead-lettered shape through PurgeWorker.Work — the actual River
// entrypoint — instead of calling PurgeDispatched directly, proving the
// worker wires both real repos correctly end to end. PurgeWorker.Work touches
// only the two StagingOutboxRepository instances (no messaging.Publisher), so
// it is cheap to exercise directly against testdb.
func TestRetentionPurge_Integration_PurgeWorkerRunsBothTables(t *testing.T) {
	ctx := context.Background()
	db := openRetentionDB(t)
	tenant := testdb.NewTenant(t, db)

	pdfRepo := fanout.NewPDFOutboxRepository(db)
	matRepo := fanout.NewMaterializeOutboxRepository(db)

	pdfOldRevisionID := testdb.DeterministicID(t, "worker-pdf-old-dispatched")
	pdfRecentRevisionID := testdb.DeterministicID(t, "worker-pdf-recent-dispatched")
	matOldRevisionID := testdb.DeterministicID(t, "worker-mat-old-dispatched")
	matRecentRevisionID := testdb.DeterministicID(t, "worker-mat-recent-dispatched")

	seedDispatchedRow(t, ctx, db, "metaldocs.pdf_dispatch_outbox", pdfRepo, tenant.ID, pdfOldRevisionID, 8*24*time.Hour)
	seedDispatchedRow(t, ctx, db, "metaldocs.pdf_dispatch_outbox", pdfRepo, tenant.ID, pdfRecentRevisionID, 0)
	seedDispatchedRow(t, ctx, db, "metaldocs.materialize_dispatch_outbox", matRepo, tenant.ID, matOldRevisionID, 8*24*time.Hour)
	seedDispatchedRow(t, ctx, db, "metaldocs.materialize_dispatch_outbox", matRepo, tenant.ID, matRecentRevisionID, 0)

	worker := NewPurgeWorker(pdfRepo, matRepo)
	if err := worker.Work(ctx, &river.Job[PurgeArgs]{}); err != nil {
		t.Fatalf("PurgeWorker.Work: %v", err)
	}

	if rowExists(t, ctx, db, "metaldocs.pdf_dispatch_outbox", tenant.ID, pdfOldRevisionID) {
		t.Error("pdf old-dispatched row still exists after PurgeWorker.Work, want gone")
	}
	if !rowExists(t, ctx, db, "metaldocs.pdf_dispatch_outbox", tenant.ID, pdfRecentRevisionID) {
		t.Error("pdf recent-dispatched row was purged by PurgeWorker.Work, want survive")
	}
	if rowExists(t, ctx, db, "metaldocs.materialize_dispatch_outbox", tenant.ID, matOldRevisionID) {
		t.Error("materialize old-dispatched row still exists after PurgeWorker.Work, want gone")
	}
	if !rowExists(t, ctx, db, "metaldocs.materialize_dispatch_outbox", tenant.ID, matRecentRevisionID) {
		t.Error("materialize recent-dispatched row was purged by PurgeWorker.Work, want survive")
	}
}

// TestRetentionPurge_Integration_BatchBounded_CapsAtBatchSizeTimesMaxIterations
// is spec.md Validation Gate criterion 2: seeds 10 eligible old-dispatched pdf
// rows, then calls PurgeDispatched with batchSize=3, maxIterations=2 (a cap
// of 6 rows, less than the full eligible set of 10). Asserts exactly 6 rows
// are deleted and 4 survive, proving the delete loop is bounded by the
// batch/iteration product rather than an unbounded single statement.
//
// Production always calls PurgeDispatched with the real BatchSize (5000) and
// MaxIterations (10) constants; this test proves the bounding mechanism
// generically using smaller numbers against the SAME code path, which is
// equivalent proof of the mechanism without needing a 5000+ row fixture.
func TestRetentionPurge_Integration_BatchBounded_CapsAtBatchSizeTimesMaxIterations(t *testing.T) {
	ctx := context.Background()
	db := openRetentionDB(t)
	tenant := testdb.NewTenant(t, db)

	pdfRepo := fanout.NewPDFOutboxRepository(db)

	const seededEligibleRows = 10
	revisionIDs := make([]string, seededEligibleRows)
	for i := 0; i < seededEligibleRows; i++ {
		revisionIDs[i] = testdb.DeterministicID(t, "batch-bound-pdf-old-dispatched-"+string(rune('a'+i)))
		seedDispatchedRow(t, ctx, db, "metaldocs.pdf_dispatch_outbox", pdfRepo, tenant.ID, revisionIDs[i], 8*24*time.Hour)
	}

	const testBatchSize = 3
	const testMaxIterations = 2
	wantDeleted := testBatchSize * testMaxIterations // 6

	cutoff := time.Now().Add(-RetentionPeriod)
	deleted, err := pdfRepo.PurgeDispatched(ctx, cutoff, testBatchSize, testMaxIterations)
	if err != nil {
		t.Fatalf("PurgeDispatched: %v", err)
	}
	if deleted != wantDeleted {
		t.Fatalf("deleted = %d, want %d (batchSize=%d x maxIterations=%d, capped below the %d eligible rows)",
			deleted, wantDeleted, testBatchSize, testMaxIterations, seededEligibleRows)
	}

	survivors := 0
	for _, revisionID := range revisionIDs {
		if rowExists(t, ctx, db, "metaldocs.pdf_dispatch_outbox", tenant.ID, revisionID) {
			survivors++
		}
	}
	wantSurvivors := seededEligibleRows - wantDeleted // 4
	if survivors != wantSurvivors {
		t.Fatalf("survivors = %d, want %d (proves the loop stopped at the batch/iteration cap, not an unbounded delete)",
			survivors, wantSurvivors)
	}
}
