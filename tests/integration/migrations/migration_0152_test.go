//go:build integration

package migrations_test

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"metaldocs/tests/integration/testdb"
)

// TestMigration0152_Columns verifies the snapshot columns were added to documents.
//
// editable_zones_schema_snapshot is intentionally absent from wantCols: it was
// dropped by archive/migrations/0157_drop_editable_zones.sql (and the sibling
// templates-side column by 0196_drop_templates_editable_zones.sql), confirmed
// absent from db/baseline/0001_current_schema.sql's public.documents definition.
func TestMigration0152_Columns(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)

	wantCols := []string{
		"placeholder_schema_snapshot",
		"placeholder_schema_hash",
		"composition_config_snapshot",
		"composition_config_hash",
		"body_docx_snapshot_s3_key",
		"body_docx_hash",
		"values_frozen_at",
		"values_hash",
		"final_docx_s3_key",
		"final_pdf_s3_key",
		"pdf_hash",
		"pdf_generated_at",
		"reconstruction_attempts",
	}

	// table_schema is the real Postgres schema ("public"), not the isolated
	// test database name testdb.Open returns as its second value — Qualified
	// ignores that value entirely (test isolation is at the database level,
	// per tests/integration/testdb/db.go), so information_schema lookups must
	// use the literal schema the runtime tables actually live in.
	for _, col := range wantCols {
		var found bool
		err := db.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_schema = 'public'
				  AND table_name   = 'documents'
				  AND column_name  = $1
			)`, col).Scan(&found)
		if err != nil {
			t.Fatalf("query column %s: %v", col, err)
		}
		if !found {
			t.Errorf("documents column %q not found in public schema", col)
		}
	}
}

// TestMigration0152_Tables verifies document_placeholder_values exists.
//
// document_editable_zone_content is intentionally not checked here: it was
// dropped by archive/migrations/0157_drop_editable_zones.sql, confirmed absent
// from db/baseline/0001_current_schema.sql (no CREATE TABLE for it anywhere in
// the baseline or db/migrations).
func TestMigration0152_Tables(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)

	for _, tbl := range []string{"document_placeholder_values"} {
		var found bool
		err := db.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM information_schema.tables
				WHERE table_schema = 'public'
				  AND table_name   = $1
			)`, tbl).Scan(&found)
		if err != nil {
			t.Fatalf("query table %s: %v", tbl, err)
		}
		if !found {
			t.Errorf("table %q not found in public schema", tbl)
		}
	}
}

// TestMigration0152_SnapshotTrigger verifies enforce_snapshot_on_submit blocks
// transitions to guarded statuses (under_review, approved, scheduled, published)
// when snapshot columns are NULL.
func TestMigration0152_SnapshotTrigger(t *testing.T) {
	ctx := context.Background()
	db, schema := testdb.Open(t)

	tenantID := testdb.DeterministicID(t, "tenant")
	userID := testdb.DeterministicID(t, "user")

	testdb.SeedUser(t, ctx, db, schema, userID, "Snapshot Test User")

	guardedStatuses := []string{"under_review", "approved", "scheduled", "published"}

	for _, status := range guardedStatuses {
		t.Run(status, func(t *testing.T) {
			docID := testdb.DeterministicID(t, fmt.Sprintf("doc-%s", status))
			testdb.SeedDocument(t, ctx, db, schema, docID, tenantID, userID)

			_, err := db.ExecContext(ctx, fmt.Sprintf(`
				UPDATE %s SET status = $2
				 WHERE id = $1::uuid`,
				testdb.Qualified(schema, "documents")),
				docID,
				status,
			)
			if err == nil {
				t.Fatalf("expected error from enforce_snapshot_on_submit trigger for status=%s, got nil", status)
			}
			if !strings.Contains(err.Error(), "snapshot columns required") &&
				!strings.Contains(err.Error(), "check_violation") &&
				!strings.Contains(err.Error(), "23514") {
				t.Fatalf("unexpected error for status=%s (wanted snapshot enforcement): %v", status, err)
			}
		})
	}
}

// TestMigration0152_PlaceholderValueInsert verifies a row can be inserted into
// document_placeholder_values when revision_id references a valid revision.
//
// document_placeholder_values.revision_id carries
// document_placeholder_values_revision_id_fkey -> public.document_revisions(id)
// (db/baseline/0001_current_schema.sql:4328-4329), NOT documents(id) — so the
// FK target must be an actual document_revisions row, seeded here via
// seedDocumentRevision (mirrors testdb.InsertDraftDocument's raw-insert idiom
// for the untripwired editor_sessions/document_revisions tables).
func TestMigration0152_PlaceholderValueInsert(t *testing.T) {
	ctx := context.Background()
	db, schema := testdb.Open(t)

	tenantID := testdb.DeterministicID(t, "tenant")
	docID := testdb.DeterministicID(t, "doc")
	userID := testdb.DeterministicID(t, "user")

	testdb.SeedUser(t, ctx, db, schema, userID, "PV Insert User")
	testdb.SeedDocument(t, ctx, db, schema, docID, tenantID, userID)
	revID := seedDocumentRevision(t, ctx, db, schema, docID, tenantID, userID)

	_, err := db.ExecContext(ctx, fmt.Sprintf(`
		INSERT INTO %s (tenant_id, revision_id, placeholder_id, source)
		VALUES ($1::uuid, $2::uuid, 'ph_title', 'user')`,
		testdb.Qualified(schema, "document_placeholder_values")),
		tenantID, revID,
	)
	if err != nil {
		t.Fatalf("insert document_placeholder_values: %v", err)
	}
}

// seedDocumentRevision inserts the editor_sessions + document_revisions rows a
// document_placeholder_values / document_editable_zone_content revision_id FK
// requires. Neither table carries a tripwire (mirrors testdb.InsertDraftDocument's
// raw-insert idiom in tests/integration/testdb/fixtures.go), so no capability
// assertion is needed. Returns the new revision id.
func seedDocumentRevision(t *testing.T, ctx context.Context, db *sql.DB, schema, docID, tenantID, userID string) string {
	t.Helper()

	var sessID string
	if err := db.QueryRowContext(ctx, `
		INSERT INTO `+testdb.Qualified(schema, "editor_sessions")+`
			(tenant_id, document_id, user_id, expires_at, last_acknowledged_revision_id, status)
		VALUES ($1::uuid, $2::uuid, $3, now() + interval '1 hour',
		        '00000000-0000-0000-0000-000000000000', 'active')
		RETURNING id::text`,
		tenantID, docID, userID,
	).Scan(&sessID); err != nil {
		t.Fatalf("seedDocumentRevision: insert editor_sessions: %v", err)
	}

	var revID string
	if err := db.QueryRowContext(ctx, `
		INSERT INTO `+testdb.Qualified(schema, "document_revisions")+`
			(document_id, parent_revision_id, session_id, storage_key, content_hash, form_data_snapshot)
		VALUES ($1::uuid, NULL, $2::uuid, '', 'aabbcc', '{}')
		RETURNING id::text`,
		docID, sessID,
	).Scan(&revID); err != nil {
		t.Fatalf("seedDocumentRevision: insert document_revisions: %v", err)
	}
	return revID
}
