package jobs

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"

	documentsdomain "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/platform/db"
)

// riverInserter captures river.Client[*sql.Tx].InsertTx's exact signature.
// *river.Client[*sql.Tx] satisfies this directly (no adapter needed) because
// its InsertTx method already has this shape once TTx is instantiated to
// *sql.Tx. Unexported and minimal so RiverLifecycleEventEnqueuer is
// unit-testable with a recording fake in place of a real River client /
// database, mirroring dispatchjobs.riverInserter.
type riverInserter interface {
	InsertTx(ctx context.Context, tx *sql.Tx, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error)
}

// RiverLifecycleEventEnqueuer implements documentsdomain.LifecycleEventEnqueuer
// using River's same-tx InsertTx, mirroring RiverScheduledPublishEnqueuer.
// The db.Tx → *sql.Tx assertion is the one allowed coupling point with River infra.
type RiverLifecycleEventEnqueuer struct {
	Client riverInserter
}

// NewLifecycleEventEnqueuer wraps a River client as a LifecycleEventEnqueuer.
func NewLifecycleEventEnqueuer(client *river.Client[*sql.Tx]) documentsdomain.LifecycleEventEnqueuer {
	return &RiverLifecycleEventEnqueuer{Client: client}
}

// EnqueueLifecycleEventTx enqueues a lifecycle-event job within tx (outbox
// pattern). tx must be a *sql.Tx. Lifecycle-fanout is request-triggered (fired
// by a user action: publish/supersede/obsolete/approve/reject), the same
// class as scheduled-publish and staging dispatch, so it is routed to the
// "temporal" queue — metaldocs-jobs never subscribes River's default queue.
func (e *RiverLifecycleEventEnqueuer) EnqueueLifecycleEventTx(ctx context.Context, tx db.Tx, args documentsdomain.LifecycleEventArgs) error {
	sqlTx, ok := tx.(*sql.Tx)
	if !ok {
		return fmt.Errorf("lifecycle_event_enqueuer: river requires *sql.Tx, got %T", tx)
	}
	_, err := e.Client.InsertTx(ctx, sqlTx, args, &river.InsertOpts{Queue: "temporal"})
	if err != nil {
		return fmt.Errorf("lifecycle_event_enqueuer: enqueue %s: %w", args.EventType, err)
	}
	return nil
}
