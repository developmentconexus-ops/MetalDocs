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
	EffectiveFrom        string     `json:"effective_from"`
	SupersededDocumentID string     `json:"superseded_document_id,omitempty"`
	EffectiveTo          *time.Time `json:"effective_to,omitempty"`
	ReviewDueAt          *time.Time `json:"review_due_at,omitempty"`
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
	// Friendly first-line mirrors of the 0274 DB CHECKs (ck_documents_effective_window,
	// ck_documents_review_due_sane) — the DB remains the enforced last line.
	if r.EffectiveTo != nil && !r.EffectiveTo.After(t) {
		return wrapValidation(fmt.Errorf("effective_to must be strictly after effective_from"))
	}
	if r.ReviewDueAt != nil && r.ReviewDueAt.Before(t) {
		return wrapValidation(fmt.Errorf("review_due_at must not be before effective_from"))
	}
	return nil
}

// PublishResponse is the response body for a successful publish or schedule-publish.
type PublishResponse struct {
	DocumentID    string `json:"document_id"`
	NewStatus     string `json:"new_status"`
	EffectiveFrom string `json:"effective_from,omitempty"`
}
