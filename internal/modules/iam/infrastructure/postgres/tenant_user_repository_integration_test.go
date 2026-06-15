//go:build integration

package postgres_test

import (
	"context"
	"sort"
	"testing"

	iampg "metaldocs/internal/modules/iam/infrastructure/postgres"
)

// TestTenantUserRepository_TenantUserIDs_Live exercises TenantUserIDs against a
// live Postgres instance. Proves: all members of a tenant are returned including
// a deactivated_at IS NOT NULL member (no active-only filter, matching the INNER
// JOIN it replaces); members of another tenant are excluded (tenant scope); an
// unknown tenant returns an empty slice.
func TestTenantUserRepository_TenantUserIDs_Live(t *testing.T) {
	db := openLiveIAMDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := iampg.NewTenantUserRepository(db)

	const tenantID = "f4500000-0000-4000-8000-0000000000aa"
	const otherTenantID = "f4500000-0000-4000-8000-0000000000bb"
	const unknownTenantID = "f4500000-0000-4000-8000-0000000000cc"
	const userActive = "tur-active-f45"
	const userDeactivated = "tur-deactivated-f45"
	const userOtherTenant = "tur-other-f45"

	cleanup := func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM metaldocs.iam_users WHERE user_id IN ($1, $2, $3)`,
			userActive, userDeactivated, userOtherTenant)
		_, _ = db.ExecContext(ctx, `DELETE FROM metaldocs.tenants WHERE id IN ($1::uuid, $2::uuid)`,
			tenantID, otherTenantID)
	}
	cleanup()
	t.Cleanup(cleanup)

	for _, tid := range []string{tenantID, otherTenantID} {
		if _, err := db.ExecContext(ctx,
			`INSERT INTO metaldocs.tenants (id, name, slug)
			 VALUES ($1::uuid, 'F4.5 Tenant '||$1, 'f45-'||$1)
			 ON CONFLICT (id) DO NOTHING`,
			tid,
		); err != nil {
			t.Fatalf("insert tenant %s: %v", tid, err)
		}
	}

	// Two members of tenantID — one deactivated; one member of otherTenantID.
	if _, err := db.ExecContext(ctx,
		`INSERT INTO metaldocs.iam_users (user_id, display_name, is_active, tenant_id, created_at, updated_at)
		 VALUES ($1, 'Active F45', TRUE, $3::uuid, now(), now()),
		        ($2, 'Deactivated F45', TRUE, $3::uuid, now(), now())`,
		userActive, userDeactivated, tenantID,
	); err != nil {
		t.Fatalf("insert tenant members: %v", err)
	}
	if _, err := db.ExecContext(ctx,
		`UPDATE metaldocs.iam_users SET deactivated_at = now() WHERE user_id = $1`, userDeactivated,
	); err != nil {
		t.Fatalf("deactivate member: %v", err)
	}
	if _, err := db.ExecContext(ctx,
		`INSERT INTO metaldocs.iam_users (user_id, display_name, is_active, tenant_id, created_at, updated_at)
		 VALUES ($1, 'Other Tenant F45', TRUE, $2::uuid, now(), now())`,
		userOtherTenant, otherTenantID,
	); err != nil {
		t.Fatalf("insert other-tenant member: %v", err)
	}

	t.Run("returns_all_members_incl_deactivated_excludes_other_tenant", func(t *testing.T) {
		ids, err := repo.TenantUserIDs(ctx, tenantID)
		if err != nil {
			t.Fatalf("TenantUserIDs: %v", err)
		}
		got := append([]string(nil), ids...)
		sort.Strings(got)
		want := []string{userActive, userDeactivated}
		sort.Strings(want)
		if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
			t.Fatalf("TenantUserIDs = %v, want exactly %v (both members incl. deactivated, no other-tenant leak)", ids, want)
		}
	})

	t.Run("unknown_tenant_returns_empty", func(t *testing.T) {
		ids, err := repo.TenantUserIDs(ctx, unknownTenantID)
		if err != nil {
			t.Fatalf("TenantUserIDs(unknown): %v", err)
		}
		if ids == nil {
			t.Fatalf("TenantUserIDs(unknown) = nil, want empty non-nil slice")
		}
		if len(ids) != 0 {
			t.Fatalf("TenantUserIDs(unknown) = %v, want empty", ids)
		}
	})
}
