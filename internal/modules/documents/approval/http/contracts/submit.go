package contracts

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	uuidPattern   = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	sha256Pattern = regexp.MustCompile(`(?i)^[0-9a-f]{64}$`)
)

// SubmitRequest is the decoded body plus header-sourced fields for the submit-for-review endpoint.
type SubmitRequest struct {
	RouteID        string `json:"route_id"`
	IdempotencyKey string
	ContentHash    string `json:"content_hash"`

	// ReasonForChange (+ optional ReasonCategory) is the F6.3 structured
	// 21 CFR Part 11 change reason. Distinct from any revision title; required
	// for REV>=1 (enforced downstream by application.SubmitRevisionForReview,
	// which knows the governed revision number), optional at REV 0. Both
	// pointers so an absent field is distinguishable from an explicit "".
	ReasonForChange *string `json:"reason_for_change,omitempty"`
	ReasonCategory  *string `json:"reason_category,omitempty"`
}

// Validate enforces RouteID is a valid UUID and ContentHash is 64 hex characters.
// reason_for_change/reason_category REV>=1-requiredness is enforced downstream
// (the service knows the governed revision number; this contract-level Validate
// does not).
func (r SubmitRequest) Validate() error {
	if err := validateUUID("route_id", r.RouteID); err != nil {
		return wrapValidation(err)
	}
	if err := validateSHA256Hex("content_hash", r.ContentHash); err != nil {
		return wrapValidation(err)
	}
	return nil
}

// SubmitResponse is the response body for a successful submit-for-review.
type SubmitResponse struct {
	InstanceID string `json:"instance_id"`
	WasReplay  bool   `json:"was_replay"`
	ETag       string `json:"etag"`
}

func validateUUID(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", field)
	}
	if !uuidPattern.MatchString(value) {
		return fmt.Errorf("%s must be a valid UUID", field)
	}
	return nil
}

func validateSHA256Hex(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", field)
	}
	if !sha256Pattern.MatchString(value) {
		return fmt.Errorf("%s must be 64 hex characters", field)
	}
	return nil
}

func validateRequired(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", field)
	}
	return nil
}
