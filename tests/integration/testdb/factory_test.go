//go:build integration
// +build integration

package testdb

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
)

// profileCodeFormat mirrors the DB CHECK profile_code_format on the taxonomy
// code columns. Minted taxonomy codes must satisfy it.
var profileCodeFormat = regexp.MustCompile(`^[a-z][a-z0-9_-]{1,63}$`)

func rowExists(t *testing.T, db *sql.DB, query string, args ...any) bool {
	t.Helper()
	var n int
	if err := db.QueryRowContext(context.Background(), query, args...).Scan(&n); err != nil {
		t.Fatalf("rowExists query %q: %v", query, err)
	}
	return n > 0
}

func TestFactory(t *testing.T) {
	t.Run("NewTenant_seeds_tenants_row", func(t *testing.T) {
		db, _ := Open(t)
		tn := NewTenant(t, db)
		if tn.ID == "" {
			t.Fatal("Tenant.ID empty")
		}
		if !rowExists(t, db, `SELECT count(*) FROM metaldocs.tenants WHERE id=$1::uuid`, tn.ID) {
			t.Fatalf("tenants row %s not found", tn.ID)
		}
	})

	t.Run("NewUser_with_role_seeds_user_and_role", func(t *testing.T) {
		db, _ := Open(t)
		u := NewUser(t, db, WithDisplayName("Snapshot Author"), WithRole("system_admin"))
		if u.ID == "" || u.TenantID == "" {
			t.Fatalf("User missing ids: %+v", u)
		}
		if u.DisplayName != "Snapshot Author" {
			t.Fatalf("DisplayName = %q, want Snapshot Author", u.DisplayName)
		}
		if !rowExists(t, db, `SELECT count(*) FROM metaldocs.iam_users WHERE user_id=$1`, u.ID) {
			t.Fatal("iam_users row not found")
		}
		if !rowExists(t, db,
			`SELECT count(*) FROM metaldocs.iam_user_roles WHERE user_id=$1 AND role_code='system_admin'`, u.ID) {
			t.Fatal("iam_user_roles system_admin row not found")
		}
	})

	t.Run("NewTaxonomy_codes_match_format_and_seed_rows", func(t *testing.T) {
		db, _ := Open(t)
		tax := NewTaxonomy(t, db)
		for _, code := range []string{tax.FamilyCode, tax.ProcessAreaCode, tax.ProfileCode} {
			if !profileCodeFormat.MatchString(code) {
				t.Fatalf("minted code %q violates profile_code_format", code)
			}
		}
		if !rowExists(t, db, `SELECT count(*) FROM metaldocs.document_profiles WHERE tenant_id=$1::uuid AND code=$2`,
			tax.TenantID, tax.ProfileCode) {
			t.Fatal("document_profiles row not found")
		}
	})

	t.Run("NewControlledDoc_autowires_parents", func(t *testing.T) {
		db, _ := Open(t)
		cd := NewControlledDoc(t, db)
		if cd.ID == "" || cd.TenantID == "" || cd.OwnerUserID == "" || cd.ProfileCode == "" {
			t.Fatalf("ControlledDoc missing fields: %+v", cd)
		}
		if !rowExists(t, db, `SELECT count(*) FROM public.controlled_documents WHERE id=$1::uuid`, cd.ID) {
			t.Fatal("controlled_documents row not found")
		}
	})

	t.Run("NewDocument_status_and_numeric_overrides", func(t *testing.T) {
		db, _ := Open(t)
		doc := NewDocument(t, db,
			WithStatus("scheduled"),
			WithRevisionNumber(1),
			WithRevisionVersion(7),
			WithScheduleGen(2),
		)
		if doc.ID == "" || doc.ControlledDocumentID == "" {
			t.Fatalf("Document missing ids: %+v", doc)
		}
		var status string
		var revVersion int
		var gen int64
		if err := db.QueryRowContext(context.Background(),
			`SELECT status, revision_version, schedule_generation FROM public.documents WHERE id=$1::uuid`,
			doc.ID,
		).Scan(&status, &revVersion, &gen); err != nil {
			t.Fatalf("load document: %v", err)
		}
		if status != "scheduled" || revVersion != 7 || gen != 2 {
			t.Fatalf("document state = (%s,%d,%d), want (scheduled,7,2)", status, revVersion, gen)
		}
	})

	t.Run("two_calls_one_db_no_collision", func(t *testing.T) {
		db, _ := Open(t)
		// Two CDs under two tenants in ONE database — the shared-DB collision
		// case. Per-call-unique taxonomy codes must keep these independent.
		cd1 := NewControlledDoc(t, db)
		cd2 := NewControlledDoc(t, db)
		if cd1.ID == cd2.ID || cd1.ProfileCode == cd2.ProfileCode {
			t.Fatalf("expected distinct CDs/codes, got %+v / %+v", cd1, cd2)
		}
		// Two taxonomies under the SAME tenant must also not collide.
		tn := NewTenant(t, db)
		taxA := NewTaxonomy(t, db, WithTenant(tn.ID))
		taxB := NewTaxonomy(t, db, WithTenant(tn.ID))
		if taxA.ProfileCode == taxB.ProfileCode {
			t.Fatalf("expected distinct profile codes under one tenant, got %q", taxA.ProfileCode)
		}
	})

	t.Run("approval_route_and_instance_chain", func(t *testing.T) {
		db, _ := Open(t)
		doc := NewDocument(t, db, WithStatus("approved"), WithRevisionVersion(9))
		route := NewApprovalRoute(t, db, WithTenant(doc.TenantID), WithProfile(docProfile(t, db, doc)))
		inst := NewApprovalInstance(t, db, WithDocument(doc), WithRoute(route), WithStatus("approved"))
		if inst.ID == "" || inst.DocumentID != doc.ID || inst.RouteID != route.ID {
			t.Fatalf("instance not wired: %+v", inst)
		}
		if !rowExists(t, db, `SELECT count(*) FROM public.approval_instances WHERE id=$1::uuid`, inst.ID) {
			t.Fatal("approval_instances row not found")
		}
	})

	t.Run("scenario_published_document", func(t *testing.T) {
		db, _ := Open(t)
		doc := Scenario{}.PublishedDocument(t, db)
		if doc.Status != "published" {
			t.Fatalf("status = %q, want published", doc.Status)
		}
		if !rowExists(t, db, `SELECT count(*) FROM public.documents WHERE id=$1::uuid AND status='published'`, doc.ID) {
			t.Fatal("published document row not found")
		}
	})

	t.Run("scenario_scheduled_revision", func(t *testing.T) {
		db, _ := Open(t)
		doc := Scenario{}.ScheduledRevision(t, db, 3)
		if doc.Status != "scheduled" || doc.ScheduleGeneration != 3 {
			t.Fatalf("scheduled revision = (%s, gen %d), want (scheduled, 3)", doc.Status, doc.ScheduleGeneration)
		}
	})
}

// docProfile reads the controlled document's profile_code for wiring an
// approval route in the self-test (routes join documents by profile).
func docProfile(t *testing.T, db *sql.DB, doc Document) string {
	t.Helper()
	var profile string
	if err := db.QueryRowContext(context.Background(),
		`SELECT profile_code FROM public.controlled_documents WHERE id=$1::uuid`,
		doc.ControlledDocumentID,
	).Scan(&profile); err != nil {
		t.Fatalf("docProfile: %v", err)
	}
	return profile
}
