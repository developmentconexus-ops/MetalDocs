package application

import (
	"context"
	"strings"

	"metaldocs/internal/modules/iam/domain"
)

type RoleCacheInvalidator interface {
	InvalidateUser(userID string)
	InvalidateUserTenant(userID, tenantID string)
}

type AdminService struct {
	repo        domain.RoleAdminRepository
	invalidator RoleCacheInvalidator
}

func NewAdminService(repo domain.RoleAdminRepository, invalidator RoleCacheInvalidator) *AdminService {
	return &AdminService{repo: repo, invalidator: invalidator}
}

func (s *AdminService) UpsertUserAndAssignRole(ctx context.Context, userID, displayName, tenantID string, role domain.Role, assignedBy string) error {
	userID = strings.TrimSpace(userID)
	displayName = strings.TrimSpace(displayName)
	tenantID = strings.TrimSpace(tenantID)
	assignedBy = strings.TrimSpace(assignedBy)

	if userID == "" {
		return domain.ErrUserNotFound
	}
	if tenantID == "" {
		return domain.ErrUserNotFound
	}
	if !domain.IsValidRole(role) {
		return domain.ErrInvalidRole
	}
	if displayName == "" {
		displayName = userID
	}
	if assignedBy == "" {
		assignedBy = "system"
	}

	if err := s.repo.UpsertUserAndAssignRole(ctx, userID, displayName, tenantID, role, assignedBy); err != nil {
		return err
	}
	if s.invalidator != nil {
		s.invalidator.InvalidateUserTenant(userID, tenantID)
	}
	return nil
}

func (s *AdminService) ReplaceUserRoles(ctx context.Context, userID, displayName, tenantID string, role domain.Role, assignedBy string) error {
	userID = strings.TrimSpace(userID)
	displayName = strings.TrimSpace(displayName)
	tenantID = strings.TrimSpace(tenantID)
	assignedBy = strings.TrimSpace(assignedBy)

	if userID == "" {
		return domain.ErrUserNotFound
	}
	if tenantID == "" {
		return domain.ErrUserNotFound
	}
	if displayName == "" {
		displayName = userID
	}
	if assignedBy == "" {
		assignedBy = "system"
	}
	if !domain.IsValidRole(role) {
		return domain.ErrInvalidRole
	}

	if err := s.repo.ReplaceUserRoles(ctx, userID, displayName, tenantID, role, assignedBy); err != nil {
		return err
	}
	if s.invalidator != nil {
		s.invalidator.InvalidateUserTenant(userID, tenantID)
	}
	return nil
}
