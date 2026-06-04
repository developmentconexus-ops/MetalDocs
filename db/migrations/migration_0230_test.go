package migrations_test

import (
	"os"
	"strings"
	"testing"
)

// TestMigration0230DecommissionsReviewerRole locks the ADR 0022 reviewer
// decommission: the migration must remove reviewer memberships, tighten the
// user_process_areas role CHECK to the 7 canonical area roles (no reviewer), and
// drop the duplicated role-list validation from the grant/revoke stored procs so
// the table CHECK is the single source of truth.
func TestMigration0230DecommissionsReviewerRole(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile("0230_authz_decommission_reviewer_role.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	sql := string(body)

	if !strings.Contains(sql, "DELETE FROM public.user_process_areas WHERE role = 'reviewer'") {
		t.Fatal("migration must remove existing reviewer memberships before tightening the CHECK")
	}
	// The table is append + soft-delete only (trg_user_process_areas_no_delete) and
	// capability-gated (trg_require_cap_asserted); both must be disabled around the
	// one-time purge and re-enabled, else the DELETE aborts on any DB with reviewer rows.
	for _, frag := range []string{
		"DISABLE TRIGGER trg_user_process_areas_no_delete",
		"DISABLE TRIGGER trg_require_cap_asserted",
		"ENABLE TRIGGER trg_user_process_areas_no_delete",
		"ENABLE TRIGGER trg_require_cap_asserted",
	} {
		if !strings.Contains(sql, frag) {
			t.Fatalf("migration must %q so the reviewer purge can run on the protected table", frag)
		}
	}
	if !strings.Contains(sql, "DROP CONSTRAINT IF EXISTS user_process_areas_role_check") ||
		!strings.Contains(sql, "ADD CONSTRAINT user_process_areas_role_check") {
		t.Fatal("migration must drop and re-add the user_process_areas role CHECK")
	}
	// The new CHECK must NOT contain reviewer and MUST contain the 7 canonical area roles.
	checkStart := strings.Index(sql, "ADD CONSTRAINT user_process_areas_role_check")
	if checkStart < 0 {
		t.Fatal("missing ADD CONSTRAINT clause")
	}
	checkClause := sql[checkStart : checkStart+400]
	if strings.Contains(checkClause, "reviewer") {
		t.Fatal("the tightened CHECK must NOT allow 'reviewer'")
	}
	for _, role := range []string{"viewer", "editor", "approver", "author", "signer", "area_admin", "qms_admin"} {
		if !strings.Contains(checkClause, "'"+role+"'") {
			t.Fatalf("the tightened CHECK must allow canonical area role %q", role)
		}
	}

	// Both stored procs must be replaced WITHOUT the duplicated role-list validation.
	if !strings.Contains(sql, "CREATE OR REPLACE FUNCTION metaldocs.grant_area_membership") ||
		!strings.Contains(sql, "CREATE OR REPLACE FUNCTION metaldocs.revoke_area_membership") {
		t.Fatal("migration must replace both grant and revoke stored procs")
	}
	if strings.Contains(sql, "_role NOT IN ('viewer', 'editor', 'reviewer', 'approver')") {
		t.Fatal("migration must DROP the stale duplicated role-list validation from the stored procs")
	}

	if !strings.Contains(sql, "schema_migrations") || !strings.Contains(sql, "'0230'") {
		t.Fatal("migration must record schema_migrations version 0230")
	}
}
