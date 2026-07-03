package docgenv2

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

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
