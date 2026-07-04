package fanout

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"metaldocs/internal/modules/iam/authz"
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
	"metaldocs.pdf_dispatch_outbox":         {},
	"metaldocs.materialize_dispatch_outbox": {},
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

// Enqueue inserts a pending outbox row inside the caller's business transaction
// and returns the newly-inserted row's id. The outbox INSERT MUST share the
// caller's business transaction (atomic dispatch). A nil tx would silently
// autocommit the outbox row outside that transaction, breaking the
// transactional-outbox guarantee — fail loud (db.Tx contract: a nil Tx is
// never valid).
//
// The (tenant_id, revision_id) ON CONFLICT is the single dedup point for
// staging dispatch: a duplicate enqueue for the same tenant+revision hits
// DO NOTHING, so RETURNING yields zero rows. That is not an error — it is a
// successful dedup skip — and is reported to the caller as an empty id with a
// nil error, letting the caller (the dispatchjobs.Enqueuer) decide to skip
// the paired River insert.
func (r *StagingOutboxRepository) Enqueue(ctx context.Context, tx db.Tx, tenantID, revisionID string, contentHash []byte) (string, error) {
	if tx == nil {
		return "", fmt.Errorf("%s enqueue: tx must not be nil", r.name)
	}
	var id string
	//nolint:gosec // table name is allowlist-validated at construction
	err := tx.QueryRowContext(ctx, fmt.Sprintf(`
INSERT INTO %s (tenant_id, revision_id, content_hash)
VALUES ($1::uuid, $2::uuid, $3)
ON CONFLICT (tenant_id, revision_id) DO NOTHING
RETURNING id`, r.table),
		tenantID, revisionID, contentHash).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// ON CONFLICT DO NOTHING skipped the insert: dedup, not a failure.
			return "", nil
		}
		return "", fmt.Errorf("%s enqueue: %w", r.name, err)
	}
	return id, nil
}

// ClaimPending intentionally claims across ALL tenants — sanctioned by ADR 0054
// (cross-tenant claim for background outbox consumers, mirroring the platform
// consumer in internal/platform/messaging/outbox/postgres/consumer.go). Tenancy
// is enforced at processing time: each returned row carries TenantID and the
// consumer must scope all per-row work to it. Do not add a tenant_id predicate.
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

// MarkDispatched records successful dispatch of one claimed row. tenantID
// (the row's own OutboxRow.TenantID, captured at claim time) is seeded via
// authz.SeedTxTenant in a dedicated per-row tx BEFORE the write, engaging the
// FORCE RLS backstop for this single-tenant processing step (M3 F3.2 —
// validation-contract.md §2.2 site 5; ADR 0054 rule 2). The cross-tenant
// ClaimPending step above is intentionally exempt (ADR 0054 rule 1) and stays
// GUC-unset.
func (r *StagingOutboxRepository) MarkDispatched(ctx context.Context, tenantID, id string) error {
	return r.inSeededTx(ctx, tenantID, func(tx *sql.Tx) error {
		//nolint:gosec // table name is allowlist-validated at construction
		res, err := tx.ExecContext(ctx, fmt.Sprintf(`
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
	})
}

// MarkFailed records a dispatch failure (retry or permanent) for one claimed
// row, seeded the same way as MarkDispatched (see its doc comment).
func (r *StagingOutboxRepository) MarkFailed(ctx context.Context, tenantID, id string, errStr string, nextRetryAt time.Time, finalize bool) error {
	return r.inSeededTx(ctx, tenantID, func(tx *sql.Tx) error {
		if finalize {
			// F-R3: set dead_lettered_at when permanently failing a row, mirroring
			// internal/platform/messaging/outbox/postgres/consumer.go:152-177.
			//nolint:gosec // table name is allowlist-validated at construction
			res, err := tx.ExecContext(ctx, fmt.Sprintf(`
UPDATE %s
   SET status='failed', last_error=$2, attempts=attempts+1, dead_lettered_at=NOW()
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
		res, err := tx.ExecContext(ctx, fmt.Sprintf(`
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
	})
}

// inSeededTx opens a tx, seeds it with tenantID via authz.SeedTxTenant BEFORE
// running fn, then commits/rolls back. This is the shared tx-wrap for the
// per-row processing writes (MarkDispatched/MarkFailed); the cross-tenant
// ClaimPending claim step does not use this helper (ADR 0054 rule 1).
func (r *StagingOutboxRepository) inSeededTx(ctx context.Context, tenantID string, fn func(tx *sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: begin tx: %w", r.name, err)
	}
	if err := authz.SeedTxTenant(ctx, tx, tenantID); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("%s: seed tenant: %w", r.name, err)
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%s: commit: %w", r.name, err)
	}
	return nil
}

// CountDeadLettered returns the number of rows that have been permanently
// dead-lettered in this outbox table. Mirrors the dead_lettered_at visibility
// pattern from internal/platform/messaging/outbox/postgres/consumer.go.
func (r *StagingOutboxRepository) CountDeadLettered(ctx context.Context) (int, error) {
	var n int
	//nolint:gosec // table name is allowlist-validated at construction
	err := r.db.QueryRowContext(ctx, fmt.Sprintf(
		`SELECT COUNT(*) FROM %s WHERE dead_lettered_at IS NOT NULL`, r.table),
	).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("%s count dead lettered: %w", r.name, err)
	}
	return n, nil
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
