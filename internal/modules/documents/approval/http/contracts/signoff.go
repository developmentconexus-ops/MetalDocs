package contracts

import "fmt"

// Decision is the wire representation of an approve/reject vote.
type Decision string

// Decision values.
const (
	DecisionApprove Decision = "approve"
	DecisionReject  Decision = "reject"
)

// SignoffRequest is the decoded body plus header-sourced fields for the signoff
// endpoint. PasswordToken carries the e-signature re-authentication credential
// (21 CFR Part 11); it is never persisted, only verified.
type SignoffRequest struct {
	Decision       Decision `json:"decision"`
	Reason         string   `json:"reason"`
	PasswordToken  string   `json:"password_token"`
	ContentHash    string   `json:"content_hash"`
	IdempotencyKey string
}

// Validate enforces Decision is approve or reject, Reason is required when
// rejecting, PasswordToken is non-empty, and ContentHash is 64 hex chars.
func (r SignoffRequest) Validate() error {
	switch r.Decision {
	case DecisionApprove:
	case DecisionReject:
	default:
		return wrapValidation(fmt.Errorf("decision must be one of: approve, reject"))
	}
	if r.Decision == DecisionReject {
		if err := validateRequired("reason", r.Reason); err != nil {
			return wrapValidation(err)
		}
	}
	if err := validateRequired("password_token", r.PasswordToken); err != nil {
		return wrapValidation(err)
	}
	if err := validateSHA256Hex("content_hash", r.ContentHash); err != nil {
		return wrapValidation(err)
	}
	return nil
}

// SignoffResponse is the response body for a successful signoff.
type SignoffResponse struct {
	SignoffID string `json:"signoff_id"`
	WasReplay bool   `json:"was_replay"`
	Outcome   string `json:"outcome"`
}
