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

// errMapping pairs an error-matching predicate with the HTTP status/problem
// code MapErr should answer when it matches. Order matters — mapErrTable is
// scanned top to bottom and the first match wins, exactly as the switch it
// replaces evaluated cases top to bottom.
type errMapping struct {
	match  func(error) bool
	status int
	code   problem.Code
}

func isErr(target error) func(error) bool {
	return func(err error) bool { return errors.Is(err, target) }
}

// mapErrTable is the templates module's domain/application-error → HTTP
// classification table. Kept as data (not a switch) so MapErr itself stays a
// single small dispatch loop — see gocyclo note below.
var mapErrTable = []errMapping{
	{isErr(domain.ErrNotFound), http.StatusNotFound, codeTplNotFound},
	{isErr(domain.ErrKeyConflict), http.StatusConflict, codeTplKeyConflict},
	{isErr(domain.ErrInvalidStateTransition), http.StatusConflict, codeTplInvalidStateTransition},
	{isErr(domain.ErrStaleBase), http.StatusConflict, codeTplStaleBase},
	{isErr(domain.ErrConcurrentTransition), http.StatusConflict, codeTplInvalidStateTransition},
	{isErr(domain.ErrStaleLockVersion), http.StatusPreconditionFailed, codeTplStaleLockVersion},
	{isErr(domain.ErrContentHashMismatch), http.StatusPreconditionFailed, codeTplContentHashMismatch},
	{isErr(domain.ErrUploadMissing), http.StatusConflict, codeTplUploadMissing},
	{isErr(domain.ErrUploadTooLarge), http.StatusRequestEntityTooLarge, codeTplUploadTooLarge},
	{isErr(domain.ErrISOSegregationViolation), http.StatusForbidden, codeTplISOSegregation},
	{func(err error) bool { return errors.As(err, new(iamauthz.ErrCapDenied)) }, http.StatusForbidden, problem.CodePermissionCapabilityDenied},
	{isErr(domain.ErrForbidden), http.StatusForbidden, codeTplForbidden},
	{isErr(domain.ErrSystemTemplateImmutable), http.StatusConflict, codeTplSystemImmutable},
	{isErr(domain.ErrArchived), http.StatusConflict, codeTplArchived},
	{isErr(domain.ErrPlaceholderNameInvalid), http.StatusUnprocessableEntity, codeTplPlaceholderNameInvalid},
	{isErr(domain.ErrPlaceholderReservedName), http.StatusUnprocessableEntity, codeTplPlaceholderNameInvalid},
	{isErr(domain.ErrPlaceholderDictionaryInvalid), http.StatusUnprocessableEntity, codeTplPlaceholderNameInvalid},
	{isErr(domain.ErrDuplicatePlaceholderName), http.StatusUnprocessableEntity, codeTplDuplicatePlaceholder},
	// ADR 0086: doc_type_code is mandatory (generic templates are gone). 422
	// is declared on createTemplate for exactly this validation.
	{isErr(domain.ErrDocTypeCodeRequired), http.StatusUnprocessableEntity, codeTplDocTypeCodeRequired},
	// ADR 0086 config-first creation gate — same 409 code the submit path
	// already raises when no active template route resolves.
	{isErr(domain.ErrApprovalRouteMissing), http.StatusConflict, codeTplApprovalRouteMissing},
	{isErr(approvalapp.ErrTemplateVersionNotFound), http.StatusNotFound, codeTplApprovalNotFound},
	{isErr(approvalapp.ErrTemplateVersionNotDraft), http.StatusConflict, codeTplApprovalConflict},
	// F-E4-1: submitting a version with no committed content_hash used to
	// reach the DB CHECK constraint and surface as a raw 500. It is a
	// missing-prerequisite conflict of exactly the same taxonomy family as
	// APPROVAL_ROUTE_MISSING, and the condition ("this version has no
	// committed DOCX content") is the one UPLOAD_MISSING already names —
	// including its friendly message (handler.go friendlyMsg) and the
	// frontend catalog entry. 409 is declared on the submit route; 422 is
	// not, so no spec change is involved.
	{isErr(approvalapp.ErrTemplateVersionNoContent), http.StatusConflict, codeTplUploadMissing},
	{isErr(approvalapp.ErrNoActiveApprovalRoute), http.StatusConflict, codeTplApprovalRouteMissing},
	{isErr(approvalapp.ErrDuplicateSubmission), http.StatusConflict, codeTplDuplicateSubmission},
	{isErr(approvalapp.ErrNoActiveInstance), http.StatusNotFound, codeTplApprovalNotFound},
	{isErr(approvalapp.ErrInstanceCompleted), http.StatusConflict, codeTplApprovalInstanceCompleted},
	{isErr(approvalapp.ErrStageNotActive), http.StatusConflict, codeTplApprovalStageNotActive},
	{isErr(approvalapp.ErrContentHashMismatch), http.StatusPreconditionFailed, codeTplContentHashMismatch},
	{isErr(approvalapp.ErrIdempotencyKeyRequired), http.StatusBadRequest, problem.CodeRequestIdempotencyKeyRequired},
	{isErr(approvaldomain.ErrEmptyEligiblePool), http.StatusUnprocessableEntity, codeTplEmptyEligiblePool},
	// M4, unit 3.2, slice 5: symmetric with the document submit mapping
	// (approval/http/errors.go) — a submit_choice-governed stage has no
	// matching/non-empty chosen_actors entry.
	{isErr(approvaldomain.ErrSubmitChoiceRequired), http.StatusUnprocessableEntity, codeTplSubmitChoiceRequired},
	// M4, unit 3.2, slice 5: chosen user fails the role x area_code
	// constraint, or chosen_actors targets a non-submit_choice stage.
	{isErr(approvaldomain.ErrSubmitChoiceConstraintViolated), http.StatusUnprocessableEntity, codeTplSubmitChoiceConstraint},
	{isErr(approvaldomain.ErrNoActiveStage), http.StatusConflict, codeTplApprovalConflict},
	{isErr(approvaldomain.ErrActorNotEligible), http.StatusForbidden, codeTplSignoffActorNotEligible},
	{isErr(approvaldomain.ErrAuthorCannotSign), http.StatusForbidden, codeTplSodSubmitterCannotSign},
}

// MapErr translates a domain/application error from the templates module into
// the HTTP status and problem+json code that should be returned to the
// caller. A nil err maps to 200 OK with an empty code; unrecognized errors
// fall back to 500 Internal Server Error. Matching walks mapErrTable in
// order and returns on the first hit, preserving the precedence the
// original switch encoded (some conditions satisfy more than one
// errors.Is/As check, so order is significant).
func MapErr(err error) (httpStatus int, code problem.Code) {
	if err == nil {
		return http.StatusOK, problem.Code{}
	}
	for _, m := range mapErrTable {
		if m.match(err) {
			return m.status, m.code
		}
	}
	return http.StatusInternalServerError, codeTplInternalError
}
