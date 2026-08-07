//go:build integration

// iam_users_tripwire_test.go — SEC-05 (P1) residual gap closed by migration
// 0259: metaldocs.iam_users now carries trg_require_cap_asserted (user.manage),
// unconditional on INSERT/DELETE and column-scoped on UPDATE (display_name,
// is_active, tenant_id, deactivated_at only — excluding the login-context and
// presence-heartbeat telemetry columns, which must remain unguarded).
//
// Three properties proven here, mirroring membership_area_scope_test.go's
// established pattern for the sibling user_process_areas tripwire:
//
//   - TestIAMUsersTripwire_Tier2DeniedWithoutGrant: the Go tier-2 path
//     (RoleAdminRepository.UpsertUserAndAssignRoleTx, which already asserts
//     user.manage per T-004 / Plan 5) rejects an actor holding no role at all,
//     via authz.ErrCapDenied, before any row is written — proving tier-2 is
//     reachable and fails closed for this table's application entry point.
//
//   - TestIAMUsersTripwire_Tier2PassesWithGrant: the same Go call succeeds once
//     the caller holds system_admin (tier-2 inheritance bypass, ADR 0022 R1),
//     and the iam_users/iam_user_roles rows are actually written.
//
//   - TestIAMUsersTripwire_DBRejectsDirectWriteWithoutAssertion: a raw SQL
//     INSERT against metaldocs.iam_users with no metaldocs.asserted_caps GUC
//     set (bypassing the Go authz layer entirely, e.g. a rogue migration or a
//     future repository method that forgets authz.Require) is rejected by the
//     DB trigger itself — the actual SEC-05 defense-in-depth property. Also
//     proves the column-scoped UPDATE exception: a telemetry-only UPDATE
//     (last_seen_at, mirroring presence.BumpLastSeen) is NOT rejected, while a
//     privileged-column UPDATE (display_name) IS.
//
//     go test -tags=integration ./tests/integration/iam/... -run IAMUsersTripwire
package iam_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"
	pgrepo "metaldocs/internal/modules/iam/infrastructure/postgres"
	"metaldocs/tests/integration/testdb"
)

// TestIAMUsersTripwire_Tier2DeniedWithoutGrant proves the Go tier-2 authz.Require
// call already present in RoleAdminRepository.UpsertUserAndAssignRoleTx (T-004,
// closed prior to SEC-05) rejects an actor with no role/capability grant, and
// that no iam_users/iam_user_roles row is written on denial (BOLA escalation
// guard, same shape as TestMembershipAreaScope_AreaAdmin_OutsideManagedArea).
func TestIAMUsersTripwire_Tier2DeniedWithoutGrant(t *testing.T) {
	db := openDB(t)
	ctx := context.Background()
	repo := pgrepo.NewRoleAdminRepository(db)

	noCapsActor := testdb.DeterministicID(t, "iamut-nocaps-actor")
	target := testdb.DeterministicID(t, "iamut-nocaps-target")
	seedIdentity(t, db, noCapsActor) // identity only — zero roles, zero capabilities
	t.Cleanup(func() { cleanupIAMUser(db, target) })

	err := repo.UpsertUserAndAssignRole(ctx, target, "IAM UT Target", devTenant, iamdomain.RoleAuthor, noCapsActor)
	if !isCapDenied(err) {
		t.Fatalf("UpsertUserAndAssignRole(no-caps actor) = %v, want authz.ErrCapDenied", err)
	}

	var count int
	if err := db.QueryRowContext(ctx,
		`SELECT count(*) FROM metaldocs.iam_users WHERE user_id = $1`, target,
	).Scan(&count); err != nil {
		t.Fatalf("count iam_users after denial: %v", err)
	}
	if count != 0 {
		t.Fatalf("denied UpsertUserAndAssignRole still wrote %d iam_users rows, want 0 (BOLA escalation guard)", count)
	}
}

// TestIAMUsersTripwire_Tier2PassesWithGrant proves the same Go call succeeds and
// actually persists once the caller holds system_admin (ADR 0022 R1 tier-2
// inheritance bypass) — the positive counterpart to the denial test above.
func TestIAMUsersTripwire_Tier2PassesWithGrant(t *testing.T) {
	db := openDB(t)
	ctx := context.Background()
	repo := pgrepo.NewRoleAdminRepository(db)

	sysAdmin := testdb.DeterministicID(t, "iamut-sysadmin-actor")
	target := testdb.DeterministicID(t, "iamut-sysadmin-target")
	seedIdentity(t, db, sysAdmin)
	seedSystemAdminRole(t, db, sysAdmin)
	t.Cleanup(func() { cleanupIAMUser(db, target) })

	if err := repo.UpsertUserAndAssignRole(ctx, target, "IAM UT Granted Target", devTenant, iamdomain.RoleAuthor, sysAdmin); err != nil {
		t.Fatalf("UpsertUserAndAssignRole(system_admin actor) = %v, want nil", err)
	}

	var displayName, roleCode string
	if err := db.QueryRowContext(ctx,
		`SELECT display_name FROM metaldocs.iam_users WHERE user_id = $1`, target,
	).Scan(&displayName); err != nil {
		t.Fatalf("read iam_users after grant: %v", err)
	}
	if displayName != "IAM UT Granted Target" {
		t.Fatalf("iam_users.display_name = %q, want %q", displayName, "IAM UT Granted Target")
	}
	if err := db.QueryRowContext(ctx,
		`SELECT role_code FROM metaldocs.iam_user_roles WHERE user_id = $1 AND tenant_id = $2::uuid`, target, devTenant,
	).Scan(&roleCode); err != nil {
		t.Fatalf("read iam_user_roles after grant: %v", err)
	}
	if roleCode != string(iamdomain.RoleAuthor) {
		t.Fatalf("iam_user_roles.role_code = %q, want %q", roleCode, iamdomain.RoleAuthor)
	}
}

// TestIAMUsersTripwire_DBRejectsDirectWriteWithoutAssertion is the actual SEC-05
// defense-in-depth proof: a raw write against metaldocs.iam_users with NO
// metaldocs.asserted_caps GUC set (i.e. bypassing authz.Require entirely, the
// scenario a future repository method or a rogue migration could hit) is
// rejected by trg_require_cap_asserted / enforce_capability_asserted() at the
// DB layer, independent of whether the Go call site remembered to assert
// user.manage. Also proves the column-scoped UPDATE exception carved out for
// the login-context/presence-heartbeat hazard (migration 0259 header comment):
// a telemetry-only UPDATE (last_seen_at) is NOT gated, while a privileged
// column UPDATE (display_name) IS.
func TestIAMUsersTripwire_DBRejectsDirectWriteWithoutAssertion(t *testing.T) {
	db := openDB(t)
	ctx := context.Background()

	userID := testdb.DeterministicID(t, "iamut-direct-write")
	t.Cleanup(func() { cleanupIAMUser(db, userID) })

	t.Run("insert_without_asserted_caps_rejected", func(t *testing.T) {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}
		defer tx.Rollback() //nolint:errcheck

		_, err = tx.ExecContext(ctx,
			`INSERT INTO metaldocs.iam_users (user_id, display_name, tenant_id)
			 VALUES ($1, 'Direct Write Probe', $2::uuid)`,
			userID, devTenant,
		)
		if !isCapabilityNotAsserted(err) {
			t.Fatalf("raw INSERT into iam_users without asserted_caps = %v, want ErrCapabilityNotAsserted (P0001)", err)
		}
	})

	t.Run("privileged_column_update_without_asserted_caps_rejected", func(t *testing.T) {
		// Seed the row under a legitimate grant first (via the Go tier-2 path).
		repo := pgrepo.NewRoleAdminRepository(db)
		actor := testdb.DeterministicID(t, "iamut-direct-write-seeder")
		seedIdentity(t, db, actor)
		seedSystemAdminRole(t, db, actor)
		if err := repo.UpsertUserAndAssignRole(ctx, userID, "Seeded For Probe", devTenant, iamdomain.RoleAuthor, actor); err != nil {
			t.Fatalf("seed target via tier-2 path: %v", err)
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}
		defer tx.Rollback() //nolint:errcheck

		_, err = tx.ExecContext(ctx,
			`UPDATE metaldocs.iam_users SET display_name = 'Tampered' WHERE user_id = $1`, userID,
		)
		if !isCapabilityNotAsserted(err) {
			t.Fatalf("raw UPDATE of iam_users.display_name without asserted_caps = %v, want ErrCapabilityNotAsserted (P0001)", err)
		}
	})

	t.Run("telemetry_only_update_not_gated", func(t *testing.T) {
		// last_seen_at mirrors internal/modules/iam/presence.PostgresRepository.
		// BumpLastSeen — this column is deliberately excluded from the trigger's
		// UPDATE OF column list so the presence heartbeat keeps working unguarded.
		// The row must already exist (seeded by the sibling subtest above via the
		// tier-2 path) for this UPDATE to have a target; if it does not (subtests
		// within a single top-level Test can run in any declared order but here
		// run sequentially in source order, non-parallel), seed it directly under
		// bypass so this subtest is self-sufficient.
		withBypass(t, db, func(tx *sql.Tx) {
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO metaldocs.iam_users (user_id, display_name, tenant_id)
				 VALUES ($1, 'Telemetry Probe', $2::uuid)
				 ON CONFLICT (user_id) DO NOTHING`,
				userID, devTenant,
			); err != nil {
				t.Fatalf("seed telemetry probe row: %v", err)
			}
		})

		if _, err := db.ExecContext(ctx,
			`UPDATE metaldocs.iam_users SET last_seen_at = now() WHERE user_id = $1`, userID,
		); err != nil {
			t.Fatalf("telemetry-only UPDATE (last_seen_at) rejected = %v, want nil (column-scoped trigger must not fire)", err)
		}
	})
}

// isCapabilityNotAsserted reports whether err is the DB tripwire's fail-closed
// rejection (RAISE EXCEPTION ... USING ERRCODE = 'P0001', message prefixed
// ErrCapabilityNotAsserted — see enforce_capability_asserted() in
// db/migrations/0259_iam_documents_tripwire.sql and its predecessors).
func isCapabilityNotAsserted(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "ErrCapabilityNotAsserted") ||
		strings.Contains(msg, "P0001")
}

// cleanupIAMUser removes a probe row via the scheduler bypass GUC (the tripwire
// now gates DELETE on iam_users same as iam_user_roles/user_process_areas).
// Error-discarding by design, matching the established cleanup convention in
// this package (e.g. seedSystemAdminRole's t.Cleanup in membership_area_scope_test.go).
func cleanupIAMUser(db *sql.DB, userID string) {
	_ = withBypassErr(db, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(), `DELETE FROM metaldocs.iam_user_roles WHERE user_id = $1`, userID)
		return err
	})
	_ = withBypassErr(db, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(), `DELETE FROM metaldocs.iam_users WHERE user_id = $1`, userID)
		return err
	})
}
