//go:build integration
// +build integration

package testdb

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

// ciRoleName is the dedicated non-owner, NOSUPERUSER + NOBYPASSRLS role
// provisioned by the bootstrap grants stage, db/grants/0001_role_grants.sql
// (re-homed there from the now-folded migration 0284_ci_rls_role.sql by the
// 2026-07-29 fold). Connecting through it makes RLS genuinely filter tenant
// reads (metaldocs_app — the owner/superuser/bypass role the rest of the
// harness uses for setup and seeding — has RLS triply inert).
const ciRoleName = "metaldocs_ci"

// ciRoleDevPassword mirrors the non-secret dev/CI fixture password baked into
// db/grants/0001_role_grants.sql. It is deliberately overridable so a prod-like
// run can point at a role whose password was rotated via
// `ALTER ROLE metaldocs_ci PASSWORD`.
const ciRoleDevPassword = "metaldocs_ci_dev"

// OpenAsCIRole returns a *sql.DB connected to an already-created per-test
// database (dbName, from Open()) AS the non-owner metaldocs_ci role. Only
// isolation-proof reads should go through this handle — all schema setup and
// row seeding must stay on the owner handle from Open(), because metaldocs_ci
// holds DML-only grants (no DDL, no ownership) and its reads are subject to the
// tenant_isolation RLS policies.
//
// The connection reuses the base DSN's host/port/dbname but overrides the role
// and password. Password precedence: METALDOCS_CI_DB_PASSWORD env, else the
// dev fixture password from migration 0284.
func OpenAsCIRole(t *testing.T, dbName string) *sql.DB {
	t.Helper()

	cfg, err := pgx.ParseConfig(DSN(t))
	if err != nil {
		t.Fatalf("parse base DSN for CI role: %v", err)
	}
	cfg.User = ciRoleName
	if pw := os.Getenv("METALDOCS_CI_DB_PASSWORD"); pw != "" {
		cfg.Password = pw
	} else {
		cfg.Password = ciRoleDevPassword
	}
	cfg.Database = dbName

	db := stdlib.OpenDB(*cfg)
	t.Cleanup(func() { _ = db.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		t.Fatalf("ping test db %s as %s: %v", dbName, ciRoleName, err)
	}
	return db
}
