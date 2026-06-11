package scheduler

import (
	"context"
	"database/sql"
	"log/slog"
)

// RunLeaseReaper deletes scheduler job leases whose expiry is more than 10
// minutes past and emits a structured log line per reclaimed lease.
//
// Job leases are system-scoped maintenance artifacts: job_name is a scheduler
// job identifier, not a tenant resource. The previous implementation tried to
// attribute each reap to a tenant by joining job_name against
// public.documents.id — a cross-schema comparison that never matched, so
// tenant resolution failed for every reaped lease, the governance insert was
// skipped, and every tick with a reap returned an error (F-19, REQ-ASYNC-4).
// governance_events.tenant_id is NOT NULL with no system-tenant convention,
// so the durable record for these infrastructure events is the structured
// log stream, per the Wave 1.7 card.
func RunLeaseReaper(db *sql.DB) JobFunc {
	return func(ctx context.Context, epoch int64) error {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		rows, err := tx.QueryContext(ctx, `
WITH locked AS (
	SELECT jl.job_name
	FROM metaldocs.job_leases jl
	WHERE jl.expires_at < now() - interval '10 minutes'
	FOR UPDATE SKIP LOCKED
),
deleted AS (
	DELETE FROM metaldocs.job_leases jl
	WHERE jl.job_name IN (SELECT job_name FROM locked)
	RETURNING jl.job_name, jl.leader_id, jl.lease_epoch
)
SELECT d.job_name, d.leader_id, d.lease_epoch
FROM deleted d
`)
		if err != nil {
			return err
		}
		defer rows.Close()

		reclaimed := 0
		for rows.Next() {
			var jobName string
			var leaderID string
			var leaseEpoch int64
			if err := rows.Scan(&jobName, &leaderID, &leaseEpoch); err != nil {
				return err
			}
			// Warn: an expired lease means a leader stopped renewing
			// (crash, partition, or shutdown without release).
			slog.WarnContext(ctx, "lease_reaper: expired lease reclaimed",
				"job", "lease_reaper",
				"job_name", jobName,
				"leader_id", leaderID,
				"lease_epoch", leaseEpoch,
				"epoch", epoch)
			reclaimed++
		}
		if err := rows.Err(); err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}

		slog.InfoContext(ctx, "lease_reaper: tick complete",
			"job", "lease_reaper",
			"epoch", epoch,
			"reclaimed", reclaimed)
		return nil
	}
}
