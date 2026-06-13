package fanout

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"metaldocs/internal/platform/db"
)

// OutboxRow is the common row shape for all staging dispatch outbox tables.
type OutboxRow struct {
	ID          string
	TenantID    string
	RevisionID  string
	ContentHash []byte
	Attempts    int
}

// stagingOutboxAllowlist restricts which tables the generic repo may address,
// preventing accidental or malicious fmt.Sprintf injection.
var stagingOutboxAllowlist = map[string]struct{}{
	"metaldocs.pdf_dispatch_outbox":          {},
	"metaldocs.materialize_dispatch_outbox":  {},
}

// StagingOutboxRepository is a generic transactional-outbox repo for any
// staging dispatch outbox table that matches the pdf/materialize schema.
// Construct with NewStagingOutboxRepository; the table name is validated
// against stagingOutboxAllowlist at construction time.
type StagingOutboxRepository struct {
	db    *sql.DB
	table string // fully-qualified, allowlist-validated
	name  string // short human name for error messages
}

// NewStagingOutboxRepository returns a repo bound to the given fully-qualified
// table name. Panics if the name is not in the allowlist (programmer error).
func NewStagingOutboxRepository(db *sql.DB, table string) *StagingOutboxRepository {
	if _, ok := stagingOutboxAllowlist[table]; !ok {
		panic(fmt.Sprintf("staging outbox: table %q not in allowlist", table))
	}
	return &StagingOutboxRepository{db: db, table: table, name: table}
}

// Enqueue inserts a pending outbox row inside the caller's business transaction.
// The outbox INSERT MUST share the caller's business transaction (atomic dispatch).
// A nil tx would silently autocommit the outbox row outside that transaction, breaking
// the transactional-outbox guarantee — fail loud (db.Tx contract: a nil Tx is never valid).
func (r *StagingOutboxRepository) Enqueue(ctx context.Context, tx db.Tx, tenantID, revisionID string, contentHash []byte) error {
	if tx == nil {
		return fmt.Errorf("%s enqueue: tx must not be nil", r.name)
	}
	//nolint:gosec // table name is allowlist-validated at construction
	_, err := tx.ExecContext(ctx, fmt.Sprintf(`
INSERT INTO %s (tenant_id, revision_id, content_hash)
VALUES ($1::uuid, $2::uuid, $3)
ON CONFLICT (tenant_id, revision_id) DO NOTHING`, r.table),
		tenantID, revisionID, contentHash)
	if err != nil {
		return fmt.Errorf("%s enqueue: %w", r.name, err)
	}
	return nil
}

// TODO(render): thread tenant scope through the worker claim path before adding
// a tenant_id predicate here; the current worker intentionally drains the shared
// outbox across tenants.
func (r *StagingOutboxRepository) ClaimPending(ctx context.Context, limit, maxAttempts int) ([]OutboxRow, error) {
	//nolint:gosec // table name is allowlist-validated at construction
	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
WITH claimed AS (
  SELECT id FROM %s
   WHERE status = 'pending' AND next_retry_at <= NOW() AND attempts < $2
   ORDER BY next_retry_at ASC
   LIMIT $1
   FOR UPDATE SKIP LOCKED
)
UPDATE %s o
   SET status='processing', claimed_at=NOW()
  FROM claimed c
 WHERE o.id = c.id
RETURNING o.id, o.tenant_id, o.revision_id, o.content_hash, o.attempts`, r.table, r.table),
		limit, maxAttempts)
	if err != nil {
		return nil, fmt.Errorf("%s claim pending: %w", r.name, err)
	}
	defer rows.Close()
	var out []OutboxRow
	for rows.Next() {
		var row OutboxRow
		if err := rows.Scan(&row.ID, &row.TenantID, &row.RevisionID, &row.ContentHash, &row.Attempts); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *StagingOutboxRepository) MarkDispatched(ctx context.Context, id string) error {
	//nolint:gosec // table name is allowlist-validated at construction
	res, err := r.db.ExecContext(ctx, fmt.Sprintf(`
UPDATE %s
   SET status='dispatched', dispatched_at=NOW()
 WHERE id=$1::uuid`, r.table), id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("%s mark dispatched: row not found: id=%s", r.name, id)
	}
	return nil
}

func (r *StagingOutboxRepository) MarkFailed(ctx context.Context, id string, errStr string, nextRetryAt time.Time, finalize bool) error {
	if finalize {
		//nolint:gosec // table name is allowlist-validated at construction
		res, err := r.db.ExecContext(ctx, fmt.Sprintf(`
UPDATE %s
   SET status='failed', last_error=$2, attempts=attempts+1
 WHERE id=$1::uuid`, r.table), id, errStr)
		if err != nil {
			return err
		}
		n, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if n == 0 {
			return fmt.Errorf("%s mark failed finalize: row not found: id=%s", r.name, id)
		}
		return nil
	}
	//nolint:gosec // table name is allowlist-validated at construction
	res, err := r.db.ExecContext(ctx, fmt.Sprintf(`
UPDATE %s
   SET status='pending', last_error=$2, attempts=attempts+1, next_retry_at=$3, claimed_at=NULL
 WHERE id=$1::uuid`, r.table), id, errStr, nextRetryAt)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("%s mark failed retry: row not found: id=%s", r.name, id)
	}
	return nil
}

// ReadState returns the latest status for the given tenant+revision.
// Returns ("", nil) when no row exists.
func (r *StagingOutboxRepository) ReadState(ctx context.Context, tenantID, revisionID string) (string, error) {
	var status string
	//nolint:gosec // table name is allowlist-validated at construction
	err := r.db.QueryRowContext(ctx, fmt.Sprintf(`
SELECT status FROM %s
 WHERE tenant_id=$1::uuid AND revision_id=$2::uuid
 ORDER BY created_at DESC LIMIT 1`, r.table),
		tenantID, revisionID,
	).Scan(&status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("%s read state: %w", r.name, err)
	}
	return status, nil
}

func (r *StagingOutboxRepository) ResetStaleClaims(ctx context.Context, olderThan time.Duration) (int, error) {
	//nolint:gosec // table name is allowlist-validated at construction
	res, err := r.db.ExecContext(ctx, fmt.Sprintf(`
UPDATE %s
   SET status='pending', claimed_at=NULL
 WHERE status='processing' AND claimed_at < NOW() - ($1 * interval '1 millisecond')`, r.table),
		olderThan.Milliseconds())
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

// NewPDFOutboxRepository returns a StagingOutboxRepository bound to the PDF dispatch table.
func NewPDFOutboxRepository(db *sql.DB) *StagingOutboxRepository {
	return NewStagingOutboxRepository(db, "metaldocs.pdf_dispatch_outbox")
}

// NewMaterializeOutboxRepository returns a StagingOutboxRepository bound to the materialize dispatch table.
func NewMaterializeOutboxRepository(db *sql.DB) *StagingOutboxRepository {
	return NewStagingOutboxRepository(db, "metaldocs.materialize_dispatch_outbox")
}
