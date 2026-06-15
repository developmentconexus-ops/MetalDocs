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

	// Use factory builders to set up the test data with capability assertions handled.
	tenant := testdb.NewTenant(t, db)
	tenantID := tenant.ID

	const (
		owner   = "user-owner"
		areaMem = "user-area"
		userGr  = "user-direct"
		none    = "user-none"
		area    = "quality"
	)

	// Create users.
	testdb.NewUser(t, db, testdb.WithTenant(tenantID), testdb.WithUserID(owner))
	testdb.NewUser(t, db, testdb.WithTenant(tenantID), testdb.WithUserID(areaMem))
	testdb.NewUser(t, db, testdb.WithTenant(tenantID), testdb.WithUserID(userGr))
	testdb.NewUser(t, db, testdb.WithTenant(tenantID), testdb.WithUserID(none))

	// Create the taxonomy (will include the process area and profile).
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tenantID), testdb.WithProfile("po"), testdb.WithCode(area))

	// Create the company-scoped controlled document.
	cdCompany := testdb.NewControlledDoc(t, db,
		testdb.WithTenant(tenantID),
		testdb.WithCode("PO-CO-001"),
		testdb.WithOwner(owner),
		testdb.WithTaxonomy(tax),
		testdb.WithName("Company CD"),
		testdb.WithStatus("active"))

	// Create the restricted controlled document.
	cdRestricted := testdb.NewControlledDoc(t, db,
		testdb.WithTenant(tenantID),
		testdb.WithCode("PO-RE-001"),
		testdb.WithOwner(owner),
		testdb.WithTaxonomy(tax),
		testdb.WithName("Restricted CD"),
		testdb.WithStatus("active"))

	// Seed area grant and direct grant for restricted CD using capability assertion.
	testdb.SeedWithCaps(t, db, `[{"cap":"membership.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `INSERT INTO `+testdb.Qualified(schema, "controlled_document_area_grants")+`
			(tenant_id, controlled_document_id, area_code) VALUES ($1::uuid,$2::uuid,$3)`, tenantID, cdRestricted.ID, area)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO `+testdb.Qualified(schema, "controlled_document_user_grants")+`
			(tenant_id, controlled_document_id, user_id) VALUES ($1::uuid,$2::uuid,$3)`, tenantID, cdRestricted.ID, userGr)
		return err
	})

	// Seed the area membership for areaMem.
	testdb.SeedWithCaps(t, db, `[{"cap":"membership.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `INSERT INTO `+testdb.Qualified(schema, "user_process_areas")+`
			(user_id, tenant_id, area_code, role, effective_from) VALUES ($1,$2::uuid,$3,'qms_admin', now())`,
			areaMem, tenantID, area)
		return err
	})

	// Documents: one per CD plus a standalone doc. InsertDraftDocument handles the
	// templates/document/session/revision FK chain; we then set the visibility-
	// relevant columns (name, created_by, controlled_document_id).
	setDoc := func(suffix, createdBy, cdID string) {
		docID, _ := testdb.InsertDraftDocument(t, db, schema, tenantID)
		mustExec(t, db, ctx, `UPDATE `+testdb.Qualified(schema, "documents")+`
			SET name=$2, created_by=$3, controlled_document_id=$4::uuid WHERE id=$1::uuid`,
			docID, "Doc "+suffix, createdBy, nullableUUID(cdID))
	}
	setDoc("company", owner, cdCompany.ID)
	setDoc("restricted", owner, cdRestricted.ID)
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
