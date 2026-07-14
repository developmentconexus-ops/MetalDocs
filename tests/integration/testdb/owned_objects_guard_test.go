//go:build integration
// +build integration

package testdb

import (
	"database/sql"
	"errors"
	"testing"
)

// TestMetaldocsOwnedObjectsMatchCatalog guards metaldocsOwnedObjects against
// drift from the real schema. That map is a hand-synced enumeration — the
// known meta-defect class in this codebase — and until now nothing checked it
// against the catalog: runtimeSchemaName just trusts the map, so a stale
// entry silently sends Qualified() callers at the wrong schema (or a table
// that isn't there at all) with no failure anywhere near the mistake. Two
// entries were exactly that (governance_events actually lives in public;
// grant_area_membership doesn't exist as a table at all), and at least one of
// them plausibly qualified a query against metaldocs.governance_events, a
// table that was never there, without ever failing because the test in
// question was being skipped. This test turns that silent drift loud. It does
// not fix the underlying meta-defect — deriving the schema from the catalog
// instead of a hand-maintained map would, but Qualified takes no DB handle,
// so that refactor is out of scope here and reported separately.
//
// Every current entry in metaldocsOwnedObjects is an ordinary table (verified
// against db/baseline/0001_current_schema.sql), so information_schema.tables
// is sufficient; it would NOT catch a view/sequence/type claimed by a future
// entry, but no such entry exists today.
//
// information_schema is privilege-filtered, so a role that cannot see a table
// reports it as absent. That direction is safe here: it can only produce a
// false FAILURE (loud, investigable), never a false PASS — this test cannot
// certify a map entry that isn't really there.
func TestMetaldocsOwnedObjectsMatchCatalog(t *testing.T) {
	db, _ := Open(t)

	var failures []string
	for object := range metaldocsOwnedObjects {
		claimedSchema := runtimeSchemaName(object)

		var actualSchema string
		err := db.QueryRow(
			`SELECT table_schema FROM information_schema.tables WHERE table_schema = $1 AND table_name = $2`,
			claimedSchema, object,
		).Scan(&actualSchema)
		switch {
		case err == nil:
			continue // present in the claimed schema, as expected
		case !errors.Is(err, sql.ErrNoRows):
			// A real query failure (connection dropped, etc.) is NOT drift, and
			// reporting it as "not found in any schema" would send the reader
			// hunting a schema bug that doesn't exist. Fail on the spot instead.
			t.Fatalf("query catalog for %s.%s: %v", claimedSchema, object, err)
		}

		var elsewhere string
		lookupErr := db.QueryRow(
			`SELECT table_schema FROM information_schema.tables WHERE table_name = $1 LIMIT 1`,
			object,
		).Scan(&elsewhere)
		switch {
		case lookupErr == nil:
			failures = append(failures, "object "+object+": map claims schema "+claimedSchema+", but found in schema "+elsewhere+" instead")
		case errors.Is(lookupErr, sql.ErrNoRows):
			failures = append(failures, "object "+object+": map claims schema "+claimedSchema+", but not found in any schema")
		default:
			t.Fatalf("search catalog for %s in any schema: %v", object, lookupErr)
		}
	}

	for _, f := range failures {
		t.Errorf("%s", f)
	}
}
