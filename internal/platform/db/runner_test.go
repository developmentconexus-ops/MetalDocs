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
	platformtenant "metaldocs/internal/platform/tenant"
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

// recordedExec captures one ExecContext call issued against the fake
// connection, so tests can assert whether (and with what args) the chokepoint
// seed statement was issued.
type recordedExec struct {
	query string
	args  []driver.NamedValue
}

type runnerFakeConn struct {
	beginErr   error
	commitErr  error
	committed  bool
	rolledBack bool
	execs      []recordedExec
	execResult driver.Result
	execErr    error
}

func (c *runnerFakeConn) Prepare(string) (driver.Stmt, error) { return nil, io.EOF }
func (c *runnerFakeConn) Close() error                        { return nil }
func (c *runnerFakeConn) Begin() (driver.Tx, error) {
	if c.beginErr != nil {
		return nil, c.beginErr
	}
	return &runnerFakeTx{conn: c}, nil
}

// ExecContext satisfies driver.ExecerContext so database/sql routes
// tx.ExecContext calls (the seed statement) through here instead of failing
// with "driver does not support ExecerContext".
func (c *runnerFakeConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	c.execs = append(c.execs, recordedExec{query: query, args: args})
	if c.execErr != nil {
		return nil, c.execErr
	}
	if c.execResult != nil {
		return c.execResult, nil
	}
	return driver.RowsAffected(0), nil
}

// BeginTx satisfies driver.ConnBeginTx so database/sql routes BeginTx calls
// through here rather than falling back to Begin.
func (c *runnerFakeConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if c.beginErr != nil {
		return nil, c.beginErr
	}
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

// ---------------------------------------------------------------------------
// F3.1 T3 — PG-1 positive behavior proof: the chokepoint auto-seeds
// metaldocs.tenant_id + metaldocs.actor_id (tx-local set_config) when BOTH
// are present in ctx, and is a no-op (no seed statement) when either is
// absent.
// ---------------------------------------------------------------------------

func identityCtx(tenantID, actorID string) context.Context {
	ctx := context.Background()
	if tenantID != "" {
		ctx = platformtenant.WithTenantID(ctx, tenantID)
	}
	if actorID != "" {
		ctx = platformtenant.WithActorID(ctx, actorID)
	}
	return ctx
}

func seedExecCount(conn *runnerFakeConn) int {
	n := 0
	for _, e := range conn.execs {
		if containsSeedCall(e.query) {
			n++
		}
	}
	return n
}

func containsSeedCall(query string) bool {
	return contains(query, "set_config") && contains(query, "metaldocs.tenant_id") && contains(query, "metaldocs.actor_id")
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})()
}

func TestDo_SeedsIdentity_WhenBothTenantAndActorPresent(t *testing.T) {
	conn := &runnerFakeConn{}
	runner := db.NewTxRunner(openRunnerFakeDB(t, conn))

	ctx := identityCtx("tenant-x", "actor-y")
	err := runner.Do(ctx, func(*sql.Tx) error { return nil })
	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	if got := seedExecCount(conn); got != 1 {
		t.Fatalf("expected exactly 1 seed exec, got %d (execs=%+v)", got, conn.execs)
	}
	got := conn.execs[0]
	if len(got.args) != 2 {
		t.Fatalf("expected 2 seed args, got %d: %+v", len(got.args), got.args)
	}
	if got.args[0].Value != "tenant-x" {
		t.Errorf("seed arg[0] (tenant) = %v, want tenant-x", got.args[0].Value)
	}
	if got.args[1].Value != "actor-y" {
		t.Errorf("seed arg[1] (actor) = %v, want actor-y", got.args[1].Value)
	}
}

func TestDo_NoSeed_WhenIdentityAbsent(t *testing.T) {
	conn := &runnerFakeConn{}
	runner := db.NewTxRunner(openRunnerFakeDB(t, conn))

	err := runner.Do(context.Background(), func(*sql.Tx) error { return nil })
	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	if got := seedExecCount(conn); got != 0 {
		t.Fatalf("expected no seed exec for identity-less ctx, got %d (execs=%+v)", got, conn.execs)
	}
}

func TestDo_NoSeed_WhenOnlyTenantPresent(t *testing.T) {
	conn := &runnerFakeConn{}
	runner := db.NewTxRunner(openRunnerFakeDB(t, conn))

	ctx := identityCtx("tenant-only", "")
	err := runner.Do(ctx, func(*sql.Tx) error { return nil })
	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	if got := seedExecCount(conn); got != 0 {
		t.Fatalf("expected no seed exec when actor absent, got %d", got)
	}
}

func TestDo_NoSeed_WhenOnlyActorPresent(t *testing.T) {
	conn := &runnerFakeConn{}
	runner := db.NewTxRunner(openRunnerFakeDB(t, conn))

	ctx := identityCtx("", "actor-only")
	err := runner.Do(ctx, func(*sql.Tx) error { return nil })
	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	if got := seedExecCount(conn); got != 0 {
		t.Fatalf("expected no seed exec when tenant absent, got %d", got)
	}
}
