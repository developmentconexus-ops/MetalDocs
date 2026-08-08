//go:build integration
// +build integration

package infrastructure_test

import (
	"context"
	"database/sql"
	"testing"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/internal/modules/documents/infrastructure"
	iamdomain "metaldocs/internal/modules/iam/domain"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/tests/integration/testdb"
)

func TestGovernedRevisionTitleColumnExistsAndCanBeRead(t *testing.T) {
	ctx := context.Background()
	db, schema := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tnt := testdb.NewTenant(t, db)
	docID, _ := testdb.InsertDraftDocument(t, db, schema, tnt.ID)

	// UPDATE public.documents — guarded by document.edit tripwire.
	testdb.SeedWithCaps(t, db, `[{"cap":"document.edit"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`UPDATE `+testdb.Qualified(schema, "documents")+` SET revision_title = $2 WHERE id = $1::uuid`,
			docID, "Correcao de procedimento",
		)
		return err
	})

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

	tnt := testdb.NewTenant(t, db)
	controlledDocumentID := testdb.DeterministicID(t, "controlled-document")
	ownerUserID := testdb.DeterministicID(t, "user")
	profileCode := "qa"
	processAreaCode := "ops"

	firstDocID, _ := testdb.InsertDraftDocument(t, db, schema, tnt.ID)
	secondDocID, _ := testdb.InsertDraftDocument(t, db, schema, tnt.ID)

	// Seed taxonomy (document_families, document_profiles, document_process_areas)
	// — all guarded by taxonomy.manage tripwire.
	testdb.SeedWithCaps(t, db, `[{"cap":"taxonomy.manage"}]`, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO `+testdb.Qualified(schema, "document_families")+`
			    (code, name, description)
			 VALUES ('procedure', 'Procedimentos', 'familia global de teste')
			 ON CONFLICT (code) DO NOTHING`,
		); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO `+testdb.Qualified(schema, "document_profiles")+`
			    (code, tenant_id, family_code, name, description, review_interval_days, alias, editable_by_role)
			 VALUES ($1, $2::uuid, 'procedure', 'Perfil QA', 'perfil de teste', 30, 'qa', 'admin')`,
			profileCode, tnt.ID,
		); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO `+testdb.Qualified(schema, "document_process_areas")+`
			    (code, tenant_id, name, description)
			 VALUES ($1, $2::uuid, 'Operacoes', 'area de teste')`,
			processAreaCode, tnt.ID,
		); err != nil {
			return err
		}
		return nil
	})

	// Seed owner user — iam_users now carries a trg_require_cap_asserted
	// tripwire (user.manage), so the insert must run inside a seedWithCaps tx
	// like testdb.NewUser does.
	testdb.SeedWithCaps(t, db, `[{"cap":"user.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO metaldocs.iam_users (user_id, display_name, tenant_id)
			 VALUES ($1, 'Owner', $2::uuid)
			 ON CONFLICT (user_id) DO NOTHING`,
			ownerUserID, tnt.ID,
		)
		return err
	})

	// Seed controlled document — guarded by controlled_documents.create tripwire.
	testdb.SeedWithCaps(t, db, `[{"cap":"controlled_documents.create"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO `+testdb.Qualified(schema, "controlled_documents")+`
			    (id, tenant_id, profile_code, process_area_code, code, title, owner_user_id, status, visibility_scope)
			 VALUES ($1::uuid, $2::uuid, $3, $4, 'QA-OPS-001', 'Documento controlado de teste', $5, 'active', 'company')`,
			controlledDocumentID, tnt.ID, profileCode, processAreaCode, ownerUserID,
		)
		return err
	})

	// Attach firstDocID to the controlled document and set snapshot columns
	// (required by enforce_snapshot_on_submit for non-draft statuses).
	// UPDATE public.documents — guarded by document.edit tripwire.
	testdb.SeedWithCaps(t, db, `[{"cap":"document.edit"}]`, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx,
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
			firstDocID, controlledDocumentID,
		); err != nil {
			return err
		}

		// Walk firstDocID through the legal status lifecycle to 'published'.
		// ADR 0085: entering 'published' stamps the ACTUAL release instant;
		// ck_documents_published_effective_from rejects a published document
		// without one (that state is F-QA4-13 — invisible to the review surfacer).
		for _, nextStatus := range []string{"under_review", "approved", "published"} {
			set := `status = $2`
			if nextStatus == "published" {
				set += `, effective_from = '2026-05-10T12:00:00Z'::timestamptz`
			}
			if _, err := tx.ExecContext(ctx,
				`UPDATE `+testdb.Qualified(schema, "documents")+`
				    SET `+set+`
				  WHERE id = $1::uuid`,
				firstDocID, nextStatus,
			); err != nil {
				return err
			}
		}
		return nil
	})

	// Attach secondDocID to the controlled document.
	// UPDATE public.documents — guarded by document.edit tripwire.
	testdb.SeedWithCaps(t, db, `[{"cap":"document.edit"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`UPDATE `+testdb.Qualified(schema, "documents")+`
			    SET controlled_document_id = $2::uuid,
			        revision_number = 1,
			        revision_title = 'Ajuste operacional',
			        created_at = '2026-05-18T12:00:00Z'::timestamptz,
			        updated_at = '2026-05-18T12:00:00Z'::timestamptz
			  WHERE id = $1::uuid`,
			secondDocID, controlledDocumentID,
		)
		return err
	})

	// Insert autosave revisions for secondDocID (public.document_revisions has no tripwire).
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

	repo := infrastructure.New(db, iamdomain.NoopUserDisplayNameReader{}, controlleddocumentsdomain.NoopCDFieldReader{}, taxonomydomain.NoopAreaCatalogReader{})
	items, err := repo.ListRevisionHistory(ctx, tnt.ID, secondDocID)
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
