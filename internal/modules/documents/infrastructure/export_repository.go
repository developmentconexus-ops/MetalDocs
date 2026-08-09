package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"metaldocs/internal/modules/documents/domain"
)

// InsertExport records a new document export, or returns the existing row
// when the (tenant, document, composite_hash) conflict fires — export rows
// are content-addressed and idempotent on that key.
func (r *Repository) InsertExport(ctx context.Context, e *domain.Export) (*domain.Export, error) {
	var inserted domain.Export
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO document_exports
		     (tenant_id, document_id, revision_id, composite_hash, storage_key, size_bytes, paper_size, landscape, docgen_v2_ver)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 ON CONFLICT (tenant_id, document_id, composite_hash) DO NOTHING
		 RETURNING id, tenant_id, document_id, revision_id, composite_hash, storage_key, size_bytes, paper_size, landscape, docgen_v2_ver`,
		e.TenantID, e.DocumentID, e.RevisionID, e.CompositeHash, e.StorageKey,
		e.SizeBytes, e.PaperSize, e.Landscape, e.DocgenV2Ver,
	).Scan(
		&inserted.ID, &inserted.TenantID, &inserted.DocumentID, &inserted.RevisionID, &inserted.CompositeHash,
		&inserted.StorageKey, &inserted.SizeBytes, &inserted.PaperSize, &inserted.Landscape, &inserted.DocgenV2Ver,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return r.GetExportByHash(ctx, e.TenantID, e.DocumentID, e.CompositeHash)
	}
	if err != nil {
		return nil, err
	}
	return &inserted, nil
}

// GetExportByHash returns the export row matching the given composite hash.
// Returns domain.ErrNotFound if no such export exists.
func (r *Repository) GetExportByHash(ctx context.Context, tenantID, documentID string, compositeHash []byte) (*domain.Export, error) {
	var e domain.Export
	err := r.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, document_id, revision_id, composite_hash, storage_key, size_bytes, paper_size, landscape, docgen_v2_ver
		 FROM document_exports
		 WHERE tenant_id = $1 AND document_id = $2 AND composite_hash = $3`,
		tenantID, documentID, compositeHash,
	).Scan(
		&e.ID, &e.TenantID, &e.DocumentID, &e.RevisionID, &e.CompositeHash,
		&e.StorageKey, &e.SizeBytes, &e.PaperSize, &e.Landscape, &e.DocgenV2Ver,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}
