package main

// Registry-binding lints (ADR 0022 Phase 5). These bind the Go capability
// registry (the single source of truth, internal/modules/iam/domain) to the
// other places capabilities are declared, so drift is a red build:
//
//   no-inline-capability   — bans Capability("literal") string conversions
//                            outside the registry (item 2; every cap is typed).
//   seed-registry-parity   — DB role_capabilities seed <-> registry, both ways
//                            (item 3).
//   wiki-capability-parity — `cap:<name>` enforcement-claim markers in wiki
//                            authz docs must exist in the registry (item 4).
//
// Like the other code-side lints these are single-file scans, no call graph.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

// deferredCaps lists registry capabilities intentionally seeded to no tenant
// role (enforced today only via the system_admin tier-2 bypass). Mirrors the
// allow-list in apps/api/cmd/metaldocs-api/permissions_test.go
// (TestEveryCapSeededOrDeferred). Empty since ADR 0022 Phase 2 seeded all four
// document-lifecycle write caps. A future routed-but-unseeded cap goes here with
// a documented deferral.
var deferredCaps = map[iamdomain.Capability]struct{}{}

// wikiAuthzDocs is the fixed set of authorization docs scanned for `cap:`
// markers. Adding a new authz doc that asserts capability grants? List it here.
var wikiAuthzDocs = []string{
	filepath.Join("wiki", "concepts", "authz-tiers.md"),
	filepath.Join("wiki", "references", "local-dev-credentials.md"),
	filepath.Join("wiki", "modules", "iam.md"),
}

// capMarkerRE matches an enforcement-claim capability reference: `cap:<name>`
// inside backticks, where <name> is the dotted capability grammar. Unmarked
// prose tokens (illustrative `doc.create`, deferred `route.view`, filenames,
// GUC names) are deliberately ignored.
var capMarkerRE = regexp.MustCompile("`cap:([a-z][a-z0-9_]*\\.[a-z0-9_.]+)`")

// seedCapRE parses capability values from the canonical role_capabilities seed.
// Mirrors seededCaps() in permissions_test.go:
//
//	INSERT INTO metaldocs.role_capabilities (...) VALUES ('role', 'cap', ...
var seedCapRE = regexp.MustCompile(`(?i)INSERT INTO\s+metaldocs\.role_capabilities[^V]*VALUES\s*\(\s*'[^']*'\s*,\s*'([^']+)'`)

// RunRegistryRules runs the three registry-binding lints. modulesRoot is the
// repo root (the api-lint second arg); paths below are resolved relative to it.
func RunRegistryRules(modulesRoot string, fset *token.FileSet) ([]Violation, error) {
	if modulesRoot == "" {
		return nil, nil
	}
	out := []Violation{}

	inline, err := checkNoInlineCapability(modulesRoot, fset)
	if err != nil {
		return nil, err
	}
	out = append(out, inline...)

	seed, err := checkSeedRegistryParity(modulesRoot)
	if err != nil {
		return nil, err
	}
	out = append(out, seed...)

	wiki, err := checkWikiCapabilityParity(modulesRoot)
	if err != nil {
		return nil, err
	}
	out = append(out, wiki...)

	return out, nil
}

// checkNoInlineCapability bans `Capability("literal")` string-conversion
// expressions (and the qualified `iamdomain.Capability("…")` / `domain.Capability("…")`
// forms) outside the registry definition file and tests. The registry is the
// single source of truth: every referenced capability must be a typed const, so
// a stray Capability("doc.bogus") — which compiles clean — fails here.
func checkNoInlineCapability(modulesRoot string, fset *token.FileSet) ([]Violation, error) {
	out := []Violation{}
	err := filepath.WalkDir(modulesRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			// Skip VCS / tooling / vendor trees: stale agent worktrees under
			// .claude can hold detached copies that would yield phantom
			// violations against code not on this branch.
			switch d.Name() {
			case ".git", ".claude", "node_modules", "vendor":
				return filepath.SkipDir
			}
			return nil
		}
		lower := strings.ToLower(path)
		if !strings.HasSuffix(lower, ".go") || strings.HasSuffix(lower, "_test.go") {
			return nil
		}
		// model.go defines the Capability type + consts (const = "literal"
		// assignments, not conversions); skip it so its definitions are exempt.
		if strings.HasSuffix(lower, filepath.Join("iam", "domain", "model.go")) {
			return nil
		}
		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return err
		}
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) != 1 {
				return true
			}
			if !isCapabilityConversion(call.Fun) {
				return true
			}
			lit, ok := call.Args[0].(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			out = append(out, Violation{
				File:    path,
				Line:    fset.Position(call.Pos()).Line,
				Rule:    "no-inline-capability",
				Message: "inline Capability(" + lit.Value + ") string conversion; reference a typed const from internal/modules/iam/domain instead (single source of truth, ADR 0022)",
			})
			return true
		})
		return nil
	})
	return out, err
}

// isCapabilityConversion reports whether expr is `Capability` or `X.Capability`.
func isCapabilityConversion(fun ast.Expr) bool {
	switch f := fun.(type) {
	case *ast.Ident:
		return f.Name == "Capability"
	case *ast.SelectorExpr:
		return f.Sel.Name == "Capability"
	}
	return false
}

// checkSeedRegistryParity binds the DB role_capabilities seed to the registry
// both directions (ADR 0022 Phase 5, item 3): every seeded capability must be in
// the registry, and every registry capability must be seeded to >=1 role (or be
// explicitly deferred). Wired into the lint binary, not just the unit test.
func checkSeedRegistryParity(modulesRoot string) ([]Violation, error) {
	seedPath := filepath.Join(modulesRoot, "db", "reference-data", "0001_product_reference_data.sql")
	raw, err := os.ReadFile(seedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // no seed under this root (e.g. lint test fixtures); skip
		}
		return nil, err
	}
	content := string(raw)

	seeded := map[iamdomain.Capability]int{} // cap -> 1-based line of first seed
	for _, m := range seedCapRE.FindAllStringSubmatchIndex(content, -1) {
		// m[2],m[3] bound capture group 1.
		capVal := iamdomain.Capability(content[m[2]:m[3]])
		if _, seen := seeded[capVal]; !seen {
			seeded[capVal] = lineOf(content, m[0])
		}
	}

	out := []Violation{}

	// (a) every seeded cap must be in the registry.
	seededSorted := make([]iamdomain.Capability, 0, len(seeded))
	for c := range seeded {
		seededSorted = append(seededSorted, c)
	}
	sort.Slice(seededSorted, func(i, j int) bool { return seededSorted[i] < seededSorted[j] })
	for _, c := range seededSorted {
		if !iamdomain.IsValidCapability(c) {
			out = append(out, Violation{
				File:    seedPath,
				Line:    seeded[c],
				Rule:    "seed-registry-parity",
				Message: "seeded capability " + string(c) + " is not in the registry (validCapabilities); fix the seed typo or add the typed const",
			})
		}
	}

	// (b) every registry cap must be seeded to >=1 role OR explicitly deferred.
	caps := iamdomain.AllCapabilities()
	sort.Slice(caps, func(i, j int) bool { return caps[i] < caps[j] })
	for _, c := range caps {
		if _, ok := seeded[c]; ok {
			continue
		}
		if _, ok := deferredCaps[c]; ok {
			continue
		}
		out = append(out, Violation{
			File:    seedPath,
			Line:    0,
			Rule:    "seed-registry-parity",
			Message: "registry capability " + string(c) + " is seeded to no role; seed it in db/reference-data/0001_product_reference_data.sql or document the deferral in deferredCaps",
		})
	}

	return out, nil
}

// checkWikiCapabilityParity binds `cap:<name>` enforcement-claim markers in the
// wiki authz docs to the registry (ADR 0022 Phase 5, item 4). Catches future
// membership.grant-style drift (a doc claiming a capability that does not exist)
// without false-positiving on prose, filenames, GUC names, or deferred caps —
// only the explicit `cap:` marker is bound.
func checkWikiCapabilityParity(modulesRoot string) ([]Violation, error) {
	out := []Violation{}
	for _, rel := range wikiAuthzDocs {
		path := filepath.Join(modulesRoot, rel)
		raw, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue // doc optional / renamed; not a lint failure
			}
			return nil, err
		}
		content := string(raw)
		for _, m := range capMarkerRE.FindAllStringSubmatchIndex(content, -1) {
			name := iamdomain.Capability(content[m[2]:m[3]])
			if !iamdomain.IsValidCapability(name) {
				out = append(out, Violation{
					File:    path,
					Line:    lineOf(content, m[0]),
					Rule:    "wiki-capability-parity",
					Message: "wiki capability marker cap:" + string(name) + " is not in the registry (validCapabilities); fix the doc or add the typed const",
				})
			}
		}
	}
	return out, nil
}

// lineOf returns the 1-based line number of byte offset off in content.
func lineOf(content string, off int) int {
	if off < 0 || off > len(content) {
		return 0
	}
	return 1 + strings.Count(content[:off], "\n")
}
