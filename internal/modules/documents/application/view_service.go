package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	v2dom "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
)

const viewDownloadTTL = 15 * time.Minute

// ViewPresigner is implemented by objectstore helpers that presign a GET URL.
// AssertedPresignGet is the tenant-guarded variant (F-O3): it asserts the key
// prefix before issuing the URL, providing defense-in-depth on the view read path.
type ViewPresigner interface {
	AssertedPresignGet(ctx context.Context, tenantID, key string, ttl time.Duration) (string, error)
}

// PDFOutboxStateReader returns the latest pdf_outbox state for a revision/document.
// Returns "" + nil when no row exists.
type PDFOutboxStateReader interface {
	ReadState(ctx context.Context, tenantID, revisionID string) (string, error)
}

// ViewService serves viewer requests by checking area-scoped RBAC, validating
// the revision's lifecycle state, and returning a presigned PDF URL.
type ViewService struct {
	runner    db.TxRunner
	presigner ViewPresigner
	outbox    PDFOutboxStateReader // optional; nil → assume pending
}

// NewViewService constructs a ViewService wired to its tx runner, presigner,
// and (optional) pdf_outbox state reader.
func NewViewService(runner db.TxRunner, presigner ViewPresigner, outbox PDFOutboxStateReader) *ViewService {
	return &ViewService{runner: runner, presigner: presigner, outbox: outbox}
}

var viewableStatuses = map[string]struct{}{
	string(v2dom.DocStatusApproved):  {},
	string(v2dom.DocStatusScheduled): {},
	string(v2dom.DocStatusPublished): {},
}

// GetViewURL authorizes the actor for viewing and returns a short-lived
// presigned URL for the document's current viewable artifact.
func (s *ViewService) GetViewURL(ctx context.Context, tenantID, actorID, docID string) (ViewResult, error) {
	ctx = authz.WithCapCache(ctx)
	var result ViewResult
	// Read-write tx: authz.Require may audit a system_admin bypass (ADR 0022 F8),
	// which INSERTs — see the Require INVARIANT note in iam/authz/authz.go.
	if err := s.runner.Do(ctx, func(tx *sql.Tx) error {
		res, err := s.resolveViewURL(ctx, tx, tenantID, docID)
		if err != nil {
			return err
		}
		result = res
		return nil
	}); err != nil {
		return ViewResult{}, err
	}
	return result, nil
}

// resolveViewURL is GetViewURL's in-tx body: it loads the document's status +
// pdf key, enforces document.view authz, rejects non-viewable lifecycle
// states, and returns either a presigned URL (final PDF already produced) or
// the outbox-derived pending/failed status.
func (s *ViewService) resolveViewURL(ctx context.Context, tx *sql.Tx, tenantID, docID string) (ViewResult, error) {
	var status string
	var pdfKey sql.NullString
	err := tx.QueryRowContext(ctx, `
		SELECT status, final_pdf_s3_key
		  FROM documents
		 WHERE tenant_id=$1::uuid AND id=$2::uuid`,
		tenantID, docID,
	).Scan(&status, &pdfKey)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ViewResult{}, v2dom.ErrNotFound
		}
		return ViewResult{}, fmt.Errorf("view: load document: %w", err)
	}

	// document.view is tenant-grade (a *.view read) — pass the "tenant" sentinel
	// so the area filter is intentionally OFF (ADR 0022 Phase 8). ADR 0022 Phase 10
	// (F2): the redundant doc.view_published cap was merged into the canonical
	// CapDocumentView — identical grant set, same tenant-grade read.
	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentView), "tenant"); err != nil {
		return ViewResult{}, err
	}

	if _, ok := viewableStatuses[status]; !ok {
		return ViewResult{}, v2dom.ErrNotFound
	}

	if pdfKey.Valid && pdfKey.String != "" {
		url, err := s.presigner.AssertedPresignGet(ctx, tenantID, pdfKey.String, viewDownloadTTL)
		if err != nil {
			return ViewResult{}, fmt.Errorf("view: presign: %w", err)
		}
		return ViewResult{PDFStatus: "ready", SignedURL: url}, nil
	}

	pdfStatus := "pending"
	if s.outbox != nil {
		if state, err := s.outbox.ReadState(ctx, tenantID, docID); err == nil && state == "failed" {
			pdfStatus = "failed"
		}
	}
	return ViewResult{PDFStatus: pdfStatus}, nil
}
