package domain

import (
	"errors"
	"testing"
	"testing/quick"
	"time"
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

// F9/ADR 0077 — ResolveEligibleIdentity is the composition point: it must
// try CheckEligibility directly first, and only widen via delegation as a
// fallback, using the SAME underlying predicate for both attempts.

func newTestDelegation(t *testing.T, delegator, delegate string) Delegation {
	t.Helper()
	now := time.Now()
	d, err := NewDelegation(DelegationParams{
		ID: delegator + "-" + delegate, TenantID: "t1",
		DelegatorID: delegator, DelegateID: delegate,
		StartsAt: now.Add(-time.Hour), EndsAt: now.Add(time.Hour),
		Reason: "test", CreatedBy: delegator, CreatedAt: now,
	})
	if err != nil {
		t.Fatalf("unexpected error building test delegation: %v", err)
	}
	return *d
}

func TestResolveEligibleIdentity_DirectMatch_NoDelegationNeeded(t *testing.T) {
	pool := []string{"actor-1", "actor-2"}
	onBehalf, err := ResolveEligibleIdentity("actor-1", pool, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if onBehalf != "" {
		t.Errorf("want empty onBehalfOf for direct match; got %q", onBehalf)
	}
}

func TestResolveEligibleIdentity_DelegateOfEligiblePool_Widened(t *testing.T) {
	pool := []string{"delegator-1"}
	delegations := []Delegation{newTestDelegation(t, "delegator-1", "delegate-1")}
	onBehalf, err := ResolveEligibleIdentity("delegate-1", pool, delegations)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if onBehalf != "delegator-1" {
		t.Errorf("want onBehalfOf=delegator-1; got %q", onBehalf)
	}
}

func TestResolveEligibleIdentity_DelegateWithNoActiveDelegation_StillIneligible(t *testing.T) {
	pool := []string{"delegator-1"}
	onBehalf, err := ResolveEligibleIdentity("delegate-1", pool, nil)
	if !errors.Is(err, ErrActorNotEligible) {
		t.Errorf("want ErrActorNotEligible; got %v", err)
	}
	if onBehalf != "" {
		t.Errorf("want empty onBehalfOf on failure; got %q", onBehalf)
	}
}

func TestResolveEligibleIdentity_DelegationForSomeoneNotInPool_Ineligible(t *testing.T) {
	pool := []string{"actor-1"}
	// delegate-1 is a delegate of delegator-1, but delegator-1 is not in the
	// pool — delegation must not widen eligibility beyond the real rule.
	delegations := []Delegation{newTestDelegation(t, "delegator-1", "delegate-1")}
	_, err := ResolveEligibleIdentity("delegate-1", pool, delegations)
	if !errors.Is(err, ErrActorNotEligible) {
		t.Errorf("want ErrActorNotEligible; got %v", err)
	}
}
