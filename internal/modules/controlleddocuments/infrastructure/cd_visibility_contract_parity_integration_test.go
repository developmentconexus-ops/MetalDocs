//go:build integration
// +build integration

package infrastructure

import (
	"context"
	"database/sql"
	"sort"
	"testing"

	"metaldocs/tests/integration/testdb"
)

// F4.1 CD search visibility + projection contract parity gate (M4 Category C4;
// ADR-0039 D3a). controlleddocuments publishes two views the search consumer
// (search/infrastructure/v2documents/reader.go) will read in F4.3 INSTEAD of CD's
// base tables (controlled_documents / controlled_document_area_grants /
// controlled_document_user_grants):
//
//   - metaldocs.v_cd_search_facts  — 1 row/CD: projection cols + is_company + owner.
//   - metaldocs.v_cd_grantee       — bounded edges: active area-grant members
//                                    (via metaldocs.v_active_user_areas) ∪ user-grants.
//
// These tests prove ZERO authz/visibility drift: the composed view decision
// (is_company OR owner=$actor OR EXISTS grantee=$actor) returns EXACTLY what a
// VERBATIM copy of the pre-M4 inline search predicate (reader.go:89-118 CD leg,
// over the raw base tables with `effective_to IS NULL`) returns — across company /
// restricted / revoked-member / user-grant / owner / no-access scopes.
//
// D6: GREEN proves view == raw baseline before F4.3 deletes the raw reads. The
// revoked-member row is the drift discriminator: the active grantee view must
// exclude it, exactly as `effective_to IS NULL` did.

// rawSearchCanSeeCD runs the VERBATIM pre-M4 search-inlined CD visibility predicate
// (reader.go:89-118 CD branch) over the raw base tables. The composed view decision
// must equal this for every (actor, cd).
func rawSearchCanSeeCD(t *testing.T, db *sql.DB, tenantID, cdID, actor string) bool {
	t.Helper()
	const q = `
SELECT EXISTS (
  SELECT 1
    FROM public.controlled_documents cd
   WHERE cd.tenant_id = $1
     AND cd.id = $2::uuid
     AND (
          cd.visibility_scope = 'company'
       OR cd.owner_user_id = $3
       OR (cd.visibility_scope = 'restricted' AND (
            EXISTS (
              SELECT 1
                FROM public.controlled_document_area_grants cdag
               WHERE cdag.tenant_id = cd.tenant_id
                 AND cdag.controlled_document_id = cd.id
                 AND EXISTS (
                   SELECT 1
                     FROM public.user_process_areas upa
                    WHERE upa.tenant_id = cd.tenant_id
                      AND upa.user_id = $3
                      AND upa.area_code = cdag.area_code
                      AND upa.effective_to IS NULL
                 )
            )
            OR EXISTS (
              SELECT 1
                FROM public.controlled_document_user_grants cdug
               WHERE cdug.tenant_id = cd.tenant_id
                 AND cdug.controlled_document_id = cd.id
                 AND cdug.user_id = $3
            )
       ))
     )
)`
	var exists bool
	if err := db.QueryRowContext(context.Background(), q, tenantID, cdID, actor).Scan(&exists); err != nil {
		t.Fatalf("rawSearchCanSeeCD: %v", err)
	}
	return exists
}

// viewSearchCanSeeCD runs the F4.1 composed decision over the published views only —
// the exact shape F4.3 will inline into the search query. NO CD base table is read.
func viewSearchCanSeeCD(t *testing.T, db *sql.DB, tenantID, cdID, actor string) bool {
	t.Helper()
	const q = `
SELECT EXISTS (
  SELECT 1
    FROM metaldocs.v_cd_search_facts f
   WHERE f.tenant_id = $1
     AND f.controlled_document_id = $2::uuid
     AND (
          f.is_company
       OR f.owner_user_id = $3
       OR EXISTS (
            SELECT 1
              FROM metaldocs.v_cd_grantee g
             WHERE g.tenant_id = f.tenant_id
               AND g.controlled_document_id = f.controlled_document_id
               AND g.grantee_user_id = $3
          )
     )
)`
	var exists bool
	if err := db.QueryRowContext(context.Background(), q, tenantID, cdID, actor).Scan(&exists); err != nil {
		t.Fatalf("viewSearchCanSeeCD: %v", err)
	}
	return exists
}

func granteeSet(t *testing.T, db *sql.DB, tenantID, cdID string) []string {
	t.Helper()
	rows, err := db.QueryContext(context.Background(),
		`SELECT grantee_user_id::text FROM metaldocs.v_cd_grantee
		  WHERE tenant_id = $1 AND controlled_document_id = $2::uuid`,
		tenantID, cdID)
	if err != nil {
		t.Fatalf("granteeSet: %v", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			t.Fatalf("granteeSet scan: %v", err)
		}
		out = append(out, id)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("granteeSet rows: %v", err)
	}
	sort.Strings(out)
	return out
}

func TestCDSearchFacts_ParityWithBaseTable(t *testing.T) {
	db, _ := testdb.Open(t)
	sc := seedCDVisibility(t, db)

	// Exactly 1 facts row per CD, projection + scalar legs == controlled_documents.
	// Columns are compared NULL-safely (department_code etc. are nullable; the view
	// passes them through verbatim — search applies its own COALESCE in F4.3).
	for name, cdID := range map[string]string{"company": sc.cdCompany, "restricted": sc.cdRestricted} {
		var (
			code, dept, profile, owner sql.NullString
			seq                        sql.NullInt64
			isCompany                  bool
		)
		err := db.QueryRowContext(context.Background(),
			`SELECT code, department_code, profile_code, sequence_num, is_company, owner_user_id::text
			   FROM metaldocs.v_cd_search_facts
			  WHERE tenant_id = $1 AND controlled_document_id = $2::uuid`,
			sc.tenantID, cdID).Scan(&code, &dept, &profile, &seq, &isCompany, &owner)
		if err != nil {
			t.Fatalf("facts row (%s): %v", name, err)
		}

		var (
			bCode, bDept, bProfile, bOwner sql.NullString
			bScope                         string
			bSeq                           sql.NullInt64
		)
		err = db.QueryRowContext(context.Background(),
			`SELECT code, department_code, profile_code, sequence_num, visibility_scope, owner_user_id::text
			   FROM public.controlled_documents
			  WHERE tenant_id = $1 AND id = $2::uuid`,
			sc.tenantID, cdID).Scan(&bCode, &bDept, &bProfile, &bSeq, &bScope, &bOwner)
		if err != nil {
			t.Fatalf("base row (%s): %v", name, err)
		}

		if code != bCode || dept != bDept || profile != bProfile || seq != bSeq || owner != bOwner {
			t.Fatalf("facts projection drift (%s): view=(%v,%v,%v,%v,%v) base=(%v,%v,%v,%v,%v)",
				name, code, dept, profile, seq, owner, bCode, bDept, bProfile, bSeq, bOwner)
		}
		if isCompany != (bScope == "company") {
			t.Fatalf("facts is_company drift (%s): view=%v base_scope=%s", name, isCompany, bScope)
		}
	}

	// Exactly 1 facts row per CD (no fan-out — projection JOIN safety).
	for name, cdID := range map[string]string{"company": sc.cdCompany, "restricted": sc.cdRestricted} {
		var n int
		if err := db.QueryRowContext(context.Background(),
			`SELECT count(*) FROM metaldocs.v_cd_search_facts WHERE tenant_id=$1 AND controlled_document_id=$2::uuid`,
			sc.tenantID, cdID).Scan(&n); err != nil {
			t.Fatalf("facts count (%s): %v", name, err)
		}
		if n != 1 {
			t.Fatalf("facts must be 1 row/CD (%s), got %d", name, n)
		}
	}
}

func TestCDGrantee_BoundedSetExcludesRevokedAndUngranted(t *testing.T) {
	db, _ := testdb.Open(t)
	sc := seedCDVisibility(t, db)

	got := granteeSet(t, db, sc.tenantID, sc.cdRestricted)
	want := []string{sc.areaMember, sc.userGrant}
	sort.Strings(want)
	if len(got) != len(want) {
		t.Fatalf("restricted grantee set drift: got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("restricted grantee set drift: got %v want %v", got, want)
		}
	}

	// Discriminators: revoked member and ungranted user must NOT be grantees.
	for _, bad := range []string{sc.revokedMem, sc.none} {
		for _, g := range got {
			if g == bad {
				t.Fatalf("grantee set must exclude %s (revoked/ungranted), saw %v", bad, got)
			}
		}
	}

	// Company CD has no grantee rows (visibility is via is_company, not edges).
	if c := granteeSet(t, db, sc.tenantID, sc.cdCompany); len(c) != 0 {
		t.Fatalf("company CD must have no grantee rows, saw %v", c)
	}
}

func TestCDVisibilityContract_ComposedDecisionParityWithRaw(t *testing.T) {
	db, _ := testdb.Open(t)
	sc := seedCDVisibility(t, db)

	actors := map[string]string{
		"owner": sc.owner, "areaMember": sc.areaMember, "revokedMem": sc.revokedMem,
		"userGrant": sc.userGrant, "none": sc.none,
	}
	cds := map[string]string{"company": sc.cdCompany, "restricted": sc.cdRestricted}

	for an, actor := range actors {
		for cn, cd := range cds {
			got := viewSearchCanSeeCD(t, db, sc.tenantID, cd, actor)
			want := rawSearchCanSeeCD(t, db, sc.tenantID, cd, actor)
			if got != want {
				t.Fatalf("composed decision drift: actor=%s cd=%s view=%v raw=%v", an, cn, got, want)
			}
		}
	}

	// Independent discriminators (not just parity): the active grantee view must
	// exclude the revoked member from the restricted CD and admit the active member.
	if viewSearchCanSeeCD(t, db, sc.tenantID, sc.cdRestricted, sc.revokedMem) {
		t.Fatal("revoked member must NOT see the restricted CD (active membership only)")
	}
	if !viewSearchCanSeeCD(t, db, sc.tenantID, sc.cdRestricted, sc.areaMember) {
		t.Fatal("active area member MUST see the restricted CD via area grant")
	}
	if viewSearchCanSeeCD(t, db, sc.tenantID, sc.cdRestricted, sc.none) {
		t.Fatal("ungranted user must NOT see the restricted CD")
	}
}
