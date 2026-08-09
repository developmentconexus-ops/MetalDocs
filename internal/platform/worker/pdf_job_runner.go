package worker

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/messaging"
	"metaldocs/internal/platform/servicebus"
)

// PDFConverter is the minimal published interface for converting a docx to
// PDF. Satisfied by *servicebus.GotenbergPDFClient.
type PDFConverter interface {
	ConvertPDF(ctx context.Context, req servicebus.ConvertPDFRequest) (servicebus.ConvertPDFResult, error)
}

// TenantID identifies the tenant a PDF write belongs to.
type TenantID string

// DocumentID identifies the document/revision a PDF write belongs to.
type DocumentID string

// StorageKey is the object-storage key a PDF was written to.
type StorageKey string

// PDFWriteRequest carries the fields needed to persist a generated PDF.
type PDFWriteRequest struct {
	TenantID    TenantID
	DocumentID  DocumentID
	StorageKey  StorageKey
	PDFHash     []byte
	GeneratedAt time.Time
}

// PDFPersister persists a generated PDF outside any transaction.
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

// StringPDFPersister is a PDF persister whose WritePDF takes plain string
// identifiers rather than the typed TenantID/DocumentID/StorageKey wrappers.
type StringPDFPersister interface {
	WritePDF(ctx context.Context, tenant, docID, s3Key string, pdfHash []byte, generatedAt time.Time) error
}

// SnapshotPDFPersister adapts a StringPDFPersister to the PDFPersister
// interface by stringifying the typed PDFWriteRequest fields.
type SnapshotPDFPersister struct {
	persister StringPDFPersister
}

// NewSnapshotPDFPersister constructs a SnapshotPDFPersister wrapping persister.
func NewSnapshotPDFPersister(persister StringPDFPersister) SnapshotPDFPersister {
	return SnapshotPDFPersister{persister: persister}
}

// WritePDF stringifies req's typed fields and delegates to the wrapped StringPDFPersister.
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

// PDFJobRunner handles EventTypePDFConvert events: converts the frozen docx
// to PDF and persists the result, optionally inside a tenant-seeded
// transaction alongside the ADR 0085 artifact fact.
type PDFJobRunner struct {
	converter    PDFConverter
	persister    PDFPersister
	artifactFact ArtifactFactRecorder
	authzSeam    TxAuthzSeam
	db           *sql.DB
}

// WithArtifactFactRecorder wires the ADR 0085 artifact-fact port. The PDF is
// the SECOND half of the artifact set, so this is the write that normally
// completes artifact-ready(G) and releases the document.
func (r *PDFJobRunner) WithArtifactFactRecorder(rec ArtifactFactRecorder) *PDFJobRunner {
	r.artifactFact = rec
	return r
}

// NewPDFJobRunner constructs a PDFJobRunner using the legacy untransacted
// write path (persister.WritePDF runs directly against the pool, with no
// RLS tenant seed). Prefer NewPDFJobRunnerWithDB for production wiring — M3
// F3.2 requires the write to run inside a SeedTxTenant-seeded tx so the FORCE
// RLS backstop engages (validation-contract.md §2.2 site 2).
func NewPDFJobRunner(converter PDFConverter, persister PDFPersister, authzSeam TxAuthzSeam) *PDFJobRunner {
	if authzSeam == nil {
		panic("pdf_job_runner: authzSeam is required")
	}
	return &PDFJobRunner{
		converter: converter,
		persister: persister,
		authzSeam: authzSeam,
	}
}

// NewPDFJobRunnerWithDB constructs a PDFJobRunner that wraps the PDF write in
// a transaction seeded with the payload's tenant via authz.SeedTxTenant
// before the write (M3 F3.2 site 2). persister must additionally implement
// PDFPersisterInTx — the caller supplies an adapter bridging its repository's
// tx-scoped write method (e.g. *documents/repository.SnapshotRepository's
// WritePDF variadic DBTX param) to WritePDFInTx(ctx, tx, req).
func NewPDFJobRunnerWithDB(converter PDFConverter, persister PDFPersister, authzSeam TxAuthzSeam, database *sql.DB) *PDFJobRunner {
	if database == nil {
		panic("pdf_job_runner: db is required for NewPDFJobRunnerWithDB")
	}
	if authzSeam == nil {
		panic("pdf_job_runner: authzSeam is required")
	}
	return &PDFJobRunner{
		converter: converter,
		persister: persister,
		authzSeam: authzSeam,
		db:        database,
	}
}

// Handle processes an EventTypePDFConvert event: converts the frozen docx to
// PDF and persists the result, transactionally when the runner was
// constructed via NewPDFJobRunnerWithDB.
func (r *PDFJobRunner) Handle(ctx context.Context, event messaging.Event) error {
	if r.authzSeam == nil {
		return fmt.Errorf("pdf job runner: authz seam not configured")
	}
	ctx = r.authzSeam.WithBackgroundBypass(ctx)
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
		return r.persistLegacy(ctx, payload, req)
	}
	return r.persistInTx(ctx, payload, req)
}

// persistLegacy runs the untransacted write path (NewPDFJobRunner) — no RLS
// tenant seed. A release generation cannot be honoured here: its fact must
// commit with the artifact write, and there is no transaction to join. Fail
// closed rather than silently dropping the fact.
func (r *PDFJobRunner) persistLegacy(ctx context.Context, payload messaging.PDFConvertPayload, req PDFWriteRequest) error {
	if payload.ReleaseGenerationID != "" {
		return fmt.Errorf("pdf job runner: release generation %s requires the transacted path (NewPDFJobRunnerWithDB)", payload.ReleaseGenerationID)
	}
	if err := r.persister.WritePDF(ctx, req); err != nil {
		return fmt.Errorf("pdf job runner: persist pdf: %w", err)
	}
	return nil
}

// persistInTx runs the transacted write path (NewPDFJobRunnerWithDB): the
// PDF write and, when a release generation is present, the ADR 0085 artifact
// fact commit together in one tenant-seeded transaction (M3 F3.2 —
// validation-contract.md §2.2 site 2).
func (r *PDFJobRunner) persistInTx(ctx context.Context, payload messaging.PDFConvertPayload, req PDFWriteRequest) error {
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
	if err := r.authzSeam.SeedTxTenant(ctx, tx, payload.TenantID); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("pdf job runner: seed tenant: %w", err)
	}
	if err := r.authzSeam.BypassSystem(ctx, tx); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("pdf job runner: bypass system: %w", err)
	}
	// ADR 0085: artifact fact in the same transaction as the artifact write,
	// and the generation guard that refuses a superseded generation's PDF.
	//
	// It runs BEFORE the documents-side write, and that order is load-bearing:
	// the recorder locks release_generations, the write locks documents, and
	// the release coordinator locks generation → documents. Taking documents
	// first here would be the reverse order and deadlock against a concurrent
	// release. The output key is already known, so nothing is lost by going
	// first.
	if payload.ReleaseGenerationID != "" {
		if r.artifactFact == nil {
			_ = tx.Rollback()
			return fmt.Errorf("pdf job runner: artifact fact recorder not configured (release generation %s)", payload.ReleaseGenerationID)
		}
		if err := r.artifactFact.RecordArtifactFactTx(ctx, tx, payload.TenantID, payload.ReleaseGenerationID, ArtifactKindFinalPDF, string(req.StorageKey)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("pdf job runner: record final pdf fact: %w", err)
		}
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
