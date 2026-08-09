package application

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/platform/servicebus"
)

// ExportRepo is the persistence port ExportService uses to read documents,
// revisions, and export dedup rows, and to record new exports.
type ExportRepo interface {
	GetDocument(ctx context.Context, tenantID, id string) (*domain.Document, error)
	GetRevision(ctx context.Context, tenantID, docID, revID string) (*domain.Revision, error)
	InsertExport(ctx context.Context, e *domain.Export) (*domain.Export, error)
	GetExportByHash(ctx context.Context, tenantID, documentID string, compositeHash []byte) (*domain.Export, error)
}

const exportDownloadTTL = 15 * time.Minute

// ExportPresigner presigns and inspects export objects in the object store.
type ExportPresigner interface {
	PresignGet(ctx context.Context, key string, ttl time.Duration) (url string, err error)
	Exists(ctx context.Context, key string) (bool, error)
	Size(ctx context.Context, key string) (int64, error)
}

// DocgenPDFClient converts a stored docx revision into a PDF via the
// docx-renderer/Gotenberg pipeline.
type DocgenPDFClient interface {
	ConvertPDF(ctx context.Context, req servicebus.ConvertPDFRequest) (servicebus.ConvertPDFResult, error)
}

// ExportService produces and serves PDF/docx exports of a document's current
// revision, deduplicating by composite content hash.
type ExportService struct {
	repo       ExportRepo
	presigner  ExportPresigner
	docgen     DocgenPDFClient
	audit      Audit
	docgenVer  string
	grammarVer string
}

// NewExportService constructs an ExportService wired to its repo, presigner,
// docgen client, and audit sink, tagging generated exports with the given
// docgen/grammar version strings.
func NewExportService(repo ExportRepo, presigner ExportPresigner, docgen DocgenPDFClient, audit Audit, docgenVer, grammarVer string) *ExportService {
	return &ExportService{
		repo:       repo,
		presigner:  presigner,
		docgen:     docgen,
		audit:      audit,
		docgenVer:  docgenVer,
		grammarVer: grammarVer,
	}
}

// ExportPDF renders (or returns a cached) PDF export of documentID's current
// revision under opts, keyed by a composite hash of content + template
// version + render options so identical requests reuse the same artifact.
func (s *ExportService) ExportPDF(ctx context.Context, tenantID, userID, documentID string, opts domain.RenderOptions) (*domain.ExportResult, error) {
	doc, err := s.repo.GetDocument(ctx, tenantID, documentID)
	if err != nil {
		return nil, err
	}
	if doc.CurrentRevisionID == "" {
		return nil, domain.ErrExportDocxMissing
	}

	rev, err := s.repo.GetRevision(ctx, tenantID, documentID, doc.CurrentRevisionID)
	if err != nil {
		return nil, err
	}

	contentHashBytes, err := hex.DecodeString(rev.ContentHash)
	if err != nil {
		return nil, fmt.Errorf("decode revision content hash: %w", err)
	}

	compositeHash, err := domain.ComputeCompositeHash(contentHashBytes, doc.TemplateVersionID, s.grammarVer, s.docgenVer, opts)
	if err != nil {
		return nil, err
	}

	storageKey := documentExportKey(tenantID, documentID, compositeHash)

	existing, err := s.repo.GetExportByHash(ctx, tenantID, documentID, compositeHash)
	if err == nil {
		s.audit.Write(ctx, tenantID, userID, "export.pdf_generated", documentID, map[string]any{"cached": true, "storage_key": existing.StorageKey})
		return &domain.ExportResult{Export: existing, Cached: true}, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}

	headFound, err := s.presigner.Exists(ctx, storageKey)
	if err != nil {
		return nil, err
	}
	if !headFound {
		_, err = s.docgen.ConvertPDF(ctx, servicebus.ConvertPDFRequest{
			DocxKey:   rev.StorageKey,
			OutputKey: storageKey,
			RenderOpts: &servicebus.PDFRenderOpts{
				PaperSize: servicebus.PaperSize(opts.PaperSize),
				Landscape: opts.LandscapeP,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("%w: %w", domain.ErrExportGotenbergFailed, err)
		}
	}

	sizeBytes, err := s.presigner.Size(ctx, storageKey)
	if err != nil {
		return nil, err
	}

	exportRow, err := domain.NewExport(tenantID, documentID, rev.ID, compositeHash, storageKey, sizeBytes, opts.PaperSize, opts.LandscapeP, s.docgenVer)
	if err != nil {
		return nil, err
	}

	exp, err := s.repo.InsertExport(ctx, exportRow)
	if err != nil {
		return nil, err
	}

	s.audit.Write(ctx, tenantID, userID, "export.pdf_generated", documentID, map[string]any{"cached": false, "storage_key": exp.StorageKey})
	return &domain.ExportResult{Export: exp, Cached: false}, nil
}

// SignExportURL presigns a short-lived GET URL for a previously generated
// export object at storageKey.
func (s *ExportService) SignExportURL(ctx context.Context, storageKey string) (string, error) {
	return s.presigner.PresignGet(ctx, storageKey, exportDownloadTTL)
}

// GetDocumentSummary returns the document row for documentID, used by export
// handlers to surface document metadata alongside the export.
func (s *ExportService) GetDocumentSummary(ctx context.Context, tenantID, documentID string) (*domain.Document, error) {
	return s.repo.GetDocument(ctx, tenantID, documentID)
}

// SignedDocxURL presigns a short-lived GET URL for documentID's current docx
// revision and records an export.docx_downloaded audit entry.
func (s *ExportService) SignedDocxURL(ctx context.Context, tenantID, userID, documentID string) (string, error) {
	doc, err := s.repo.GetDocument(ctx, tenantID, documentID)
	if err != nil {
		return "", err
	}
	if doc.CurrentRevisionID == "" {
		return "", domain.ErrExportDocxMissing
	}

	rev, err := s.repo.GetRevision(ctx, tenantID, documentID, doc.CurrentRevisionID)
	if err != nil {
		return "", err
	}

	url, err := s.presigner.PresignGet(ctx, rev.StorageKey, exportDownloadTTL)
	if err != nil {
		return "", err
	}

	s.audit.Write(ctx, tenantID, userID, "export.docx_downloaded", documentID, map[string]any{"revision_id": rev.ID, "storage_key": rev.StorageKey})
	return url, nil
}
