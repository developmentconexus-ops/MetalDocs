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

// Aliases onto the canonical catalog — this package introduces no wire strings
// of its own. ADR 0089 step 3: problem.Code is now a closed struct type issued
// only by the registry, so a catalog code is a `var`, and an alias to one cannot
// be `const`. Bindings are unchanged; annex §2.3 re-points them in step 11.
var (
	codeTplNotFound               = problem.CodeNotFound
	codeTplKeyConflict            = problem.CodeAlreadyExists
	codeTplInvalidStateTransition = problem.CodeStateTransitionInvalid
	codeTplStaleBase              = problem.CodeStaleBase
	codeTplStaleLockVersion       = problem.CodeConcurrentModification
	codeTplContentHashMismatch    = problem.CodeConflict
	codeTplUploadMissing          = problem.CodeUploadMissing
	codeTplUploadTooLarge         = problem.CodeRequestBodyTooLarge
	codeTplISOSegregation         = problem.CodeISOSegregationViolation
	codeTplForbidden              = problem.CodeAuthForbidden
	codeTplSystemImmutable        = problem.CodeSystemTemplateImmutable
	codeTplArchived               = problem.CodeStateTransitionInvalid
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
		return http.StatusOK, problem.Code{}
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
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden, codeTplForbidden
	case errors.Is(err, domain.ErrSystemTemplateImmutable):
		return http.StatusConflict, codeTplSystemImmutable
	case errors.Is(err, domain.ErrArchived):
		return http.StatusConflict, codeTplArchived
	case errors.Is(err, domain.ErrPlaceholderNameInvalid):
		return http.StatusUnprocessableEntity, codeTplPlaceholderNameInvalid
	case errors.Is(err, domain.ErrPlaceholderReservedName):
		return http.StatusUnprocessableEntity, codeTplPlaceholderNameInvalid
	case errors.Is(err, domain.ErrPlaceholderDictionaryInvalid):
		return http.StatusUnprocessableEntity, codeTplPlaceholderNameInvalid
	case errors.Is(err, domain.ErrDuplicatePlaceholderName):
		return http.StatusUnprocessableEntity, codeTplDuplicatePlaceholder
	case errors.Is(err, domain.ErrDocTypeCodeRequired):
		// ADR 0086: doc_type_code is mandatory (generic templates are gone).
		// 422 is declared on createTemplate for exactly this validation.
		return http.StatusUnprocessableEntity, codeTplInvalidRequest
	case errors.Is(err, domain.ErrApprovalRouteMissing):
		// ADR 0086 config-first creation gate — same 409 code the submit path
		// already raises when no active template route resolves.
		return http.StatusConflict, codeTplApprovalRouteMissing
	case errors.Is(err, approvalapp.ErrTemplateVersionNotFound):
		return http.StatusNotFound, codeTplApprovalNotFound
	case errors.Is(err, approvalapp.ErrTemplateVersionNotDraft):
		return http.StatusConflict, codeTplApprovalConflict
	case errors.Is(err, approvalapp.ErrTemplateVersionNoContent):
		// F-E4-1: submitting a version with no committed content_hash used to
		// reach the DB CHECK constraint and surface as a raw 500. It is a
		// missing-prerequisite conflict of exactly the same taxonomy family as
		// APPROVAL_ROUTE_MISSING, and the condition ("this version has no
		// committed DOCX content") is the one UPLOAD_MISSING already names —
		// including its friendly message (handler.go friendlyMsg) and the
		// frontend catalog entry. 409 is declared on the submit route; 422 is
		// not, so no spec change is involved.
		return http.StatusConflict, codeTplUploadMissing
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
	case errors.Is(err, approvaldomain.ErrSubmitChoiceRequired):
		// M4, unit 3.2, slice 5: symmetric with the document submit mapping
		// (approval/http/errors.go) — a submit_choice-governed stage has no
		// matching/non-empty chosen_actors entry.
		return http.StatusUnprocessableEntity, codeTplApprovalConflict
	case errors.Is(err, approvaldomain.ErrSubmitChoiceConstraintViolated):
		// M4, unit 3.2, slice 5: chosen user fails the role x area_code
		// constraint, or chosen_actors targets a non-submit_choice stage.
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
