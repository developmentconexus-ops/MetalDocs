//go:build integration
// +build integration

package infrastructure

import (
	"context"
	"database/sql"
	"sort"
	"testing"
	"time"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/tests/integration/testdb"
)

// F3.2 CD membership-leg parity gate (M3 Category C, C1+C2; ADR-0039 D3a).
//
// CanRead (:492) and List (:150) both gate restricted-visibility on a foreign
// membership EXISTS over iam's user_process_areas. M3/F3.2 repoints that leg to
// the published view metaldocs.v_active_user_areas (which encodes effective_to
// IS NULL, ADR 0037 D1). These tests prove ZERO authz drift: the repo methods
// (post-repoint = view form) return EXACTLY what verbatim inline copies of the
// deleted raw user_process_areas SQL return — across company / restricted-area /
// restricted-area-but-REVOKED / restricted-user / owner / no-access scopes.
//
// D6: run GREEN pre-repoint (raw repo == raw baseline — sanity) then GREEN
// post-repoint (view repo == raw baseline — parity). The REVOKED-membership row
// is the drift discriminator: the active view must exclude it, exactly as
// `effective_to IS NULL` did.

type cdScenario struct {
	tenantID     string
	cdCompany    string
	cdRestricted string
	owner        string // owns both CDs
	areaMember   string // ACTIVE member of the granted area
	revokedMem   string // REVOKED member of the granted area (must NOT see restricted)
	userGrant    string // direct user-grant on the restricted CD
	none         string // no grant of any kind
	area         string
}

func seedCDVisibility(t *testing.T, db *sql.DB) cdScenario {
	t.Helper()
	ctx := context.Background()

	tenant := testdb.NewTenant(t, db)
	tenantID := tenant.ID

	owner := testdb.NewUser(t, db, testdb.WithTenant(tenantID)).ID
	areaMember := testdb.NewUser(t, db, testdb.WithTenant(tenantID)).ID
	revokedMem := testdb.NewUser(t, db, testdb.WithTenant(tenantID)).ID
	userGrant := testdb.NewUser(t, db, testdb.WithTenant(tenantID)).ID
	none := testdb.NewUser(t, db, testdb.WithTenant(tenantID)).ID

	const area = "quality"
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenantID))

	// Fixed-code process area so area-grant + membership FKs resolve.
	testdb.SeedWithCaps(t, db, `[{"cap":"taxonomy.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO metaldocs.document_process_areas (code, tenant_id, name)
			 VALUES ($1, $2::uuid, 'Quality')
			 ON CONFLICT (tenant_id, code) DO NOTHING`,
			area, tenantID,
		)
		return err
	})

	cdCompany := testdb.DeterministicID(t, "cd-company")
	cdRestricted := testdb.DeterministicID(t, "cd-restricted")
	testdb.SeedWithCaps(t, db, `[{"cap":"controlled_documents.create"}]`, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO public.controlled_documents
			   (id, tenant_id, profile_code, process_area_code, code, title, owner_user_id, visibility_scope, status)
			 VALUES ($1::uuid,$2::uuid,$3,$4,'PO-CO-001','Company CD',$5,'company','active')`,
			cdCompany, tenantID, tax.ProfileCode, area, owner,
		); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx,
			`INSERT INTO public.controlled_documents
			   (id, tenant_id, profile_code, process_area_code, code, title, owner_user_id, visibility_scope, status)
			 VALUES ($1::uuid,$2::uuid,$3,$4,'PO-RE-001','Restricted CD',$5,'restricted','active')`,
			cdRestricted, tenantID, tax.ProfileCode, area, owner,
		)
		return err
	})

	// Grants on the restricted CD (no tripwire on grant tables).
	if _, err := db.ExecContext(ctx, `INSERT INTO public.controlled_document_area_grants
		(tenant_id, controlled_document_id, area_code) VALUES ($1::uuid,$2::uuid,$3)`,
		tenantID, cdRestricted, area); err != nil {
		t.Fatalf("seed area grant: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO public.controlled_document_user_grants
		(tenant_id, controlled_document_id, user_id) VALUES ($1::uuid,$2::uuid,$3)`,
		tenantID, cdRestricted, userGrant); err != nil {
		t.Fatalf("seed user grant: %v", err)
	}

	now := time.Now().UTC()
	testdb.SeedWithCaps(t, db, `[{"cap":"membership.manage"}]`, func(tx *sql.Tx) error {
		// ACTIVE member.
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO public.user_process_areas
			   (user_id, tenant_id, area_code, role, effective_from)
			 VALUES ($1,$2::uuid,$3,'qms_admin',$4)`,
			areaMember, tenantID, area, now.Add(-2*time.Hour),
		); err != nil {
			return err
		}
		// REVOKED member (past effective_to > effective_from, revoked_by set) — must be
		// EXCLUDED by the active view, exactly as effective_to IS NULL excluded it.
		_, err := tx.ExecContext(ctx,
			`INSERT INTO public.user_process_areas
			   (user_id, tenant_id, area_code, role, effective_from, effective_to, granted_by, revoked_by)
			 VALUES ($1,$2::uuid,$3,'qms_admin',$4,$5,$1,$1)`,
			revokedMem, tenantID, area, now.Add(-3*time.Hour), now.Add(-1*time.Hour),
		)
		return err
	})

	return cdScenario{
		tenantID: tenantID, cdCompany: cdCompany, cdRestricted: cdRestricted,
		owner: owner, areaMember: areaMember, revokedMem: revokedMem,
		userGrant: userGrant, none: none, area: area,
	}
}

// rawCanRead runs the VERBATIM pre-F3.2 CanRead SQL — membership leg over
// public.user_process_areas with `effective_to IS NULL`. The post-repoint
// repo.CanRead must equal this for every (actor, cd).
func rawCanRead(t *testing.T, db *sql.DB, tenantID, cdID, actor string) bool {
	t.Helper()
	const q = `
SELECT EXISTS (
  SELECT 1
    FROM controlled_documents cd
   WHERE cd.tenant_id = $1
     AND cd.id = $2::uuid
     AND (
          cd.visibility_scope = 'company'
       OR cd.owner_user_id = $3
       OR (
            cd.visibility_scope = 'restricted'
            AND (
                 EXISTS (
                   SELECT 1
                     FROM controlled_document_area_grants cdag
                    WHERE cdag.tenant_id = cd.tenant_id
                      AND cdag.controlled_document_id = cd.id
                      AND EXISTS (
                        SELECT 1
                          FROM user_process_areas upa
                         WHERE upa.tenant_id = cd.tenant_id
                           AND upa.user_id = $3
                           AND upa.area_code = cdag.area_code
                           AND upa.effective_to IS NULL
                      )
                 )
                 OR EXISTS (
                   SELECT 1
                     FROM controlled_document_user_grants cdug
                    WHERE cdug.tenant_id = cd.tenant_id
                      AND cdug.controlled_document_id = cd.id
                      AND cdug.user_id = $3
                 )
            )
       )
     )
)`
	var exists bool
	if err := db.QueryRowContext(context.Background(), q, tenantID, cdID, actor).Scan(&exists); err != nil {
		t.Fatalf("rawCanRead: %v", err)
	}
	return exists
}

// rawListIDs runs the VERBATIM pre-F3.2 List visibility SQL (ActorUserID-only
// filter shape) — membership leg over public.user_process_areas with
// `effective_to IS NULL`. The post-repoint repo.List visible-id set must equal it.
func rawListIDs(t *testing.T, db *sql.DB, tenantID, actor string) []string {
	t.Helper()
	const q = `
SELECT id::text
FROM controlled_documents
WHERE tenant_id = $1
 AND (
      visibility_scope = 'company'
   OR owner_user_id = $2
   OR (
        visibility_scope = 'restricted'
        AND (
             EXISTS (
               SELECT 1
                 FROM controlled_document_area_grants cdag
                WHERE cdag.tenant_id = controlled_documents.tenant_id
                  AND cdag.controlled_document_id = controlled_documents.id
                  AND EXISTS (
                    SELECT 1
                      FROM user_process_areas upa
                     WHERE upa.tenant_id = controlled_documents.tenant_id
                       AND upa.user_id = $2
                       AND upa.area_code = cdag.area_code
                       AND upa.effective_to IS NULL
                  )
             )
             OR EXISTS (
               SELECT 1
                 FROM controlled_document_user_grants cdug
                WHERE cdug.tenant_id = controlled_documents.tenant_id
                  AND cdug.controlled_document_id = controlled_documents.id
                  AND cdug.user_id = $2
             )
        )
      )
 )
ORDER BY created_at DESC, id DESC`
	rows, err := db.QueryContext(context.Background(), q, tenantID, actor)
	if err != nil {
		t.Fatalf("rawListIDs: %v", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			t.Fatalf("rawListIDs scan: %v", err)
		}
		out = append(out, id)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rawListIDs rows: %v", err)
	}
	return out
}

func repoListIDs(t *testing.T, repo *PostgresControlledDocumentRepository, tenantID, actor string) []string {
	t.Helper()
	a := actor
	limit := 50
	items, _, err := repo.List(context.Background(), tenantID, controlleddocumentsdomain.CDFilter{
		ActorUserID: &a,
		Limit:       limit,
	})
	if err != nil {
		t.Fatalf("repo.List(%s): %v", actor, err)
	}
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, it.ID)
	}
	return out
}

func sortedSet(in []string) []string {
	cp := append([]string(nil), in...)
	sort.Strings(cp)
	return cp
}

func TestCanRead_ViewParityWithRaw(t *testing.T) {
	db, _ := testdb.Open(t)
	sc := seedCDVisibility(t, db)
	repo := NewPostgresControlledDocumentRepository(db, nil)

	actors := map[string]string{
		"owner": sc.owner, "areaMember": sc.areaMember, "revokedMem": sc.revokedMem,
		"userGrant": sc.userGrant, "none": sc.none,
	}
	cds := map[string]string{"company": sc.cdCompany, "restricted": sc.cdRestricted}

	for an, actor := range actors {
		for cn, cd := range cds {
			got, err := repo.CanRead(context.Background(), sc.tenantID, cd, actor)
			if err != nil {
				t.Fatalf("repo.CanRead(%s,%s): %v", an, cn, err)
			}
			want := rawCanRead(t, db, sc.tenantID, cd, actor)
			if got != want {
				t.Fatalf("CanRead drift: actor=%s cd=%s repo=%v raw=%v", an, cn, got, want)
			}
		}
	}

	// Discriminators (independent of parity): the active view must EXCLUDE the
	// revoked member from the restricted CD, and admit the active member.
	if repoCan, _ := repo.CanRead(context.Background(), sc.tenantID, sc.cdRestricted, sc.revokedMem); repoCan {
		t.Fatal("revoked member must NOT read the restricted CD (active membership only)")
	}
	if repoCan, _ := repo.CanRead(context.Background(), sc.tenantID, sc.cdRestricted, sc.areaMember); !repoCan {
		t.Fatal("active area member MUST read the restricted CD via area grant")
	}
	if repoCan, _ := repo.CanRead(context.Background(), sc.tenantID, sc.cdRestricted, sc.none); repoCan {
		t.Fatal("ungranted user must NOT read the restricted CD")
	}
}

func TestList_ViewParityWithRaw(t *testing.T) {
	db, _ := testdb.Open(t)
	sc := seedCDVisibility(t, db)
	repo := NewPostgresControlledDocumentRepository(db, nil)

	for an, actor := range map[string]string{
		"owner": sc.owner, "areaMember": sc.areaMember, "revokedMem": sc.revokedMem,
		"userGrant": sc.userGrant, "none": sc.none,
	} {
		got := sortedSet(repoListIDs(t, repo, sc.tenantID, actor))
		want := sortedSet(rawListIDs(t, db, sc.tenantID, actor))
		if len(got) != len(want) {
			t.Fatalf("List drift: actor=%s repo=%v raw=%v", an, got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("List drift: actor=%s repo=%v raw=%v", an, got, want)
			}
		}
	}

	// Discriminator: revoked member sees ONLY the company CD (not the restricted one).
	revoked := sortedSet(repoListIDs(t, repo, sc.tenantID, sc.revokedMem))
	if len(revoked) != 1 || revoked[0] != sc.cdCompany {
		t.Fatalf("revoked member must see only the company CD, saw %v", revoked)
	}
	// Active member sees both.
	active := sortedSet(repoListIDs(t, repo, sc.tenantID, sc.areaMember))
	if len(active) != 2 {
		t.Fatalf("active member must see company + restricted, saw %v", active)
	}
}
