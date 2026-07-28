package domain

import (
	"context"
	"time"
)

// AreaCapabilityReader is the narrow read surface cross-module consumers use to
// ask "in which process areas does this actor hold capability X?" without
// reaching across the module boundary into iam's owned grant tables
// (metaldocs.role_capabilities ⋈ public.user_process_areas). iam owns the
// capability model; callers depend on this port instead of issuing raw SQL, and
// they NEVER reason about role names (ADR 0022: capabilities, never roles).
//
// This is a read model for pre-flight UI narrowing — it is not an authorization
// decision. The binding decision stays with the two-tier PDP (tier-1 route
// capability + tier-2 authz.Require in-tx + DB tripwire); a caller that acted on
// a stale area list still gets denied at write time.
type AreaCapabilityReader interface {
	// AreasWithCapability returns the actor's capability reach in tenantID as of
	// now:
	//   - tenantWide is true when the actor is system_admin, i.e. the tier-2
	//     inheritance bypass authz.Require already honors (ADR 0022 R1). Such an
	//     actor holds the capability EVERYWHERE, so areaCodes (which only lists
	//     explicit area grants) must not be treated as the full reach.
	//   - areaCodes lists the distinct area codes with an active grant carrying
	//     capability, sorted ascending. Empty (non-nil) means no explicit grant.
	//
	// now is the caller's service clock (ADR 0022 Phase 7 — the active-membership
	// predicate uses the same clock as the rest of the request, not DB wall time).
	AreasWithCapability(ctx context.Context, tenantID, actorUserID, capability string, now time.Time) (tenantWide bool, areaCodes []string, err error)
}

// NoopAreaCapabilityReader is the explicit null-object for callers that never
// exercise a capability-reach read. It reports no system-admin bypass and no
// areas, so any consumer that narrows on it narrows to nothing — fail closed.
type NoopAreaCapabilityReader struct{}

// AreasWithCapability always returns (false, empty, nil).
func (NoopAreaCapabilityReader) AreasWithCapability(_ context.Context, _, _, _ string, _ time.Time) (bool, []string, error) {
	return false, []string{}, nil
}
