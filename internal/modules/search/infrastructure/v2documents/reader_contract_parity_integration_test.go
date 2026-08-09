//go:build integration
// +build integration

package v2documents

import (
	"context"
	"database/sql"
	"testing"

	"github.com/lib/pq"

	searchdomain "metaldocs/internal/modules/search/domain"
	taxonomyinfra "metaldocs/internal/modules/taxonomy/infrastructure"
	"metaldocs/tests/integration/testdb"
)

// F4.3 D6 parity-before-delete gate (M4 Category C4; ADR-0039 D3a). The reader is
// rewritten to read the three published views (v_document_search_facts,
// v_cd_search_facts, v_cd_grantee) INSTEAD of five foreign base tables. This test
// freezes a VERBATIM copy of the pre-M4 raw query (the base-table reads being
// deleted) and proves the rewritten reader returns the SAME ordered document set,
// per actor — including the revoked-member and ungranted-user discriminators — so
// the seam provably changes no authz/visibility/ordering. Run before AND after the
// rewrite: it equals the raw baseline both ways (raw==contract).
//
// frozenRawListDocumentsQuery is the EXACT reader.go:54-121 query as it stood before
// F4.3 (base tables public.documents / public.controlled_documents /
// controlled_document_area_grants / controlled_document_user_grants /
// user_process_areas). Kept here as the permanent parity baseline.
const frozenRawListDocumentsQuery = `
SELECT
	d.id,
	d.name,
	COALESCE(d.status, ''),
	COALESCE(d.profile_code_snapshot, ''),
	COALESCE(d.profile_code_snapshot, cd.profile_code),
	COALESCE(d.process_area_code_snapshot, ''),
	COALESCE(cd.department_code, ''),
	d.created_by::text,
	COALESCE(cd.code, d.code, ''),
	COALESCE(cd.sequence_num, d.revision_number, 0),
	d.effective_from,
	d.effective_to,
	d.created_at
FROM public.documents d
LEFT JOIN public.controlled_documents cd
  ON cd.id = d.controlled_document_id
 AND cd.tenant_id = d.tenant_id
WHERE d.tenant_id = $1
  AND d.archived_at IS NULL
  AND ($2 = '' OR LOWER(COALESCE(d.name, '')) LIKE '%' || $2 || '%' ESCAPE '\')
  AND ($3 = '' OR UPPER(COALESCE(d.status, '')) = $3)
  AND ($4 = '' OR LOWER(COALESCE(d.profile_code_snapshot, '')) = $4)
  AND ($5 = '' OR COALESCE(d.profile_code_snapshot, cd.profile_code) = ANY($14))
  AND ($6 = '' OR LOWER(COALESCE(d.process_area_code_snapshot, '')) = $6)
  AND ($7 = '' OR COALESCE(cd.department_code, '') = $7)
  AND ($8 = '' OR d.created_by::text = $8)
  AND ($9::timestamptz IS NULL OR (d.effective_to IS NOT NULL AND d.effective_to <= $9::timestamptz))
  AND ($10::timestamptz IS NULL OR (d.effective_to IS NOT NULL AND d.effective_to >= $10::timestamptz))
  AND (
    (d.controlled_document_id IS NULL AND d.created_by::text = $13)
    OR (cd.id IS NOT NULL AND (
         cd.visibility_scope = 'company'
      OR cd.owner_user_id = $13
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
                     AND upa.user_id = $13
                     AND upa.area_code = cdag.area_code
                     AND upa.effective_to IS NULL
                )
           )
           OR EXISTS (
             SELECT 1
               FROM public.controlled_document_user_grants cdug
              WHERE cdug.tenant_id = cd.tenant_id
                AND cdug.controlled_document_id = cd.id
                AND cdug.user_id = $13
           )
      ))
    ))
  )
ORDER BY d.created_at DESC, d.id DESC
LIMIT $11 OFFSET $12
`

// rawVisibleIDs runs the frozen raw query for actor and returns the ordered doc ids.
func rawVisibleIDs(t *testing.T, db *sql.DB, ctx context.Context, tenantID, actor string) []string {
	t.Helper()
	rows, err := db.QueryContext(ctx, frozenRawListDocumentsQuery,
		tenantID, "", "", "", "", "", "", "", nil, nil, 50, 0, actor, pq.Array([]string{}))
	if err != nil {
		t.Fatalf("raw query (%s): %v", actor, err)
	}
	defer rows.Close()
	holders := make([]sql.RawBytes, 13)
	dest := make([]any, 13)
	for i := range holders {
		dest[i] = &holders[i]
	}
	var ids []string
	for rows.Next() {
		if err := rows.Scan(dest...); err != nil {
			t.Fatalf("raw scan (%s): %v", actor, err)
		}
		ids = append(ids, string(holders[0]))
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("raw rows (%s): %v", actor, err)
	}
	return ids
}

// TestListDocuments_ContractParityWithFrozenRaw also guards REQ-SEARCH-1: it
// proves the rewritten reader changes "no authz/visibility/ordering" (see the
// file header) relative to the frozen pre-rewrite baseline query — the reader
// is bound to the same visibility predicate before and after, never a search-
// owned decision. See ADR 0095.
func TestListDocuments_ContractParityWithFrozenRaw(t *testing.T) {
	ctx := context.Background()
	db, schema := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tenant := testdb.NewTenant(t, db)
	tenantID := tenant.ID
	owner := testdb.NewUser(t, db, testdb.WithTenant(tenantID)).ID
	areaMem := testdb.NewUser(t, db, testdb.WithTenant(tenantID)).ID
	userGr := testdb.NewUser(t, db, testdb.WithTenant(tenantID)).ID
	revokedMem := testdb.NewUser(t, db, testdb.WithTenant(tenantID)).ID
	none := testdb.NewUser(t, db, testdb.WithTenant(tenantID)).ID

	const area = "quality"
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenantID))
	testdb.SeedWithCaps(t, db, `[{"cap":"taxonomy.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO metaldocs.document_process_areas (code, tenant_id, name)
			 VALUES ($1,$2::uuid,'Quality') ON CONFLICT (tenant_id, code) DO NOTHING`,
			area, tenantID)
		return err
	})

	cdCompany := testdb.DeterministicID(t, "cd-company")
	cdRestricted := testdb.DeterministicID(t, "cd-restricted")
	testdb.SeedWithCaps(t, db, `[{"cap":"controlled_documents.create"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO public.controlled_documents
			   (id, tenant_id, profile_code, process_area_code, code, title, owner_user_id, visibility_scope, status)
			 VALUES ($1::uuid,$2::uuid,$3,$4,'PO-CO-001','Company CD',$5,'company','active')`,
			cdCompany, tenantID, tax.ProfileCode, area, owner)
		return err
	})
	testdb.SeedWithCaps(t, db, `[{"cap":"controlled_documents.create"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO public.controlled_documents
			   (id, tenant_id, profile_code, process_area_code, code, title, owner_user_id, visibility_scope, status)
			 VALUES ($1::uuid,$2::uuid,$3,$4,'PO-RE-001','Restricted CD',$5,'restricted','active')`,
			cdRestricted, tenantID, tax.ProfileCode, area, owner)
		return err
	})

	mustExec(t, db, ctx, `INSERT INTO public.controlled_document_area_grants
		(tenant_id, controlled_document_id, area_code) VALUES ($1::uuid,$2::uuid,$3)`,
		tenantID, cdRestricted, area)
	mustExec(t, db, ctx, `INSERT INTO public.controlled_document_user_grants
		(tenant_id, controlled_document_id, user_id) VALUES ($1::uuid,$2::uuid,$3)`,
		tenantID, cdRestricted, userGr)

	// areaMem is an ACTIVE member; revokedMem is a REVOKED member (effective_to set) —
	// the anti-drift discriminator: the active grantee view must exclude it, exactly
	// as the raw `effective_to IS NULL` predicate did.
	testdb.SeedWithCaps(t, db, `[{"cap":"membership.manage"}]`, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO public.user_process_areas (user_id, tenant_id, area_code, role, effective_from)
			 VALUES ($1,$2::uuid,$3,'qms_admin', now())`, areaMem, tenantID, area); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx,
			`INSERT INTO public.user_process_areas (user_id, tenant_id, area_code, role, effective_from, effective_to, revoked_by)
			 VALUES ($1,$2::uuid,$3,'qms_admin', now() - interval '2 days', now() - interval '1 day', $4)`,
			revokedMem, tenantID, area, owner)
		return err
	})

	setDoc := func(suffix, createdBy, cdID string) {
		docID, _ := testdb.InsertDraftDocument(t, db, schema, tenantID)
		testdb.SeedWithCaps(t, db, `[{"cap":"document.edit"}]`, func(tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx,
				`UPDATE `+testdb.Qualified(schema, "documents")+`
				 SET name=$2, created_by=$3, controlled_document_id=$4::uuid WHERE id=$1::uuid`,
				docID, "Doc "+suffix, createdBy, nullableUUID(cdID))
			return err
		})
	}
	setDoc("company", owner, cdCompany)
	setDoc("restricted", owner, cdRestricted)
	setDoc("standalone", owner, "")

	reader := NewReader(db, taxonomyinfra.NewFamilyCodeResolverRepository(db))
	contractIDs := func(actor string) []string {
		docs, err := reader.ListDocuments(ctx, searchdomain.Query{TenantID: tenantID, ActorUserID: actor}, 50, 0)
		if err != nil {
			t.Fatalf("ListDocuments(%s): %v", actor, err)
		}
		ids := make([]string, 0, len(docs))
		for _, d := range docs {
			ids = append(ids, d.ID)
		}
		return ids
	}

	for _, actor := range []struct{ name, id string }{
		{"owner", owner}, {"areaMember", areaMem}, {"userGrant", userGr},
		{"revokedMem", revokedMem}, {"none", none},
	} {
		raw := rawVisibleIDs(t, db, ctx, tenantID, actor.id)
		got := contractIDs(actor.id)
		if len(raw) != len(got) {
			t.Fatalf("actor %s: contract returned %d docs, raw baseline %d (%v vs %v)",
				actor.name, len(got), len(raw), got, raw)
		}
		for i := range raw {
			if raw[i] != got[i] {
				t.Fatalf("actor %s: order/set drift at %d: contract=%v raw=%v", actor.name, i, got, raw)
			}
		}
	}
}
