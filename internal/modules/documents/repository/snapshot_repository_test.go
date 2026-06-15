//go:build integration
// +build integration

package repository_test

import (
	"context"
	"encoding/hex"
	"testing"
	"time"

	"metaldocs/internal/modules/documents/repository"
	"metaldocs/internal/platform/tenant"
	"metaldocs/tests/integration/testdb"
)

const snapshotTestTenantID = tenant.DevTenantID

func TestSnapshotRepository_WriteFreeze_PersistsHashAndFrozenAt(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	if _, err := db.ExecContext(ctx, `SET search_path TO metaldocs, public`); err != nil {
		t.Fatalf("set search_path: %v", err)
	}

	tenantObj := testdb.NewTenant(t, db, testdb.WithTenant(snapshotTestTenantID))
	doc := testdb.NewDocument(t, db, testdb.WithTenant(tenantObj.ID), testdb.WithStatus("draft"))
	repo := repository.NewSnapshotRepository(db)

	hash, err := hex.DecodeString("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err != nil {
		t.Fatalf("decode hash: %v", err)
	}
	frozenAt := time.Date(2026, 4, 23, 18, 0, 0, 0, time.UTC)

	if err := repo.WriteFreeze(ctx, tenantObj.ID, doc.ID, hash, frozenAt); err != nil {
		t.Fatalf("WriteFreeze: %v", err)
	}

	var gotHash []byte
	var gotFrozenAt *time.Time
	if err := db.QueryRowContext(ctx, `
		SELECT values_hash, values_frozen_at
		  FROM public.documents
		 WHERE tenant_id=$1::uuid AND id=$2::uuid`,
		tenantObj.ID, doc.ID,
	).Scan(&gotHash, &gotFrozenAt); err != nil {
		t.Fatalf("read freeze columns: %v", err)
	}
	if hex.EncodeToString(gotHash) != hex.EncodeToString(hash) {
		t.Fatalf("values_hash mismatch: got %x want %x", gotHash, hash)
	}
	if gotFrozenAt == nil || !gotFrozenAt.Equal(frozenAt) {
		t.Fatalf("values_frozen_at mismatch: got %v want %v", gotFrozenAt, frozenAt)
	}
}

func TestSnapshotRepository_WriteFinalDocx_PersistsKeyAndContentHash(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)
	db.SetMaxOpenConns(1)

	if _, err := db.ExecContext(ctx, `SET search_path TO metaldocs, public`); err != nil {
		t.Fatalf("set search_path: %v", err)
	}

	tenantObj := testdb.NewTenant(t, db, testdb.WithTenant(snapshotTestTenantID))
	doc := testdb.NewDocument(t, db, testdb.WithTenant(tenantObj.ID), testdb.WithStatus("draft"))
	repo := repository.NewSnapshotRepository(db)

	contentHash, err := hex.DecodeString("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	if err != nil {
		t.Fatalf("decode hash: %v", err)
	}
	s3Key := "final/doc.docx"

	if err := repo.WriteFinalDocx(ctx, tenantObj.ID, doc.ID, s3Key, contentHash); err != nil {
		t.Fatalf("WriteFinalDocx: %v", err)
	}

	var gotKey string
	var gotHash []byte
	if err := db.QueryRowContext(ctx, `
		SELECT coalesce(final_docx_s3_key, ''), content_hash
		  FROM public.documents
		 WHERE tenant_id=$1::uuid AND id=$2::uuid`,
		tenantObj.ID, doc.ID,
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

	if _, err := db.ExecContext(ctx, `SET search_path TO metaldocs, public`); err != nil {
		t.Fatalf("set search_path: %v", err)
	}

	tenantObj := testdb.NewTenant(t, db, testdb.WithTenant(snapshotTestTenantID))
	doc := testdb.NewDocument(t, db, testdb.WithTenant(tenantObj.ID), testdb.WithStatus("draft"))
	repo := repository.NewSnapshotRepository(db)

	if _, err := repo.ReadFinalDocxS3Key(ctx, tenantObj.ID, doc.ID); err == nil {
		t.Fatal("ReadFinalDocxS3Key on unfrozen document: got nil error, want error")
	}

	contentHash, err := hex.DecodeString("dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd")
	if err != nil {
		t.Fatalf("decode hash: %v", err)
	}
	s3Key := "final/read-doc.docx"

	if err := repo.WriteFinalDocx(ctx, tenantObj.ID, doc.ID, s3Key, contentHash); err != nil {
		t.Fatalf("WriteFinalDocx: %v", err)
	}

	got, err := repo.ReadFinalDocxS3Key(ctx, tenantObj.ID, doc.ID)
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

	if _, err := db.ExecContext(ctx, `SET search_path TO metaldocs, public`); err != nil {
		t.Fatalf("set search_path: %v", err)
	}

	tenantObj := testdb.NewTenant(t, db, testdb.WithTenant(snapshotTestTenantID))
	doc := testdb.NewDocument(t, db, testdb.WithTenant(tenantObj.ID), testdb.WithStatus("draft"))
	repo := repository.NewSnapshotRepository(db)

	pdfHash, err := hex.DecodeString("cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc")
	if err != nil {
		t.Fatalf("decode hash: %v", err)
	}
	pdfKey := "final/doc.pdf"
	generated := time.Date(2026, 4, 23, 19, 0, 0, 0, time.UTC)

	if err := repo.WritePDF(ctx, tenantObj.ID, doc.ID, pdfKey, pdfHash, generated); err != nil {
		t.Fatalf("WritePDF: %v", err)
	}

	var gotKey string
	var gotHash []byte
	var gotAt *time.Time
	if err := db.QueryRowContext(ctx, `
		SELECT coalesce(final_pdf_s3_key,''), pdf_hash, pdf_generated_at
		  FROM public.documents
		 WHERE tenant_id=$1::uuid AND id=$2::uuid`,
		tenantObj.ID, doc.ID,
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
