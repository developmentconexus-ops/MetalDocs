package migrate

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestApply_IgnoresNonNumericBaselineMarker(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	dir := t.TempDir()
	writeMigrationFile(t, dir, "0001_first.sql", "BEGIN;\nSELECT 1;\nCOMMIT;\n")
	writeMigrationFile(t, dir, "0002_second.sql", "BEGIN;\nSELECT 2;\nCOMMIT;\n")

	mock.ExpectExec(`SELECT pg_advisory_lock\(\$1\)`).
		WithArgs(migrateLockKey).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(`SELECT version FROM public\.schema_migrations`).
		WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("baseline_marker"))
	mock.ExpectExec("BEGIN;\nSELECT 1;\nCOMMIT;\n").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("BEGIN;\nSELECT 2;\nCOMMIT;\n").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`SELECT pg_advisory_unlock\(\$1\)`).
		WithArgs(migrateLockKey).
		WillReturnResult(sqlmock.NewResult(0, 0))

	if err := Apply(context.Background(), db, dir, slog.Default()); err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestApply_SkipsOnlyAppliedVersions(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	dir := t.TempDir()
	writeMigrationFile(t, dir, "0001_first.sql", "BEGIN;\nSELECT 1;\nCOMMIT;\n")
	writeMigrationFile(t, dir, "0002_second.sql", "BEGIN;\nSELECT 2;\nCOMMIT;\n")
	writeMigrationFile(t, dir, "0003_third.sql", "BEGIN;\nSELECT 3;\nCOMMIT;\n")

	mock.ExpectExec(`SELECT pg_advisory_lock\(\$1\)`).
		WithArgs(migrateLockKey).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(`SELECT version FROM public\.schema_migrations`).
		WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("0002"))
	mock.ExpectExec("BEGIN;\nSELECT 1;\nCOMMIT;\n").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("BEGIN;\nSELECT 3;\nCOMMIT;\n").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`SELECT pg_advisory_unlock\(\$1\)`).
		WithArgs(migrateLockKey).
		WillReturnResult(sqlmock.NewResult(0, 0))

	if err := Apply(context.Background(), db, dir, slog.Default()); err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestLoadApplied_AllowsMissingSchemaMigrationsTable(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	mock.ExpectQuery(`SELECT version FROM public\.schema_migrations`).
		WillReturnError(&pgconn.PgError{Code: "42P01"})

	applied, err := loadApplied(context.Background(), db)
	if err != nil {
		t.Fatalf("loadApplied returned error: %v", err)
	}
	if len(applied) != 0 {
		t.Fatalf("loadApplied returned %d applied versions, want 0", len(applied))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestLoadApplied_PropagatesQueryErrors(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	wantErr := errors.New("permission denied")
	mock.ExpectQuery(`SELECT version FROM public\.schema_migrations`).
		WillReturnError(wantErr)

	_, err = loadApplied(context.Background(), db)
	if !errors.Is(err, wantErr) {
		t.Fatalf("loadApplied error = %v, want wrapping %v", err, wantErr)
	}
	if !strings.Contains(err.Error(), "load applied") {
		t.Fatalf("loadApplied error = %q, want operation context", err.Error())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestApply_RejectsMigrationWithoutExplicitTransactionGuard(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	dir := t.TempDir()
	writeMigrationFile(t, dir, "0001_first.sql", "SELECT 1;")

	mock.ExpectExec(`SELECT pg_advisory_lock\(\$1\)`).
		WithArgs(migrateLockKey).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(`SELECT version FROM public\.schema_migrations`).
		WillReturnRows(sqlmock.NewRows([]string{"version"}))
	mock.ExpectExec(`SELECT pg_advisory_unlock\(\$1\)`).
		WithArgs(migrateLockKey).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = Apply(context.Background(), db, dir, slog.Default())
	if err == nil || !strings.Contains(err.Error(), "explicit BEGIN") {
		t.Fatalf("Apply error = %v, want explicit BEGIN failure", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

// TestApplyGrants_RunsEveryFileSortedWithoutLedger pins the grants stage
// contract: every *.sql in the directory runs, in lexical order, on every call
// — no schema_migrations lookup and no ledger write, so an edit to the
// privilege posture reaches a long-lived volume that never re-runs initdb.
func TestApplyGrants_RunsEveryFileSortedWithoutLedger(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	dir := t.TempDir()
	writeMigrationFile(t, dir, "0002_second.sql", "BEGIN;\nSELECT 2;\nCOMMIT;\n")
	writeMigrationFile(t, dir, "0001_first.sql", "BEGIN;\nSELECT 1;\nCOMMIT;\n")
	writeMigrationFile(t, dir, "README.md", "not sql\n")

	mock.ExpectExec(`SELECT pg_advisory_lock\(\$1\)`).
		WithArgs(migrateLockKey).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("BEGIN;\nSELECT 1;\nCOMMIT;\n").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("BEGIN;\nSELECT 2;\nCOMMIT;\n").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`SELECT pg_advisory_unlock\(\$1\)`).
		WithArgs(migrateLockKey).
		WillReturnResult(sqlmock.NewResult(0, 0))

	if err := ApplyGrants(context.Background(), db, dir, slog.Default()); err != nil {
		t.Fatalf("ApplyGrants returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}

	// Second call re-runs the same files: the stage is deliberately ledger-less.
	mock.ExpectExec(`SELECT pg_advisory_lock\(\$1\)`).
		WithArgs(migrateLockKey).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("BEGIN;\nSELECT 1;\nCOMMIT;\n").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("BEGIN;\nSELECT 2;\nCOMMIT;\n").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`SELECT pg_advisory_unlock\(\$1\)`).
		WithArgs(migrateLockKey).
		WillReturnResult(sqlmock.NewResult(0, 0))

	if err := ApplyGrants(context.Background(), db, dir, slog.Default()); err != nil {
		t.Fatalf("second ApplyGrants returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations (second run): %v", err)
	}
}

// TestApplyGrants_FailsClosedOnMissingOrEmptyDir pins the no-fallback rule: an
// API that cannot find the privilege stage must refuse to boot rather than
// start with an unknown ACL posture.
func TestApplyGrants_FailsClosedOnMissingOrEmptyDir(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	missing := filepath.Join(t.TempDir(), "does-not-exist")
	if err := ApplyGrants(context.Background(), db, missing, slog.Default()); err == nil ||
		!strings.Contains(err.Error(), "read grants dir") {
		t.Fatalf("ApplyGrants(missing dir) error = %v, want read-grants-dir failure", err)
	}

	empty := t.TempDir()
	if err := ApplyGrants(context.Background(), db, empty, slog.Default()); err == nil ||
		!strings.Contains(err.Error(), "no .sql files") {
		t.Fatalf("ApplyGrants(empty dir) error = %v, want mandatory-stage failure", err)
	}

	// Neither path may reach the database.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

// TestApplyGrants_RejectsFileWithoutExplicitTransactionGuard keeps the grants
// stage under the same all-or-nothing transaction rule as migrations: a file
// that half-applies would leave a partial privilege posture behind.
func TestApplyGrants_RejectsFileWithoutExplicitTransactionGuard(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	dir := t.TempDir()
	writeMigrationFile(t, dir, "0001_grants.sql", "GRANT SELECT ON t TO r;")

	mock.ExpectExec(`SELECT pg_advisory_lock\(\$1\)`).
		WithArgs(migrateLockKey).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`SELECT pg_advisory_unlock\(\$1\)`).
		WithArgs(migrateLockKey).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = ApplyGrants(context.Background(), db, dir, slog.Default())
	if err == nil || !strings.Contains(err.Error(), "explicit BEGIN") {
		t.Fatalf("ApplyGrants error = %v, want explicit BEGIN failure", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

// TestCommittedGrantsStageParses guards the real db/grants bundle against the
// same transaction-guard rule the runtime enforces, so a hand-edit that drops
// BEGIN/COMMIT is caught here instead of at API startup.
func TestCommittedGrantsStageParses(t *testing.T) {
	t.Parallel()

	dir := filepath.Join("..", "..", "..", "db", "grants")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}
	found := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		found++
		body, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		if err := requireExplicitTransactionGuard(string(body)); err != nil {
			t.Errorf("db/grants/%s: %v", e.Name(), err)
		}
	}
	if found == 0 {
		t.Fatal("db/grants holds no .sql files; ApplyGrants fails closed on that")
	}
}

func TestRequireExplicitTransactionGuard_AllowsCommentPrefixedTransaction(t *testing.T) {
	t.Parallel()

	body := "-- heading\n/* block */\nBEGIN;\nSELECT 1;\nCOMMIT;\n"
	if err := requireExplicitTransactionGuard(body); err != nil {
		t.Fatalf("requireExplicitTransactionGuard returned error: %v", err)
	}
}

func writeMigrationFile(t *testing.T, dir, name, body string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
