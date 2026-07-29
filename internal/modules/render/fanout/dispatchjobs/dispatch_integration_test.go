//go:build integration
// +build integration

package dispatchjobs

// Real-DB equivalence proofs for the M5 F5.3 staging pdf/materialize dispatch
// migration off the poll loop onto River (T1-T4, see commits 268f68da,
// f9a713c2, 8242584c, a6a0d868). Mirrors the testdb-factory convention used by
// internal/modules/approval/jobs/scheduled_publish_job_test.go: a
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
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"strings"
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
	db, _ := testdb.OpenFreshDatabase(t)
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
	// Migration 0312: this outbox carries the MATERIALIZED frozen-docx hash.
	frozenDocxHash := []byte{0xAB, 0xCD, 0xEF, 0x01}
	finalDocxKey := "tenants/" + tenant.ID + "/" + revisionID + "/frozen.docx"

	enqueuer := newTestEnqueuer(t, db)

	tx := seedTenantTx(t, ctx, db, tenant.ID)
	if err := enqueuer.EnqueuePDFTx(ctx, tx, tenant.ID, revisionID, frozenDocxHash, finalDocxKey, ""); err != nil {
		_ = tx.Rollback()
		t.Fatalf("EnqueuePDFTx: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit business tx: %v", err)
	}

	// Renamed-column round-trip: the enqueue wrote the hash under
	// frozen_docx_hash, the name that matches what it actually carries.
	var gotFrozenDocxHash []byte
	if err := db.QueryRowContext(ctx, `
		SELECT frozen_docx_hash FROM metaldocs.pdf_dispatch_outbox
		 WHERE tenant_id = $1::uuid AND revision_id = $2::uuid`,
		tenant.ID, revisionID,
	).Scan(&gotFrozenDocxHash); err != nil {
		t.Fatalf("read pdf_dispatch_outbox.frozen_docx_hash: %v", err)
	}
	if !bytes.Equal(gotFrozenDocxHash, frozenDocxHash) {
		t.Fatalf("frozen_docx_hash = %x, want %x", gotFrozenDocxHash, frozenDocxHash)
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

	// The worker reads final_docx_s3_key off its OWN job args (EnqueuePDFTx
	// threads it there); it never re-reads the outbox row. Hand-building the
	// args here therefore has to mirror what EnqueuePDFTx wrote, or the
	// final_docx_s3_key assertion below is asserting against a fixture the
	// production path would never produce.
	fields := dispatchFields{
		TenantID:   tenant.ID,
		RevisionID: revisionID,
		OutboxID:   outboxID,
	}

	if err := worker.Work(ctx, &river.Job[PDFDispatchArgs]{
		JobRow: &rivertype.JobRow{Attempt: 1, MaxAttempts: 25},
		Args: PDFDispatchArgs{
			dispatchFields: fields,
			FrozenDocxHash: frozenDocxHash,
			FinalDocxS3Key: finalDocxKey,
		},
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
	if !strings.Contains(payload, finalDocxKey) {
		t.Fatalf("published pdf-convert payload %q does not carry final_docx_s3_key %q", payload, finalDocxKey)
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
	// Migration 0312: this outbox carries the resolved-placeholder VALUES hash.
	valuesHash := []byte{0x11, 0x22, 0x33}

	enqueuer := newTestEnqueuer(t, db)

	tx := seedTenantTx(t, ctx, db, tenant.ID)
	if err := enqueuer.EnqueueMaterializeTx(ctx, tx, tenant.ID, revisionID, valuesHash, ""); err != nil {
		_ = tx.Rollback()
		t.Fatalf("EnqueueMaterializeTx: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit business tx: %v", err)
	}

	// Renamed-column round-trip for the other table.
	var gotValuesHash []byte
	if err := db.QueryRowContext(ctx, `
		SELECT values_hash FROM metaldocs.materialize_dispatch_outbox
		 WHERE tenant_id = $1::uuid AND revision_id = $2::uuid`,
		tenant.ID, revisionID,
	).Scan(&gotValuesHash); err != nil {
		t.Fatalf("read materialize_dispatch_outbox.values_hash: %v", err)
	}
	if !bytes.Equal(gotValuesHash, valuesHash) {
		t.Fatalf("values_hash = %x, want %x", gotValuesHash, valuesHash)
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
		TenantID:   tenant.ID,
		RevisionID: revisionID,
		OutboxID:   outboxID,
	}

	if err := worker.Work(ctx, &river.Job[MaterializeDispatchArgs]{
		JobRow: &rivertype.JobRow{Attempt: 1, MaxAttempts: 25},
		Args: MaterializeDispatchArgs{
			dispatchFields: fields,
			ValuesHash:     valuesHash,
		},
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
	frozenDocxHash := []byte{0x01}
	finalDocxKey := "tenants/" + tenant.ID + "/" + revisionID + "/frozen.docx"

	enqueuer := newTestEnqueuer(t, db)

	tx := seedTenantTx(t, ctx, db, tenant.ID)
	if err := enqueuer.EnqueuePDFTx(ctx, tx, tenant.ID, revisionID, frozenDocxHash, finalDocxKey, ""); err != nil {
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

	// Round-trip: the renderer-produced key is persisted on the staging row
	// (migration 0309 column), the event-contract snapshot the dispatch reads.
	var gotKey string
	if err := db.QueryRowContext(ctx, `
		SELECT final_docx_s3_key FROM metaldocs.pdf_dispatch_outbox
		 WHERE tenant_id = $1::uuid AND revision_id = $2::uuid`,
		tenant.ID, revisionID,
	).Scan(&gotKey); err != nil {
		t.Fatalf("load final_docx_s3_key: %v", err)
	}
	if gotKey != finalDocxKey {
		t.Fatalf("pdf_dispatch_outbox.final_docx_s3_key = %q, want %q", gotKey, finalDocxKey)
	}

	if n := countRiverJobs(t, ctx, db, "pdf_dispatch"); n != 1 {
		t.Fatalf("river_job rows of kind pdf_dispatch = %d, want 1", n)
	}

	// Worker-side arg decode (migration 0312 hard cutover): the args River
	// persisted must carry the renamed JSON field and decode back into the
	// typed field the worker reads. Asserting the raw JSON too is deliberate —
	// a struct-only assertion would still pass if the tag silently reverted.
	var rawArgs []byte
	if err := db.QueryRowContext(ctx,
		`SELECT args FROM river_job WHERE kind = 'pdf_dispatch'`,
	).Scan(&rawArgs); err != nil {
		t.Fatalf("load river_job args: %v", err)
	}
	if !bytes.Contains(rawArgs, []byte(`"frozen_docx_hash"`)) {
		t.Fatalf("river_job args %s missing frozen_docx_hash field", rawArgs)
	}
	if bytes.Contains(rawArgs, []byte(`"content_hash"`)) {
		t.Fatalf("river_job args %s still carry the old content_hash field", rawArgs)
	}
	var decoded PDFDispatchArgs
	if err := json.Unmarshal(rawArgs, &decoded); err != nil {
		t.Fatalf("decode PDFDispatchArgs: %v", err)
	}
	if !bytes.Equal(decoded.FrozenDocxHash, frozenDocxHash) {
		t.Fatalf("decoded FrozenDocxHash = %x, want %x", decoded.FrozenDocxHash, frozenDocxHash)
	}
	if decoded.RevisionID != revisionID || decoded.TenantID != tenant.ID {
		t.Fatalf("decoded identity = (%s, %s), want (%s, %s)",
			decoded.TenantID, decoded.RevisionID, tenant.ID, revisionID)
	}
}

// Materialize-side mirror of the arg-decode proof above: the values hash must
// travel under values_hash, and the old shared name must be gone.
func TestEnqueuer_Integration_EnqueueMaterializeTx_ArgsCarryValuesHash(t *testing.T) {
	ctx := context.Background()
	db := openDispatchDB(t)

	tenant := testdb.NewTenant(t, db)
	revisionID := testdb.DeterministicID(t, "materialize-args-proof")
	valuesHash := []byte{0x0A, 0x0B, 0x0C}

	enqueuer := newTestEnqueuer(t, db)

	tx := seedTenantTx(t, ctx, db, tenant.ID)
	if err := enqueuer.EnqueueMaterializeTx(ctx, tx, tenant.ID, revisionID, valuesHash, ""); err != nil {
		_ = tx.Rollback()
		t.Fatalf("EnqueueMaterializeTx: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit business tx: %v", err)
	}

	var rawArgs []byte
	if err := db.QueryRowContext(ctx,
		`SELECT args FROM river_job WHERE kind = 'materialize_dispatch'`,
	).Scan(&rawArgs); err != nil {
		t.Fatalf("load river_job args: %v", err)
	}
	if !bytes.Contains(rawArgs, []byte(`"values_hash"`)) {
		t.Fatalf("river_job args %s missing values_hash field", rawArgs)
	}
	if bytes.Contains(rawArgs, []byte(`"content_hash"`)) {
		t.Fatalf("river_job args %s still carry the old content_hash field", rawArgs)
	}
	var decoded MaterializeDispatchArgs
	if err := json.Unmarshal(rawArgs, &decoded); err != nil {
		t.Fatalf("decode MaterializeDispatchArgs: %v", err)
	}
	if !bytes.Equal(decoded.ValuesHash, valuesHash) {
		t.Fatalf("decoded ValuesHash = %x, want %x", decoded.ValuesHash, valuesHash)
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
	frozenDocxHash := []byte{0xFE}
	finalDocxKey := "tenants/" + tenant.ID + "/" + revisionID + "/frozen.docx"

	enqueuer := newTestEnqueuer(t, db)

	tx1 := seedTenantTx(t, ctx, db, tenant.ID)
	if err := enqueuer.EnqueuePDFTx(ctx, tx1, tenant.ID, revisionID, frozenDocxHash, finalDocxKey, ""); err != nil {
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
	if err := enqueuer.EnqueuePDFTx(ctx, tx2, tenant.ID, revisionID, frozenDocxHash, finalDocxKey, ""); err != nil {
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
