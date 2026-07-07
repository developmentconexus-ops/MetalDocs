package domain

import "errors"

var (
	// ErrAuthorCannotSign is returned by CheckSoD when the signing actor is the document's author.
	ErrAuthorCannotSign = errors.New("SoD: document author cannot sign their own revision")
	// ErrActorAlreadySigned is returned by CheckSoD when the actor already signed off in an earlier stage of the same instance.
	ErrActorAlreadySigned = errors.New("SoD: actor has already signed in a prior stage of this instance")
)

// CheckSoD validates Separation-of-Duties rules. Pure function — no DB, no globals.
//
// This is the single, shared SoD predicate for the approval module (F7
// unification, spec.md §5/W7): both DecisionService.RecordSignoff
// (approval-stage signoffs) and ReviewVerdictService.RecordVerdict
// (review-stage verdicts) call this same function for the author-cannot-act-
// on-their-own-submission rule — neither reimplements the check inline.
//
// Rules:
//   - actorUserID must not equal authorUserID
//   - actorUserID must not appear in any priorSignoffs (same instance, earlier stages)
//
// Callers with no same-shape prior-record source (e.g. review verdicts, which
// live in a differently-shaped table with their own per-stage-actor
// uniqueness enforcement) pass nil for priorSignoffs — the cross-stage clause
// is then simply a no-op, and only the author self-check applies.
func CheckSoD(authorUserID string, actorUserID string, priorSignoffs []Signoff) error {
	if actorUserID == authorUserID {
		return ErrAuthorCannotSign
	}
	for _, s := range priorSignoffs {
		if s.ActorUserID() == actorUserID {
			return ErrActorAlreadySigned
		}
	}
	return nil
}
