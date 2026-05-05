package domain

import "errors"

// ErrActorNotEligible is returned when an actor is not in the eligible_actor_ids
// snapshot frozen at submit time.
var ErrActorNotEligible = errors.New("eligibility: actor is not in the eligible_actor_ids snapshot for the active stage")

// CheckEligibility verifies actorUserID is present in the eligible-actor snapshot
// frozen at submit time. Pure function — no DB, no globals. Mirrors sod.go shape.
func CheckEligibility(actorUserID string, eligibleActorIDs []string) error {
	for _, id := range eligibleActorIDs {
		if id == actorUserID {
			return nil
		}
	}
	return ErrActorNotEligible
}
