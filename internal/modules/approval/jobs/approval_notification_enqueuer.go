package jobs

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/riverqueue/river"

	approvaldomain "metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/platform/db"
)

// RiverApprovalNotificationEnqueuer implements
// approvaldomain.ApprovalNotificationEnqueuer over River's same-tx InsertTx,
// mirroring RiverLifecycleEventEnqueuer exactly. The db.Tx → *sql.Tx assertion
// is the one allowed coupling point with River infra.
type RiverApprovalNotificationEnqueuer struct {
	Client riverInserter
}

// NewApprovalNotificationEnqueuer wraps a River client as the approval-owned
// notification enqueuer.
func NewApprovalNotificationEnqueuer(client *river.Client[*sql.Tx]) approvaldomain.ApprovalNotificationEnqueuer {
	return &RiverApprovalNotificationEnqueuer{Client: client}
}

// EnqueueApprovalNotificationTx enqueues the notification inside tx (outbox
// pattern). Routed to the "temporal" queue for the same reason lifecycle fanout
// is: it is request-triggered work, and metaldocs-jobs never subscribes River's
// default queue.
func (e *RiverApprovalNotificationEnqueuer) EnqueueApprovalNotificationTx(ctx context.Context, tx db.Tx, args approvaldomain.ApprovalNotificationArgs) error {
	sqlTx, ok := tx.(*sql.Tx)
	if !ok {
		return fmt.Errorf("approval_notification_enqueuer: river requires *sql.Tx, got %T", tx)
	}
	if _, err := e.Client.InsertTx(ctx, sqlTx, args, &river.InsertOpts{Queue: "temporal"}); err != nil {
		return fmt.Errorf("approval_notification_enqueuer: enqueue %s: %w", args.EventType, err)
	}
	return nil
}

// DeferredApprovalNotificationEnqueuer breaks the same wiring cycle
// DeferredReleaseEvaluationEnqueuer breaks (see release_evaluate_job.go): in
// the jobs binary, the worker set (which needs this enqueuer, e.g. the F8
// approval_sla_surfacer) must be constructed BEFORE the River client that the
// enqueuer inserts into exists. Bind is called once, during startup, before
// the client starts working jobs — so no synchronisation is needed and an
// unbound use is a wiring bug that fails loud rather than silently dropping
// overdue notifications.
type DeferredApprovalNotificationEnqueuer struct {
	inner approvaldomain.ApprovalNotificationEnqueuer
}

// NewDeferredApprovalNotificationEnqueuer returns an unbound enqueuer.
func NewDeferredApprovalNotificationEnqueuer() *DeferredApprovalNotificationEnqueuer {
	return &DeferredApprovalNotificationEnqueuer{}
}

// Bind supplies the River client once it exists.
func (e *DeferredApprovalNotificationEnqueuer) Bind(client *river.Client[*sql.Tx]) {
	e.inner = NewApprovalNotificationEnqueuer(client)
}

// EnqueueApprovalNotificationTx delegates to the bound enqueuer.
func (e *DeferredApprovalNotificationEnqueuer) EnqueueApprovalNotificationTx(ctx context.Context, tx db.Tx, args approvaldomain.ApprovalNotificationArgs) error {
	if e.inner == nil {
		return fmt.Errorf("approval_notification_enqueuer: deferred enqueuer used before Bind (startup wiring bug)")
	}
	return e.inner.EnqueueApprovalNotificationTx(ctx, tx, args)
}
