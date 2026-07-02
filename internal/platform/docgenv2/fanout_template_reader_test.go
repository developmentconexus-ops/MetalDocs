package docgenv2

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// newFanoutUnderTest wires a FanoutTemplateReader over two independent sqlmock
// DBs (one per concrete reader, since TemplatesTemplateReader/TemplateReader
// each own a *sql.DB) so canonical and legacy hits can be asserted
// independently.
func newFanoutUnderTest(t *testing.T) (f *FanoutTemplateReader, canonicalMock, legacyMock sqlmock.Sqlmock, cleanup func()) {
	t.Helper()

	canonicalDB, canonicalMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New (canonical): %v", err)
	}
	legacyDB, legacyMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New (legacy): %v", err)
	}

	primary := NewTemplatesTemplateReader(canonicalDB)
	secondary := NewTemplateReader(legacyDB, &fakeObjectReader{payload: []byte(`{"fields":[]}`)})
	f = NewFanoutTemplateReader(primary, secondary)

	cleanup = func() {
		canonicalDB.Close()
		legacyDB.Close()
	}
	return f, canonicalMock, legacyMock, cleanup
}

const canonicalQuery = `
	SELECT tv.docx_storage_key
	FROM templates_template_version tv
	JOIN templates_template tpl ON tpl.id = tv.template_id
	WHERE tv.id = $1
	  AND (tpl.tenant_id = $2::uuid OR tpl.tenant_id = $3::uuid)
	  AND tv.status = 'published'`

const legacyQuery = `
	SELECT tv.docx_storage_key, tv.schema_storage_key
	FROM template_versions tv
	JOIN templates tpl ON tpl.id = tv.template_id
	WHERE tv.id = $1
	  AND (tpl.tenant_id = $2 OR tpl.tenant_id = $3)
	  AND tv.status = 'published'`

func TestFanoutTemplateReader_CanonicalHit_NoLegacyCall(t *testing.T) {
	f, canonicalMock, legacyMock, cleanup := newFanoutUnderTest(t)
	defer cleanup()

	before := LegacyTemplateReadCount()

	canonicalMock.ExpectQuery(regexp.QuoteMeta(canonicalQuery)).
		WithArgs("tv-1", "tenant-1", systemTemplateTenantID).
		WillReturnRows(sqlmock.NewRows([]string{"docx_storage_key"}).AddRow("s3://canonical-docx"))

	docxKey, schemaKey, schemaJSON, err := f.GetPublishedVersion(context.Background(), "tenant-1", "tv-1")
	if err != nil {
		t.Fatalf("GetPublishedVersion: %v", err)
	}
	if docxKey != "s3://canonical-docx" {
		t.Fatalf("docxKey = %q, want s3://canonical-docx", docxKey)
	}
	if schemaKey != "" || schemaJSON != "" {
		t.Fatalf("schemaKey/schemaJSON = %q/%q, want empty (canonical reader never has S3 schema)", schemaKey, schemaJSON)
	}
	if err := canonicalMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("canonical sql expectations: %v", err)
	}
	// The legacy mock has zero expectations queued; any query against it would
	// fail ExpectationsWereMet with "not expected" — asserting it stays clean
	// proves the legacy reader was never invoked.
	if err := legacyMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("legacy reader was called on a canonical hit: %v", err)
	}
	if got := LegacyTemplateReadCount(); got != before {
		t.Fatalf("legacyTemplateReadTotal = %d, want unchanged at %d (canonical hit must not count as legacy)", got, before)
	}
}

func TestFanoutTemplateReader_CanonicalNotFound_FallsBackToLegacy(t *testing.T) {
	f, canonicalMock, legacyMock, cleanup := newFanoutUnderTest(t)
	defer cleanup()

	before := LegacyTemplateReadCount()

	canonicalMock.ExpectQuery(regexp.QuoteMeta(canonicalQuery)).
		WithArgs("tv-1", "tenant-1", systemTemplateTenantID).
		WillReturnError(sql.ErrNoRows)

	legacyMock.ExpectQuery(regexp.QuoteMeta(legacyQuery)).
		WithArgs("tv-1", "tenant-1", systemTemplateTenantID).
		WillReturnRows(sqlmock.NewRows([]string{"docx_storage_key", "schema_storage_key"}).AddRow("s3://legacy-docx", ""))

	docxKey, schemaKey, schemaJSON, err := f.GetPublishedVersion(context.Background(), "tenant-1", "tv-1")
	if err != nil {
		t.Fatalf("GetPublishedVersion: %v", err)
	}
	if docxKey != "s3://legacy-docx" {
		t.Fatalf("docxKey = %q, want s3://legacy-docx", docxKey)
	}
	if schemaKey != "" || schemaJSON != "" {
		t.Fatalf("schemaKey/schemaJSON = %q/%q, want empty", schemaKey, schemaJSON)
	}
	if err := canonicalMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("canonical sql expectations: %v", err)
	}
	if err := legacyMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("legacy sql expectations: %v", err)
	}
	if got := LegacyTemplateReadCount(); got != before+1 {
		t.Fatalf("legacyTemplateReadTotal = %d, want %d (fallback hit must increment the counter)", got, before+1)
	}
}

func TestFanoutTemplateReader_CanonicalError_SurfacesWithoutFallback(t *testing.T) {
	f, canonicalMock, legacyMock, cleanup := newFanoutUnderTest(t)
	defer cleanup()

	before := LegacyTemplateReadCount()

	dbErr := sql.ErrConnDone // a non-not-found DB error
	canonicalMock.ExpectQuery(regexp.QuoteMeta(canonicalQuery)).
		WithArgs("tv-1", "tenant-1", systemTemplateTenantID).
		WillReturnError(dbErr)

	_, _, _, err := f.GetPublishedVersion(context.Background(), "tenant-1", "tv-1")
	if err == nil {
		t.Fatal("GetPublishedVersion: want error, got nil")
	}
	if sql.ErrNoRows == dbErr {
		t.Fatal("test setup invariant broken: dbErr must not be sql.ErrNoRows")
	}

	if err := canonicalMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("canonical sql expectations: %v", err)
	}
	// A real DB error must surface immediately, not be silently masked by a
	// legacy-table lookup — assert the legacy reader was never invoked.
	if err := legacyMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("legacy reader was called on a canonical DB error: %v", err)
	}
	if got := LegacyTemplateReadCount(); got != before {
		t.Fatalf("legacyTemplateReadTotal = %d, want unchanged at %d (no fallback attempted on real DB error)", got, before)
	}
}

func TestFanoutTemplateReader_BothNotFound_ReturnsErrNoRows(t *testing.T) {
	f, canonicalMock, legacyMock, cleanup := newFanoutUnderTest(t)
	defer cleanup()

	before := LegacyTemplateReadCount()

	canonicalMock.ExpectQuery(regexp.QuoteMeta(canonicalQuery)).
		WithArgs("tv-1", "tenant-1", systemTemplateTenantID).
		WillReturnError(sql.ErrNoRows)
	legacyMock.ExpectQuery(regexp.QuoteMeta(legacyQuery)).
		WithArgs("tv-1", "tenant-1", systemTemplateTenantID).
		WillReturnError(sql.ErrNoRows)

	_, _, _, err := f.GetPublishedVersion(context.Background(), "tenant-1", "tv-1")
	if err == nil {
		t.Fatal("GetPublishedVersion: want error, got nil")
	}

	if err := canonicalMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("canonical sql expectations: %v", err)
	}
	if err := legacyMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("legacy sql expectations: %v", err)
	}
	// Both branches exhausted without a legacy row — no fallback "hit" occurred,
	// so the residual-legacy-usage counter must not move.
	if got := LegacyTemplateReadCount(); got != before {
		t.Fatalf("legacyTemplateReadTotal = %d, want unchanged at %d (no successful legacy hit)", got, before)
	}
}
