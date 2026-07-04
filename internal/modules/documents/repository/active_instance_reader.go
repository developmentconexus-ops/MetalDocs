package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	approvaldomain "metaldocs/internal/modules/documents/approval/domain"
	documentsdomain "metaldocs/internal/modules/documents/domain"
)

// ActiveInstanceReaderPG is the documents-owned Postgres adapter for the
// active-instance projection read-port (ADR-0039 D3(b)). It runs the same
// FULL OUTER JOIN projection over documents/document_revisions plus the derived
// in-progress approval_instances lookup that controlleddocuments' GetActiveInstance
// historically ran inline — all three tables are documents-owned (the approval
// sub-context maps to the documents top-level module), so this read is
// intra-module. Status filters use the owner's typed constants as query
// parameters, never bare literals.
type ActiveInstanceReaderPG struct {
	db *sql.DB
}

// NewActiveInstanceReaderPG builds the adapter over the documents pool. Wired at
// the composition root and injected into the controlleddocuments repository.
func NewActiveInstanceReaderPG(db *sql.DB) *ActiveInstanceReaderPG {
	return &ActiveInstanceReaderPG{db: db}
}

var _ documentsdomain.ActiveInstanceReader = (*ActiveInstanceReaderPG)(nil)

// activeInstanceStatuses is the active (non-published) status set, expressed in
// the owner's vocabulary (documents/domain.DocStatus*). Passed as $3..$6.
var activeInstanceStatuses = []any{
	string(documentsdomain.DocStatusDraft),
	string(documentsdomain.DocStatusUnderReview),
	string(documentsdomain.DocStatusApproved),
	string(documentsdomain.DocStatusScheduled),
}

func (r *ActiveInstanceReaderPG) ActiveInstanceForControlledDocument(ctx context.Context, tenantID, controlledDocumentID string) (*documentsdomain.ActiveInstanceView, error) {
	var (
		docID          sql.NullString
		contentHash    sql.NullString
		revisionVer    sql.NullInt64
		status         sql.NullString
		publishedDocID sql.NullString
	)
	args := append([]any{tenantID, controlledDocumentID}, activeInstanceStatuses...)
	args = append(args, string(documentsdomain.DocStatusPublished))
	err := r.db.QueryRowContext(ctx, `
SELECT active.id,
       COALESCE(active.content_hash_at_submit,
                (SELECT r.content_hash FROM document_revisions r
                  WHERE r.document_id = active.id
                  ORDER BY r.created_at DESC LIMIT 1)),
       active.revision_version,
       active.status,
       pub.id::text
  FROM (SELECT id, content_hash_at_submit, revision_version, status
          FROM documents
         WHERE tenant_id = $1::uuid
           AND controlled_document_id = $2::uuid
           AND status IN ($3, $4, $5, $6)
         LIMIT 1) active
  FULL OUTER JOIN
       (SELECT id FROM documents
         WHERE tenant_id = $1::uuid
           AND controlled_document_id = $2::uuid
           AND status = $7
         ORDER BY revision_number DESC
         LIMIT 1) pub ON TRUE`,
		args...,
	).Scan(&docID, &contentHash, &revisionVer, &status, &publishedDocID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("active instance projection: %w", err)
	}

	// Both sides NULL means the controlled document simply has no documents.
	if !docID.Valid && !publishedDocID.Valid {
		return nil, nil
	}

	view := &documentsdomain.ActiveInstanceView{}
	if docID.Valid {
		v := docID.String
		view.DocumentID = &v
	}
	if contentHash.Valid {
		v := contentHash.String
		view.ContentHash = &v
	}
	if revisionVer.Valid {
		v := int(revisionVer.Int64)
		view.RevisionVersion = &v
	}
	if status.Valid {
		v := status.String
		view.Status = &v
	}
	if publishedDocID.Valid {
		v := publishedDocID.String
		view.PublishedDocumentID = &v
	}

	// When the active document is under review, fetch the in-progress approval
	// instance id (B4). approval_instances is documents-owned (approval sub-context).
	if view.DocumentID != nil && view.Status != nil && *view.Status == string(documentsdomain.DocStatusUnderReview) {
		var approvalInstanceID sql.NullString
		err := r.db.QueryRowContext(ctx, `
SELECT id::text
  FROM approval_instances
 WHERE document_id = $1::uuid
   AND tenant_id = $2::uuid
   AND status = $3
 ORDER BY submitted_at DESC
 LIMIT 1`,
			*view.DocumentID, tenantID, string(approvaldomain.InstanceInProgress),
		).Scan(&approvalInstanceID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("active approval instance: %w", err)
		}
		if approvalInstanceID.Valid {
			v := approvalInstanceID.String
			view.ApprovalInstanceID = &v
		}
	}

	return view, nil
}
