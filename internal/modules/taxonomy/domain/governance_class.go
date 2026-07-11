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
	// GovernanceLivre — ungoverned material (rascunhos). No route is permitted.
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
	// RoutePolicyNoRoutePermitted — no approval route may exist for the
	// profile. Derived from GovernanceLivre.
	RoutePolicyNoRoutePermitted RoutePolicy = "no_route_permitted"
)

// RoutePolicy derives the route-signature policy for this profile. It is pure
// and total over the three valid classes. An unset or unknown class
// fail-closes to RoutePolicyRequireApprovalStage — an integrity gate never
// silently drops the signature requirement on ambiguous input (no-fallback
// principle). The DB trigger (migration 0295) is the authoritative last line
// regardless.
func (p *DocumentProfile) RoutePolicy() RoutePolicy {
	switch p.GovernanceClass {
	case GovernanceSimples:
		return RoutePolicyApprovalOptional
	case GovernanceLivre:
		return RoutePolicyNoRoutePermitted
	default:
		return RoutePolicyRequireApprovalStage
	}
}
