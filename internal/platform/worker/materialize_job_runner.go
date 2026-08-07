package worker

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/messaging"
)

// TxAuthzSeam is the consumer-owned port onto the iam/authz background-tx
// primitives these runners need (REQ-TOP-2: internal/platform must not
// import internal/modules). Implemented by a thin adapter over
// iam/authz at the production construction site
// (apps/worker/cmd/metaldocs-worker/main.go); see WithBackgroundBypass,
// SeedTxTenant, and BypassSystem in internal/modules/iam/authz for the
// semantics each method bridges.
type TxAuthzSeam interface {
	WithBackgroundBypass(ctx context.Context) context.Context
	SeedTxTenant(ctx context.Context, tx *sql.Tx, tenantID string) error
	BypassSystem(ctx context.Context, tx *sql.Tx) error
}

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
	WriteFinalDocxInTx(ctx context.Context, tx db.Tx, tenantID, revisionID, s3Key string, contentHash []byte) error
}

// MaterializePDFEnqueuer is the minimal published interface for the staging
// pdf dispatch Enqueuer (render/fanout/dispatchjobs), owned here (the
// consumer) and satisfied by *dispatchjobs.Enqueuer. It inserts the paired
// (outbox row, River job) atomically inside tx (M5 F5.3 T3).
type MaterializePDFEnqueuer interface {
	EnqueuePDFTx(ctx context.Context, tx db.Tx, tenantID, revisionID string, frozenDocxHash []byte, finalDocxS3Key, releaseGenerationID string) error
}

// ArtifactFactRecorder is the consumer-owned port onto the approval module's
// release-generation facts (ADR 0085). Implemented by
// approval/application.ArtifactFactWriter.
//
// It runs INSIDE the artifact-persistence transaction: the artifact pointer
// and the fact that records it commit together or not at all, and a write
// belonging to a superseded generation is refused, rolling the whole write
// back. The worker never touches public.release_generations itself.
type ArtifactFactRecorder interface {
	RecordArtifactFactTx(ctx context.Context, tx db.Tx, tenantID, releaseGenerationID, kind, s3Key string) error
}

// ArtifactKindFinalDocx / ArtifactKindFinalPDF are the two halves of the ADR
// 0085 artifact set, as the recorder port names them.
const (
	ArtifactKindFinalDocx = "final_docx"
	ArtifactKindFinalPDF  = "final_pdf"
)

// MaterializeJobRunner handles EventTypeMaterializeFanout events.
// It calls the docx-renderer fanout and, in a single transaction,
// persists the final docx key and enqueues the PDF dispatch outbox row.
type MaterializeJobRunner struct {
	invoker      MaterializeInvoker
	finalDocx    MaterializeFinalDocxPersister
	pdfOutbox    MaterializePDFEnqueuer
	artifactFact ArtifactFactRecorder
	authzSeam    TxAuthzSeam
	db           *sql.DB
}

// WithArtifactFactRecorder wires the ADR 0085 artifact-fact port. Required for
// any render that carries a release generation; renders without one (legacy,
// non-approval) never call it.
func (r *MaterializeJobRunner) WithArtifactFactRecorder(rec ArtifactFactRecorder) *MaterializeJobRunner {
	r.artifactFact = rec
	return r
}

func NewMaterializeJobRunner(
	invoker MaterializeInvoker,
	finalDocx MaterializeFinalDocxPersister,
	pdfOutbox MaterializePDFEnqueuer,
	authzSeam TxAuthzSeam,
	db *sql.DB,
) *MaterializeJobRunner {
	if authzSeam == nil {
		panic("materialize_job_runner: authzSeam is required")
	}
	return &MaterializeJobRunner{
		invoker:   invoker,
		finalDocx: finalDocx,
		pdfOutbox: pdfOutbox,
		authzSeam: authzSeam,
		db:        db,
	}
}

func (r *MaterializeJobRunner) Handle(ctx context.Context, event messaging.Event) error {
	if r.authzSeam == nil {
		return fmt.Errorf("materialize job runner: authz seam not configured")
	}
	ctx = r.authzSeam.WithBackgroundBypass(ctx)
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
	// M3 F3.2 (validation-contract.md §2.2 site 1) — seed the tenant-only RLS
	// backstop GUC before any write/lock in this single-tenant processing tx.
	if err := r.authzSeam.SeedTxTenant(ctx, tx, payload.TenantID); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("materialize job runner: seed tenant: %w", err)
	}
	if err := r.authzSeam.BypassSystem(ctx, tx); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("materialize job runner: bypass system: %w", err)
	}
	// ADR 0085: the artifact fact is part of THIS transaction. It also acts as
	// the generation guard — recording refuses a generation that is no longer
	// the document's freeze head, rolling the artifact write back with it.
	//
	// It runs BEFORE the documents-side write, and that order is load-bearing:
	// the recorder locks release_generations, the write locks documents, and
	// the release coordinator locks generation → documents. Taking documents
	// first here would be the reverse order and deadlock against a concurrent
	// release. The S3 key is already known, so nothing is lost by going first.
	if payload.ReleaseGenerationID != "" {
		if r.artifactFact == nil {
			_ = tx.Rollback()
			return fmt.Errorf("materialize job runner: artifact fact recorder not configured (release generation %s)", payload.ReleaseGenerationID)
		}
		if err := r.artifactFact.RecordArtifactFactTx(ctx, tx, payload.TenantID, payload.ReleaseGenerationID, ArtifactKindFinalDocx, result.FinalDocxS3Key); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("materialize job runner: record final docx fact: %w", err)
		}
	}
	if err := r.finalDocx.WriteFinalDocxInTx(ctx, tx, payload.TenantID, payload.RevisionID, result.FinalDocxS3Key, result.ContentHash); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("materialize job runner: write final docx: %w", err)
	}
	// result.FinalDocxS3Key is the renderer-produced key just written to
	// documents (WriteFinalDocxInTx). Thread it into the pdf dispatch snapshot
	// so the docgen_v2_pdf event carries it (F-QA2-2); the worker must never
	// re-derive this key. result.ContentHash is that same artifact's hash — it
	// lands in pdf_dispatch_outbox.frozen_docx_hash (F-QA4-10, migration 0312).
	if err := r.pdfOutbox.EnqueuePDFTx(ctx, tx, payload.TenantID, payload.RevisionID, result.ContentHash, result.FinalDocxS3Key, payload.ReleaseGenerationID); err != nil {
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
