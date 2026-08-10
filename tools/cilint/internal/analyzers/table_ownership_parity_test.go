package analyzers

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Ownership is enforced from table-ownership.json and governed in
// wiki/database/tables/<table>.md. Two places can disagree, and before #87/A1
// review B4 they did: the analyzer called governance_events approval-owned
// because approval writes it, while the governed doc (and ADR 0044) said audit.
// Nothing failed, and the A4.0 census silently lost two foreign writes.
//
// This test is the mechanism that makes that impossible. It runs under
// `go test ./...`, which the registry's go-test-unit check runs on every PR, so
// a doc/catalog disagreement is a red build rather than a judgement call. The
// catalog is not a second truth: it is the enforcement projection of the docs,
// and this test is the projection's proof.

// wikiOwnerLine matches the governed owner declaration. Both spellings occur in
// the tree: "> **Owner:** audit" (blockquote) and "**Owner:** approval module
// (ADR 0085 — ...)" (bare, with a qualifier).
var wikiOwnerLine = regexp.MustCompile(`\*\*Owner:\*\*\s*(.+)`)

// repoRoot walks up from this package to the module root.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for i := 0; i < 8; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, "wiki", "database", "tables")); err == nil {
				return dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatal("could not locate the repo root (looking for go.mod next to wiki/database/tables)")
	return ""
}

// governedOwner extracts the owner from a table doc, normalised to a bare
// label: "approval module (ADR 0085 — ...)" -> "approval".
func governedOwner(body string) (string, bool) {
	m := wikiOwnerLine.FindStringSubmatch(body)
	if m == nil {
		return "", false
	}
	owner := m[1]
	if i := strings.Index(owner, "("); i >= 0 {
		owner = owner[:i]
	}
	owner = strings.TrimSpace(owner)
	owner = strings.TrimSuffix(owner, " module")
	return strings.TrimSpace(owner), owner != ""
}

func TestTableOwnershipMatchesGovernedDocs(t *testing.T) {
	root := repoRoot(t)
	modules := moduleDirs(t, root)

	for _, e := range ownershipEntries {
		path := filepath.Join(root, "wiki", "database", "tables", e.Table+".md")
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("%s: no governed doc at wiki/database/tables/%s.md — ownership must be governed before it is enforced (%v)",
				e.Table, e.Table, err)
			continue
		}
		wikiOwner, ok := governedOwner(string(raw))
		if !ok {
			t.Errorf("%s: wiki/database/tables/%s.md declares no **Owner:** line; the catalog says %q with nothing governing it",
				e.Table, e.Table, e.Owner)
			continue
		}

		normalised := wikiOwner
		if alias, isAlias := ownershipAliases[normalised]; isAlias {
			normalised = alias
		}

		// A table classified out of module enforcement still has a governed
		// label, and it still has to match: the class says "no module owns
		// this", not "the docs stop applying". Recording the label verbatim is
		// what makes a doc change that re-homes the table a red build.
		if e.Class != classModule {
			if e.WikiOwner != wikiOwner {
				t.Errorf("%s (class %s): wiki_owner=%q but the governed doc says %q",
					e.Table, e.Class, e.WikiOwner, wikiOwner)
			}
			if modules[normalised] {
				t.Errorf("%s: the governed doc names the module %q, so it is module-owned, not %q",
					e.Table, normalised, e.Class)
			}
			continue
		}

		if normalised == e.Owner {
			if e.WikiOwner != "" {
				t.Errorf("%s: wiki_owner=%q is set but the governed doc already agrees with owner=%q — drop the acknowledgement",
					e.Table, e.WikiOwner, e.Owner)
			}
			continue
		}

		// Divergence is allowed only where the governed label is not a module
		// directory (a platform-layer owner, or a product surface) AND the entry
		// acknowledges the exact governed label. Anything else is drift.
		if e.WikiOwner == "" {
			t.Errorf("%s: governed doc says owner %q, catalog says %q — reconcile the doc or record wiki_owner with the reason",
				e.Table, wikiOwner, e.Owner)
			continue
		}
		if e.WikiOwner != wikiOwner {
			t.Errorf("%s: wiki_owner=%q no longer matches the governed doc, which now says %q",
				e.Table, e.WikiOwner, wikiOwner)
			continue
		}
		// The old form of this rule asked whether the label contained a "/",
		// which is a guess at "is this a module name". The tree answers it
		// exactly: if the governed label IS one of the module directories, then
		// the catalog naming a different owner is drift, full stop.
		if modules[normalised] {
			t.Errorf("%s: the governed doc names the module %q and the catalog says %q — that is drift, not a platform-layer divergence",
				e.Table, normalised, e.Owner)
		}
		if e.Note == "" {
			t.Errorf("%s: diverges from the governed doc (%q vs %q) with no note explaining why", e.Table, wikiOwner, e.Owner)
		}
	}
}

// The exemption list is the other half of B5: modes must be a closed set, and a
// write exemption must carry its own rationale. A typo'd mode would silently
// authorise nothing (safe) or, worse, be read as a wildcard by a future edit.
func TestExemptionModesAreValidAndJustified(t *testing.T) {
	for _, lists := range [][]hgSite{hgExempt, hgPendingRemediation} {
		for _, site := range lists {
			switch site.mode {
			case hgModeReads, hgModeWrites:
			default:
				t.Errorf("%s x %s: mode %q is not one of %q/%q",
					site.fileSuffix, site.table, site.mode, hgModeReads, hgModeWrites)
			}
			if strings.TrimSpace(site.why) == "" {
				t.Errorf("%s x %s: exemption carries no rationale", site.fileSuffix, site.table)
			}
		}
	}
}
