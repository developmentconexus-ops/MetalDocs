// Package document_review_surfacer holds the River periodic job that surfaces
// documents due for periodic eQMS review (M6 F6.2 T4). It contains NO raw SQL
// against public.documents — per validation-contract.md §4.1, it calls only
// the documents module's published ports: ReviewDueReader.ListDueForReview
// (read, for count/log), ReviewDueReader.ListTenantsWithDueReviews (the
// cross-tenant enumeration read), and ReviewSurfaceWriter.MarkSurfaced (the
// idempotent, per-tenant-seeded side effect — validation-contract.md
// §4.2/§4.3).
package document_review_surfacer

import (
	"context"
	"database/sql"
	"errors"
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

// run executes one review-due-surfacer tick in two phases, per
// validation-contract.md §4.2/§4.3:
//
//  1. a single cross-tenant SYSTEM READ (ListTenantsWithDueReviews) under the
//     scheduler bypass with no tenant GUC seeded — enumerating tenants is
//     safe unseeded, exactly like stuck_instance_watchdog.listStuckInstances;
//  2. for EACH tenant returned, a NEW tx that seeds authz.BypassSystem THEN
//     authz.SeedTxTenant(tenantID) before calling ListDueForReview (for the
//     count/log) and MarkSurfaced (the idempotent side effect), passing that
//     same tenantID as an explicit parameter to both — mirroring
//     stuck_instance_watchdog.emitStuckAlert's per-instance SeedTxTenant
//     pattern. Isolation is now primarily by the explicit tenant_id predicate
//     each port carries; FORCE RLS on the seeded GUC is the backstop.
//
// Per validation-contract.md §4.1, every call goes through documents-owned
// ports (ListTenantsWithDueReviews, ListDueForReview, MarkSurfaced); this
// package holds no raw SQL against public.documents.
func run(ctx context.Context, database *sql.DB, reader documentsdomain.ReviewDueReader, writer documentsdomain.ReviewSurfaceWriter, now time.Time) error {
	tenantIDs, err := listTenantsWithDueReviews(ctx, database, reader, now)
	if err != nil {
		slog.ErrorContext(ctx, "document_review_surfacer: list tenants with due reviews failed",
			"job", JobName, "error", err)
		return err
	}

	var totalDue, totalSurfaced int
	var runErr error

	for _, tenantID := range tenantIDs {
		due, surfaced, err := surfaceTenant(ctx, database, reader, writer, tenantID, now)
		if err != nil {
			slog.ErrorContext(ctx, "document_review_surfacer: surface tenant failed",
				"job", JobName, "tenant_id", tenantID, "error", err)
			runErr = errors.Join(runErr, err)
			continue
		}
		totalDue += due
		totalSurfaced += surfaced
	}

	slog.InfoContext(ctx, "document_review_surfacer: tick complete",
		"job", JobName,
		"tenants_with_due", len(tenantIDs),
		"due_count", totalDue,
		"surfaced_count", totalSurfaced)

	return runErr
}

// listTenantsWithDueReviews opens a single bypass tx (no tenant GUC seeded —
// this is a system-level cross-tenant read, safe unseeded) and returns the
// distinct tenant_ids with at least one review-due document.
func listTenantsWithDueReviews(ctx context.Context, database *sql.DB, reader documentsdomain.ReviewDueReader, now time.Time) ([]string, error) {
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err := authz.BypassSystem(ctx, tx); err != nil {
		return nil, err
	}

	tenantIDs, err := reader.ListTenantsWithDueReviews(ctx, tx, now)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return tenantIDs, nil
}

// surfaceTenant opens a NEW tx seeded to exactly one tenant
// (authz.BypassSystem then authz.SeedTxTenant(tenantID)) and, within it, reads
// the tenant's due documents (for the count/log) and marks them surfaced,
// passing tenantID explicitly to both ports so scoping is correct-by-
// construction; RLS on the seeded GUC backstops the write.
func surfaceTenant(ctx context.Context, database *sql.DB, reader documentsdomain.ReviewDueReader, writer documentsdomain.ReviewSurfaceWriter, tenantID string, now time.Time) (dueCount int, surfacedCount int, err error) {
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback()

	if err := authz.BypassSystem(ctx, tx); err != nil {
		return 0, 0, err
	}
	if err := authz.SeedTxTenant(ctx, tx, tenantID); err != nil {
		return 0, 0, err
	}

	due, err := reader.ListDueForReview(ctx, tx, tenantID, now, BatchSize)
	if err != nil {
		return 0, 0, err
	}

	surfaced, err := writer.MarkSurfaced(ctx, tx, tenantID, now)
	if err != nil {
		return 0, 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, err
	}

	return len(due), len(surfaced), nil
}
