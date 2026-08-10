package approvalhttp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	approvalapi "metaldocs/internal/modules/approval/api"
	"metaldocs/internal/modules/approval/application"
	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/http/contracts"
	"metaldocs/internal/modules/approval/infrastructure"
	approvalsignature "metaldocs/internal/modules/approval/infrastructure/signature"
	v2dom "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/platform/idempotency"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/strictjson"
)

const internalErrorMessage = "internal error"

// Module-owned problem codes for the approval bounded context.
//
// ADR 0089 execution step 6 (annex §2.2, rows #48-#115): every RegisterLegacy is
// gone. Each code now carries a semantic family prefix from the closed set and
// its status is BOUND to the code, so the same condition cannot answer 409 here
// and 412 elsewhere. Families are never module-named — the wire code tells the
// client what to do, and `signoff.` / `sod.` / `route.` / `approval.` / `submit.`
// / `freeze.` became lies the moment approval was extracted to a top-level module
// (ADR 0082). They are re-homed per annex §1.3.
//
// Codes bound to a problem.Code* platform var are wire strings shared with
// another bounded context (annex §2.9 collision classes C-1, C-3, C-5, C-6, C-7,
// C-8, C-9, C-13, C-16); the registry's duplicate guard forbids declaring them
// twice, and a module may not import another module's delivery package, so the
// platform catalog is their single legal declaration site.
var (
	approvalCodeInternalUnknown       = problem.CodeInternalUnknown
	approvalCodeConflictStaleRevision = problem.Register("approval", "conflict.stale_revision", 409)
	approvalCodeNotFoundInstance      = problem.Register("approval", "notfound.approval_instance", 404)
	// approvalCodeNotFoundInstanceNotVisible (F8, spec.md §6.3) is the
	// DISTINCT 404 code for infrastructure.ErrInstanceNotVisible —
	// deliberately its own code (not approvalCodeNotFoundInstance) so
	// monitoring/logs can tell "instance genuinely does not exist" apart from
	// "instance exists but is outside this actor's visibility boundary",
	// even though both return the same 404 status to the client (the client
	// response body must not leak which case it is).
	approvalCodeNotFoundInstanceNotVisible = problem.Register("approval", "notfound.approval_instance_not_visible", 404)
	approvalCodeConflictDuplicate          = problem.CodeConflictDuplicateSubmission
	approvalCodeSignoffDuplicate           = problem.Register("approval", "conflict.signoff_duplicate", 409)

	// annex R-13: the caller supplied a supersede target id that fails a business
	// rule. No race is asserted, so 409 was wrong — 422 (RFC 9110 §15.5.21).
	approvalCodeValidationSupersedeTargetInvalid = problem.Register("approval", "validation.supersede_target_invalid", 422)

	approvalCodeStateInstanceCompleted = problem.CodeStateApprovalInstanceCompleted

	// annex R-14 / §2.8 #164: "no active stage" and "instance completed" are
	// different conditions with different operator remedies, so domain.ErrNoActiveStage
	// no longer folds onto state.approval_instance_completed.
	approvalCodeStateApprovalStageNotActive = problem.CodeStateApprovalStageNotActive

	approvalCodeRouteInUse             = problem.Register("approval", "state.approval_route_in_use", 409)
	approvalCodeRouteDuplicateProfile  = problem.Register("approval", "conflict.approval_route_duplicate_profile", 409)
	approvalCodeSignoffNotEligible     = problem.CodePermissionSignoffActorNotEligible
	approvalCodeSodSubmitterCannotSign = problem.CodePermissionSodSubmitterCannotSign
	approvalCodeSodCrossStageDuplicate = problem.Register("approval", "permission.sod_cross_stage_duplicate", 403)
	approvalCodeFreezeEffDateMissing   = problem.Register("approval", "validation.effective_date_required", 422)

	approvalCodePreconditionIfMatch = problem.RegisterWithStatus("approval", "precondition.if_match_required", 428,
		"428 Precondition Required (RFC 6585 §3): no precondition FAILED — the server refuses to act "+
			"without one, which is the opposite of the 412 the precondition. family defaults to")

	approvalCodeValidationIfMatchBad = problem.Register("approval", "request.if_match_malformed", 400)

	// C-8 / C-13: the Idempotency-Key contract is a platform contract, not an
	// approval dialect — a caller must see the same code whether or not the route
	// is wrapped by the idempotency middleware.
	approvalCodeIdempotencyRequired    = problem.CodeRequestIdempotencyKeyRequired
	approvalCodeIdempotencyKeyConflict = problem.CodeConflictIdempotencyKeyReused

	approvalCodePreconditionHashMismatch = problem.CodePreconditionContentHashMismatch
	approvalCodeAuthnSignatureInvalid    = problem.Register("approval", "auth.signature_invalid", 401)

	// C-6: a throttle is a throttle. The signature re-auth limiter emits the one
	// platform ratelimit code rather than an authn.* dialect.
	approvalCodeAuthnRateLimited = problem.CodeRateLimitExceeded

	approvalCodeInternalDBPrivilege      = problem.Register("approval", "internal.db_privilege_missing", 500)
	approvalCodeInternalDBUnknown        = problem.Register("approval", "internal.db_unknown", 500)
	approvalCodeInternalSigMisconfigured = problem.Register("approval", "internal.signature_misconfigured", 500)

	// The five oapi-codegen parameter/header binding failures are request-SHAPE
	// defects (annex §1.4 rule 1), not value rejections — hence request., not
	// validation.
	approvalCodeValidationParamFormat    = problem.Register("approval", "request.param_format", 400)
	approvalCodeValidationParamUnmarshal = problem.Register("approval", "request.param_unmarshal", 400)
	approvalCodeValidationParamRequired  = problem.Register("approval", "request.param_required", 400)
	approvalCodeValidationHeaderRequired = problem.Register("approval", "request.header_required", 400)
	approvalCodeValidationParamTooMany   = problem.Register("approval", "request.param_too_many_values", 400)

	approvalCodeAuthzCapDenied     = problem.CodePermissionCapabilityDenied
	approvalCodeApprovalUnresolved = problem.Register("approval", "state.approval_blocked_unresolved_comments", 409)

	// annex R-15: the sibling reason codes (validation.reason_for_change_required,
	// validation.reason_category_invalid) are already 422; 400 vs 422 for the same
	// field class was incoherent.
	approvalCodeValidationReasonRequired = problem.Register("approval", "validation.reason_required", 422)

	approvalCodeNotFoundRoute      = problem.Register("approval", "notfound.approval_route", 404)
	approvalCodeStateRouteInactive = problem.Register("approval", "state.approval_route_inactive", 409)

	approvalCodeTimeout = problem.RegisterWithStatus("approval", "internal.upstream_timeout", 504,
		"504 Gateway Timeout: the deadline was exceeded waiting on a dependency, which is a narrower "+
			"server fault than the internal. family default of 500")

	approvalCodeValidationJSONDecode    = problem.CodeRequestJSONDecode
	approvalCodeValidationJSONTypeError = problem.Register("approval", "request.json_type_error", 400)
	approvalCodeValidationEmptyBody     = problem.CodeRequestEmptyBody

	approvalCodeValidationContentType = problem.CodeRequestContentTypeUnsupported

	// C-5 / C-1: both collapse onto platform codes.
	approvalCodeValidationBodyTooLarge   = problem.CodeRequestBodyTooLarge
	approvalCodeValidationRequestInvalid = problem.CodeRequestInvalid

	approvalCodeValidationProfileUnknown             = problem.Register("approval", "validation.profile_unknown", 422)
	approvalCodeValidationReasonForChangeRequired    = problem.Register("approval", "validation.reason_for_change_required", 422)
	approvalCodeValidationReasonCategoryInvalid      = problem.Register("approval", "validation.reason_category_invalid", 422)
	approvalCodeValidationRevisionTitleRequired      = problem.Register("approval", "validation.revision_title_required", 422)
	approvalCodeValidationDocumentSubjectKeyMismatch = problem.Register("approval", "validation.document_subject_key_mismatch", 422)
	approvalCodeValidationTemplateSubjectKeyMismatch = problem.Register("approval", "validation.template_subject_key_mismatch", 422)
	approvalCodeStateDocumentNotDraft                = problem.Register("approval", "state.document_not_draft", 409)

	// annex R-16: same class as validation.profile_unknown (422); the 400 was
	// inherited from the removed finalize handler.
	approvalCodeValidationProfileNotConfigured = problem.Register("approval", "validation.profile_not_configured", 422)

	approvalCodeStateApprovalRouteMissing          = problem.CodeStateApprovalRouteMissing
	approvalCodeNotFoundDocument                   = problem.Register("approval", "notfound.document", 404)
	approvalCodeStateDocumentNotPublished          = problem.Register("approval", "state.document_not_published", 409)
	approvalCodeConflictMarkReviewedStaleRevision  = problem.Register("approval", "conflict.mark_reviewed_stale_revision", 409)
	approvalCodeValidationReviewDueBeforeEffective = problem.Register("approval", "validation.review_due_before_effective", 422)
	approvalCodeValidationEffectiveToNotAfterFrom  = problem.Register("approval", "validation.effective_to_not_after_effective_from", 422)
	approvalCodeValidationEmptyEligiblePool        = problem.CodeValidationEmptyEligiblePool
	approvalCodeValidationSubmitChoiceRequired     = problem.CodeValidationSubmitChoiceRequired
	approvalCodeValidationSubmitChoiceConstraint   = problem.CodeValidationSubmitChoiceConstraintViolated

	// Route-shape policy rejections (per-profile governance policy, G1/ADR 0081,
	// livre arm superseded by ADR 0087). All 422: the request is well-formed,
	// the resulting route shape is not permitted for the profile's class.
	approvalCodeValidationRouteStagesNotPermitted = problem.Register("approval", "validation.route_stages_not_permitted", 422)
	approvalCodeValidationApprovalStageRequired   = problem.Register("approval", "validation.approval_stage_required", 422)
	approvalCodeValidationRouteStageRequired      = problem.Register("approval", "validation.route_stage_required", 422)

	// F9/ADR 0077 — approval delegation.
	approvalCodeValidationSelfDelegation   = problem.Register("approval", "validation.self_delegation", 422)
	approvalCodeValidationDelegationWindow = problem.Register("approval", "validation.delegation_window_invalid", 422)
	approvalCodeNotFoundDelegation         = problem.Register("approval", "notfound.delegation", 404)

	// R3/G2 — review-verdict stage-kind guard. annex R-17: the old state. prefix
	// contradicted both the 422 and the condition, which rejects a SUPPLIED
	// verdict rather than reporting a lifecycle state.
	approvalCodeStateVerdictReadyOnApprovalStage = problem.Register("approval", "validation.verdict_ready_on_approval_stage", 422)
	approvalCodeInternalVerdictWrongStageKind    = problem.Register("approval", "internal.verdict_wrong_stage_kind", 500)

	// R5/unit 2.3 G3 — fast-forward ("Aprovar já") composition guards.
	approvalCodeStateFastForwardStageNotCompleted = problem.Register("approval", "state.fast_forward_stage_not_completed", 409)
	approvalCodeStateFastForwardNotEligible       = problem.Register("approval", "state.fast_forward_not_eligible", 409)

	// Task 9 — SLA extension ("extend one instance's deadline without a
	// second route"). Three distinct conditions get three distinct codes: a
	// backwards/non-forward request is a business-rule rejection of an
	// otherwise well-formed request (422); a missing active stage/deadline is
	// an illegal state for the write (409); a blank reason is a friendly
	// first-line validation failure (422), mirroring the module's other
	// reason-required cases.
	approvalCodeSLANotForward     = problem.Register("approval", "validation.sla_extension_not_forward", 422)
	approvalCodeSLANoActiveStage  = problem.Register("approval", "state.sla_extension_no_active_stage", 409)
	approvalCodeSLAReasonRequired = problem.Register("approval", "validation.sla_extension_reason_required", 422)
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

// errorMapping is one entry of approvalErrorMappings: a predicate over err
// plus the HTTP status and problem.Code to use when it matches.
type errorMapping struct {
	match  func(err error) bool
	status int
	code   problem.Code
}

// matchIs reports errors.Is(err, target) for any target in targets — an
// error matches the mapping if it matches ANY of them (mirrors a switch case
// with multiple comma-separated conditions sharing one body).
func matchIs(targets ...error) func(error) bool {
	return func(err error) bool {
		for _, target := range targets {
			if errors.Is(err, target) {
				return true
			}
		}
		return false
	}
}

// matchAs reports errors.As(err, &target) for a fresh zero-value T — the
// generic counterpart of a `var x T; errors.As(err, &x)` case clause.
func matchAs[T error]() func(error) bool {
	return func(err error) bool {
		var target T
		return errors.As(err, &target)
	}
}

// approvalErrorMappings is the ordered dispatch table MapErrorToResponse
// evaluates top to bottom, taking the status/code of the first match — the
// flat equivalent of the switch statement this table replaced (a plain
// outer `switch { case errors.Is(...): ... }` falls through to its default
// only when nothing above matched, which is exactly the same "first match
// wins across everything" semantics as one ordered list). Order matters:
// entries are transcribed in their original evaluation order and must not be
// reordered.
var approvalErrorMappings = []errorMapping{
	{matchIs(infrastructure.ErrStaleRevision), http.StatusConflict, approvalCodeConflictStaleRevision},
	{matchIs(infrastructure.ErrNoActiveInstance), http.StatusNotFound, approvalCodeNotFoundInstance},
	// F8, spec.md §6.3: cross-boundary = not-found. Same 404 status as
	// ErrNoActiveInstance but a distinct problem.Code (see the constant's doc
	// comment) so server-side logs/monitoring can distinguish the two cases
	// without the client-visible response revealing which one fired.
	{matchIs(infrastructure.ErrInstanceNotVisible), http.StatusNotFound, approvalCodeNotFoundInstanceNotVisible},
	{matchIs(infrastructure.ErrDuplicateSubmission), http.StatusConflict, approvalCodeConflictDuplicate},
	{matchIs(infrastructure.ErrActorAlreadySigned), http.StatusConflict, approvalCodeSignoffDuplicate},
	{matchIs(infrastructure.ErrInvalidSupersedeTarget), http.StatusConflict, approvalCodeValidationSupersedeTargetInvalid},
	{matchIs(infrastructure.ErrInstanceCompleted), http.StatusConflict, approvalCodeStateInstanceCompleted},
	{matchIs(infrastructure.ErrRouteInUse), http.StatusConflict, approvalCodeRouteInUse},
	{matchIs(infrastructure.ErrDuplicateRouteProfile), http.StatusConflict, approvalCodeRouteDuplicateProfile},
	{matchIs(domain.ErrActorNotEligible), http.StatusForbidden, approvalCodeSignoffNotEligible},
	{matchIs(domain.ErrAuthorCannotSign), http.StatusForbidden, approvalCodeSodSubmitterCannotSign},
	{matchIs(domain.ErrActorAlreadySigned), http.StatusForbidden, approvalCodeSodCrossStageDuplicate},
	{matchIs(v2dom.ErrEffectiveDateMissing), http.StatusUnprocessableEntity, approvalCodeFreezeEffDateMissing},
	// F6.3 §5.3: REV>=1 submit without a structured reason is a friendly
	// first-line validation failure — 422 (semantically valid JSON,
	// business-rule rejection), mirroring the effective-date-missing case.
	{matchIs(application.ErrReasonForChangeRequired), http.StatusUnprocessableEntity, approvalCodeValidationReasonForChangeRequired},
	{matchIs(application.ErrReasonCategoryInvalid), http.StatusUnprocessableEntity, approvalCodeValidationReasonCategoryInvalid},
	// ADR 0073: canonical /submit now returns this (was finalize-only). REV>=1
	// submit without a revision title is a friendly business-rule rejection —
	// 422, mirroring the reason-for-change case above.
	{matchIs(application.ErrRevisionTitleRequired), http.StatusUnprocessableEntity, approvalCodeValidationRevisionTitleRequired},
	// M3 P3.S2b-2: a document-kind route create with subject_key != profile_code
	// is a friendly first-line validation failure (semantically valid JSON,
	// business-rule rejection) — 422, mirroring the reason-for-change/revision-
	// title cases above.
	{matchIs(application.ErrDocumentSubjectKeyMismatch), http.StatusUnprocessableEntity, approvalCodeValidationDocumentSubjectKeyMismatch},
	// ADR 0086: a template route is profile-keyed, so subject_key != profile_code
	// is the same class of friendly business-rule rejection as the document case.
	{matchIs(application.ErrTemplateSubjectKeyMismatch), http.StatusUnprocessableEntity, approvalCodeValidationTemplateSubjectKeyMismatch},
	// In-tx submit resolution surfaces the finalize-era sentinels (ADR 0073).
	// Document not in draft = illegal state for the submit write → 409.
	{matchIs(v2dom.ErrDocumentNotDraft), http.StatusConflict, approvalCodeStateDocumentNotDraft},
	// Controlled document has no profile → the server cannot resolve a route.
	// Actionable request problem (finalize mapped it 400 ValidationError).
	{matchIs(v2dom.ErrProfileNotConfigured), http.StatusBadRequest, approvalCodeValidationProfileNotConfigured},
	// No active approval route for the profile (finalize mapped it 409).
	{matchIs(v2dom.ErrApprovalRouteMissing), http.StatusConflict, approvalCodeStateApprovalRouteMissing},
	{matchIs(ErrIfMatchRequired), http.StatusPreconditionRequired, approvalCodePreconditionIfMatch},
	{matchIs(ErrIfMatchMalformed), http.StatusBadRequest, approvalCodeValidationIfMatchBad},
	{matchIs(ErrIdempotencyRequired, idempotency.ErrKeyRequired, application.ErrIdempotencyKeyRequired), http.StatusBadRequest, approvalCodeIdempotencyRequired},
	// F-QA4-6: bespoke-replay handlers enforce the same UUID wire rule the
	// idempotency.Require middleware enforces, and deliberately surface the
	// PLATFORM code here (not a module-local dialect) so a malformed key
	// looks identical to clients whether or not the route is wrapped.
	{matchIs(idempotency.ErrKeyInvalid), http.StatusBadRequest, problem.CodeRequestIdempotencyKeyInvalid},
	// Same Idempotency-Key reused with a different request fingerprint: the
	// caller must rotate the key for a genuinely new attempt.
	{matchIs(idempotency.ErrConflict), http.StatusConflict, approvalCodeIdempotencyKeyConflict},
	{matchIs(ErrContentHashMismatch), http.StatusPreconditionFailed, approvalCodePreconditionHashMismatch},
	{matchIs(approvalsignature.ErrInvalidCredentials), http.StatusUnauthorized, approvalCodeAuthnSignatureInvalid},
	{matchIs(approvalsignature.ErrRateLimited), http.StatusTooManyRequests, approvalCodeAuthnRateLimited},
	// Signature-verifier misconfiguration (registry/limiter unwired). Unreachable
	// in production wiring; distinct 500 code so monitoring separates it from a
	// DB-layer 500. Client body stays the generic "internal error".
	{matchIs(application.ErrReauthNotConfigured, approvalsignature.ErrUnknownSignatureMethod, approvalsignature.ErrRateLimiterConfig), http.StatusInternalServerError, approvalCodeInternalSigMisconfigured},
	{matchIs(infrastructure.ErrInsufficientPrivilege), http.StatusInternalServerError, approvalCodeInternalDBPrivilege},
	{matchIs(infrastructure.ErrUnknownDB), http.StatusInternalServerError, approvalCodeInternalDBUnknown},
	// annex R-14: "no active stage" and "the instance is completed" are
	// different conditions with different operator remedies, so this sentinel
	// no longer folds onto state.approval_instance_completed — which now
	// carries only infrastructure.ErrInstanceCompleted. Status is unchanged.
	{matchIs(domain.ErrNoActiveStage), http.StatusConflict, approvalCodeStateApprovalStageNotActive},
	// R3/G2: a `ready` verdict targeting an approval-kind stage is a
	// business-rule rejection of a semantically-valid request — 422,
	// mirroring the v2dom.ErrEffectiveDateMissing / ErrReasonForChangeRequired
	// precedent above.
	{matchIs(domain.ErrVerdictReadyOnApprovalStage), http.StatusUnprocessableEntity, approvalCodeStateVerdictReadyOnApprovalStage},
	// Only reachable if an active stage carries a kind outside the
	// DB-constrained {review,approval} set — corrupted internal state, not a
	// client-caused condition. 500 with a distinct internal.* code, matching
	// the ErrInsufficientPrivilege / ErrUnknownDB precedent above (not the 422
	// business-rule class used for ready-on-approval).
	{matchIs(domain.ErrVerdictWrongStageKind), http.StatusInternalServerError, approvalCodeInternalVerdictWrongStageKind},
	// ADR 0087: the profile is livre — its route is configured to require no
	// approval, so it must carry ZERO stages. Adding one is a business-rule
	// rejection of a structurally-valid request (422), and it gets a
	// dedicated code so the route builder can say precisely why. The DB
	// trigger (assert_route_shape, migration 0316) rejects the same shape at
	// COMMIT; this is the friendly first line.
	{matchIs(domain.ErrRouteStagesNotPermittedForProfile), http.StatusUnprocessableEntity, approvalCodeValidationRouteStagesNotPermitted},
	// controlado: the route must contain >=1 approval-kind stage.
	{matchIs(domain.ErrApprovalStageRequired), http.StatusUnprocessableEntity, approvalCodeValidationApprovalStageRequired},
	// simples: review-only is fine, stageless is not (that shape is livre's).
	{matchIs(domain.ErrRouteStageRequired), http.StatusUnprocessableEntity, approvalCodeValidationRouteStageRequired},
	{matchIs(application.ErrDocumentNotFound), http.StatusNotFound, approvalCodeNotFoundDocument},
	// Friendly first-line precondition (mark-reviewed requires published);
	// 409 mirrors the other illegal-state-for-write cases above (stale
	// revision / instance completed), not a validation 4xx.
	{matchIs(application.ErrDocumentNotPublished), http.StatusConflict, approvalCodeStateDocumentNotPublished},
	// 409, mirroring infrastructure.ErrStaleRevision above: the mark-reviewed
	// route's openapi response set is {400,401,403,404,409,428,500} — no
	// 412 — so the OCC conflict is a 409 Conflict, not 412.
	{matchIs(application.ErrMarkReviewedStaleRevision), http.StatusConflict, approvalCodeConflictMarkReviewedStaleRevision},
	{matchIs(application.ErrReviewDueBeforeEffective), http.StatusUnprocessableEntity, approvalCodeValidationReviewDueBeforeEffective},
	{matchIs(application.ErrEffectiveToNotAfterEffectiveFrom), http.StatusUnprocessableEntity, approvalCodeValidationEffectiveToNotAfterFrom},
	{matchIs(domain.ErrSelfDelegation), http.StatusUnprocessableEntity, approvalCodeValidationSelfDelegation},
	{matchIs(domain.ErrInvalidDelegationWindow), http.StatusUnprocessableEntity, approvalCodeValidationDelegationWindow},
	{matchIs(application.ErrDelegationNotFoundOrNotOwned), http.StatusNotFound, approvalCodeNotFoundDelegation},
	// Shortening a deadline is a different act with different governance
	// consequences and is out of scope for this route — 422, a
	// business-rule rejection of a semantically valid request.
	{matchIs(application.ErrSLAExtensionNotForward), http.StatusUnprocessableEntity, approvalCodeSLANotForward},
	// No active stage, or the active stage carries no deadline — nothing
	// to extend. Illegal state for the write, mirroring the module's
	// other "no active stage" / "instance completed" 409s.
	{matchIs(application.ErrSLAExtensionNoActiveStage), http.StatusConflict, approvalCodeSLANoActiveStage},
	{matchIs(application.ErrSLAExtensionReasonRequired), http.StatusUnprocessableEntity, approvalCodeSLAReasonRequired},
	// R5: the actor's `ready` verdict was recorded but did not satisfy the
	// active stage's quorum (e.g. all_of with a co-reviewer still pending) —
	// an illegal state for the fast-forward composite write, not a client
	// validation error. 409, mirroring the other illegal-state-for-write
	// cases above (stale revision / instance completed).
	{matchIs(domain.ErrFastForwardStageNotCompleted), http.StatusConflict, approvalCodeStateFastForwardStageNotCompleted},
	// R5: the verdict completed the stage, but the instance is already
	// approved, there is no next stage, or the actor is not in the next
	// (approval-kind) stage's eligible pool — the composite fast-forward
	// write cannot proceed as one transaction. 409, same class as above.
	{matchIs(domain.ErrFastForwardNotEligible), http.StatusConflict, approvalCodeStateFastForwardNotEligible},
	// F2 (W6): a stage whose eligibility resolution yields zero actors is a
	// business-rule rejection at submit time (422), not a 500 — and this
	// mapping also closes a pre-existing gap for decision_service.go's
	// quorum-evaluation path, which returned this sentinel unmapped before.
	{matchIs(domain.ErrEmptyEligiblePool), http.StatusUnprocessableEntity, approvalCodeValidationEmptyEligiblePool},
	// M4, unit 3.2, slice 5: a submit_choice-governed stage has no
	// matching (or empty) chosen_actors entry — fail-closed business-rule
	// rejection at submit time, 422 mirroring ErrEmptyEligiblePool above.
	{matchIs(domain.ErrSubmitChoiceRequired), http.StatusUnprocessableEntity, approvalCodeValidationSubmitChoiceRequired},
	// M4, unit 3.2, slice 5: either a chosen user does not satisfy the
	// submit_choice selector's role x area_code constraint, or a
	// chosen_actors entry targets a stage_order with no submit_choice
	// selector (no-fallback principle) — both 422.
	{matchIs(domain.ErrSubmitChoiceConstraintViolated), http.StatusUnprocessableEntity, approvalCodeValidationSubmitChoiceConstraint},

	// The remaining entries were the inner `default:` switch — reached only
	// when nothing above matched, which this flat table reproduces simply by
	// listing them after every case above (first match still wins).
	{matchAs[*ValidationError](), http.StatusBadRequest, approvalCodeValidationRequestInvalid},
	{matchAs[*approvalapi.InvalidParamFormatError](), http.StatusBadRequest, approvalCodeValidationParamFormat},
	{matchAs[*approvalapi.UnmarshalingParamError](), http.StatusBadRequest, approvalCodeValidationParamUnmarshal},
	{matchAs[*approvalapi.RequiredParamError](), http.StatusBadRequest, approvalCodeValidationParamRequired},
	{matchAs[*approvalapi.RequiredHeaderError](), http.StatusBadRequest, approvalCodeValidationHeaderRequired},
	{matchAs[*approvalapi.TooManyValuesForParamError](), http.StatusBadRequest, approvalCodeValidationParamTooMany},
	{matchAs[authz.ErrCapDenied](), http.StatusForbidden, approvalCodeAuthzCapDenied},
	{matchIs(application.ErrApprovalBlockedByUnresolvedComments), http.StatusConflict, approvalCodeApprovalUnresolved},
	{matchIs(application.ErrReasonRequired, application.ErrRouteDeactivateReasonRequired), http.StatusBadRequest, approvalCodeValidationReasonRequired},
	{matchIs(application.ErrInvalidObsoleteSource), http.StatusBadRequest, approvalCodeValidationRequestInvalid},
	// FK violation on (tenant_id, profile_code) → the document profile
	// does not exist for this tenant. Actionable 4xx, never a 500.
	{matchIs(application.ErrRouteProfileUnknown), http.StatusUnprocessableEntity, approvalCodeValidationProfileUnknown},
	{matchIs(application.ErrRouteNotFound), http.StatusNotFound, approvalCodeNotFoundRoute},
	{matchIs(application.ErrRouteAlreadyInactive), http.StatusConflict, approvalCodeStateRouteInactive},
	{matchIs(context.DeadlineExceeded, context.Canceled), http.StatusGatewayTimeout, approvalCodeTimeout},
	{matchAs[*json.SyntaxError](), http.StatusBadRequest, approvalCodeValidationJSONDecode},
	{matchIs(io.ErrUnexpectedEOF), http.StatusBadRequest, approvalCodeValidationJSONDecode},
	{func(err error) bool { return err != nil && strings.HasPrefix(err.Error(), "json: unknown field") }, http.StatusBadRequest, approvalCodeValidationJSONDecode},
	{matchAs[*json.UnmarshalTypeError](), http.StatusBadRequest, approvalCodeValidationJSONTypeError},
	{matchIs(io.EOF), http.StatusBadRequest, approvalCodeValidationEmptyBody},
	{matchIs(strictjson.ErrContentType), http.StatusUnsupportedMediaType, approvalCodeValidationContentType},
	{matchIs(strictjson.ErrBodyTooLarge), http.StatusRequestEntityTooLarge, approvalCodeValidationBodyTooLarge},
	{matchIs(strictjson.ErrEmptyBody), http.StatusBadRequest, approvalCodeValidationEmptyBody},
	{matchIs(contracts.ErrValidation), http.StatusBadRequest, approvalCodeValidationRequestInvalid},
}

// MapErrorToResponse translates a domain/repository/contract error into an RFC
// 9457 problem+json response, choosing the HTTP status and typed problem.Code
// that best matches err. Unrecognized errors fall back to a generic 500 with
// approvalCodeInternalUnknown — the client never sees internal error detail.
func MapErrorToResponse(err error) *problem.Problem {
	statusCode := http.StatusInternalServerError
	code := approvalCodeInternalUnknown

	for _, m := range approvalErrorMappings {
		if m.match(err) {
			statusCode, code = m.status, m.code
			break
		}
	}

	return problem.New(statusCode, code, responseTitle(err, statusCode))
}

// WriteError is approval's error-TRANSLATION seam (R-ERR-2): it maps a domain
// or infrastructure error onto the module's problem catalog and hands the
// result to the canonical writer. It owns the mapping and nothing else — no
// status/header/body emission, no logging of its own — so 4xx and 5xx cannot
// drift from the rest of the API.
//
// The cause goes to problem.RespondCause because the client body for a 5xx is
// deliberately generic: the underlying error must still reach the log.
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	problem.RespondCause(w, r, MapErrorToResponse(err), err)
}

// WriteJSON marshals body and writes it with the given status code. A marshal
// failure is an error response, so it leaves through the canonical problem
// writer instead of a hand-rolled application/json payload.
func WriteJSON(w http.ResponseWriter, r *http.Request, status int, body any) {
	payload, err := json.Marshal(body)
	if err != nil {
		problem.RespondCause(w, r,
			problem.New(http.StatusInternalServerError, approvalCodeInternalUnknown, internalErrorMessage), err)
		return
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
