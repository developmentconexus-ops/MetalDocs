package contracts

// SupersedeRequest is the decoded body plus header-sourced fields for the supersede endpoint.
type SupersedeRequest struct {
	SupersededDocumentID string `json:"superseded_document_id"`
	IdempotencyKey       string
	IfMatchVersion       int
}

// Validate enforces SupersededDocumentID is a valid UUID.
func (r SupersedeRequest) Validate() error {
	return wrapValidation(validateUUID("superseded_document_id", r.SupersededDocumentID))
}

// SupersedeResponse is the response body for a successful supersede.
type SupersedeResponse struct {
	DocumentID   string `json:"document_id"`
	SupersededID string `json:"superseded_id"`
}
