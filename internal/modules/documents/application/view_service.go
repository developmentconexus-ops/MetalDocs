package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	v2dom "metaldocs/internal/modules/documents/domain"
	documentshttp "metaldocs/internal/modules/documents/http"
	"metaldocs/internal/modules/iam/authz"
)

// ViewPresigner is implemented by objectstore helpers that presign a GET URL.
type ViewPresigner interface {
	PresignObjectGET(ctx context.Context, storageKey string) (string, error)
}

// PDFOutboxStateReader returns the latest pdf_outbox state for a revision/document.
// Returns "" + nil when no row exists.
type PDFOutboxStateReader interface {
	ReadState(ctx context.Context, tenantID, revisionID string) (string, error)
}

// ViewService serves viewer requests by checking area-scoped RBAC, validating
// the revision's lifecycle state, and returning a presigned PDF URL.
type ViewService struct {
	db        *sql.DB
	presigner ViewPresigner
	outbox    PDFOutboxStateReader // optional; nil → assume pending
}

func NewViewService(db *sql.DB, presigner ViewPresigner, outbox PDFOutboxStateReader) *ViewService {
	return &ViewService{db: db, presigner: presigner, outbox: outbox}
}

var viewableStatuses = map[string]struct{}{
	"approved":  {},
	"scheduled": {},
	"published": {},
}

func (s *ViewService) GetViewURL(ctx context.Context, tenantID, actorID, docID string) (documentshttp.ViewResult, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return documentshttp.ViewResult{}, fmt.Errorf("view: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	ctx = authz.WithCapCache(ctx)
	if err := authz.SeedTxIdentity(ctx, tx, tenantID, actorID); err != nil {
		return documentshttp.ViewResult{}, err
	}

	var status, areaCode string
	var pdfKey sql.NullString
	err = tx.QueryRowContext(ctx, `
		SELECT status, coalesce(process_area_code_snapshot,''), final_pdf_s3_key
		  FROM documents
		 WHERE tenant_id=$1::uuid AND id=$2::uuid`,
		tenantID, docID,
	).Scan(&status, &areaCode, &pdfKey)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return documentshttp.ViewResult{}, v2dom.ErrNotFound
		}
		return documentshttp.ViewResult{}, fmt.Errorf("view: load document: %w", err)
	}

	area := areaCode
	if area == "" {
		area = "tenant"
	}
	if err := authz.Require(ctx, tx, "doc.view_published", area); err != nil {
		return documentshttp.ViewResult{}, err
	}

	if _, ok := viewableStatuses[status]; !ok {
		return documentshttp.ViewResult{}, v2dom.ErrNotFound
	}

	if pdfKey.Valid && pdfKey.String != "" {
		url, err := s.presigner.PresignObjectGET(ctx, pdfKey.String)
		if err != nil {
			return documentshttp.ViewResult{}, fmt.Errorf("view: presign: %w", err)
		}
		return documentshttp.ViewResult{PDFStatus: "ready", SignedURL: url}, nil
	}

	pdfStatus := "pending"
	if s.outbox != nil {
		if state, err := s.outbox.ReadState(ctx, tenantID, docID); err == nil && state == "failed" {
			pdfStatus = "failed"
		}
	}
	return documentshttp.ViewResult{PDFStatus: pdfStatus}, nil
}
