// Package problem defines the canonical error code catalog used by
// MetalDocs Problem responses (RFC 9457). See spec §5.2.
package problem

// PLATFORM CATALOG (ADR 0089 execution step 5, annex §2.1 + §2.10).
//
// Every code here is registered under the closed semantic family set: the
// pre-0089 SCREAMING_SNAKE names are gone, the 11 dead constants are deleted,
// and `RegisterLegacy` no longer appears in this file. Codes are grouped by
// family, and the group order is the family order of annex §1.2.
//
// Four registrations are MERGES forced by the registry's duplicate guard: the
// annex renames two old wire strings onto one new string, so they can no longer
// be two declarations (C-1..C-19 in annex §2.9):
//
//	auth.unauthenticated          <- UNAUTHENTICATED + AUTH_UNAUTHORIZED   (C-2)
//	permission.capability_denied  <- FORBIDDEN_CAPABILITY + authz.capability_denied (C-3)
//	internal.unknown              <- INTERNAL_ERROR + internal.unknown     (C-7)
//	state.approval_route_missing  <- APPROVAL_ROUTE_MISSING + state.approval_route_missing (C-9)
//
// The `shared` module tag marks a code two or more bounded contexts genuinely
// emit; it is documentation and lint metadata only and NEVER reaches the wire
// (see Registration.Module).
//
// TWO KINDS of `shared` code live here, and the second is why this block keeps
// growing after the module sweeps:
//
//  1. Wire strings that were ALREADY byte-identical in two modules before the
//     rename — annex §2.10's nine (S-1..S-9).
//  2. Wire strings that CONVERGE only AFTER the rename, because two modules
//     describe the same condition and the semantic taxonomy correctly gives that
//     condition one name (C-10..C-14, C-16..C-18, R-2/R-4/R-6/R-14/R-22).
//
// Both land in the same place for the same reason: Register panics at init on a
// duplicate, and a module may not import another module's delivery package, so
// the platform catalog is the only legal single declaration site. That is annex
// §2.10's own argument applied to a set it did not enumerate — not a new
// decision, but ADR 0089 decision 2 plus the module-boundary invariant.
//
// Do NOT add module-owned codes here — a new code belongs in its owning module,
// registered with problem.Register under a semantic family.

// ---------------------------------------------------------------------------
// request. — the request itself is malformed, unparseable, or protocol-
// unacceptable. Fix the syntax/shape and retry. Family default 400.
// ---------------------------------------------------------------------------

var (
	// CodeRequestInvalid is the generic request-shape rejection (annex #1, C-1).
	CodeRequestInvalid = Register("platform", "request.invalid", 400)

	// CodeRequestCursorInvalid is returned when a pagination cursor is malformed.
	CodeRequestCursorInvalid = Register("platform", "request.cursor_invalid", 400)

	// CodeRequestMethodNotAllowed is returned when the target exists but does
	// not support the request's HTTP method.
	CodeRequestMethodNotAllowed = RegisterWithStatus("platform", "request.method_not_allowed", 405,
		"405 is the dedicated status for an unsupported method on an existing target (RFC 9110 §15.5.6); "+
			"the request is otherwise well-formed, so the request. family is right and 400 is not")

	// CodeRequestBodyTooLarge is returned when the request body exceeds the
	// configured size limit.
	CodeRequestBodyTooLarge = RegisterWithStatus("platform", "request.body_too_large", 413,
		"413 Content Too Large is the dedicated status for an oversized body (RFC 9110 §15.5.14)")

	// CodeRequestCursorExpired is returned when the pagination cursor
	// addressed an item that has been permanently removed.
	CodeRequestCursorExpired = RegisterWithStatus("platform", "request.cursor_expired", 410,
		"410 Gone: the cursor addressed an item that has been permanently removed, so the caller must "+
			"restart pagination rather than fix the request syntax")

	// CodeRequestIdempotencyKeyInvalid is returned when a supplied
	// Idempotency-Key does not meet the required shape.
	CodeRequestIdempotencyKeyInvalid = Register("platform", "request.idempotency_key_invalid", 400)
	// CodeRequestIdempotencyKeyRequired is returned when a route that
	// requires an Idempotency-Key received none.
	CodeRequestIdempotencyKeyRequired = Register("platform", "request.idempotency_key_required", 400)

	// CodeRequestJSONDecode and CodeRequestEmptyBody are shared by approval and
	// the documents fill-in surface (annex §2.10 S-2/S-3, C-16).
	CodeRequestJSONDecode = Register("shared", "request.json_decode", 400)
	// CodeRequestEmptyBody is returned when the request body is required but
	// empty (see CodeRequestJSONDecode's doc comment for the shared context).
	CodeRequestEmptyBody = Register("shared", "request.empty_body", 400)

	// CodeRequestContentTypeUnsupported is the third leg of C-16: approval (annex
	// row #86) and the documents fill-in surface (row #123) both rename onto this
	// one wire string at the same 415, so it can no longer be declared twice.
	// Status is unchanged in both modules.
	CodeRequestContentTypeUnsupported = RegisterWithStatus("shared", "request.content_type_unsupported", 415,
		"415 Unsupported Media Type is the dedicated status for a rejected Content-Type (RFC 9110 §15.5.16)")
)

// ---------------------------------------------------------------------------
// auth. — the caller's identity is missing, unproven, or blocked. Default 401.
// ---------------------------------------------------------------------------

var (
	// CodeAuthUnauthenticated is the single no/invalid-session code (annex C-2).
	CodeAuthUnauthenticated = Register("platform", "auth.unauthenticated", 401)

	// CodeAuthInvalidCredentials is returned when the supplied credentials do
	// not match any account.
	CodeAuthInvalidCredentials = Register("platform", "auth.invalid_credentials", 401)

	// The five 403 registrations below deviate from the auth. default of 401 for
	// one shared reason: identity is PROVEN and the account or tenant claim is
	// what blocks the call (RFC 9110 §15.5.4). A 401 would wrongly invite the
	// client to re-authenticate against a credential that is already valid.

	// CodeAuthAccountLocked is returned when identity is proven but the account
	// is administratively blocked, so re-authenticating cannot help.
	CodeAuthAccountLocked = RegisterWithStatus("platform", "auth.account_locked", 403,
		"identity is proven; the account is administratively blocked, so re-authenticating cannot help")
	// CodeAuthAccountInactive is returned when identity is proven but the
	// account is deactivated, so re-authenticating cannot help.
	CodeAuthAccountInactive = RegisterWithStatus("platform", "auth.account_inactive", 403,
		"identity is proven; the account is deactivated, so re-authenticating cannot help")
	// CodeAuthTenantForbidden is returned when identity is proven but the
	// caller may not act on the requested tenant.
	CodeAuthTenantForbidden = RegisterWithStatus("platform", "auth.tenant_forbidden", 403,
		"identity is proven; the caller may not act on the requested tenant")
	// CodeAuthTenantRequired is returned when identity is proven but the
	// session carries no tenant claim for a tenant-scoped route.
	CodeAuthTenantRequired = RegisterWithStatus("platform", "auth.tenant_required", 403,
		"identity is proven; the session carries no tenant claim for a tenant-scoped route")
	// CodeAuthPasswordChangeRequired is returned when identity is proven but
	// the session is refused until the mandatory password change completes.
	CodeAuthPasswordChangeRequired = RegisterWithStatus("platform", "auth.password_change_required", 403,
		"identity is proven; the session is refused until the mandatory password change completes")
)

// ---------------------------------------------------------------------------
// permission. — the caller is identified but lacks the required capability in
// the required area. Default 403.
// ---------------------------------------------------------------------------

var (
	// CodePermissionCapabilityDenied is the single capability-denial code
	// (annex C-3): it absorbs the old FORBIDDEN_CAPABILITY catalog constant and
	// the shared authz.capability_denied string emitted by approval and fill-in.
	CodePermissionCapabilityDenied = Register("shared", "permission.capability_denied", 403)

	// CodePermissionDenied replaces AUTH_FORBIDDEN, which was mis-prefixed: it is
	// a capability/ownership denial, not an identity failure (annex #33).
	CodePermissionDenied = Register("platform", "permission.denied", 403)

	// CodePermissionOriginForbidden is returned when the request's Origin
	// header fails the origin-protection check.
	CodePermissionOriginForbidden = Register("platform", "permission.origin_forbidden", 403)

	// CodePermissionISOSegregationViolation is returned when the requested
	// action would violate an ISO segregation-of-duties rule.
	CodePermissionISOSegregationViolation = Register("platform", "permission.iso_segregation_violation", 403)

	// The two sign-off eligibility denials below are emitted by BOTH approval (its
	// own handlers) and templates (which classifies the approval kernel's published
	// application/domain error surface at its own delivery edge). templates used to
	// flatten both onto the generic capability denial — annex C-11 / C-12, R-4.
	// A module may not import another module's delivery package, so the shared
	// platform catalog is the single legal declaration site.

	// CodePermissionSignoffActorNotEligible is returned when the caller is not
	// an eligible actor for the sign-off stage.
	CodePermissionSignoffActorNotEligible = Register("shared", "permission.signoff_actor_not_eligible", 403)
	// CodePermissionSodSubmitterCannotSign is returned when the submitter of a
	// change is not permitted to also sign it off (segregation of duties).
	CodePermissionSodSubmitterCannotSign = Register("shared", "permission.sod_submitter_cannot_sign", 403)
)

// ---------------------------------------------------------------------------
// notfound. — the addressed subject does not exist or is invisible to this
// caller. Default 404.
// ---------------------------------------------------------------------------

var (
	// CodeNotFoundResource is the generic "addressed subject does not exist" code.
	CodeNotFoundResource = Register("platform", "notfound.resource", 404)

	// CodeNotFoundMembership is returned when the addressed membership
	// (user-to-tenant or user-to-group) does not exist.
	CodeNotFoundMembership = Register("platform", "notfound.membership", 404)

	// CodeNotFoundDocumentProfile is shared with controlleddocuments and
	// taxonomy (annex §2.10 S-6/S-8).
	CodeNotFoundDocumentProfile = Register("shared", "notfound.document_profile", 404)
	// CodeNotFoundProcessArea is shared with controlleddocuments and taxonomy
	// (annex §2.10 S-6/S-8).
	CodeNotFoundProcessArea = Register("shared", "notfound.process_area", 404)
)

// ---------------------------------------------------------------------------
// state. — the subject exists but is in the wrong lifecycle state; retrying is
// futile until a different operation changes the state. Default 409.
// ---------------------------------------------------------------------------

var (
	// CodeStateTransitionInvalid is returned when the requested lifecycle
	// transition is not valid from the subject's current state.
	CodeStateTransitionInvalid = Register("platform", "state.transition_invalid", 409)

	// CodeStateApprovalRouteMissing is the single "no active approval route" code
	// (annex C-9), shared by approval, controlleddocuments and templates.
	CodeStateApprovalRouteMissing = Register("shared", "state.approval_route_missing", 409)

	// CodeStateUploadExpired is returned when the upload session existed but
	// has permanently expired.
	CodeStateUploadExpired = RegisterWithStatus("platform", "state.upload_expired", 410,
		"410 Gone: unlike state.upload_missing, the upload session did exist and has permanently "+
			"expired, which is the condition 410 asserts (RFC 9110 §15.5.11)")

	// CodeStateUploadMissing is 409, not the pre-0089 documents value of 410
	// (ADR 0089 decision 3, annex R-1): the upload never happened, so 410 Gone —
	// which asserts the resource existed and was permanently removed — is false.
	CodeStateUploadMissing = Register("platform", "state.upload_missing", 409)

	// CodeStateSystemTemplateImmutable is returned when the requested mutation
	// targets a system-owned template, which is never editable.
	CodeStateSystemTemplateImmutable = Register("platform", "state.system_template_immutable", 409)

	// CodeStateDocumentProfileArchived is shared with controlleddocuments and
	// taxonomy (annex §2.10 S-7/S-9).
	CodeStateDocumentProfileArchived = Register("shared", "state.document_profile_archived", 409)
	// CodeStateProcessAreaArchived is shared with controlleddocuments and
	// taxonomy (annex §2.10 S-7/S-9).
	CodeStateProcessAreaArchived = Register("shared", "state.process_area_archived", 409)

	// The two approval-lifecycle codes below are emitted by approval and by
	// templates, whose delivery edge classifies the approval kernel's published
	// error surface for the template submit/signoff path.
	//
	// annex R-14 / §2.8 #164 keeps them apart deliberately: "the instance is
	// finished" and "this stage is not the active one" have different operator
	// remedies, and templates used to flatten both onto conflict.generic.

	// CodeStateApprovalInstanceCompleted is returned when the approval instance
	// has already finished.
	CodeStateApprovalInstanceCompleted = Register("shared", "state.approval_instance_completed", 409)
	// CodeStateApprovalStageNotActive is returned when the addressed stage is
	// not the approval instance's currently active stage.
	CodeStateApprovalStageNotActive = Register("shared", "state.approval_stage_not_active", 409)
)

// ---------------------------------------------------------------------------
// conflict. — race, uniqueness collision, or idempotency-key collision; a retry
// against refreshed state may succeed. Default 409.
// ---------------------------------------------------------------------------

var (
	// CodeConflictAlreadyExists is returned when the resource being created
	// already exists.
	CodeConflictAlreadyExists = Register("platform", "conflict.already_exists", 409)

	// CodeConflictConcurrentModification is returned when a concurrent write
	// raced the caller's operation.
	CodeConflictConcurrentModification = Register("platform", "conflict.concurrent_modification", 409)

	// CodeConflictIdempotencyKeyReused is 409, not the pre-0089 422 (ADR 0089
	// decision 3, annex R-8): the key is well-formed and meaningful; the fault is
	// a collision with a prior request, not unprocessable content.
	CodeConflictIdempotencyKeyReused = Register("platform", "conflict.idempotency_key_reused", 409)

	// CodeConflictGeneric is the catch-all 409 (annex #24, ex-CONFLICT_ERROR).
	CodeConflictGeneric = Register("platform", "conflict.generic", 409)

	// CodeConflictStaleBase is returned when the caller's base version is
	// stale relative to the current state.
	CodeConflictStaleBase = Register("platform", "conflict.stale_base", 409)
	// CodeConflictMembershipExists is returned when the membership being
	// created already exists.
	CodeConflictMembershipExists = Register("platform", "conflict.membership_exists", 409)

	// CodeConflictDuplicateSubmission is shared by approval and templates (annex
	// C-10, R-4): a second submit while an approval instance is already open.
	// templates used to answer conflict.generic.
	CodeConflictDuplicateSubmission = Register("shared", "conflict.duplicate_submission", 409)
)

// ---------------------------------------------------------------------------
// precondition. — a precondition the CALLER SUPPLIED (If-Match, expected content
// hash, lock version) no longer holds. Default 412.
// ---------------------------------------------------------------------------

var (
	// CodePreconditionContentHashMismatch is the single content-hash-mismatch code
	// (annex C-17 / R-2, ratified in ADR 0089 decision 3): the caller DECLARED a
	// precondition and it failed (RFC 9110 §15.5.13). It was 412 in approval, 422
	// in documents and 409 in templates; 412 wins everywhere. Approval and
	// templates both emit it, so it lives here rather than in either module.
	CodePreconditionContentHashMismatch = Register("shared", "precondition.content_hash_mismatch", 412)
)

// ---------------------------------------------------------------------------
// validation. — the request parses; a supplied value or a business rule over
// supplied values rejects it. Default 422.
// ---------------------------------------------------------------------------

var (
	// CodeValidationRoleUnknown is 422, not the pre-0089 400 (annex R-12): the
	// body parses and the supplied value names a role that does not exist.
	CodeValidationRoleUnknown = Register("platform", "validation.role_unknown", 422)

	// CodeValidationFailed is the canonical GENERIC 422 (annex row #121, R-6). It
	// was declared by the documents fill-in surface; taxonomy and tokens need the
	// same string for their R-6 residual sites, which were answering 422 while
	// carrying request.invalid (registered @400) — a code/status contradiction the
	// registry's status binding forbids.
	CodeValidationFailed = Register("shared", "validation.failed", 422)

	// CodeValidationTemplateProfileMismatch is the single "this template version
	// does not belong to that document profile" code (annex C-14 / R-22, rows #139
	// + #150). It was `template_invalid` @422 in controlleddocuments and
	// TEMPLATE_PROFILE_MISMATCH @409 in taxonomy; 422 wins — the caller supplied an
	// invalid template/profile pair and nothing is racing.
	CodeValidationTemplateProfileMismatch = Register("shared", "validation.template_profile_mismatch", 422)

	// The three submit-time business-rule rejections below are emitted by approval
	// and by templates (annex C-18 / R-4). Statuses already agreed at 422; only the
	// codes diverged, because templates flattened all three onto conflict.generic
	// and thereby made the FE message map useless for template submit.

	// CodeValidationEmptyEligiblePool is returned when the configured stage
	// resolves to zero eligible actors, so no one could ever act on it.
	CodeValidationEmptyEligiblePool = Register("shared", "validation.empty_eligible_pool", 422)
	// CodeValidationSubmitChoiceRequired is returned when the stage needs the
	// submitter to pick an actor and the request did not carry one.
	CodeValidationSubmitChoiceRequired = Register("shared", "validation.submit_choice_required", 422)
	// CodeValidationSubmitChoiceConstraintViolated is returned when the submitter
	// did pick, but the choice breaks a stage constraint (eligibility, SoD).
	CodeValidationSubmitChoiceConstraintViolated = Register("shared", "validation.submit_choice_constraint_violated", 422)
)

// ---------------------------------------------------------------------------
// ratelimit. — the caller is being throttled. Default 429.
// ---------------------------------------------------------------------------

var (
	// CodeRateLimitExceeded is emitted by the rate-limit middlewares
	// (platform/ratelimit + platform/security) and, per annex C-6, by
	// approval's signature re-auth limiter once step 6 lands.
	CodeRateLimitExceeded = Register("platform", "ratelimit.exceeded", 429)
)

// ---------------------------------------------------------------------------
// internal. — server fault: bug, misconfiguration, unimplemented surface, or
// upstream failure. Default 500.
// ---------------------------------------------------------------------------

var (
	// CodeInternalUnknown is the single unmapped-server-fault code (annex C-7):
	// it absorbs the old INTERNAL_ERROR catalog constant and the shared
	// internal.unknown string emitted by approval and fill-in.
	CodeInternalUnknown = Register("shared", "internal.unknown", 500)

	// CodeInternalNotImplemented is returned for a wired-but-unimplemented surface.
	CodeInternalNotImplemented = RegisterWithStatus("platform", "internal.not_implemented", 501,
		"501 Not Implemented is the dedicated status for a wired-but-unimplemented surface (RFC 9110 §15.6.2)")
)
