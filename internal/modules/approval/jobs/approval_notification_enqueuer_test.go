package jobs

import (
	"context"
	"database/sql"
	"testing"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"

	approvaldomain "metaldocs/internal/modules/approval/domain"
)

type recordingInserter struct {
	args  river.JobArgs
	opts  *river.InsertOpts
	calls int
}

func (r *recordingInserter) InsertTx(_ context.Context, _ *sql.Tx, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error) {
	r.args = args
	r.opts = opts
	r.calls++
	return &rivertype.JobInsertResult{}, nil
}

func TestEnqueueApprovalNotificationTx_RoutesToTemporalQueue(t *testing.T) {
	rec := &recordingInserter{}
	enq := &RiverApprovalNotificationEnqueuer{Client: rec}

	err := enq.EnqueueApprovalNotificationTx(context.Background(), &sql.Tx{}, approvaldomain.ApprovalNotificationArgs{
		EventID:          "11111111-1111-1111-1111-111111111111",
		TenantID:         "22222222-2222-2222-2222-222222222222",
		EventType:        approvaldomain.EventTypeApprovalPending,
		SubjectKind:      "template",
		SubjectID:        "33333333-3333-3333-3333-333333333333",
		RecipientUserIDs: []string{"u1", "u2"},
	})
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	if rec.calls != 1 {
		t.Fatalf("InsertTx calls = %d, want 1", rec.calls)
	}
	if rec.opts == nil || rec.opts.Queue != "temporal" {
		t.Fatalf("queue = %v, want temporal — metaldocs-jobs never subscribes River's default queue", rec.opts)
	}
}

func TestEnqueueApprovalNotificationTx_RejectsNonSQLTx(t *testing.T) {
	enq := &RiverApprovalNotificationEnqueuer{Client: &recordingInserter{}}

	err := enq.EnqueueApprovalNotificationTx(context.Background(), notASQLTx{}, approvaldomain.ApprovalNotificationArgs{})
	if err == nil {
		t.Fatal("err = nil, want a type error — River needs *sql.Tx and silently doing nothing would drop the notification")
	}
}

// notASQLTx is a minimal db.Tx implementation that is NOT *sql.Tx, mirroring
// the sibling lifecycle_event_enqueuer_test.go's wrongTx.
type notASQLTx struct{}

func (notASQLTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return nil, nil
}
func (notASQLTx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return nil, nil
}
func (notASQLTx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return nil
}
