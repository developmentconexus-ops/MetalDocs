package main

import (
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

// writeFile writes content to <dir>/<rel>, creating parent dirs.
func writeFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	path := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}

func countRule(vs []Violation, rule string) int {
	n := 0
	for _, v := range vs {
		if v.Rule == rule {
			n++
		}
	}
	return n
}

// --- no-inline-capability -------------------------------------------------

func TestNoInlineCapability_CleanTree(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pkg/clean.go", "package pkg\nfunc f() string { return string(someVar) }\nvar someVar = \"x\"\n")
	got, err := checkNoInlineCapability(dir, token.NewFileSet())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "no-inline-capability"); n != 0 {
		t.Fatalf("clean tree: want 0 violations, got %d: %+v", n, got)
	}
}

func TestNoInlineCapability_BitesOnDrift(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pkg/drift.go", "package pkg\nvar _ = Capability(\"doc.bogus\")\nfunc Capability(s string) string { return s }\n")
	got, err := checkNoInlineCapability(dir, token.NewFileSet())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "no-inline-capability"); n != 1 {
		t.Fatalf("drift: want 1 violation, got %d: %+v", n, got)
	}
}

func TestNoInlineCapability_QualifiedConversionBites(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pkg/q.go", "package pkg\nvar _ = iamdomain.Capability(\"doc.bogus\")\n")
	got, err := checkNoInlineCapability(dir, token.NewFileSet())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "no-inline-capability"); n != 1 {
		t.Fatalf("qualified drift: want 1, got %d: %+v", n, got)
	}
}

func TestNoInlineCapability_IgnoresTestsAndModelAndNonLiteral(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pkg/x_test.go", "package pkg\nvar _ = Capability(\"in.test\")\n")
	writeFile(t, dir, "internal/modules/iam/domain/model.go", "package domain\ntype Capability string\nvar _ = Capability(\"defn\")\n")
	writeFile(t, dir, "pkg/nonlit.go", "package pkg\nfunc g(s string){ _ = Capability(s) }\nfunc Capability(s string) string { return s }\n")
	got, err := checkNoInlineCapability(dir, token.NewFileSet())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "no-inline-capability"); n != 0 {
		t.Fatalf("exemptions: want 0, got %d: %+v", n, got)
	}
}

// --- seed-registry-parity -------------------------------------------------

// fullSeed builds a role_capabilities seed that grants every registry cap, plus
// any extra raw cap strings supplied. omit removes a real cap from the seed.
func fullSeed(extra []string, omit map[string]bool) string {
	var b strings.Builder
	for _, c := range iamdomain.AllCapabilities() {
		if omit[string(c)] {
			continue
		}
		b.WriteString("INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('viewer', '" + string(c) + "', '') ON CONFLICT DO NOTHING;\n")
	}
	for _, e := range extra {
		b.WriteString("INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES ('viewer', '" + e + "', '') ON CONFLICT DO NOTHING;\n")
	}
	return b.String()
}

func TestSeedRegistryParity_Clean(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "db/reference-data/0001_product_reference_data.sql", fullSeed(nil, nil))
	got, err := checkSeedRegistryParity(dir)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "seed-registry-parity"); n != 0 {
		t.Fatalf("clean: want 0, got %d: %+v", n, got)
	}
}

func TestSeedRegistryParity_SeededNotInRegistry(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "db/reference-data/0001_product_reference_data.sql", fullSeed([]string{"doc.bogus"}, nil))
	got, err := checkSeedRegistryParity(dir)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "seed-registry-parity"); n != 1 {
		t.Fatalf("seeded-not-in-registry: want 1, got %d: %+v", n, got)
	}
	if !strings.Contains(got[len(got)-1].Message, "doc.bogus") && !containsMsg(got, "doc.bogus") {
		t.Fatalf("expected doc.bogus in message: %+v", got)
	}
}

func TestSeedRegistryParity_RegistryNotSeeded(t *testing.T) {
	dir := t.TempDir()
	// Omit one real cap from the seed -> (b) must fire for exactly that cap.
	caps := iamdomain.AllCapabilities()
	missing := string(caps[0])
	writeFile(t, dir, "db/reference-data/0001_product_reference_data.sql", fullSeed(nil, map[string]bool{missing: true}))
	got, err := checkSeedRegistryParity(dir)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "seed-registry-parity"); n != 1 {
		t.Fatalf("registry-not-seeded: want 1, got %d: %+v", n, got)
	}
	if !containsMsg(got, missing) {
		t.Fatalf("expected %q in message: %+v", missing, got)
	}
}

// --- wiki-capability-parity ----------------------------------------------

func TestWikiCapabilityParity_Clean(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, filepath.Join("wiki", "concepts", "authz-tiers.md"),
		"An area_admin holds `cap:membership.manage` and the illustrative `doc.create` example is ignored.\n")
	got, err := checkWikiCapabilityParity(dir)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "wiki-capability-parity"); n != 0 {
		t.Fatalf("clean: want 0, got %d: %+v", n, got)
	}
}

func TestWikiCapabilityParity_BitesOnDrift(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, filepath.Join("wiki", "references", "local-dev-credentials.md"),
		"area_admin grants `cap:membership.grant` to areas.\n")
	got, err := checkWikiCapabilityParity(dir)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "wiki-capability-parity"); n != 1 {
		t.Fatalf("drift: want 1, got %d: %+v", n, got)
	}
	if !containsMsg(got, "membership.grant") {
		t.Fatalf("expected membership.grant in message: %+v", got)
	}
}

func TestWikiCapabilityParity_IgnoresUnmarkedProse(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, filepath.Join("wiki", "concepts", "authz-tiers.md"),
		"Deferred `route.view`, illustrative `doc.create`, file `permissions.go`, GUC `metaldocs.actor_id` — none are markers.\n")
	got, err := checkWikiCapabilityParity(dir)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "wiki-capability-parity"); n != 0 {
		t.Fatalf("unmarked prose: want 0, got %d: %+v", n, got)
	}
}

// --- authz-area-scope-binding (ADR 0022 Phase 7 core control) -------------

// fixtureCapModel writes a minimal iam/domain/model.go the guard parses for the
// const-name -> value map. Values must be REAL capabilities because the guard
// classifies them via the compiled iamdomain.IsAreaGrade.
const fixtureCapModel = "package domain\n" +
	"type Capability string\n" +
	"const (\n" +
	"\tCapDocumentEdit Capability = \"document.edit\"\n" + // area-grade
	"\tCapDocumentView Capability = \"document.view\"\n" + // tenant-grade
	")\n"

func TestAuthzAreaScopeBinding_BitesOnAreaGradeTenant(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, filepath.Join("internal", "modules", "iam", "domain", "model.go"), fixtureCapModel)
	writeFile(t, dir, "repo/r.go",
		"package repo\nfunc f() { authz.Require(nil, nil, string(iamdomain.CapDocumentEdit), \"tenant\") }\n")
	got, err := checkAuthzAreaScopeBinding(dir, token.NewFileSet())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "authz-area-scope-binding"); n != 1 {
		t.Fatalf("area-grade+\"tenant\": want 1 violation, got %d: %+v", n, got)
	}
	if !containsMsg(got, "document.edit") {
		t.Fatalf("violation message should name the cap: %+v", got)
	}
}

func TestAuthzAreaScopeBinding_GreenOnRealArea(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, filepath.Join("internal", "modules", "iam", "domain", "model.go"), fixtureCapModel)
	writeFile(t, dir, "repo/r.go",
		"package repo\nfunc f(areaCode string) { authz.Require(nil, nil, string(iamdomain.CapDocumentEdit), areaCode) }\n")
	got, err := checkAuthzAreaScopeBinding(dir, token.NewFileSet())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "authz-area-scope-binding"); n != 0 {
		t.Fatalf("real area arg: want 0, got %d: %+v", n, got)
	}
}

func TestAuthzAreaScopeBinding_IgnoresTenantGradeAndVariableCap(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, filepath.Join("internal", "modules", "iam", "domain", "model.go"), fixtureCapModel)
	// Tenant-grade cap with "tenant" is legitimate; a variable cap is unresolvable.
	writeFile(t, dir, "repo/r.go",
		"package repo\nfunc f(cap string) {\n\tauthz.Require(nil, nil, string(iamdomain.CapDocumentView), \"tenant\")\n\tauthz.Require(nil, nil, cap, \"tenant\")\n}\n")
	got, err := checkAuthzAreaScopeBinding(dir, token.NewFileSet())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "authz-area-scope-binding"); n != 0 {
		t.Fatalf("tenant-grade + variable cap: want 0, got %d: %+v", n, got)
	}
}

// --- golden seed row count (ADR 0022 Phase 7) -----------------------------

// expectedRoleCapabilityRows is the exact number of role_capabilities INSERT
// rows the seed parser (seedCapRE) must capture from the canonical baseline. It
// is a golden value: a reformat that splits/merges INSERTs, an accidental row
// drop, or a multi-row VALUES rewrite that the single-row regex cannot parse all
// change this count and fail loudly — protecting seed-registry-parity, which
// silently under-counts if the regex stops matching a row.
const expectedRoleCapabilityRows = 94

func TestSeedRowCount_GoldenMatchesParser(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// scripts/api-lint/ -> repo root is two levels up.
	seedPath := filepath.Join(filepath.Dir(thisFile), "..", "..",
		"db", "reference-data", "0001_product_reference_data.sql")
	raw, err := os.ReadFile(seedPath)
	if err != nil {
		t.Fatalf("read seed %s: %v", seedPath, err)
	}
	got := len(seedCapRE.FindAllStringSubmatchIndex(string(raw), -1))
	if got != expectedRoleCapabilityRows {
		t.Fatalf("seedCapRE captured %d role_capabilities rows, want golden %d — if this change is intentional (a real grant added/removed), update expectedRoleCapabilityRows; if not, a reformat or multi-row VALUES rewrite broke the single-row parser (ADR 0022 Phase 7).", got, expectedRoleCapabilityRows)
	}
}

func containsMsg(vs []Violation, sub string) bool {
	for _, v := range vs {
		if strings.Contains(v.Message, sub) {
			return true
		}
	}
	return false
}
