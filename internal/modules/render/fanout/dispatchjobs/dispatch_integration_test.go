//go:build integration
// +build integration

package dispatchjobs

// Real-DB equivalence proofs for the M5 F5.3 staging pdf/materialize dispatch
// migration off the poll loop onto River (T1-T4, see commits 268f68da,
// f9a713c2, 8242584c, a6a0d868). Mirrors the testdb-factory convention used by
// internal/modules/documents/approval/jobs/scheduled_publish_job_test.go: a
// template-cloned database per test, real repos/publisher/River client, no
// fakes for the system-under-test.
//
// Coverage:
//  1. TestPDFDispatchWorker_Integration_PublishesAndMarksDispatched — enqueue
//     via Enqueuer.EnqueuePDFTx inside a business tx, commit, then run
//     PDFDispatchWorker.Work directly against real repo+publisher+testdb.
//  2. TestMaterializeDispatchWorker_Integration_PublishesAndMarksDispatched —
//     same shape for the materialize kind.
//  3. TestEnqueuer_Integration_DedupSkip_NoSecondRiverInsert — enqueuing the
//     same (tenantID, revisionID) twice via EnqueuePDFTx; the second call must
//     not insert a second river_job row (repo dedup returns empty id).
//  4. TestEnqueuer_Integration_EnqueuePDFTx_InsertsOutboxRowAndRiverJob — proves
//     EnqueuePDFTx/EnqueueMaterializeTx really call InsertTx with a real
//     *sql.Tx against testdb (both the staging outbox row and the paired
//     river_job row exist after commit).
//
// Execution is deferred to the M5-close live drive (no DATABASE_URL in this
// authoring session) — see the commit message for the exact -run commands.

import (
	"context"
	"database/sql"
	"testing"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"

	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/modules/render/fanout"
	riverjobs "metaldocs/internal/platform/jobs/river"
	"metaldocs/internal/platform/messaging"
	"metaldocs/internal/platform/messaging/outbox/postgres"
	"metaldocs/tests/integration/testdb"
)

// openDispatchDB opens a template-cloned DB pinned to a single connection, as
// scheduled_publish_job_test.go does, so tx-local GUCs survive within one
// connection and the runtime search path resolves the metaldocs-schema
// staging outbox tables.
func openDispatchDB(t *testing.T) *sql.DB {
	t.Helper()
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)
	if _, err := db.ExecContext(context.Background(), `SET search_path TO metaldocs, public`); err != nil {
		t.Fatalf("set search_path: %v", err)
	}
	return db
}

// newTestEnqueuer builds a real dispatchjobs.Enqueuer bound to a real
// river.Client[*sql.Tx] (matching production's ClientBundle construction) and
// the two real StagingOutboxRepository instances.
func newTestEnqueuer(t *testing.T, db *sql.DB) *Enqueuer {
	t.Helper()
	bundle, err := riverjobs.NewClientBundle(db, riverjobs.Config{
		Queues:              map[string]river.QueueConfig{"temporal": {MaxWorkers: 1}},
		SkipUnknownJobCheck: true,
	}, river.NewWorkers())
	if err != nil {
		t.Fatalf("new river client bundle: %v", err)
	}
	pdfRepo := fanout.NewPDFOutboxRepository(db)
	matRepo := fanout.NewMaterializeOutboxRepository(db)
	return NewEnqueuer(bundle.Client, pdfRepo, matRepo, 25)
}

// seedTenantTx begins a real business tx and seeds it with tenantID via
// authz.SeedTxTenant (the tenant-seeded write path, matching how production
// callers of EnqueuePDFTx/EnqueueMaterializeTx run under an authz-seeded
// TxRunner.Do transaction), returning the live *sql.Tx for the caller to
// enqueue into and commit.
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

func countOutboxEvents(t *testing.T, ctx context.Context, db *sql.DB, idemKey string) (int, string) {
	t.Helper()
	var n int
	var payload string
	err := db.QueryRowContext(ctx, `
		SELECT count(*), coalesce(max(payload::text), '')
		  FROM metaldocs.outbox_events
		 WHERE idempotency_key = $1`, idemKey,
	).Scan(&n, &payload)
	if err != nil {
		t.Fatalf("count outbox_events: %v", err)
	}
	return n, payload
}

func countRiverJobs(t *testing.T, ctx context.Context, db *sql.DB, kind string) int {
	t.Helper()
	var n int
	if err := db.QueryRowContext(ctx,
		`SELECT count(*) FROM river_job WHERE kind = $1`, kind,
	).Scan(&n); err != nil {
		t.Fatalf("count river_job: %v", err)
	}
	return n
}

// TestPDFDispatchWorker_Integration_PublishesAndMarksDispatched proves the
// enqueue -> publish -> outbox-row-state chain end to end for the pdf kind
// against a real database: EnqueuePDFTx inside a business tx (commit), then
// PDFDispatchWorker.Work run directly (skipping River's own dequeue loop,
// which is River's tested library code) against the real repo + a real
// messaging.Publisher + testdb.
func TestPDFDispatchWorker_Integration_PublishesAndMarksDispatched(t *testing.T) {
	ctx := context.Background()
	db := openDispatchDB(t)

	tenant := testdb.NewTenant(t, db)
	revisionID := testdb.DeterministicID(t, "pdf-revision")
	contentHash := []byte{0xAB, 0xCD, 0xEF, 0x01}

	enqueuer := newTestEnqueuer(t, db)

	tx := seedTenantTx(t, ctx, db, tenant.ID)
	if err := enqueuer.EnqueuePDFTx(ctx, tx, tenant.ID, revisionID, contentHash); err != nil {
		_ = tx.Rollback()
		t.Fatalf("EnqueuePDFTx: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit business tx: %v", err)
	}

	pdfRepo := fanout.NewPDFOutboxRepository(db)
	publisher := postgres.NewPublisher(db)
	worker := NewPDFDispatchWorker(publisher, pdfRepo)

	var outboxID string
	if err := db.QueryRowContext(ctx, `
		SELECT id::text FROM metaldocs.pdf_dispatch_outbox
		 WHERE tenant_id = $1::uuid AND revision_id = $2::uuid`,
		tenant.ID, revisionID,
	).Scan(&outboxID); err != nil {
		t.Fatalf("load pdf_dispatch_outbox row: %v", err)
	}

	fields := dispatchFields{
		TenantID:    tenant.ID,
		RevisionID:  revisionID,
		ContentHash: contentHash,
		OutboxID:    outboxID,
	}

	if err := worker.Work(ctx, &river.Job[PDFDispatchArgs]{
		JobRow: &rivertype.JobRow{Attempt: 1, MaxAttempts: 25},
		Args:   PDFDispatchArgs{dispatchFields: fields},
	}); err != nil {
		t.Fatalf("PDFDispatchWorker.Work: %v", err)
	}

	wantKey := "docgen_v2_pdf:" + tenant.ID + ":" + revisionID
	n, payload := countOutboxEvents(t, ctx, db, wantKey)
	if n != 1 {
		t.Fatalf("outbox_events rows for key %q = %d, want 1", wantKey, n)
	}
	if payload == "" {
		t.Fatal("expected non-empty payload for the pdf-convert outbox event")
	}

	var eventType string
	if err := db.QueryRowContext(ctx, `
		SELECT event_type FROM metaldocs.outbox_events WHERE idempotency_key = $1`, wantKey,
	).Scan(&eventType); err != nil {
		t.Fatalf("load outbox event_type: %v", err)
	}
	if eventType != string(messaging.EventTypePDFConvert) {
		t.Fatalf("event_type = %q, want %q", eventType, messaging.EventTypePDFConvert)
	}

	status, err := pdfRepo.ReadState(ctx, tenant.ID, revisionID)
	if err != nil {
		t.Fatalf("ReadState: %v", err)
	}
	if status != "dispatched" {
		t.Fatalf("pdf_dispatch_outbox status = %q, want dispatched", status)
	}
}

// TestMaterializeDispatchWorker_Integration_PublishesAndMarksDispatched is the
// materialize-kind mirror of the pdf test above.
func TestMaterializeDispatchWorker_Integration_PublishesAndMarksDispatched(t *testing.T) {
	ctx := context.Background()
	db := openDispatchDB(t)

	tenant := testdb.NewTenant(t, db)
	revisionID := testdb.DeterministicID(t, "materialize-revision")
	contentHash := []byte{0x11, 0x22, 0x33}

	enqueuer := newTestEnqueuer(t, db)

	tx := seedTenantTx(t, ctx, db, tenant.ID)
	if err := enqueuer.EnqueueMaterializeTx(ctx, tx, tenant.ID, revisionID, contentHash); err != nil {
		_ = tx.Rollback()
		t.Fatalf("EnqueueMaterializeTx: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit business tx: %v", err)
	}

	matRepo := fanout.NewMaterializeOutboxRepository(db)
	publisher := postgres.NewPublisher(db)
	worker := NewMaterializeDispatchWorker(publisher, matRepo)

	var outboxID string
	if err := db.QueryRowContext(ctx, `
		SELECT id::text FROM metaldocs.materialize_dispatch_outbox
		 WHERE tenant_id = $1::uuid AND revision_id = $2::uuid`,
		tenant.ID, revisionID,
	).Scan(&outboxID); err != nil {
		t.Fatalf("load materialize_dispatch_outbox row: %v", err)
	}

	fields := dispatchFields{
		TenantID:    tenant.ID,
		RevisionID:  revisionID,
		ContentHash: contentHash,
		OutboxID:    outboxID,
	}

	if err := worker.Work(ctx, &river.Job[MaterializeDispatchArgs]{
		JobRow: &rivertype.JobRow{Attempt: 1, MaxAttempts: 25},
		Args:   MaterializeDispatchArgs{dispatchFields: fields},
	}); err != nil {
		t.Fatalf("MaterializeDispatchWorker.Work: %v", err)
	}

	wantKey := "materialize_fanout:" + tenant.ID + ":" + revisionID
	n, payload := countOutboxEvents(t, ctx, db, wantKey)
	if n != 1 {
		t.Fatalf("outbox_events rows for key %q = %d, want 1", wantKey, n)
	}
	if payload == "" {
		t.Fatal("expected non-empty payload for the materialize-fanout outbox event")
	}

	var eventType string
	if err := db.QueryRowContext(ctx, `
		SELECT event_type FROM metaldocs.outbox_events WHERE idempotency_key = $1`, wantKey,
	).Scan(&eventType); err != nil {
		t.Fatalf("load outbox event_type: %v", err)
	}
	if eventType != string(messaging.EventTypeMaterializeFanout) {
		t.Fatalf("event_type = %q, want %q", eventType, messaging.EventTypeMaterializeFanout)
	}

	status, err := matRepo.ReadState(ctx, tenant.ID, revisionID)
	if err != nil {
		t.Fatalf("ReadState: %v", err)
	}
	if status != "dispatched" {
		t.Fatalf("materialize_dispatch_outbox status = %q, want dispatched", status)
	}
}

// TestEnqueuer_Integration_EnqueuePDFTx_InsertsOutboxRowAndRiverJob proves
// EnqueuePDFTx really calls InsertTx with a real *sql.Tx against testdb: after
// commit, both the staging outbox row and the paired river_job row exist.
func TestEnqueuer_Integration_EnqueuePDFTx_InsertsOutboxRowAndRiverJob(t *testing.T) {
	ctx := context.Background()
	db := openDispatchDB(t)

	tenant := testdb.NewTenant(t, db)
	revisionID := testdb.DeterministicID(t, "pdf-insert-proof")
	contentHash := []byte{0x01}

	enqueuer := newTestEnqueuer(t, db)

	tx := seedTenantTx(t, ctx, db, tenant.ID)
	if err := enqueuer.EnqueuePDFTx(ctx, tx, tenant.ID, revisionID, contentHash); err != nil {
		_ = tx.Rollback()
		t.Fatalf("EnqueuePDFTx: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit business tx: %v", err)
	}

	pdfRepo := fanout.NewPDFOutboxRepository(db)
	status, err := pdfRepo.ReadState(ctx, tenant.ID, revisionID)
	if err != nil {
		t.Fatalf("ReadState: %v", err)
	}
	if status != "pending" {
		t.Fatalf("pdf_dispatch_outbox status = %q, want pending", status)
	}

	if n := countRiverJobs(t, ctx, db, "pdf_dispatch"); n != 1 {
		t.Fatalf("river_job rows of kind pdf_dispatch = %d, want 1", n)
	}
}

// TestEnqueuer_Integration_DedupSkip_NoSecondRiverInsert enqueues the SAME
// (tenantID, revisionID) twice via EnqueuePDFTx. The staging outbox
// ON CONFLICT DO NOTHING dedup means the second Enqueue call returns an empty
// id, so enqueueTx must return nil WITHOUT a second River InsertTx — asserted
// against real river_job rows in testdb, not a fake.
func TestEnqueuer_Integration_DedupSkip_NoSecondRiverInsert(t *testing.T) {
	ctx := context.Background()
	db := openDispatchDB(t)

	tenant := testdb.NewTenant(t, db)
	revisionID := testdb.DeterministicID(t, "pdf-dedup-revision")
	contentHash := []byte{0xFE}

	enqueuer := newTestEnqueuer(t, db)

	tx1 := seedTenantTx(t, ctx, db, tenant.ID)
	if err := enqueuer.EnqueuePDFTx(ctx, tx1, tenant.ID, revisionID, contentHash); err != nil {
		_ = tx1.Rollback()
		t.Fatalf("first EnqueuePDFTx: %v", err)
	}
	if err := tx1.Commit(); err != nil {
		t.Fatalf("commit first business tx: %v", err)
	}

	if n := countRiverJobs(t, ctx, db, "pdf_dispatch"); n != 1 {
		t.Fatalf("river_job rows after first enqueue = %d, want 1", n)
	}

	tx2 := seedTenantTx(t, ctx, db, tenant.ID)
	if err := enqueuer.EnqueuePDFTx(ctx, tx2, tenant.ID, revisionID, contentHash); err != nil {
		_ = tx2.Rollback()
		t.Fatalf("second (dedup) EnqueuePDFTx: %v", err)
	}
	if err := tx2.Commit(); err != nil {
		t.Fatalf("commit second business tx: %v", err)
	}

	// Dedup: still exactly one outbox row and one river_job row for this
	// (tenant, revision) pair — the second call must not have inserted again.
	pdfRepo := fanout.NewPDFOutboxRepository(db)
	status, err := pdfRepo.ReadState(ctx, tenant.ID, revisionID)
	if err != nil {
		t.Fatalf("ReadState: %v", err)
	}
	if status != "pending" {
		t.Fatalf("pdf_dispatch_outbox status = %q, want pending", status)
	}

	var outboxRowCount int
	if err := db.QueryRowContext(ctx, `
		SELECT count(*) FROM metaldocs.pdf_dispatch_outbox
		 WHERE tenant_id = $1::uuid AND revision_id = $2::uuid`,
		tenant.ID, revisionID,
	).Scan(&outboxRowCount); err != nil {
		t.Fatalf("count pdf_dispatch_outbox rows: %v", err)
	}
	if outboxRowCount != 1 {
		t.Fatalf("pdf_dispatch_outbox rows for (tenant, revision) = %d, want 1 (unique constraint + dedup)", outboxRowCount)
	}

	if n := countRiverJobs(t, ctx, db, "pdf_dispatch"); n != 1 {
		t.Fatalf("river_job rows after dedup-skip enqueue = %d, want still 1 (no second InsertTx)", n)
	}
}
