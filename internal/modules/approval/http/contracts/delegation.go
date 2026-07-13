package contracts

import (
	"fmt"
	"time"
)

// CreateDelegationRequest is the decoded body for the create-delegation endpoint.
// DelegatorID is deliberately absent — it is always the caller's own identity,
// derived server-side from the authenticated session (never client-supplied,
// F9/ADR 0077 §privilege-escalation guard).
type CreateDelegationRequest struct {
	DelegateID string    `json:"delegate_id"`
	StartsAt   time.Time `json:"starts_at"`
	EndsAt     time.Time `json:"ends_at"`
	Reason     string    `json:"reason"`
}

// Validate enforces DelegateID and Reason are non-empty. StartsAt/EndsAt
// window validity (ends_at > starts_at) and the self-delegation rule are
// domain invariants enforced by domain.NewDelegation, not duplicated here.
func (r CreateDelegationRequest) Validate() error {
	if err := validateRequired("delegate_id", r.DelegateID); err != nil {
		return wrapValidation(err)
	}
	if err := validateRequired("reason", r.Reason); err != nil {
		return wrapValidation(err)
	}
	if r.StartsAt.IsZero() {
		return wrapValidation(fmt.Errorf("starts_at is required"))
	}
	if r.EndsAt.IsZero() {
		return wrapValidation(fmt.Errorf("ends_at is required"))
	}
	return nil
}

// DelegationResponse is the response body for a successful create-delegation
// call, mirroring the OpenAPI ApprovalDelegation schema.
type DelegationResponse struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	DelegatorID string    `json:"delegator_id"`
	DelegateID  string    `json:"delegate_id"`
	StartsAt    time.Time `json:"starts_at"`
	EndsAt      time.Time `json:"ends_at"`
	Reason      string    `json:"reason"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}
