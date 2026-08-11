package memory

import (
	"context"
	"sync"

	"metaldocs/internal/modules/audit/domain"
)

// ExportJobRepository keeps export jobs in process memory.
type ExportJobRepository struct {
	mu   sync.Mutex
	jobs map[string]domain.ExportJob
}

// NewExportJobRepository constructs an empty in-memory ExportJobRepository.
func NewExportJobRepository() *ExportJobRepository {
	return &ExportJobRepository{jobs: map[string]domain.ExportJob{}}
}

// Save upserts job by ID. Never returns an error.
func (r *ExportJobRepository) Save(_ context.Context, job domain.ExportJob) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.jobs[job.ID] = job
	return nil
}

// Len reports how many export jobs are currently stored. Test-only helper —
// used to assert that a refused export (e.g. ErrExportTenantErased) left no
// row behind, not merely that ExportEvents returned an error.
func (r *ExportJobRepository) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.jobs)
}

// Get returns the job for exportID scoped to tenantID. Returns
// domain.ErrExportJobNotFound when absent or owned by a different tenant.
func (r *ExportJobRepository) Get(_ context.Context, tenantID, exportID string) (domain.ExportJob, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	job, ok := r.jobs[exportID]
	if !ok || job.TenantID.String() != tenantID {
		return domain.ExportJob{}, domain.ErrExportJobNotFound
	}
	return job, nil
}

// GetByDownloadToken returns the job for exportID iff its stored
// DownloadToken matches token. Returns domain.ErrExportJobNotFound on any
// mismatch or empty stored token.
func (r *ExportJobRepository) GetByDownloadToken(_ context.Context, exportID, token string) (domain.ExportJob, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	job, ok := r.jobs[exportID]
	if !ok || job.DownloadToken == "" || job.DownloadToken != token {
		return domain.ExportJob{}, domain.ErrExportJobNotFound
	}
	return job, nil
}
