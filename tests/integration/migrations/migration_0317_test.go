//go:build integration

package migrations_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"metaldocs/internal/platform/tenant"
	"metaldocs/tests/integration/testdb"
)

// TestMigration0317_ContentHashAlwaysRequired proves ADR 0088's authoritative
// DB line against a real bootstrapped database (baseline + reference data +
// grants + the full forward tail — exactly what a deployed database has).
//
// The invariant: a template version row can never exist without a real
// content_hash, in ANY status. Before 0317 the CHECK carried a
// `status = 'draft' OR …` escape hatch, and that hatch was how the field's
// second, incompatible meaning ("the user has actually edited this") was
// expressed — by absence of the first. Removing it is what makes "version
// without content" unreachable rather than merely discouraged.
func TestMigration0317_ContentHashAlwaysRequired(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.OpenFreshDatabase(t)

	def := constraintDef(t, ctx, db, "public.templates_template_version", "chk_template_version_content_hash_non_draft")
	if def == "" {
		t.Fatal("chk_template_version_content_hash_non_draft is missing after bootstrap — ADR 0088's authoritative line is gone, not merely loosened")
	}
	// The escape hatch must be gone. Asserting on the definition text (which is
	// what pg_get_constraintdef reports, and therefore what a DBA reads) is the
	// direct statement of the rule; the behavioural proof follows below.
	if strings.Contains(def, "draft") {
		t.Fatalf("chk_template_version_content_hash_non_draft still carries a draft escape hatch after 0317: %s", def)
	}
	if !strings.Contains(def, "64") {
		t.Fatalf("chk_template_version_content_hash_non_draft no longer requires a 64-char hash after 0317: %s", def)
	}

	if n := ledgerCount(t, ctx, db, "0317"); n != 1 {
		t.Fatalf("expected exactly one schema_migrations row for version '0317' after bootstrap, got %d", n)
	}
}

// TestMigration0317_RejectsContentLessInsert is the behavioural half: the
// constraint must actually refuse a content-less DRAFT — the precise row shape
// that was legal before ADR 0088 and that produced the operator-reported
// 409 UPLOAD_MISSING on submitting a freshly created blank template.
//
// The insert is attempted with the same capability + tenant context the
// application would hold, so what is proved is the CHECK rejecting it, not a
// tripwire or an RLS policy rejecting it first for an unrelated reason.
func TestMigration0317_RejectsContentLessInsert(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.OpenFreshDatabase(t)

	for _, tc := range []struct {
		name string
		hash string
	}{
		{name: "empty hash", hash: ""},
		{name: "short hash", hash: "deadbeef"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := insertContentLessDraft(t, ctx, db, tc.hash)
			if err == nil {
				t.Fatalf("insert of a draft version with content_hash %q succeeded — ADR 0088's invariant is not enforced at the DB line", tc.hash)
			}
			if !strings.Contains(err.Error(), "content_hash") {
				t.Fatalf("insert failed for an unrelated reason (want a content_hash CHECK violation): %v", err)
			}
		})
	}
}

// TestMigration0317_NoContentLessRowsSurvive proves the backfill half: after
// bootstrap there is no content-less version row anywhere. On a fresh database
// the purge is a no-op over an empty set, so what this really pins is that
// neither the baseline nor the reference-data bundle seeds a row the tightened
// CHECK would reject — i.e. bootstrap and the invariant agree.
func TestMigration0317_NoContentLessRowsSurvive(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.OpenFreshDatabase(t)

	var n int
	if err := db.QueryRowContext(ctx, `
		SELECT count(*) FROM public.templates_template_version
		 WHERE content_hash IS NULL OR length(content_hash) <> 64`).Scan(&n); err != nil {
		t.Fatalf("count content-less template version rows: %v", err)
	}
	if n != 0 {
		t.Fatalf("expected zero content-less template version rows after bootstrap, got %d", n)
	}

	// And the system blank template — the row every from-scratch template now
	// materializes from (ADR 0088 §3) — must be present and carry a real hash.
	// Without it, creation fails closed, so this is a bootstrap precondition,
	// not a nicety.
	var key, hash string
	if err := db.QueryRowContext(ctx, `
		SELECT docx_storage_key, content_hash
		  FROM public.templates_template_version
		 WHERE template_id = '00000000-0000-0000-0000-000000000101'::uuid
		   AND status = 'published'`).Scan(&key, &hash); err != nil {
		t.Fatalf("load system blank template version (ADR 0088 materialization source): %v", err)
	}
	if key != "system/templates/blank.docx" {
		t.Fatalf("system blank docx_storage_key = %q, want system/templates/blank.docx", key)
	}
	if len(hash) != 64 {
		t.Fatalf("system blank content_hash is not a sha256 (len=%d): %q", len(hash), hash)
	}
}

func constraintDef(t *testing.T, ctx context.Context, db *sql.DB, table, name string) string {
	t.Helper()
	var def sql.NullString
	err := db.QueryRowContext(ctx, `
		SELECT pg_catalog.pg_get_constraintdef(c.oid)
		  FROM pg_catalog.pg_constraint c
		 WHERE c.conname = $1 AND c.conrelid = $2::regclass`, name, table).Scan(&def)
	if err == sql.ErrNoRows {
		return ""
	}
	if err != nil {
		t.Fatalf("read constraint definition %s on %s: %v", name, table, err)
	}
	return def.String
}

// insertContentLessDraft attempts the forbidden write with the capability and
// tenant context the application itself would hold, so a failure is attributable
// to the CHECK rather than to trg_require_cap_asserted or RLS.
func insertContentLessDraft(t *testing.T, ctx context.Context, db *sql.DB, hash string) error {
	t.Helper()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	testdb.SetCapsOnTx(t, tx, `[{"cap":"template.create"}]`)
	if _, err := tx.ExecContext(ctx, `SELECT set_config('metaldocs.tenant_id', $1, true)`,
		tenant.DevTenantID); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO public.templates_template_version (
			id, template_id, tenant_id, version_number, status,
			docx_storage_key, content_hash, metadata_schema, placeholder_schema, author_id
		) VALUES (
			gen_random_uuid(),
			'00000000-0000-0000-0000-000000000101'::uuid,
			'ffffffff-ffff-ffff-ffff-ffffffffffff'::uuid,
			9317,
			'draft',
			'system/templates/0317-probe.docx',
			$1,
			'{}'::jsonb, '[]'::jsonb, 'system'
		)`, hash)
	return err
}
