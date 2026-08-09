package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strings"
)

// testdb-bypass-guard closes a class of defect that has recurred at least
// five times: a `_test.go` file reads DATABASE_URL/METALDOCS_DATABASE_URL
// out of the environment and opens it directly (sql.Open, or a raw pgx entry
// point — pgx.Connect, pgxpool.New, pgxpool.Connect) instead of going through
// tests/integration/testdb.Open (ADR 0034). Locally that works, because a
// developer's own database was populated by
// scripts/dev-bootstrap-baseline.ps1. In CI it does not: the workflow starts
// a bare postgres:16 service container with no schema, and the bypassing
// test fails with `relation "..." does not exist` — the cause of
// test-full.yml being red on every push to main from 2026-06-29 until commit
// 1a0663f5. Five prior fixes each left a comment saying "this file is
// special" instead of a check saying "this file is wrong"; this is the
// check.
//
// The scan itself (trackedTestFiles) is repo-wide — `git ls-files
// "*_test.go"`, not scoped to any subtree — because a bypassing test can
// land under cmd/ or scripts/ just as easily as under internal/ or tests/;
// see the registry.go comment on this check's (deliberately absent) Paths
// for why that check-vs-declared-scope distinction matters for CI selection.
//
// Detection is AST-based rather than a text grep: fileBypassesTestdb only
// looks at string literal tokens (never comment text) and at genuine
// call expressions matching one of the recognised connection forms (see
// dbConnectionCalls / isDBConnectionCall), so a file that merely *mentions*
// DATABASE_URL in a comment cannot false-positive, and the reported line
// number is the real call site, not a regex match position. This is more
// precise than the equivalent grep-shaped rule in
// scripts/check-test-discipline.sh and was chosen because the two facts this
// check must not get wrong (comment vs. code, and the exact line to point a
// reader at) are both AST-native and grep-fragile.

// testdbBypassEntry is one allowlisted exception, with its reason recorded
// inline so a future reader never has to go spelunking for why a file is
// exempt.
type testdbBypassEntry struct {
	path   string
	reason string
}

// testdbBypassAllowlist is the fixed set of test files permitted to open a
// raw DATABASE_URL/METALDOCS_DATABASE_URL connection instead of going
// through testdb.Open. Today it has exactly one entry. Any addition needs
// its own reason, inline, same discipline as scripts/check-test-discipline.sh's
// R2/R3/R4 allowlists.
var testdbBypassAllowlist = []testdbBypassEntry{
	{
		path: "tests/integration/scenarios/e2e_happy_test.go",
		reason: "openRequiredDirectDB (line ~1039) deliberately verifies a " +
			"running E2E server's own database — a different physical " +
			"database than a leased testdb clone would give it, so this " +
			"is not a unit-of-schema test that testdb.Open could serve.",
	},
}

// testdbBypassExcludedDir owns the primitives this check guards: the
// factory itself reads DATABASE_URL/METALDOCS_DATABASE_URL and calls
// sql.Open to build the leased clones every other test borrows, so scanning
// it would flag the framework for doing its job — the same exclusion
// scripts/check-test-discipline.sh applies to the same directory.
const testdbBypassExcludedDir = "tests/integration/testdb/"

// isAllowlisted reports whether relPath is one of the named exceptions.
func isAllowlisted(entries []testdbBypassEntry, relPath string) bool {
	for _, e := range entries {
		if e.path == relPath {
			return true
		}
	}
	return false
}

// validateAllowlistEntries is the fail-closed half of the allowlist: an
// entry whose path no longer exists on disk (renamed, deleted) is reported
// as a finding of its own rather than silently ignored — a stale entry
// would otherwise widen the hole this check exists to close. Same
// discipline as tools/verify/audit.go's A5 rule, which fails closed when its
// own input file is missing or malformed. `exists` is injected so this stays
// testable without touching the filesystem.
func validateAllowlistEntries(entries []testdbBypassEntry, exists func(string) bool) []string {
	var findings []string
	for _, e := range entries {
		if !exists(e.path) {
			findings = append(findings, fmt.Sprintf(
				"testdb-bypass-guard: allowlist entry %q does not exist on disk — a stale allowlist entry silently widens the hole; fix or remove it",
				e.path))
		}
	}
	return findings
}

// dbConnectionCalls is the fixed set of package.Function call forms this
// check treats as "opens a real database connection". Each is a raw driver
// entry point that a test could point straight at a DATABASE_URL-derived DSN
// instead of going through testdb.Open — sql.Open is the database/sql form,
// the other three are the pgx equivalents (a direct connection and the two
// pool constructors across pgx's v4/v5 naming). Deliberately a flat literal
// table rather than a cleverer match on import paths or types: the brief for
// this guard preferred an obviously-correct rule over one that has to be
// trusted.
var dbConnectionCalls = map[string]string{
	"sql":     "Open",
	"pgx":     "Connect",
	"pgxpool": "New",
}

// isDBConnectionCall reports whether call is exactly one of the recognised
// `package.Function(...)` forms in dbConnectionCalls — sql.Open, pgx.Connect,
// or pgxpool.New — or the additional pgxpool.Connect form (pgx v4).
func isDBConnectionCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	if want, ok := dbConnectionCalls[ident.Name]; ok && sel.Sel.Name == want {
		return true
	}
	// pgxpool.Connect is pgx v4's pool constructor; v5 renamed it to
	// pgxpool.New (already covered above via dbConnectionCalls). Both are
	// live in the wild, so both are recognised.
	return ident.Name == "pgxpool" && sel.Sel.Name == "Connect"
}

// fileBypassesTestdb parses one Go source file (already-read bytes, so this
// stays pure and testable without disk I/O) and reports whether it both
// references DATABASE_URL/METALDOCS_DATABASE_URL in a string literal — never
// in a comment, since comments are not literal tokens the parser walks —
// and calls one of the recognised raw-connection forms (sql.Open,
// pgx.Connect, pgxpool.New, pgxpool.Connect; see dbConnectionCalls). line is
// the position of the first such call, valid only when bypass is true.
//
// Known limitation: both signals must appear in this same file. A helper in
// one file that reads DATABASE_URL and returns it to a consumer in another
// file of the same package, which then opens the connection, defeats this
// check — the AST walk never looks across files. Accepted for a
// deliberately simple, obviously-correct rule; not a defect to silently
// widen later without noting it here.
func fileBypassesTestdb(filename string, src []byte) (line int, bypass bool, err error) {
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, filename, src, 0)
	if err != nil {
		return 0, false, err
	}

	referencesDBURL := false
	openLine := 0
	ast.Inspect(astFile, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.BasicLit:
			if node.Kind == token.STRING && strings.Contains(node.Value, "DATABASE_URL") {
				referencesDBURL = true
			}
		case *ast.CallExpr:
			if openLine == 0 && isDBConnectionCall(node) {
				openLine = fset.Position(node.Pos()).Line
			}
		}
		return true
	})

	return openLine, referencesDBURL && openLine != 0, nil
}

// bypassMessage is the reader-facing finding: what fired, what to do about
// it, and where to copy from. A CI failure that sends someone hunting is a
// half-built check.
func bypassMessage(relPath string, line int) string {
	return fmt.Sprintf(
		"%s:%d: bypasses the testdb factory — reads DATABASE_URL/METALDOCS_DATABASE_URL and calls sql.Open, pgx.Connect, "+
			"pgxpool.New, or pgxpool.Connect directly. "+
			"Use testdb.Open(t) instead (see internal/modules/iam/infrastructure/postgres/role_provider_integration_test.go "+
			"for a worked example). Locally your own database already has a schema, so this passes; CI's bare postgres:16 "+
			"service container never gets one from this path, and the test fails with `relation \"...\" does not exist`. "+
			"This class of defect has recurred at least five times; most recently fixed in commit 1a0663f5.",
		relPath, line)
}

// scanTestdbBypassFindings reads each candidate _test.go file (relative
// paths, as produced by `git ls-files`) and reports the ones that bypass the
// factory. The testdb framework directory and allowlisted paths are
// excluded before any file is read, not filtered out of the findings after
// the fact, so a caller can trust that every reported path is a real,
// actionable violation.
func scanTestdbBypassFindings(entries []testdbBypassEntry, files []string) ([]string, error) {
	var findings []string
	for _, relPath := range files {
		if strings.HasPrefix(relPath, testdbBypassExcludedDir) || isAllowlisted(entries, relPath) {
			continue
		}
		// Same G304 idiom as tools/verify/audit.go's parseWorkflowJobs and
		// parseRequiredGateKeys — a suppression this codebase already relies
		// on for the identical "path comes from a fixed, non-user-controlled
		// enumeration" shape, not a one-off exemption invented here.
		src, err := os.ReadFile(relPath) //nolint:gosec // G304 — relPath comes from `git ls-files`, a fixed tracked-file enumeration, not external input.
		if err != nil {
			return nil, fmt.Errorf("%s: %w", relPath, err)
		}
		line, bad, err := fileBypassesTestdb(relPath, src)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", relPath, err)
		}
		if bad {
			findings = append(findings, bypassMessage(relPath, line))
		}
	}
	return findings, nil
}

// trackedTestFiles lists every git-tracked `_test.go` file, relative to the
// repo root (the process's cwd by the time this runs — main() chdirs into
// repoRoot() before dispatching any check). Scoped to tracked files only, so
// an untracked scratch file or a sibling git worktree under .claude/ can
// never be scanned or skew the count.
func trackedTestFiles() ([]string, error) {
	out, err := capture("git", "ls-files", "*_test.go")
	if err != nil {
		return nil, fmt.Errorf("git ls-files: %w", err)
	}
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

func fileExistsOnDisk(relPath string) bool {
	_, err := os.Stat(relPath)
	return err == nil
}

// printTestdbBypassGuard is the --testdb-bypass-guard entry point: the I/O
// shell around the pure functions above, mirroring printAudit's split in
// tools/verify/audit.go (parse/scan is pure and tested without a real repo;
// this function is the thin, untested-by-design wrapper that talks to the
// filesystem and stdout).
func printTestdbBypassGuard() int {
	findings := validateAllowlistEntries(testdbBypassAllowlist, fileExistsOnDisk)

	files, err := trackedTestFiles()
	if err != nil {
		fmt.Fprintf(os.Stderr, "verify --testdb-bypass-guard: %v\n", err)
		return 1
	}
	scanFindings, err := scanTestdbBypassFindings(testdbBypassAllowlist, files)
	if err != nil {
		fmt.Fprintf(os.Stderr, "verify --testdb-bypass-guard: %v\n", err)
		return 1
	}
	findings = append(findings, scanFindings...)
	sort.Strings(findings)

	fmt.Printf("verify --testdb-bypass-guard: %d test files scanned, %d findings\n", len(files), len(findings))
	if len(findings) == 0 {
		return 0
	}
	fmt.Println()
	for _, f := range findings {
		fmt.Printf("  %s\n", f)
	}
	return 1
}
