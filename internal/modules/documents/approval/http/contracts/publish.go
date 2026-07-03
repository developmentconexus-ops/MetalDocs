package contracts

import (
	"fmt"
	"time"
)

// PublishRequest is populated from HTTP headers, not decoded from the request body.
type PublishRequest struct {
	IdempotencyKey string
	IfMatchVersion int
}

// SchedulePublishRequest is the decoded body plus header-sourced fields for the
// schedule-publish endpoint.
type SchedulePublishRequest struct {
	EffectiveFrom        string `json:"effective_from"`
	SupersededDocumentID string `json:"superseded_document_id,omitempty"`
	IdempotencyKey       string
	IfMatchVersion       int
}

// Validate enforces EffectiveFrom is a UTC RFC3339 timestamp and, when present,
// SupersededDocumentID is a valid UUID.
func (r SchedulePublishRequest) Validate() error {
	if err := validateRequired("effective_from", r.EffectiveFrom); err != nil {
		return wrapValidation(err)
	}
	t, err := time.Parse(time.RFC3339, r.EffectiveFrom)
	if err != nil {
		return wrapValidation(fmt.Errorf("effective_from must be parseable RFC3339: %w", err))
	}
	_, offset := t.Zone()
	if offset != 0 {
		return wrapValidation(fmt.Errorf("effective_from must be UTC"))
	}
	if r.SupersededDocumentID != "" {
		if err := validateUUID("superseded_document_id", r.SupersededDocumentID); err != nil {
			return wrapValidation(err)
		}
	}
	return nil
}

// PublishResponse is the response body for a successful publish or schedule-publish.
type PublishResponse struct {
	DocumentID    string `json:"document_id"`
	NewStatus     string `json:"new_status"`
	EffectiveFrom string `json:"effective_from,omitempty"`
}
