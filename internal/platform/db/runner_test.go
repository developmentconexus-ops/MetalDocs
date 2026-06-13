package db_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"testing"

	"metaldocs/internal/platform/db"
)

// ---------------------------------------------------------------------------
// Fake driver that records begin/commit/rollback so the TxRunner lifecycle is
// observable without a real database.
// ---------------------------------------------------------------------------

type runnerFakeTx struct{ conn *runnerFakeConn }

func (t *runnerFakeTx) Commit() error {
	t.conn.committed = true
	return t.conn.commitErr
}

func (t *runnerFakeTx) Rollback() error {
	t.conn.rolledBack = true
	return nil
}

type runnerFakeConn struct {
	beginErr   error
	commitErr  error
	committed  bool
	rolledBack bool
	readOnly   bool
}

func (c *runnerFakeConn) Prepare(string) (driver.Stmt, error) { return nil, io.EOF }
func (c *runnerFakeConn) Close() error                        { return nil }
func (c *runnerFakeConn) Begin() (driver.Tx, error) {
	if c.beginErr != nil {
		return nil, c.beginErr
	}
	return &runnerFakeTx{conn: c}, nil
}

// BeginTx satisfies driver.ConnBeginTx so database/sql routes BeginTx calls
// through here rather than falling back to Begin. This allows DoReadOnly's
// ReadOnly flag to be observed and prevents the "driver does not support
// read-only transactions" error that would otherwise occur.
func (c *runnerFakeConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if c.beginErr != nil {
		return nil, c.beginErr
	}
	c.readOnly = opts.ReadOnly
	return &runnerFakeTx{conn: c}, nil
}

type runnerFakeDriver struct{ conn *runnerFakeConn }

func (d *runnerFakeDriver) Open(string) (driver.Conn, error) { return d.conn, nil }

func openRunnerFakeDB(t *testing.T, conn *runnerFakeConn) *sql.DB {
	t.Helper()
	name := fmt.Sprintf("runner_fake_%p", conn)
	sql.Register(name, &runnerFakeDriver{conn: conn})
	database, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("open runner fake db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	return database
}

func TestNewTxRunnerPanicsOnNil(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on nil *sql.DB")
		}
	}()
	db.NewTxRunner(nil)
}

func TestDoCommitsOnSuccess(t *testing.T) {
	conn := &runnerFakeConn{}
	runner := db.NewTxRunner(openRunnerFakeDB(t, conn))

	called := false
	err := runner.Do(context.Background(), func(*sql.Tx) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	if !called {
		t.Fatal("fn was not called")
	}
	if !conn.committed {
		t.Fatal("expected commit")
	}
	if conn.rolledBack {
		t.Fatal("unexpected rollback on success")
	}
}

func TestDoRollsBackAndReturnsErrUnwrapped(t *testing.T) {
	conn := &runnerFakeConn{}
	runner := db.NewTxRunner(openRunnerFakeDB(t, conn))

	sentinel := errors.New("domain failure")
	err := runner.Do(context.Background(), func(*sql.Tx) error { return sentinel })
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel returned unwrapped, got %v", err)
	}
	if !conn.rolledBack {
		t.Fatal("expected rollback on fn error")
	}
	if conn.committed {
		t.Fatal("unexpected commit on fn error")
	}
}

func TestDoRePanicsAndRollsBack(t *testing.T) {
	conn := &runnerFakeConn{}
	runner := db.NewTxRunner(openRunnerFakeDB(t, conn))

	defer func() {
		if recover() == nil {
			t.Fatal("expected re-panic")
		}
		if !conn.rolledBack {
			t.Fatal("expected rollback on panic")
		}
		if conn.committed {
			t.Fatal("unexpected commit on panic")
		}
	}()
	_ = runner.Do(context.Background(), func(*sql.Tx) error { panic("boom") })
}

func TestDoReadOnlyRePanicsAndRollsBack(t *testing.T) {
	conn := &runnerFakeConn{}
	runner := db.NewTxRunner(openRunnerFakeDB(t, conn))

	defer func() {
		if recover() == nil {
			t.Fatal("expected re-panic")
		}
		if !conn.rolledBack {
			t.Fatal("expected rollback on panic")
		}
		if conn.committed {
			t.Fatal("unexpected commit on panic")
		}
		if !conn.readOnly {
			t.Fatal("expected read-only begin")
		}
	}()
	_ = runner.DoReadOnly(context.Background(), func(*sql.Tx) error { panic("boom") })
}

func TestDoWrapsBeginError(t *testing.T) {
	beginErr := errors.New("begin failed")
	conn := &runnerFakeConn{beginErr: beginErr}
	runner := db.NewTxRunner(openRunnerFakeDB(t, conn))

	called := false
	err := runner.Do(context.Background(), func(*sql.Tx) error {
		called = true
		return nil
	})
	if !errors.Is(err, beginErr) {
		t.Fatalf("expected wrapped begin error, got %v", err)
	}
	if called {
		t.Fatal("fn must not run when begin fails")
	}
}

func TestDoWrapsCommitError(t *testing.T) {
	commitErr := errors.New("commit failed")
	conn := &runnerFakeConn{commitErr: commitErr}
	runner := db.NewTxRunner(openRunnerFakeDB(t, conn))

	err := runner.Do(context.Background(), func(*sql.Tx) error { return nil })
	if !errors.Is(err, commitErr) {
		t.Fatalf("expected wrapped commit error, got %v", err)
	}
	if !conn.committed {
		t.Fatal("expected commit attempt")
	}
}

func TestDoReadOnlyCommitsOnSuccessAndSetsReadOnly(t *testing.T) {
	conn := &runnerFakeConn{}
	runner := db.NewTxRunner(openRunnerFakeDB(t, conn))

	err := runner.DoReadOnly(context.Background(), func(*sql.Tx) error { return nil })
	if err != nil {
		t.Fatalf("DoReadOnly returned error: %v", err)
	}
	if !conn.committed {
		t.Fatal("expected commit")
	}
	if conn.rolledBack {
		t.Fatal("unexpected rollback on success")
	}
	if !conn.readOnly {
		t.Fatal("expected ReadOnly=true on the transaction options")
	}
}

func TestDoReadOnlyRollsBackAndReturnsErrUnwrapped(t *testing.T) {
	conn := &runnerFakeConn{}
	runner := db.NewTxRunner(openRunnerFakeDB(t, conn))

	sentinel := errors.New("read-only domain failure")
	err := runner.DoReadOnly(context.Background(), func(*sql.Tx) error { return sentinel })
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel returned unwrapped, got %v", err)
	}
	if !conn.rolledBack {
		t.Fatal("expected rollback on fn error")
	}
	if conn.committed {
		t.Fatal("unexpected commit on fn error")
	}
}
