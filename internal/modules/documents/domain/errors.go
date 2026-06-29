package domain

import "errors"

var ErrValidationFailed = errors.New("placeholder value validation failed")

// ErrDictionaryTokenMissing signals a referenced dictionary token has no entry
// for the tenant at creation time (SP-2 D7). The CD layer maps it to 422.
var ErrDictionaryTokenMissing = errors.New("documents: referenced dictionary token not found")
var ErrEffectiveDateMissing = errors.New("effective_date missing")

// ErrDocumentNotDraft is returned when finalizeDocument targets a document that
// is not in 'draft' status (409 StateTransitionInvalid).
var ErrDocumentNotDraft = errors.New("document is not in draft status")

// ErrProfileNotConfigured is returned when the controlled document linked to the
// document has no profile_code set (400 ValidationError).
var ErrProfileNotConfigured = errors.New("controlled document has no profile configured")

// ErrApprovalRouteMissing is returned when no active approval route exists for
// the document's profile (409 ApprovalRouteMissing).
var ErrApprovalRouteMissing = errors.New("no active approval route for this profile")

// FinalizePrereqs holds the data that finalizeDocument reads from the database
// before delegating to the approval submit service.
type FinalizePrereqs struct {
	RevisionVersion int64
	RevisionNumber  int64
	// ControlledDocumentID is empty when the document is not linked to a
	// controlled document.
	ControlledDocumentID string
	// ProfileCode is empty when no controlled document is linked.
	ProfileCode string
	// RouteID is the active approval route id for the document's profile.
	RouteID string
	// ContentHash is the latest autosaved content hash; may be empty.
	ContentHash string
}
