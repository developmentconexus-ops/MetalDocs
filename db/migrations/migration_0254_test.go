package migrations_test

import (
	"os"
	"strings"
	"testing"
)

// TestMigration0254SelfRegistersInLedger locks the boot-safety + idempotency
// invariants for the staging outbox dead_lettered_at migration (mirrors the
// 0250/0252/0253 convention). The runner does not auto-insert; the file must
// self-register in public.schema_migrations.
func TestMigration0254SelfRegistersInLedger(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile("0254_staging_outbox_dead_lettered_at.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	sql := string(body)

	if !strings.Contains(sql, "BEGIN;") || !strings.Contains(sql, "COMMIT;") {
		t.Fatal("migration must be wrapped in BEGIN/COMMIT")
	}
	if !strings.Contains(sql, "INSERT INTO public.schema_migrations") {
		t.Fatal("migration must insert into public.schema_migrations")
	}
	if !strings.Contains(sql, "'0254'") {
		t.Fatal("migration must record schema_migrations version '0254'")
	}
	if !strings.Contains(sql, "ON CONFLICT (version) DO NOTHING") {
		t.Fatal("schema_migrations insert must be idempotent (ON CONFLICT (version) DO NOTHING)")
	}
	// Ledger insert must sit inside the transaction (before COMMIT).
	insertIdx := strings.Index(sql, "INSERT INTO public.schema_migrations")
	commitIdx := strings.LastIndex(sql, "COMMIT;")
	if insertIdx < 0 || commitIdx < 0 || insertIdx > commitIdx {
		t.Fatal("ledger insert must appear inside the BEGIN/COMMIT block, before COMMIT")
	}
}

// TestMigration0254ColumnsPresentBothTables asserts the migration adds
// dead_lettered_at to both staging outbox tables using IF NOT EXISTS.
func TestMigration0254ColumnsPresentBothTables(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile("0254_staging_outbox_dead_lettered_at.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	sql := string(body)

	if !strings.Contains(sql, "pdf_dispatch_outbox") {
		t.Fatal("migration must alter pdf_dispatch_outbox")
	}
	if !strings.Contains(sql, "materialize_dispatch_outbox") {
		t.Fatal("migration must alter materialize_dispatch_outbox")
	}
	if !strings.Contains(sql, "dead_lettered_at timestamptz") {
		t.Fatal("migration must add dead_lettered_at timestamptz column")
	}
	if count := strings.Count(sql, "ADD COLUMN IF NOT EXISTS dead_lettered_at"); count != 2 {
		t.Fatalf("migration must use ADD COLUMN IF NOT EXISTS for both tables, got %d occurrences", count)
	}
}
