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

func NewExportJobRepository() *ExportJobRepository {
	return &ExportJobRepository{jobs: map[string]domain.ExportJob{}}
}

func (r *ExportJobRepository) Save(_ context.Context, job domain.ExportJob) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.jobs[job.ID] = job
	return nil
}

func (r *ExportJobRepository) Get(_ context.Context, tenantID, exportID string) (domain.ExportJob, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	job, ok := r.jobs[exportID]
	if !ok || job.TenantID != tenantID {
		return domain.ExportJob{}, domain.ErrExportJobNotFound
	}
	return job, nil
}

func (r *ExportJobRepository) GetByDownloadToken(_ context.Context, exportID, token string) (domain.ExportJob, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	job, ok := r.jobs[exportID]
	if !ok || job.DownloadToken == "" || job.DownloadToken != token {
		return domain.ExportJob{}, domain.ErrExportJobNotFound
	}
	return job, nil
}
