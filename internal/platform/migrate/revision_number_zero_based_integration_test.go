//go:build integration

// package migrate_test (external), not migrate: this file uses no unexported
// migrate symbols, and internal/platform/migrate's C6 fix (PR #110 review --
// testdb.listSQLFiles now shares migrate.IsApplicableSQLFile instead of its
// own copy) made testdb import migrate. An in-package (package migrate) test
// file importing testdb would cycle: migrate(test) -> testdb -> migrate.
package migrate_test

import (
	"context"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"metaldocs/tests/integration/testdb"
)

// TestDocumentsRevisionNumberZeroBasedShiftAvoidsUniqueCollisions runs against
// a reset-safe leased clone from the canonical testdb factory (ADR 0034), NOT
// a raw sql.Open against the shared dev database. The prior body read
// DATABASE_URL/METALDOCS_DATABASE_URL directly and skipped-as-green when
// unset — the cluster-4 framework bypass. testdb.Open owns t.Helper, the
// leased-DB reset, cleanup, and a fail-loud ping. The test itself only ever
// touched a session-local TEMP TABLE, so this migration changes nothing about
// what schema state the assertions depend on — it only fixes how the
// connection is obtained.
func TestDocumentsRevisionNumberZeroBasedShiftAvoidsUniqueCollisions(t *testing.T) {
	db, _ := testdb.Open(t)

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `
		CREATE TEMP TABLE documents (
			id text PRIMARY KEY,
			controlled_document_id text,
			revision_number integer NOT NULL
		) ON COMMIT DROP;
		CREATE UNIQUE INDEX ux_documents_cd_revision_test
			ON documents (controlled_document_id, revision_number)
			WHERE controlled_document_id IS NOT NULL;
		INSERT INTO documents (id, controlled_document_id, revision_number)
		VALUES
			('rev-01', 'cd-1', 1),
			('rev-02', 'cd-1', 2),
			('legacy-uncontrolled', NULL, 1);
	`); err != nil {
		t.Fatalf("seed temp documents: %v", err)
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE documents
		   SET revision_number = revision_number + 1000000000
		 WHERE controlled_document_id IS NOT NULL
		   AND revision_number > 0;

		UPDATE documents
		   SET revision_number = revision_number - 1000000001
		 WHERE controlled_document_id IS NOT NULL
		   AND revision_number >= 1000000001;

		UPDATE documents
		   SET revision_number = revision_number - 1
		 WHERE controlled_document_id IS NULL
		   AND revision_number > 0;
	`); err != nil {
		t.Fatalf("two-phase zero-based shift: %v", err)
	}

	rows, err := tx.QueryContext(ctx, `SELECT id, revision_number FROM documents ORDER BY id`)
	if err != nil {
		t.Fatalf("query shifted rows: %v", err)
	}
	defer rows.Close()

	got := map[string]int{}
	for rows.Next() {
		var id string
		var revisionNumber int
		if err := rows.Scan(&id, &revisionNumber); err != nil {
			t.Fatalf("scan shifted row: %v", err)
		}
		got[id] = revisionNumber
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate shifted rows: %v", err)
	}

	want := map[string]int{
		"rev-01":              0,
		"rev-02":              1,
		"legacy-uncontrolled": 0,
	}
	for id, wantRevisionNumber := range want {
		if got[id] != wantRevisionNumber {
			t.Fatalf("%s revision_number = %d, want %d (all rows: %#v)", id, got[id], wantRevisionNumber, got)
		}
	}
}
