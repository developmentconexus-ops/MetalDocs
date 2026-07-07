package approvalhttp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	approvalapi "metaldocs/internal/modules/documents/approval/api"
	"metaldocs/internal/modules/documents/approval/application"
	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/modules/documents/approval/http/contracts"
	approvalsignature "metaldocs/internal/modules/documents/approval/infrastructure/signature"
	"metaldocs/internal/modules/documents/approval/infrastructure"
	v2dom "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/platform/idempotency"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/strictjson"
)

const internalErrorMessage = "internal error"

// Module-local typed codes for the approval domain (dot-notation taxonomy).
const (
	approvalCodeInternalUnknown                   problem.Code = "internal.unknown"
	approvalCodeConflictStaleRevision             problem.Code = "conflict.stale_revision"
	approvalCodeNotFoundInstance                  problem.Code = "not_found.instance"
	approvalCodeConflictDuplicate                 problem.Code = "conflict.duplicate_submission"
	approvalCodeSignoffDuplicate                  problem.Code = "signoff.duplicate"
	approvalCodePublishInvalidSupersede           problem.Code = "publish.invalid_supersede_target"
	approvalCodeStateInstanceCompleted            problem.Code = "state.instance_completed"
	approvalCodeRouteInUse                        problem.Code = "route.in_use"
	approvalCodeRouteDuplicateProfile             problem.Code = "route.duplicate_profile"
	approvalCodeSignoffNotEligible                problem.Code = "signoff.not_eligible"
	approvalCodeSodSubmitterCannotSign            problem.Code = "sod.submitter_cannot_sign"
	approvalCodeSodCrossStageDuplicate            problem.Code = "sod.cross_stage_duplicate"
	approvalCodeFreezeEffDateMissing              problem.Code = "freeze.effective_date_missing"
	approvalCodePreconditionIfMatch               problem.Code = "precondition.if_match_required"
	approvalCodeValidationIfMatchBad              problem.Code = "validation.if_match_malformed"
	approvalCodeIdempotencyRequired               problem.Code = "idempotency.key_required"
	approvalCodeIdempotencyKeyConflict            problem.Code = "idempotency.key_conflict"
	approvalCodePreconditionHashMismatch          problem.Code = "precondition.content_hash_mismatch"
	approvalCodeAuthnSignatureInvalid             problem.Code = "authn.signature_invalid"
	approvalCodeAuthnRateLimited                  problem.Code = "authn.rate_limited"
	approvalCodeInternalDBPrivilege               problem.Code = "internal.db_privilege_missing"
	approvalCodeInternalDBUnknown                 problem.Code = "internal.db_unknown"
	approvalCodeInternalSigMisconfigured          problem.Code = "internal.signature_misconfigured"
	approvalCodeValidationParamFormat             problem.Code = "validation.param_format"
	approvalCodeValidationParamUnmarshal          problem.Code = "validation.param_unmarshal"
	approvalCodeValidationParamRequired           problem.Code = "validation.param_required"
	approvalCodeValidationHeaderRequired          problem.Code = "validation.header_required"
	approvalCodeValidationParamTooMany            problem.Code = "validation.param_too_many_values"
	approvalCodeAuthzCapDenied                    problem.Code = "authz.capability_denied"
	approvalCodeApprovalUnresolved                problem.Code = "approval.unresolved_comments"
	approvalCodeValidationReasonRequired          problem.Code = "validation.reason_required"
	approvalCodeNotFoundRoute                     problem.Code = "not_found.route"
	approvalCodeStateRouteInactive                problem.Code = "state.route_inactive"
	approvalCodeTimeout                           problem.Code = "timeout"
	approvalCodeValidationJSONDecode              problem.Code = "validation.json_decode"
	approvalCodeValidationJSONTypeError           problem.Code = "validation.json_type_error"
	approvalCodeValidationEmptyBody               problem.Code = "validation.empty_body"
	approvalCodeValidationContentType             problem.Code = "validation.content_type"
	approvalCodeValidationBodyTooLarge            problem.Code = "validation.body_too_large"
	approvalCodeValidationRequestInvalid          problem.Code = "validation.request_invalid"
	approvalCodeValidationProfileUnknown          problem.Code = "validation.profile_unknown"
	approvalCodeValidationReasonForChangeRequired problem.Code = "validation.reason_for_change_required"
	approvalCodeValidationReasonCategoryInvalid   problem.Code = "validation.reason_category_invalid"
	approvalCodeValidationRevisionTitleRequired   problem.Code = "validation.revision_title_required"
	approvalCodeStateDocumentNotDraft             problem.Code = "state.document_not_draft"
	approvalCodeValidationProfileNotConfigured    problem.Code = "validation.profile_not_configured"
	approvalCodeStateApprovalRouteMissing         problem.Code = "state.approval_route_missing"
	approvalCodeNotFoundDocument                  problem.Code = "not_found.document"
	approvalCodeStateDocumentNotPublished          problem.Code = "state.document_not_published"
	approvalCodeConflictMarkReviewedStaleRevision  problem.Code = "conflict.mark_reviewed_stale_revision"
	approvalCodeValidationReviewDueBeforeEffective problem.Code = "validation.review_due_before_effective"
	approvalCodeValidationEffectiveToNotAfterFrom  problem.Code = "validation.effective_to_not_after_effective_from"
)

// ValidationError is a generic request-validation failure mapped to HTTP 400
// with approvalCodeValidationRequestInvalid by MapErrorToResponse.
type ValidationError struct {
	msg string
}

// NewValidationError constructs a ValidationError with the given message.
func NewValidationError(msg string) *ValidationError {
	return &ValidationError{msg: msg}
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return e.msg
}

// MapErrorToResponse translates a domain/repository/contract error into an RFC
// 9457 problem+json response, choosing the HTTP status and typed problem.Code
// that best matches err. Unrecognized errors fall back to a generic 500 with
// approvalCodeInternalUnknown — the client never sees internal error detail.
func MapErrorToResponse(err error) *problem.Problem {
	statusCode := http.StatusInternalServerError
	code := approvalCodeInternalUnknown

	switch {
	case errors.Is(err, infrastructure.ErrStaleRevision):
		statusCode = http.StatusConflict
		code = approvalCodeConflictStaleRevision
	case errors.Is(err, infrastructure.ErrNoActiveInstance):
		statusCode = http.StatusNotFound
		code = approvalCodeNotFoundInstance
	case errors.Is(err, infrastructure.ErrDuplicateSubmission):
		statusCode = http.StatusConflict
		code = approvalCodeConflictDuplicate
	case errors.Is(err, infrastructure.ErrActorAlreadySigned):
		statusCode = http.StatusConflict
		code = approvalCodeSignoffDuplicate
	case errors.Is(err, infrastructure.ErrInvalidScheduledSupersedeTarget):
		statusCode = http.StatusConflict
		code = approvalCodePublishInvalidSupersede
	case errors.Is(err, infrastructure.ErrInstanceCompleted):
		statusCode = http.StatusConflict
		code = approvalCodeStateInstanceCompleted
	case errors.Is(err, infrastructure.ErrRouteInUse):
		statusCode = http.StatusConflict
		code = approvalCodeRouteInUse
	case errors.Is(err, infrastructure.ErrDuplicateRouteProfile):
		statusCode = http.StatusConflict
		code = approvalCodeRouteDuplicateProfile
	case errors.Is(err, domain.ErrActorNotEligible):
		statusCode = http.StatusForbidden
		code = approvalCodeSignoffNotEligible
	case errors.Is(err, domain.ErrAuthorCannotSign):
		statusCode = http.StatusForbidden
		code = approvalCodeSodSubmitterCannotSign
	case errors.Is(err, domain.ErrActorAlreadySigned):
		statusCode = http.StatusForbidden
		code = approvalCodeSodCrossStageDuplicate
	case errors.Is(err, v2dom.ErrEffectiveDateMissing):
		statusCode = http.StatusUnprocessableEntity
		code = approvalCodeFreezeEffDateMissing
	case errors.Is(err, application.ErrReasonForChangeRequired):
		// F6.3 §5.3: REV>=1 submit without a structured reason is a friendly
		// first-line validation failure — 422 (semantically valid JSON,
		// business-rule rejection), mirroring the effective-date-missing case.
		statusCode = http.StatusUnprocessableEntity
		code = approvalCodeValidationReasonForChangeRequired
	case errors.Is(err, application.ErrReasonCategoryInvalid):
		statusCode = http.StatusUnprocessableEntity
		code = approvalCodeValidationReasonCategoryInvalid
	case errors.Is(err, application.ErrRevisionTitleRequired):
		// ADR 0073: canonical /submit now returns this (was finalize-only). REV>=1
		// submit without a revision title is a friendly business-rule rejection —
		// 422, mirroring the reason-for-change case above.
		statusCode = http.StatusUnprocessableEntity
		code = approvalCodeValidationRevisionTitleRequired
	case errors.Is(err, v2dom.ErrDocumentNotDraft):
		// In-tx submit resolution surfaces the finalize-era sentinels (ADR 0073).
		// Document not in draft = illegal state for the submit write → 409.
		statusCode = http.StatusConflict
		code = approvalCodeStateDocumentNotDraft
	case errors.Is(err, v2dom.ErrProfileNotConfigured):
		// Controlled document has no profile → the server cannot resolve a route.
		// Actionable request problem (finalize mapped it 400 ValidationError).
		statusCode = http.StatusBadRequest
		code = approvalCodeValidationProfileNotConfigured
	case errors.Is(err, v2dom.ErrApprovalRouteMissing):
		// No active approval route for the profile (finalize mapped it 409).
		statusCode = http.StatusConflict
		code = approvalCodeStateApprovalRouteMissing
	case errors.Is(err, ErrIfMatchRequired):
		statusCode = http.StatusPreconditionRequired
		code = approvalCodePreconditionIfMatch
	case errors.Is(err, ErrIfMatchMalformed):
		statusCode = http.StatusBadRequest
		code = approvalCodeValidationIfMatchBad
	case errors.Is(err, ErrIdempotencyRequired),
		errors.Is(err, application.ErrIdempotencyKeyRequired):
		statusCode = http.StatusBadRequest
		code = approvalCodeIdempotencyRequired
	case errors.Is(err, idempotency.ErrConflict):
		// Same Idempotency-Key reused with a different request fingerprint: the
		// caller must rotate the key for a genuinely new attempt.
		statusCode = http.StatusConflict
		code = approvalCodeIdempotencyKeyConflict
	case errors.Is(err, ErrContentHashMismatch):
		statusCode = http.StatusPreconditionFailed
		code = approvalCodePreconditionHashMismatch
	case errors.Is(err, approvalsignature.ErrInvalidCredentials):
		statusCode = http.StatusUnauthorized
		code = approvalCodeAuthnSignatureInvalid
	case errors.Is(err, approvalsignature.ErrRateLimited):
		statusCode = http.StatusTooManyRequests
		code = approvalCodeAuthnRateLimited
	case errors.Is(err, application.ErrReauthNotConfigured),
		errors.Is(err, approvalsignature.ErrUnknownSignatureMethod),
		errors.Is(err, approvalsignature.ErrRateLimiterConfig):
		// Signature-verifier misconfiguration (registry/limiter unwired). Unreachable
		// in production wiring; distinct 500 code so monitoring separates it from a
		// DB-layer 500. Client body stays the generic "internal error".
		statusCode = http.StatusInternalServerError
		code = approvalCodeInternalSigMisconfigured
	case errors.Is(err, infrastructure.ErrInsufficientPrivilege):
		statusCode = http.StatusInternalServerError
		code = approvalCodeInternalDBPrivilege
	case errors.Is(err, infrastructure.ErrUnknownDB):
		statusCode = http.StatusInternalServerError
		code = approvalCodeInternalDBUnknown
	case errors.Is(err, domain.ErrNoActiveStage):
		statusCode = http.StatusConflict
		code = approvalCodeStateInstanceCompleted
	case errors.Is(err, application.ErrDocumentNotFound):
		statusCode = http.StatusNotFound
		code = approvalCodeNotFoundDocument
	case errors.Is(err, application.ErrDocumentNotPublished):
		// Friendly first-line precondition (mark-reviewed requires published);
		// 409 mirrors the other illegal-state-for-write cases above (stale
		// revision / instance completed), not a validation 4xx.
		statusCode = http.StatusConflict
		code = approvalCodeStateDocumentNotPublished
	case errors.Is(err, application.ErrMarkReviewedStaleRevision):
		// 409, mirroring infrastructure.ErrStaleRevision above: the mark-reviewed
		// route's openapi response set is {400,401,403,404,409,428,500} — no
		// 412 — so the OCC conflict is a 409 Conflict, not 412.
		statusCode = http.StatusConflict
		code = approvalCodeConflictMarkReviewedStaleRevision
	case errors.Is(err, application.ErrReviewDueBeforeEffective):
		statusCode = http.StatusUnprocessableEntity
		code = approvalCodeValidationReviewDueBeforeEffective
	case errors.Is(err, application.ErrEffectiveToNotAfterEffectiveFrom):
		statusCode = http.StatusUnprocessableEntity
		code = approvalCodeValidationEffectiveToNotAfterFrom
	default:
		var capabilityDenied authz.ErrCapDenied
		var syntaxErr *json.SyntaxError
		var typeErr *json.UnmarshalTypeError
		var validationErr *ValidationError
		var invalidParamErr *approvalapi.InvalidParamFormatError
		var unmarshalParamErr *approvalapi.UnmarshalingParamError
		var requiredParamErr *approvalapi.RequiredParamError
		var requiredHeaderErr *approvalapi.RequiredHeaderError
		var tooManyValuesErr *approvalapi.TooManyValuesForParamError

		switch {
		case errors.As(err, &validationErr):
			statusCode = http.StatusBadRequest
			code = approvalCodeValidationRequestInvalid
		case errors.As(err, &invalidParamErr):
			statusCode = http.StatusBadRequest
			code = approvalCodeValidationParamFormat
		case errors.As(err, &unmarshalParamErr):
			statusCode = http.StatusBadRequest
			code = approvalCodeValidationParamUnmarshal
		case errors.As(err, &requiredParamErr):
			statusCode = http.StatusBadRequest
			code = approvalCodeValidationParamRequired
		case errors.As(err, &requiredHeaderErr):
			statusCode = http.StatusBadRequest
			code = approvalCodeValidationHeaderRequired
		case errors.As(err, &tooManyValuesErr):
			statusCode = http.StatusBadRequest
			code = approvalCodeValidationParamTooMany
		case errors.As(err, &capabilityDenied):
			statusCode = http.StatusForbidden
			code = approvalCodeAuthzCapDenied
		case errors.Is(err, application.ErrApprovalBlockedByUnresolvedComments):
			statusCode = http.StatusConflict
			code = approvalCodeApprovalUnresolved
		case errors.Is(err, application.ErrReasonRequired),
			errors.Is(err, application.ErrRouteDeactivateReasonRequired):
			statusCode = http.StatusBadRequest
			code = approvalCodeValidationReasonRequired
		case errors.Is(err, application.ErrInvalidObsoleteSource):
			statusCode = http.StatusBadRequest
			code = approvalCodeValidationRequestInvalid
		case errors.Is(err, application.ErrEffectiveDateInPast):
			statusCode = http.StatusBadRequest
			code = approvalCodeValidationRequestInvalid
		case errors.Is(err, application.ErrRouteProfileUnknown):
			// FK violation on (tenant_id, profile_code) → the document profile
			// does not exist for this tenant. Actionable 4xx, never a 500.
			statusCode = http.StatusUnprocessableEntity
			code = approvalCodeValidationProfileUnknown
		case errors.Is(err, application.ErrRouteNotFound):
			statusCode = http.StatusNotFound
			code = approvalCodeNotFoundRoute
		case errors.Is(err, application.ErrRouteAlreadyInactive):
			statusCode = http.StatusConflict
			code = approvalCodeStateRouteInactive
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			statusCode = http.StatusGatewayTimeout
			code = approvalCodeTimeout
		case errors.As(err, &syntaxErr):
			statusCode = http.StatusBadRequest
			code = approvalCodeValidationJSONDecode
		case errors.Is(err, io.ErrUnexpectedEOF):
			statusCode = http.StatusBadRequest
			code = approvalCodeValidationJSONDecode
		case err != nil && strings.HasPrefix(err.Error(), "json: unknown field"):
			statusCode = http.StatusBadRequest
			code = approvalCodeValidationJSONDecode
		case errors.As(err, &typeErr):
			statusCode = http.StatusBadRequest
			code = approvalCodeValidationJSONTypeError
		case errors.Is(err, io.EOF):
			statusCode = http.StatusBadRequest
			code = approvalCodeValidationEmptyBody
		case errors.Is(err, strictjson.ErrContentType):
			statusCode = http.StatusUnsupportedMediaType
			code = approvalCodeValidationContentType
		case errors.Is(err, strictjson.ErrBodyTooLarge):
			statusCode = http.StatusRequestEntityTooLarge
			code = approvalCodeValidationBodyTooLarge
		case errors.Is(err, strictjson.ErrEmptyBody):
			statusCode = http.StatusBadRequest
			code = approvalCodeValidationEmptyBody
		case errors.Is(err, contracts.ErrValidation):
			statusCode = http.StatusBadRequest
			code = approvalCodeValidationRequestInvalid
		}
	}

	return problem.New(statusCode, code, responseTitle(err, statusCode))
}

// WriteError maps err via MapErrorToResponse and writes it as a problem+json
// response. Any 5xx is also logged server-side, since the client body is
// intentionally generic and must not leak internal error detail.
func WriteError(w http.ResponseWriter, err error) {
	prob := MapErrorToResponse(err)
	// Never swallow server-side failures: the client gets a generic "internal
	// error" body, so the underlying cause must be logged for diagnosis.
	if prob.Status >= http.StatusInternalServerError {
		slog.Error("approval handler error",
			slog.Int("status", prob.Status),
			slog.String("code", string(prob.Code)),
			slog.Any("error", err),
		)
	}
	if writeErr := problem.Write(w, prob); writeErr != nil {
		WriteJSON(w, http.StatusInternalServerError, problem.New(http.StatusInternalServerError, approvalCodeInternalUnknown, internalErrorMessage))
	}
}

// WriteJSON marshals body and writes it with the given status code. If body
// fails to marshal, it falls back to a generic 500 internal-error payload
// rather than writing a broken response.
func WriteJSON(w http.ResponseWriter, status int, body any) {
	payload, err := json.Marshal(body)
	if err != nil {
		fallback := problem.New(http.StatusInternalServerError, approvalCodeInternalUnknown, internalErrorMessage)
		payload, err = json.Marshal(fallback)
		if err != nil {
			payload = []byte(`{"status":500,"title":"internal error"}`)
		}
		status = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(payload)
}

func responseTitle(err error, statusCode int) string {
	if statusCode >= http.StatusInternalServerError {
		return internalErrorMessage
	}
	if err == nil {
		return internalErrorMessage
	}
	return err.Error()
}
