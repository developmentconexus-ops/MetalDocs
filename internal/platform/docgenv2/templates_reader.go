package docgenv2

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

const systemTemplateTenantID = "ffffffff-ffff-ffff-ffff-ffffffffffff"

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
		  AND (tpl.tenant_id = $2::uuid OR tpl.tenant_id = $3::uuid)
		  AND tv.status = 'published'`,
		templateVersionID, tenantID, systemTemplateTenantID,
	).Scan(&docxKey)
	if err != nil {
		return "", "", "", fmt.Errorf("templates reader: %w", err)
	}
	return docxKey, "", "", nil
}
