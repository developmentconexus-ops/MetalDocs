package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestListDocuments_ExcludesArchivedByDefault(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	r := New(db)

	mock.ExpectQuery(`archived_at IS NULL`).
		WithArgs("tenant-1").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "tenant_id", "template_version_id", "name", "status", "form_data_json",
			"current_revision_id", "active_session_id", "archived_at",
			"created_at", "updated_at", "created_by", "controlled_document_id", "code",
		}))

	if _, err := r.ListDocuments(context.Background(), "tenant-1"); err != nil {
		t.Fatalf("ListDocuments: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestListDocumentsForUser_ExcludesArchivedByDefault(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	r := New(db)

	mock.ExpectQuery(`archived_at IS NULL`).
		WithArgs("tenant-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "tenant_id", "template_version_id", "name", "status", "form_data_json",
			"current_revision_id", "active_session_id", "archived_at",
			"created_at", "updated_at", "created_by", "controlled_document_id", "code",
		}))

	if _, err := r.ListDocumentsForUser(context.Background(), "tenant-1", "user-1"); err != nil {
		t.Fatalf("ListDocumentsForUser: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
