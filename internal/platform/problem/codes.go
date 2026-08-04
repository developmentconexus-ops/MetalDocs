// Package problem defines the canonical error code catalog used by
// MetalDocs Problem responses (RFC 9457). See spec §5.2.
package problem

// LEGACY CATALOG — every declaration in this file is TEMPORARY.
//
// These are the 47 pre-ADR-0089 catalog codes, registered verbatim: same wire
// strings, same statuses, zero behaviour change. They go through RegisterLegacy
// because their names are SCREAMING_SNAKE and predate the closed semantic family
// set, so Register would (correctly) reject them.
//
// ADR 0089 execution step 5 rewrites this file from annex §2.1: each surviving
// row becomes Register / RegisterWithStatus under its new dotted name, the 11
// dead rows are deleted, and RegisterLegacy itself is removed. Do NOT add
// entries here — a new code belongs in its owning module, registered with
// problem.Register under a semantic family.
//
// The declarations are `var`, not `const`, because Register* runs at init.
// Nothing depended on their const-ness.

// HTTP-status-level codes (used in Problem.code).
var (
	CodeValidationError        = RegisterLegacy("platform", "VALIDATION_ERROR", 400)
	CodeUnknownField           = RegisterLegacy("platform", "UNKNOWN_FIELD", 400)
	CodeUnknownFilter          = RegisterLegacy("platform", "UNKNOWN_FILTER", 400)
	CodeInvalidSortField       = RegisterLegacy("platform", "INVALID_SORT_FIELD", 400)
	CodeInvalidCursor          = RegisterLegacy("platform", "INVALID_CURSOR", 400)
	CodeIncludeNotSupported    = RegisterLegacy("platform", "INCLUDE_NOT_SUPPORTED", 400)
	CodeUnauthenticated        = RegisterLegacy("platform", "UNAUTHENTICATED", 401)
	CodeForbiddenCapability    = RegisterLegacy("platform", "FORBIDDEN_CAPABILITY", 403)
	CodeForbiddenArea          = RegisterLegacy("platform", "FORBIDDEN_AREA", 403)
	CodeForbiddenOrigin        = RegisterLegacy("platform", "FORBIDDEN_ORIGIN", 403)
	CodeNotFound               = RegisterLegacy("platform", "NOT_FOUND", 404)
	CodeMethodNotAllowed       = RegisterLegacy("platform", "METHOD_NOT_ALLOWED", 405)
	CodeAlreadyExists          = RegisterLegacy("platform", "ALREADY_EXISTS", 409)
	CodeStateTransitionInvalid = RegisterLegacy("platform", "STATE_TRANSITION_INVALID", 409)
	CodeConcurrentModification = RegisterLegacy("platform", "CONCURRENT_MODIFICATION", 409)
	CodeIdempotencyKeyReused   = RegisterLegacy("platform", "IDEMPOTENCY_KEY_REUSED", 422) // R-8 will move this to 409
	CodeIdempotencyKeyInvalid  = RegisterLegacy("platform", "IDEMPOTENCY_KEY_INVALID", 400)
	CodeIdempotencyReplay      = RegisterLegacy("platform", "IDEMPOTENCY_REPLAY", 409)
	CodeRequestBodyTooLarge    = RegisterLegacy("platform", "REQUEST_BODY_TOO_LARGE", 413)
	CodeRateLimited            = RegisterLegacy("platform", "RATE_LIMITED", 429) // emitted by the rate-limit middlewares (platform/ratelimit + platform/security)
	CodeInternalError          = RegisterLegacy("platform", "INTERNAL_ERROR", 500)
	CodeNotImplemented         = RegisterLegacy("platform", "NOT_IMPLEMENTED", 501) // 501: endpoint/feature wired but not yet implemented
	CodeCursorExpired          = RegisterLegacy("platform", "CURSOR_EXPIRED", 410)  // 410: pagination cursor refers to an item that no longer exists
	CodeConflict               = RegisterLegacy("platform", "CONFLICT_ERROR", 409)
	CodeIdempotencyKeyRequired = RegisterLegacy("platform", "IDEMPOTENCY_KEY_REQUIRED", 400)

	CodeAuthUnauthorized           = RegisterLegacy("platform", "AUTH_UNAUTHORIZED", 401)
	CodeAuthInvalidCredentials     = RegisterLegacy("platform", "AUTH_INVALID_CREDENTIALS", 401)
	CodeAuthAccountLocked          = RegisterLegacy("platform", "AUTH_ACCOUNT_LOCKED", 403)
	CodeAuthAccountInactive        = RegisterLegacy("platform", "AUTH_ACCOUNT_INACTIVE", 403)
	CodeAuthTenantForbidden        = RegisterLegacy("platform", "AUTH_TENANT_FORBIDDEN", 403)
	CodeAuthTenantRequired         = RegisterLegacy("platform", "AUTH_TENANT_REQUIRED", 403)
	CodeAuthPasswordChangeRequired = RegisterLegacy("platform", "AUTH_PASSWORD_CHANGE_REQUIRED", 403)
	CodeAuthForbidden              = RegisterLegacy("platform", "AUTH_FORBIDDEN", 403)
)

// Documents/templates domain-specific codes (used in Problem.code).
var (
	CodeApprovalRouteMissing = RegisterLegacy("platform", "APPROVAL_ROUTE_MISSING", 409)
	CodeUploadExpired        = RegisterLegacy("platform", "UPLOAD_EXPIRED", 410)
	// CodeUploadMissing is registered at the documents status (410). Templates
	// emits the same code at 409; ADR 0089 ruling R-1 unifies both on 409. The
	// binding is inert until call sites move to NewFor, so recording the current
	// documents value here changes nothing on the wire.
	CodeUploadMissing           = RegisterLegacy("platform", "UPLOAD_MISSING", 410)
	CodeStaleBase               = RegisterLegacy("platform", "STALE_BASE", 409)
	CodeISOSegregationViolation = RegisterLegacy("platform", "ISO_SEGREGATION_VIOLATION", 403)
	CodeSystemTemplateImmutable = RegisterLegacy("platform", "SYSTEM_TEMPLATE_IMMUTABLE", 409)
	// CodePreconditionRequired is the canonical-catalog counterpart of the
	// approval/http package's local approvalCodePreconditionIfMatch
	// ("precondition.if_match_required"). Reserved wire-contract code: the
	// finalize handler that consumed it was removed (ADR 0073); retained for the
	// FE error-code catalog and any future If-Match/OCC precondition surface.
	CodePreconditionRequired = RegisterLegacy("platform", "PRECONDITION_REQUIRED", 428)
)

// IAM area-membership domain codes (used in Problem.code by the memberships
// handler). These have been part of the wire contract — and the frontend
// error-code catalog — since Wave 1; they are promoted to typed constants here
// (H-4) with their exact existing string values, so the contract is unchanged.
var (
	CodeMembershipExists   = RegisterLegacy("platform", "MEMBERSHIP_EXISTS", 409)
	CodeMembershipNotFound = RegisterLegacy("platform", "MEMBERSHIP_NOT_FOUND", 404)
	CodeUnknownRole        = RegisterLegacy("platform", "UNKNOWN_ROLE", 400)
)

// SHARED LEGACY CODES — one wire string emitted by two or more modules.
//
// These strings already existed twice in the tree, declared independently by two
// modules (or by one module and a raw literal in another). The registry's
// duplicate guard makes an independent second declaration a build break, which
// is the whole point of ADR 0089 decision 2 — so the declaration has to move to
// a place both emitters can reference. `platform/problem` is that place; it is
// the only package both sides already import, and having one module import
// another module's delivery package would violate the module boundary.
//
// This is a REGISTRATION collapse, not a semantic one: the wire strings and the
// per-call-site statuses are untouched. ADR 0089 execution step 10 does the
// semantic collapse (which sentinel maps to which arm) and re-homes these into
// their owning module or into the renamed catalog.
var (
	CodeSharedInternalUnknown           = RegisterLegacy("shared", "internal.unknown", 500)             // approval + documents fill-in
	CodeSharedValidationJSONDecode      = RegisterLegacy("shared", "validation.json_decode", 400)       // approval + documents fill-in
	CodeSharedValidationEmptyBody       = RegisterLegacy("shared", "validation.empty_body", 400)        // approval + documents fill-in
	CodeSharedAuthzCapabilityDenied     = RegisterLegacy("shared", "authz.capability_denied", 403)      // approval + documents fill-in
	CodeSharedStateApprovalRouteMissing = RegisterLegacy("shared", "state.approval_route_missing", 409) // approval + controlleddocuments
	CodeSharedProfileNotFound           = RegisterLegacy("shared", "PROFILE_NOT_FOUND", 404)            // taxonomy + controlleddocuments
	CodeSharedProfileArchived           = RegisterLegacy("shared", "PROFILE_ARCHIVED", 409)             // taxonomy + controlleddocuments
	CodeSharedAreaNotFound              = RegisterLegacy("shared", "AREA_NOT_FOUND", 404)               // taxonomy + controlleddocuments
	CodeSharedAreaArchived              = RegisterLegacy("shared", "AREA_ARCHIVED", 409)                // taxonomy + controlleddocuments
)

// Field-level codes (used in FieldError.code). Separate namespace, no status.
var (
	FieldCodeRequired      = RegisterField("platform", "REQUIRED")
	FieldCodeInvalidFormat = RegisterField("platform", "INVALID_FORMAT")
	FieldCodeOutOfRange    = RegisterField("platform", "OUT_OF_RANGE")
	FieldCodeInvalidEnum   = RegisterField("platform", "INVALID_ENUM")
)
