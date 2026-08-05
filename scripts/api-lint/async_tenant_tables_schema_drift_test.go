package main

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"testing"
)

// forceRLSTableRE matches `ALTER TABLE ONLY <schema>.<table> FORCE ROW LEVEL
// SECURITY;` lines in the pg_dump-style baseline schema, capturing the bare
// table name (schema prefix stripped, mirroring async-tenant-tables.txt's own
// convention).
var forceRLSTableRE = regexp.MustCompile(`(?m)^ALTER TABLE ONLY [a-z_]+\.([a-z_]+) FORCE ROW LEVEL SECURITY;`)

// loadSchemaForceRLSTables parses db/baseline/0001_current_schema.sql for
// every table carrying FORCE ROW LEVEL SECURITY — the runtime truth for which
// tables are tenant-scoped and RLS-enforced.
func loadSchemaForceRLSTables(t *testing.T, root string) map[string]struct{} {
	t.Helper()
	path := filepath.Join(root, "db", "baseline", "0001_current_schema.sql")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read baseline schema %s: %v", path, err)
	}
	out := map[string]struct{}{}
	for _, m := range forceRLSTableRE.FindAllSubmatch(raw, -1) {
		out[string(m[1])] = struct{}{}
	}
	if len(out) == 0 {
		t.Fatalf("no FORCE ROW LEVEL SECURITY tables parsed from %s — regex or file drifted", path)
	}
	return out
}

// TestAsyncTenantTables_MatchesSchemaForceRLS guards against the
// hand-synced-enumeration meta-defect (docs/engineering/defect-class-catalog.md):
// scripts/api-lint/async-tenant-tables.txt is a hand-maintained mirror of the
// baseline schema's FORCE ROW LEVEL SECURITY table set, and the two are free
// to drift silently whenever a migration adds/renames/drops a tenant-scoped
// table without a matching edit to the data file. Both ASYNC-TENANT-SEED
// (async_tenant_seed_rule.go) and SOLE-RLS-ASYNC-READ
// (sole_rls_read_rule.go) load this data file as their tenant-table scope, so
// a stale entry silently disarms the backstop for that table's async
// write/read sites. This test fails the moment the sets diverge in either
// direction (missing from the file, or listed with no FORCE RLS in schema).
func TestAsyncTenantTables_MatchesSchemaForceRLS(t *testing.T) {
	root := repoRoot(t)

	schemaTables := loadSchemaForceRLSTables(t, root)
	fileTables, err := loadAsyncTenantTables(root)
	if err != nil {
		t.Fatalf("loadAsyncTenantTables: %v", err)
	}

	var missingFromFile []string
	for table := range schemaTables {
		if _, ok := fileTables[table]; !ok {
			missingFromFile = append(missingFromFile, table)
		}
	}
	var staleInFile []string
	for table := range fileTables {
		if _, ok := schemaTables[table]; !ok {
			staleInFile = append(staleInFile, table)
		}
	}
	sort.Strings(missingFromFile)
	sort.Strings(staleInFile)

	if len(missingFromFile) > 0 || len(staleInFile) > 0 {
		t.Fatalf(
			"scripts/api-lint/async-tenant-tables.txt has drifted from db/baseline/0001_current_schema.sql's FORCE ROW LEVEL SECURITY set — the ASYNC-TENANT-SEED and SOLE-RLS-ASYNC-READ backstops are silently inert for the affected tables.\n"+
				"FORCE RLS in schema but missing from the file (add these): %v\n"+
				"listed in the file but no FORCE RLS in schema (remove these): %v",
			missingFromFile, staleInFile,
		)
	}
}
