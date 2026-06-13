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
//   - audit emission on every grant + revoke success path.
package httpdelivery

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	auditdomain "metaldocs/internal/modules/audit/domain"
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

type MembershipHandler struct {
	svc      *iamapp.AreaMembershipService
	verifier MembershipUserTenantVerifier
	audit    auditdomain.Writer
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

func NewMembershipHandler(svc *iamapp.AreaMembershipService, verifier MembershipUserTenantVerifier, audit auditdomain.Writer) *MembershipHandler {
	return &MembershipHandler{svc: svc, verifier: verifier, audit: audit}
}

func (h *MembershipHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/iam/area-memberships", h.listMemberships)
	mux.HandleFunc("POST /api/v1/iam/area-memberships", h.grantMembership)
	// F-DELETE-SHAPE (api-contract-hardening Phase E2): the revoked membership is
	// identified by path parameters (resource identity) rather than query params.
	mux.HandleFunc("DELETE /api/v1/iam/area-memberships/{user_id}/{area_code}", h.revokeMembership)
}

// listMemberships — operationId listAreaMemberships.
func (h *MembershipHandler) listMemberships(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		h.writeProblem(w, problem.New(http.StatusNotImplemented, "INTERNAL_ERROR", "Membership service is not configured"))
		return
	}

	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error"))
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
	actor, ok := authn.UserIDFromContext(r.Context())
	actor = strings.TrimSpace(actor)
	if !ok || actor == "" {
		h.writeProblem(w, problem.New(http.StatusForbidden, "AUTH_FORBIDDEN", "Insufficient permissions"))
		return
	}
	tenantWide, hasManagedAreas, err := h.svc.DirectoryScope(r.Context(), tenantID, actor, string(iamdomain.CapMembershipManage))
	if err != nil {
		slog.Error("iam memberships: resolve directory scope failed", "err", err)
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list memberships"))
		return
	}

	if !tenantWide && !hasManagedAreas {
		// Self-only: ignore any userId filter that isn't the actor (no probing).
		if userID != "" && !strings.EqualFold(userID, actor) {
			h.writeProblem(w, problem.New(http.StatusForbidden, "AUTH_FORBIDDEN", "Insufficient permissions"))
			return
		}
		userID = actor
	}

	// When a specific target user is in scope, preserve the cross-tenant 404
	// guard (an attacker must not distinguish "exists elsewhere" from "absent").
	if userID != "" {
		if !h.guardMembershipUserInTenant(w, r, tenantID, userID) {
			return
		}
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
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list memberships"))
		return
	}
	dtos := make([]membershipDTO, 0, len(items))
	for _, m := range items {
		dtos = append(dtos, toMembershipDTO(m))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": dtos})
}

// grantMembership — operationId grantAreaMembership.
func (h *MembershipHandler) grantMembership(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		h.writeProblem(w, problem.New(http.StatusNotImplemented, "INTERNAL_ERROR", "Membership service is not configured"))
		return
	}

	var req grantMembershipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON payload"))
		return
	}
	userID := strings.TrimSpace(req.UserID)
	areaCode := strings.TrimSpace(req.AreaCode)
	role := strings.ToLower(strings.TrimSpace(req.Role))
	if userID == "" || areaCode == "" || role == "" {
		h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "userId, areaCode and role are required"))
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
		h.writeProblem(w, problem.New(http.StatusForbidden, "AUTH_FORBIDDEN", "Self-grant is not permitted"))
		return
	}
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error"))
		return
	}

	if !h.guardMembershipUserInTenant(w, r, tenantID, userID) {
		return
	}

	grantedBy, ok := authn.UserIDFromContext(r.Context())
	if !ok {
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error"))
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
		h.writeMembershipError(w, err, "Failed to grant membership")
		return
	}

	h.recordMembershipAudit(r, userID, "iam.area_membership.granted", map[string]any{
		"user_id":   userID,
		"tenant_id": tenantID,
		"area_code": areaCode,
		"role":     role,
	})

	writeJSON(w, http.StatusCreated, map[string]any{
		"user_id":   userID,
		"tenant_id": tenantID,
		"area_code": areaCode,
		"role":     role,
	})
}

// revokeMembership — operationId revokeAreaMembership.
func (h *MembershipHandler) revokeMembership(w http.ResponseWriter, r *http.Request) {
	if h.svc == nil {
		h.writeProblem(w, problem.New(http.StatusNotImplemented, "INTERNAL_ERROR", "Membership service is not configured"))
		return
	}

	userID := strings.TrimSpace(r.PathValue("user_id"))
	areaCode := strings.TrimSpace(r.PathValue("area_code"))
	if userID == "" || areaCode == "" {
		h.writeProblem(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", "user_id and area_code are required"))
		return
	}
	// ADR 0022 Phase 3: tier-2 area enforcement (repository authz.Require with the
	// target areaCode) replaces the removed RoleSystemAdmin handler gate; an actor
	// without membership.manage in this area gets 403 via ErrCapDenied.

	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error"))
		return
	}

	if !h.guardMembershipUserInTenant(w, r, tenantID, userID) {
		return
	}

	revokedBy, ok := authn.UserIDFromContext(r.Context())
	if !ok {
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error"))
		return
	}

	err = h.svc.Revoke(r.Context(), userID, tenantID, areaCode, revokedBy)
	if err != nil {
		h.writeMembershipError(w, err, "Failed to revoke membership")
		return
	}

	h.recordMembershipAudit(r, userID, "iam.area_membership.revoked", map[string]any{
		"user_id":   userID,
		"tenant_id": tenantID,
		"area_code": areaCode,
	})

	w.WriteHeader(http.StatusNoContent)
}

// ─── helpers ───────────────────────────────────────────────────────────────

func (h *MembershipHandler) writeProblem(w http.ResponseWriter, p *problem.Problem) {
	if werr := problem.Write(w, p); werr != nil {
		slog.Warn("iam memberships: write response failed", "err", werr)
	}
}

func (h *MembershipHandler) writeMembershipError(w http.ResponseWriter, err error, defaultMessage string) {
	var capDenied authz.ErrCapDenied
	switch {
	case errors.As(err, &capDenied):
		// ADR 0022 Phase 3: tier-2 area denial (area_admin acting outside a
		// managed area, or actor lacking membership.manage there) → 403, not 500.
		h.writeProblem(w, problem.New(http.StatusForbidden, "AUTH_FORBIDDEN", "Insufficient permissions"))
	case errors.Is(err, iamapp.ErrMembershipExists):
		h.writeProblem(w, problem.New(http.StatusConflict, "MEMBERSHIP_EXISTS", "Membership already exists for this user and area with the same role"))
	case errors.Is(err, iamapp.ErrMembershipNotFound):
		h.writeProblem(w, problem.New(http.StatusNotFound, "MEMBERSHIP_NOT_FOUND", "Membership not found"))
	case errors.Is(err, iamapp.ErrUnknownRole):
		h.writeProblem(w, problem.New(http.StatusBadRequest, "UNKNOWN_ROLE", "Unknown role"))
	default:
		slog.Error("iam memberships: handler error", "err", err)
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", defaultMessage))
	}
}

// guardMembershipUserInTenant mirrors PeopleHandler.guardUserInTenant — cross-
// tenant probes get 404 (not 403) so an attacker cannot infer "exists in
// another tenant". Returns true on match. When the verifier is not wired
// (e.g. SQLDB-less boot path), the guard fails closed with NOT_IMPLEMENTED.
func (h *MembershipHandler) guardMembershipUserInTenant(w http.ResponseWriter, r *http.Request, tenantID, userID string) bool {
	if h.verifier == nil {
		h.writeProblem(w, problem.New(http.StatusNotImplemented, "INTERNAL_ERROR", "Membership tenant verifier is not configured"))
		return false
	}
	if err := h.verifier.VerifyUserInTenant(r.Context(), tenantID, userID); err != nil {
		if errors.Is(err, iamapp.ErrUserNotInTenant) {
			h.writeProblem(w, problem.New(http.StatusNotFound, "NOT_FOUND", "User not found"))
			return false
		}
		slog.Error("iam memberships: verify user-in-tenant failed", "err", err)
		h.writeProblem(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to verify user"))
		return false
	}
	return true
}

// recordMembershipAudit is log-and-continue — audit failures must never fail
// the mutating request once the write has been committed.
func (h *MembershipHandler) recordMembershipAudit(r *http.Request, userID, action string, payload map[string]any) {
	if h.audit == nil {
		return
	}
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		slog.Warn("audit: skip — tenant missing from context", "action", action, "user_id", userID, "err", err)
		return
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		slog.Warn("audit: skip — payload marshal failed", "action", action, "user_id", userID, "err", err)
		return
	}
	now := time.Now().UTC()
	if err := h.audit.Record(r.Context(), auditdomain.Event{
		ID:           "evt_" + uuid.NewString(),
		OccurredAt:   now,
		ActorID:      authenticatedActor(r),
		Action:       action,
		ResourceType: "area_membership",
		ResourceID:   userID,
		PayloadJSON:  string(payloadJSON),
		TraceID:      r.Header.Get("X-Trace-Id"),
		TenantID:     tenantID,
	}); err != nil {
		slog.Warn("audit: failed to record", "action", action, "user_id", userID, "err", err)
	}
}

func tenantIDFromRequest(r *http.Request) (string, error) {
	return tenant.FromContext(r.Context())
}

// isSelf reports whether the authenticated actor IS the target user. Used by
// grantMembership to block self-escalation: a CapMembershipManage holder
// (typically area_admin) must not be able to hand themselves additional roles.
func isSelf(ctx context.Context, targetUserID string) bool {
	actor := strings.TrimSpace(authenticatedActorFromContext(ctx))
	if actor == "" {
		return false
	}
	return strings.EqualFold(actor, strings.TrimSpace(targetUserID))
}

func authenticatedActorFromContext(ctx context.Context) string {
	userID, _ := authn.UserIDFromContext(ctx)
	return userID
}
