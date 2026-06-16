package postgres_test

import (
	"context"
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"metaldocs/internal/platform/db/postgres"
)

func TestOpen_EmitsOTelSpan(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))

	dsn := "postgres://metaldocs:metaldocs@localhost:5432/metaldocs_test?sslmode=disable"
	db, err := postgres.OpenWithTracerProvider(dsn, tp)
	if err != nil {
		t.Skipf("no test DB: %v", err)
	}
	defer db.Close()

	_, err = db.QueryContext(context.Background(), "SELECT 1")
	if err != nil {
		t.Skipf("query: %v", err)
	}

	spans := sr.Ended()
	if len(spans) == 0 {
		t.Fatal("want ≥1 span from go.sql.*, got 0")
	}
	found := false
	for _, s := range spans {
		n := s.Name()
		if n == "go.sql.query" || n == "go.sql.exec" || n == "go.sql" || n == "go.sql.rows" {
			found = true
			break
		}
	}
	if !found {
		names := make([]string, len(spans))
		for i, s := range spans {
			names[i] = s.Name()
		}
		t.Fatalf("want go.sql.* span, got %v", names)
	}
}
