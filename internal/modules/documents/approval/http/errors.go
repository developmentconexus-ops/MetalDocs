package approvalhttp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	approvalapi "metaldocs/internal/modules/documents/approval/api"
	"metaldocs/internal/modules/documents/approval/application"
	"metaldocs/internal/modules/documents/approval/domain"
	"metaldocs/internal/modules/documents/approval/http/contracts"
	approvalsignature "metaldocs/internal/modules/documents/approval/infrastructure/signature"
	"metaldocs/internal/modules/documents/approval/repository"
	v2dom "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/platform/problem"
)

const internalErrorMessage = "internal error"

// Module-local typed codes for the approval domain (dot-notation taxonomy).
const (
	approvalCodeInternalUnknown          problem.Code = "internal.unknown"
	approvalCodeConflictStaleRevision    problem.Code = "conflict.stale_revision"
	approvalCodeNotFoundInstance         problem.Code = "not_found.instance"
	approvalCodeConflictDuplicate        problem.Code = "conflict.duplicate_submission"
	approvalCodeSignoffDuplicate         problem.Code = "signoff.duplicate"
	approvalCodePublishInvalidSupersede  problem.Code = "publish.invalid_supersede_target"
	approvalCodeStateInstanceCompleted   problem.Code = "state.instance_completed"
	approvalCodeRouteInUse               problem.Code = "route.in_use"
	approvalCodeRouteDuplicateProfile    problem.Code = "route.duplicate_profile"
	approvalCodeSignoffNotEligible       problem.Code = "signoff.not_eligible"
	approvalCodeSodSubmitterCannotSign   problem.Code = "sod.submitter_cannot_sign"
	approvalCodeSodCrossStageDuplicate   problem.Code = "sod.cross_stage_duplicate"
	approvalCodeFreezeEffDateMissing     problem.Code = "freeze.effective_date_missing"
	approvalCodePreconditionIfMatch      problem.Code = "precondition.if_match_required"
	approvalCodeValidationIfMatchBad     problem.Code = "validation.if_match_malformed"
	approvalCodeIdempotencyRequired      problem.Code = "idempotency.key_required"
	approvalCodePreconditionHashMismatch problem.Code = "precondition.content_hash_mismatch"
	approvalCodeAuthnSignatureInvalid    problem.Code = "authn.signature_invalid"
	approvalCodeInternalDBPrivilege      problem.Code = "internal.db_privilege_missing"
	approvalCodeInternalDBUnknown        problem.Code = "internal.db_unknown"
	approvalCodeValidationParamFormat    problem.Code = "validation.param_format"
	approvalCodeValidationParamUnmarshal problem.Code = "validation.param_unmarshal"
	approvalCodeValidationParamRequired  problem.Code = "validation.param_required"
	approvalCodeValidationHeaderRequired problem.Code = "validation.header_required"
	approvalCodeValidationParamTooMany   problem.Code = "validation.param_too_many_values"
	approvalCodeAuthzCapDenied           problem.Code = "authz.capability_denied"
	approvalCodeApprovalUnresolved       problem.Code = "approval.unresolved_comments"
	approvalCodeValidationReasonRequired problem.Code = "validation.reason_required"
	approvalCodeNotFoundRoute            problem.Code = "not_found.route"
	approvalCodeTimeout                  problem.Code = "timeout"
	approvalCodeValidationJSONDecode     problem.Code = "validation.json_decode"
	approvalCodeValidationJSONTypeError  problem.Code = "validation.json_type_error"
	approvalCodeValidationEmptyBody      problem.Code = "validation.empty_body"
	approvalCodeValidationContentType    problem.Code = "validation.content_type"
	approvalCodeValidationBodyTooLarge   problem.Code = "validation.body_too_large"
	approvalCodeValidationRequestInvalid problem.Code = "validation.request_invalid"
)

type ValidationError struct {
	msg string
}

func NewValidationError(msg string) *ValidationError {
	return &ValidationError{msg: msg}
}

func (e *ValidationError) Error() string {
	return e.msg
}

func MapErrorToResponse(err error) *problem.Problem {
	statusCode := http.StatusInternalServerError
	code := approvalCodeInternalUnknown

	var capabilityDenied authz.ErrCapDenied
	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	var invalidParamErr *approvalapi.InvalidParamFormatError
	var unmarshalParamErr *approvalapi.UnmarshalingParamError
	var requiredParamErr *approvalapi.RequiredParamError
	var requiredHeaderErr *approvalapi.RequiredHeaderError
	var tooManyValuesErr *approvalapi.TooManyValuesForParamError
	var validationErr *ValidationError

	switch {
	case errors.Is(err, repository.ErrStaleRevision):
		statusCode = http.StatusConflict
		code = approvalCodeConflictStaleRevision
	case errors.Is(err, repository.ErrNoActiveInstance):
		statusCode = http.StatusNotFound
		code = approvalCodeNotFoundInstance
	case errors.Is(err, repository.ErrDuplicateSubmission):
		statusCode = http.StatusConflict
		code = approvalCodeConflictDuplicate
	case errors.Is(err, repository.ErrActorAlreadySigned):
		statusCode = http.StatusConflict
		code = approvalCodeSignoffDuplicate
	case errors.Is(err, repository.ErrInvalidScheduledSupersedeTarget):
		statusCode = http.StatusConflict
		code = approvalCodePublishInvalidSupersede
	case errors.Is(err, repository.ErrInstanceCompleted):
		statusCode = http.StatusConflict
		code = approvalCodeStateInstanceCompleted
	case errors.Is(err, repository.ErrRouteInUse):
		statusCode = http.StatusConflict
		code = approvalCodeRouteInUse
	case errors.Is(err, repository.ErrDuplicateRouteProfile):
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
	case errors.Is(err, ErrIfMatchRequired):
		statusCode = http.StatusPreconditionRequired
		code = approvalCodePreconditionIfMatch
	case errors.Is(err, ErrIfMatchMalformed):
		statusCode = http.StatusBadRequest
		code = approvalCodeValidationIfMatchBad
	case errors.Is(err, ErrIdempotencyRequired):
		statusCode = http.StatusBadRequest
		code = approvalCodeIdempotencyRequired
	case errors.Is(err, ErrContentHashMismatch):
		statusCode = http.StatusPreconditionFailed
		code = approvalCodePreconditionHashMismatch
	case errors.Is(err, approvalsignature.ErrInvalidCredentials):
		statusCode = http.StatusUnauthorized
		code = approvalCodeAuthnSignatureInvalid
	case errors.Is(err, repository.ErrInsufficientPrivilege):
		statusCode = http.StatusInternalServerError
		code = approvalCodeInternalDBPrivilege
	case errors.Is(err, repository.ErrUnknownDB):
		statusCode = http.StatusInternalServerError
		code = approvalCodeInternalDBUnknown
	case errors.Is(err, domain.ErrNoActiveStage):
		statusCode = http.StatusConflict
		code = approvalCodeStateInstanceCompleted
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
	case errors.Is(err, application.ErrReasonRequired):
		statusCode = http.StatusBadRequest
		code = approvalCodeValidationReasonRequired
	case errors.Is(err, application.ErrInvalidObsoleteSource):
		statusCode = http.StatusBadRequest
		code = approvalCodeValidationRequestInvalid
	case errors.Is(err, application.ErrEffectiveDateInPast):
		statusCode = http.StatusBadRequest
		code = approvalCodeValidationRequestInvalid
	case errors.Is(err, application.ErrRouteNotFound):
		statusCode = http.StatusNotFound
		code = approvalCodeNotFoundRoute
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		statusCode = http.StatusGatewayTimeout
		code = approvalCodeTimeout
	case errors.As(err, &syntaxErr):
		statusCode = http.StatusBadRequest
		code = approvalCodeValidationJSONDecode
	case errors.Is(err, io.ErrUnexpectedEOF):
		statusCode = http.StatusBadRequest
		code = approvalCodeValidationJSONDecode
	case errors.As(err, &typeErr):
		statusCode = http.StatusBadRequest
		code = approvalCodeValidationJSONTypeError
	case errors.Is(err, io.EOF):
		statusCode = http.StatusBadRequest
		code = approvalCodeValidationEmptyBody
	case errors.Is(err, contracts.ErrContentType):
		statusCode = http.StatusUnsupportedMediaType
		code = approvalCodeValidationContentType
	case errors.Is(err, contracts.ErrBodyTooLarge):
		statusCode = http.StatusRequestEntityTooLarge
		code = approvalCodeValidationBodyTooLarge
	case errors.Is(err, contracts.ErrEmptyBody):
		statusCode = http.StatusBadRequest
		code = approvalCodeValidationEmptyBody
	case errors.As(err, &validationErr):
		statusCode = http.StatusBadRequest
		code = approvalCodeValidationRequestInvalid
	}

	return problem.New(statusCode, code, responseTitle(err, statusCode))
}

func WriteError(w http.ResponseWriter, err error) {
	prob := MapErrorToResponse(err)
	if writeErr := problem.Write(w, prob); writeErr != nil {
		WriteJSON(w, http.StatusInternalServerError, problem.New(http.StatusInternalServerError, approvalCodeInternalUnknown, internalErrorMessage))
	}
}

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
