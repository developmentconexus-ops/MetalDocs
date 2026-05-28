package application

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"
)

func TestLegacyCodeFromDocumentID(t *testing.T) {
	t.Run("valid uuid-like id", func(t *testing.T) {
		got, err := legacyCodeFromDocumentID("2af9e3dd-48e2-4aef-8ecc-57f34eb35de5")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "MIG-2AF9E3DD" {
			t.Fatalf("legacy code = %q, want MIG-2AF9E3DD", got)
		}
	})

	t.Run("short id returns error", func(t *testing.T) {
		_, err := legacyCodeFromDocumentID("abc")
		if err == nil {
			t.Fatalf("expected error for short id, got nil")
		}
	})
}

type backfillDriver struct {
	state *backfillDriverState
}

type backfillDriverState struct {
	lockConnID   int
	unlockConnID int
	nextConnID   int
}

func (d *backfillDriver) Open(_ string) (driver.Conn, error) {
	d.state.nextConnID++
	return &backfillConn{id: d.state.nextConnID, state: d.state}, nil
}

type backfillConn struct {
	id    int
	state *backfillDriverState
}

func (c *backfillConn) Prepare(query string) (driver.Stmt, error) {
	return &backfillStmt{connID: c.id, state: c.state, query: query}, nil
}

func (c *backfillConn) Close() error { return nil }

func (c *backfillConn) Begin() (driver.Tx, error) { return backfillTx{}, nil }

func (c *backfillConn) BeginTx(_ context.Context, _ driver.TxOptions) (driver.Tx, error) {
	return backfillTx{}, nil
}

type backfillTx struct{}

func (backfillTx) Commit() error   { return nil }
func (backfillTx) Rollback() error { return nil }

type backfillStmt struct {
	connID int
	state  *backfillDriverState
	query  string
}

func (s *backfillStmt) Close() error  { return nil }
func (s *backfillStmt) NumInput() int { return -1 }

func (s *backfillStmt) Exec(_ []driver.Value) (driver.Result, error) {
	if strings.Contains(strings.ToLower(s.query), "pg_advisory_unlock") {
		s.state.unlockConnID = s.connID
	}
	return backfillResult(1), nil
}

func (s *backfillStmt) Query(_ []driver.Value) (driver.Rows, error) {
	lower := strings.ToLower(s.query)
	switch {
	case strings.Contains(lower, "pg_try_advisory_lock"):
		s.state.lockConnID = s.connID
		return &backfillRows{cols: []string{"locked"}, rows: [][]driver.Value{{true}}}, nil
	case strings.Contains(lower, "from documents"):
		return &backfillRows{cols: []string{"id", "tenant_id"}, rows: [][]driver.Value{}}, nil
	default:
		return &backfillRows{cols: []string{"ok"}, rows: [][]driver.Value{}}, nil
	}
}

type backfillRows struct {
	cols []string
	rows [][]driver.Value
	idx  int
}

func (r *backfillRows) Columns() []string { return r.cols }
func (r *backfillRows) Close() error      { return nil }

func (r *backfillRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.rows) {
		return io.EOF
	}
	for i := range r.rows[r.idx] {
		dest[i] = r.rows[r.idx][i]
	}
	r.idx++
	return nil
}

type backfillResult int64

func (r backfillResult) LastInsertId() (int64, error) { return 0, nil }
func (r backfillResult) RowsAffected() (int64, error) { return int64(r), nil }

func TestBackfillLegacyDocuments_UnlocksUsingSameSession(t *testing.T) {
	t.Parallel()

	state := &backfillDriverState{}
	name := fmt.Sprintf("backfill_driver_%p", state)
	sql.Register(name, &backfillDriver{state: state})

	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	if err := BackfillLegacyDocuments(context.Background(), db, slog.New(slog.NewTextHandler(io.Discard, nil))); err != nil {
		t.Fatalf("BackfillLegacyDocuments: %v", err)
	}
	if state.lockConnID == 0 || state.unlockConnID == 0 {
		t.Fatalf("lockConnID=%d unlockConnID=%d, want both recorded", state.lockConnID, state.unlockConnID)
	}
	if state.lockConnID != state.unlockConnID {
		t.Fatalf("lock conn = %d, unlock conn = %d, want same session", state.lockConnID, state.unlockConnID)
	}
}
