// routes_memberships.go — IAM area-membership HTTP wiring (PR-1 hardening).
//
// Surface matches OpenAPI ops listAreaMemberships / grantAreaMembership /
// revokeAreaMembership (api/openapi/v1/openapi.yaml). Hand-rolled rather than
// codegen-served — IAM is still pre-codegen on the BE side per ADR 0012
// partial rollout; mirrors PeopleHandler's hand-rolled-with-spec-tagged pattern.
//
// Hardening added in PR-1:
//   - cross-tenant 404 via MembershipUserTenantVerifier (mirrors PeopleHandler
//     guardUserInTenant — cross-tenant probes must not distinguish "exists
//     elsewhere" from "does not exist").
//   - duplicate-grant 409 via iamapp.ErrMembershipExists (same role on active
//     row → revoke first).
//
// Audit: grant/revoke governance events are written in-tx by
// AreaMembershipService (AuditMembershipLogger), not here — a post-commit
// emission would duplicate that row (H-3a).
package httpdelivery

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	iamapi "metaldocs/internal/modules/iam/api"
	iamapp "metaldocs/internal/modules/iam/application"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

// MembershipUserTenantVerifier is the narrow interface MembershipHandler needs
// to enforce tenant scoping on the target user. Satisfied by *iamapp.PeopleService
// (VerifyUserInTenant). Kept as a local interface so MembershipHandler does
// not transitively pull in the entire PeopleService surface.
type MembershipUserTenantVerifier interface {
	VerifyUserInTenant(ctx context.Context, tenantID, userID string) error
}

// MembershipHandler serves the /iam/area-memberships list/grant/revoke
// surface. Authorization is tier-1 (route capability) plus tier-2
// (AreaMembershipService's in-repo authz.Require on the target area); the
// handler itself enforces only the self-grant lockout and the cross-tenant
// 404 guard (ADR 0022 Phase 3/4 — no role-string literals here).
type MembershipHandler struct {
	svc      *iamapp.AreaMembershipService
	verifier MembershipUserTenantVerifier
}

type grantMembershipRequest struct {
	UserID   string `json:"user_id"`
	AreaCode string `json:"area_code"`
	Role     string `json:"role"`
}

type membershipDTO struct {
	UserID        string  `json:"user_id"`
	TenantID      string  `json:"tenant_id"`
	AreaCode      string  `json:"area_code"`
	Role          string  `json:"role"`
	EffectiveFrom string  `json:"effective_from"`
	EffectiveTo   *string `json:"effective_to"`
	GrantedBy     *string `json:"granted_by"`
}

type listMembershipsResponse struct {
	Items []membershipDTO `json:"items"`
}

func toMembershipDTO(m iamdomain.UserProcessArea) membershipDTO {
	dto := membershipDTO{
		UserID:        m.UserID,
		TenantID:      m.TenantID,
		AreaCode:      m.AreaCode,
		Role:          string(m.Role),
		EffectiveFrom: m.EffectiveFrom.Format(time.RFC3339),
		GrantedBy:     m.GrantedBy,
	}
	if m.EffectiveTo != nil {
		s := m.EffectiveTo.Format(time.RFC3339)
		dto.EffectiveTo = &s
	}
	return dto
}

// NewMembershipHandler constructs the handler. Both svc and verifier should
// be non-nil in production; a nil svc causes every route to answer 501, and
// a nil verifier fails the tenant guard closed with 501 (see
// guardMembershipUserInTenant).
func NewMembershipHandler(svc *iamapp.AreaMembershipService, verifier MembershipUserTenantVerifier) *MembershipHandler {
	return &MembershipHandler{svc: svc, verifier: verifier}
}

// listMemberships — operationId listAreaMemberships.
func (h *MembershipHandler) listMemberships(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		problem.Respond(w, r, problem.New(http.StatusNotImplemented, problem.CodeInternalUnknown, "Membership service is not configured"))
		return
	}

	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error"))
		return
	}

	userID := strings.TrimSpace(r.URL.Query().Get("user_id"))
	areaCode := strings.TrimSpace(r.URL.Query().Get("area_code"))
	role := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("role")))

	// Directory scope (ADR 0022 Phase 4 — capability/area-aware, resolved at the
	// data layer so the handler never pivots on a role-name string, ADR 0021).
	// membership.view is held by every role, so a tenant-wide directory must be
	// gated tighter than the route-level view gate. Three tiers:
	//   - tenant-wide  → system_admin (tier-2 inheritance bypass, R1): full
	//     tenant directory with optional filters.
	//   - managed-areas → area_admin (holds membership.manage in ≥1 area):
	//     memberships only in their managed areas, filtered IN SQL (R3).
	//   - self-only    → every other role: own memberships regardless of the
	//     userId filter they pass.
	userID, tenantWide, hasManagedAreas, actor, ok := h.resolveMembershipsListScope(w, r, tenantID, userID)
	if !ok {
		return
	}

	var items []iamdomain.UserProcessArea
	if !tenantWide && hasManagedAreas {
		// area_admin: managed-area restriction enforced in SQL (R3).
		items, err = h.svc.ListByTenantInManagedAreas(r.Context(), tenantID, userID, areaCode, role, actor, string(iamdomain.CapMembershipManage))
	} else {
		// tenant-wide (system_admin) or self-only (userID pinned to actor above).
		items, err = h.svc.ListByTenant(r.Context(), tenantID, userID, areaCode, role)
	}
	if err != nil {
		slog.Error("iam memberships: list failed", "err", err)
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "Failed to list memberships"))
		return
	}
	dtos := make([]membershipDTO, 0, len(items))
	for _, m := range items {
		dtos = append(dtos, toMembershipDTO(m))
	}
	writeJSON(w, http.StatusOK, listMembershipsResponse{Items: dtos})
}

// resolveMembershipsListScope resolves the requesting actor's directory
// scope for listMemberships, applies the self-only userID restriction when
// the actor holds neither tenant-wide nor managed-area visibility, and
// enforces the cross-tenant existence guard when a specific target user is
// in scope. On failure it writes the problem+json response itself and
// returns ok=false; callers must return immediately without writing again.
func (h *MembershipHandler) resolveMembershipsListScope(w http.ResponseWriter, r *http.Request, tenantID, userID string) (resolvedUserID string, tenantWide, hasManagedAreas bool, actor string, ok bool) {
	actor, hasActor := authn.UserIDFromContext(r.Context())
	actor = strings.TrimSpace(actor)
	if !hasActor || actor == "" {
		problem.Respond(w, r, problem.New(http.StatusForbidden, problem.CodePermissionDenied, "Insufficient permissions"))
		return "", false, false, actor, false
	}
	tenantWide, hasManagedAreas, err := h.svc.DirectoryScope(r.Context(), tenantID, actor, string(iamdomain.CapMembershipManage))
	if err != nil {
		slog.Error("iam memberships: resolve directory scope failed", "err", err)
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "Failed to list memberships"))
		return "", false, false, actor, false
	}

	if !tenantWide && !hasManagedAreas {
		// Self-only: ignore any userId filter that isn't the actor (no probing).
		if userID != "" && !strings.EqualFold(userID, actor) {
			problem.Respond(w, r, problem.New(http.StatusForbidden, problem.CodePermissionDenied, "Insufficient permissions"))
			return "", false, false, actor, false
		}
		userID = actor
	}

	// When a specific target user is in scope, preserve the cross-tenant 404
	// guard (an attacker must not distinguish "exists elsewhere" from "absent").
	if userID != "" {
		if !h.guardMembershipUserInTenant(w, r, tenantID, userID) {
			return "", false, false, actor, false
		}
	}

	return userID, tenantWide, hasManagedAreas, actor, true
}

// grantMembership — operationId grantAreaMembership.
func (h *MembershipHandler) grantMembership(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		problem.Respond(w, r, problem.New(http.StatusNotImplemented, problem.CodeInternalUnknown, "Membership service is not configured"))
		return
	}

	var req grantMembershipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, "Invalid JSON payload"))
		return
	}
	userID := strings.TrimSpace(req.UserID)
	areaCode := strings.TrimSpace(req.AreaCode)
	role := strings.ToLower(strings.TrimSpace(req.Role))
	if userID == "" || areaCode == "" || role == "" {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, "userId, areaCode and role are required"))
		return
	}
	// ADR 0022 Phase 3: authorization is tier-1 (route cap) + tier-2 (area,
	// enforced in the repository's authz.Require with the target areaCode). The
	// former canManageMembershipTarget RoleSystemAdmin handler gate is removed —
	// area_admin is now authorized within their managed area, and tier-2 returns
	// ErrCapDenied (→ 403) when the actor lacks membership.manage in that area.
	//
	// Self-grant lockdown remains a business invariant (not an authz question): a
	// CapMembershipManage holder must not hand themselves additional area roles.
	// System admins bypass via the non-self path (target != actor); area admins
	// lose self-escalation here but retain cross-target grants in their area.
	if isSelf(r.Context(), userID) {
		problem.Respond(w, r, problem.New(http.StatusForbidden, problem.CodePermissionDenied, "Self-grant is not permitted"))
		return
	}
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error"))
		return
	}

	if !h.guardMembershipUserInTenant(w, r, tenantID, userID) {
		return
	}

	grantedBy, ok := authn.UserIDFromContext(r.Context())
	if !ok {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error"))
		return
	}
	err = h.svc.Grant(
		r.Context(),
		userID,
		tenantID,
		areaCode,
		iamdomain.Role(role),
		grantedBy,
	)
	if err != nil {
		h.writeMembershipError(w, r, err, "Failed to grant membership")
		return
	}

	// Audit is written in-tx by AreaMembershipService.Grant (AuditMembershipLogger);
	// no post-commit emission here (would be a duplicate row — H-3a).

	writeJSON(w, http.StatusCreated, iamapi.GrantAreaMembershipResponse{
		UserId:   userID,
		TenantId: tenantID,
		AreaCode: areaCode,
		Role:     iamapi.UserRole(role),
	})
}

// revokeMembership — operationId revokeAreaMembership.
func (h *MembershipHandler) revokeMembership(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		problem.Respond(w, r, problem.New(http.StatusNotImplemented, problem.CodeInternalUnknown, "Membership service is not configured"))
		return
	}

	userID := strings.TrimSpace(r.PathValue("user_id"))
	areaCode := strings.TrimSpace(r.PathValue("area_code"))
	if userID == "" || areaCode == "" {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, "user_id and area_code are required"))
		return
	}
	// ADR 0022 Phase 3: tier-2 area enforcement (repository authz.Require with the
	// target areaCode) replaces the removed RoleSystemAdmin handler gate; an actor
	// without membership.manage in this area gets 403 via ErrCapDenied.

	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error"))
		return
	}

	if !h.guardMembershipUserInTenant(w, r, tenantID, userID) {
		return
	}

	revokedBy, ok := authn.UserIDFromContext(r.Context())
	if !ok {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error"))
		return
	}

	err = h.svc.Revoke(r.Context(), userID, tenantID, areaCode, revokedBy)
	if err != nil {
		h.writeMembershipError(w, r, err, "Failed to revoke membership")
		return
	}

	// Audit is written in-tx by AreaMembershipService.Revoke (AuditMembershipLogger);
	// no post-commit emission here (would be a duplicate row — H-3a).

	w.WriteHeader(http.StatusNoContent)
}

// ─── helpers ───────────────────────────────────────────────────────────────

func (h *MembershipHandler) writeMembershipError(w http.ResponseWriter, r *http.Request, err error, defaultMessage string) {
	var capDenied authz.ErrCapDenied
	switch {
	case errors.As(err, &capDenied):
		// ADR 0022 Phase 3: tier-2 area denial (area_admin acting outside a
		// managed area, or actor lacking membership.manage there) → 403, not 500.
		problem.Respond(w, r, problem.New(http.StatusForbidden, problem.CodePermissionDenied, "Insufficient permissions"))
	case errors.Is(err, iamapp.ErrMembershipExists):
		problem.Respond(w, r, problem.New(http.StatusConflict, problem.CodeConflictMembershipExists, "Membership already exists for this user and area with the same role"))
	case errors.Is(err, iamapp.ErrMembershipNotFound):
		problem.Respond(w, r, problem.New(http.StatusNotFound, problem.CodeNotFoundMembership, "Membership not found"))
	case errors.Is(err, iamapp.ErrUnknownRole):
		problem.Respond(w, r, problem.NewFor(problem.CodeValidationRoleUnknown, "Unknown role"))
	default:
		slog.Error("iam memberships: handler error", "err", err)
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, defaultMessage))
	}
}

// guardMembershipUserInTenant mirrors PeopleHandler.guardUserInTenant — cross-
// tenant probes get 404 (not 403) so an attacker cannot infer "exists in
// another tenant". Returns true on match. When the verifier is not wired
// (e.g. SQLDB-less boot path), the guard fails closed with NOT_IMPLEMENTED.
func (h *MembershipHandler) guardMembershipUserInTenant(w http.ResponseWriter, r *http.Request, tenantID, userID string) bool {
	if h.verifier == nil {
		problem.Respond(w, r, problem.New(http.StatusNotImplemented, problem.CodeInternalUnknown, "Membership tenant verifier is not configured"))
		return false
	}
	if err := h.verifier.VerifyUserInTenant(r.Context(), tenantID, userID); err != nil {
		if errors.Is(err, iamapp.ErrUserNotInTenant) {
			problem.Respond(w, r, problem.New(http.StatusNotFound, problem.CodeNotFoundResource, "User not found"))
			return false
		}
		slog.Error("iam memberships: verify user-in-tenant failed", "err", err)
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, problem.CodeInternalUnknown, "Failed to verify user"))
		return false
	}
	return true
}

func tenantIDFromRequest(r *http.Request) (string, error) {
	return tenant.FromContext(r.Context())
}

// isSelf reports whether the authenticated actor IS the target user. Used by
// grantMembership to block self-escalation: a CapMembershipManage holder
// (typically area_admin) must not be able to hand themselves additional roles.
func isSelf(ctx context.Context, targetUserID string) bool {
	// A3.3 class B (identity OPTIONAL by explicit policy): this is a PREDICATE,
	// not a gate — "is the actor the same person as the target?". With no actor
	// the honest answer is false (nobody is the target), and answering false
	// does NOT open the route: the actor-required decision on this path is the
	// `grantedBy, ok := authn.UserIDFromContext(...)` check in grantMembership
	// / revokeMembership, which 401s before isSelf can matter. The presence
	// result is therefore read explicitly and consumed here rather than
	// discarded into a "" that would compare equal to a blank target.
	actor, ok := authn.UserIDFromContext(ctx)
	if !ok {
		return false
	}
	return strings.EqualFold(actor, strings.TrimSpace(targetUserID))
}
