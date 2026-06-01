package worker

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"metaldocs/internal/platform/messaging"
)

// MaterializeInvoker calls the docx-renderer fanout and returns the result.
// Implemented by *docapp.FreezeService via the MaterializeResult adapter in main.
type MaterializeInvoker interface {
	Materialize(ctx context.Context, tenantID, revisionID string) (MaterializeFanoutResult, error)
}

// MaterializeFanoutResult holds the fanout output.
type MaterializeFanoutResult struct {
	FinalDocxS3Key string
	ContentHash    []byte
}

// MaterializeFinalDocxPersister writes the final docx key + hash transactionally.
type MaterializeFinalDocxPersister interface {
	WriteFinalDocxInTx(ctx context.Context, tx *sql.Tx, tenantID, revisionID, s3Key string, contentHash []byte) error
}

// MaterializePDFEnqueuer enqueues a pdf_dispatch_outbox row inside a transaction.
type MaterializePDFEnqueuer interface {
	Enqueue(ctx context.Context, tx *sql.Tx, tenantID, revisionID string, contentHash []byte) error
}

// MaterializeJobRunner handles EventTypeMaterializeFanout events.
// It calls the docx-renderer fanout and, in a single transaction,
// persists the final docx key and enqueues the PDF dispatch outbox row.
type MaterializeJobRunner struct {
	invoker   MaterializeInvoker
	finalDocx MaterializeFinalDocxPersister
	pdfOutbox MaterializePDFEnqueuer
	db        *sql.DB
}

func NewMaterializeJobRunner(
	invoker MaterializeInvoker,
	finalDocx MaterializeFinalDocxPersister,
	pdfOutbox MaterializePDFEnqueuer,
	db *sql.DB,
) *MaterializeJobRunner {
	return &MaterializeJobRunner{
		invoker:   invoker,
		finalDocx: finalDocx,
		pdfOutbox: pdfOutbox,
		db:        db,
	}
}

func (r *MaterializeJobRunner) Handle(ctx context.Context, event messaging.Event) error {
	payload, err := messaging.MaterializeFanoutPayloadFrom(event)
	if err != nil {
		return fmt.Errorf("materialize job runner: %w", err)
	}
	if payload.TenantID == "" || payload.RevisionID == "" {
		return fmt.Errorf("materialize job runner: missing payload fields")
	}

	// HTTP call to docx-renderer — outside tx.
	result, err := r.invoker.Materialize(ctx, payload.TenantID, payload.RevisionID)
	if err != nil {
		return fmt.Errorf("materialize job runner: %w", err)
	}

	// WriteFinalDocx + PDF enqueue atomically.
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("materialize job runner: begin tx: %w", err)
	}
	if err := r.finalDocx.WriteFinalDocxInTx(ctx, tx, payload.TenantID, payload.RevisionID, result.FinalDocxS3Key, result.ContentHash); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("materialize job runner: write final docx: %w", err)
	}
	if err := r.pdfOutbox.Enqueue(ctx, tx, payload.TenantID, payload.RevisionID, result.ContentHash); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("materialize job runner: enqueue pdf outbox: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("materialize job runner: commit: %w", err)
	}

	slog.InfoContext(ctx, "materialize complete",
		"tenant_id", payload.TenantID,
		"revision_id", payload.RevisionID,
		"final_docx_key", result.FinalDocxS3Key,
	)
	return nil
}
