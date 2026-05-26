package contracts

import "fmt"

type Decision string

const (
	DecisionApprove Decision = "approve"
	DecisionReject  Decision = "reject"
)

type SignoffRequest struct {
	Decision       Decision `json:"decision"`
	Reason         string   `json:"reason"`
	PasswordToken  string   `json:"password_token"`
	ContentHash    string   `json:"content_hash"`
	IdempotencyKey string
}

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

type SignoffResponse struct {
	SignoffID string `json:"signoff_id"`
	WasReplay bool   `json:"was_replay"`
	Outcome   string `json:"outcome"`
}
