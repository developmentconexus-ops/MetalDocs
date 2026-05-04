//go:build integration
// +build integration

package repository_test

import (
	"context"
	"fmt"
	"testing"

	"metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/documents/repository"
	"metaldocs/tests/integration/testdb"
)

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

	firstDocID, _, _, err := repo.CreateDocument(ctx, newDocument("Revision 1", "PO-REV-001"), "hash-1")
	if err != nil {
		t.Fatalf("CreateDocument first: %v", err)
	}

	if _, err := db.ExecContext(ctx, `UPDATE documents SET status='published' WHERE id=$1::uuid`, firstDocID); err != nil {
		t.Fatalf("publish first document: %v", err)
	}

	secondDocID, _, _, err := repo.CreateDocument(ctx, newDocument("Revision 2", "PO-REV-002"), "hash-2")
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
