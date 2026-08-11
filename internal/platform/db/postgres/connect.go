package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/XSAM/otelsql"
	_ "github.com/jackc/pgx/v5/stdlib"
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
