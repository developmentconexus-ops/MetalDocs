package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"metaldocs/internal/modules/audit/domain"
)

type Writer struct {
	db *sql.DB
}

const auditHashChainLockID int64 = 90120260513004
const auditIntegrityValidationWindow = 10000
const auditIntegrityIssueLimit = 256

func NewWriter(db *sql.DB) *Writer {
	return &Writer{db: db}
}

func (w *Writer) Record(ctx context.Context, event domain.Event) error {
	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin audit hash-chain tx: %w", err)
	}
	// Best-effort rollback for the uncommitted path; Commit makes sql.ErrTxDone expected.
	defer func() { _ = tx.Rollback() }()

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
WITH recent AS (
  SELECT audit_sequence, id, occurred_at, actor_id, action, resource_type, resource_id,
         payload, trace_id, tenant_id, prev_hash, row_hash
  FROM metaldocs.audit_events
  ORDER BY audit_sequence DESC
  LIMIT $1
),
ordered AS (
  SELECT audit_sequence, id, occurred_at, actor_id, action, resource_type, resource_id,
         payload, trace_id, tenant_id, prev_hash, row_hash,
         ROW_NUMBER() OVER (ORDER BY audit_sequence) AS rn,
         LAG(row_hash, 1, '') OVER (ORDER BY audit_sequence) AS previous_row_hash
  FROM recent
)
SELECT audit_sequence, id, prev_hash, row_hash, expected_prev_hash,
       metaldocs.audit_event_row_hash(prev_hash, id, occurred_at, actor_id, action, resource_type, resource_id, payload, trace_id, tenant_id) AS expected_row_hash
FROM (
  SELECT *, CASE WHEN rn = 1 THEN prev_hash ELSE previous_row_hash END AS expected_prev_hash
  FROM ordered
) checked
ORDER BY audit_sequence
`
	rows, err := w.db.QueryContext(ctx, q, auditIntegrityValidationWindow)
	if err != nil {
		return nil, fmt.Errorf("validate audit integrity: %w", err)
	}
	defer rows.Close()

	issues := make([]domain.IntegrityIssue, 0, 4)
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
		if len(issues) >= auditIntegrityIssueLimit {
			break
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

	sqlText, args := buildListQuery(query, limit)
	rows, err := w.db.QueryContext(ctx, sqlText, args...)
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

// CountEvents counts rows matching the same filter (cursor / limit ignored).
func (w *Writer) CountEvents(ctx context.Context, query domain.ListEventsQuery) (int64, error) {
	sqlText, args := buildCountQuery(query)
	var n int64
	if err := w.db.QueryRowContext(ctx, sqlText, args...).Scan(&n); err != nil {
		return 0, fmt.Errorf("count audit events: %w", err)
	}
	return n, nil
}

func buildListQuery(q domain.ListEventsQuery, limit int) (string, []any) {
	where, args := buildWhere(q)
	if !q.Cursor.IsZero() {
		args = append(args, q.Cursor.OccurredAt, q.Cursor.ID)
		where = append(where, fmt.Sprintf("(occurred_at, id) < ($%d, $%d)", len(args)-1, len(args)))
	}
	args = append(args, limit)
	limitPos := len(args)

	var b strings.Builder
	b.WriteString(`SELECT id, occurred_at, actor_id, action, resource_type, resource_id, payload::text, trace_id, tenant_id
FROM metaldocs.audit_events
WHERE `)
	b.WriteString(strings.Join(where, " AND "))
	b.WriteString(`
ORDER BY occurred_at DESC, id DESC
LIMIT $`)
	b.WriteString(strconv.Itoa(limitPos))
	return b.String(), args
}

func buildCountQuery(q domain.ListEventsQuery) (string, []any) {
	where, args := buildWhere(q)
	var b strings.Builder
	b.WriteString(`SELECT COUNT(*) FROM metaldocs.audit_events WHERE `)
	b.WriteString(strings.Join(where, " AND "))
	return b.String(), args
}

func buildWhere(q domain.ListEventsQuery) ([]string, []any) {
	where := []string{}
	args := []any{}

	add := func(clause string, value any) {
		args = append(args, value)
		where = append(where, fmt.Sprintf(clause, len(args)))
	}

	add("tenant_id = $%d", strings.TrimSpace(q.TenantID))
	if v := strings.TrimSpace(q.ResourceType); v != "" {
		add("resource_type = $%d", v)
	}
	if v := strings.TrimSpace(q.ResourceID); v != "" {
		add("resource_id = $%d", v)
	}
	if v := strings.TrimSpace(q.ActorID); v != "" {
		add("actor_id = $%d", v)
	}
	if v := strings.TrimSpace(q.Action); v != "" {
		if strings.HasSuffix(v, "*") {
			add("action LIKE $%d", strings.TrimSuffix(v, "*")+"%")
		} else {
			add("action = $%d", v)
		}
	}
	if !q.OccurredAfter.IsZero() {
		add("occurred_at >= $%d", q.OccurredAfter)
	}
	if !q.OccurredBefore.IsZero() {
		add("occurred_at < $%d", q.OccurredBefore)
	}
	if needle := strings.TrimSpace(q.Query); needle != "" {
		pattern := "%" + needle + "%"
		args = append(args, pattern)
		idx := len(args)
		where = append(where, fmt.Sprintf(
			"(action ILIKE $%d OR actor_id ILIKE $%d OR resource_id ILIKE $%d OR payload::text ILIKE $%d)",
			idx, idx, idx, idx,
		))
	}
	return where, args
}
