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
const (
	approvalCodeInternalUnknown       problem.Code = "internal.unknown"
	approvalCodeConflictStaleRevision problem.Code = "conflict.stale_revision"
	approvalCodeNotFoundInstance      problem.Code = "not_found.instance"
	// approvalCodeNotFoundInstanceNotVisible (F8, spec.md §6.3) is the
	// DISTINCT 404 code for infrastructure.ErrInstanceNotVisible —
	// deliberately its own code (not approvalCodeNotFoundInstance) so
	// monitoring/logs can tell "instance genuinely does not exist" apart from
	// "instance exists but is outside this actor's visibility boundary",
	// even though both return the same 404 status to the client (the client
	// response body must not leak which case it is).
	approvalCodeNotFoundInstanceNotVisible           problem.Code = "not_found.instance_not_visible"
	approvalCodeConflictDuplicate                    problem.Code = "conflict.duplicate_submission"
	approvalCodeSignoffDuplicate                     problem.Code = "signoff.duplicate"
	approvalCodePublishInvalidSupersede              problem.Code = "publish.invalid_supersede_target"
	approvalCodeStateInstanceCompleted               problem.Code = "state.instance_completed"
	approvalCodeRouteInUse                           problem.Code = "route.in_use"
	approvalCodeRouteDuplicateProfile                problem.Code = "route.duplicate_profile"
	approvalCodeSignoffNotEligible                   problem.Code = "signoff.not_eligible"
	approvalCodeSodSubmitterCannotSign               problem.Code = "sod.submitter_cannot_sign"
	approvalCodeSodCrossStageDuplicate               problem.Code = "sod.cross_stage_duplicate"
	approvalCodeFreezeEffDateMissing                 problem.Code = "freeze.effective_date_missing"
	approvalCodePreconditionIfMatch                  problem.Code = "precondition.if_match_required"
	approvalCodeValidationIfMatchBad                 problem.Code = "validation.if_match_malformed"
	approvalCodeIdempotencyRequired                  problem.Code = "idempotency.key_required"
	approvalCodeIdempotencyKeyConflict               problem.Code = "idempotency.key_conflict"
	approvalCodePreconditionHashMismatch             problem.Code = "precondition.content_hash_mismatch"
	approvalCodeAuthnSignatureInvalid                problem.Code = "authn.signature_invalid"
	approvalCodeAuthnRateLimited                     problem.Code = "authn.rate_limited"
	approvalCodeInternalDBPrivilege                  problem.Code = "internal.db_privilege_missing"
	approvalCodeInternalDBUnknown                    problem.Code = "internal.db_unknown"
	approvalCodeInternalSigMisconfigured             problem.Code = "internal.signature_misconfigured"
	approvalCodeValidationParamFormat                problem.Code = "validation.param_format"
	approvalCodeValidationParamUnmarshal             problem.Code = "validation.param_unmarshal"
	approvalCodeValidationParamRequired              problem.Code = "validation.param_required"
	approvalCodeValidationHeaderRequired             problem.Code = "validation.header_required"
	approvalCodeValidationParamTooMany               problem.Code = "validation.param_too_many_values"
	approvalCodeAuthzCapDenied                       problem.Code = "authz.capability_denied"
	approvalCodeApprovalUnresolved                   problem.Code = "approval.unresolved_comments"
	approvalCodeValidationReasonRequired             problem.Code = "validation.reason_required"
	approvalCodeNotFoundRoute                        problem.Code = "not_found.route"
	approvalCodeStateRouteInactive                   problem.Code = "state.route_inactive"
	approvalCodeTimeout                              problem.Code = "timeout"
	approvalCodeValidationJSONDecode                 problem.Code = "validation.json_decode"
	approvalCodeValidationJSONTypeError              problem.Code = "validation.json_type_error"
	approvalCodeValidationEmptyBody                  problem.Code = "validation.empty_body"
	approvalCodeValidationContentType                problem.Code = "validation.content_type"
	approvalCodeValidationBodyTooLarge               problem.Code = "validation.body_too_large"
	approvalCodeValidationRequestInvalid             problem.Code = "validation.request_invalid"
	approvalCodeValidationProfileUnknown             problem.Code = "validation.profile_unknown"
	approvalCodeValidationReasonForChangeRequired    problem.Code = "validation.reason_for_change_required"
	approvalCodeValidationReasonCategoryInvalid      problem.Code = "validation.reason_category_invalid"
	approvalCodeValidationRevisionTitleRequired      problem.Code = "validation.revision_title_required"
	approvalCodeValidationDocumentSubjectKeyMismatch problem.Code = "validation.document_subject_key_mismatch"
	approvalCodeStateDocumentNotDraft                problem.Code = "state.document_not_draft"
	approvalCodeValidationProfileNotConfigured       problem.Code = "validation.profile_not_configured"
	approvalCodeStateApprovalRouteMissing            problem.Code = "state.approval_route_missing"
	approvalCodeNotFoundDocument                     problem.Code = "not_found.document"
	approvalCodeStateDocumentNotPublished            problem.Code = "state.document_not_published"
	approvalCodeConflictMarkReviewedStaleRevision    problem.Code = "conflict.mark_reviewed_stale_revision"
	approvalCodeValidationReviewDueBeforeEffective   problem.Code = "validation.review_due_before_effective"
	approvalCodeValidationEffectiveToNotAfterFrom    problem.Code = "validation.effective_to_not_after_effective_from"
	approvalCodeValidationEmptyEligiblePool          problem.Code = "validation.empty_eligible_pool"
	approvalCodeValidationSubmitChoiceRequired       problem.Code = "validation.submit_choice_required"
	approvalCodeValidationSubmitChoiceConstraint     problem.Code = "validation.submit_choice_constraint_violated"

	// F9/ADR 0077 — approval delegation.
	approvalCodeValidationSelfDelegation   problem.Code = "validation.self_delegation"
	approvalCodeValidationDelegationWindow problem.Code = "validation.delegation_window_invalid"
	approvalCodeNotFoundDelegation         problem.Code = "not_found.delegation"

	// R3/G2 — review-verdict stage-kind guard.
	approvalCodeStateVerdictReadyOnApprovalStage problem.Code = "state.verdict_ready_on_approval_stage"
	approvalCodeInternalVerdictWrongStageKind    problem.Code = "internal.verdict_wrong_stage_kind"

	// R5/unit 2.3 G3 — fast-forward ("Aprovar já") composition guards.
	approvalCodeStateFastForwardStageNotCompleted problem.Code = "state.fast_forward_stage_not_completed"
	approvalCodeStateFastForwardNotEligible       problem.Code = "state.fast_forward_not_eligible"
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
	case errors.Is(err, infrastructure.ErrInvalidScheduledSupersedeTarget):
		statusCode = http.StatusConflict
		code = approvalCodePublishInvalidSupersede
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
		case errors.Is(err, application.ErrEffectiveDateInPast):
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
			slog.String("code", string(prob.Code)),
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
