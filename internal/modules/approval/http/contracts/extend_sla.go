package contracts

import (
	"fmt"
	"time"
)

// ExtendSLARequest is the decoded body for the extend-sla endpoint
// (POST /approval/instances/{instance_id}/extend-sla, Task 9). No If-Match:
// the forward-only rule is enforced by a read and a write on the same row
// inside one transaction, so an OCC precondition would be contract surface
// solving a problem the transaction already solves.
type ExtendSLARequest struct {
	DueAt          string `json:"due_at"`
	Reason         string `json:"reason"`
	IdempotencyKey string
}

// Validate enforces DueAt is a parseable RFC3339 UTC timestamp and Reason is
// non-blank. A blank (or whitespace-only) reason is caught here as a
// friendly first-line rejection; the service enforces the same rule again
// (trimmed) as the authoritative gate.
func (r ExtendSLARequest) Validate() error {
	if err := validateRequired("due_at", r.DueAt); err != nil {
		return wrapValidation(err)
	}
	t, err := time.Parse(time.RFC3339, r.DueAt)
	if err != nil {
		return wrapValidation(fmt.Errorf("due_at must be parseable RFC3339: %w", err))
	}
	if _, offset := t.Zone(); offset != 0 {
		return wrapValidation(fmt.Errorf("due_at must be UTC"))
	}
	if err := validateRequired("reason", r.Reason); err != nil {
		return wrapValidation(err)
	}
	return nil
}

// DueAtTime parses DueAt as RFC3339. Callers should call Validate first so
// the error path here is unreachable in practice.
func (r ExtendSLARequest) DueAtTime() (time.Time, error) {
	return time.Parse(time.RFC3339, r.DueAt)
}
