package approvalhttp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

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

func MapErrorToResponse(err error) *problem.Problem {
	statusCode := http.StatusInternalServerError
	code := "internal.unknown"

	switch {
	case errors.Is(err, repository.ErrStaleRevision):
		statusCode = http.StatusConflict
		code = "conflict.stale_revision"
	case errors.Is(err, repository.ErrNoActiveInstance):
		statusCode = http.StatusNotFound
		code = "not_found.instance"
	case errors.Is(err, repository.ErrDuplicateSubmission):
		statusCode = http.StatusConflict
		code = "conflict.duplicate_submission"
	case errors.Is(err, repository.ErrActorAlreadySigned):
		statusCode = http.StatusConflict
		code = "signoff.duplicate"
	case errors.Is(err, repository.ErrInstanceCompleted):
		statusCode = http.StatusConflict
		code = "state.instance_completed"
	case errors.Is(err, repository.ErrRouteInUse):
		statusCode = http.StatusConflict
		code = "route.in_use"
	case errors.Is(err, repository.ErrDuplicateRouteProfile):
		statusCode = http.StatusConflict
		code = "route.duplicate_profile"
	case errors.Is(err, domain.ErrActorNotEligible):
		statusCode = http.StatusForbidden
		code = "signoff.not_eligible"
	case errors.Is(err, domain.ErrAuthorCannotSign):
		statusCode = http.StatusForbidden
		code = "sod.submitter_cannot_sign"
	case errors.Is(err, domain.ErrActorAlreadySigned):
		statusCode = http.StatusForbidden
		code = "sod.cross_stage_duplicate"
	case errors.Is(err, v2dom.ErrEffectiveDateMissing):
		statusCode = http.StatusUnprocessableEntity
		code = "freeze.effective_date_missing"
	case errors.Is(err, ErrIfMatchRequired):
		statusCode = http.StatusPreconditionRequired
		code = "precondition.if_match_required"
	case errors.Is(err, ErrIfMatchMalformed):
		statusCode = http.StatusBadRequest
		code = "validation.if_match_malformed"
	case errors.Is(err, ErrIdempotencyRequired):
		statusCode = http.StatusBadRequest
		code = "idempotency.key_required"
	case errors.Is(err, ErrContentHashMismatch):
		statusCode = http.StatusPreconditionFailed
		code = "precondition.content_hash_mismatch"
	case errors.Is(err, approvalsignature.ErrInvalidCredentials):
		statusCode = http.StatusUnauthorized
		code = "authn.signature_invalid"
	case errors.Is(err, repository.ErrInsufficientPrivilege):
		statusCode = http.StatusInternalServerError
		code = "internal.db_privilege_missing"
	case errors.Is(err, repository.ErrUnknownDB):
		statusCode = http.StatusInternalServerError
		code = "internal.db_unknown"
	case errors.Is(err, domain.ErrNoActiveStage):
		statusCode = http.StatusConflict
		code = "state.instance_completed"
	default:
		var capabilityDenied authz.ErrCapDenied
		var syntaxErr *json.SyntaxError
		var typeErr *json.UnmarshalTypeError

		switch {
		case errors.As(err, &capabilityDenied):
			statusCode = http.StatusForbidden
			code = "authz.capability_denied"
		case errors.Is(err, application.ErrApprovalBlockedByUnresolvedComments):
			statusCode = http.StatusConflict
			code = "approval.unresolved_comments"
		case errors.Is(err, application.ErrReasonRequired):
			statusCode = http.StatusBadRequest
			code = "validation.reason_required"
		case errors.Is(err, application.ErrRouteNotFound):
			statusCode = http.StatusNotFound
			code = "not_found.route"
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			statusCode = http.StatusGatewayTimeout
			code = "timeout"
		case errors.As(err, &syntaxErr):
			statusCode = http.StatusBadRequest
			code = "validation.json_decode"
		case errors.Is(err, io.ErrUnexpectedEOF):
			statusCode = http.StatusBadRequest
			code = "validation.json_decode"
		case errors.As(err, &typeErr):
			statusCode = http.StatusBadRequest
			code = "validation.json_type_error"
		case errors.Is(err, io.EOF):
			statusCode = http.StatusBadRequest
			code = "validation.empty_body"
		case errors.Is(err, contracts.ErrContentType):
			statusCode = http.StatusUnsupportedMediaType
			code = "validation.content_type"
		case errors.Is(err, contracts.ErrBodyTooLarge):
			statusCode = http.StatusRequestEntityTooLarge
			code = "validation.body_too_large"
		case errors.Is(err, contracts.ErrEmptyBody):
			statusCode = http.StatusBadRequest
			code = "validation.empty_body"
		case errors.Is(err, contracts.ErrDuplicateKey):
			statusCode = http.StatusBadRequest
			code = "validation.duplicate_key"
		case looksLikeValidationError(err):
			statusCode = http.StatusBadRequest
			code = "validation.request_invalid"
		}
	}

	return problem.New(statusCode, code, responseTitle(err, statusCode))
}

func WriteError(w http.ResponseWriter, err error) {
	prob := MapErrorToResponse(err)
	if writeErr := problem.Write(w, prob); writeErr != nil {
		WriteJSON(w, http.StatusInternalServerError, problem.New(http.StatusInternalServerError, "internal.unknown", internalErrorMessage))
	}
}

func WriteJSON(w http.ResponseWriter, status int, body any) {
	payload, err := json.Marshal(body)
	if err != nil {
		fallback := problem.New(http.StatusInternalServerError, "internal.unknown", internalErrorMessage)
		payload, _ = json.Marshal(fallback)
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

func looksLikeValidationError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, " is required") ||
		strings.Contains(msg, " must be ") ||
		strings.Contains(msg, " must not be ") ||
		strings.HasPrefix(msg, "json: unknown field")
}
