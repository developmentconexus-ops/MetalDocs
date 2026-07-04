package stuck_instance_watchdog

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"

	"metaldocs/internal/modules/documents/approval/application"
	"metaldocs/internal/modules/iam/authz"
	platformdb "metaldocs/internal/platform/db"
)

type recordingEmitter struct {
	mu     sync.Mutex
	events []application.GovernanceEvent
}

func (r *recordingEmitter) Emit(_ context.Context, _ platformdb.Tx, e application.GovernanceEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, e)
	return nil
}

func (r *recordingEmitter) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.events)
}

func (r *recordingEmitter) first() application.GovernanceEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.events) == 0 {
		return application.GovernanceEvent{}
	}
	return r.events[0]
}

type watchdogDBState struct {
	mu         sync.Mutex
	stuckRows  []StuckInstance
	setConfigN int
}

type watchdogDriver struct {
	state *watchdogDBState
}

func (d *watchdogDriver) Open(_ string) (driver.Conn, error) {
	return &watchdogConn{state: d.state}, nil
}

type watchdogConn struct {
	state *watchdogDBState
}

func (c *watchdogConn) Prepare(query string) (driver.Stmt, error) {
	return &watchdogStmt{state: c.state, query: query}, nil
}

func (c *watchdogConn) Close() error { return nil }

func (c *watchdogConn) Begin() (driver.Tx, error) { return watchdogTx{}, nil }

func (c *watchdogConn) BeginTx(_ context.Context, _ driver.TxOptions) (driver.Tx, error) {
	return watchdogTx{}, nil
}

type watchdogTx struct{}

func (watchdogTx) Commit() error   { return nil }
func (watchdogTx) Rollback() error { return nil }

type watchdogStmt struct {
	state *watchdogDBState
	query string
}

func (s *watchdogStmt) Close() error  { return nil }
func (s *watchdogStmt) NumInput() int { return -1 }

func (s *watchdogStmt) Exec(_ []driver.Value) (driver.Result, error) {
	if strings.Contains(strings.ToLower(s.query), "set_config('metaldocs.bypass_authz'") {
		s.state.mu.Lock()
		s.state.setConfigN++
		s.state.mu.Unlock()
	}
	return watchdogResult(1), nil
}

func (s *watchdogStmt) Query(_ []driver.Value) (driver.Rows, error) {
	q := strings.ToLower(s.query)
	if strings.Contains(q, "from approval_instances") {
		s.state.mu.Lock()
		rows := append([]StuckInstance(nil), s.state.stuckRows...)
		s.state.mu.Unlock()
		out := make([][]driver.Value, 0, len(rows))
		for _, row := range rows {
			out = append(out, []driver.Value{
				row.ID,
				row.TenantID,
				row.DocumentID,
				row.SubmittedBy,
				row.DriftPolicy,
			})
		}
		return &watchdogRows{
			cols: []string{"id", "tenant_id", "document_id", "submitted_by", "drift_policy"},
			rows: out,
		}, nil
	}
	return &watchdogRows{cols: []string{"ok"}, rows: [][]driver.Value{}}, nil
}

type watchdogRows struct {
	cols []string
	rows [][]driver.Value
	idx  int
}

func (r *watchdogRows) Columns() []string { return r.cols }
func (r *watchdogRows) Close() error      { return nil }

func (r *watchdogRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.idx])
	r.idx++
	return nil
}

type watchdogResult int64

func (r watchdogResult) LastInsertId() (int64, error) { return 0, nil }
func (r watchdogResult) RowsAffected() (int64, error) { return int64(r), nil }

func newWatchdogDB(t *testing.T, state *watchdogDBState) *sql.DB {
	t.Helper()
	name := fmt.Sprintf("watchdog_%p", state)
	sql.Register(name, &watchdogDriver{state: state})
	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("open watchdog db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestWatchdog_NoStuck(t *testing.T) {
	t.Parallel()

	state := &watchdogDBState{}
	db := newWatchdogDB(t, state)
	emitter := &recordingEmitter{}

	if err := run(authz.WithBackgroundBypass(context.Background()), db, emitter); err != nil {
		t.Fatalf("job returned error: %v", err)
	}

	if got := emitter.count(); got != 0 {
		t.Fatalf("alerts emitted = %d, want 0", got)
	}
}

// TestListStuckInstances_UsesStageSnapshotDriftPolicy guards the read source
// for drift_policy: the watchdog still reads (and reports, in the alert
// payload) the active stage snapshot's on_eligibility_drift_snapshot column —
// this is unrelated to the removed auto_cancel branch (ADR 0068), it is the
// informational field surfaced in every stuck_alert event.
func TestListStuckInstances_UsesStageSnapshotDriftPolicy(t *testing.T) {
	src, err := os.ReadFile("job.go")
	if err != nil {
		t.Fatalf("read job.go: %v", err)
	}

	body := string(src)
	if strings.Contains(body, "ar.on_eligibility_drift") {
		t.Fatal("watchdog must not read on_eligibility_drift from approval_routes; the column belongs to route stages/stage snapshots")
	}
	if !strings.Contains(body, "asi.on_eligibility_drift_snapshot") {
		t.Fatal("watchdog must read drift policy from active approval_stage_instances snapshot")
	}
}

func TestWatchdog_AlertOnly(t *testing.T) {
	t.Parallel()

	state := &watchdogDBState{
		stuckRows: []StuckInstance{
			{ID: "inst-1", TenantID: "tenant-1", DocumentID: "doc-1", SubmittedBy: "u1", DriftPolicy: "reduce_quorum"},
		},
	}
	db := newWatchdogDB(t, state)
	emitter := &recordingEmitter{}

	if err := run(authz.WithBackgroundBypass(context.Background()), db, emitter); err != nil {
		t.Fatalf("job returned error: %v", err)
	}

	if got := emitter.count(); got != 1 {
		t.Fatalf("alerts emitted = %d, want 1", got)
	}
	ev := emitter.first()
	if ev.EventType != "approval.instance.stuck_alert" {
		t.Fatalf("event type = %q, want approval.instance.stuck_alert", ev.EventType)
	}
	if ev.ResourceType != "approval_instance" {
		t.Fatalf("resource type = %q, want approval_instance", ev.ResourceType)
	}
	if ev.ActorUserID != SystemActor {
		t.Fatalf("actor_user_id = %q, want %q", ev.ActorUserID, SystemActor)
	}
}

// TestWatchdog_AlertOnly_MultipleInstances proves every stuck instance gets
// exactly one alert regardless of its drift_policy value — there is no branch
// in run() that special-cases any policy string anymore (ADR 0068).
func TestWatchdog_AlertOnly_MultipleInstances(t *testing.T) {
	t.Parallel()

	state := &watchdogDBState{
		stuckRows: []StuckInstance{
			{ID: "inst-1", TenantID: "tenant-1", DocumentID: "doc-1", SubmittedBy: "u1", DriftPolicy: "reduce_quorum"},
			{ID: "inst-2", TenantID: "tenant-1", DocumentID: "doc-2", SubmittedBy: "u2", DriftPolicy: "keep_snapshot"},
			{ID: "inst-3", TenantID: "tenant-2", DocumentID: "doc-3", SubmittedBy: "u3", DriftPolicy: "fail_stage"},
		},
	}
	db := newWatchdogDB(t, state)
	emitter := &recordingEmitter{}

	if err := run(authz.WithBackgroundBypass(context.Background()), db, emitter); err != nil {
		t.Fatalf("job returned error: %v", err)
	}

	if got := emitter.count(); got != 3 {
		t.Fatalf("alerts emitted = %d, want 3", got)
	}
	for _, ev := range []application.GovernanceEvent{emitter.events[0], emitter.events[1], emitter.events[2]} {
		if ev.EventType != "approval.instance.stuck_alert" {
			t.Fatalf("event type = %q, want approval.instance.stuck_alert", ev.EventType)
		}
		if ev.ActorUserID != SystemActor {
			t.Fatalf("actor_user_id = %q, want %q", ev.ActorUserID, SystemActor)
		}
	}
}
