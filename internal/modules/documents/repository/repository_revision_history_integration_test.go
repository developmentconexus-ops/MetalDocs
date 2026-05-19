//go:build integration
// +build integration

package repository_test

import (
	"context"
	"database/sql"
	"testing"

	"metaldocs/internal/modules/documents/repository"
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

func TestListRevisionHistory_ReturnsGovernedDocumentsNotAutosaveRows(t *testing.T) {
	ctx := context.Background()
	db, schema := testdb.Open(t)
	db.SetMaxOpenConns(1)

	if _, err := db.ExecContext(ctx,
		`SELECT set_config('metaldocs.asserted_caps', '[{"cap":"document.create"},{"cap":"document.edit"},{"cap":"registry.create"},{"cap":"taxonomy.manage"}]', false)`,
	); err != nil {
		t.Fatalf("set asserted_caps: %v", err)
	}

	tenantID := testdb.DeterministicID(t, "tenant-history")
	controlledDocumentID := testdb.DeterministicID(t, "controlled-document")
	ownerUserID := testdb.DeterministicID(t, "user")
	profileCode := "qa"
	processAreaCode := "ops"

	firstDocID, _ := testdb.InsertDraftDocument(t, db, schema, tenantID)
	secondDocID, _ := testdb.InsertDraftDocument(t, db, schema, tenantID)

	if _, err := db.ExecContext(ctx,
		`INSERT INTO `+testdb.Qualified(schema, "document_families")+`
		    (code, name, description)
		 VALUES ('procedure', 'Procedimentos', 'familia global de teste')`,
	); err != nil {
		t.Fatalf("insert document_family: %v", err)
	}

	if _, err := db.ExecContext(ctx,
		`INSERT INTO `+testdb.Qualified(schema, "document_profiles")+`
		    (code, tenant_id, family_code, name, description, review_interval_days, alias, editable_by_role)
		 VALUES ($1, $2::uuid, 'procedure', 'Perfil QA', 'perfil de teste', 30, 'qa', 'admin')`,
		profileCode,
		tenantID,
	); err != nil {
		t.Fatalf("insert document_profile: %v", err)
	}

	if _, err := db.ExecContext(ctx,
		`INSERT INTO `+testdb.Qualified(schema, "document_process_areas")+`
		    (code, tenant_id, name, description)
		 VALUES ($1, $2::uuid, 'Operacoes', 'area de teste')`,
		processAreaCode,
		tenantID,
	); err != nil {
		t.Fatalf("insert document_process_area: %v", err)
	}

	if _, err := db.ExecContext(ctx,
		`INSERT INTO `+testdb.Qualified(schema, "controlled_documents")+`
		    (id, tenant_id, profile_code, process_area_code, code, title, owner_user_id, status, visibility_scope)
		 VALUES ($1::uuid, $2::uuid, $3, $4, 'QA-OPS-001', 'Documento controlado de teste', $5, 'active', 'company')`,
		controlledDocumentID,
		tenantID,
		profileCode,
		processAreaCode,
		ownerUserID,
	); err != nil {
		t.Fatalf("insert controlled_document: %v", err)
	}

	if _, err := db.ExecContext(ctx,
		`UPDATE `+testdb.Qualified(schema, "documents")+`
		    SET controlled_document_id = $2::uuid,
		        revision_number = 0,
		        revision_title = 'Primeira versao',
		        created_at = '2026-05-10T12:00:00Z'::timestamptz,
		        updated_at = '2026-05-10T12:00:00Z'::timestamptz,
		        placeholder_schema_snapshot = '[]'::jsonb,
		        placeholder_schema_hash = decode(repeat('11', 32), 'hex'),
		        composition_config_snapshot = '{}'::jsonb,
		        composition_config_hash = decode(repeat('22', 32), 'hex'),
		        body_docx_snapshot_s3_key = 'snapshots/doc-1.docx',
		        body_docx_hash = decode(repeat('33', 32), 'hex')
		  WHERE id = $1::uuid`,
		firstDocID,
		controlledDocumentID,
	); err != nil {
		t.Fatalf("prepare historical governed document %s: %v", firstDocID, err)
	}

	for _, nextStatus := range []string{"under_review", "approved", "published"} {
		if _, err := db.ExecContext(ctx,
			`UPDATE `+testdb.Qualified(schema, "documents")+`
			    SET status = $2
			  WHERE id = $1::uuid`,
			firstDocID,
			nextStatus,
		); err != nil {
			t.Fatalf("transition historical governed document to %s: %v", nextStatus, err)
		}
	}

	if _, err := db.ExecContext(ctx,
		`UPDATE `+testdb.Qualified(schema, "documents")+`
		    SET controlled_document_id = $2::uuid,
		        revision_number = 1,
		        revision_title = 'Ajuste operacional',
		        created_at = '2026-05-18T12:00:00Z'::timestamptz,
		        updated_at = '2026-05-18T12:00:00Z'::timestamptz
		  WHERE id = $1::uuid`,
		secondDocID,
		controlledDocumentID,
	); err != nil {
		t.Fatalf("prepare current governed document %s: %v", secondDocID, err)
	}

	for i := 0; i < 3; i++ {
		if _, err := db.ExecContext(ctx,
			`INSERT INTO `+testdb.Qualified(schema, "document_revisions")+`
			    (document_id, parent_revision_id, session_id, storage_key, content_hash, form_data_snapshot)
			 VALUES ($1::uuid, NULL, (
			     SELECT active_session_id FROM `+testdb.Qualified(schema, "documents")+` WHERE id = $1::uuid
			 ), $2, $3, '{}'::jsonb)`,
			secondDocID,
			"autosave/"+testdb.DeterministicID(t, "storage"),
			testdb.DeterministicID(t, "hash-"+string(rune('a'+i))),
		); err != nil {
			t.Fatalf("insert technical autosave %d: %v", i, err)
		}
	}

	repo := repository.New(db)
	items, err := repo.ListRevisionHistory(ctx, tenantID, secondDocID)
	if err != nil {
		t.Fatalf("ListRevisionHistory: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}
	if items[0].DocumentID != secondDocID || items[0].RevisionTitle != "Ajuste operacional" || !items[0].IsCurrent {
		t.Fatalf("items[0] = %#v", items[0])
	}
	if items[0].RevisionNumber != 1 {
		t.Fatalf("items[0].RevisionNumber = %d, want 1", items[0].RevisionNumber)
	}
	if items[1].DocumentID != firstDocID || items[1].RevisionTitle != "Primeira versao" || items[1].IsCurrent {
		t.Fatalf("items[1] = %#v", items[1])
	}
	if items[1].RevisionNumber != 0 {
		t.Fatalf("items[1].RevisionNumber = %d, want 0", items[1].RevisionNumber)
	}
}
