package jobs_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"

	"metaldocs/internal/modules/approval/jobs"
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

// fakeRiverInserter is a recording fake for the unexported riverInserter
// dependency, letting the enqueuer be unit-tested without a real *sql.Tx or
// database (constructing a real *sql.Tx is not possible outside package
// database/sql). It captures the args/opts passed to InsertTx. Mirrors
// dispatchjobs.fakeRiverInserter.
type fakeRiverInserter struct {
	calls   int
	gotArgs river.JobArgs
	gotOpts *river.InsertOpts
	err     error
}

func (f *fakeRiverInserter) InsertTx(_ context.Context, _ *sql.Tx, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error) {
	f.calls++
	f.gotArgs = args
	f.gotOpts = opts
	if f.err != nil {
		return nil, f.err
	}
	return &rivertype.JobInsertResult{}, nil
}

func TestRiverLifecycleEventEnqueuer_RoutesToTemporalQueue(t *testing.T) {
	fake := &fakeRiverInserter{}
	enc := &jobs.RiverLifecycleEventEnqueuer{Client: fake}

	args := documentsdomain.LifecycleEventArgs{EventType: "published"}
	err := enc.EnqueueLifecycleEventTx(context.Background(), &sql.Tx{}, args)
	if err != nil {
		t.Fatalf("EnqueueLifecycleEventTx: %v", err)
	}
	if fake.calls != 1 {
		t.Fatalf("InsertTx calls = %d, want 1", fake.calls)
	}
	if fake.gotOpts == nil || fake.gotOpts.Queue != "temporal" {
		t.Errorf("InsertOpts.Queue = %+v, want temporal", fake.gotOpts)
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
