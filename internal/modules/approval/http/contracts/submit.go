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
//
// ADR 0073: route_id and content_hash are OPTIONAL. When omitted the submit
// service resolves both server-side, in-tx (active route by profile, head
// content hash) — an author cannot supply either (route listing needs an admin
// read the author lacks; the hash is server-authoritative). Explicit values,
// when present, keep their exact prior semantics (integration callers).
type SubmitRequest struct {
	RouteID        string `json:"route_id"`
	IdempotencyKey string
	ContentHash    string `json:"content_hash"`

	// RevisionTitle is the human revision label (was finalize-only, ADR 0073).
	// Optional at REV 0 (service defaults it); required for REV>=1 — enforced
	// downstream by application.SubmitRevisionForReview against the governed
	// revision number, not here.
	RevisionTitle string `json:"revision_title"`

	// ReasonForChange (+ optional ReasonCategory) is the F6.3 structured
	// 21 CFR Part 11 change reason. Distinct from any revision title; required
	// for REV>=1 (enforced downstream by application.SubmitRevisionForReview,
	// which knows the governed revision number), optional at REV 0. Both
	// pointers so an absent field is distinguishable from an explicit "".
	ReasonForChange *string `json:"reason_for_change,omitempty"`
	ReasonCategory  *string `json:"reason_category,omitempty"`

	// ChosenActors (M4, unit 3.2, slice 5) carries per-stage caller-chosen
	// actors for stages governed by a submit_choice selector. Optional at the
	// wire level — omitted or empty is legal for routes with no
	// submit_choice stage; the application service fails closed
	// (ErrSubmitChoiceRequired / ErrSubmitChoiceConstraintViolated) when a
	// submit_choice stage is present and the entry is missing, empty, or
	// targets the wrong stage.
	ChosenActors []SubmitChosenActors `json:"chosen_actors,omitempty"`
}

// SubmitChosenActors is the wire shape of one chosen_actors entry: the
// stage_order it targets plus the chosen user_ids.
type SubmitChosenActors struct {
	StageOrder int      `json:"stage_order"`
	UserIDs    []string `json:"user_ids"`
}

// Validate is format-only and optionality-aware (ADR 0073): route_id and
// content_hash are validated ONLY when non-empty (a valid UUID / 64 hex chars);
// an empty value is legal and signals server-side in-tx resolution. The REV>=1
// requiredness of revision_title / reason_for_change is enforced downstream (the
// service knows the governed revision number; this contract-level Validate does not).
func (r SubmitRequest) Validate() error {
	if strings.TrimSpace(r.RouteID) != "" {
		if err := validateUUID("route_id", r.RouteID); err != nil {
			return wrapValidation(err)
		}
	}
	if strings.TrimSpace(r.ContentHash) != "" {
		if err := validateSHA256Hex("content_hash", r.ContentHash); err != nil {
			return wrapValidation(err)
		}
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
