package migrations_test

import (
	"os"
	"strings"
	"testing"
)

// TestMigration0252SelfRegistersInLedger locks the boot-safety + idempotency
// invariants for the one-published-per-template partial-unique index migration
// (mirrors the 0250 convention). The runner does not auto-insert; the file must
// self-register in public.schema_migrations.
func TestMigration0252SelfRegistersInLedger(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile("0252_template_one_published_per_template_unique.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	sql := string(body)

	if !strings.Contains(sql, "BEGIN;") || !strings.Contains(sql, "COMMIT;") {
		t.Fatal("migration must be wrapped in BEGIN/COMMIT")
	}
	// Idempotency: the index add must be safe to re-run.
	if !strings.Contains(sql, "CREATE UNIQUE INDEX IF NOT EXISTS uq_one_published_per_template") {
		t.Fatal("migration must add the one-published-per-template unique index with IF NOT EXISTS")
	}
	// Partial predicate must filter to published rows only.
	if !strings.Contains(sql, "WHERE status = 'published'") {
		t.Fatal("migration unique index must carry the partial predicate WHERE status = 'published'")
	}
	// Ledger self-registration: present, correct 4-digit version, idempotent.
	if !strings.Contains(sql, "INSERT INTO public.schema_migrations") {
		t.Fatal("migration must insert into public.schema_migrations")
	}
	if !strings.Contains(sql, "'0252'") {
		t.Fatal("migration must record schema_migrations version '0252'")
	}
	if !strings.Contains(sql, "ON CONFLICT (version) DO NOTHING") {
		t.Fatal("schema_migrations insert must be idempotent (ON CONFLICT (version) DO NOTHING)")
	}
	// The ledger insert must sit inside the transaction (before COMMIT).
	insertIdx := strings.Index(sql, "INSERT INTO public.schema_migrations")
	commitIdx := strings.LastIndex(sql, "COMMIT;")
	if insertIdx < 0 || commitIdx < 0 || insertIdx > commitIdx {
		t.Fatal("ledger insert must appear inside the BEGIN/COMMIT block, before COMMIT")
	}
}

// TestMigration0252DoesNotRecreateRevisionIndex confirms that migration 0252
// does NOT attempt to recreate ux_templates_version_revision, which was
// already added by migration 0233. Recreating it would fail on any DB that
// has already run 0233 (even with IF NOT EXISTS, it signals misattribution).
func TestMigration0252DoesNotRecreateRevisionIndex(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile("0252_template_one_published_per_template_unique.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	sql := string(body)

	if strings.Contains(sql, "ux_templates_version_revision") {
		t.Fatal("migration 0252 must NOT touch ux_templates_version_revision (owned by 0233)")
	}
}
