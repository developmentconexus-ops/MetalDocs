package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"metaldocs/internal/platform/messaging"
)

type Consumer struct {
	db         *sql.DB
	claimLease time.Duration
}

func NewConsumer(db *sql.DB, claimLease time.Duration) *Consumer {
	if claimLease <= 0 {
		claimLease = 30 * time.Second
	}
	return &Consumer{db: db, claimLease: claimLease}
}

func (c *Consumer) ClaimUnpublished(ctx context.Context, limit int) ([]messaging.Event, error) {
	if limit <= 0 {
		limit = 20
	}

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin claim outbox tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

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
	rows, err := tx.QueryContext(ctx, q, limit, durationToPostgresInterval(c.claimLease))
	if err != nil {
		return nil, fmt.Errorf("claim unpublished outbox events: %w", err)
	}
	defer rows.Close()

	var events []messaging.Event
	for rows.Next() {
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
			return nil, fmt.Errorf("scan outbox event: %w", err)
		}
		event.EventID = messaging.EventID(eventID)
		event.EventType = messaging.EventType(eventType)
		event.AggregateType = messaging.AggregateType(aggregateType)
		event.AggregateID = messaging.AggregateID(aggregateID)
		event.IdempotencyKey = messaging.IdempotencyKey(idempotencyKey)
		event.TraceID = messaging.TraceID(traceID)
		event.OccurredAtRFC3339 = occurredAt.UTC().Format(time.RFC3339)
		if len(payloadJSON) > 0 {
			payload, err := messaging.DecodePayload(event.EventType, payloadJSON)
			if err != nil {
				return nil, fmt.Errorf("unmarshal outbox payload: %w", err)
			}
			event.Payload = payload
		} else {
			payload, err := messaging.DecodePayload(event.EventType, []byte("{}"))
			if err != nil {
				return nil, fmt.Errorf("unmarshal outbox payload: %w", err)
			}
			event.Payload = payload
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate outbox rows: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit claim outbox tx: %w", err)
	}
	return events, nil
}

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
