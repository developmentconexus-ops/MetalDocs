package domain

// ViewerFacts is the pure, server-derived eligibility truth for a single
// viewer against a single instance's active stage (ADR 0078). Display facts
// only — enforcement stays exactly at authz.Require (tier-2) plus the
// write-path CheckEligibility/CheckSoD call sites; this value never gates
// the server.
type ViewerFacts struct {
	IsAuthor                bool
	EligibleForActiveStage  bool
	HasSignedActiveStage    bool
	ViaDelegationFromUserID string // "" when none (directly eligible, not eligible, or no active stage)
}

// ViewerEligibility computes ViewerFacts for viewerID against authorID and
// the instance's active stage. Composes EXCLUSIVELY on the write-path's
// single eligibility source (ResolveEligibleIdentity) and single SoD source
// (CheckSoD) — this function must never reimplement delegation matching or
// introduce a second SoD rule (ADR 0078, the anti-split-brain requirement).
//
// stages is the full stage list of the instance, used to compute
// priorSignoffs (this viewer's signoffs on stages OTHER than active) for the
// cross-stage double-sign clause of CheckSoD.
func ViewerEligibility(viewerID, authorID string, active *StageInstance, delegations []Delegation, stages []StageInstance) ViewerFacts {
	facts := ViewerFacts{
		IsAuthor: viewerID == authorID,
	}

	if active == nil {
		return facts
	}

	onBehalf, err := ResolveEligibleIdentity(viewerID, active.EligibleActorIDs, delegations)

	var priorSignoffs []Signoff
	for i := range stages {
		if stages[i].ID == active.ID {
			continue
		}
		for _, sig := range stages[i].Signoffs {
			if sig != nil {
				priorSignoffs = append(priorSignoffs, *sig)
			}
		}
	}

	passesSoD := CheckSoD(authorID, viewerID, onBehalf, priorSignoffs) == nil
	eligible := err == nil && passesSoD

	facts.EligibleForActiveStage = eligible
	if eligible && onBehalf != "" {
		facts.ViaDelegationFromUserID = onBehalf
	}

	for _, sig := range active.Signoffs {
		if sig != nil && sig.ActorUserID() == viewerID {
			facts.HasSignedActiveStage = true
			break
		}
	}

	return facts
}
