//go:build integration
// +build integration

package worker

// Real-DB tripwire proof for the PDFJobRunner write path (F9-2, sibling to the
// materialize consumer fix in materialize_job_runner.go). Handle opens a tx,
// seeds only the tenant GUC via authz.SeedTxTenant, then writes
// documents.final_pdf_s3_key — an UPDATE guarded by trg_require_cap_asserted
// (document.edit). Without a capability bypass this UPDATE fails closed with
// ErrCapabilityNotAsserted (P0001); this is the same class of defect as F9-1.
//
// Coverage:
//  1. TestPDFJobRunner_Integration_Handle_PersistsFinalPDFKey — proves the
//     fixed Handle (background bypass + tx-local BypassSystem) writes
//     documents.final_pdf_s3_key end to end against a real testdb.
//  2. TestPDFJobRunner_Integration_SeedTenantOnly_FailsClosed — pins the
//     defect: seeding ONLY the tenant GUC (no BypassSystem) on the same
//     guarded UPDATE must still fail with ErrCapabilityNotAsserted, proving
//     the tripwire itself stays armed and the fix is additive, not a
//     trigger/SQL change.

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"metaldocs/internal/modules/documents/infrastructure"
	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/messaging"
	"metaldocs/internal/platform/servicebus"
	"metaldocs/internal/platform/tenant"
	"metaldocs/tests/integration/testdb"
)

const pdfJobRunnerTestTenantID = tenant.DevTenantID

// snapshotPDFTxAdapterForTest mirrors apps/worker/cmd/metaldocs-worker/main.go's
// unexported snapshotPDFTxAdapter, bridging infrastructure.SnapshotRepository
// to both PDFPersister and PDFPersisterInTx so it can be passed to
// NewPDFJobRunnerWithDB exactly as production wiring does.
type snapshotPDFTxAdapterForTest struct {
	repo *infrastructure.SnapshotRepository
}

func (a snapshotPDFTxAdapterForTest) WritePDFInTx(ctx context.Context, tx db.Tx, req PDFWriteRequest) error {
	return a.repo.WritePDF(
		ctx,
		string(req.TenantID),
		string(req.DocumentID),
		string(req.StorageKey),
		req.PDFHash,
		req.GeneratedAt,
		tx,
	)
}

func (a snapshotPDFTxAdapterForTest) WritePDF(ctx context.Context, req PDFWriteRequest) error {
	return a.repo.WritePDF(
		ctx,
		string(req.TenantID),
		string(req.DocumentID),
		string(req.StorageKey),
		req.PDFHash,
		req.GeneratedAt,
	)
}

// fixedPDFConverter is a real-DB-test double for PDFConverter: no external PDF
// rendering happens in these tests, only the DB write path is under test.
type fixedPDFConverter struct {
	outputKey   string
	contentHash string
}

func (c fixedPDFConverter) ConvertPDF(_ context.Context, req servicebus.ConvertPDFRequest) (servicebus.ConvertPDFResult, error) {
	return servicebus.ConvertPDFResult{
		OutputKey:   c.outputKey,
		ContentHash: c.contentHash,
	}, nil
}

// testAuthzSeam mirrors apps/worker/cmd/metaldocs-worker/main.go's
// authzSeamAdapter, bridging the real iam/authz package to TxAuthzSeam for
// this integration test's system-under-test.
type testAuthzSeam struct{}

func (testAuthzSeam) WithBackgroundBypass(ctx context.Context) context.Context {
	return authz.WithBackgroundBypass(ctx)
}

func (testAuthzSeam) SeedTxTenant(ctx context.Context, tx *sql.Tx, tenantID string) error {
	return authz.SeedTxTenant(ctx, tx, tenantID)
}

func (testAuthzSeam) BypassSystem(ctx context.Context, tx *sql.Tx) error {
	return authz.BypassSystem(ctx, tx)
}

func TestPDFJobRunner_Integration_Handle_PersistsFinalPDFKey(t *testing.T) {
	ctx := context.Background()
	sqlDB, _ := testdb.Open(t)
	sqlDB.SetMaxOpenConns(1)

	docID, tnt := testdb.InsertDraftDocument(t, sqlDB, "", pdfJobRunnerTestTenantID)

	repo := infrastructure.NewSnapshotRepository(sqlDB)
	adapter := snapshotPDFTxAdapterForTest{repo: repo}
	converter := fixedPDFConverter{
		outputKey:   "tenants/" + tnt + "/revisions/" + docID + "/final.pdf",
		contentHash: "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
	}

	runner := NewPDFJobRunnerWithDB(converter, adapter, testAuthzSeam{}, sqlDB)

	event := messaging.Event{
		EventID:   "event-1",
		EventType: messaging.EventTypePDFConvert,
		Payload: messaging.PDFConvertPayload{
			TenantID:       tnt,
			RevisionID:     docID,
			FinalDocxS3Key: "tenants/" + tnt + "/revisions/" + docID + "/final.docx",
		},
	}

	if err := runner.Handle(ctx, event); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	var gotKey string
	if err := sqlDB.QueryRowContext(ctx, `
		SELECT coalesce(final_pdf_s3_key, '')
		  FROM public.documents
		 WHERE tenant_id=$1::uuid AND id=$2::uuid`,
		tnt, docID,
	).Scan(&gotKey); err != nil {
		t.Fatalf("read final_pdf_s3_key: %v", err)
	}
	if gotKey != converter.outputKey {
		t.Fatalf("final_pdf_s3_key = %q, want %q", gotKey, converter.outputKey)
	}
}

// TestPDFJobRunner_Integration_SeedTenantOnly_FailsClosed pins the pre-fix
// defect shape: seeding ONLY the tenant GUC (authz.SeedTxTenant, no
// authz.BypassSystem) on the same guarded UPDATE that Handle performs must
// still fail closed with ErrCapabilityNotAsserted (P0001). This proves the
// tripwire remains armed independent of the Handle fix above.
func TestPDFJobRunner_Integration_SeedTenantOnly_FailsClosed(t *testing.T) {
	ctx := context.Background()
	sqlDB, _ := testdb.Open(t)
	sqlDB.SetMaxOpenConns(1)

	docID, tnt := testdb.InsertDraftDocument(t, sqlDB, "", pdfJobRunnerTestTenantID)
	repo := infrastructure.NewSnapshotRepository(sqlDB)

	tx, err := sqlDB.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	if err := authz.SeedTxTenant(ctx, tx, tnt); err != nil {
		t.Fatalf("seed tenant: %v", err)
	}

	err = repo.WritePDF(ctx, tnt, docID, "tenants/x/final.pdf",
		[]byte{0xAA, 0xBB}, time.Now().UTC(), tx)
	if err == nil {
		t.Fatal("WritePDF with tenant-only seed: got nil error, want ErrCapabilityNotAsserted")
	}
	if !strings.Contains(err.Error(), "ErrCapabilityNotAsserted") {
		t.Fatalf("WritePDF error = %v, want ErrCapabilityNotAsserted", err)
	}
}
