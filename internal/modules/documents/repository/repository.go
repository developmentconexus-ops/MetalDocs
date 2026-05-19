package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"

	"metaldocs/internal/modules/documents/domain"
	templatesdomain "metaldocs/internal/modules/templates/domain"

	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
)

// isInvalidUUID returns true when err is a Postgres error with SQLSTATE 22P02
// (invalid text representation) — typically raised when a UUID column receives
// a malformed string. We treat this as "row doesn't exist" rather than an
// internal server error.
func isInvalidUUID(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "22P02"
}

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository { return &Repository{db: db} }

// CreateDocument inserts document + initial session + initial revision in one
// CreateDocument is the legacy (non-tx) wrapper. Only DuplicateDocument uses it.
// Atomic flow uses CreateDocumentTx directly with a caller-owned tx.
func (r *Repository) CreateDocument(ctx context.Context, d *domain.Document, initialContentHash string, requiredPlaceholders []templatesdomain.Placeholder) (docID, revID, sessionID string, err error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return "", "", "", err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	docID, revID, sessionID, err = r.CreateDocumentTx(ctx, tx, d, initialContentHash, "", requiredPlaceholders)
	if err != nil {
		return "", "", "", err
	}
	if err = tx.Commit(); err != nil {
		return "", "", "", err
	}
	return docID, revID, sessionID, nil
}

// CreateDocumentTx performs the DB-only portion of CreateDocument inside the
// caller-owned tx: inserts document + initial editor_session + initial
// revision, seeds template snapshot columns and required placeholder rows.
//
// initialStorageKey:
//   - "" for the legacy CreateDocument flow, where the caller must finalize
//     storage_key after S3 AdoptTempObject (the rendered docx key is content-
//     addressed and only known post-render).
//   - the template's published docx key for the registry atomic-create flow,
//     which uses template-passthrough — no render side-effect, safe to write
//     atomically with the document/revision insert so the editor opens
//     immediately on first GET.
//
// It does NOT touch S3. Callers that need S3 finalization (overwriting
// storage_key with a rendered key) must do so after tx.Commit() via
// SetRevisionStorageKey.
func (r *Repository) CreateDocumentTx(ctx context.Context, tx *sql.Tx, d *domain.Document, initialContentHash, initialStorageKey string, requiredPlaceholders []templatesdomain.Placeholder) (docID, revID, sessionID string, err error) {
	// Serialise revision_number allocation per (tenant, controlled_document).
	// pg_advisory_xact_lock auto-releases at COMMIT/ROLLBACK.
	if d.ControlledDocumentID != nil && *d.ControlledDocumentID != "" {
		lockKey := d.TenantID + ":" + *d.ControlledDocumentID
		if _, err := tx.ExecContext(ctx,
			`SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, lockKey,
		); err != nil {
			return "", "", "", fmt.Errorf("acquire revision lock: %w", err)
		}
	}

	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentCreate), "tenant"); err != nil {
		return "", "", "", fmt.Errorf("create document: authz check: %w", err)
	}

	// Deferrable FKs allow inserting doc -> session -> revision in any order in tx.
	// COALESCE handles hypothetical NULL CD (schema enforces NOT NULL since migration 0129).

	var createdByDisplayName sql.NullString
	if err := tx.QueryRowContext(ctx, `SELECT display_name FROM metaldocs.iam_users WHERE user_id = $1`, d.CreatedBy).Scan(&createdByDisplayName); err != nil && err != sql.ErrNoRows {
		return "", "", "", fmt.Errorf("create document: lookup created_by display name: %w", err)
	}

	var areaName sql.NullString
	if d.ProcessAreaCodeSnapshot != nil && *d.ProcessAreaCodeSnapshot != "" {
		if err := tx.QueryRowContext(ctx, `SELECT name FROM metaldocs.document_process_areas WHERE tenant_id=$1::uuid AND code=$2`, d.TenantID, *d.ProcessAreaCodeSnapshot).Scan(&areaName); err != nil && err != sql.ErrNoRows {
			return "", "", "", fmt.Errorf("create document: lookup area name: %w", err)
		}
	}

	if err := tx.QueryRowContext(ctx,
		`INSERT INTO documents (tenant_id, template_version_id, name, status, form_data_json, created_by, controlled_document_id, profile_code_snapshot, process_area_code_snapshot, code, revision_number, created_by_display_name_snapshot, area_name_snapshot)
		 SELECT $1, $2, $3, 'draft', $4, $5, $6, $7, $8, $9,
		        COALESCE((SELECT MAX(d2.revision_number) + 1 FROM documents d2 WHERE d2.controlled_document_id = $6), 0),
		        $10, $11
		 RETURNING id`,
		d.TenantID, d.TemplateVersionID, d.Name, d.FormDataJSON, d.CreatedBy, d.ControlledDocumentID, d.ProfileCodeSnapshot, d.ProcessAreaCodeSnapshot, d.Code, createdByDisplayName, areaName,
	).Scan(&docID); err != nil {
		return "", "", "", fmt.Errorf("insert document: %w", err)
	}

	if err := tx.QueryRowContext(ctx,
		`INSERT INTO editor_sessions (document_id, user_id, expires_at, last_acknowledged_revision_id, status)
		 VALUES ($1, $2, now() + interval '5 minutes', '00000000-0000-0000-0000-000000000000', 'active') RETURNING id`,
		docID, d.CreatedBy,
	).Scan(&sessionID); err != nil {
		return "", "", "", fmt.Errorf("insert session: %w", err)
	}

	if err := tx.QueryRowContext(ctx,
		`INSERT INTO document_revisions (document_id, parent_revision_id, session_id, storage_key, content_hash, form_data_snapshot)
		 VALUES ($1, NULL, $2, $3, $4, $5) RETURNING id`,
		docID, sessionID, initialStorageKey, initialContentHash, d.FormDataJSON,
	).Scan(&revID); err != nil {
		return "", "", "", fmt.Errorf("insert revision: %w", err)
	}

	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), "tenant"); err != nil {
		return "", "", "", fmt.Errorf("initialize document pointers: authz check: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE editor_sessions SET last_acknowledged_revision_id = $1 WHERE id = $2`,
		revID, sessionID,
	); err != nil {
		return "", "", "", fmt.Errorf("update session ack: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE documents SET current_revision_id = $1, active_session_id = $2, updated_at = now() WHERE id = $3`,
		revID, sessionID, docID,
	); err != nil {
		return "", "", "", fmt.Errorf("update document pointers: %w", err)
	}

	// Write snapshot columns in same tx (C2/C4: atomic creation).
	if snap := d.TemplateSnapshot; snap.PlaceholderSchemaJSON != nil || snap.CompositionJSON != nil || snap.BodyDocxS3Key != "" {
		h := snap.Hashes()
		if _, err := tx.ExecContext(ctx, `
			UPDATE public.documents
			   SET placeholder_schema_snapshot = $1,
			       placeholder_schema_hash     = $2,
			       composition_config_snapshot = $3,
			       composition_config_hash     = $4,
			       body_docx_snapshot_s3_key   = $5,
			       body_docx_hash              = $6
			 WHERE id = $7`,
			snap.PlaceholderSchemaJSON, h.PlaceholderSchemaHash,
			snap.CompositionJSON, h.CompositionHash,
			snap.BodyDocxS3Key, h.BodyDocxHash,
			docID,
		); err != nil {
			return "", "", "", fmt.Errorf("write snapshot: %w", err)
		}
	}

	// Seed required placeholder_values rows in same tx (C4).
	for _, p := range requiredPlaceholders {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO public.document_placeholder_values
			    (tenant_id, revision_id, placeholder_id, source, created_at, updated_at)
			VALUES ($1, $2, $3, 'default', NOW(), NOW())
			ON CONFLICT DO NOTHING`,
			d.TenantID, revID, p.ID,
		); err != nil {
			return "", "", "", fmt.Errorf("seed placeholder %q: %w", p.ID, err)
		}
	}

	return docID, revID, sessionID, nil
}

// SetRevisionStorageKey finalizes storage_key after S3 AdoptTempObject.
// Legacy pattern — only used by Service.CreateDocument (via DuplicateDocument).
// Atomic flow sets storage_key in CreateDocumentTx; this method is bypassed.
func (r *Repository) SetRevisionStorageKey(ctx context.Context, revID, storageKey string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE document_revisions SET storage_key = $1 WHERE id = $2 AND storage_key = ''`,
		storageKey, revID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("revision %s already has storage_key set", revID)
	}
	return nil
}

func (r *Repository) GetDocument(ctx context.Context, tenantID, id string) (*domain.Document, error) {
	var d domain.Document
	err := r.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, template_version_id, name, status, form_data_json,
		        coalesce(current_revision_id::text, ''), coalesce(active_session_id::text, ''),
		        archived_at, created_at, updated_at, created_by, revision_title,
		        controlled_document_id, profile_code_snapshot, process_area_code_snapshot,
		        coalesce(code,''), revision_version
		 FROM documents WHERE id=$1 AND tenant_id=$2`, id, tenantID,
	).Scan(&d.ID, &d.TenantID, &d.TemplateVersionID, &d.Name, &d.Status, &d.FormDataJSON,
		&d.CurrentRevisionID, &d.ActiveSessionID, &d.ArchivedAt,
		&d.CreatedAt, &d.UpdatedAt, &d.CreatedBy, &d.RevisionTitle, &d.ControlledDocumentID, &d.ProfileCodeSnapshot,
		&d.ProcessAreaCodeSnapshot, &d.Code,
		&d.RevisionVersion)
	if errors.Is(err, sql.ErrNoRows) || isInvalidUUID(err) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *Repository) UpdateDocumentName(ctx context.Context, tenantID, docID, name string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), "tenant"); err != nil {
		return fmt.Errorf("update document name: authz check: %w", err)
	}

	res, err := tx.ExecContext(ctx,
		`UPDATE documents SET name=$2, updated_at=now() WHERE id=$1 AND tenant_id=$3`,
		docID, name, tenantID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return tx.Commit()
}

func (r *Repository) UpdateDocumentNameTx(ctx context.Context, tx *sql.Tx, tenantID, docID, name string) error {
	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), "tenant"); err != nil {
		return fmt.Errorf("update document name: authz check: %w", err)
	}
	res, err := tx.ExecContext(ctx,
		`UPDATE documents SET name=$2, updated_at=now() WHERE id=$1 AND tenant_id=$3`,
		docID, name, tenantID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *Repository) ListDocuments(ctx context.Context, tenantID string) ([]domain.Document, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, tenant_id, template_version_id, name, status, form_data_json,
		        coalesce(current_revision_id::text, ''), coalesce(active_session_id::text, ''),
		        archived_at, created_at, updated_at, created_by,
		        controlled_document_id, coalesce(code,'')
		 FROM documents WHERE tenant_id=$1 AND archived_at IS NULL ORDER BY updated_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Document{}
	for rows.Next() {
		var d domain.Document
		if err := rows.Scan(&d.ID, &d.TenantID, &d.TemplateVersionID, &d.Name, &d.Status, &d.FormDataJSON,
			&d.CurrentRevisionID, &d.ActiveSessionID, &d.ArchivedAt,
			&d.CreatedAt, &d.UpdatedAt, &d.CreatedBy, &d.ControlledDocumentID, &d.Code); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// ListDocumentsForUser restricts metadata leakage for document_filler role -
// returns only docs the actor created. Admins / template_* roles use the
// unrestricted ListDocuments path instead.
func (r *Repository) ListDocumentsForUser(ctx context.Context, tenantID, userID string) ([]domain.Document, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, tenant_id, template_version_id, name, status, form_data_json,
		        coalesce(current_revision_id::text, ''), coalesce(active_session_id::text, ''),
		        archived_at, created_at, updated_at, created_by,
		        controlled_document_id, coalesce(code,'')
		 FROM documents WHERE tenant_id=$1 AND created_by=$2 AND archived_at IS NULL ORDER BY updated_at DESC`, tenantID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Document{}
	for rows.Next() {
		var d domain.Document
		if err := rows.Scan(&d.ID, &d.TenantID, &d.TemplateVersionID, &d.Name, &d.Status, &d.FormDataJSON,
			&d.CurrentRevisionID, &d.ActiveSessionID, &d.ArchivedAt,
			&d.CreatedAt, &d.UpdatedAt, &d.CreatedBy, &d.ControlledDocumentID, &d.Code); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

type ListOptions struct {
	Page            int
	PageSize        int
	CreatedBy       string
	Status          []string
	AreaCode        string
	ProfileCode     string
	Q               string
	IncludeArchived bool
}

func (o ListOptions) Offset() int {
	if o.Page < 1 {
		return 0
	}
	return (o.Page - 1) * o.Limit()
}

func (o ListOptions) Limit() int {
	if o.PageSize == 0 {
		return 20
	}
	if o.PageSize < 1 {
		return 1
	}
	if o.PageSize > 50 {
		return 50
	}
	return o.PageSize
}

func buildDocumentFilter(tenantID string, opts ListOptions) (whereClause string, args []interface{}) {
	conds := []string{"tenant_id = $1"}
	args = []interface{}{tenantID}
	if !opts.IncludeArchived {
		conds = append(conds, "archived_at IS NULL")
	}

	if opts.CreatedBy != "" {
		args = append(args, opts.CreatedBy)
		conds = append(conds, fmt.Sprintf("created_by = $%d", len(args)))
	}
	if len(opts.Status) > 0 {
		args = append(args, pq.Array(opts.Status))
		conds = append(conds, fmt.Sprintf("status = ANY($%d)", len(args)))
	}
	if opts.AreaCode != "" {
		args = append(args, opts.AreaCode)
		conds = append(conds, fmt.Sprintf("process_area_code_snapshot = $%d", len(args)))
	}
	if opts.ProfileCode != "" {
		args = append(args, opts.ProfileCode)
		conds = append(conds, fmt.Sprintf("profile_code_snapshot = $%d", len(args)))
	}
	if opts.Q != "" {
		args = append(args, "%"+opts.Q+"%")
		conds = append(conds, fmt.Sprintf("name ILIKE $%d", len(args)))
	}

	return strings.Join(conds, " AND "), args
}

func (r *Repository) ListDocumentsPaginated(ctx context.Context, tenantID string, opts ListOptions) ([]*domain.Document, error) {
	where, args := buildDocumentFilter(tenantID, opts)
	args = append(args, opts.Limit(), opts.Offset())
	limitIdx := len(args) - 1
	offsetIdx := len(args)
	q := fmt.Sprintf(`SELECT id, tenant_id, template_version_id, name, status, form_data_json,
			coalesce(current_revision_id::text, ''), coalesce(active_session_id::text, ''),
			archived_at, created_at, updated_at, created_by,
			controlled_document_id, coalesce(code,'')
		FROM documents
		WHERE %s
		ORDER BY updated_at DESC
		LIMIT $%d OFFSET $%d`, where, limitIdx, offsetIdx)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]*domain.Document, 0)
	for rows.Next() {
		var d domain.Document
		if err := rows.Scan(&d.ID, &d.TenantID, &d.TemplateVersionID, &d.Name, &d.Status, &d.FormDataJSON,
			&d.CurrentRevisionID, &d.ActiveSessionID, &d.ArchivedAt,
			&d.CreatedAt, &d.UpdatedAt, &d.CreatedBy, &d.ControlledDocumentID, &d.Code); err != nil {
			return nil, err
		}
		out = append(out, &d)
	}
	return out, rows.Err()
}

func (r *Repository) CountDocuments(ctx context.Context, tenantID string, opts ListOptions) (int64, error) {
	where, args := buildDocumentFilter(tenantID, opts)
	q := fmt.Sprintf(`SELECT COUNT(*) FROM documents WHERE %s`, where)
	var n int64
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func (r *Repository) StatsByStatus(ctx context.Context, tenantID string, opts ListOptions) (map[string]int64, error) {
	where, args := buildDocumentFilter(tenantID, opts)
	q := fmt.Sprintf(`SELECT status, COUNT(*) FROM documents WHERE %s GROUP BY status`, where)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]int64)
	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		out[status] = count
	}
	return out, rows.Err()
}

func (r *Repository) StatsByArea(ctx context.Context, tenantID string, opts ListOptions) (map[string]int64, error) {
	where, args := buildDocumentFilter(tenantID, opts)
	q := fmt.Sprintf(`SELECT COALESCE(process_area_code_snapshot, '') AS area, COUNT(*) FROM documents WHERE %s GROUP BY area`, where)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]int64)
	for rows.Next() {
		var area string
		var count int64
		if err := rows.Scan(&area, &count); err != nil {
			return nil, err
		}
		out[area] = count
	}
	return out, rows.Err()
}

func (r *Repository) UpdateDocumentStatus(ctx context.Context, tenantID, id string, cur, next domain.DocumentStatus, stampTime bool) error {
	col := ""
	if stampTime {
		if next == domain.DocStatusArchived {
			col = "archived_at  = now(),"
		}
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), "tenant"); err != nil {
		return fmt.Errorf("update document status: authz check: %w", err)
	}

	res, err := tx.ExecContext(ctx,
		fmt.Sprintf(`UPDATE documents SET status=$1, %s updated_at=now() WHERE id=$2 AND tenant_id=$3 AND status=$4`, col),
		next, id, tenantID, cur)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrInvalidStateTransition
	}
	return tx.Commit()
}

// AcquireSession attempts to claim the single active-session slot for a doc.
// Relies on partial unique index idx_one_active_session_per_doc.
// Returns (newSession, nil) on success. Returns existing active session
// with ErrSessionTaken if another user holds it.
func (r *Repository) AcquireSession(ctx context.Context, tenantID, docID, userID string) (*domain.Session, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, "SELECT set_config('metaldocs.tenant_id', $1, true)", tenantID); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, "SELECT set_config('metaldocs.actor_id', $1, true)", userID); err != nil {
		return nil, err
	}
	ctx = authz.WithCapCache(ctx)
	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), "tenant"); err != nil {
		return nil, fmt.Errorf("acquire session: authz check: %w", err)
	}

	var existingID, existingUser, existingStatus, existingAck string
	err = tx.QueryRowContext(ctx,
		`SELECT id::text, user_id::text, status, coalesce(last_acknowledged_revision_id::text,'') FROM editor_sessions
		 WHERE document_id=$1 AND status='active' FOR UPDATE`, docID,
	).Scan(&existingID, &existingUser, &existingStatus, &existingAck)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if err == nil {
		// Caller already holds it - refresh.
		if existingUser == userID {
			var refreshedExpiresAt time.Time
			if err := tx.QueryRowContext(ctx, `UPDATE editor_sessions SET expires_at = now() + interval '5 minutes' WHERE id=$1 RETURNING expires_at`, existingID).Scan(&refreshedExpiresAt); err != nil {
				return nil, err
			}
			s := &domain.Session{ID: existingID, DocumentID: docID, UserID: userID, LastAcknowledgedRevisionID: existingAck, Status: domain.SessionActive, ExpiresAt: refreshedExpiresAt}
			return s, tx.Commit()
		}
		// Someone else owns it.
		s := &domain.Session{ID: existingID, DocumentID: docID, UserID: existingUser, Status: domain.SessionActive}
		return s, domain.ErrSessionTaken
	}

	// No active session. Grab current_revision_id from documents.
	var curRev string
	if err := tx.QueryRowContext(ctx,
		`SELECT coalesce(current_revision_id::text,'') FROM documents WHERE id=$1 AND tenant_id=$2`, docID, tenantID,
	).Scan(&curRev); err != nil {
		return nil, err
	}
	if curRev == "" {
		return nil, fmt.Errorf("document has no current revision")
	}

	var newID string
	var newExpiresAt time.Time
	if err := tx.QueryRowContext(ctx,
		`INSERT INTO editor_sessions (document_id, user_id, expires_at, last_acknowledged_revision_id, status)
		 VALUES ($1, $2, now() + interval '5 minutes', $3, 'active') RETURNING id, expires_at`,
		docID, userID, curRev,
	).Scan(&newID, &newExpiresAt); err != nil {
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, `UPDATE documents SET active_session_id=$1, updated_at=now() WHERE id=$2`, newID, docID); err != nil {
		return nil, err
	}

	return &domain.Session{ID: newID, DocumentID: docID, UserID: userID, LastAcknowledgedRevisionID: curRev, Status: domain.SessionActive, ExpiresAt: newExpiresAt}, tx.Commit()
}

func (r *Repository) HeartbeatSession(ctx context.Context, sessionID, userID string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE editor_sessions SET expires_at = now() + interval '5 minutes'
		 WHERE id=$1 AND user_id=$2 AND status='active'`, sessionID, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrSessionInactive
	}
	return nil
}

func (r *Repository) ReleaseSession(ctx context.Context, tenantID, sessionID, userID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, "SELECT set_config('metaldocs.tenant_id', $1, true)", tenantID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "SELECT set_config('metaldocs.actor_id', $1, true)", userID); err != nil {
		return err
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), "tenant"); err != nil {
		return fmt.Errorf("release session: authz check: %w", err)
	}
	res, err := tx.ExecContext(ctx,
		`UPDATE editor_sessions SET status='released', released_at=now()
		 WHERE id=$1 AND user_id=$2 AND status='active'`, sessionID, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrSessionInactive
	}
	if _, err := tx.ExecContext(ctx, `UPDATE documents SET active_session_id=NULL, updated_at=now() WHERE active_session_id=$1`, sessionID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) ForceReleaseSession(ctx context.Context, tenantID, adminID, sessionID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, "SELECT set_config('metaldocs.tenant_id', $1, true)", tenantID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "SELECT set_config('metaldocs.actor_id', $1, true)", adminID); err != nil {
		return err
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), "tenant"); err != nil {
		return fmt.Errorf("force release session: authz check: %w", err)
	}
	res, err := tx.ExecContext(ctx,
		`UPDATE editor_sessions SET status='force_released', released_at=now()
		 WHERE id=$1 AND status='active'`, sessionID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrSessionInactive
	}
	if _, err := tx.ExecContext(ctx, `UPDATE documents SET active_session_id=NULL, updated_at=now() WHERE active_session_id=$1`, sessionID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) ExpireStaleSessions(ctx context.Context, now time.Time) (int, error) {
	// Single atomic CTE: expire sessions and clear document pointers in one tx.
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	if err := authz.BypassSystem(ctx, tx); err != nil {
		return 0, err
	}

	var n int
	err = tx.QueryRowContext(ctx, `
		WITH expired AS (
			UPDATE editor_sessions SET status='expired'
			WHERE status='active' AND expires_at < $1
			RETURNING id
		),
		cleared AS (
			UPDATE documents SET active_session_id=NULL, updated_at=now()
			WHERE active_session_id IN (SELECT id FROM expired)
			RETURNING 1
		)
		SELECT count(*) FROM expired`, now,
	).Scan(&n)
	if err != nil {
		return 0, err
	}
	return n, tx.Commit()
}

func (r *Repository) PresignReserve(ctx context.Context, sessionID, userID, docID, baseRevisionID, contentHash, storageKey string, expiresAt time.Time) (pendingID string, err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	// Verify session active, holder matches, and session points at base.
	var sessUser, sessDoc, sessAck, sessStatus string
	err = tx.QueryRowContext(ctx,
		`SELECT user_id::text, document_id::text, last_acknowledged_revision_id::text, status
		 FROM editor_sessions WHERE id=$1 FOR UPDATE`, sessionID,
	).Scan(&sessUser, &sessDoc, &sessAck, &sessStatus)
	if err != nil {
		return "", err
	}
	if sessStatus != string(domain.SessionActive) {
		return "", domain.ErrSessionInactive
	}
	if sessUser != userID || sessDoc != docID {
		return "", domain.ErrSessionNotHolder
	}
	if sessAck != baseRevisionID {
		return "", domain.ErrStaleBase
	}

	// Idempotent on (session, base, hash): ON CONFLICT returns existing row's id.
	err = tx.QueryRowContext(ctx,
		`INSERT INTO autosave_pending_uploads
		   (session_id, document_id, base_revision_id, content_hash, storage_key, expires_at)
		 VALUES ($1,$2,$3,$4,$5,$6)
		 ON CONFLICT (session_id, base_revision_id, content_hash)
		 DO UPDATE SET presigned_at = autosave_pending_uploads.presigned_at
		 RETURNING id`,
		sessionID, docID, baseRevisionID, contentHash, storageKey, expiresAt,
	).Scan(&pendingID)
	if err != nil {
		return "", err
	}
	return pendingID, tx.Commit()
}

// CommitResult + PendingCommitMeta + RestoreResult are mirrored in application
// (same-shape type aliases) so handlers depend only on application types.
type CommitResult struct {
	RevisionID      string
	RevisionNum     int64
	AlreadyConsumed bool
	FileSizeBytes   *int64
	PageCount       *int
	PageCountSource *string
}

type PendingCommitMeta struct {
	SessionID           string
	DocumentID          string
	BaseRevisionID      string
	ExpectedContentHash string
	StorageKey          string
	ExpiresAt           time.Time
	ConsumedAt          *time.Time
}

type RestoreResult struct {
	NewRevisionID   string
	NewRevisionNum  int64
	CheckpointRevID string
	// Idempotent is true when ON CONFLICT (document_id, content_hash) DO
	// UPDATE SET id = id fired - the current head already matched the
	// checkpoint hash, so no new row was inserted. Used by the handler to
	// surface `idempotent: true` on the restore response.
	Idempotent bool
}

// GetPendingForCommit returns the minimal metadata the service needs before
// performing server-authoritative hash verification. Short, unlocked read;
// CommitUpload re-locks and re-checks under FOR UPDATE.
func (r *Repository) GetPendingForCommit(ctx context.Context, pendingID string) (*PendingCommitMeta, error) {
	var m PendingCommitMeta
	err := r.db.QueryRowContext(ctx,
		`SELECT session_id::text, document_id::text, base_revision_id::text,
		        content_hash, storage_key, expires_at, consumed_at
		 FROM autosave_pending_uploads WHERE id=$1`, pendingID,
	).Scan(&m.SessionID, &m.DocumentID, &m.BaseRevisionID,
		&m.ExpectedContentHash, &m.StorageKey, &m.ExpiresAt, &m.ConsumedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrPendingNotFound
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// CommitUpload encodes every DB-level rejection branch below as explicit errors.
// Callers translate each to the matching HTTP status (404/409/410/422 per spec).
// Server-authoritative content_hash verification happens in Service BEFORE this
// method is called; `serverComputedHash` is the hash the service streamed from
// S3 and therefore trusted. CommitUpload still compares it to pending.content_hash
// under FOR UPDATE to catch TOCTOU races.
func (r *Repository) setAuthzGUC(ctx context.Context, tx *sql.Tx, stmt, tenantID, actorID string) error {
	if _, err := tx.ExecContext(ctx, stmt, tenantID, actorID); err != nil {
		return fmt.Errorf("set authz guc: %w", err)
	}
	return nil
}

func (r *Repository) CommitUpload(ctx context.Context, tenantID, sessionID, userID, docID, pendingID, serverComputedHash string, formDataSnapshot []byte, fileSizeBytes int64, pageCount *int, pageCountSource *string) (*CommitResult, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	authzGUCStmt := `
		SELECT
			set_config('metaldocs.tenant_id', $1, true),
			set_config('metaldocs.actor_id', $2, true)
	`
	if err := r.setAuthzGUC(ctx, tx, authzGUCStmt, tenantID, userID); err != nil {
		return nil, err
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), "tenant"); err != nil {
		return nil, fmt.Errorf("commit autosave: authz check: %w", err)
	}

	// Lock pending row.
	var p domain.PendingUpload
	err = tx.QueryRowContext(ctx,
		`SELECT id::text, session_id::text, document_id::text, base_revision_id::text, content_hash,
		        storage_key, expires_at, consumed_at
		 FROM autosave_pending_uploads WHERE id=$1 FOR UPDATE`, pendingID,
	).Scan(&p.ID, &p.SessionID, &p.DocumentID, &p.BaseRevisionID, &p.ContentHash, &p.StorageKey, &p.ExpiresAt, &p.ConsumedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrPendingNotFound
	}
	if err != nil {
		return nil, err
	}

	if p.SessionID != sessionID || p.DocumentID != docID {
		return nil, domain.ErrMisbound
	}
	if p.ConsumedAt != nil {
		// Idempotent replay: the upload already committed. Session state is intentionally
		// not checked — the revision exists regardless of whether the session is still
		// active. Return the existing revision so the client can ack it safely.
		var rid string
		var rnum int64
		var replayFileSize sql.NullInt64
		var replayPageCount sql.NullInt64
		var replayPageCountSource sql.NullString
		if err := tx.QueryRowContext(ctx,
			`SELECT id::text, revision_num, file_size_bytes, page_count, page_count_source FROM document_revisions
			 WHERE document_id=$1 AND content_hash=$2`, docID, p.ContentHash,
		).Scan(&rid, &rnum, &replayFileSize, &replayPageCount, &replayPageCountSource); err != nil {
			return nil, fmt.Errorf("replay lookup: %w", err)
		}
		res := &CommitResult{RevisionID: rid, RevisionNum: rnum, AlreadyConsumed: true}
		if replayFileSize.Valid {
			res.FileSizeBytes = &replayFileSize.Int64
		}
		if replayPageCount.Valid {
			pc := int(replayPageCount.Int64)
			res.PageCount = &pc
		}
		if replayPageCountSource.Valid {
			res.PageCountSource = &replayPageCountSource.String
		}
		return res, tx.Commit()
	}
	if time.Now().After(p.ExpiresAt) {
		return nil, domain.ErrExpiredUpload
	}

	// Re-verify session still active + holder + ack still matches base.
	var sessUser, sessAck, sessStatus string
	err = tx.QueryRowContext(ctx,
		`SELECT user_id::text, last_acknowledged_revision_id::text, status
		 FROM editor_sessions WHERE id=$1 FOR UPDATE`, sessionID,
	).Scan(&sessUser, &sessAck, &sessStatus)
	if err != nil {
		return nil, err
	}
	if sessStatus != string(domain.SessionActive) {
		return nil, domain.ErrSessionInactive
	}
	if sessUser != userID {
		return nil, domain.ErrSessionNotHolder
	}
	if sessAck != p.BaseRevisionID {
		return nil, domain.ErrStaleBase
	}

	// TOCTOU guard: service verified S3 hash matches pending.content_hash moments
	// before this call, but a concurrent tx could have rewritten the pending row.
	// Re-check under lock.
	if serverComputedHash != p.ContentHash {
		return nil, domain.ErrContentHashMismatch
	}

	var revID string
	var revNum int64
	var committedFileSize sql.NullInt64
	var committedPageCount sql.NullInt64
	var committedPageCountSource sql.NullString
	if err := tx.QueryRowContext(ctx,
		`INSERT INTO document_revisions
		   (document_id, parent_revision_id, session_id, storage_key, content_hash, form_data_snapshot, file_size_bytes, page_count, page_count_source)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		 RETURNING id::text, revision_num, file_size_bytes, page_count, page_count_source`,
		docID, p.BaseRevisionID, sessionID, p.StorageKey, p.ContentHash, formDataSnapshot, fileSizeBytes, pageCount, pageCountSource,
	).Scan(&revID, &revNum, &committedFileSize, &committedPageCount, &committedPageCountSource); err != nil {
		return nil, fmt.Errorf("insert revision: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE documents SET current_revision_id=$1, form_data_json=$2, updated_at=now() WHERE id=$3`,
		revID, formDataSnapshot, docID,
	); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE editor_sessions SET last_acknowledged_revision_id=$1 WHERE id=$2`, revID, sessionID,
	); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE autosave_pending_uploads SET consumed_at=now() WHERE id=$1`, pendingID,
	); err != nil {
		return nil, err
	}

	res := &CommitResult{RevisionID: revID, RevisionNum: revNum}
	if committedFileSize.Valid {
		res.FileSizeBytes = &committedFileSize.Int64
	}
	if committedPageCount.Valid {
		pc := int(committedPageCount.Int64)
		res.PageCount = &pc
	}
	if committedPageCountSource.Valid {
		res.PageCountSource = &committedPageCountSource.String
	}
	return res, tx.Commit()
}

func (r *Repository) DeleteExpiredPending(ctx context.Context, olderThan time.Time) (int, error) {
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM autosave_pending_uploads WHERE expires_at < $1 AND consumed_at IS NULL`,
		olderThan)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

func (r *Repository) CreateCheckpoint(ctx context.Context, docID, actorUserID, label string) (*domain.Checkpoint, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var revID string
	if err := tx.QueryRowContext(ctx,
		`SELECT current_revision_id::text FROM documents WHERE id=$1 FOR UPDATE`, docID,
	).Scan(&revID); err != nil {
		return nil, err
	}
	if revID == "" {
		return nil, fmt.Errorf("document has no current revision")
	}

	var nextVer int
	if err := tx.QueryRowContext(ctx,
		`SELECT coalesce(max(version_num),0)+1 FROM document_checkpoints WHERE document_id=$1`, docID,
	).Scan(&nextVer); err != nil {
		return nil, err
	}

	cp := &domain.Checkpoint{DocumentID: docID, RevisionID: revID, VersionNum: nextVer, Label: label, CreatedBy: actorUserID}
	if err := tx.QueryRowContext(ctx,
		`INSERT INTO document_checkpoints (document_id, revision_id, version_num, label, created_by)
		 VALUES ($1,$2,$3,$4,$5) RETURNING id::text, created_at`,
		docID, revID, nextVer, label, actorUserID,
	).Scan(&cp.ID, &cp.CreatedAt); err != nil {
		return nil, err
	}

	return cp, tx.Commit()
}

func (r *Repository) ListCheckpoints(ctx context.Context, docID string) ([]domain.Checkpoint, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id::text, document_id::text, revision_id::text, version_num, coalesce(label,''), created_at, created_by::text
		 FROM document_checkpoints WHERE document_id=$1 ORDER BY version_num DESC`, docID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Checkpoint{}
	for rows.Next() {
		var c domain.Checkpoint
		if err := rows.Scan(&c.ID, &c.DocumentID, &c.RevisionID, &c.VersionNum, &c.Label, &c.CreatedAt, &c.CreatedBy); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *Repository) ListRevisionHistory(ctx context.Context, tenantID, docID string) ([]domain.RevisionHistoryItem, error) {
	const q = `
WITH anchor AS (
  SELECT controlled_document_id, id AS current_document_id
  FROM documents
  WHERE tenant_id = $1::uuid AND id = $2::uuid
)
SELECT d.id::text,
       d.revision_number,
       COALESCE(d.revision_title, ''),
       d.status,
       d.created_at,
       (d.id = a.current_document_id) AS is_current
FROM anchor a
JOIN documents d
  ON (a.controlled_document_id IS NOT NULL AND d.controlled_document_id = a.controlled_document_id)
  OR (a.controlled_document_id IS NULL AND d.id = a.current_document_id)
WHERE d.tenant_id = $1::uuid
ORDER BY d.revision_number DESC, d.created_at DESC
`

	rows, err := r.db.QueryContext(ctx, q, tenantID, docID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.RevisionHistoryItem, 0)
	for rows.Next() {
		var item domain.RevisionHistoryItem
		if err := rows.Scan(
			&item.DocumentID,
			&item.RevisionNumber,
			&item.RevisionTitle,
			&item.Status,
			&item.CreatedAt,
			&item.IsCurrent,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) GetRevision(ctx context.Context, docID, revID string) (*domain.Revision, error) {
	var rv domain.Revision
	err := r.db.QueryRowContext(ctx,
		`SELECT id::text, document_id::text, revision_num, coalesce(parent_revision_id::text,''), session_id::text, storage_key, content_hash, form_data_snapshot, created_at
		 FROM document_revisions WHERE id=$1 AND document_id=$2`, revID, docID,
	).Scan(&rv.ID, &rv.DocumentID, &rv.RevisionNum, &rv.ParentRevisionID, &rv.SessionID, &rv.StorageKey, &rv.ContentHash, &rv.FormDataSnapshot, &rv.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) || isInvalidUUID(err) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &rv, nil
}

// RestoreCheckpoint is forward-only: it resolves the checkpoint's revision,
// copies that revision's storage_key + content_hash + form_data_snapshot into
// a NEW revision appended to head (parent = current_revision_id). It never
// rewrites history or rewinds revision_num. Session holder/active checks apply
// - restore is a session-authoritative action equivalent to a large autosave.
func (r *Repository) RestoreCheckpoint(ctx context.Context, docID, actorUserID string, versionNum int) (*RestoreResult, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Lock document + require active session held by actor.
	var sessID, sessUser, sessStatus string
	var curRev string
	if err := tx.QueryRowContext(ctx,
		`SELECT coalesce(active_session_id::text,''), coalesce(current_revision_id::text,'')
		 FROM documents WHERE id=$1 FOR UPDATE`, docID,
	).Scan(&sessID, &curRev); err != nil {
		return nil, err
	}
	if sessID == "" {
		return nil, domain.ErrSessionInactive
	}
	if err := tx.QueryRowContext(ctx,
		`SELECT user_id::text, status FROM editor_sessions WHERE id=$1 FOR UPDATE`, sessID,
	).Scan(&sessUser, &sessStatus); err != nil {
		return nil, err
	}
	if sessStatus != string(domain.SessionActive) {
		return nil, domain.ErrSessionInactive
	}
	if sessUser != actorUserID {
		return nil, domain.ErrSessionNotHolder
	}

	// Resolve checkpoint.
	var cpRevID, cpStorageKey, cpContentHash string
	var cpFormData []byte
	err = tx.QueryRowContext(ctx,
		`SELECT cp.revision_id::text, r.storage_key, r.content_hash, r.form_data_snapshot
		 FROM document_checkpoints cp
		 JOIN document_revisions r ON r.id = cp.revision_id
		 WHERE cp.document_id=$1 AND cp.version_num=$2`, docID, versionNum,
	).Scan(&cpRevID, &cpStorageKey, &cpContentHash, &cpFormData)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrCheckpointNotFound
	}
	if err != nil {
		return nil, err
	}

	// Append a new head revision pointing at the checkpoint's storage_key. The
	// content_hash is reused (content-addressed storage means no new upload).
	// NOTE: document_revisions.UNIQUE (document_id, content_hash) means restoring
	// to a hash that already exists as head is a no-op - ON CONFLICT returns the
	// existing row (xmax != 0 in pg). We detect that via `xmax::text::bigint <> 0`
	// on the RETURNING row to set RestoreResult.Idempotent = true.
	var newRevID string
	var newRevNum int64
	var idempotent bool
	err = tx.QueryRowContext(ctx,
		`INSERT INTO document_revisions
		   (document_id, parent_revision_id, session_id, storage_key, content_hash, form_data_snapshot)
		 VALUES ($1,$2,$3,$4,$5,$6)
		 ON CONFLICT (document_id, content_hash)
		 DO UPDATE SET id = document_revisions.id
		 RETURNING id::text, revision_num, (xmax::text::bigint <> 0)`,
		docID, curRev, sessID, cpStorageKey, cpContentHash, cpFormData,
	).Scan(&newRevID, &newRevNum, &idempotent)
	if err != nil {
		return nil, fmt.Errorf("restore insert: %w", err)
	}

	// On idempotent (head already equals checkpoint hash): do NOT rewrite
	// documents.current_revision_id or the session ack - they already match.
	// On fresh insert: advance head + session ack.
	if !idempotent {
		if _, err := tx.ExecContext(ctx,
			`UPDATE documents SET current_revision_id=$1, form_data_json=$2, updated_at=now() WHERE id=$3`,
			newRevID, cpFormData, docID,
		); err != nil {
			return nil, err
		}
		if _, err := tx.ExecContext(ctx,
			`UPDATE editor_sessions SET last_acknowledged_revision_id=$1 WHERE id=$2`, newRevID, sessID,
		); err != nil {
			return nil, err
		}
	}

	return &RestoreResult{
		NewRevisionID:   newRevID,
		NewRevisionNum:  newRevNum,
		CheckpointRevID: cpRevID,
		Idempotent:      idempotent,
	}, tx.Commit()
}

// IsDocumentOwner returns true iff the document was created by userID under
// tenantID. Used by handler-level defense-in-depth for document_filler routes.
func (r *Repository) IsDocumentOwner(ctx context.Context, tenantID, docID, userID string) (bool, error) {
	var ok bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM documents WHERE id=$1 AND tenant_id=$2 AND created_by=$3)`,
		docID, tenantID, userID,
	).Scan(&ok)
	return ok, err
}

func (r *Repository) CreateComment(ctx context.Context, tenantID, documentID, authorID string, in domain.CommentCreateInput) (*domain.Comment, error) {
	row := r.db.QueryRowContext(ctx,
		`INSERT INTO document_comments (
			tenant_id, document_id, library_comment_id, parent_library_id, author_id, author_display, content_json
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id::text, tenant_id::text, document_id::text, library_comment_id, parent_library_id, author_id, author_display,
		          content_json, resolved_at, resolved_by, created_at, updated_at`,
		tenantID, documentID, in.LibraryCommentID, in.ParentLibraryID, authorID, in.AuthorDisplay, in.ContentJSON,
	)
	comment, err := scanComment(row)
	if err != nil {
		return nil, err
	}
	return comment, nil
}

func (r *Repository) ListComments(ctx context.Context, tenantID, documentID string) ([]domain.Comment, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id::text, tenant_id::text, document_id::text, library_comment_id, parent_library_id, author_id, author_display,
		        content_json, resolved_at, resolved_by, created_at, updated_at
		 FROM document_comments
		 WHERE tenant_id=$1 AND document_id=$2
		 ORDER BY created_at ASC`,
		tenantID, documentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Comment, 0)
	for rows.Next() {
		comment, err := scanComment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *comment)
	}
	return out, rows.Err()
}

func (r *Repository) UpdateComment(ctx context.Context, tenantID, documentID string, libraryID int, userID string, in domain.CommentUpdateInput) (*domain.Comment, error) {
	now := time.Now().UTC()
	q := `
		UPDATE document_comments
		SET content_json = COALESCE($4::jsonb, content_json),
		    resolved_at = CASE
		      WHEN $5::bool IS NULL THEN resolved_at
		      WHEN $5::bool THEN $6
		      ELSE NULL
		    END,
		    resolved_by = CASE
		      WHEN $5::bool IS NULL THEN resolved_by
		      WHEN $5::bool THEN $7
		      ELSE NULL
		    END,
		    updated_at = $6
		WHERE tenant_id=$1 AND document_id=$2 AND library_comment_id=$3
		RETURNING id::text, tenant_id::text, document_id::text, library_comment_id, parent_library_id, author_id, author_display,
		          content_json, resolved_at, resolved_by, created_at, updated_at`

	var content any
	if in.ContentJSON != nil {
		content = *in.ContentJSON
	}
	comment, err := scanComment(r.db.QueryRowContext(ctx, q, tenantID, documentID, libraryID, content, in.Done, now, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrCommentNotFound
	}
	if err != nil {
		return nil, err
	}
	return comment, nil
}

func (r *Repository) DeleteComment(ctx context.Context, tenantID, documentID string, libraryID int) error {
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM document_comments
		 WHERE tenant_id=$1 AND document_id=$2 AND (library_comment_id=$3 OR parent_library_id=$3)`,
		tenantID, documentID, libraryID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrCommentNotFound
	}
	return nil
}

type commentScanner interface {
	Scan(dest ...any) error
}

// MarkArchived sets archived_at on a document without changing its status.
// Idempotent: WHERE archived_at IS NULL means double-call is a no-op.
func (r *Repository) MarkArchived(ctx context.Context, tenantID, docID, actorID string) error {
	_ = actorID // reserved for future audit column; audit via Service layer
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), "tenant"); err != nil {
		return fmt.Errorf("mark archived: authz check: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE public.documents
		   SET archived_at = now(), updated_at = now()
		 WHERE tenant_id = $1 AND id = $2 AND archived_at IS NULL`,
		tenantID, docID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// Unarchive clears archived_at, restoring the document to active queries.
func (r *Repository) Unarchive(ctx context.Context, tenantID, docID, actorID string) error {
	_ = actorID
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), "tenant"); err != nil {
		return fmt.Errorf("unarchive: authz check: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE public.documents
		   SET archived_at = NULL, updated_at = now()
		 WHERE tenant_id = $1 AND id = $2 AND archived_at IS NOT NULL`,
		tenantID, docID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func scanComment(row commentScanner) (*domain.Comment, error) {
	var (
		idText       string
		tenantText   string
		documentText string
		content      []byte
		c            domain.Comment
	)
	if err := row.Scan(
		&idText, &tenantText, &documentText, &c.LibraryCommentID, &c.ParentLibraryID, &c.AuthorID, &c.AuthorDisplay,
		&content, &c.ResolvedAt, &c.ResolvedBy, &c.CreatedAt, &c.UpdatedAt,
	); err != nil {
		return nil, err
	}
	id, err := uuid.Parse(idText)
	if err != nil {
		return nil, fmt.Errorf("parse comment id: %w", err)
	}
	tenantID, err := uuid.Parse(tenantText)
	if err != nil {
		return nil, fmt.Errorf("parse tenant id: %w", err)
	}
	documentID, err := uuid.Parse(documentText)
	if err != nil {
		return nil, fmt.Errorf("parse document id: %w", err)
	}
	c.ID = id
	c.TenantID = tenantID
	c.DocumentID = documentID
	c.ContentJSON = content
	return &c, nil
}
