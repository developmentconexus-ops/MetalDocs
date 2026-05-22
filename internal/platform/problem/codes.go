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
	CodeAlreadyExists          Code = "ALREADY_EXISTS"
	CodeStateTransitionInvalid Code = "STATE_TRANSITION_INVALID"
	CodeConcurrentModification Code = "CONCURRENT_MODIFICATION"
	CodeIdempotencyKeyReused   Code = "IDEMPOTENCY_KEY_REUSED"
	CodeIdempotencyReplay      Code = "IDEMPOTENCY_REPLAY"
	CodeRateLimited            Code = "RATE_LIMITED" // reserved; not enforced yet
	CodeInternalError          Code = "INTERNAL_ERROR"
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

// Field-level codes (used in FieldError.code).
const (
	FieldCodeRequired      Code = "REQUIRED"
	FieldCodeInvalidFormat Code = "INVALID_FORMAT"
	FieldCodeOutOfRange    Code = "OUT_OF_RANGE"
	FieldCodeInvalidEnum   Code = "INVALID_ENUM"
)
