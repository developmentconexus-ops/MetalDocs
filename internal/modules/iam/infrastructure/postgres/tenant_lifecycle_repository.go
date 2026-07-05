// tenant_lifecycle_repository.go — M7 F7.3 Task E: the persistence adapter
// for metaldocs.tenant_lifecycle_jobs. Split from tenant_repository.go (which
// owns metaldocs.tenants itself) because this repository's RunJob-side
// methods (LoadLifecycleJob/MarkLifecycleJob*) run OFF any authz-seeded tx
// (they are called by the River worker, off the pool, sometimes before any
// tenant GUC is seeded) — a distinct lifecycle from InsertTenantTx/
// TenantLookupTx's always-tx-scoped shape.
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

// TenantLifecycleRepository implements the application-layer
// lifecycleJobInserter and lifecycleJobStore ports against
// metaldocs.tenant_lifecycle_jobs.
type TenantLifecycleRepository struct {
	db *sql.DB
}

// NewTenantLifecycleRepository constructs a pool-backed repository. db is
// used directly by the run-side (load/mark) methods, which do not run
// inside the enqueue tx; InsertLifecycleJobTx always uses the caller's tx.
func NewTenantLifecycleRepository(db *sql.DB) *TenantLifecycleRepository {
	return &TenantLifecycleRepository{db: db}
}

// InsertLifecycleJobTx inserts a pending tenant_lifecycle_jobs row inside
// the caller's tx (M7 F7.3 enqueue path — the tripwire arm on this table
// requires tenant.export/tenant.erase already asserted on tx, and RLS
// requires the target tenant's GUC already seeded; both are the caller's
// responsibility, done before this call per TenantLifecycleService.request).
func (r *TenantLifecycleRepository) InsertLifecycleJobTx(ctx context.Context, tx *sql.Tx, jobID, tenantID, kind, requestedBy string) error {
	_, err := tx.ExecContext(ctx, `
INSERT INTO metaldocs.tenant_lifecycle_jobs (id, tenant_id, kind, status, requested_by)
VALUES ($1::uuid, $2::uuid, $3, 'pending', $4)
`, jobID, tenantID, kind, requestedBy)
	if err != nil {
		return fmt.Errorf("insert tenant lifecycle job: %w", err)
	}
	return nil
}

// LoadLifecycleJob reads one row by id, off any tx (pool read — the River
// worker's RunJob has no authz-seeded tx of its own at this point; the
// row's own tenant_id becomes the seed target for the run-side fan-out tx).
func (r *TenantLifecycleRepository) LoadLifecycleJob(ctx context.Context, jobID string) (iamdomain.TenantLifecycleJob, error) {
	var job iamdomain.TenantLifecycleJob
	var objectKey, errMsg sql.NullString
	err := r.db.QueryRowContext(ctx, `
SELECT id, tenant_id, kind, status, requested_by, object_key, error
  FROM metaldocs.tenant_lifecycle_jobs
 WHERE id = $1::uuid
`, jobID).Scan(&job.ID, &job.TenantID, &job.Kind, &job.Status, &job.RequestedBy, &objectKey, &errMsg)
	if errors.Is(err, sql.ErrNoRows) {
		return iamdomain.TenantLifecycleJob{}, fmt.Errorf("tenant lifecycle job %s: %w", jobID, sql.ErrNoRows)
	}
	if err != nil {
		return iamdomain.TenantLifecycleJob{}, fmt.Errorf("load tenant lifecycle job: %w", err)
	}
	job.ObjectKey = objectKey.String
	job.Error = errMsg.String
	return job, nil
}

// MarkLifecycleJobRunning transitions the job to status 'running'. Explicit
// tenant_id-free predicate is fine here (id is already globally unique,
// this is a system/pool-level write, not a tenant-scoped read of business
// data) — mirrors the run-side's pool-handle access pattern throughout.
func (r *TenantLifecycleRepository) MarkLifecycleJobRunning(ctx context.Context, jobID string) error {
	_, err := r.db.ExecContext(ctx, `
UPDATE metaldocs.tenant_lifecycle_jobs SET status = 'running' WHERE id = $1::uuid
`, jobID)
	if err != nil {
		return fmt.Errorf("mark tenant lifecycle job running: %w", err)
	}
	return nil
}

// MarkLifecycleJobReady transitions the job to 'ready', recording objectKey
// (empty for erase jobs, which have no artifact) and completed_at = now().
func (r *TenantLifecycleRepository) MarkLifecycleJobReady(ctx context.Context, jobID, objectKey string) error {
	_, err := r.db.ExecContext(ctx, `
UPDATE metaldocs.tenant_lifecycle_jobs
   SET status = 'ready', object_key = NULLIF($2, ''), completed_at = now(), error = NULL
 WHERE id = $1::uuid
`, jobID, objectKey)
	if err != nil {
		return fmt.Errorf("mark tenant lifecycle job ready: %w", err)
	}
	return nil
}

// MarkLifecycleJobFailed transitions the job to 'failed', recording errMsg
// and completed_at = now().
func (r *TenantLifecycleRepository) MarkLifecycleJobFailed(ctx context.Context, jobID, errMsg string) error {
	_, err := r.db.ExecContext(ctx, `
UPDATE metaldocs.tenant_lifecycle_jobs
   SET status = 'failed', error = $2, completed_at = now()
 WHERE id = $1::uuid
`, jobID, errMsg)
	if err != nil {
		return fmt.Errorf("mark tenant lifecycle job failed: %w", err)
	}
	return nil
}
