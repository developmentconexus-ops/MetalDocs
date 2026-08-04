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

// Catalog aliases + the four codes templates genuinely owns.
//
// ADR 0089 execution step 11 (annex §2.3 + §2.8 rows #159, #161, #162, #163).
// Most bindings here are aliases onto the platform catalog and introduce no wire
// string of their own; the four problem.Register calls below are the conditions
// annex §3 ruled were being flattened onto a catalog code whose family or status
// contradicted them.
var (
	codeTplNotFound               = problem.CodeNotFoundResource
	codeTplKeyConflict            = problem.CodeConflictAlreadyExists
	codeTplInvalidStateTransition = problem.CodeStateTransitionInvalid
	codeTplStaleBase              = problem.CodeConflictStaleBase

	// annex #159 / R-7: domain.ErrStaleLockVersion is an OCC precondition on a
	// caller-SUPPLIED lock version, and this site has always answered 412 — but
	// it carried CONCURRENT_MODIFICATION, whose 409 documents still emits. The
	// name now matches the status class instead of contradicting it.
	codeTplStaleLockVersion = problem.Register("templates", "precondition.lock_version_stale", 412)

	// annex R-2 / C-17, ratified in ADR 0089 decision 3: the caller SUPPLIES the
	// expected hash and the server refuses because the world moved — textbook RFC
	// 9110 §15.5.13. This site answered 409 conflict.generic; it is now 412
	// precondition.content_hash_mismatch, the same code and status approval and
	// documents answer. BEHAVIOUR-VISIBLE: 409 -> 412.
	codeTplContentHashMismatch = problem.CodePreconditionContentHashMismatch

	codeTplUploadMissing   = problem.CodeStateUploadMissing
	codeTplUploadTooLarge  = problem.CodeRequestBodyTooLarge
	codeTplISOSegregation  = problem.CodePermissionISOSegregationViolation
	codeTplForbidden       = problem.CodePermissionDenied
	codeTplSystemImmutable = problem.CodeStateSystemTemplateImmutable
	codeTplArchived        = problem.CodeStateTransitionInvalid

	// annex #161 / R-6: these three sites answer 422 while carrying
	// VALIDATION_ERROR, whose registered default is 400 (request.invalid). One
	// code cannot default to both, so the semantic rejection gets its own.
	codeTplPlaceholderNameInvalid = problem.Register("templates", "validation.placeholder_name_invalid", 422)

	// annex #162 / R-18: the duplicate is inside the caller's OWN payload, so it
	// is a content defect (422), not a collision with stored state — which is
	// what ALREADY_EXISTS (409) claimed.
	codeTplDuplicatePlaceholder = problem.Register("templates", "validation.placeholder_name_duplicate", 422)

	codeTplInternalError = problem.CodeInternalUnknown

	// annex #163 / R-6: the ADR 0086 doc_type_code gate answered 422 on
	// VALIDATION_ERROR (400). codeTplInvalidRequest below stays bound to the
	// generic 400 code for the genuinely request-shaped rejection at
	// handler.go:207.
	codeTplDocTypeCodeRequired = problem.Register("templates", "validation.doc_type_code_required", 422)

	codeTplInvalidRequest = problem.CodeRequestInvalid
	codeTplInvalidBody    = problem.CodeRequestInvalid
	codeTplInvalidLimit   = problem.CodeRequestInvalid
	codeTplInvalidParam   = problem.CodeRequestInvalid

	// Approval-kernel entry point codes (M3 P3.S2b-4, R2a): the thin
	// submit-for-approval/signoff handlers delegate to
	// approval/application's kernel services and classify the published
	// (application-layer) error surface here — never approval/infrastructure,
	// per the module-boundary allow-model.
	codeTplApprovalRouteMissing = problem.CodeStateApprovalRouteMissing
	codeTplApprovalNotFound     = problem.CodeNotFoundResource

	// codeTplApprovalConflict was a single generic 409/422 bucket for eight
	// distinct approval-kernel conditions, which is exactly what made the FE
	// message map useless for template submit (annex R-10). Annex R-4 / R-14 /
	// C-10 / C-18 split it: templates adopts approval's codes verbatim, so the same
	// condition answers the same code and status from either surface. Only
	// approvalapp.ErrTemplateVersionNotDraft — a templates-lifecycle state, not an
	// approval one — keeps the generic 409.
	codeTplApprovalConflict = problem.CodeConflictGeneric

	// The eight adopted codes below are declared in the platform catalog's shared
	// block: approval owns the conditions, but templates may not import approval's
	// delivery package (module-boundary invariant, ADR 0082) and one wire string
	// may have only one registration.
	codeTplDuplicateSubmission       = problem.CodeConflictDuplicateSubmission
	codeTplApprovalInstanceCompleted = problem.CodeStateApprovalInstanceCompleted
	codeTplApprovalStageNotActive    = problem.CodeStateApprovalStageNotActive
	codeTplEmptyEligiblePool         = problem.CodeValidationEmptyEligiblePool
	codeTplSubmitChoiceRequired      = problem.CodeValidationSubmitChoiceRequired
	codeTplSubmitChoiceConstraint    = problem.CodeValidationSubmitChoiceConstraintViolated
	codeTplSignoffActorNotEligible   = problem.CodePermissionSignoffActorNotEligible
	codeTplSodSubmitterCannotSign    = problem.CodePermissionSodSubmitterCannotSign
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
		return http.StatusPreconditionFailed, codeTplContentHashMismatch
	case errors.Is(err, domain.ErrUploadMissing):
		return http.StatusConflict, codeTplUploadMissing
	case errors.Is(err, domain.ErrUploadTooLarge):
		return http.StatusRequestEntityTooLarge, codeTplUploadTooLarge
	case errors.Is(err, domain.ErrISOSegregationViolation):
		return http.StatusForbidden, codeTplISOSegregation
	case errors.As(err, new(iamauthz.ErrCapDenied)):
		return http.StatusForbidden, problem.CodePermissionCapabilityDenied
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
		return http.StatusUnprocessableEntity, codeTplDocTypeCodeRequired
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
		return http.StatusPreconditionFailed, codeTplContentHashMismatch
	case errors.Is(err, approvalapp.ErrIdempotencyKeyRequired):
		return http.StatusBadRequest, problem.CodeRequestIdempotencyKeyRequired
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
		return http.StatusForbidden, problem.CodePermissionCapabilityDenied
	case errors.Is(err, approvaldomain.ErrAuthorCannotSign):
		return http.StatusForbidden, problem.CodePermissionCapabilityDenied
	default:
		return http.StatusInternalServerError, codeTplInternalError
	}
}
