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
	DueAt  string `json:"due_at"`
	Reason string `json:"reason"`

	// parsedDueAt caches the RFC3339 parse performed by Validate, so
	// DueAtTime never re-parses (and never has a reachable error path).
	parsedDueAt time.Time
}

// Validate enforces DueAt is a parseable RFC3339 timestamp. openapi.yaml
// declares due_at as `format: date-time`, which admits any UTC offset (an
// offset and UTC denote the same instant); the service normalizes with
// .UTC() before writing, so no additional offset restriction is enforced
// here — doing so would reject requests the published contract promises to
// accept.
//
// Reason is intentionally NOT checked here. A blank reason must surface as
// the registered 422 validation.sla_extension_reason_required
// (application.ErrSLAExtensionReasonRequired) rather than the generic 400
// validation.request_invalid every other field failure produces here, and
// this package does not depend on the application package's error
// sentinels. ExtendSLAHandler checks Reason directly against
// application.ErrSLAExtensionReasonRequired instead.
func (r *ExtendSLARequest) Validate() error {
	if err := validateRequired("due_at", r.DueAt); err != nil {
		return wrapValidation(err)
	}
	t, err := time.Parse(time.RFC3339, r.DueAt)
	if err != nil {
		return wrapValidation(fmt.Errorf("due_at must be parseable RFC3339: %w", err))
	}
	r.parsedDueAt = t
	return nil
}

// DueAtTime returns the RFC3339 timestamp parsed by Validate. Callers must
// call Validate first.
func (r ExtendSLARequest) DueAtTime() time.Time {
	return r.parsedDueAt
}
