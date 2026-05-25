package domain

import "fmt"

// QuorumOutcome is the result of evaluating quorum for a stage.
type QuorumOutcome string

const (
	QuorumPending       QuorumOutcome = "pending"
	QuorumApprovedStage QuorumOutcome = "approved_stage"
	QuorumRejectedStage QuorumOutcome = "rejected_stage"
	QuorumError         QuorumOutcome = "error"
)

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

func EvaluateQuorumResult(stage StageInstance, approvals []Signoff, rejections []Signoff, effectiveDenominator int) QuorumResult {
	if len(stage.EligibleActorIDs) == 0 {
		return QuorumResult{Outcome: QuorumError, Reason: "empty_eligible_set"}
	}
	if effectiveDenominator == 0 {
		return QuorumResult{Outcome: QuorumRejectedStage}
	}

	approveCount := 0
	rejectCount := 0

	eligible := make(map[string]bool, len(stage.EligibleActorIDs))
	for _, id := range stage.EligibleActorIDs {
		eligible[id] = true
	}
	for _, s := range approvals {
		if eligible[s.ActorUserID()] {
			approveCount++
		}
	}
	for _, s := range rejections {
		if eligible[s.ActorUserID()] {
			rejectCount++
		}
	}

	switch stage.QuorumSnapshot {
	case QuorumAny1Of:
		if approveCount >= 1 {
			return QuorumResult{Outcome: QuorumApprovedStage}
		}
		if rejectCount >= 1 {
			return QuorumResult{Outcome: QuorumRejectedStage}
		}
		return QuorumResult{Outcome: QuorumPending}

	case QuorumAllOf:
		if rejectCount >= 1 {
			return QuorumResult{Outcome: QuorumRejectedStage}
		}
		if approveCount >= effectiveDenominator {
			return QuorumResult{Outcome: QuorumApprovedStage}
		}
		return QuorumResult{Outcome: QuorumPending}

	case QuorumMofN:
		m := 1
		if stage.QuorumMSnapshot != nil {
			m = *stage.QuorumMSnapshot
		}
		if approveCount >= m {
			return QuorumResult{Outcome: QuorumApprovedStage}
		}
		if rejectCount > effectiveDenominator-m {
			return QuorumResult{Outcome: QuorumRejectedStage}
		}
		return QuorumResult{Outcome: QuorumPending}
	default:
		return QuorumResult{Outcome: QuorumError, Reason: fmt.Sprintf("unknown quorum type: %q", stage.QuorumSnapshot)}
	}
}
