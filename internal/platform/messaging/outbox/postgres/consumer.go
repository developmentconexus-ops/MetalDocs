package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	platformdb "metaldocs/internal/platform/db"
	"metaldocs/internal/platform/messaging"
)

// Consumer implements messaging.Consumer against the Postgres-backed
// transactional outbox table.
type Consumer struct {
	db         *sql.DB
	runner     platformdb.TxRunner
	claimLease time.Duration
}

// NewConsumer constructs a Consumer backed by db. claimLease bounds how long
// a claimed event is held before it becomes eligible for reclaim; it defaults
// to 30 seconds when zero or negative.
func NewConsumer(db *sql.DB, claimLease time.Duration) *Consumer {
	if claimLease <= 0 {
		claimLease = 30 * time.Second
	}
	return &Consumer{db: db, runner: platformdb.NewTxRunner(db), claimLease: claimLease}
}

// ClaimUnpublished claims up to limit unpublished outbox events for
// processing, using SKIP LOCKED so concurrent consumers do not collide.
func (c *Consumer) ClaimUnpublished(ctx context.Context, limit int) ([]messaging.Event, error) {
	if limit <= 0 {
		limit = 20
	}

	const q = `
-- TODO(phase11): heartbeat_lease is mirrored here as claimLease; make the DB interval configurable in the follow-up migration/worker sweep.
-- TODO(phase11): this claim query depends on the current partial-index status predicate; revisit together with the pending DB index fix.
WITH candidates AS (
  SELECT event_id
  FROM metaldocs.outbox_events
  WHERE published_at IS NULL
    AND dead_lettered_at IS NULL
    AND (next_attempt_at IS NULL OR next_attempt_at <= NOW())
  ORDER BY occurred_at ASC
  FOR UPDATE SKIP LOCKED
  LIMIT $1
),
claimed AS (
  UPDATE metaldocs.outbox_events oe
  SET attempt_count = oe.attempt_count + 1,
      last_attempt_at = NOW(),
      next_attempt_at = NOW() + $2::interval
  FROM candidates c
  WHERE oe.event_id = c.event_id
  RETURNING oe.event_id, oe.event_type, oe.aggregate_type, oe.aggregate_id, oe.occurred_at,
            oe.version, oe.attempt_count, oe.idempotency_key, oe.producer, oe.trace_id, oe.payload
)
SELECT event_id, event_type, aggregate_type, aggregate_id, occurred_at, version,
       attempt_count, idempotency_key, producer, trace_id, payload
FROM claimed
ORDER BY occurred_at ASC
`
	var events []messaging.Event
	err := c.runner.Do(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, q, limit, durationToPostgresInterval(c.claimLease))
		if err != nil {
			return fmt.Errorf("query unpublished outbox events: %w", err)
		}
		defer func() { _ = rows.Close() }()

		for rows.Next() {
			event, err := scanUnpublishedEvent(rows)
			if err != nil {
				return err
			}
			events = append(events, event)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("iterate outbox rows: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("claim unpublished outbox events: %w", err)
	}
	return events, nil
}

func scanUnpublishedEvent(rows *sql.Rows) (messaging.Event, error) {
	var event messaging.Event
	var occurredAt time.Time
	var payloadJSON []byte
	var eventID string
	var eventType string
	var aggregateType string
	var aggregateID string
	var idempotencyKey string
	var traceID string
	if err := rows.Scan(
		&eventID,
		&eventType,
		&aggregateType,
		&aggregateID,
		&occurredAt,
		&event.Version,
		&event.AttemptCount,
		&idempotencyKey,
		&event.Producer,
		&traceID,
		&payloadJSON,
	); err != nil {
		return messaging.Event{}, fmt.Errorf("scan outbox event: %w", err)
	}
	event.EventID = messaging.EventID(eventID)
	event.EventType = messaging.EventType(eventType)
	event.AggregateType = messaging.AggregateType(aggregateType)
	event.AggregateID = messaging.AggregateID(aggregateID)
	event.IdempotencyKey = messaging.IdempotencyKey(idempotencyKey)
	event.TraceID = messaging.TraceID(traceID)
	event.OccurredAtRFC3339 = occurredAt.UTC().Format(time.RFC3339)

	payload, err := decodeUnpublishedPayload(event.EventType, payloadJSON)
	if err != nil {
		return messaging.Event{}, fmt.Errorf("unmarshal outbox payload: %w", err)
	}
	event.Payload = payload
	return event, nil
}

func decodeUnpublishedPayload(eventType messaging.EventType, payloadJSON []byte) (messaging.Payload, error) {
	if len(payloadJSON) == 0 {
		payloadJSON = []byte("{}")
	}
	return messaging.DecodePayload(eventType, payloadJSON)
}

// MarkPublished marks the given outbox events as published, clearing their
// retry state.
func (c *Consumer) MarkPublished(ctx context.Context, eventIDs []messaging.EventID) error {
	if len(eventIDs) == 0 {
		return nil
	}
	placeholders := make([]string, 0, len(eventIDs))
	args := make([]any, 0, len(eventIDs)+1)
	for idx, eventID := range eventIDs {
		placeholders = append(placeholders, fmt.Sprintf("$%d", idx+1))
		args = append(args, strings.TrimSpace(string(eventID)))
	}
	// #nosec G201 -- the only interpolated values are computed placeholder positions ($1,$2,...) and their comma-joined list; every actual event-id value is bound as a query arg via args, never interpolated into the SQL text.
	q := fmt.Sprintf(`
UPDATE metaldocs.outbox_events
SET published_at = $%d,
    next_attempt_at = NULL,
    last_error = NULL
WHERE event_id IN (%s)
`, len(args)+1, strings.Join(placeholders, ", "))
	args = append(args, time.Now().UTC())
	if _, err := c.db.ExecContext(ctx, q, args...); err != nil {
		return fmt.Errorf("mark outbox events published: %w", err)
	}
	return nil
}

// MarkFailed records a failed publish attempt, scheduling a retry or
// dead-lettering the event per failure's fields.
func (c *Consumer) MarkFailed(ctx context.Context, failure messaging.FailedEvent) error {
	if strings.TrimSpace(string(failure.EventID)) == "" {
		return fmt.Errorf("event id is required")
	}

	var nextAttempt any
	if failure.NextAttemptAt != nil {
		nextAttempt = failure.NextAttemptAt.UTC()
	}
	var deadLettered any
	if failure.DeadLetteredAt != nil {
		deadLettered = failure.DeadLetteredAt.UTC()
	}

	const q = `
UPDATE metaldocs.outbox_events
SET published_at = NULL,
    last_error = $2,
    next_attempt_at = $3,
    dead_lettered_at = $4
WHERE event_id = $1
`
	if _, err := c.db.ExecContext(ctx, q, strings.TrimSpace(string(failure.EventID)), strings.TrimSpace(failure.LastError), nextAttempt, deadLettered); err != nil {
		return fmt.Errorf("mark outbox event failed: %w", err)
	}
	return nil
}

func durationToPostgresInterval(value time.Duration) string {
	seconds := int(value.Seconds())
	if seconds < 1 {
		seconds = 1
	}
	return fmt.Sprintf("%d seconds", seconds)
}
