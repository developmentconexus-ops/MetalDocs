//go:build integration

package postgres_test

import (
	"context"
	"database/sql"
	"sort"
	"testing"

	iampg "metaldocs/internal/modules/iam/infrastructure/postgres"
	"metaldocs/tests/integration/testdb"
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
		// iam_user_roles and tenants carry trg_require_cap_asserted (user.manage /
		// tenant.onboard) — assert both tx-locally so this best-effort cleanup
		// actually deletes rows.
		if tx, err := db.BeginTx(ctx, nil); err == nil {
			testdb.SetCapsOnTx(t, tx, `[{"cap":"user.manage"},{"cap":"tenant.onboard"}]`)
			_, _ = tx.ExecContext(ctx, `DELETE FROM metaldocs.iam_user_roles WHERE user_id IN ($1, $2)`,
				userMulti, userOther)
			_, _ = tx.ExecContext(ctx, `DELETE FROM metaldocs.iam_users WHERE user_id IN ($1, $2)`,
				userMulti, userOther)
			_, _ = tx.ExecContext(ctx, `DELETE FROM metaldocs.tenants WHERE id IN ($1::uuid, $2::uuid, $3::uuid)`,
				tenantA, tenantB, otherTenant)
			_ = tx.Commit()
		}
	}
	cleanup()
	t.Cleanup(cleanup)

	// iam_user_roles.tenant_id FKs to metaldocs.tenants(id) — seed the three
	// tenants before inserting roles.
	for _, tid := range []string{tenantA, tenantB, otherTenant} {
		tid := tid
		seedWithCapsIAM(t, db, `[{"cap":"tenant.onboard"}]`, func(tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx,
				`INSERT INTO metaldocs.tenants (id, name, slug)
				 VALUES ($1::uuid, 'F5.2 Tenant '||$1, 'f52-'||$1)
				 ON CONFLICT (id) DO NOTHING`,
				tid,
			)
			return err
		})
	}

	// iam_user_roles.user_id FKs to metaldocs.iam_users(user_id) (PK is user_id
	// alone — one home-tenant row per user) — seed a home iam_users row for each
	// user before inserting roles; iam_user_roles memberships in other tenants
	// are independent of the user's home tenant.
	seedWithCapsIAM(t, db, `[{"cap":"user.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO metaldocs.iam_users (user_id, display_name, is_active, tenant_id, created_at, updated_at)
			 VALUES ($1, 'Multi F52', TRUE, $2::uuid, now(), now()),
			        ($3, 'Other F52', TRUE, $4::uuid, now(), now())`,
			userMulti, tenantA, userOther, otherTenant,
		)
		return err
	})

	// userMulti: one role in tenantA + one role in tenantB (schema now enforces
	// uq_iam_user_roles_user_tenant UNIQUE (tenant_id, user_id) — one role per
	// user per tenant — so the "DISTINCT tenants for a user" proof seeds two
	// distinct tenant memberships instead of two roles in the same tenant).
	// SEC-05 / migration 0259 (and pre-existing since migration 0188):
	// iam_user_roles carries trg_require_cap_asserted (user.manage) — assert the
	// cap tx-locally via seedWithCapsIAM.
	seedWithCapsIAM(t, db, `[{"cap":"user.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO metaldocs.iam_user_roles (user_id, role_code, tenant_id)
			 VALUES ($1, 'author', $2::uuid),
			        ($1, 'viewer', $3::uuid)`,
			userMulti, tenantA, tenantB,
		)
		return err
	})
	// userOther: a role in a third tenant — must not leak into userMulti's result.
	seedWithCapsIAM(t, db, `[{"cap":"user.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO metaldocs.iam_user_roles (user_id, role_code, tenant_id)
			 VALUES ($1, 'viewer', $2::uuid)`,
			userOther, otherTenant,
		)
		return err
	})

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
