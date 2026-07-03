// Package repository implements the templates module's PostgreSQL persistence
// layer: the Repository type backing application.Repository, plus the row
// scanning and mapping helpers it depends on.
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	auditdomain "metaldocs/internal/modules/audit/domain"
	"metaldocs/internal/modules/templates/application"
	"metaldocs/internal/modules/templates/domain"
	"metaldocs/internal/platform/db"
	tenant "metaldocs/internal/platform/tenant"
)

// isInvalidUUID returns true when err is a Postgres error with SQLSTATE 22P02
// (invalid text representation). We map malformed UUID lookups to not found.
func isInvalidUUID(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "22P02"
}

func rowsAffected(res sql.Result) (int64, error) {
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected: %w", err)
	}
	return n, nil
}

// Repository implements the templates application.Repository port against
// PostgreSQL, using database/sql for its own operations and accepting a
// caller-supplied db.Tx for the *Tx variants of its methods.
type Repository struct {
	db    *sql.DB
	audit auditdomain.Writer
}

// New constructs a Repository backed by the given database handle.
func New(db *sql.DB) *Repository { return &Repository{db: db} }

// WithAudit attaches an audit event writer to the Repository and returns it
// for chaining. Without a writer, AppendAudit and AppendAuditTx are no-ops.
func (r *Repository) WithAudit(w auditdomain.Writer) *Repository {
	r.audit = w
	return r
}

var _ application.Repository = (*Repository)(nil)

// CreateTemplate inserts a new template row for the tenant, executing
// directly against the repository's database handle. It maps a unique-key
// violation on (tenant_id, key) to domain.ErrKeyConflict.
func (r *Repository) CreateTemplate(ctx context.Context, t *domain.Template) error {
	const q = `
INSERT INTO templates_template (
	id, tenant_id, doc_type_code, key, name, description,
	latest_version, published_version_id, created_by, system_owned, created_at, archived_at
) VALUES (
	$1, $2::uuid, $3, $4, $5, $6,
	$7, $8, $9, $10, $11, $12
)`
	_, err := r.db.ExecContext(ctx, q,
		t.ID, t.TenantID, t.DocTypeCode, t.Key, t.Name, t.Description,
		t.LatestVersion, t.PublishedVersionID, t.CreatedBy, t.SystemOwned, t.CreatedAt, t.ArchivedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrKeyConflict
		}
		return err
	}
	return nil
}

// GetTemplate loads a template by ID, scoped to the tenant, joining in the
// latest and published version's revision numbers. It returns
// domain.ErrNotFound when the row is missing or id is not a valid UUID.
func (r *Repository) GetTemplate(ctx context.Context, tenantID, id string) (*domain.TemplateRead, error) {
	const q = `
SELECT
	t.id::text, t.tenant_id::text, t.doc_type_code, t.key, t.name, t.description,
	t.latest_version, lv.id::text, lv.revision_number, lv.status,
	t.published_version_id::text, pv.version_number, pv.revision_number, pv.status,
	t.created_by, t.system_owned, t.created_at, t.updated_at, t.archived_at
FROM templates_template t
LEFT JOIN templates_template_version pv ON pv.id = t.published_version_id
LEFT JOIN templates_template_version lv ON lv.template_id = t.id AND lv.version_number = t.latest_version
WHERE t.id = $1 AND t.tenant_id = $2::uuid`

	tmpl, err := scanTemplateRead(r.db.QueryRowContext(ctx, q, id, tenantID))
	if errors.Is(err, sql.ErrNoRows) || isInvalidUUID(err) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

// GetTemplateByKey loads a template by its tenant-scoped unique key, joining
// in the latest and published version's revision numbers. It returns
// domain.ErrNotFound when no matching row exists.
func (r *Repository) GetTemplateByKey(ctx context.Context, tenantID, key string) (*domain.TemplateRead, error) {
	const q = `
SELECT
	t.id::text, t.tenant_id::text, t.doc_type_code, t.key, t.name, t.description,
	t.latest_version, lv.id::text, lv.revision_number, lv.status,
	t.published_version_id::text, pv.version_number, pv.revision_number, pv.status,
	t.created_by, t.system_owned, t.created_at, t.updated_at, t.archived_at
FROM templates_template t
LEFT JOIN templates_template_version pv ON pv.id = t.published_version_id
LEFT JOIN templates_template_version lv ON lv.template_id = t.id AND lv.version_number = t.latest_version
WHERE t.tenant_id = $1::uuid AND t.key = $2`

	t, err := scanTemplateRead(r.db.QueryRowContext(ctx, q, tenantID, key))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return t, nil
}

// templatesDefaultLimit / templatesMaxLimit mirror the /templates OpenAPI
// contract's declared bounds (limit: min 1, max 200, default 50 — see
// api/openapi/v1/openapi.yaml), which are deliberately wider than the
// platform-wide internal/platform/pagination defaults (20/100). /templates is
// x-pagination-exempt: true ("a permanent bounded-list exemption, not a
// deferred cursor migration"), so this repo clamps defensively to the
// contract's own bounds rather than adopting the platform primitive, which
// would silently narrow the wire-promised max=200 down to 100 (T-011).
const (
	templatesDefaultLimit = 50
	templatesMaxLimit     = 200
)

// clampTemplatesLimit bounds limit to the /templates contract's declared
// range: <=0 -> default (50), >200 -> 200. The HTTP handler already rejects
// limit>200 with a 400 before reaching the repo; this clamp is a defensive
// second line so no caller (handler or otherwise) can send an unbounded or
// zero/negative LIMIT to Postgres.
func clampTemplatesLimit(limit int) int {
	if limit <= 0 {
		return templatesDefaultLimit
	}
	if limit > templatesMaxLimit {
		return templatesMaxLimit
	}
	return limit
}

// ListTemplates returns non-system templates for the tenant, optionally
// filtered by doc type code (a nil filter also always includes generic
// templates, i.e. those with an empty doc_type_code) and, when
// f.PublishedOnly is true, restricted to templates that currently have a
// published version, newest first, clamped to the /templates contract's
// pagination bounds via clampTemplatesLimit.
func (r *Repository) ListTemplates(ctx context.Context, f application.ListFilter) ([]*domain.TemplateRead, error) {
	const q = `
SELECT
	t.id::text, t.tenant_id::text, t.doc_type_code, t.key, t.name, t.description,
	t.latest_version, lv.id::text, lv.revision_number, lv.status,
	t.published_version_id::text, pv.version_number, pv.revision_number, pv.status,
	t.created_by, t.system_owned, t.created_at, t.updated_at, t.archived_at
FROM templates_template t
LEFT JOIN templates_template_version pv ON pv.id = t.published_version_id
LEFT JOIN templates_template_version lv ON lv.template_id = t.id AND lv.version_number = t.latest_version
WHERE t.tenant_id = $1::uuid
  AND t.system_owned = false
  -- A profile filter ($2) returns templates scoped to that profile PLUS generic
  -- templates (doc_type_code = ''), which by product definition apply to every
  -- profile — the template wizard's "Genérico" scope promises "templates
  -- genéricos aparecem para todos os perfis". A NULL filter returns every
  -- non-system template (management listing).
  AND ($2::text IS NULL OR t.doc_type_code = $2 OR t.doc_type_code = '')
  -- published=true ($5) restricts to templates that currently have a
  -- published version (the selectable-to-clone set for the new-document
  -- wizard); false/omitted keeps the management view (every status).
  AND ($5::bool IS NOT TRUE OR t.published_version_id IS NOT NULL)
ORDER BY t.created_at DESC, t.id DESC
LIMIT $3 OFFSET $4`

	limit := clampTemplatesLimit(f.Limit)
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}

	rows, err := r.db.QueryContext(
		ctx,
		q,
		f.TenantID,
		f.DocTypeCode,
		limit,
		offset,
		f.PublishedOnly,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]*domain.TemplateRead, 0)
	for rows.Next() {
		t, scanErr := scanTemplateRead(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// UpdateTemplate updates a template's mutable fields, scoped to the tenant,
// executing directly against the repository's database handle. It returns
// domain.ErrNotFound when no row matches (id, tenant_id).
func (r *Repository) UpdateTemplate(ctx context.Context, t *domain.Template) error {
	const q = `
UPDATE templates_template
SET
	doc_type_code = $3,
	key = $4,
	name = $5,
	description = $6,
	latest_version = $7,
	published_version_id = $8,
	system_owned = $9,
	archived_at = $10,
	updated_at = now()
WHERE id = $1 AND tenant_id = $2::uuid`

	res, err := r.db.ExecContext(ctx, q,
		t.ID, t.TenantID, t.DocTypeCode, t.Key, t.Name, t.Description,
		t.LatestVersion, t.PublishedVersionID, t.SystemOwned, t.ArchivedAt,
	)
	if err != nil {
		return err
	}
	n, err := rowsAffected(res)
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// CreateVersion inserts a new template version, executing directly against
// the repository's database handle. The revision_number is allocated
// atomically per template_id (ADR 0013) and scanned back into v.
func (r *Repository) CreateVersion(ctx context.Context, v *domain.TemplateVersion) error {
	metadataJSON, placeholderJSON, err := marshalVersionSchemas(v)
	if err != nil {
		return err
	}

	// ADR 0013: revision_number allocated atomically per template_id, mirroring
	// documents.revision_number (internal/modules/documents/repository/repository.go).
	// COALESCE(MAX(rn)+1, 0) yields 0 for the first version of a template and
	// increments thereafter. The DEFAULT 0 in the schema is a safety floor.
	const q = `
INSERT INTO templates_template_version (
	id, tenant_id, template_id, version_number, revision_number, status, docx_storage_key, content_hash,
	metadata_schema, placeholder_schema, author_id,
	pending_reviewer_role, pending_approver_role, reviewer_id, approver_id,
	submitted_at, reviewed_at, approved_at, published_at, obsoleted_at, lock_version, created_at
) VALUES (
	$1, $21::uuid, $2, $3,
	COALESCE((SELECT MAX(rn.revision_number) + 1
	            FROM templates_template_version rn
	           WHERE rn.template_id = $2), 0),
	$4, $5, $6,
	$7, $8, $9,
	$10, $11, $12, $13,
	$14, $15, $16, $17, $18, $19, $20
)
RETURNING revision_number`
	row := r.db.QueryRowContext(ctx, q,
		v.ID, v.TemplateID, v.VersionNumber, string(v.Status), v.DocxStorageKey, v.ContentHash,
		metadataJSON, placeholderJSON, v.AuthorID,
		v.PendingReviewerRole, v.PendingApproverRole, v.ReviewerID, v.ApproverID,
		v.SubmittedAt, v.ReviewedAt, v.ApprovedAt, v.PublishedAt, v.ObsoletedAt, v.LockVersion, v.CreatedAt,
		v.TenantID,
	)
	if err = row.Scan(&v.RevisionNumber); err != nil {
		return fmt.Errorf("templates repository create version: %w", err)
	}
	return nil
}

// CreateVersionTx inserts a new template version using the caller-supplied
// transaction. Same revision_number allocation (ADR 0013) as CreateVersion.
func (r *Repository) CreateVersionTx(ctx context.Context, tx db.Tx, v *domain.TemplateVersion) error {
	metadataJSON, placeholderJSON, err := marshalVersionSchemas(v)
	if err != nil {
		return err
	}
	// ADR 0013: see CreateVersion comment.
	const q = `
INSERT INTO templates_template_version (
	id, tenant_id, template_id, version_number, revision_number, status, docx_storage_key, content_hash,
	metadata_schema, placeholder_schema, author_id,
	pending_reviewer_role, pending_approver_role, reviewer_id, approver_id,
	submitted_at, reviewed_at, approved_at, published_at, obsoleted_at, lock_version, created_at
) VALUES (
	$1, $21::uuid, $2, $3,
	COALESCE((SELECT MAX(rn.revision_number) + 1
	            FROM templates_template_version rn
	           WHERE rn.template_id = $2), 0),
	$4, $5, $6,
	$7, $8, $9,
	$10, $11, $12, $13,
	$14, $15, $16, $17, $18, $19, $20
)
RETURNING revision_number`
	row := tx.QueryRowContext(ctx, q,
		v.ID, v.TemplateID, v.VersionNumber, string(v.Status), v.DocxStorageKey, v.ContentHash,
		metadataJSON, placeholderJSON, v.AuthorID,
		v.PendingReviewerRole, v.PendingApproverRole, v.ReviewerID, v.ApproverID,
		v.SubmittedAt, v.ReviewedAt, v.ApprovedAt, v.PublishedAt, v.ObsoletedAt, v.LockVersion, v.CreatedAt,
		v.TenantID,
	)
	if err := row.Scan(&v.RevisionNumber); err != nil {
		return fmt.Errorf("templates repository create version tx: %w", err)
	}
	return nil
}

// GetVersion loads a template version by its (templateID, version_number)
// pair, scoped to the tenant via a join on templates_template. It returns
// domain.ErrNotFound when the row is missing or an id is not a valid UUID.
func (r *Repository) GetVersion(ctx context.Context, tenantID, templateID string, n int) (*domain.TemplateVersion, error) {
	const q = `
SELECT
	v.id::text, v.template_id::text, v.version_number, v.revision_number, v.status, v.docx_storage_key, v.content_hash,
	v.metadata_schema, v.placeholder_schema, v.author_id,
	v.pending_reviewer_role, v.pending_approver_role, v.reviewer_id, v.approver_id,
	v.submitted_at, v.reviewed_at, v.approved_at, v.published_at, v.obsoleted_at, v.lock_version, v.created_at
FROM templates_template_version v
JOIN templates_template t ON t.id = v.template_id
WHERE v.template_id = $1 AND v.version_number = $2 AND t.tenant_id = $3::uuid`

	v, err := scanTemplateVersion(r.db.QueryRowContext(ctx, q, templateID, n, tenantID))
	if errors.Is(err, sql.ErrNoRows) || isInvalidUUID(err) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return v, nil
}

// GetVersionByID loads a template version by its own ID, scoped to the
// tenant via a join on templates_template. It returns domain.ErrNotFound
// when the row is missing or id is not a valid UUID.
func (r *Repository) GetVersionByID(ctx context.Context, tenantID, id string) (*domain.TemplateVersion, error) {
	const q = `
SELECT
	v.id::text, v.template_id::text, v.version_number, v.revision_number, v.status, v.docx_storage_key, v.content_hash,
	v.metadata_schema, v.placeholder_schema, v.author_id,
	v.pending_reviewer_role, v.pending_approver_role, v.reviewer_id, v.approver_id,
	v.submitted_at, v.reviewed_at, v.approved_at, v.published_at, v.obsoleted_at, v.lock_version, v.created_at
FROM templates_template_version v
JOIN templates_template t ON t.id = v.template_id
WHERE v.id = $1 AND t.tenant_id = $2::uuid`

	v, err := scanTemplateVersion(r.db.QueryRowContext(ctx, q, id, tenantID))
	if errors.Is(err, sql.ErrNoRows) || isInvalidUUID(err) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return v, nil
}

// UpdateVersion updates a template version's workflow fields, scoped to the
// tenant, executing directly against the repository's database handle. It
// uses optimistic concurrency on lock_version: the WHERE clause requires the
// caller's v.LockVersion to still match, returning domain.ErrStaleLockVersion
// when it does not and domain.ErrNotFound when the row does not exist at
// all. On success v.LockVersion is incremented in place.
func (r *Repository) UpdateVersion(ctx context.Context, tenantID string, v *domain.TemplateVersion) error {
	return updateVersion(ctx, r.db, tenantID, v)
}

// CreateTemplateTx inserts a new template row for the tenant using the
// caller-supplied transaction. It maps a unique-key violation on
// (tenant_id, key) to domain.ErrKeyConflict.
func (r *Repository) CreateTemplateTx(ctx context.Context, tx db.Tx, t *domain.Template) error {
	const q = `
INSERT INTO templates_template (
	id, tenant_id, doc_type_code, key, name, description,
	latest_version, published_version_id, created_by, system_owned, created_at, archived_at
) VALUES (
	$1, $2::uuid, $3, $4, $5, $6,
	$7, $8, $9, $10, $11, $12
)`
	_, err := tx.ExecContext(ctx, q,
		t.ID, t.TenantID, t.DocTypeCode, t.Key, t.Name, t.Description,
		t.LatestVersion, t.PublishedVersionID, t.CreatedBy, t.SystemOwned, t.CreatedAt, t.ArchivedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrKeyConflict
		}
		return err
	}
	return nil
}

type versionUpdateExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func updateVersion(ctx context.Context, db versionUpdateExecutor, tenantID string, v *domain.TemplateVersion) error {
	metadataJSON, placeholderJSON, err := marshalVersionSchemas(v)
	if err != nil {
		return err
	}

	const q = `
UPDATE templates_template_version
SET
	status = $2,
	docx_storage_key = $3,
	content_hash = $4,
	metadata_schema = $5,
	placeholder_schema = $6,
	pending_reviewer_role = $7,
	pending_approver_role = $8,
	reviewer_id = $9,
	approver_id = $10,
	submitted_at = $11,
	reviewed_at = $12,
	approved_at = $13,
	published_at = $14,
	obsoleted_at = $15,
	lock_version = lock_version + 1
WHERE id = $1
  AND EXISTS (
    SELECT 1 FROM templates_template t
    WHERE t.id = templates_template_version.template_id
      AND t.tenant_id = $17::uuid
  )
  AND lock_version = $16`
	res, err := db.ExecContext(ctx, q,
		v.ID, string(v.Status), v.DocxStorageKey, v.ContentHash,
		metadataJSON, placeholderJSON,
		v.PendingReviewerRole, v.PendingApproverRole, v.ReviewerID, v.ApproverID,
		v.SubmittedAt, v.ReviewedAt, v.ApprovedAt, v.PublishedAt, v.ObsoletedAt, v.LockVersion, tenantID,
	)
	if err != nil {
		return err
	}
	n, err := rowsAffected(res)
	if err != nil {
		return err
	}
	if n > 0 {
		v.LockVersion++
		return nil
	}

	var exists bool
	if err := db.QueryRowContext(ctx, `
SELECT EXISTS (
  SELECT 1
  FROM templates_template_version v
  JOIN templates_template t ON t.id = v.template_id
  WHERE v.id = $1
    AND t.tenant_id = $2::uuid
)`, v.ID, tenantID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return domain.ErrNotFound
	}
	return domain.ErrStaleLockVersion
}

// UpdateTemplateTx updates a template's mutable fields, scoped to the
// tenant, using the caller-supplied transaction. It returns
// domain.ErrNotFound when no row matches (id, tenant_id).
func (r *Repository) UpdateTemplateTx(ctx context.Context, tx db.Tx, t *domain.Template) error {
	const q = `
UPDATE templates_template
SET
	doc_type_code = $3,
	key = $4,
	name = $5,
	description = $6,
	latest_version = $7,
	published_version_id = $8,
	system_owned = $9,
	archived_at = $10,
	updated_at = now()
WHERE id = $1 AND tenant_id = $2::uuid`
	res, err := tx.ExecContext(ctx, q,
		t.ID, t.TenantID, t.DocTypeCode, t.Key, t.Name, t.Description,
		t.LatestVersion, t.PublishedVersionID, t.SystemOwned, t.ArchivedAt,
	)
	if err != nil {
		return err
	}
	n, err := rowsAffected(res)
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// UpdateVersionTx updates a template version's workflow fields, scoped to
// the tenant, using the caller-supplied transaction. Same optimistic
// concurrency semantics on lock_version as UpdateVersion.
func (r *Repository) UpdateVersionTx(ctx context.Context, tx db.Tx, tenantID string, v *domain.TemplateVersion) error {
	return updateVersion(ctx, tx, tenantID, v)
}

type draftCASExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

// UpdateVersionSchemaCAS compare-and-swaps a template version's metadata and
// placeholder schemas, executing directly against the repository's database
// handle. The UPDATE's WHERE clause requires lock_version to equal
// expectedLockVersion; on a mismatch it returns domain.ErrStaleLockVersion
// (or domain.ErrNotFound if the version does not exist), and on success it
// sets v.LockVersion to expectedLockVersion+1.
func (r *Repository) UpdateVersionSchemaCAS(ctx context.Context, tenantID string, v *domain.TemplateVersion, expectedLockVersion int) error {
	return updateVersionSchemaCAS(ctx, r.db, tenantID, v, expectedLockVersion)
}

// UpdateVersionSchemaCASTx compare-and-swaps a template version's metadata
// and placeholder schemas using the caller-supplied transaction. Same
// lock_version CAS semantics as UpdateVersionSchemaCAS.
func (r *Repository) UpdateVersionSchemaCASTx(ctx context.Context, tx db.Tx, tenantID string, v *domain.TemplateVersion, expectedLockVersion int) error {
	return updateVersionSchemaCAS(ctx, tx, tenantID, v, expectedLockVersion)
}

func updateVersionSchemaCAS(ctx context.Context, db draftCASExecutor, tenantID string, v *domain.TemplateVersion, expectedLockVersion int) error {
	metadataJSON, placeholderJSON, err := marshalVersionSchemas(v)
	if err != nil {
		return err
	}
	const q = `
UPDATE templates_template_version
SET
	metadata_schema = $3,
	placeholder_schema = $4,
	lock_version = lock_version + 1
WHERE id = $1
  AND EXISTS (
    SELECT 1 FROM templates_template t
    WHERE t.id = templates_template_version.template_id
      AND t.tenant_id = $5::uuid
  )
  AND lock_version = $2`
	res, err := db.ExecContext(ctx, q, v.ID, expectedLockVersion, metadataJSON, placeholderJSON, tenantID)
	if err != nil {
		return err
	}
	n, err := rowsAffected(res)
	if err != nil {
		return err
	}
	if n > 0 {
		v.LockVersion = expectedLockVersion + 1
		return nil
	}

	var exists bool
	if err := db.QueryRowContext(ctx, `
SELECT EXISTS (
  SELECT 1
  FROM templates_template_version v
  JOIN templates_template t ON t.id = v.template_id
  WHERE v.id = $1
    AND t.tenant_id = $2::uuid
)`, v.ID, tenantID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return domain.ErrNotFound
	}
	return domain.ErrStaleLockVersion
}

// ObsoletePreviousPublished marks every other published version of
// templateID as obsolete (keeping keepVersionID untouched), executing
// directly against the repository's database handle. The UPDATE joins back
// to templates_template on tenant_id so it cannot cross tenant boundaries
// even if upstream validation is bypassed (F-DB5).
func (r *Repository) ObsoletePreviousPublished(ctx context.Context, tenantID, templateID, keepVersionID string) error {
	// Explicit tenant predicate via JOIN to the parent templates_template row so
	// the UPDATE cannot cross tenant boundaries even if upstream validation is
	// bypassed.  This closes the concrete gap identified in F-DB5: previously the
	// WHERE clause filtered only by template_id+status with no tenant guard.
	const q = `
UPDATE templates_template_version v
SET status = 'obsolete', obsoleted_at = now()
FROM templates_template t
WHERE v.template_id = $1
  AND v.status = 'published'
  AND v.id <> $2
  AND t.id = v.template_id
  AND t.tenant_id = $3::uuid`
	res, err := r.db.ExecContext(ctx, q, templateID, keepVersionID, tenantID)
	if err != nil {
		return fmt.Errorf("templates repository obsolete previous published: %w", err)
	}
	if _, err := rowsAffected(res); err != nil {
		return fmt.Errorf("templates repository obsolete previous published: %w", err)
	}
	return nil
}

// ObsoletePreviousPublishedTx marks every other published version of
// templateID as obsolete (keeping keepVersionID untouched), using the
// caller-supplied transaction. This is the primary production path, called
// inside the publish/approve transaction; same tenant guard as
// ObsoletePreviousPublished (F-DB5).
func (r *Repository) ObsoletePreviousPublishedTx(ctx context.Context, tx db.Tx, tenantID, templateID, keepVersionID string) error {
	// Same tenant guard as ObsoletePreviousPublished — applied here because this
	// path is the primary production path (called inside the publish/approve tx).
	const q = `
UPDATE templates_template_version v
SET status = 'obsolete', obsoleted_at = now()
FROM templates_template t
WHERE v.template_id = $1
  AND v.status = 'published'
  AND v.id <> $2
  AND t.id = v.template_id
  AND t.tenant_id = $3::uuid`
	res, err := tx.ExecContext(ctx, q, templateID, keepVersionID, tenantID)
	if err != nil {
		return fmt.Errorf("templates repository obsolete previous published tx: %w", err)
	}
	if _, err := rowsAffected(res); err != nil {
		return fmt.Errorf("templates repository obsolete previous published tx: %w", err)
	}
	return nil
}

// GetApprovalConfig loads the reviewer/approver role configuration for a
// template, scoped to the tenant via a subquery on templates_template. It
// returns domain.ErrNotFound when no config row exists or templateID is not
// a valid UUID.
func (r *Repository) GetApprovalConfig(ctx context.Context, tenantID, templateID string) (*domain.ApprovalConfig, error) {
	const q = `
SELECT template_id::text, reviewer_role, approver_role
FROM templates_approval_config
WHERE template_id = $1
  AND template_id IN (SELECT id FROM templates_template WHERE tenant_id = $2::uuid)`
	var (
		cfg      domain.ApprovalConfig
		reviewer sql.NullString
	)
	err := r.db.QueryRowContext(ctx, q, templateID, tenantID).Scan(&cfg.TemplateID, &reviewer, &cfg.ApproverRole)
	if errors.Is(err, sql.ErrNoRows) || isInvalidUUID(err) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if reviewer.Valid {
		cfg.ReviewerRole = &reviewer.String
	}
	return &cfg, nil
}

// UpsertApprovalConfig inserts or updates a template's reviewer/approver
// role configuration, executing directly against the repository's database
// handle (ON CONFLICT (template_id) DO UPDATE).
func (r *Repository) UpsertApprovalConfig(ctx context.Context, c *domain.ApprovalConfig) error {
	const q = `
INSERT INTO templates_approval_config (template_id, reviewer_role, approver_role)
VALUES ($1, $2, $3)
ON CONFLICT (template_id) DO UPDATE
SET reviewer_role = EXCLUDED.reviewer_role,
    approver_role = EXCLUDED.approver_role`
	res, err := r.db.ExecContext(ctx, q, c.TemplateID, c.ReviewerRole, c.ApproverRole)
	if err != nil {
		return fmt.Errorf("templates repository upsert approval config: %w", err)
	}
	if _, err := rowsAffected(res); err != nil {
		return fmt.Errorf("templates repository upsert approval config: %w", err)
	}
	return nil
}

// UpsertApprovalConfigTx inserts or updates a template's reviewer/approver
// role configuration using the caller-supplied transaction (ON CONFLICT
// (template_id) DO UPDATE).
func (r *Repository) UpsertApprovalConfigTx(ctx context.Context, tx db.Tx, c *domain.ApprovalConfig) error {
	const q = `
INSERT INTO templates_approval_config (template_id, reviewer_role, approver_role)
VALUES ($1, $2, $3)
ON CONFLICT (template_id) DO UPDATE
SET reviewer_role = EXCLUDED.reviewer_role,
    approver_role = EXCLUDED.approver_role`
	res, err := tx.ExecContext(ctx, q, c.TemplateID, c.ReviewerRole, c.ApproverRole)
	if err != nil {
		return fmt.Errorf("templates repository upsert approval config tx: %w", err)
	}
	if _, err := rowsAffected(res); err != nil {
		return fmt.Errorf("templates repository upsert approval config tx: %w", err)
	}
	return nil
}

// AppendAudit records a template audit event via the repository's attached
// audit.Writer, executing directly (not inside a caller transaction). It is
// a no-op returning nil when no writer was attached via WithAudit.
func (r *Repository) AppendAudit(ctx context.Context, entry *domain.AuditEvent) error {
	if r.audit == nil {
		return nil
	}
	event, err := r.auditEvent(ctx, entry)
	if err != nil {
		return err
	}
	return r.audit.Record(ctx, event)
}

// AppendAuditTx records a template audit event via the repository's
// attached audit.Writer, using the caller-supplied transaction so the audit
// row commits atomically with the rest of the transaction's writes. It is a
// no-op returning nil when no writer was attached via WithAudit.
func (r *Repository) AppendAuditTx(ctx context.Context, tx db.Tx, entry *domain.AuditEvent) error {
	if r.audit == nil {
		return nil
	}
	event, err := r.auditEvent(ctx, entry)
	if err != nil {
		return err
	}
	return r.audit.RecordTx(ctx, tx, event)
}

func (r *Repository) auditEvent(ctx context.Context, entry *domain.AuditEvent) (auditdomain.Event, error) {
	// audit_events has no version_id column; carry it inside the payload so
	// ListAudit (which reads audit_events since Wave 1.8, F-07-sub-split)
	// can surface it. Copy before mutating — Details belongs to the caller.
	details := entry.Details
	if entry.VersionID != nil && *entry.VersionID != "" {
		details = make(map[string]any, len(entry.Details)+1)
		for k, v := range entry.Details {
			details[k] = v
		}
		details["version_id"] = *entry.VersionID
	}
	payload, err := json.Marshal(details)
	if err != nil {
		return auditdomain.Event{}, fmt.Errorf("templates audit marshal details: %w", err)
	}
	tenantID := entry.TenantID
	if tenantID == "" {
		if tid, err := tenant.FromContext(ctx); err == nil {
			tenantID = tid
		}
	}
	return auditdomain.Event{
		ID:           uuid.NewString(),
		OccurredAt:   time.Now().UTC(),
		ActorID:      entry.ActorID,
		Action:       string(entry.Action),
		ResourceType: "template",
		ResourceID:   entry.TemplateID,
		PayloadJSON:  string(payload),
		TenantID:     tenantID,
	}, nil
}

// ListAudit reads template audit history from the canonical audit_events
// sink, where AppendAudit[Tx] writes (F-07-sub-split, Wave 1.8). Events
// recorded in the legacy templates_audit_log before the write-sink migration
// are an accepted historical seam — they are not surfaced here.
func (r *Repository) ListAudit(ctx context.Context, tenantID, templateID string, limit, offset int) ([]*domain.AuditEvent, error) {
	const q = `
SELECT tenant_id::text, resource_id, actor_id, action, payload, occurred_at
FROM metaldocs.audit_events
WHERE resource_type = 'template'
  AND resource_id = $1
  AND tenant_id = $2::uuid
ORDER BY occurred_at DESC
LIMIT $3 OFFSET $4`
	rows, err := r.db.QueryContext(ctx, q, templateID, tenantID, limit, offset)
	if isInvalidUUID(err) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]*domain.AuditEvent, 0)
	for rows.Next() {
		var (
			event   domain.AuditEvent
			details []byte
		)
		if err := rows.Scan(&event.TenantID, &event.TemplateID, &event.ActorID, &event.Action, &details, &event.OccurredAt); err != nil {
			return nil, err
		}
		event.Details = map[string]any{}
		if len(details) > 0 {
			if err := unmarshalAuditDetails(details, &event.Details); err != nil {
				return nil, err
			}
		}
		// version_id travels inside the payload (audit_events has no
		// dedicated column); lift it back onto the domain field.
		if v, ok := event.Details["version_id"].(string); ok && v != "" {
			event.VersionID = &v
			delete(event.Details, "version_id")
		}
		out = append(out, &event)
	}
	return out, rows.Err()
}

// TenantIDsWithTemplates returns the distinct tenant_ids that own at least one
// template. The orphan-object GC sweeper iterates these to bound its per-tenant
// object-store enumeration to tenants that actually have template objects.
func (r *Repository) TenantIDsWithTemplates(ctx context.Context) ([]string, error) {
	const q = `SELECT DISTINCT tenant_id::text FROM templates_template`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("templates repository tenant ids with templates: %w", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("templates repository tenant ids with templates scan: %w", err)
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// ReferencedTemplateObjectRef is one template-version row's identity plus its
// stored docx object key, as needed to compute the set of object keys a tenant's
// templates reference. The schema object key is NOT a column — it is derived from
// (TemplateID, VersionNumber) by the application layer — so the coordinates are
// returned alongside the stored docx key.
type ReferencedTemplateObjectRef struct {
	TemplateID     string
	VersionNumber  int
	DocxStorageKey string
}

// ReferencedTemplateObjectRefs returns every template-version row for tenantID,
// carrying the stored docx_storage_key plus the (template_id, version_number)
// coordinates the caller needs to derive the version's schema key. This is the
// DB source of truth for "which template objects are referenced": the orphan-GC
// sweeper unions the docx keys with the derived schema keys to form the complete
// referenced set, then deletes only objects absent from it.
func (r *Repository) ReferencedTemplateObjectRefs(ctx context.Context, tenantID string) ([]ReferencedTemplateObjectRef, error) {
	const q = `
SELECT v.template_id::text, v.version_number, v.docx_storage_key
FROM templates_template_version v
JOIN templates_template t ON t.id = v.template_id
WHERE t.tenant_id = $1::uuid`
	rows, err := r.db.QueryContext(ctx, q, tenantID)
	if isInvalidUUID(err) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("templates repository referenced template object refs: %w", err)
	}
	defer rows.Close()
	var out []ReferencedTemplateObjectRef
	for rows.Next() {
		var ref ReferencedTemplateObjectRef
		if err := rows.Scan(&ref.TemplateID, &ref.VersionNumber, &ref.DocxStorageKey); err != nil {
			return nil, fmt.Errorf("templates repository referenced template object refs scan: %w", err)
		}
		out = append(out, ref)
	}
	return out, rows.Err()
}
