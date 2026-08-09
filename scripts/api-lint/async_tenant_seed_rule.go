package main

// ASYNC-TENANT-SEED — M3 F3.2 (docs/superpowers/milestones/global-maximum-
// remediation/milestone-3-tenancy-chokepoint/validation-contract.md §2.3 —
// binding). Blocking; no reported-only tier (registered in RunCodeRules
// beside F3.1's SEED-CHOKEPOINT).
//
// The 33-table FORCE ROW LEVEL SECURITY backstop only engages for async
// (worker/jobs) processing when the enclosing tx has seeded
// metaldocs.tenant_id via authz.SeedTxTenant (or the actor-bearing
// authz.SeedTxIdentity). This rule AST-scans every non-test .go file for a
// function that both (a) executes a DB write (Exec/ExecContext with a
// string-literal INSERT/UPDATE/DELETE SQL statement) against a table in the
// checked-in tenant-scoped table set (async-tenant-tables.txt) and (b) does
// NOT call SeedTxTenant/SeedTxIdentity anywhere in that same function body —
// unless the call site is on the §2.4 sanctioned allowlist
// (async-tenant-seed-allowlist.txt).
//
// Coverage scope (documented limitation, not a gap): function/handler-local
// AST only — mirrors the M2 TRIPWIRE-ARM-DRIFT technique. Cross-file
// claim->process edges (e.g. a claim step that hands a row to a separately
// defined processing function) are NOT resolved by call-graph analysis; they
// are covered by the allowlist + the negative RLS integration proof instead
// (contract §2.3, recorded bounded defer).

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

// asyncHandlerRoots are the worker/jobs handler package roots this rule
// scans (contract §2.3: "AST-scan the worker/jobs code paths — the async
// binaries' handler packages"), NOT the whole modulesRoot. Sync API-path
// repository code lives outside these roots and is covered by the F3.1
// TxRunner chokepoint (auto-seed from ctx), not by manual SeedTxTenant calls
// — scanning the whole tree would false-positive on every such sync write.
// Repo-relative, forward-slash, matched as a path prefix.
var asyncHandlerRoots = []string{
	"apps/worker",
	"apps/jobs",
	"internal/platform/worker",
	"internal/modules/approval/jobs",
	"internal/modules/notifications/infrastructure",
	"internal/modules/render/fanout",
}

// asyncHandlerFiles are individual files that are async (worker/jobs)
// handlers even though they live inside an otherwise sync-API-path package —
// e.g. RunScheduledPublishJob is a River-job entrypoint colocated in
// internal/modules/approval/application alongside the sync
// decision/publish/submit/obsolete/supersede services (those are called from
// the F3.1-chokepoint-covered API path, not this rule's concern; scanning
// the whole package would false-positive on every one of their repo writes).
var asyncHandlerFiles = map[string]struct{}{
	"internal/modules/approval/application/scheduler_service.go": {},
}

func isAsyncHandlerPath(relSlash string) bool {
	if _, ok := asyncHandlerFiles[relSlash]; ok {
		return true
	}
	for _, root := range asyncHandlerRoots {
		if relSlash == root || strings.HasPrefix(relSlash, root+"/") {
			return true
		}
	}
	return false
}

func asyncTenantSeedAllowlistPath(modulesRoot string) string {
	return filepath.Join(modulesRoot, "scripts", "api-lint", "async-tenant-seed-allowlist.txt")
}

func asyncTenantTablesPath(modulesRoot string) string {
	return filepath.Join(modulesRoot, "scripts", "api-lint", "async-tenant-tables.txt")
}

// loadAsyncTenantSeedAllowlist reads the frozen F3.2 sanctioned-exception
// list. Format mirrors seed-chokepoint-allowlist.txt / tripwire-allowlist.txt:
// `relative/path.go:LINE  # reason`, `#`-comments and blank lines ignored. A
// missing file is not an error — an empty/absent allowlist is a legitimate
// state (nothing sanctioned yet), same rationale as loadSeedChokepointAllowlist.
func loadAsyncTenantSeedAllowlist(modulesRoot string) (map[string]struct{}, error) {
	path := asyncTenantSeedAllowlistPath(modulesRoot)
	raw, err := os.ReadFile(path) // #nosec G304 -- path is modulesRoot (CLI-supplied repo root) joined with the fixed literal "scripts/api-lint/async-tenant-seed-allowlist.txt"; not external input.
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

// loadAsyncTenantTables reads the checked-in tenant-scoped (FORCE RLS) table
// data file — the set of bare table names a write must match to be in scope
// for this rule at all. Kept as a data file (not hardcoded) so the set can't
// silently drift from db/baseline/0001_current_schema.sql (contract §2.3).
func loadAsyncTenantTables(modulesRoot string) (map[string]struct{}, error) {
	path := asyncTenantTablesPath(modulesRoot)
	raw, err := os.ReadFile(path) // #nosec G304 -- path is modulesRoot (CLI-supplied repo root) joined with the fixed literal "scripts/api-lint/async-tenant-tables.txt"; not external input.
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
		out[line] = struct{}{}
	}
	return out, nil
}

// asyncTenantSeedKey is the stable allowlist key for a call site: repo-
// relative forward-slash path + ":" + line number (same shape as
// seedChokepointKey).
func asyncTenantSeedKey(modulesRoot, path string, line int) string {
	rel, err := filepath.Rel(modulesRoot, path)
	if err != nil {
		rel = path
	}
	return filepath.ToSlash(rel) + ":" + strconv.Itoa(line)
}

// funcHasTenantSeedCall reports whether fn's body contains a call whose
// selector name is SeedTxTenant or SeedTxIdentity (matched by selector name
// only, so it is robust to import aliasing — mirrors the SEED-CHOKEPOINT /
// TRIPWIRE-ARM-DRIFT robustness note).
func funcHasTenantSeedCall(fn *ast.FuncDecl) bool {
	found := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if found {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if sel.Sel.Name == "SeedTxTenant" || sel.Sel.Name == "SeedTxIdentity" {
			found = true
			return false
		}
		return true
	})
	return found
}

// funcMutatingWrites collects every Exec/ExecContext call in fn's body whose
// first (Exec) or second (ExecContext) argument is a string-literal SQL
// statement parseable as an INSERT/UPDATE/DELETE, via the same
// parseMutatingTable helper used by checkTripwireArmDrift.
func funcMutatingWrites(fn *ast.FuncDecl, fset *token.FileSet) []mutatingSQLWrite {
	var writes []mutatingSQLWrite
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
		if table, op, ok := parseMutatingTable(sqlText); ok {
			writes = append(writes, mutatingSQLWrite{table: table, op: op, line: fset.Position(call.Pos()).Line})
		}
		return true
	})
	return writes
}

// checkAsyncTenantSeed implements contract §2.3: AST-scan root for .go files
// (skipping _test.go and .gen.go) and, for every function, collect its
// mutating writes against tables in the tenant-scoped set
// (async-tenant-tables.txt). If the function contains no
// SeedTxTenant/SeedTxIdentity call, every such write is a violation unless
// its file:line is on async-tenant-seed-allowlist.txt.
//
// Also emits ASYNC-TENANT-SEED-ALLOWLIST-STALE for any allowlist entry that
// matches no live unseeded tenant-table write site, mirroring
// checkSeedChokepoint's stale-allowlist behavior so the list cannot rot.
func checkAsyncTenantSeed(root string, strict bool) ([]Violation, error) {
	tenantTables, err := loadAsyncTenantTables(root)
	if err != nil {
		return nil, err
	}
	if len(tenantTables) == 0 {
		// No data file under this root (e.g. an isolated lint fixture root) —
		// nothing to check against; not an error (mirrors checkTripwireArmDrift's
		// empty-registry short-circuit).
		return nil, nil
	}

	allow, err := loadAsyncTenantSeedAllowlist(root)
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
			writes := funcMutatingWrites(fn, fset)
			if len(writes) == 0 {
				continue
			}
			seeded := funcHasTenantSeedCall(fn)
			for _, w := range writes {
				if _, isTenantScoped := tenantTables[w.table]; !isTenantScoped {
					continue
				}
				key := asyncTenantSeedKey(root, path, w.line)
				if _, ok := allow[key]; ok {
					matched[key] = true
					continue
				}
				if seeded {
					continue
				}
				out = append(out, Violation{
					File: path,
					Line: w.line,
					Rule: "ASYNC-TENANT-SEED",
					Message: fmt.Sprintf(
						"function %s writes tenant-scoped table %s (op %s) at %s without a SeedTxTenant/SeedTxIdentity call in the same function — the FORCE ROW LEVEL SECURITY backstop is inert for this write; seed the processing tx with authz.SeedTxTenant(ctx, tx, <row tenant>) before the write, or add %s to scripts/api-lint/async-tenant-seed-allowlist.txt with a recorded sanctioned-category reason (M3 F3.2 validation-contract.md §2.2/§2.4)",
						fn.Name.Name, w.table, w.op, key, key,
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
				File:    asyncTenantSeedAllowlistPath(root),
				Line:    0,
				Rule:    "ASYNC-TENANT-SEED-ALLOWLIST-STALE",
				Message: fmt.Sprintf("async-tenant-seed-allowlist.txt entry %s matches no live unseeded tenant-table write site; remove the stale line (M3 F3.2 validation-contract.md §2.4)", key),
			})
		}
	}

	return out, nil
}
