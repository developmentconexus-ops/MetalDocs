package application

import (
	"context"
	"log"
	"strings"

	"metaldocs/internal/modules/iam/domain"
)

// DevRoleProvider is a deterministic in-memory provider used for local memory mode.
type DevRoleProvider struct {
	rolesByUser      map[string][]domain.Role
	allowedTenantID string
}

func NewDevRoleProvider(rolesByUser map[string][]domain.Role, allowedTenantID string) *DevRoleProvider {
	if rolesByUser == nil {
		rolesByUser = map[string][]domain.Role{}
	}
	allowedTenantID = strings.TrimSpace(allowedTenantID)
	log.Printf("iam: using dev role provider; restrict access to tenant %q", allowedTenantID)
	return &DevRoleProvider{rolesByUser: rolesByUser, allowedTenantID: allowedTenantID}
}

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
