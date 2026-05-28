package repository

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMarkArchived_StampsTimestampWithoutStatusChange(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	r := New(db)

	mock.ExpectBegin()
	expectDocumentEditAuthz(t, mock)
	mock.ExpectExec(`UPDATE public\.documents`).
		WithArgs("tenant-1", "doc-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := r.MarkArchived(context.Background(), "tenant-1", "doc-1", "user-1"); err != nil {
		t.Fatalf("MarkArchived: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMarkArchived_NoRowsReturnsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	r := New(db)

	mock.ExpectBegin()
	expectDocumentEditAuthz(t, mock)
	mock.ExpectExec(`UPDATE public\.documents`).
		WithArgs("tenant-1", "doc-1").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err = r.MarkArchived(context.Background(), "tenant-1", "doc-1", "user-1")
	if err == nil || !strings.Contains(err.Error(), "not found or already in target state") {
		t.Fatalf("MarkArchived error = %v, want not found/target state", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUnarchive_ClearsTimestamp(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	r := New(db)

	mock.ExpectBegin()
	expectDocumentEditAuthz(t, mock)
	mock.ExpectExec(`UPDATE public\.documents`).
		WithArgs("tenant-1", "doc-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := r.Unarchive(context.Background(), "tenant-1", "doc-1", "user-1"); err != nil {
		t.Fatalf("Unarchive: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUnarchive_NoRowsReturnsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	r := New(db)

	mock.ExpectBegin()
	expectDocumentEditAuthz(t, mock)
	mock.ExpectExec(`UPDATE public\.documents`).
		WithArgs("tenant-1", "doc-1").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err = r.Unarchive(context.Background(), "tenant-1", "doc-1", "user-1")
	if err == nil || !strings.Contains(err.Error(), "not found or already in target state") {
		t.Fatalf("Unarchive error = %v, want not found/target state", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func expectDocumentEditAuthz(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	mock.ExpectExec(`(?s)SELECT\s+set_config\('metaldocs\.tenant_id', \$1, true\),\s*set_config\('metaldocs\.actor_id', \$2, true\)`).
		WithArgs("tenant-1", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.actor_id', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow("user-1"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.tenant_id', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow("tenant-1"))
	mock.ExpectQuery(`SELECT EXISTS`).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.asserted_caps', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow(""))
	mock.ExpectExec(regexp.QuoteMeta("SELECT set_config('metaldocs.asserted_caps'")).
		WillReturnResult(sqlmock.NewResult(0, 0))
}
