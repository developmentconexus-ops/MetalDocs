//go:build integration

package migrations_test

import (
	"context"
	"os"
	"testing"

	"metaldocs/tests/integration/testdb"
)

func TestMigration0154_FileExists(t *testing.T) {
	if _, err := os.Stat("../../../migrations/0154_capability_doc_edit_draft.sql"); err != nil {
		t.Fatalf("expected migration file 0154_capability_doc_edit_draft.sql: %v", err)
	}
}

// TestMigration0154_DocEditDraftMergedToDocumentEdit asserts the ADR 0022
// Phase 10 end state. The historical migration 0154 seeded the phantom cap
// doc.edit_draft (file still exists — see TestMigration0154_FileExists), but
// Phase 10 (migration 0227 + curated baseline) merged that redundant cap into
// the canonical document.edit: doc.edit_draft is reverse-seeded out and the
// edit-draft surface now enforces document.edit. So on a fully-migrated DB the
// phantom cap is GONE and document.edit holds the role set it used to mirror.
func TestMigration0154_DocEditDraftMergedToDocumentEdit(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)

	// The phantom cap was reversed out by Phase 10 (migration 0227).
	var phantomRows int
	if err := db.QueryRowContext(ctx, `
		SELECT count(*) FROM metaldocs.role_capabilities
		 WHERE capability = 'doc.edit_draft'`).Scan(&phantomRows); err != nil {
		t.Fatalf("count doc.edit_draft rows: %v", err)
	}
	if phantomRows != 0 {
		t.Fatalf("doc.edit_draft still seeded (%d rows); ADR 0022 Phase 10 should have merged it into document.edit", phantomRows)
	}

	// The canonical cap that subsumed it must be seeded to the same roles.
	roles := []string{"author", "qms_admin"}
	for _, role := range roles {
		var found bool
		err := db.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT 1
				  FROM metaldocs.role_capabilities
				 WHERE capability = 'document.edit'
				   AND role = $1
			)`, role).Scan(&found)
		if err != nil {
			t.Fatalf("query role_capabilities for role=%s: %v", role, err)
		}
		if !found {
			t.Fatalf("missing canonical capability document.edit for role=%s", role)
		}
	}
}
