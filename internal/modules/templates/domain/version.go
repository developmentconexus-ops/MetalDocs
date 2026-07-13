package domain

import (
	"errors"
	"time"
)

// VersionStatus is the lifecycle state of a template version.
type VersionStatus string

// Lifecycle states of a template version, in forward order.
const (
	// VersionStatusDraft is the initial state of a newly created version,
	// editable by its author.
	VersionStatusDraft VersionStatus = "draft"
	// VersionStatusUnderReview marks a version submitted for review/approval
	// and no longer editable by its author.
	VersionStatusUnderReview VersionStatus = "under_review"
	// VersionStatusApproved marks a version that has passed review and is
	// awaiting publication.
	VersionStatusApproved VersionStatus = "approved"
	// VersionStatusPublished marks a version that is the live, in-force
	// revision of its template.
	VersionStatusPublished VersionStatus = "published"
	// VersionStatusObsolete marks a version that has been superseded by a
	// later published version.
	VersionStatusObsolete VersionStatus = "obsolete"
)

// TemplateVersion is one immutable-content revision of a Template moving
// through the review/approval lifecycle. It carries the DOCX storage key and
// content hash, the metadata and placeholder schemas, the actor IDs collected
// at each transition, and an optimistic-lock version.
type TemplateVersion struct {
	ID                string
	TenantID          string
	TemplateID        string
	VersionNumber     int
	RevisionNumber    int
	Status            VersionStatus
	DocxStorageKey    string
	ContentHash       string
	MetadataSchema    MetadataSchema
	PlaceholderSchema []Placeholder
	AuthorID          string
	ReviewerID        *string
	ApproverID        *string
	SubmittedAt       *time.Time
	ReviewedAt        *time.Time
	ApprovedAt        *time.Time
	PublishedAt       *time.Time
	ObsoletedAt       *time.Time
	LockVersion       int
	CreatedAt         time.Time
}

// NewTemplateVersionDraft constructs a new version in draft status with an
// empty content hash; the hash is set once the DOCX content is persisted.
func NewTemplateVersionDraft(id, tenantID, templateID, authorID, docxStorageKey string, versionNumber int, metadata MetadataSchema, placeholders []Placeholder, createdAt time.Time) *TemplateVersion {
	return &TemplateVersion{
		ID:                id,
		TenantID:          tenantID,
		TemplateID:        templateID,
		VersionNumber:     versionNumber,
		Status:            VersionStatusDraft,
		DocxStorageKey:    docxStorageKey,
		ContentHash:       "",
		MetadataSchema:    metadata,
		PlaceholderSchema: placeholders,
		AuthorID:          authorID,
		CreatedAt:         createdAt,
	}
}

// CanTransition reports whether moving from the version's current status to
// next is a legal lifecycle transition. hasReviewer selects the review path:
// with a reviewer, under_review must pass through approved before publish;
// without one, under_review may publish directly. Returns
// ErrInvalidStateTransition when the move is not allowed.
func (v *TemplateVersion) CanTransition(next VersionStatus, hasReviewer bool) error {
	switch v.Status {
	case VersionStatusDraft:
		if next == VersionStatusUnderReview {
			return nil
		}
	case VersionStatusUnderReview:
		if next == VersionStatusDraft {
			return nil
		}
		if next == VersionStatusApproved && hasReviewer {
			return nil
		}
		if next == VersionStatusPublished && !hasReviewer {
			return nil
		}
	case VersionStatusApproved:
		if next == VersionStatusPublished || next == VersionStatusDraft {
			return nil
		}
	case VersionStatusPublished:
		if next == VersionStatusObsolete {
			return nil
		}
	}
	return ErrInvalidStateTransition
}

var (
	// ErrInvalidStateTransition is returned when a lifecycle move is not
	// permitted from the version's current status.
	ErrInvalidStateTransition = errors.New("templates: invalid_state_transition")
	// ErrContentHashMismatch is returned when uploaded content does not match
	// the expected content hash.
	ErrContentHashMismatch = errors.New("templates: content_hash_mismatch")
	// ErrStaleBase is returned when an autosave is based on an outdated
	// revision of the draft.
	ErrStaleBase = errors.New("templates: stale_base")
	// ErrStaleLockVersion is returned when an optimistic-lock check fails
	// because the version row changed since it was read.
	ErrStaleLockVersion = errors.New("templates: stale_lock_version")
)
