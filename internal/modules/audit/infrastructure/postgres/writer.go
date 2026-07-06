package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"metaldocs/internal/modules/audit/domain"
	platcrypto "metaldocs/internal/platform/crypto"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/pagination"
	"metaldocs/internal/platform/sqlescape"
)

// PayloadCrypto is the audit module's own narrow port for per-tenant payload
// envelope encryption (M7 F7.3 item 2). It is declared audit-side so this
// package never imports the security module — the composition root wires an
// adapter over security's published TenantCrypto port (ADR 0070 decision 4).
//
// EncryptForTenant returns (envelope, encrypted=true, nil) when tenantID has
// an active DEK. When the tenant has no key at all (never onboarded with
// crypto, or crypto is globally disabled) or its key has been crypto-shredded
// (ErrKeyDestroyed — a tombstone tenant), the adapter must map that to
// (_, false, nil) so RecordTx falls through to legacy plaintext storage
// instead of surfacing an error — a destroyed-key tenant still needs its
// lifecycle tombstone events to land (spec item 6.5). Only a genuine,
// unexpected failure (DB error, malformed KEK, etc.) should propagate as a
// non-nil error, which aborts the write.
//
// DecryptForTenant mirrors the read side: (plaintext, nil) on success. The
// adapter maps ErrKeyDestroyed to a sentinel the reader recognises so it can
// substitute the redacted marker instead of failing the whole list/export
// call; other errors propagate.
//
// EncryptForTenantTx is the tx-aware variant RecordTx uses when it is
// running inside a *sql.Tx (always, in practice — RecordTx's tx parameter is
// db.Tx, and *sql.Tx satisfies it at runtime; see sealPayload's type
// assertion). It exists to close a same-transaction key-visibility gap: a
// caller that provisions a tenant's crypto key and writes an audit event for
// that same tenant in the SAME still-open transaction (e.g.
// OnboardTenantService's tenant.onboarded event, written in the same tx as
// ProvisionTenantKeyTx) must have its key lookup read through that tx, not
// the pool — a pool read cannot see the uncommitted insert and always
// misses, which previously caused the payload to silently fall through to
// plaintext. Same fall-through contract as EncryptForTenant otherwise.
type PayloadCrypto interface {
	EncryptForTenant(ctx context.Context, tenantID string, plaintext []byte) (envelope string, encrypted bool, err error)
	EncryptForTenantTx(ctx context.Context, tx *sql.Tx, tenantID string, plaintext []byte) (envelope string, encrypted bool, err error)
	DecryptForTenant(ctx context.Context, tenantID string, envelope string) (plaintext []byte, err error)
}

// redactedPayload is what ListEvents substitutes for an envelope payload it
// cannot decrypt because the tenant's key has been crypto-shredded (erasure).
// Plaintext legacy rows and successfully-decrypted rows are never touched.
const redactedPayload = `{"redacted":"crypto-shredded"}`

// Writer is the postgres-backed implementation of domain.Writer/Reader/
// Counter/IntegrityValidator against metaldocs.audit_events. Every insert
// takes the audit-hash-chain advisory lock so the prev_hash/row_hash chain is
// computed and appended without a race between concurrent writers.
type Writer struct {
	db     *sql.DB
	crypto PayloadCrypto
}

const auditHashChainLockID int64 = 90120260513004
const auditIntegrityValidationWindow = 10000
const auditIntegrityIssueLimit = 256

// NewWriter constructs a Writer backed by db. The payload is stored plaintext
// (today's legacy behavior) unless WithPayloadCrypto is used to wire an
// encryptor.
func NewWriter(db *sql.DB) *Writer {
	return &Writer{db: db}
}

// WithPayloadCrypto wires crypto into the Writer, returning it for chaining.
// When crypto is non-nil, RecordTx seals new event payloads under the
// event's tenant DEK (when one exists) and ListEvents decrypts envelopes for
// display. Composition root only — audit itself never constructs a crypto
// implementation.
func (w *Writer) WithPayloadCrypto(crypto PayloadCrypto) *Writer {
	w.crypto = crypto
	return w
}

// Record appends event in its own new transaction (begin, RecordTx, commit).
// Use this for the standard post-commit case, where the caller's regulated
// mutation has already committed and this is an independent, separate write.
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

// RecordTx appends event inside the caller's own transaction tx, so the audit
// insert commits or rolls back atomically with the caller's regulated
// mutation. It takes a transaction-scoped advisory lock (pg_advisory_xact_lock)
// on the hash chain before computing prev_hash/row_hash, serializing
// concurrent writers for the duration of tx.
func (w *Writer) RecordTx(ctx context.Context, tx db.Tx, event domain.Event) error {
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, auditHashChainLockID); err != nil {
		return fmt.Errorf("lock audit hash chain: %w", err)
	}

	payloadJSON, err := w.sealPayload(ctx, tx, event)
	if err != nil {
		return fmt.Errorf("seal audit payload: %w", err)
	}
	event.PayloadJSON = payloadJSON

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

// sealPayload returns the string to store in the payload column: the
// envelope JSON when w.crypto is wired and event.TenantID has an active DEK,
// otherwise event.PayloadJSON unchanged (today's legacy plaintext
// behavior — byte-identical when crypto is nil). The hash chain hashes
// whatever this function returns, so encryption is transparent to
// tamper-evidence: chain semantics are untouched either way.
//
// tx is type-asserted to *sql.Tx so the key lookup can go through the SAME
// transaction as this INSERT: RecordTx always receives a live *sql.Tx in
// practice (db.Tx is satisfied only by *sql.Tx at runtime), so this closes
// the same-tx key-visibility gap (a tenant_keys row inserted earlier in this
// same tx, not yet committed, is invisible to a pool read). If the assertion
// ever fails (a future non-*sql.Tx db.Tx implementation), this falls back to
// EncryptForTenant's pool-backed read rather than erroring.
func (w *Writer) sealPayload(ctx context.Context, tx db.Tx, event domain.Event) (string, error) {
	if w.crypto == nil {
		return event.PayloadJSON, nil
	}
	var envelope string
	var encrypted bool
	var err error
	if sqlTx, ok := tx.(*sql.Tx); ok && sqlTx != nil {
		envelope, encrypted, err = w.crypto.EncryptForTenantTx(ctx, sqlTx, event.TenantID, []byte(event.PayloadJSON))
	} else {
		envelope, encrypted, err = w.crypto.EncryptForTenant(ctx, event.TenantID, []byte(event.PayloadJSON))
	}
	if err != nil {
		return "", err
	}
	if !encrypted {
		// No key (never provisioned / crypto disabled) or key destroyed
		// (tombstone tenant, spec item 6.5) — fall through to plaintext.
		return event.PayloadJSON, nil
	}
	return envelope, nil
}

// openPayload reverses sealPayload for display: when payloadJSON is a
// recognised crypto envelope, it is decrypted under tenantID's DEK. A
// decrypt failure caused by a destroyed key (crypto-shredded tenant) yields
// the redacted marker instead of propagating an error — the intended,
// permanent effect of erasure, not a bug. Plaintext rows (IsEnvelope false)
// and the nil-crypto composition-root path pass through untouched.
func (w *Writer) openPayload(ctx context.Context, tenantID, payloadJSON string) string {
	if w.crypto == nil || !platcrypto.IsEnvelope(payloadJSON) {
		return payloadJSON
	}
	plaintext, err := w.crypto.DecryptForTenant(ctx, tenantID, payloadJSON)
	if err != nil {
		return redactedPayload
	}
	return string(plaintext)
}

// ValidateIntegrity re-derives prev_hash/row_hash for the most recent
// auditIntegrityValidationWindow rows and reports any mismatch as an
// IntegrityIssue, capped at auditIntegrityIssueLimit. Used by the
// audit-integrity-validator janitor to detect tampering with audit_events.
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

// ListEvents queries metaldocs.audit_events for query's filter, ordered
// newest-first with keyset pagination on (occurred_at, id). Fetches one extra
// row past the page (limit+1 probe) to compute hasMore precisely, trimming
// the probe before returning.
func (w *Writer) ListEvents(ctx context.Context, query domain.ListEventsQuery) ([]domain.Event, bool, error) {
	limit := query.Limit
	if limit <= 0 {
		limit = pagination.DefaultLimit
	}

	// +1 probe row to detect hasMore: a trailing row beyond the page means a
	// further page exists. Trimmed below so the caller never sees the probe.
	sqlText, args := buildListQuery(query, limit+1)
	rows, err := w.db.QueryContext(ctx, sqlText, args...)
	if err != nil {
		return nil, false, fmt.Errorf("list audit events: %w", err)
	}
	defer rows.Close()

	items := make([]domain.Event, 0, limit+1)
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
			return nil, false, fmt.Errorf("scan audit event: %w", err)
		}
		event.PayloadJSON = w.openPayload(ctx, event.TenantID, event.PayloadJSON)
		items = append(items, event)
	}
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("iterate audit events: %w", err)
	}
	if len(items) > limit {
		return items[:limit], true, nil
	}
	return items, false, nil
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
			escaped := sqlescape.LikeEscape(strings.TrimSuffix(v, "*"))
			add(`action LIKE $%d ESCAPE '\'`, escaped+"%")
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
		pattern := "%" + sqlescape.LikeEscape(needle) + "%"
		args = append(args, pattern)
		idx := len(args)
		// Restrict free-text search to the indexed/structured columns only.
		// payload::text ILIKE was dropped (F-10): it forced a full-scan of the
		// JSONB column with no index support, making every search O(rows).
		// The OpenAPI description "action / payload summary" is satisfied by
		// searching action, actor_id, and resource_id — the columns users
		// realistically filter on. If payload-content search is ever required,
		// add a GIN index on the payload column and a separate query parameter.
		where = append(where, fmt.Sprintf(
			`(action ILIKE $%d ESCAPE '\' OR actor_id ILIKE $%d ESCAPE '\' OR resource_id ILIKE $%d ESCAPE '\')`,
			idx, idx, idx,
		))
	}
	return where, args
}
