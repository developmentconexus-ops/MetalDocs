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
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

// deferredCaps lists registry capabilities intentionally seeded to no tenant
// role (enforced today only via the system_admin tier-2 bypass). Mirrors the
// allow-list in apps/api/cmd/metaldocs-api/permissions_test.go
// (TestEveryCapSeededOrDeferred). A routed-but-unseeded cap goes here with a
// documented deferral.
//
//	distribution.read — minted in mission frontend-screen-completion M2/F2.1c
//	(ADR-0042). Sensitive coverage surface; deliberately NOT seeded to any
//	tenant role by the agent — the operator grants it to roles separately.
//
//	notification.read — minted in mission frontend-screen-completion M3/F3.1
//	(ADR-0043). Self-scope inbox read; deliberately NOT seeded to any tenant
//	role by the agent — the operator grants it to roles separately.
var deferredCaps = map[iamdomain.Capability]struct{}{
	iamdomain.CapDistributionRead: {},
	iamdomain.CapNotificationRead: {},
}

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
func RunRegistryRules(modulesRoot string, fset *token.FileSet, strict bool) ([]Violation, error) {
	if modulesRoot == "" {
		return nil, nil
	}
	out := []Violation{}

	inline, err := checkNoInlineCapability(modulesRoot, fset)
	if err != nil {
		return nil, err
	}
	out = append(out, inline...)

	seed, err := checkSeedRegistryParity(modulesRoot, strict)
	if err != nil {
		return nil, err
	}
	out = append(out, seed...)

	wiki, err := checkWikiCapabilityParity(modulesRoot, strict)
	if err != nil {
		return nil, err
	}
	out = append(out, wiki...)

	binding, err := checkAuthzAreaScopeBinding(modulesRoot, fset, strict)
	if err != nil {
		return nil, err
	}
	out = append(out, binding...)

	rawcap, err := checkNoRawStringCapability(modulesRoot, fset)
	if err != nil {
		return nil, err
	}
	out = append(out, rawcap...)

	rolestr, err := checkNoRoleStringInDelivery(modulesRoot, fset, strict)
	if err != nil {
		return nil, err
	}
	out = append(out, rolestr...)

	tier1, err := checkNoRawStringTier1Authz(modulesRoot, fset)
	if err != nil {
		return nil, err
	}
	out = append(out, tier1...)

	return out, nil
}

// requireCoreFile converts a missing EXPECTED core registry file into a hard
// error under strict mode (production/CI root), while preserving the legitimate
// "unit-test fixture root, file absent" path in non-strict mode. This is the
// fix for the silent-empty swallow (ADR 0022 Phase 13): before, a wrong
// modulesRoot made every join miss and each helper returned an empty set that
// passed trivially. `what` names the file for the operator.
func requireCoreFile(strict bool, what, path string) error {
	if !strict {
		return nil
	}
	return fmt.Errorf("api-lint strict: %s not found at %s — wrong repo-root? (expected the repo root, not internal/modules); a missing core file under a production root is a hard error, not an empty pass (ADR 0022 Phase 13)", what, path)
}

// phantomRoleDialect lists role-name string literals that may not appear in a
// delivery/http authz gate, beyond the canonical Role catalog parsed from
// model.go. All three are TRUE phantoms — decommissioned role names that are not
// canonical anywhere (no Go const, no DB CHECK, no seed, no issuance path). They
// compile clean but name roles that cannot exist:
//
//   - template_author, document_filler: legacy "docx v2" dialect role names
//     (ADR 0022 Phase 9).
//   - reviewer: legacy area role decommissioned in ADR 0022 — removed from the
//     user_process_areas role CHECK + stored procs (migration 0230). Granted zero
//     capabilities; redundant with `approver` + the controlled-document workflow
//     reviewer columns.
//
// They are kept banned so a removed role can never silently reappear in a gate.
// All three are unioned with the parsed canonical Role catalog to form the
// banned set.
var phantomRoleDialect = []string{"template_author", "document_filler", "reviewer"}

// checkNoRoleStringInDelivery is the ADR 0022 Phase 9 core control: it bans a
// role-name string literal in a delivery/http file from appearing in (a) a
// const ValueSpec value, (b) an ==/!= comparison operand, or (c) a switch/case
// clause expression (`switch role { case "system_admin": }`). Authorization is
// the capability model's job; a role string baked into a handler gate is a
// second, incoherent authz dialect (e.g. roleAdmin/roleDocumentFiller gating on
// phantom roles absent from the registry). The banned set is the canonical Role
// catalog parsed from internal/modules/iam/domain/model.go UNION the legacy
// phantom dialect. A role literal used as DATA (composite literal element,
// struct field, function arg) is intentionally NOT flagged — only the three
// authz-gate shapes (const value, equality comparison, switch/case) bite.
// Scoped to non-test .go files whose path contains "/delivery/http/".
func checkNoRoleStringInDelivery(modulesRoot string, fset *token.FileSet, strict bool) ([]Violation, error) {
	banned, err := bannedRoleSet(modulesRoot, fset, strict)
	if err != nil {
		return nil, err
	}
	if len(banned) == 0 {
		// No model.go under this root (e.g. lint test fixtures); nothing to ban.
		return nil, nil
	}

	out := []Violation{}
	walkErr := filepath.WalkDir(modulesRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
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
		if !strings.Contains(filepath.ToSlash(path), "/delivery/http/") {
			return nil
		}
		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return err
		}
		// flagged tracks BasicLit nodes already reported so a literal that is
		// both (it cannot be) is never double-counted across the two passes.
		flagged := map[*ast.BasicLit]struct{}{}
		flag := func(lit *ast.BasicLit) {
			val, uerr := strconv.Unquote(lit.Value)
			if uerr != nil {
				return
			}
			if _, ok := banned[val]; !ok {
				return
			}
			if _, seen := flagged[lit]; seen {
				return
			}
			flagged[lit] = struct{}{}
			out = append(out, Violation{
				File:    path,
				Line:    fset.Position(lit.Pos()).Line,
				Rule:    "no-rolestring-in-delivery",
				Message: "role-name string literal " + lit.Value + " in a delivery/http authz gate; authorize through the capability model (string(iamdomain.CapXxx) via authz.Require), not a role string — phantom/legacy roles compile clean but gate incoherently (ADR 0022 Phase 9)",
			})
		}
		ast.Inspect(file, func(n ast.Node) bool {
			switch e := n.(type) {
			case *ast.GenDecl: // (a) const ValueSpec value (var-declared data is exempt)
				if e.Tok != token.CONST {
					return true
				}
				for _, spec := range e.Specs {
					vs, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}
					for _, v := range vs.Values {
						if lit, ok := v.(*ast.BasicLit); ok && lit.Kind == token.STRING {
							flag(lit)
						}
					}
				}
			case *ast.BinaryExpr: // (b) ==/!= comparison operand
				if e.Op != token.EQL && e.Op != token.NEQ {
					return true
				}
				if lit, ok := e.X.(*ast.BasicLit); ok && lit.Kind == token.STRING {
					flag(lit)
				}
				if lit, ok := e.Y.(*ast.BasicLit); ok && lit.Kind == token.STRING {
					flag(lit)
				}
			case *ast.CaseClause: // (c) switch/case clause expression operand
				for _, expr := range e.List {
					if lit, ok := expr.(*ast.BasicLit); ok && lit.Kind == token.STRING {
						flag(lit)
					}
				}
			}
			return true
		})
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}
	return out, nil
}

// bannedRoleSet builds the role-name set the delivery gate may not name as a
// string literal: the canonical Role values parsed from model.go UNION the
// legacy phantom dialect. Returns an empty map when no model.go is found.
func bannedRoleSet(modulesRoot string, fset *token.FileSet, strict bool) (map[string]struct{}, error) {
	roles, err := parseRoleConsts(modulesRoot, fset, strict)
	if err != nil {
		return nil, err
	}
	if len(roles) == 0 {
		return nil, nil
	}
	banned := make(map[string]struct{}, len(roles)+len(phantomRoleDialect))
	for _, r := range roles {
		banned[string(r)] = struct{}{}
	}
	for _, p := range phantomRoleDialect {
		banned[p] = struct{}{}
	}
	return banned, nil
}

// parseRoleConsts parses internal/modules/iam/domain/model.go and returns the
// canonical Role const-name -> value map (e.g. "RoleSystemAdmin" ->
// "system_admin"). Mirrors parseCapabilityConsts but matches Role-typed consts.
func parseRoleConsts(modulesRoot string, fset *token.FileSet, strict bool) (map[string]iamdomain.Role, error) {
	modelPath := filepath.Join(modulesRoot, "internal", "modules", "iam", "domain", "model.go")
	raw, err := os.ReadFile(modelPath) // #nosec G304 -- path is modulesRoot (CLI-supplied repo root) joined with the fixed literal internal/modules/iam/domain/model.go; not external input.
	if err != nil {
		if os.IsNotExist(err) {
			if cerr := requireCoreFile(strict, "iam/domain/model.go (Role registry)", modelPath); cerr != nil {
				return nil, cerr
			}
			return nil, nil
		}
		return nil, err
	}
	file, err := parser.ParseFile(fset, modelPath, raw, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}
	out := map[string]iamdomain.Role{}
	for _, decl := range file.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.CONST {
			continue
		}
		for _, spec := range gd.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			typeIdent, ok := vs.Type.(*ast.Ident)
			if !ok || typeIdent.Name != "Role" {
				continue
			}
			if len(vs.Names) != 1 || len(vs.Values) != 1 {
				continue
			}
			lit, ok := vs.Values[0].(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				continue
			}
			val, uerr := strconv.Unquote(lit.Value)
			if uerr != nil {
				continue
			}
			out[vs.Names[0].Name] = iamdomain.Role(val)
		}
	}
	return out, nil
}

// checkNoRawStringCapability is the ADR 0022 Phase 8 core control: it bans a raw
// string literal capability argument to authz.Require / authz.RequireAll. This
// closes the fourth dialect the SSOT lint was structurally blind to: no-inline-
// capability bans only the Capability("…") conversion syntax (a *ast.CallExpr),
// and authz-area-scope-binding only inspects the area arg of resolvable typed
// consts — neither rejects a bare `"document.publish"` string passed as the cap. The
// registry is the single source of truth: every cap argument must be a typed
// const (string(iamdomain.CapXxx)), never a literal that can name an unregistered
// (phantom) or misspelled capability and still compile + pass for system_admin.
//
// For authz.Require the capability is the 3rd arg (index 2). For authz.RequireAll
// every argument is treated as a capability position, so any BasicLit string arg
// is flagged. A variable / selector / conversion cap arg (e.g. the `cap` param at
// templates/lifecycle.go, or string(iamdomain.CapXxx)) is NOT a *ast.BasicLit and
// is correctly ignored. Scoped to non-test .go files.
func checkNoRawStringCapability(modulesRoot string, fset *token.FileSet) ([]Violation, error) {
	out := []Violation{}
	err := filepath.WalkDir(modulesRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
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
		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return err
		}
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			pkg, ok := sel.X.(*ast.Ident)
			if !ok || pkg.Name != "authz" {
				return true
			}
			// capArgs are the argument indexes that carry a capability.
			var capArgs []int
			switch sel.Sel.Name {
			case "Require":
				if len(call.Args) < 3 {
					return true
				}
				capArgs = []int{2}
			case "RequireAll":
				for i := range call.Args {
					capArgs = append(capArgs, i)
				}
			default:
				return true
			}
			for _, idx := range capArgs {
				lit, ok := call.Args[idx].(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					continue
				}
				out = append(out, Violation{
					File:    path,
					Line:    fset.Position(call.Pos()).Line,
					Rule:    "no-rawstring-capability",
					Message: "raw-string capability literal " + lit.Value + " passed to authz." + sel.Sel.Name + "; reference a typed const from internal/modules/iam/domain (string(iamdomain.CapXxx)) — the single source of truth (ADR 0022 Phase 8). A raw string can name a phantom/unregistered cap and still compile + pass for system_admin",
				})
			}
			return true
		})
		return nil
	})
	return out, err
}

// checkAuthzAreaScopeBinding is the ADR 0022 Phase 7 core control: it bans
// `authz.Require(ctx, tx, <areaGradeCap>, "tenant")` — an area-grade capability
// (per the typed IsAreaGrade classification) enforced tenant-wide because its
// tier-2 call site passed the literal "tenant" sentinel. This is the missing
// declaration<->enforcement binding (Principle 5): the typed scope map declared
// the cap area-grade, but nothing forced the call site to pass a real area. A
// regression (or a new area-grade write defaulting to "tenant") is now a red build.
//
// LIMITATION (mirrors no-inline-capability's literal-only limitation): the rule
// resolves the capability argument only when it is a typed registry const
// (`string(iamdomain.CapXxx)` / `string(CapXxx)`). A capability passed via a
// variable or parameter (e.g. a `cap string` function param) cannot be resolved
// to a value at lint time and is skipped. Such call sites are covered by the
// per-call-site fixes + tests, not by this static scan. Scoped to authz.Require
// (authz.RequireAll has a different argument shape).
func checkAuthzAreaScopeBinding(modulesRoot string, fset *token.FileSet, strict bool) ([]Violation, error) {
	constToValue, err := parseCapabilityConsts(modulesRoot, fset, strict)
	if err != nil {
		return nil, err
	}
	if len(constToValue) == 0 {
		// No registry under this root (e.g. lint test fixtures); nothing to bind.
		return nil, nil
	}

	out := []Violation{}
	walkErr := filepath.WalkDir(modulesRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
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
		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return err
		}
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) < 4 {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "Require" {
				return true
			}
			pkg, ok := sel.X.(*ast.Ident)
			if !ok || pkg.Name != "authz" {
				return true
			}
			// 4th arg (index 3) is areaCode; only the literal "tenant" sentinel bites.
			areaLit, ok := call.Args[3].(*ast.BasicLit)
			if !ok || areaLit.Kind != token.STRING {
				return true
			}
			if val, uerr := strconv.Unquote(areaLit.Value); uerr != nil || val != "tenant" {
				return true
			}
			// 3rd arg (index 2) is the capability; resolve a typed const to its value.
			capName, ok := capConstName(call.Args[2])
			if !ok {
				return true // variable/unresolvable cap — documented skip
			}
			capVal, known := constToValue[capName]
			if !known || !iamdomain.IsAreaGrade(capVal) {
				return true
			}
			out = append(out, Violation{
				File:    path,
				Line:    fset.Position(call.Pos()).Line,
				Rule:    "authz-area-scope-binding",
				Message: "area-grade capability " + string(capVal) + " enforced with literal \"tenant\" (area filter OFF); pass the resource's real areaCode to authz.Require (ADR 0022 Phase 7 — declared area-grade in capability_scope.go must be enforced area-scoped) or reclassify the cap to ScopeTenant",
			})
			return true
		})
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}
	return out, nil
}

// capConstName extracts the capability const identifier from a tier-2 capability
// argument. Handles the canonical `string(iamdomain.CapXxx)` / `string(CapXxx)`
// conversion and a bare `iamdomain.CapXxx` / `CapXxx`. Returns the const name
// (e.g. "CapDocumentEdit") and whether it was resolvable.
func capConstName(arg ast.Expr) (string, bool) {
	// Unwrap a single-arg string(...) conversion.
	if call, ok := arg.(*ast.CallExpr); ok && len(call.Args) == 1 {
		if id, ok := call.Fun.(*ast.Ident); ok && id.Name == "string" {
			arg = call.Args[0]
		}
	}
	switch e := arg.(type) {
	case *ast.SelectorExpr: // iamdomain.CapXxx / domain.CapXxx
		return e.Sel.Name, true
	case *ast.Ident: // CapXxx (same package)
		return e.Name, true
	}
	return "", false
}

// parseCapabilityConsts parses internal/modules/iam/domain/model.go and returns
// the registry const-name -> Capability-value map (e.g. "CapDocumentEdit" ->
// "document.edit"). The AST guard needs this because Go consts are not
// reflectable: the call site carries the const identifier, not its value.
func parseCapabilityConsts(modulesRoot string, fset *token.FileSet, strict bool) (map[string]iamdomain.Capability, error) {
	modelPath := filepath.Join(modulesRoot, "internal", "modules", "iam", "domain", "model.go")
	raw, err := os.ReadFile(modelPath) // #nosec G304 -- path is modulesRoot (CLI-supplied repo root) joined with the fixed literal internal/modules/iam/domain/model.go; not external input.
	if err != nil {
		if os.IsNotExist(err) {
			if cerr := requireCoreFile(strict, "iam/domain/model.go (Capability registry)", modelPath); cerr != nil {
				return nil, cerr
			}
			return nil, nil
		}
		return nil, err
	}
	file, err := parser.ParseFile(fset, modelPath, raw, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}
	out := map[string]iamdomain.Capability{}
	for _, decl := range file.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.CONST {
			continue
		}
		for _, spec := range gd.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			typeIdent, ok := vs.Type.(*ast.Ident)
			if !ok || typeIdent.Name != "Capability" {
				continue
			}
			if len(vs.Names) != 1 || len(vs.Values) != 1 {
				continue
			}
			lit, ok := vs.Values[0].(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				continue
			}
			val, uerr := strconv.Unquote(lit.Value)
			if uerr != nil {
				continue
			}
			out[vs.Names[0].Name] = iamdomain.Capability(val)
		}
	}
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
func checkSeedRegistryParity(modulesRoot string, strict bool) ([]Violation, error) {
	seedPath := filepath.Join(modulesRoot, "db", "reference-data", "0001_product_reference_data.sql")
	raw, err := os.ReadFile(seedPath) // #nosec G304 -- path is modulesRoot (CLI-supplied repo root) joined with the fixed literal db/reference-data/0001_product_reference_data.sql; not external input.
	if err != nil {
		if os.IsNotExist(err) {
			if cerr := requireCoreFile(strict, "db/reference-data/0001_product_reference_data.sql (role_capabilities seed)", seedPath); cerr != nil {
				return nil, cerr
			}
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
func checkWikiCapabilityParity(modulesRoot string, strict bool) ([]Violation, error) {
	out := []Violation{}
	for _, rel := range wikiAuthzDocs {
		path := filepath.Join(modulesRoot, rel)
		raw, err := os.ReadFile(path) // #nosec G304 -- rel iterates the fixed wikiAuthzDocs literal slice, joined onto modulesRoot (CLI-supplied repo root); not external input.
		if err != nil {
			if os.IsNotExist(err) {
				// In strict mode (production root) these are FIXED, known authz
				// docs — a missing one means a wrong root or an un-tracked rename,
				// either of which must fail loudly rather than silently pass. In
				// non-strict mode (fixture roots) a missing doc is skipped, and if
				// a doc is legitimately renamed update wikiAuthzDocs in the same
				// change (ADR 0022 Phase 13).
				if cerr := requireCoreFile(strict, "wiki authz doc "+rel, path); cerr != nil {
					return nil, cerr
				}
				continue
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

// checkNoRawStringTier1Authz is the ADR 0022 tier-1 companion to
// checkNoRawStringCapability. While no-rawstring-capability covers authz.Require
// / authz.RequireAll (tier-2 application-layer calls), tier-1 delivery gates go
// through the AuthzFunc field `h.authz(r, tenantID, area, cap)` — a method-value
// call where the capability is the 4th argument (index 3). This rule bans a raw
// capability-shaped string literal in that position.
//
// Detection: any call expression whose callee is a SelectorExpr with selector
// name "authz" (i.e. `h.authz(...)` — the canonical delivery-layer pattern) AND
// whose 4th argument is a BasicLit string. A variable / typed-const conversion
// (`string(iamdomain.CapXxx)`) is not a BasicLit and is correctly ignored.
// Scoped to non-test .go files whose path contains "/delivery/http/".
func checkNoRawStringTier1Authz(modulesRoot string, fset *token.FileSet) ([]Violation, error) {
	out := []Violation{}
	err := filepath.WalkDir(modulesRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
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
		if !strings.Contains(filepath.ToSlash(path), "/delivery/http/") {
			return nil
		}
		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return err
		}
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			// Match: X.authz(...) — the AuthzFunc field call pattern.
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "authz" {
				return true
			}
			// The capability is the 4th argument (index 3).
			if len(call.Args) < 4 {
				return true
			}
			lit, ok := call.Args[3].(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			out = append(out, Violation{
				File:    path,
				Line:    fset.Position(call.Pos()).Line,
				Rule:    "no-rawstring-tier1-authz",
				Message: "raw-string capability literal " + lit.Value + " passed as the tier-1 authz action arg; reference a typed const from internal/modules/iam/domain (string(iamdomain.CapXxx)) — a phantom capability compiles clean but permanently locks the route (ADR 0022 F-11)",
			})
			return true
		})
		return nil
	})
	return out, err
}

// lineOf returns the 1-based line number of byte offset off in content.
func lineOf(content string, off int) int {
	if off < 0 || off > len(content) {
		return 0
	}
	return 1 + strings.Count(content[:off], "\n")
}
