package infrastructure

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestPostgresControlledDocumentRepository_GetByIDLoadsVisibilityGrants(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewPostgresControlledDocumentRepository(db)
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT id::text, tenant_id::text, profile_code, process_area_code, department_code,
       code, sequence_num, title, owner_user_id, coalesce(override_template_version_id::text, ''),
       visibility_scope, status, created_at, updated_at
FROM controlled_documents
WHERE tenant_id = $1 AND id = $2`)).
		WithArgs("tenant-1", "cd-1").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "tenant_id", "profile_code", "process_area_code", "department_code",
			"code", "sequence_num", "title", "owner_user_id", "override_template_version_id",
			"visibility_scope", "status", "created_at", "updated_at",
		}).AddRow("cd-1", "tenant-1", "POP", "QA", nil, "POP-QA-001", 1, "Procedure", "owner-1", "", "restricted", "active", now, now))

	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT area_code
FROM controlled_document_area_grants
WHERE tenant_id = $1 AND controlled_document_id = $2
ORDER BY area_code`)).
		WithArgs("tenant-1", "cd-1").
		WillReturnRows(sqlmock.NewRows([]string{"area_code"}).AddRow("QA").AddRow("RH"))

	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT user_id
FROM controlled_document_user_grants
WHERE tenant_id = $1 AND controlled_document_id = $2
ORDER BY user_id`)).
		WithArgs("tenant-1", "cd-1").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow("user-2"))

	doc, err := repo.GetByID(context.Background(), "tenant-1", "cd-1")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if doc.Visibility.Scope != "restricted" {
		t.Fatalf("scope = %q, want restricted", doc.Visibility.Scope)
	}
	if got := doc.Visibility.AreaCodes; len(got) != 2 || got[0] != "QA" || got[1] != "RH" {
		t.Fatalf("area grants = %#v, want [QA RH]", got)
	}
	if got := doc.Visibility.UserIDs; len(got) != 1 || got[0] != "user-2" {
		t.Fatalf("user grants = %#v, want [user-2]", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestPostgresControlledDocumentRepository_UpdateStatus_RowsAffectedError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewPostgresControlledDocumentRepository(db)
	rowsErr := errors.New("rows affected failed")
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE controlled_documents SET status = $1, updated_at = $2 WHERE tenant_id = $3 AND id = $4`)).
		WithArgs("obsolete", sqlmock.AnyArg(), "tenant-1", "cd-1").
		WillReturnResult(sqlmock.NewErrorResult(rowsErr))

	err = repo.UpdateStatus(context.Background(), "tenant-1", "cd-1", "obsolete", time.Now().UTC())
	if !errors.Is(err, rowsErr) {
		t.Fatalf("expected rows affected error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
