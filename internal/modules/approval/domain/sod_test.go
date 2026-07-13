package domain

import (
	"errors"
	"testing"
)

func TestSoDSelfSign(t *testing.T) {
	err := CheckSoD("author-1", "author-1", "", nil)
	if !errors.Is(err, ErrAuthorCannotSign) {
		t.Errorf("want ErrAuthorCannotSign; got %v", err)
	}
}

func TestSoDReSignAcrossStages(t *testing.T) {
	prior := []Signoff{makeSignoff("actor-1", DecisionApprove)}
	err := CheckSoD("author-1", "actor-1", "", prior)
	if !errors.Is(err, ErrActorAlreadySigned) {
		t.Errorf("want ErrActorAlreadySigned; got %v", err)
	}
}

func TestSoDFreshActor(t *testing.T) {
	prior := []Signoff{makeSignoff("actor-1", DecisionApprove)}
	err := CheckSoD("author-1", "fresh-actor", "", prior)
	if err != nil {
		t.Errorf("fresh actor should pass SoD; got %v", err)
	}
}

func TestSoDEmptyPrior(t *testing.T) {
	err := CheckSoD("author-1", "actor-2", "", nil)
	if err != nil {
		t.Errorf("empty prior signoffs should pass SoD; got %v", err)
	}
}

// F9 (ADR 0077): delegate inherits delegator's SoD constraint.

func TestSoDDelegateActingForAuthor_Rejected(t *testing.T) {
	// delegate-1 is not the author, but is acting on behalf of author-1
	// (the author), via an active delegation — must be rejected exactly like
	// a direct self-sign.
	err := CheckSoD("author-1", "delegate-1", "author-1", nil)
	if !errors.Is(err, ErrAuthorCannotSign) {
		t.Errorf("want ErrAuthorCannotSign for delegate acting as author; got %v", err)
	}
}

func TestSoDDelegateActingForNonAuthor_Allowed(t *testing.T) {
	// delegate-1 acting on behalf of some other eligible actor (not the
	// author) is unaffected.
	err := CheckSoD("author-1", "delegate-1", "actor-2", nil)
	if err != nil {
		t.Errorf("delegate acting for a non-author delegator should pass SoD; got %v", err)
	}
}
