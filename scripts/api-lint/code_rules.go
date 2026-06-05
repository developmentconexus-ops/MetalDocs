package main

// Code-side lints are intentionally single-file scans; they do not build a call graph.

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
)

func RunCodeRules(specPath, modulesRoot string, strict bool) ([]Violation, error) {
	if modulesRoot == "" {
		return nil, nil
	}

	data, err := os.ReadFile(specPath)
	if err != nil {
		return nil, err
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}

	root := doc.Content[0]
	paths := mapGet(root, "paths")
	if paths == nil {
		return nil, fmt.Errorf("%s: missing paths", specPath)
	}

	fset := token.NewFileSet()
	index, err := indexModuleFuncs(modulesRoot, fset)
	if err != nil {
		return nil, err
	}

	out := []Violation{}
	for i := 0; i+1 < len(paths.Content); i += 2 {
		pathVal := paths.Content[i+1]
		for j := 0; j+1 < len(pathVal.Content); j += 2 {
			op := pathVal.Content[j+1]
			opID := scalarValue(mapGet(op, "operationId"))
			if opID == "" {
				continue
			}
			out = append(out, checkAuthzCallPresent(specPath, opID, op, index, fset)...)
		}
	}

	tripwire, err := checkTripwirePairing(modulesRoot, fset, strict)
	if err != nil {
		return nil, err
	}
	out = append(out, tripwire...)

	// ADR 0022 Phase 5 — registry-binding lints (inline-cap ban, seed parity,
	// wiki parity). Reuses the same fset/modulesRoot.
	registry, err := RunRegistryRules(modulesRoot, fset, strict)
	if err != nil {
		return nil, err
	}
	out = append(out, registry...)

	return out, nil
}

type indexedFunc struct {
	File string
	Decl *ast.FuncDecl
}

func indexModuleFuncs(modulesRoot string, fset *token.FileSet) (map[string][]indexedFunc, error) {
	out := map[string][]indexedFunc{}
	err := filepath.WalkDir(modulesRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			// Skip VCS / tooling / vendor / fixture trees (match the registry
			// walkers). A stale agent worktree under .claude holds detached code
			// copies whose duplicate receiver methods would otherwise produce
			// spurious "multiple handler matches" noise (ADR 0022 Phase 11 F5).
			switch d.Name() {
			case ".git", ".claude", "node_modules", "vendor", "testdata":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(path), ".go") || strings.HasSuffix(strings.ToLower(path), "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return err
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil {
				continue
			}
			out[fn.Name.Name] = append(out[fn.Name.Name], indexedFunc{File: path, Decl: fn})
		}
		return nil
	})
	return out, err
}

// checkAuthzCallPresent is DORMANT by design (ADR 0022 Phase 11 F6 — documented,
// not deleted). It only fires for an operation that declares x-authz-area or
// x-authz-custom:true WITHOUT x-authz-skip-area, and then expects the operationId-
// named handler to itself call authz.Require(req.Body.X | req.<Op>Params.X). No
// MetalDocs module matches that shape: tier-2 area enforcement lives in tx-layer
// services where the area is DB-derived (un-spoofable, per ADR 0007 tx-coupling),
// not request-supplied, and every area-grade op therefore carries x-authz-skip-area
// (which silences this rule) — so the rule's live count is 0 and has been across
// all phases. Activating it for real requires a call-graph / "source: derived"
// lint-engine rewrite (tracked as Phase 5+ residual), which is out of scope here.
// It is kept (a) so the x-authz-custom: true escape hatch keeps working if a future
// codegen handler does inline authz.Require, and (b) as the documented seam for that
// rewrite. Do NOT treat its 0-count as "all area ops are statically verified" — the
// authz-area-scope-binding AST guard (registry_rules.go) is the real per-call-site
// binding; this is a complementary, currently-unreachable spec-shape check.
func checkAuthzCallPresent(specPath, opID string, op *yaml.Node, index map[string][]indexedFunc, fset *token.FileSet) []Violation {
	// Gate: rule applies only when the op declares x-authz-area or x-authz-custom: true.
	// x-authz-skip-area silences the rule explicitly (also implies the gate is open).
	if mapGet(op, "x-authz-skip-area") != nil {
		return nil
	}
	hasArea := mapGet(op, "x-authz-area") != nil
	custom := strings.EqualFold(scalarValue(mapGet(op, "x-authz-custom")), "true")
	if !hasArea && !custom {
		return nil
	}

	handlerName := pascalCase(opID)
	candidates := index[handlerName]
	if len(candidates) == 0 {
		return []Violation{{
			File:    specPath,
			Line:    op.Line,
			Rule:    "authz-call-present",
			Message: fmt.Sprintf("handler %s not found for operation %s", handlerName, opID),
		}}
	}
	if len(candidates) > 1 {
		fmt.Printf("warning: multiple handler matches for %s (%s), using first\n", opID, handlerName)
	}

	fn := candidates[0].Decl
	expectedFn, expectedExpr := expectedAuthzCall(opID, op)
	hasAny, hasExpectedFn, matchedArg, actualArg := inspectAuthzCalls(fn, fset, expectedFn, expectedExpr)

	if custom {
		if !hasAny {
			return []Violation{{
				File:    candidates[0].File,
				Line:    fset.Position(fn.Pos()).Line,
				Rule:    "authz-call-present",
				Message: fmt.Sprintf("handler %s does not call authz.Require or authz.RequireAll", handlerName),
			}}
		}
		return nil
	}

	if !hasExpectedFn {
		return []Violation{{
			File:    candidates[0].File,
			Line:    fset.Position(fn.Pos()).Line,
			Rule:    "authz-call-present",
			Message: fmt.Sprintf("handler %s does not call authz.%s", handlerName, expectedFn),
		}}
	}

	if matchedArg {
		return nil
	}

	if actualArg == "" {
		actualArg = "<missing>"
	}
	return []Violation{{
		File:    candidates[0].File,
		Line:    fset.Position(fn.Pos()).Line,
		Rule:    "authz-call-present",
		Message: fmt.Sprintf("handler %s calls authz.%s with arg %s; expected %s", handlerName, expectedFn, actualArg, expectedExpr),
	}}
}

func inspectAuthzCalls(fn *ast.FuncDecl, fset *token.FileSet, expectedFn, expectedExpr string) (hasAny, hasExpectedFn, matchedArg bool, actualArg string) {
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		x, ok := sel.X.(*ast.Ident)
		if !ok || x.Name != "authz" {
			return true
		}
		if sel.Sel.Name != "Require" && sel.Sel.Name != "RequireAll" {
			return true
		}
		hasAny = true
		if sel.Sel.Name != expectedFn {
			return true
		}
		hasExpectedFn = true
		if len(call.Args) == 0 {
			if actualArg == "" {
				actualArg = "<missing>"
			}
			return true
		}
		arg := renderExpr(fset, call.Args[len(call.Args)-1])
		if arg == expectedExpr {
			matchedArg = true
			return false
		}
		if actualArg == "" {
			actualArg = arg
		}
		return true
	})
	return hasAny, hasExpectedFn, matchedArg, actualArg
}

func expectedAuthzCall(opID string, op *yaml.Node) (fnName, expr string) {
	area := mapGet(op, "x-authz-area")
	source := strings.ToLower(strings.TrimSpace(scalarValue(mapGet(area, "source"))))
	if source == "" {
		source = "body"
	}
	field := strings.TrimSpace(scalarValue(mapGet(area, "field")))
	multi := strings.EqualFold(scalarValue(mapGet(area, "multi")), "true")

	fnName = "Require"
	if multi {
		fnName = "RequireAll"
	}

	if source == "path" {
		expr = "req." + pascalCase(opID) + "Params"
		if field != "" {
			expr += "." + pascalCasePath(field)
		}
		return fnName, expr
	}

	expr = "req.Body"
	if field != "" {
		expr += "." + pascalCasePath(field)
	}
	return fnName, expr
}

func pascalCasePath(path string) string {
	parts := strings.Split(path, ".")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if strings.TrimSpace(p) == "" {
			continue
		}
		out = append(out, pascalCase(p))
	}
	return strings.Join(out, ".")
}

// pascalCase mirrors oapi-codegen's method-name derivation for operation IDs.
//
// Behaviour:
//   - If the input already has no separators (`_`, `-`, `.`), preserve internal
//     case and just uppercase the leading rune. This matches camelCase op-ids
//     like `listTemplatesV2` → `ListTemplatesV2` and initialism-rich names like
//     `recordMDDMShadowDiff` → `RecordMDDMShadowDiff`.
//   - If the input has separators (e.g. `area_code`, `zone.area_code`), split,
//     lowercase each segment, then uppercase the first rune of each.
//
// TODO(initialisms): for snake_case fields like `template_id`, real codegen
// can emit `TemplateID` rather than `TemplateId` depending on initialism
// config. Revisit once a real handler trips this.
func pascalCase(s string) string {
	if s == "" {
		return ""
	}
	if !strings.ContainsAny(s, "_-.") {
		r := []rune(s)
		r[0] = unicode.ToUpper(r[0])
		return string(r)
	}
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == '.'
	})
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		r := []rune(strings.ToLower(p))
		r[0] = unicode.ToUpper(r[0])
		b.WriteString(string(r))
	}
	return b.String()
}

func renderExpr(fset *token.FileSet, expr ast.Expr) string {
	var b bytes.Buffer
	_ = printer.Fprint(&b, fset, expr)
	return b.String()
}

// checkTripwirePairing flags repository functions that run mutating SQL without
// an authz.Require call in the same function body. It is a single-file scan with
// no call graph, so it cannot see tier-2 enforcement that lives one layer up (the
// tx-layer service that owns the authz decision and passes the tx down). The
// known-legitimate false positives are frozen in scripts/api-lint/tripwire-
// allowlist.txt so the LIVE tripwire count is 0 and any NEW violation is a hard
// red (ADR 0022 Phase 11 F5). A stale allow-list entry (one that no longer
// matches any live violation) is itself reported so the list cannot rot.
//
// The walker skips .git/.claude/node_modules/vendor/testdata (matching the
// registry walkers): a stale agent worktree under .claude held ~26 phantom
// duplicate violations against code not on this branch, and the api-lint testdata
// fixtures (intentionally un-paired repositories) added one more.
func checkTripwirePairing(modulesRoot string, fset *token.FileSet, strict bool) ([]Violation, error) {
	allow, err := loadTripwireAllowlist(modulesRoot, strict)
	if err != nil {
		return nil, err
	}
	matched := make(map[string]bool, len(allow))

	out := []Violation{}
	walkErr := filepath.WalkDir(modulesRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", ".claude", "node_modules", "vendor", "testdata":
				return filepath.SkipDir
			}
			return nil
		}
		name := strings.ToLower(filepath.Base(path))
		if !strings.Contains(name, "repository") {
			return nil
		}
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return err
		}

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			hasAuthz, hasMutatingSQL := scanTripwireFunc(fn)
			if !hasMutatingSQL || hasAuthz {
				continue
			}
			key := tripwireKey(modulesRoot, path, fn.Name.Name)
			if _, ok := allow[key]; ok {
				matched[key] = true
				continue
			}
			out = append(out, Violation{
				File:    path,
				Line:    fset.Position(fn.Pos()).Line,
				Rule:    "tripwire-pairing",
				Message: fmt.Sprintf("mutating SQL in %s without authz.Require call (new violation: add tier-2 authz.Require, or if enforced one layer up, allow-list %s in scripts/api-lint/tripwire-allowlist.txt)", fn.Name.Name, key),
			})
		}
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}

	// Report stale allow-list entries so the baseline cannot rot: an entry listed
	// but matched by no live violation means the function was fixed or removed and
	// its line must be deleted.
	if len(allow) > 0 {
		stale := make([]string, 0)
		for key := range allow {
			if !matched[key] {
				stale = append(stale, key)
			}
		}
		sort.Strings(stale)
		for _, key := range stale {
			out = append(out, Violation{
				File:    tripwireAllowlistPath(modulesRoot),
				Line:    0,
				Rule:    "tripwire-allowlist-stale",
				Message: fmt.Sprintf("allow-list entry %s matches no live tripwire violation; remove the stale line (ADR 0022 Phase 11 F5)", key),
			})
		}
	}
	return out, nil
}

// tripwireKey is the stable allow-list key for a flagged function: the repo-
// relative forward-slash path plus the function name. Line-independent so the
// baseline survives edits above the function.
func tripwireKey(modulesRoot, path, fn string) string {
	rel, err := filepath.Rel(modulesRoot, path)
	if err != nil {
		rel = path
	}
	return filepath.ToSlash(rel) + "|" + fn
}

func tripwireAllowlistPath(modulesRoot string) string {
	return filepath.Join(modulesRoot, "scripts", "api-lint", "tripwire-allowlist.txt")
}

// loadTripwireAllowlist reads the frozen tripwire baseline. In strict mode a
// missing allow-list is a HARD ERROR: under a production repo root the file MUST
// exist, and treating it as empty is exactly what let a wrong modulesRoot pass
// trivially before (ADR 0022 Phase 13 — the silent-empty swallow). In non-strict
// mode (unit-test fixture roots under testdata/, where the allow-list is
// legitimately absent) a missing file still yields an empty set so fixtures keep
// reporting their intentional violations.
func loadTripwireAllowlist(modulesRoot string, strict bool) (map[string]struct{}, error) {
	path := tripwireAllowlistPath(modulesRoot)
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if strict {
				return nil, fmt.Errorf("api-lint strict: tripwire allow-list not found at %s — wrong repo-root? (expected the repo root, not internal/modules); a missing core file under a production root is a hard error, not an empty pass (ADR 0022 Phase 13)", path)
			}
			return map[string]struct{}{}, nil
		}
		return nil, err
	}
	out := map[string]struct{}{}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out[line] = struct{}{}
	}
	return out, nil
}

func scanTripwireFunc(fn *ast.FuncDecl) (hasAuthz, hasMutatingSQL bool) {
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if x, ok := sel.X.(*ast.Ident); ok && x.Name == "authz" && (sel.Sel.Name == "Require" || sel.Sel.Name == "RequireAll") {
			hasAuthz = true
		}
		// db.Exec(sql, args...)            → SQL is Args[0]
		// db.ExecContext(ctx, sql, args...)→ SQL is Args[1]
		var sqlIdx int
		switch sel.Sel.Name {
		case "Exec":
			sqlIdx = 0
		case "ExecContext":
			sqlIdx = 1
		default:
			return true
		}
		if len(call.Args) <= sqlIdx {
			return true
		}
		raw, ok := call.Args[sqlIdx].(*ast.BasicLit)
		if !ok || raw.Kind != token.STRING {
			return true
		}
		sqlText, err := strconv.Unquote(raw.Value)
		if err != nil {
			sqlText = raw.Value
		}
		s := strings.ToUpper(sqlText)
		if strings.Contains(s, "INSERT INTO") || strings.Contains(s, "UPDATE") || strings.Contains(s, "DELETE FROM") {
			hasMutatingSQL = true
		}
		return true
	})
	return hasAuthz, hasMutatingSQL
}
