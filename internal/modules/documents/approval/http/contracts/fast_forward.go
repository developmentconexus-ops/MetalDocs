package contracts

// FastForwardRequest is the decoded body for the fast-forward ("Aprovar já")
// endpoint. It carries the same e-signature ceremony fields as SignoffRequest
// (PasswordToken, ContentHash) plus an optional Comment recorded on the
// `ready` review-verdict leg — there is no Decision field: fast-forward
// always records an approve on both legs.
type FastForwardRequest struct {
	Comment       string `json:"comment"`
	PasswordToken string `json:"password_token"`
	ContentHash   string `json:"content_hash"`
}

// Validate enforces PasswordToken is non-empty and ContentHash is 64 hex
// chars, mirroring SignoffRequest.Validate's ceremony-field checks.
func (r FastForwardRequest) Validate() error {
	if err := validateRequired("password_token", r.PasswordToken); err != nil {
		return wrapValidation(err)
	}
	if err := validateSHA256Hex("content_hash", r.ContentHash); err != nil {
		return wrapValidation(err)
	}
	return nil
}

// FastForwardResponse is the response body for a successful fast-forward.
type FastForwardResponse struct {
	SignoffID string `json:"signoff_id"`
	WasReplay bool   `json:"was_replay"`
	Outcome   string `json:"outcome"`
}
