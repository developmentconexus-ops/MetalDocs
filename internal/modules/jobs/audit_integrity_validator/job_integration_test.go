//go:build integration
// +build integration

package audit_integrity_validator

// M5 F5.2 T6 — P3 proof. After the River job-consolidation migration removed
// the custom lease scheduler + its advisory lock (ADR 0067), this test proves
// the extracted run(...) body still detects a tampered audit hash chain
// identically to the pre-migration behavior, against real Postgres via the
// canonical testdb factory (ADR 0034) and the real auditpg.Writer — no
// sqlmock.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"metaldocs/internal/modules/audit/domain"
	auditpg "metaldocs/internal/modules/audit/infrastructure/postgres"
	"metaldocs/tests/integration/testdb"
)

// TestIntegration_AuditValidator_P3_DetectsTamperedChain writes a small clean
// chain via the real Writer.Record (production hash-chain path), then tampers
// one row's row_hash directly with SQL (simulating out-of-band corruption),
// and asserts run(ctx, validator) surfaces ErrIntegrityViolation.
func TestIntegration_AuditValidator_P3_DetectsTamperedChain(t *testing.T) {
	db, _ := testdb.Open(t)
	tenant := testdb.NewTenant(t, db)
	writer := auditpg.NewWriter(db)

	var eventIDs []string
	for i := 0; i < 3; i++ {
		id := uuid.NewString()
		eventIDs = append(eventIDs, id)
		event := domain.Event{
			ID:           id,
			OccurredAt:   time.Now().UTC(),
			ActorID:      "test-actor",
			Action:       "audit_integrity_validator.p3_test",
			ResourceType: "test_resource",
			ResourceID:   "res-" + id,
			PayloadJSON:  `{"seed":"p3"}`,
			TraceID:      "trace-" + id,
			TenantID:     tenant.ID,
		}
		if err := writer.Record(context.Background(), event); err != nil {
			t.Fatalf("seed chain event %d: %v", i, err)
		}
	}

	// Tamper the middle row's row_hash directly — an out-of-band write no
	// production code path performs, simulating DB-level corruption/tampering.
	if _, err := db.ExecContext(context.Background(),
		`UPDATE metaldocs.audit_events SET row_hash = 'tampered0000000000000000000000000000000000000000000000000000' WHERE id = $1`,
		eventIDs[1],
	); err != nil {
		t.Fatalf("tamper row: %v", err)
	}

	err := run(context.Background(), writer)
	if err == nil {
		t.Fatal("expected run() to return an error for a tampered chain, got nil")
	}
	if !errors.Is(err, ErrIntegrityViolation) {
		t.Fatalf("run() error = %v, want wrapping ErrIntegrityViolation", err)
	}
}

// TestIntegration_AuditValidator_P3_CleanChainNoIssue proves a clean,
// untampered chain produces no integrity violation.
func TestIntegration_AuditValidator_P3_CleanChainNoIssue(t *testing.T) {
	db, _ := testdb.Open(t)
	tenant := testdb.NewTenant(t, db)
	writer := auditpg.NewWriter(db)

	for i := 0; i < 3; i++ {
		id := uuid.NewString()
		event := domain.Event{
			ID:           id,
			OccurredAt:   time.Now().UTC(),
			ActorID:      "test-actor",
			Action:       "audit_integrity_validator.p3_clean_test",
			ResourceType: "test_resource",
			ResourceID:   "res-" + id,
			PayloadJSON:  `{"seed":"p3-clean"}`,
			TraceID:      "trace-" + id,
			TenantID:     tenant.ID,
		}
		if err := writer.Record(context.Background(), event); err != nil {
			t.Fatalf("seed chain event %d: %v", i, err)
		}
	}

	if err := run(context.Background(), writer); err != nil {
		t.Fatalf("run() on clean chain returned error: %v", err)
	}
}
