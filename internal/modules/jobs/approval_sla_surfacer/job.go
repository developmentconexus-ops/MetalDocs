// Package approval_sla_surfacer holds the River periodic job that surfaces
// approval stage instances overdue against their SLA due_at (F8, milestone
// 2b approval-kernel-backend, design spec.md §4/W4). It contains NO raw SQL
// against public.approval_stage_instances — it calls only the approval
// module's published ports: SLAOverdueReader.ListOverdueStages (read, for
// count/log), SLAOverdueReader.ListTenantsWithOverdueStages (the cross-tenant
// enumeration read), SLASurfaceWriter.MarkStagesOverdueSurfaced (the
// idempotent, per-tenant-seeded side effect), and (Task 7)
// ApprovalNotificationEnqueuer.EnqueueApprovalNotificationTx, which tells the
// stage's eligible actors — in the SAME tx as the marker write — that they
// missed their deadline.
//
// This is a genuine SIBLING job to document_review_surfacer, not an
// extension of it: an approval stage's due_at (this instance's SLA clock,
// activated per-stage at submit/advance time) is a distinct concept from a
// published document's periodic review_due_at (per-document eQMS review
// cadence) — F8 Interview #2. Reusing one job/port/column for both would
// conflate two unrelated due-date mechanisms. Alert-only (ADR 0068): this job
// never mutates instance/stage status or due_at itself, only the
// sla_surfaced_at idempotency marker — mirrors stuck_instance_watchdog and
// document_review_surfacer's own alert-only contract.
package approval_sla_surfacer

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/riverqueue/river"

	approvaldomain "metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/iam/authz"
	platformdb "metaldocs/internal/platform/db"
)

// JobName identifies this job type to River and in logs.
const JobName = "approval_sla_surfacer"

// BatchSize caps how many overdue stages the read-port reports per tick (for
// the count/log only — MarkStagesOverdueSurfaced's UPDATE has no LIMIT and
// covers every eligible row in one statement).
const BatchSize = 100

// ApprovalSLASurfacerArgs is the (empty) River job payload for the SLA
// surfacer tick. The job carries no per-run parameters — all tick behavior
// is derived from the database at run time.
type ApprovalSLASurfacerArgs struct{}

// Kind implements river.JobArgs, identifying this job type to River.
func (ApprovalSLASurfacerArgs) Kind() string { return JobName }

// ApprovalSLASurfacerWorker is the River worker that runs the approval SLA
// surfacer tick. Cluster-wide single-runner is provided by River's
// leader-elected periodic insert plus its queue dequeue semantics; no
// advisory lock is taken here (mirrors document_review_surfacer, ADR 0067
// §H-PRE-1).
type ApprovalSLASurfacerWorker struct {
	river.WorkerDefaults[ApprovalSLASurfacerArgs]

	runner   platformdb.TxRunner
	reader   approvaldomain.SLAOverdueReader
	writer   approvaldomain.SLASurfaceWriter
	notifier approvaldomain.ApprovalNotificationEnqueuer
}

// NewWorker constructs an ApprovalSLASurfacerWorker. runner replaces a raw
// *sql.DB (A5.1, Lane E issue #92): this is a system-path TxRunner consumer —
// ctx carries only authz.WithBackgroundBypass, never a platform/tenant
// identity, so the TxRunner chokepoint's ctx-seed is a deliberate no-op here,
// identical to the prior raw db.BeginTx(ctx, nil) behavior. The manual
// authz.BypassSystem / authz.SeedTxTenant calls below are unchanged.
func NewWorker(runner platformdb.TxRunner, reader approvaldomain.SLAOverdueReader, writer approvaldomain.SLASurfaceWriter, notifier approvaldomain.ApprovalNotificationEnqueuer) *ApprovalSLASurfacerWorker {
	return &ApprovalSLASurfacerWorker{
		runner:   runner,
		reader:   reader,
		writer:   writer,
		notifier: notifier,
	}
}

// Work runs one surfacer tick under a background-bypass authz context (no
// HTTP-request identity exists here).
func (w *ApprovalSLASurfacerWorker) Work(ctx context.Context, job *river.Job[ApprovalSLASurfacerArgs]) error {
	ctx = authz.WithBackgroundBypass(ctx)
	return run(ctx, w.runner, w.reader, w.writer, w.notifier, time.Now().UTC())
}

// run executes one SLA-surfacer tick in two phases, mirroring
// document_review_surfacer.run:
//
//  1. a single cross-tenant SYSTEM READ (ListTenantsWithOverdueStages) under
//     the scheduler bypass with no tenant GUC seeded — enumerating tenants
//     is safe unseeded;
//  2. for EACH tenant returned, a NEW tx that seeds authz.BypassSystem THEN
//     authz.SeedTxTenant(tenantID) before calling ListOverdueStages (for the
//     count/log) and MarkStagesOverdueSurfaced (the idempotent side effect),
//     passing that same tenantID explicitly to both.
func run(ctx context.Context, runner platformdb.TxRunner, reader approvaldomain.SLAOverdueReader, writer approvaldomain.SLASurfaceWriter, notifier approvaldomain.ApprovalNotificationEnqueuer, now time.Time) error {
	tenantIDs, err := listTenantsWithOverdueStages(ctx, runner, reader, now)
	if err != nil {
		slog.ErrorContext(ctx, "approval_sla_surfacer: list tenants with overdue stages failed",
			"job", JobName, "error", err)
		return err
	}

	var totalOverdue, totalSurfaced int
	var runErr error

	for _, tenantID := range tenantIDs {
		overdue, surfaced, err := surfaceTenant(ctx, runner, reader, writer, notifier, tenantID, now)
		if err != nil {
			slog.ErrorContext(ctx, "approval_sla_surfacer: surface tenant failed",
				"job", JobName, "tenant_id", tenantID, "error", err)
			runErr = errors.Join(runErr, err)
			continue
		}
		totalOverdue += overdue
		totalSurfaced += surfaced
	}

	slog.InfoContext(ctx, "approval_sla_surfacer: tick complete",
		"job", JobName,
		"tenants_with_overdue", len(tenantIDs),
		"overdue_count", totalOverdue,
		"surfaced_count", totalSurfaced)

	return runErr
}

// listTenantsWithOverdueStages opens a single bypass tx (no tenant GUC
// seeded — system-level cross-tenant read, safe unseeded) and returns the
// distinct tenant_ids with at least one overdue, not-yet-surfaced active
// stage instance.
func listTenantsWithOverdueStages(ctx context.Context, runner platformdb.TxRunner, reader approvaldomain.SLAOverdueReader, now time.Time) ([]string, error) {
	var tenantIDs []string
	err := runner.Do(ctx, func(tx *sql.Tx) error {
		if err := authz.BypassSystem(ctx, tx); err != nil {
			return err
		}

		ids, err := reader.ListTenantsWithOverdueStages(ctx, tx, now)
		if err != nil {
			return err
		}
		tenantIDs = ids
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tenantIDs, nil
}

// surfaceTenant opens a NEW tx seeded to exactly one tenant
// (authz.BypassSystem then authz.SeedTxTenant(tenantID)) and, within it,
// reads the tenant's overdue stages (for the count/log) and marks them
// surfaced, passing tenantID explicitly to both ports.
func surfaceTenant(ctx context.Context, runner platformdb.TxRunner, reader approvaldomain.SLAOverdueReader, writer approvaldomain.SLASurfaceWriter, notifier approvaldomain.ApprovalNotificationEnqueuer, tenantID string, now time.Time) (overdueCount int, surfacedCount int, err error) {
	err = runner.Do(ctx, func(tx *sql.Tx) error {
		if err := authz.BypassSystem(ctx, tx); err != nil {
			return err
		}
		if err := authz.SeedTxTenant(ctx, tx, tenantID); err != nil {
			return err
		}

		overdue, err := reader.ListOverdueStages(ctx, tx, tenantID, now, BatchSize)
		if err != nil {
			return err
		}

		surfaced, err := writer.MarkStagesOverdueSurfaced(ctx, tx, tenantID, now)
		if err != nil {
			return err
		}

		// Same tx as the marker: the outbox discipline. If the enqueue fails, the
		// marker rolls back too, so the next tick retries the whole thing rather
		// than leaving a stage marked-but-unannounced.
		for _, s := range surfaced {
			if len(s.EligibleActorIDs) == 0 {
				continue
			}
			if err := notifier.EnqueueApprovalNotificationTx(ctx, tx, approvaldomain.ApprovalNotificationArgs{
				EventID:          uuid.New().String(),
				TenantID:         s.TenantID,
				EventType:        approvaldomain.EventTypeApprovalOverdue,
				SubjectKind:      s.SubjectKind,
				SubjectID:        s.SubjectID,
				RecipientUserIDs: s.EligibleActorIDs,
				// No ActorUserID: an SLA tick has no human cause.
				OccurredAt: now,
			}); err != nil {
				return err
			}
		}

		overdueCount = len(overdue)
		surfacedCount = len(surfaced)
		return nil
	})
	if err != nil {
		return 0, 0, err
	}
	return overdueCount, surfacedCount, nil
}
