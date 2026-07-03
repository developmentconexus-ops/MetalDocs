package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

// ErrCapabilityDenied is returned by CanDo when the user does not hold the
// requested capability in the tenant (and is not system_admin).
var ErrCapabilityDenied = errors.New("capability denied")

// queryExecutor is the narrow read surface CapabilityService needs. *sql.DB
// satisfies it; tests can substitute a fake. (Accept interfaces, return structs.)
type queryExecutor interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// CapabilityService is the tier-1 HTTP-edge capability check (CanDo), backed
// by a DB-level EXISTS query over direct role grants and group-derived
// grants. Distinct from tier-2 authz.Require, which runs in-tx with
// area-scoping. Safe for concurrent use; holds no mutable state beyond db.
type CapabilityService struct {
	db queryExecutor
}

// NewCapabilityService constructs the service over the given query executor
// (typically *sql.DB; tests may substitute a fake satisfying queryExecutor).
func NewCapabilityService(db queryExecutor) *CapabilityService {
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

// IsSystemAdmin reports whether the user holds the system_admin role in the
// tenant via a direct role assignment or group membership. system_admin
// bypasses ownership and capability checks (mirrors CanDo / authz.Require).
func (s *CapabilityService) IsSystemAdmin(ctx context.Context, userID, tenantID string) (bool, error) {
	const adminQuery = `
SELECT EXISTS (
  SELECT 1
    FROM metaldocs.iam_user_roles ur
   WHERE ur.user_id   = $1
     AND ur.tenant_id = $2::uuid
     AND ur.role_code = 'system_admin'
  UNION ALL
  SELECT 1
    FROM metaldocs.iam_group_members gm
    JOIN metaldocs.iam_groups g ON g.id = gm.group_id
    JOIN metaldocs.iam_group_roles gr ON gr.group_id = gm.group_id
   WHERE gm.user_id   = $1
     AND gm.tenant_id = $2::uuid
     AND g.tenant_id  = $2::uuid
     AND gr.role      = 'system_admin'
)`
	var isAdmin bool
	if err := s.db.QueryRowContext(ctx, adminQuery, userID, tenantID).Scan(&isAdmin); err != nil {
		return false, fmt.Errorf("is system admin: %w", err)
	}
	return isAdmin, nil
}

// CapsByUserID returns the distinct capability codes the user holds in the
// tenant via direct roles or group roles. system_admin returns the full known
// capability set (mirrors the bypass semantics of CanDo / authz.Require).
//
// Intended for tier-1 UX hints (e.g. /auth/me); backend mutation paths must
// still call CanDo or authz.Require — never trust this result as enforcement.
func (s *CapabilityService) CapsByUserID(ctx context.Context, userID, tenantID string) ([]iamdomain.Capability, error) {
	isAdmin, err := s.IsSystemAdmin(ctx, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("caps for user: admin check: %w", err)
	}
	if isAdmin {
		return iamdomain.AllCapabilities(), nil
	}

	const query = `
SELECT DISTINCT rc.capability
  FROM metaldocs.iam_user_roles ur
  JOIN metaldocs.role_capabilities rc ON rc.role = ur.role_code
 WHERE ur.user_id = $1
   AND ur.tenant_id = $2::uuid
UNION
SELECT DISTINCT rc.capability
  FROM metaldocs.iam_group_members gm
  JOIN metaldocs.iam_groups g ON g.id = gm.group_id
  JOIN metaldocs.iam_group_roles gr ON gr.group_id = gm.group_id
  JOIN metaldocs.role_capabilities rc ON rc.role = gr.role
 WHERE gm.user_id   = $1
   AND gm.tenant_id = $2::uuid
   AND g.tenant_id  = $2::uuid
`
	rows, err := s.db.QueryContext(ctx, query, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("caps for user: query: %w", err)
	}
	defer rows.Close()

	caps := make([]iamdomain.Capability, 0, 8)
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, fmt.Errorf("caps for user: scan: %w", err)
		}
		c := iamdomain.Capability(raw)
		if iamdomain.IsValidCapability(c) {
			caps = append(caps, c)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("caps for user: rows: %w", err)
	}
	return caps, nil
}
