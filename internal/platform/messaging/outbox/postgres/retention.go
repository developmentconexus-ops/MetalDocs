package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Retention owns the terminal-row purge for metaldocs.outbox_events.
//
// It lives in this package — beside publisher.go and consumer.go — because
// this package is the sole owner of SQL against metaldocs.outbox_events (a
// platform table, owner platform/workers). The River job that ticks the purge
// lives in internal/modules/jobs/outbox_retention; it holds no SQL.
//
// WHY a purge exists at all: nothing ever deleted from outbox_events. The
// table grew forever, and — because Publisher.Publish dedupes on
// UNIQUE (idempotency_key) with ON CONFLICT DO NOTHING — every terminal row
// also pinned its idempotency key forever, so a later legitimate re-publish on
// the same key was silently swallowed. Retention bounds both, but it does NOT
// fix the swallow itself: a re-publish INSIDE the retention window is still
// dropped and still returns nil. That is a separate publisher defect (see the
// package docs in wiki/backend/platform/async-messaging.md §2.2).
type Retention struct {
	db *sql.DB
}

// NewRetention constructs a Retention bound to db.
func NewRetention(db *sql.DB) *Retention {
	return &Retention{db: db}
}

// PurgePublished deletes rows that were successfully published before cutoff,
// in bounded batches. Dead-lettered rows are excluded here even when they also
// carry a published_at (a replayed dead-letter): they age out on the separate,
// much longer dead-letter window instead, via PurgeDeadLettered.
func (r *Retention) PurgePublished(ctx context.Context, cutoff time.Time, batchSize, maxIterations int) (int, error) {
	// published_at IS NOT NULL is redundant with `published_at < $1` under SQL
	// three-valued logic (NULL < x yields NULL, which never matches). It is
	// written out anyway so the fail-closed invariant — an unpublished,
	// non-dead-lettered row is NEVER eligible — is visible in the statement
	// itself and survives a future edit to the cutoff comparison.
	const q = `
DELETE FROM metaldocs.outbox_events
WHERE ctid IN (
	SELECT ctid FROM metaldocs.outbox_events
	WHERE published_at IS NOT NULL
	  AND published_at < $1
	  AND dead_lettered_at IS NULL
	LIMIT $2
)`
	return r.purgeBatched(ctx, "published", q, cutoff, batchSize, maxIterations)
}

// PurgeDeadLettered deletes rows that were dead-lettered before cutoff, in
// bounded batches. Callers pass a deliberately long cutoff: a dead-lettered
// row is the only durable record that a domain event never reached its
// consumer, so it is forensic evidence first and a retention liability second.
func (r *Retention) PurgeDeadLettered(ctx context.Context, cutoff time.Time, batchSize, maxIterations int) (int, error) {
	// dead_lettered_at IS NOT NULL is likewise redundant-but-explicit; see the
	// comment in PurgePublished.
	const q = `
DELETE FROM metaldocs.outbox_events
WHERE ctid IN (
	SELECT ctid FROM metaldocs.outbox_events
	WHERE dead_lettered_at IS NOT NULL
	  AND dead_lettered_at < $1
	LIMIT $2
)`
	return r.purgeBatched(ctx, "dead lettered", q, cutoff, batchSize, maxIterations)
}

// purgeBatched runs the given single-cutoff delete in a bounded loop, stopping
// early once a batch affects 0 rows (nothing left to delete) or maxIterations
// is reached — the same ctid-subquery-with-LIMIT idiom used by
// internal/modules/jobs/idempotency_janitor/job.go:56-78 and
// internal/modules/render/fanout.StagingOutboxRepository.PurgeDispatched.
//
// Non-positive bounds are an error, not a silent no-op and not an unbounded
// delete: a caller that passes 0 has a bug, and neither degradation is safe to
// pick on its behalf.
func (r *Retention) purgeBatched(ctx context.Context, class, query string, cutoff time.Time, batchSize, maxIterations int) (int, error) {
	if batchSize <= 0 {
		return 0, fmt.Errorf("purge %s outbox events: batch size must be positive, got %d", class, batchSize)
	}
	if maxIterations <= 0 {
		return 0, fmt.Errorf("purge %s outbox events: max iterations must be positive, got %d", class, maxIterations)
	}

	totalDeleted := 0
	for i := 0; i < maxIterations; i++ {
		result, err := r.db.ExecContext(ctx, query, cutoff.UTC(), batchSize)
		if err != nil {
			return totalDeleted, fmt.Errorf("purge %s outbox events: %w", class, err)
		}

		n, err := result.RowsAffected()
		if err != nil {
			return totalDeleted, fmt.Errorf("purge %s outbox events: rows affected: %w", class, err)
		}
		totalDeleted += int(n)
		if n == 0 {
			break
		}
	}
	return totalDeleted, nil
}
