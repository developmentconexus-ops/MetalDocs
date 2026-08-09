package docgenv2

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"metaldocs/internal/modules/documents/application"
	"metaldocs/internal/modules/documents/domain"
)

var _ application.SnapshotTemplateReader = (*TemplatesSnapshotReader)(nil)

// TemplatesSnapshotReader implements application.SnapshotTemplateReader by
// loading published template versions authored via the templates module.
type TemplatesSnapshotReader struct {
	db *sql.DB
}

// NewTemplatesSnapshotReader constructs a TemplatesSnapshotReader backed by db.
func NewTemplatesSnapshotReader(db *sql.DB) *TemplatesSnapshotReader {
	return &TemplatesSnapshotReader{db: db}
}

// LoadForSnapshot returns the placeholder schema and DOCX storage key for the
// published template version, for embedding into a document snapshot.
func (r *TemplatesSnapshotReader) LoadForSnapshot(ctx context.Context, tenantID, templateVersionID string) (domain.TemplateSnapshot, error) {
	if r.db == nil {
		return domain.TemplateSnapshot{}, errors.New("templates snapshot reader: db is nil")
	}

	var phJSON, docxKey string
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(tv.placeholder_schema::text, '{}'),
		       COALESCE(tv.docx_storage_key, '')
		  FROM templates_template_version tv
		  JOIN templates_template tpl ON tpl.id = tv.template_id
		 WHERE tv.id = $1::uuid
		   AND tpl.tenant_id = $2::uuid
		   AND tv.status = 'published'`,
		templateVersionID, tenantID,
	).Scan(&phJSON, &docxKey)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.TemplateSnapshot{}, domain.ErrSnapshotTemplateNotFound
	}
	if err != nil {
		return domain.TemplateSnapshot{}, fmt.Errorf("templates snapshot reader: %w", err)
	}

	return domain.TemplateSnapshot{
		PlaceholderSchemaJSON: []byte(phJSON),
		CompositionJSON:       []byte("{}"),
		BodyDocxBytes:         nil,
		BodyDocxS3Key:         docxKey,
	}, nil
}
