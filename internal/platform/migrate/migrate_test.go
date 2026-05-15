package migrate

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestApply_IgnoresNonNumericBaselineMarker(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	dir := t.TempDir()
	writeMigrationFile(t, dir, "0001_first.sql", "SELECT 1;")
	writeMigrationFile(t, dir, "0002_second.sql", "SELECT 2;")

	mock.ExpectQuery(`SELECT version FROM public\.schema_migrations`).
		WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("baseline_marker"))
	mock.ExpectExec("SELECT 1;").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("SELECT 2;").WillReturnResult(sqlmock.NewResult(0, 0))

	if err := Apply(context.Background(), db, dir, slog.Default()); err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestApply_PreservesNumericHighWaterMark(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	dir := t.TempDir()
	writeMigrationFile(t, dir, "0001_first.sql", "SELECT 1;")
	writeMigrationFile(t, dir, "0002_second.sql", "SELECT 2;")
	writeMigrationFile(t, dir, "0003_third.sql", "SELECT 3;")

	mock.ExpectQuery(`SELECT version FROM public\.schema_migrations`).
		WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("0002"))
	mock.ExpectExec("SELECT 3;").WillReturnResult(sqlmock.NewResult(0, 0))

	if err := Apply(context.Background(), db, dir, slog.Default()); err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func writeMigrationFile(t *testing.T, dir, name, body string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
