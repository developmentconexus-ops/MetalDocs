package http

import (
	"errors"
	"net/http"

	approvalapp "metaldocs/internal/modules/approval/application"
	approvaldomain "metaldocs/internal/modules/approval/domain"
	iamauthz "metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/modules/templates/domain"
	"metaldocs/internal/platform/problem"
)

const (
	codeTplNotFound               = problem.CodeNotFound
	codeTplKeyConflict            = problem.CodeAlreadyExists
	codeTplInvalidStateTransition = problem.CodeStateTransitionInvalid
	codeTplStaleBase              = problem.CodeStaleBase
	codeTplStaleLockVersion       = problem.CodeConcurrentModification
	codeTplContentHashMismatch    = problem.CodeConflict
	codeTplUploadMissing          = problem.CodeUploadMissing
	codeTplUploadTooLarge         = problem.CodeRequestBodyTooLarge
	codeTplISOSegregation         = problem.CodeISOSegregationViolation
	codeTplForbiddenRole          = problem.CodeForbiddenCapability
	codeTplForbidden              = problem.CodeAuthForbidden
	codeTplSystemImmutable        = problem.CodeSystemTemplateImmutable
	codeTplArchived               = problem.CodeStateTransitionInvalid
	codeTplInvalidApprovalConfig  = problem.CodeValidationError
	codeTplPlaceholderNameInvalid = problem.CodeValidationError
	codeTplDuplicatePlaceholder   = problem.CodeAlreadyExists
	codeTplInternalError          = problem.CodeInternalError
	codeTplInvalidRequest         = problem.CodeValidationError
	codeTplInvalidBody            = problem.CodeValidationError
	codeTplInvalidLimit           = problem.CodeValidationError
	codeTplInvalidParam           = problem.CodeValidationError

	// Approval-kernel entry point codes (M3 P3.S2b-4, R2a): the thin
	// submit-for-approval/signoff handlers delegate to
	// approval/application's kernel services and classify the published
	// (application-layer) error surface here — never approval/infrastructure,
	// per the module-boundary allow-model.
	codeTplApprovalRouteMissing = problem.CodeApprovalRouteMissing
	codeTplApprovalConflict     = problem.CodeConflict
	codeTplApprovalNotFound     = problem.CodeNotFound
)

// MapErr translates a domain/application error from the templates module into
// the HTTP status and problem+json code that should be returned to the
// caller. A nil err maps to 200 OK with an empty code; unrecognized errors
// fall back to 500 Internal Server Error.
func MapErr(err error) (httpStatus int, code problem.Code) {
	switch {
	case err == nil:
		return http.StatusOK, ""
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, codeTplNotFound
	case errors.Is(err, domain.ErrKeyConflict):
		return http.StatusConflict, codeTplKeyConflict
	case errors.Is(err, domain.ErrInvalidStateTransition):
		return http.StatusConflict, codeTplInvalidStateTransition
	case errors.Is(err, domain.ErrStaleBase):
		return http.StatusConflict, codeTplStaleBase
	case errors.Is(err, domain.ErrConcurrentTransition):
		return http.StatusConflict, codeTplInvalidStateTransition
	case errors.Is(err, domain.ErrStaleLockVersion):
		return http.StatusPreconditionFailed, codeTplStaleLockVersion
	case errors.Is(err, domain.ErrContentHashMismatch):
		return http.StatusConflict, codeTplContentHashMismatch
	case errors.Is(err, domain.ErrUploadMissing):
		return http.StatusConflict, codeTplUploadMissing
	case errors.Is(err, domain.ErrUploadTooLarge):
		return http.StatusRequestEntityTooLarge, codeTplUploadTooLarge
	case errors.Is(err, domain.ErrISOSegregationViolation):
		return http.StatusForbidden, codeTplISOSegregation
	case errors.As(err, new(iamauthz.ErrCapDenied)):
		return http.StatusForbidden, problem.CodeForbiddenCapability
	case errors.Is(err, domain.ErrForbiddenRole):
		return http.StatusForbidden, codeTplForbiddenRole
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden, codeTplForbidden
	case errors.Is(err, domain.ErrSystemTemplateImmutable):
		return http.StatusConflict, codeTplSystemImmutable
	case errors.Is(err, domain.ErrArchived):
		return http.StatusConflict, codeTplArchived
	case errors.Is(err, domain.ErrInvalidApprovalConfig):
		return http.StatusBadRequest, codeTplInvalidApprovalConfig
	case errors.Is(err, domain.ErrPlaceholderNameInvalid):
		return http.StatusUnprocessableEntity, codeTplPlaceholderNameInvalid
	case errors.Is(err, domain.ErrPlaceholderReservedName):
		return http.StatusUnprocessableEntity, codeTplPlaceholderNameInvalid
	case errors.Is(err, domain.ErrPlaceholderDictionaryInvalid):
		return http.StatusUnprocessableEntity, codeTplPlaceholderNameInvalid
	case errors.Is(err, domain.ErrDuplicatePlaceholderName):
		return http.StatusUnprocessableEntity, codeTplDuplicatePlaceholder
	case errors.Is(err, approvalapp.ErrTemplateVersionNotFound):
		return http.StatusNotFound, codeTplApprovalNotFound
	case errors.Is(err, approvalapp.ErrTemplateVersionNotDraft):
		return http.StatusConflict, codeTplApprovalConflict
	case errors.Is(err, approvalapp.ErrNoActiveApprovalRoute):
		return http.StatusConflict, codeTplApprovalRouteMissing
	case errors.Is(err, approvalapp.ErrDuplicateSubmission):
		return http.StatusConflict, codeTplApprovalConflict
	case errors.Is(err, approvalapp.ErrNoActiveInstance):
		return http.StatusNotFound, codeTplApprovalNotFound
	case errors.Is(err, approvalapp.ErrInstanceCompleted):
		return http.StatusConflict, codeTplApprovalConflict
	case errors.Is(err, approvalapp.ErrStageNotActive):
		return http.StatusConflict, codeTplApprovalConflict
	case errors.Is(err, approvalapp.ErrContentHashMismatch):
		return http.StatusConflict, codeTplContentHashMismatch
	case errors.Is(err, approvalapp.ErrIdempotencyKeyRequired):
		return http.StatusBadRequest, problem.CodeIdempotencyKeyRequired
	case errors.Is(err, approvaldomain.ErrEmptyEligiblePool):
		return http.StatusUnprocessableEntity, codeTplApprovalConflict
	case errors.Is(err, approvaldomain.ErrNoActiveStage):
		return http.StatusConflict, codeTplApprovalConflict
	case errors.Is(err, approvaldomain.ErrActorNotEligible):
		return http.StatusForbidden, problem.CodeForbiddenCapability
	case errors.Is(err, approvaldomain.ErrAuthorCannotSign):
		return http.StatusForbidden, problem.CodeForbiddenCapability
	default:
		return http.StatusInternalServerError, codeTplInternalError
	}
}
