package httpdelivery

import (
	"context"
	"net/http"

	iamapi "metaldocs/internal/modules/iam/api"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

// RoleCapabilitiesReader is the read-only repository surface needed by the
// "Roles & Capabilities" Admin Center tab. The handler depends on this
// interface so tests can stub it without an import cycle to postgres.
type RoleCapabilitiesReader interface {
	ListRoleCapabilities(ctx context.Context) ([]iamdomain.RoleCapabilityLink, error)
}

// RolesCapsHandler serves the read-only matrix endpoints
//
//	GET /api/v1/iam/roles
//	GET /api/v1/iam/capabilities
//	GET /api/v1/iam/role-capabilities
//
// All three are gated on CapMembershipView at the permissions table. They
// require authenticated tenant context (middleware-enforced) even though
// the catalogues themselves are global.
type RolesCapsHandler struct {
	roleCaps RoleCapabilitiesReader

	// Catalogues cached at construction time — the role + capability lists
	// only change via migration (and a process restart).
	cachedRoles []iamapi.RoleDescriptor
	cachedCaps  []iamapi.CapabilityDescriptor
}

func NewRolesCapsHandler(roleCaps RoleCapabilitiesReader) *RolesCapsHandler {
	return &RolesCapsHandler{
		roleCaps:    roleCaps,
		cachedRoles: buildRoleCatalog(),
		cachedCaps:  buildCapabilityCatalog(),
	}
}

func (h *RolesCapsHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/iam/roles", h.listRoles)
	mux.HandleFunc("GET /api/v1/iam/capabilities", h.listCapabilities)
	mux.HandleFunc("GET /api/v1/iam/role-capabilities", h.listRoleCapabilities)
}

func (h *RolesCapsHandler) listRoles(w http.ResponseWriter, r *http.Request) {
	if _, err := tenant.FromContext(r.Context()); err != nil {
		_ = problem.Write(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error"))
		return
	}
	writeJSON(w, http.StatusOK, iamapi.ListRolesResponse{Items: h.cachedRoles})
}

func (h *RolesCapsHandler) listCapabilities(w http.ResponseWriter, r *http.Request) {
	if _, err := tenant.FromContext(r.Context()); err != nil {
		_ = problem.Write(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error"))
		return
	}
	writeJSON(w, http.StatusOK, iamapi.ListCapabilitiesResponse{Items: h.cachedCaps})
}

func (h *RolesCapsHandler) listRoleCapabilities(w http.ResponseWriter, r *http.Request) {
	if _, err := tenant.FromContext(r.Context()); err != nil {
		_ = problem.Write(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error"))
		return
	}
	if h.roleCaps == nil {
		_ = problem.Write(w, problem.New(http.StatusNotImplemented, "INTERNAL_ERROR", "role_capabilities repository is not configured"))
		return
	}
	links, err := h.roleCaps.ListRoleCapabilities(r.Context())
	if err != nil {
		_ = problem.Write(w, problem.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list role capabilities"))
		return
	}
	items := make([]iamapi.RoleCapabilityLink, 0, len(links))
	for _, link := range links {
		items = append(items, iamapi.RoleCapabilityLink{
			Role:       iamapi.UserRole(link.Role),
			Capability: string(link.Capability),
		})
	}
	writeJSON(w, http.StatusOK, iamapi.ListRoleCapabilitiesResponse{Items: items})
}

func buildRoleCatalog() []iamapi.RoleDescriptor {
	roles := iamdomain.CanonicalRoles()
	out := make([]iamapi.RoleDescriptor, 0, len(roles))
	for _, role := range roles {
		out = append(out, iamapi.RoleDescriptor{
			Code:        iamapi.UserRole(role.Code),
			Label:       role.Label,
			Description: role.Description,
			Category:    iamapi.RoleDescriptorCategory(role.Category),
		})
	}
	return out
}

func buildCapabilityCatalog() []iamapi.CapabilityDescriptor {
	caps := iamdomain.CapabilityCatalog()
	out := make([]iamapi.CapabilityDescriptor, 0, len(caps))
	for _, cap := range caps {
		out = append(out, iamapi.CapabilityDescriptor{
			Code:        string(cap.Code),
			Description: cap.Description,
			Category:    cap.Category,
		})
	}
	return out
}
