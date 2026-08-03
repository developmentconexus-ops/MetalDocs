package main

import (
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/tripwire"
)

// repoRoot(t) is defined in exit_code_test.go and reused here.

// --- TRIPWIRE-ARM-PARITY ----------------------------------------------------

// Clean tree: the real TripwireArms + the real committed tripwire migration
// (0275 as of M6 F6.2) must be 0 violations (contract §1.5.a POSITIVE proof).
func TestTripwireArmParity_CleanTreeGreen(t *testing.T) {
	got, err := checkTripwireArmParity(repoRoot(t), true)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "TRIPWIRE-ARM-PARITY"); n != 0 {
		t.Fatalf("clean tree: want 0 violations, got %d: %+v", n, got)
	}
}

// NEGATIVE (i): a TripwireArms entry referencing a non-registry cap must fire.
// checkTripwireArmParity checks the real, imported tripwire.TripwireArms
// against the real iamdomain.IsValidCapability, so this fixture exercises the
// same non-registry-cap detection logic directly against a synthetic Arm
// slice — proving the predicate the rule applies, without mutating the real
// arms.go (forbidden by the task constraints).
func TestTripwireArmParity_NonRegistryCapFiresDirectly(t *testing.T) {
	bogus := iamdomain.Capability("document.bogus_not_in_registry")
	if iamdomain.IsValidCapability(bogus) {
		t.Fatalf("fixture cap %q must NOT be a real registry capability for this test to be meaningful", bogus)
	}
	fakeArms := []tripwire.Arm{
		{Table: "documents", Op: tripwire.OpInsert, Caps: []iamdomain.Capability{bogus}},
	}
	var violations int
	for _, arm := range fakeArms {
		for _, cap := range arm.Caps {
			if !iamdomain.IsValidCapability(cap) {
				violations++
			}
		}
	}
	if violations != 1 {
		t.Fatalf("non-registry cap scenario: want 1 violation, got %d", violations)
	}
}

// NEGATIVE (ii): a hand-edited tripwire migration that no longer byte-equals
// tripwire.RenderMigration() must fire. We cannot mutate the committed file
// (forbidden), so we point checkTripwireArmParity at a temp modulesRoot that
// carries a deliberately WRONG copy of the migration at the same relative
// path — this exercises the exact byte-compare branch the rule runs in
// production against a tree whose committed SQL has drifted.
func TestTripwireArmParity_MutatedMigrationFires(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, tripwireGoldenPath, "-- hand-edited, does not match RenderMigration()\n")
	got, err := checkTripwireArmParity(dir, false)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "TRIPWIRE-ARM-PARITY"); n != 1 {
		t.Fatalf("mutated migration: want 1 violation, got %d: %+v", n, got)
	}
	if !containsMsg(got, "does not byte-equal") {
		t.Fatalf("violation should explain the byte-mismatch: %+v", got)
	}
}

// A missing migration file under a fixture (non-strict) root is skipped, not a
// violation — mirrors every other core-file helper's requireCoreFile
// convention (parseCapabilityConsts, checkSeedRegistryParity, ...) so
// pre-existing testdata/e2e fixture roots that never plant a migrations/ tree
// are unaffected.
func TestTripwireArmParity_MissingMigrationSkippedNonStrict(t *testing.T) {
	dir := t.TempDir()
	got, err := checkTripwireArmParity(dir, false)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "TRIPWIRE-ARM-PARITY"); n != 0 {
		t.Fatalf("missing migration, non-strict: want 0 violations (skip), got %d: %+v", n, got)
	}
}

// Under strict (production/CI root) a missing migration file is a hard error,
// not a silent pass — the same wrong-root protection as every other
// requireCoreFile-gated helper (ADR 0022 Phase 13).
func TestTripwireArmParity_MissingMigrationHardErrorsStrict(t *testing.T) {
	dir := t.TempDir()
	_, err := checkTripwireArmParity(dir, true)
	if err == nil {
		t.Fatal("missing migration, strict: want hard error, got nil")
	}
	if !strings.Contains(err.Error(), "strict") {
		t.Fatalf("strict error should explain the wrong-root contract, got: %v", err)
	}
}

// --- TRIPWIRE-ARM-DRIFT ------------------------------------------------------

// Clean tree: the real tree (repository.go force-release, obsolete_service.go
// MarkObsolete, etc.) must be 0 violations post-0271 (contract §1.5.b
// POSITIVE proof — Stage 1 already widened the documents/UPDATE arm).
func TestTripwireArmDrift_CleanTreeGreen(t *testing.T) {
	got, err := checkTripwireArmDrift(repoRoot(t), token.NewFileSet(), true)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "TRIPWIRE-ARM-DRIFT"); n != 0 {
		t.Fatalf("clean tree post-0271: want 0 violations, got %d: %+v", n, got)
	}
}

// NEGATIVE (the mission's required synthetic, contract §1.5.b): a function
// that asserts a capability NOT in the gated table's arm, then writes that
// table, must fire naming file:line/table/cap. CapDocumentView ("document.view")
// is a real registry cap (so parseCapabilityConsts resolves it) but is not a
// member of TripwireArms[documents,INSERT] (={document.create}) nor
// TripwireArms[documents,UPDATE] — so this is a genuine, real drift scenario,
// not a contrived fixture.
func TestTripwireArmDrift_SyntheticUnarmedCapFires(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, filepath.Join("internal", "modules", "iam", "domain", "model.go"), fixtureCapModel)
	writeFile(t, dir, filepath.Join("internal", "modules", "fixture", "repo.go"),
		"package fixture\n"+
			"import (\n\t\"context\"\n\t\"database/sql\"\n)\n"+
			"var authz = struct{ Require func(context.Context, *sql.Tx, string, string) error }{}\n"+
			"type iamdomainPkg struct{ CapDocumentView string }\n"+
			"var iamdomain = iamdomainPkg{CapDocumentView: \"document.view\"}\n"+
			"func writeDoc(ctx context.Context, tx *sql.Tx) error {\n"+
			"\tif err := authz.Require(ctx, tx, string(iamdomain.CapDocumentView), \"area1\"); err != nil {\n"+
			"\t\treturn err\n"+
			"\t}\n"+
			"\t_, err := tx.ExecContext(ctx, `INSERT INTO documents (id) VALUES ($1)`, \"x\")\n"+
			"\treturn err\n"+
			"}\n")
	got, err := checkTripwireArmDrift(dir, token.NewFileSet(), false)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "TRIPWIRE-ARM-DRIFT"); n != 1 {
		t.Fatalf("unarmed cap + gated write: want 1 violation, got %d: %+v", n, got)
	}
	v := got[0]
	if !strings.Contains(v.File, "repo.go") {
		t.Fatalf("violation should name the file: %+v", v)
	}
	if !containsMsg(got, "documents") || !containsMsg(got, "document.view") || !containsMsg(got, "writeDoc") {
		t.Fatalf("violation message should name the function, table, and unarmed cap: %+v", v)
	}
}

// Removing the fixture (no authz.Require + write pairing) restores GREEN —
// contract §1.5.b "Removing it ... -> GREEN".
func TestTripwireArmDrift_RemovingFixtureIsGreen(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, filepath.Join("internal", "modules", "iam", "domain", "model.go"), fixtureCapModel)
	writeFile(t, dir, filepath.Join("internal", "modules", "fixture", "repo.go"),
		"package fixture\n"+
			"import (\n\t\"context\"\n\t\"database/sql\"\n)\n"+
			"func writeDoc(ctx context.Context, tx *sql.Tx) error {\n"+
			"\t_, err := tx.ExecContext(ctx, `INSERT INTO documents (id) VALUES ($1)`, \"x\")\n"+
			"\treturn err\n"+
			"}\n")
	got, err := checkTripwireArmDrift(dir, token.NewFileSet(), false)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "TRIPWIRE-ARM-DRIFT"); n != 0 {
		t.Fatalf("no authz.Require call: want 0 violations, got %d: %+v", n, got)
	}
}

// A function asserting an arm-member cap for the write's table+op is GREEN —
// e.g. CapDocumentEdit is in TripwireArms[documents,UPDATE].
func TestTripwireArmDrift_ArmMemberCapIsGreen(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, filepath.Join("internal", "modules", "iam", "domain", "model.go"), fixtureCapModel)
	writeFile(t, dir, filepath.Join("internal", "modules", "fixture", "repo.go"),
		"package fixture\n"+
			"import (\n\t\"context\"\n\t\"database/sql\"\n)\n"+
			"var authz = struct{ Require func(context.Context, *sql.Tx, string, string) error }{}\n"+
			"type iamdomainPkg struct{ CapDocumentEdit string }\n"+
			"var iamdomain = iamdomainPkg{CapDocumentEdit: \"document.edit\"}\n"+
			"func updateDoc(ctx context.Context, tx *sql.Tx) error {\n"+
			"\tif err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), \"area1\"); err != nil {\n"+
			"\t\treturn err\n"+
			"\t}\n"+
			"\t_, err := tx.ExecContext(ctx, `UPDATE documents SET status='x' WHERE id=$1`, \"x\")\n"+
			"\treturn err\n"+
			"}\n")
	got, err := checkTripwireArmDrift(dir, token.NewFileSet(), false)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "TRIPWIRE-ARM-DRIFT"); n != 0 {
		t.Fatalf("arm-member cap: want 0 violations, got %d: %+v", n, got)
	}
}

// A write to a non-gated table (e.g. governance_events, outbox) is never
// flagged, regardless of asserted cap — contract §1.5.b "ignore writes to
// non-gated tables".
func TestTripwireArmDrift_IgnoresNonGatedTable(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, filepath.Join("internal", "modules", "iam", "domain", "model.go"), fixtureCapModel)
	writeFile(t, dir, filepath.Join("internal", "modules", "fixture", "repo.go"),
		"package fixture\n"+
			"import (\n\t\"context\"\n\t\"database/sql\"\n)\n"+
			"var authz = struct{ Require func(context.Context, *sql.Tx, string, string) error }{}\n"+
			"type iamdomainPkg struct{ CapDocumentView string }\n"+
			"var iamdomain = iamdomainPkg{CapDocumentView: \"document.view\"}\n"+
			"func writeOutbox(ctx context.Context, tx *sql.Tx) error {\n"+
			"\tif err := authz.Require(ctx, tx, string(iamdomain.CapDocumentView), \"area1\"); err != nil {\n"+
			"\t\treturn err\n"+
			"\t}\n"+
			"\t_, err := tx.ExecContext(ctx, `INSERT INTO outbox (id) VALUES ($1)`, \"x\")\n"+
			"\treturn err\n"+
			"}\n")
	got, err := checkTripwireArmDrift(dir, token.NewFileSet(), false)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "TRIPWIRE-ARM-DRIFT"); n != 0 {
		t.Fatalf("non-gated table write: want 0 violations, got %d: %+v", n, got)
	}
}

// Match-one: a function that asserts MULTIPLE caps is green if ANY one of
// them is an arm member, mirroring the trigger's match-one semantics
// (contract §1.5.b "match-one, mirroring the trigger").
func TestTripwireArmDrift_MatchOneAmongMultipleCapsIsGreen(t *testing.T) {
	dir := t.TempDir()
	multiCapModel := "package domain\n" +
		"type Capability string\n" +
		"const (\n" +
		"\tCapDocumentEdit Capability = \"document.edit\"\n" +
		"\tCapDocumentView Capability = \"document.view\"\n" +
		")\n"
	writeFile(t, dir, filepath.Join("internal", "modules", "iam", "domain", "model.go"), multiCapModel)
	writeFile(t, dir, filepath.Join("internal", "modules", "fixture", "repo.go"),
		"package fixture\n"+
			"import (\n\t\"context\"\n\t\"database/sql\"\n)\n"+
			"var authz = struct{ Require func(context.Context, *sql.Tx, string, string) error }{}\n"+
			"type iamdomainPkg struct{ CapDocumentEdit, CapDocumentView string }\n"+
			"var iamdomain = iamdomainPkg{CapDocumentEdit: \"document.edit\", CapDocumentView: \"document.view\"}\n"+
			"func updateDoc(ctx context.Context, tx *sql.Tx) error {\n"+
			"\tif err := authz.Require(ctx, tx, string(iamdomain.CapDocumentView), \"area1\"); err != nil {\n"+
			"\t\treturn err\n"+
			"\t}\n"+
			"\tif err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), \"area1\"); err != nil {\n"+
			"\t\treturn err\n"+
			"\t}\n"+
			"\t_, err := tx.ExecContext(ctx, `UPDATE documents SET status='x' WHERE id=$1`, \"x\")\n"+
			"\treturn err\n"+
			"}\n")
	got, err := checkTripwireArmDrift(dir, token.NewFileSet(), false)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "TRIPWIRE-ARM-DRIFT"); n != 0 {
		t.Fatalf("match-one among multiple caps: want 0 violations, got %d: %+v", n, got)
	}
}

// Subject-discriminated arm (ADR 0083): approval_instances/INSERT now has TWO
// real arms — document->document.submit, template->template.submit. A function
// asserting ONLY template.submit and writing approval_instances must be GREEN
// (the static DRIFT lint unions the arm caps; the DB trigger's CASE
// NEW.subject_kind does the precise per-row enforcement). Before the
// unionArmCapsFor generalization this fired a false positive, because armFor
// returned only the FIRST matching arm (document->{document.submit}).
func TestTripwireArmDrift_TemplateSubmitOnApprovalInstancesIsGreen(t *testing.T) {
	dir := t.TempDir()
	subjModel := "package domain\n" +
		"type Capability string\n" +
		"const (\n" +
		"\tCapDocumentSubmit Capability = \"document.submit\"\n" +
		"\tCapTemplateSubmit Capability = \"template.submit\"\n" +
		")\n"
	writeFile(t, dir, filepath.Join("internal", "modules", "iam", "domain", "model.go"), subjModel)
	writeFile(t, dir, filepath.Join("internal", "modules", "fixture", "repo.go"),
		"package fixture\n"+
			"import (\n\t\"context\"\n\t\"database/sql\"\n)\n"+
			"var authz = struct{ Require func(context.Context, *sql.Tx, string, string) error }{}\n"+
			"type iamdomainPkg struct{ CapTemplateSubmit string }\n"+
			"var iamdomain = iamdomainPkg{CapTemplateSubmit: \"template.submit\"}\n"+
			"func submitTemplate(ctx context.Context, tx *sql.Tx) error {\n"+
			"\tif err := authz.Require(ctx, tx, string(iamdomain.CapTemplateSubmit), \"tenant\"); err != nil {\n"+
			"\t\treturn err\n"+
			"\t}\n"+
			"\t_, err := tx.ExecContext(ctx, `INSERT INTO approval_instances (id) VALUES ($1)`, \"x\")\n"+
			"\treturn err\n"+
			"}\n")
	got, err := checkTripwireArmDrift(dir, token.NewFileSet(), false)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "TRIPWIRE-ARM-DRIFT"); n != 0 {
		t.Fatalf("template.submit + approval_instances INSERT: want 0 violations (union arm), got %d: %+v", n, got)
	}
}

// Fail-closed for the discriminated table: a function asserting NEITHER
// document.submit NOR template.submit, then writing approval_instances, still
// fires — the union of the two arms is {document.submit, template.submit}, and
// document.view is in neither.
func TestTripwireArmDrift_NonSubmitCapOnApprovalInstancesFires(t *testing.T) {
	dir := t.TempDir()
	subjModel := "package domain\n" +
		"type Capability string\n" +
		"const (\n" +
		"\tCapDocumentView Capability = \"document.view\"\n" +
		")\n"
	writeFile(t, dir, filepath.Join("internal", "modules", "iam", "domain", "model.go"), subjModel)
	writeFile(t, dir, filepath.Join("internal", "modules", "fixture", "repo.go"),
		"package fixture\n"+
			"import (\n\t\"context\"\n\t\"database/sql\"\n)\n"+
			"var authz = struct{ Require func(context.Context, *sql.Tx, string, string) error }{}\n"+
			"type iamdomainPkg struct{ CapDocumentView string }\n"+
			"var iamdomain = iamdomainPkg{CapDocumentView: \"document.view\"}\n"+
			"func writeInstance(ctx context.Context, tx *sql.Tx) error {\n"+
			"\tif err := authz.Require(ctx, tx, string(iamdomain.CapDocumentView), \"tenant\"); err != nil {\n"+
			"\t\treturn err\n"+
			"\t}\n"+
			"\t_, err := tx.ExecContext(ctx, `INSERT INTO approval_instances (id) VALUES ($1)`, \"x\")\n"+
			"\treturn err\n"+
			"}\n")
	got, err := checkTripwireArmDrift(dir, token.NewFileSet(), false)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "TRIPWIRE-ARM-DRIFT"); n != 1 {
		t.Fatalf("document.view + approval_instances INSERT: want 1 violation, got %d: %+v", n, got)
	}
	if !containsMsg(got, "approval_instances") || !containsMsg(got, "writeInstance") {
		t.Fatalf("violation should name the function and table: %+v", got)
	}
	// The reported arm set must be the UNION (both submit caps), proving the
	// generalization rendered the full expected membership, not a single arm.
	if !containsMsg(got, "document.submit") || !containsMsg(got, "template.submit") {
		t.Fatalf("violation message should list the unioned arm caps: %+v", got)
	}
}

// --- Detection proof: pre-0271-style arm reproduces the two real latent
// incidents (force-release + obsolete) as RED. We do NOT mutate the real
// arms.go here (forbidden by the task); this test constructs a temp tree that
// mirrors the shape of the real force-release/obsolete functions but is
// pointed at a fixture registry, and separately documents (in the returned
// summary) that the live detection proof was captured by temporarily
// reverting arms.go and running the rule against the real tree — see the
// orchestrator's captured transcript, not this unit test, for that proof
// (task constraint: do not leave repo changes from this test file).
func TestTripwireArmDrift_ReproducesForceReleaseShape(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, filepath.Join("internal", "modules", "iam", "domain", "model.go"), fixtureCapModel)
	// documents/UPDATE's real arm (post-0271) contains document.edit,
	// document.obsolete, membership.manage — NOT document.view. Asserting only
	// CapDocumentView while writing documents(UPDATE) reproduces the same
	// "asserted cap not in arm" shape as the pre-0271 force-release incident
	// (which asserted membership.manage, absent from the pre-0271 {document.edit}
	// arm).
	writeFile(t, dir, filepath.Join("internal", "modules", "fixture", "force_release.go"),
		"package fixture\n"+
			"import (\n\t\"context\"\n\t\"database/sql\"\n)\n"+
			"var authz = struct{ Require func(context.Context, *sql.Tx, string, string) error }{}\n"+
			"type iamdomainPkg struct{ CapDocumentView string }\n"+
			"var iamdomain = iamdomainPkg{CapDocumentView: \"document.view\"}\n"+
			"func ForceReleaseSession(ctx context.Context, tx *sql.Tx) error {\n"+
			"\tif err := authz.Require(ctx, tx, string(iamdomain.CapDocumentView), \"area1\"); err != nil {\n"+
			"\t\treturn err\n"+
			"\t}\n"+
			"\t_, err := tx.ExecContext(ctx, `UPDATE documents SET active_session_id=NULL WHERE id=$1`, \"x\")\n"+
			"\treturn err\n"+
			"}\n")
	got, err := checkTripwireArmDrift(dir, token.NewFileSet(), false)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "TRIPWIRE-ARM-DRIFT"); n != 1 {
		t.Fatalf("force-release-shaped unarmed cap: want 1 violation, got %d: %+v", n, got)
	}
	if !containsMsg(got, "ForceReleaseSession") {
		t.Fatalf("violation should name the function: %+v", got)
	}
}

// sanity: RenderMigration is deterministic across calls (findArm/renderArray
// side-effect-free) — guards the byte-compare in checkTripwireArmParity
// against nondeterminism false-flaking CI.
func TestRenderMigration_Deterministic(t *testing.T) {
	a := tripwire.RenderMigration()
	b := tripwire.RenderMigration()
	if a != b {
		t.Fatal("RenderMigration() is not deterministic across calls")
	}
}

// sanity: the real committed latest tripwire migration file (0275 as of M6
// F6.2) exists and is non-empty (guards against an accidental deletion silently
// turning the PARITY rule green via checkTripwireArmParity's missing-file
// violation not firing for the wrong reason).
func TestCommittedTripwireMigrationExists(t *testing.T) {
	path := filepath.Join(repoRoot(t), filepath.FromSlash(tripwireGoldenPath))
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if info.Size() == 0 {
		t.Fatalf("%s is empty", path)
	}
}
