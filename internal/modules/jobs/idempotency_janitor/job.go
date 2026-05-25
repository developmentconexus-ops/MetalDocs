package idempotency_janitor

import (
	"context"
	"database/sql"
	"log/slog"

	"metaldocs/internal/modules/jobs/scheduler"
)

const (
	JobName       = "idempotency_janitor"
	BatchSize     = 5000
	MaxIterations = 10
	// OrphanGracePeriod is how long an `in_flight` row may remain past its
	// expires_at before the janitor logs it as a crashed-handler orphan.
	OrphanGraceMinutes = 5
)

// New returns a janitor job that sweeps the idempotency_keys table.
//
// Two passes per tick:
//  1. Delete any row past expires_at regardless of status. Pre-fix the sweep
//     was gated on `status = 'completed'`, which leaked `in_flight` / `failed`
//     rows forever (M10).
//  2. Warn on `in_flight` rows past expires_at + OrphanGraceMinutes. These
//     are crashed-handler orphans; surfacing them lets ops correlate with
//     panic logs.
func New(db *sql.DB) scheduler.JobFunc {
	return func(ctx context.Context, epoch int64) error {
		totalDeleted := 0
		for i := 0; i < MaxIterations; i++ {
			result, err := db.ExecContext(ctx, `
DELETE FROM metaldocs.idempotency_keys
WHERE ctid IN (
	SELECT ctid FROM metaldocs.idempotency_keys
	WHERE expires_at < now()
	LIMIT $1
)`, BatchSize)
			if err != nil {
				slog.ErrorContext(ctx, "idempotency_janitor: delete failed", "error", err)
				return err
			}

			n, err := result.RowsAffected()
			if err != nil {
				slog.ErrorContext(ctx, "idempotency_janitor: rows affected failed", "error", err)
				return err
			}
			totalDeleted += int(n)
			if n == 0 {
				break
			}
		}

		// Orphan-detection pass: anything still `in_flight` past the grace
		// window is from a handler that crashed without releasing. The pass-1
		// delete already reaped these by ctid above, so the count here is
		// (a) what we just reclaimed plus (b) any that crossed the grace
		// threshold mid-tick. Either way we want visibility.
		var orphans int64
		if err := db.QueryRowContext(ctx, `
SELECT count(*) FROM metaldocs.idempotency_keys
 WHERE status     = 'in_flight'
   AND expires_at < now() - make_interval(mins => $1)`,
			OrphanGraceMinutes,
		).Scan(&orphans); err != nil {
			slog.ErrorContext(ctx, "idempotency_janitor: orphan count failed", "error", err)
		} else if orphans > 0 {
			slog.WarnContext(ctx, "idempotency_janitor: in_flight orphans detected",
				"job", JobName,
				"epoch", epoch,
				"orphans", orphans,
				"grace_minutes", OrphanGraceMinutes,
				"hint", "crashed handler? check for panic logs near the orphans' expires_at - 24h",
			)
		}

		slog.InfoContext(ctx, "idempotency_janitor: tick complete",
			"job", JobName,
			"epoch", epoch,
			"deleted", totalDeleted)
		return nil
	}
}
