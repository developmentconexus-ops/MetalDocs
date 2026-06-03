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
FROM public.user_process_areas
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

// ListByTenant returns active memberships across the whole tenant, with
// optional exact-match filters. An empty filter string means "no filter" for
// that column; the ($n = '' OR col = $n) form keeps the statement static and
// injection-safe regardless of which filters are supplied.
func (r *UserAreaRepository) ListByTenant(ctx context.Context, tenantID, userID, areaCode, role string, now time.Time) ([]iamdomain.UserProcessArea, error) {
	const q = `
SELECT user_id, tenant_id::text, area_code, role, effective_from, effective_to, granted_by
FROM public.user_process_areas
WHERE tenant_id = $1::uuid
  AND effective_from <= $2
  AND (effective_to IS NULL OR effective_to > $2)
  AND ($3 = '' OR user_id   = $3)
  AND ($4 = '' OR area_code = $4)
  AND ($5 = '' OR role      = $5)
ORDER BY user_id ASC, area_code ASC, effective_from DESC
`
	rows, err := r.db.QueryContext(ctx, q, tenantID, now, userID, areaCode, role)
	if err != nil {
		return nil, fmt.Errorf("query tenant process areas: %w", err)
	}
	defer rows.Close()

	result := make([]iamdomain.UserProcessArea, 0, 16)
	for rows.Next() {
		item, err := scanUserProcessArea(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tenant process areas: %w", err)
	}
	return result, nil
}

func (r *UserAreaRepository) Insert(ctx context.Context, membership iamdomain.UserProcessArea) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin insert area tx: %w", err)
	}
	defer rollbackTx("insert_user_process_area", tx)

	ctx = authz.WithCapCache(ctx)
	if err := authz.SeedTxIdentity(ctx, tx, membership.TenantID, grantedByActor(membership.GrantedBy)); err != nil {
		return fmt.Errorf("iam: seed authz Insert area: %w", err)
	}
	// ADR 0022 Phase 3: area-scoped tier-2 — pass the membership's real area so
	// area_admin is authorized only within areas where they hold membership.manage.
	// system_admin still bypasses (authz.Require short-circuits before the area filter).
	if err := authz.Require(ctx, tx, string(iamdomain.CapMembershipManage), membership.AreaCode); err != nil {
		return fmt.Errorf("iam: authz check Insert area: %w", err)
	}

	const q = `
INSERT INTO public.user_process_areas
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

func (r *UserAreaRepository) CloseActive(ctx context.Context, userID, tenantID, areaCode string, effectiveTo time.Time, actorID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin close area tx: %w", err)
	}
	defer rollbackTx("close_active_user_process_area", tx)

	ctx = authz.WithCapCache(ctx)
	if err := authz.SeedTxIdentity(ctx, tx, tenantID, actorID); err != nil {
		return fmt.Errorf("iam: seed authz CloseActive area: %w", err)
	}
	// ADR 0022 Phase 3: area-scoped tier-2 — see Insert.
	if err := authz.Require(ctx, tx, string(iamdomain.CapMembershipManage), areaCode); err != nil {
		return fmt.Errorf("iam: authz check CloseActive area: %w", err)
	}

	const q = `
UPDATE public.user_process_areas
SET effective_to = $4,
    revoked_by   = $5
WHERE user_id = $1
  AND tenant_id = $2::uuid
  AND area_code = $3
  AND effective_to IS NULL
`
	result, err := tx.ExecContext(ctx, q, userID, tenantID, areaCode, effectiveTo, actorID)
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

	// ADR 0022 Phase 3: the single area-scoped tier-2 check below governs both the
	// close (old) and insert (new) legs, so they MUST target the same area. The
	// service only ever rebuilds in place; enforce that as a hard repository
	// invariant so a future caller can't close a row in an area it isn't
	// authorized for by pairing it with an authorized new area.
	if oldMembership.AreaCode != newMembership.AreaCode {
		return fmt.Errorf("iam: GrantAtomic area mismatch: old=%q new=%q", oldMembership.AreaCode, newMembership.AreaCode)
	}

	ctx = authz.WithCapCache(ctx)
	if err := authz.SeedTxIdentity(ctx, tx, newMembership.TenantID, grantedByActor(newMembership.GrantedBy)); err != nil {
		return fmt.Errorf("iam: seed authz GrantAtomic: %w", err)
	}
	// Area-scoped tier-2 — old/new share one area (asserted above), so the new
	// area governs the check.
	if err := authz.Require(ctx, tx, string(iamdomain.CapMembershipManage), newMembership.AreaCode); err != nil {
		return fmt.Errorf("iam: authz check GrantAtomic: %w", err)
	}

	const closeQ = `
UPDATE public.user_process_areas
SET effective_to = $5,
    revoked_by   = $6
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
		grantedByActor(newMembership.GrantedBy),
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
INSERT INTO public.user_process_areas
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
FROM public.user_process_areas
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

func grantedByActor(actorID *string) string {
	if actorID == nil {
		return ""
	}
	return *actorID
}
