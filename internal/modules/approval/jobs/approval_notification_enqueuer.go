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
