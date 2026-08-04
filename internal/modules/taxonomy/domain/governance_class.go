package domain

// GovernanceClass classifies how strictly a document profile is governed. It
// is the taxonomy-owned knowledge-level type that drives the per-profile
// route-signature policy (see RoutePolicy). It is DB-mirrored by the
// document_profiles.governance_class column + CHECK constraint (migration 0295).
type GovernanceClass string

// GovernanceClass values. These mirror the DB CHECK exactly.
const (
	// GovernanceControlado — fully governed documents that MUST be signed
	// (POP, IT, desenho, FMEA). Their route requires ≥1 approval-kind stage.
	GovernanceControlado GovernanceClass = "controlado"
	// GovernanceSimples — reviewed-but-not-signed documents (nota fiscal,
	// orçamento, relatório de rotina). A review-only route is allowed.
	GovernanceSimples GovernanceClass = "simples"
	// GovernanceLivre — material that circulates without approval (rascunhos).
	// It is NOT route-less: config-first is universal (ADR 0087), so a livre
	// profile requires an explicitly configured ACTIVE route that carries ZERO
	// stages. Submitting against it completes instantly.
	GovernanceLivre GovernanceClass = "livre"
)

// Validate returns ErrInvalidGovernanceClass unless c is exactly one of the
// three known classes.
func (c GovernanceClass) Validate() error {
	switch c {
	case GovernanceControlado, GovernanceSimples, GovernanceLivre:
		return nil
	default:
		return ErrInvalidGovernanceClass
	}
}

// RoutePolicy is the narrow, approval-facing consequence of a GovernanceClass:
// the published-language value that crosses the approval↔taxonomy boundary. It
// carries no taxonomy internals — approval reasons only about these three
// route-shape outcomes, never about the class itself.
type RoutePolicy string

// RoutePolicy values.
const (
	// RoutePolicyRequireApprovalStage — the route must contain at least one
	// approval-kind (signature) stage. Derived from GovernanceControlado.
	RoutePolicyRequireApprovalStage RoutePolicy = "require_approval_stage"
	// RoutePolicyApprovalOptional — a review-only route is permitted. Derived
	// from GovernanceSimples.
	RoutePolicyApprovalOptional RoutePolicy = "approval_optional"
	// RoutePolicyNoApprovalStages — an ACTIVE route is required (config-first is
	// universal) and it MUST carry zero stages: submitting against it completes
	// instantly, with no approval burden. Derived from GovernanceLivre.
	// Replaces the pre-ADR-0087 RoutePolicyNoRoutePermitted, under which livre
	// profiles could own no route — and therefore no document or template — at
	// all (ADR 0087 supersedes the livre arm of ADR 0081).
	RoutePolicyNoApprovalStages RoutePolicy = "no_approval_stages"
)

// RoutePolicy derives the route-signature policy for this profile. It is pure
// and total over the three valid classes. An unset or unknown class
// fail-closes to RoutePolicyRequireApprovalStage — an integrity gate never
// silently drops the signature requirement on ambiguous input (no-fallback
// principle). The DB trigger (assert_route_shape, migration 0316) is the
// authoritative last line regardless.
func (p *DocumentProfile) RoutePolicy() RoutePolicy {
	switch p.GovernanceClass {
	case GovernanceSimples:
		return RoutePolicyApprovalOptional
	case GovernanceLivre:
		return RoutePolicyNoApprovalStages
	default:
		return RoutePolicyRequireApprovalStage
	}
}
