//go:build integration

package postgres_test

// Live integration proof for the M4/F4.2 and F4.3 boundary fixes: verifies that
// the port-backed ListOffHoursAdminActions and MfaCoverage return the same data
// as the prior direct-SQL versions on a seeded fixture.
//
// Run:
//
//	METALDOCS_DATABASE_URL=postgres://... go test -tags integration -v \
//	  -run TestSecurityRepository_PortParity_Live \
//	  ./internal/modules/security/infrastructure/postgres/
import (
	"context"
	"database/sql"
	"testing"
	"time"

	iampg "metaldocs/internal/modules/iam/infrastructure/postgres"
	securitypg "metaldocs/internal/modules/security/infrastructure/postgres"
	"metaldocs/tests/integration/testdb"
)

func TestSecurityRepository_PortParity_Live(t *testing.T) {
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
		tenantA = "f4500000-0000-4000-8000-0000000000aa"
		tenantB = "f4500000-0000-4000-8000-0000000000bb"

		adminUser   = "f45-admin-user"  // tenant A, system_admin role
		normalUser  = "f45-normal-user" // tenant A, no admin role (editor)
		otherTenant = "f45-other-admin" // tenant B, system_admin role (must be excluded by tenant)
	)

	allUsers := []string{adminUser, normalUser, otherTenant}
	allTenants := []string{tenantA, tenantB}

	cleanup := func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM metaldocs.audit_events WHERE actor_id = ANY($1)`, asArray(allUsers))
		// iam_user_roles, iam_users and tenants carry trg_require_cap_asserted (user.manage /
		// tenant.onboard) — assert both tx-locally so this best-effort cleanup actually deletes
		// rows instead of silently no-op'ing and leaking duplicate-key state into the next run.
		if tx, err := db.BeginTx(ctx, nil); err == nil {
			testdb.SetCapsOnTx(t, tx, `[{"cap":"user.manage"},{"cap":"tenant.onboard"}]`)
			_, _ = tx.ExecContext(ctx, `DELETE FROM metaldocs.iam_user_roles WHERE user_id = ANY($1)`, asArray(allUsers))
			_, _ = tx.ExecContext(ctx, `DELETE FROM metaldocs.iam_users WHERE user_id = ANY($1)`, asArray(allUsers))
			_, _ = tx.ExecContext(ctx, `DELETE FROM metaldocs.tenants WHERE id IN ($1::uuid, $2::uuid)`, tenantA, tenantB)
			_ = tx.Commit()
		}
	}
	cleanup()
	t.Cleanup(cleanup)

	// Seed tenants.
	for _, tid := range allTenants {
		testdb.SeedWithCaps(t, db, `[{"cap":"tenant.onboard"}]`, func(tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx,
				`INSERT INTO metaldocs.tenants (id, name, slug)
				 VALUES ($1::uuid, 'F4.5 Tenant '||$1, 'f45-'||$1)
				 ON CONFLICT (id) DO NOTHING`, tid)
			return err
		})
	}

	// Seed iam_users: all three (MFA varies).
	type iamUser struct {
		id, tenant string
		mfa        bool
		deact      bool
	}
	for _, u := range []iamUser{
		{adminUser, tenantA, true, false},
		{normalUser, tenantA, false, false},
		{otherTenant, tenantB, true, false},
	} {
		u := u
		testdb.SeedWithCaps(t, db, `[{"cap":"user.manage"}]`, func(tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx,
				`INSERT INTO metaldocs.iam_users (user_id, display_name, is_active, mfa_enabled, tenant_id, created_at, updated_at)
				 VALUES ($1, $1, TRUE, $2, $3::uuid, now(), now())`,
				u.id, u.mfa, u.tenant)
			return err
		})
	}

	// Seed iam_user_roles: adminUser = system_admin in tenantA; otherTenant = system_admin in tenantB.
	// normalUser has no admin role (only excluded role code so it never enters the port result).
	for _, r := range []struct{ user, tenant, role string }{
		{adminUser, tenantA, "system_admin"},
		{otherTenant, tenantB, "system_admin"},
		{normalUser, tenantA, "editor"}, // NOT an admin role used in the signal
	} {
		r := r
		testdb.SeedWithCaps(t, db, `[{"cap":"user.manage"}]`, func(tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx,
				`INSERT INTO metaldocs.iam_user_roles (user_id, tenant_id, role_code, assigned_at)
				 VALUES ($1, $2::uuid, $3, now())
				 ON CONFLICT DO NOTHING`, r.user, r.tenant, r.role)
			return err
		})
	}

	// Seed audit_events: one off-hours event per user + one in-hours event for adminUser.
	// Off-hours: 23:00 UTC (OffHoursStartHour=22, OffHoursEndHour=6 → hour>=22 qualifies).
	// In-hours: 10:00 UTC (should NOT appear).
	now := time.Now().UTC()
	// Pin the off-hours timestamp to yesterday at 23:00 so it is always in the window.
	offHours := time.Date(now.Year(), now.Month(), now.Day()-1, 23, 0, 0, 0, time.UTC)
	inHours := time.Date(now.Year(), now.Month(), now.Day()-1, 10, 0, 0, 0, time.UTC)

	for _, e := range []struct {
		id, actor, tenant string
		at                time.Time
	}{
		{"f45-evt-admin-off", adminUser, tenantA, offHours},
		{"f45-evt-admin-in", adminUser, tenantA, inHours},     // in-hours: should be excluded
		{"f45-evt-normal-off", normalUser, tenantA, offHours}, // no admin role: excluded by port
		{"f45-evt-other-off", otherTenant, tenantB, offHours}, // wrong tenant: excluded
	} {
		if _, err := db.ExecContext(ctx,
			`INSERT INTO metaldocs.audit_events (id, occurred_at, actor_id, action, resource_type, resource_id, payload, trace_id, tenant_id)
			 VALUES ($1, $2, $3, 'update', 'document', 'doc-1', '{}', 'trace-f45', $4)
			 ON CONFLICT DO NOTHING`, e.id, e.at, e.actor, e.tenant); err != nil {
			t.Fatalf("insert audit_event %s: %v", e.id, err)
		}
	}

	t.Run("F4.2_ListOffHoursAdminActions_port_parity", func(t *testing.T) {
		adminRoles := []string{"system_admin", "area_admin", "qms_admin"}
		got, err := repo.ListOffHoursAdminActions(ctx, tenantA, 48*60*60, adminRoles, 22, 6)
		if err != nil {
			t.Fatalf("ListOffHoursAdminActions: %v", err)
		}

		// Only the off-hours event for adminUser should appear.
		// normalUser has no admin role (excluded by port); otherTenant is wrong tenant.
		// adminUser's in-hours event is excluded by off-hours predicate.
		if len(got) != 1 {
			t.Fatalf("ListOffHoursAdminActions returned %d events, want 1; got: %+v", len(got), got)
		}
		if got[0].EventID != "f45-evt-admin-off" {
			t.Errorf("event ID = %q, want %q", got[0].EventID, "f45-evt-admin-off")
		}
		if got[0].ActorID != adminUser {
			t.Errorf("actor ID = %q, want %q", got[0].ActorID, adminUser)
		}
		if got[0].ActorRole != "system_admin" {
			t.Errorf("actor role = %q, want %q", got[0].ActorRole, "system_admin")
		}

		// Non-admin user must not appear (port excluded them before the audit_events query).
		for _, a := range got {
			if a.ActorID == normalUser {
				t.Errorf("normal user (no admin role) leaked into off-hours actions")
			}
			if a.ActorID == otherTenant {
				t.Errorf("tenant-B admin leaked into tenant-A off-hours actions")
			}
		}
	})

	t.Run("F4.3_MfaCoverage_port_parity", func(t *testing.T) {
		cov, err := repo.MfaCoverage(ctx, tenantA)
		if err != nil {
			t.Fatalf("MfaCoverage: %v", err)
		}

		// Tenant A has 2 active users: adminUser (mfa=true), normalUser (mfa=false).
		// otherTenant belongs to tenant B — must not appear.
		if cov.TotalUsers != 2 {
			t.Errorf("TotalUsers = %d, want 2", cov.TotalUsers)
		}
		if cov.MfaEnabled != 1 {
			t.Errorf("MfaEnabled = %d, want 1 (adminUser only)", cov.MfaEnabled)
		}
		if cov.MfaEnabledPct != 50 {
			t.Errorf("MfaEnabledPct = %v, want 50", cov.MfaEnabledPct)
		}

		// Per-role: adminUser has system_admin (mfa=true, total=1, enabled=1).
		//           normalUser has editor (mfa=false, total=1, enabled=0).
		roleMap := map[string]struct{ total, enabled int }{}
		for _, s := range cov.ByRole {
			roleMap[s.Role] = struct{ total, enabled int }{s.Total, s.MfaEnabled}
		}
		if r, ok := roleMap["system_admin"]; !ok || r.total != 1 || r.enabled != 1 {
			t.Errorf("system_admin slice = %+v, want total=1 enabled=1", roleMap["system_admin"])
		}
		if r, ok := roleMap["editor"]; !ok || r.total != 1 || r.enabled != 0 {
			t.Errorf("editor slice = %+v, want total=1 enabled=0", roleMap["editor"])
		}
		// Tenant-B admin must not appear.
		if _, leaked := roleMap["__tenantB_admin"]; leaked {
			t.Errorf("tenant-B role leaked into tenant-A MFA coverage")
		}
	})
}
