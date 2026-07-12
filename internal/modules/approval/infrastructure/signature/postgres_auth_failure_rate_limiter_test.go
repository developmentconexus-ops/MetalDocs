package signature

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// newMockLimiter returns a PostgresAuthFailureRateLimiter backed by a sqlmock DB.
func newMockLimiter(t *testing.T) (*PostgresAuthFailureRateLimiter, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return NewPostgresAuthFailureRateLimiter(db), mock
}

// TestPostgresLimiter_Allow_NoRow — no counter row → allowed.
func TestPostgresLimiter_Allow_NoRow(t *testing.T) {
	l, mock := newMockLimiter(t)
	mock.ExpectQuery(`SELECT fail_count`).
		WithArgs("u1").
		WillReturnError(sql.ErrNoRows)

	ok, err := l.Allow(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("want allowed when no row exists")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// TestPostgresLimiter_Allow_BelowThreshold — count < maxFailures → allowed.
func TestPostgresLimiter_Allow_BelowThreshold(t *testing.T) {
	l, mock := newMockLimiter(t)
	rows := sqlmock.NewRows([]string{"fail_count", "window_start"}).
		AddRow(3, time.Now().UTC())
	mock.ExpectQuery(`SELECT fail_count`).
		WithArgs("u1").
		WillReturnRows(rows)

	ok, err := l.Allow(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("want allowed when count < maxFailures")
	}
}

// TestPostgresLimiter_Allow_AtThreshold — count == maxFailures, window active → blocked.
func TestPostgresLimiter_Allow_AtThreshold(t *testing.T) {
	l, mock := newMockLimiter(t)
	rows := sqlmock.NewRows([]string{"fail_count", "window_start"}).
		AddRow(maxFailures, time.Now().UTC())
	mock.ExpectQuery(`SELECT fail_count`).
		WithArgs("u1").
		WillReturnRows(rows)

	ok, err := l.Allow(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("want blocked when count >= maxFailures within window")
	}
}

// TestPostgresLimiter_Allow_ExpiredWindow — count >= maxFailures but window expired → allowed.
func TestPostgresLimiter_Allow_ExpiredWindow(t *testing.T) {
	l, mock := newMockLimiter(t)
	stale := time.Now().UTC().Add(-(windowDur + time.Second))
	rows := sqlmock.NewRows([]string{"fail_count", "window_start"}).
		AddRow(maxFailures, stale)
	mock.ExpectQuery(`SELECT fail_count`).
		WithArgs("u1").
		WillReturnRows(rows)

	ok, err := l.Allow(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("want allowed when window has expired")
	}
}

// TestPostgresLimiter_RecordFailure — executes UPSERT with three timestamptz
// args: ($1=actorID, $2=now, $3=windowExpiredBefore). No time.Duration is sent
// to Postgres — precomputed in Go to avoid the "timestamptz + bigint" type error
// that pgx/v5/stdlib causes when encoding time.Duration as bigint.
func TestPostgresLimiter_RecordFailure(t *testing.T) {
	l, mock := newMockLimiter(t)
	// $1 = actorID (string), $2 = now (time.Time/timestamptz),
	// $3 = windowExpiredBefore (time.Time/timestamptz).
	mock.ExpectExec(`INSERT INTO public.auth_failure_counters`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := l.RecordFailure(context.Background(), "u1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// TestPostgresLimiter_Reset — executes DELETE without error.
func TestPostgresLimiter_Reset(t *testing.T) {
	l, mock := newMockLimiter(t)
	mock.ExpectExec(`DELETE FROM public.auth_failure_counters`).
		WithArgs("u1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := l.Reset(context.Background(), "u1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// TestPostgresLimiter_Reset_NoRow — DELETE on missing row is not an error.
func TestPostgresLimiter_Reset_NoRow(t *testing.T) {
	l, mock := newMockLimiter(t)
	mock.ExpectExec(`DELETE FROM public.auth_failure_counters`).
		WithArgs("u1").
		WillReturnResult(sqlmock.NewResult(0, 0))

	if err := l.Reset(context.Background(), "u1"); err != nil {
		t.Fatalf("unexpected error on reset with no row: %v", err)
	}
}

// TestPostgresLimiter_ImplementsInterface verifies the type satisfies the port.
func TestPostgresLimiter_ImplementsInterface(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()
	var _ AuthFailureRateLimiter = NewPostgresAuthFailureRateLimiter(db)
}
