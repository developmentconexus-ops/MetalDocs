package approvalhttp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
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

// Module-local typed codes for the approval domain (dot-notation taxonomy).
//
// ADR 0089 step 3: these were `const … problem.Code = "…"`. problem.Code is now a
// closed struct type that only the registry can issue, so they are package-level
// `var` registrations. Strings and statuses are unchanged. RegisterLegacy skips
// the semantic-family validation these pre-taxonomy names would fail; execution
// steps 6 and 10 rename them per annex §2.2 and delete RegisterLegacy.
//
// The five codes bound to problem.CodeShared* are wire strings ALSO emitted by
// another module; the registry's duplicate guard forbids declaring them twice.
var (
	approvalCodeInternalUnknown = problem.CodeSharedInternalUnknown
	approvalCodeConflictStaleRevision = problem.RegisterLegacy("approval", "conflict.stale_revision", 409)
	approvalCodeNotFoundInstance = problem.RegisterLegacy("approval", "not_found.instance", 404)
	// approvalCodeNotFoundInstanceNotVisible (F8, spec.md §6.3) is the
	// DISTINCT 404 code for infrastructure.ErrInstanceNotVisible —
	// deliberately its own code (not approvalCodeNotFoundInstance) so
	// monitoring/logs can tell "instance genuinely does not exist" apart from
	// "instance exists but is outside this actor's visibility boundary",
	// even though both return the same 404 status to the client (the client
	// response body must not leak which case it is).
	approvalCodeNotFoundInstanceNotVisible = problem.RegisterLegacy("approval", "not_found.instance_not_visible", 404)
	approvalCodeConflictDuplicate = problem.RegisterLegacy("approval", "conflict.duplicate_submission", 409)
	approvalCodeSignoffDuplicate = problem.RegisterLegacy("approval", "signoff.duplicate", 409)
	approvalCodeSubmitInvalidSupersede = problem.RegisterLegacy("approval", "submit.invalid_supersede_target", 409)
	approvalCodeStateInstanceCompleted = problem.RegisterLegacy("approval", "state.instance_completed", 409)
	approvalCodeRouteInUse = problem.RegisterLegacy("approval", "route.in_use", 409)
	approvalCodeRouteDuplicateProfile = problem.RegisterLegacy("approval", "route.duplicate_profile", 409)
	approvalCodeSignoffNotEligible = problem.RegisterLegacy("approval", "signoff.not_eligible", 403)
	approvalCodeSodSubmitterCannotSign = problem.RegisterLegacy("approval", "sod.submitter_cannot_sign", 403)
	approvalCodeSodCrossStageDuplicate = problem.RegisterLegacy("approval", "sod.cross_stage_duplicate", 403)
	approvalCodeFreezeEffDateMissing = problem.RegisterLegacy("approval", "freeze.effective_date_missing", 422)
	approvalCodePreconditionIfMatch = problem.RegisterLegacy("approval", "precondition.if_match_required", 428)
	approvalCodeValidationIfMatchBad = problem.RegisterLegacy("approval", "validation.if_match_malformed", 400)
	approvalCodeIdempotencyRequired = problem.RegisterLegacy("approval", "idempotency.key_required", 400)
	approvalCodeIdempotencyKeyConflict = problem.RegisterLegacy("approval", "idempotency.key_conflict", 409)
	approvalCodePreconditionHashMismatch = problem.RegisterLegacy("approval", "precondition.content_hash_mismatch", 412)
	approvalCodeAuthnSignatureInvalid = problem.RegisterLegacy("approval", "authn.signature_invalid", 401)
	approvalCodeAuthnRateLimited = problem.RegisterLegacy("approval", "authn.rate_limited", 429)
	approvalCodeInternalDBPrivilege = problem.RegisterLegacy("approval", "internal.db_privilege_missing", 500)
	approvalCodeInternalDBUnknown = problem.RegisterLegacy("approval", "internal.db_unknown", 500)
	approvalCodeInternalSigMisconfigured = problem.RegisterLegacy("approval", "internal.signature_misconfigured", 500)
	approvalCodeValidationParamFormat = problem.RegisterLegacy("approval", "validation.param_format", 400)
	approvalCodeValidationParamUnmarshal = problem.RegisterLegacy("approval", "validation.param_unmarshal", 400)
	approvalCodeValidationParamRequired = problem.RegisterLegacy("approval", "validation.param_required", 400)
	approvalCodeValidationHeaderRequired = problem.RegisterLegacy("approval", "validation.header_required", 400)
	approvalCodeValidationParamTooMany = problem.RegisterLegacy("approval", "validation.param_too_many_values", 400)
	approvalCodeAuthzCapDenied = problem.CodeSharedAuthzCapabilityDenied
	approvalCodeApprovalUnresolved = problem.RegisterLegacy("approval", "approval.unresolved_comments", 409)
	approvalCodeValidationReasonRequired = problem.RegisterLegacy("approval", "validation.reason_required", 400)
	approvalCodeNotFoundRoute = problem.RegisterLegacy("approval", "not_found.route", 404)
	approvalCodeStateRouteInactive = problem.RegisterLegacy("approval", "state.route_inactive", 409)
	approvalCodeTimeout = problem.RegisterLegacy("approval", "timeout", 504)
	approvalCodeValidationJSONDecode = problem.CodeSharedValidationJSONDecode
	approvalCodeValidationJSONTypeError = problem.RegisterLegacy("approval", "validation.json_type_error", 400)
	approvalCodeValidationEmptyBody = problem.CodeSharedValidationEmptyBody
	approvalCodeValidationContentType = problem.RegisterLegacy("approval", "validation.content_type", 415)
	approvalCodeValidationBodyTooLarge = problem.RegisterLegacy("approval", "validation.body_too_large", 413)
	approvalCodeValidationRequestInvalid = problem.RegisterLegacy("approval", "validation.request_invalid", 400)
	approvalCodeValidationProfileUnknown = problem.RegisterLegacy("approval", "validation.profile_unknown", 422)
	approvalCodeValidationReasonForChangeRequired = problem.RegisterLegacy("approval", "validation.reason_for_change_required", 422)
	approvalCodeValidationReasonCategoryInvalid = problem.RegisterLegacy("approval", "validation.reason_category_invalid", 422)
	approvalCodeValidationRevisionTitleRequired = problem.RegisterLegacy("approval", "validation.revision_title_required", 422)
	approvalCodeValidationDocumentSubjectKeyMismatch = problem.RegisterLegacy("approval", "validation.document_subject_key_mismatch", 422)
	approvalCodeValidationTemplateSubjectKeyMismatch = problem.RegisterLegacy("approval", "validation.template_subject_key_mismatch", 422)
	approvalCodeStateDocumentNotDraft = problem.RegisterLegacy("approval", "state.document_not_draft", 409)
	approvalCodeValidationProfileNotConfigured = problem.RegisterLegacy("approval", "validation.profile_not_configured", 400)
	approvalCodeStateApprovalRouteMissing = problem.CodeSharedStateApprovalRouteMissing
	approvalCodeNotFoundDocument = problem.RegisterLegacy("approval", "not_found.document", 404)
	approvalCodeStateDocumentNotPublished = problem.RegisterLegacy("approval", "state.document_not_published", 409)
	approvalCodeConflictMarkReviewedStaleRevision = problem.RegisterLegacy("approval", "conflict.mark_reviewed_stale_revision", 409)
	approvalCodeValidationReviewDueBeforeEffective = problem.RegisterLegacy("approval", "validation.review_due_before_effective", 422)
	approvalCodeValidationEffectiveToNotAfterFrom = problem.RegisterLegacy("approval", "validation.effective_to_not_after_effective_from", 422)
	approvalCodeValidationEmptyEligiblePool = problem.RegisterLegacy("approval", "validation.empty_eligible_pool", 422)
	approvalCodeValidationSubmitChoiceRequired = problem.RegisterLegacy("approval", "validation.submit_choice_required", 422)
	approvalCodeValidationSubmitChoiceConstraint = problem.RegisterLegacy("approval", "validation.submit_choice_constraint_violated", 422)

	// Route-shape policy rejections (per-profile governance policy, G1/ADR 0081,
	// livre arm superseded by ADR 0087). All 422: the request is well-formed,
	// the resulting route shape is not permitted for the profile's class.
	approvalCodeValidationRouteStagesNotPermitted = problem.RegisterLegacy("approval", "validation.route_stages_not_permitted", 422)
	approvalCodeValidationApprovalStageRequired = problem.RegisterLegacy("approval", "validation.approval_stage_required", 422)
	approvalCodeValidationRouteStageRequired = problem.RegisterLegacy("approval", "validation.route_stage_required", 422)

	// F9/ADR 0077 — approval delegation.
	approvalCodeValidationSelfDelegation = problem.RegisterLegacy("approval", "validation.self_delegation", 422)
	approvalCodeValidationDelegationWindow = problem.RegisterLegacy("approval", "validation.delegation_window_invalid", 422)
	approvalCodeNotFoundDelegation = problem.RegisterLegacy("approval", "not_found.delegation", 404)

	// R3/G2 — review-verdict stage-kind guard.
	approvalCodeStateVerdictReadyOnApprovalStage = problem.RegisterLegacy("approval", "state.verdict_ready_on_approval_stage", 422)
	approvalCodeInternalVerdictWrongStageKind = problem.RegisterLegacy("approval", "internal.verdict_wrong_stage_kind", 500)

	// R5/unit 2.3 G3 — fast-forward ("Aprovar já") composition guards.
	approvalCodeStateFastForwardStageNotCompleted = problem.RegisterLegacy("approval", "state.fast_forward_stage_not_completed", 409)
	approvalCodeStateFastForwardNotEligible = problem.RegisterLegacy("approval", "state.fast_forward_not_eligible", 409)
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
	case errors.Is(err, infrastructure.ErrInstanceNotVisible):
		// F8, spec.md §6.3: cross-boundary = not-found. Same 404 status as
		// ErrNoActiveInstance but a distinct problem.Code (see the constant's
		// doc comment) so server-side logs/monitoring can distinguish the two
		// cases without the client-visible response revealing which one fired.
		statusCode = http.StatusNotFound
		code = approvalCodeNotFoundInstanceNotVisible
	case errors.Is(err, infrastructure.ErrDuplicateSubmission):
		statusCode = http.StatusConflict
		code = approvalCodeConflictDuplicate
	case errors.Is(err, infrastructure.ErrActorAlreadySigned):
		statusCode = http.StatusConflict
		code = approvalCodeSignoffDuplicate
	case errors.Is(err, infrastructure.ErrInvalidSupersedeTarget):
		statusCode = http.StatusConflict
		code = approvalCodeSubmitInvalidSupersede
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
	case errors.Is(err, application.ErrDocumentSubjectKeyMismatch):
		// M3 P3.S2b-2: a document-kind route create with subject_key != profile_code
		// is a friendly first-line validation failure (semantically valid JSON,
		// business-rule rejection) — 422, mirroring the reason-for-change/revision-
		// title cases above.
		statusCode = http.StatusUnprocessableEntity
		code = approvalCodeValidationDocumentSubjectKeyMismatch
	case errors.Is(err, application.ErrTemplateSubjectKeyMismatch):
		// ADR 0086: a template route is profile-keyed, so subject_key != profile_code
		// is the same class of friendly business-rule rejection as the document case.
		statusCode = http.StatusUnprocessableEntity
		code = approvalCodeValidationTemplateSubjectKeyMismatch
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
		errors.Is(err, idempotency.ErrKeyRequired),
		errors.Is(err, application.ErrIdempotencyKeyRequired):
		statusCode = http.StatusBadRequest
		code = approvalCodeIdempotencyRequired
	case errors.Is(err, idempotency.ErrKeyInvalid):
		// F-QA4-6: bespoke-replay handlers enforce the same UUID wire rule the
		// idempotency.Require middleware enforces, and deliberately surface the
		// PLATFORM code here (not a module-local dialect) so a malformed key
		// looks identical to clients whether or not the route is wrapped.
		statusCode = http.StatusBadRequest
		code = problem.CodeIdempotencyKeyInvalid
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
	case errors.Is(err, domain.ErrVerdictReadyOnApprovalStage):
		// R3/G2: a `ready` verdict targeting an approval-kind stage is a
		// business-rule rejection of a semantically-valid request — 422,
		// mirroring the v2dom.ErrEffectiveDateMissing / ErrReasonForChangeRequired
		// precedent above.
		statusCode = http.StatusUnprocessableEntity
		code = approvalCodeStateVerdictReadyOnApprovalStage
	case errors.Is(err, domain.ErrVerdictWrongStageKind):
		// Only reachable if an active stage carries a kind outside the
		// DB-constrained {review,approval} set — corrupted internal state, not a
		// client-caused condition. 500 with a distinct internal.* code, matching
		// the ErrInsufficientPrivilege / ErrUnknownDB precedent above (not the 422
		// business-rule class used for ready-on-approval).
		statusCode = http.StatusInternalServerError
		code = approvalCodeInternalVerdictWrongStageKind
	case errors.Is(err, domain.ErrRouteStagesNotPermittedForProfile):
		// ADR 0087: the profile is livre — its route is configured to require no
		// approval, so it must carry ZERO stages. Adding one is a business-rule
		// rejection of a structurally-valid request (422), and it gets a
		// dedicated code so the route builder can say precisely why. The DB
		// trigger (assert_route_shape, migration 0316) rejects the same shape at
		// COMMIT; this is the friendly first line.
		statusCode = http.StatusUnprocessableEntity
		code = approvalCodeValidationRouteStagesNotPermitted
	case errors.Is(err, domain.ErrApprovalStageRequired):
		// controlado: the route must contain >=1 approval-kind stage.
		statusCode = http.StatusUnprocessableEntity
		code = approvalCodeValidationApprovalStageRequired
	case errors.Is(err, domain.ErrRouteStageRequired):
		// simples: review-only is fine, stageless is not (that shape is livre's).
		statusCode = http.StatusUnprocessableEntity
		code = approvalCodeValidationRouteStageRequired
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
	case errors.Is(err, domain.ErrSelfDelegation):
		statusCode = http.StatusUnprocessableEntity
		code = approvalCodeValidationSelfDelegation
	case errors.Is(err, domain.ErrInvalidDelegationWindow):
		statusCode = http.StatusUnprocessableEntity
		code = approvalCodeValidationDelegationWindow
	case errors.Is(err, application.ErrDelegationNotFoundOrNotOwned):
		statusCode = http.StatusNotFound
		code = approvalCodeNotFoundDelegation
	case errors.Is(err, domain.ErrFastForwardStageNotCompleted):
		// R5: the actor's `ready` verdict was recorded but did not satisfy the
		// active stage's quorum (e.g. all_of with a co-reviewer still pending) —
		// an illegal state for the fast-forward composite write, not a client
		// validation error. 409, mirroring the other illegal-state-for-write
		// cases above (stale revision / instance completed).
		statusCode = http.StatusConflict
		code = approvalCodeStateFastForwardStageNotCompleted
	case errors.Is(err, domain.ErrFastForwardNotEligible):
		// R5: the verdict completed the stage, but the instance is already
		// approved, there is no next stage, or the actor is not in the next
		// (approval-kind) stage's eligible pool — the composite fast-forward
		// write cannot proceed as one transaction. 409, same class as above.
		statusCode = http.StatusConflict
		code = approvalCodeStateFastForwardNotEligible
	case errors.Is(err, domain.ErrEmptyEligiblePool):
		// F2 (W6): a stage whose eligibility resolution yields zero actors is a
		// business-rule rejection at submit time (422), not a 500 — and this
		// mapping also closes a pre-existing gap for decision_service.go's
		// quorum-evaluation path, which returned this sentinel unmapped before.
		statusCode = http.StatusUnprocessableEntity
		code = approvalCodeValidationEmptyEligiblePool
	case errors.Is(err, domain.ErrSubmitChoiceRequired):
		// M4, unit 3.2, slice 5: a submit_choice-governed stage has no
		// matching (or empty) chosen_actors entry — fail-closed business-rule
		// rejection at submit time, 422 mirroring ErrEmptyEligiblePool above.
		statusCode = http.StatusUnprocessableEntity
		code = approvalCodeValidationSubmitChoiceRequired
	case errors.Is(err, domain.ErrSubmitChoiceConstraintViolated):
		// M4, unit 3.2, slice 5: either a chosen user does not satisfy the
		// submit_choice selector's role x area_code constraint, or a
		// chosen_actors entry targets a stage_order with no submit_choice
		// selector (no-fallback principle) — both 422.
		statusCode = http.StatusUnprocessableEntity
		code = approvalCodeValidationSubmitChoiceConstraint
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
			slog.String("code", prob.Code.String()),
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
