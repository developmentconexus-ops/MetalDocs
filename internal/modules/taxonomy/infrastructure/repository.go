package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/taxonomy/domain"
)

type ProfileRepository struct {
	db *sql.DB
}

const (
	maxTaxonomyListRows  = 1000
	maxTaxonomyTreeDepth = 20
)

type taxonomyTx struct {
	tx *sql.Tx
}

func (t taxonomyTx) Commit() error   { return t.tx.Commit() }
func (t taxonomyTx) Rollback() error { return t.tx.Rollback() }

func NewProfileRepository(db *sql.DB) *ProfileRepository {
	return &ProfileRepository{db: db}
}

func (r *ProfileRepository) BeginTx(ctx context.Context) (domain.FamilyTx, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin profile tx: %w", err)
	}
	return taxonomyTx{tx: tx}, nil
}

func (r *ProfileRepository) GetByCode(ctx context.Context, tenantID string, code domain.ProfileCode) (*domain.DocumentProfile, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin get profile tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := setAuthzGUC(ctx, tx); err != nil {
		return nil, fmt.Errorf("query profile %q: %w", code, err)
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentView), "tenant"); err != nil {
		return nil, fmt.Errorf("taxonomy: authz check Get profile: %w", err)
	}

	const q = `
SELECT code, tenant_id, family_code, name, description, alias, review_interval_days,
       default_template_version_id, owner_user_id, editable_by_role, archived_at, created_at
FROM metaldocs.document_profiles
WHERE tenant_id = $1 AND code = $2`

	var profile domain.DocumentProfile
	var defaultTemplateVersionID sql.NullString
	var ownerUserID sql.NullString
	err = tx.QueryRowContext(ctx, q, tenantID, code).Scan(
		&profile.Code,
		&profile.TenantID,
		&profile.FamilyCode,
		&profile.Name,
		&profile.Description,
		&profile.Alias,
		&profile.ReviewIntervalDays,
		&defaultTemplateVersionID,
		&ownerUserID,
		&profile.EditableByRole,
		&profile.ArchivedAt,
		&profile.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrProfileNotFound
	}
	if err != nil {
		return nil, err
	}
	profile.DefaultTemplateVersionID = nullStringPtr(defaultTemplateVersionID)
	profile.OwnerUserID = nullStringPtr(ownerUserID)
	return &profile, nil
}

func (r *ProfileRepository) List(ctx context.Context, tenantID string, includeArchived bool) ([]domain.DocumentProfile, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin list profiles tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := setAuthzGUC(ctx, tx); err != nil {
		return nil, err
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentView), "tenant"); err != nil {
		return nil, fmt.Errorf("taxonomy: authz check List profiles: %w", err)
	}

	q := `
SELECT code, tenant_id, family_code, name, description, alias, review_interval_days,
       default_template_version_id, owner_user_id, editable_by_role, archived_at, created_at
FROM metaldocs.document_profiles
WHERE tenant_id = $1`
	if !includeArchived {
		q += " AND archived_at IS NULL"
	}
	q += " ORDER BY code ASC LIMIT " + strconv.Itoa(maxTaxonomyListRows) // TODO: add pagination instead of returning the full profile catalog.

	rows, err := tx.QueryContext(ctx, q, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.DocumentProfile, 0)
	for rows.Next() {
		var profile domain.DocumentProfile
		var defaultTemplateVersionID sql.NullString
		var ownerUserID sql.NullString
		if err := rows.Scan(
			&profile.Code,
			&profile.TenantID,
			&profile.FamilyCode,
			&profile.Name,
			&profile.Description,
			&profile.Alias,
			&profile.ReviewIntervalDays,
			&defaultTemplateVersionID,
			&ownerUserID,
			&profile.EditableByRole,
			&profile.ArchivedAt,
			&profile.CreatedAt,
		); err != nil {
			return nil, err
		}
		profile.DefaultTemplateVersionID = nullStringPtr(defaultTemplateVersionID)
		profile.OwnerUserID = nullStringPtr(ownerUserID)
		out = append(out, profile)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *ProfileRepository) Create(ctx context.Context, p *domain.DocumentProfile) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin create profile tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := setAuthzGUC(ctx, tx); err != nil {
		return fmt.Errorf("insert profile %q: %w", p.Code, err)
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapTaxonomyManage), "tenant"); err != nil {
		return fmt.Errorf("taxonomy: authz check Create profile: %w", err)
	}
	const q = `
INSERT INTO metaldocs.document_profiles
    (code, tenant_id, family_code, name, description, alias, review_interval_days, default_template_version_id, owner_user_id, editable_by_role, archived_at)
VALUES
    ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	if _, err := tx.ExecContext(
		ctx,
		q,
		p.Code,
		p.TenantID,
		p.FamilyCode,
		p.Name,
		p.Description,
		p.Alias,
		p.ReviewIntervalDays,
		stringPtrToNull(p.DefaultTemplateVersionID),
		stringPtrToNull(p.OwnerUserID),
		p.EditableByRole,
		p.ArchivedAt,
	); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *ProfileRepository) Update(ctx context.Context, p *domain.DocumentProfile) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin update profile tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := setAuthzGUC(ctx, tx); err != nil {
		return err
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapTaxonomyManage), "tenant"); err != nil {
		return fmt.Errorf("taxonomy: authz check Update profile: %w", err)
	}
	const q = `
UPDATE metaldocs.document_profiles
SET family_code = $1,
    name = $2,
    description = $3,
    alias = $4,
    review_interval_days = $5,
    default_template_version_id = $6,
    owner_user_id = $7,
    editable_by_role = $8,
    archived_at = $9
WHERE tenant_id = $10 AND code = $11`

	result, err := tx.ExecContext(
		ctx,
		q,
		p.FamilyCode,
		p.Name,
		p.Description,
		p.Alias,
		p.ReviewIntervalDays,
		stringPtrToNull(p.DefaultTemplateVersionID),
		stringPtrToNull(p.OwnerUserID),
		p.EditableByRole,
		p.ArchivedAt,
		p.TenantID,
		p.Code,
	)
	if err != nil {
		return fmt.Errorf("update profile %q: %w", p.Code, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("profile update rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrProfileNotFound
	}
	return tx.Commit()
}

func (r *ProfileRepository) GetByCodeForUpdate(ctx context.Context, tx domain.FamilyTx, tenantID string, code domain.ProfileCode) (*domain.DocumentProfile, error) {
	sqlTx, ok := tx.(taxonomyTx)
	if !ok {
		return nil, fmt.Errorf("invalid taxonomy tx type %T", tx)
	}
	if err := setAuthzGUC(ctx, sqlTx.tx); err != nil {
		return nil, fmt.Errorf("query profile for update %q: %w", code, err)
	}
	if err := authz.Require(ctx, sqlTx.tx, string(iamdomain.CapTaxonomyManage), "tenant"); err != nil {
		return nil, fmt.Errorf("taxonomy: authz check Get profile for update: %w", err)
	}
	const q = `
SELECT code, tenant_id, family_code, name, description, alias, review_interval_days,
       default_template_version_id, owner_user_id, editable_by_role, archived_at, created_at
FROM metaldocs.document_profiles
WHERE tenant_id = $1 AND code = $2
FOR UPDATE`
	var profile domain.DocumentProfile
	var defaultTemplateVersionID sql.NullString
	var ownerUserID sql.NullString
	err := sqlTx.tx.QueryRowContext(ctx, q, tenantID, code).Scan(
		&profile.Code, &profile.TenantID, &profile.FamilyCode, &profile.Name, &profile.Description,
		&profile.Alias, &profile.ReviewIntervalDays, &defaultTemplateVersionID, &ownerUserID,
		&profile.EditableByRole, &profile.ArchivedAt, &profile.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrProfileNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query profile for update %q: %w", code, err)
	}
	profile.DefaultTemplateVersionID = nullStringPtr(defaultTemplateVersionID)
	profile.OwnerUserID = nullStringPtr(ownerUserID)
	return &profile, nil
}

func (r *ProfileRepository) UpdateTx(ctx context.Context, tx domain.FamilyTx, p *domain.DocumentProfile) error {
	sqlTx, ok := tx.(taxonomyTx)
	if !ok {
		return fmt.Errorf("invalid taxonomy tx type %T", tx)
	}
	if err := setAuthzGUC(ctx, sqlTx.tx); err != nil {
		return err
	}
	if err := authz.Require(ctx, sqlTx.tx, string(iamdomain.CapTaxonomyManage), "tenant"); err != nil {
		return fmt.Errorf("taxonomy: authz check Update profile: %w", err)
	}
	const q = `
UPDATE metaldocs.document_profiles
SET family_code = $1,
    name = $2,
    description = $3,
    alias = $4,
    review_interval_days = $5,
    default_template_version_id = $6,
    owner_user_id = $7,
    editable_by_role = $8,
    archived_at = $9
WHERE tenant_id = $10 AND code = $11`
	result, err := sqlTx.tx.ExecContext(ctx, q, p.FamilyCode, p.Name, p.Description, p.Alias, p.ReviewIntervalDays, stringPtrToNull(p.DefaultTemplateVersionID), stringPtrToNull(p.OwnerUserID), p.EditableByRole, p.ArchivedAt, p.TenantID, p.Code)
	if err != nil {
		return fmt.Errorf("update profile tx %q: %w", p.Code, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("profile tx update rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrProfileNotFound
	}
	return nil
}

type AreaRepository struct {
	db *sql.DB
}

func NewAreaRepository(db *sql.DB) *AreaRepository {
	return &AreaRepository{db: db}
}

func (r *AreaRepository) BeginTx(ctx context.Context) (domain.FamilyTx, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin area tx: %w", err)
	}
	return taxonomyTx{tx: tx}, nil
}

func (r *AreaRepository) GetByCode(ctx context.Context, tenantID string, code domain.AreaCode) (*domain.ProcessArea, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin get area tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := setAuthzGUC(ctx, tx); err != nil {
		return nil, fmt.Errorf("query area %q: %w", code, err)
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentView), "tenant"); err != nil {
		return nil, fmt.Errorf("taxonomy: authz check Get area: %w", err)
	}

	const q = `
SELECT code, tenant_id, name, description, parent_code, owner_user_id, default_approver_role, archived_at, created_at
FROM metaldocs.document_process_areas
WHERE tenant_id = $1 AND code = $2`

	var area domain.ProcessArea
	var parentCode sql.NullString
	var ownerUserID sql.NullString
	var defaultApproverRole sql.NullString
	err = tx.QueryRowContext(ctx, q, tenantID, code).Scan(
		&area.Code,
		&area.TenantID,
		&area.Name,
		&area.Description,
		&parentCode,
		&ownerUserID,
		&defaultApproverRole,
		&area.ArchivedAt,
		&area.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrAreaNotFound
	}
	if err != nil {
		return nil, err
	}
	area.ParentCode = nullAreaCodePtr(parentCode)
	area.OwnerUserID = nullStringPtr(ownerUserID)
	area.DefaultApproverRole = nullStringPtr(defaultApproverRole)
	return &area, nil
}

func (r *AreaRepository) List(ctx context.Context, tenantID string, includeArchived bool) ([]domain.ProcessArea, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin list areas tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := setAuthzGUC(ctx, tx); err != nil {
		return nil, err
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentView), "tenant"); err != nil {
		return nil, fmt.Errorf("taxonomy: authz check List areas: %w", err)
	}

	q := `
SELECT code, tenant_id, name, description, parent_code, owner_user_id, default_approver_role, archived_at, created_at
FROM metaldocs.document_process_areas
WHERE tenant_id = $1`
	if !includeArchived {
		q += " AND archived_at IS NULL"
	}
	q += " ORDER BY code ASC LIMIT " + strconv.Itoa(maxTaxonomyListRows) // TODO: add pagination instead of returning the full area catalog.

	rows, err := tx.QueryContext(ctx, q, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.ProcessArea, 0)
	for rows.Next() {
		var area domain.ProcessArea
		var parentCode sql.NullString
		var ownerUserID sql.NullString
		var defaultApproverRole sql.NullString
		if err := rows.Scan(
			&area.Code,
			&area.TenantID,
			&area.Name,
			&area.Description,
			&parentCode,
			&ownerUserID,
			&defaultApproverRole,
			&area.ArchivedAt,
			&area.CreatedAt,
		); err != nil {
			return nil, err
		}
		area.ParentCode = nullAreaCodePtr(parentCode)
		area.OwnerUserID = nullStringPtr(ownerUserID)
		area.DefaultApproverRole = nullStringPtr(defaultApproverRole)
		out = append(out, area)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *AreaRepository) Create(ctx context.Context, a *domain.ProcessArea) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin create area tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := setAuthzGUC(ctx, tx); err != nil {
		return fmt.Errorf("insert area %q: %w", a.Code, err)
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapTaxonomyManage), "tenant"); err != nil {
		return fmt.Errorf("taxonomy: authz check Create area: %w", err)
	}
	const q = `
INSERT INTO metaldocs.document_process_areas
    (code, tenant_id, name, description, parent_code, owner_user_id, default_approver_role, archived_at)
VALUES
    ($1, $2, $3, $4, $5, $6, $7, $8)`

	if _, err := tx.ExecContext(
		ctx,
		q,
		a.Code,
		a.TenantID,
		a.Name,
		a.Description,
		areaCodePtrToNull(a.ParentCode),
		stringPtrToNull(a.OwnerUserID),
		stringPtrToNull(a.DefaultApproverRole),
		a.ArchivedAt,
	); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *AreaRepository) Update(ctx context.Context, a *domain.ProcessArea) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin update area tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := setAuthzGUC(ctx, tx); err != nil {
		return err
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapTaxonomyManage), "tenant"); err != nil {
		return fmt.Errorf("taxonomy: authz check Update area: %w", err)
	}
	const q = `
UPDATE metaldocs.document_process_areas
SET name = $1,
    description = $2,
    parent_code = $3,
    owner_user_id = $4,
    default_approver_role = $5,
    archived_at = $6
WHERE tenant_id = $7 AND code = $8`

	result, err := tx.ExecContext(
		ctx,
		q,
		a.Name,
		a.Description,
		areaCodePtrToNull(a.ParentCode),
		stringPtrToNull(a.OwnerUserID),
		stringPtrToNull(a.DefaultApproverRole),
		a.ArchivedAt,
		a.TenantID,
		a.Code,
	)
	if err != nil {
		return fmt.Errorf("update area %q: %w", a.Code, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("area update rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrAreaNotFound
	}
	return tx.Commit()
}

func (r *AreaRepository) GetByCodeForUpdate(ctx context.Context, tx domain.FamilyTx, tenantID string, code domain.AreaCode) (*domain.ProcessArea, error) {
	sqlTx, ok := tx.(taxonomyTx)
	if !ok {
		return nil, fmt.Errorf("invalid taxonomy tx type %T", tx)
	}
	if err := setAuthzGUC(ctx, sqlTx.tx); err != nil {
		return nil, fmt.Errorf("query area for update %q: %w", code, err)
	}
	if err := authz.Require(ctx, sqlTx.tx, string(iamdomain.CapTaxonomyManage), "tenant"); err != nil {
		return nil, fmt.Errorf("taxonomy: authz check Get area for update: %w", err)
	}
	const q = `
SELECT code, tenant_id, name, description, parent_code, owner_user_id, default_approver_role, archived_at, created_at
FROM metaldocs.document_process_areas
WHERE tenant_id = $1 AND code = $2
FOR UPDATE`
	var area domain.ProcessArea
	var parentCode sql.NullString
	var ownerUserID sql.NullString
	var defaultApproverRole sql.NullString
	err := sqlTx.tx.QueryRowContext(ctx, q, tenantID, code).Scan(
		&area.Code, &area.TenantID, &area.Name, &area.Description, &parentCode, &ownerUserID, &defaultApproverRole, &area.ArchivedAt, &area.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrAreaNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query area for update %q: %w", code, err)
	}
	area.ParentCode = nullAreaCodePtr(parentCode)
	area.OwnerUserID = nullStringPtr(ownerUserID)
	area.DefaultApproverRole = nullStringPtr(defaultApproverRole)
	return &area, nil
}

func (r *AreaRepository) ListAncestorsTx(ctx context.Context, tx domain.FamilyTx, tenantID string, code domain.AreaCode) ([]domain.AreaCode, error) {
	sqlTx, ok := tx.(taxonomyTx)
	if !ok {
		return nil, fmt.Errorf("invalid taxonomy tx type %T", tx)
	}
	const q = `
	WITH RECURSIVE ancestors AS (
    SELECT p.code, p.parent_code, 1 AS depth
    FROM metaldocs.document_process_areas p
    WHERE p.tenant_id = $1
      AND p.code = (
          SELECT parent_code
          FROM metaldocs.document_process_areas
          WHERE tenant_id = $1 AND code = $2
      )
    UNION
    SELECT p.code, p.parent_code, a.depth + 1
    FROM metaldocs.document_process_areas p
    INNER JOIN ancestors a ON p.tenant_id = $1 AND p.code = a.parent_code
    WHERE a.depth < $3
)
SELECT code FROM ancestors`
	rows, err := sqlTx.tx.QueryContext(ctx, q, tenantID, code, maxTaxonomyTreeDepth)
	if err != nil {
		return nil, fmt.Errorf("query area ancestors tx %q: %w", code, err)
	}
	defer rows.Close()
	ancestors := make([]domain.AreaCode, 0)
	for rows.Next() {
		var ancestorCode domain.AreaCode
		if err := rows.Scan(&ancestorCode); err != nil {
			return nil, fmt.Errorf("scan area ancestor tx %q: %w", code, err)
		}
		ancestors = append(ancestors, ancestorCode)
	}
	return ancestors, rows.Err()
}

func (r *AreaRepository) UpdateTx(ctx context.Context, tx domain.FamilyTx, a *domain.ProcessArea) error {
	sqlTx, ok := tx.(taxonomyTx)
	if !ok {
		return fmt.Errorf("invalid taxonomy tx type %T", tx)
	}
	if err := setAuthzGUC(ctx, sqlTx.tx); err != nil {
		return err
	}
	if err := authz.Require(ctx, sqlTx.tx, string(iamdomain.CapTaxonomyManage), "tenant"); err != nil {
		return fmt.Errorf("taxonomy: authz check Update area: %w", err)
	}
	const q = `
UPDATE metaldocs.document_process_areas
SET name = $1,
    description = $2,
    parent_code = $3,
    owner_user_id = $4,
    default_approver_role = $5,
    archived_at = $6
WHERE tenant_id = $7 AND code = $8`
	result, err := sqlTx.tx.ExecContext(ctx, q, a.Name, a.Description, areaCodePtrToNull(a.ParentCode), stringPtrToNull(a.OwnerUserID), stringPtrToNull(a.DefaultApproverRole), a.ArchivedAt, a.TenantID, a.Code)
	if err != nil {
		return fmt.Errorf("update area tx %q: %w", a.Code, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("area tx update rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrAreaNotFound
	}
	return nil
}

func (r *AreaRepository) ListAncestors(ctx context.Context, tenantID string, code domain.AreaCode) ([]domain.AreaCode, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin list area ancestors tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := setAuthzGUC(ctx, tx); err != nil {
		return nil, fmt.Errorf("query area ancestors %q: %w", code, err)
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentView), "tenant"); err != nil {
		return nil, fmt.Errorf("taxonomy: authz check List area ancestors: %w", err)
	}

	const q = `
	WITH RECURSIVE ancestors AS (
    SELECT p.code, p.parent_code, 1 AS depth
    FROM metaldocs.document_process_areas p
    WHERE p.tenant_id = $1
      AND p.code = (
          SELECT parent_code
          FROM metaldocs.document_process_areas
          WHERE tenant_id = $1 AND code = $2
      )
    UNION
    SELECT p.code, p.parent_code, a.depth + 1
    FROM metaldocs.document_process_areas p
    INNER JOIN ancestors a ON p.tenant_id = $1 AND p.code = a.parent_code
    WHERE a.depth < $3
)
SELECT code FROM ancestors`

	rows, err := tx.QueryContext(ctx, q, tenantID, code, maxTaxonomyTreeDepth)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ancestors := make([]domain.AreaCode, 0)
	for rows.Next() {
		var ancestorCode domain.AreaCode
		if err := rows.Scan(&ancestorCode); err != nil {
			return nil, fmt.Errorf("scan area ancestor %q: %w", code, err)
		}
		ancestors = append(ancestors, ancestorCode)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ancestors, nil
}

func stringPtrToNull(v *string) sql.NullString {
	if v == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *v, Valid: true}
}

func nullStringPtr(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	value := v.String
	return &value
}

func areaCodePtrToNull(v *domain.AreaCode) sql.NullString {
	if v == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: string(*v), Valid: true}
}

func nullAreaCodePtr(v sql.NullString) *domain.AreaCode {
	if !v.Valid {
		return nil
	}
	value := domain.AreaCode(v.String)
	return &value
}
