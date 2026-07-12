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

// ErrRouteNotPermittedForProfile is returned by Route.Validate when the
// profile's governance policy forbids any approval route (livre class →
// RoutePolicyNoRoutePermitted). Friendly first line for the DB trigger.
var ErrRouteNotPermittedForProfile = errors.New("approval: profile governance policy permits no approval route")

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
