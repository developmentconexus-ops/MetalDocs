package main

import (
	"go/token"
	"strings"
	"testing"
)

// This file is the ADR 0022 Phase 13 end-to-end self-test. The per-rule unit
// tests in registry_rules_test.go call each helper directly and so NEVER
// exercised the real composition (RunCodeRules) or the path-resolution contract
// — which is exactly why they stayed green while the CI invocation passed a
// wrong root that made every binding helper silently load nothing. These tests
// drive the REAL entrypoint (RunCodeRules, the function main() calls) against a
// tree with a PLANTED violation of each AST-only guard, proving the guards are
// wired to the real root, and prove strict mode turns a wrong root into a hard
// error instead of a clean pass.

// e2eCapRoleModel is a minimal iam/domain/model.go the binding guards parse for
// the Capability and Role const maps. Values are REAL capabilities/roles because
// the guards classify them via the compiled iamdomain (IsAreaGrade) and union
// RoleSystemAdmin into the banned delivery-role set. CapDocumentEdit is
// area-grade; that is what makes the planted Require(...,"tenant") bite.
const e2eCapRoleModel = "package domain\n" +
	"type Capability string\n" +
	"type Role string\n" +
	"const (\n" +
	"\tCapDocumentEdit Capability = \"document.edit\"\n" + // area-grade
	"\tCapDocumentView Capability = \"document.view\"\n" + // tenant-grade
	")\n" +
	"const RoleSystemAdmin Role = \"system_admin\"\n"

// e2eMinimalSpec is a syntactically valid OpenAPI doc with zero operations, so
// RunSpecRules/checkAuthzCallPresent contribute nothing and the planted code
// guards are the only violations under test.
const e2eMinimalSpec = "openapi: 3.0.0\n" +
	"info:\n  title: e2e\n  version: \"1\"\n" +
	"paths: {}\n"

// plantTree lays down model.go + minimal spec under dir and returns the spec
// path. The four planted-violation files are added by the caller when wanted.
func plantTree(t *testing.T, dir string) string {
	t.Helper()
	writeFile(t, dir, "internal/modules/iam/domain/model.go", e2eCapRoleModel)
	writeFile(t, dir, "api/openapi/v1/openapi.yaml", e2eMinimalSpec)
	return dir + "/api/openapi/v1/openapi.yaml"
}

// TestEndToEnd_BindingGuardsBiteViaRealEntrypoint plants one violation of each
// AST-only guard and asserts RunCodeRules (the real composition) catches every
// one. This is the regression the unit tests missed: it fails if the guards are
// ever again wired to a root where they load nothing.
func TestEndToEnd_BindingGuardsBiteViaRealEntrypoint(t *testing.T) {
	dir := t.TempDir()
	spec := plantTree(t, dir)

	// (1) raw-string capability passed to authz.Require (area is a var, not
	//     "tenant", so ONLY no-rawstring-capability bites here).
	writeFile(t, dir, "internal/modules/x/rawcap.go",
		"package x\nfunc f(area string) { authz.Require(nil, nil, \"document.bogus\", area) }\n")
	// (2) role-name string literal in a delivery/http authz gate.
	writeFile(t, dir, "internal/modules/x/delivery/http/role.go",
		"package http\nconst roleAdmin = \"system_admin\"\n")
	// (3) inline Capability("literal") conversion outside the registry.
	writeFile(t, dir, "internal/modules/x/inline.go",
		"package x\nvar _ = Capability(\"doc.inline\")\n")
	// (4) area-grade capability enforced with the literal "tenant" sentinel.
	writeFile(t, dir, "internal/modules/x/areatenant.go",
		"package x\nfunc g() { authz.Require(nil, nil, string(iamdomain.CapDocumentEdit), \"tenant\") }\n")

	got, err := RunCodeRules(spec, dir, false)
	if err != nil {
		t.Fatalf("RunCodeRules: %v", err)
	}

	wantOnePer := []string{
		"no-rawstring-capability",
		"no-rolestring-in-delivery",
		"no-inline-capability",
		"authz-area-scope-binding",
	}
	for _, rule := range wantOnePer {
		if n := countRule(got, rule); n != 1 {
			t.Fatalf("planted %s: want 1 via real entrypoint, got %d\nfull got=%+v", rule, n, got)
		}
	}
	if len(got) != len(wantOnePer) {
		t.Fatalf("planted tree: want exactly %d violations, got %d\nfull got=%+v", len(wantOnePer), len(got), got)
	}
}

// TestEndToEnd_CleanTreeGuardsZero proves the same real entrypoint returns the
// blocking guards at ZERO on a tree with no planted violations — the green-state
// half of the self-test (the gate must be 0 on a clean tree before CI enables
// it).
func TestEndToEnd_CleanTreeGuardsZero(t *testing.T) {
	dir := t.TempDir()
	spec := plantTree(t, dir)
	// One benign, fully-typed call site that must NOT trip any guard.
	writeFile(t, dir, "internal/modules/x/clean.go",
		"package x\nfunc f(area string) { authz.Require(nil, nil, string(iamdomain.CapDocumentEdit), area) }\n")

	got, err := RunCodeRules(spec, dir, false)
	if err != nil {
		t.Fatalf("RunCodeRules: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("clean tree via real entrypoint: want 0 violations, got %d\nfull got=%+v", len(got), got)
	}
}

// TestEndToEnd_StrictWrongRootErrors is the direct regression for CI bug #1: a
// wrong modulesRoot (here, a tree with no core registry files) used to make
// every binding helper return an empty set and PASS. Under strict mode it must
// now be a hard error, so a misconfigured CI invocation can never again
// masquerade as clean.
func TestEndToEnd_StrictWrongRootErrors(t *testing.T) {
	dir := t.TempDir()
	// Only the spec exists; no model.go, no seed, no allow-list, no wiki docs —
	// the shape a wrong root produces.
	writeFile(t, dir, "api/openapi/v1/openapi.yaml", e2eMinimalSpec)
	spec := dir + "/api/openapi/v1/openapi.yaml"

	_, err := RunCodeRules(spec, dir, true)
	if err == nil {
		t.Fatal("strict mode on a root missing core files: want hard error, got nil (silent-empty swallow not killed)")
	}
	if !strings.Contains(err.Error(), "strict") {
		t.Fatalf("strict error should explain the wrong-root contract, got: %v", err)
	}

	// Same wrong root, non-strict: must NOT error (fixture/back-compat path) —
	// the binding helpers legitimately skip when files are absent.
	if _, err := RunCodeRules(spec, dir, false); err != nil {
		t.Fatalf("non-strict on missing core files: want no error (fixture path), got %v", err)
	}
}

// TestStrict_EachCoreFileHelperErrorsButFixturePathSurvives nails the
// silent-empty swallow at the helper level: every core-file reader must hard-
// error in strict mode when its file is absent (wrong root) AND must keep the
// non-strict skip so the testdata fixture roots still work.
func TestStrict_EachCoreFileHelperErrorsButFixturePathSurvives(t *testing.T) {
	dir := t.TempDir() // empty: no model.go, no seed, no allow-list, no wiki docs
	fset := token.NewFileSet()

	type helper struct {
		name   string
		strict func() error
		soft   func() error
	}
	helpers := []helper{
		{"parseCapabilityConsts",
			func() error { _, e := parseCapabilityConsts(dir, fset, true); return e },
			func() error { _, e := parseCapabilityConsts(dir, fset, false); return e }},
		{"parseRoleConsts",
			func() error { _, e := parseRoleConsts(dir, fset, true); return e },
			func() error { _, e := parseRoleConsts(dir, fset, false); return e }},
		{"checkSeedRegistryParity",
			func() error { _, e := checkSeedRegistryParity(dir, true); return e },
			func() error { _, e := checkSeedRegistryParity(dir, false); return e }},
		{"checkWikiCapabilityParity",
			func() error { _, e := checkWikiCapabilityParity(dir, true); return e },
			func() error { _, e := checkWikiCapabilityParity(dir, false); return e }},
		{"loadTripwireAllowlist",
			func() error { _, e := loadTripwireAllowlist(dir, true); return e },
			func() error { _, e := loadTripwireAllowlist(dir, false); return e }},
	}
	for _, h := range helpers {
		if err := h.strict(); err == nil {
			t.Fatalf("%s strict: want hard error on missing core file, got nil", h.name)
		} else if !strings.Contains(err.Error(), "strict") {
			t.Fatalf("%s strict error should explain wrong-root, got: %v", h.name, err)
		}
		if err := h.soft(); err != nil {
			t.Fatalf("%s non-strict: want fixture skip (no error), got %v", h.name, err)
		}
	}
}
