// Package document_review_surfacer holds the River periodic job that surfaces
// documents due for periodic eQMS review (M6 F6.2 T4). It contains NO raw SQL
// against public.documents — per validation-contract.md §4.1, it calls only
// the documents module's published ports: ReviewDueReader.ListDueForReview
// (read, for count/log) and ReviewSurfaceWriter.MarkSurfaced (the idempotent
// side effect).
package document_review_surfacer

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/riverqueue/river"

	documentsdomain "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/iam/authz"
)

// JobName identifies this job type to River and in logs.
const JobName = "document_review_surfacer"

// BatchSize caps how many due documents the read-port reports per tick (for
// the count/log only — MarkSurfaced's UPDATE has no LIMIT and covers every
// eligible row in one statement).
const BatchSize = 100

// DocumentReviewSurfacerArgs is the (empty) River job payload for the
// review-due surfacer tick. The job carries no per-run parameters — all tick
// behavior is derived from the database at run time.
type DocumentReviewSurfacerArgs struct{}

// Kind implements river.JobArgs, identifying this job type to River.
func (DocumentReviewSurfacerArgs) Kind() string { return JobName }

// DocumentReviewSurfacerWorker is the River worker that runs the review-due
// surfacer tick. Cluster-wide single-runner is provided by River's
// leader-elected periodic insert plus its queue dequeue semantics; no
// advisory lock is taken here (mirrors stuck_instance_watchdog, ADR 0067
// §H-PRE-1).
type DocumentReviewSurfacerWorker struct {
	river.WorkerDefaults[DocumentReviewSurfacerArgs]

	database *sql.DB
	reader   documentsdomain.ReviewDueReader
	writer   documentsdomain.ReviewSurfaceWriter
}

// NewWorker constructs a DocumentReviewSurfacerWorker.
func NewWorker(database *sql.DB, reader documentsdomain.ReviewDueReader, writer documentsdomain.ReviewSurfaceWriter) *DocumentReviewSurfacerWorker {
	return &DocumentReviewSurfacerWorker{
		database: database,
		reader:   reader,
		writer:   writer,
	}
}

// Work runs one surfacer tick under a background-bypass authz context (no
// HTTP-request identity exists here).
func (w *DocumentReviewSurfacerWorker) Work(ctx context.Context, job *river.Job[DocumentReviewSurfacerArgs]) error {
	ctx = authz.WithBackgroundBypass(ctx)
	return run(ctx, w.database, w.reader, w.writer, time.Now().UTC())
}

// run executes one review-due-surfacer tick. It first calls the documents
// read-port (ListDueForReview) for an observability count of what is due,
// then calls the write-port (MarkSurfaced) for the idempotent side effect —
// per validation-contract.md §4.1, both calls go through documents-owned
// ports; this package holds no raw SQL against public.documents.
//
// Cross-tenant scope: mirrors stuck_instance_watchdog.listStuckInstances —
// the tx runs under the scheduler bypass (authz.BypassSystem) with NO tenant
// GUC seeded. public.documents' tenant_isolation RLS policy treats an unset
// metaldocs.tenant_id GUC as "all tenants" (its NULLIF(...) IS NULL branch),
// so this single tx sweeps every tenant's due documents in one MarkSurfaced
// UPDATE — not a per-tenant iterate+seed loop. This is deliberate: the
// contract requires ONE idempotent marker write per due document per cycle,
// and iterating tenants would multiply transactions for no isolation benefit
// (there is no per-tenant side effect to keep separate, unlike
// emitStuckAlert's per-instance governance event).
func run(ctx context.Context, database *sql.DB, reader documentsdomain.ReviewDueReader, writer documentsdomain.ReviewSurfaceWriter, now time.Time) error {
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := authz.BypassSystem(ctx, tx); err != nil {
		return err
	}

	due, err := reader.ListDueForReview(ctx, tx, now, BatchSize)
	if err != nil {
		slog.ErrorContext(ctx, "document_review_surfacer: list due for review failed",
			"job", JobName, "error", err)
		return err
	}

	surfaced, err := writer.MarkSurfaced(ctx, tx, now)
	if err != nil {
		slog.ErrorContext(ctx, "document_review_surfacer: mark surfaced failed",
			"job", JobName, "error", err)
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	slog.InfoContext(ctx, "document_review_surfacer: tick complete",
		"job", JobName,
		"due_count", len(due),
		"surfaced_count", len(surfaced))

	return nil
}
