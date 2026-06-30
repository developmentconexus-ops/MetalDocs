package migrations_test

import (
	"os"
	"strings"
	"testing"
)

// TestMigration0251SelfRegistersInLedger locks the boot-safety + idempotency
// invariants for the template_version CHECK constraints migration (mirrors the
// 0250 convention). The runner does not auto-insert; the file must self-register
// in public.schema_migrations.
func TestMigration0251SelfRegistersInLedger(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile("0251_template_version_check_constraints.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	sql := string(body)

	if !strings.Contains(sql, "BEGIN;") || !strings.Contains(sql, "COMMIT;") {
		t.Fatal("migration must be wrapped in BEGIN/COMMIT")
	}

	// Idempotency: each constraint add must be guarded by a pg_constraint existence check.
	for _, conname := range []string{
		"chk_template_version_status",
		"chk_template_version_content_hash",
		"chk_template_version_content_hash_non_draft",
		"chk_template_version_number_positive",
	} {
		if !strings.Contains(sql, "'"+conname+"'") {
			t.Fatalf("migration must guard constraint %q with a pg_constraint existence check", conname)
		}
	}

	// Each constraint definition must be present.
	for _, fragment := range []string{
		// F-DB1: status closed-enum
		"'draft', 'in_review', 'approved', 'published', 'obsolete'",
		// F-DB4a: content_hash format
		"content_hash = '' OR length(content_hash) = 64",
		// F-DB4b: content_hash lifecycle
		"status = 'draft' OR length(content_hash) = 64",
		// F-DB7: version_number positivity
		"version_number >= 1",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("migration must contain CHECK fragment: %q", fragment)
		}
	}

	// Ledger self-registration: present, correct 4-digit version, idempotent.
	if !strings.Contains(sql, "INSERT INTO public.schema_migrations") {
		t.Fatal("migration must insert into public.schema_migrations")
	}
	if !strings.Contains(sql, "'0251'") {
		t.Fatal("migration must record schema_migrations version '0251'")
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

	// Idempotency pattern: guarded by DO $$ BEGIN ... END $$, not raw ALTER TABLE.
	if !strings.Contains(sql, "DO $$") {
		t.Fatal("migration must use DO $$ idempotency guard (not raw ALTER TABLE)")
	}
	if !strings.Contains(sql, "pg_constraint") {
		t.Fatal("migration must check pg_constraint for idempotent constraint adds")
	}
}
