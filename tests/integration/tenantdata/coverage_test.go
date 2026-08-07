//go:build integration

// coverage_test.go — M7 F7.3 Task D TDD: the anti-rot guard for
// TenantDataPort registration (spec §2.2, validation-contract.md §0.1/§2.2).
//
// TestTenantDataPortCoverage queries the live schema (via the curated
// bootstrap, tests/integration/testdb.Open) for every BASE TABLE carrying a
// tenant_id or actor_tenant_id column, across both the metaldocs and public
// schemas, and asserts that set equals EXACTLY the union of every registered
// port's Tables() plus the explicit allowlist (tenants root row, tenant_keys
// crypto-shred row — both handled outside any TenantDataPort by design, see
// registry.go's doc comment). Any new tenant-scoped table that ships without
// a registered port fails this test by name — the whole point of the guard
// (F7.4/future-module consumer contract, spec item 4).
//
//	go test -tags=integration ./tests/integration/tenantdata/... -run TestTenantDataPortCoverage
package tenantdata_test

import (
	"context"
	"database/sql"
	"sort"
	"testing"

	"metaldocs/internal/composition/tenantdata/registry"
	"metaldocs/internal/platform/tenantdata"
	"metaldocs/tests/integration/testdb"
)

// allowlist enumerates tenant-scoped assets that are deliberately NOT
// covered by any TenantDataPort (spec item 2 special cases / registry.go
// doc comment):
//   - metaldocs.tenants: the root row, tombstoned (not deleted) by the
//     erase orchestrator directly.
//   - metaldocs.tenant_keys: the per-tenant crypto key row, destroyed by
//     the security module's crypto-shred step, not a TenantDataPort.
var allowlist = map[string]struct{}{
	"metaldocs.tenants":     {},
	"metaldocs.tenant_keys": {},
}

// censusQuery finds every BASE TABLE (never a VIEW) in the metaldocs or
// public schema carrying a tenant_id or actor_tenant_id column — the same
// definition the F7.3 contract §0.1 census uses. VIEW is excluded because
// metaldocs.user_process_areas is a read-only view over the real base table
// public.user_process_areas; counting both would double-count one asset.
const censusQuery = `
SELECT c.table_schema || '.' || c.table_name
  FROM information_schema.columns c
  JOIN information_schema.tables t
    ON t.table_schema = c.table_schema AND t.table_name = c.table_name
 WHERE c.table_schema IN ('metaldocs', 'public')
   AND c.column_name IN ('tenant_id', 'actor_tenant_id')
   AND t.table_type = 'BASE TABLE'
 ORDER BY 1`

func schemaCensus(t *testing.T, db *sql.DB) []string {
	t.Helper()
	rows, err := db.QueryContext(context.Background(), censusQuery)
	if err != nil {
		t.Fatalf("schema census query: %v", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan census row: %v", err)
		}
		tables = append(tables, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate census rows: %v", err)
	}
	return tables
}

func registeredTables(db *sql.DB) map[string]string {
	var ports []tenantdata.Port = registry.AllTenantDataPorts(db)
	owners := make(map[string]string)
	for _, port := range ports {
		for _, table := range port.Tables() {
			owners[table] = port.Module()
		}
	}
	return owners
}

// TestTenantDataPortCoverage is the positive proof: the live census, plus
// the explicit allowlist, equals exactly the union of every registered
// port's Tables() — both directions.
func TestTenantDataPortCoverage(t *testing.T) {
	db, _ := testdb.Open(t)
	defer db.Close()

	census := schemaCensus(t, db)
	owners := registeredTables(db)

	var missingPort []string
	for _, table := range census {
		if _, allowed := allowlist[table]; allowed {
			continue
		}
		if _, ok := owners[table]; !ok {
			missingPort = append(missingPort, table)
		}
	}
	if len(missingPort) > 0 {
		sort.Strings(missingPort)
		t.Fatalf("tenant-scoped table(s) with NO registered TenantDataPort: %v — "+
			"every module owning a tenant_id/actor_tenant_id table must register a port "+
			"in internal/composition/tenantdata/registry.go (F7.3 spec §2.2)", missingPort)
	}

	censusSet := make(map[string]struct{}, len(census))
	for _, table := range census {
		censusSet[table] = struct{}{}
	}

	var phantomTables []string
	for table := range owners {
		if _, allowed := allowlist[table]; allowed {
			// A port must never claim an allowlisted table.
			phantomTables = append(phantomTables, table+" (also allowlisted — remove from one)")
			continue
		}
		if _, inCensus := censusSet[table]; !inCensus {
			phantomTables = append(phantomTables, table)
		}
	}
	if len(phantomTables) > 0 {
		sort.Strings(phantomTables)
		t.Fatalf("registered port(s) claim table(s) that do NOT exist (or lack a tenant column) in the live schema: %v", phantomTables)
	}
}

// TestTenantDataPortCoverage_SyntheticGapGoesRed is the negative proof
// (spec item 4 / validation gate row "Coverage test: census fully ported;
// synthetic gap goes RED"): a synthetic census entry with no registered
// port must fail the same assertion TestTenantDataPortCoverage runs,
// proving the guard actually detects a real gap rather than vacuously
// passing.
func TestTenantDataPortCoverage_SyntheticGapGoesRed(t *testing.T) {
	db, _ := testdb.Open(t)
	defer db.Close()

	census := schemaCensus(t, db)
	// Inject a table that certainly has no registered port.
	census = append(census, "public.synthetic_unported_tenant_table")

	owners := registeredTables(db)

	var missingPort []string
	for _, table := range census {
		if _, allowed := allowlist[table]; allowed {
			continue
		}
		if _, ok := owners[table]; !ok {
			missingPort = append(missingPort, table)
		}
	}

	if len(missingPort) == 0 {
		t.Fatalf("expected the synthetic unported table to be flagged, got no missing tables — the coverage guard is not detecting gaps")
	}
	found := false
	for _, m := range missingPort {
		if m == "public.synthetic_unported_tenant_table" {
			found = true
		}
	}
	if !found {
		t.Fatalf("missingPort = %v, want it to include the synthetic gap table", missingPort)
	}
}

// TestAllTenantDataPorts_TablesNonEmpty is the unit-level sanity check (spec
// item 5): every registered port must own at least one table.
func TestAllTenantDataPorts_TablesNonEmpty(t *testing.T) {
	db, _ := testdb.Open(t)
	defer db.Close()

	for _, port := range registry.AllTenantDataPorts(db) {
		if len(port.Tables()) == 0 {
			t.Errorf("port %q: Tables() is empty, want at least one tenant-scoped table", port.Module())
		}
	}
}

// TestAllTenantDataPorts_ModuleNamesUnique is the unit-level sanity check
// (spec item 5): no two registered ports may declare the same Module() name.
func TestAllTenantDataPorts_ModuleNamesUnique(t *testing.T) {
	db, _ := testdb.Open(t)
	defer db.Close()

	seen := make(map[string]bool)
	for _, port := range registry.AllTenantDataPorts(db) {
		name := port.Module()
		if seen[name] {
			t.Errorf("duplicate TenantDataPort Module() name: %q", name)
		}
		seen[name] = true
	}
}
