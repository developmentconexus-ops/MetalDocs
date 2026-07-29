package domain

import (
	"errors"
	"time"
)

type DocumentStatus string

const (
	DocStatusDraft       DocumentStatus = "draft"
	DocStatusUnderReview DocumentStatus = "under_review" // migration 0142 removed draft→finalized; submit = under_review // cilint:allow-legacy (historical note)
	DocStatusApproved    DocumentStatus = "approved"
	DocStatusScheduled   DocumentStatus = "scheduled"
	DocStatusPublished   DocumentStatus = "published"
	DocStatusSuperseded  DocumentStatus = "superseded"
	DocStatusObsolete    DocumentStatus = "obsolete"
	DocStatusArchived    DocumentStatus = "archived"
)

type SessionStatus string

const (
	SessionActive        SessionStatus = "active"
	SessionExpired       SessionStatus = "expired"
	SessionReleased      SessionStatus = "released"
	SessionForceReleased SessionStatus = "force_released"
)

type Document struct {
	ID                string         `json:"id"`
	TenantID          string         `json:"tenant_id"`
	TemplateVersionID string         `json:"template_version_id"`
	Name              string         `json:"name"`
	Status            DocumentStatus `json:"status"`
	FormDataJSON      []byte         `json:"form_data_json"`
	CurrentRevisionID string         `json:"current_revision_id"`
	RevisionVersion   int64          `json:"revision_version"`
	ActiveSessionID   string         `json:"active_session_id"`
	ValuesFrozenAt    *time.Time     `json:"values_frozen_at,omitempty"`
	ArchivedAt        *time.Time     `json:"archived_at,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	CreatedBy         string         `json:"created_by"`
	RevisionNumber    int64          `json:"revision_number"`
	RevisionTitle     *string        `json:"revision_title,omitempty"`
	// Bridge fields (Spec 1 — added as nullable for Phase A; NOT NULL enforced in migration 0129)
	ControlledDocumentID           *string `json:"controlled_document_id,omitempty"`
	ProfileCodeSnapshot            *string `json:"profile_code_snapshot,omitempty"`
	ProcessAreaCodeSnapshot        *string `json:"process_area_code_snapshot,omitempty"`
	Code                           string  `json:"code"`
	CurrentRevisionFileSizeBytes   *int64  `json:"current_revision_file_size_bytes,omitempty"`
	CurrentRevisionPageCount       *int    `json:"current_revision_page_count,omitempty"`
	CurrentRevisionPageCountSource *string `json:"current_revision_page_count_source,omitempty"`
	// Review/expiry model (M6 F6.2, migrations 0274/0276). effective_from/
	// effective_to already existed on the row; review_due_at/last_reviewed_at
	// are new. All four nullable — legacy rows keep NULL (no review cycle set).
	EffectiveFrom  *time.Time `json:"effective_from,omitempty"`
	EffectiveTo    *time.Time `json:"effective_to,omitempty"`
	ReviewDueAt    *time.Time `json:"review_due_at,omitempty"`
	LastReviewedAt *time.Time `json:"last_reviewed_at,omitempty"`
	// ReviewSurfacedAt (migration 0276, M6 F6.2 T4) records when the River
	// periodic surfacer flagged this document for its current review_due_at
	// cycle. F6.4 D2 makes this the worklist marker the review-due filter
	// reads (see buildDocumentFilter's ReviewDue branch in repository.go).
	ReviewSurfacedAt *time.Time `json:"review_surfaced_at,omitempty"`
	// Release is the ADR 0085 Stage C readiness-hold projection for this
	// document's LATEST release generation (public.release_generations,
	// highest generation_seq). nil when the document has no release
	// generation at all — pre-approval drafts and legacy rows that were never
	// backfilled. Composed into the SINGLE detail read query (GetDocument's
	// LEFT JOIN LATERAL), never a second round-trip.
	Release *ReleaseProjection `json:"release,omitempty"`
	// TemplateVersionID is already present above — now semantically write-once after this migration

	// TemplateSnapshot is the frozen template payload. Populated by Service.Create
	// from SnapshotService.ResolveTemplate before CreateDocument INSERT so all
	// snapshot columns are written atomically with the documents row.
	TemplateSnapshot TemplateSnapshot `json:"-"`
}

// ReleaseState is the coarse, server-derived release status of a release
// generation (ADR 0085 Stage C readiness-hold projection). It is DERIVED, never
// stored: released_at IS NOT NULL => ReleaseStateReleased, otherwise
// ReleaseStateHold. Mirrored by the `state` enum on the OpenAPI
// DocumentReleaseProjection schema.
type ReleaseState string

const (
	// ReleaseStateHold — the coordinator has not (yet) released this
	// generation. HoldReason carries why, when the coordinator has evaluated it.
	ReleaseStateHold ReleaseState = "hold"
	// ReleaseStateReleased — a winning release transaction stamped released_at.
	ReleaseStateReleased ReleaseState = "released"
)

// ReleaseProjection is the read-model view of ONE release generation — the
// latest one for a document. It is a pure projection: every field is read
// straight out of the row (or, for State, derived from released_at) and the
// documents module never writes it. The hold_reason vocabulary is owned by the
// approval module (approval/domain.ReleaseHoldReason, DB-pinned by
// ck_release_generations_hold_reason in migration 0310) and is carried here as
// an opaque pass-through string, exactly like CurrentRevisionPageCountSource.
type ReleaseProjection struct {
	GenerationID string       `json:"generation_id"`
	State        ReleaseState `json:"state"`
	// HoldReason/HoldDetail are nil on a released generation (the release tx
	// clears them) and on a hold the coordinator has not evaluated yet.
	HoldReason *string `json:"hold_reason,omitempty"`
	HoldDetail *string `json:"hold_detail,omitempty"`
	// PlannedEffectiveFrom is the PLAN half of ADR 0085's planned/actual split
	// and is sourced from documents.planned_effective_from (NOT from the
	// generation row). nil means "effective on release".
	PlannedEffectiveFrom *time.Time `json:"planned_effective_from,omitempty"`
	ReleasedAt           *time.Time `json:"released_at,omitempty"`
	LastEvaluatedAt      *time.Time `json:"last_evaluated_at,omitempty"`
}

type Session struct {
	ID                         string
	DocumentID                 string
	UserID                     string
	AcquiredAt                 time.Time
	ExpiresAt                  time.Time
	ReleasedAt                 *time.Time
	LastAcknowledgedRevisionID string
	Status                     SessionStatus
}

type Revision struct {
	ID               string
	DocumentID       string
	RevisionNum      int64
	ParentRevisionID string
	SessionID        string
	StorageKey       string
	ContentHash      string
	FormDataSnapshot []byte
	CreatedAt        time.Time
}

type RevisionHistoryItem struct {
	DocumentID     string
	RevisionNumber int64
	RevisionTitle  string
	Status         DocumentStatus
	CreatedAt      time.Time
	IsCurrent      bool
}

type PendingUpload struct {
	ID             string
	SessionID      string
	DocumentID     string
	BaseRevisionID string
	ContentHash    string
	StorageKey     string
	PresignedAt    time.Time
	ExpiresAt      time.Time
	ConsumedAt     *time.Time
}

type Checkpoint struct {
	ID         string
	DocumentID string
	RevisionID string
	VersionNum int
	Label      string
	CreatedAt  time.Time
	CreatedBy  string
}

var (
	ErrInvalidStateTransition = errors.New("invalid_state_transition")
	ErrSessionInactive        = errors.New("session_inactive")
	ErrSessionNotHolder       = errors.New("session_not_holder")
	ErrStaleBase              = errors.New("stale_base")
	ErrMisbound               = errors.New("misbound")
	ErrExpiredUpload          = errors.New("expired_upload")
	ErrContentHashMismatch    = errors.New("content_hash_mismatch")
	ErrPendingNotFound        = errors.New("pending_not_found")
	ErrAlreadyConsumed        = errors.New("already_consumed")
	ErrSessionTaken           = errors.New("session_taken")
	ErrForbidden              = errors.New("forbidden")
	ErrUploadMissing          = errors.New("upload_missing")
	ErrUploadTooLarge         = errors.New("upload_too_large")
	ErrInvalidPageCount       = errors.New("invalid_page_count")
	ErrCheckpointNotFound     = errors.New("checkpoint_not_found")
	ErrDocumentNotOwner       = errors.New("document_not_owner")
	ErrNotFound               = errors.New("not_found")
	ErrInvalidName            = errors.New("invalid_name")
	ErrCommentNotFound        = errors.New("comment_not_found")
	ErrCommentInvalid         = errors.New("comment_invalid")
)
