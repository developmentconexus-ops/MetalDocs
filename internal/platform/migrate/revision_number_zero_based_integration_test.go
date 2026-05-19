//go:build integration

package migrate

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestDocumentsRevisionNumberZeroBasedShiftAvoidsUniqueCollisions(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		dsn = strings.TrimSpace(os.Getenv("METALDOCS_DATABASE_URL"))
	}
	if dsn == "" {
		t.Skip("DATABASE_URL or METALDOCS_DATABASE_URL is not set")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

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
