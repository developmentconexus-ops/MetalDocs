package worker

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/messaging"
	"metaldocs/internal/platform/servicebus"
)

type PDFConverter interface {
	ConvertPDF(ctx context.Context, req servicebus.ConvertPDFRequest) (servicebus.ConvertPDFResult, error)
}

type TenantID string
type DocumentID string
type StorageKey string

type PDFWriteRequest struct {
	TenantID    TenantID
	DocumentID  DocumentID
	StorageKey  StorageKey
	PDFHash     []byte
	GeneratedAt time.Time
}

type PDFPersister interface {
	WritePDF(ctx context.Context, req PDFWriteRequest) error
}

// PDFPersisterInTx is a tx-aware variant of PDFPersister: it runs the write
// inside the caller-supplied tx rather than opening its own (M3 F3.2 —
// validation-contract.md §2.2 site 2). When a PDFJobRunner is constructed
// with a *sql.DB (NewPDFJobRunner), the tx-wrapping path is used only if the
// persister also implements this interface; otherwise the runner falls back
// to the legacy untransacted PDFPersister.WritePDF for backward compatibility.
type PDFPersisterInTx interface {
	WritePDFInTx(ctx context.Context, tx db.Tx, req PDFWriteRequest) error
}

type StringPDFPersister interface {
	WritePDF(ctx context.Context, tenant, docID, s3Key string, pdfHash []byte, generatedAt time.Time) error
}

type SnapshotPDFPersister struct {
	persister StringPDFPersister
}

func NewSnapshotPDFPersister(persister StringPDFPersister) SnapshotPDFPersister {
	return SnapshotPDFPersister{persister: persister}
}

func (p SnapshotPDFPersister) WritePDF(ctx context.Context, req PDFWriteRequest) error {
	return p.persister.WritePDF(
		ctx,
		string(req.TenantID),
		string(req.DocumentID),
		string(req.StorageKey),
		req.PDFHash,
		req.GeneratedAt,
	)
}

type PDFJobRunner struct {
	converter PDFConverter
	persister PDFPersister
	db        *sql.DB
}

// NewPDFJobRunner constructs a PDFJobRunner using the legacy untransacted
// write path (persister.WritePDF runs directly against the pool, with no
// RLS tenant seed). Prefer NewPDFJobRunnerWithDB for production wiring — M3
// F3.2 requires the write to run inside a SeedTxTenant-seeded tx so the FORCE
// RLS backstop engages (validation-contract.md §2.2 site 2).
func NewPDFJobRunner(converter PDFConverter, persister PDFPersister) *PDFJobRunner {
	return &PDFJobRunner{
		converter: converter,
		persister: persister,
	}
}

// NewPDFJobRunnerWithDB constructs a PDFJobRunner that wraps the PDF write in
// a transaction seeded with the payload's tenant via authz.SeedTxTenant
// before the write (M3 F3.2 site 2). persister must additionally implement
// PDFPersisterInTx — the caller supplies an adapter bridging its repository's
// tx-scoped write method (e.g. *documents/repository.SnapshotRepository's
// WritePDF variadic DBTX param) to WritePDFInTx(ctx, tx, req).
func NewPDFJobRunnerWithDB(converter PDFConverter, persister PDFPersister, database *sql.DB) *PDFJobRunner {
	if database == nil {
		panic("pdf_job_runner: db is required for NewPDFJobRunnerWithDB")
	}
	return &PDFJobRunner{
		converter: converter,
		persister: persister,
		db:        database,
	}
}

func (r *PDFJobRunner) Handle(ctx context.Context, event messaging.Event) error {
	payload, err := messaging.PDFConvertPayloadFrom(event)
	if err != nil {
		return fmt.Errorf("pdf job runner: %w", err)
	}
	if payload.TenantID == "" || payload.RevisionID == "" {
		return fmt.Errorf("pdf job runner: missing payload fields")
	}
	// FinalDocxS3Key is required: it is the renderer-produced key (computed by
	// apps/docx-renderer/src/routes/fanout.ts:frozenDocxKey and persisted via the
	// producer), never a value the worker may guess. A missing key is a malformed
	// event and must fail rather than silently fall back to a hardcoded default.
	if payload.FinalDocxS3Key == "" {
		return fmt.Errorf("pdf job runner: missing final_docx_s3_key")
	}

	outputKey := workerPDFKey(payload.TenantID, payload.RevisionID)
	result, err := r.converter.ConvertPDF(ctx, servicebus.ConvertPDFRequest{
		DocxKey:   payload.FinalDocxS3Key,
		OutputKey: outputKey,
	})
	if err != nil {
		return fmt.Errorf("pdf job runner: convert pdf: %w", err)
	}

	hashBytes, err := hex.DecodeString(result.ContentHash)
	if err != nil {
		return fmt.Errorf("pdf job runner: decode content hash: %w", err)
	}

	req := PDFWriteRequest{
		TenantID:    TenantID(payload.TenantID),
		DocumentID:  DocumentID(payload.RevisionID),
		StorageKey:  StorageKey(result.OutputKey),
		PDFHash:     hashBytes,
		GeneratedAt: time.Now().UTC(),
	}

	if r.db == nil {
		// Legacy untransacted path (NewPDFJobRunner) — no RLS tenant seed.
		if err := r.persister.WritePDF(ctx, req); err != nil {
			return fmt.Errorf("pdf job runner: persist pdf: %w", err)
		}
		return nil
	}

	inTx, ok := r.persister.(PDFPersisterInTx)
	if !ok {
		return fmt.Errorf("pdf job runner: persister %T does not implement PDFPersisterInTx (required when constructed via NewPDFJobRunnerWithDB)", r.persister)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("pdf job runner: begin tx: %w", err)
	}
	// M3 F3.2 (validation-contract.md §2.2 site 2) — seed the tenant-only RLS
	// backstop GUC before the write in this single-tenant processing tx.
	if err := authz.SeedTxTenant(ctx, tx, payload.TenantID); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("pdf job runner: seed tenant: %w", err)
	}
	if err := inTx.WritePDFInTx(ctx, tx, req); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("pdf job runner: persist pdf: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("pdf job runner: commit: %w", err)
	}
	return nil
}
