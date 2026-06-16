//go:build integration

package postgres_test

// Live integration proof for the M4/F4.6 boundary fix: the security repository no
// longer JOINs metaldocs.iam_users for display names or tenant scope. Tenant scope
// for the auth_identities-coupled methods comes from the iam-owned TenantUserReader
// (F4.5, = ANY(member-ids)); display names from the iam-owned UserDisplayNameReader
// (F4.1). ListNewDeviceLogins scopes via auth_sessions.tenant_id directly.
//
// Run:
//
//	METALDOCS_DATABASE_URL=postgres://... go test -tags integration -v \
//	  -run TestSecurityRepository_NoIamUsersJoin_Live \
//	  ./internal/modules/security/infrastructure/postgres/
import (
	"context"
	"database/sql"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	iampg "metaldocs/internal/modules/iam/infrastructure/postgres"
	securitypg "metaldocs/internal/modules/security/infrastructure/postgres"
)

func openLiveSecurityDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		dsn = strings.TrimSpace(os.Getenv("METALDOCS_DATABASE_URL"))
	}
	if dsn == "" {
		t.Skip("no DATABASE_URL or METALDOCS_DATABASE_URL set — skipping live DB probe")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("sql.Open(pgx): %v", err)
	}
	if err := db.PingContext(context.Background()); err != nil {
		t.Fatalf("db.Ping: %v", err)
	}
	return db
}

func TestSecurityRepository_NoIamUsersJoin_Live(t *testing.T) {
	db := openLiveSecurityDB(t)
	defer db.Close()
	ctx := context.Background()

	repo := securitypg.NewRepository(db,
		iampg.NewUserDisplayNameRepository(db),
		iampg.NewTenantUserRepository(db),
		iampg.NewAdminRoleMemberRepository(db),
		iampg.NewMfaUserRepository(db),
	)

	const (
		tenantA = "f4600000-0000-4000-8000-0000000000aa"
		tenantB = "f4600000-0000-4000-8000-0000000000bb"

		userActive = "sec-active-f46"   // tenant A member, locked, recent failures
		userDeact  = "sec-deact-f46"    // tenant A member (deactivated), locked
		userOther  = "sec-other-f46"    // tenant B member, locked
		userGhost  = "sec-ghost-f46"    // NO iam_users row; auth_identity + tenant-A session
	)
	now := time.Now().UTC()

	allUsers := []string{userActive, userDeact, userOther, userGhost}
	cleanup := func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM metaldocs.auth_sessions WHERE user_id = ANY($1)`, asArray(allUsers))
		_, _ = db.ExecContext(ctx, `DELETE FROM metaldocs.auth_identities WHERE user_id = ANY($1)`, asArray(allUsers))
		_, _ = db.ExecContext(ctx, `DELETE FROM metaldocs.iam_users WHERE user_id = ANY($1)`, asArray(allUsers))
		_, _ = db.ExecContext(ctx, `DELETE FROM metaldocs.tenants WHERE id IN ($1::uuid, $2::uuid)`, tenantA, tenantB)
	}
	cleanup()
	t.Cleanup(cleanup)

	// Tenants.
	for _, tid := range []string{tenantA, tenantB} {
		if _, err := db.ExecContext(ctx,
			`INSERT INTO metaldocs.tenants (id, name, slug)
			 VALUES ($1::uuid, 'F4.6 Tenant '||$1, 'f46-'||$1)
			 ON CONFLICT (id) DO NOTHING`, tid); err != nil {
			t.Fatalf("insert tenant %s: %v", tid, err)
		}
	}

	// iam_users membership (NO row for userGhost — proves JOIN removal + fallback).
	type member struct {
		id, name, tenant string
		deactivated      bool
	}
	for _, m := range []member{
		{userActive, "Active Sec", tenantA, false},
		{userDeact, "Deact Sec", tenantA, true},
		{userOther, "Other Sec", tenantB, false},
	} {
		if _, err := db.ExecContext(ctx,
			`INSERT INTO metaldocs.iam_users (user_id, display_name, is_active, tenant_id, created_at, updated_at)
			 VALUES ($1, $2, TRUE, $3::uuid, now(), now())`, m.id, m.name, m.tenant); err != nil {
			t.Fatalf("insert iam_user %s: %v", m.id, err)
		}
		if m.deactivated {
			if _, err := db.ExecContext(ctx,
				`UPDATE metaldocs.iam_users SET deactivated_at = now() WHERE user_id = $1`, m.id); err != nil {
				t.Fatalf("deactivate %s: %v", m.id, err)
			}
		}
	}

	// auth_identities — all four; locked; userActive over the failed-login threshold.
	type identity struct {
		id            string
		failedLogins  int
		lockedUntil   time.Time
		lastFailedAt  sql.NullTime
		lastFailedIP  string
	}
	idents := []identity{
		{userActive, 6, now.Add(time.Hour), sql.NullTime{Time: now.Add(-time.Minute), Valid: true}, "10.0.0.1"},
		{userDeact, 0, now.Add(time.Hour), sql.NullTime{}, ""},
		{userOther, 9, now.Add(time.Hour), sql.NullTime{Time: now.Add(-time.Minute), Valid: true}, "10.0.0.2"},
		{userGhost, 9, now.Add(time.Hour), sql.NullTime{Time: now.Add(-time.Minute), Valid: true}, "10.0.0.3"},
	}
	for _, id := range idents {
		if _, err := db.ExecContext(ctx, `
INSERT INTO metaldocs.auth_identities
  (user_id, username, email, display_name, is_active, password_hash, password_algo,
   must_change_password, failed_login_attempts, locked_until, last_failed_login_at,
   last_failed_login_ip, created_at, updated_at)
VALUES ($1, $1, NULL, 'AI', TRUE, 'hash', 'bcrypt', FALSE, $2, $3, $4, NULLIF($5,''), NOW(), NOW())
`, id.id, id.failedLogins, id.lockedUntil, id.lastFailedAt, id.lastFailedIP); err != nil {
			t.Fatalf("insert auth_identity %s: %v", id.id, err)
		}
	}

	// auth_sessions for new-device: userActive (member) + userGhost (no iam_users) in
	// tenant A; userOther in tenant B (must be excluded by tenant scope).
	type sess struct {
		sid, uid, tenant, ua string
	}
	for i, s := range []sess{
		{"f46-sess-active", userActive, tenantA, "DeviceX"},
		{"f46-sess-ghost", userGhost, tenantA, "GhostUA"},
		{"f46-sess-other", userOther, tenantB, "BDev"},
	} {
		if _, err := db.ExecContext(ctx, `
INSERT INTO metaldocs.auth_sessions
  (session_id, user_id, tenant_id, created_at, expires_at, revoked_at, ip_address, user_agent, last_seen_at)
VALUES ($1, $2, $3::uuid, $4, $5, NULL, '127.0.0.1', $6, $4)
`, s.sid, s.uid, s.tenant, now.Add(-time.Duration(i+1)*time.Minute), now.Add(time.Hour), s.ua); err != nil {
			t.Fatalf("insert auth_session %s: %v", s.sid, err)
		}
	}

	t.Run("ListLockouts_members_only_with_port_names", func(t *testing.T) {
		got, err := repo.ListLockouts(ctx, tenantA)
		if err != nil {
			t.Fatalf("ListLockouts: %v", err)
		}
		names := map[string]string{}
		for _, l := range got {
			names[l.UserID] = l.DisplayName
		}
		if _, leaked := names[userOther]; leaked {
			t.Fatalf("tenant scope broken: tenant-B member leaked into tenant-A lockouts")
		}
		if _, leaked := names[userGhost]; leaked {
			t.Fatalf("membership filter broken: non-member (no iam_users row) leaked into lockouts")
		}
		if names[userActive] != "Active Sec" {
			t.Fatalf("userActive display_name = %q, want %q (port enrichment)", names[userActive], "Active Sec")
		}
		if names[userDeact] != "Deact Sec" {
			t.Fatalf("userDeact display_name = %q, want %q (deactivated member still surfaces, no deactivated_at filter)", names[userDeact], "Deact Sec")
		}
	})

	t.Run("CountRecentFailedLoginsByUser_tenant_scoped", func(t *testing.T) {
		got, err := repo.CountRecentFailedLoginsByUser(ctx, tenantA, 15*60, 5)
		if err != nil {
			t.Fatalf("CountRecentFailedLoginsByUser: %v", err)
		}
		if _, leaked := got[userOther]; leaked {
			t.Fatalf("tenant scope broken: tenant-B member in failed-login map")
		}
		if _, leaked := got[userGhost]; leaked {
			t.Fatalf("membership filter broken: non-member in failed-login map")
		}
		s, ok := got[userActive]
		if !ok {
			t.Fatalf("userActive (6 failures, recent) missing from failed-login map: %v", got)
		}
		if s.DisplayName != "Active Sec" {
			t.Fatalf("failed-login display_name = %q, want %q", s.DisplayName, "Active Sec")
		}
		if _, ok := got[userDeact]; ok {
			t.Fatalf("userDeact (0 failures) should not meet threshold")
		}
	})

	t.Run("CountRecentLockouts_tenant_scoped_count", func(t *testing.T) {
		n, err := repo.CountRecentLockouts(ctx, tenantA, 60*60)
		if err != nil {
			t.Fatalf("CountRecentLockouts: %v", err)
		}
		if n != 2 {
			t.Fatalf("CountRecentLockouts(tenantA) = %d, want 2 (userActive+userDeact; tenant-B + non-member excluded)", n)
		}
	})

	t.Run("ListNewDeviceLogins_scoped_via_session_tenant_with_fallback", func(t *testing.T) {
		got, err := repo.ListNewDeviceLogins(ctx, tenantA, 24*60*60, 90*24*60*60)
		if err != nil {
			t.Fatalf("ListNewDeviceLogins: %v", err)
		}
		names := map[string]string{}
		var sids []string
		for _, d := range got {
			names[d.UserID] = d.DisplayName
			sids = append(sids, d.SessionID)
		}
		sort.Strings(sids)
		for _, s := range sids {
			if s == "f46-sess-other" {
				t.Fatalf("tenant scope broken: tenant-B session leaked into tenant-A new-device list")
			}
		}
		if names[userActive] != "Active Sec" {
			t.Fatalf("new-device member display_name = %q, want %q", names[userActive], "Active Sec")
		}
		// userGhost has no iam_users row → DisplayNames omits → fallback to user_id.
		if names[userGhost] != userGhost {
			t.Fatalf("new-device fallback display_name = %q, want %q (missing→user_id; proves JOIN gone)", names[userGhost], userGhost)
		}
	})
}

// asArray adapts a []string to a Postgres text[] parameter without importing pq
// in the test (the repo owns the pq dependency); a comma-joined literal is fine
// for the small fixed cleanup sets here.
func asArray(ss []string) interface{} {
	return "{" + strings.Join(ss, ",") + "}"
}
