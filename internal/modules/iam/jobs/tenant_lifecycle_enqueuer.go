// tenant_lifecycle_enqueuer.go — M7 F7.3 Task E: the River-backed
// implementation of iamdomain.TenantLifecycleEnqueuer. Mirrors
// documents/approval/jobs/lifecycle_event_enqueuer.go exactly: a narrow
// riverInserter interface capturing river.Client[*sql.Tx].InsertTx's real
// signature (so the production client satisfies it with no wrapper), a
// thin struct wrapping it, and a same-tx InsertTx call using queue
// "temporal" (every request-triggered, non-default-queue job in this
// codebase uses "temporal" — metaldocs-jobs never subscribes River's
// default queue).
package jobs

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
)

// riverInserter captures river.Client[*sql.Tx].InsertTx's exact signature.
// *river.Client[*sql.Tx] satisfies this directly (no adapter needed).
// Unexported and minimal so RiverTenantLifecycleEnqueuer is unit-testable
// with a recording fake in place of a real River client/database.
type riverInserter interface {
	InsertTx(ctx context.Context, tx *sql.Tx, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error)
}

// RiverTenantLifecycleEnqueuer implements iamdomain.TenantLifecycleEnqueuer
// using River's same-tx InsertTx.
type RiverTenantLifecycleEnqueuer struct {
	client riverInserter
}

// NewTenantLifecycleEnqueuer wraps a River client as a
// iamdomain.TenantLifecycleEnqueuer.
func NewTenantLifecycleEnqueuer(client *river.Client[*sql.Tx]) iamdomain.TenantLifecycleEnqueuer {
	return &RiverTenantLifecycleEnqueuer{client: client}
}

// EnqueueTenantLifecycleJobTx enqueues a tenant_lifecycle job within tx
// (transactional-outbox pairing with the tenant_lifecycle_jobs row insert
// TenantLifecycleService performs in the same tx). tx must be a *sql.Tx.
func (e *RiverTenantLifecycleEnqueuer) EnqueueTenantLifecycleJobTx(ctx context.Context, tx db.Tx, args iamdomain.TenantLifecycleJobArgs) error {
	sqlTx, ok := tx.(*sql.Tx)
	if !ok {
		return fmt.Errorf("tenant_lifecycle_enqueuer: river requires *sql.Tx, got %T", tx)
	}
	_, err := e.client.InsertTx(ctx, sqlTx, args, &river.InsertOpts{Queue: "temporal"})
	if err != nil {
		return fmt.Errorf("tenant_lifecycle_enqueuer: enqueue job %s for tenant %s: %w", args.JobKind, args.TenantID, err)
	}
	return nil
}
