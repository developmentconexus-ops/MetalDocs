package docgenv2

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

const systemTemplateTenantID = "ffffffff-ffff-ffff-ffff-ffffffffffff"

// schemaMaxBytes is the per-call size cap passed to ObjectReader.AssertedReadObject for
// template schema files. Callers may not exceed this; the limit is enforced inside
// the kernel (VerifiedStore.ReadObject) before bytes reach the application.
const schemaMaxBytes int64 = 1 * 1024 * 1024 // 1 MiB — unchanged from prior hard-code

// ObjectReader abstracts object-store reads so TemplateReader is not coupled to a
// concrete MinIO client. *objectstore.VerifiedStore satisfies this interface via its
// AssertedReadObject method (F-O6): the tenant-guarded read so a mis-scoped schema
// key cannot fetch a cross-tenant object (the system tenant is allowed for built-in
// template schemas, matching the SQL scoping below).
type ObjectReader interface {
	AssertedReadObject(ctx context.Context, tenantID, key string, maxBytes int64) ([]byte, error)
}

type TemplateReader struct {
	db     *sql.DB
	store  ObjectReader
}

// NewTemplateReader constructs a TemplateReader backed by an ObjectReader (typically
// *objectstore.VerifiedStore). The store replaces the raw minio.Client so all object
// reads go through the kernel's size-limit enforcement (F-O6).
func NewTemplateReader(db *sql.DB, store ObjectReader) *TemplateReader {
	return &TemplateReader{db: db, store: store}
}

func (t *TemplateReader) GetPublishedVersion(ctx context.Context, tenantID, templateVersionID string) (docxKey, schemaKey, schemaJSON string, err error) {
	if t.db == nil {
		return "", "", "", errors.New("template reader db is nil")
	}
	if err := t.db.QueryRowContext(ctx, `
		SELECT tv.docx_storage_key, tv.schema_storage_key
		FROM template_versions tv
		JOIN templates tpl ON tpl.id = tv.template_id
		WHERE tv.id = $1
		  AND (tpl.tenant_id = $2 OR tpl.tenant_id = $3)
		  AND tv.status = 'published'`,
		templateVersionID, tenantID, systemTemplateTenantID,
	).Scan(&docxKey, &schemaKey); err != nil {
		return "", "", "", fmt.Errorf("get published version: %w", err)
	}

	if schemaKey == "" {
		return docxKey, schemaKey, "", nil
	}
	if t.store == nil {
		return "", "", "", errors.New("template reader object store is nil")
	}
	payload, err := t.store.AssertedReadObject(ctx, tenantID, schemaKey, schemaMaxBytes)
	if err != nil {
		return "", "", "", fmt.Errorf("read schema object: %w", err)
	}
	return docxKey, schemaKey, string(payload), nil
}
