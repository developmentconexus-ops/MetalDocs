//go:build integration
// +build integration

package v2documents

import (
	"context"
	"database/sql"
	"testing"

	searchdomain "metaldocs/internal/modules/search/domain"
	"metaldocs/tests/integration/testdb"
)

// TestListDocuments_EnforcesUnifiedVisibility proves the v2 search reader returns
// a document only when the caller can see it per the unified model (AD-3):
// company-scoped → everyone; restricted → owner / area-grant / direct user-grant;
// standalone (no controlled document) → creator only. No system_admin bypass.
func TestListDocuments_EnforcesUnifiedVisibility(t *testing.T) {
	ctx := context.Background()
	db, schema := testdb.Open(t)
	db.SetMaxOpenConns(1)

	// Seed tenant + users via factory (no tripwire on tenants or iam_users).
	tenant := testdb.NewTenant(t, db)
	tenantID := tenant.ID

	ownerUser := testdb.NewUser(t, db, testdb.WithTenant(tenantID))
	owner := ownerUser.ID

	areaMember := testdb.NewUser(t, db, testdb.WithTenant(tenantID))
	areaMem := areaMember.ID

	directUser := testdb.NewUser(t, db, testdb.WithTenant(tenantID))
	userGr := directUser.ID

	noneUser := testdb.NewUser(t, db, testdb.WithTenant(tenantID))
	none := noneUser.ID

	const area = "quality"

	// Seed taxonomy (document_families / document_process_areas / document_profiles)
	// via factory — all three carry the taxonomy.manage tripwire; factory wraps them.
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenantID))

	// Seed the process area code we need for grants (area = "quality"). The
	// taxonomy factory minted a random code; seed a separate process-area row with
	// the fixed "quality" code so the area-grant and membership FK resolve.
	testdb.SeedWithCaps(t, db, `[{"cap":"taxonomy.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO metaldocs.document_process_areas (code, tenant_id, name)
			 VALUES ($1, $2::uuid, 'Quality')
			 ON CONFLICT (tenant_id, code) DO NOTHING`,
			area, tenantID,
		)
		return err
	})

	// Seed two controlled documents with visibility_scope via SeedWithCaps
	// (controlled_documents INSERT requires controlled_documents.create).
	cdCompany := testdb.DeterministicID(t, "cd-company")
	cdRestricted := testdb.DeterministicID(t, "cd-restricted")

	testdb.SeedWithCaps(t, db, `[{"cap":"controlled_documents.create"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO public.controlled_documents
			   (id, tenant_id, profile_code, process_area_code, code, title, owner_user_id, visibility_scope, status)
			 VALUES ($1::uuid,$2::uuid,$3,$4,'PO-CO-001','Company CD',$5,'company','active')`,
			cdCompany, tenantID, tax.ProfileCode, area, owner,
		)
		return err
	})
	testdb.SeedWithCaps(t, db, `[{"cap":"controlled_documents.create"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO public.controlled_documents
			   (id, tenant_id, profile_code, process_area_code, code, title, owner_user_id, visibility_scope, status)
			 VALUES ($1::uuid,$2::uuid,$3,$4,'PO-RE-001','Restricted CD',$5,'restricted','active')`,
			cdRestricted, tenantID, tax.ProfileCode, area, owner,
		)
		return err
	})

	// Area + user grants on the restricted CD — no tripwire on these tables.
	mustExec(t, db, ctx, `INSERT INTO public.controlled_document_area_grants
		(tenant_id, controlled_document_id, area_code) VALUES ($1::uuid,$2::uuid,$3)`,
		tenantID, cdRestricted, area)
	mustExec(t, db, ctx, `INSERT INTO public.controlled_document_user_grants
		(tenant_id, controlled_document_id, user_id) VALUES ($1::uuid,$2::uuid,$3)`,
		tenantID, cdRestricted, userGr)

	// areaMem is an active member of the area — user_process_areas carries
	// the membership.manage tripwire.
	testdb.SeedWithCaps(t, db, `[{"cap":"membership.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO public.user_process_areas
			   (user_id, tenant_id, area_code, role, effective_from)
			 VALUES ($1,$2::uuid,$3,'qms_admin', now())`,
			areaMem, tenantID, area,
		)
		return err
	})

	// Documents — InsertDraftDocument handles the template/session/revision chain
	// (already factory-safe). The UPDATE to wire name/created_by/controlled_document_id
	// touches documents (UPDATE tripwire: document.edit), so wrap in SeedWithCaps.
	setDoc := func(suffix, createdBy, cdID string) {
		docID, _ := testdb.InsertDraftDocument(t, db, schema, tenantID)
		testdb.SeedWithCaps(t, db, `[{"cap":"document.edit"}]`, func(tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx,
				`UPDATE `+testdb.Qualified(schema, "documents")+`
				 SET name=$2, created_by=$3, controlled_document_id=$4::uuid WHERE id=$1::uuid`,
				docID, "Doc "+suffix, createdBy, nullableUUID(cdID),
			)
			return err
		})
	}
	setDoc("company", owner, cdCompany)
	setDoc("restricted", owner, cdRestricted)
	setDoc("standalone", owner, "")

	reader := NewReader(db)
	visibleTitles := func(actor string) map[string]bool {
		docs, err := reader.ListDocuments(ctx, searchdomain.Query{TenantID: tenantID, ActorUserID: actor}, 50, 0)
		if err != nil {
			t.Fatalf("ListDocuments(%s): %v", actor, err)
		}
		out := map[string]bool{}
		for _, d := range docs {
			out[d.Title] = true
		}
		return out
	}

	cases := []struct {
		actor string
		want  map[string]bool
	}{
		// no grants → only the company-wide doc.
		{none, map[string]bool{"Doc company": true}},
		// area member → company + restricted (via area grant), not the standalone.
		{areaMem, map[string]bool{"Doc company": true, "Doc restricted": true}},
		// direct user grant → company + restricted.
		{userGr, map[string]bool{"Doc company": true, "Doc restricted": true}},
		// owner → company + restricted (owner) + standalone (creator).
		{owner, map[string]bool{"Doc company": true, "Doc restricted": true, "Doc standalone": true}},
	}
	for _, tc := range cases {
		got := visibleTitles(tc.actor)
		if len(got) != len(tc.want) {
			t.Fatalf("actor %s saw %v, want exactly %v", tc.actor, keys(got), keys(tc.want))
		}
		for title := range tc.want {
			if !got[title] {
				t.Fatalf("actor %s did not see %q (saw %v)", tc.actor, title, keys(got))
			}
		}
	}

	// The ungranted user must never see the restricted document.
	if visibleTitles(none)["Doc restricted"] {
		t.Fatal("ungranted user must not see the restricted document")
	}
}

func mustExec(t *testing.T, db *sql.DB, ctx context.Context, q string, args ...any) {
	t.Helper()
	if _, err := db.ExecContext(ctx, q, args...); err != nil {
		t.Fatalf("exec %q: %v", q, err)
	}
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func nullableUUID(id string) any {
	if id == "" {
		return nil
	}
	return id
}
