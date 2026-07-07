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

// ResolveEligibleIdentity widens eligibility for delegation (F9/ADR 0077)
// WITHOUT introducing a second, possibly-divergent membership rule: it tries
// CheckEligibility directly for actorUserID first (the common, non-delegated
// case), and on failure tries CheckEligibility again — the SAME function,
// SAME eligibleActorIDs pool — for each active delegation's DelegatorID.
// Delegation only ever supplies additional candidate identities to the one
// real predicate; it never bypasses it.
//
// Returns onBehalfOf == "" when the actor is directly eligible (acting as
// themselves). Returns onBehalfOf == the winning delegation's DelegatorID
// when eligibility was satisfied via delegation. Returns ErrActorNotEligible
// (unchanged sentinel) when neither the direct check nor any delegation
// resolves — callers get the same error/event/audit shape as before F9 for
// the true-ineligible case.
func ResolveEligibleIdentity(actorUserID string, eligibleActorIDs []string, delegations []Delegation) (onBehalfOf string, err error) {
	if err := CheckEligibility(actorUserID, eligibleActorIDs); err == nil {
		return "", nil
	}

	for _, d := range delegations {
		if d.DelegateID() != actorUserID {
			continue
		}
		if err := CheckEligibility(d.DelegatorID(), eligibleActorIDs); err == nil {
			return d.DelegatorID(), nil
		}
	}

	return "", ErrActorNotEligible
}
