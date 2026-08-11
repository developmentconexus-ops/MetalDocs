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

	// ADR 0022 / H-PRE-1 — NO-READONLY-TX-OPTIONS: sql.TxOptions{ReadOnly:
	// true} is banned outright outside test code. Replaces
	// checkAuthzRequireRWTx (retired alongside DoReadOnly in A5.2, Lane E
	// issue #92): that guard flagged one *consequence* of a read-only tx
	// (a DoReadOnly closure invoking authz.Require, whose F8 bypass audit
	// INSERT a read-only tx rejects — ADR 0022 Phase 11). TxRunner has no
	// read-only variant anymore, so the remaining risk is a caller reaching
	// past TxRunner for a raw db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true}).
	// Banning the construct itself, not one call shape built from it, is
	// strictly stronger (unrepresentable beats merely-checked) and stays
	// honest for the whole window until A5.1's later ban on raw .BeginTx(
	// outside internal/platform/db/ makes tx-opening itself unrepresentable.
	// See checkNoReadOnlyTxOptions's doc below for the full reasoning record.
	noReadOnlyTx, err := checkNoReadOnlyTxOptions(modulesRoot, fset)
	if err != nil {
		return nil, err
	}
	out = append(out, noReadOnlyTx...)

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

	// M3 F3.2 (validation-contract.md §2.3) — ASYNC-TENANT-SEED: a worker/jobs
	// write against a tenant-scoped (FORCE RLS) table must sit in a function
	// that also calls SeedTxTenant/SeedTxIdentity, unless allowlisted.
	asyncTenantSeed, err := checkAsyncTenantSeed(modulesRoot, strict)
	if err != nil {
		return nil, err
	}
	out = append(out, asyncTenantSeed...)

	// M7 F7.4 §4.3 — SOLE-RLS-ASYNC-READ: a worker/jobs SELECT against a
	// tenant-scoped (FORCE RLS) table must carry an explicit tenant_id/
	// actor_tenant_id predicate in its own SQL text, unless allowlisted.
	soleRLSRead, err := checkSoleRLSAsyncRead(modulesRoot, strict)
	if err != nil {
		return nil, err
	}
	out = append(out, soleRLSRead...)

	// ADR 0089 step 12 — PROBLEM-DUMP-IMPORT: a package that registers Problem
	// codes but is not linked into cmd/problem-codes-dump contributes nothing to
	// the registry, so every generated artifact silently omits its codes.
	problemDump, err := checkProblemDumpImports(modulesRoot, fset, strict)
	if err != nil {
		return nil, err
	}
	out = append(out, problemDump...)

	// F1.2 Part B — WIRE-NULLABLE-OMITEMPTY: closes the gap that let Part A's
	// drift (route.go QuorumM/DueInDays, instance_read.go
	// CompletedAt/Decision) exist in the first place. SHAPE-NULLABLE-NOT-
	// REQUIRED (spec_rules.go) is spec-only and structurally blind to the Go
	// side; this cross-checks hand-written wire structs under http/ against
	// the spec's nullable-and-required response properties.
	wire, err := checkWireNullableOmitempty(specPath, modulesRoot)
	if err != nil {
		return nil, err
	}
	out = append(out, wire...)

	return out, nil
}

// paginationCodecExemptFiles are the ONLY files allowed to name
// base64.StdEncoding. The rule protects keyset cursors (query-string
// round-tripping), so exemptions are limited to files whose StdEncoding use
// is provably not a cursor:
//   - pagination/cursor.go owns the canonical codec (now RawURLEncoding) and
//     is allow-listed defensively (cursor_test.go is the behavioral guard);
//   - platform/crypto/envelope.go encodes wrapped-DEK/ciphertext blobs stored
//     in DB text columns (M7 crypto-shred) — padded StdEncoding is already the
//     stored-value format, never a query string;
//   - platform/config/tenant_crypto.go decodes the METALDOCS_TENANT_KEK env
//     var, which follows the industry-standard padded-base64 key convention.
var paginationCodecExemptFiles = map[string]bool{
	"internal/platform/pagination/cursor.go":    true,
	"internal/platform/crypto/envelope.go":      true,
	"internal/platform/config/tenant_crypto.go": true,
}

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
		if rel, err := filepath.Rel(modulesRoot, path); err == nil && paginationCodecExemptFiles[filepath.ToSlash(rel)] {
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

// checkNoReadOnlyTxOptions flags any sql.TxOptions{ReadOnly: true} composite
// literal outside test code — addressed (&sql.TxOptions{...}) or not, field
// order irrelevant. TxRunner.Do (internal/platform/db/runner.go) is the sole
// production tx-opening chokepoint and has no read-only variant since
// DoReadOnly was deleted (A5.2, Lane E issue #92); a hand-rolled read-only tx
// bypasses that chokepoint AND reintroduces the exact failure mode DoReadOnly
// used to enable: authz.Require's system_admin/BypassSystem short-circuit
// audits the bypass in-tx with an INSERT (ADR 0022 Phase 11, H-PRE-1), which
// Postgres rejects inside a READ ONLY transaction. Banning the literal
// construct is unconditional — it does not need to trace a call graph to
// authz.Require the way the retired checkAuthzRequireRWTx did, so it also
// catches a future read-only tx that never touches authz.Require but still
// bypasses the chokepoint. Fix is always TxRunner.Do (or, until A5.1's
// BeginTx-outside-runner.go ban lands, BeginTx(ctx, nil)).
//
// The selector qualifier is resolved per file from that file's own imports,
// not hard-coded to the literal identifier "sql": Go lets any file alias
// database/sql to another name (import dbsql "database/sql"), and a
// hard-coded "sql" comparison would let &dbsql.TxOptions{ReadOnly: true}
// pass silently, reopening exactly the gap this rule exists to close.
// Blank (_) and dot (.) imports of database/sql are not resolved here — a
// dot import turns the literal into a bare TxOptions{...} composite lit
// with no SelectorExpr at all, which is a different AST shape this
// selector-based check does not attempt to match; that is a known,
// narrower gap than the one this fix closes, not a regression from it.
func checkNoReadOnlyTxOptions(modulesRoot string, fset *token.FileSet) ([]Violation, error) {
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
		sqlIdents := map[string]bool{}
		for _, imp := range file.Imports {
			importPath, unquoteErr := strconv.Unquote(imp.Path.Value)
			if unquoteErr != nil || importPath != "database/sql" {
				continue
			}
			switch {
			case imp.Name == nil:
				sqlIdents["sql"] = true
			case imp.Name.Name == "_" || imp.Name.Name == ".":
				// Not a selector-qualified reference; see doc comment above.
			default:
				sqlIdents[imp.Name.Name] = true
			}
		}
		ast.Inspect(file, func(n ast.Node) bool {
			lit, ok := n.(*ast.CompositeLit)
			if !ok {
				return true
			}
			sel, ok := lit.Type.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "TxOptions" {
				return true
			}
			if x, ok := sel.X.(*ast.Ident); !ok || !sqlIdents[x.Name] {
				return true
			}
			for _, elt := range lit.Elts {
				kv, ok := elt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				key, ok := kv.Key.(*ast.Ident)
				if !ok || key.Name != "ReadOnly" {
					continue
				}
				val, ok := kv.Value.(*ast.Ident)
				if !ok || val.Name != "true" {
					continue
				}
				out = append(out, Violation{
					File:    path,
					Line:    fset.Position(lit.Pos()).Line,
					Rule:    "no-readonly-tx-options",
					Message: "sql.TxOptions{ReadOnly: true} is banned outside test code — TxRunner has no read-only variant (DoReadOnly deleted A5.2, issue #92); a raw read-only tx rejects authz.Require's F8 bypass audit INSERT (ADR 0022 Phase 11, H-PRE-1). Open the tx with TxRunner.Do instead.",
				})
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
	raw, err := os.ReadFile(path) // #nosec G304 -- path is modulesRoot (CLI-supplied repo root) joined with the fixed literal "scripts/api-lint/tripwire-allowlist.txt"; not external input.
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
