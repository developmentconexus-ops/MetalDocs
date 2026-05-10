// Package problem defines the canonical error code catalog used by
// MetalDocs Problem responses (RFC 9457). See spec §5.2.
package problem

// HTTP-status-level codes (used in Problem.code).
const (
	CodeValidationError        = "VALIDATION_ERROR"
	CodeUnknownField           = "UNKNOWN_FIELD"
	CodeUnknownFilter          = "UNKNOWN_FILTER"
	CodeInvalidSortField       = "INVALID_SORT_FIELD"
	CodeInvalidCursor          = "INVALID_CURSOR"
	CodeIncludeNotSupported    = "INCLUDE_NOT_SUPPORTED"
	CodeUnauthenticated        = "UNAUTHENTICATED"
	CodeForbiddenCapability    = "FORBIDDEN_CAPABILITY"
	CodeForbiddenArea          = "FORBIDDEN_AREA"
	CodeNotFound               = "NOT_FOUND"
	CodeAlreadyExists          = "ALREADY_EXISTS"
	CodeStateTransitionInvalid = "STATE_TRANSITION_INVALID"
	CodeConcurrentModification = "CONCURRENT_MODIFICATION"
	CodeIdempotencyKeyReused   = "IDEMPOTENCY_KEY_REUSED"
	CodeIdempotencyReplay      = "IDEMPOTENCY_REPLAY"
	CodeRateLimited            = "RATE_LIMITED" // reserved; not enforced yet
	CodeInternalError          = "INTERNAL_ERROR"
)

// Field-level codes (used in FieldError.code).
const (
	FieldCodeRequired      = "REQUIRED"
	FieldCodeInvalidFormat = "INVALID_FORMAT"
	FieldCodeOutOfRange    = "OUT_OF_RANGE"
	FieldCodeInvalidEnum   = "INVALID_ENUM"
)
