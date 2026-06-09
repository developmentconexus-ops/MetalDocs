package http

import (
	"errors"
	"net/http"

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
)

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
	case errors.Is(err, domain.ErrStaleLockVersion):
		return http.StatusPreconditionFailed, codeTplStaleLockVersion
	case errors.Is(err, domain.ErrContentHashMismatch):
		return http.StatusConflict, codeTplContentHashMismatch
	case errors.Is(err, domain.ErrUploadMissing):
		return http.StatusConflict, codeTplUploadMissing
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
	case errors.Is(err, domain.ErrDuplicatePlaceholderName):
		return http.StatusUnprocessableEntity, codeTplDuplicatePlaceholder
	default:
		return http.StatusInternalServerError, codeTplInternalError
	}
}
