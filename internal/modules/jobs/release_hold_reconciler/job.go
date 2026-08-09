// Package release_hold_reconciler holds the River periodic job that sweeps
// release generations stuck in readiness hold and ALERTS on them (ADR 0085
// Stage C W2, §"Observability, attribution, reconciliation": "a River periodic
// maintenance job (ADR 0067 pattern, metaldocs-jobs) sweeps generations stuck
// in readiness hold beyond a threshold (lost fact, dead-lettered consumer,
// dead timer) and alerts — alert-only posture per ADR 0068; duplicate-safe
// because recovery is always re-run the idempotent evaluation").
//
// It contains NO raw SQL against public.release_generations: it calls only the
// approval module's published read-port ReleaseHoldReader
// (ListTenantsWithStuckHolds for the cross-tenant enumeration,
// ListStuckHolds for the per-tenant read) — mirroring approval_sla_surfacer's
// boundary discipline.
//
// ALERT-ONLY, and strictly so (ADR 0068): this job never mutates a generation,
// a document, or a hold reason, and it does NOT re-enqueue the coordinator
// evaluation. Auto-recovery was deliberately NOT taken here even though the
// evaluation is idempotent and duplicate-safe: an evaluation stamps
// last_evaluated_at, which is the very signal this sweep measures, so a
// self-healing sweep would reset its own detector every tick and turn a
// chronic hold into a silent oscillation. Recovery ("re-run the idempotent
// evaluation") stays an operator lever on top of an honest alert.
package release_hold_reconciler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/riverqueue/river"

	"metaldocs/internal/modules/approval/application"
	approvaldomain "metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/platform/db"
)

const (
	// JobName identifies this job type to River and in logs.
	JobName = "release_hold_reconciler"

	// HoldStuckAfter is the reconciliation threshold: a generation still
	// unreleased AND not evaluated within this window is reported stuck.
	// Constant (not env-tunable) for the same reason as
	// stuck_instance_watchdog.StuckAfter — the threshold is part of the
	// operational contract, not a per-deployment knob. Sized well above the
	// happy-path latency of the whole materialization pipeline (render dispatch
	// + DOCX + PDF), so crossing it means a fact, a consumer or a timer is
	// genuinely lost rather than merely slow.
	HoldStuckAfter = 30 * time.Minute

	// BatchSize caps how many stuck generations are reported per tenant per
	// tick. A tenant with more than this many stuck generations has a systemic
	// failure that the first BatchSize alerts already prove.
	BatchSize = 100

	// SystemActor is the stable system principal the alert is attributed to.
	// Never a user id: nobody caused this hold by acting.
	SystemActor = "system:release-hold-reconciler"

	// AlertEventType is the governance event this job emits per stuck
	// generation — the durable half of the alert (the slog warning is the
	// operational half). Mirrors stuck_instance_watchdog's
	// approval.instance.stuck_alert.
	AlertEventType = "release.generation.stuck_alert"

	// alertRule names the detection rule in the event payload, so an operator
	// reading governance_events knows which sweep produced the row.
	alertRule = "release_hold_stuck_30m"
)

// governanceEmitter is the narrow write surface this job needs (the approval
// module's EventEmitter). Declared locally so the job depends on a method set,
// not on a concrete emitter — mirrors stuck_instance_watchdog.
type governanceEmitter interface {
	Emit(ctx context.Context, tx db.Tx, e application.GovernanceEvent) error
}

// ReleaseHoldReconcilerArgs is the (empty) River job payload for the
// reconciliation tick. The job carries no per-run parameters — all tick
// behavior is derived from the database at run time.
type ReleaseHoldReconcilerArgs struct{}

// Kind implements river.JobArgs, identifying this job type to River.
func (ReleaseHoldReconcilerArgs) Kind() string { return JobName }

// ReleaseHoldReconcilerWorker is the River worker that runs the reconciliation
// tick. Cluster-wide single-runner is provided by River's leader-elected
// periodic insert plus its queue dequeue semantics; no advisory lock is taken
// here (ADR 0067 §H-PRE-1).
type ReleaseHoldReconcilerWorker struct {
	river.WorkerDefaults[ReleaseHoldReconcilerArgs]

	database *sql.DB
	reader   approvaldomain.ReleaseHoldReader
	emitter  governanceEmitter
}

// NewWorker constructs a ReleaseHoldReconcilerWorker.
func NewWorker(database *sql.DB, reader approvaldomain.ReleaseHoldReader, emitter governanceEmitter) *ReleaseHoldReconcilerWorker {
	return &ReleaseHoldReconcilerWorker{
		database: database,
		reader:   reader,
		emitter:  emitter,
	}
}

// Work runs one reconciliation tick under a background-bypass authz context
// (no HTTP-request identity exists here — ADR 0022 Phase 7).
func (w *ReleaseHoldReconcilerWorker) Work(ctx context.Context, job *river.Job[ReleaseHoldReconcilerArgs]) error {
	ctx = authz.WithBackgroundBypass(ctx)
	return run(ctx, w.database, w.reader, w.emitter, time.Now().UTC())
}

// run executes one reconciliation tick in two phases, mirroring
// approval_sla_surfacer.run:
//
//  1. a single cross-tenant SYSTEM READ (ListTenantsWithStuckHolds) under the
//     scheduler bypass with no tenant GUC seeded — enumerating tenants is safe
//     unseeded;
//  2. for EACH tenant returned, a NEW tx that seeds authz.BypassSystem THEN
//     authz.SeedTxTenant(tenantID) before reading that tenant's stuck holds and
//     emitting one alert per generation (governance_events is tenant-scoped and
//     FORCE RLS, so the seed is what lets the write land).
//
// A tenant whose alerts fail does not abort the sweep: the error is joined and
// the remaining tenants are still reported.
func run(ctx context.Context, database *sql.DB, reader approvaldomain.ReleaseHoldReader, emitter governanceEmitter, now time.Time) error {
	tenantIDs, err := listTenantsWithStuckHolds(ctx, database, reader, now)
	if err != nil {
		slog.ErrorContext(ctx, "release_hold_reconciler: list tenants with stuck holds failed",
			"job", JobName, "error", err)
		return err
	}

	stuckDetected := 0
	alertsEmitted := 0
	var runErr error

	for _, tenantID := range tenantIDs {
		detected, emitted, err := reconcileTenant(ctx, database, reader, emitter, tenantID, now)
		if err != nil {
			slog.ErrorContext(ctx, "release_hold_reconciler: reconcile tenant failed",
				"job", JobName, "tenant_id", tenantID, "error", err)
			runErr = errors.Join(runErr, err)
			continue
		}
		stuckDetected += detected
		alertsEmitted += emitted
	}

	slog.InfoContext(ctx, "release_hold_reconciler: tick complete",
		"job", JobName,
		"tenants_with_stuck_holds", len(tenantIDs),
		"stuck_detected", stuckDetected,
		"alerts_emitted", alertsEmitted)

	return runErr
}

// listTenantsWithStuckHolds opens a single bypass tx (no tenant GUC seeded —
// system-level cross-tenant read, safe unseeded) and returns the distinct
// tenant_ids that currently hold at least one stuck generation.
func listTenantsWithStuckHolds(ctx context.Context, database *sql.DB, reader approvaldomain.ReleaseHoldReader, now time.Time) ([]string, error) {
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	if err := authz.BypassSystem(ctx, tx); err != nil {
		return nil, err
	}

	tenantIDs, err := reader.ListTenantsWithStuckHolds(ctx, tx, now, HoldStuckAfter)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return tenantIDs, nil
}

// reconcileTenant opens a NEW tx seeded to exactly one tenant
// (authz.BypassSystem then authz.SeedTxTenant(tenantID)) and, within it, reads
// that tenant's stuck holds and emits one alert per generation. Read and alerts
// share the tx so the alerts are a consistent snapshot of what the sweep saw.
func reconcileTenant(ctx context.Context, database *sql.DB, reader approvaldomain.ReleaseHoldReader, emitter governanceEmitter, tenantID string, now time.Time) (detected int, emitted int, err error) {
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return 0, 0, err
	}
	defer func() { _ = tx.Rollback() }()

	if err := authz.BypassSystem(ctx, tx); err != nil {
		return 0, 0, err
	}
	// Carrier-less reconciler ctx → the TxRunner chokepoint never runs here
	// (this is a raw db.BeginTx). Seed the tenant so FORCE RLS backstops both
	// the release_generations read and the governance_events write.
	if err := authz.SeedTxTenant(ctx, tx, tenantID); err != nil {
		return 0, 0, err
	}

	stuck, err := reader.ListStuckHolds(ctx, tx, tenantID, now, HoldStuckAfter, BatchSize)
	if err != nil {
		return 0, 0, err
	}

	for _, hold := range stuck {
		if err := alertStuckHold(ctx, tx, emitter, hold, now); err != nil {
			return 0, 0, err
		}
		emitted++
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, err
	}
	return len(stuck), emitted, nil
}

// alertStuckHold is the whole action of this job: a structured warning plus one
// durable governance event. No state is touched (ADR 0068).
func alertStuckHold(ctx context.Context, tx *sql.Tx, emitter governanceEmitter, hold approvaldomain.StuckReleaseHoldView, now time.Time) error {
	ageSeconds := int64(now.Sub(hold.CreatedAt) / time.Second)
	holdReason := hold.HoldReason
	if holdReason == "" {
		// A generation that was NEVER evaluated carries no reason. Saying
		// "never_evaluated" is the honest label; leaving it blank would read as
		// a missing field rather than the most severe case.
		holdReason = "never_evaluated"
	}

	slog.WarnContext(ctx, "release_hold_reconciler: release generation stuck in readiness hold",
		"job", JobName,
		"generation_id", hold.GenerationID,
		"tenant_id", hold.TenantID,
		"document_id", hold.DocumentID,
		"approval_instance_id", hold.ApprovalInstanceID,
		"hold_reason", holdReason,
		"hold_detail", hold.HoldDetail,
		"age_seconds", ageSeconds,
		"last_evaluated_at", lastEvaluatedLog(hold.LastEvaluatedAt),
		"threshold_seconds", int64(HoldStuckAfter/time.Second))

	payload, err := json.Marshal(map[string]any{
		"release_generation_id": hold.GenerationID,
		"document_id":           hold.DocumentID,
		"approval_instance_id":  hold.ApprovalInstanceID,
		"hold_reason":           holdReason,
		"hold_detail":           hold.HoldDetail,
		"age_seconds":           ageSeconds,
		"last_evaluated_at":     lastEvaluatedLog(hold.LastEvaluatedAt),
		"threshold_seconds":     int64(HoldStuckAfter / time.Second),
		"reconciler_rule":       alertRule,
	})
	if err != nil {
		return err
	}

	return emitter.Emit(ctx, tx, application.GovernanceEvent{
		TenantID:     hold.TenantID,
		EventType:    AlertEventType,
		ActorUserID:  SystemActor,
		ResourceType: "release_generation",
		ResourceID:   hold.GenerationID,
		Reason:       "release_hold_stuck_alert",
		PayloadJSON:  payload,
		OccurredAt:   now,
	})
}

// lastEvaluatedLog renders the nullable last-evaluation instant for logs and
// payloads: "" when the generation was never evaluated (no substitute value —
// a zero timestamp would read as a real 0001-01-01 evaluation).
func lastEvaluatedLog(at *time.Time) string {
	if at == nil {
		return ""
	}
	return at.UTC().Format(time.RFC3339)
}
