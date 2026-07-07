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
