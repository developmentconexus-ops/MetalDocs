package contracts

import "fmt"

// Verdict is the wire representation of a review-stage runtime verdict (F4).
type Verdict string

// Verdict values.
const (
	VerdictReady          Verdict = "ready"
	VerdictRequestChanges Verdict = "request_changes"
)

// ReviewVerdictRequest is the decoded body for the review-verdict endpoint.
type ReviewVerdictRequest struct {
	Verdict Verdict `json:"verdict"`
	Comment string  `json:"comment"`
}

// Validate enforces Verdict is ready or request_changes, and Comment is
// required when Verdict is request_changes (spec.md §2 consumer contract).
func (r ReviewVerdictRequest) Validate() error {
	switch r.Verdict {
	case VerdictReady:
	case VerdictRequestChanges:
		if err := validateRequired("comment", r.Comment); err != nil {
			return wrapValidation(err)
		}
	default:
		return wrapValidation(fmt.Errorf("verdict must be one of: ready, request_changes"))
	}
	return nil
}

// ReviewVerdictResponse is the response body for a successfully recorded verdict.
type ReviewVerdictResponse struct {
	VerdictID string `json:"verdict_id"`
	WasReplay bool   `json:"was_replay"`
	Outcome   string `json:"outcome"`
	// FastForwardEligible (R5, unit 2.3 G3) is true iff the verdict completed
	// the review stage AND the actor is eligible on the now-active approval
	// stage. On idempotent replay this is always false (not recomputed).
	FastForwardEligible bool `json:"fast_forward_eligible"`
	// NextStageID is set only when FastForwardEligible is true.
	NextStageID string `json:"next_stage_id,omitempty"`
}
