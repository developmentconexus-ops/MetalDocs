package domain

import (
	"errors"
	"fmt"
)

// QuorumPolicy defines how many signoffs satisfy a stage.
type QuorumPolicy string

// QuorumPolicy values understood by EvaluateQuorumResult.
const (
	QuorumAny1Of QuorumPolicy = "any_1_of"
	QuorumAllOf  QuorumPolicy = "all_of"
	QuorumMofN   QuorumPolicy = "m_of_n"
)

// DriftPolicy defines behavior when eligible actors change after stage opens.
type DriftPolicy string

// DriftPolicy values understood by ApplyEligibilityDrift.
const (
	DriftReduceQuorum DriftPolicy = "reduce_quorum"
	DriftFailStage    DriftPolicy = "fail_stage"
	DriftKeepSnapshot DriftPolicy = "keep_snapshot"
)

// StageKind distinguishes a review stage (comment/feedback, no approve/reject
// authority) from an approval stage (binding sign-off). F1 (milestone 2b)
// substrate only — no service wiring; every existing route/instance defaults
// to StageKindApproval so current behavior is unchanged.
type StageKind string

// StageKind values understood by Validate. These mirror the DB CHECK
// constraint added by migration 0286 exactly.
const (
	StageKindReview   StageKind = "review"
	StageKindApproval StageKind = "approval"
)

// Validate returns ErrInvalidStageKind unless k is exactly StageKindReview or
// StageKindApproval.
func (k StageKind) Validate() error {
	switch k {
	case StageKindReview, StageKindApproval:
		return nil
	default:
		return ErrInvalidStageKind
	}
}

// Stage is a single step in an approval route.
type Stage struct {
	Order              int
	Name               string
	RequiredRole       string
	RequiredCapability string
	AreaCode           string
	Quorum             QuorumPolicy
	QuorumM            *int
	OnEligibilityDrift DriftPolicy
	Kind               StageKind
	DueInDays          *int
}

// Route is the per-profile approval route configuration.
type Route struct {
	ID          string
	TenantID    string
	ProfileCode string
	Version     int
	Stages      []Stage
}

// Validate enforces route structural invariants.
func (r Route) Validate() error {
	if len(r.Stages) == 0 {
		return errors.New("route must have at least one stage")
	}

	names := make(map[string]bool, len(r.Stages))
	for i, s := range r.Stages {
		// Dense order starting at 1.
		if s.Order != i+1 {
			return fmt.Errorf("stage order must be dense starting at 1: stage at index %d has order %d, expected %d", i, s.Order, i+1)
		}

		// Quorum + QuorumM consistency.
		if s.Quorum == QuorumMofN {
			if s.QuorumM == nil {
				return fmt.Errorf("stage %q: quorum m_of_n requires QuorumM", s.Name)
			}
			if *s.QuorumM < 1 {
				return fmt.Errorf("stage %q: QuorumM must be >= 1", s.Name)
			}
		} else {
			if s.QuorumM != nil {
				return fmt.Errorf("stage %q: QuorumM must be nil for quorum %s", s.Name, s.Quorum)
			}
		}

		// Unique names.
		if names[s.Name] {
			return fmt.Errorf("duplicate stage name %q in route", s.Name)
		}
		names[s.Name] = true
	}
	return nil
}
