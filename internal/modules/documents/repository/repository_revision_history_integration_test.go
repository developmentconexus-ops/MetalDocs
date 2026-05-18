//go:build integration
// +build integration

package repository_test

import (
	"context"
	"database/sql"
	"testing"

	"metaldocs/tests/integration/testdb"
)

func TestGovernedRevisionTitleColumnExistsAndCanBeRead(t *testing.T) {
	ctx := context.Background()
	db, schema := testdb.Open(t)
	db.SetMaxOpenConns(1)

	if _, err := db.ExecContext(ctx,
		`SELECT set_config('metaldocs.asserted_caps', '[{"cap":"document.create"},{"cap":"document.edit"}]', false)`,
	); err != nil {
		t.Fatalf("set asserted_caps: %v", err)
	}

	tenantID := testdb.DeterministicID(t, "tenant")
	docID, _ := testdb.InsertDraftDocument(t, db, schema, tenantID)

	if _, err := db.ExecContext(ctx,
		`UPDATE `+testdb.Qualified(schema, "documents")+` SET revision_title = $2 WHERE id = $1::uuid`,
		docID, "Correcao de procedimento",
	); err != nil {
		t.Fatalf("update revision_title: %v", err)
	}

	var title sql.NullString
	if err := db.QueryRowContext(ctx,
		`SELECT revision_title FROM `+testdb.Qualified(schema, "documents")+` WHERE id = $1::uuid`,
		docID,
	).Scan(&title); err != nil {
		t.Fatalf("scan revision_title: %v", err)
	}

	if !title.Valid || title.String != "Correcao de procedimento" {
		t.Fatalf("revision_title = %#v", title)
	}
}
