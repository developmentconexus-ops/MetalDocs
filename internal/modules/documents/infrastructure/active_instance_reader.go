package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	approvaldomain "metaldocs/internal/modules/approval/domain"
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

// ActiveInstanceForControlledDocument returns the projection of the active
// (non-terminal) document instance for a controlled document, or nil when none exists.
func (r *ActiveInstanceReaderPG) ActiveInstanceForControlledDocument(ctx context.Context, tenantID, controlledDocumentID string) (*documentsdomain.ActiveInstanceView, error) {
	view, err := r.loadActiveInstanceProjection(ctx, tenantID, controlledDocumentID)
	if err != nil || view == nil {
		return nil, err
	}

	// frozenContentHash carries the pin when an under-review document has an
	// in-progress approval instance that has already frozen (F1/F5 freeze
	// boundary). Folded into the SAME query already fetching the in-progress
	// instance id (identical WHERE predicate) rather than a third round-trip.
	var frozenContentHash sql.NullString

	// When the active document is under review, fetch the in-progress approval
	// instance id (B4). approval_instances is documents-owned (approval sub-context).
	if view.DocumentID != nil && view.Status != nil && *view.Status == string(documentsdomain.DocStatusUnderReview) {
		instanceID, frozenHash, err := r.loadInProgressApprovalInstance(ctx, tenantID, *view.DocumentID)
		if err != nil {
			return nil, err
		}
		view.ApprovalInstanceID = instanceID
		frozenContentHash = frozenHash
	}

	contentHash, err := r.resolveContentHash(ctx, view.DocumentID, frozenContentHash)
	if err != nil {
		return nil, err
	}
	view.ContentHash = contentHash

	return view, nil
}

// loadActiveInstanceProjection runs the FULL OUTER JOIN of the tenant's
// active (non-published) document and its latest published document for
// controlledDocumentID, and maps the scanned row into an ActiveInstanceView.
// Returns (nil, nil) when there is no active AND no published document —
// there is nothing for the caller to build an active-instance view around.
func (r *ActiveInstanceReaderPG) loadActiveInstanceProjection(ctx context.Context, tenantID, controlledDocumentID string) (*documentsdomain.ActiveInstanceView, error) {
	var (
		docID          sql.NullString
		revisionVer    sql.NullInt64
		status         sql.NullString
		publishedDocID sql.NullString
	)
	args := append([]any{tenantID, controlledDocumentID}, activeInstanceStatuses...)
	args = append(args, string(documentsdomain.DocStatusPublished))
	// No-fallback (F6, spec §11): this projection no longer selects or
	// resolves any content hash itself. content_hash_at_submit COALESCE over
	// a live document_revisions hash silently substituted a value that could
	// be captured AFTER an in-progress instance froze — floating content the
	// signer never reviewed. The hash is now resolved by two explicit,
	// status-scoped follow-up queries below (frozen pin, else head revision).
	err := r.db.QueryRowContext(ctx, `
SELECT active.id,
       active.revision_version,
       active.status,
       pub.id::text
  FROM (SELECT id, revision_version, status
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
	).Scan(&docID, &revisionVer, &status, &publishedDocID)
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
	return view, nil
}

// loadInProgressApprovalInstance fetches the in-progress approval instance
// for documentID (B4), returning its id plus the frozen_content_hash pin
// recorded on it (invalid/empty when the instance hasn't frozen yet).
// approval_instances is documents-owned (approval sub-context).
func (r *ActiveInstanceReaderPG) loadInProgressApprovalInstance(ctx context.Context, tenantID, documentID string) (*string, sql.NullString, error) {
	var (
		approvalInstanceID sql.NullString
		frozenContentHash  sql.NullString
	)
	err := r.db.QueryRowContext(ctx, `
SELECT id::text, frozen_content_hash
  FROM approval_instances
 WHERE document_id = $1::uuid
   AND tenant_id = $2::uuid
   AND status = $3
 ORDER BY submitted_at DESC
 LIMIT 1`,
		documentID, tenantID, string(approvaldomain.InstanceInProgress),
	).Scan(&approvalInstanceID, &frozenContentHash)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, sql.NullString{}, fmt.Errorf("active approval instance: %w", err)
	}
	var id *string
	if approvalInstanceID.Valid {
		v := approvalInstanceID.String
		id = &v
	}
	return id, frozenContentHash, nil
}

// resolveContentHash implements the no-fallback (F6) two-step hash
// resolution, never a COALESCE. Step 1 — an already-frozen in-progress
// instance's pin is authoritative (it is the SAME value signoff will compare
// against). Step 2 — otherwise (no instance, or an in-progress instance that
// has not yet frozen, or the document isn't under_review at all) fall
// through to the head document_revisions hash, the only meaningful "current
// content" answer while nothing is pinned yet.
func (r *ActiveInstanceReaderPG) resolveContentHash(ctx context.Context, documentID *string, frozenContentHash sql.NullString) (*string, error) {
	if frozenContentHash.Valid {
		v := frozenContentHash.String
		return &v, nil
	}
	if documentID == nil {
		return nil, nil
	}
	var headHash sql.NullString
	err := r.db.QueryRowContext(ctx, `
SELECT content_hash
  FROM document_revisions
 WHERE document_id = $1::uuid
 ORDER BY created_at DESC
 LIMIT 1`,
		*documentID,
	).Scan(&headHash)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("active document head revision hash: %w", err)
	}
	if headHash.Valid {
		v := headHash.String
		return &v, nil
	}
	return nil, nil
}
