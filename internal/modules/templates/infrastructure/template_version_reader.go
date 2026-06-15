package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"metaldocs/internal/platform/tenant"
)

const templateVersionQuery = `
	SELECT v.status, t.doc_type_code
	FROM templates_template_version v
	JOIN templates_template t ON t.id = v.template_id
	WHERE v.id = $1
	  AND t.tenant_id = $2::uuid
`

type TemplateVersionReader struct {
	db *sql.DB
}

func NewTemplateVersionReader(db *sql.DB) *TemplateVersionReader {
	if db == nil {
		panic("templates: template version reader db must not be nil")
	}
	return &TemplateVersionReader{db: db}
}

func (r *TemplateVersionReader) IsPublished(ctx context.Context, versionID string) (bool, string, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return false, "", err
	}
	var status sql.NullString
	var profileCode sql.NullString
	err = r.db.QueryRowContext(ctx, templateVersionQuery, versionID, tenantID).Scan(&status, &profileCode)
	if errors.Is(err, sql.ErrNoRows) {
		return false, "", nil
	}
	if err != nil {
		return false, "", err
	}
	if !status.Valid || status.String != "published" {
		return false, profileCode.String, nil
	}
	return true, profileCode.String, nil
}

// GetTemplateVersionState returns the raw status and owning-template doc_type_code
// for versionID, scoped to tenantID. status is nil when the version is absent or
// its status column is NULL; not-found is (nil, "", nil). tenantID is explicit
// (not ctx-derived) so cross-module / off-tx callers (controlled-documents create)
// pass it directly.
func (r *TemplateVersionReader) GetTemplateVersionState(ctx context.Context, tenantID, versionID string) (*string, string, error) {
	var status sql.NullString
	var docTypeCode sql.NullString
	err := r.db.QueryRowContext(ctx, templateVersionQuery, versionID, tenantID).Scan(&status, &docTypeCode)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", err
	}
	if !status.Valid {
		return nil, docTypeCode.String, nil
	}
	state := status.String
	return &state, docTypeCode.String, nil
}
