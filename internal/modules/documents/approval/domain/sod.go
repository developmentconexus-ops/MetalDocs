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
// Rules:
//   - actorUserID must not equal authorUserID
//   - actorUserID must not appear in any priorSignoffs (same instance, earlier stages)
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
