//go:build integration
// +build integration

package infrastructure

import (
	"context"
	"database/sql"
	"testing"

	"metaldocs/tests/integration/testdb"
)

// F2.1a — obligated-reader read contract.
//
// metaldocs.v_cd_obligated_readers must return, for every (tenant_id, cd):
//   - one row per active user_grant            (source='user_grant',    area_code=NULL)
//   - one row per active area_grant member     (source='area_grant',    area_code=<area>)
//   - one row per active tenant user, for each company-scope CD
//                                              (source='company_scope', area_code=NULL)
//
// DISTINCT BY (tenant_id, cd, user_id) with source precedence
// user_grant > area_grant > company_scope. Revoked area members MUST be excluded.
// v_cd_grantee MUST NOT be modified.

type obligatedRow struct {
	UserID   string
	AreaCode sql.NullString
	Source   string
}

func obligatedSet(t *testing.T, db *sql.DB, tenantID, cdID string) []obligatedRow {
	t.Helper()
	rows, err := db.QueryContext(context.Background(),
		`SELECT user_id::text, area_code, source
		   FROM metaldocs.v_cd_obligated_readers
		  WHERE tenant_id = $1 AND controlled_document_id = $2::uuid
		  ORDER BY user_id, source`,
		tenantID, cdID)
	if err != nil {
		t.Fatalf("obligatedSet query: %v", err)
	}
	defer rows.Close()
	var out []obligatedRow
	for rows.Next() {
		var r obligatedRow
		if err := rows.Scan(&r.UserID, &r.AreaCode, &r.Source); err != nil {
			t.Fatalf("obligatedSet scan: %v", err)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("obligatedSet rows: %v", err)
	}
	return out
}

// seedObligatedExtraOverlap adds one user that is BOTH user-granted AND an active
// area member on the restricted CD — the source-precedence discriminator.
func seedObligatedExtraOverlap(t *testing.T, db *sql.DB, sc cdScenario) string {
	t.Helper()
	ctx := context.Background()

	overlap := testdb.NewUser(t, db, testdb.WithTenant(sc.tenantID)).ID

	if _, err := db.ExecContext(ctx,
		`INSERT INTO public.controlled_document_user_grants
		   (tenant_id, controlled_document_id, user_id)
		 VALUES ($1::uuid,$2::uuid,$3)`,
		sc.tenantID, sc.cdRestricted, overlap); err != nil {
		t.Fatalf("seed overlap user grant: %v", err)
	}

	testdb.SeedWithCaps(t, db, `[{"cap":"membership.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO public.user_process_areas
			   (user_id, tenant_id, area_code, role, effective_from)
			 VALUES ($1,$2::uuid,$3,'qms_admin', now() - interval '2 hours')`,
			overlap, sc.tenantID, sc.area,
		)
		return err
	})
	return overlap
}

func TestObligatedReaders_RestrictedCD_ThreeLegsWithPrecedence(t *testing.T) {
	db, _ := testdb.Open(t)
	sc := seedCDVisibility(t, db)
	overlap := seedObligatedExtraOverlap(t, db, sc)

	got := obligatedSet(t, db, sc.tenantID, sc.cdRestricted)

	want := map[string]obligatedRow{
		sc.areaMember: {UserID: sc.areaMember, AreaCode: sql.NullString{String: sc.area, Valid: true}, Source: "area_grant"},
		sc.userGrant:  {UserID: sc.userGrant, AreaCode: sql.NullString{}, Source: "user_grant"},
		overlap:       {UserID: overlap, AreaCode: sql.NullString{}, Source: "user_grant"},
	}

	if len(got) != len(want) {
		t.Fatalf("restricted obligated set size: got %d want %d (rows=%+v)", len(got), len(want), got)
	}
	for _, r := range got {
		w, ok := want[r.UserID]
		if !ok {
			t.Fatalf("unexpected obligated row: %+v", r)
		}
		if r.Source != w.Source {
			t.Fatalf("source drift for user %s: got %q want %q", r.UserID, r.Source, w.Source)
		}
		if r.AreaCode.Valid != w.AreaCode.Valid || r.AreaCode.String != w.AreaCode.String {
			t.Fatalf("area_code drift for user %s: got %+v want %+v", r.UserID, r.AreaCode, w.AreaCode)
		}
	}

	for _, bad := range []string{sc.revokedMem, sc.owner, sc.none} {
		for _, r := range got {
			if r.UserID == bad {
				t.Fatalf("user %s must NOT be in restricted obligated set, got %+v", bad, r)
			}
		}
	}
}

func TestObligatedReaders_CompanyCD_AllActiveTenantUsers(t *testing.T) {
	db, _ := testdb.Open(t)
	sc := seedCDVisibility(t, db)

	// Anchor the expected company-scope cardinality to the fixture's actual active-user
	// count (read from the upstream contract metaldocs.v_active_user_areas) so a future
	// seedCDVisibility change that adds active members surfaces as a setup-drift error
	// rather than a silent set-size mismatch.
	var activeUsers int
	if err := db.QueryRowContext(context.Background(),
		`SELECT count(DISTINCT user_id) FROM metaldocs.v_active_user_areas WHERE tenant_id = $1`,
		sc.tenantID).Scan(&activeUsers); err != nil {
		t.Fatalf("active users count: %v", err)
	}
	if activeUsers != 1 {
		t.Fatalf("fixture drift: seedCDVisibility now produces %d active tenant users; this test was authored against exactly 1 (areaMember) — re-read the fixture before asserting", activeUsers)
	}

	got := obligatedSet(t, db, sc.tenantID, sc.cdCompany)

	if len(got) != activeUsers {
		t.Fatalf("company obligated set size: got %d want %d (rows=%+v)", len(got), activeUsers, got)
	}
	r := got[0]
	if r.UserID != sc.areaMember {
		t.Fatalf("company obligated user: got %s want %s", r.UserID, sc.areaMember)
	}
	if r.Source != "company_scope" {
		t.Fatalf("company source: got %q want %q", r.Source, "company_scope")
	}
	if r.AreaCode.Valid {
		t.Fatalf("company area_code: got %v want NULL", r.AreaCode)
	}
}

func TestObligatedReaders_ViewShape(t *testing.T) {
	db, _ := testdb.Open(t)
	const q = `
	  SELECT column_name, data_type, is_nullable
	    FROM information_schema.columns
	   WHERE table_schema = 'metaldocs'
	     AND table_name   = 'v_cd_obligated_readers'
	   ORDER BY ordinal_position`
	rows, err := db.QueryContext(context.Background(), q)
	if err != nil {
		t.Fatalf("view shape query: %v", err)
	}
	defer rows.Close()
	type col struct{ name, typ, nullable string }
	var got []col
	for rows.Next() {
		var c col
		if err := rows.Scan(&c.name, &c.typ, &c.nullable); err != nil {
			t.Fatalf("scan: %v", err)
		}
		got = append(got, c)
	}
	// PostgreSQL marks every column of a UNION ALL-derived view as is_nullable=YES
	// in information_schema (conservative default — view metadata cannot prove a
	// UNION branch never yields NULL). Runtime NOT NULL is enforced upstream by
	// the base-table constraints on controlled_document_user_grants /
	// controlled_document_area_grants / v_active_user_areas / v_cd_search_facts,
	// not by this view's column metadata. The contract that matters here is
	// the column NAME, TYPE, and ORDER — those are real drift signals.
	want := []col{
		{"tenant_id", "uuid", "YES"},
		{"controlled_document_id", "uuid", "YES"},
		{"user_id", "text", "YES"},
		{"area_code", "text", "YES"},
		{"source", "text", "YES"},
	}
	if len(got) != len(want) {
		t.Fatalf("column count: got %d want %d (got=%+v)", len(got), len(want), got)
	}
	// Compare in ordinal_position order (no sort) — the declared column order is part of
	// the published contract; reordering would surface as a real drift, not silent re-pass.
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("column %d drift: got %+v want %+v", i, got[i], want[i])
		}
	}
}
