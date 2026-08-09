// Package infrastructure is the taxonomy module's Postgres-backed
// implementation of the domain repository ports (ProfileRepository,
// AreaRepository, FamilyRepository) plus the cross-module read adapters
// (AreaCatalogReaderPG, FamilyCodeResolverRepository). Every write path
// seeds the authz GUCs (setAuthzGUC) and calls authz.Require before
// mutating a row, pairing with the trg_require_cap_asserted DB tripwire.
package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"

	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/taxonomy/domain"
)

// ProfileRepository is the *sql.DB-backed implementation of
// domain.ProfileRepository. Non-Tx methods each own a single-statement
// transaction; Tx-suffixed methods run inside a caller-supplied
// domain.FamilyTx (concretely a taxonomyTx).
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

func (t taxonomyTx) Commit() error   { return mapProfilePolicyErr(t.tx.Commit()) }
func (t taxonomyTx) Rollback() error { return t.tx.Rollback() }

// mapProfilePolicyErr translates the G1 governance-policy DB triggers'
// P0001 exceptions (migration 0295) into taxonomy domain sentinels so the
// application/domain layers never inspect driver error codes. The
// reclassify guard is a DEFERRABLE INITIALLY DEFERRED constraint trigger, so
// its violation surfaces at COMMIT — hence the mapping lives on taxonomyTx.
// Commit. Non-matching errors pass through unchanged.
func mapProfilePolicyErr(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "P0001" {
		switch {
		case strings.HasPrefix(pgErr.Message, "ErrClassChangeRouteConflict"):
			return domain.ErrClassChangeRouteConflict
		case strings.HasPrefix(pgErr.Message, "ErrRouteViolatesProfilePolicy"):
			return domain.ErrRouteViolatesProfilePolicy
		}
	}
	return err
}

// db.Tx forwarding — satisfies domain.FamilyTx which embeds db.Tx (F-07).
func (t taxonomyTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}
func (t taxonomyTx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}
func (t taxonomyTx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return t.tx.QueryRowContext(ctx, query, args...)
}

// NewProfileRepository builds a ProfileRepository backed by db.
func NewProfileRepository(db *sql.DB) *ProfileRepository {
	return &ProfileRepository{db: db}
}

// BeginTx opens a new database transaction and returns it wrapped as a
// domain.FamilyTx (taxonomyTx). The caller owns Commit/Rollback.
func (r *ProfileRepository) BeginTx(ctx context.Context) (domain.FamilyTx, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin profile tx: %w", err)
	}
	return taxonomyTx{tx: tx}, nil
}

// profileByCodeQuery is the single (tenantID, code) profile read shared by the
// identity-gated GetByCode and the background GetByCodeSystem; both scan it
// through scanProfileRow, so the column list exists once.
const profileByCodeQuery = `
SELECT code, tenant_id, family_code, name, description, alias, review_interval_days,
       default_template_version_id, owner_user_id, editable_by_role, governance_class, archived_at, created_at
FROM metaldocs.document_profiles
WHERE tenant_id = $1 AND code = $2`

// scanProfileRow scans one profileByCodeQuery row, mapping sql.ErrNoRows to
// domain.ErrProfileNotFound.
func scanProfileRow(row *sql.Row) (*domain.DocumentProfile, error) {
	var profile domain.DocumentProfile
	var defaultTemplateVersionID sql.NullString
	var ownerUserID sql.NullString
	err := row.Scan(
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
		&profile.GovernanceClass,
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

// GetByCode reads the profile for (tenantID, code) inside a short-lived
// read transaction, seeding the authz GUCs and requiring CapTaxonomyView.
// Returns domain.ErrProfileNotFound if no such row exists.
func (r *ProfileRepository) GetByCode(ctx context.Context, tenantID string, code domain.ProfileCode) (*domain.DocumentProfile, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin get profile tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := setAuthzGUC(ctx, tx); err != nil {
		return nil, fmt.Errorf("query profile %q: %w", code, err)
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapTaxonomyView), "tenant"); err != nil {
		return nil, fmt.Errorf("taxonomy: authz check Get profile: %w", err)
	}

	return scanProfileRow(tx.QueryRowContext(ctx, profileByCodeQuery, tenantID, code))
}

// GetByCodeSystem reads the profile for (tenantID, code) on the background
// system path: River workers carry no HTTP identity, so setAuthzGUC's ctx
// tenant/actor resolution and the CapTaxonomyView check cannot apply. Instead
// the tenant GUC is seeded from the EXPLICIT tenantID parameter (RLS scoping,
// never ctx), and authz.BypassSystem stands in for tier-2 — it is fail-closed
// (ErrBypassNotBackground unless ctx carries authz.WithBackgroundBypass, so
// this method is unreachable from any HTTP path) and audits every invocation
// per ADR 0022 Phase 11 F8. Precedent: the release coordinator's own phase-1
// preflight tx (SeedTxTenant + BypassSystem, release_coordinator.go).
// Returns domain.ErrProfileNotFound if no such row exists.
//
// COMMIT CONSTRAINT — this read transaction is NOT throwaway: BypassSystem's F8
// bypass-audit event is written by the sink INSIDE this tx (bypass_audit.go
// RecordBypass), so returning without Commit would roll the audit record back
// and let the bypass happen unaudited. Every exit AFTER a successful
// BypassSystem therefore commits first — including the not-found path, where the
// bypass and the read attempt still happened. The deferred Rollback stays as the
// safety net for the pre-bypass and infrastructure-error paths (a Rollback after
// Commit is a no-op).
func (r *ProfileRepository) GetByCodeSystem(ctx context.Context, tenantID string, code domain.ProfileCode) (*domain.DocumentProfile, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin get profile system tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := authz.SeedTxTenant(ctx, tx, tenantID); err != nil {
		return nil, fmt.Errorf("taxonomy: seed tenant for system profile read %q: %w", code, err)
	}
	if err := authz.BypassSystem(ctx, tx); err != nil {
		return nil, fmt.Errorf("taxonomy: system profile read %q: %w", code, err)
	}

	profile, scanErr := scanProfileRow(tx.QueryRowContext(ctx, profileByCodeQuery, tenantID, code))
	if scanErr != nil && !errors.Is(scanErr, domain.ErrProfileNotFound) {
		return nil, scanErr
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit system profile read %q: %w", code, err)
	}
	if scanErr != nil {
		return nil, scanErr
	}
	return profile, nil
}

// List returns profiles for tenantID, ordered by code, capped at
// maxTaxonomyListRows (no real pagination — T-015). Requires
// CapTaxonomyView.
func (r *ProfileRepository) List(ctx context.Context, tenantID string, includeArchived bool) ([]domain.DocumentProfile, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin list profiles tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := setAuthzGUC(ctx, tx); err != nil {
		return nil, err
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapTaxonomyView), "tenant"); err != nil {
		return nil, fmt.Errorf("taxonomy: authz check List profiles: %w", err)
	}

	q := `
SELECT code, tenant_id, family_code, name, description, alias, review_interval_days,
       default_template_version_id, owner_user_id, editable_by_role, governance_class, archived_at, created_at
FROM metaldocs.document_profiles
WHERE tenant_id = $1`
	if !includeArchived {
		q += " AND archived_at IS NULL"
	}
	q += " ORDER BY code ASC LIMIT " + strconv.Itoa(maxTaxonomyListRows) // #nosec G202 -- maxTaxonomyListRows is a package-level compile-time int constant, never derived from user/tenant input. TODO: add pagination instead of returning the full profile catalog.

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
			&profile.GovernanceClass,
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

// Create inserts p inside its own transaction, requiring CapTaxonomyManage
// before the INSERT; the transaction commits only on success.
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
    (code, tenant_id, family_code, name, description, alias, review_interval_days, default_template_version_id, owner_user_id, editable_by_role, governance_class, archived_at)
VALUES
    ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

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
		p.GovernanceClass,
		p.ArchivedAt,
	); err != nil {
		return err
	}
	return tx.Commit()
}

// CreateTx inserts a new DocumentProfile using the caller-owned transaction.
// The caller is responsible for committing or rolling back the transaction.
func (r *ProfileRepository) CreateTx(ctx context.Context, tx domain.FamilyTx, p *domain.DocumentProfile) error {
	sqlTx, ok := tx.(taxonomyTx)
	if !ok {
		return fmt.Errorf("taxonomy: ProfileRepository.CreateTx: unsupported tx type %T", tx)
	}
	if err := setAuthzGUC(ctx, sqlTx.tx); err != nil {
		return fmt.Errorf("insert profile %q: %w", p.Code, err)
	}
	if err := authz.Require(ctx, sqlTx.tx, string(iamdomain.CapTaxonomyManage), "tenant"); err != nil {
		return fmt.Errorf("taxonomy: authz check CreateTx profile: %w", err)
	}
	const q = `
INSERT INTO metaldocs.document_profiles
    (code, tenant_id, family_code, name, description, alias, review_interval_days, default_template_version_id, owner_user_id, editable_by_role, governance_class, archived_at)
VALUES
    ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := sqlTx.tx.ExecContext(
		ctx, q,
		p.Code, p.TenantID, p.FamilyCode, p.Name, p.Description, p.Alias,
		p.ReviewIntervalDays, stringPtrToNull(p.DefaultTemplateVersionID),
		stringPtrToNull(p.OwnerUserID), p.EditableByRole, p.GovernanceClass, p.ArchivedAt,
	)
	return err
}

// Update writes p's mutable fields inside its own transaction, requiring
// CapTaxonomyManage before the UPDATE, and commits only on success.
// Returns domain.ErrProfileNotFound if the (tenant_id, code) row does not
// exist.
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

// GetByCodeForUpdate reads the profile for (tenantID, code) with a SELECT
// ... FOR UPDATE inside tx, requiring CapTaxonomyManage. tx must be the
// taxonomyTx returned by BeginTx. Returns domain.ErrProfileNotFound if no
// such row exists.
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
       default_template_version_id, owner_user_id, editable_by_role, governance_class, archived_at, created_at
FROM metaldocs.document_profiles
WHERE tenant_id = $1 AND code = $2
FOR UPDATE`
	var profile domain.DocumentProfile
	var defaultTemplateVersionID sql.NullString
	var ownerUserID sql.NullString
	err := sqlTx.tx.QueryRowContext(ctx, q, tenantID, code).Scan(
		&profile.Code, &profile.TenantID, &profile.FamilyCode, &profile.Name, &profile.Description,
		&profile.Alias, &profile.ReviewIntervalDays, &defaultTemplateVersionID, &ownerUserID,
		&profile.EditableByRole, &profile.GovernanceClass, &profile.ArchivedAt, &profile.CreatedAt,
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

// UpdateTx writes p's mutable fields inside the caller-owned tx, requiring
// CapTaxonomyManage before the UPDATE. The caller is responsible for
// committing or rolling back. Returns domain.ErrProfileNotFound if the
// (tenant_id, code) row does not exist.
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

// SetGovernanceClassTx writes ONLY the governance_class column for
// (tenantID, code) inside the caller-owned tx, requiring CapTaxonomyManage.
// It is the narrow reclassification write (the generic Update/UpdateTx path
// never touches governance_class). Returns domain.ErrProfileNotFound if the
// row does not exist. Any route-shape conflict is enforced by the deferred
// reclassify-guard trigger and surfaces at Commit (mapped to
// domain.ErrClassChangeRouteConflict by taxonomyTx.Commit).
func (r *ProfileRepository) SetGovernanceClassTx(ctx context.Context, tx domain.FamilyTx, tenantID string, code domain.ProfileCode, class domain.GovernanceClass) error {
	sqlTx, ok := tx.(taxonomyTx)
	if !ok {
		return fmt.Errorf("invalid taxonomy tx type %T", tx)
	}
	if err := setAuthzGUC(ctx, sqlTx.tx); err != nil {
		return err
	}
	if err := authz.Require(ctx, sqlTx.tx, string(iamdomain.CapTaxonomyManage), "tenant"); err != nil {
		return fmt.Errorf("taxonomy: authz check set governance_class: %w", err)
	}
	const q = `
UPDATE metaldocs.document_profiles
SET governance_class = $1
WHERE tenant_id = $2 AND code = $3`
	result, err := sqlTx.tx.ExecContext(ctx, q, class, tenantID, code)
	if err != nil {
		return fmt.Errorf("set governance_class %q: %w", code, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("set governance_class rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrProfileNotFound
	}
	return nil
}

// AreaRepository is the *sql.DB-backed implementation of
// domain.AreaRepository. Non-Tx methods each own a single-statement
// transaction; Tx-suffixed methods run inside a caller-supplied
// domain.FamilyTx (concretely a taxonomyTx).
type AreaRepository struct {
	db *sql.DB
}

// NewAreaRepository builds an AreaRepository backed by db.
func NewAreaRepository(db *sql.DB) *AreaRepository {
	return &AreaRepository{db: db}
}

// BeginTx opens a new database transaction and returns it wrapped as a
// domain.FamilyTx (taxonomyTx). The caller owns Commit/Rollback.
func (r *AreaRepository) BeginTx(ctx context.Context) (domain.FamilyTx, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin area tx: %w", err)
	}
	return taxonomyTx{tx: tx}, nil
}

// GetByCode reads the area for (tenantID, code) inside a short-lived read
// transaction, seeding the authz GUCs and requiring CapTaxonomyView.
// Returns domain.ErrAreaNotFound if no such row exists.
func (r *AreaRepository) GetByCode(ctx context.Context, tenantID string, code domain.AreaCode) (*domain.ProcessArea, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin get area tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := setAuthzGUC(ctx, tx); err != nil {
		return nil, fmt.Errorf("query area %q: %w", code, err)
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapTaxonomyView), "tenant"); err != nil {
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

// List returns process areas for tenantID, ordered by code, capped at
// maxTaxonomyListRows (no real pagination — T-015). Requires
// CapTaxonomyView.
func (r *AreaRepository) List(ctx context.Context, tenantID string, includeArchived bool) ([]domain.ProcessArea, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin list areas tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := setAuthzGUC(ctx, tx); err != nil {
		return nil, err
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapTaxonomyView), "tenant"); err != nil {
		return nil, fmt.Errorf("taxonomy: authz check List areas: %w", err)
	}

	q := `
SELECT code, tenant_id, name, description, parent_code, owner_user_id, default_approver_role, archived_at, created_at
FROM metaldocs.document_process_areas
WHERE tenant_id = $1`
	if !includeArchived {
		q += " AND archived_at IS NULL"
	}
	q += " ORDER BY code ASC LIMIT " + strconv.Itoa(maxTaxonomyListRows) // #nosec G202 -- maxTaxonomyListRows is a package-level compile-time int constant, never derived from user/tenant input. TODO: add pagination instead of returning the full area catalog.

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

// Create inserts a inside its own transaction, requiring CapTaxonomyManage
// before the INSERT; the transaction commits only on success.
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

// CreateTx inserts a new ProcessArea using the caller-owned transaction.
// The caller is responsible for committing or rolling back the transaction.
func (r *AreaRepository) CreateTx(ctx context.Context, tx domain.FamilyTx, a *domain.ProcessArea) error {
	sqlTx, ok := tx.(taxonomyTx)
	if !ok {
		return fmt.Errorf("taxonomy: AreaRepository.CreateTx: unsupported tx type %T", tx)
	}
	if err := setAuthzGUC(ctx, sqlTx.tx); err != nil {
		return fmt.Errorf("insert area %q: %w", a.Code, err)
	}
	if err := authz.Require(ctx, sqlTx.tx, string(iamdomain.CapTaxonomyManage), "tenant"); err != nil {
		return fmt.Errorf("taxonomy: authz check CreateTx area: %w", err)
	}
	const q = `
INSERT INTO metaldocs.document_process_areas
    (code, tenant_id, name, description, parent_code, owner_user_id, default_approver_role, archived_at)
VALUES
    ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := sqlTx.tx.ExecContext(
		ctx, q,
		a.Code, a.TenantID, a.Name, a.Description,
		areaCodePtrToNull(a.ParentCode), stringPtrToNull(a.OwnerUserID),
		stringPtrToNull(a.DefaultApproverRole), a.ArchivedAt,
	)
	return err
}

// Update writes a's mutable fields inside its own transaction, requiring
// CapTaxonomyManage before the UPDATE, and commits only on success.
// Returns domain.ErrAreaNotFound if the (tenant_id, code) row does not
// exist.
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

// GetByCodeForUpdate reads the area for (tenantID, code) with a SELECT ...
// FOR UPDATE inside tx, requiring CapTaxonomyManage. tx must be the
// taxonomyTx returned by BeginTx. Returns domain.ErrAreaNotFound if no such
// row exists.
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

// ListAncestorsTx walks the parent_code chain for (tenantID, code) up to
// maxTaxonomyTreeDepth inside the caller-owned tx, returning the ancestor
// codes. It performs no authz check of its own (read-only recursive CTE);
// callers that need a capability check must gate the surrounding
// operation.
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

// UpdateTx writes a's mutable fields inside the caller-owned tx, requiring
// CapTaxonomyManage before the UPDATE. The caller is responsible for
// committing or rolling back. Returns domain.ErrAreaNotFound if the
// (tenant_id, code) row does not exist.
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

// ListAncestors walks the parent_code chain for (tenantID, code) up to
// maxTaxonomyTreeDepth inside its own short-lived transaction, seeding the
// authz GUCs and requiring CapTaxonomyView. Used by the (now-deleted)
// SetParent cycle check; retained as a standalone read.
func (r *AreaRepository) ListAncestors(ctx context.Context, tenantID string, code domain.AreaCode) ([]domain.AreaCode, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin list area ancestors tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := setAuthzGUC(ctx, tx); err != nil {
		return nil, fmt.Errorf("query area ancestors %q: %w", code, err)
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapTaxonomyView), "tenant"); err != nil {
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
