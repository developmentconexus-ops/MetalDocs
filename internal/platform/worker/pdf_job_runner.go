package worker

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

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
}

func NewPDFJobRunner(converter PDFConverter, persister PDFPersister) *PDFJobRunner {
	return &PDFJobRunner{
		converter: converter,
		persister: persister,
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
	docxKey := payload.FinalDocxS3Key
	if docxKey == "" {
		docxKey = fmt.Sprintf("tenants/%s/revisions/%s/frozen.docx", payload.TenantID, payload.RevisionID)
	}

	outputKey := fmt.Sprintf("tenants/%s/revisions/%s/final.pdf", payload.TenantID, payload.RevisionID)
	result, err := r.converter.ConvertPDF(ctx, servicebus.ConvertPDFRequest{
		DocxKey:   docxKey,
		OutputKey: outputKey,
	})
	if err != nil {
		return fmt.Errorf("pdf job runner: convert pdf: %w", err)
	}

	hashBytes, err := hex.DecodeString(result.ContentHash)
	if err != nil {
		return fmt.Errorf("pdf job runner: decode content hash: %w", err)
	}

	if err := r.persister.WritePDF(ctx, PDFWriteRequest{
		TenantID:    TenantID(payload.TenantID),
		DocumentID:  DocumentID(payload.RevisionID),
		StorageKey:  StorageKey(result.OutputKey),
		PDFHash:     hashBytes,
		GeneratedAt: time.Now().UTC(),
	}); err != nil {
		return fmt.Errorf("pdf job runner: persist pdf: %w", err)
	}
	return nil
}
