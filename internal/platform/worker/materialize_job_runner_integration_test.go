//go:build integration
// +build integration

package worker_test

// Real-DB proof for QR-B (F9): MaterializeJobRunner.Handle must write the
// final docx pointer through the metaldocs authz bypass bridge, not just the
// SeedTxTenant RLS backstop — WriteFinalDocxInTx's UPDATE public.documents
// carries trg_require_cap_asserted (document.edit), which SeedTxTenant alone
// never satisfies (mirrors the document_review_surfacer.job.go precedent:
// WithBackgroundBypass at the composition root + BypassSystem inside the
// tx). See internal/platform/worker/materialize_job_runner.go Handle.
//
// Coverage:
//  1. TestMaterializeJobRunner_Integration_PersistsFinalDocx — the fixed
//     Handle path (WithBackgroundBypass + SeedTxTenant + BypassSystem)
//     against a real testdb + real SnapshotRepository-backed persister;
//     asserts no error and documents.final_docx_s3_key/content_hash persisted.
//  2. TestMaterializeJobRunner_Integration_TripwireStaysArmed_WithoutBypass —
//     negative pin: SeedTxTenant only (no BypassSystem), same UPDATE, must
//     fail closed with the P0001 ErrCapabilityNotAsserted trigger error. This
//     proves the tripwire is not being loosened/removed by the fix.

import (
	"bytes"
	"context"
	"database/sql"
	"strings"
	"testing"

	"metaldocs/internal/modules/documents/infrastructure"
	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/messaging"
	"metaldocs/internal/platform/worker"
	"metaldocs/tests/integration/testdb"
)

// snapshotFinalDocxAdapter mirrors apps/worker/cmd/metaldocs-worker/main.go:45-52
// (snapshotFinalDocxAdapter), bridging the real SnapshotRepository to
// workerapp.MaterializeFinalDocxPersister for this test's system-under-test.
type snapshotFinalDocxAdapter struct {
	repo *infrastructure.SnapshotRepository
}

func (a snapshotFinalDocxAdapter) WriteFinalDocxInTx(ctx context.Context, tx db.Tx, tenantID, revisionID, s3Key string, contentHash []byte) error {
	return a.repo.WriteFinalDocx(ctx, tenantID, revisionID, s3Key, contentHash, tx)
}

// fakeMaterializeInvoker returns a fixed fanout result without calling the
// docx-renderer — Handle's HTTP-fanout step is outside this test's scope.
type fakeMaterializeInvoker struct {
	result worker.MaterializeFanoutResult
}

func (f fakeMaterializeInvoker) Materialize(ctx context.Context, tenantID, revisionID string) (worker.MaterializeFanoutResult, error) {
	return f.result, nil
}

// noopPDFEnqueuer satisfies MaterializePDFEnqueuer without a real River
// client/staging outbox — the pdf dispatch enqueue is outside this test's
// scope (F9 only proves the final-docx write path).
type noopPDFEnqueuer struct{}

func (noopPDFEnqueuer) EnqueuePDFTx(ctx context.Context, tx db.Tx, tenantID, revisionID string, contentHash []byte, finalDocxS3Key, releaseGenerationID string) error {
	return nil
}

// testAuthzSeam mirrors apps/worker/cmd/metaldocs-worker/main.go's
// authzSeamAdapter, bridging the real iam/authz package to
// worker.TxAuthzSeam for this integration test's system-under-test.
type testAuthzSeam struct{}

func (testAuthzSeam) WithBackgroundBypass(ctx context.Context) context.Context {
	return authz.WithBackgroundBypass(ctx)
}

func (testAuthzSeam) SeedTxTenant(ctx context.Context, tx *sql.Tx, tenantID string) error {
	return authz.SeedTxTenant(ctx, tx, tenantID)
}

func (testAuthzSeam) BypassSystem(ctx context.Context, tx *sql.Tx) error {
	return authz.BypassSystem(ctx, tx)
}

func openRunnerDB(t *testing.T) *sql.DB {
	t.Helper()
	database, _ := testdb.Open(t)
	database.SetMaxOpenConns(1)
	return database
}

// TestMaterializeJobRunner_Integration_PersistsFinalDocx proves the fixed
// Handle path (WithBackgroundBypass + SeedTxTenant + BypassSystem) writes the
// final docx pointer through the document.edit tripwire against a real
// database — before the fix this failed with P0001 ErrCapabilityNotAsserted.
func TestMaterializeJobRunner_Integration_PersistsFinalDocx(t *testing.T) {
	ctx := context.Background()
	database := openRunnerDB(t)

	docID, tenantID := testdb.InsertDraftDocument(t, database, "", testdb.NewTenant(t, database).ID)

	wantResult := worker.MaterializeFanoutResult{
		FinalDocxS3Key: "final/materialize-runner-test.docx",
		ContentHash:    bytes.Repeat([]byte{0xAB}, 32), // documents_content_hash_len requires exactly 32 bytes
	}

	repo := infrastructure.NewSnapshotRepository(database)
	runner := worker.NewMaterializeJobRunner(
		fakeMaterializeInvoker{result: wantResult},
		snapshotFinalDocxAdapter{repo: repo},
		noopPDFEnqueuer{},
		testAuthzSeam{},
		database,
	)

	event := messaging.Event{
		EventID:   messaging.EventID("evt-" + docID),
		EventType: messaging.EventTypeMaterializeFanout,
		Payload: messaging.MaterializeFanoutPayload{
			TenantID:   tenantID,
			RevisionID: docID,
		},
	}

	if err := runner.Handle(ctx, event); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	var gotKey string
	var gotHash []byte
	if err := database.QueryRowContext(ctx, `
		SELECT coalesce(final_docx_s3_key, ''), content_hash
		  FROM public.documents
		 WHERE tenant_id=$1::uuid AND id=$2::uuid`,
		tenantID, docID,
	).Scan(&gotKey, &gotHash); err != nil {
		t.Fatalf("read final columns: %v", err)
	}
	if gotKey != wantResult.FinalDocxS3Key {
		t.Fatalf("final_docx_s3_key = %q, want %q", gotKey, wantResult.FinalDocxS3Key)
	}
	if !bytes.Equal(gotHash, wantResult.ContentHash) {
		t.Fatalf("content_hash = %x, want %x", gotHash, wantResult.ContentHash)
	}
}

// TestMaterializeJobRunner_Integration_TripwireStaysArmed_WithoutBypass pins
// the negative: the same WriteFinalDocxInTx UPDATE, in a tx seeded ONLY with
// authz.SeedTxTenant (no authz.BypassSystem), must fail closed with the
// document.edit tripwire's P0001 ErrCapabilityNotAsserted. This is the exact
// pre-fix failure mode and must remain armed — the fix must not have loosened
// or removed the trigger.
func TestMaterializeJobRunner_Integration_TripwireStaysArmed_WithoutBypass(t *testing.T) {
	ctx := context.Background()
	database := openRunnerDB(t)

	docID, tenantID := testdb.InsertDraftDocument(t, database, "", testdb.NewTenant(t, database).ID)
	repo := infrastructure.NewSnapshotRepository(database)

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	if err := authz.SeedTxTenant(ctx, tx, tenantID); err != nil {
		t.Fatalf("SeedTxTenant: %v", err)
	}

	err = repo.WriteFinalDocx(ctx, tenantID, docID, "final/should-not-persist.docx", []byte{0x01}, tx)
	if err == nil {
		t.Fatal("WriteFinalDocx with SeedTxTenant only: got nil error, want tripwire P0001 error")
	}
	if !strings.Contains(err.Error(), "ErrCapabilityNotAsserted") && !isCapabilityNotAssertedPGError(err) {
		t.Fatalf("WriteFinalDocx error = %v, want document.edit tripwire (ErrCapabilityNotAsserted / P0001)", err)
	}
}

// isCapabilityNotAssertedPGError matches the raw P0001 tripwire error text
// (RAISE EXCEPTION message), covering the case where the driver returns the
// raw pq/pgx error rather than a wrapped authz.ErrCapabilityNotAsserted.
func isCapabilityNotAssertedPGError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "capability") && (strings.Contains(msg, "not asserted") || strings.Contains(msg, "p0001"))
}
