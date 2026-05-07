// Package pgtest provides a test helper that opens a connection to the real
// Postgres test database. Tests are skipped when DATABASE_URL is not set.
// The metaldocs.idempotency_keys table is assumed to exist (from migration 0147).
package pgtest

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// OpenAndMigrate connects to the test database and returns the connection.
// The test is skipped if no DB URL is configured.
// Assumes migrations have been applied (metaldocs schema + tables exist).
func OpenAndMigrate(t *testing.T) *sql.DB {
	t.Helper()

	dsn := strings.TrimSpace(os.Getenv("METALDOCS_DATABASE_URL"))
	if dsn == "" {
		dsn = strings.TrimSpace(os.Getenv("DATABASE_URL"))
	}
	if dsn == "" {
		t.Skip("DATABASE_URL / METALDOCS_DATABASE_URL not set")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("pgtest: sql.Open: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		t.Skipf("pgtest: DB unreachable: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	return db
}
