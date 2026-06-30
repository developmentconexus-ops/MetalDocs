package docgenv2

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// fakeObjectReader captures the tenant-guarded read args so the test can assert
// TemplateReader forwards tenantID + the schema size cap into the kernel.
type fakeObjectReader struct {
	gotTenantID string
	gotKey      string
	gotMaxBytes int64
	payload     []byte
	err         error
}

func (f *fakeObjectReader) AssertedReadObject(_ context.Context, tenantID, key string, maxBytes int64) ([]byte, error) {
	f.gotTenantID = tenantID
	f.gotKey = key
	f.gotMaxBytes = maxBytes
	return f.payload, f.err
}

func TestTemplateReader_GetPublishedVersion_ReadsSchemaThroughGuardedReader(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := &fakeObjectReader{payload: []byte(`{"fields":[]}`)}
	reader := NewTemplateReader(db, store)

	const schemaKey = "tenants/tenant-1/templates/tpl-9/v3/schema.json"
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT tv.docx_storage_key, tv.schema_storage_key
		FROM template_versions tv
		JOIN templates tpl ON tpl.id = tv.template_id
		WHERE tv.id = $1
		  AND (tpl.tenant_id = $2 OR tpl.tenant_id = $3)
		  AND tv.status = 'published'`)).
		WithArgs("tv-1", "tenant-1", systemTemplateTenantID).
		WillReturnRows(sqlmock.NewRows([]string{"docx_storage_key", "schema_storage_key"}).AddRow("s3://docx", schemaKey))

	_, gotSchemaKey, schemaJSON, err := reader.GetPublishedVersion(context.Background(), "tenant-1", "tv-1")
	if err != nil {
		t.Fatalf("GetPublishedVersion: %v", err)
	}
	if schemaJSON != `{"fields":[]}` {
		t.Fatalf("schemaJSON = %q, want the read payload", schemaJSON)
	}
	if gotSchemaKey != schemaKey {
		t.Fatalf("schemaKey = %q, want %q", gotSchemaKey, schemaKey)
	}
	// The guarded read must receive the caller's tenant, the DB key, and the cap.
	if store.gotTenantID != "tenant-1" {
		t.Fatalf("forwarded tenantID = %q, want tenant-1", store.gotTenantID)
	}
	if store.gotKey != schemaKey {
		t.Fatalf("forwarded key = %q, want %q", store.gotKey, schemaKey)
	}
	if store.gotMaxBytes != schemaMaxBytes {
		t.Fatalf("forwarded maxBytes = %d, want %d", store.gotMaxBytes, schemaMaxBytes)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestTemplateReader_GetPublishedVersion_AllowsSystemTemplateTenant(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	reader := NewTemplateReader(db, nil)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT tv.docx_storage_key, tv.schema_storage_key
		FROM template_versions tv
		JOIN templates tpl ON tpl.id = tv.template_id
		WHERE tv.id = $1
		  AND (tpl.tenant_id = $2 OR tpl.tenant_id = $3)
		  AND tv.status = 'published'`)).
		WithArgs("tv-1", "tenant-1", systemTemplateTenantID).
		WillReturnRows(sqlmock.NewRows([]string{"docx_storage_key", "schema_storage_key"}).AddRow("s3://docx", ""))

	docxKey, schemaKey, schemaJSON, err := reader.GetPublishedVersion(context.Background(), "tenant-1", "tv-1")
	if err != nil {
		t.Fatalf("GetPublishedVersion: %v", err)
	}
	if docxKey != "s3://docx" {
		t.Fatalf("docxKey = %q, want s3://docx", docxKey)
	}
	if schemaKey != "" || schemaJSON != "" {
		t.Fatalf("schemaKey/schemaJSON = %q/%q, want empty", schemaKey, schemaJSON)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestTemplatesTemplateReader_GetPublishedVersion_AllowsSystemTemplateTenant(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	reader := NewTemplatesTemplateReader(db)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT tv.docx_storage_key
		FROM templates_template_version tv
		JOIN templates_template tpl ON tpl.id = tv.template_id
		WHERE tv.id = $1
		  AND (tpl.tenant_id = $2::uuid OR tpl.tenant_id = $3::uuid)
		  AND tv.status = 'published'`)).
		WithArgs("template-version-1", "tenant-1", systemTemplateTenantID).
		WillReturnRows(sqlmock.NewRows([]string{"docx_storage_key"}).AddRow("s3://template-docx"))

	docxKey, schemaKey, schemaJSON, err := reader.GetPublishedVersion(context.Background(), "tenant-1", "template-version-1")
	if err != nil {
		t.Fatalf("GetPublishedVersion: %v", err)
	}
	if docxKey != "s3://template-docx" {
		t.Fatalf("docxKey = %q, want s3://template-docx", docxKey)
	}
	if schemaKey != "" || schemaJSON != "" {
		t.Fatalf("schemaKey/schemaJSON = %q/%q, want empty", schemaKey, schemaJSON)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
