package stuck_instance_watchdog

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/riverqueue/river"

	"metaldocs/internal/modules/documents/approval/application"
	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/platform/db"
)

const (
	JobName     = "stuck_instance_watchdog"
	StuckAfter  = 7 * 24 * time.Hour
	BatchSize   = 50
	SystemActor = "system:watchdog"
)

type StuckInstance struct {
	ID          string
	TenantID    string
	DocumentID  string
	SubmittedBy string
	DriftPolicy string
}

type governanceEmitter interface {
	Emit(ctx context.Context, tx db.Tx, e application.GovernanceEvent) error
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

	database *sql.DB
	emitter  governanceEmitter
}

// NewWorker constructs a StuckInstanceWatchdogWorker.
func NewWorker(database *sql.DB, emitter governanceEmitter) *StuckInstanceWatchdogWorker {
	return &StuckInstanceWatchdogWorker{
		database: database,
		emitter:  emitter,
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
	return run(ctx, w.database, w.emitter)
}

// run executes one watchdog tick. Cluster-wide single-runner is provided by
// River leader-election + queue dequeue semantics; the lease-scheduler-era
// advisory lock is not part of this body (ADR 0067 §H-PRE-1). The watchdog is
// alert-only (ADR 0068): every stuck instance emits exactly one
// approval.instance.stuck_alert governance event — no cancel side effect.
func run(ctx context.Context, database *sql.DB, emitter governanceEmitter) error {
	stuck, err := listStuckInstances(ctx, database)
	if err != nil {
		slog.ErrorContext(ctx, "stuck_instance_watchdog: list stuck instances failed",
			"job", JobName, "error", err)
		return err
	}

	stuckDetected := len(stuck)
	alertsEmitted := 0
	var runErr error

	for _, inst := range stuck {
		if err := emitStuckAlert(ctx, database, emitter, inst); err != nil {
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

func listStuckInstances(ctx context.Context, db *sql.DB) ([]StuckInstance, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err := authz.BypassSystem(ctx, tx); err != nil {
		return nil, err
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
		return nil, err
	}
	defer rows.Close()

	out := make([]StuckInstance, 0, BatchSize)
	for rows.Next() {
		var inst StuckInstance
		if err := rows.Scan(&inst.ID, &inst.TenantID, &inst.DocumentID, &inst.SubmittedBy, &inst.DriftPolicy); err != nil {
			return nil, err
		}
		out = append(out, inst)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return out, nil
}

func emitStuckAlert(ctx context.Context, db *sql.DB, emitter governanceEmitter, inst StuckInstance) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := authz.BypassSystem(ctx, tx); err != nil {
		return err
	}
	// Carrier-less watchdog ctx → the F3.1 TxRunner chokepoint never runs here
	// (this is a raw db.BeginTx). Seed the stuck instance's tenant so FORCE RLS
	// backstops the governance_events write (a tenant-scoped table); the alert
	// concerns exactly one tenant's instance (validation-contract §2.6/§4: no
	// tenant-scoped async write runs unseeded outside the §2.4 allowlist).
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

	if err := emitter.Emit(ctx, tx, application.GovernanceEvent{
		TenantID:     inst.TenantID,
		EventType:    "approval.instance.stuck_alert",
		ActorUserID:  SystemActor,
		ResourceType: "approval_instance",
		ResourceID:   inst.ID,
		Reason:       "stuck_watchdog_alert",
		PayloadJSON:  payload,
		OccurredAt:   time.Now().UTC(),
	}); err != nil {
		return err
	}

	return tx.Commit()
}
