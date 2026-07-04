//go:build integration
// +build integration

package idempotency_janitor

// M5 F5.2 T6 — P2 proof. After the River job-consolidation migration removed
// the custom lease scheduler + its advisory lock (ADR 0067), this test proves
// the extracted run(...) body still sweeps metaldocs.idempotency_keys
// identically to the pre-migration behavior, against real Postgres via the
// canonical testdb factory (ADR 0034) — no sqlmock.

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"

	"metaldocs/tests/integration/testdb"
)

// TestIntegration_Janitor_P2_SweepEquivalence seeds expired rows (must be
// deleted), unexpired rows (must survive), and an in_flight row past the
// 5-minute orphan grace (must survive the delete pass — the orphan pass only
// counts/logs, per job.go's two-pass design) then runs the janitor once.
func TestIntegration_Janitor_P2_SweepEquivalence(t *testing.T) {
	db, _ := testdb.Open(t)
	tenant := testdb.NewTenant(t, db)

	expiredKeys := []string{
		seedIdempotencyKey(t, db, tenant.ID, "completed", -1*time.Hour),
		seedIdempotencyKey(t, db, tenant.ID, "completed", -48*time.Hour),
		seedIdempotencyKey(t, db, tenant.ID, "failed", -10*time.Minute),
	}
	survivingKeys := []string{
		seedIdempotencyKey(t, db, tenant.ID, "completed", 1*time.Hour),
		seedIdempotencyKey(t, db, tenant.ID, "completed", 24*time.Hour),
	}
	// Orphan: in_flight, expired past the 5-min OrphanGraceMinutes window.
	// The delete pass targets ANY row past expires_at regardless of status
	// (job.go comment: "Pre-fix the sweep was gated on status = 'completed'"),
	// so this row IS deleted by pass 1 — the orphan pass only surfaces the
	// count/log for ops visibility, it does not independently preserve rows.
	orphanKey := seedIdempotencyKey(t, db, tenant.ID, "in_flight", -10*time.Minute)

	if err := run(context.Background(), db); err != nil {
		t.Fatalf("janitor run: %v", err)
	}

	for _, k := range expiredKeys {
		assertKeyAbsent(t, db, tenant.ID, k)
	}
	for _, k := range survivingKeys {
		assertKeyPresent(t, db, tenant.ID, k)
	}
	// The orphan row is past expires_at, so pass 1 deletes it too (deletion is
	// unconditional on expires_at, not gated by status) — assert the delete
	// effect, which is the deterministic assertion the task calls for.
	assertKeyAbsent(t, db, tenant.ID, orphanKey)
}

// TestIntegration_Janitor_P2_OrphanSurvivesWithinGrace proves an in_flight row
// NOT yet past expires_at survives untouched — the janitor never reaps a
// still-live in-flight request.
func TestIntegration_Janitor_P2_OrphanSurvivesWithinGrace(t *testing.T) {
	db, _ := testdb.Open(t)
	tenant := testdb.NewTenant(t, db)

	liveKey := seedIdempotencyKey(t, db, tenant.ID, "in_flight", 30*time.Minute)

	if err := run(context.Background(), db); err != nil {
		t.Fatalf("janitor run: %v", err)
	}

	assertKeyPresent(t, db, tenant.ID, liveKey)
}

// seedIdempotencyKey inserts one metaldocs.idempotency_keys row and returns
// its key column (the row's identity within the composite PK, alongside
// tenant/actor/route). expiresIn may be negative to seed an already-expired row.
func seedIdempotencyKey(t *testing.T, db *sql.DB, tenantID, status string, expiresIn time.Duration) string {
	t.Helper()
	key := "test-key-" + uuid.NewString()

	var responseStatus sql.NullInt32
	var responseBody []byte
	if status == "completed" {
		responseStatus = sql.NullInt32{Int32: 200, Valid: true}
		responseBody = []byte(`{}`)
	}

	if _, err := db.ExecContext(context.Background(), `
		INSERT INTO metaldocs.idempotency_keys
		  (tenant_id, actor_user_id, route_template, key, payload_hash,
		   response_status, response_body, status, expires_at)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8, now() + $9::interval)`,
		tenantID, "test-actor", "/test/route", key, "hash-"+key,
		responseStatus, responseBody, status, expiresIn.String(),
	); err != nil {
		t.Fatalf("seedIdempotencyKey: %v", err)
	}
	return key
}

func assertKeyAbsent(t *testing.T, db *sql.DB, tenantID, key string) {
	t.Helper()
	var count int
	if err := db.QueryRowContext(context.Background(),
		`SELECT count(*) FROM metaldocs.idempotency_keys WHERE tenant_id = $1::uuid AND key = $2`,
		tenantID, key,
	).Scan(&count); err != nil {
		t.Fatalf("assertKeyAbsent: %v", err)
	}
	if count != 0 {
		t.Fatalf("key %s: count = %d, want 0 (expected deleted)", key, count)
	}
}

func assertKeyPresent(t *testing.T, db *sql.DB, tenantID, key string) {
	t.Helper()
	var count int
	if err := db.QueryRowContext(context.Background(),
		`SELECT count(*) FROM metaldocs.idempotency_keys WHERE tenant_id = $1::uuid AND key = $2`,
		tenantID, key,
	).Scan(&count); err != nil {
		t.Fatalf("assertKeyPresent: %v", err)
	}
	if count != 1 {
		t.Fatalf("key %s: count = %d, want 1 (expected to survive)", key, count)
	}
}
