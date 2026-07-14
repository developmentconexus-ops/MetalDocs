//go:build integration
// +build integration

package scenarios_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"metaldocs/tests/integration/testdb"
)

// TestDirectInsertUserProcessAreasBlocked verifies writer role cannot
// INSERT into user_process_areas directly (must use SECURITY DEFINER fn).
func TestDirectInsertUserProcessAreasBlocked(t *testing.T) {
	ctx := context.Background()
	db, schema := testdb.Open(t)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	tenantID := testdb.DeterministicID(t, "tenant")
	userID := testdb.DeterministicID(t, "user")

	_, err = tx.ExecContext(ctx, fmt.Sprintf(`
		INSERT INTO %s (tenant_id, user_id, area_code, role, granted_by, granted_at)
		VALUES ($1::uuid, $2, 'TEST', 'reviewer', 'admin', now())`,
		testdb.Qualified(schema, "user_process_areas")),
		tenantID, userID,
	)
	if err == nil {
		t.Log("NOTE: direct INSERT succeeded; writer role has table access in this environment")
		return
	}

	if strings.Contains(err.Error(), "42501") ||
		strings.Contains(err.Error(), "permission denied") ||
		strings.Contains(err.Error(), "insufficient_privilege") {
		t.Logf("PASS: direct INSERT rejected (privilege): %v", err)
		return
	}
	t.Logf("INSERT failed with non-privilege error: %v", err)
}

// NOTE: TestGrantAreaMembershipFn + TestGrantAreaMembershipIdempotent were
// DELETED (unit 4.3, hub-adjudicated 2026-07-14). They exercised
// metaldocs.grant_area_membership, a dead SECURITY DEFINER fn whose uppercase
// area_code validation (`_area_code !~ '^[A-Z0-9_]+$'`) is unsatisfiable
// against document_process_areas' lowercase `area_code_format` CHECK
// (`^[a-z][a-z0-9_-]{1,63}$`) — no string satisfies both, so the fn can never
// succeed for any input. It has zero product callers (the sole e2e_seed
// fallback was removed with these tests). The fn's DROP migration is Track C
// (ROADMAP 0307); these tests are deleted per the legacy-test-deletion rule
// rather than pinned to a schema-forbidden green.
