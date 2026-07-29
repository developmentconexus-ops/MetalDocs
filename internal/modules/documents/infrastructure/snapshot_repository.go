package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"metaldocs/internal/modules/documents/domain"
)

// SnapshotRepository reads and writes the template snapshot columns on documents.
type SnapshotRepository struct {
	db     *sql.DB
	schema string // optional schema prefix; empty = bare table name
}

type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// NewSnapshotRepository creates a SnapshotRepository using bare table names.
// In tests, use NewSnapshotRepositoryWithSchema to point at the isolated test schema.
func NewSnapshotRepository(db *sql.DB) *SnapshotRepository {
	return &SnapshotRepository{db: db}
}

// NewSnapshotRepositoryWithSchema creates a SnapshotRepository that qualifies
// table names with the given schema. Used by integration tests.
func NewSnapshotRepositoryWithSchema(db *sql.DB, schema string) *SnapshotRepository {
	return &SnapshotRepository{db: db, schema: schema}
}

func (r *SnapshotRepository) table(name string) string {
	if r.schema == "" {
		return name
	}
	return fmt.Sprintf("%q.%q", r.schema, name)
}

func requireRowsAffected(result sql.Result, action string) error {
	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s rows affected: %w", action, err)
	}
	if n == 0 {
		return fmt.Errorf("%s: not found or already in target state", action)
	}
	return nil
}

// ReadSnapshotWithFreezeAt reads snapshot columns and values_frozen_at for idempotency checks.
func (r *SnapshotRepository) ReadSnapshotWithFreezeAt(ctx context.Context, tenantID, docID string, q ...DBTX) (domain.TemplateSnapshot, *time.Time, error) {
	exec := DBTX(r.db)
	if len(q) > 0 && q[0] != nil {
		exec = q[0]
	}
	return r.readSnapshot(ctx, exec, tenantID, docID)
}

// ReadFreezeAt reads only values_frozen_at — used by Pin to check idempotency without
// fetching the snapshot blobs that Pin does not consume.
func (r *SnapshotRepository) ReadFreezeAt(ctx context.Context, tenantID, docID string, q ...DBTX) (*time.Time, error) {
	exec := DBTX(r.db)
	if len(q) > 0 && q[0] != nil {
		exec = q[0]
	}
	var valuesFrozenAt *time.Time
	err := exec.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT values_frozen_at
		  FROM %s
		 WHERE tenant_id = $1::uuid AND id = $2::uuid`, r.table("documents")),
		tenantID, docID,
	).Scan(&valuesFrozenAt)
	return valuesFrozenAt, err
}

func (r *SnapshotRepository) readSnapshot(ctx context.Context, exec DBTX, tenantID, docID string) (domain.TemplateSnapshot, *time.Time, error) {
	var s domain.TemplateSnapshot
	var valuesFrozenAt *time.Time
	err := exec.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT placeholder_schema_snapshot,
		       composition_config_snapshot,
		       coalesce(body_docx_snapshot_s3_key, ''),
		       values_frozen_at
		  FROM %s
		 WHERE tenant_id = $1::uuid AND id = $2::uuid`, r.table("documents")),
		tenantID, docID,
	).Scan(
		&s.PlaceholderSchemaJSON,
		&s.CompositionJSON,
		&s.BodyDocxS3Key,
		&valuesFrozenAt,
	)
	return s, valuesFrozenAt, err
}

// ReadCurrentRevisionRef returns the document's current editor revision — the
// authored body the approver actually reviewed — as {id, storage_key,
// content_hash}. It is the FREEZE-TIME read: Pin uses it to decide which
// revision to pin.
//
// F-QA3-1 (operator ruling: option (a)): the freeze pipeline materializes the
// EDITOR revision, not the template snapshot. The template snapshot only seeds
// the initial clone; every later edit lands on document_revisions and becomes
// the frozen truth. Returns a zero RevisionRef when the document has no current
// revision so the caller can fail closed (no-fallback principle) instead of
// silently rendering an empty template body.
//
// document_revisions carries no tenant_id of its own — tenancy is enforced by
// the join through documents, which is tenant-predicated here.
func (r *SnapshotRepository) ReadCurrentRevisionRef(ctx context.Context, tenantID, docID string, q ...DBTX) (domain.RevisionRef, error) {
	return r.readRevisionRef(ctx, "current_revision_id", tenantID, docID, q...)
}

// ReadFrozenRevisionRef returns the revision PINNED at freeze time
// (documents.frozen_revision_id, migration 0313) — the revision Materialize
// must render and verify. It deliberately does NOT fall back to
// current_revision_id: a document whose pin is absent (frozen before 0313) has
// no recorded lineage, and rendering the head instead would re-open exactly the
// drift window the pin closes.
func (r *SnapshotRepository) ReadFrozenRevisionRef(ctx context.Context, tenantID, docID string, q ...DBTX) (domain.RevisionRef, error) {
	return r.readRevisionRef(ctx, "frozen_revision_id", tenantID, docID, q...)
}

// readRevisionRef is the shared body for the two revision-ref reads above.
// pointerColumn is a package-internal literal (never caller-supplied), so the
// Sprintf carries no injection surface.
func (r *SnapshotRepository) readRevisionRef(ctx context.Context, pointerColumn, tenantID, docID string, q ...DBTX) (domain.RevisionRef, error) {
	exec := DBTX(r.db)
	if len(q) > 0 && q[0] != nil {
		exec = q[0]
	}
	var ref domain.RevisionRef
	err := exec.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT coalesce(rev.id::text, ''),
		       coalesce(rev.storage_key, ''),
		       coalesce(rev.content_hash, '')
		  FROM %s AS d
		  LEFT JOIN %s AS rev ON rev.id = d.%s
		 WHERE d.tenant_id = $1::uuid AND d.id = $2::uuid`,
		r.table("documents"), r.table("document_revisions"), pointerColumn),
		tenantID, docID,
	).Scan(&ref.ID, &ref.StorageKey, &ref.ContentHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.RevisionRef{}, fmt.Errorf("document not found: %s", docID)
		}
		return domain.RevisionRef{}, err
	}
	return ref, nil
}

// WriteFreeze stamps the freeze state on a document: the values hash, the
// freeze timestamp, and the lineage pin naming WHICH document_revisions row was
// frozen (migration 0313). All three land in ONE update so a freeze can never
// be half-recorded — a values_hash without a pin would be a freeze nobody can
// verify afterwards.
//
// frozenRevisionID is a document_revisions.id (NOT the documents.id that the
// freeze pipeline's `revisionID` parameter carries). An empty string writes a
// NULL pin, which is the honest record for a document that has no revision at
// all; Materialize then fails closed on it rather than guessing a body.
func (r *SnapshotRepository) WriteFreeze(ctx context.Context, tenantID, docID string, valuesHash []byte, frozenRevisionID string, frozenAt time.Time, q ...DBTX) error {
	exec := DBTX(r.db)
	if len(q) > 0 && q[0] != nil {
		exec = q[0]
	}
	result, err := exec.ExecContext(ctx, fmt.Sprintf(`
        UPDATE %s
           SET values_hash=$1, values_frozen_at=$2, frozen_revision_id=NULLIF($3,'')::uuid
         WHERE tenant_id=$4::uuid AND id=$5::uuid`, r.table("documents")),
		valuesHash, frozenAt, frozenRevisionID, tenantID, docID)
	if err != nil {
		return fmt.Errorf("write freeze: %w", err)
	}
	return requireRowsAffected(result, "write freeze")
}

// WriteFinalDocx persists the fanout output pointer and content hash onto a document.
func (r *SnapshotRepository) WriteFinalDocx(ctx context.Context, tenantID, docID, s3Key string, contentHash []byte, q ...DBTX) error {
	exec := DBTX(r.db)
	if len(q) > 0 && q[0] != nil {
		exec = q[0]
	}
	result, err := exec.ExecContext(ctx, fmt.Sprintf(`
        UPDATE %s
           SET final_docx_s3_key=$1, content_hash=$2
         WHERE tenant_id=$3::uuid AND id=$4::uuid`, r.table("documents")),
		s3Key, contentHash, tenantID, docID)
	if err != nil {
		return fmt.Errorf("write final docx: %w", err)
	}
	return requireRowsAffected(result, "write final docx")
}

// ReadFinalDocxS3Key returns the frozen DOCX S3 key for a document.
// Returns an error if the document does not exist or has not been frozen yet.
func (r *SnapshotRepository) ReadFinalDocxS3Key(ctx context.Context, tenantID, docID string) (string, error) {
	var key sql.NullString
	err := r.db.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT final_docx_s3_key
		  FROM %s
		 WHERE tenant_id=$1::uuid AND id=$2::uuid`, r.table("documents")),
		tenantID, docID).Scan(&key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("document not found: %s", docID)
		}
		return "", err
	}
	if !key.Valid || key.String == "" {
		return "", fmt.Errorf("document not frozen: no final_docx_s3_key for %s", docID)
	}
	return key.String, nil
}

// WritePDF persists the rendered PDF pointer, hash, and generation timestamp.
// An optional DBTX (q) lets the caller run this inside its own transaction
// (M3 F3.2 — the pdf job runner wraps this write in a SeedTxTenant-seeded tx);
// omitted, it runs directly against the pool as before.
func (r *SnapshotRepository) WritePDF(ctx context.Context, tenantID, docID, s3Key string, pdfHash []byte, generatedAt time.Time, q ...DBTX) error {
	exec := DBTX(r.db)
	if len(q) > 0 && q[0] != nil {
		exec = q[0]
	}
	result, err := exec.ExecContext(ctx, fmt.Sprintf(`
        UPDATE %s
           SET final_pdf_s3_key=$1, pdf_hash=$2, pdf_generated_at=$3
         WHERE tenant_id=$4::uuid AND id=$5::uuid`, r.table("documents")),
		s3Key, pdfHash, generatedAt, tenantID, docID)
	if err != nil {
		return fmt.Errorf("write pdf: %w", err)
	}
	return requireRowsAffected(result, "write pdf")
}

func (r *SnapshotRepository) ResolveTenantByDocumentID(ctx context.Context, docID string) (string, error) {
	var tenantID string
	if err := r.db.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT tenant_id::text
		  FROM %s
		 WHERE id = $1::uuid`, r.table("documents")),
		docID,
	).Scan(&tenantID); err != nil {
		return "", err
	}
	return tenantID, nil
}

// AppendReconstruction appends a forensic attempt entry onto documents.reconstruction_attempts.
// Never touches final_docx_s3_key or content_hash.
func (r *SnapshotRepository) AppendReconstruction(ctx context.Context, tenantID, docID string, entry []byte) error {
	result, err := r.db.ExecContext(ctx, fmt.Sprintf(`
        UPDATE %s
           SET reconstruction_attempts = reconstruction_attempts || $1::jsonb
         WHERE tenant_id=$2::uuid AND id=$3::uuid`, r.table("documents")),
		entry, tenantID, docID)
	if err != nil {
		return fmt.Errorf("append reconstruction: %w", err)
	}
	return requireRowsAffected(result, "append reconstruction")
}
