package domain

import (
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
	RequiredCapability string
	Quorum             QuorumPolicy
	QuorumM            *int
	OnEligibilityDrift DriftPolicy
	Kind               StageKind
	DueInDays          *int
	Selectors          []ActorSelector
}

// Route is the per-profile approval route configuration.
type Route struct {
	ID string
	// Name is the route's configured display name (approval_routes.name).
	// Populated by LoadRoute (SLICE 8a, unit 3.2) for read paths that surface
	// it (e.g. RoutePreview); not read or asserted by the submit-time
	// resolution/validation path, which never needed it.
	Name        string
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
// caller (Validate stays pure — it never reads taxonomy).
//
// Stage-count is a POLICY question, not a structural one (ADR 0087): a livre
// profile's route is a configured ZERO-stage route, so "at least one stage"
// cannot be a blanket precondition any more. An empty policy (structural-only,
// e.g. submit's re-validation of an already-bound active route and unit tests)
// therefore imposes no stage-count and no signature constraint — it checks only
// that whatever stages exist are internally well-formed.
func (r Route) Validate(policy taxonomydomain.RoutePolicy) error {
	names := make(map[string]bool, len(r.Stages))
	for i, s := range r.Stages {
		if err := validateStageStructure(s, i, names); err != nil {
			return err
		}
	}
	return validateRouteShapePolicy(policy, r.Stages)
}

// validateStageStructure enforces one stage's internal invariants: dense
// ordering, Quorum/QuorumM consistency, name uniqueness within the route
// (via the shared names set), and that every selector is internally
// consistent for its kind.
func validateStageStructure(s Stage, i int, names map[string]bool) error {
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
	} else if s.QuorumM != nil {
		return fmt.Errorf("stage %q: QuorumM must be nil for quorum %s", s.Name, s.Quorum)
	}

	// Unique names.
	if names[s.Name] {
		return fmt.Errorf("duplicate stage name %q in route", s.Name)
	}
	names[s.Name] = true

	// Every stage must name at least one way to resolve its eligible
	// actor pool (M4, unit 3.2), and every selector must be internally
	// consistent for its kind. Selectors is the sole source of truth
	// post-slice-6b; the flat RequiredRole/AreaCode → selector synthesis
	// happens at the HTTP boundary (route_admin_handler.go) before a
	// Stage ever reaches domain code.
	effective := s.Selectors
	if len(effective) == 0 {
		return ErrStageNoSelector
	}
	for _, sel := range effective {
		if err := sel.Validate(); err != nil {
			return fmt.Errorf("stage %q: %w", s.Name, err)
		}
	}
	return nil
}

// validateRouteShapePolicy enforces the per-profile governance route-shape
// policy (G1, reworked by ADR 0087). Belt-and-suspenders to the DB
// deferrable trigger; the switch is total over the published policy values,
// with unknown/empty fail-closing to "no extra constraint" (the DB trigger
// remains authoritative, so nothing weakens the signature guarantee).
func validateRouteShapePolicy(policy taxonomydomain.RoutePolicy, stages []Stage) error {
	switch policy {
	case taxonomydomain.RoutePolicyNoApprovalStages:
		// livre: the route is required but must carry ZERO stages — its whole
		// point is that submitting against it completes instantly.
		if len(stages) > 0 {
			return ErrRouteStagesNotPermittedForProfile
		}
	case taxonomydomain.RoutePolicyRequireApprovalStage:
		// controlado: the route must bind at least one signature stage.
		if !hasApprovalStage(stages) {
			return ErrApprovalStageRequired
		}
	case taxonomydomain.RoutePolicyApprovalOptional:
		// simples: review-only route allowed, but a stageless route is not — a
		// simples profile is reviewed, just not signed (pre-ADR-0087 semantics,
		// preserved exactly: the old blanket structural check enforced this).
		if len(stages) == 0 {
			return ErrRouteStageRequired
		}
	case "":
		// Structural-only: no stage-count and no signature constraint.
	}
	return nil
}
