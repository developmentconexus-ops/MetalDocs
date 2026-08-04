// Package domain holds the approval bounded-context's pure domain model:
// the instance/stage aggregate, quorum and drift evaluation, signoff value
// objects, and the Spec 2 document-state graph. Types and functions here are
// side-effect free — no DB access, no clock reads beyond what callers pass in.
package domain

import "errors"

// ErrEmptyEligiblePool is returned when a stage has no eligible actors to evaluate quorum against.
var ErrEmptyEligiblePool = errors.New("approval: empty eligible pool")

// ErrInvalidStageKind is returned by StageKind.Validate for any value other
// than StageKindReview or StageKindApproval.
var ErrInvalidStageKind = errors.New("approval: invalid stage kind")

// ErrRouteStagesNotPermittedForProfile is returned by Route.Validate when the
// profile's governance policy requires a ZERO-stage route (livre class →
// RoutePolicyNoApprovalStages) but the route carries stages. Friendly first
// line for the DB trigger (assert_route_shape's livre arm, migration 0316).
//
// It replaces the pre-ADR-0087 ErrRouteNotPermittedForProfile: under the old
// model a livre profile could own no route at all, which made it unable to own
// documents or templates either. Config-first is now universal.
var ErrRouteStagesNotPermittedForProfile = errors.New("approval: profile governance policy permits no approval stages on its route")

// ErrRouteStageRequired is returned by Route.Validate when the profile's
// governance policy requires the route to carry at least one stage
// (simples class → RoutePolicyApprovalOptional) but the route has none.
// A zero-stage route is legal ONLY under RoutePolicyNoApprovalStages.
var ErrRouteStageRequired = errors.New("approval: profile governance policy requires at least one route stage")

// ErrApprovalStageRequired is returned by Route.Validate when the profile's
// governance policy requires at least one approval-kind (signature) stage
// (controlado class → RoutePolicyRequireApprovalStage) but the route has none.
var ErrApprovalStageRequired = errors.New("approval: profile governance policy requires at least one approval stage")

// ErrFastForwardStageNotCompleted (R5, unit 2.3 G3) is returned by
// FastForwardService.RecordFastForward when the `ready` verdict leg did not
// complete the review stage — quorum is still pending (e.g. an all_of stage
// with other reviewers still outstanding). The signoff leg is never attempted
// in this case.
var ErrFastForwardStageNotCompleted = errors.New("approval: fast-forward verdict did not complete the review stage")

// ErrFastForwardNotEligible (R5, unit 2.3 G3) is returned by
// FastForwardService.RecordFastForward when the review stage completed but
// there is no now-active approval-kind stage the acting actor is eligible on
// (instance fully approved with no next stage, next stage is review-kind, or
// the actor is not in the next stage's eligible pool). The signoff leg is
// never attempted in this case.
var ErrFastForwardNotEligible = errors.New("approval: fast-forward not eligible on the now-active stage")

// ErrInvalidSelectorKind is returned by ActorSelector.Validate for any Kind
// other than the four SelectorKind values (M4, unit 3.2).
var ErrInvalidSelectorKind = errors.New("approval: invalid actor selector kind")

// ErrSelectorFieldsInvalid is returned by ActorSelector.Validate when the
// UserID/Role/AreaCode fields don't match the presence/absence contract for
// the selector's Kind.
var ErrSelectorFieldsInvalid = errors.New("approval: actor selector fields invalid for its kind")

// ErrStageNoSelector is returned by Route.Validate when a stage has zero
// ActorSelectors — every stage must name at least one way to resolve its
// eligible actor pool.
var ErrStageNoSelector = errors.New("approval: stage must have at least one actor selector")

// ErrSubmitChoiceRequired is returned at submit time (M4, unit 3.2, slice 5)
// when a stage carries a submit_choice selector but the caller's
// chosen_actors has no entry for that stage's Order, or the entry's UserIDs
// is empty. Fail-closed: a submit_choice stage can never silently fall back
// to an empty pool.
var ErrSubmitChoiceRequired = errors.New("approval: submit_choice stage requires chosen actors")

// ErrTemplateVersionNoContent is returned when a template version carries no
// committed content hash at the moment it would enter approval. It lives here
// rather than in the application package because the templates-side submit-lock
// adapter must speak it across the TemplateVersionSubmitWriter port: that
// adapter may import approval/domain (it already does) but never
// approval/application. application.ErrTemplateVersionNoContent aliases this
// value, so errors.Is holds for both the fast-path guard and the CAS leg.
var ErrTemplateVersionNoContent = errors.New("approval: template version has no committed content")

// ErrSubmitChoiceConstraintViolated is returned at submit time (M4, unit 3.2,
// slice 5) when either (a) a chosen user does not satisfy the submit_choice
// selector's role x area_code constraint (not an active member in that
// role/area for the tenant), or (b) a chosen_actors entry targets a
// stage_order that has no submit_choice selector at all. Both are fail-closed
// per the no-fallback principle — never a silent accept.
var ErrSubmitChoiceConstraintViolated = errors.New("approval: chosen actor does not satisfy submit_choice constraint")
