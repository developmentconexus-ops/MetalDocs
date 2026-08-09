package postgres_test

import (
	"context"
	"strings"
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"metaldocs/internal/platform/db/postgres"
)

// TestOpen_EmitsOTelSpan verifies that OpenWithTracerProvider wraps the pgx driver with
// github.com/XSAM/otelsql, causing every database/sql call to emit a span.
//
// This test uses an unreachable DSN so no real DB is required. The otelsql connector
// creates the span BEFORE calling the underlying driver's Connect — so sql.connector.connect
// is emitted even when the TCP connection is refused.
func TestOpen_EmitsOTelSpan(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))

	// Valid DSN format, unreachable address — sql.Open succeeds; Ping emits a span then fails.
	dsn := "postgres://invalid:invalid@127.0.0.1:9999/noexist?sslmode=disable&connect_timeout=1"
	db, err := postgres.OpenWithTracerProvider(dsn, tp)
	if err != nil {
		t.Fatalf("OpenWithTracerProvider returned unexpected error: %v", err)
	}
	defer func() { _ = db.Close() }()

	// PingContext triggers a connection attempt. The otelsql connector emits
	// sql.connector.connect before the underlying pgx dial — so the span is recorded
	// even when the connection is refused.
	_ = db.PingContext(context.Background())

	spans := sr.Ended()
	if len(spans) == 0 {
		t.Fatal("want ≥1 span from github.com/XSAM/otelsql, got 0 — driver wrapper may not be active")
	}

	// github.com/XSAM/otelsql emits sql.* span names (sql.connector.connect, sql.conn.ping,
	// sql.conn.query, sql.conn.exec, etc.) — NOT go.sql.* which belongs to a different library.
	found := false
	for _, s := range spans {
		if strings.HasPrefix(s.Name(), "sql.") {
			found = true
			break
		}
	}
	if !found {
		names := make([]string, len(spans))
		for i, s := range spans {
			names[i] = s.Name()
		}
		t.Fatalf("want span with sql.* prefix from github.com/XSAM/otelsql, got %v", names)
	}
}
