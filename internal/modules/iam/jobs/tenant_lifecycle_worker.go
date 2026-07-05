// tenant_lifecycle_worker.go — M7 F7.3 Task E: the metaldocs-jobs River
// worker that consumes TenantLifecycleJobArgs and runs the export/erase
// fan-out via TenantLifecycleService.RunJob. Registered in
// apps/jobs/cmd/metaldocs-jobs/main.go alongside the other river.AddWorker
// calls (mirrors dispatchjobs/workers.go's worker registration shape).
package jobs

import (
	"context"

	"github.com/riverqueue/river"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

// tenantLifecycleRunner is the narrow application-layer surface the worker
// depends on. Implemented by *iamapp.TenantLifecycleService; declared here
// (not imported from application) so this file's only import-cycle-risk
// dependency is on iamdomain, mirroring the enqueuer's dependency shape.
type tenantLifecycleRunner interface {
	RunJob(ctx context.Context, jobID string) error
}

// TenantLifecycleWorker is a River worker that consumes
// iamdomain.TenantLifecycleJobArgs jobs and runs the export/erase fan-out.
// The worker itself performs no authz/GUC work directly — RunJob's erase
// path establishes its own background-bypass context internally (H-PRE-1:
// WithBackgroundBypass is called inside RunJob's erase branch, at the
// fan-out tx boundary, not here at the worker's Work entrypoint) — matching
// the boundary NotificationsFanoutWorker draws between "worker as thin River
// adapter" and "service as the place authz context is established".
type TenantLifecycleWorker struct {
	river.WorkerDefaults[iamdomain.TenantLifecycleJobArgs]
	service tenantLifecycleRunner
}

// NewTenantLifecycleWorker builds a ready worker. service is required.
func NewTenantLifecycleWorker(service tenantLifecycleRunner) *TenantLifecycleWorker {
	if service == nil {
		panic("tenant_lifecycle_worker: service is required")
	}
	return &TenantLifecycleWorker{service: service}
}

// Work runs the job identified by args.JobID. Idempotent: RunJob is a
// no-op when the job is already status "ready" (see
// TenantLifecycleService.RunJob), so a River retry after a transient
// failure (e.g. worker crash between fan-out completion and status update)
// never double-runs a completed export/erase.
func (w *TenantLifecycleWorker) Work(ctx context.Context, job *river.Job[iamdomain.TenantLifecycleJobArgs]) error {
	return w.service.RunJob(ctx, job.Args.JobID)
}
