//go:build integration
// +build integration

package repository_test

import (
	"context"
	"strings"
	"testing"

	"metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/documents/repository"
	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/tests/integration/testdb"
)

// TestCreateDocumentTx_StorageKeyInvariant verifies that CreateDocumentTx persists
// the initialStorageKey as-is on the revision row: non-empty for atomic-create,
// empty string for legacy flow.
func TestCreateDocumentTx_StorageKeyInvariant(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tnt := testdb.NewTenant(t, db)
	actor := testdb.NewUser(t, db, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tnt.ID))
	cd := testdb.NewControlledDoc(t, db,
		testdb.WithTenant(tnt.ID),
		testdb.WithTaxonomy(tax),
		testdb.WithOwner(actor.ID),
	)

	templateVersionID := testdb.DeterministicID(t, "template-version")
	repo := repository.New(db, iamdomain.NoopUserDisplayNameReader{}, controlleddocumentsdomain.NoopCDFieldReader{}, taxonomydomain.NoopAreaCatalogReader{})

	newDocument := func(name, code string) *domain.Document {
		profileCode := cd.ProfileCode
		processAreaCode := cd.ProcessAreaCode
		controlledDocID := cd.ID
		return &domain.Document{
			TenantID:                tnt.ID,
			TemplateVersionID:       templateVersionID,
			Name:                    name,
			FormDataJSON:            []byte(`{}`),
			CreatedBy:               actor.ID,
			ControlledDocumentID:    &controlledDocID,
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

		// Assert document.create + document.edit tx-locally (mirrors production authz layer).
		testdb.SetCapsOnTx(t, tx, `[{"cap":"document.create"},{"cap":"document.edit"}]`)

		wantKey := "tenants/test/templates/published.docx"
		_, revID, _, err := repo.CreateDocumentTx(ctx, tx, newDocument("Atomic Doc", "PO-SK-001"), "hash-atomic", wantKey, nil)
		if err != nil {
			t.Fatalf("CreateDocumentTx: %v", err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatalf("commit: %v", err)
		}

		var gotKey string
		if err := db.QueryRowContext(ctx,
			`SELECT storage_key FROM public.document_revisions WHERE id = $1::uuid`, revID,
		).Scan(&gotKey); err != nil {
			t.Fatalf("query storage_key: %v", err)
		}
		if gotKey != wantKey {
			t.Fatalf("storage_key = %q, want %q", gotKey, wantKey)
		}
	})

	t.Run("EmptyStorageKey", func(t *testing.T) {
		// Supersede the first doc so the second create succeeds:
		// ux_documents_cd_active permits one active document per CD.
		testdb.SupersedeActiveDocumentForCD(t, db, cd.ID)

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}
		defer tx.Rollback()

		// Assert caps tx-locally before repo call.
		testdb.SetCapsOnTx(t, tx, `[{"cap":"document.create"},{"cap":"document.edit"}]`)

		_, revID, _, err := repo.CreateDocumentTx(ctx, tx, newDocument("Legacy Doc", "PO-SK-002"), "hash-legacy", "", nil)
		if err != nil {
			t.Fatalf("CreateDocumentTx: %v", err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatalf("commit: %v", err)
		}

		var gotKey string
		if err := db.QueryRowContext(ctx,
			`SELECT storage_key FROM public.document_revisions WHERE id = $1::uuid`, revID,
		).Scan(&gotKey); err != nil {
			t.Fatalf("query storage_key: %v", err)
		}
		if gotKey != "" {
			t.Fatalf("storage_key = %q, want empty string", gotKey)
		}
	})
}

// TestCreateDocumentTx_RevisionNumberIncrementsForSameCD verifies that two documents
// created for the same controlled_document_id get revision_number 0 and 1.
func TestCreateDocumentTx_RevisionNumberIncrementsForSameCD(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tnt := testdb.NewTenant(t, db)
	actor := testdb.NewUser(t, db, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tnt.ID))
	cd := testdb.NewControlledDoc(t, db,
		testdb.WithTenant(tnt.ID),
		testdb.WithTaxonomy(tax),
		testdb.WithOwner(actor.ID),
	)

	templateVersionID := testdb.DeterministicID(t, "template-version")
	repo := repository.New(db, iamdomain.NoopUserDisplayNameReader{}, controlleddocumentsdomain.NoopCDFieldReader{}, taxonomydomain.NoopAreaCatalogReader{})
	newDocument := func(name, code string) *domain.Document {
		profileCode := cd.ProfileCode
		processAreaCode := cd.ProcessAreaCode
		controlledDocID := cd.ID
		return &domain.Document{
			TenantID:                tnt.ID,
			TemplateVersionID:       templateVersionID,
			Name:                    name,
			FormDataJSON:            []byte(`{}`),
			CreatedBy:               actor.ID,
			ControlledDocumentID:    &controlledDocID,
			ProfileCodeSnapshot:     &profileCode,
			ProcessAreaCodeSnapshot: &processAreaCode,
			Code:                    code,
		}
	}

	tx1, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx1: %v", err)
	}
	testdb.SetCapsOnTx(t, tx1, `[{"cap":"document.create"},{"cap":"document.edit"}]`)
	firstDocID, _, _, err := repo.CreateDocumentTx(ctx, tx1, newDocument("Revision 1", "PO-REV-001"), "hash-1", "", nil)
	if err != nil {
		t.Fatalf("CreateDocumentTx first: %v", err)
	}
	if err := tx1.Commit(); err != nil {
		t.Fatalf("commit tx1: %v", err)
	}

	// Supersede the first doc so the second create succeeds.
	testdb.SupersedeActiveDocumentForCD(t, db, cd.ID)

	tx2, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx2: %v", err)
	}
	testdb.SetCapsOnTx(t, tx2, `[{"cap":"document.create"},{"cap":"document.edit"}]`)
	secondDocID, _, _, err := repo.CreateDocumentTx(ctx, tx2, newDocument("Revision 2", "PO-REV-002"), "hash-2", "", nil)
	if err != nil {
		t.Fatalf("CreateDocumentTx second: %v", err)
	}
	if err := tx2.Commit(); err != nil {
		t.Fatalf("commit tx2: %v", err)
	}

	var firstRevision, secondRevision int
	if err := db.QueryRowContext(ctx,
		`SELECT revision_number FROM public.documents WHERE id = $1::uuid`, firstDocID,
	).Scan(&firstRevision); err != nil {
		t.Fatalf("query first revision_number: %v", err)
	}
	if err := db.QueryRowContext(ctx,
		`SELECT revision_number FROM public.documents WHERE id = $1::uuid`, secondDocID,
	).Scan(&secondRevision); err != nil {
		t.Fatalf("query second revision_number: %v", err)
	}

	if firstRevision != 0 {
		t.Fatalf("first revision_number = %d, want 0", firstRevision)
	}
	if secondRevision != 1 {
		t.Fatalf("second revision_number = %d, want 1", secondRevision)
	}
}

// TestCreateDocumentTx_RejectsEmptyName verifies the documents_name_not_empty
// CHECK constraint blocks rows with empty/whitespace-only name values.
func TestCreateDocumentTx_RejectsEmptyName(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tnt := testdb.NewTenant(t, db)
	actor := testdb.NewUser(t, db, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tnt.ID))
	cd := testdb.NewControlledDoc(t, db,
		testdb.WithTenant(tnt.ID),
		testdb.WithTaxonomy(tax),
		testdb.WithOwner(actor.ID),
	)

	templateVersionID := testdb.DeterministicID(t, "template-version")
	repo := repository.New(db, iamdomain.NoopUserDisplayNameReader{}, controlleddocumentsdomain.NoopCDFieldReader{}, taxonomydomain.NoopAreaCatalogReader{})

	profileCode := cd.ProfileCode
	processAreaCode := cd.ProcessAreaCode
	controlledDocID := cd.ID
	doc := &domain.Document{
		TenantID:                tnt.ID,
		TemplateVersionID:       templateVersionID,
		Name:                    "",
		FormDataJSON:            []byte(`{}`),
		CreatedBy:               actor.ID,
		ControlledDocumentID:    &controlledDocID,
		ProfileCodeSnapshot:     &profileCode,
		ProcessAreaCodeSnapshot: &processAreaCode,
		Code:                    "PO-EMPTY-CHECK",
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	// Assert caps tx-locally before repo call.
	testdb.SetCapsOnTx(t, tx, `[{"cap":"document.create"},{"cap":"document.edit"}]`)

	_, _, _, err = repo.CreateDocumentTx(ctx, tx, doc, "hash", "", nil)
	if err == nil || !strings.Contains(err.Error(), "documents_name_not_empty") {
		t.Fatalf("expected CHECK violation documents_name_not_empty, got: %v", err)
	}
}

func TestGetDocument_ReturnsSnapshotMetadata(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	tnt := testdb.NewTenant(t, db)
	actor := testdb.NewUser(t, db, testdb.WithTenant(tnt.ID), testdb.WithRole("system_admin"))
	tax := testdb.NewTaxonomy(t, db, testdb.WithTenant(tnt.ID))
	cd := testdb.NewControlledDoc(t, db,
		testdb.WithTenant(tnt.ID),
		testdb.WithTaxonomy(tax),
		testdb.WithOwner(actor.ID),
	)

	templateVersionID := testdb.DeterministicID(t, "template-version")
	repo := repository.New(db, iamdomain.NoopUserDisplayNameReader{}, controlleddocumentsdomain.NoopCDFieldReader{}, taxonomydomain.NoopAreaCatalogReader{})

	profileCode := cd.ProfileCode
	processAreaCode := cd.ProcessAreaCode
	controlledDocID := cd.ID
	doc := &domain.Document{
		TenantID:                tnt.ID,
		TemplateVersionID:       templateVersionID,
		Name:                    "Snapshot Metadata Doc",
		FormDataJSON:            []byte(`{}`),
		CreatedBy:               actor.ID,
		ControlledDocumentID:    &controlledDocID,
		ProfileCodeSnapshot:     &profileCode,
		ProcessAreaCodeSnapshot: &processAreaCode,
		Code:                    "POP-GENERAL-001",
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	// Assert caps tx-locally before repo call.
	testdb.SetCapsOnTx(t, tx, `[{"cap":"document.create"},{"cap":"document.edit"}]`)

	docID, _, _, err := repo.CreateDocumentTx(ctx, tx, doc, "hash-snapshot", "", nil)
	if err != nil {
		t.Fatalf("CreateDocumentTx: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}

	got, err := repo.GetDocument(ctx, tnt.ID, docID)
	if err != nil {
		t.Fatalf("GetDocument: %v", err)
	}
	if got.ProfileCodeSnapshot == nil || *got.ProfileCodeSnapshot != profileCode {
		t.Fatalf("ProfileCodeSnapshot = %#v, want %q", got.ProfileCodeSnapshot, profileCode)
	}
	if got.ProcessAreaCodeSnapshot == nil || *got.ProcessAreaCodeSnapshot != processAreaCode {
		t.Fatalf("ProcessAreaCodeSnapshot = %#v, want %q", got.ProcessAreaCodeSnapshot, processAreaCode)
	}
}
