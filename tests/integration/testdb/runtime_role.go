//go:build integration
// +build integration

package testdb

import (
	"context"
	"database/sql"
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

	cfg, err := pgx.ParseConfig(DSN(t))
	if err != nil {
		t.Fatalf("parse base DSN for runtime role: %v", err)
	}
	cfg.User = runtimeRoleName
	if pw := os.Getenv("METALDOCS_RUNTIME_DB_PASSWORD"); pw != "" {
		cfg.Password = pw
	} else {
		cfg.Password = runtimeRoleDevPassword
	}
	cfg.Database = dbName

	db := stdlib.OpenDB(*cfg)
	t.Cleanup(func() { _ = db.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		t.Fatalf("ping test db %s as %s: %v", dbName, runtimeRoleName, err)
	}
	return db
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
