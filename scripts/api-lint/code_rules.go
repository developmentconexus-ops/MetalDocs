package main

// Code-side lints are intentionally single-file scans; they do not build a call graph.

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// RunCodeRules runs the code-side lints that read the Go module tree:
//   - tripwire-pairing: mutating SQL in a repository must pair with authz.Require
//     (single-file scan; legitimate one-layer-up enforcement is allow-listed).
//   - the ADR 0022 registry-binding lints (RunRegistryRules).
//
// The former `authz-call-present` spec→handler scan was DELETED in api-contract-
// hardening Phase F (FD-1 = Option A). It was dormant by design (0 hits across
// every phase): it expected each area op's handler to itself call
// authz.Require(req.Body.AreaCode), but MetalDocs derives the area from the DB row
// inside the tx (un-spoofable, ADR 0007 tx-coupling) and enforces it via the
// Postgres tripwire (migration 0142b) + the tier-2 authz.Require in the tx-layer
// service. The rule modelled the wrong architecture, so every area op carried a
// negative `x-authz-skip-area` escape hatch. Phase F replaces those negative
// markers with honest positive ones (`x-authz-area: {source: tx, derived_from}` /
// `x-authz-area-none`) validated by AUTHZ-DRIFT in spec_rules.go. The real static
// guarantees that remain: tripwire-pairing (below), the authz-area-scope-binding
// AST guard in registry_rules.go, and the DB trigger.
func RunCodeRules(specPath, modulesRoot string, strict bool) ([]Violation, error) {
	if modulesRoot == "" {
		return nil, nil
	}

	fset := token.NewFileSet()

	out := []Violation{}
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

	// Family 2 · B2 — keyset cursors must share one base64 dialect so the list
	// endpoints cannot diverge again.
	codec, err := checkPaginationCodec(modulesRoot, fset)
	if err != nil {
		return nil, err
	}
	out = append(out, codec...)

	// ADR 0022 — authz.Require must run in a read-WRITE tx (the F8 bypass audit
	// INSERTs). No DoReadOnly closure may invoke a tier-2 require.
	rwTx, err := checkAuthzRequireRWTx(modulesRoot, fset)
	if err != nil {
		return nil, err
	}
	out = append(out, rwTx...)

	// M2 F2.1 (validation-contract.md §1.5) — TRIPWIRE-ARM-PARITY: TripwireArms
	// caps must be registry-real and RenderMigration() must byte-equal the
	// committed 0271 migration.
	parity, err := checkTripwireArmParity(modulesRoot, strict)
	if err != nil {
		return nil, err
	}
	out = append(out, parity...)

	// M2 F2.1 (validation-contract.md §1.5) — TRIPWIRE-ARM-DRIFT: a function
	// that asserts a cap and writes a gated table must assert an arm-member cap
	// for that table+op (the incident-class catcher; function-local generalization
	// of checkTripwirePairing above from presence to arm-membership).
	drift, err := checkTripwireArmDrift(modulesRoot, fset, strict)
	if err != nil {
		return nil, err
	}
	out = append(out, drift...)

	// M3 F3.1 (validation-contract.md §1.5) — SEED-CHOKEPOINT: the TxRunner
	// chokepoint auto-seeds tenant+actor from ctx; any remaining manual
	// authz.SeedTxIdentity call outside the chokepoint/definition files and
	// outside the declared allowlist is census drift.
	seedChokepoint, err := checkSeedChokepoint(modulesRoot, strict)
	if err != nil {
		return nil, err
	}
	out = append(out, seedChokepoint...)

	return out, nil
}

// requireSelectors are the tier-2 enforcement call names a DoReadOnly closure
// must never contain. "Require"/"RequireAll" cover direct authz.Require calls;
// the lowercase "require" covers the application-service seam field
// (authzRequireFunc, e.g. tokens.Service.require) that delegates to authz.Require
// without naming the package — name-based AST scans would otherwise miss it.
var requireSelectors = map[string]struct{}{
	"Require":    {},
	"RequireAll": {},
	"require":    {},
}

// checkAuthzRequireRWTx flags any DoReadOnly(...) call whose closure body invokes
// a tier-2 require (authz.Require / the require seam). authz.Require's
// system_admin & BypassSystem short-circuits audit the bypass in-tx with an
// INSERT (ADR 0022 Phase 11 F8, fail-closed); a Postgres READ ONLY transaction
// rejects that INSERT, so the path 500s the moment the actor is an admin while
// staying latent for everyone else. DoReadOnly is the single read-only-tx
// chokepoint in the tree, so this static guard fully covers the regression
// surface with zero runtime cost. The fix is always DoReadOnly → Do (or
// BeginTx(ctx, nil)). Single-file AST scan, matching the other code rules.
func checkAuthzRequireRWTx(modulesRoot string, fset *token.FileSet) ([]Violation, error) {
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
		lower := strings.ToLower(path)
		if !strings.HasSuffix(lower, ".go") || strings.HasSuffix(lower, "_test.go") || strings.HasSuffix(lower, ".gen.go") {
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
			if !ok || sel.Sel.Name != "DoReadOnly" {
				return true
			}
			for _, arg := range call.Args {
				lit, ok := arg.(*ast.FuncLit)
				if !ok || lit.Body == nil {
					continue
				}
				if closureCallsRequire(lit.Body) {
					out = append(out, Violation{
						File:    path,
						Line:    fset.Position(call.Pos()).Line,
						Rule:    "authz-require-rw-tx",
						Message: "authz.Require invoked inside a DoReadOnly closure — the F8 bypass audit INSERTs and a READ ONLY tx rejects it (ADR 0022 Phase 11). Open the tx with Do or BeginTx(ctx, nil), not DoReadOnly.",
					})
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

// closureCallsRequire reports whether a function body contains a tier-2 require
// call (see requireSelectors).
func closureCallsRequire(body *ast.BlockStmt) bool {
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if _, hit := requireSelectors[sel.Sel.Name]; hit {
			found = true
			return false
		}
		return true
	})
	return found
}

// paginationCursorFile is the ONE file allowed to name base64.StdEncoding's
// sibling — actually it owns the canonical codec and now uses RawURLEncoding, so
// it is allow-listed defensively in case a future edit reaches for the wrong
// dialect there (the unit tests in cursor_test.go are the behavioral guard).
const paginationCursorFile = "internal/platform/pagination/cursor.go"

// checkPaginationCodec flags any non-generated source that references
// base64.StdEncoding. Keyset cursors must use the shared URL-safe codec in
// internal/platform/pagination (RawURLEncoding); a StdEncoding cursor would be
// padding-/+/-fragile in a query string and incompatible with the shared
// helpers. Generated *.gen.go files are exempt: oapi-codegen uses StdEncoding to
// gunzip embedded specs, which is unrelated to cursors and contract-owned.
func checkPaginationCodec(modulesRoot string, fset *token.FileSet) ([]Violation, error) {
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
		lower := strings.ToLower(path)
		// _test.go is exempt: test fixtures legitimately reference base64.StdEncoding
		// to build/decode expected values and never participate in the runtime cursor
		// path the rule protects.
		if !strings.HasSuffix(lower, ".go") || strings.HasSuffix(lower, "_test.go") || strings.HasSuffix(lower, ".gen.go") {
			return nil
		}
		// Anchored repo-relative match (mirrors tripwireKey): a bare HasSuffix could
		// false-exempt any file whose path merely ends in cursor.go.
		if rel, err := filepath.Rel(modulesRoot, path); err == nil && filepath.ToSlash(rel) == paginationCursorFile {
			return nil
		}
		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return err
		}
		ast.Inspect(file, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "StdEncoding" {
				return true
			}
			x, ok := sel.X.(*ast.Ident)
			if !ok || x.Name != "base64" {
				return true
			}
			out = append(out, Violation{
				File:    path,
				Line:    fset.Position(sel.Pos()).Line,
				Rule:    "pagination-codec",
				Message: "base64.StdEncoding outside internal/platform/pagination/cursor.go — keyset cursors must use the shared URL-safe codec (pagination.EncodeCursor/DecodeCursor); StdEncoding is query-string-fragile (Family 2 · B2)",
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
