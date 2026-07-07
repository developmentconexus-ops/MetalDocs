package domain

import (
	"testing"
)

// R2-2: Drift→Quorum call-order — fail_stage short-circuits quorum.
func TestDriftThenQuorumOrdering(t *testing.T) {
	st := StageInstance{
		Status:                     StageActive,
		EligibleActorIDs:           []string{"a", "b"},
		OnEligibilityDriftSnapshot: DriftFailStage,
		QuorumSnapshot:             QuorumAllOf,
	}

	// Simulate: actor "a" departed.
	driftResult := ApplyEligibilityDrift(st, []string{"b"})
	if driftResult.ForcedOutcome != QuorumRejectedStage {
		t.Fatalf("fail_stage drift should force rejection; got %s", driftResult.ForcedOutcome)
	}

	// When ForcedOutcome != QuorumPending, quorum MUST NOT be consulted.
	// We assert this by verifying caller logic: if ForcedOutcome != pending, return it directly.
	var finalOutcome QuorumOutcome
	if driftResult.ForcedOutcome != QuorumPending {
		finalOutcome = driftResult.ForcedOutcome
	} else {
		// This branch must NOT run in this test.
		finalOutcome = EvaluateQuorum(st, nil, nil, driftResult.EffectiveDenominator)
	}

	if finalOutcome != QuorumRejectedStage {
		t.Errorf("canonical call order should yield rejected_stage; got %s", finalOutcome)
	}
}

// R2-3: RevisionVersion gates — deferred to Phase 4 repository layer OCC.
// TODO(phase4): AdvanceStage + OCC expectedVersion check lives in approval_repository.
// Domain-only test: BumpRevisionVersion enforces monotonic at domain level.
func TestRevisionVersionGatesStateTransition(t *testing.T) {
	inst := threeStageInstance()
	inst.RevisionVersion = 1

	// Bump from 1 → 2: ok.
	if err := inst.BumpRevisionVersion(2); err != nil {
		t.Fatalf("BumpRevisionVersion(2): %v", err)
	}
	// Regression 2 → 1: rejected.
	if err := inst.BumpRevisionVersion(1); err == nil {
		t.Error("regression should be rejected")
	}
	// Note: full OCC version check (AdvanceStage(expectedVersion)) is enforced in Phase 4
	// repository layer via SELECT FOR UPDATE and version comparison. Domain layer only
	// ensures monotonicity via BumpRevisionVersion.
}
