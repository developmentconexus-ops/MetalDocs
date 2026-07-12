// Package contracts holds the approval bounded-context's HTTP request/response
// DTOs and their Validate methods. These are the hand-maintained shapes consumed
// by the handlers in the sibling http package; the wire contract itself is owned
// by api/openapi and generated separately.
package contracts

// CancelRequest is the decoded body plus header-sourced fields for the cancel endpoint.
type CancelRequest struct {
	Reason         string `json:"reason"`
	IdempotencyKey string
	IfMatchVersion int
}

// Validate enforces that Reason is non-empty.
func (r CancelRequest) Validate() error {
	return wrapValidation(validateRequired("reason", r.Reason))
}

// CancelResponse is the response body for a successful cancel.
type CancelResponse struct {
	DocumentID string `json:"document_id"`
}
