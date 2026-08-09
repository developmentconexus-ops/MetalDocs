//go:build integration
// +build integration

package infrastructure_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"testing"

	"metaldocs/internal/modules/templates/domain"
	"metaldocs/internal/modules/templates/infrastructure"
	"metaldocs/tests/integration/testdb"
)

// contentHashFor derives a syntactically valid 64-hex content_hash from a
// fixture's own docx storage key, so chk_template_version_content_hash_non_draft
// (migration 0317, ADR 0088) is satisfied with a value tied to the fixture
// rather than an unexplained literal. This test's subject is tenant-id RLS
// parity, not content-hash semantics, so the hash's provenance (storage key,
// not actual docx bytes) has no bearing on what is asserted below.
func contentHashFor(storageKey string) string {
	sum := sha256.Sum256([]byte(storageKey))
	return hex.EncodeToString(sum[:])
}

// TestTemplateVersion_TenantID_RLSParity proves migration 0256: the
// templates_template_version table now carries its own tenant_id column and is
// covered by the canonical FORCE-RLS tenant_isolation policy (REQ-TEN-1 / F-DB5),
// instead of relying solely on the JOIN to templates_template.
//
// Two independent guarantees are asserted:
//
//	(a) a normal version insert through the repository, in a tenant context,
//	    succeeds and persists the command's tenant_id onto the new row; and
//	(b) the RLS USING clause filters cross-tenant reads — under a NOBYPASSRLS
//	    role with metaldocs.tenant_id set to tenant A, tenant B's version row is
//	    invisible even on a direct table read with no JOIN/WHERE tenant predicate.
//
// (b) must SET ROLE to a NOBYPASSRLS role: the dev/test connection role is a
// superuser with BYPASSRLS (see migration 0234/0237 caveat, ADR 0022 Phase 5),
// for which RLS never applies. The production app role is NOSUPERUSER +
// NOBYPASSRLS, so this is the faithful way to exercise the deployed policy.
func TestTemplateVersion_TenantID_RLSParity(t *testing.T) {
	ctx := context.Background()
	db, dbName := testdb.OpenFreshDatabase(t)

	db.SetMaxIdleConns(0)
	db.SetMaxIdleConns(4)
	db.SetMaxOpenConns(4)

	tntA := testdb.NewTenant(t, db)
	tntB := testdb.NewTenant(t, db)
	actorA := testdb.NewUser(t, db, testdb.WithTenant(tntA.ID)).ID
	actorB := testdb.NewUser(t, db, testdb.WithTenant(tntB.ID)).ID

	templateA := testdb.DeterministicID(t, "rls-template-a")
	templateB := testdb.DeterministicID(t, "rls-template-b")
	versionA := testdb.DeterministicID(t, "rls-version-a")
	versionB := testdb.DeterministicID(t, "rls-version-b")

	repo := infrastructure.New(db)

	// ── (a) repository insert in tenant context persists tenant_id ───────────
	// Seed the parent templates raw (cap-gated), then insert versions through the
	// repository's CreateVersionTx so the production INSERT column list + binding
	// is the system under test.
	testdb.SeedWithCaps(t, db, `[{"cap":"template.create"}]`, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO public.templates_template (
				id, tenant_id, doc_type_code, key, name, latest_version, published_version_id, created_by
			) VALUES
				($1::uuid, $2::uuid, 'po', 'rls-tpl-a', 'RLS Template A', 1, NULL, $3),
				($4::uuid, $5::uuid, 'po', 'rls-tpl-b', 'RLS Template B', 1, NULL, $6)`,
			templateA, tntA.ID, actorA, templateB, tntB.ID, actorB,
		); err != nil {
			return err
		}
		// enforce_template_version_tenant_consistent (migration 0255) requires
		// the tx-local metaldocs.tenant_id GUC to match the parent template's
		// tenant; seed it per row alongside the asserted caps.
		if _, err := tx.ExecContext(ctx, `SELECT set_config('metaldocs.tenant_id', $1, true)`, tntA.ID); err != nil {
			return err
		}
		if err := repo.CreateVersionTx(ctx, tx, &domain.TemplateVersion{
			ID:                versionA,
			TenantID:          tntA.ID,
			TemplateID:        templateA,
			VersionNumber:     1,
			Status:            domain.VersionStatusDraft,
			MetadataSchema:    domain.MetadataSchema{},
			PlaceholderSchema: []domain.Placeholder{},
			AuthorID:          actorA,
			DocxStorageKey:    "templates/rls-a/body.docx",
			ContentHash:       contentHashFor("templates/rls-a/body.docx"),
		}); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `SELECT set_config('metaldocs.tenant_id', $1, true)`, tntB.ID); err != nil {
			return err
		}
		return repo.CreateVersionTx(ctx, tx, &domain.TemplateVersion{
			ID:                versionB,
			TenantID:          tntB.ID,
			TemplateID:        templateB,
			VersionNumber:     1,
			Status:            domain.VersionStatusDraft,
			MetadataSchema:    domain.MetadataSchema{},
			PlaceholderSchema: []domain.Placeholder{},
			AuthorID:          actorB,
			DocxStorageKey:    "templates/rls-b/body.docx",
			ContentHash:       contentHashFor("templates/rls-b/body.docx"),
		})
	})

	t.Run("insert_persists_tenant_id", func(t *testing.T) {
		var gotTenant string
		if err := db.QueryRowContext(ctx,
			`SELECT tenant_id::text FROM public.templates_template_version WHERE id = $1::uuid`,
			versionA,
		).Scan(&gotTenant); err != nil {
			t.Fatalf("read back version tenant_id: %v", err)
		}
		if gotTenant != tntA.ID {
			t.Fatalf("version A tenant_id = %q, want %q (insert must persist command tenant)", gotTenant, tntA.ID)
		}
	})

	// ── (b) RLS USING clause filters cross-tenant reads ──────────────────────
	t.Run("rls_filters_cross_tenant_read", func(t *testing.T) {
		// Create a NOBYPASSRLS role and grant it read on the table, so FORCE RLS
		// actually applies (the connection role is a BYPASSRLS superuser). The
		// role name is UNIQUE per fresh-clone database. Roles are cluster-global
		// but this test's db is a per-run clone: a FIXED global name collides
		// across runs when a prior clone's DROP DATABASE times out under load and
		// orphans the clone still holding this role's GRANT — the next run's
		// DROP ROLE then fails 2BP01 (dependent objects in the surviving orphan;
		// confirmed via pg_shdepend at the 4.6 exit gate). Deriving the name from
		// dbName makes every run's role dependency-free at creation. Cleanup does
		// DROP OWNED BY first so the role's grants in THIS db are revoked before
		// DROP ROLE — without it the cleanup silently leaks the role every run
		// (its GRANTs still exist when the cleanup fires, before the db is dropped).
		rlsRole := "rls_tester_tpl_version_" + dbName
		if _, err := db.ExecContext(ctx, `DROP ROLE IF EXISTS `+rlsRole); err != nil {
			t.Fatalf("drop role: %v", err)
		}
		if _, err := db.ExecContext(ctx, `CREATE ROLE `+rlsRole+` NOLOGIN NOBYPASSRLS`); err != nil {
			t.Fatalf("create role: %v", err)
		}
		t.Cleanup(func() {
			_, _ = db.ExecContext(ctx, `DROP OWNED BY `+rlsRole)
			_, _ = db.ExecContext(ctx, `DROP ROLE IF EXISTS `+rlsRole)
		})
		if _, err := db.ExecContext(ctx, `GRANT USAGE ON SCHEMA public TO `+rlsRole); err != nil {
			t.Fatalf("grant schema usage: %v", err)
		}
		if _, err := db.ExecContext(ctx, `GRANT SELECT ON public.templates_template_version TO `+rlsRole); err != nil {
			t.Fatalf("grant select: %v", err)
		}

		// Single connection so SET ROLE + GUC + query all run together.
		conn, err := db.Conn(ctx)
		if err != nil {
			t.Fatalf("acquire conn: %v", err)
		}
		defer conn.Close()
		defer func() { _, _ = conn.ExecContext(ctx, `RESET ROLE`) }()

		if _, err := conn.ExecContext(ctx, `SELECT set_config('metaldocs.tenant_id', $1, false)`, tntA.ID); err != nil {
			t.Fatalf("set tenant_id GUC: %v", err)
		}
		if _, err := conn.ExecContext(ctx, `SET ROLE `+rlsRole); err != nil {
			t.Fatalf("set role: %v", err)
		}

		// Direct table read, NO join, NO tenant WHERE predicate: isolation must
		// come entirely from the RLS policy. Tenant A's row visible, B's hidden.
		var visibleA bool
		if err := conn.QueryRowContext(ctx,
			`SELECT EXISTS (SELECT 1 FROM public.templates_template_version WHERE id = $1::uuid)`,
			versionA,
		).Scan(&visibleA); err != nil {
			t.Fatalf("query tenant A visibility: %v", err)
		}
		if !visibleA {
			t.Fatal("tenant A's own version must be visible under metaldocs.tenant_id = A")
		}

		var visibleB bool
		if err := conn.QueryRowContext(ctx,
			`SELECT EXISTS (SELECT 1 FROM public.templates_template_version WHERE id = $1::uuid)`,
			versionB,
		).Scan(&visibleB); err != nil {
			t.Fatalf("query tenant B visibility: %v", err)
		}
		if visibleB {
			t.Fatal("RLS breach: tenant B's version is visible while metaldocs.tenant_id = A — tenant_isolation policy not enforced")
		}
	})
}
