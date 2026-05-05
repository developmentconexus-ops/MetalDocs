package domain

import (
	"errors"
	"testing"
	"testing/quick"
)

func TestCheckEligibility_PresentReturnsNil(t *testing.T) {
	err := CheckEligibility("u2", []string{"u1", "u2", "u3"})
	if err != nil {
		t.Fatalf("got %v, want nil", err)
	}
}

func TestCheckEligibility_AbsentReturnsErr(t *testing.T) {
	err := CheckEligibility("u9", []string{"u1", "u2"})
	if !errors.Is(err, ErrActorNotEligible) {
		t.Fatalf("got %v, want ErrActorNotEligible", err)
	}
}

func TestCheckEligibility_EmptyListReturnsErr(t *testing.T) {
	err := CheckEligibility("u1", nil)
	if !errors.Is(err, ErrActorNotEligible) {
		t.Fatalf("got %v", err)
	}
}

func TestCheckEligibility_PropertyMembership(t *testing.T) {
	prop := func(actor string, list []string) bool {
		want := false
		for _, id := range list {
			if id == actor {
				want = true
				break
			}
		}
		got := CheckEligibility(actor, list) == nil
		return got == want
	}
	if err := quick.Check(prop, nil); err != nil {
		t.Fatalf("property failed: %v", err)
	}
}
