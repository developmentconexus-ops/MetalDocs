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

// TestDirectInsertUserProcessAreasBlocked verifies that a write to
// user_process_areas is refused when no capability has been asserted on the
// transaction. Enforcement is the DB tripwire requiring membership.manage —
// NOT a role GRANT: metaldocs_ci (see testdb.OpenAsCIRole) holds DML on this
// table, so a privilege error would mean the tripwire never ran. Per
// CLAUDE.md authz is capabilities, never roles, and the tripwire is the last
// line; asserting SQLSTATE 42501 here would be asserting the wrong layer.
//
// Runs on the lease pool: the INSERT is inside a tx that is unconditionally
// rolled back, so even an unexpectedly-successful write is undone before any
// sibling test can observe it. A physical clone would add a per-test DROP
// DATABASE, which blocks on a forced checkpoint on this storage stack.
func TestDirectInsertUserProcessAreasBlocked(t *testing.T) {
	ctx := context.Background()
	_, dbName := testdb.Open(t)
	db := testdb.OpenAsCIRole(t, dbName)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	tenantID := testdb.DeterministicID(t, "tenant")
	userID := testdb.DeterministicID(t, "user")

	// user_process_areas (db/baseline/0001_current_schema.sql) has no
	// granted_at column — only effective_from/effective_to. effective_from
	// is NOT NULL with no default, so it must be supplied explicitly.
	_, err = tx.ExecContext(ctx, fmt.Sprintf(`
		INSERT INTO %s (tenant_id, user_id, area_code, role, granted_by, effective_from)
		VALUES ($1::uuid, $2, 'TEST', 'reviewer', 'admin', now())`,
		testdb.Qualified(dbName, "user_process_areas")),
		tenantID, userID,
	)
	if err == nil {
		t.Fatalf("direct INSERT into user_process_areas succeeded with no capability asserted; the membership.manage tripwire did not fire")
	}

	// Pinning the capability name, not just the SQLSTATE: a tripwire that fired
	// demanding some other capability would be a different invariant passing
	// under this test's name.
	if !strings.Contains(err.Error(), "P0001") ||
		!strings.Contains(err.Error(), "ErrCapabilityNotAsserted") ||
		!strings.Contains(err.Error(), "membership.manage") {
		t.Fatalf("expected the membership.manage capability tripwire (SQLSTATE P0001 ErrCapabilityNotAsserted), got: %v", err)
	}
}

// NOTE: there is no SECURITY DEFINER fn to route this write through — see
// below. The sanctioned path asserts membership.manage on the transaction.
//
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
