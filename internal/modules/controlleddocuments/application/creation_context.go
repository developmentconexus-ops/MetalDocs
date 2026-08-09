package application

import (
	"context"
	"errors"
	"fmt"

	approvaldomain "metaldocs/internal/modules/approval/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/authn"
)

// ProfileLister is the cross-module catalog-read port into taxonomy used to
// build the creation context. Segregated from ProfileReader (per-code lookup)
// so neither port's implementers carry methods they do not need.
type ProfileLister interface {
	ListProfiles(ctx context.Context, tenantID string) ([]taxonomydomain.DocumentProfile, error)
}

// AreaLister is the cross-module catalog-read port into taxonomy for process
// areas. See ProfileLister.
type AreaLister interface {
	ListAreas(ctx context.Context, tenantID string) ([]taxonomydomain.ProcessArea, error)
}

// ErrCreationContextUnconfigured is returned when CreationContext runs without
// its readers wired — a composition-root bug, not a user error. It fails closed
// rather than returning a partially-computed (and therefore misleading)
// authorization view.
var ErrCreationContextUnconfigured = errors.New("controlled_documents: creation context readers not configured")

// CreationContextProfile is one selectable document profile plus its approval
// readiness. HasActiveRoute false means creation under this profile is rejected
// with the same 409 the submit path emits (state.approval_route_missing).
type CreationContextProfile struct {
	Code           string
	Name           string
	HasActiveRoute bool
}

// CreationContextArea is one process area the caller may create into.
type CreationContextArea struct {
	Code string
	Name string
}

// CreationContext is the pre-flight read model for the create-document form.
type CreationContext struct {
	Profiles []CreationContextProfile
	Areas    []CreationContextArea
}

// CreationContext composes the create form's pre-flight data: the tenant's
// active profiles annotated with approval-route readiness, and the process
// areas the CALLER may actually create into.
//
// The area narrowing is computed server-side from the caller's capability
// grants (iamdomain.AreaCapabilityReader) intersected with taxonomy's active
// area catalog — the client never receives the full catalog to filter. The
// intersection is a Go-level join on purpose: capability grants are iam-owned
// and the area catalog is taxonomy-owned, so a single SQL join across both
// would breach the module boundary. system_admin (tenantWide) short-circuits to
// the whole active catalog, mirroring the tier-2 inheritance bypass
// authz.Require already honors (ADR 0022 R1).
//
// This endpoint deliberately does NOT call authz.Require: its response body IS
// the authorization computation, and Require needs a concrete area — which is
// precisely what the caller is asking the server to discover. Tier-1 already
// gates the route on controlled_documents.create, and every state change still
// passes tier-2 + the DB tripwire, so this read grants nothing.
//
// Runs entirely off-tx (three plain reads); nothing here takes the audit
// advisory lock, so HS-PRE-1 does not apply.
func (s *ControlledDocumentService) CreationContext(ctx context.Context, tenantID string) (*CreationContext, error) {
	actorUserID, ok := authn.UserIDFromContext(ctx)
	if !ok {
		return nil, ErrActorMissing
	}
	if s.profileList == nil || s.areaList == nil || s.areaCaps == nil || s.routes == nil {
		return nil, ErrCreationContextUnconfigured
	}

	profiles, err := s.profileList.ListProfiles(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("controlled_documents: list profiles for creation context: %w", err)
	}
	readyKeys, err := s.routes.ActiveRouteSubjectKeys(ctx, tenantID, string(approvaldomain.SubjectKindDocument))
	if err != nil {
		return nil, fmt.Errorf("controlled_documents: read approval route readiness: %w", err)
	}

	areas, err := s.areaList.ListAreas(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("controlled_documents: list areas for creation context: %w", err)
	}
	tenantWide, grantedAreas, err := s.areaCaps.AreasWithCapability(
		ctx, tenantID, actorUserID, string(iamdomain.CapControlledDocumentCreate), s.now().UTC(),
	)
	if err != nil {
		return nil, fmt.Errorf("controlled_documents: read caller create-capability areas: %w", err)
	}

	return &CreationContext{
		Profiles: creationContextActiveProfiles(profiles, readyKeys),
		Areas:    creationContextVisibleAreas(areas, tenantWide, grantedAreas),
	}, nil
}

// creationContextActiveProfiles filters profiles to the active ones,
// annotated with their approval-route readiness from readyKeys. Extracted
// from CreationContext; behavior unchanged.
func creationContextActiveProfiles(profiles []taxonomydomain.DocumentProfile, readyKeys map[string]struct{}) []CreationContextProfile {
	out := make([]CreationContextProfile, 0, len(profiles))
	for _, p := range profiles {
		if !p.IsActive() {
			continue
		}
		_, ready := readyKeys[string(p.Code)]
		out = append(out, CreationContextProfile{
			Code:           string(p.Code),
			Name:           p.Name,
			HasActiveRoute: ready,
		})
	}
	return out
}

// creationContextVisibleAreas filters areas to the active ones the caller
// may create into: the full active set when tenantWide, otherwise narrowed
// to grantedAreas. Extracted from CreationContext; behavior unchanged.
func creationContextVisibleAreas(areas []taxonomydomain.ProcessArea, tenantWide bool, grantedAreas []string) []CreationContextArea {
	granted := make(map[string]struct{}, len(grantedAreas))
	for _, code := range grantedAreas {
		granted[code] = struct{}{}
	}
	out := make([]CreationContextArea, 0)
	for _, a := range areas {
		if !a.IsActive() {
			continue
		}
		if !tenantWide {
			if _, allowed := granted[string(a.Code)]; !allowed {
				continue
			}
		}
		out = append(out, CreationContextArea{Code: string(a.Code), Name: a.Name})
	}
	return out
}
