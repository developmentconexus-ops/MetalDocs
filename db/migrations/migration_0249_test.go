package migrations_test

import (
	"os"
	"strings"
	"testing"
)

// TestMigration0249WidensDocumentPlaceholderValuesSourceCheck locks the SP-2
// invariant: the migration must widen the source CHECK constraint on
// document_placeholder_values to include 'dictionary' while preserving the
// three existing values ('user', 'computed', 'default'). The constraint is
// dropped then re-added (idempotent pattern) and must not replace the closed
// set — only extend it.
func TestMigration0249WidensDocumentPlaceholderValuesSourceCheck(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile("0249_document_placeholder_values_dictionary_source.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	sql := string(body)

	// Must be wrapped in an explicit transaction (idempotent style).
	if !strings.Contains(sql, "BEGIN;") || !strings.Contains(sql, "COMMIT;") {
		t.Fatal("migration must be wrapped in BEGIN/COMMIT")
	}

	// Idempotency: drop before add.
	if !strings.Contains(sql, "DROP CONSTRAINT IF EXISTS document_placeholder_values_source_check") {
		t.Fatal("migration must DROP CONSTRAINT IF EXISTS before re-adding (idempotent)")
	}
	if !strings.Contains(sql, "ADD CONSTRAINT document_placeholder_values_source_check") {
		t.Fatal("migration must ADD CONSTRAINT document_placeholder_values_source_check")
	}

	// Locate the ADD CONSTRAINT clause to scope the value-set assertions.
	addIdx := strings.Index(sql, "ADD CONSTRAINT document_placeholder_values_source_check")
	if addIdx < 0 {
		t.Fatal("missing ADD CONSTRAINT clause")
	}
	end := addIdx + 300
	if end > len(sql) {
		end = len(sql)
	}
	checkClause := sql[addIdx:end]

	// New value must be present.
	if !strings.Contains(checkClause, "'dictionary'") {
		t.Fatal("the widened CHECK must include 'dictionary'")
	}

	// Existing values must be preserved (additive, not replacement).
	for _, val := range []string{"'user'", "'computed'", "'default'"} {
		if !strings.Contains(checkClause, val) {
			t.Fatalf("the widened CHECK must still include %s (additive, not replacement)", val)
		}
	}

	// The constraint targets the correct table.
	if !strings.Contains(sql, "ALTER TABLE public.document_placeholder_values") {
		t.Fatal("migration must target public.document_placeholder_values")
	}
}
