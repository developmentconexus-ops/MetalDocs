package domain

import (
	"errors"
	"testing"
	"time"
)

func validDelegationParams() DelegationParams {
	now := time.Now()
	return DelegationParams{
		ID:          "d1",
		TenantID:    "t1",
		DelegatorID: "user-a",
		DelegateID:  "user-b",
		StartsAt:    now,
		EndsAt:      now.Add(24 * time.Hour),
		Reason:      "on leave",
		CreatedBy:   "user-a",
		CreatedAt:   now,
	}
}

func TestNewDelegation_Valid(t *testing.T) {
	p := validDelegationParams()
	d, err := NewDelegation(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.DelegatorID() != "user-a" || d.DelegateID() != "user-b" {
		t.Errorf("identities did not round-trip: %+v", d)
	}
	if d.Reason() != "on leave" {
		t.Errorf("reason did not round-trip")
	}
}

func TestNewDelegation_SelfDelegationRejected(t *testing.T) {
	p := validDelegationParams()
	p.DelegateID = p.DelegatorID
	_, err := NewDelegation(p)
	if !errors.Is(err, ErrSelfDelegation) {
		t.Errorf("want ErrSelfDelegation; got %v", err)
	}
}

func TestNewDelegation_InvalidWindow_EndsBeforeStarts(t *testing.T) {
	p := validDelegationParams()
	p.EndsAt = p.StartsAt.Add(-time.Hour)
	_, err := NewDelegation(p)
	if !errors.Is(err, ErrInvalidDelegationWindow) {
		t.Errorf("want ErrInvalidDelegationWindow; got %v", err)
	}
}

func TestNewDelegation_InvalidWindow_EndsEqualsStarts(t *testing.T) {
	p := validDelegationParams()
	p.EndsAt = p.StartsAt
	_, err := NewDelegation(p)
	if !errors.Is(err, ErrInvalidDelegationWindow) {
		t.Errorf("want ErrInvalidDelegationWindow; got %v", err)
	}
}

func TestDelegation_IsActiveAt(t *testing.T) {
	p := validDelegationParams()
	d, err := NewDelegation(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !d.IsActiveAt(p.StartsAt) {
		t.Errorf("expected active at window start (inclusive)")
	}
	if d.IsActiveAt(p.EndsAt) {
		t.Errorf("expected inactive at window end (exclusive)")
	}
	if d.IsActiveAt(p.StartsAt.Add(-time.Minute)) {
		t.Errorf("expected inactive before window start")
	}
	if !d.IsActiveAt(p.StartsAt.Add(time.Hour)) {
		t.Errorf("expected active within window")
	}
}
