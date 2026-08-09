// Package analyzers runs all cilint analyzers and aggregates findings.
package analyzers

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Finding is a lint violation.
type Finding struct {
	Analyzer string
	File     string
	Line     int
	Message  string
}

// allowedTxPackages are the layers that legitimately own database transactions
// in this codebase's clean/hexagonal architecture: the data adapters
// (infrastructure/repository), the application/service layer that orchestrates
// cross-aggregate atomic writes (ADR 0011), shared platform transaction infra
// (the db.TxRunner port, idempotency store, transactional outbox — ADR 0009),
// worker job runners (ADR 0015), and the test seed harness. Transactions must
// NOT be opened in the presentation layer (delivery/http, api/), so those stay
// flagged — that is the real smell this rule guards against.
var allowedTxPackages = []string{
	"/infrastructure/",
	"/repository/",
	"/application/",
	"/platform/db/", // TxRunner port — canonical centralized tx lifecycle owner (H-1d)
	"/platform/idempotency/",
	"/platform/messaging/",
	"/platform/worker/",
	"jobs/",
	"internal/test/",            // integration seed harness (build-tagged, non-prod)
	"tests/integration/testdb/", // M4c unified testdb fixture framework — seed helpers (seedWithCaps) own a tx to set_config asserted-caps + governed write on one conn; same seed-harness category as internal/test/
}

// RunAll runs every analyzer over the given patterns and aggregates findings.
func RunAll(targets []string) []Finding {
	files := collectGoFiles(targets)
	var out []Finding
	out = append(out, TxOwnership(files)...)
	out = append(out, LegacyVocab(files)...)
	out = append(out, OutboxPair(files)...)
	out = append(out, PlatformBoundary(files)...)
	out = append(out, PostCommitAudit(files)...)
	out = append(out, DeliveryAuditSink(files)...)
	out = append(out, NoSQLTxInDomain(files)...)
	out = append(out, NoDualMode(files)...)
	out = append(out, NoResponseMap(files)...)
	out = append(out, HGCrossModule(files)...)
	return out
}

func collectGoFiles(patterns []string) []string {
	var files []string
	for _, pat := range patterns {
		pat = strings.TrimSuffix(pat, "/...")
		if pat == "./..." || pat == "." {
			pat = "."
		}
		_ = filepath.WalkDir(pat, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil //nolint:nilerr // best-effort walk: skip unreadable entries
			}
			if d.IsDir() {
				// Skip vendored, dependency, and tooling-scratch trees. Without
				// this the walker descends into .claude/worktrees (agent working
				// copies), scanning duplicate Go files and double-reporting.
				name := d.Name()
				if name == "vendor" || name == "node_modules" ||
					(name != "." && strings.HasPrefix(name, ".")) {
					return fs.SkipDir
				}
				return nil
			}
			if strings.HasSuffix(path, ".go") &&
				!strings.Contains(path, "_test.go") &&
				!strings.Contains(path, "vendor/") {
				files = append(files, path)
			}
			return nil
		})
	}
	return files
}

func parseFile(fset *token.FileSet, path string) (*token.File, any) {
	f, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return nil, nil
	}
	tf := fset.File(f.Pos())
	return tf, f
}

func readSource(path string) string {
	b, _ := os.ReadFile(path) // #nosec G304 -- path comes from collectGoFiles' filepath.WalkDir over the repo tree, not external input.
	return string(b)
}

func inAllowedPackage(path string) bool {
	for _, pkg := range allowedTxPackages {
		if strings.Contains(filepath.ToSlash(path), pkg) {
			return true
		}
	}
	return false
}
