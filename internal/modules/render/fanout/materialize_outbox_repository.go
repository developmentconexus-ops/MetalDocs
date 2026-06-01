package fanout

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type MaterializeOutboxRow struct {
	ID          string
	TenantID    string
	RevisionID  string
	ContentHash []byte
	Attempts    int
}

type MaterializeOutboxRepository struct{ db *sql.DB }

func NewMaterializeOutboxRepository(db *sql.DB) *MaterializeOutboxRepository {
	return &MaterializeOutboxRepository{db: db}
}

func (r *MaterializeOutboxRepository) Enqueue(ctx context.Context, tx *sql.Tx, tenantID, revisionID string, contentHash []byte) error {
	var exec interface {
		ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	} = r.db
	if tx != nil {
		exec = tx
	}
	_, err := exec.ExecContext(ctx, `
INSERT INTO metaldocs.materialize_dispatch_outbox (tenant_id, revision_id, content_hash)
VALUES ($1::uuid, $2::uuid, $3)
ON CONFLICT (tenant_id, revision_id) DO NOTHING`,
		tenantID, revisionID, contentHash)
	if err != nil {
		return fmt.Errorf("materialize outbox enqueue: %w", err)
	}
	return nil
}

func (r *MaterializeOutboxRepository) ClaimPending(ctx context.Context, limit, maxAttempts int) ([]MaterializeOutboxRow, error) {
	rows, err := r.db.QueryContext(ctx, `
WITH claimed AS (
  SELECT id FROM metaldocs.materialize_dispatch_outbox
   WHERE status = 'pending' AND next_retry_at <= NOW() AND attempts < $2
   ORDER BY next_retry_at ASC
   LIMIT $1
   FOR UPDATE SKIP LOCKED
)
UPDATE metaldocs.materialize_dispatch_outbox o
   SET status='processing', claimed_at=NOW()
  FROM claimed c
 WHERE o.id = c.id
RETURNING o.id, o.tenant_id, o.revision_id, o.content_hash, o.attempts`, limit, maxAttempts)
	if err != nil {
		return nil, fmt.Errorf("materialize outbox claim pending: %w", err)
	}
	defer rows.Close()
	var out []MaterializeOutboxRow
	for rows.Next() {
		var row MaterializeOutboxRow
		if err := rows.Scan(&row.ID, &row.TenantID, &row.RevisionID, &row.ContentHash, &row.Attempts); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *MaterializeOutboxRepository) MarkDispatched(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `
UPDATE metaldocs.materialize_dispatch_outbox
   SET status='dispatched', dispatched_at=NOW()
 WHERE id=$1::uuid`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("materialize outbox mark dispatched: row not found: id=%s", id)
	}
	return nil
}

func (r *MaterializeOutboxRepository) MarkFailed(ctx context.Context, id string, errStr string, nextRetryAt time.Time, finalize bool) error {
	if finalize {
		res, err := r.db.ExecContext(ctx, `
UPDATE metaldocs.materialize_dispatch_outbox
   SET status='failed', last_error=$2, attempts=attempts+1
 WHERE id=$1::uuid`, id, errStr)
		if err != nil {
			return err
		}
		n, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if n == 0 {
			return fmt.Errorf("materialize outbox mark failed finalize: row not found: id=%s", id)
		}
		return nil
	}
	res, err := r.db.ExecContext(ctx, `
UPDATE metaldocs.materialize_dispatch_outbox
   SET status='pending', last_error=$2, attempts=attempts+1, next_retry_at=$3, claimed_at=NULL
 WHERE id=$1::uuid`, id, errStr, nextRetryAt)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("materialize outbox mark failed retry: row not found: id=%s", id)
	}
	return nil
}

// ReadState returns the current status for the given tenant+revision.
// Returns ("", nil) when no row exists yet.
func (r *MaterializeOutboxRepository) ReadState(ctx context.Context, tenantID, revisionID string) (string, error) {
	var status string
	err := r.db.QueryRowContext(ctx, `
SELECT status FROM metaldocs.materialize_dispatch_outbox
 WHERE tenant_id=$1::uuid AND revision_id=$2::uuid
 ORDER BY created_at DESC LIMIT 1`,
		tenantID, revisionID,
	).Scan(&status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("materialize outbox read state: %w", err)
	}
	return status, nil
}

func (r *MaterializeOutboxRepository) ResetStaleClaims(ctx context.Context, olderThan time.Duration) (int, error) {
	res, err := r.db.ExecContext(ctx, `
UPDATE metaldocs.materialize_dispatch_outbox
   SET status='pending', claimed_at=NULL
 WHERE status='processing' AND claimed_at < NOW() - ($1 * interval '1 millisecond')`,
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
