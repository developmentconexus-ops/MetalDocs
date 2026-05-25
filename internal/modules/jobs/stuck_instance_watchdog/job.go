package stuck_instance_watchdog

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"metaldocs/internal/modules/documents/approval/application"
	"metaldocs/internal/modules/jobs/scheduler"
)

const (
	JobName     = "stuck_instance_watchdog"
	StuckAfter  = 7 * 24 * time.Hour
	BatchSize   = 50
	SystemActor = "system:watchdog"
	BypassGUC   = "watchdog"
)

type StuckInstance struct {
	ID          string
	TenantID    string
	DocumentID  string
	SubmittedBy string
	DriftPolicy string
}

type cancelSvcInterface interface {
	CancelInstance(ctx context.Context, db *sql.DB, in application.CancelInput) (application.CancelResult, error)
}

type governanceEmitter interface {
	Emit(ctx context.Context, tx *sql.Tx, e application.GovernanceEvent) error
}

func New(db *sql.DB, cancelSvc cancelSvcInterface, emitter governanceEmitter) scheduler.JobFunc {
	return func(ctx context.Context, epoch int64) error {
		unlock, err := acquireRunLock(ctx, db)
		if err != nil {
			slog.ErrorContext(ctx, "stuck_instance_watchdog: acquire run lock failed",
				"job", JobName, "epoch", epoch, "error", err)
			return err
		}
		defer unlock()

		stuck, err := listStuckInstances(ctx, db)
		if err != nil {
			slog.ErrorContext(ctx, "stuck_instance_watchdog: list stuck instances failed",
				"job", JobName, "epoch", epoch, "error", err)
			return err
		}

		stuckDetected := len(stuck)
		autoCancelled := 0
		alertsEmitted := 0
		var runErr error

		for _, inst := range stuck {
			if inst.DriftPolicy == "auto_cancel" {
				_, err := cancelSvc.CancelInstance(ctx, db, application.CancelInput{
					TenantID:                inst.TenantID,
					InstanceID:              inst.ID,
					ExpectedRevisionVersion: 0,
					ActorUserID:             SystemActor,
					Reason:                  "stuck_watchdog_auto_cancel",
					BypassAuthz:             true,
				})
				if err != nil {
					slog.ErrorContext(ctx, "stuck_instance_watchdog: auto-cancel failed",
						"job", JobName, "epoch", epoch, "instance_id", inst.ID, "tenant_id", inst.TenantID, "error", err)
					runErr = errors.Join(runErr, err)
					continue
				}
				autoCancelled++
				continue
			}

			if err := emitStuckAlert(ctx, db, emitter, inst); err != nil {
				slog.ErrorContext(ctx, "stuck_instance_watchdog: emit stuck alert failed",
					"job", JobName, "epoch", epoch, "instance_id", inst.ID, "tenant_id", inst.TenantID, "error", err)
				runErr = errors.Join(runErr, err)
				continue
			}
			alertsEmitted++
		}

		slog.InfoContext(ctx, "stuck_instance_watchdog: tick complete",
			"job", JobName,
			"epoch", epoch,
			"stuck_detected", stuckDetected,
			"auto_cancelled", autoCancelled,
			"alerts_emitted", alertsEmitted)

		return runErr
	}
}

func acquireRunLock(ctx context.Context, db *sql.DB) (func(), error) {
	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, err
	}

	var locked bool
	if err := conn.QueryRowContext(ctx, `SELECT pg_try_advisory_lock(hashtext($1))`, JobName).Scan(&locked); err != nil {
		_ = conn.Close()
		return nil, err
	}
	if !locked {
		_ = conn.Close()
		return nil, fmt.Errorf("watchdog run lock not acquired")
	}

	return func() {
		_, _ = conn.ExecContext(context.Background(), `SELECT pg_advisory_unlock(hashtext($1))`, JobName)
		_ = conn.Close()
	}, nil
}

func listStuckInstances(ctx context.Context, db *sql.DB) ([]StuckInstance, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err := setBypassAuthzGUC(ctx, tx); err != nil {
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
  AND ai.submitted_at < now() - interval '7 days'
LIMIT $1`, BatchSize)
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

	if err := setBypassAuthzGUC(ctx, tx); err != nil {
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

func setBypassAuthzGUC(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `SELECT set_config('metaldocs.bypass_authz', $1, true)`, BypassGUC)
	return err
}
