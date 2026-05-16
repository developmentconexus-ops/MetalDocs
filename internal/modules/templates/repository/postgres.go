package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	auditdomain "metaldocs/internal/modules/audit/domain"
	"metaldocs/internal/modules/templates/application"
	"metaldocs/internal/modules/templates/domain"
	tenant "metaldocs/internal/platform/tenant"
)

// isInvalidUUID returns true when err is a Postgres error with SQLSTATE 22P02
// (invalid text representation). We map malformed UUID lookups to not found.
func isInvalidUUID(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "22P02"
}

type Repository struct {
	db    *sql.DB
	audit auditdomain.Writer
}

func New(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) WithAudit(w auditdomain.Writer) *Repository {
	r.audit = w
	return r
}

var _ application.Repository = (*Repository)(nil)

func (r *Repository) CreateTemplate(ctx context.Context, t *domain.Template) error {
	const q = `
INSERT INTO templates_template (
	id, tenant_id, doc_type_code, key, name, description, areas, visibility,
	specific_areas, latest_version, published_version_id, created_by, system_owned, created_at, archived_at
) VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8,
	$9, $10, $11, $12, $13, $14, $15
)`
	_, err := r.db.ExecContext(ctx, q,
		t.ID, t.TenantID, t.DocTypeCode, t.Key, t.Name, t.Description, t.Areas, string(t.Visibility),
		t.SpecificAreas, t.LatestVersion, t.PublishedVersionID, t.CreatedBy, t.SystemOwned, t.CreatedAt, t.ArchivedAt,
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

func (r *Repository) GetTemplate(ctx context.Context, tenantID, id string) (*domain.Template, error) {
	const q = `
SELECT
	id::text, tenant_id, doc_type_code, key, name, description, array_to_json(areas)::text, visibility, array_to_json(specific_areas)::text,
	latest_version, published_version_id::text, created_by, system_owned, created_at, archived_at
FROM templates_template
WHERE id = $1 AND tenant_id = $2`

	t, err := scanTemplate(r.db.QueryRowContext(ctx, q, id, tenantID))
	if errors.Is(err, sql.ErrNoRows) || isInvalidUUID(err) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *Repository) GetTemplateByKey(ctx context.Context, tenantID, key string) (*domain.Template, error) {
	const q = `
SELECT
	id::text, tenant_id, doc_type_code, key, name, description, array_to_json(areas)::text, visibility, array_to_json(specific_areas)::text,
	latest_version, published_version_id::text, created_by, system_owned, created_at, archived_at
FROM templates_template
WHERE tenant_id = $1 AND key = $2`

	t, err := scanTemplate(r.db.QueryRowContext(ctx, q, tenantID, key))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *Repository) ListTemplates(ctx context.Context, f application.ListFilter) ([]*domain.Template, error) {
	const q = `
SELECT
	id::text, tenant_id, doc_type_code, key, name, description, array_to_json(areas)::text, visibility, array_to_json(specific_areas)::text,
	latest_version, published_version_id::text, created_by, system_owned, created_at, archived_at
FROM templates_template
WHERE tenant_id = $1
  AND system_owned = false
  AND ($2::text IS NULL OR doc_type_code = $2)
  AND (cardinality($3::text[]) = 0 OR areas && $3::text[])
  AND (
    visibility = 'public'
    OR (visibility = 'internal' AND NOT $6::boolean)
    OR (visibility = 'specific' AND cardinality($7::text[]) > 0 AND specific_areas && $7::text[])
  )
ORDER BY created_at DESC
LIMIT $4 OFFSET $5`

	rows, err := r.db.QueryContext(
		ctx,
		q,
		f.TenantID,
		f.DocTypeCode,
		normalizedTextArray(f.AreaAny),
		f.Limit,
		f.Offset,
		f.IsExternalViewer,
		normalizedTextArray(f.ActorAreas),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]*domain.Template, 0)
	for rows.Next() {
		t, scanErr := scanTemplate(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *Repository) UpdateTemplate(ctx context.Context, t *domain.Template) error {
	const q = `
UPDATE templates_template
SET
	doc_type_code = $3,
	key = $4,
	name = $5,
	description = $6,
	areas = $7,
	visibility = $8,
	specific_areas = $9,
	latest_version = $10,
	published_version_id = $11,
	system_owned = $12,
	archived_at = $13
WHERE id = $1 AND tenant_id = $2`

	res, err := r.db.ExecContext(ctx, q,
		t.ID, t.TenantID, t.DocTypeCode, t.Key, t.Name, t.Description,
		t.Areas, string(t.Visibility), t.SpecificAreas, t.LatestVersion, t.PublishedVersionID, t.SystemOwned, t.ArchivedAt,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *Repository) CreateVersion(ctx context.Context, v *domain.TemplateVersion) error {
	metadataJSON, placeholderJSON, err := marshalVersionSchemas(v)
	if err != nil {
		return err
	}

	const q = `
INSERT INTO templates_template_version (
	id, template_id, version_number, status, docx_storage_key, content_hash,
	metadata_schema, placeholder_schema, author_id,
	pending_reviewer_role, pending_approver_role, reviewer_id, approver_id,
	submitted_at, reviewed_at, approved_at, published_at, obsoleted_at, lock_version, created_at
) VALUES (
	$1, $2, $3, $4, $5, $6,
	$7, $8, $9,
	$10, $11, $12, $13,
	$14, $15, $16, $17, $18, $19, $20
)`
	_, err = r.db.ExecContext(ctx, q,
		v.ID, v.TemplateID, v.VersionNumber, string(v.Status), v.DocxStorageKey, v.ContentHash,
		metadataJSON, placeholderJSON, v.AuthorID,
		v.PendingReviewerRole, v.PendingApproverRole, v.ReviewerID, v.ApproverID,
		v.SubmittedAt, v.ReviewedAt, v.ApprovedAt, v.PublishedAt, v.ObsoletedAt, v.LockVersion, v.CreatedAt,
	)
	return err
}

func (r *Repository) CreateVersionTx(ctx context.Context, tx *sql.Tx, v *domain.TemplateVersion) error {
	metadataJSON, placeholderJSON, err := marshalVersionSchemas(v)
	if err != nil {
		return err
	}
	const q = `
INSERT INTO templates_template_version (
	id, template_id, version_number, status, docx_storage_key, content_hash,
	metadata_schema, placeholder_schema, author_id,
	pending_reviewer_role, pending_approver_role, reviewer_id, approver_id,
	submitted_at, reviewed_at, approved_at, published_at, obsoleted_at, lock_version, created_at
) VALUES (
	$1, $2, $3, $4, $5, $6,
	$7, $8, $9,
	$10, $11, $12, $13,
	$14, $15, $16, $17, $18, $19, $20
)`
	_, err = tx.ExecContext(ctx, q,
		v.ID, v.TemplateID, v.VersionNumber, string(v.Status), v.DocxStorageKey, v.ContentHash,
		metadataJSON, placeholderJSON, v.AuthorID,
		v.PendingReviewerRole, v.PendingApproverRole, v.ReviewerID, v.ApproverID,
		v.SubmittedAt, v.ReviewedAt, v.ApprovedAt, v.PublishedAt, v.ObsoletedAt, v.LockVersion, v.CreatedAt,
	)
	return err
}

func (r *Repository) GetVersion(ctx context.Context, templateID string, n int) (*domain.TemplateVersion, error) {
	const q = `
SELECT
	id::text, template_id::text, version_number, status, docx_storage_key, content_hash,
	metadata_schema, placeholder_schema, author_id,
	pending_reviewer_role, pending_approver_role, reviewer_id, approver_id,
	submitted_at, reviewed_at, approved_at, published_at, obsoleted_at, lock_version, created_at
FROM templates_template_version
WHERE template_id = $1 AND version_number = $2`

	v, err := scanTemplateVersion(r.db.QueryRowContext(ctx, q, templateID, n))
	if errors.Is(err, sql.ErrNoRows) || isInvalidUUID(err) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (r *Repository) GetVersionByID(ctx context.Context, id string) (*domain.TemplateVersion, error) {
	const q = `
SELECT
	id::text, template_id::text, version_number, status, docx_storage_key, content_hash,
	metadata_schema, placeholder_schema, author_id,
	pending_reviewer_role, pending_approver_role, reviewer_id, approver_id,
	submitted_at, reviewed_at, approved_at, published_at, obsoleted_at, lock_version, created_at
FROM templates_template_version
WHERE id = $1`

	v, err := scanTemplateVersion(r.db.QueryRowContext(ctx, q, id))
	if errors.Is(err, sql.ErrNoRows) || isInvalidUUID(err) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (r *Repository) UpdateVersion(ctx context.Context, v *domain.TemplateVersion) error {
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
	lock_version = $16
WHERE id = $1`
	res, err := r.db.ExecContext(ctx, q,
		v.ID, string(v.Status), v.DocxStorageKey, v.ContentHash,
		metadataJSON, placeholderJSON,
		v.PendingReviewerRole, v.PendingApproverRole, v.ReviewerID, v.ApproverID,
		v.SubmittedAt, v.ReviewedAt, v.ApprovedAt, v.PublishedAt, v.ObsoletedAt, v.LockVersion,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *Repository) CreateTemplateTx(ctx context.Context, tx *sql.Tx, t *domain.Template) error {
	const q = `
INSERT INTO templates_template (
	id, tenant_id, doc_type_code, key, name, description, areas, visibility,
	specific_areas, latest_version, published_version_id, created_by, system_owned, created_at, archived_at
) VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8,
	$9, $10, $11, $12, $13, $14, $15
)`
	_, err := tx.ExecContext(ctx, q,
		t.ID, t.TenantID, t.DocTypeCode, t.Key, t.Name, t.Description, t.Areas, string(t.Visibility),
		t.SpecificAreas, t.LatestVersion, t.PublishedVersionID, t.CreatedBy, t.SystemOwned, t.CreatedAt, t.ArchivedAt,
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

func (r *Repository) UpdateTemplateTx(ctx context.Context, tx *sql.Tx, t *domain.Template) error {
	const q = `
UPDATE templates_template
SET
	doc_type_code = $3,
	key = $4,
	name = $5,
	description = $6,
	areas = $7,
	visibility = $8,
	specific_areas = $9,
	latest_version = $10,
	published_version_id = $11,
	system_owned = $12,
	archived_at = $13
WHERE id = $1 AND tenant_id = $2`
	res, err := tx.ExecContext(ctx, q,
		t.ID, t.TenantID, t.DocTypeCode, t.Key, t.Name, t.Description,
		t.Areas, string(t.Visibility), t.SpecificAreas, t.LatestVersion, t.PublishedVersionID, t.SystemOwned, t.ArchivedAt,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *Repository) UpdateVersionTx(ctx context.Context, tx *sql.Tx, v *domain.TemplateVersion) error {
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
	lock_version = $16
WHERE id = $1`
	res, err := tx.ExecContext(ctx, q,
		v.ID, string(v.Status), v.DocxStorageKey, v.ContentHash,
		metadataJSON, placeholderJSON,
		v.PendingReviewerRole, v.PendingApproverRole, v.ReviewerID, v.ApproverID,
		v.SubmittedAt, v.ReviewedAt, v.ApprovedAt, v.PublishedAt, v.ObsoletedAt, v.LockVersion,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *Repository) UpdateVersionDraftCAS(ctx context.Context, versionID string, expectedLockVersion int, docxStorageKey, docxContentHash string) error {
	const q = `
UPDATE templates_template_version
SET
	docx_storage_key = $3,
	content_hash = $4,
	lock_version = lock_version + 1
WHERE id = $1
  AND lock_version = $2`
	res, err := r.db.ExecContext(ctx, q, versionID, expectedLockVersion, docxStorageKey, docxContentHash)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n > 0 {
		return nil
	}

	var exists bool
	if err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM templates_template_version WHERE id = $1)", versionID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return domain.ErrNotFound
	}
	return domain.ErrStaleLockVersion
}

func (r *Repository) ObsoletePreviousPublished(ctx context.Context, templateID, keepVersionID string) error {
	const q = `
UPDATE templates_template_version
SET status = 'obsolete', obsoleted_at = now()
WHERE template_id = $1 AND status = 'published' AND id <> $2`
	_, err := r.db.ExecContext(ctx, q, templateID, keepVersionID)
	return err
}

func (r *Repository) ObsoletePreviousPublishedTx(ctx context.Context, tx *sql.Tx, templateID, keepVersionID string) error {
	const q = `
UPDATE templates_template_version
SET status = 'obsolete', obsoleted_at = now()
WHERE template_id = $1 AND status = 'published' AND id <> $2`
	_, err := tx.ExecContext(ctx, q, templateID, keepVersionID)
	return err
}

func (r *Repository) GetApprovalConfig(ctx context.Context, templateID string) (*domain.ApprovalConfig, error) {
	const q = `
SELECT template_id::text, reviewer_role, approver_role
FROM templates_approval_config
WHERE template_id = $1`
	var (
		cfg      domain.ApprovalConfig
		reviewer sql.NullString
	)
	err := r.db.QueryRowContext(ctx, q, templateID).Scan(&cfg.TemplateID, &reviewer, &cfg.ApproverRole)
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

func (r *Repository) UpsertApprovalConfig(ctx context.Context, c *domain.ApprovalConfig) error {
	const q = `
INSERT INTO templates_approval_config (template_id, reviewer_role, approver_role)
VALUES ($1, $2, $3)
ON CONFLICT (template_id) DO UPDATE
SET reviewer_role = EXCLUDED.reviewer_role,
    approver_role = EXCLUDED.approver_role`
	_, err := r.db.ExecContext(ctx, q, c.TemplateID, c.ReviewerRole, c.ApproverRole)
	return err
}

func (r *Repository) UpsertApprovalConfigTx(ctx context.Context, tx *sql.Tx, c *domain.ApprovalConfig) error {
	const q = `
INSERT INTO templates_approval_config (template_id, reviewer_role, approver_role)
VALUES ($1, $2, $3)
ON CONFLICT (template_id) DO UPDATE
SET reviewer_role = EXCLUDED.reviewer_role,
    approver_role = EXCLUDED.approver_role`
	_, err := tx.ExecContext(ctx, q, c.TemplateID, c.ReviewerRole, c.ApproverRole)
	return err
}

func (r *Repository) AppendAudit(ctx context.Context, entry *domain.AuditEvent) error {
	if r.audit == nil {
		return nil
	}
	payload, _ := json.Marshal(entry.Details)
	tenantID := entry.TenantID
	if tenantID == "" {
		if tid, err := tenant.FromContext(ctx); err == nil {
			tenantID = tid
		}
	}
	return r.audit.Record(ctx, auditdomain.Event{
		ID:           uuid.NewString(),
		OccurredAt:   time.Now().UTC(),
		ActorID:      entry.ActorID,
		Action:       string(entry.Action),
		ResourceType: "template",
		ResourceID:   entry.TemplateID,
		PayloadJSON:  string(payload),
		TenantID:     tenantID,
	})
}

func (r *Repository) AppendAuditTx(ctx context.Context, _ *sql.Tx, entry *domain.AuditEvent) error {
	return r.AppendAudit(ctx, entry)
}

func (r *Repository) ListAudit(ctx context.Context, templateID string, limit, offset int) ([]*domain.AuditEvent, error) {
	const q = `
SELECT tenant_id, template_id::text, version_id::text, actor_id, action, details, occurred_at
FROM templates_audit_log
WHERE template_id = $1
ORDER BY occurred_at DESC
LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx, q, templateID, limit, offset)
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
			event     domain.AuditEvent
			versionID sql.NullString
			details   []byte
		)
		if err := rows.Scan(&event.TenantID, &event.TemplateID, &versionID, &event.ActorID, &event.Action, &details, &event.OccurredAt); err != nil {
			return nil, err
		}
		if versionID.Valid {
			event.VersionID = &versionID.String
		}
		event.Details = map[string]any{}
		if len(details) > 0 {
			if err := unmarshalAuditDetails(details, &event.Details); err != nil {
				return nil, err
			}
		}
		out = append(out, &event)
	}
	return out, rows.Err()
}
