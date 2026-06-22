package jobs_test

import (
	"context"
	"database/sql"
	"testing"

	"metaldocs/internal/modules/documents/approval/jobs"
	documentsdomain "metaldocs/internal/modules/documents/domain"
)

// Verify compile-time: RiverLifecycleEventEnqueuer implements LifecycleEventEnqueuer.
var _ documentsdomain.LifecycleEventEnqueuer = (*jobs.RiverLifecycleEventEnqueuer)(nil)

func TestRiverLifecycleEventEnqueuer_WrongTxType(t *testing.T) {
	enc := &jobs.RiverLifecycleEventEnqueuer{} // Client nil — only tests type assertion
	err := enc.EnqueueLifecycleEventTx(context.Background(), wrongTx{}, documentsdomain.LifecycleEventArgs{})
	if err == nil {
		t.Fatal("want error for wrong tx type, got nil")
	}
}

// wrongTx is a db.Tx that is NOT *sql.Tx, triggering the type assertion failure.
type wrongTx struct{}

func (wrongTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return nil, nil
}
func (wrongTx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return nil, nil
}
func (wrongTx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row { return nil }
