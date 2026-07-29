//go:build integration
// +build integration

package infrastructure_test

import (
	"context"
	"database/sql"
	"encoding/hex"
	"testing"
	"time"

	"metaldocs/internal/modules/documents/infrastructure"
	"metaldocs/internal/platform/tenant"
	"metaldocs/tests/integration/testdb"
)

const snapshotTestTenantID = tenant.DevTenantID

func TestSnapshotRepository_WriteFreeze_PersistsHashFrozenAtAndPin(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	docID, tnt := testdb.InsertDraftDocument(t, db, "", snapshotTestTenantID)
	repo := infrastructure.NewSnapshotRepository(db)

	// The seam is what supplies the pin in production, so read it the same way
	// here — a literal revision id would prove nothing about the join.
	current, err := repo.ReadCurrentRevisionRef(ctx, tnt, docID)
	if err != nil {
		t.Fatalf("ReadCurrentRevisionRef: %v", err)
	}
	if current.ID == "" {
		t.Fatal("fixture document has no current revision; cannot exercise the pin")
	}

	hash, err := hex.DecodeString("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err != nil {
		t.Fatalf("decode hash: %v", err)
	}
	frozenAt := time.Date(2026, 4, 23, 18, 0, 0, 0, time.UTC)

	// WriteFreeze updates public.documents — guarded by document.edit tripwire.
	// Pass the tx via the variadic DBTX parameter; cap set tx-locally via SeedWithCaps.
	testdb.SeedWithCaps(t, db, `[{"cap":"document.edit"}]`, func(tx *sql.Tx) error {
		return repo.WriteFreeze(ctx, tnt, docID, hash, current.ID, frozenAt, tx)
	})

	var gotHash []byte
	var gotFrozenAt *time.Time
	var gotPin *string
	if err := db.QueryRowContext(ctx, `
		SELECT values_hash, values_frozen_at, frozen_revision_id::text
		  FROM public.documents
		 WHERE tenant_id=$1::uuid AND id=$2::uuid`,
		tnt, docID,
	).Scan(&gotHash, &gotFrozenAt, &gotPin); err != nil {
		t.Fatalf("read freeze columns: %v", err)
	}
	if hex.EncodeToString(gotHash) != hex.EncodeToString(hash) {
		t.Fatalf("values_hash mismatch: got %x want %x", gotHash, hash)
	}
	if gotFrozenAt == nil || !gotFrozenAt.Equal(frozenAt) {
		t.Fatalf("values_frozen_at mismatch: got %v want %v", gotFrozenAt, frozenAt)
	}
	// Migration 0313: the lineage pin lands in the SAME update as the hash.
	if gotPin == nil || *gotPin != current.ID {
		t.Fatalf("frozen_revision_id mismatch: got %v want %s", gotPin, current.ID)
	}

	// And the frozen-side read must now resolve that pin to the same revision.
	frozen, err := repo.ReadFrozenRevisionRef(ctx, tnt, docID)
	if err != nil {
		t.Fatalf("ReadFrozenRevisionRef: %v", err)
	}
	if frozen.ID != current.ID {
		t.Fatalf("ReadFrozenRevisionRef id = %q, want %q", frozen.ID, current.ID)
	}
	if frozen.ContentHash != current.ContentHash || frozen.ContentHash == "" {
		t.Fatalf("ReadFrozenRevisionRef content_hash = %q, want %q", frozen.ContentHash, current.ContentHash)
	}
}

// An unfrozen document has a NULL pin, and the frozen-side read must report
// that as "nothing pinned" rather than silently resolving to the current head —
// falling back to the head is the exact drift the pin exists to prevent.
func TestSnapshotRepository_ReadFrozenRevisionRef_EmptyWhenUnpinned(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	docID, tnt := testdb.InsertDraftDocument(t, db, "", snapshotTestTenantID)
	repo := infrastructure.NewSnapshotRepository(db)

	current, err := repo.ReadCurrentRevisionRef(ctx, tnt, docID)
	if err != nil {
		t.Fatalf("ReadCurrentRevisionRef: %v", err)
	}
	if current.ID == "" {
		t.Fatal("fixture document has no current revision")
	}

	frozen, err := repo.ReadFrozenRevisionRef(ctx, tnt, docID)
	if err != nil {
		t.Fatalf("ReadFrozenRevisionRef: %v", err)
	}
	if frozen.ID != "" || frozen.StorageKey != "" || frozen.ContentHash != "" {
		t.Fatalf("unpinned document returned %+v, want zero RevisionRef", frozen)
	}
}

func TestSnapshotRepository_WriteFinalDocx_PersistsKeyAndContentHash(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	docID, tnt := testdb.InsertDraftDocument(t, db, "", snapshotTestTenantID)
	repo := infrastructure.NewSnapshotRepository(db)

	contentHash, err := hex.DecodeString("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	if err != nil {
		t.Fatalf("decode hash: %v", err)
	}
	s3Key := "final/doc.docx"

	// WriteFinalDocx updates public.documents — guarded by document.edit tripwire.
	// Pass the tx via the variadic DBTX parameter; cap set tx-locally via SeedWithCaps.
	testdb.SeedWithCaps(t, db, `[{"cap":"document.edit"}]`, func(tx *sql.Tx) error {
		return repo.WriteFinalDocx(ctx, tnt, docID, s3Key, contentHash, tx)
	})

	var gotKey string
	var gotHash []byte
	if err := db.QueryRowContext(ctx, `
		SELECT coalesce(final_docx_s3_key, ''), content_hash
		  FROM public.documents
		 WHERE tenant_id=$1::uuid AND id=$2::uuid`,
		tnt, docID,
	).Scan(&gotKey, &gotHash); err != nil {
		t.Fatalf("read final columns: %v", err)
	}
	if gotKey != s3Key {
		t.Fatalf("final_docx_s3_key mismatch: got %q want %q", gotKey, s3Key)
	}
	if hex.EncodeToString(gotHash) != hex.EncodeToString(contentHash) {
		t.Fatalf("content_hash mismatch: got %x want %x", gotHash, contentHash)
	}
}

func TestSnapshotRepository_ReadFinalDocxS3Key(t *testing.T) {
	if testing.Short() {
		t.Skip("requires postgres")
	}

	ctx := context.Background()
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	docID, tnt := testdb.InsertDraftDocument(t, db, "", snapshotTestTenantID)
	repo := infrastructure.NewSnapshotRepository(db)

	if _, err := repo.ReadFinalDocxS3Key(ctx, tnt, docID); err == nil {
		t.Fatal("ReadFinalDocxS3Key on unfrozen document: got nil error, want error")
	}

	contentHash, err := hex.DecodeString("dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd")
	if err != nil {
		t.Fatalf("decode hash: %v", err)
	}
	s3Key := "final/read-doc.docx"

	// WriteFinalDocx updates public.documents — guarded by document.edit tripwire.
	testdb.SeedWithCaps(t, db, `[{"cap":"document.edit"}]`, func(tx *sql.Tx) error {
		return repo.WriteFinalDocx(ctx, tnt, docID, s3Key, contentHash, tx)
	})

	got, err := repo.ReadFinalDocxS3Key(ctx, tnt, docID)
	if err != nil {
		t.Fatalf("ReadFinalDocxS3Key: %v", err)
	}
	if got != s3Key {
		t.Fatalf("ReadFinalDocxS3Key = %q, want %q", got, s3Key)
	}
}

func TestSnapshotRepository_WritePDF_PersistsAllColumns(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	docID, tnt := testdb.InsertDraftDocument(t, db, "", snapshotTestTenantID)
	repo := infrastructure.NewSnapshotRepository(db)

	pdfHash, err := hex.DecodeString("cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc")
	if err != nil {
		t.Fatalf("decode hash: %v", err)
	}
	pdfKey := "final/doc.pdf"
	generated := time.Date(2026, 4, 23, 19, 0, 0, 0, time.UTC)

	// WritePDF uses r.db.ExecContext directly (no DBTX variadic). Assert
	// document.edit session-locally via SetCapsOnDB — safe because this test DB
	// is isolated (cloned per-test, dropped after, MaxOpenConns=1).
	testdb.SetCapsOnDB(t, db, `[{"cap":"document.edit"}]`)
	if err := repo.WritePDF(ctx, tnt, docID, pdfKey, pdfHash, generated); err != nil {
		t.Fatalf("WritePDF: %v", err)
	}

	var gotKey string
	var gotHash []byte
	var gotAt *time.Time
	if err := db.QueryRowContext(ctx, `
		SELECT coalesce(final_pdf_s3_key,''), pdf_hash, pdf_generated_at
		  FROM public.documents
		 WHERE tenant_id=$1::uuid AND id=$2::uuid`,
		tnt, docID,
	).Scan(&gotKey, &gotHash, &gotAt); err != nil {
		t.Fatalf("read pdf columns: %v", err)
	}
	if gotKey != pdfKey {
		t.Fatalf("final_pdf_s3_key mismatch: got %q want %q", gotKey, pdfKey)
	}
	if hex.EncodeToString(gotHash) != hex.EncodeToString(pdfHash) {
		t.Fatalf("pdf_hash mismatch: got %x want %x", gotHash, pdfHash)
	}
	if gotAt == nil || !gotAt.Equal(generated) {
		t.Fatalf("pdf_generated_at mismatch: got %v want %v", gotAt, generated)
	}
}
