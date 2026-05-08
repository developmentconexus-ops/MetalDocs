//go:build integration
// +build integration

package repository_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/documents/repository"
	"metaldocs/tests/integration/testdb"
)

// TestCreateDocumentTx_StorageKeyInvariant verifies that CreateDocumentTx persists
// the initialStorageKey as-is on the revision row: non-empty for atomic-create,
// empty string for legacy flow.
func TestCreateDocumentTx_StorageKeyInvariant(t *testing.T) {
	ctx := context.Background()
	db, schema := testdb.Open(t)
	db.SetMaxOpenConns(1)

	if _, err := db.ExecContext(ctx, fmt.Sprintf(`SET search_path TO %q`, schema)); err != nil {
		t.Fatalf("set search_path: %v", err)
	}

	tenantID := testdb.DeterministicID(t, "tenant")
	actorID := testdb.DeterministicID(t, "actor")
	templateVersionID := testdb.DeterministicID(t, "template-version")
	controlledDocumentID := testdb.DeterministicID(t, "controlled-document")

	if _, err := db.ExecContext(ctx, `
		INSERT INTO controlled_documents (
			id, tenant_id, profile_code, process_area_code, code, title, owner_user_id, status
		) VALUES (
			$1::uuid, $2::uuid, 'po', 'quality', 'PO-SK-001',
			'StorageKey Invariant CD', $3::uuid, 'active'
		)`,
		controlledDocumentID, tenantID, actorID,
	); err != nil {
		t.Fatalf("seed controlled_documents: %v", err)
	}

	profileCode := "po"
	processAreaCode := "quality"
	repo := repository.New(db)

	newDocument := func(name, code string) *domain.Document {
		return &domain.Document{
			TenantID:                tenantID,
			TemplateVersionID:       templateVersionID,
			Name:                    name,
			FormDataJSON:            []byte(`{}`),
			CreatedBy:               actorID,
			ControlledDocumentID:    &controlledDocumentID,
			ProfileCodeSnapshot:     &profileCode,
			ProcessAreaCodeSnapshot: &processAreaCode,
			Code:                    code,
		}
	}

	t.Run("NonEmptyStorageKey", func(t *testing.T) {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}
		defer tx.Rollback()

		wantKey := "tenants/test/templates/published.docx"
		_, revID, _, err := repo.CreateDocumentTx(ctx, tx, newDocument("Atomic Doc", "PO-SK-001"), "hash-atomic", wantKey, nil)
		if err != nil {
			t.Fatalf("CreateDocumentTx: %v", err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatalf("commit: %v", err)
		}

		var gotKey string
		if err := db.QueryRowContext(ctx, `SELECT storage_key FROM document_revisions WHERE id = $1::uuid`, revID).Scan(&gotKey); err != nil {
			t.Fatalf("query storage_key: %v", err)
		}
		if gotKey != wantKey {
			t.Fatalf("storage_key = %q, want %q", gotKey, wantKey)
		}
	})

	t.Run("EmptyStorageKey_LegacyPath", func(t *testing.T) {
		// Publish first doc so the sequence advances and the second insert succeeds.
		if _, err := db.ExecContext(ctx, `UPDATE documents SET status='published' WHERE controlled_document_id = $1::uuid`, controlledDocumentID); err != nil {
			t.Fatalf("publish first: %v", err)
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}
		defer tx.Rollback()

		_, revID, _, err := repo.CreateDocumentTx(ctx, tx, newDocument("Legacy Doc", "PO-SK-002"), "hash-legacy", "", nil)
		if err != nil {
			t.Fatalf("CreateDocumentTx: %v", err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatalf("commit: %v", err)
		}

		var gotKey string
		if err := db.QueryRowContext(ctx, `SELECT storage_key FROM document_revisions WHERE id = $1::uuid`, revID).Scan(&gotKey); err != nil {
			t.Fatalf("query storage_key: %v", err)
		}
		if gotKey != "" {
			t.Fatalf("storage_key = %q, want empty string", gotKey)
		}
	})
}

// TestCreateDocument_RevisionNumberIncrementsForSameCD verifies that two documents
// created for the same controlled_document_id get revision_number 1 and 2.
func TestCreateDocument_RevisionNumberIncrementsForSameCD(t *testing.T) {
	ctx := context.Background()
	db, schema := testdb.Open(t)
	db.SetMaxOpenConns(1)

	if _, err := db.ExecContext(ctx, fmt.Sprintf(`SET search_path TO %q`, schema)); err != nil {
		t.Fatalf("set search_path: %v", err)
	}

	tenantID := testdb.DeterministicID(t, "tenant")
	actorID := testdb.DeterministicID(t, "actor")
	templateVersionID := testdb.DeterministicID(t, "template-version")
	controlledDocumentID := testdb.DeterministicID(t, "controlled-document")

	if _, err := db.ExecContext(ctx, `
		INSERT INTO controlled_documents (
			id, tenant_id, profile_code, process_area_code, code, title, owner_user_id, status
		) VALUES (
			$1::uuid, $2::uuid, 'po', 'quality', 'PO-REV-001',
			'Revision Test Controlled Document', $3::uuid, 'active'
		)`,
		controlledDocumentID, tenantID, actorID,
	); err != nil {
		t.Fatalf("seed controlled_documents: %v", err)
	}

	profileCode := "po"
	processAreaCode := "quality"
	repo := repository.New(db)
	newDocument := func(name, code string) *domain.Document {
		return &domain.Document{
			TenantID:                tenantID,
			TemplateVersionID:       templateVersionID,
			Name:                    name,
			FormDataJSON:            []byte(`{}`),
			CreatedBy:               actorID,
			ControlledDocumentID:    &controlledDocumentID,
			ProfileCodeSnapshot:     &profileCode,
			ProcessAreaCodeSnapshot: &processAreaCode,
			Code:                    code,
		}
	}

	firstDocID, _, _, err := repo.CreateDocument(ctx, newDocument("Revision 1", "PO-REV-001"), "hash-1", nil)
	if err != nil {
		t.Fatalf("CreateDocument first: %v", err)
	}

	if _, err := db.ExecContext(ctx, `UPDATE documents SET status='published' WHERE id=$1::uuid`, firstDocID); err != nil {
		t.Fatalf("publish first document: %v", err)
	}

	secondDocID, _, _, err := repo.CreateDocument(ctx, newDocument("Revision 2", "PO-REV-002"), "hash-2", nil)
	if err != nil {
		t.Fatalf("CreateDocument second: %v", err)
	}

	var firstRevision, secondRevision int
	if err := db.QueryRowContext(ctx, `SELECT revision_number FROM documents WHERE id = $1::uuid`, firstDocID).Scan(&firstRevision); err != nil {
		t.Fatalf("query first revision_number: %v", err)
	}
	if err := db.QueryRowContext(ctx, `SELECT revision_number FROM documents WHERE id = $1::uuid`, secondDocID).Scan(&secondRevision); err != nil {
		t.Fatalf("query second revision_number: %v", err)
	}

	if firstRevision != 1 {
		t.Fatalf("first revision_number = %d, want 1", firstRevision)
	}
	if secondRevision != 2 {
		t.Fatalf("second revision_number = %d, want 2", secondRevision)
	}
}

// TestCreateDocument_RejectsEmptyName verifies the documents_name_not_empty
// CHECK constraint blocks rows with empty/whitespace-only name values.
func TestCreateDocument_RejectsEmptyName(t *testing.T) {
	ctx := context.Background()
	db, schema := testdb.Open(t)
	db.SetMaxOpenConns(1)

	if _, err := db.ExecContext(ctx, fmt.Sprintf(`SET search_path TO %q`, schema)); err != nil {
		t.Fatalf("set search_path: %v", err)
	}

	tenantID := testdb.DeterministicID(t, "tenant")
	actorID := testdb.DeterministicID(t, "actor")
	templateVersionID := testdb.DeterministicID(t, "template-version")
	controlledDocumentID := testdb.DeterministicID(t, "controlled-document")

	if _, err := db.ExecContext(ctx, `
		INSERT INTO controlled_documents (
			id, tenant_id, profile_code, process_area_code, code, title, owner_user_id, status
		) VALUES (
			$1::uuid, $2::uuid, 'po', 'quality', 'PO-EMPTY-CHECK',
			'Empty Name Check CD', $3::uuid, 'active'
		)`,
		controlledDocumentID, tenantID, actorID,
	); err != nil {
		t.Fatalf("seed controlled_documents: %v", err)
	}

	profileCode := "po"
	processAreaCode := "quality"
	repo := repository.New(db)

	doc := &domain.Document{
		TenantID:                tenantID,
		TemplateVersionID:       templateVersionID,
		Name:                    "",
		FormDataJSON:            []byte(`{}`),
		CreatedBy:               actorID,
		ControlledDocumentID:    &controlledDocumentID,
		ProfileCodeSnapshot:     &profileCode,
		ProcessAreaCodeSnapshot: &processAreaCode,
		Code:                    "PO-EMPTY-CHECK",
	}

	_, _, _, err := repo.CreateDocument(ctx, doc, "hash", nil)
	if err == nil || !strings.Contains(err.Error(), "documents_name_not_empty") {
		t.Fatalf("expected CHECK violation documents_name_not_empty, got: %v", err)
	}
}
