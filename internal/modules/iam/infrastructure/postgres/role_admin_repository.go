package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
)

// RoleAdminRepository writes user+role assignments against
// metaldocs.iam_users / iam_user_roles. Tx-owning callers (UpsertUserAndAssignRole,
// ReplaceUserRoles) open and commit their own transaction; tx-aware variants
// (UpsertUserAndAssignRoleTx, ReplaceUserRolesTx) run inside a caller-supplied
// *sql.Tx and require CapUserManage in-tx via authz.Require.
type RoleAdminRepository struct {
	db *sql.DB
}

// NewRoleAdminRepository constructs a pool-backed RoleAdminRepository.
func NewRoleAdminRepository(db *sql.DB) *RoleAdminRepository {
	return &RoleAdminRepository{db: db}
}

// BeginTx opens a database transaction for callers that need to compose
// RoleAdminRepository's tx-aware methods with other work. The caller owns
// Commit/Rollback.
func (r *RoleAdminRepository) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, opts)
}

// HasAnyRole reports whether at least one iam_user_roles row exists for role
// within tenantID.
func (r *RoleAdminRepository) HasAnyRole(ctx context.Context, role iamdomain.Role, tenantID string) (bool, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM metaldocs.iam_user_roles
WHERE role_code = $1
  AND tenant_id = $2::uuid
`, string(role), tenantID).Scan(&count); err != nil {
		return false, fmt.Errorf("count role assignments: %w", err)
	}
	return count > 0, nil
}

// UpsertUserAndAssignRole opens and commits its own transaction around
// UpsertUserAndAssignRoleTx, rolling back on any error.
func (r *RoleAdminRepository) UpsertUserAndAssignRole(ctx context.Context, userID, displayName, tenantID string, role iamdomain.Role, assignedBy string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin iam tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := r.UpsertUserAndAssignRoleTx(ctx, tx, userID, displayName, tenantID, role, assignedBy); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit iam tx: %w", err)
	}
	return nil
}

// UpsertUserAndAssignRoleTx requires CapUserManage in-tx (tenant-grade, area
// filter off), then upserts the iam_users row and replaces the user's role
// set with {role} via DELETE-then-INSERT under the (tenant_id, user_id)
// uniqueness on iam_user_roles — idempotent under retry. The caller owns
// Commit/Rollback on tx.
func (r *RoleAdminRepository) UpsertUserAndAssignRoleTx(ctx context.Context, tx *sql.Tx, userID, displayName, tenantID string, role iamdomain.Role, assignedBy string) error {
	ctx = authz.WithCapCache(ctx)
	if err := authz.SeedTxIdentity(ctx, tx, tenantID, assignedBy); err != nil {
		return fmt.Errorf("seed iam user.manage authorization: %w", err)
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapUserManage), "tenant"); err != nil {
		return fmt.Errorf("require iam user.manage authorization: %w", err)
	}

	// tenant_id is NOT NULL on iam_users with a sentinel default; it must be set
	// explicitly so new rows carry the caller's real tenant (re-audit Major #2,
	// M5/F5.7). EXCLUDED.tenant_id on conflict also repairs any pre-existing
	// sentinel row on the next upsert.
	const upsertUser = `
INSERT INTO metaldocs.iam_users (user_id, display_name, tenant_id, is_active, updated_at)
VALUES ($1, $2, $3::uuid, TRUE, NOW())
ON CONFLICT (user_id)
DO UPDATE SET display_name = EXCLUDED.display_name, tenant_id = EXCLUDED.tenant_id, is_active = TRUE, updated_at = NOW()
`
	if _, err := tx.ExecContext(ctx, upsertUser, userID, displayName, tenantID); err != nil {
		return fmt.Errorf("upsert iam user: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		// TODO: migration 0002 uses cascading FKs from iam_user_roles; keep role replacement explicit and review any future hard-delete path carefully.
		`DELETE FROM metaldocs.iam_user_roles WHERE tenant_id = $1::uuid AND user_id = $2`,
		tenantID, userID); err != nil {
		return fmt.Errorf("delete prior iam roles: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO metaldocs.iam_user_roles (user_id, tenant_id, role_code, assigned_at, assigned_by)
VALUES ($1, $2::uuid, $3, NOW(), $4)
`, userID, tenantID, string(role), assignedBy); err != nil {
		return fmt.Errorf("insert iam role: %w", err)
	}
	return nil
}

// ReplaceUserRoles writes the user+role assignments.
func (r *RoleAdminRepository) ReplaceUserRoles(ctx context.Context, userID, displayName, tenantID string, role iamdomain.Role, assignedBy string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin iam replace tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := r.ReplaceUserRolesTx(ctx, tx, userID, displayName, tenantID, role, assignedBy); err != nil {
		return err
	}
	return tx.Commit()
}

// ReplaceUserRolesTx requires CapUserManage in-tx (tenant-grade, area filter
// off), then upserts the iam_users row and replaces the user's role set with
// {role} via the same DELETE-then-INSERT pattern as UpsertUserAndAssignRoleTx.
// The caller owns Commit/Rollback on tx.
func (r *RoleAdminRepository) ReplaceUserRolesTx(ctx context.Context, tx *sql.Tx, userID, displayName, tenantID string, role iamdomain.Role, assignedBy string) error {
	ctx = authz.WithCapCache(ctx)
	if err := authz.SeedTxIdentity(ctx, tx, tenantID, assignedBy); err != nil {
		return fmt.Errorf("seed iam user.manage authorization: %w", err)
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapUserManage), "tenant"); err != nil {
		return fmt.Errorf("require iam user.manage authorization: %w", err)
	}

	// tenant_id is NOT NULL on iam_users with a sentinel default; set it
	// explicitly so new rows carry the caller's real tenant (re-audit Major #2,
	// M5/F5.7). EXCLUDED.tenant_id on conflict also repairs any pre-existing
	// sentinel row on the next upsert.
	const upsertUser = `
INSERT INTO metaldocs.iam_users (user_id, display_name, tenant_id, is_active, updated_at)
VALUES ($1, $2, $3::uuid, TRUE, NOW())
ON CONFLICT (user_id)
DO UPDATE SET display_name = EXCLUDED.display_name, tenant_id = EXCLUDED.tenant_id, updated_at = NOW()
`
	if _, err := tx.ExecContext(ctx, upsertUser, userID, displayName, tenantID); err != nil {
		return fmt.Errorf("upsert iam user: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		// TODO: migration 0002 uses cascading FKs from iam_user_roles; keep role replacement explicit and review any future hard-delete path carefully.
		`DELETE FROM metaldocs.iam_user_roles WHERE tenant_id = $1::uuid AND user_id = $2`,
		tenantID, userID); err != nil {
		return fmt.Errorf("delete prior iam roles: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO metaldocs.iam_user_roles (user_id, tenant_id, role_code, assigned_at, assigned_by)
VALUES ($1, $2::uuid, $3, NOW(), $4)
`, userID, tenantID, string(role), assignedBy); err != nil {
		return fmt.Errorf("insert iam role: %w", err)
	}
	return nil
}
