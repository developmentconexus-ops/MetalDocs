package stuck_instance_watchdog

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/riverqueue/river"

	"metaldocs/internal/modules/approval/application"
	"metaldocs/internal/modules/iam/authz"
	platformdb "metaldocs/internal/platform/db"
)

const (
	// JobName identifies this job type to River and in logs.
	JobName = "stuck_instance_watchdog"
	// StuckAfter is the age threshold past which an in-progress approval
	// instance is reported stuck.
	StuckAfter = 7 * 24 * time.Hour
	// BatchSize caps how many stuck instances are reported per tick.
	BatchSize = 50
	// SystemActor is the stable system principal alerts are attributed to.
	SystemActor = "system:watchdog"
)

// StuckInstance is one approval instance the watchdog found stuck
// in_progress past StuckAfter.
type StuckInstance struct {
	ID          string
	TenantID    string
	DocumentID  string
	SubmittedBy string
	DriftPolicy string
}

type governanceEmitter interface {
	Emit(ctx context.Context, tx platformdb.Tx, e application.GovernanceEvent) error
}

// StuckInstanceWatchdogArgs is the (empty) River job payload for the
// stuck-instance watchdog tick. The job carries no per-run parameters — all
// tick behavior is derived from the database at run time.
type StuckInstanceWatchdogArgs struct{}

// Kind implements river.JobArgs, identifying this job type to River.
func (StuckInstanceWatchdogArgs) Kind() string { return JobName }

// StuckInstanceWatchdogWorker is the River worker that runs the watchdog tick.
// Cluster-wide single-runner is provided by River's leader-elected periodic
// insert plus its queue dequeue semantics; no advisory lock is taken here
// (ADR 0067 §H-PRE-1 retires the lease-scheduler-era advisory lock as
// redundant under River).
type StuckInstanceWatchdogWorker struct {
	river.WorkerDefaults[StuckInstanceWatchdogArgs]

	runner  platformdb.TxRunner
	emitter governanceEmitter
}

// NewWorker constructs a StuckInstanceWatchdogWorker. runner replaces a raw
// *sql.DB (A5.1, Lane E issue #92): this is a system-path TxRunner consumer —
// ctx carries only authz.WithBackgroundBypass, never a platform/tenant
// identity, so the TxRunner chokepoint's ctx-seed is a deliberate no-op here,
// identical to the prior raw db.BeginTx(ctx, nil) behavior. The manual
// authz.BypassSystem / authz.SeedTxTenant calls below are unchanged.
func NewWorker(runner platformdb.TxRunner, emitter governanceEmitter) *StuckInstanceWatchdogWorker {
	return &StuckInstanceWatchdogWorker{
		runner:  runner,
		emitter: emitter,
	}
}

// Work runs one watchdog tick under a background-bypass authz context (no
// HTTP-request identity exists here).
func (w *StuckInstanceWatchdogWorker) Work(ctx context.Context, job *river.Job[StuckInstanceWatchdogArgs]) error {
	// Background root: permits listStuckInstances/emitStuckAlert's own
	// authz.BypassSystem reads (fail-closed off any HTTP path — ADR 0022
	// Phase 7, CWE-269). The watchdog is alert-only (ADR 0068) — no cancel
	// path roots here.
	ctx = authz.WithBackgroundBypass(ctx)
	return run(ctx, w.runner, w.emitter)
}

// run executes one watchdog tick. Cluster-wide single-runner is provided by
// River leader-election + queue dequeue semantics; the lease-scheduler-era
// advisory lock is not part of this body (ADR 0067 §H-PRE-1). The watchdog is
// alert-only (ADR 0068): every stuck instance emits exactly one
// approval.instance.stuck_alert governance event — no cancel side effect.
func run(ctx context.Context, runner platformdb.TxRunner, emitter governanceEmitter) error {
	stuck, err := listStuckInstances(ctx, runner)
	if err != nil {
		slog.ErrorContext(ctx, "stuck_instance_watchdog: list stuck instances failed",
			"job", JobName, "error", err)
		return err
	}

	stuckDetected := len(stuck)
	alertsEmitted := 0
	var runErr error

	for _, inst := range stuck {
		if err := emitStuckAlert(ctx, runner, emitter, inst); err != nil {
			slog.ErrorContext(ctx, "stuck_instance_watchdog: emit stuck alert failed",
				"job", JobName, "instance_id", inst.ID, "tenant_id", inst.TenantID, "error", err)
			runErr = errors.Join(runErr, err)
			continue
		}
		alertsEmitted++
	}

	slog.InfoContext(ctx, "stuck_instance_watchdog: tick complete",
		"job", JobName,
		"stuck_detected", stuckDetected,
		"alerts_emitted", alertsEmitted)

	return runErr
}

func listStuckInstances(ctx context.Context, runner platformdb.TxRunner) ([]StuckInstance, error) {
	var out []StuckInstance
	err := runner.Do(ctx, func(tx *sql.Tx) error {
		if err := authz.BypassSystem(ctx, tx); err != nil {
			return err
		}

		rows, err := tx.QueryContext(ctx, `
SELECT
    ai.id::text,
    ai.tenant_id::text,
    ai.document_id::text,
    ai.submitted_by,
    COALESCE(asi.on_eligibility_drift_snapshot, '') AS drift_policy
FROM approval_instances ai
LEFT JOIN approval_stage_instances asi
  ON asi.approval_instance_id = ai.id
 AND asi.status = 'active'
WHERE ai.status = 'in_progress'
  AND ai.submitted_at < now() - ($2 * interval '1 second')
LIMIT $1`, BatchSize, int64(StuckAfter/time.Second))
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()

		out = make([]StuckInstance, 0, BatchSize)
		for rows.Next() {
			var inst StuckInstance
			if err := rows.Scan(&inst.ID, &inst.TenantID, &inst.DocumentID, &inst.SubmittedBy, &inst.DriftPolicy); err != nil {
				return err
			}
			out = append(out, inst)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func emitStuckAlert(ctx context.Context, runner platformdb.TxRunner, emitter governanceEmitter, inst StuckInstance) error {
	return runner.Do(ctx, func(tx *sql.Tx) error {
		if err := authz.BypassSystem(ctx, tx); err != nil {
			return err
		}
		// Carrier-less watchdog ctx → the TxRunner chokepoint's ctx-seed is a
		// deliberate no-op here (see NewWorker doc). Seed the stuck instance's
		// tenant so FORCE RLS backstops the governance_events write (a
		// tenant-scoped table); the alert concerns exactly one tenant's
		// instance (validation-contract §2.6/§4: no tenant-scoped async write
		// runs unseeded outside the §2.4 allowlist).
		if err := authz.SeedTxTenant(ctx, tx, inst.TenantID); err != nil {
			return err
		}

		payload, err := json.Marshal(map[string]any{
			"instance_id":   inst.ID,
			"document_id":   inst.DocumentID,
			"submitted_by":  inst.SubmittedBy,
			"drift_policy":  inst.DriftPolicy,
			"watchdog_rule": "stuck_instance_7d",
		})
		if err != nil {
			return err
		}

		return emitter.Emit(ctx, tx, application.GovernanceEvent{
			TenantID:     inst.TenantID,
			EventType:    "approval.instance.stuck_alert",
			ActorUserID:  SystemActor,
			ResourceType: "approval_instance",
			ResourceID:   inst.ID,
			Reason:       "stuck_watchdog_alert",
			PayloadJSON:  payload,
			OccurredAt:   time.Now().UTC(),
		})
	})
}
