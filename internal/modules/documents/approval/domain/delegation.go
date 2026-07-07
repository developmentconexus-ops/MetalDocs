package domain

import (
	"errors"
	"time"
)

// ErrSelfDelegation is returned by NewDelegation when DelegatorID == DelegateID.
var ErrSelfDelegation = errors.New("delegation: delegator and delegate must not be the same user")

// ErrInvalidDelegationWindow is returned by NewDelegation when EndsAt does not
// come strictly after StartsAt.
var ErrInvalidDelegationWindow = errors.New("delegation: ends_at must be after starts_at")

// Delegation is an immutable value object representing a time-windowed grant
// of eligibility from one user (delegator) to another (delegate). See F9/ADR
// 0077: delegation widens the INPUT to CheckEligibility/CheckSoD, it never
// grants a capability and never bypasses authz.Require.
type Delegation struct {
	id          string
	tenantID    string
	delegatorID string
	delegateID  string
	startsAt    time.Time
	endsAt      time.Time
	reason      string
	createdBy   string
	createdAt   time.Time
}

// Getters — no setters exist; immutable after construction.

// ID returns the delegation's identifier.
func (d *Delegation) ID() string { return d.id }

// TenantID returns the owning tenant's identifier.
func (d *Delegation) TenantID() string { return d.tenantID }

// DelegatorID returns the user_id delegating their eligibility.
func (d *Delegation) DelegatorID() string { return d.delegatorID }

// DelegateID returns the user_id receiving the delegated eligibility.
func (d *Delegation) DelegateID() string { return d.delegateID }

// StartsAt returns the window's start (inclusive).
func (d *Delegation) StartsAt() time.Time { return d.startsAt }

// EndsAt returns the window's end (exclusive).
func (d *Delegation) EndsAt() time.Time { return d.endsAt }

// Reason returns the human-supplied justification for this grant.
func (d *Delegation) Reason() string { return d.reason }

// CreatedBy returns the user_id who created this delegation record. Always
// equal to DelegatorID today (a user may only delegate their own
// eligibility), kept as a distinct field to mirror this module's other audit
// tables and to allow a future oversee-assisted creation path without a
// column-purpose change.
func (d *Delegation) CreatedBy() string { return d.createdBy }

// CreatedAt returns when this delegation record was created.
func (d *Delegation) CreatedAt() time.Time { return d.createdAt }

// IsActiveAt reports whether this delegation covers the instant asOf
// (starts_at <= asOf < ends_at). Real-time, no caching — callers must always
// call this (or the equivalent DB predicate) at actual use-time, never from a
// previously-loaded snapshot, so that revocation (row deletion) and window
// expiry are honored immediately (F9/ADR 0077 §5).
func (d *Delegation) IsActiveAt(asOf time.Time) bool {
	return !asOf.Before(d.startsAt) && asOf.Before(d.endsAt)
}

// DelegationParams holds constructor inputs for NewDelegation.
type DelegationParams struct {
	ID          string
	TenantID    string
	DelegatorID string
	DelegateID  string
	StartsAt    time.Time
	EndsAt      time.Time
	Reason      string
	CreatedBy   string
	CreatedAt   time.Time
}

// NewDelegation constructs an immutable Delegation value object. App-checks-
// first line (DB CHECK constraints on approval_delegations are the
// tripwire-last-line backstop for the same two rules):
//   - DelegatorID must not equal DelegateID (ErrSelfDelegation)
//   - EndsAt must be strictly after StartsAt (ErrInvalidDelegationWindow)
func NewDelegation(p DelegationParams) (*Delegation, error) {
	if p.ID == "" {
		return nil, errors.New("delegation: ID is required")
	}
	if p.TenantID == "" {
		return nil, errors.New("delegation: TenantID is required")
	}
	if p.DelegatorID == "" {
		return nil, errors.New("delegation: DelegatorID is required")
	}
	if p.DelegateID == "" {
		return nil, errors.New("delegation: DelegateID is required")
	}
	if p.DelegatorID == p.DelegateID {
		return nil, ErrSelfDelegation
	}
	if p.StartsAt.IsZero() || p.EndsAt.IsZero() {
		return nil, errors.New("delegation: StartsAt and EndsAt are required")
	}
	if !p.EndsAt.After(p.StartsAt) {
		return nil, ErrInvalidDelegationWindow
	}
	if p.Reason == "" {
		return nil, errors.New("delegation: Reason is required")
	}
	if p.CreatedBy == "" {
		return nil, errors.New("delegation: CreatedBy is required")
	}
	if p.CreatedAt.IsZero() {
		return nil, errors.New("delegation: CreatedAt must not be zero")
	}

	return &Delegation{
		id:          p.ID,
		tenantID:    p.TenantID,
		delegatorID: p.DelegatorID,
		delegateID:  p.DelegateID,
		startsAt:    p.StartsAt,
		endsAt:      p.EndsAt,
		reason:      p.Reason,
		createdBy:   p.CreatedBy,
		createdAt:   p.CreatedAt,
	}, nil
}
