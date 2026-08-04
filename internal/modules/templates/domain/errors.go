package domain

import "errors"

// ErrISOSegregationViolation is returned when a reviewer/approver action would
// violate ISO segregation of duties (actor conflicts with author or reviewer).
var ErrISOSegregationViolation = errors.New("templates: iso_segregation_violation")

// ErrForbidden is returned when the actor lacks permission for the operation.
var ErrForbidden = errors.New("templates: forbidden")

// ErrUploadMissing is returned when a required DOCX upload is absent.
var ErrUploadMissing = errors.New("templates: upload_missing")

// ErrUploadTooLarge is returned when an uploaded DOCX exceeds the size limit.
var ErrUploadTooLarge = errors.New("templates: upload_too_large")

// ErrPlaceholderIDEmpty is returned when a placeholder has an empty ID.
var ErrPlaceholderIDEmpty = errors.New("placeholder id empty")

// ErrDuplicatePlaceholderID is returned when two placeholders share an ID.
var ErrDuplicatePlaceholderID = errors.New("duplicate placeholder id")

// ErrPlaceholderNameInvalid is returned when a placeholder name fails the
// naming rules.
var ErrPlaceholderNameInvalid = errors.New("placeholder name invalid")

// ErrDuplicatePlaceholderName is returned when two placeholders share a name.
var ErrDuplicatePlaceholderName = errors.New("duplicate placeholder name")

// ErrInvalidConstraint is returned when a schema constraint value is invalid
// (e.g. retention_days below 1, or a malformed placeholder constraint).
var ErrInvalidConstraint = errors.New("invalid constraint")

// ErrPlaceholderCycle is returned when placeholder visible_if conditions form
// a dependency cycle.
var ErrPlaceholderCycle = errors.New("placeholder visibility cycle")

// ErrUnknownResolver is returned when a computed placeholder references a
// resolver key that is not registered.
var ErrUnknownResolver = errors.New("unknown resolver")

// ErrPlaceholderNotInCatalog is returned when a catalog placeholder name is
// not part of the fixed catalog.
var ErrPlaceholderNotInCatalog = errors.New("placeholder name not in fixed catalog")

// ErrPlaceholderNotComputed is returned when a catalog placeholder is not
// declared as computed with a resolver_key.
var ErrPlaceholderNotComputed = errors.New("catalog placeholders must be computed with resolver_key")

// ErrPlaceholderReservedName is returned when a dictionary placeholder uses a
// name reserved for native tokens.
var ErrPlaceholderReservedName = errors.New("dictionary placeholder name is reserved (native token)")

// ErrPlaceholderDictionaryInvalid is returned when a dictionary placeholder
// sets resolver_key or computed, which are not allowed for that type.
var ErrPlaceholderDictionaryInvalid = errors.New("dictionary placeholder must not set resolver_key or computed")

// ErrDocTypeCodeRequired is returned when a template is created without a
// doc_type_code. Generic templates were exterminated by ADR 0086 (migration
// 0315's templates_template_doc_type_code_required_check is the DB backstop):
// a template with no profile has no template approval route, so it could never
// be submitted. HTTP layer maps this to 422.
var ErrDocTypeCodeRequired = errors.New("templates: doc_type_code_required")

// ErrApprovalRouteMissing is returned when a template is created for a profile
// that has no ACTIVE template approval route (ADR 0086 config-first gate).
// Mirrors controlleddocuments' identically-named sentinel; HTTP layer maps it
// to 409 APPROVAL_ROUTE_MISSING.
var ErrApprovalRouteMissing = errors.New("templates: approval_route_missing")

// ErrContentMaterializationFailed is returned when the server cannot produce
// the object a new template version must point at (ADR 0088 §2 — every version
// is born with the verified hash of its own object).
//
// It covers EVERY failure inside the server-side store-then-reference copy,
// because every caller of that primitive copies bytes the SERVER already owns:
// blank creation copies the system blank asset, and spawning the next revision
// copies the previous version's object. Concretely: the source row is absent or
// carries no storage key / no 64-hex hash; the copy itself fails; the copied
// object is missing, hash-mismatched, or oversized on read-back.
//
// None of those are user errors, and the distinction between them is an
// OPERATIONS distinction, not a client one — no caller can act on it, so it is
// carried in the wrapped error chain for logs rather than in the classification.
// Mapping any of them to the user-facing upload sentinels (as the pre-ADR-0088
// code did) tells the user to fix an upload they never made and cannot make,
// while hiding an object-store failure behind a 4xx.
//
// The HTTP layer therefore leaves this on the default 500 arm, deliberately: a
// 4xx would blame the request when the deployment is what is broken. It is also
// never a licence to fall back to a content-less version (no-fallback
// principle) — the state ADR 0088 exists to make unreachable must not be
// re-opened by a degraded path.
//
// Operator fix, when the source is the system blank: re-run the reference-data
// bundle and the object seeding (deploy/assets/system-blank.docx ->
// system/templates/blank.docx).
var ErrContentMaterializationFailed = errors.New("templates: content_materialization_failed")

// ErrTransactionRequired is returned when an operation that must run inside a
// transaction is invoked without one.
var ErrTransactionRequired = errors.New("templates: transaction_required")

// ErrConcurrentTransition is returned when a status-transition write loses a
// CAS race against a concurrent transition on the same version.  The caller
// should retry the full read-modify-write cycle.  HTTP layer maps this to 409.
var ErrConcurrentTransition = errors.New("templates: concurrent_transition")
