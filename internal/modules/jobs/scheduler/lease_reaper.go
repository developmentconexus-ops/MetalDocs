package scheduler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/lib/pq"
)

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
SELECT
	d.job_name,
	d.leader_id,
	d.lease_epoch,
	(SELECT doc.tenant_id::text FROM public.documents doc WHERE doc.id::text = d.job_name LIMIT 1) AS tenant_id
FROM deleted d
`)
		if err != nil {
			return err
		}
		defer rows.Close()

		reclaimed := 0
		jobNames := make([]string, 0)
		payloads := make([]string, 0)
		tenantIDs := make([]string, 0)
		var rowErrs []error
		for rows.Next() {
			var jobName string
			var leaderID string
			var leaseEpoch int64
			var tenantID sql.NullString
			if err := rows.Scan(&jobName, &leaderID, &leaseEpoch, &tenantID); err != nil {
				slog.ErrorContext(ctx, "lease_reaper: scan reclaimed lease failed", "error", err)
				rowErrs = append(rowErrs, err)
				continue
			}

			payloadJSON, err := json.Marshal(map[string]any{
				"job_name":    jobName,
				"leader_id":   leaderID,
				"lease_epoch": leaseEpoch,
			})
			if err != nil {
				slog.ErrorContext(ctx, "lease_reaper: marshal governance payload failed",
					"job_name", jobName,
					"leader_id", leaderID,
					"lease_epoch", leaseEpoch,
					"error", err)
				rowErrs = append(rowErrs, err)
				continue
			}

			if !tenantID.Valid || tenantID.String == "" {
				err := fmt.Errorf("lease_reaper: tenant attribution unavailable for job %q", jobName)
				slog.ErrorContext(ctx, "lease_reaper: resolve tenant failed",
					"job_name", jobName,
					"leader_id", leaderID,
					"lease_epoch", leaseEpoch,
					"error", err)
				rowErrs = append(rowErrs, err)
				continue
			}

			jobNames = append(jobNames, jobName)
			payloads = append(payloads, string(payloadJSON))
			tenantIDs = append(tenantIDs, tenantID.String)
			reclaimed++
		}
		if err := rows.Err(); err != nil {
			return err
		}
		if len(jobNames) > 0 {
			if _, err := tx.ExecContext(ctx, `
INSERT INTO governance_events
	(tenant_id, event_type, actor_user_id, resource_type, resource_id, reason, payload_json, occurred_at)
SELECT
	u.tenant_id::uuid,
	'lease.reaped',
	'system:reaper',
	'job_lease',
	u.job_name,
	'expired_lease_reclaimed',
	u.payload_json::jsonb,
	now()
FROM unnest($1::text[], $2::text[], $3::text[]) AS u(job_name, payload_json, tenant_id)
`, pq.Array(jobNames), pq.Array(payloads), pq.Array(tenantIDs)); err != nil {
				return err
			}
		}
		if err := tx.Commit(); err != nil {
			return err
		}

		slog.InfoContext(ctx, "lease_reaper: tick complete",
			"job", "lease_reaper",
			"epoch", epoch,
			"reclaimed", reclaimed)
		return errors.Join(rowErrs...)
	}
}
