package application

import (
	"context"
	"log/slog"
	"strings"

	"metaldocs/internal/modules/iam/domain"
)

// DevRoleProvider is a deterministic in-memory provider used for local memory mode only.
type DevRoleProvider struct {
	rolesByUser     map[string][]domain.Role
	allowedTenantID string
}

// NewDevRoleProvider constructs the dev-mode provider. A nil rolesByUser map
// is replaced with an empty one. Only allowedTenantID resolves any user;
// every other tenant is treated as unknown.
func NewDevRoleProvider(rolesByUser map[string][]domain.Role, allowedTenantID string) *DevRoleProvider {
	if rolesByUser == nil {
		rolesByUser = map[string][]domain.Role{}
	}
	allowedTenantID = strings.TrimSpace(allowedTenantID)
	slog.Info("iam: using dev role provider", "allowed_tenant_id", allowedTenantID)
	return &DevRoleProvider{rolesByUser: rolesByUser, allowedTenantID: allowedTenantID}
}

// RolesByUserID returns a defensive copy of the user's configured roles.
// Returns domain.ErrUserNotFound when tenantID does not match the allowed
// tenant or userID is unknown, and domain.ErrNoRolesAssigned when the user
// is known but has zero roles configured.
func (p *DevRoleProvider) RolesByUserID(_ context.Context, userID, tenantID string) ([]domain.Role, error) {
	if strings.TrimSpace(tenantID) != p.allowedTenantID {
		return nil, domain.ErrUserNotFound
	}
	id := strings.TrimSpace(userID)
	if id == "" {
		return nil, domain.ErrUserNotFound
	}
	roles, ok := p.rolesByUser[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	if len(roles) == 0 {
		return nil, domain.ErrNoRolesAssigned
	}
	out := make([]domain.Role, len(roles))
	copy(out, roles)
	return out, nil
}

// UserActiveInTenant returns true iff tenantID matches the allowed tenant and
// userID is present in the known-user set. Dev mode has no deactivation state.
func (p *DevRoleProvider) UserActiveInTenant(_ context.Context, tenantID, userID string) (bool, error) {
	if strings.TrimSpace(tenantID) != p.allowedTenantID {
		return false, nil
	}
	id := strings.TrimSpace(userID)
	if id == "" {
		return false, nil
	}
	_, ok := p.rolesByUser[id]
	return ok, nil
}

// RolesByUserIDs resolves roles for multiple users in a single call (dev mode).
// Mirrors the batch semantics: absent/inactive users are omitted from the map;
// active users with no roles are present with an empty slice.
func (p *DevRoleProvider) RolesByUserIDs(_ context.Context, tenantID string, userIDs []string) (map[string][]domain.Role, error) {
	out := make(map[string][]domain.Role, len(userIDs))
	if strings.TrimSpace(tenantID) != p.allowedTenantID {
		return out, nil
	}
	for _, uid := range userIDs {
		id := strings.TrimSpace(uid)
		if id == "" {
			continue
		}
		roles, ok := p.rolesByUser[id]
		if !ok {
			continue
		}
		clone := make([]domain.Role, len(roles))
		copy(clone, roles)
		out[id] = clone
	}
	return out, nil
}
