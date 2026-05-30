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

func wrapRepositoryErr(op string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, domain.ErrNotFound) ||
		errors.Is(err, domain.ErrKeyConflict) ||
		errors.Is(err, domain.ErrStaleLockVersion) {
		return err
	}
	return fmt.Errorf("templates repository %s: %w", op, err)
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
	$1, $2::uuid, $3, $4, $5, $6, '{}'::text[], 'public',
	'{}'::text[], $7, $8, $9, $10, $11, $12
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

func (r *Repository) GetTemplate(ctx context.Context, tenantID, id string) (*domain.Template, error) {
	const q = `
SELECT
	t.id::text, t.tenant_id::text, t.doc_type_code, t.key, t.name, t.description,
	t.latest_version, t.published_version_id::text, pv.version_number, pv.revision_number,
	t.created_by, t.system_owned, t.created_at, t.archived_at
FROM templates_template t
LEFT JOIN templates_template_version pv ON pv.id = t.published_version_id
WHERE t.id = $1 AND t.tenant_id = $2::uuid`

	tmpl, err := scanTemplate(r.db.QueryRowContext(ctx, q, id, tenantID))
	if errors.Is(err, sql.ErrNoRows) || isInvalidUUID(err) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

func (r *Repository) GetTemplateByKey(ctx context.Context, tenantID, key string) (*domain.Template, error) {
	const q = `
SELECT
	t.id::text, t.tenant_id::text, t.doc_type_code, t.key, t.name, t.description,
	t.latest_version, t.published_version_id::text, pv.version_number, pv.revision_number,
	t.created_by, t.system_owned, t.created_at, t.archived_at
FROM templates_template t
LEFT JOIN templates_template_version pv ON pv.id = t.published_version_id
WHERE t.tenant_id = $1::uuid AND t.key = $2`

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
	t.id::text, t.tenant_id::text, t.doc_type_code, t.key, t.name, t.description,
	t.latest_version, t.published_version_id::text, pv.version_number, pv.revision_number,
	t.created_by, t.system_owned, t.created_at, t.archived_at
FROM templates_template t
LEFT JOIN templates_template_version pv ON pv.id = t.published_version_id
WHERE t.tenant_id = $1::uuid
  AND t.system_owned = false
  AND ($2::text IS NULL OR t.doc_type_code = $2)
ORDER BY t.created_at DESC
LIMIT $3 OFFSET $4`

	rows, err := r.db.QueryContext(
		ctx,
		q,
		f.TenantID,
		f.DocTypeCode,
		f.Limit,
		f.Offset,
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
	latest_version = $7,
	published_version_id = $8,
	system_owned = $9,
	archived_at = $10
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
	id, template_id, version_number, revision_number, status, docx_storage_key, content_hash,
	metadata_schema, placeholder_schema, author_id,
	pending_reviewer_role, pending_approver_role, reviewer_id, approver_id,
	submitted_at, reviewed_at, approved_at, published_at, obsoleted_at, lock_version, created_at
) VALUES (
	$1, $2, $3,
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
	)
	if err = row.Scan(&v.RevisionNumber); err != nil {
		return fmt.Errorf("templates repository create version: %w", err)
	}
	return nil
}

func (r *Repository) CreateVersionTx(ctx context.Context, tx *sql.Tx, v *domain.TemplateVersion) error {
	metadataJSON, placeholderJSON, err := marshalVersionSchemas(v)
	if err != nil {
		return err
	}
	// ADR 0013: see CreateVersion comment.
	const q = `
INSERT INTO templates_template_version (
	id, template_id, version_number, revision_number, status, docx_storage_key, content_hash,
	metadata_schema, placeholder_schema, author_id,
	pending_reviewer_role, pending_approver_role, reviewer_id, approver_id,
	submitted_at, reviewed_at, approved_at, published_at, obsoleted_at, lock_version, created_at
) VALUES (
	$1, $2, $3,
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
	)
	if err := row.Scan(&v.RevisionNumber); err != nil {
		return fmt.Errorf("templates repository create version tx: %w", err)
	}
	return nil
}

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

func (r *Repository) UpdateVersion(ctx context.Context, tenantID string, v *domain.TemplateVersion) error {
	return updateVersion(ctx, r.db, tenantID, v)
}

func (r *Repository) CreateTemplateTx(ctx context.Context, tx *sql.Tx, t *domain.Template) error {
	const q = `
INSERT INTO templates_template (
	id, tenant_id, doc_type_code, key, name, description, areas, visibility,
	specific_areas, latest_version, published_version_id, created_by, system_owned, created_at, archived_at
) VALUES (
	$1, $2::uuid, $3, $4, $5, $6, '{}'::text[], 'public',
	'{}'::text[], $7, $8, $9, $10, $11, $12
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

func (r *Repository) UpdateTemplateTx(ctx context.Context, tx *sql.Tx, t *domain.Template) error {
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
	archived_at = $10
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

func (r *Repository) UpdateVersionTx(ctx context.Context, tx *sql.Tx, tenantID string, v *domain.TemplateVersion) error {
	return updateVersion(ctx, tx, tenantID, v)
}

func (r *Repository) UpdateVersionDraftCAS(ctx context.Context, tenantID, versionID string, expectedLockVersion int, docxStorageKey, docxContentHash string) error {
	return updateVersionDraftCAS(ctx, r.db, tenantID, versionID, expectedLockVersion, docxStorageKey, docxContentHash)
}

func (r *Repository) UpdateVersionDraftCASTx(ctx context.Context, tx *sql.Tx, tenantID, versionID string, expectedLockVersion int, docxStorageKey, docxContentHash string) error {
	return updateVersionDraftCAS(ctx, tx, tenantID, versionID, expectedLockVersion, docxStorageKey, docxContentHash)
}

type draftCASExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func updateVersionDraftCAS(ctx context.Context, db draftCASExecutor, tenantID, versionID string, expectedLockVersion int, docxStorageKey, docxContentHash string) error {
	const q = `
UPDATE templates_template_version
SET
	docx_storage_key = $3,
	content_hash = $4,
	lock_version = lock_version + 1
WHERE id = $1
  AND EXISTS (
    SELECT 1 FROM templates_template t
    WHERE t.id = templates_template_version.template_id
      AND t.tenant_id = $5::uuid
  )
  AND lock_version = $2`
	res, err := db.ExecContext(ctx, q, versionID, expectedLockVersion, docxStorageKey, docxContentHash, tenantID)
	if err != nil {
		return err
	}
	n, err := rowsAffected(res)
	if err != nil {
		return err
	}
	if n > 0 {
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
)`, versionID, tenantID).Scan(&exists); err != nil {
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
	res, err := r.db.ExecContext(ctx, q, templateID, keepVersionID)
	if err != nil {
		return fmt.Errorf("templates repository obsolete previous published: %w", err)
	}
	if _, err := rowsAffected(res); err != nil {
		return fmt.Errorf("templates repository obsolete previous published: %w", err)
	}
	return nil
}

func (r *Repository) ObsoletePreviousPublishedTx(ctx context.Context, tx *sql.Tx, templateID, keepVersionID string) error {
	const q = `
UPDATE templates_template_version
SET status = 'obsolete', obsoleted_at = now()
WHERE template_id = $1 AND status = 'published' AND id <> $2`
	res, err := tx.ExecContext(ctx, q, templateID, keepVersionID)
	if err != nil {
		return fmt.Errorf("templates repository obsolete previous published tx: %w", err)
	}
	if _, err := rowsAffected(res); err != nil {
		return fmt.Errorf("templates repository obsolete previous published tx: %w", err)
	}
	return nil
}

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

func (r *Repository) UpsertApprovalConfigTx(ctx context.Context, tx *sql.Tx, c *domain.ApprovalConfig) error {
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

func (r *Repository) AppendAuditTx(ctx context.Context, tx *sql.Tx, entry *domain.AuditEvent) error {
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
	payload, err := json.Marshal(entry.Details)
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

func (r *Repository) ListAudit(ctx context.Context, tenantID, templateID string, limit, offset int) ([]*domain.AuditEvent, error) {
	const q = `
SELECT tenant_id::text, template_id::text, version_id::text, actor_id, action, details, occurred_at
FROM templates_audit_log
WHERE template_id = $1
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
