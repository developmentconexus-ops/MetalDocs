package main

// SEED-CHOKEPOINT — M3 F3.1 (docs/superpowers/milestones/global-maximum-
// remediation/milestone-3-tenancy-chokepoint/validation-contract.md §1.5 —
// binding). Blocking; no reported-only tier (registered in RunCodeRules).
//
// The TxRunner chokepoint (internal/platform/db/runner.go) now auto-seeds
// metaldocs.tenant_id + metaldocs.actor_id from the ctx-carried platform
// identity on every Do/DoReadOnly. Any remaining manual
// authz.SeedTxIdentity(...) call site outside the chokepoint file, outside
// the authz package's own definition file, outside _test.go, and outside the
// declared allowlist is census drift: either it should have been collapsed
// (chokepoint now covers it) or it is a genuine distinct-identity exception
// that must be recorded in scripts/api-lint/seed-chokepoint-allowlist.txt.

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

// seedChokepointExemptFiles are repo-relative (forward-slash) paths that may
// call authz.SeedTxIdentity without being flagged or allowlisted: the
// TxRunner chokepoint itself (which inlines the same SQL but never calls
// SeedTxIdentity — exempted defensively) and the primitive's own definition
// file.
var seedChokepointExemptFiles = map[string]struct{}{
	"internal/platform/db/runner.go":        {},
	"internal/modules/iam/authz/context.go": {},
}

func seedChokepointAllowlistPath(modulesRoot string) string {
	return filepath.Join(modulesRoot, "scripts", "api-lint", "seed-chokepoint-allowlist.txt")
}

// loadSeedChokepointAllowlist reads the frozen F3.1 allowlist. Format mirrors
// tripwire-allowlist.txt: `relative/path.go:LINE  # reason`, `#`-comments and
// blank lines ignored. A missing file is not an error (empty allowlist is the
// ideal outcome per contract §1.4) in either strict or non-strict mode — this
// differs from loadTripwireAllowlist because an empty/absent
// seed-chokepoint-allowlist.txt is a legitimate, desired end state, not a
// wrong-repo-root signal.
func loadSeedChokepointAllowlist(modulesRoot string) (map[string]struct{}, error) {
	path := seedChokepointAllowlistPath(modulesRoot)
	raw, err := os.ReadFile(path)
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
		// Strip a trailing `# reason` comment if present.
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

// seedChokepointKey is the stable allowlist key for a call site: repo-
// relative forward-slash path + ":" + line number.
func seedChokepointKey(modulesRoot, path string, line int) string {
	rel, err := filepath.Rel(modulesRoot, path)
	if err != nil {
		rel = path
	}
	return filepath.ToSlash(rel) + ":" + strconv.Itoa(line)
}

// checkSeedChokepoint implements contract §1.5: AST-scan root for .go files
// (skipping _test.go) and flag every authz.SeedTxIdentity(...) call
// expression that is not exempt (chokepoint/definition file), not in a
// _test.go file, and not listed in seed-chokepoint-allowlist.txt. Matching is
// by call selector name "SeedTxIdentity" on any ident/import-aliased
// selector base, so it is robust to import aliasing
// (`authz "metaldocs/internal/modules/iam/authz"` or a differently-named
// alias) — mirrors the "match by call selector name" robustness note in the
// task brief and the pattern already used by checkTripwireArmDrift/
// checkTripwirePairing for authz.Require.
//
// Also emits SEED-CHOKEPOINT-ALLOWLIST-STALE for any allowlist path:line
// entry that matches no live SeedTxIdentity call site, mirroring
// checkTripwirePairing's tripwire-allowlist-stale behavior (code_rules.go) so
// the allowlist cannot rot silently.
func checkSeedChokepoint(root string, strict bool) ([]Violation, error) {
	allow, err := loadSeedChokepointAllowlist(root)
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
		if !strings.HasSuffix(lower, ".go") {
			return nil
		}
		if strings.HasSuffix(lower, "_test.go") {
			return nil // exempt: test fixtures legitimately seed identity directly
		}

		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			rel = path
		}
		relSlash := filepath.ToSlash(rel)
		_, exempt := seedChokepointExemptFiles[relSlash]

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
			if sel.Sel.Name != "SeedTxIdentity" {
				return true
			}
			if exempt {
				return true
			}
			line := fset.Position(call.Pos()).Line
			key := seedChokepointKey(root, path, line)
			if _, ok := allow[key]; ok {
				matched[key] = true
				return true
			}
			out = append(out, Violation{
				File: path,
				Line: line,
				Rule: "SEED-CHOKEPOINT",
				Message: fmt.Sprintf(
					"manual authz.SeedTxIdentity call at %s outside the TxRunner chokepoint (internal/platform/db/runner.go) and outside scripts/api-lint/seed-chokepoint-allowlist.txt — the chokepoint now auto-seeds tenant+actor from ctx for every Do/DoReadOnly; remove this call if the seeded identity is provably the ctx identity, or add it to the allowlist (category A/B) with a recorded reason if it is a genuine distinct-identity exception (M3 F3.1 validation-contract.md §1.3/§1.4)",
					key,
				),
			})
			return true
		})
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}

	// Report stale allowlist entries so the list cannot rot (mirrors
	// tripwire-allowlist-stale in code_rules.go).
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
				File:    seedChokepointAllowlistPath(root),
				Line:    0,
				Rule:    "SEED-CHOKEPOINT-ALLOWLIST-STALE",
				Message: fmt.Sprintf("seed-chokepoint-allowlist.txt entry %s matches no live authz.SeedTxIdentity call site; remove the stale line (M3 F3.1 validation-contract.md §1.4)", key),
			})
		}
	}

	return out, nil
}
