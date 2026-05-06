package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestStatsByStatus_GroupsCorrectly(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	r := New(db)
	rows := sqlmock.NewRows([]string{"status", "count"}).
		AddRow("draft", int64(2)).
		AddRow("under_review", int64(1))

	mock.ExpectQuery(`SELECT status, COUNT\(\*\) FROM documents WHERE .* GROUP BY status`).
		WithArgs("tenant-1").
		WillReturnRows(rows)

	got, err := r.StatsByStatus(context.Background(), "tenant-1", ListOptions{})
	if err != nil {
		t.Fatalf("StatsByStatus: %v", err)
	}
	if got["draft"] != 2 || got["under_review"] != 1 {
		t.Fatalf("unexpected stats: %#v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStatsByArea_GroupsCorrectly(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	r := New(db)
	rows := sqlmock.NewRows([]string{"area", "count"}).
		AddRow("", int64(1)).
		AddRow("QA", int64(2))

	mock.ExpectQuery(`SELECT COALESCE\(process_area_code_snapshot, ''\) AS area, COUNT\(\*\) FROM documents WHERE .* GROUP BY area`).
		WithArgs("tenant-1").
		WillReturnRows(rows)

	got, err := r.StatsByArea(context.Background(), "tenant-1", ListOptions{})
	if err != nil {
		t.Fatalf("StatsByArea: %v", err)
	}
	if got[""] != 1 || got["QA"] != 2 {
		t.Fatalf("unexpected stats: %#v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
