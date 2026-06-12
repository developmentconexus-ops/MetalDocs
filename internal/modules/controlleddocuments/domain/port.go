package domain

import (
	"context"
	"time"
)

// ActiveDocumentInstance is the result of a GetActiveInstance repository query.
// All pointer fields are nil when the corresponding row/column does not exist.
type ActiveDocumentInstance struct {
	// DocumentID is the active (non-published) document id, if any.
	DocumentID *string
	// ContentHash is the content hash from the active document, if any.
	ContentHash *string
	// RevisionVersion is the active document revision version, if any.
	RevisionVersion *int
	// ApprovalState is the active document status string, if any.
	ApprovalState *string
	// PublishedDocumentID is the most recent published document id, if any.
	PublishedDocumentID *string
	// ApprovalInstanceID is the in-progress approval instance id for the
	// active document when its status is 'under_review'. Nil otherwise.
	ApprovalInstanceID *string
}

type ControlledDocumentRepository interface {
	// Read operations.
	GetByID(ctx context.Context, tenantID, id string) (*ControlledDocument, error)
	GetByCode(ctx context.Context, tenantID, profileCode, code string) (*ControlledDocument, error)
	CodeExists(ctx context.Context, tenantID, profileCode, code string) (bool, error)
	List(ctx context.Context, tenantID string, filter CDFilter) ([]ControlledDocument, bool, error)
	CanRead(ctx context.Context, tenantID, controlledDocumentID, actorUserID string) (bool, error)
	// GetActiveInstance returns the active document instance for the given
	// controlled document, including the published document id and (when
	// applicable) the in-progress approval instance id.
	// Returns nil, nil when neither an active nor a published document exists.
	GetActiveInstance(ctx context.Context, tenantID, controlledDocumentID string) (*ActiveDocumentInstance, error)

	// Write operations.
	Create(ctx context.Context, doc *ControlledDocument) error
	CreateTx(ctx context.Context, tx DBTX, doc *ControlledDocument) error
	UpdateStatus(ctx context.Context, tenantID, id string, status CDStatus, updatedAt time.Time) error
	UpdateStatusTx(ctx context.Context, tx DBTX, tenantID, id string, status CDStatus, updatedAt time.Time) error
}

type CDFilter struct {
	ProfileCode     *string
	ProcessAreaCode *string
	UserAreaCodes   []string
	DepartmentCode  *string
	OwnerUserID     *string
	Status          *CDStatus
	Query           *string
	ActorUserID     *string
	Limit           int
	// Cursor is the opaque forward keyset cursor (FD-2). Empty = first page.
	Cursor string
}
