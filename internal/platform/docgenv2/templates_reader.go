package docgenv2

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// TemplatesTemplateReader implements documents/application.TemplateReader
// for templates authored via the templates module. Schema is stored in the
// database (not S3), so schemaKey is always "" and schemaJSON is always "".
type TemplatesTemplateReader struct {
	db *sql.DB
}

func NewTemplatesTemplateReader(db *sql.DB) *TemplatesTemplateReader {
	return &TemplatesTemplateReader{db: db}
}

func (r *TemplatesTemplateReader) GetPublishedVersion(ctx context.Context, tenantID, templateVersionID string) (docxKey, schemaKey, schemaJSON string, err error) {
	if r.db == nil {
		return "", "", "", errors.New("templates template reader: db is nil")
	}
	err = r.db.QueryRowContext(ctx, `
		SELECT tv.docx_storage_key
		FROM templates_template_version tv
		JOIN templates_template tpl ON tpl.id = tv.template_id
		WHERE tv.id = $1
		  AND (tpl.tenant_id = $2 OR tpl.tenant_id = $3)
		  AND tv.status = 'published'`,
		templateVersionID, tenantID, systemTemplateTenantID,
	).Scan(&docxKey)
	if err != nil {
		return "", "", "", fmt.Errorf("templates reader: %w", err)
	}
	return docxKey, "", "", nil
}

// FanoutTemplateReader tries the primary reader first; if it returns sql.ErrNoRows,
// it falls back to the secondary reader.
// TODO: extract interfaces for testing so fanout behavior can be exercised
// without constructing concrete storage-backed readers.
type FanoutTemplateReader struct {
	primary   *TemplateReader
	secondary *TemplatesTemplateReader
}

func NewFanoutTemplateReader(primary *TemplateReader, secondary *TemplatesTemplateReader) *FanoutTemplateReader {
	return &FanoutTemplateReader{primary: primary, secondary: secondary}
}

func (f *FanoutTemplateReader) GetPublishedVersion(ctx context.Context, tenantID, templateVersionID string) (docxKey, schemaKey, schemaJSON string, err error) {
	docxKey, schemaKey, schemaJSON, err = f.primary.GetPublishedVersion(ctx, tenantID, templateVersionID)
	if err == nil {
		return docxKey, schemaKey, schemaJSON, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", "", "", fmt.Errorf("templates reader secondary: %w", err)
	}
	return f.secondary.GetPublishedVersion(ctx, tenantID, templateVersionID)
}
