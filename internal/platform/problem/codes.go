// Package problem defines the canonical error code catalog used by
// MetalDocs Problem responses (RFC 9457). See spec §5.2.
package problem

// Code is the type for canonical error codes in Problem responses.
// Using a distinct type prevents arbitrary strings from being used as codes.
type Code string

// HTTP-status-level codes (used in Problem.code).
const (
	CodeValidationError        Code = "VALIDATION_ERROR"
	CodeUnknownField           Code = "UNKNOWN_FIELD"
	CodeUnknownFilter          Code = "UNKNOWN_FILTER"
	CodeInvalidSortField       Code = "INVALID_SORT_FIELD"
	CodeInvalidCursor          Code = "INVALID_CURSOR"
	CodeIncludeNotSupported    Code = "INCLUDE_NOT_SUPPORTED"
	CodeUnauthenticated        Code = "UNAUTHENTICATED"
	CodeForbiddenCapability    Code = "FORBIDDEN_CAPABILITY"
	CodeForbiddenArea          Code = "FORBIDDEN_AREA"
	CodeForbiddenOrigin        Code = "FORBIDDEN_ORIGIN"
	CodeNotFound               Code = "NOT_FOUND"
	CodeMethodNotAllowed       Code = "METHOD_NOT_ALLOWED"
	CodeAlreadyExists          Code = "ALREADY_EXISTS"
	CodeStateTransitionInvalid Code = "STATE_TRANSITION_INVALID"
	CodeConcurrentModification Code = "CONCURRENT_MODIFICATION"
	CodeIdempotencyKeyReused   Code = "IDEMPOTENCY_KEY_REUSED"
	CodeIdempotencyKeyInvalid  Code = "IDEMPOTENCY_KEY_INVALID"
	CodeIdempotencyReplay      Code = "IDEMPOTENCY_REPLAY"
	CodeRequestBodyTooLarge    Code = "REQUEST_BODY_TOO_LARGE"
	CodeRateLimited            Code = "RATE_LIMITED" // emitted by the rate-limit middlewares (platform/ratelimit + platform/security)
	CodeInternalError          Code = "INTERNAL_ERROR"
	CodeNotImplemented         Code = "NOT_IMPLEMENTED" // 501: endpoint/feature wired but not yet implemented
	CodeCursorExpired          Code = "CURSOR_EXPIRED"  // 410: pagination cursor refers to an item that no longer exists
	CodeConflict               Code = "CONFLICT_ERROR"
	CodeIdempotencyKeyRequired Code = "IDEMPOTENCY_KEY_REQUIRED"

	CodeAuthUnauthorized           Code = "AUTH_UNAUTHORIZED"
	CodeAuthInvalidCredentials     Code = "AUTH_INVALID_CREDENTIALS"
	CodeAuthAccountLocked          Code = "AUTH_ACCOUNT_LOCKED"
	CodeAuthAccountInactive        Code = "AUTH_ACCOUNT_INACTIVE"
	CodeAuthTenantForbidden        Code = "AUTH_TENANT_FORBIDDEN"
	CodeAuthTenantRequired         Code = "AUTH_TENANT_REQUIRED"
	CodeAuthPasswordChangeRequired Code = "AUTH_PASSWORD_CHANGE_REQUIRED"
	CodeAuthForbidden              Code = "AUTH_FORBIDDEN"
)

// Documents/templates domain-specific codes (used in Problem.code).
const (
	CodeApprovalRouteMissing    Code = "APPROVAL_ROUTE_MISSING"
	CodeUploadExpired           Code = "UPLOAD_EXPIRED"
	CodeUploadMissing           Code = "UPLOAD_MISSING"
	CodeStaleBase               Code = "STALE_BASE"
	CodeISOSegregationViolation Code = "ISO_SEGREGATION_VIOLATION"
	CodeSystemTemplateImmutable Code = "SYSTEM_TEMPLATE_IMMUTABLE"
	// CodePreconditionRequired is the canonical-catalog counterpart of the
	// approval/http package's local approvalCodePreconditionIfMatch
	// ("precondition.if_match_required") — that package is excluded from the
	// guarded-package catalog check (dotted taxonomy predates this catalog), but
	// documents/delivery/http is guarded, so CON-01's finalize handler (mandatory
	// If-Match, OCC parity with /documents/{id}/submit) uses this typed constant.
	CodePreconditionRequired Code = "PRECONDITION_REQUIRED"
)

// IAM area-membership domain codes (used in Problem.code by the memberships
// handler). These have been part of the wire contract — and the frontend
// error-code catalog — since Wave 1; they are promoted to typed constants here
// (H-4) with their exact existing string values, so the contract is unchanged.
const (
	CodeMembershipExists   Code = "MEMBERSHIP_EXISTS"
	CodeMembershipNotFound Code = "MEMBERSHIP_NOT_FOUND"
	CodeUnknownRole        Code = "UNKNOWN_ROLE"
)

// Field-level codes (used in FieldError.code).
const (
	FieldCodeRequired      Code = "REQUIRED"
	FieldCodeInvalidFormat Code = "INVALID_FORMAT"
	FieldCodeOutOfRange    Code = "OUT_OF_RANGE"
	FieldCodeInvalidEnum   Code = "INVALID_ENUM"
)
