//go:build integration

package signature

// Live integration probe for PostgresAuthFailureRateLimiter (F-20e critical fix).
//
// Verifies that RecordFailure uses only timestamptz parameters — no time.Duration
// binding — so the UPSERT succeeds against a real Postgres instance.
//
// Run:
//   DATABASE_URL=postgres://... go test -tags integration -run TestPostgresLimiter_Live ./internal/modules/documents/approval/infrastructure/signature/
//   METALDOCS_DATABASE_URL=postgres://... go test -tags integration -run TestPostgresLimiter_Live ./internal/modules/documents/approval/infrastructure/signature/
//
// The test creates a throwaway actor UUID and deletes it in cleanup.

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func testLiveDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		dsn = strings.TrimSpace(os.Getenv("METALDOCS_DATABASE_URL"))
	}
	if dsn == "" {
		t.Skip("no DATABASE_URL or METALDOCS_DATABASE_URL set — skipping live DB probe")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("sql.Open(pgx): %v", err)
	}
	if err := db.PingContext(context.Background()); err != nil {
		t.Fatalf("db.Ping: %v", err)
	}
	return db
}

// TestPostgresLimiter_Live exercises the real code path against metaldocs-postgres.
// Asserts:
//
//	(i)   first RecordFailure inserts count=1
//	(ii)  second RecordFailure within window → count=2
//	(iii) 5 failures → Allow returns false
//	(iv)  backdated window_start (direct UPDATE) → next RecordFailure resets count=1
//	(v)   Reset deletes the row → Allow returns true
func TestPostgresLimiter_Live(t *testing.T) {
	db := testLiveDB(t)
	actor := fmt.Sprintf("integration-test-probe-%d", time.Now().UnixNano())

	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM public.auth_failure_counters WHERE actor_id = $1`, actor)
		_ = db.Close()
	})

	l := NewPostgresAuthFailureRateLimiter(db)
	ctx := context.Background()

	// (i) First failure: row should be inserted with count=1.
	if err := l.RecordFailure(ctx, actor); err != nil {
		t.Fatalf("(i) RecordFailure #1 failed: %v", err)
	}
	count := readCount(t, db, actor)
	if count != 1 {
		t.Fatalf("(i) want count=1 after first failure, got %d", count)
	}
	t.Logf("(i) first failure: count=%d ✓", count)

	// (ii) Second failure within window → count=2.
	if err := l.RecordFailure(ctx, actor); err != nil {
		t.Fatalf("(ii) RecordFailure #2 failed: %v", err)
	}
	count = readCount(t, db, actor)
	if count != 2 {
		t.Fatalf("(ii) want count=2 after second failure, got %d", count)
	}
	t.Logf("(ii) second failure: count=%d ✓", count)

	// (iii) 5 total failures → Allow returns false.
	for i := 3; i <= 5; i++ {
		if err := l.RecordFailure(ctx, actor); err != nil {
			t.Fatalf("(iii) RecordFailure #%d failed: %v", i, err)
		}
	}
	count = readCount(t, db, actor)
	if count != 5 {
		t.Fatalf("(iii) want count=5, got %d", count)
	}
	ok, err := l.Allow(ctx, actor)
	if err != nil {
		t.Fatalf("(iii) Allow error: %v", err)
	}
	if ok {
		t.Fatal("(iii) want Allow=false after 5 failures, got true")
	}
	t.Logf("(iii) 5 failures: count=%d, Allow=false ✓", count)

	// (iv) Backdate window_start so it falls outside the window.
	// Next RecordFailure must treat the window as expired and reset count=1.
	expired := time.Now().UTC().Add(-(windowDur + 5*time.Second))
	if _, err := db.ExecContext(ctx,
		`UPDATE public.auth_failure_counters SET window_start = $1 WHERE actor_id = $2`,
		expired, actor,
	); err != nil {
		t.Fatalf("(iv) backdate UPDATE failed: %v", err)
	}
	if err := l.RecordFailure(ctx, actor); err != nil {
		t.Fatalf("(iv) RecordFailure after backdate failed: %v", err)
	}
	count = readCount(t, db, actor)
	if count != 1 {
		t.Fatalf("(iv) want count=1 after expired window reset, got %d", count)
	}
	t.Logf("(iv) expired window reset: count=%d ✓", count)

	// (v) Reset deletes the row; Allow should return true.
	if err := l.Reset(ctx, actor); err != nil {
		t.Fatalf("(v) Reset failed: %v", err)
	}
	ok, err = l.Allow(ctx, actor)
	if err != nil {
		t.Fatalf("(v) Allow after Reset error: %v", err)
	}
	if !ok {
		t.Fatal("(v) want Allow=true after Reset, got false")
	}
	t.Logf("(v) Reset then Allow=true ✓")
}

func readCount(t *testing.T, db *sql.DB, actor string) int {
	t.Helper()
	var n int
	err := db.QueryRowContext(context.Background(),
		`SELECT fail_count FROM public.auth_failure_counters WHERE actor_id = $1`, actor,
	).Scan(&n)
	if err != nil {
		t.Fatalf("readCount: %v", err)
	}
	return n
}
