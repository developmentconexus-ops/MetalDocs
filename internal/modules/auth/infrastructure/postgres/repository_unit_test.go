package postgres

import (
	"context"
	"database/sql/driver"
	"errors"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"

	authdomain "metaldocs/internal/modules/auth/domain"
)

func TestCreateUserTx_UniqueViolationReturnsDomainError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	mock.ExpectBegin()
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}
	mock.ExpectExec(regexp.QuoteMeta(`
INSERT INTO metaldocs.auth_identities (user_id, username, email, display_name, is_active, password_hash, password_algo, must_change_password, last_login_at, failed_login_attempts, locked_until, created_at, updated_at)
VALUES ($1, $2, NULLIF($3, ''), $4, $5, $6, $7, $8, NULL, 0, NULL, NOW(), NOW())
`)).
		WithArgs("alice", "alice", "alice@example.com", "Alice", true, "hash", "bcrypt", true).
		WillReturnError(&pq.Error{Code: "23505"})

	err = repo.CreateUserTx(context.Background(), tx, authdomain.CreateUserParams{
		UserID:             "alice",
		Username:           "alice",
		Email:              "alice@example.com",
		DisplayName:        "Alice",
		PasswordHash:       "hash",
		PasswordAlgo:       "bcrypt",
		MustChangePassword: true,
		IsActive:           true,
	})
	if !errors.Is(err, authdomain.ErrUserAlreadyExists) {
		t.Fatalf("expected ErrUserAlreadyExists, got %v", err)
	}
	_ = tx.Rollback()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestBootstrapAdmin_OnConflictDoNothingReturnsFalse(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`
INSERT INTO metaldocs.auth_identities (user_id, username, email, display_name, is_active, password_hash, password_algo, must_change_password, last_login_at, failed_login_attempts, locked_until, created_at, updated_at)
VALUES ($1, $2, NULLIF($3, ''), $4, TRUE, $5, $6, $7, NULL, 0, NULL, NOW(), NOW())
ON CONFLICT (user_id)
DO NOTHING
`)).
		WithArgs("admin", "admin", "admin@example.com", "Admin", "hash", "bcrypt", true).
		WillReturnResult(driver.RowsAffected(0))
	mock.ExpectCommit()

	created, err := repo.BootstrapAdmin(context.Background(), authdomain.BootstrapAdminParams{
		UserID:             "admin",
		Username:           "admin",
		Email:              "admin@example.com",
		DisplayName:        "Admin",
		PasswordHash:       "hash",
		PasswordAlgo:       "bcrypt",
		MustChangePassword: true,
	})
	if err != nil {
		t.Fatalf("BootstrapAdmin: %v", err)
	}
	if created {
		t.Fatal("expected created=false when ON CONFLICT DO NOTHING affects zero rows")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestFindIdentityByIdentifier_NormalizesLookupArgument(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT i.user_id, i.username, COALESCE(i.email, ''), i.display_name, i.password_hash, i.password_algo,
       i.must_change_password, i.last_login_at, i.failed_login_attempts, i.locked_until, i.is_active,
       i.created_at, i.updated_at
FROM metaldocs.auth_identities i
WHERE lower(i.username) = $1
UNION ALL
SELECT i.user_id, i.username, COALESCE(i.email, ''), i.display_name, i.password_hash, i.password_algo,
       i.must_change_password, i.last_login_at, i.failed_login_attempts, i.locked_until, i.is_active,
       i.created_at, i.updated_at
FROM metaldocs.auth_identities i
WHERE i.email IS NOT NULL
  AND lower(i.email) = $1
  AND lower(i.username) <> $1
LIMIT 1
`)).
		WithArgs("alice@example.com").
		WillReturnRows(sqlmock.NewRows([]string{
			"user_id", "username", "email", "display_name", "password_hash", "password_algo",
			"must_change_password", "last_login_at", "failed_login_attempts", "locked_until", "is_active",
			"created_at", "updated_at",
		}).AddRow("u1", "alice", "alice@example.com", "Alice", "hash", "bcrypt", false, nil, 0, nil, true, time.Now(), time.Now()))

	_, err = repo.FindIdentityByIdentifier(context.Background(), "  Alice@Example.com  ")
	if err != nil {
		t.Fatalf("FindIdentityByIdentifier: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestTouchSession_NoOpInsideGraceWindowReturnsNil(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	seenAt := time.Date(2026, 5, 25, 12, 0, 31, 0, time.UTC)
	mock.ExpectExec(regexp.QuoteMeta(`
UPDATE metaldocs.auth_sessions
SET last_seen_at = $2
WHERE session_id = $1
  AND last_seen_at < $2 - INTERVAL '30 seconds'
`)).
		WithArgs("session-1", seenAt).
		WillReturnResult(driver.RowsAffected(0))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS (SELECT 1 FROM metaldocs.auth_sessions WHERE session_id = $1)`)).
		WithArgs("session-1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	if err := repo.TouchSession(context.Background(), "session-1", seenAt); err != nil {
		t.Fatalf("TouchSession: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestTouchSession_MissingSessionReturnsNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	seenAt := time.Date(2026, 5, 25, 12, 0, 31, 0, time.UTC)
	mock.ExpectExec(regexp.QuoteMeta(`
UPDATE metaldocs.auth_sessions
SET last_seen_at = $2
WHERE session_id = $1
  AND last_seen_at < $2 - INTERVAL '30 seconds'
`)).
		WithArgs("missing-session", seenAt).
		WillReturnResult(driver.RowsAffected(0))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS (SELECT 1 FROM metaldocs.auth_sessions WHERE session_id = $1)`)).
		WithArgs("missing-session").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	err = repo.TouchSession(context.Background(), "missing-session", seenAt)
	if !errors.Is(err, authdomain.ErrSessionNotFound) {
		t.Fatalf("TouchSession error = %v, want ErrSessionNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
