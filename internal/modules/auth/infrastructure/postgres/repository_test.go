//go:build integration

package postgres_test

import (
	"context"
	"database/sql"
	"errors"
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

func TestUpdateUserMissingIdentityReturnsNotFound(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	repo := authpostgres.NewRepository(db)

	err := repo.UpdateUser(ctx, authdomain.UpdateUserParams{
		UserID:      fmt.Sprintf("missing-user-%d", time.Now().UnixNano()),
		DisplayName: ptr("Missing"),
	})
	if !errors.Is(err, authdomain.ErrIdentityNotFound) {
		t.Fatalf("UpdateUser error = %v, want ErrIdentityNotFound", err)
	}
}

func TestListOnlineUsers_FiltersByTenant(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	repo := authpostgres.NewRepository(db)

	tenantA := "11111111-1111-1111-1111-111111111111"
	tenantB := "22222222-2222-2222-2222-222222222222"
	userA := fmt.Sprintf("tenant-a-user-%d", time.Now().UnixNano())
	userB := fmt.Sprintf("tenant-b-user-%d", time.Now().UnixNano())
	sessionA := fmt.Sprintf("session-a-%d", time.Now().UnixNano())
	sessionB := fmt.Sprintf("session-b-%d", time.Now().UnixNano())
	now := time.Now().UTC()

	for _, identity := range []struct {
		userID      string
		username    string
		displayName string
	}{
		{userA, userA, "Tenant A"},
		{userB, userB, "Tenant B"},
	} {
		_, err := db.ExecContext(ctx, `
INSERT INTO metaldocs.auth_identities
  (user_id, username, email, display_name, is_active, password_hash, password_algo,
   must_change_password, failed_login_attempts, locked_until, created_at, updated_at)
VALUES ($1, $2, NULL, $3, TRUE, 'hash', 'bcrypt', FALSE, 0, NULL, NOW(), NOW())
`, identity.userID, identity.username, identity.displayName)
		if err != nil {
			t.Fatalf("insert auth identity %s: %v", identity.userID, err)
		}
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM metaldocs.auth_sessions WHERE session_id IN ($1, $2)`, sessionA, sessionB)
		_, _ = db.ExecContext(ctx, `DELETE FROM metaldocs.auth_identities WHERE user_id IN ($1, $2)`, userA, userB)
	})

	for _, session := range []struct {
		sessionID  string
		userID     string
		tenantID   string
		lastSeenAt time.Time
	}{
		{sessionA, userA, tenantA, now},
		{sessionB, userB, tenantB, now},
	} {
		_, err := db.ExecContext(ctx, `
INSERT INTO metaldocs.auth_sessions
  (session_id, user_id, tenant_id, created_at, expires_at, revoked_at, ip_address, user_agent, last_seen_at)
VALUES ($1, $2, $3::uuid, $4, $5, NULL, '127.0.0.1', 'test', $6)
`, session.sessionID, session.userID, session.tenantID, now.Add(-time.Minute), now.Add(time.Hour), session.lastSeenAt)
		if err != nil {
			t.Fatalf("insert auth session %s: %v", session.sessionID, err)
		}
	}

	items, err := repo.ListOnlineUsers(ctx, tenantA, now.Add(-10*time.Minute))
	if err != nil {
		t.Fatalf("ListOnlineUsers: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 online user for tenant A, got %d", len(items))
	}
	if items[0].UserID != userA {
		t.Fatalf("expected tenant A user %q, got %q", userA, items[0].UserID)
	}
}

func ptr(value string) *string {
	return &value
}
