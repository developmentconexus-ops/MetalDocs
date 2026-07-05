package contracts

import (
	"fmt"
	"time"
)

// MarkReviewedRequest is the decoded body plus header-sourced fields for the
// mark-reviewed endpoint (POST /documents/{documentId}/review, F6.2).
type MarkReviewedRequest struct {
	ReviewDueAt    string     `json:"review_due_at"`
	EffectiveTo    *time.Time `json:"effective_to,omitempty"`
	IdempotencyKey string
	IfMatchVersion int
}

// Validate enforces ReviewDueAt is a parseable RFC3339 UTC timestamp.
func (r MarkReviewedRequest) Validate() error {
	if err := validateRequired("review_due_at", r.ReviewDueAt); err != nil {
		return wrapValidation(err)
	}
	t, err := time.Parse(time.RFC3339, r.ReviewDueAt)
	if err != nil {
		return wrapValidation(fmt.Errorf("review_due_at must be parseable RFC3339: %w", err))
	}
	_, offset := t.Zone()
	if offset != 0 {
		return wrapValidation(fmt.Errorf("review_due_at must be UTC"))
	}
	return nil
}

// ParsedReviewDueAt parses ReviewDueAt as RFC3339. Callers should call
// Validate first so the error path here is unreachable in practice.
func (r MarkReviewedRequest) ParsedReviewDueAt() (time.Time, error) {
	return time.Parse(time.RFC3339, r.ReviewDueAt)
}

// MarkReviewedResponse is the response body for a successful mark-reviewed.
// Mirrors PublishResponse's shape (document_id + new_status); new_status is
// always "published" since mark-reviewed never changes status.
type MarkReviewedResponse struct {
	DocumentID string `json:"document_id"`
	NewStatus  string `json:"new_status"`
}
