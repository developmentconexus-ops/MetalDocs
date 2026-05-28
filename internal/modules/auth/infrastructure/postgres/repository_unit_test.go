package postgres

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	authdomain "metaldocs/internal/modules/auth/domain"
)

func TestUpdateUser_NoOpChecksExistenceOnly(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT EXISTS (
	SELECT 1
	FROM metaldocs.auth_identities
	WHERE user_id = $1
)`)).
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	repo := NewRepository(db)
	if err := repo.UpdateUser(context.Background(), authdomain.UpdateUserParams{UserID: "user-1"}); err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateUser_NoOpMissingIdentityReturnsNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT EXISTS (
	SELECT 1
	FROM metaldocs.auth_identities
	WHERE user_id = $1
)`)).
		WithArgs("missing-user").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	repo := NewRepository(db)
	err = repo.UpdateUser(context.Background(), authdomain.UpdateUserParams{UserID: "missing-user"})
	if !errors.Is(err, authdomain.ErrIdentityNotFound) {
		t.Fatalf("UpdateUser error = %v, want ErrIdentityNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestTouchSession_UsesGraceWindowComparison(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	seenAt := time.Date(2026, time.May, 27, 17, 0, 0, 0, time.UTC)

mock.ExpectExec(regexp.QuoteMeta(`
UPDATE metaldocs.auth_sessions
SET last_seen_at = $2
WHERE session_id = $1
  AND last_seen_at < (($2)::timestamptz - INTERVAL '30 seconds')
`)).
		WithArgs("session-1", seenAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewRepository(db)
	if err := repo.TouchSession(context.Background(), "session-1", seenAt); err != nil {
		t.Fatalf("TouchSession: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
