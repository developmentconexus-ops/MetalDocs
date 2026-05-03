package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var ErrCapabilityDenied = errors.New("capability denied")

type CapabilityService struct {
	db *sql.DB
}

func NewCapabilityService(db *sql.DB) *CapabilityService {
	return &CapabilityService{db: db}
}

func (s *CapabilityService) CanDo(ctx context.Context, userID, tenantID, capability string) error {
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
    JOIN metaldocs.iam_group_roles gr ON gr.group_id = gm.group_id
   WHERE gm.user_id = $1
     AND gm.tenant_id = $2::uuid
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
    JOIN metaldocs.iam_group_roles gr ON gr.group_id = gm.group_id
    JOIN metaldocs.role_capabilities rc ON rc.role = gr.role
   WHERE gm.user_id = $1
     AND gm.tenant_id = $2::uuid
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
