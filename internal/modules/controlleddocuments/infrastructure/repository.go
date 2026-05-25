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
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
)

type PostgresControlledDocumentRepository struct {
	db *sql.DB
}

func NewPostgresControlledDocumentRepository(db *sql.DB) *PostgresControlledDocumentRepository {
	return &PostgresControlledDocumentRepository{db: db}
}

func (r *PostgresControlledDocumentRepository) GetByID(ctx context.Context, tenantID, id string) (*controlleddocumentsdomain.ControlledDocument, error) {
	const q = `
SELECT id::text, tenant_id::text, profile_code, process_area_code, department_code,
       code, sequence_num, title, owner_user_id, coalesce(override_template_version_id::text, ''),
       visibility_scope, status, created_at, updated_at
FROM controlled_documents
WHERE tenant_id = $1 AND id = $2`
	doc, err := scanControlledDocument(r.db.QueryRowContext(ctx, q, tenantID, id))
	if err != nil {
		return nil, err
	}
	if doc.Visibility.Scope == controlleddocumentsdomain.VisibilityScopeRestricted {
		vis, err := r.loadVisibilityGrants(ctx, tenantID, doc.ID)
		if err != nil {
			return nil, err
		}
		doc.Visibility = vis
	}
	return doc, nil
}

func (r *PostgresControlledDocumentRepository) GetByCode(ctx context.Context, tenantID, profileCode, code string) (*controlleddocumentsdomain.ControlledDocument, error) {
	const q = `
SELECT id::text, tenant_id::text, profile_code, process_area_code, department_code,
       code, sequence_num, title, owner_user_id, coalesce(override_template_version_id::text, ''),
       visibility_scope, status, created_at, updated_at
FROM controlled_documents
WHERE tenant_id = $1 AND profile_code = $2 AND code = $3`
	doc, err := scanControlledDocument(r.db.QueryRowContext(ctx, q, tenantID, profileCode, code))
	if err != nil {
		return nil, err
	}
	if doc.Visibility.Scope == controlleddocumentsdomain.VisibilityScopeRestricted {
		vis, err := r.loadVisibilityGrants(ctx, tenantID, doc.ID)
		if err != nil {
			return nil, err
		}
		doc.Visibility = vis
	}
	return doc, nil
}

func (r *PostgresControlledDocumentRepository) CodeExists(ctx context.Context, tenantID, profileCode, code string) (bool, error) {
	const q = `SELECT EXISTS(
		SELECT 1 FROM controlled_documents WHERE tenant_id = $1 AND profile_code = $2 AND code = $3
	)`
	var exists bool
	if err := r.db.QueryRowContext(ctx, q, tenantID, profileCode, code).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *PostgresControlledDocumentRepository) List(ctx context.Context, tenantID string, filter controlleddocumentsdomain.CDFilter) ([]controlleddocumentsdomain.ControlledDocument, error) {
	q := `
SELECT id::text, tenant_id::text, profile_code, process_area_code, department_code,
       code, sequence_num, title, owner_user_id, coalesce(override_template_version_id::text, ''),
       visibility_scope, status, created_at, updated_at
FROM controlled_documents
WHERE tenant_id = $1`
	args := []any{tenantID}
	idx := 2

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
		q += fmt.Sprintf(" AND (code ILIKE $%d OR title ILIKE $%d)", idx, idx)
		args = append(args, "%"+strings.TrimSpace(*filter.Query)+"%")
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
                      FROM user_process_areas upa
                     WHERE upa.tenant_id = controlled_documents.tenant_id
                       AND upa.user_id = $%d
                       AND upa.area_code = cdag.area_code
                       AND upa.effective_to IS NULL
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
	q += " ORDER BY created_at DESC"
	if filter.Limit > 0 {
		q += fmt.Sprintf(" LIMIT $%d", idx)
		args = append(args, filter.Limit)
		idx++
	}
	if filter.Offset > 0 {
		q += fmt.Sprintf(" OFFSET $%d", idx)
		args = append(args, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]controlleddocumentsdomain.ControlledDocument, 0)
	for rows.Next() {
		doc, err := scanControlledDocument(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *doc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := r.hydrateVisibilityGrants(ctx, tenantID, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *PostgresControlledDocumentRepository) loadVisibilityGrants(ctx context.Context, tenantID, controlledDocumentID string) (controlleddocumentsdomain.Visibility, error) {
	areas, err := r.loadAreaGrants(ctx, tenantID, controlledDocumentID)
	if err != nil {
		return controlleddocumentsdomain.Visibility{}, err
	}
	users, err := r.loadUserGrants(ctx, tenantID, controlledDocumentID)
	if err != nil {
		return controlleddocumentsdomain.Visibility{}, err
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
		return nil, err
	}
	defer rows.Close()

	out := []string{}
	for rows.Next() {
		var areaCode string
		if err := rows.Scan(&areaCode); err != nil {
			return nil, err
		}
		out = append(out, areaCode)
	}
	return out, rows.Err()
}

func (r *PostgresControlledDocumentRepository) loadUserGrants(ctx context.Context, tenantID, controlledDocumentID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT user_id
FROM controlled_document_user_grants
WHERE tenant_id = $1 AND controlled_document_id = $2
ORDER BY user_id`, tenantID, controlledDocumentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []string{}
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		out = append(out, userID)
	}
	return out, rows.Err()
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
		return err
	}
	defer areaRows.Close()
	for areaRows.Next() {
		var docID, areaCode string
		if err := areaRows.Scan(&docID, &areaCode); err != nil {
			return err
		}
		if idx, ok := indexByID[docID]; ok {
			docs[idx].Visibility.AreaCodes = append(docs[idx].Visibility.AreaCodes, areaCode)
		}
	}
	if err := areaRows.Err(); err != nil {
		return err
	}

	userRows, err := r.db.QueryContext(ctx, `
SELECT controlled_document_id::text, user_id
FROM controlled_document_user_grants
WHERE tenant_id = $1 AND controlled_document_id = ANY($2)
ORDER BY controlled_document_id, user_id`, tenantID, pgtype.FlatArray[string](ids))
	if err != nil {
		return err
	}
	defer userRows.Close()
	for userRows.Next() {
		var docID, userID string
		if err := userRows.Scan(&docID, &userID); err != nil {
			return err
		}
		if idx, ok := indexByID[docID]; ok {
			docs[idx].Visibility.UserIDs = append(docs[idx].Visibility.UserIDs, userID)
		}
	}
	return userRows.Err()
}

func (r *PostgresControlledDocumentRepository) Create(ctx context.Context, doc *controlleddocumentsdomain.ControlledDocument) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin create CD tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := authz.Require(ctx, tx, string(iamdomain.CapControlledDocumentCreate), "tenant"); err != nil {
		return fmt.Errorf("registry: authz check Create: %w", err)
	}
	if err := r.createWithQueryer(ctx, tx, doc); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *PostgresControlledDocumentRepository) CreateTx(ctx context.Context, tx controlleddocumentsdomain.DBTX, doc *controlleddocumentsdomain.ControlledDocument) error {
	if tx == nil {
		return errors.New("nil transaction")
	}
	sqlTx, ok := tx.(*sql.Tx)
	if !ok {
		return errors.New("create tx requires *sql.Tx for authz")
	}
	if err := authz.Require(ctx, sqlTx, string(iamdomain.CapControlledDocumentCreate), "tenant"); err != nil {
		return fmt.Errorf("registry: authz check CreateTx: %w", err)
	}
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

func (r *PostgresControlledDocumentRepository) UpdateStatus(ctx context.Context, tenantID, id string, status controlleddocumentsdomain.CDStatus, updatedAt time.Time) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE controlled_documents SET status = $1, updated_at = $2 WHERE tenant_id = $3 AND id = $4`,
		status, updatedAt, tenantID, id,
	)
	if err != nil {
		return err
	}
	n, rowsErr := res.RowsAffected()
	if rowsErr != nil {
		return rowsErr
	}
	if n == 0 {
		return controlleddocumentsdomain.ErrCDNotFound
	}
	return nil
}

func (r *PostgresControlledDocumentRepository) UpdateStatusTx(ctx context.Context, tx controlleddocumentsdomain.DBTX, tenantID, id string, status controlleddocumentsdomain.CDStatus, updatedAt time.Time) error {
	res, err := tx.ExecContext(ctx,
		`UPDATE controlled_documents SET status = $1, updated_at = $2 WHERE tenant_id = $3 AND id = $4`,
		status, updatedAt, tenantID, id,
	)
	if err != nil {
		return err
	}
	n, rowsErr := res.RowsAffected()
	if rowsErr != nil {
		return rowsErr
	}
	if n == 0 {
		return controlleddocumentsdomain.ErrCDNotFound
	}
	return nil
}

func (r *PostgresControlledDocumentRepository) CanRead(ctx context.Context, tenantID, controlledDocumentID, actorUserID string) (bool, error) {
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
                          FROM user_process_areas upa
                         WHERE upa.tenant_id = cd.tenant_id
                           AND upa.user_id = $3
                           AND upa.area_code = cdag.area_code
                           AND upa.effective_to IS NULL
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
		return false, fmt.Errorf("check controlled document visibility: %w", err)
	}
	return exists, nil
}

type PostgresSequenceAllocator struct {
	db *sql.DB
}

func NewPostgresSequenceAllocator(db *sql.DB) *PostgresSequenceAllocator {
	return &PostgresSequenceAllocator{db: db}
}

// EnsureCounter initialises a sequence counter for the given tenant/profile/area combination.
func (a *PostgresSequenceAllocator) EnsureCounter(ctx context.Context, tenantID, profileCode, areaCode string) error {
	return a.ensureCounterViaExec(ctx, a.db, tenantID, profileCode, areaCode)
}

func (a *PostgresSequenceAllocator) ensureCounterViaExec(ctx context.Context, exec controlleddocumentsdomain.DBTX, tenantID, profileCode, areaCode string) error {
	_, err := exec.ExecContext(ctx, `
		INSERT INTO cd_sequence_counters (tenant_id, profile_code, process_area_code, next_seq)
		VALUES ($1, $2, $3, 1)
		ON CONFLICT (tenant_id, profile_code, process_area_code) DO NOTHING`,
		tenantID, profileCode, areaCode,
	)
	if err != nil {
		return fmt.Errorf("ensure sequence counter: %w", err)
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
		return 0, fmt.Errorf("peek sequence query: %w", err)
	}
	return next, nil
}

// NextAndIncrement atomically increments and returns the next sequence number.
func (a *PostgresSequenceAllocator) NextAndIncrement(ctx context.Context, tx controlleddocumentsdomain.DBTX, tenantID, profileCode, areaCode string) (int, error) {
	var exec controlleddocumentsdomain.DBTX = a.db
	if tx != nil {
		exec = tx
	}

	if err := a.ensureCounterViaExec(ctx, exec, tenantID, profileCode, areaCode); err != nil {
		return 0, fmt.Errorf("ensure sequence counter before increment: %w", err)
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
		return 0, fmt.Errorf("next sequence increment query: %w", err)
	}
	return next, nil
}

type PostgresTemplateVersionChecker struct {
	db *sql.DB
}

func NewPostgresTemplateVersionChecker(db *sql.DB) *PostgresTemplateVersionChecker {
	return &PostgresTemplateVersionChecker{db: db}
}

func (c *PostgresTemplateVersionChecker) GetTemplateVersionState(ctx context.Context, tenantID, templateVersionID string) (*string, string, error) {
	var status sql.NullString
	var profileCode sql.NullString
	err := c.db.QueryRowContext(ctx, `
		SELECT v.status, t.profile_code
		FROM templates_template_version v
		JOIN templates_template t ON t.id = v.template_id
		WHERE v.id = $1
		  AND t.tenant_id = $2`, templateVersionID, tenantID,
	).Scan(&status, &profileCode)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", fmt.Errorf("get template version state query: %w", err)
	}
	if !status.Valid {
		return nil, profileCode.String, nil
	}
	state := status.String
	return &state, profileCode.String, nil
}

type TaxonomyProfileReader struct {
	db *sql.DB
}

func NewTaxonomyProfileReader(db *sql.DB) *TaxonomyProfileReader {
	return &TaxonomyProfileReader{db: db}
}

func (r *TaxonomyProfileReader) GetByCode(ctx context.Context, tenantID, code string) (*taxonomydomain.DocumentProfile, error) {
	const q = `
SELECT code, tenant_id, family_code, name, description, review_interval_days,
       default_template_version_id, owner_user_id, editable_by_role, archived_at, created_at
FROM metaldocs.document_profiles
WHERE tenant_id = $1 AND code = $2`

	var profile taxonomydomain.DocumentProfile
	var defaultTemplateVersionID sql.NullString
	var ownerUserID sql.NullString
	err := r.db.QueryRowContext(ctx, q, tenantID, code).Scan(
		&profile.Code,
		&profile.TenantID,
		&profile.FamilyCode,
		&profile.Name,
		&profile.Description,
		&profile.ReviewIntervalDays,
		&defaultTemplateVersionID,
		&ownerUserID,
		&profile.EditableByRole,
		&profile.ArchivedAt,
		&profile.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, taxonomydomain.ErrProfileNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get document profile by code query: %w", err)
	}
	profile.DefaultTemplateVersionID = nullStringPtr(defaultTemplateVersionID)
	profile.OwnerUserID = nullStringPtr(ownerUserID)
	return &profile, nil
}

type TaxonomyAreaReader struct {
	db *sql.DB
}

func NewTaxonomyAreaReader(db *sql.DB) *TaxonomyAreaReader { return &TaxonomyAreaReader{db: db} }

func (r *TaxonomyAreaReader) GetByCode(ctx context.Context, tenantID, code string) (*taxonomydomain.ProcessArea, error) {
	const q = `
SELECT code, tenant_id, name, description, parent_code, owner_user_id, default_approver_role, archived_at, created_at
FROM metaldocs.document_process_areas
WHERE tenant_id = $1 AND code = $2`

	var area taxonomydomain.ProcessArea
	var parentCode sql.NullString
	var ownerUserID sql.NullString
	var defaultApproverRole sql.NullString
	err := r.db.QueryRowContext(ctx, q, tenantID, code).Scan(
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
		return nil, taxonomydomain.ErrAreaNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get process area by code query: %w", err)
	}
	area.ParentCode = nullStringPtr(parentCode)
	area.OwnerUserID = nullStringPtr(ownerUserID)
	area.DefaultApproverRole = nullStringPtr(defaultApproverRole)
	return &area, nil
}

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
		return nil, err
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
		return nil, err
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
