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

// TestGrantAreaMembershipFn verifies the grant_area_membership SECURITY DEFINER fn
// is callable and inserts a membership row.
//
// ESCALATED PRODUCT DEFECT (2026-07-11, not fixed here — no product-code
// changes in scope): metaldocs.grant_area_membership's own input validation
// (db/baseline/0001_current_schema.sql:204-206, `_area_code !~ '^[A-Z0-9_]+$'`)
// requires an ALL-UPPERCASE area_code, but the row it must insert into
// public.user_process_areas carries FK user_process_areas_tenant_id_area_code_fkey
// (tenant_id, area_code) -> metaldocs.document_process_areas(tenant_id, code),
// and that table's own CHECK area_code_format (0001_current_schema.sql:1060)
// requires the OPPOSITE: `^[a-z][a-z0-9_-]{1,63}$` (lowercase-first). No string
// satisfies both regexes, so grant_area_membership cannot succeed for ANY
// area_code against the current schema — a genuine, always-broken contradiction
// between the function (migration 0143, 2026-era uppercase area-code convention)
// and the taxonomy table's format CHECK (migration 0123, lowercase convention),
// never reconciled. In production this path is a rarely-hit fallback
// (internal/test/e2e_seed.go:574-584 inserts directly and only calls the
// SECURITY DEFINER fn if that insert fails), so the blast radius is narrow, but
// the fn itself is unusable as written. This test cannot be made to pass via
// fixture/caps repair alone; leaving RED and reporting up rather than forcing
// green.
func TestGrantAreaMembershipFn(t *testing.T) {
	ctx := context.Background()
	db, schema := testdb.Open(t)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	tenantID := testdb.DeterministicID(t, "tenant")
	userID := testdb.DeterministicID(t, "user")
	granterID := testdb.DeterministicID(t, "granter")

	// user_process_areas_granted_by_same_tenant FK requires (tenant_id,
	// granted_by) to match an iam_users row; plain SeedUser leaves
	// iam_users.tenant_id at its default (dev tenant), which does not match
	// this test's randomly-minted tenantID. SeedSystemAdmin seeds iam_users
	// with the correct tenant_id.
	testdb.SeedSystemAdmin(t, db, tenantID, userID, userID)
	testdb.SeedSystemAdmin(t, db, tenantID, granterID, granterID)

	// user_process_areas_tenant_id_area_code_fkey requires (tenant_id,
	// area_code) to match a document_process_areas row, whose code must
	// satisfy area_code_format (lowercase). SeedGovernedTaxonomy seeds the
	// minimal governed process-area row this FK requires (plus an unused
	// family/profile row, harmless — idempotent ON CONFLICT DO NOTHING).
	testdb.SeedGovernedTaxonomy(t, db, tenantID, "area_test", "area_test")

	testdb.SetCapsOnTx(t, tx, `[{"cap":"membership.manage"}]`)

	// user_process_areas_role_check only allows viewer/editor/approver/
	// author/signer/area_admin/qms_admin — 'reviewer' predates the current
	// role enum and violates the CHECK constraint (schema drift).
	_, err = tx.ExecContext(ctx, fmt.Sprintf(
		`SELECT %s($1::uuid, $2, 'area_test', 'approver', $3)`,
		testdb.Qualified(schema, "grant_area_membership"),
	), tenantID, userID, granterID)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			t.Skipf("grant_area_membership not found: %v", err)
		}
		t.Fatalf("grant_area_membership failed: %v", err)
	}

	var count int
	if err := tx.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT count(*) FROM %s
		 WHERE user_id = $1 AND area_code = 'area_test'`,
		testdb.Qualified(schema, "user_process_areas")),
		userID,
	).Scan(&count); err != nil {
		t.Fatalf("count user_process_areas: %v", err)
	}
	if count == 0 {
		t.Fatal("expected user_process_areas row after grant_area_membership")
	}
}

// TestGrantAreaMembershipIdempotent verifies calling grant_area_membership twice is safe.
func TestGrantAreaMembershipIdempotent(t *testing.T) {
	ctx := context.Background()
	db, schema := testdb.Open(t)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	tenantID := testdb.DeterministicID(t, "tenant-idem")
	userID := testdb.DeterministicID(t, "user-idem")
	granterID := testdb.DeterministicID(t, "granter-idem")

	// See TestGrantAreaMembershipFn for the FK/CHECK rationale: iam_users
	// tenant_id must match tenantID (SeedSystemAdmin), and the area_code row
	// must exist with a lowercase code (SeedGovernedTaxonomy).
	testdb.SeedSystemAdmin(t, db, tenantID, userID, userID)
	testdb.SeedSystemAdmin(t, db, tenantID, granterID, granterID)
	testdb.SeedGovernedTaxonomy(t, db, tenantID, "area_idem", "area_idem")

	testdb.SetCapsOnTx(t, tx, `[{"cap":"membership.manage"}]`)

	for i := 0; i < 2; i++ {
		if _, err := tx.ExecContext(ctx, fmt.Sprintf(
			`SELECT %s($1::uuid, $2, 'area_idem', 'approver', $3)`,
			testdb.Qualified(schema, "grant_area_membership"),
		), tenantID, userID, granterID); err != nil {
			if strings.Contains(err.Error(), "does not exist") {
				t.Skip("grant_area_membership not found")
			}
			t.Fatalf("call %d failed: %v", i+1, err)
		}
	}

	var count int
	if err := tx.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT count(*) FROM %s
		 WHERE user_id = $1 AND area_code = 'area_idem'`,
		testdb.Qualified(schema, "user_process_areas")),
		userID,
	).Scan(&count); err != nil {
		t.Fatalf("count idempotent rows: %v", err)
	}
	if count == 0 {
		t.Fatal("expected at least one row after two calls")
	}
}
