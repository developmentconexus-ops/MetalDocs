//go:build integration
// +build integration

package authz_test

import (
	"context"
	"strings"
	"testing"

	"metaldocs/internal/modules/iam/authz"
	"metaldocs/tests/integration/testdb"
)

// TestSeedTxTenant_RLSBackstop_LeakBeforeBlockedAfter is the F3.2 negative RLS
// integration proof (validation-contract.md §2.5, spec.md PG-4) — the only
// test class that can pin the async RLS backstop end-to-end (application
// tests are sqlmock and cannot exercise a real Postgres RLS policy).
//
// Table: metaldocs.notifications (NOT public.documents). documents carries the
// M2 capability write-tripwire (trg_require_cap_asserted -> P0001 when
// metaldocs.asserted_caps is unset), which fires BEFORE the RLS tenant_isolation
// policy is the deciding control — so a documents-based proof conflates two
// controls and the leak-before / retenant assertions observe P0001 instead of
// the RLS outcome (see F3.5). metaldocs.notifications is a FORCE-RLS
// tenant_isolation table (0001_current_schema.sql:1437,1453,4632) that carries
// no capability tripwire and is a real F3.2 async seed site (notifications
// fanout worker), so RLS is the sole deciding control end-to-end.
//
// Shape (binding, contract §2.5):
//   - Setup: two tenants A and B, each with a metaldocs.notifications row.
//   - NEGATIVE-before / leak demonstration: in a tx with NO tenant GUC seeded
//     (today's pre-F3.2 worker behavior), a cross-tenant SELECT/UPDATE against
//     tenant B's row succeeds/sees the row.
//   - POSITIVE-after / backstop engaged: in a tx that calls
//     authz.SeedTxTenant(ctx, tx, A) first, the same cross-tenant access to
//     tenant B's row is blocked: SELECT/UPDATE/DELETE -> 0 rows; INSERT/UPDATE
//     producing a B-tenant row -> error 42501 (new row violates row-level
//     security policy).
//
// Must run under a NOBYPASSRLS role (the dev/test connection role is a
// BYPASSRLS superuser, for which RLS never applies) — mirrors
// internal/modules/templates/repository/tenant_id_rls_integration_test.go.
//
// Run scope (binding): `go test -tags integration -run 'RLS|Tenant|Isolation' ./...`
// only — never the full suite.
func TestSeedTxTenant_RLSBackstop_LeakBeforeBlockedAfter(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)

	tntA := testdb.NewTenant(t, db)
	tntB := testdb.NewTenant(t, db)

	notifA := testdb.NewNotification(t, db, testdb.WithTenant(tntA.ID))
	notifB := testdb.NewNotification(t, db, testdb.WithTenant(tntB.ID))

	// A NOBYPASSRLS role is required to faithfully exercise FORCE RLS — the
	// pooled dev/test connection role is a BYPASSRLS superuser.
	const rlsRole = "rls_tester_async_tenant_seed"
	if _, err := db.ExecContext(ctx, `DROP ROLE IF EXISTS `+rlsRole); err != nil {
		t.Fatalf("drop role: %v", err)
	}
	if _, err := db.ExecContext(ctx, `CREATE ROLE `+rlsRole+` NOLOGIN NOBYPASSRLS`); err != nil {
		t.Fatalf("create role: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, `DROP ROLE IF EXISTS `+rlsRole) })
	if _, err := db.ExecContext(ctx, `GRANT USAGE ON SCHEMA public TO `+rlsRole); err != nil {
		t.Fatalf("grant schema usage: %v", err)
	}
	if _, err := db.ExecContext(ctx, `GRANT USAGE ON SCHEMA metaldocs TO `+rlsRole); err != nil {
		t.Fatalf("grant schema usage: %v", err)
	}
	if _, err := db.ExecContext(ctx, `GRANT SELECT, UPDATE, INSERT, DELETE ON metaldocs.notifications TO `+rlsRole); err != nil {
		t.Fatalf("grant dml: %v", err)
	}

	// ── NEGATIVE-before: no tenant GUC seeded — cross-tenant leak ────────────
	t.Run("leak_before_no_seed", func(t *testing.T) {
		conn, err := db.Conn(ctx)
		if err != nil {
			t.Fatalf("acquire conn: %v", err)
		}
		defer conn.Close()
		defer func() { _, _ = conn.ExecContext(ctx, `RESET ROLE`) }()

		if _, err := conn.ExecContext(ctx, `SET ROLE `+rlsRole); err != nil {
			t.Fatalf("set role: %v", err)
		}

		tx, err := conn.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}
		defer func() { _ = tx.Rollback() }()

		// No authz.SeedTxTenant call in this tx — this is today's pre-fix
		// worker/jobs behavior. NULL-permissive tenant_isolation policy means
		// an unset metaldocs.tenant_id GUC lets a cross-tenant row through.
		var visibleB bool
		if err := tx.QueryRowContext(ctx,
			`SELECT EXISTS (SELECT 1 FROM metaldocs.notifications WHERE id = $1::uuid)`,
			notifB.ID,
		).Scan(&visibleB); err != nil {
			t.Fatalf("query tenant B visibility (unseeded): %v", err)
		}
		if !visibleB {
			t.Fatal("leak demonstration failed: expected tenant B's row to be visible with NO tenant GUC seeded (NULL-permissive policy) — if this now fails, the backstop may already be engaged some other way; investigate before trusting the after-seed assertion")
		}

		res, err := tx.ExecContext(ctx,
			`UPDATE metaldocs.notifications SET status = 'SENT' WHERE id = $1::uuid`,
			notifB.ID,
		)
		if err != nil {
			t.Fatalf("cross-tenant update (unseeded) should succeed pre-fix, got error: %v", err)
		}
		n, err := res.RowsAffected()
		if err != nil {
			t.Fatalf("rows affected: %v", err)
		}
		if n != 1 {
			t.Fatalf("leak demonstration failed: expected the unseeded cross-tenant UPDATE to affect tenant B's row (1), got %d", n)
		}
	})

	// ── POSITIVE-after: authz.SeedTxTenant(A) seeded — backstop engaged ──────
	t.Run("blocked_after_seed_tenant_a", func(t *testing.T) {
		conn, err := db.Conn(ctx)
		if err != nil {
			t.Fatalf("acquire conn: %v", err)
		}
		defer conn.Close()
		defer func() { _, _ = conn.ExecContext(ctx, `RESET ROLE`) }()

		if _, err := conn.ExecContext(ctx, `SET ROLE `+rlsRole); err != nil {
			t.Fatalf("set role: %v", err)
		}

		t.Run("select_update_delete_see_zero_rows", func(t *testing.T) {
			tx, err := conn.BeginTx(ctx, nil)
			if err != nil {
				t.Fatalf("begin tx: %v", err)
			}
			defer func() { _ = tx.Rollback() }()

			if err := authz.SeedTxTenant(ctx, tx, tntA.ID); err != nil {
				t.Fatalf("SeedTxTenant: %v", err)
			}

			// SELECT of B's row -> invisible (0 rows), even with no WHERE tenant
			// predicate — isolation comes entirely from the RLS policy.
			var visibleB bool
			if err := tx.QueryRowContext(ctx,
				`SELECT EXISTS (SELECT 1 FROM metaldocs.notifications WHERE id = $1::uuid)`,
				notifB.ID,
			).Scan(&visibleB); err != nil {
				t.Fatalf("query tenant B visibility (seeded A): %v", err)
			}
			if visibleB {
				t.Fatal("RLS breach: tenant B's notification is visible via SELECT while metaldocs.tenant_id = A")
			}

			// UPDATE of B's row -> 0 rows affected (row invisible to the tx).
			res, err := tx.ExecContext(ctx,
				`UPDATE metaldocs.notifications SET status = 'READ' WHERE id = $1::uuid`,
				notifB.ID,
			)
			if err != nil {
				t.Fatalf("cross-tenant update (seeded A) errored instead of affecting 0 rows: %v", err)
			}
			n, err := res.RowsAffected()
			if err != nil {
				t.Fatalf("rows affected: %v", err)
			}
			if n != 0 {
				t.Fatalf("RLS breach: cross-tenant UPDATE of tenant B's row affected %d rows while metaldocs.tenant_id = A, want 0", n)
			}

			// DELETE of B's row -> 0 rows affected.
			res, err = tx.ExecContext(ctx, `DELETE FROM metaldocs.notifications WHERE id = $1::uuid`, notifB.ID)
			if err != nil {
				t.Fatalf("cross-tenant delete (seeded A) errored instead of affecting 0 rows: %v", err)
			}
			n, err = res.RowsAffected()
			if err != nil {
				t.Fatalf("rows affected: %v", err)
			}
			if n != 0 {
				t.Fatalf("RLS breach: cross-tenant DELETE of tenant B's row affected %d rows while metaldocs.tenant_id = A, want 0", n)
			}

			// Tenant A's own row remains visible under its own seed — proves the
			// block above is tenant-scoped, not a blanket RLS lockout.
			var visibleA bool
			if err := tx.QueryRowContext(ctx,
				`SELECT EXISTS (SELECT 1 FROM metaldocs.notifications WHERE id = $1::uuid)`,
				notifA.ID,
			).Scan(&visibleA); err != nil {
				t.Fatalf("query tenant A visibility (seeded A): %v", err)
			}
			if !visibleA {
				t.Fatal("tenant A's own notification must remain visible under metaldocs.tenant_id = A")
			}
		})

		t.Run("insert_or_update_producing_b_row_is_42501", func(t *testing.T) {
			tx, err := conn.BeginTx(ctx, nil)
			if err != nil {
				t.Fatalf("begin tx: %v", err)
			}
			defer func() { _ = tx.Rollback() }()

			if err := authz.SeedTxTenant(ctx, tx, tntA.ID); err != nil {
				t.Fatalf("SeedTxTenant: %v", err)
			}

			// An UPDATE that would retarget tenant A's own notification row to
			// tenant B violates the WITH CHECK half of tenant_isolation (the
			// USING clause doubles as WITH CHECK per the policy shape) ->
			// 42501, not a silent 0-row no-op.
			_, err = tx.ExecContext(ctx,
				`UPDATE metaldocs.notifications SET tenant_id = $2::uuid WHERE id = $1::uuid`,
				notifA.ID, tntB.ID,
			)
			if err == nil {
				t.Fatal("RLS breach: UPDATE re-tenanting tenant A's row to tenant B must be rejected by the WITH CHECK policy while metaldocs.tenant_id = A")
			}
			if !strings.Contains(err.Error(), "42501") {
				t.Fatalf("expected SQLSTATE 42501 (row-level security policy violation), got: %v", err)
			}
		})
	})
}
