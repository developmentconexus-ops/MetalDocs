package scheduler

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
)

type reapedLease struct {
	jobName    string
	leaderID   string
	leaseEpoch int64
}

type leaseReaperDBState struct {
	mu sync.Mutex

	leasesToReap   []reapedLease
	governanceRows int
	deleteQueries  int
}

func (s *leaseReaperDBState) snapshotLeases() []reapedLease {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]reapedLease, len(s.leasesToReap))
	copy(out, s.leasesToReap)
	return out
}

func (s *leaseReaperDBState) recordDeleteQuery() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deleteQueries++
}

func (s *leaseReaperDBState) recordGovernanceInsert() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.governanceRows++
}

func (s *leaseReaperDBState) governanceInsertCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.governanceRows
}

func (s *leaseReaperDBState) deleteQueryCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.deleteQueries
}

type leaseReaperDriver struct {
	state *leaseReaperDBState
}

func (d *leaseReaperDriver) Open(_ string) (driver.Conn, error) {
	return &leaseReaperConn{state: d.state}, nil
}

type leaseReaperConn struct {
	state *leaseReaperDBState
}

func (c *leaseReaperConn) Prepare(query string) (driver.Stmt, error) {
	return &leaseReaperStmt{state: c.state, query: query}, nil
}

func (c *leaseReaperConn) Close() error { return nil }

func (c *leaseReaperConn) Begin() (driver.Tx, error) { return leaseReaperTx{}, nil }

func (c *leaseReaperConn) BeginTx(_ context.Context, _ driver.TxOptions) (driver.Tx, error) {
	return leaseReaperTx{}, nil
}

type leaseReaperTx struct{}

func (leaseReaperTx) Commit() error   { return nil }
func (leaseReaperTx) Rollback() error { return nil }

type leaseReaperStmt struct {
	state *leaseReaperDBState
	query string
}

func (s *leaseReaperStmt) Close() error  { return nil }
func (s *leaseReaperStmt) NumInput() int { return -1 }

func (s *leaseReaperStmt) Exec(_ []driver.Value) (driver.Result, error) {
	if strings.Contains(strings.ToLower(s.query), "insert into governance_events") {
		s.state.recordGovernanceInsert()
	}
	return leaseReaperResult(1), nil
}

func (s *leaseReaperStmt) Query(_ []driver.Value) (driver.Rows, error) {
	lower := strings.ToLower(s.query)
	if strings.Contains(lower, "delete from metaldocs.job_leases") {
		s.state.recordDeleteQuery()
		// The reaper must not join job names against tenant tables: leases
		// are system-scoped (F-19 fix removed the public.documents lookup).
		if strings.Contains(lower, "public.documents") {
			return nil, fmt.Errorf("lease reaper query must not reference public.documents")
		}
		leases := s.state.snapshotLeases()
		rows := make([][]driver.Value, 0, len(leases))
		for _, l := range leases {
			rows = append(rows, []driver.Value{l.jobName, l.leaderID, l.leaseEpoch})
		}
		return &leaseReaperRows{
			cols: []string{"job_name", "leader_id", "lease_epoch"},
			rows: rows,
		}, nil
	}
	return &leaseReaperRows{cols: []string{"ok"}, rows: [][]driver.Value{}}, nil
}

type leaseReaperRows struct {
	cols []string
	rows [][]driver.Value
	idx  int
}

func (r *leaseReaperRows) Columns() []string { return r.cols }
func (r *leaseReaperRows) Close() error      { return nil }

func (r *leaseReaperRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.idx])
	r.idx++
	return nil
}

type leaseReaperResult int64

func (r leaseReaperResult) LastInsertId() (int64, error) { return 0, nil }
func (r leaseReaperResult) RowsAffected() (int64, error) { return int64(r), nil }

func newLeaseReaperDB(t *testing.T, state *leaseReaperDBState) *sql.DB {
	t.Helper()
	name := fmt.Sprintf("lease_reaper_%p", state)
	sql.Register(name, &leaseReaperDriver{state: state})
	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("open lease reaper db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestLeaseReaper_NoExpired(t *testing.T) {
	t.Parallel()

	state := &leaseReaperDBState{}
	db := newLeaseReaperDB(t, state)

	fn := RunLeaseReaper(db)
	if err := fn(context.Background(), 1); err != nil {
		t.Fatalf("job returned error: %v", err)
	}
	if got := state.deleteQueryCount(); got != 1 {
		t.Fatalf("delete queries = %d, want 1", got)
	}
}

// TestLeaseReaper_SystemJobReapsWithoutError is the regression test for the
// F-19 JOIN bug: scheduler job names never match document ids, so the old
// tenant-attribution path errored on EVERY reap. Reaping must now succeed
// for system jobs and must not write tenant-scoped governance events.
func TestLeaseReaper_SystemJobReapsWithoutError(t *testing.T) {
	t.Parallel()

	state := &leaseReaperDBState{
		leasesToReap: []reapedLease{
			{jobName: "effective_date_publisher", leaderID: "node-1", leaseEpoch: 9},
			{jobName: "stuck_instance_watchdog", leaderID: "node-2", leaseEpoch: 4},
		},
	}
	db := newLeaseReaperDB(t, state)

	fn := RunLeaseReaper(db)
	if err := fn(context.Background(), 2); err != nil {
		t.Fatalf("job returned error: %v", err)
	}
	if got := state.governanceInsertCount(); got != 0 {
		t.Fatalf("governance inserts = %d, want 0 (leases are system-scoped)", got)
	}
}
