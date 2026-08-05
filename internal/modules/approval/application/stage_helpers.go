package application

import "metaldocs/internal/modules/approval/domain"

// activeStageEligibleIDs returns the eligible actors of the one active stage,
// or nil when there is none (a livre route's zero-stage instance, ADR 0087).
// Deliberately not "all stages": a stage that has not opened yet is not
// waiting on anyone, so it contributes no approval.pending recipients.
//
// Shared by TemplateSubmitService.SubmitTemplateVersionForReview and
// SubmitService.SubmitRevisionForReview — imported by both, not duplicated.
func activeStageEligibleIDs(stages []domain.StageInstance) []string {
	for _, s := range stages {
		if s.Status == domain.StageActive {
			return s.EligibleActorIDs
		}
	}
	return nil
}
