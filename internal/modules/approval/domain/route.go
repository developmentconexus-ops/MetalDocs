package domain

import (
	"errors"
	"fmt"

	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
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
	Selectors          []ActorSelector
}

// EffectiveSelectors returns the actor selectors that govern this stage during
// the flat-column → selector coexistence window. If explicit Selectors are set
// they win; otherwise a legacy stage's RequiredRole/AreaCode is synthesized
// into a single role_in_fixed_area selector — the in-memory mirror of the 0303
// backfill. A stage with neither returns nil (rejected by Route.Validate).
//
// TRANSITION bridge: deleted in slice 6 / migration 0304 when the flat
// RequiredRole/RequiredCapability/AreaCode columns drop and Selectors becomes
// the sole source of truth.
func (s Stage) EffectiveSelectors() []ActorSelector {
	if len(s.Selectors) >= 1 {
		return s.Selectors
	}
	if s.RequiredRole != "" {
		return []ActorSelector{{Kind: SelectorRoleInFixedArea, Role: s.RequiredRole, AreaCode: s.AreaCode}}
	}
	return nil
}

// Route is the per-profile approval route configuration.
type Route struct {
	ID          string
	TenantID    string
	ProfileCode string
	// Subject generalizes what this route governs (M3 kernel extraction, ADR
	// 0082 / P2.S2). For every route constructed by the current document code
	// path, Subject == {SubjectKindDocument, ProfileCode} by construction —
	// ProfileCode remains the canonical field for existing consumers, Subject
	// is the field the repository persists to subject_kind/subject_key.
	Subject Subject
	Version int
	Stages  []Stage
}

// hasApprovalStage reports whether the route contains at least one approval-kind
// (binding signature) stage. An empty Kind counts as approval: the DB column
// defaults to 'approval' (migration 0286) and every pre-G1 route/instance is an
// approval route, so an unset in-memory Kind must not be read as "review-only" —
// that keeps the controlado requirement fail-closed (never wrongly blocks a
// legitimate approval route whose Kind is defaulted at insert time).
func hasApprovalStage(stages []Stage) bool {
	for _, s := range stages {
		if s.Kind == StageKindApproval || s.Kind == "" {
			return true
		}
	}
	return false
}

// Validate enforces route structural invariants and the per-profile governance
// route-signature policy (G1). policy is the published-language consequence of
// the owning profile's governance class, resolved off-tx and passed in by the
// caller (Validate stays pure — it never reads taxonomy). An empty policy (unset,
// e.g. in structural-only unit tests) imposes no signature constraint.
func (r Route) Validate(policy taxonomydomain.RoutePolicy) error {
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

		// Every stage must name at least one way to resolve its eligible
		// actor pool (M4, unit 3.2), and every effective selector must be
		// internally consistent for its kind. EffectiveSelectors bridges the
		// coexistence window: an explicit-Selectors stage validates on those;
		// a legacy RequiredRole/AreaCode stage validates on the synthesized
		// role_in_fixed_area selector.
		effective := s.EffectiveSelectors()
		if len(effective) == 0 {
			return ErrStageNoSelector
		}
		for _, sel := range effective {
			if err := sel.Validate(); err != nil {
				return fmt.Errorf("stage %q: %w", s.Name, err)
			}
		}
	}

	// Per-profile governance route-signature policy (G1). Belt-and-suspenders to
	// the DB deferrable trigger; the switch is total over the published policy
	// values, with unknown/empty fail-closing to "no extra constraint" (the DB
	// trigger remains authoritative, so nothing weakens the signature guarantee).
	switch policy {
	case taxonomydomain.RoutePolicyNoRoutePermitted:
		// livre: no approval route may exist for the profile.
		return ErrRouteNotPermittedForProfile
	case taxonomydomain.RoutePolicyRequireApprovalStage:
		// controlado: the route must bind at least one signature stage.
		if !hasApprovalStage(r.Stages) {
			return ErrApprovalStageRequired
		}
	case taxonomydomain.RoutePolicyApprovalOptional, "":
		// simples (and unset in structural-only tests): review-only route allowed.
	}
	return nil
}
