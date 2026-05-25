package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"metaldocs/internal/modules/audit/domain"
)

type Writer struct {
	db *sql.DB
}

const auditHashChainLockID int64 = 90120260513004

func NewWriter(db *sql.DB) *Writer {
	return &Writer{db: db}
}

func (w *Writer) Record(ctx context.Context, event domain.Event) error {
	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin audit hash-chain tx: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			slog.Error("audit rollback failed", "error", rollbackErr)
		}
	}()

	if err := w.RecordTx(ctx, tx, event); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit audit hash-chain tx: %w", err)
	}
	return nil
}

func (w *Writer) RecordTx(ctx context.Context, tx *sql.Tx, event domain.Event) error {
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, auditHashChainLockID); err != nil {
		return fmt.Errorf("lock audit hash chain: %w", err)
	}

	const q = `
WITH previous AS (
  SELECT row_hash
  FROM metaldocs.audit_events
  WHERE row_hash <> ''
  ORDER BY audit_sequence DESC
  LIMIT 1
),
prepared AS (
  SELECT COALESCE((SELECT row_hash FROM previous), '') AS prev_hash,
         $7::jsonb AS payload_json
)
INSERT INTO metaldocs.audit_events (
  id, occurred_at, actor_id, action, resource_type, resource_id, payload, trace_id, tenant_id, prev_hash, row_hash
)
SELECT $1, $2, $3, $4, $5, $6, payload_json, $8, $9, prev_hash,
       metaldocs.audit_event_row_hash(prev_hash, $1, $2, $3, $4, $5, $6, payload_json, $8, $9)
FROM prepared
`
	if _, err := tx.ExecContext(ctx, q,
		event.ID, event.OccurredAt, event.ActorID, event.Action,
		event.ResourceType, event.ResourceID, event.PayloadJSON, event.TraceID, event.TenantID,
	); err != nil {
		return fmt.Errorf("insert audit event (tx): %w", err)
	}
	return nil
}

func (w *Writer) ValidateIntegrity(ctx context.Context) ([]domain.IntegrityIssue, error) {
	const q = `
WITH ordered AS (
  SELECT audit_sequence, id, occurred_at, actor_id, action, resource_type, resource_id,
         payload, trace_id, tenant_id, prev_hash, row_hash,
         ROW_NUMBER() OVER (ORDER BY audit_sequence) AS rn,
         LAG(row_hash, 1, '') OVER (ORDER BY audit_sequence) AS previous_row_hash
  FROM metaldocs.audit_events
)
SELECT audit_sequence, id, prev_hash, row_hash, expected_prev_hash,
       metaldocs.audit_event_row_hash(prev_hash, id, occurred_at, actor_id, action, resource_type, resource_id, payload, trace_id, tenant_id) AS expected_row_hash
FROM (
  SELECT *, CASE WHEN rn = 1 THEN prev_hash ELSE previous_row_hash END AS expected_prev_hash
  FROM ordered
) checked
ORDER BY audit_sequence
`
	rows, err := w.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("validate audit integrity: %w", err)
	}
	defer rows.Close()

	var issues []domain.IntegrityIssue
	for rows.Next() {
		var sequence int64
		var id, prevHash, rowHash, expectedPrevHash, expectedRowHash string
		if err := rows.Scan(&sequence, &id, &prevHash, &rowHash, &expectedPrevHash, &expectedRowHash); err != nil {
			return nil, fmt.Errorf("scan audit integrity row: %w", err)
		}
		if prevHash != expectedPrevHash {
			issues = append(issues, domain.IntegrityIssue{
				Sequence:         sequence,
				EventID:          id,
				Kind:             domain.IntegrityIssuePrevHashMismatch,
				ExpectedPrevHash: expectedPrevHash,
				ActualPrevHash:   prevHash,
			})
		}
		if rowHash != expectedRowHash {
			issues = append(issues, domain.IntegrityIssue{
				Sequence:     sequence,
				EventID:      id,
				Kind:         domain.IntegrityIssueRowHashMismatch,
				ExpectedHash: expectedRowHash,
				ActualHash:   rowHash,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate audit integrity rows: %w", err)
	}
	return issues, nil
}

func (w *Writer) ListEvents(ctx context.Context, query domain.ListEventsQuery) ([]domain.Event, error) {
	limit := query.Limit
	if limit <= 0 {
		limit = 50
	}

	const q = `
SELECT id, occurred_at, actor_id, action, resource_type, resource_id, payload::text, trace_id, tenant_id
FROM metaldocs.audit_events
WHERE ($1 = '' OR resource_type = $1)
  AND ($2 = '' OR resource_id = $2)
  AND ($3 = '' OR tenant_id = $3)
ORDER BY occurred_at DESC, id DESC
LIMIT $4
`

	rows, err := w.db.QueryContext(ctx, q,
		strings.TrimSpace(query.ResourceType),
		strings.TrimSpace(query.ResourceID),
		strings.TrimSpace(query.TenantID),
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list audit events: %w", err)
	}
	defer rows.Close()

	items := make([]domain.Event, 0, limit)
	for rows.Next() {
		var event domain.Event
		if err := rows.Scan(
			&event.ID,
			&event.OccurredAt,
			&event.ActorID,
			&event.Action,
			&event.ResourceType,
			&event.ResourceID,
			&event.PayloadJSON,
			&event.TraceID,
			&event.TenantID,
		); err != nil {
			return nil, fmt.Errorf("scan audit event: %w", err)
		}
		items = append(items, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate audit events: %w", err)
	}
	return items, nil
}
