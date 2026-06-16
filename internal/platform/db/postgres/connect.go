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
// Signature is unchanged from the pre-OTel version.
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
