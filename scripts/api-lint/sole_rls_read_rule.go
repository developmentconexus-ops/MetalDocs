package main

// SOLE-RLS-ASYNC-READ — M7 F7.4 §4.3 (carried from M6 F6.4's fix-feature
// false-negative: the review-due ports had a tenant-scoped read relying
// solely on RLS, closed by adding an explicit tenant_id predicate — this rule
// makes that class a static, blocking guard instead of a one-off fix).
// Blocking; no reported-only tier (registered in RunCodeRules beside M3's
// ASYNC-TENANT-SEED).
//
// M3's ASYNC-TENANT-SEED covers async WRITES: a worker/jobs mutating write
// against a FORCE-RLS tenant-scoped table must sit in a function that also
// seeds metaldocs.tenant_id via SeedTxTenant/SeedTxIdentity, or the backstop
// is inert. This rule covers the READ side of the same seam. The
// tenant_isolation policy carries a NULL-GUC-bypass branch —
// `(NULLIF(current_setting('metaldocs.tenant_id',true),'') IS NULL) OR (col =
// ...)` — so a SELECT against a tenant-scoped table, issued from an async
// processing tx whose GUC was never seeded (or was seeded for a DIFFERENT
// tenant than the row set the caller actually wants), silently returns ALL
// tenants' rows instead of erroring. Belt-and-suspenders: an async
// tenant-scoped read must carry an explicit `tenant_id` or `actor_tenant_id`
// predicate in its own SQL text — the exact class M6 F6.4 fixed on the
// schedule-publish / review-due ports — so correctness does not depend
// solely on the GUC having been seeded correctly upstream.
//
// Scope: this rule AST-scans the SAME async handler roots as
// checkAsyncTenantSeed (asyncHandlerRoots / asyncHandlerFiles, defined in
// async_tenant_seed_rule.go) — NOT the whole modulesRoot. Sync API-path reads
// are covered by the F3.1 TxRunner GUC auto-seed chokepoint (ctx-derived,
// always correct for the requesting actor's own tenant), not by an explicit
// per-query predicate; scanning the whole tree would false-positive on every
// such sync read. That is a deliberate M3 design boundary and this rule does
// not contradict it.
//
// Coverage scope (documented limitation, matching ASYNC-TENANT-SEED):
// function-local AST only — the SQL string literal must be the call's first
// (Query/QueryRow) or second (QueryContext/QueryRowContext) argument. Query
// strings built via fmt.Sprintf/concatenation are not resolved (same
// limitation class as parseMutatingTable's write-side scan).

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
)

var sqlSelectFromRE = regexp.MustCompile(`(?is)SELECT\b.*?\bFROM\s+([a-zA-Z_][a-zA-Z0-9_.]*)`)

// parseSelectTable extracts the bare table name from a SELECT ... FROM
// <table> string, mirroring parseMutatingTable's approach for INSERT/UPDATE/
// DELETE. Returns ok=false if the statement is not a recognized single SELECT
// referencing exactly one leading FROM table.
func parseSelectTable(sql string) (table string, ok bool) {
	m := sqlSelectFromRE.FindStringSubmatch(sql)
	if m == nil {
		return "", false
	}
	return bareTableName(m[1]), true
}

// hasExplicitTenantPredicate reports whether sql's text contains an explicit
// tenant predicate token — `tenant_id` or `actor_tenant_id` — anywhere in the
// statement (contract §4.3: "contains NEITHER the token tenant_id NOR
// actor_tenant_id"). Case-insensitive to be robust to hand-written SQL style.
func hasExplicitTenantPredicate(sql string) bool {
	lower := strings.ToLower(sql)
	return strings.Contains(lower, "tenant_id") || strings.Contains(lower, "actor_tenant_id")
}

// soleRLSReadAllowlistPath returns the sanctioned-exception list path,
// mirroring asyncTenantSeedAllowlistPath.
func soleRLSReadAllowlistPath(modulesRoot string) string {
	return filepath.Join(modulesRoot, "scripts", "api-lint", "sole-rls-read-allowlist.txt")
}

// loadSoleRLSReadAllowlist reads the sanctioned-exception list. Format and
// semantics mirror loadAsyncTenantSeedAllowlist exactly: `relative/path.go:LINE
// # reason`, `#`-comments and blank lines ignored, trailing inline comments
// stripped, missing file is not an error.
func loadSoleRLSReadAllowlist(modulesRoot string) (map[string]struct{}, error) {
	path := soleRLSReadAllowlistPath(modulesRoot)
	raw, err := os.ReadFile(path) // #nosec G304 -- path is modulesRoot (CLI-supplied repo root) joined with the fixed literal "scripts/api-lint/sole-rls-read-allowlist.txt"; not external input.
	if err != nil {
		if os.IsNotExist(err) {
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
		if i := strings.Index(line, "#"); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		if line == "" {
			continue
		}
		out[filepath.ToSlash(line)] = struct{}{}
	}
	return out, nil
}

// soleRLSReadKey is the stable allowlist key for a call site: repo-relative
// forward-slash path + ":" + line number (same shape as asyncTenantSeedKey).
func soleRLSReadKey(modulesRoot, path string, line int) string {
	rel, err := filepath.Rel(modulesRoot, path)
	if err != nil {
		rel = path
	}
	return filepath.ToSlash(rel) + ":" + strconv.Itoa(line)
}

// soleRLSRead is one Query/QueryContext/QueryRow/QueryRowContext call site
// whose first string-literal SQL argument is a SELECT against a tenant-scoped
// table with no explicit tenant predicate in its own SQL text.
type soleRLSRead struct {
	table string
	line  int
}

// funcSoleRLSReads collects every Query/QueryContext/QueryRow/QueryRowContext
// call in fn's body whose SQL argument is a SELECT against a table in
// tenantTables and whose SQL text carries neither `tenant_id` nor
// `actor_tenant_id`.
func funcSoleRLSReads(fn *ast.FuncDecl, fset *token.FileSet, tenantTables map[string]struct{}) []soleRLSRead {
	var reads []soleRLSRead
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		var sqlIdx int
		switch sel.Sel.Name {
		case "Query", "QueryRow":
			sqlIdx = 0
		case "QueryContext", "QueryRowContext":
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
		table, ok := parseSelectTable(sqlText)
		if !ok {
			return true
		}
		if _, isTenantScoped := tenantTables[table]; !isTenantScoped {
			return true
		}
		if hasExplicitTenantPredicate(sqlText) {
			return true
		}
		reads = append(reads, soleRLSRead{table: table, line: fset.Position(call.Pos()).Line})
		return true
	})
	return reads
}

// checkSoleRLSAsyncRead implements M7 F7.4 §4.3: AST-scan the async handler
// roots (asyncHandlerRoots/asyncHandlerFiles) for .go files (skipping
// _test.go and .gen.go) and, for every function, collect its
// Query/QueryContext/QueryRow/QueryRowContext calls whose SQL is a SELECT
// against a table in the tenant-scoped set (async-tenant-tables.txt — REUSES
// M3's data file) with no explicit tenant_id/actor_tenant_id token anywhere
// in the SQL text. Every such read is a violation unless its file:line is on
// sole-rls-read-allowlist.txt.
func checkSoleRLSAsyncRead(root string, strict bool) ([]Violation, error) {
	tenantTables, err := loadAsyncTenantTables(root)
	if err != nil {
		return nil, err
	}
	if len(tenantTables) == 0 {
		// No data file under this root (e.g. an isolated lint fixture root) —
		// nothing to check against; not an error (mirrors checkAsyncTenantSeed's
		// empty-registry short-circuit).
		return nil, nil
	}

	allow, err := loadSoleRLSReadAllowlist(root)
	if err != nil {
		return nil, err
	}
	matched := make(map[string]bool, len(allow))

	fset := token.NewFileSet()
	out := []Violation{}

	walkErr := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
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
		if !strings.HasSuffix(lower, ".go") || strings.HasSuffix(lower, "_test.go") || strings.HasSuffix(lower, ".gen.go") {
			return nil
		}

		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			rel = path
		}
		if !isAsyncHandlerPath(filepath.ToSlash(rel)) {
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
			reads := funcSoleRLSReads(fn, fset, tenantTables)
			for _, r := range reads {
				key := soleRLSReadKey(root, path, r.line)
				if _, ok := allow[key]; ok {
					matched[key] = true
					continue
				}
				out = append(out, Violation{
					File: path,
					Line: r.line,
					Rule: "SOLE-RLS-ASYNC-READ",
					Message: fmt.Sprintf(
						"function %s reads tenant-scoped table %s at %s via a SELECT with no explicit tenant_id/actor_tenant_id predicate — the read relies solely on the tenant_isolation RLS policy, whose NULL-GUC branch returns all tenants' rows if the enclosing async tx's GUC was not seeded (or was seeded for a different tenant); add an explicit tenant_id/actor_tenant_id predicate to the SQL, or add %s to scripts/api-lint/sole-rls-read-allowlist.txt with a recorded sanctioned reason (M7 F7.4 §4.3)",
						fn.Name.Name, r.table, key, key,
					),
				})
			}
		}
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}

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
				File:    soleRLSReadAllowlistPath(root),
				Line:    0,
				Rule:    "SOLE-RLS-ASYNC-READ-ALLOWLIST-STALE",
				Message: fmt.Sprintf("sole-rls-read-allowlist.txt entry %s matches no live unpredicated tenant-table read site; remove the stale line (M7 F7.4 §4.3)", key),
			})
		}
	}

	return out, nil
}
