package application

import (
	"context"

	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
)

// ProfilePolicyReader is the narrow approval→taxonomy read port (G1). Given a
// tenant and a profile code it returns that profile's route-signature policy —
// the published-language consequence of its governance class. It is the ONLY
// taxonomy surface the approval subsystem consumes: approval never sees
// GovernanceClass or any other taxonomy internal, only the 3-valued RoutePolicy.
//
// Mirrors the interface-segregation pattern in
// controlleddocuments/infrastructure/taxonomy_reader.go. The production adapter
// resolves the policy in taxonomy's own short tx (a separate connection, so the
// read never sits inside an approval lock-holding tx — H-PRE-1) while preserving
// the taxonomy authz GUC + CapTaxonomyView.
type ProfilePolicyReader interface {
	RoutePolicy(ctx context.Context, tenantID, profileCode string) (taxonomydomain.RoutePolicy, error)
}
