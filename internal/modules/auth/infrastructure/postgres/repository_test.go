//go:build integration

package postgres_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"

	authdomain "metaldocs/internal/modules/auth/domain"
	authpostgres "metaldocs/internal/modules/auth/infrastructure/postgres"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// TestUpdateUserResetLockState verifies that UpdateUser runs the UPDATE even
// when only ResetLockState is set (no Email/NewPasswordHash/MustChangePassword).
func TestUpdateUserResetLockState(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	// Insert a locked test user directly.
	userID := fmt.Sprintf("test-unlock-%d", time.Now().UnixNano())
	lockedUntil := time.Now().Add(10 * time.Minute)
	_, err := db.ExecContext(ctx, `
INSERT INTO metaldocs.auth_identities
  (user_id, username, email, display_name, is_active, password_hash, password_algo,
   must_change_password, failed_login_attempts, locked_until, created_at, updated_at)
VALUES ($1, $1, NULL, 'Test Unlock', TRUE, 'hash', 'bcrypt', FALSE, 3, $2, NOW(), NOW())
`, userID, lockedUntil)
	if err != nil {
		t.Fatalf("insert test user: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM metaldocs.auth_identities WHERE user_id = $1`, userID)
	})

	repo := authpostgres.NewRepository(db)

	err = repo.UpdateUser(ctx, authdomain.UpdateUserParams{
		UserID:         userID,
		ResetLockState: true,
	})
	if err != nil {
		t.Fatalf("UpdateUser with ResetLockState=true: %v", err)
	}

	var failedAttempts int
	var lockedUntilOut sql.NullTime
	if err := db.QueryRowContext(ctx, `
SELECT failed_login_attempts, locked_until
FROM metaldocs.auth_identities WHERE user_id = $1
`, userID).Scan(&failedAttempts, &lockedUntilOut); err != nil {
		t.Fatalf("select after update: %v", err)
	}

	if failedAttempts != 0 {
		t.Errorf("expected failed_login_attempts=0, got %d", failedAttempts)
	}
	if lockedUntilOut.Valid {
		t.Errorf("expected locked_until=NULL, got %v", lockedUntilOut.Time)
	}
}
