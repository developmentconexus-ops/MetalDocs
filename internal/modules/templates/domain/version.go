package domain

import (
	"errors"
	"time"
)

type VersionStatus string

const (
	VersionStatusDraft     VersionStatus = "draft"
	VersionStatusInReview  VersionStatus = "in_review"
	VersionStatusApproved  VersionStatus = "approved"
	VersionStatusPublished VersionStatus = "published"
	VersionStatusObsolete  VersionStatus = "obsolete"
)

type TemplateVersion struct {
	ID                  string
	TemplateID          string
	VersionNumber       int
	RevisionNumber      int
	Status              VersionStatus
	DocxStorageKey      string
	ContentHash         string
	MetadataSchema      MetadataSchema
	PlaceholderSchema   []Placeholder
	AuthorID            string
	PendingReviewerRole *string
	PendingApproverRole string
	ReviewerID          *string
	ApproverID          *string
	SubmittedAt         *time.Time
	ReviewedAt          *time.Time
	ApprovedAt          *time.Time
	PublishedAt         *time.Time
	ObsoletedAt         *time.Time
	LockVersion         int
	CreatedAt           time.Time
}

func NewTemplateVersionDraft(id, templateID, authorID, docxStorageKey string, versionNumber int, metadata MetadataSchema, placeholders []Placeholder, createdAt time.Time) *TemplateVersion {
	return &TemplateVersion{
		ID:                id,
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

func (v *TemplateVersion) CanTransition(next VersionStatus, hasReviewer bool) error {
	switch v.Status {
	case VersionStatusDraft:
		if next == VersionStatusInReview {
			return nil
		}
	case VersionStatusInReview:
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
	ErrInvalidStateTransition = errors.New("templates: invalid_state_transition")
	ErrContentHashMismatch    = errors.New("templates: content_hash_mismatch")
	ErrStaleBase              = errors.New("templates: stale_base")
	ErrStaleLockVersion       = errors.New("templates: stale_lock_version")
)
