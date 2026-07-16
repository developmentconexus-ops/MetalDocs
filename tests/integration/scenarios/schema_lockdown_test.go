//go:build integration
// +build integration

package scenarios_test

import (
	"context"
	"strings"
	"testing"

	"metaldocs/tests/integration/testdb"
)

// These tests connect as metaldocs_ci (NOSUPERUSER, NOBYPASSRLS, DML-only
// grants — see testdb.OpenAsCIRole). metaldocs_app, used everywhere else in
// this package, is superuser + BYPASSRLS and cannot exercise a privilege
// denial at all.
//
// They run on the lease pool, not a physical clone, even though they issue
// DDL: Postgres DDL is transactional, and every statement below runs inside a
// tx that is unconditionally rolled back. A statement that unexpectedly
// SUCCEEDS is therefore still undone before any sibling test observes the
// database — the rollback contains the blast radius, so a clone would buy no
// isolation the tx does not already provide. It would cost one forced
// checkpoint per test: OpenFreshDatabase DROPs its database in t.Cleanup, and
// on this storage stack DROP DATABASE blocks on IPC:CheckpointDone for an
// unbounded time.

func TestWriterCannotDropTable(t *testing.T) {
	ctx := context.Background()
	_, dbName := testdb.Open(t)
	db := testdb.OpenAsCIRole(t, dbName)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `DROP TABLE public.documents`)
	require42501(t, err, "DROP TABLE public.documents")
}

func TestWriterCannotAlterTable(t *testing.T) {
	ctx := context.Background()
	_, dbName := testdb.Open(t)
	db := testdb.OpenAsCIRole(t, dbName)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `ALTER TABLE public.documents ADD COLUMN _test_col text`)
	require42501(t, err, "ALTER TABLE public.documents")
}

func TestWriterCannotCreateTable(t *testing.T) {
	ctx := context.Background()
	_, dbName := testdb.Open(t)
	db := testdb.OpenAsCIRole(t, dbName)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `CREATE TABLE metaldocs._test_lockdown (id serial)`)
	require42501(t, err, "CREATE TABLE metaldocs._test_lockdown")
}

// TestWriterCanReadApprovalTables is the complement of the three DDL-denial
// tests above: it runs through the SAME metaldocs_ci connection, proving
// reads are granted to the very role that DDL is denied to (not just that
// some unrelated superuser handle can read).
func TestWriterCanReadApprovalTables(t *testing.T) {
	ctx := context.Background()
	_, dbName := testdb.Open(t)
	db := testdb.OpenAsCIRole(t, dbName)

	if _, err := db.ExecContext(ctx, `SELECT 1 FROM public.approval_instances LIMIT 0`); err != nil {
		t.Fatalf("read approval_instances should succeed: %v", err)
	}
}

// require42501 asserts the operation was denied on privilege grounds
// (SQLSTATE 42501 / "permission denied"). Any other outcome — including the
// statement succeeding, or failing for an unrelated reason such as a
// dependency error (2BP01) — is a test failure, not a pass: it means the
// assertion did not actually observe the privilege boundary it claims to.
func require42501(t *testing.T, err error, op string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s unexpectedly succeeded; metaldocs_ci should be denied by privilege", op)
	}
	msg := err.Error()
	if strings.Contains(msg, "42501") || strings.Contains(strings.ToLower(msg), "permission denied") {
		return
	}
	t.Fatalf("%s failed for a non-privilege reason (expected SQLSTATE 42501/permission denied): %v", op, err)
}
