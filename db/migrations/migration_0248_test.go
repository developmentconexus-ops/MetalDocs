package migrations_test

import (
	"os"
	"strings"
	"testing"
)

// TestMigration0248SelfRegistersInLedger locks the boot-safety invariant: the
// migration must self-register in the schema_migrations ledger so the runner
// records 0248 as applied and does not re-execute it on every boot (the runner
// does not auto-insert; files must include the ledger insert — see
// internal/platform/migrate/migrate.go). Convention (0240–0246):
// INSERT ... VALUES ('0248', ...) ON CONFLICT (version) DO NOTHING.
func TestMigration0248SelfRegistersInLedger(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile("0248_token_dictionary_entries.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	sql := string(body)

	// Must be wrapped in an explicit transaction (idempotent style).
	if !strings.Contains(sql, "BEGIN;") || !strings.Contains(sql, "COMMIT;") {
		t.Fatal("migration must be wrapped in BEGIN/COMMIT")
	}

	// Idempotency: re-running the rest of the file must be safe.
	if !strings.Contains(sql, "CREATE TABLE IF NOT EXISTS metaldocs.token_dictionary_entries") {
		t.Fatal("migration must create metaldocs.token_dictionary_entries with IF NOT EXISTS (idempotent)")
	}

	// Ledger self-registration: present, correct 4-digit version, idempotent.
	if !strings.Contains(sql, "INSERT INTO public.schema_migrations") {
		t.Fatal("migration must insert into public.schema_migrations")
	}
	if !strings.Contains(sql, "'0248'") {
		t.Fatal("migration must record schema_migrations version '0248'")
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
