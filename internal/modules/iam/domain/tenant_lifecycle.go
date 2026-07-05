package domain

import (
	"context"

	"metaldocs/internal/platform/db"
)

// TenantLifecycleJobArgs is the River job payload enqueued alongside the
// tenant_lifecycle_jobs row (M7 F7.3). Declared in domain (not application)
// so the River adapter (internal/modules/iam/jobs, mirroring
// documents/approval/jobs/lifecycle_event_enqueuer.go) can implement
// river.JobArgs' Kind() method on this exact type without importing
// application, matching documentsdomain.LifecycleEventArgs' placement.
type TenantLifecycleJobArgs struct {
	JobID    string `json:"job_id"`
	TenantID string `json:"tenant_id"`
	JobKind  string `json:"kind"`
}

// Kind implements river.JobArgs, identifying this job type to River. Named
// distinctly from the JobKind field (export/erase) to avoid a field/method
// name collision — River's Kind() is the job TYPE registered with
// river.AddWorker, not this job's export/erase business kind.
func (TenantLifecycleJobArgs) Kind() string { return "tenant_lifecycle" }

// TenantLifecycleEnqueuer is the narrow port TenantLifecycleService depends
// on to enqueue the paired River job inside the same tx as the
// tenant_lifecycle_jobs INSERT (transactional-outbox pairing). Implemented
// by *jobs.RiverTenantLifecycleEnqueuer (River-backed) in production;
// TenantLifecycleService depends only on this interface, never on River
// directly — mirrors documentsdomain.LifecycleEventEnqueuer.
type TenantLifecycleEnqueuer interface {
	EnqueueTenantLifecycleJobTx(ctx context.Context, tx db.Tx, args TenantLifecycleJobArgs) error
}

// TenantLifecycleJob is the shared (application + infrastructure) view of
// one metaldocs.tenant_lifecycle_jobs row (M7 F7.3). Hosted in domain, not
// application or infrastructure/postgres, so the postgres repository (a
// leaf, never importing application) and the application-layer
// TenantLifecycleService (which depends downward on domain, per every other
// iam service) can share one type without either layer depending on the
// other.
type TenantLifecycleJob struct {
	ID          string
	TenantID    string
	Kind        string
	Status      string
	RequestedBy string
	ObjectKey   string
	Error       string
}
