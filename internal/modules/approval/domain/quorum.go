package domain

import "fmt"

// QuorumOutcome is the result of evaluating quorum for a stage.
type QuorumOutcome string

// QuorumOutcome values returned by EvaluateQuorum / EvaluateQuorumResult.
const (
	QuorumPending       QuorumOutcome = "pending"
	QuorumApprovedStage QuorumOutcome = "approved_stage"
	QuorumRejectedStage QuorumOutcome = "rejected_stage"
	QuorumError         QuorumOutcome = "error"
)

// QuorumResult is the outcome of EvaluateQuorumResult plus an optional
// human-readable Reason, populated for QuorumError and forced outcomes.
type QuorumResult struct {
	Outcome QuorumOutcome
	Reason  string
}

// ComputeEffectiveDenominator intersects the stage's snapshot eligible set with the
// current eligible set, returning the count. Callers never hand-compute this.
func ComputeEffectiveDenominator(stage StageInstance, currentEligible []string) int {
	if len(stage.EligibleActorIDs) == 0 || len(currentEligible) == 0 {
		return 0
	}
	snapshot := make(map[string]bool, len(stage.EligibleActorIDs))
	for _, id := range stage.EligibleActorIDs {
		snapshot[id] = true
	}
	count := 0
	for _, id := range currentEligible {
		if snapshot[id] {
			count++
		}
	}
	return count
}

// EvaluateQuorum evaluates whether signoffs satisfy the stage's quorum policy.
// Signoffs from actors NOT in EligibleActorIDs are ignored.
func EvaluateQuorum(stage StageInstance, approvals []Signoff, rejections []Signoff, effectiveDenominator int) QuorumOutcome {
	return EvaluateQuorumResult(stage, approvals, rejections, effectiveDenominator).Outcome
}

// EvaluateQuorumResult is the QuorumResult-returning counterpart of EvaluateQuorum,
// carrying a Reason string for QuorumError and denominator-forced outcomes.
// Signoffs from actors NOT in EligibleActorIDs are ignored.
func EvaluateQuorumResult(stage StageInstance, approvals []Signoff, rejections []Signoff, effectiveDenominator int) QuorumResult {
	if len(stage.EligibleActorIDs) == 0 {
		return QuorumResult{Outcome: QuorumError, Reason: "empty_eligible_set"}
	}
	if effectiveDenominator == 0 {
		return QuorumResult{Outcome: QuorumRejectedStage}
	}

	eligible := eligibleActorSet(stage.EligibleActorIDs)
	approveCount := countEligibleSignoffs(eligible, approvals)
	rejectCount := countEligibleSignoffs(eligible, rejections)

	switch stage.QuorumSnapshot {
	case QuorumAny1Of:
		return evaluateAny1Of(approveCount, rejectCount)
	case QuorumAllOf:
		return evaluateAllOf(approveCount, rejectCount, effectiveDenominator)
	case QuorumMofN:
		return evaluateMofN(stage.QuorumMSnapshot, approveCount, rejectCount, effectiveDenominator)
	default:
		return QuorumResult{Outcome: QuorumError, Reason: fmt.Sprintf("unknown quorum type: %q", stage.QuorumSnapshot)}
	}
}

// eligibleActorSet builds a membership set from a stage's eligible actor IDs.
func eligibleActorSet(eligibleActorIDs []string) map[string]bool {
	eligible := make(map[string]bool, len(eligibleActorIDs))
	for _, id := range eligibleActorIDs {
		eligible[id] = true
	}
	return eligible
}

// countEligibleSignoffs counts signoffs cast by actors present in eligible.
// Signoffs from actors NOT in the eligible set are ignored.
func countEligibleSignoffs(eligible map[string]bool, signoffs []Signoff) int {
	count := 0
	for _, s := range signoffs {
		if eligible[s.ActorUserID()] {
			count++
		}
	}
	return count
}

func evaluateAny1Of(approveCount, rejectCount int) QuorumResult {
	if approveCount >= 1 {
		return QuorumResult{Outcome: QuorumApprovedStage}
	}
	if rejectCount >= 1 {
		return QuorumResult{Outcome: QuorumRejectedStage}
	}
	return QuorumResult{Outcome: QuorumPending}
}

func evaluateAllOf(approveCount, rejectCount, effectiveDenominator int) QuorumResult {
	if rejectCount >= 1 {
		return QuorumResult{Outcome: QuorumRejectedStage}
	}
	if approveCount >= effectiveDenominator {
		return QuorumResult{Outcome: QuorumApprovedStage}
	}
	return QuorumResult{Outcome: QuorumPending}
}

func evaluateMofN(quorumMSnapshot *int, approveCount, rejectCount, effectiveDenominator int) QuorumResult {
	m := 1
	switch {
	case quorumMSnapshot == nil:
	case *quorumMSnapshot < 1:
		return QuorumResult{Outcome: QuorumRejectedStage}
	default:
		m = *quorumMSnapshot
	}
	if approveCount >= m {
		return QuorumResult{Outcome: QuorumApprovedStage}
	}
	if rejectCount > effectiveDenominator-m {
		return QuorumResult{Outcome: QuorumRejectedStage}
	}
	return QuorumResult{Outcome: QuorumPending}
}
