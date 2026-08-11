//go:build integration
// +build integration

package testdb

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

// runtimeRoleName is the dedicated non-owner, NOSUPERUSER + NOBYPASSRLS
// application role provisioned by the bootstrap grants stage,
// db/grants/0001_role_grants.sql (issue #88 / axis A6.1). Unlike metaldocs_ci
// (ci_role.go), which is a test-only role that deliberately keeps DML on
// audit_events, metaldocs_runtime inherits the same audit_events/
// outbox_events hardening as metaldocs_app -- it is the candidate identity
// for the request-serving/worker/jobs connection once A6.2 points compose at
// it. It is not yet wired into any compose/env default.
const runtimeRoleName = "metaldocs_runtime"

// runtimeRoleDevPassword mirrors the non-secret dev/CI fixture password baked
// into db/grants/0001_role_grants.sql. Overridable so a prod-like run can
// point at a role whose password was rotated via `ALTER ROLE metaldocs_runtime
// PASSWORD`.
const runtimeRoleDevPassword = "metaldocs_runtime_dev"

// OpenAsRuntimeRole returns a *sql.DB connected to an already-created
// per-test database (dbName, from Open()) AS the non-owner, non-bypassrls
// metaldocs_runtime role -- the A6.1 safe-identity fixture used to prove the
// boot assertion (internal/platform/db/postgres.AssertSafeIdentity) accepts a
// genuinely safe identity, not just rejects unsafe ones. Only DML-shaped
// operations should go through this handle: metaldocs_runtime holds no DDL
// rights (see the grants file comment for why).
//
// Password precedence: METALDOCS_RUNTIME_DB_PASSWORD env, else the dev
// fixture password from db/grants/0001_role_grants.sql.
func OpenAsRuntimeRole(t *testing.T, dbName string) *sql.DB {
	t.Helper()
	pw := os.Getenv("METALDOCS_RUNTIME_DB_PASSWORD")
	if pw == "" {
		pw = runtimeRoleDevPassword
	}
	return OpenAsRole(t, dbName, runtimeRoleName, pw)
}

// OpenAsRole returns a *sql.DB connected to an already-created per-test
// database (dbName, from Open()) AS an arbitrary role/password pair. It is
// the general form OpenAsRuntimeRole is built on; tests that need a role
// OTHER than metaldocs_runtime -- e.g. a throwaway role created via
// CreateThrowawayRole, so a test can mutate role ATTRIBUTES without touching
// a cluster-global fixture role every other test package also depends on --
// call this directly.
func OpenAsRole(t *testing.T, dbName, role, password string) *sql.DB {
	t.Helper()

	cfg, err := pgx.ParseConfig(DSN(t))
	if err != nil {
		t.Fatalf("parse base DSN for role %s: %v", role, err)
	}
	cfg.User = role
	cfg.Password = password
	cfg.Database = dbName

	db := stdlib.OpenDB(*cfg)
	t.Cleanup(func() { _ = db.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		t.Fatalf("ping test db %s as %s: %v", dbName, role, err)
	}
	return db
}

// CreateThrowawayRole creates a uniquely named, LOGIN role for tests that
// need to mutate role ATTRIBUTES (e.g. flip SUPERUSER live) without touching
// a cluster-global fixture role like metaldocs_runtime. Roles are
// cluster-global, not per-database: mutating metaldocs_runtime directly races
// every other test PACKAGE that asserts it is safe (Go runs test packages in
// parallel by default), and if the process dies between the mutation and its
// t.Cleanup revert, the shared fixture stays permanently unsafe for every
// later run. A throwaway role makes both failure modes structurally
// impossible -- nothing else in the cluster depends on its attributes.
//
// The role and its password are dropped via t.Cleanup.
func CreateThrowawayRole(t *testing.T, adminDB *sql.DB) (roleName, password string) {
	t.Helper()

	roleName = "metaldocs_throwaway_" + randomSuffix(t)
	password = randomSuffix(t)

	createStmt := fmt.Sprintf(
		"CREATE ROLE %s NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE NOINHERIT LOGIN PASSWORD %s",
		quoteIdent(roleName), quoteLiteral(password),
	)
	if _, err := adminDB.ExecContext(context.Background(), createStmt); err != nil {
		t.Fatalf("create throwaway role %s: %v", roleName, err)
	}
	t.Cleanup(func() {
		dropStmt := fmt.Sprintf("DROP ROLE IF EXISTS %s", quoteIdent(roleName))
		if _, err := adminDB.ExecContext(context.Background(), dropStmt); err != nil {
			t.Errorf("cleanup: drop throwaway role %s: %v", roleName, err)
		}
	})
	return roleName, password
}

// RuntimeRoleDSN returns the ambient test DSN (DSN(t)) rewritten to connect
// AS metaldocs_runtime instead of whatever ambient identity DATABASE_URL /
// METALDOCS_DATABASE_URL names -- same host/port/database, same query
// params, only the userinfo changes. Composition-root tests that call
// bootstrap.Build{API,Worker,Jobs}Dependencies (which open their own
// connection straight from env, not from an already-leased *sql.DB) use this
// to point the boot-fatal A6.1 gate (postgres.AssertSafeIdentity) at a
// genuinely safe identity via t.Setenv("METALDOCS_DATABASE_URL", ...),
// instead of asserting against the ambient identity, which is superuser +
// BYPASSRLS in dev today and is exactly the posture the gate exists to
// refuse.
func RuntimeRoleDSN(t *testing.T) string {
	t.Helper()

	u, err := url.Parse(DSN(t))
	if err != nil {
		t.Fatalf("parse base DSN for runtime role: %v", err)
	}
	pw := os.Getenv("METALDOCS_RUNTIME_DB_PASSWORD")
	if pw == "" {
		pw = runtimeRoleDevPassword
	}
	u.User = url.UserPassword(runtimeRoleName, pw)
	return u.String()
}
