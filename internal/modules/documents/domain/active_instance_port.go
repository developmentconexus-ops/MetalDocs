package domain

import "context"

// ActiveInstanceView is the documents-owned projection of a controlled
// document's active/published documents (plus the in-progress approval instance
// when the active document is under review). It is the published read-port shape
// consumed by controlleddocuments' GetActiveInstance (ADR-0039 D3(b)); documents
// owns documents/document_revisions/approval_instances, so this is the single
// cross-module read surface for that projection.
//
// All pointer fields are nil when the corresponding row/column does not exist.
type ActiveInstanceView struct {
	// DocumentID is the active (non-published) document id, if any.
	DocumentID *string
	// ContentHash is the active document's content hash: content_hash_at_submit
	// when set, else the latest document_revisions.content_hash.
	ContentHash *string
	// RevisionVersion is the active document's revision_version, if any.
	RevisionVersion *int
	// Status is the active document's status (owner vocabulary), if any.
	Status *string
	// PublishedDocumentID is the most-recent published document id, if any.
	PublishedDocumentID *string
	// ApprovalInstanceID is the in-progress approval instance id for the active
	// document when its status is under_review. Nil otherwise.
	ApprovalInstanceID *string
}

// ActiveInstanceReader is the documents-owned read-port for the active-instance
// projection. The interface lives in the owner's domain package; the Postgres
// adapter lives in documents/infrastructure and is wired at the composition
// root. Consumers (controlleddocuments) depend on this interface and never name
// the documents base tables in their own SQL.
type ActiveInstanceReader interface {
	// ActiveInstanceForControlledDocument returns the projection for
	// (tenantID, controlledDocumentID), or nil when neither an active nor a
	// published document exists. Off-tx (runs on the owner's pool).
	ActiveInstanceForControlledDocument(ctx context.Context, tenantID, controlledDocumentID string) (*ActiveInstanceView, error)
}

// NoopActiveInstanceReader is the fail-closed default: it reports no active or
// published document. Used as the nil-guard default at wiring and in unit tests
// that do not exercise GetActiveInstance.
type NoopActiveInstanceReader struct{}

func (NoopActiveInstanceReader) ActiveInstanceForControlledDocument(context.Context, string, string) (*ActiveInstanceView, error) {
	return nil, nil
}
