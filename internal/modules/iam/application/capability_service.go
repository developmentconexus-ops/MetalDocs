package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

var ErrCapabilityDenied = errors.New("capability denied")

type CapabilityService struct {
	// TODO: replace *sql.DB with a query executor interface for testability.
	db *sql.DB
}

func NewCapabilityService(db *sql.DB) *CapabilityService {
	return &CapabilityService{db: db}
}

// CanDo enforces a tier-1 tenant-level capability check.
// See wiki/decisions/0007-two-tier-authz.md for the boundary between tier-1
// (this function, HTTP middleware) and tier-2 (authz.Require, in-transaction).
//
// Returns nil if the user holds the capability in the given tenant via either:
//   - direct role in iam_user_roles
//   - membership in iam_groups + iam_group_roles
//
// system_admin role bypasses all capability checks.
//
// Returns ErrCapabilityDenied otherwise.
func (s *CapabilityService) CanDo(ctx context.Context, userID, tenantID, capability string) error {
	if !iamdomain.IsValidCapability(iamdomain.Capability(capability)) {
		return fmt.Errorf("check capability: %w", ErrCapabilityDenied)
	}

	const query = `
SELECT EXISTS (
  SELECT 1
    FROM metaldocs.iam_user_roles ur
   WHERE ur.user_id = $1
     AND ur.tenant_id = $2::uuid
     AND ur.role_code = 'system_admin'
  UNION ALL
  SELECT 1
    FROM metaldocs.iam_group_members gm
    JOIN metaldocs.iam_groups g ON g.id = gm.group_id
    JOIN metaldocs.iam_group_roles gr ON gr.group_id = gm.group_id
   WHERE gm.user_id = $1
     AND gm.tenant_id = $2::uuid
     AND g.tenant_id = $2::uuid
     AND gr.role = 'system_admin'
  UNION ALL
  SELECT 1
    FROM metaldocs.iam_user_roles ur
    JOIN metaldocs.role_capabilities rc ON rc.role = ur.role_code
   WHERE ur.user_id = $1
     AND ur.tenant_id = $2::uuid
     AND rc.capability = $3
  UNION ALL
  SELECT 1
    FROM metaldocs.iam_group_members gm
    JOIN metaldocs.iam_groups g ON g.id = gm.group_id
    JOIN metaldocs.iam_group_roles gr ON gr.group_id = gm.group_id
    JOIN metaldocs.role_capabilities rc ON rc.role = gr.role
   WHERE gm.user_id = $1
     AND gm.tenant_id = $2::uuid
     AND g.tenant_id = $2::uuid
     AND rc.capability = $3
)`

	var allowed bool
	if err := s.db.QueryRowContext(ctx, query, userID, tenantID, capability).Scan(&allowed); err != nil {
		return fmt.Errorf("check capability: %w", err)
	}
	if !allowed {
		return ErrCapabilityDenied
	}
	return nil
}
