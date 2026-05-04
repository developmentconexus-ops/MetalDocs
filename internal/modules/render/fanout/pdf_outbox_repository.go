package fanout

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// OutboxTx is the subset of *sql.Tx used by PDFOutboxRepository.Enqueue.
// Defined locally to avoid an import cycle with documents/repository.
type OutboxTx interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type OutboxRow struct {
	ID          string
	TenantID    string
	RevisionID  string
	ContentHash []byte
	Attempts    int
}

type PDFOutboxRepository struct{ db *sql.DB }

func NewPDFOutboxRepository(db *sql.DB) *PDFOutboxRepository {
	return &PDFOutboxRepository{db: db}
}

func (r *PDFOutboxRepository) Enqueue(ctx context.Context, tx OutboxTx, tenantID, revisionID string, contentHash []byte) error {
	var exec OutboxTx = r.db
	if tx != nil {
		exec = tx
	}
	_, err := exec.ExecContext(ctx, `
INSERT INTO metaldocs.pdf_dispatch_outbox (tenant_id, revision_id, content_hash)
VALUES ($1::uuid, $2::uuid, $3)
ON CONFLICT (tenant_id, revision_id) DO NOTHING`,
		tenantID, revisionID, contentHash)
	if err != nil {
		return fmt.Errorf("pdf outbox enqueue: %w", err)
	}
	return nil
}

func (r *PDFOutboxRepository) ClaimPending(ctx context.Context, limit int) ([]OutboxRow, error) {
	rows, err := r.db.QueryContext(ctx, `
WITH claimed AS (
  SELECT id FROM metaldocs.pdf_dispatch_outbox
   WHERE status = 'pending' AND next_retry_at <= NOW() AND attempts < 5
   ORDER BY next_retry_at ASC
   LIMIT $1
   FOR UPDATE SKIP LOCKED
)
UPDATE metaldocs.pdf_dispatch_outbox o
   SET status='processing', claimed_at=NOW()
  FROM claimed c
 WHERE o.id = c.id
RETURNING o.id, o.tenant_id, o.revision_id, o.content_hash, o.attempts`, limit)
	if err != nil {
		return nil, fmt.Errorf("claim pending: %w", err)
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

func (r *PDFOutboxRepository) MarkDispatched(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `
UPDATE metaldocs.pdf_dispatch_outbox
   SET status='dispatched', dispatched_at=NOW()
 WHERE id=$1::uuid`, id)
	return err
}

func (r *PDFOutboxRepository) MarkFailed(ctx context.Context, id string, errStr string, nextRetryAt time.Time, finalize bool) error {
	if finalize {
		_, err := r.db.ExecContext(ctx, `
UPDATE metaldocs.pdf_dispatch_outbox
   SET status='failed', last_error=$2, attempts=attempts+1
 WHERE id=$1::uuid`, id, errStr)
		return err
	}
	_, err := r.db.ExecContext(ctx, `
UPDATE metaldocs.pdf_dispatch_outbox
   SET status='pending', last_error=$2, attempts=attempts+1, next_retry_at=$3, claimed_at=NULL
 WHERE id=$1::uuid`, id, errStr, nextRetryAt)
	return err
}

func (r *PDFOutboxRepository) ResetStaleClaims(ctx context.Context, olderThan time.Duration) (int, error) {
	res, err := r.db.ExecContext(ctx, `
UPDATE metaldocs.pdf_dispatch_outbox
   SET status='pending', claimed_at=NULL
 WHERE status='processing' AND claimed_at < NOW() - $1::interval`,
		fmt.Sprintf("%d milliseconds", olderThan.Milliseconds()))
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}
