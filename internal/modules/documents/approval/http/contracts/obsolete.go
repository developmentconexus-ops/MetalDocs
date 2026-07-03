package contracts

// ObsoleteRequest is the decoded body plus header-sourced fields for the mark-obsolete endpoint.
type ObsoleteRequest struct {
	Reason         string `json:"reason"`
	IdempotencyKey string
	IfMatchVersion int
}

// Validate enforces that Reason is non-empty.
func (r ObsoleteRequest) Validate() error {
	return wrapValidation(validateRequired("reason", r.Reason))
}

// ObsoleteResponse is the response body for a successful mark-obsolete.
type ObsoleteResponse struct {
	DocumentID string `json:"document_id"`
}
