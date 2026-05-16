package docgenv2

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestTemplateReader_GetPublishedVersion_AllowsSystemTemplateTenant(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	reader := NewTemplateReader(db, nil, "")

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
		  AND (tpl.tenant_id = $2 OR tpl.tenant_id = $3)
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
