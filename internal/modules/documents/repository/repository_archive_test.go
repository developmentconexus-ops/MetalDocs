package repository

import (
	"context"
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

	mock.ExpectExec(`UPDATE metaldocs\.documents`).
		WithArgs("tenant-1", "doc-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := r.MarkArchived(context.Background(), "tenant-1", "doc-1", "user-1"); err != nil {
		t.Fatalf("MarkArchived: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUnarchive_ClearsTimestamp(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	r := New(db)

	mock.ExpectExec(`UPDATE metaldocs\.documents`).
		WithArgs("tenant-1", "doc-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := r.Unarchive(context.Background(), "tenant-1", "doc-1", "user-1"); err != nil {
		t.Fatalf("Unarchive: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
