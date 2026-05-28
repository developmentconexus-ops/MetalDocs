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

	leasesToReap    []reapedLease
	governanceRows  int
	insertedJobRefs []string
	documentTenants map[string]string
}

func (s *leaseReaperDBState) snapshotLeases() []reapedLease {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]reapedLease, len(s.leasesToReap))
	copy(out, s.leasesToReap)
	return out
}

func (s *leaseReaperDBState) recordGovernanceInsert(jobName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.governanceRows++
	s.insertedJobRefs = append(s.insertedJobRefs, jobName)
}

func (s *leaseReaperDBState) governanceInsertCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.governanceRows
}

func (s *leaseReaperDBState) insertedJobs() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, len(s.insertedJobRefs))
	copy(out, s.insertedJobRefs)
	return out
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

func (s *leaseReaperStmt) Exec(args []driver.Value) (driver.Result, error) {
	if strings.Contains(strings.ToLower(s.query), "insert into governance_events") {
		if len(args) > 0 {
			switch values := args[0].(type) {
			case string:
				s.state.recordGovernanceInsert(values)
			case []string:
				for _, jobName := range values {
					s.state.recordGovernanceInsert(jobName)
				}
			case []any:
				for _, raw := range values {
					if jobName, ok := raw.(string); ok {
						s.state.recordGovernanceInsert(jobName)
					}
				}
			}
		}
	}
	return leaseReaperResult(1), nil
}

func (s *leaseReaperStmt) Query(args []driver.Value) (driver.Rows, error) {
	lower := strings.ToLower(s.query)
	if strings.Contains(lower, "delete from metaldocs.job_leases") {
		leases := s.state.snapshotLeases()
		rows := make([][]driver.Value, 0, len(leases))
		for _, l := range leases {
			var tenantID any
			if s.state.documentTenants != nil {
				if resolved, ok := s.state.documentTenants[l.jobName]; ok {
					tenantID = resolved
				}
			}
			rows = append(rows, []driver.Value{l.jobName, l.leaderID, l.leaseEpoch, tenantID})
		}
		return &leaseReaperRows{
			cols: []string{"job_name", "leader_id", "lease_epoch", "tenant_id"},
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
	for i := range r.rows[r.idx] {
		dest[i] = r.rows[r.idx][i]
	}
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
	if got := state.governanceInsertCount(); got != 0 {
		t.Fatalf("governance inserts = %d, want 0", got)
	}
}

func TestLeaseReaper_OneExpired(t *testing.T) {
	t.Parallel()

	state := &leaseReaperDBState{
		leasesToReap: []reapedLease{
			{jobName: "effective_date_publisher", leaderID: "node-1", leaseEpoch: 9},
		},
		documentTenants: map[string]string{
			"effective_date_publisher": "11111111-1111-1111-1111-111111111111",
		},
	}
	db := newLeaseReaperDB(t, state)

	fn := RunLeaseReaper(db)
	if err := fn(context.Background(), 2); err != nil {
		t.Fatalf("job returned error: %v", err)
	}
	if got := state.governanceInsertCount(); got != 1 {
		t.Fatalf("governance inserts = %d, want 1", got)
	}
}

func TestLeaseReaper_FreshLeaseUntouched(t *testing.T) {
	t.Parallel()

	state := &leaseReaperDBState{
		leasesToReap: []reapedLease{
			{jobName: "stale_job", leaderID: "node-2", leaseEpoch: 4},
		},
		documentTenants: map[string]string{
			"stale_job": "22222222-2222-2222-2222-222222222222",
		},
	}
	db := newLeaseReaperDB(t, state)

	fn := RunLeaseReaper(db)
	if err := fn(context.Background(), 3); err != nil {
		t.Fatalf("job returned error: %v", err)
	}

	for _, job := range state.insertedJobs() {
		if job == "fresh_job" {
			t.Fatalf("fresh lease should not be touched")
		}
	}
}

func TestLeaseReaper_SkipsGovernanceInsertWhenTenantAttributionUnavailable(t *testing.T) {
	t.Parallel()

	state := &leaseReaperDBState{
		leasesToReap: []reapedLease{
			{jobName: "global_job", leaderID: "node-9", leaseEpoch: 12},
		},
		documentTenants: map[string]string{},
	}
	db := newLeaseReaperDB(t, state)

	fn := RunLeaseReaper(db)
	err := fn(context.Background(), 4)
	if err == nil {
		t.Fatal("expected tenant attribution error")
	}
	if got := state.governanceInsertCount(); got != 0 {
		t.Fatalf("governance inserts = %d, want 0", got)
	}
	if !strings.Contains(err.Error(), "tenant attribution unavailable") {
		t.Fatalf("error = %v, want tenant attribution unavailable", err)
	}
}
