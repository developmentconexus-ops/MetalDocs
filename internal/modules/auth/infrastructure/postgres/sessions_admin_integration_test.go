//go:build integration

package postgres_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	_ "github.com/lib/pq"

	authdomain "metaldocs/internal/modules/auth/domain"
	authpostgres "metaldocs/internal/modules/auth/infrastructure/postgres"
)

// TestListActiveSessions_NoIamUsersJoin proves the M4/F4.4 boundary fix: the
// query no longer INNER JOINs metaldocs.iam_users. A session whose user has an
// auth_identities row but NO iam_users membership row is still returned (the old
// JOIN would have dropped it), and tenant scoping is preserved via
// auth_sessions.tenant_id alone.
func TestListActiveSessions_NoIamUsersJoin(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	repo := authpostgres.NewRepository(db, nil)

	// Seeded tenants (FK fk_auth_sessions_tenant requires real tenant rows).
	tenantA := "ffffffff-ffff-ffff-ffff-ffffffffffff"
	tenantB := "e1f4f8d7-73bd-44bd-b8f5-b01f9cca53e1"
	stamp := time.Now().UnixNano()
	userNoIam := fmt.Sprintf("f44-no-iam-user-%d", stamp)
	userOther := fmt.Sprintf("f44-other-user-%d", stamp)
	sessionA := fmt.Sprintf("f44-session-a-%d", stamp)
	sessionB := fmt.Sprintf("f44-session-b-%d", stamp)
	now := time.Now().UTC()

	// auth_identities rows satisfy the auth_sessions FK; NEITHER user gets an
	// iam_users row, so under the old INNER JOIN both sessions would vanish.
	for _, uid := range []string{userNoIam, userOther} {
		if _, err := db.ExecContext(ctx, `
INSERT INTO metaldocs.auth_identities
  (user_id, username, email, display_name, is_active, password_hash, password_algo,
   must_change_password, failed_login_attempts, locked_until, created_at, updated_at)
VALUES ($1, $1, NULL, 'F44 No IAM', TRUE, 'hash', 'bcrypt', FALSE, 0, NULL, NOW(), NOW())
`, uid); err != nil {
			t.Fatalf("insert auth identity %s: %v", uid, err)
		}
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM metaldocs.auth_sessions WHERE session_id IN ($1, $2)`, sessionA, sessionB)
		_, _ = db.ExecContext(ctx, `DELETE FROM metaldocs.auth_identities WHERE user_id IN ($1, $2)`, userNoIam, userOther)
	})

	for _, s := range []struct {
		sessionID string
		userID    string
		tenantID  string
	}{
		{sessionA, userNoIam, tenantA},
		{sessionB, userOther, tenantB},
	} {
		if _, err := db.ExecContext(ctx, `
INSERT INTO metaldocs.auth_sessions
  (session_id, user_id, tenant_id, created_at, expires_at, revoked_at, ip_address, user_agent, last_seen_at)
VALUES ($1, $2, $3::uuid, $4, $5, NULL, '127.0.0.1', 'test', $6)
`, s.sessionID, s.userID, s.tenantID, now.Add(-time.Minute), now.Add(time.Hour), now); err != nil {
			t.Fatalf("insert auth session %s: %v", s.sessionID, err)
		}
	}

	items, err := repo.ListActiveSessions(ctx, authdomain.SessionAdminQuery{TenantID: tenantA})
	if err != nil {
		t.Fatalf("ListActiveSessions: %v", err)
	}

	var found bool
	for _, it := range items {
		if it.SessionID == sessionB {
			t.Fatalf("tenant scope broken: tenant-B session leaked into tenant-A result")
		}
		if it.SessionID == sessionA {
			found = true
			if it.UserID != userNoIam {
				t.Fatalf("session-a user_id = %q, want %q", it.UserID, userNoIam)
			}
		}
	}
	if !found {
		t.Fatalf("session-a (user with no iam_users row) was not returned — INNER JOIN still present? got %d items", len(items))
	}
}
