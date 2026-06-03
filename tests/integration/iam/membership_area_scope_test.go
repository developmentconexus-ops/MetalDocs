//go:build integration

// membership_area_scope_test.go — ADR 0022 Phase 3 (membership area-scoping).
//
// Exercises the REAL tier-2 authz path: AreaMembershipService over the postgres
// UserAreaRepository, which now passes the membership's real areaCode (not the
// literal "tenant") to authz.Require. Asserts the behavioral contract that the
// in-memory unit tests cannot reach (they have no authz layer):
//
//   - area_admin CAN grant/revoke within a managed area
//   - area_admin CANNOT grant/revoke outside a managed area  → ErrCapDenied (403)
//   - R1 (ADR amendment): system_admin with NO per-area membership row is NOT
//     blocked by the missing sub-scope grant — the tier-2 bypass short-circuits
//     before the area-filtered capability query.
//
//   go test -tags=integration ./tests/integration/iam/... -run Membership
package iam_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	iamapp "metaldocs/internal/modules/iam/application"
	iamauthz "metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	pgrepo "metaldocs/internal/modules/iam/infrastructure/postgres"
	"metaldocs/tests/integration/testdb"
)

const (
	// Real seeded process areas in the dev tenant (user_process_areas has an FK
	// to document_process_areas(tenant_id, code)). The test area_admin holds
	// membership.manage in areaManaged only.
	areaManaged   = "qualidade"
	areaUnmanaged = "rh"
)

func newAreaMembershipService(db *sql.DB) *iamapp.AreaMembershipService {
	return iamapp.NewAreaMembershipService(pgrepo.NewUserAreaRepository(db), nil)
}

// seedIdentity inserts a minimal iam_users row (no tripwire on this table).
func seedIdentity(t *testing.T, db *sql.DB, userID string) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(),
		`INSERT INTO metaldocs.iam_users (user_id, display_name, tenant_id)
		 VALUES ($1, $1, $2::uuid)
		 ON CONFLICT (user_id) DO UPDATE SET tenant_id = EXCLUDED.tenant_id`,
		userID, devTenant,
	); err != nil {
		t.Fatalf("seed iam_users %s: %v", userID, err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), `DELETE FROM metaldocs.iam_users WHERE user_id = $1`, userID) //nolint:errcheck
	})
}

// seedAreaAdminMembership inserts an active area_admin row in the given area.
// user_process_areas carries the tripwire trigger trg_require_cap_asserted, so a
// plain INSERT is rejected; seeding uses the scheduler bypass GUC (the same hatch
// the async scheduler uses) within a transaction. Cleanup closes the row by
// setting effective_to (direct DELETE is blocked by trg_user_process_areas_no_delete).
func seedAreaAdminMembership(t *testing.T, db *sql.DB, userID, areaCode string) {
	t.Helper()
	ctx := context.Background()
	withBypass(t, db, func(tx *sql.Tx) {
		// Close any straggler active row first so reruns can't trip the
		// ux_user_process_areas_one_active partial unique index.
		// revoked_by must reference a same-tenant user (FK); the row's own seeded
		// user_id is always valid.
		if _, err := tx.ExecContext(ctx,
			`UPDATE public.user_process_areas
			    SET effective_to = now(), revoked_by = $1
			  WHERE user_id = $1 AND tenant_id = $2::uuid AND area_code = $3 AND effective_to IS NULL`,
			userID, devTenant, areaCode,
		); err != nil {
			t.Fatalf("seed close straggler %s/%s: %v", userID, areaCode, err)
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO public.user_process_areas
			   (user_id, tenant_id, area_code, role, effective_from, effective_to, granted_by, revoked_by)
			 VALUES ($1, $2::uuid, $3, 'area_admin', now() - interval '1 hour', NULL, $1, NULL)`,
			userID, devTenant, areaCode,
		); err != nil {
			t.Fatalf("seed area_admin %s/%s: %v", userID, areaCode, err)
		}
	})
	t.Cleanup(func() { closeAllActive(db, userID) })
}

// withBypass runs fn inside a transaction with the scheduler authz bypass set,
// committing on success. Used only for test seeding of tripwire-guarded tables.
func withBypass(t *testing.T, db *sql.DB, fn func(tx *sql.Tx)) {
	t.Helper()
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin bypass tx: %v", err)
	}
	defer tx.Rollback() //nolint:errcheck
	if _, err := tx.ExecContext(ctx, `SELECT set_config('metaldocs.bypass_authz', 'scheduler', true)`); err != nil {
		t.Fatalf("set bypass: %v", err)
	}
	fn(tx)
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit bypass tx: %v", err)
	}
}

// closeAllActive revokes every active membership row for a user (test teardown).
func closeAllActive(db *sql.DB, userID string) {
	_ = withBypassErr(db, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(),
			`UPDATE public.user_process_areas
			    SET effective_to = now(), revoked_by = $1
			  WHERE user_id = $1 AND tenant_id = $2::uuid AND effective_to IS NULL`,
			userID, devTenant,
		)
		return err
	})
}

func withBypassErr(db *sql.DB, fn func(tx *sql.Tx) error) error {
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck
	if _, err := tx.ExecContext(ctx, `SELECT set_config('metaldocs.bypass_authz', 'scheduler', true)`); err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func activeMembershipCount(t *testing.T, db *sql.DB, userID, areaCode string) int {
	t.Helper()
	var n int
	if err := db.QueryRowContext(context.Background(),
		`SELECT count(*) FROM public.user_process_areas
		  WHERE user_id = $1 AND tenant_id = $2::uuid AND area_code = $3 AND effective_to IS NULL`,
		userID, devTenant, areaCode,
	).Scan(&n); err != nil {
		t.Fatalf("count active %s/%s: %v", userID, areaCode, err)
	}
	return n
}

// TestMembershipAreaScope_AreaAdmin_WithinManagedArea: area_admin grants and then
// revokes a membership inside the area where they hold membership.manage. Both
// succeed via tier-2 (area-scoped) authz.
func TestMembershipAreaScope_AreaAdmin_WithinManagedArea(t *testing.T) {
	db := openDB(t)
	svc := newAreaMembershipService(db)
	ctx := context.Background()

	areaAdmin := testdb.DeterministicID(t, "area-admin")
	target := testdb.DeterministicID(t, "target")
	seedIdentity(t, db, areaAdmin)
	seedIdentity(t, db, target)
	seedAreaAdminMembership(t, db, areaAdmin, areaManaged)
	t.Cleanup(func() { closeAllActive(db, target) })

	if err := svc.Grant(ctx, target, devTenant, areaManaged, iamdomain.RoleAuthor, areaAdmin); err != nil {
		t.Fatalf("area_admin grant within managed area = %v, want nil", err)
	}
	if got := activeMembershipCount(t, db, target, areaManaged); got != 1 {
		t.Fatalf("active rows after grant = %d, want 1", got)
	}
	if err := svc.Revoke(ctx, target, devTenant, areaManaged, areaAdmin); err != nil {
		t.Fatalf("area_admin revoke within managed area = %v, want nil", err)
	}
	if got := activeMembershipCount(t, db, target, areaManaged); got != 0 {
		t.Fatalf("active rows after revoke = %d, want 0", got)
	}
}

// TestMembershipAreaScope_AreaAdmin_OutsideManagedArea: area_admin attempting to
// grant/revoke in an area where they hold nothing is denied by tier-2 → ErrCapDenied.
func TestMembershipAreaScope_AreaAdmin_OutsideManagedArea(t *testing.T) {
	db := openDB(t)
	svc := newAreaMembershipService(db)
	ctx := context.Background()

	areaAdmin := testdb.DeterministicID(t, "area-admin")
	target := testdb.DeterministicID(t, "target")
	seedIdentity(t, db, areaAdmin)
	seedIdentity(t, db, target)
	seedAreaAdminMembership(t, db, areaAdmin, areaManaged) // managed area only
	t.Cleanup(func() { closeAllActive(db, target) })

	// Grant in the UNMANAGED area → denied before any row is written.
	err := svc.Grant(ctx, target, devTenant, areaUnmanaged, iamdomain.RoleAuthor, areaAdmin)
	if !isCapDenied(err) {
		t.Fatalf("area_admin grant outside managed area = %v, want ErrCapDenied", err)
	}
	if got := activeMembershipCount(t, db, target, areaUnmanaged); got != 0 {
		t.Fatalf("denied grant still wrote %d rows, want 0 (BOLA escalation guard)", got)
	}

	// Pre-seed an active row in the unmanaged area (via bypass), then confirm
	// the area_admin cannot revoke it either.
	withBypass(t, db, func(tx *sql.Tx) {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO public.user_process_areas
			   (user_id, tenant_id, area_code, role, effective_from, effective_to, granted_by, revoked_by)
			 VALUES ($1, $2::uuid, $3, 'author', now() - interval '1 hour', NULL, $4, NULL)`,
			target, devTenant, areaUnmanaged, areaAdmin,
		); err != nil {
			t.Fatalf("seed unmanaged active row: %v", err)
		}
	})
	if err := svc.Revoke(ctx, target, devTenant, areaUnmanaged, areaAdmin); !isCapDenied(err) {
		t.Fatalf("area_admin revoke outside managed area = %v, want ErrCapDenied", err)
	}
	if got := activeMembershipCount(t, db, target, areaUnmanaged); got != 1 {
		t.Fatalf("denied revoke closed the row (active=%d, want 1)", got)
	}
}

// TestMembershipAreaScope_SystemAdmin_BypassNotBlockedByMissingArea is the R1
// acceptance criterion (ADR 0022 amendment): a system_admin holding NO
// user_process_areas row is NOT blocked by the missing per-area grant — the
// tier-2 system_admin bypass short-circuits before the area-filtered query.
func TestMembershipAreaScope_SystemAdmin_BypassNotBlockedByMissingArea(t *testing.T) {
	db := openDB(t)
	svc := newAreaMembershipService(db)
	ctx := context.Background()

	sysAdmin := testdb.DeterministicID(t, "sys-admin")
	target := testdb.DeterministicID(t, "target")
	seedIdentity(t, db, sysAdmin)
	seedIdentity(t, db, target)
	seedSystemAdminRole(t, db, sysAdmin)
	t.Cleanup(func() { closeAllActive(db, target) })

	// No user_process_areas row for sysAdmin in ANY area, yet the grant succeeds
	// in an arbitrary area (and the revoke that follows).
	if err := svc.Grant(ctx, target, devTenant, areaUnmanaged, iamdomain.RoleAuthor, sysAdmin); err != nil {
		t.Fatalf("system_admin grant w/o area row = %v, want nil (R1 inheritance)", err)
	}
	if got := activeMembershipCount(t, db, target, areaUnmanaged); got != 1 {
		t.Fatalf("active rows after system_admin grant = %d, want 1", got)
	}
	if err := svc.Revoke(ctx, target, devTenant, areaUnmanaged, sysAdmin); err != nil {
		t.Fatalf("system_admin revoke w/o area row = %v, want nil (R1 inheritance)", err)
	}
}

// seedSystemAdminRole grants the actor a tenant-wide system_admin role row.
func seedSystemAdminRole(t *testing.T, db *sql.DB, userID string) {
	t.Helper()
	if err := withBypassErr(db, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(),
			`INSERT INTO metaldocs.iam_user_roles (user_id, role_code, tenant_id, assigned_by)
			 VALUES ($1, 'system_admin', $2::uuid, $1)
			 ON CONFLICT (tenant_id, user_id) DO UPDATE SET role_code = EXCLUDED.role_code`,
			userID, devTenant,
		)
		return err
	}); err != nil {
		t.Fatalf("seed system_admin role %s: %v", userID, err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), `DELETE FROM metaldocs.iam_user_roles WHERE user_id = $1`, userID) //nolint:errcheck
	})
}

func isCapDenied(err error) bool {
	var denied iamauthz.ErrCapDenied
	return errors.As(err, &denied)
}
