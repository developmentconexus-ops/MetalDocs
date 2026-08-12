// Package infrastructure is the controlled-documents module's
// Postgres-backed implementation of the domain repository ports
// (ControlledDocumentRepository, SequenceAllocator) plus the cross-module
// read adapters (CDFieldReaderPG, TaxonomyProfileReader,
// TaxonomyAreaReader). Every write path calls authz.Require before
// mutating a row, pairing with the trg_require_cap_asserted DB tripwire;
// visibility reads/writes cover the controlled_document_area_grants and
// controlled_document_user_grants side tables for restricted-scope
// documents.
package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	documentsdomain "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/pagination"
	"metaldocs/internal/platform/sqlescape"
)

// PostgresControlledDocumentRepository is the Postgres-backed
// domain.ControlledDocumentRepository implementation, holding the
// metaldocs.controlled_documents table plus its visibility grant side
// tables.
type PostgresControlledDocumentRepository struct {
	db     *sql.DB
	runner db.TxRunner
	// activeInstance is the documents-owned read-port for the active-instance
	// projection (M2/F2.2, ADR-0039 D3(b)). CD no longer reads documents/
	// document_revisions/approval_instances directly.
	activeInstance documentsdomain.ActiveInstanceReader
}

// NewPostgresControlledDocumentRepository builds a
// PostgresControlledDocumentRepository backed by db. activeInstance
// defaults to documentsdomain.NoopActiveInstanceReader when nil.
func NewPostgresControlledDocumentRepository(database *sql.DB, activeInstance documentsdomain.ActiveInstanceReader) *PostgresControlledDocumentRepository {
	if activeInstance == nil {
		activeInstance = documentsdomain.NoopActiveInstanceReader{}
	}
	return &PostgresControlledDocumentRepository{db: database, runner: db.NewTxRunner(database), activeInstance: activeInstance}
}

// GetByID returns the controlled document by (tenantID, id), hydrating
// visibility grants when the document's scope is restricted. Returns
// ErrCDNotFound when no matching row exists.
func (r *PostgresControlledDocumentRepository) GetByID(ctx context.Context, tenantID, id string) (*controlleddocumentsdomain.ControlledDocument, error) {
	const q = `
SELECT id::text, tenant_id::text, profile_code, process_area_code, department_code,
       code, sequence_num, title, owner_user_id, coalesce(override_template_version_id::text, ''),
       visibility_scope, status, created_at, updated_at
FROM controlled_documents
WHERE tenant_id = $1 AND id = $2`
	doc, err := scanControlledDocument(r.db.QueryRowContext(ctx, q, tenantID, id))
	if err != nil {
		return nil, fmt.Errorf("get controlled document by id: %w", err)
	}
	if doc.Visibility.Scope == controlleddocumentsdomain.VisibilityScopeRestricted {
		vis, err := r.loadVisibilityGrants(ctx, tenantID, doc.ID)
		if err != nil {
			return nil, fmt.Errorf("load controlled document visibility grants: %w", err)
		}
		doc.Visibility = vis
	}
	return doc, nil
}

// GetByCode returns the controlled document by (tenantID, profileCode,
// code), hydrating visibility grants when the document's scope is
// restricted. Returns ErrCDNotFound when no matching row exists.
func (r *PostgresControlledDocumentRepository) GetByCode(ctx context.Context, tenantID, profileCode, code string) (*controlleddocumentsdomain.ControlledDocument, error) {
	const q = `
SELECT id::text, tenant_id::text, profile_code, process_area_code, department_code,
       code, sequence_num, title, owner_user_id, coalesce(override_template_version_id::text, ''),
       visibility_scope, status, created_at, updated_at
FROM controlled_documents
WHERE tenant_id = $1 AND profile_code = $2 AND code = $3`
	doc, err := scanControlledDocument(r.db.QueryRowContext(ctx, q, tenantID, profileCode, code))
	if err != nil {
		return nil, fmt.Errorf("get controlled document by code: %w", err)
	}
	if doc.Visibility.Scope == controlleddocumentsdomain.VisibilityScopeRestricted {
		vis, err := r.loadVisibilityGrants(ctx, tenantID, doc.ID)
		if err != nil {
			return nil, fmt.Errorf("load controlled document visibility grants: %w", err)
		}
		doc.Visibility = vis
	}
	return doc, nil
}

// CodeExists reports whether a controlled document with (tenantID,
// profileCode, code) already exists, regardless of status.
func (r *PostgresControlledDocumentRepository) CodeExists(ctx context.Context, tenantID, profileCode, code string) (bool, error) {
	const q = `SELECT EXISTS(
		SELECT 1 FROM controlled_documents WHERE tenant_id = $1 AND profile_code = $2 AND code = $3
	)`
	var exists bool
	if err := r.db.QueryRowContext(ctx, q, tenantID, profileCode, code).Scan(&exists); err != nil {
		return false, fmt.Errorf("check controlled document code exists: %w", err)
	}
	return exists, nil
}

// appendListFilterClauses appends List's optional WHERE clauses (profile,
// area, department, owner, status, query, and the actor visibility gate) to
// q in the same order and with the same $N placeholder bookkeeping as
// before extraction, returning the extended query, args, and the next free
// placeholder index. Extracted from List; behavior unchanged.
func appendListFilterClauses(q string, args []any, idx int, filter controlleddocumentsdomain.CDFilter) (string, []any, int) {
	if filter.ProfileCode != nil {
		q += fmt.Sprintf(" AND profile_code = $%d", idx)
		args = append(args, *filter.ProfileCode)
		idx++
	}
	if filter.ProcessAreaCode != nil {
		q += fmt.Sprintf(" AND process_area_code = $%d", idx)
		args = append(args, *filter.ProcessAreaCode)
		idx++
	}
	if len(filter.UserAreaCodes) > 0 {
		q += fmt.Sprintf(" AND process_area_code = ANY($%d)", idx)
		args = append(args, pgtype.FlatArray[string](filter.UserAreaCodes))
		idx++
	}
	if filter.DepartmentCode != nil {
		q += fmt.Sprintf(" AND department_code = $%d", idx)
		args = append(args, *filter.DepartmentCode)
		idx++
	}
	if filter.OwnerUserID != nil {
		q += fmt.Sprintf(" AND owner_user_id = $%d", idx)
		args = append(args, *filter.OwnerUserID)
		idx++
	}
	if filter.Status != nil {
		q += fmt.Sprintf(" AND status = $%d", idx)
		args = append(args, *filter.Status)
		idx++
	}
	if filter.Query != nil && strings.TrimSpace(*filter.Query) != "" {
		// GIN trigram indexes idx_controlled_documents_code_trgm / _title_trgm (migration 0239) accelerate this ILIKE.
		q += fmt.Sprintf(" AND (code ILIKE $%d ESCAPE '\\' OR title ILIKE $%d ESCAPE '\\')", idx, idx)
		args = append(args, "%"+sqlescape.LikeEscape(strings.TrimSpace(*filter.Query))+"%")
		idx++
	}
	if filter.ActorUserID != nil {
		q += fmt.Sprintf(`
 AND (
      visibility_scope = 'company'
   OR owner_user_id = $%d
   OR (
        visibility_scope = 'restricted'
        AND (
             EXISTS (
               SELECT 1
                 FROM controlled_document_area_grants cdag
                WHERE cdag.tenant_id = controlled_documents.tenant_id
                  AND cdag.controlled_document_id = controlled_documents.id
                  AND EXISTS (
                    SELECT 1
                      FROM metaldocs.v_active_user_areas upa
                     WHERE upa.tenant_id = controlled_documents.tenant_id
                       AND upa.user_id = $%d
                       AND upa.area_code = cdag.area_code
                  )
             )
             OR EXISTS (
               SELECT 1
                 FROM controlled_document_user_grants cdug
                WHERE cdug.tenant_id = controlled_documents.tenant_id
                  AND cdug.controlled_document_id = controlled_documents.id
                  AND cdug.user_id = $%d
             )
        )
      )
 )`, idx, idx, idx)
		args = append(args, *filter.ActorUserID)
		idx++
	}
	return q, args, idx
}

// List returns up to filter.Limit controlled documents ordered by
// (created_at DESC, id DESC) using an opaque keyset cursor (FD-2). hasMore
// reports whether a further page exists; the caller builds the next cursor from
// the last returned document's (CreatedAt, ID).
func (r *PostgresControlledDocumentRepository) List(ctx context.Context, tenantID string, filter controlleddocumentsdomain.CDFilter) (items []controlleddocumentsdomain.ControlledDocument, hasMore bool, err error) {
	// The area-grant EXISTS subquery reads iam's PUBLISHED active-membership view
	// metaldocs.v_active_user_areas (M3/F3.2, ADR-0039 D3a) — the view encodes the
	// active-now predicate `effective_to IS NULL` (ADR 0037 D1), so CD names no
	// iam base table and re-derives no temporal predicate.
	q := `
SELECT id::text, tenant_id::text, profile_code, process_area_code, department_code,
       code, sequence_num, title, owner_user_id, coalesce(override_template_version_id::text, ''),
       visibility_scope, status, created_at, updated_at
FROM controlled_documents
WHERE tenant_id = $1`
	args := []any{tenantID}
	idx := 2

	q, args, idx = appendListFilterClauses(q, args, idx, filter)

	cursorTS, cursorID, err := pagination.DecodeCursor(filter.Cursor)
	if err != nil {
		return nil, false, err
	}
	if cursorTS != "" {
		q += fmt.Sprintf(" AND (created_at, id) < ($%d::timestamptz, $%d::uuid)", idx, idx+1)
		args = append(args, cursorTS, cursorID)
		idx += 2
	}

	limit := pagination.ClampLimit(filter.Limit)
	q += fmt.Sprintf(" ORDER BY created_at DESC, id DESC LIMIT $%d", idx) // #nosec G202 -- idx is a computed placeholder index (int), not a value; the actual limit value is bound via args below.
	args = append(args, limit+1)                                          // +1 probe row to detect hasMore

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, false, fmt.Errorf("list controlled documents: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]controlleddocumentsdomain.ControlledDocument, 0, limit+1)
	for rows.Next() {
		doc, err := scanControlledDocument(rows)
		if err != nil {
			return nil, false, fmt.Errorf("scan controlled document list row: %w", err)
		}
		out = append(out, *doc)
	}
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("iterate controlled document list rows: %w", err)
	}
	if len(out) > limit {
		out = out[:limit]
		hasMore = true
	}
	if err := r.hydrateVisibilityGrants(ctx, tenantID, out); err != nil {
		return nil, false, fmt.Errorf("hydrate controlled document visibility grants: %w", err)
	}
	return out, hasMore, nil
}

func (r *PostgresControlledDocumentRepository) loadVisibilityGrants(ctx context.Context, tenantID, controlledDocumentID string) (controlleddocumentsdomain.Visibility, error) {
	areas, err := r.loadAreaGrants(ctx, tenantID, controlledDocumentID)
	if err != nil {
		return controlleddocumentsdomain.Visibility{}, fmt.Errorf("load controlled document area grants: %w", err)
	}
	users, err := r.loadUserGrants(ctx, tenantID, controlledDocumentID)
	if err != nil {
		return controlleddocumentsdomain.Visibility{}, fmt.Errorf("load controlled document user grants: %w", err)
	}
	return controlleddocumentsdomain.NewVisibility(string(controlleddocumentsdomain.VisibilityScopeRestricted), areas, users, "")
}

func (r *PostgresControlledDocumentRepository) loadAreaGrants(ctx context.Context, tenantID, controlledDocumentID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT area_code
FROM controlled_document_area_grants
WHERE tenant_id = $1 AND controlled_document_id = $2
ORDER BY area_code`, tenantID, controlledDocumentID)
	if err != nil {
		return nil, fmt.Errorf("query controlled document area grants: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := []string{}
	for rows.Next() {
		var areaCode string
		if err := rows.Scan(&areaCode); err != nil {
			return nil, fmt.Errorf("scan controlled document area grant: %w", err)
		}
		out = append(out, areaCode)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate controlled document area grants: %w", err)
	}
	return out, nil
}

func (r *PostgresControlledDocumentRepository) loadUserGrants(ctx context.Context, tenantID, controlledDocumentID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT user_id
FROM controlled_document_user_grants
WHERE tenant_id = $1 AND controlled_document_id = $2
ORDER BY user_id`, tenantID, controlledDocumentID)
	if err != nil {
		return nil, fmt.Errorf("query controlled document user grants: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := []string{}
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("scan controlled document user grant: %w", err)
		}
		out = append(out, userID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate controlled document user grants: %w", err)
	}
	return out, nil
}

func (r *PostgresControlledDocumentRepository) hydrateVisibilityGrants(ctx context.Context, tenantID string, docs []controlleddocumentsdomain.ControlledDocument) error {
	ids := make([]string, 0, len(docs))
	indexByID := make(map[string]int, len(docs))
	for i := range docs {
		if docs[i].Visibility.Scope != controlleddocumentsdomain.VisibilityScopeRestricted {
			continue
		}
		ids = append(ids, docs[i].ID)
		indexByID[docs[i].ID] = i
		docs[i].Visibility.AreaCodes = []string{}
		docs[i].Visibility.UserIDs = []string{}
	}
	if len(ids) == 0 {
		return nil
	}

	areaRows, err := r.db.QueryContext(ctx, `
SELECT controlled_document_id::text, area_code
FROM controlled_document_area_grants
WHERE tenant_id = $1 AND controlled_document_id = ANY($2)
ORDER BY controlled_document_id, area_code`, tenantID, pgtype.FlatArray[string](ids))
	if err != nil {
		return fmt.Errorf("query controlled document area grants: %w", err)
	}
	defer func() { _ = areaRows.Close() }()
	for areaRows.Next() {
		var docID, areaCode string
		if err := areaRows.Scan(&docID, &areaCode); err != nil {
			return fmt.Errorf("scan controlled document area grant: %w", err)
		}
		if idx, ok := indexByID[docID]; ok {
			docs[idx].Visibility.AreaCodes = append(docs[idx].Visibility.AreaCodes, areaCode)
		}
	}
	if err := areaRows.Err(); err != nil {
		return fmt.Errorf("iterate controlled document area grants: %w", err)
	}

	userRows, err := r.db.QueryContext(ctx, `
SELECT controlled_document_id::text, user_id
FROM controlled_document_user_grants
WHERE tenant_id = $1 AND controlled_document_id = ANY($2)
ORDER BY controlled_document_id, user_id`, tenantID, pgtype.FlatArray[string](ids))
	if err != nil {
		return fmt.Errorf("query controlled document user grants: %w", err)
	}
	defer func() { _ = userRows.Close() }()
	for userRows.Next() {
		var docID, userID string
		if err := userRows.Scan(&docID, &userID); err != nil {
			return fmt.Errorf("scan controlled document user grant: %w", err)
		}
		if idx, ok := indexByID[docID]; ok {
			docs[idx].Visibility.UserIDs = append(docs[idx].Visibility.UserIDs, userID)
		}
	}
	if err := userRows.Err(); err != nil {
		return fmt.Errorf("iterate controlled document user grants: %w", err)
	}
	return nil
}

// Create persists doc in its own transaction, running the tier-2 authz
// check (CapControlledDocumentCreate against doc.ProcessAreaCode) before
// the insert. On success doc.ID is populated. Returns ErrCDCodeTaken (or
// ErrCDArchivedCodeReuse when the conflicting row is non-active) on a
// unique-code violation.
func (r *PostgresControlledDocumentRepository) Create(ctx context.Context, doc *controlleddocumentsdomain.ControlledDocument) error {
	if err := r.runner.Do(ctx, func(tx *sql.Tx) error {
		// ADR 0022 Phase 7: authorize CD create against its target process area.
		if err := authz.Require(ctx, tx, string(iamdomain.CapControlledDocumentCreate), doc.ProcessAreaCode); err != nil {
			return fmt.Errorf("registry: authz check Create: %w", err)
		}
		if err := r.createWithQueryer(ctx, tx, doc); err != nil {
			return fmt.Errorf("create controlled document with queryer: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("create controlled document tx: %w", err)
	}
	return nil
}

// CreateTx persists doc inside the caller's tx (ADR 0011 atomic
// CD+first-revision create). It performs no authz check of its own — the
// service layer is the mandatory gate for CD create (F-CD6); a redundant
// check here would double the DB round-trip with no correctness benefit.
func (r *PostgresControlledDocumentRepository) CreateTx(ctx context.Context, tx db.Tx, doc *controlleddocumentsdomain.ControlledDocument) error {
	if tx == nil {
		return errors.New("nil transaction")
	}
	// AuthZ for CD create is the service layer's mandatory gate (ADR 0022 Phase 7).
	// The service calls authz.Require(ctx, tx, CapControlledDocumentCreate, processAreaCode)
	// before invoking CreateTx on both the manual-code and auto-code paths; a
	// redundant Require here would double the DB round-trip on every create with
	// no correctness benefit (both always agree). The repository is the storage
	// layer only; capability enforcement must not be split across layers.
	return r.createWithQueryer(ctx, tx, doc)
}

func (r *PostgresControlledDocumentRepository) createWithQueryer(ctx context.Context, qr queryRower, doc *controlleddocumentsdomain.ControlledDocument) error {
	const insertQ = `
INSERT INTO controlled_documents
	(tenant_id, profile_code, process_area_code, department_code, code, sequence_num, title, owner_user_id,
	 override_template_version_id, visibility_scope, status, created_at, updated_at)
VALUES
	($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING id::text`
	var id string
	err := qr.QueryRowContext(
		ctx,
		insertQ,
		doc.TenantID,
		doc.ProfileCode,
		doc.ProcessAreaCode,
		stringPtrToNull(doc.DepartmentCode),
		doc.Code,
		intPtrToNull(doc.SequenceNum),
		doc.Title,
		doc.OwnerUserID,
		stringPtrToNull(doc.OverrideTemplateVersionID),
		doc.Visibility.Scope,
		doc.Status,
		doc.CreatedAt,
		doc.UpdatedAt,
	).Scan(&id)
	if err == nil {
		doc.ID = id
		return r.createVisibilityGrants(ctx, qr, doc)
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		existing, getErr := r.GetByCode(ctx, doc.TenantID, doc.ProfileCode, doc.Code)
		if getErr == nil && existing.Status != controlleddocumentsdomain.CDStatusActive {
			return controlleddocumentsdomain.ErrCDArchivedCodeReuse
		}
		return controlleddocumentsdomain.ErrCDCodeTaken
	}
	return fmt.Errorf("insert controlled document: %w", err)
}

func (r *PostgresControlledDocumentRepository) createVisibilityGrants(ctx context.Context, qr queryRower, doc *controlleddocumentsdomain.ControlledDocument) error {
	if doc.Visibility.Scope != controlleddocumentsdomain.VisibilityScopeRestricted {
		return nil
	}
	for _, areaCode := range doc.Visibility.AreaCodes {
		if _, err := qr.ExecContext(ctx, `
INSERT INTO controlled_document_area_grants (tenant_id, controlled_document_id, area_code)
VALUES ($1, $2::uuid, $3)
ON CONFLICT (tenant_id, controlled_document_id, area_code) DO NOTHING`, doc.TenantID, doc.ID, areaCode); err != nil {
			return fmt.Errorf("insert controlled document area grant: %w", err)
		}
	}
	for _, userID := range doc.Visibility.UserIDs {
		if _, err := qr.ExecContext(ctx, `
INSERT INTO controlled_document_user_grants (tenant_id, controlled_document_id, user_id)
VALUES ($1, $2::uuid, $3)
ON CONFLICT (tenant_id, controlled_document_id, user_id) DO NOTHING`, doc.TenantID, doc.ID, userID); err != nil {
			return fmt.Errorf("insert controlled document user grant: %w", err)
		}
	}
	return nil
}

// UpdateStatus updates the controlled document's status/updated_at
// off-tx. Returns ErrCDNotFound when no row matches (tenantID, id).
func (r *PostgresControlledDocumentRepository) UpdateStatus(ctx context.Context, tenantID, id string, status controlleddocumentsdomain.CDStatus, updatedAt time.Time) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE controlled_documents SET status = $1, updated_at = $2 WHERE tenant_id = $3 AND id = $4`,
		status, updatedAt, tenantID, id,
	)
	if err != nil {
		return fmt.Errorf("update controlled document status: %w", err)
	}
	n, rowsErr := res.RowsAffected()
	if rowsErr != nil {
		return fmt.Errorf("check controlled document status rows affected: %w", rowsErr)
	}
	if n == 0 {
		return controlleddocumentsdomain.ErrCDNotFound
	}
	return nil
}

// UpdateStatusTx updates the controlled document's status/updated_at
// inside the caller's tx (e.g. paired with a FOR UPDATE lock + authz
// check in the service's changeStatus). Returns ErrCDNotFound when no
// row matches (tenantID, id).
func (r *PostgresControlledDocumentRepository) UpdateStatusTx(ctx context.Context, tx db.Tx, tenantID, id string, status controlleddocumentsdomain.CDStatus, updatedAt time.Time) error {
	res, err := tx.ExecContext(ctx,
		`UPDATE controlled_documents SET status = $1, updated_at = $2 WHERE tenant_id = $3 AND id = $4`,
		status, updatedAt, tenantID, id,
	)
	if err != nil {
		return fmt.Errorf("update controlled document status in tx: %w", err)
	}
	n, rowsErr := res.RowsAffected()
	if rowsErr != nil {
		return fmt.Errorf("check controlled document status tx rows affected: %w", rowsErr)
	}
	if n == 0 {
		return controlleddocumentsdomain.ErrCDNotFound
	}
	return nil
}

// CanRead reports whether actorUserID may read the controlled document:
// true when the document is company-scoped, actorUserID is the owner, or
// (for restricted scope) actorUserID's active area membership or an
// explicit user grant covers it. This is the visibility-scope EXISTS
// check; it is distinct from — and runs before — the tier-2 capability
// check in callers like GetActiveInstance.
func (r *PostgresControlledDocumentRepository) CanRead(ctx context.Context, tenantID, controlledDocumentID, actorUserID string) (bool, error) {
	// The area-grant EXISTS subquery reads iam's PUBLISHED active-membership view
	// metaldocs.v_active_user_areas (M3/F3.2, ADR-0039 D3a) — the view encodes the
	// active-now predicate `effective_to IS NULL` (ADR 0037 D1), so CD names no
	// iam base table and re-derives no temporal predicate.
	const q = `
SELECT EXISTS (
  SELECT 1
    FROM controlled_documents cd
   WHERE cd.tenant_id = $1
     AND cd.id = $2::uuid
     AND (
          cd.visibility_scope = 'company'
       OR cd.owner_user_id = $3
       OR (
            cd.visibility_scope = 'restricted'
            AND (
                 EXISTS (
                   SELECT 1
                     FROM controlled_document_area_grants cdag
                    WHERE cdag.tenant_id = cd.tenant_id
                      AND cdag.controlled_document_id = cd.id
                      AND EXISTS (
                        SELECT 1
                          FROM metaldocs.v_active_user_areas upa
                         WHERE upa.tenant_id = cd.tenant_id
                           AND upa.user_id = $3
                           AND upa.area_code = cdag.area_code
                      )
                 )
                 OR EXISTS (
                   SELECT 1
                     FROM controlled_document_user_grants cdug
                    WHERE cdug.tenant_id = cd.tenant_id
                      AND cdug.controlled_document_id = cd.id
                      AND cdug.user_id = $3
                 )
            )
       )
     )
)`
	var exists bool
	if err := r.db.QueryRowContext(ctx, q, tenantID, controlledDocumentID, actorUserID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check controlled document read access: %w", err)
	}
	return exists, nil
}

// GetActiveInstance fetches the active document instance (plus published
// document id and, when under review, the in-progress approval instance id)
// for the given controlled document. Returns nil, nil when no active or
// published document exists for this controlled document.
func (r *PostgresControlledDocumentRepository) GetActiveInstance(ctx context.Context, tenantID, controlledDocumentID string) (*controlleddocumentsdomain.ActiveDocumentInstance, error) {
	// B2/B3/B4 ported (M2/F2.2, ADR-0039 D3(b)): the active-instance projection
	// over documents/document_revisions/approval_instances is resolved through the
	// documents-owned ActiveInstanceReader port; CD maps the view 1:1 onto its own
	// ActiveDocumentInstance and no longer names those base tables in its own SQL.
	view, err := r.activeInstance.ActiveInstanceForControlledDocument(ctx, tenantID, controlledDocumentID)
	if err != nil {
		return nil, fmt.Errorf("get active document instance: %w", err)
	}
	if view == nil {
		return nil, nil
	}
	return &controlleddocumentsdomain.ActiveDocumentInstance{
		DocumentID:          view.DocumentID,
		ContentHash:         view.ContentHash,
		RevisionVersion:     view.RevisionVersion,
		ApprovalState:       view.Status,
		PublishedDocumentID: view.PublishedDocumentID,
		ApprovalInstanceID:  view.ApprovalInstanceID,
	}, nil
}

// PostgresSequenceAllocator is the Postgres-backed
// domain.SequenceAllocator implementation, backed by the
// cd_sequence_counters table (one row per tenant/profile/area).
type PostgresSequenceAllocator struct {
	db *sql.DB
}

// NewPostgresSequenceAllocator builds a PostgresSequenceAllocator backed
// by db.
func NewPostgresSequenceAllocator(db *sql.DB) *PostgresSequenceAllocator {
	return &PostgresSequenceAllocator{db: db}
}

// EnsureCounter initialises a sequence counter for the given tenant/profile/area combination.
func (a *PostgresSequenceAllocator) EnsureCounter(ctx context.Context, tenantID, profileCode, areaCode string) error {
	return a.ensureCounterViaExec(ctx, a.db, tenantID, profileCode, areaCode)
}

func (a *PostgresSequenceAllocator) ensureCounterViaExec(ctx context.Context, exec db.Tx, tenantID, profileCode, areaCode string) error {
	_, err := exec.ExecContext(ctx, `
		INSERT INTO cd_sequence_counters (tenant_id, profile_code, process_area_code, next_seq)
		VALUES ($1, $2, $3, 1)
		ON CONFLICT (tenant_id, profile_code, process_area_code) DO NOTHING`,
		tenantID, profileCode, areaCode,
	)
	if err != nil {
		return fmt.Errorf("ensure controlled document sequence counter: %w", err)
	}
	return nil
}

// Peek returns the current next_seq value without incrementing it.
// Returns 1 if no counter exists yet for the given combination.
func (a *PostgresSequenceAllocator) Peek(ctx context.Context, tenantID, profileCode, areaCode string) (int, error) {
	var next int
	err := a.db.QueryRowContext(ctx, `
		SELECT next_seq
		FROM cd_sequence_counters
		WHERE tenant_id = $1 AND profile_code = $2 AND process_area_code = $3`,
		tenantID, profileCode, areaCode,
	).Scan(&next)
	if errors.Is(err, sql.ErrNoRows) {
		return 1, nil
	}
	if err != nil {
		return 0, fmt.Errorf("peek controlled document sequence counter: %w", err)
	}
	return next, nil
}

// NextAndIncrement atomically increments and returns the next sequence number.
// tx must be non-nil: the FOR UPDATE row lock only serializes concurrent
// allocations when the SELECT…UPDATE run in the SAME transaction as the caller's
// insert. A nil tx would silently autocommit each statement, dropping the lock
// the instant this method returns and reopening the duplicate-sequence race the
// lock exists to close (db.Tx contract; the nodualmode linter doesn't reach
// infrastructure/).
func (a *PostgresSequenceAllocator) NextAndIncrement(ctx context.Context, tx db.Tx, tenantID, profileCode, areaCode string) (int, error) {
	if tx == nil {
		return 0, fmt.Errorf("controlled_documents: NextAndIncrement requires a non-nil tx (no autocommit fallback)")
	}
	exec := tx

	if err := a.ensureCounterViaExec(ctx, exec, tenantID, profileCode, areaCode); err != nil {
		return 0, fmt.Errorf("prepare controlled document sequence counter: %w", err)
	}

	var next int
	if err := exec.QueryRowContext(ctx, `
		WITH locked AS (
			SELECT next_seq
			FROM cd_sequence_counters
			WHERE tenant_id = $1 AND profile_code = $2 AND process_area_code = $3
			FOR UPDATE
		)
		UPDATE cd_sequence_counters c
		SET next_seq = locked.next_seq + 1
		FROM locked
		WHERE c.tenant_id = $1 AND c.profile_code = $2 AND c.process_area_code = $3
		RETURNING locked.next_seq`,
		tenantID, profileCode, areaCode,
	).Scan(&next); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, controlleddocumentsdomain.ErrSequenceCounterNotFound
		}
		return 0, fmt.Errorf("increment controlled document sequence counter: %w", err)
	}
	return next, nil
}

// Template-version state reads moved to the templates-owned port
// (templates/domain.TemplateVersionPort.GetTemplateVersionState, impl
// templates/infrastructure.TemplateVersionReader) in M4 F4.2 — controlled-documents
// no longer queries templates_* tables. The reader is wired as the module's
// TemplateVersionChecker in module.go.

type rowScanner interface {
	Scan(dest ...any) error
}

type queryRower interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func scanControlledDocument(row rowScanner) (*controlleddocumentsdomain.ControlledDocument, error) {
	var (
		doc                     controlleddocumentsdomain.ControlledDocument
		departmentCode          sql.NullString
		sequenceNum             sql.NullInt64
		overrideTemplateVersion string
		visibilityScope         string
	)
	if err := row.Scan(
		&doc.ID,
		&doc.TenantID,
		&doc.ProfileCode,
		&doc.ProcessAreaCode,
		&departmentCode,
		&doc.Code,
		&sequenceNum,
		&doc.Title,
		&doc.OwnerUserID,
		&overrideTemplateVersion,
		&visibilityScope,
		&doc.Status,
		&doc.CreatedAt,
		&doc.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, controlleddocumentsdomain.ErrCDNotFound
		}
		return nil, fmt.Errorf("scan controlled document: %w", err)
	}
	doc.DepartmentCode = nullStringPtr(departmentCode)
	if sequenceNum.Valid {
		v := int(sequenceNum.Int64)
		doc.SequenceNum = &v
	}
	if strings.TrimSpace(overrideTemplateVersion) != "" {
		doc.OverrideTemplateVersionID = &overrideTemplateVersion
	}
	vis, err := controlleddocumentsdomain.NewVisibility(visibilityScope, nil, nil, "")
	if err != nil {
		return nil, fmt.Errorf("build controlled document visibility: %w", err)
	}
	doc.Visibility = vis
	return &doc, nil
}

func stringPtrToNull(v *string) sql.NullString {
	if v == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *v, Valid: true}
}

func intPtrToNull(v *int) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*v), Valid: true}
}

func nullStringPtr(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	value := v.String
	return &value
}
