package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"metaldocs/internal/modules/audit/domain"
)

// ExportJobRepository is the postgres-backed implementation. PR-6 stores the
// rendered CSV/JSONL blob inline in the row's payload column. A future async
// path can store the blob in MinIO/S3 and only keep metadata + object_key.
type ExportJobRepository struct {
	db *sql.DB
}

func NewExportJobRepository(db *sql.DB) *ExportJobRepository {
	return &ExportJobRepository{db: db}
}

const exportJobColumns = `id, tenant_id, actor_id, format, filter_json, status, object_key, download_token,
       expires_at, error_message, estimated_rows, actual_rows, payload, created_at, completed_at`

func (r *ExportJobRepository) Save(ctx context.Context, job domain.ExportJob) error {
	const q = `
INSERT INTO metaldocs.audit_export_jobs (
  id, tenant_id, actor_id, format, filter_json, status, object_key, download_token,
  expires_at, error_message, estimated_rows, actual_rows, payload, created_at, completed_at
) VALUES ($1, $2, $3, $4, $5::jsonb, $6, NULLIF($7, ''), NULLIF($8, ''),
          $9, NULLIF($10, ''), $11, $12, $13, $14, $15)
`
	var completedAt any
	if !job.CompletedAt.IsZero() {
		completedAt = job.CompletedAt
	}
	var expiresAt any
	if !job.ExpiresAt.IsZero() {
		expiresAt = job.ExpiresAt
	}
	if _, err := r.db.ExecContext(ctx, q,
		job.ID, job.TenantID, job.ActorID, string(job.Format), job.FilterJSON, string(job.Status),
		job.ObjectKey, job.DownloadToken, expiresAt, job.ErrorMessage,
		job.EstimatedRows, job.ActualRows, job.Payload, job.CreatedAt, completedAt,
	); err != nil {
		return fmt.Errorf("insert audit export job: %w", err)
	}
	return nil
}

func (r *ExportJobRepository) Get(ctx context.Context, tenantID, exportID string) (domain.ExportJob, error) {
	q := `SELECT ` + exportJobColumns + ` FROM metaldocs.audit_export_jobs WHERE id = $1 AND tenant_id = $2`
	return scanJob(r.db.QueryRowContext(ctx, q, exportID, tenantID))
}

func (r *ExportJobRepository) GetByDownloadToken(ctx context.Context, exportID, token string) (domain.ExportJob, error) {
	q := `SELECT ` + exportJobColumns + ` FROM metaldocs.audit_export_jobs WHERE id = $1 AND download_token = $2`
	return scanJob(r.db.QueryRowContext(ctx, q, exportID, token))
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanJob(row rowScanner) (domain.ExportJob, error) {
	var (
		job          domain.ExportJob
		format       string
		status       string
		objectKey    sql.NullString
		token        sql.NullString
		expiresAt    sql.NullTime
		errorMessage sql.NullString
		completedAt  sql.NullTime
		payload      []byte
	)
	err := row.Scan(
		&job.ID, &job.TenantID, &job.ActorID, &format, &job.FilterJSON, &status, &objectKey, &token,
		&expiresAt, &errorMessage, &job.EstimatedRows, &job.ActualRows, &payload, &job.CreatedAt, &completedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ExportJob{}, domain.ErrExportJobNotFound
		}
		return domain.ExportJob{}, fmt.Errorf("scan audit export job: %w", err)
	}
	job.Format = domain.ExportFormat(format)
	job.Status = domain.ExportStatus(status)
	if objectKey.Valid {
		job.ObjectKey = objectKey.String
	}
	if token.Valid {
		job.DownloadToken = token.String
	}
	if expiresAt.Valid {
		job.ExpiresAt = expiresAt.Time
	}
	if errorMessage.Valid {
		job.ErrorMessage = errorMessage.String
	}
	if completedAt.Valid {
		job.CompletedAt = completedAt.Time
	}
	job.Payload = payload
	return job, nil
}
