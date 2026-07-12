package tripwire

import (
	"os"
	"path/filepath"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

// TestTripwireArms_CapsAreRegistryReal pins TRIPWIRE-ARM-PARITY sub-clause 1:
// every capability referenced by TripwireArms must be a real, registered
// iam capability (validation-contract.md §1.5.a). A cap typo or a stale
// constant here would otherwise silently generate a migration arm that can
// never be satisfied (or worse, is satisfied by the wrong cap).
func TestTripwireArms_CapsAreRegistryReal(t *testing.T) {
	for _, arm := range TripwireArms {
		for _, cap := range arm.Caps {
			if !iamdomain.IsValidCapability(cap) {
				t.Errorf("TripwireArms[%s,%s]: capability %q is not a registered iam capability (IsValidCapability=false)", arm.Table, arm.Op, cap)
			}
		}
	}
}

// TestTripwireArms_MatchesContractTable pins the 18-entry table in M2
// validation-contract.md §1.2 exactly, as extended by M6 validation-contract.md
// §3 (documents/UPDATE additionally gains document.review), M7
// validation-contract.md §5 (F7.2, ADR 0070: +1 tenants/INSERT arm gated on
// tenant.onboard — 19 entries), M7 §5 (F7.3, ADR 0070: +1
// tenant_lifecycle_jobs/INSERT arm gated on tenant.export OR tenant.erase — 20
// entries), ADR 0083 / M3 P3.S2b-3b-0 (approval_instances/INSERT arm #1
// splits into two subject-discriminated entries — 21 entries), and ADR 0083's
// follow-on / M3 P3.S2b-3b-iii-a (approval_signoffs/INSERT arm #2 splits into
// two parent-lookup-discriminated entries — 22 entries). Divergence is HS-7.
// The key includes WhenValue so the discriminated entries (same table+op,
// different subject) are distinct rows rather than colliding.
func TestTripwireArms_MatchesContractTable(t *testing.T) {
	type key struct {
		table     string
		op        Op
		whenValue string
	}

	want := map[key][]iamdomain.Capability{
		// ADR 0083: approval_instances/INSERT is subject-discriminated —
		// document rows require exactly document.submit, template rows
		// require exactly template.submit, never unioned.
		{"approval_instances", OpInsert, "document"}: {iamdomain.CapDocumentSubmit},
		{"approval_instances", OpInsert, "template"}: {iamdomain.CapTemplateSubmit},
		// ADR 0083 follow-on: approval_signoffs/INSERT is
		// parent-lookup-discriminated — parent subject_kind='document'
		// requires exactly document.signoff, parent subject_kind='template'
		// requires exactly template.approve, never unioned.
		{"approval_signoffs", OpInsert, "document"}: {iamdomain.CapDocumentSignoff},
		{"approval_signoffs", OpInsert, "template"}: {iamdomain.CapTemplateApprove},
		{"iam_user_roles", OpAny, ""}:                {iamdomain.CapUserManage},
		{"user_process_areas", OpAny, ""}:            {iamdomain.CapMembershipManage},
		{"documents", OpInsert, ""}:                  {iamdomain.CapDocumentCreate},
		{"documents", OpUpdate, ""}:                  {iamdomain.CapDocumentEdit, iamdomain.CapDocumentObsolete, iamdomain.CapMembershipManage, iamdomain.CapDocumentReview},
		{"controlled_documents", OpInsert, ""}:       {iamdomain.CapControlledDocumentCreate},
		{"controlled_documents", OpUpdate, ""}:       {iamdomain.CapControlledDocumentObsolete, iamdomain.CapControlledDocumentSupersede},
		{"cd_sequence_counters", OpAny, ""}:          {iamdomain.CapControlledDocumentCreate},
		{"document_profiles", OpAny, ""}:             {iamdomain.CapTaxonomyManage},
		{"document_process_areas", OpAny, ""}:        {iamdomain.CapTaxonomyManage},
		{"document_families", OpAny, ""}:             {iamdomain.CapTaxonomyManage},
		{"templates_template", OpAny, ""}: {
			iamdomain.CapTemplateCreate, iamdomain.CapTemplateEdit, iamdomain.CapTemplateSubmit,
			iamdomain.CapTemplateApprove, iamdomain.CapTemplatePublish, iamdomain.CapTemplateArchive,
		},
		{"templates_template_version", OpAny, ""}: {
			iamdomain.CapTemplateCreate, iamdomain.CapTemplateEdit, iamdomain.CapTemplateSubmit,
			iamdomain.CapTemplateReview, iamdomain.CapTemplateApprove, iamdomain.CapTemplatePublish,
		},
		{"iam_users", OpAny, ""}:         {iamdomain.CapUserManage},
		{"iam_groups", OpAny, ""}:        {iamdomain.CapUserManage},
		{"iam_group_members", OpAny, ""}: {iamdomain.CapUserManage},
		{"iam_group_roles", OpAny, ""}:   {iamdomain.CapUserManage},
		// M7 F7.2 (ADR 0070, M7 validation-contract.md §5 touchpoint 6):
		// onboarding INSERTs metaldocs.tenants under tenant.onboard.
		{"tenants", OpInsert, ""}: {iamdomain.CapTenantOnboard},
		// M7 F7.3 (ADR 0070, M7 validation-contract.md §5 touchpoint 6):
		// export/erase enqueue INSERTs metaldocs.tenant_lifecycle_jobs under
		// tenant.export OR tenant.erase (match-one).
		{"tenant_lifecycle_jobs", OpInsert, ""}: {iamdomain.CapTenantExport, iamdomain.CapTenantErase},
	}

	if len(TripwireArms) != len(want) {
		t.Fatalf("TripwireArms has %d entries, contract §1.2 requires %d", len(TripwireArms), len(want))
	}

	seen := make(map[key]bool, len(TripwireArms))
	for _, arm := range TripwireArms {
		k := key{arm.Table, arm.Op, arm.WhenValue}
		if seen[k] {
			t.Errorf("duplicate TripwireArms entry for (%s, %s, whenValue=%q)", arm.Table, arm.Op, arm.WhenValue)
		}
		seen[k] = true

		wantCaps, ok := want[k]
		if !ok {
			t.Errorf("TripwireArms has unexpected entry (%s, %s, whenValue=%q) not present in contract §1.2", arm.Table, arm.Op, arm.WhenValue)
			continue
		}
		if len(arm.Caps) != len(wantCaps) {
			t.Errorf("(%s,%s,%q): got %d caps %v, want %d caps %v", arm.Table, arm.Op, arm.WhenValue, len(arm.Caps), arm.Caps, len(wantCaps), wantCaps)
			continue
		}
		for i, c := range arm.Caps {
			if c != wantCaps[i] {
				t.Errorf("(%s,%s,%q) cap[%d]: got %q, want %q", arm.Table, arm.Op, arm.WhenValue, i, c, wantCaps[i])
			}
		}
	}
	for k := range want {
		if !seen[k] {
			t.Errorf("contract §1.2 entry (%s, %s) missing from TripwireArms", k.table, k.op)
		}
	}
}

// TestRenderMigration_MatchesCommittedFile is the golden test: RenderMigration()
// must byte-equal the latest committed tripwire migration. M2 pinned 0271; M6
// F6.2 re-rendered it to 0275; M7 F7.2 re-rendered it to 0277 (tenants/INSERT
// arm); M7 F7.3 re-rendered it to 0279 (tenant_lifecycle_jobs/INSERT arm +
// one-time attachment, ADR 0070), then to db/migrations/0283_*.sql (DELETE
// paths RETURN OLD — RETURN NEW on BEFORE DELETE silently cancelled DELETEs),
// then to db/migrations/0299_*.sql (ADR 0083, M3 P3.S2b-3b-0:
// approval_instances/INSERT arm subject-discriminated into a nested CASE
// NEW.subject_kind; numbered 0299 to avoid the 0284_ci_rls_role prefix
// collision — migrate.Apply dedupes by 4-digit prefix only), then to
// db/migrations/0300_*.sql (ADR 0083 follow-on, M3 P3.S2b-3b-iii-a:
// approval_signoffs/INSERT arm parent-lookup-discriminated via a SELECT of
// the parent approval_instances row's subject_kind + a nested CASE), so the
// golden target advances with the latest rendered migration (M7
// validation-contract.md §5, M6 §3, M2 §1.4/§1.5.a, ADR 0083 + follow-on).
func TestRenderMigration_MatchesCommittedFile(t *testing.T) {
	repoRoot, err := findRepoRoot()
	if err != nil {
		t.Fatalf("locate repo root: %v", err)
	}
	migrationPath := filepath.Join(repoRoot, "db", "migrations", "0300_tripwire_signoff_parent_discriminator.sql")

	committed, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read committed migration %s: %v (has it been generated via cmd/gen-tripwire?)", migrationPath, err)
	}

	rendered := RenderMigration()

	if rendered != string(committed) {
		t.Errorf("RenderMigration() does not byte-equal committed %s.\n--- rendered len=%d ---\n%s\n--- committed len=%d ---\n%s",
			migrationPath, len(rendered), rendered, len(committed), string(committed))
	}
}

// TestRenderMigration_Deterministic pins determinism: RenderMigration() must
// return identical output on every call (validation-contract.md §1.4).
func TestRenderMigration_Deterministic(t *testing.T) {
	first := RenderMigration()
	second := RenderMigration()
	if first != second {
		t.Fatalf("RenderMigration() is not deterministic across calls")
	}
}

// findRepoRoot walks up from the working directory looking for go.mod.
func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}
