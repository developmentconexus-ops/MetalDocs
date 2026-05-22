package http

import (
	"errors"
	"net/http"

	"metaldocs/internal/modules/templates/domain"
	"metaldocs/internal/platform/problem"
)

const (
	codeTplNotFound               problem.Code = "not_found"
	codeTplKeyConflict            problem.Code = "key_conflict"
	codeTplInvalidStateTransition problem.Code = "invalid_state_transition"
	codeTplStaleBase              problem.Code = "stale_base"
	codeTplStaleLockVersion       problem.Code = "stale_lock_version"
	codeTplContentHashMismatch    problem.Code = "content_hash_mismatch"
	codeTplUploadMissing          problem.Code = "upload_missing"
	codeTplISOSegregation         problem.Code = "iso_segregation_violation"
	codeTplForbiddenRole          problem.Code = "forbidden_role"
	codeTplForbidden              problem.Code = "forbidden"
	codeTplSystemImmutable        problem.Code = "SYSTEM_TEMPLATE_IMMUTABLE"
	codeTplArchived               problem.Code = "archived"
	codeTplInvalidApprovalConfig  problem.Code = "invalid_approval_config"
	codeTplPlaceholderNameInvalid problem.Code = "placeholder_name_invalid"
	codeTplDuplicatePlaceholder   problem.Code = "duplicate_placeholder_name"
	codeTplInternalError          problem.Code = "internal_error"
	codeTplInvalidRequest         problem.Code = "invalid_request"
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
