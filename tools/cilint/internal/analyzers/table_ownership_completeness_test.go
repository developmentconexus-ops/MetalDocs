package analyzers

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// REVERSE COMPLETENESS (#87/A1 review F1).
//
// The parity test walks catalog → docs. That direction proves every entry is
// governed; it cannot prove every governed table has an entry, and it did not:
// 19 of the 56 live base tables were absent from table-ownership.json, and
// hgViolation returns "not a violation" for a table it does not know. A foreign
// `UPDATE document_images SET ...` was therefore invisible — the guard's central
// A4.0 claim, that a new foreign write is mechanically rejected, was false for a
// third of the schema.
//
// The fix is not more entries; entries rot. The fix is that the catalog must BE
// the universe. These tests derive the universe from the governed schema — the
// same db/baseline the db-docs-coverage check already treats as table truth,
// plus the forward migrations on top of it — and fail on any divergence in
// either direction. No new source of truth is introduced: the schema decides
// which tables exist, wiki/database/tables decides who owns them, and this file
// is the projection both of them are held against.

var (
	createTableRe = regexp.MustCompile(`(?i)CREATE\s+TABLE(?:\s+IF\s+NOT\s+EXISTS)?\s+(?:[a-z0-9_]+\.)?([a-z0-9_]+)`)
	dropTableRe   = regexp.MustCompile(`(?i)DROP\s+TABLE(?:\s+IF\s+EXISTS)?\s+([^;]+);`)
	tableTokenRe  = regexp.MustCompile(`(?:[a-z0-9_]+\.)?([a-z0-9_]+)`)
)

// governedSchemaFiles returns the SQL that defines the LIVE schema, in apply
// order: the folded baseline first, then the forward migrations. archive/ is
// deliberately not here — those migrations describe tables that no longer exist,
// which is exactly what the `retired` list records instead.
func governedSchemaFiles(t *testing.T, root string) []string {
	t.Helper()
	var out []string
	for _, dir := range []string{
		filepath.Join(root, "db", "baseline"),
		filepath.Join(root, "db", "migrations"),
	} {
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("read %s: %v", dir, err)
		}
		var names []string
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
				names = append(names, e.Name())
			}
		}
		sort.Strings(names)
		for _, n := range names {
			out = append(out, filepath.Join(dir, n))
		}
	}
	if len(out) == 0 {
		t.Fatal("no schema SQL found under db/baseline or db/migrations — the universe cannot be derived")
	}
	return out
}

// liveBaseTables applies CREATE/DROP TABLE in file order and returns what is
// left standing: the authoritative base-table universe.
func liveBaseTables(t *testing.T, root string) map[string]bool {
	t.Helper()
	live := map[string]bool{}
	for _, path := range governedSchemaFiles(t, root) {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		sql := string(raw)
		for _, m := range createTableRe.FindAllStringSubmatch(sql, -1) {
			live[strings.ToLower(m[1])] = true
		}
		// A migration that drops a table takes it out of the universe, so the
		// catalog is not asked to classify something that no longer exists.
		for _, m := range dropTableRe.FindAllStringSubmatch(sql, -1) {
			for _, tok := range strings.Split(m[1], ",") {
				tok = strings.TrimSpace(tok)
				tok = strings.TrimSuffix(tok, " CASCADE")
				tok = strings.TrimSuffix(tok, " cascade")
				if sub := tableTokenRe.FindStringSubmatch(strings.TrimSpace(tok)); sub != nil {
					delete(live, strings.ToLower(sub[1]))
				}
			}
		}
	}
	return live
}

func catalogTables() map[string]ownershipEntry {
	out := make(map[string]ownershipEntry, len(ownershipEntries))
	for _, e := range ownershipEntries {
		out[e.Table] = e
	}
	return out
}

// A new live base table that nobody classified is the exact hole F1 named: the
// analyzer would treat every access to it as compliant.
func TestEveryLiveBaseTableIsClassified(t *testing.T) {
	root := repoRoot(t)
	live := liveBaseTables(t, root)
	cat := catalogTables()

	var missing []string
	for table := range live {
		if _, ok := cat[table]; !ok {
			missing = append(missing, table)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("%d live base table(s) are absent from table-ownership.json: %s\n"+
			"An absent table is an UNGUARDED table — hgViolation reports nothing for a table it does not know, "+
			"so a foreign write against any of these would pass silently. Classify each one "+
			"(module/platform/historical/out-of-scope); omission is not a classification.",
			len(missing), strings.Join(missing, ", "))
	}
}

// The other direction: a table that left the schema must leave the catalog, or
// the catalog starts describing a database that does not exist.
func TestEveryCatalogedTableIsLive(t *testing.T) {
	root := repoRoot(t)
	live := liveBaseTables(t, root)

	var phantom []string
	for _, e := range ownershipEntries {
		if !live[e.Table] {
			phantom = append(phantom, e.Table)
		}
	}
	sort.Strings(phantom)
	if len(phantom) > 0 {
		t.Errorf("table-ownership.json classifies %d table(s) the live schema does not have: %s\n"+
			"Either db/baseline + db/migrations dropped them (move the entry to `retired`) or the entry is a typo.",
			len(phantom), strings.Join(phantom, ", "))
	}
}

// Duplicates are caught by the loader's panic, which is the fail-closed path.
// This test states the property in the place someone reads for it, and it also
// covers the live/retired overlap.
func TestCatalogHasNoDuplicateTables(t *testing.T) {
	seen := map[string]string{}
	for _, e := range ownershipEntries {
		if prev, dup := seen[e.Table]; dup {
			t.Errorf("%s is classified twice (%s, then %s)", e.Table, prev, e.Class)
		}
		seen[e.Table] = e.Class
	}
	for _, r := range ownershipRetired {
		if prev, dup := seen[r.Table]; dup {
			t.Errorf("%s is listed as retired and as %s — it is one or the other", r.Table, prev)
		}
		seen[r.Table] = "retired"
	}
}

// Ownership can only be enforced against a module that exists. A typo'd owner
// would make the table effectively unowned for every reader (no module's name
// matches it), which reads as "compliant" everywhere.
func TestModuleOwnersAreRealModules(t *testing.T) {
	root := repoRoot(t)
	modules := moduleDirs(t, root)

	for _, e := range ownershipEntries {
		if e.Class != classModule {
			continue
		}
		if !modules[e.Owner] {
			t.Errorf("%s is owned by %q, which is not a directory under internal/modules/", e.Table, e.Owner)
		}
	}
}

// The documented set and the live set must reconcile with no leftovers on either
// side. This is what gives document_attachments a checked answer instead of a
// remembered one: it is documented, it is not in the schema, so it must be
// listed as retired — and if the baseline ever brings it back, this test demands
// a classification for it.
func TestGovernedDocsReconcileWithLiveAndRetired(t *testing.T) {
	root := repoRoot(t)
	dir := filepath.Join(root, "wiki", "database", "tables")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}

	documented := map[string]bool{}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			documented[strings.TrimSuffix(e.Name(), ".md")] = true
		}
	}

	accounted := map[string]bool{}
	for _, e := range ownershipEntries {
		accounted[e.Table] = true
	}
	for _, r := range ownershipRetired {
		accounted[r.Table] = true
	}

	var undocumented, unaccounted []string
	for table := range accounted {
		if !documented[table] {
			undocumented = append(undocumented, table)
		}
	}
	for table := range documented {
		if !accounted[table] {
			unaccounted = append(unaccounted, table)
		}
	}
	sort.Strings(undocumented)
	sort.Strings(unaccounted)

	if len(undocumented) > 0 {
		t.Errorf("catalogued but not documented under wiki/database/tables: %s", strings.Join(undocumented, ", "))
	}
	if len(unaccounted) > 0 {
		t.Errorf("documented under wiki/database/tables but neither classified nor retired: %s\n"+
			"A doc with no entry is how an omission hides: add the table to `tables` if the schema has it, "+
			"or to `retired` with the migration that removed it.", strings.Join(unaccounted, ", "))
	}
}

// FAIL CLOSED. Each case below is a catalog that would still load, still let
// cilint run, and still print "0 findings" — while classifying less than it
// looks like it does. The whole F1 defect was one of these shapes at the file
// level (a table simply absent), so the loader refusing them is the property,
// and this is where it is proved rather than asserted.
func TestOwnershipLoadFailsClosed(t *testing.T) {
	cases := []struct {
		name string
		cat  ownershipCatalog
		want string
	}{
		{
			name: "module-owned with no owner",
			cat:  ownershipCatalog{Tables: []ownershipEntry{{Table: "documents", Class: classModule}}},
			want: "module-owned with no owner",
		},
		{
			name: "no class at all",
			cat:  ownershipCatalog{Tables: []ownershipEntry{{Table: "documents", Owner: "documents"}}},
			want: "has no class",
		},
		{
			name: "class nobody defined",
			cat:  ownershipCatalog{Tables: []ownershipEntry{{Table: "documents", Class: "shared-ish", Note: "n"}}},
			want: "unknown class",
		},
		{
			name: "unenforced without a reason",
			cat:  ownershipCatalog{Tables: []ownershipEntry{{Table: "outbox_events", Class: classPlatform}}},
			want: "with no reason",
		},
		{
			name: "unenforced but carrying an owner",
			cat:  ownershipCatalog{Tables: []ownershipEntry{{Table: "outbox_events", Class: classPlatform, Owner: "jobs", Note: "n"}}},
			want: "pick one",
		},
		{
			name: "the same table twice",
			cat: ownershipCatalog{Tables: []ownershipEntry{
				{Table: "documents", Class: classModule, Owner: "documents"},
				{Table: "documents", Class: classModule, Owner: "approval"},
			}},
			want: "twice",
		},
		{
			name: "live and retired at once",
			cat: ownershipCatalog{
				Tables:  []ownershipEntry{{Table: "documents", Class: classModule, Owner: "documents"}},
				Retired: []retiredEntry{{Table: "documents", Note: "n"}},
			},
			want: "both live and retired",
		},
		{
			name: "retired with no reason",
			cat:  ownershipCatalog{Retired: []retiredEntry{{Table: "job_leases"}}},
			want: "needs a table and a reason",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ownerByTable(tc.cat)
			if err == nil {
				t.Fatalf("catalog was accepted; it must be refused (%s)", tc.want)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("refused for the wrong reason: got %q, want it to mention %q", err, tc.want)
			}
		})
	}
}

// The live catalog must survive its own gate.
func TestLiveCatalogLoads(t *testing.T) {
	owners, err := ownerByTable(mustOwnershipCatalog())
	if err != nil {
		t.Fatalf("the shipped catalog does not load: %v", err)
	}
	if len(owners) == 0 {
		t.Fatal("the shipped catalog produced no owners")
	}
}

// moduleDirs is the set of top-level bounded-context module directories. It is
// read from the tree rather than listed here, so adding module 16 does not need
// an edit in this file.
func moduleDirs(t *testing.T, root string) map[string]bool {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(root, "internal", "modules"))
	if err != nil {
		t.Fatalf("read internal/modules: %v", err)
	}
	out := map[string]bool{}
	for _, e := range entries {
		if e.IsDir() {
			out[e.Name()] = true
		}
	}
	return out
}
