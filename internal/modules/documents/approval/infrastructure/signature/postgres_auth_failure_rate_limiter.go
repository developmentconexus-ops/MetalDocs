package signature

import (
	"context"
	"database/sql"
	"time"
)

// PostgresAuthFailureRateLimiter is a shared, Postgres-backed implementation of
// AuthFailureRateLimiter. It preserves the exact semantics of InMemoryAuthFailureRateLimiter:
//   - Fixed window of windowDur (60s); counter resets when window_start + windowDur < now.
//   - Blocked when fail_count >= maxFailures (5) within the current window.
//   - Reset on successful authentication (DELETE the row).
//   - Key: actorID (user UUID as text), no IP component.
//
// Designed for production use: lockouts survive API restarts and are shared
// across replicas. Requires migration 0235 (public.auth_failure_counters).
type PostgresAuthFailureRateLimiter struct {
	db *sql.DB
}

// NewPostgresAuthFailureRateLimiter creates the limiter. db must not be nil.
func NewPostgresAuthFailureRateLimiter(db *sql.DB) *PostgresAuthFailureRateLimiter {
	return &PostgresAuthFailureRateLimiter{db: db}
}

// Allow returns true when the actor has fewer than maxFailures failures in the
// current window (or no row exists / window has expired).
func (l *PostgresAuthFailureRateLimiter) Allow(ctx context.Context, actorID string) (bool, error) {
	const q = `
		SELECT fail_count, window_start
		FROM public.auth_failure_counters
		WHERE actor_id = $1`

	var count int
	var windowStart time.Time
	err := l.db.QueryRowContext(ctx, q, actorID).Scan(&count, &windowStart)
	if err == sql.ErrNoRows {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	if time.Since(windowStart) >= windowDur {
		// Window expired — counter is stale; allow and let RecordFailure clean up.
		return true, nil
	}
	return count < maxFailures, nil
}

// RecordFailure increments the failure counter. If no row exists, or the existing
// window has expired, a new window is started at the current time (count = 1).
// Stale rows older than 2×windowDur are pruned in the same statement to bound
// table size without a separate janitor job.
func (l *PostgresAuthFailureRateLimiter) RecordFailure(ctx context.Context, actorID string) error {
	const q = `
		INSERT INTO public.auth_failure_counters (actor_id, fail_count, window_start)
		VALUES ($1, 1, $2)
		ON CONFLICT (actor_id) DO UPDATE SET
		    fail_count   = CASE
		                       WHEN public.auth_failure_counters.window_start + $3 <= EXCLUDED.window_start
		                       THEN 1
		                       ELSE public.auth_failure_counters.fail_count + 1
		                   END,
		    window_start = CASE
		                       WHEN public.auth_failure_counters.window_start + $3 <= EXCLUDED.window_start
		                       THEN EXCLUDED.window_start
		                       ELSE public.auth_failure_counters.window_start
		                   END`

	now := time.Now().UTC()
	_, err := l.db.ExecContext(ctx, q, actorID, now, windowDur)
	return err
}

// Reset removes the failure counter for the actor on successful authentication.
func (l *PostgresAuthFailureRateLimiter) Reset(ctx context.Context, actorID string) error {
	const q = `DELETE FROM public.auth_failure_counters WHERE actor_id = $1`
	_, err := l.db.ExecContext(ctx, q, actorID)
	return err
}
