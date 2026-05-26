package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
)

type UserAreaRepository struct {
	db *sql.DB
}

func NewUserAreaRepository(db *sql.DB) *UserAreaRepository {
	return &UserAreaRepository{db: db}
}

func (r *UserAreaRepository) ListActive(ctx context.Context, userID, tenantID string, now time.Time) ([]iamdomain.UserProcessArea, error) {
	const q = `
SELECT user_id, tenant_id::text, area_code, role, effective_from, effective_to, granted_by
FROM user_process_areas
WHERE user_id = $1
  AND tenant_id = $2::uuid
  AND effective_from <= $3
  AND (effective_to IS NULL OR effective_to > $3)
ORDER BY area_code ASC, effective_from DESC
`
	rows, err := r.db.QueryContext(ctx, q, userID, tenantID, now)
	if err != nil {
		return nil, fmt.Errorf("query active user process areas: %w", err)
	}
	defer rows.Close()

	result := make([]iamdomain.UserProcessArea, 0, 8)
	for rows.Next() {
		item, err := scanUserProcessArea(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active user process areas: %w", err)
	}
	return result, nil
}

func (r *UserAreaRepository) Insert(ctx context.Context, membership iamdomain.UserProcessArea) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin insert area tx: %w", err)
	}
	defer rollbackTx("insert_user_process_area", tx)

	if err := authz.Require(ctx, tx, string(iamdomain.CapMembershipManage), "tenant"); err != nil {
		return fmt.Errorf("iam: authz check Insert area: %w", err)
	}

	const q = `
INSERT INTO user_process_areas
  (user_id, tenant_id, area_code, role, effective_from, effective_to, granted_by)
VALUES
  ($1, $2::uuid, $3, $4, $5, $6, $7)
`
	// TODO: migration 0136 hangs these FKs off a non-PK unique key on iam_users; keep caller identity writes aligned with iam_users uniqueness.
	result, err := tx.ExecContext(
		ctx, q,
		membership.UserID,
		membership.TenantID,
		membership.AreaCode,
		string(membership.Role),
		membership.EffectiveFrom,
		membership.EffectiveTo,
		membership.GrantedBy,
	)
	if err != nil {
		return fmt.Errorf("insert user process area: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("insert user process area rows affected: %w", err)
	}
	if rowsAffected == 0 {
		slog.Warn("iam user area zero rows",
			"action", "insert",
			"user_id", membership.UserID,
			"tenant_id", membership.TenantID,
			"area_code", membership.AreaCode,
			"role", membership.Role,
		)
		return fmt.Errorf("insert user process area: no rows inserted")
	}
	return tx.Commit()
}

func (r *UserAreaRepository) CloseActive(ctx context.Context, userID, tenantID, areaCode string, effectiveTo time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin close area tx: %w", err)
	}
	defer rollbackTx("close_active_user_process_area", tx)

	if err := authz.Require(ctx, tx, string(iamdomain.CapMembershipManage), "tenant"); err != nil {
		return fmt.Errorf("iam: authz check CloseActive area: %w", err)
	}

	const q = `
UPDATE user_process_areas
SET effective_to = $4
WHERE user_id = $1
  AND tenant_id = $2::uuid
  AND area_code = $3
  AND effective_to IS NULL
`
	result, err := tx.ExecContext(ctx, q, userID, tenantID, areaCode, effectiveTo)
	if err != nil {
		return fmt.Errorf("close active user process area: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("close active user process area rows affected: %w", err)
	}
	if rowsAffected == 0 {
		slog.Warn("iam user area zero rows",
			"action", "close_active",
			"user_id", userID,
			"tenant_id", tenantID,
			"area_code", areaCode,
		)
		return fmt.Errorf("close active user process area: no rows updated")
	}
	return tx.Commit()
}

func (r *UserAreaRepository) GrantAtomic(ctx context.Context, oldMembership, newMembership iamdomain.UserProcessArea) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin grant transaction: %w", err)
	}
	defer rollbackTx("grant_atomic_user_process_area", tx)

	if err := authz.Require(ctx, tx, string(iamdomain.CapMembershipManage), "tenant"); err != nil {
		return fmt.Errorf("iam: authz check GrantAtomic: %w", err)
	}

	const closeQ = `
UPDATE user_process_areas
SET effective_to = $5
WHERE user_id = $1
  AND tenant_id = $2::uuid
  AND area_code = $3
  AND effective_from = $4
  AND effective_to IS NULL
`
	closeResult, err := tx.ExecContext(
		ctx,
		closeQ,
		oldMembership.UserID,
		oldMembership.TenantID,
		oldMembership.AreaCode,
		oldMembership.EffectiveFrom,
		newMembership.EffectiveFrom,
	)
	if err != nil {
		return fmt.Errorf("close active membership in grant transaction: %w", err)
	}
	rowsAffected, err := closeResult.RowsAffected()
	if err != nil {
		return fmt.Errorf("read affected rows for close active membership: %w", err)
	}
	if rowsAffected == 0 {
		slog.Warn("iam user area zero rows",
			"action", "grant_atomic_close",
			"user_id", oldMembership.UserID,
			"tenant_id", oldMembership.TenantID,
			"area_code", oldMembership.AreaCode,
		)
		return fmt.Errorf("close active membership in grant transaction: no rows updated")
	}

	const insertQ = `
INSERT INTO user_process_areas
  (user_id, tenant_id, area_code, role, effective_from, effective_to, granted_by)
VALUES
  ($1, $2::uuid, $3, $4, $5, $6, $7)
`
	// TODO: migration 0136 hangs these FKs off a non-PK unique key on iam_users; keep caller identity writes aligned with iam_users uniqueness.
	if _, err := tx.ExecContext(
		ctx,
		insertQ,
		newMembership.UserID,
		newMembership.TenantID,
		newMembership.AreaCode,
		string(newMembership.Role),
		newMembership.EffectiveFrom,
		newMembership.EffectiveTo,
		newMembership.GrantedBy,
	); err != nil {
		return fmt.Errorf("insert membership in grant transaction: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit grant transaction: %w", err)
	}
	return nil
}

func (r *UserAreaRepository) GetActiveByUserAndArea(ctx context.Context, userID, tenantID, areaCode string, now time.Time) (*iamdomain.UserProcessArea, error) {
	const q = `
SELECT user_id, tenant_id::text, area_code, role, effective_from, effective_to, granted_by
FROM user_process_areas
WHERE user_id = $1
  AND tenant_id = $2::uuid
  AND area_code = $3
  AND effective_from <= $4
  AND (effective_to IS NULL OR effective_to > $4)
ORDER BY effective_from DESC
LIMIT 1
`
	row := r.db.QueryRowContext(ctx, q, userID, tenantID, areaCode, now)
	item, err := scanUserProcessArea(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanUserProcessArea(s scanner) (iamdomain.UserProcessArea, error) {
	var item iamdomain.UserProcessArea
	var effectiveTo sql.NullTime
	var grantedBy sql.NullString
	if err := s.Scan(
		&item.UserID,
		&item.TenantID,
		&item.AreaCode,
		&item.Role,
		&item.EffectiveFrom,
		&effectiveTo,
		&grantedBy,
	); err != nil {
		return iamdomain.UserProcessArea{}, fmt.Errorf("scan user process area: %w", err)
	}
	if effectiveTo.Valid {
		value := effectiveTo.Time
		item.EffectiveTo = &value
	}
	if grantedBy.Valid {
		value := grantedBy.String
		item.GrantedBy = &value
	}
	return item, nil
}

func rollbackTx(action string, tx *sql.Tx) {
	if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		slog.Warn("iam tx rollback failed", "action", action, "error", err)
	}
}
