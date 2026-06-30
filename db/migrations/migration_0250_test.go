package migrations_test

import (
	"os"
	"strings"
	"testing"
)

// TestMigration0250SelfRegistersInLedger locks the boot-safety + idempotency
// invariants for the docx_storage_key de-share + UNIQUE migration (mirrors the
// 0248 convention). The runner does not auto-insert; the file must self-register
// in public.schema_migrations.
func TestMigration0250SelfRegistersInLedger(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile("0250_template_docx_storage_key_unique.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	sql := string(body)

	if !strings.Contains(sql, "BEGIN;") || !strings.Contains(sql, "COMMIT;") {
		t.Fatal("migration must be wrapped in BEGIN/COMMIT")
	}
	// Idempotency: the constraint add must be safe to re-run.
	if !strings.Contains(sql, "CREATE UNIQUE INDEX IF NOT EXISTS uq_templates_template_version_docx_storage_key") {
		t.Fatal("migration must add the docx_storage_key unique index with IF NOT EXISTS")
	}
	// Ledger self-registration: present, correct 4-digit version, idempotent.
	if !strings.Contains(sql, "INSERT INTO public.schema_migrations") {
		t.Fatal("migration must insert into public.schema_migrations")
	}
	if !strings.Contains(sql, "'0250'") {
		t.Fatal("migration must record schema_migrations version '0250'")
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
