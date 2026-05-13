package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
)

type RoleAdminRepository struct {
	db *sql.DB
}

func NewRoleAdminRepository(db *sql.DB) *RoleAdminRepository {
	return &RoleAdminRepository{db: db}
}

func (r *RoleAdminRepository) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, opts)
}

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

func (r *RoleAdminRepository) UpsertUserAndAssignRole(ctx context.Context, userID, displayName, tenantID string, role iamdomain.Role, assignedBy string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin iam tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := authz.Require(ctx, tx, string(iamdomain.CapUserManage), "tenant"); err != nil {
		return fmt.Errorf("require iam user.manage authorization: %w", err)
	}

	const upsertUser = `
INSERT INTO metaldocs.iam_users (user_id, display_name, is_active, updated_at)
VALUES ($1, $2, TRUE, NOW())
ON CONFLICT (user_id)
DO UPDATE SET display_name = EXCLUDED.display_name, is_active = TRUE, updated_at = NOW()
`
	if _, err := tx.ExecContext(ctx, upsertUser, userID, displayName); err != nil {
		return fmt.Errorf("upsert iam user: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
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

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit iam tx: %w", err)
	}
	return nil
}

// ReplaceUserRoles writes the user+role assignment. The schema constraint
// UNIQUE(tenant_id, user_id) means at most ONE role row per user per tenant.
// If the input slice has multiple roles, only the last one is written.
func (r *RoleAdminRepository) ReplaceUserRoles(ctx context.Context, userID, displayName, tenantID string, roles []iamdomain.Role, assignedBy string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin iam replace tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := r.ReplaceUserRolesTx(ctx, tx, userID, displayName, tenantID, roles, assignedBy); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *RoleAdminRepository) ReplaceUserRolesTx(ctx context.Context, tx *sql.Tx, userID, displayName, tenantID string, roles []iamdomain.Role, assignedBy string) error {
	if err := authz.Require(ctx, tx, string(iamdomain.CapUserManage), "tenant"); err != nil {
		return fmt.Errorf("require iam user.manage authorization: %w", err)
	}

	const upsertUser = `
INSERT INTO metaldocs.iam_users (user_id, display_name, is_active, updated_at)
VALUES ($1, $2, TRUE, NOW())
ON CONFLICT (user_id)
DO UPDATE SET display_name = EXCLUDED.display_name, updated_at = NOW()
`
	if _, err := tx.ExecContext(ctx, upsertUser, userID, displayName); err != nil {
		return fmt.Errorf("upsert iam user: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM metaldocs.iam_user_roles WHERE tenant_id = $1::uuid AND user_id = $2`,
		tenantID, userID); err != nil {
		return fmt.Errorf("delete prior iam roles: %w", err)
	}

	var lastRole string
	for _, role := range roles {
		if code := strings.TrimSpace(string(role)); code != "" {
			lastRole = code
		}
	}
	if lastRole == "" {
		return nil
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO metaldocs.iam_user_roles (user_id, tenant_id, role_code, assigned_at, assigned_by)
VALUES ($1, $2::uuid, $3, NOW(), $4)
`, userID, tenantID, lastRole, assignedBy); err != nil {
		return fmt.Errorf("insert iam role: %w", err)
	}
	return nil
}
