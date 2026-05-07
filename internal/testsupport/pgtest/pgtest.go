// Package pgtest provides a lightweight Postgres test helper that creates an
// isolated schema per test and applies only the DDL needed for platform-level
// integration tests (currently: idempotency_keys).
//
// Tests that call [OpenAndMigrate] are skipped automatically when
// DATABASE_URL / METALDOCS_DATABASE_URL is not set.
package pgtest

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// OpenAndMigrate connects to the test database, creates a unique isolated
// schema, applies the minimal DDL required for platform integration tests, and
// returns a *sql.DB whose search_path is set to that schema. The schema is
// dropped in t.Cleanup. The test is skipped if no DB URL is configured.
func OpenAndMigrate(t *testing.T) *sql.DB {
	t.Helper()

	dsn := strings.TrimSpace(os.Getenv("METALDOCS_DATABASE_URL"))
	if dsn == "" {
		dsn = strings.TrimSpace(os.Getenv("DATABASE_URL"))
	}
	if dsn == "" {
		t.Skip("DATABASE_URL / METALDOCS_DATABASE_URL not set")
	}

	// Admin connection: create the isolated schema and apply DDL.
	adminDB, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("pgtest: sql.Open: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := adminDB.PingContext(ctx); err != nil {
		_ = adminDB.Close()
		t.Skipf("pgtest: DB unreachable: %v", err)
	}

	schema := "metaldocs_pgtest_" + randomSuffix(t)
	if _, err := adminDB.ExecContext(ctx, "CREATE SCHEMA "+quoteIdent(schema)); err != nil {
		_ = adminDB.Close()
		t.Fatalf("pgtest: create schema: %v", err)
	}
	_ = adminDB.Close()

	// Apply minimal DDL using the schema-scoped connection so all statements
	// reference the isolated schema via search_path.
	schemaDB, err := sql.Open("pgx", dsnWithSearchPath(dsn, schema))
	if err != nil {
		dropSchema(t, dsn, schema)
		t.Fatalf("pgtest: sql.Open (schema): %v", err)
	}

	if err := applyDDL(ctx, schemaDB); err != nil {
		_ = schemaDB.Close()
		dropSchema(t, dsn, schema)
		t.Fatalf("pgtest: apply DDL: %v", err)
	}

	t.Cleanup(func() {
		_ = schemaDB.Close()
		dropSchema(t, dsn, schema)
	})

	return schemaDB
}

// applyDDL creates the tables required by platform integration tests.
// Mirrors migrations/0147_idempotency_keys.sql; omits GRANTs (not needed
// in test schemas) and relaxes tenant_id to TEXT so tests can use arbitrary
// identifiers without constructing valid UUIDs.
func applyDDL(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS idempotency_keys (
			tenant_id        TEXT        NOT NULL,
			actor_user_id    TEXT        NOT NULL,
			route_template   TEXT        NOT NULL,
			key              TEXT        NOT NULL,
			payload_hash     TEXT        NOT NULL,
			response_status  INT         NOT NULL,
			response_body    TEXT        NOT NULL,
			status           TEXT        NOT NULL CHECK (status IN ('in_flight', 'completed', 'failed')),
			created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
			expires_at       TIMESTAMPTZ NOT NULL,
			PRIMARY KEY (tenant_id, actor_user_id, route_template, key)
		)`)
	return err
}

func dropSchema(t *testing.T, dsn, schema string) {
	t.Helper()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return
	}
	defer db.Close()
	_, _ = db.ExecContext(context.Background(), "DROP SCHEMA IF EXISTS "+quoteIdent(schema)+" CASCADE")
}

// dsnWithSearchPath returns the DSN with search_path set to schema so every
// pooled connection targets the isolated schema without a SET per-query.
func dsnWithSearchPath(dsn, schema string) string {
	u, err := url.Parse(dsn)
	if err != nil {
		return dsn
	}
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	return u.String()
}

func quoteIdent(v string) string {
	return `"` + strings.ReplaceAll(v, `"`, `""`) + `"`
}

func randomSuffix(t *testing.T) string {
	t.Helper()
	b := make([]byte, 5)
	if _, err := rand.Read(b); err != nil {
		t.Fatalf("pgtest: rand: %v", err)
	}
	return fmt.Sprintf("%x", b)
}
