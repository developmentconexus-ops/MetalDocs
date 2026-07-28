package infrastructure

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

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/internal/modules/documents/domain"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
	templatesdomain "metaldocs/internal/modules/templates/domain"

	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/pagination"
	"metaldocs/internal/platform/sqlescape"
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
	db          *sql.DB
	displayName iamdomain.UserDisplayNameReader
	cdRead      controlleddocumentsdomain.CDFieldReader
	areaCatalog taxonomydomain.AreaCatalogReader
}

// New constructs a Repository. displayName, cdRead and areaCatalog are required
// collaborators — use iamdomain.NoopUserDisplayNameReader{} /
// controlleddocumentsdomain.NoopCDFieldReader{} / taxonomydomain.
// NoopAreaCatalogReader{} for tests or paths that do not exercise them. cdRead is
// the controlleddocuments-owned read-port for controlled_documents fields
// (ADR-0039 D3(b), M2/F2.1); areaCatalog is the taxonomy-owned read-port for
// document_process_areas (ADR-0039 D3(b), M2/F2.3) — this module reads the area
// name through it (in-tx, non-recording) instead of issuing raw cross-module
// SQL. A nil areaCatalog is defaulted to the Noop reader.
func New(db *sql.DB, displayName iamdomain.UserDisplayNameReader, cdRead controlleddocumentsdomain.CDFieldReader, areaCatalog taxonomydomain.AreaCatalogReader) *Repository {
	if areaCatalog == nil {
		areaCatalog = taxonomydomain.NoopAreaCatalogReader{}
	}
	return &Repository{db: db, displayName: displayName, cdRead: cdRead, areaCatalog: areaCatalog}
}

// mustSQLTx asserts a db.Tx to *sql.Tx. All callers of this repository pass
// *sql.Tx at runtime; the assertion fails only if a test double is misused.
func mustSQLTx(tx db.Tx) *sql.Tx {
	sqltx, ok := tx.(*sql.Tx)
	if !ok {
		panic(fmt.Sprintf("documents/repository: expected *sql.Tx, got %T", tx))
	}
	return sqltx
}

// loadDocumentArea returns the document's process_area_code_snapshot within tx.
// document.create/edit are area-grade (ADR 0022 Phase 7): the area is passed to
// tier-2 authz.Require so an actor is authorized only within the document's
// process area. Empty string when the snapshot is NULL/empty — fails closed for
// non-system actors (system_admin still bypasses tier-2). A missing row yields
// ("", nil) so the authz check, not a row-existence probe, decides the outcome.
func loadDocumentArea(ctx context.Context, tx db.Tx, tenantID, docID string) (string, error) {
	var area sql.NullString
	err := tx.QueryRowContext(ctx,
		`SELECT process_area_code_snapshot FROM documents WHERE tenant_id=$1 AND id=$2`,
		tenantID, docID).Scan(&area)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("load document area: %w", err)
	}
	return area.String, nil
}

// loadDocumentAreaBySession resolves the document area from an editor session id,
// for session-keyed edit ops that carry no docID. Same fail-closed semantics.
func loadDocumentAreaBySession(ctx context.Context, tx db.Tx, tenantID, sessionID string) (string, error) {
	var area sql.NullString
	err := tx.QueryRowContext(ctx,
		`SELECT d.process_area_code_snapshot
		   FROM editor_sessions s
		   JOIN documents d ON d.id = s.document_id AND d.tenant_id = s.tenant_id
		  WHERE s.tenant_id=$1::uuid AND s.id=$2`,
		tenantID, sessionID).Scan(&area)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("load document area by session: %w", err)
	}
	return area.String, nil
}

// docAreaSnapshot derefs a document's nullable area snapshot to a tier-2 area arg.
func docAreaSnapshot(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// CreateDocumentTx inserts document + initial editor_session + initial
// revision, seeds template snapshot columns and required placeholder rows,
// all inside the caller-owned tx.
//
// initialStorageKey is the template's published docx key (template-passthrough):
// storage_key is written atomically with the insert so the editor opens
// immediately on first GET.
//
// It does NOT touch S3.
func (r *Repository) CreateDocumentTx(ctx context.Context, tx db.Tx, d *domain.Document, initialContentHash, initialStorageKey string, requiredPlaceholders []templatesdomain.Placeholder) (docID, revID, sessionID string, err error) {
	ctx = authz.WithCapCache(ctx)
	if err := authz.SeedTxIdentity(ctx, mustSQLTx(tx), d.TenantID, d.CreatedBy); err != nil {
		return "", "", "", err
	}

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

	// ADR 0022 Phase 7: a document is created INTO a process area — authorize
	// against that area (area-grade). system_admin bypasses tier-2.
	docArea := docAreaSnapshot(d.ProcessAreaCodeSnapshot)
	if err := authz.Require(ctx, mustSQLTx(tx), string(iamdomain.CapDocumentCreate), docArea); err != nil {
		return "", "", "", fmt.Errorf("create document: authz check: %w", err)
	}

	// Deferrable FKs allow inserting doc -> session -> revision in any order in tx.
	// COALESCE handles hypothetical NULL CD (schema enforces NOT NULL since migration 0129).

	// M4/F4.1: read display_name via the iam-owned port (off-tx, tenant-scoped).
	// The port reads the iam connection pool so the snapshot is never inside the
	// create tx (H-PRE-1). Becomes tenant-scoped — deliberate correctness
	// tightening (spec §non-goals, two bounded divergences).
	rawDisplayName, err := r.displayName.DisplayName(ctx, d.TenantID, d.CreatedBy)
	if err != nil {
		return "", "", "", fmt.Errorf("create document: lookup created_by display name: %w", err)
	}
	var createdByDisplayName sql.NullString
	if rawDisplayName != "" {
		createdByDisplayName = sql.NullString{String: rawDisplayName, Valid: true}
	}

	// M2/F2.3: read the area name through the taxonomy-owned AreaCatalogReader
	// port (ADR-0039 D3(b)) instead of raw cross-module SQL. tx is passed so the
	// read runs inside this create tx and stays non-recording (HS-PRE-1); found
	// == false reproduces the prior sql.ErrNoRows branch (NULL area_name_snapshot).
	var areaName sql.NullString
	if d.ProcessAreaCodeSnapshot != nil && *d.ProcessAreaCodeSnapshot != "" {
		name, found, err := r.areaCatalog.AreaName(ctx, tx, d.TenantID, *d.ProcessAreaCodeSnapshot)
		if err != nil {
			return "", "", "", fmt.Errorf("create document: lookup area name: %w", err)
		}
		if found {
			areaName = sql.NullString{String: name, Valid: true}
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
		`INSERT INTO editor_sessions (tenant_id, document_id, user_id, expires_at, last_acknowledged_revision_id, status)
		 VALUES ($1, $2, $3, now() + interval '5 minutes', '00000000-0000-0000-0000-000000000000', 'active') RETURNING id`,
		d.TenantID, docID, d.CreatedBy,
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

	if err := authz.Require(ctx, mustSQLTx(tx), string(iamdomain.CapDocumentEdit), docArea); err != nil {
		return "", "", "", fmt.Errorf("initialize document pointers: authz check: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE editor_sessions SET last_acknowledged_revision_id = $1 WHERE id = $2 AND tenant_id = $3::uuid`,
		revID, sessionID, d.TenantID,
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

// SeedDictionaryValuesTx pins resolved dictionary placeholder values (keyed by
// placeholder ID) into document_placeholder_values inside the caller-owned tx,
// with source='dictionary' (SP-2 D1). Upserts so it overrides any valueless
// source='default' seed for the same placeholder. Requires migration 0249 (the
// source CHECK widened to include 'dictionary').
func (r *Repository) SeedDictionaryValuesTx(ctx context.Context, tx db.Tx, tenantID, revisionID string, values map[string]string) error {
	for placeholderID, value := range values {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO public.document_placeholder_values
			    (tenant_id, revision_id, placeholder_id, value_text, source, created_at, updated_at)
			VALUES ($1::uuid, $2::uuid, $3, $4, 'dictionary', NOW(), NOW())
			ON CONFLICT (tenant_id, revision_id, placeholder_id) DO UPDATE SET
				value_text = EXCLUDED.value_text,
				source     = 'dictionary',
				updated_at = NOW()`,
			tenantID, revisionID, placeholderID, value,
		); err != nil {
			return fmt.Errorf("seed dictionary placeholder %q: %w", placeholderID, err)
		}
	}
	return nil
}

func (r *Repository) GetDocument(ctx context.Context, tenantID, id string) (*domain.Document, error) {
	var d domain.Document
	var currentFileSize sql.NullInt64
	var currentPageCount sql.NullInt64
	var currentPageCountSource sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT d.id, d.tenant_id, d.template_version_id, d.name, d.status, d.form_data_json,
		        coalesce(d.current_revision_id::text, ''), coalesce(d.active_session_id::text, ''),
		        d.archived_at, d.created_at, d.updated_at, d.created_by, d.revision_title,
		        d.controlled_document_id, d.profile_code_snapshot, d.process_area_code_snapshot,
		        coalesce(d.code,''), d.revision_version, d.revision_number,
		        cr.file_size_bytes, cr.page_count, cr.page_count_source,
		        d.effective_from, d.effective_to, d.review_due_at, d.last_reviewed_at, d.review_surfaced_at,
		        d.values_frozen_at
		 FROM documents d
		 LEFT JOIN document_revisions cr
		   ON cr.id = d.current_revision_id
		  AND cr.document_id = d.id
		 WHERE d.id=$1 AND d.tenant_id=$2`, id, tenantID,
	).Scan(&d.ID, &d.TenantID, &d.TemplateVersionID, &d.Name, &d.Status, &d.FormDataJSON,
		&d.CurrentRevisionID, &d.ActiveSessionID, &d.ArchivedAt,
		&d.CreatedAt, &d.UpdatedAt, &d.CreatedBy, &d.RevisionTitle, &d.ControlledDocumentID, &d.ProfileCodeSnapshot,
		&d.ProcessAreaCodeSnapshot, &d.Code,
		&d.RevisionVersion, &d.RevisionNumber, &currentFileSize, &currentPageCount, &currentPageCountSource,
		&d.EffectiveFrom, &d.EffectiveTo, &d.ReviewDueAt, &d.LastReviewedAt, &d.ReviewSurfacedAt,
		&d.ValuesFrozenAt)
	if errors.Is(err, sql.ErrNoRows) || isInvalidUUID(err) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if currentFileSize.Valid {
		fileSize := currentFileSize.Int64
		d.CurrentRevisionFileSizeBytes = &fileSize
	}
	if currentPageCount.Valid {
		pageCount := int(currentPageCount.Int64)
		d.CurrentRevisionPageCount = &pageCount
	}
	if currentPageCountSource.Valid {
		pageCountSource := currentPageCountSource.String
		d.CurrentRevisionPageCountSource = &pageCountSource
	}
	return &d, nil
}

func (r *Repository) UpdateDocumentName(ctx context.Context, tenantID, actorID, docID, name string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	ctx = authz.WithCapCache(ctx)
	if err := authz.SeedTxIdentity(ctx, tx, tenantID, actorID); err != nil {
		return err
	}
	docArea, err := loadDocumentArea(ctx, tx, tenantID, docID)
	if err != nil {
		return err
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), docArea); err != nil {
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

func (r *Repository) UpdateDocumentNameTx(ctx context.Context, tx db.Tx, tenantID, actorID, docID, name string) error {
	docArea, err := loadDocumentArea(ctx, tx, tenantID, docID)
	if err != nil {
		return err
	}
	if err := authz.Require(ctx, mustSQLTx(tx), string(iamdomain.CapDocumentEdit), docArea); err != nil {
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

// ListDocumentsForUser returns only docs the actor created (ownership-scoped
// list path; system_admin uses the unrestricted ListDocuments path). ADR 0022
// Phase 9.
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
	// Cursor is the opaque forward keyset cursor (FD-2). Empty = first page.
	Cursor          string
	PageSize        int
	CreatedBy       string
	Status          []string
	AreaCode        string
	ProfileCode     string
	Q               string
	IncludeArchived bool
	// ReviewDue filters to documents currently due for periodic review (M6
	// F6.2): published, effective, and review_due_at <= now(). See
	// buildDocumentFilter for the exact predicate.
	ReviewDue bool
}

// Limit clamps the page size to the design-system bounds (default 20, max 100).
func (o ListOptions) Limit() int {
	return pagination.ClampLimit(o.PageSize)
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
		// Escape LIKE metacharacters (%, _, \) so user input cannot inject
		// wildcards (CWE-943: info-inference + CPU-exhaustion). Requires the
		// matching ESCAPE clause below.
		args = append(args, "%"+sqlescape.LikeEscape(opts.Q)+"%")
		conds = append(conds, fmt.Sprintf("name ILIKE $%d ESCAPE '\\'", len(args)))
	}
	if opts.ReviewDue {
		// F6.4 D2 (worklist model, superseding the F6.2 recompute filter):
		// this returns the surfacer's worklist, not a live recompute of
		// "due right now". A document only appears once the River periodic
		// surfacer (jobs/document_review_surfacer) has flagged it for its
		// CURRENT review_due_at cycle — review_surfaced_at IS NOT NULL AND
		// review_surfaced_at >= review_due_at (the same idempotency
		// invariant ReviewSurfaceWriterPG.MarkSurfaced maintains: it only
		// (re)stamps review_surfaced_at when the prior mark is stale for the
		// current cycle, repository/review_surface_writer.go). This is a
		// derived/materialized read of the shared due-core eligibility
		// (repository/review_due_reader.go's `dueCorePredicate` — same
		// published + effective-window definition, restated here with the
		// filter builder's own dynamic status/effective-window fragment
		// because dueCorePredicate is a fixed-$1 SQL string and cannot be
		// spliced into this arg-numbered WHERE clause). D3 single-source is
		// honored at the reader/writer layer; this is the derived worklist
		// state, not a second definition of "due".
		//
		// mark-reviewed advances review_due_at (repository's mark-reviewed
		// path), which makes review_surfaced_at < review_due_at true again
		// and auto-expels the document from this filter until the surfacer
		// re-flags it on its next due cycle — no separate "un-surface" step
		// needed.
		conds = append(conds, `status = 'published'
		   AND effective_from IS NOT NULL AND effective_from <= now()
		   AND (effective_to IS NULL OR effective_to > now())
		   AND review_surfaced_at IS NOT NULL
		   AND review_surfaced_at >= review_due_at`)
	}

	return strings.Join(conds, " AND "), args
}

// ListDocumentsPaginated returns up to opts.Limit() documents ordered by
// (updated_at DESC, id DESC) using an opaque keyset cursor (FD-2). hasMore
// reports whether a further page exists (a limit+1 probe row was found and
// trimmed). The caller builds the next cursor from the last returned document's
// (UpdatedAt, ID).
func (r *Repository) ListDocumentsPaginated(ctx context.Context, tenantID string, opts ListOptions) (items []*domain.Document, total int64, hasMore bool, err error) {
	where, args := buildDocumentFilter(tenantID, opts)

	cursorTS, cursorID, err := pagination.DecodeCursor(opts.Cursor)
	if err != nil {
		return nil, 0, false, err
	}

	// The grand total must be counted over the base-filtered set BEFORE the cursor
	// predicate (a window over the cursor-filtered rows would only count the remaining
	// page tail). A CTE computes COUNT(*) OVER() on the base filter; the cursor + limit
	// are applied in the outer query. Page rows and total then share one MVCC snapshot
	// (one statement), eliminating the list/count TOCTOU race.
	var cursorClause string
	if cursorTS != "" {
		args = append(args, cursorTS, cursorID)
		cursorClause = fmt.Sprintf(" WHERE (updated_at, id) < ($%d::timestamptz, $%d)", len(args)-1, len(args))
	}

	limit := opts.Limit()
	args = append(args, limit+1) // +1 probe row to detect hasMore
	q := fmt.Sprintf(`WITH filtered AS (
			SELECT id, tenant_id, template_version_id, name, status, form_data_json,
				coalesce(current_revision_id::text, '') AS current_revision_id,
				coalesce(active_session_id::text, '') AS active_session_id,
				archived_at, created_at, updated_at, created_by,
				controlled_document_id, coalesce(code,'') AS code,
				profile_code_snapshot, process_area_code_snapshot,
				revision_version, revision_number,
				effective_from, effective_to, review_due_at, last_reviewed_at, review_surfaced_at,
				values_frozen_at,
				COUNT(*) OVER() AS total_count
			FROM documents
			WHERE %s
		)
		SELECT id, tenant_id, template_version_id, name, status, form_data_json,
			current_revision_id, active_session_id, archived_at, created_at, updated_at,
			created_by, controlled_document_id, code,
			profile_code_snapshot, process_area_code_snapshot,
			revision_version, revision_number,
			effective_from, effective_to, review_due_at, last_reviewed_at, review_surfaced_at,
			values_frozen_at, total_count
		FROM filtered%s
		ORDER BY updated_at DESC, id DESC
		LIMIT $%d`, where, cursorClause, len(args))

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, false, err
	}
	defer rows.Close()

	out := make([]*domain.Document, 0, limit+1)
	for rows.Next() {
		var d domain.Document
		var rowTotal int64
		if err := rows.Scan(&d.ID, &d.TenantID, &d.TemplateVersionID, &d.Name, &d.Status, &d.FormDataJSON,
			&d.CurrentRevisionID, &d.ActiveSessionID, &d.ArchivedAt,
			&d.CreatedAt, &d.UpdatedAt, &d.CreatedBy, &d.ControlledDocumentID, &d.Code,
			&d.ProfileCodeSnapshot, &d.ProcessAreaCodeSnapshot,
			&d.RevisionVersion, &d.RevisionNumber,
			&d.EffectiveFrom, &d.EffectiveTo, &d.ReviewDueAt, &d.LastReviewedAt, &d.ReviewSurfacedAt,
			&d.ValuesFrozenAt, &rowTotal); err != nil {
			return nil, 0, false, err
		}
		total = rowTotal // identical on every row (window over the full filtered set)
		out = append(out, &d)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, false, err
	}
	// Empty result (e.g. a cursor past the end) carries no total_count row; total stays 0.
	// Keyset clients read total on the first page only, so this edge is benign.
	if len(out) > limit {
		return out[:limit], total, true, nil
	}
	return out, total, false, nil
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

func (r *Repository) UpdateDocumentStatus(ctx context.Context, tenantID, actorID, id string, cur, next domain.DocumentStatus, stampTime bool) error {
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

	ctx = authz.WithCapCache(ctx)
	if err := authz.SeedTxIdentity(ctx, tx, tenantID, actorID); err != nil {
		return err
	}
	docArea, err := loadDocumentArea(ctx, tx, tenantID, id)
	if err != nil {
		return err
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), docArea); err != nil {
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
	ctx = authz.WithCapCache(ctx)
	if err := authz.SeedTxIdentity(ctx, tx, tenantID, userID); err != nil {
		return nil, err
	}
	docArea, err := loadDocumentArea(ctx, tx, tenantID, docID)
	if err != nil {
		return nil, err
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), docArea); err != nil {
		return nil, fmt.Errorf("acquire session: authz check: %w", err)
	}

	var existingID, existingUser, existingStatus, existingAck string
	err = tx.QueryRowContext(ctx,
		`SELECT id::text, user_id::text, status, coalesce(last_acknowledged_revision_id::text,'') FROM editor_sessions
		 WHERE tenant_id=$1::uuid AND document_id=$2 AND status='active' FOR UPDATE`, tenantID, docID,
	).Scan(&existingID, &existingUser, &existingStatus, &existingAck)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if err == nil {
		// Caller already holds it - refresh.
		if existingUser == userID {
			var refreshedExpiresAt time.Time
			if err := tx.QueryRowContext(ctx, `UPDATE editor_sessions SET expires_at = now() + interval '5 minutes' WHERE id=$1 AND tenant_id=$2::uuid RETURNING expires_at`, existingID, tenantID).Scan(&refreshedExpiresAt); err != nil {
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
		`INSERT INTO editor_sessions (tenant_id, document_id, user_id, expires_at, last_acknowledged_revision_id, status)
		 VALUES ($1, $2, $3, now() + interval '5 minutes', $4, 'active') RETURNING id, expires_at`,
		tenantID, docID, userID, curRev,
	).Scan(&newID, &newExpiresAt); err != nil {
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, `UPDATE documents SET active_session_id=$1, updated_at=now() WHERE id=$2`, newID, docID); err != nil {
		return nil, err
	}

	return &domain.Session{ID: newID, DocumentID: docID, UserID: userID, LastAcknowledgedRevisionID: curRev, Status: domain.SessionActive, ExpiresAt: newExpiresAt}, tx.Commit()
}

func (r *Repository) HeartbeatSession(ctx context.Context, tenantID, sessionID, userID string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE editor_sessions SET expires_at = now() + interval '5 minutes'
		 WHERE id=$1 AND tenant_id=$2::uuid AND user_id=$3 AND status='active'`, sessionID, tenantID, userID)
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
	ctx = authz.WithCapCache(ctx)
	if err := authz.SeedTxIdentity(ctx, tx, tenantID, userID); err != nil {
		return err
	}
	docArea, err := loadDocumentAreaBySession(ctx, tx, tenantID, sessionID)
	if err != nil {
		return err
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), docArea); err != nil {
		return fmt.Errorf("release session: authz check: %w", err)
	}
	res, err := tx.ExecContext(ctx,
		`UPDATE editor_sessions SET status='released', released_at=now()
		 WHERE id=$1 AND tenant_id=$2::uuid AND user_id=$3 AND status='active'`, sessionID, tenantID, userID)
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
	ctx = authz.WithCapCache(ctx)
	if err := authz.SeedTxIdentity(ctx, tx, tenantID, adminID); err != nil {
		return err
	}
	docArea, err := loadDocumentAreaBySession(ctx, tx, tenantID, sessionID)
	if err != nil {
		return err
	}
	// ADR 0022 Phase 11 (F4): force-release is an administrative session takeover
	// (releasing ANOTHER user's editor lock), so tier-1 gates it on membership.manage
	// (apps/api/cmd/metaldocs-api/permissions.go — /session/force-release). Tier-2
	// now agrees: the area-scoped check uses CapMembershipManage (not CapDocumentEdit)
	// so both tiers answer one coherent question — "is the actor an admin for this
	// document's area?" — and the destructive takeover stays narrowed to area admins
	// (area_admin in-area + system_admin via bypass), never every document editor.
	// membership.manage is area-grade: docArea is fail-closed "" on a missing row,
	// which denies non-system actors.
	if err := authz.Require(ctx, tx, string(iamdomain.CapMembershipManage), docArea); err != nil {
		return fmt.Errorf("force release session: authz check: %w", err)
	}
	res, err := tx.ExecContext(ctx,
		`UPDATE editor_sessions SET status='force_released', released_at=now()
		 WHERE id=$1 AND tenant_id=$2::uuid AND status='active'`, sessionID, tenantID)
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

// ForceReleaseSessionTx performs the force-release using the caller-owned tx.
// The caller is responsible for authz GUC setup, commit, and rollback.
func (r *Repository) ForceReleaseSessionTx(ctx context.Context, tx db.Tx, tenantID, adminID, sessionID string) error {
	ctx = authz.WithCapCache(ctx)
	docArea, err := loadDocumentAreaBySession(ctx, tx, tenantID, sessionID)
	if err != nil {
		return err
	}
	if err := authz.Require(ctx, mustSQLTx(tx), string(iamdomain.CapMembershipManage), docArea); err != nil {
		return fmt.Errorf("force release session: authz check: %w", err)
	}
	res, err := tx.ExecContext(ctx,
		`UPDATE editor_sessions SET status='force_released', released_at=now()
		 WHERE id=$1 AND tenant_id=$2::uuid AND status='active'`, sessionID, tenantID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrSessionInactive
	}
	_, err = tx.ExecContext(ctx, `UPDATE documents SET active_session_id=NULL, updated_at=now() WHERE active_session_id=$1`, sessionID)
	return err
}

// MarkArchivedTx sets archived_at using the caller-owned tx.
// The caller is responsible for authz GUC setup, commit, and rollback.
func (r *Repository) MarkArchivedTx(ctx context.Context, tx db.Tx, tenantID, docID, actorID string) error {
	ctx = authz.WithCapCache(ctx)
	docArea, err := loadDocumentArea(ctx, tx, tenantID, docID)
	if err != nil {
		return err
	}
	if err := authz.Require(ctx, mustSQLTx(tx), string(iamdomain.CapDocumentEdit), docArea); err != nil {
		return fmt.Errorf("mark archived: authz check: %w", err)
	}
	res, err := tx.ExecContext(ctx, `
		UPDATE public.documents
		   SET archived_at = now(), updated_at = now()
		 WHERE tenant_id = $1 AND id = $2 AND archived_at IS NULL`,
		tenantID, docID)
	if err != nil {
		return fmt.Errorf("mark archived: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("mark archived rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("mark archived: not found or already in target state")
	}
	return nil
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
		WITH batch AS (
			SELECT id
			FROM editor_sessions
			WHERE status='active' AND expires_at < $1
			ORDER BY expires_at, id
			LIMIT 500
		),
		expired AS (
			UPDATE editor_sessions SET status='expired'
			WHERE id IN (SELECT id FROM batch)
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

func (r *Repository) PresignReserve(ctx context.Context, tenantID, sessionID, userID, docID, baseRevisionID, contentHash, storageKey string, expiresAt time.Time) (pendingID string, err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	// Verify session active, holder matches, and session points at base.
	var sessUser, sessDoc, sessAck, sessStatus string
	err = tx.QueryRowContext(ctx,
		`SELECT user_id::text, document_id::text, last_acknowledged_revision_id::text, status
		 FROM editor_sessions WHERE id=$1 AND tenant_id=$2::uuid FOR UPDATE`, sessionID, tenantID,
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
func (r *Repository) GetPendingForCommit(ctx context.Context, tenantID, pendingID string) (*PendingCommitMeta, error) {
	var m PendingCommitMeta
	err := r.db.QueryRowContext(ctx,
		`SELECT p.session_id::text, p.document_id::text, p.base_revision_id::text,
		        p.content_hash, p.storage_key, p.expires_at, p.consumed_at
		 FROM autosave_pending_uploads p
		 JOIN documents d ON d.id = p.document_id
		 WHERE p.id=$1 AND d.tenant_id=$2::uuid`, pendingID, tenantID,
	).Scan(&m.SessionID, &m.DocumentID, &m.BaseRevisionID,
		&m.ExpectedContentHash, &m.StorageKey, &m.ExpiresAt, &m.ConsumedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrPendingNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get pending for commit: %w", err)
	}
	return &m, nil
}

// CommitUpload encodes every DB-level rejection branch below as explicit errors.
// Callers translate each to the matching HTTP status (404/409/410/422 per spec).
// Server-authoritative content_hash verification happens in Service BEFORE this
// method is called; `serverComputedHash` is the hash the service streamed from
// S3 and therefore trusted. CommitUpload still compares it to pending.content_hash
// under FOR UPDATE to catch TOCTOU races.
func (r *Repository) CommitUpload(ctx context.Context, tenantID, sessionID, userID, docID, pendingID, serverComputedHash string, formDataSnapshot []byte, fileSizeBytes int64, pageCount *int, pageCountSource *string) (*CommitResult, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	ctx = authz.WithCapCache(ctx)
	if err := authz.SeedTxIdentity(ctx, tx, tenantID, userID); err != nil {
		return nil, err
	}
	docArea, err := loadDocumentArea(ctx, tx, tenantID, docID)
	if err != nil {
		return nil, err
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), docArea); err != nil {
		return nil, fmt.Errorf("commit autosave: authz check: %w", err)
	}

	// Lock pending row.
	var p domain.PendingUpload
	err = tx.QueryRowContext(ctx,
		`SELECT p.id::text, p.session_id::text, p.document_id::text, p.base_revision_id::text, p.content_hash,
		        p.storage_key, p.expires_at, p.consumed_at
		 FROM autosave_pending_uploads p
		 JOIN documents d ON d.id = p.document_id
		 WHERE p.id=$1 AND d.tenant_id=$2::uuid FOR UPDATE`, pendingID, tenantID,
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
		 FROM editor_sessions WHERE id=$1 AND tenant_id=$2::uuid FOR UPDATE`, sessionID, tenantID,
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

	// form_data_snapshot is OPTIONAL in the contract (commitDocumentAutosave in
	// api/openapi/v1/openapi.yaml): an autosave that only changes the artifact
	// carries no form data. Partial-update semantics — absent means "leave the
	// form data as it is", so the effective snapshot is the document's current
	// form_data_json. Writing the nil through instead would hit the
	// documents.form_data_json NOT NULL constraint (a contract-conforming
	// request turning into a 500), and substituting `{}` would wipe real user
	// data. The same preserved value is stamped on the new revision so revision
	// history and checkpoint restore (which copies form_data_snapshot back into
	// documents.form_data_json) stay coherent. Race-safe: the documents row is
	// already held FOR UPDATE by the pending-upload lock above, so nothing can
	// change form_data_json between this read and the UPDATE below.
	effectiveFormData := formDataSnapshot
	if len(effectiveFormData) == 0 {
		if err := tx.QueryRowContext(ctx,
			`SELECT form_data_json FROM documents WHERE id=$1 AND tenant_id=$2::uuid`, docID, tenantID,
		).Scan(&effectiveFormData); err != nil {
			return nil, fmt.Errorf("preserve current form data: %w", err)
		}
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
		docID, p.BaseRevisionID, sessionID, p.StorageKey, p.ContentHash, effectiveFormData, fileSizeBytes, pageCount, pageCountSource,
	).Scan(&revID, &revNum, &committedFileSize, &committedPageCount, &committedPageCountSource); err != nil {
		return nil, fmt.Errorf("insert revision: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE documents SET current_revision_id=$1, form_data_json=$2, updated_at=now() WHERE id=$3 AND tenant_id=$4::uuid`,
		revID, effectiveFormData, docID, tenantID,
	); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE editor_sessions SET last_acknowledged_revision_id=$1 WHERE id=$2 AND tenant_id=$3::uuid`, revID, sessionID, tenantID,
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

func (r *Repository) SyncCurrentRevisionArtifactMetadata(ctx context.Context, tenantID, sessionID, userID, docID string, fileSizeBytes int64, pageCount *int, pageCountSource *string) (*CommitResult, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	ctx = authz.WithCapCache(ctx)
	if err := authz.SeedTxIdentity(ctx, tx, tenantID, userID); err != nil {
		return nil, err
	}
	docArea, err := loadDocumentArea(ctx, tx, tenantID, docID)
	if err != nil {
		return nil, err
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), docArea); err != nil {
		return nil, fmt.Errorf("sync artifact metadata: authz check: %w", err)
	}

	var currentRevisionID, activeSessionID, status string
	if err := tx.QueryRowContext(ctx,
		`SELECT coalesce(current_revision_id::text,''), coalesce(active_session_id::text,''), status
		   FROM documents
		  WHERE id = $1 AND tenant_id = $2
		  FOR UPDATE`,
		docID, tenantID,
	).Scan(&currentRevisionID, &activeSessionID, &status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	if status != string(domain.DocStatusDraft) {
		return nil, domain.ErrInvalidStateTransition
	}
	if activeSessionID == "" {
		return nil, domain.ErrSessionInactive
	}
	if activeSessionID != sessionID {
		return nil, domain.ErrSessionNotHolder
	}

	var sessUser, sessStatus string
	if err := tx.QueryRowContext(ctx,
		`SELECT user_id::text, status
		   FROM editor_sessions
		  WHERE id = $1 AND tenant_id = $2::uuid
		  FOR UPDATE`,
		sessionID, tenantID,
	).Scan(&sessUser, &sessStatus); err != nil {
		return nil, err
	}
	if sessStatus != string(domain.SessionActive) {
		return nil, domain.ErrSessionInactive
	}
	if sessUser != userID {
		return nil, domain.ErrSessionNotHolder
	}

	var committedFileSize sql.NullInt64
	var committedPageCount sql.NullInt64
	var committedPageCountSource sql.NullString
	if err := tx.QueryRowContext(ctx,
		`UPDATE document_revisions
		    SET file_size_bytes = $1,
		        page_count = $2,
		        page_count_source = $3
		  WHERE id = $4
		    AND document_id = $5
		RETURNING file_size_bytes, page_count, page_count_source`,
		fileSizeBytes, pageCount, pageCountSource, currentRevisionID, docID,
	).Scan(&committedFileSize, &committedPageCount, &committedPageCountSource); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	res := &CommitResult{RevisionID: currentRevisionID}
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
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	if err := authz.BypassSystem(ctx, tx); err != nil {
		return 0, err
	}

	res, err := tx.ExecContext(ctx,
		`WITH batch AS (
			SELECT id
			FROM autosave_pending_uploads
			WHERE expires_at < $1 AND consumed_at IS NULL
			ORDER BY expires_at, id
			LIMIT 500
		)
		DELETE FROM autosave_pending_uploads p
		USING batch
		WHERE p.id = batch.id`,
		olderThan)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return int(n), nil
}

func (r *Repository) CreateCheckpoint(ctx context.Context, tenantID, docID, actorUserID, label string) (*domain.Checkpoint, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Tier-2 authz: document.edit is area-grade (ADR 0022); mirrors CommitUpload.
	// RW tx: the F8 bypass audit may INSERT (ADR 0022 Phase 11).
	ctx = authz.WithCapCache(ctx)
	if err := authz.SeedTxIdentity(ctx, tx, tenantID, actorUserID); err != nil {
		return nil, err
	}
	docArea, err := loadDocumentArea(ctx, tx, tenantID, docID)
	if err != nil {
		return nil, err
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), docArea); err != nil {
		return nil, fmt.Errorf("create checkpoint: authz check: %w", err)
	}

	var revID string
	if err := tx.QueryRowContext(ctx,
		`SELECT current_revision_id::text FROM documents WHERE id=$1 AND tenant_id=$2::uuid FOR UPDATE`, docID, tenantID,
	).Scan(&revID); err != nil {
		return nil, err
	}
	if revID == "" {
		return nil, fmt.Errorf("document has no current revision")
	}

	var nextVer int
	if err := tx.QueryRowContext(ctx,
		`SELECT coalesce(max(cp.version_num),0)+1
		   FROM document_checkpoints cp
		   JOIN documents d ON d.id = cp.document_id
		  WHERE cp.document_id=$1 AND d.tenant_id=$2::uuid`, docID, tenantID,
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

func (r *Repository) ListCheckpoints(ctx context.Context, tenantID, docID string) ([]domain.Checkpoint, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT cp.id::text, cp.document_id::text, cp.revision_id::text, cp.version_num, coalesce(cp.label,''), cp.created_at, cp.created_by::text
		 FROM document_checkpoints cp
		 JOIN documents d ON d.id = cp.document_id
		 WHERE cp.document_id=$1 AND d.tenant_id=$2::uuid ORDER BY cp.version_num DESC`, docID, tenantID)
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

func (r *Repository) GetRevision(ctx context.Context, tenantID, docID, revID string) (*domain.Revision, error) {
	var rv domain.Revision
	err := r.db.QueryRowContext(ctx,
		`SELECT r.id::text, r.document_id::text, r.revision_num, coalesce(r.parent_revision_id::text,''), r.session_id::text, r.storage_key, r.content_hash, r.form_data_snapshot, r.created_at
		 FROM document_revisions r
		 JOIN documents d ON d.id = r.document_id
		 WHERE r.id=$1 AND r.document_id=$2 AND d.tenant_id=$3::uuid`, revID, docID, tenantID,
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
func (r *Repository) RestoreCheckpoint(ctx context.Context, tenantID, docID, actorUserID string, versionNum int) (*RestoreResult, error) {
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
		 FROM documents WHERE id=$1 AND tenant_id=$2::uuid FOR UPDATE`, docID, tenantID,
	).Scan(&sessID, &curRev); err != nil {
		return nil, err
	}
	if sessID == "" {
		return nil, domain.ErrSessionInactive
	}
	if err := tx.QueryRowContext(ctx,
		`SELECT user_id::text, status FROM editor_sessions WHERE id=$1 AND tenant_id=$2::uuid FOR UPDATE`, sessID, tenantID,
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
		 JOIN documents d ON d.id = cp.document_id
		 WHERE cp.document_id=$1 AND cp.version_num=$2 AND d.tenant_id=$3::uuid`, docID, versionNum, tenantID,
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
			`UPDATE documents SET current_revision_id=$1, form_data_json=$2, updated_at=now() WHERE id=$3 AND tenant_id=$4::uuid`,
			newRevID, cpFormData, docID, tenantID,
		); err != nil {
			return nil, err
		}
		if _, err := tx.ExecContext(ctx,
			`UPDATE editor_sessions SET last_acknowledged_revision_id=$1 WHERE id=$2 AND tenant_id=$3::uuid`, newRevID, sessID, tenantID,
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
// tenantID. Used by handler-level ownership defense-in-depth (ADR 0022 Phase 9).
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
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	ctx = authz.WithCapCache(ctx)
	if err := authz.SeedTxIdentity(ctx, tx, tenantID, actorID); err != nil {
		return err
	}
	docArea, err := loadDocumentArea(ctx, tx, tenantID, docID)
	if err != nil {
		return err
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), docArea); err != nil {
		return fmt.Errorf("mark archived: authz check: %w", err)
	}

	res, err := tx.ExecContext(ctx, `
		UPDATE public.documents
		   SET archived_at = now(), updated_at = now()
		 WHERE tenant_id = $1 AND id = $2 AND archived_at IS NULL`,
		tenantID, docID)
	if err != nil {
		return fmt.Errorf("mark archived: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("mark archived rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("mark archived: not found or already in target state")
	}
	return tx.Commit()
}

// Unarchive clears archived_at, restoring the document to active queries.
func (r *Repository) Unarchive(ctx context.Context, tenantID, docID, actorID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	ctx = authz.WithCapCache(ctx)
	if err := authz.SeedTxIdentity(ctx, tx, tenantID, actorID); err != nil {
		return err
	}
	docArea, err := loadDocumentArea(ctx, tx, tenantID, docID)
	if err != nil {
		return err
	}
	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), docArea); err != nil {
		return fmt.Errorf("unarchive: authz check: %w", err)
	}

	res, err := tx.ExecContext(ctx, `
		UPDATE public.documents
		   SET archived_at = NULL, updated_at = now()
		 WHERE tenant_id = $1 AND id = $2 AND archived_at IS NOT NULL`,
		tenantID, docID)
	if err != nil {
		return fmt.Errorf("unarchive: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("unarchive rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("unarchive: not found or already in target state")
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

// IsEligibleApprover reports whether actorUserID is listed in the
// eligible_actor_ids of the currently-open (pending) approval stage for the
// document's active (in_progress) approval instance.
func (r *Repository) IsEligibleApprover(ctx context.Context, tenantID, documentID, actorUserID string) (bool, error) {
	const q = `
		SELECT EXISTS (
			SELECT 1
			  FROM approval_stage_instances asi
			  JOIN approval_instances ai ON ai.id = asi.approval_instance_id
			 WHERE ai.tenant_id   = $1::uuid
			   AND ai.document_id = $2::uuid
			   AND ai.status      = 'in_progress'
			   AND asi.status     = 'pending'
			   AND asi.eligible_actor_ids @> to_jsonb(ARRAY[$3::text])
		)`
	var ok bool
	if err := r.db.QueryRowContext(ctx, q, tenantID, documentID, actorUserID).Scan(&ok); err != nil {
		return false, fmt.Errorf("documents: is-eligible-approver: %w", err)
	}
	return ok, nil
}
