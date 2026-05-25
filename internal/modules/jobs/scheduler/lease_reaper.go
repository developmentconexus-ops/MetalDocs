package scheduler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/lib/pq"

	"metaldocs/internal/platform/tenant"
)

func RunLeaseReaper(db *sql.DB) JobFunc {
	return func(ctx context.Context, epoch int64) error {
		tenantID, err := tenant.FromContext(ctx)
		if err != nil {
			tenantID = tenant.DevTenantID
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		rows, err := tx.QueryContext(ctx, `
DELETE FROM metaldocs.job_leases
WHERE job_name IN (
	SELECT job_name FROM metaldocs.job_leases
	WHERE expires_at < now() - interval '10 minutes'
	FOR UPDATE SKIP LOCKED
)
RETURNING job_name, leader_id, lease_epoch
`)
		if err != nil {
			return err
		}
		defer rows.Close()

		reclaimed := 0
		jobNames := make([]string, 0)
		payloads := make([]string, 0)
		var rowErrs []error
		for rows.Next() {
			var jobName string
			var leaderID string
			var leaseEpoch int64
			if err := rows.Scan(&jobName, &leaderID, &leaseEpoch); err != nil {
				rowErrs = append(rowErrs, err)
				continue
			}

			payloadJSON, err := json.Marshal(map[string]any{
				"job_name":    jobName,
				"leader_id":   leaderID,
				"lease_epoch": leaseEpoch,
			})
			if err != nil {
				rowErrs = append(rowErrs, err)
				continue
			}

			jobNames = append(jobNames, jobName)
			payloads = append(payloads, string(payloadJSON))
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
	COALESCE(
		(SELECT d.tenant_id FROM public.documents d WHERE d.id::text = u.job_name LIMIT 1),
		$3::uuid
	),
	'lease.reaped',
	'system:reaper',
	'job_lease',
	u.job_name,
	'expired_lease_reclaimed',
	u.payload_json::jsonb,
	now()
FROM unnest($1::text[], $2::text[]) AS u(job_name, payload_json)
`, pq.Array(jobNames), pq.Array(payloads), tenantID); err != nil {
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
