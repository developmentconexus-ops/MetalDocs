package jobs

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/riverqueue/river"

	documentsdomain "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/platform/db"
)

// RiverLifecycleEventEnqueuer implements documentsdomain.LifecycleEventEnqueuer
// using River's same-tx InsertTx, mirroring RiverScheduledPublishEnqueuer.
// The db.Tx → *sql.Tx assertion is the one allowed coupling point with River infra.
type RiverLifecycleEventEnqueuer struct {
	Client *river.Client[*sql.Tx]
}

// NewLifecycleEventEnqueuer wraps a River client as a LifecycleEventEnqueuer.
func NewLifecycleEventEnqueuer(client *river.Client[*sql.Tx]) documentsdomain.LifecycleEventEnqueuer {
	return &RiverLifecycleEventEnqueuer{Client: client}
}

func (e *RiverLifecycleEventEnqueuer) EnqueueLifecycleEventTx(ctx context.Context, tx db.Tx, args documentsdomain.LifecycleEventArgs) error {
	sqlTx, ok := tx.(*sql.Tx)
	if !ok {
		return fmt.Errorf("lifecycle_event_enqueuer: river requires *sql.Tx, got %T", tx)
	}
	_, err := e.Client.InsertTx(ctx, sqlTx, args, nil)
	if err != nil {
		return fmt.Errorf("lifecycle_event_enqueuer: enqueue %s: %w", args.EventType, err)
	}
	return nil
}
