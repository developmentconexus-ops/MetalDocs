package repository_test

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"metaldocs/internal/modules/documents/repository"
)

type recordingDBTX struct {
	execCalled bool
}

func (r *recordingDBTX) ExecContext(_ context.Context, _ string, _ ...any) (sql.Result, error) {
	r.execCalled = true
	return sqlmock.NewResult(1, 1), nil
}

func (r *recordingDBTX) QueryRowContext(_ context.Context, _ string, _ ...any) *sql.Row {
	return nil
}

func TestUpsertValue_UsesTxWhenProvided(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	tx := &recordingDBTX{}
	repo := repository.NewFillInRepository(db)
	value := "computed"

	if err := repo.UpsertValue(ctx, repository.PlaceholderValue{
		TenantID:      "00000000-0000-0000-0000-000000000001",
		RevisionID:    "00000000-0000-0000-0000-000000000002",
		PlaceholderID: "computed-placeholder",
		ValueText:     &value,
		Source:        "computed",
	}, tx); err != nil {
		t.Fatalf("UpsertValue: %v", err)
	}

	if !tx.execCalled {
		t.Fatal("expected UpsertValue to execute through provided DBTX")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB executor call: %v", err)
	}
}

func TestUpsertValue_FallsBackToDB(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO document_placeholder_values")).
		WithArgs(
			"00000000-0000-0000-0000-000000000001",
			"00000000-0000-0000-0000-000000000002",
			"user-placeholder",
			sqlmock.AnyArg(),
			nil,
			"user",
			nil,
			nil,
			[]byte(nil),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := repository.NewFillInRepository(db)
	value := "user value"

	if err := repo.UpsertValue(ctx, repository.PlaceholderValue{
		TenantID:      "00000000-0000-0000-0000-000000000001",
		RevisionID:    "00000000-0000-0000-0000-000000000002",
		PlaceholderID: "user-placeholder",
		ValueText:     &value,
		Source:        "user",
	}); err != nil {
		t.Fatalf("UpsertValue: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
