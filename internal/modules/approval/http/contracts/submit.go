package contracts

import (
	"fmt"
	"regexp"
	"strings"
	"time"
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

	// Publication plan (ADR 0085). Publication is not an endpoint — the release
	// coordinator releases an approved revision once every readiness fact holds.
	// The plan is the author's ONLY input to that decision and is declared here,
	// at submission, or not at all. Absent plan = release immediately on
	// readiness.
	//
	// PlannedEffectiveFrom is a string (not *time.Time) so a non-UTC offset can
	// be rejected rather than silently normalized — the same rule the retiring
	// schedule-publish contract enforced.
	PlannedEffectiveFrom string     `json:"planned_effective_from,omitempty"`
	EffectiveTo          *time.Time `json:"effective_to,omitempty"`
	ReviewDueAt          *time.Time `json:"review_due_at,omitempty"`
	// SupersededDocumentID names a CROSS-document supersede target: a published
	// document of a DIFFERENT controlled document this revision retires on
	// release. Same-controlled-document supersession is implicit (the
	// coordinator discovers the incumbent head) and naming it is rejected.
	SupersededDocumentID string `json:"superseded_document_id,omitempty"`
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
	return r.validatePlan()
}

// validatePlan is the friendly first line for the ADR 0085 publication plan.
// Every rule below mirrors a DB CHECK the release transaction still enforces
// against the ACTUAL release instant (ck_documents_effective_window,
// ck_documents_review_due_sane) — this is a courtesy 400, never the last line.
// The whole plan is optional; an entirely absent plan validates trivially.
func (r SubmitRequest) validatePlan() error {
	if strings.TrimSpace(r.SupersededDocumentID) != "" {
		if err := validateUUID("superseded_document_id", r.SupersededDocumentID); err != nil {
			return wrapValidation(err)
		}
	}

	planned := strings.TrimSpace(r.PlannedEffectiveFrom)
	if planned == "" {
		// No planned date: effective_to / review_due_at have no anchor to be
		// compared against here. Both remain individually legal and are
		// re-checked against the actual release instant in the DB.
		return nil
	}
	t, err := time.Parse(time.RFC3339, planned)
	if err != nil {
		return wrapValidation(fmt.Errorf("planned_effective_from must be parseable RFC3339: %w", err))
	}
	if _, offset := t.Zone(); offset != 0 {
		return wrapValidation(fmt.Errorf("planned_effective_from must be UTC"))
	}
	if r.EffectiveTo != nil && !r.EffectiveTo.After(t) {
		return wrapValidation(fmt.Errorf("effective_to must be strictly after planned_effective_from"))
	}
	if r.ReviewDueAt != nil && r.ReviewDueAt.Before(t) {
		return wrapValidation(fmt.Errorf("review_due_at must not be before planned_effective_from"))
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
