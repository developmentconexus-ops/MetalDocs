package infrastructure

import (
	"context"

	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
)

// taxonomyProfileGetter is the narrow read slice of the taxonomy profile
// repository the approval subsystem consumes for G1. The canonical taxonomy
// ProfileRepository satisfies it; depending on the slice keeps approval from
// coupling to taxonomy's write surface (interface segregation).
type taxonomyProfileGetter interface {
	GetByCode(ctx context.Context, tenantID string, code taxonomydomain.ProfileCode) (*taxonomydomain.DocumentProfile, error)
}

// ProfilePolicyReader adapts the canonical taxonomy profile repository to the
// approval application.ProfilePolicyReader port. It reads the profile through
// the taxonomy reader that enforces the authz GUC and CapTaxonomyView in
// taxonomy's own short tx (a separate connection — so the read never sits inside
// an approval lock-holding tx, H-PRE-1), then derives the narrow RoutePolicy.
// approval thus sees only the 3-valued policy, never GovernanceClass.
type ProfilePolicyReader struct {
	profiles taxonomyProfileGetter
}

// NewProfilePolicyReader builds a ProfilePolicyReader backed by profiles (the
// canonical taxonomy profile repository).
func NewProfilePolicyReader(profiles taxonomyProfileGetter) *ProfilePolicyReader {
	return &ProfilePolicyReader{profiles: profiles}
}

// RoutePolicy satisfies application.ProfilePolicyReader by fetching the profile
// and returning its derived route-signature policy. A read error (including a
// cross-tenant/not-found miss) is propagated unchanged so callers fail closed.
func (r *ProfilePolicyReader) RoutePolicy(ctx context.Context, tenantID, profileCode string) (taxonomydomain.RoutePolicy, error) {
	profile, err := r.profiles.GetByCode(ctx, tenantID, taxonomydomain.ProfileCode(profileCode))
	if err != nil {
		return "", err
	}
	return profile.RoutePolicy(), nil
}
