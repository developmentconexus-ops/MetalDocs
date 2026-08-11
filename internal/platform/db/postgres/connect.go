package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/XSAM/otelsql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
	"go.opentelemetry.io/otel/trace"
)

// Open opens a postgres connection using the global OTel tracer provider.
// Signature is unchanged from the pre-OTel version. It performs no identity
// check -- callers that will serve traffic or process work under the
// returned connection must use OpenServing instead. Open itself stays
// available for internal tooling that legitimately needs the bootstrap/
// owner identity (apps/dbprovision, scripts/release-backfill).
func Open(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := openDB(dsn)
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return db, nil
}

// OpenServing opens a postgres connection and enforces the A6.1 boot-fatal
// identity gate (AssertSafeIdentity) before returning it: the connected
// identity must not be SUPERUSER or BYPASSRLS, or OpenServing closes the
// connection and returns the refusal error instead of a usable *sql.DB.
//
// Every composition root that serves live traffic or processes work under
// this connection -- metaldocs-api, metaldocs-worker, metaldocs-jobs -- must
// open its runtime DB connection through OpenServing, not Open. Open+
// AssertSafeIdentity as two hand-paired calls was A6.1's original shape and
// is a convention, not a guard: nothing stops a new composition root from
// calling Open alone, forgetting the assertion, and compiling clean under a
// superuser identity. OpenServing makes that mistake require deliberately
// reaching past the safe constructor, not just forgetting a second line.
//
// Open remains available, unchanged, for internal tooling that legitimately
// needs the bootstrap/owner identity to perform DDL or role provisioning
// (apps/dbprovision/cmd/metaldocs-dbprovision, scripts/release-backfill) --
// those binaries are not serving paths and must run with elevated
// privileges by design.
func OpenServing(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := Open(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	if err := AssertSafeIdentity(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

// OpenWithTracerProvider opens a connection with an explicit TracerProvider.
// Intended for test use — callers that don't need a live DB skip the ping.
func OpenWithTracerProvider(dsn string, tp trace.TracerProvider) (*sql.DB, error) {
	return openDB(dsn, otelsql.WithTracerProvider(tp))
}

// OpenAsRole opens a dedicated *sql.DB whose every physical connection issues
// `SET ROLE role` immediately after connecting, before that connection is
// ever handed out for a query -- via pgx stdlib's AfterConnect hook, which
// runs unconditionally on every new physical connection a *sql.DB backed by
// this connector ever establishes (verified against stdlib's connector.Connect
// source: AfterConnect fires before the connection is returned, on the
// initial connect AND on every later reconnect the pool performs -- pool
// growth, a connection dropped after driver.ErrBadConn, etc.). ResetSession
// re-asserts the same SET ROLE on every pooled-connection reuse too, as a
// second, independent line of defense.
//
// This is the A6.1 "DDL runs pinned to the owner role" mechanism (issue #88).
// An earlier shape of this binary tried to prove that pinning with
// db.SetMaxOpenConns(1) plus a single `SET ROLE` statement issued once on a
// shared *sql.DB -- that looked like it bound every later call to one
// session, but did not: SET ROLE is session state, and nothing in
// database/sql's pool contract guarantees a logical *sql.DB stays on one
// physical connection across separate top-level calls. A pool reconnect
// silently drops back to the connection's original (bootstrap-superuser)
// identity, so DDL could run un-role-switched with no error and no signal.
// AfterConnect/ResetSession close that gap structurally: the role switch is
// re-applied to every connection this pool can ever hand out, so it survives
// reconnects instead of merely surviving as long as nothing forces one.
//
// Deliberately not traced via otelsql like openDB below: otelsql wraps the
// driver.Conn behind an unexported type, which would make the AfterConnect/
// ResetSession hooks unreachable (they must run beneath any tracing wrapper,
// on the real pgx connection). OpenAsRole is only ever used by
// apps/dbprovision's short-lived, single-run DDL stage, where losing spans is
// an acceptable trade for a structurally provable role pin.
//
// role must name an existing, reachable role (typically NOLOGIN, reached only
// via SET ROLE from a session already authenticated as a role permitted to
// assume it -- see db/grants/0000_identity_roles.sql for metaldocs_owner).
// dsn authenticates as whatever identity is calling OpenAsRole; that identity
// must itself be permitted to `SET ROLE role`.
func OpenAsRole(ctx context.Context, dsn string, role string) (*sql.DB, error) {
	connConfig, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn for role %q: %w", role, err)
	}

	setRoleSQL := "SET ROLE " + pgx.Identifier{role}.Sanitize()
	setRole := func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, setRoleSQL)
		return err
	}

	connector := stdlib.GetConnector(*connConfig,
		stdlib.OptionAfterConnect(setRole),
		stdlib.OptionResetSession(func(ctx context.Context, conn *pgx.Conn) error {
			return setRole(ctx, conn)
		}),
	)

	db := sql.OpenDB(connector)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres as role %q: %w", role, err)
	}

	return db, nil
}

func openDB(dsn string, extra ...otelsql.Option) (*sql.DB, error) {
	opts := []otelsql.Option{
		otelsql.WithAttributes(semconv.DBSystemNamePostgreSQL),
	}
	opts = append(opts, extra...)

	db, err := otelsql.Open("pgx", dsn, opts...)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	return db, nil
}
