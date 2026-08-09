package analyzers

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// legacyPattern matches deprecated vocabulary in Go and TypeScript sources.
//
// NOTE: "archived" is intentionally NOT banned. Per ADR 0010 (soft-archive via
// timestamp), `archived`/`archived_at` is a CURRENT, load-bearing concept across
// documents, taxonomy areas/profiles, templates, controlled documents, and
// search. Only the document-approval Spec-2 graph treats it as legacy, and that
// code rejects it explicitly. "finalized" remains banned as a document status
// (use published). There is no per-line suppression directive: legitimate
// legacy-rejection/cutover code must reference the retired literal through the
// canonical constant homes (internal/platform/legacystatus in Go,
// frontend/apps/web/src/lib/legacyStatus.ts in TypeScript), which are
// structural, reviewable path exclusions in legacyExcludeDirs below — never an
// inline, self-service escape hatch.
var legacyPattern = regexp.MustCompile(`(?i)\b(finalized|document\.finalize|document\.archive)\b`)

// legacyExcludeDirs are excluded from legacy vocab checks. Every entry here is
// a structural, reviewable path — the only form of exclusion this analyzer
// supports. There is deliberately no inline per-line directive: an exclusion
// must show up in this diff, not scattered invisibly through the tree.
var legacyExcludeDirs = []string{
	"migrations/",
	"fixtures/",
	"testdata/",
	"/api-types/",            // generated OpenAPI client types mirror the backend contract
	"tools/cilint/",          // the linter's own source defines the banned-word pattern
	"cutover_service.go",     // legacy-status cutover preflight (drains finalized/archived; ADR/migration 0142)
	"platform/legacystatus/", // canonical Go home for the retired "finalized" literal (internal/platform/legacystatus)
	"lib/legacyStatus.ts",    // canonical TS home for the retired "finalized" literal (frontend/apps/web/src/lib/legacyStatus.ts)
}

// LegacyVocab reports legacy vocabulary in .go, .ts, and .tsx files
// excluding test fixtures and historical event type enums.
func LegacyVocab(goFiles []string) []Finding {
	// Expand to include .ts/.tsx in addition to .go
	var allFiles []string
	allFiles = append(allFiles, goFiles...)

	// Walk frontend sources too
	_ = filepath.WalkDir("frontend/apps/web/src", func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			// Best-effort walk: an unreadable entry (or an absent frontend
			// tree) skips that entry instead of aborting the whole sweep.
			return nil //nolint:nilerr // propagating err would kill the walk on the first unreadable path
		}
		if strings.HasSuffix(path, ".ts") || strings.HasSuffix(path, ".tsx") {
			allFiles = append(allFiles, path)
		}
		return nil
	})

	var out []Finding

	for _, path := range allFiles {
		if isLegacyExcluded(path) {
			continue
		}

		f, err := os.Open(path) // #nosec G304 -- path is drawn from allFiles, itself sourced from the caller's goFiles list (repo-tree walk) plus a filepath.WalkDir over frontend/apps/web/src; not external input.
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()

			// Skip historical event type constant definitions (e.g., EventTypeArchived = "archived")
			if strings.Contains(line, "EventType") && strings.Contains(line, "=") {
				continue
			}

			if legacyPattern.MatchString(line) {
				matches := legacyPattern.FindAllString(line, -1)
				out = append(out, Finding{
					Analyzer: "legacyvocab",
					File:     path,
					Line:     lineNum,
					Message:  "legacy vocabulary found: " + strings.Join(matches, ", ") + " — use current terminology (published/cancelled/obsolete)",
				})
			}
		}
		_ = f.Close()
	}
	return out
}

func isLegacyExcluded(path string) bool {
	slash := strings.ReplaceAll(path, "\\", "/")
	for _, excl := range legacyExcludeDirs {
		if strings.Contains(slash, excl) {
			return true
		}
	}
	return false
}
