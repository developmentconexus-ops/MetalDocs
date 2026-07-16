//go:build integration
// +build integration

// rls_truth_test.go — M7 F7.4 §4.5 "negative+positive proof": proves tenant
// RLS is genuinely enforced by connecting as the dedicated non-owner,
// NOSUPERUSER + NOBYPASSRLS role `metaldocs_ci` (migration 0284_ci_rls_role.sql)
// instead of the dev/test owner role metaldocs_app (SUPERUSER + BYPASSRLS +
// owner of every table). Every prior "RLS is enforced" assertion in this repo
// (see the bounded-limitation notes in tenant_crypto_test.go and the M6/F6.4
// GMR memory) ran against metaldocs_app and was a FALSE GREEN: RLS was
// unconditionally bypassed regardless of policy correctness, because a
// superuser/bypassrls/table-owner connection never evaluates row security
// policies at all. This test kills that false green by proving isolation
// under a role that cannot bypass RLS.
//
// POLICY IDIOM: every tenant_isolation policy in this codebase is
//
//	(NULLIF(current_setting('metaldocs.tenant_id', true), '') IS NULL)
//	OR (<tenant column> = (NULLIF(current_setting('metaldocs.tenant_id', true), ''))::uuid)
//
// So an UNSET/empty GUC passes ALL rows — a deliberate escape hatch for
// janitors/bootstrap code that runs without a tenant context (ADR 0027
// amendment). This means the isolation proof below is "wrong-tenant GUC
// blocks the other tenant's row", not "no-GUC blocks everything" — case (d)
// below pins the escape hatch explicitly so it is never mistaken for a leak.
//
//	go test -tags=integration ./tests/integration/security/... -run TestRLSTruth
package security_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"metaldocs/tests/integration/testdb"
)

func TestRLSTruth_NonOwnerRoleEnforcesIsolation(t *testing.T) {
	db, dbName := testdb.OpenFreshDatabase(t) // owner handle: metaldocs_app (superuser+bypass) — seeding only.

	// Timeout scopes only this test's own queries — created AFTER Open() so the
	// one-time template-clone/rebuild (which can take minutes on first run in a
	// process) is not counted against it.
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// ── Seed: two tenants + one metaldocs.tenant_keys row each ────────────────
	// tenant_keys is FORCE-RLS + tenant_isolation policy (migration 0281) and,
	// crucially, carries NO cap-asserted INSERT tripwire (unlike
	// tenant_lifecycle_jobs/tenants/iam_users), so the owner can seed rows
	// directly without asserting a capability — keeping this proof focused on
	// RLS, not on the tripwire layer. One row per tenant (tenant_id is PK/FK).
	tenantA := testdb.NewTenant(t, db)
	tenantB := testdb.NewTenant(t, db)

	if _, err := db.ExecContext(ctx,
		`INSERT INTO metaldocs.tenant_keys (tenant_id, wrapped_dek) VALUES ($1::uuid, $2)`,
		tenantA.ID, []byte{0x00},
	); err != nil {
		t.Fatalf("seed tenant A tenant_keys row: %v", err)
	}
	if _, err := db.ExecContext(ctx,
		`INSERT INTO metaldocs.tenant_keys (tenant_id, wrapped_dek) VALUES ($1::uuid, $2)`,
		tenantB.ID, []byte{0x01},
	); err != nil {
		t.Fatalf("seed tenant B tenant_keys row: %v", err)
	}

	// ── Role-attribute assertions (real catalog, owner conn) ──────────────────
	var rolsuper, rolbypassrls bool
	if err := db.QueryRowContext(ctx,
		`SELECT rolsuper, rolbypassrls FROM pg_roles WHERE rolname = 'metaldocs_ci'`,
	).Scan(&rolsuper, &rolbypassrls); err != nil {
		t.Fatalf("read pg_roles for metaldocs_ci: %v", err)
	}
	if rolsuper {
		t.Fatalf("metaldocs_ci rolsuper = true, want false (a superuser bypasses RLS unconditionally)")
	}
	if rolbypassrls {
		t.Fatalf("metaldocs_ci rolbypassrls = true, want false (BYPASSRLS makes RLS inert for this role)")
	}

	var ownedTables int
	if err := db.QueryRowContext(ctx,
		`SELECT count(*) FROM pg_class c
		   JOIN pg_roles r ON c.relowner = r.oid
		  WHERE r.rolname = 'metaldocs_ci' AND c.relkind = 'r'`,
	).Scan(&ownedTables); err != nil {
		t.Fatalf("count tables owned by metaldocs_ci: %v", err)
	}
	if ownedTables != 0 {
		t.Fatalf("metaldocs_ci owns %d tables, want 0 (a table owner is exempt from its own RLS unless FORCE is set)", ownedTables)
	}

	// ── Connect as the non-owner CI role for the actual isolation reads ───────
	ci := testdb.OpenAsCIRole(t, dbName)

	// (a) POSITIVE: GUC = tenant A -> only A's row is visible, 0 of B's.
	func() {
		tx, err := ci.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("(a) begin tx: %v", err)
		}
		defer tx.Rollback()

		if _, err := tx.ExecContext(ctx, `SET LOCAL metaldocs.tenant_id = '`+tenantA.ID+`'`); err != nil {
			t.Fatalf("(a) SET LOCAL tenant_id=A: %v", err)
		}

		var total int
		if err := tx.QueryRowContext(ctx,
			`SELECT count(*) FROM metaldocs.tenant_keys WHERE tenant_id IN ($1::uuid, $2::uuid)`,
			tenantA.ID, tenantB.ID,
		).Scan(&total); err != nil {
			t.Fatalf("(a) count seeded rows under tenant A GUC: %v", err)
		}
		if total != 1 {
			t.Fatalf("(a) tenant_keys visible seeded-row count under tenant A GUC = %d, want 1 (RLS must hide tenant B's row)", total)
		}

		var bVisible int
		if err := tx.QueryRowContext(ctx,
			`SELECT count(*) FROM metaldocs.tenant_keys WHERE tenant_id = $1::uuid`, tenantB.ID,
		).Scan(&bVisible); err != nil {
			t.Fatalf("(a) count tenant B's key under tenant A GUC: %v", err)
		}
		if bVisible != 0 {
			t.Fatalf("(a) tenant B's key visible under tenant A GUC = %d rows, want 0 (cross-tenant leak)", bVisible)
		}
	}()

	// (b) ISOLATION/NEGATIVE: GUC = tenant B -> tenant A's specific job row is invisible.
	func() {
		tx, err := ci.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("(b) begin tx: %v", err)
		}
		defer tx.Rollback()

		if _, err := tx.ExecContext(ctx, `SET LOCAL metaldocs.tenant_id = '`+tenantB.ID+`'`); err != nil {
			t.Fatalf("(b) SET LOCAL tenant_id=B: %v", err)
		}

		var aVisible int
		if err := tx.QueryRowContext(ctx,
			`SELECT count(*) FROM metaldocs.tenant_keys WHERE tenant_id = $1::uuid`, tenantA.ID,
		).Scan(&aVisible); err != nil {
			t.Fatalf("(b) count tenant A's key under tenant B GUC: %v", err)
		}
		if aVisible != 0 {
			t.Fatalf("(b) tenant A's key visible under tenant B GUC = %d rows, want 0 -- RLS did NOT block cross-tenant read under the non-owner role", aVisible)
		}
	}()

	// (c) BYPASS FALSE-GREEN CONTRAST: same wrong-tenant GUC on the OWNER conn
	// (metaldocs_app, superuser+bypass) DOES see tenant A's row -- this is the
	// exact false green that (a)/(b) above are proving is now closed for the
	// non-owner role. Documented, not fixed: bypass is expected owner behavior.
	func() {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("(c) begin owner tx: %v", err)
		}
		defer tx.Rollback()

		if _, err := tx.ExecContext(ctx, `SET LOCAL metaldocs.tenant_id = '`+tenantB.ID+`'`); err != nil {
			t.Fatalf("(c) SET LOCAL tenant_id=B on owner conn: %v", err)
		}

		var aVisibleToOwner int
		if err := tx.QueryRowContext(ctx,
			`SELECT count(*) FROM metaldocs.tenant_keys WHERE tenant_id = $1::uuid`, tenantA.ID,
		).Scan(&aVisibleToOwner); err != nil {
			t.Fatalf("(c) count tenant A's key under tenant B GUC on owner conn: %v", err)
		}
		if aVisibleToOwner != 1 {
			t.Fatalf("(c) tenant A's key visible to OWNER conn under wrong-tenant GUC = %d, want 1 (documents the pre-F7.4 false-green: owner/superuser/bypassrls ignores RLS entirely)", aVisibleToOwner)
		}
	}()

	// (d) NULL-GUC ESCAPE HATCH PIN: no GUC set at all -> both rows visible.
	// This is the deliberate janitor/bootstrap branch of the policy idiom
	// (NULLIF(...) IS NULL -> pass all rows), NOT a leak. Pinning it here so a
	// future change to the idiom can't silently flip it into an actual leak
	// without failing this assertion.
	func() {
		tx, err := ci.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("(d) begin tx: %v", err)
		}
		defer tx.Rollback()

		var total int
		if err := tx.QueryRowContext(ctx,
			`SELECT count(*) FROM metaldocs.tenant_keys WHERE tenant_id IN ($1::uuid, $2::uuid)`,
			tenantA.ID, tenantB.ID,
		).Scan(&total); err != nil {
			t.Fatalf("(d) count seeded rows with no GUC set: %v", err)
		}
		if total != 2 {
			t.Fatalf("(d) tenant_keys visible seeded-row count with no GUC set = %d, want 2 (the null-GUC branch is a deliberate escape hatch for GUC-less janitor/bootstrap sessions, not a leak -- see ADR 0027 amendment)", total)
		}
	}()

	// ── §4.4 approval_signoffs: catalog + non-owner queryability ──────────────
	var forceRLS bool
	if err := db.QueryRowContext(ctx,
		`SELECT relforcerowsecurity FROM pg_class WHERE relname = 'approval_signoffs'`,
	).Scan(&forceRLS); err != nil {
		t.Fatalf("read relforcerowsecurity for approval_signoffs: %v", err)
	}
	if !forceRLS {
		t.Fatalf("approval_signoffs relforcerowsecurity = false, want true")
	}

	var policyCount int
	if err := db.QueryRowContext(ctx,
		`SELECT count(*) FROM pg_policy p
		   JOIN pg_class c ON p.polrelid = c.oid
		  WHERE c.relname = 'approval_signoffs' AND p.polname = 'tenant_isolation'`,
	).Scan(&policyCount); err != nil {
		t.Fatalf("count tenant_isolation policy on approval_signoffs: %v", err)
	}
	if policyCount != 1 {
		t.Fatalf("approval_signoffs tenant_isolation policy count = %d, want 1", policyCount)
	}

	var policyExpr string
	if err := db.QueryRowContext(ctx,
		`SELECT pg_get_expr(p.polqual, p.polrelid)
		   FROM pg_policy p
		   JOIN pg_class c ON p.polrelid = c.oid
		  WHERE c.relname = 'approval_signoffs' AND p.polname = 'tenant_isolation'`,
	).Scan(&policyExpr); err != nil {
		t.Fatalf("read tenant_isolation policy expr on approval_signoffs: %v", err)
	}
	if !strings.Contains(policyExpr, "actor_tenant_id") {
		t.Fatalf("approval_signoffs tenant_isolation policy expr = %q, want it to reference actor_tenant_id", policyExpr)
	}

	// Non-owner queryability: the table is empty (heavy BEFORE-INSERT triggers
	// mean this test does not seed real signoffs), but the query must succeed
	// with no permission error under metaldocs_ci -- proving the FORCE RLS
	// policy is active and queryable for a non-owner role, not merely present
	// in the catalog.
	func() {
		tx, err := ci.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("(signoffs) begin tx: %v", err)
		}
		defer tx.Rollback()

		if _, err := tx.ExecContext(ctx, `SET LOCAL metaldocs.tenant_id = '`+uuid.NewString()+`'`); err != nil {
			t.Fatalf("(signoffs) SET LOCAL tenant_id: %v", err)
		}

		var count int
		if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM public.approval_signoffs`).Scan(&count); err != nil {
			t.Fatalf("(signoffs) SELECT count(*) under metaldocs_ci: %v (query must succeed -- RLS-active, not a permission error)", err)
		}
		if count != 0 {
			t.Fatalf("(signoffs) count = %d, want 0 (no rows seeded in this test)", count)
		}
	}()
}
