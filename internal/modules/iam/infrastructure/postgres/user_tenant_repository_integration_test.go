//go:build integration

package postgres_test

import (
	"context"
	"sort"
	"testing"

	iampg "metaldocs/internal/modules/iam/infrastructure/postgres"
)

// TestUserTenantRepository_UserTenantIDs_Live exercises UserTenantIDs against a live
// Postgres instance (H-G site #1 parity, M5/F5.2). Proves: a user holding multiple
// roles across two tenants yields each tenant exactly once, sorted (DISTINCT +
// ORDER BY); another user's tenant is excluded; an unknown user returns an empty
// non-nil slice. This is the behavioral parity proof for the auth GetUserTenants
// move — labeled live, not fixture.
func TestUserTenantRepository_UserTenantIDs_Live(t *testing.T) {
	db := openLiveIAMDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := iampg.NewUserTenantRepository(db)

	const tenantA = "f5200000-0000-4000-8000-0000000000aa"
	const tenantB = "f5200000-0000-4000-8000-0000000000bb"
	const otherTenant = "f5200000-0000-4000-8000-0000000000cc"
	const userMulti = "utr-multi-f52"
	const userOther = "utr-other-f52"
	const userUnknown = "utr-unknown-f52"

	cleanup := func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM metaldocs.iam_user_roles WHERE user_id IN ($1, $2)`,
			userMulti, userOther)
	}
	cleanup()
	t.Cleanup(cleanup)

	// userMulti: two roles in tenantA (DISTINCT must collapse) + one in tenantB.
	if _, err := db.ExecContext(ctx,
		`INSERT INTO metaldocs.iam_user_roles (user_id, role_code, tenant_id)
		 VALUES ($1, 'author', $2::uuid),
		        ($1, 'editor', $2::uuid),
		        ($1, 'viewer', $3::uuid)`,
		userMulti, tenantA, tenantB,
	); err != nil {
		t.Fatalf("insert userMulti roles: %v", err)
	}
	// userOther: a role in a third tenant — must not leak into userMulti's result.
	if _, err := db.ExecContext(ctx,
		`INSERT INTO metaldocs.iam_user_roles (user_id, role_code, tenant_id)
		 VALUES ($1, 'viewer', $2::uuid)`,
		userOther, otherTenant,
	); err != nil {
		t.Fatalf("insert userOther role: %v", err)
	}

	t.Run("distinct_sorted_tenants_for_user_excludes_others", func(t *testing.T) {
		ids, err := repo.UserTenantIDs(ctx, userMulti)
		if err != nil {
			t.Fatalf("UserTenantIDs: %v", err)
		}
		want := []string{tenantA, tenantB}
		sort.Strings(want)
		if len(ids) != 2 || ids[0] != want[0] || ids[1] != want[1] {
			t.Fatalf("UserTenantIDs = %v, want exactly %v (distinct, sorted, no other-user leak)", ids, want)
		}
	})

	t.Run("unknown_user_returns_empty", func(t *testing.T) {
		ids, err := repo.UserTenantIDs(ctx, userUnknown)
		if err != nil {
			t.Fatalf("UserTenantIDs(unknown): %v", err)
		}
		if ids == nil {
			t.Fatalf("UserTenantIDs(unknown) = nil, want empty non-nil slice")
		}
		if len(ids) != 0 {
			t.Fatalf("UserTenantIDs(unknown) = %v, want empty", ids)
		}
	})
}
