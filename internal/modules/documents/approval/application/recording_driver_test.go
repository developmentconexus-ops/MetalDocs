package application

// recording_driver_test.go — minimal in-memory database/sql/driver.Conn that
// records every executed statement's query text (no argument binding, no row
// data). Shared fixture for tests that only need to assert "was this SQL
// string executed" (e.g. SQLEmitter's governance_events INSERT) — not tied to
// any single service. Extracted from the former membership_tx_test.go when
// WithMembershipContext was deleted as dead code (ARC-07 / T-011): the helper
// itself had zero non-test call sites and targeted a DB role
// (metaldocs_membership_writer) and GUC (metaldocs.verified_capability) that
// exist nowhere else in the codebase or migrations, but this generic driver
// fixture is still consumed by service_invariants_test.go.

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"testing"
)

type recordingConn struct {
	executed []string
	failAt   string
}

type noopResult struct{}

func (noopResult) LastInsertId() (int64, error) { return 0, nil }
func (noopResult) RowsAffected() (int64, error) { return 1, nil }

type emptyRows struct{}

func (emptyRows) Columns() []string         { return nil }
func (emptyRows) Close() error              { return nil }
func (emptyRows) Next([]driver.Value) error { return io.EOF }

func (c *recordingConn) Prepare(query string) (driver.Stmt, error) {
	return &recordingStmt{conn: c, query: query}, nil
}
func (c *recordingConn) Close() error              { return nil }
func (c *recordingConn) Begin() (driver.Tx, error) { return c, nil }
func (c *recordingConn) Commit() error             { return nil }
func (c *recordingConn) Rollback() error           { return nil }

type recordingStmt struct {
	conn  *recordingConn
	query string
}

func (s *recordingStmt) Close() error  { return nil }
func (s *recordingStmt) NumInput() int { return -1 }
func (s *recordingStmt) Exec(args []driver.Value) (driver.Result, error) {
	s.conn.executed = append(s.conn.executed, s.query)
	if s.conn.failAt != "" && s.query == s.conn.failAt {
		return nil, fmt.Errorf("injected error: %s", s.query)
	}
	return noopResult{}, nil
}
func (s *recordingStmt) Query(args []driver.Value) (driver.Rows, error) {
	return emptyRows{}, nil
}

type recordingDriverInstance struct{ conn *recordingConn }

func (d *recordingDriverInstance) Open(_ string) (driver.Conn, error) { return d.conn, nil }

// newTestDB registers a unique driver and returns a *sql.DB backed by conn.
// Unused by service_invariants_test.go today (it registers its own name via
// newSQLEmitterTestDB) but kept for parity with the original fixture in case
// a future test wants the one-liner.
func newTestDB(t *testing.T, conn *recordingConn) *sql.DB {
	t.Helper()
	name := fmt.Sprintf("recording_%p", conn)
	sql.Register(name, &recordingDriverInstance{conn: conn})
	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}
