package dispatchjobs

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"

	"metaldocs/internal/modules/render/fanout"
	"metaldocs/internal/platform/db"
)

// stagingEnqueuer captures the one StagingOutboxRepository method the
// materialize dispatch path needs. Kept unexported and minimal so tests can
// supply a fake without standing up a real *fanout.StagingOutboxRepository,
// mirroring outboxMarker in workers.go.
type stagingEnqueuer interface {
	Enqueue(ctx context.Context, tx db.Tx, tenantID, revisionID string, valuesHash []byte, releaseGenerationID string) (string, error)
}

// pdfStagingEnqueuer captures the pdf-specific enqueue, which additionally
// persists the renderer-produced final_docx_s3_key snapshot (F-QA2-2). Satisfied
// by *fanout.StagingOutboxRepository.EnqueuePDF.
type pdfStagingEnqueuer interface {
	EnqueuePDF(ctx context.Context, tx db.Tx, tenantID, revisionID string, frozenDocxHash []byte, finalDocxS3Key, releaseGenerationID string) (string, error)
}

// riverInserter captures river.Client[*sql.Tx].InsertTx's exact signature.
// *river.Client[*sql.Tx] satisfies this directly (no adapter needed) because
// its InsertTx method already has this shape once TTx is instantiated to
// *sql.Tx. Unexported and minimal so the Enqueuer is unit-testable with a
// recording fake in place of a real River client / database.
type riverInserter interface {
	InsertTx(ctx context.Context, tx *sql.Tx, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error)
}

// Enqueuer enqueues staging pdf/materialize dispatch as a paired
// (outbox row, River job) write inside the caller's business transaction
// (transactional outbox). The outbox row's (tenant_id, revision_id)
// ON CONFLICT DO NOTHING is the single dedup point: StagingOutboxRepository.Enqueue
// returns an empty id when a duplicate is skipped, and Enqueuer treats that as
// "nothing to dispatch" — no River job is inserted for a dedup skip.
type Enqueuer struct {
	client      riverInserter
	pdfRepo     pdfStagingEnqueuer
	matRepo     stagingEnqueuer
	maxAttempts int
}

// NewEnqueuer constructs an Enqueuer bound to a real River client and the two
// staging outbox repos. maxAttempts <= 0 means "let River's own default
// apply" (InsertOpts.MaxAttempts is left unset) rather than forcing a
// single-attempt job.
func NewEnqueuer(client *river.Client[*sql.Tx], pdfRepo, matRepo *fanout.StagingOutboxRepository, maxAttempts int) *Enqueuer {
	return newEnqueuerWithInserter(client, pdfRepo, matRepo, maxAttempts)
}

// newEnqueuerWithInserter is the unexported constructor taking the narrower
// riverInserter/pdfStagingEnqueuer/stagingEnqueuer interfaces, used directly by
// unit tests to inject fakes; NewEnqueuer is the public entry point for
// production wiring.
func newEnqueuerWithInserter(client riverInserter, pdfRepo pdfStagingEnqueuer, matRepo stagingEnqueuer, maxAttempts int) *Enqueuer {
	return &Enqueuer{client: client, pdfRepo: pdfRepo, matRepo: matRepo, maxAttempts: maxAttempts}
}

// EnqueuePDFTx enqueues a pdf_dispatch_outbox row (carrying the renderer-produced
// finalDocxS3Key snapshot, F-QA2-2) and, only when a new row was actually
// inserted (dedup skip returns an empty id), a paired River PDFDispatchArgs job
// threading that key into the args — both inside tx. finalDocxS3Key is
// REQUIRED: EnqueuePDF fails closed on an empty key.
//
// frozenDocxHash is the materialized frozen-docx hash (F-QA4-10): it lands in
// pdf_dispatch_outbox.frozen_docx_hash and in PDFDispatchArgs.FrozenDocxHash.
func (e *Enqueuer) EnqueuePDFTx(ctx context.Context, tx db.Tx, tenantID, revisionID string, frozenDocxHash []byte, finalDocxS3Key, releaseGenerationID string) error {
	id, err := e.pdfRepo.EnqueuePDF(ctx, tx, tenantID, revisionID, frozenDocxHash, finalDocxS3Key, releaseGenerationID)
	if err != nil {
		return err
	}
	if id == "" {
		// ON CONFLICT DO NOTHING skipped the insert (dedup) — no River job to enqueue.
		return nil
	}

	return e.insertRiverJob(ctx, tx, PDFDispatchArgs{
		dispatchFields: dispatchFields{
			TenantID:            tenantID,
			RevisionID:          revisionID,
			OutboxID:            id,
			ReleaseGenerationID: releaseGenerationID,
		},
		FrozenDocxHash: frozenDocxHash,
		FinalDocxS3Key: finalDocxS3Key,
	})
}

// EnqueueMaterializeTx enqueues a materialize_dispatch_outbox row and, only
// when a new row was actually inserted (dedup skip returns an empty id), a
// paired River MaterializeDispatchArgs job — both inside tx.
//
// valuesHash is the resolved-placeholder values hash pinned at freeze
// (F-QA4-10): it lands in materialize_dispatch_outbox.values_hash and in
// MaterializeDispatchArgs.ValuesHash.
func (e *Enqueuer) EnqueueMaterializeTx(ctx context.Context, tx db.Tx, tenantID, revisionID string, valuesHash []byte, releaseGenerationID string) error {
	id, err := e.matRepo.Enqueue(ctx, tx, tenantID, revisionID, valuesHash, releaseGenerationID)
	if err != nil {
		return err
	}
	if id == "" {
		// ON CONFLICT DO NOTHING skipped the insert (dedup) — no River job to enqueue.
		return nil
	}

	return e.insertRiverJob(ctx, tx, MaterializeDispatchArgs{
		dispatchFields: dispatchFields{
			TenantID:            tenantID,
			RevisionID:          revisionID,
			OutboxID:            id,
			ReleaseGenerationID: releaseGenerationID,
		},
		ValuesHash: valuesHash,
	})
}

// insertRiverJob inserts the paired River job sharing the caller's *sql.Tx.
// tx must be a *sql.Tx (River's InsertTx requires it) — a non-*sql.Tx fails
// loud rather than silently skipping the paired insert.
func (e *Enqueuer) insertRiverJob(ctx context.Context, tx db.Tx, args river.JobArgs) error {
	sqlTx, ok := tx.(*sql.Tx)
	if !ok {
		return fmt.Errorf("staging dispatch: river requires *sql.Tx, got %T", tx)
	}

	opts := &river.InsertOpts{Queue: "temporal"}
	if e.maxAttempts > 0 {
		opts.MaxAttempts = e.maxAttempts
	}

	_, err := e.client.InsertTx(ctx, sqlTx, args, opts)
	return err
}
