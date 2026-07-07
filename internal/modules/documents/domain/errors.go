package domain

import "errors"

var ErrValidationFailed = errors.New("placeholder value validation failed")

// ErrPlaceholderNotAuthorEditable signals an author fill-in write targeting a
// row whose current source is governed (computed/dictionary). Maps to 409 (SP-2 D11).
var ErrPlaceholderNotAuthorEditable = errors.New("documents: placeholder is not author-editable")

// ErrDictionaryTokenMissing signals a referenced dictionary token has no entry
// for the tenant at creation time (SP-2 D7). The CD layer maps it to 422.
var ErrDictionaryTokenMissing = errors.New("documents: referenced dictionary token not found")
var ErrEffectiveDateMissing = errors.New("effective_date missing")

// ErrDocumentNotDraft is returned when in-tx submit resolution targets a
// document that is not in 'draft' status (409 StateTransitionInvalid). ADR 0073.
var ErrDocumentNotDraft = errors.New("document is not in draft status")

// ErrProfileNotConfigured is returned when the controlled document linked to the
// document has no profile_code set (400 ValidationError). ADR 0073.
var ErrProfileNotConfigured = errors.New("controlled document has no profile configured")

// ErrApprovalRouteMissing is returned when no active approval route exists for
// the document's profile (409 ApprovalRouteMissing). ADR 0073.
var ErrApprovalRouteMissing = errors.New("no active approval route for this profile")
