package migrations_test

import (
	"os"
	"strings"
	"testing"
)

// TestMigration0256SelfRegistersInLedger locks the boot-safety + idempotency
// invariants for the templates_template_version tenant_id migration (mirrors
// 0253 convention). The runner does not auto-insert; the file must self-register
// in public.schema_migrations.
func TestMigration0256SelfRegistersInLedger(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile("0256_templates_template_version_tenant_id.sql")
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
	if !strings.Contains(sql, "'0256'") {
		t.Fatal("migration must record schema_migrations version '0256'")
	}
	if !strings.Contains(sql, "ON CONFLICT (version) DO NOTHING") {
		t.Fatal("schema_migrations insert must be idempotent (ON CONFLICT (version) DO NOTHING)")
	}
	insertIdx := strings.Index(sql, "INSERT INTO public.schema_migrations")
	commitIdx := strings.LastIndex(sql, "COMMIT;")
	if insertIdx < 0 || commitIdx < 0 || insertIdx > commitIdx {
		t.Fatal("ledger insert must appear inside the BEGIN/COMMIT block, before COMMIT")
	}
}

// TestMigration0256ColumnPresent asserts the migration adds tenant_id with NOT
// NULL and the ADD COLUMN uses IF NOT EXISTS for idempotency.
func TestMigration0256ColumnPresent(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile("0256_templates_template_version_tenant_id.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	sql := string(body)

	if !strings.Contains(sql, "ADD COLUMN IF NOT EXISTS tenant_id uuid") {
		t.Fatal("migration must add tenant_id column with IF NOT EXISTS for idempotency")
	}
	if !strings.Contains(sql, "ALTER COLUMN tenant_id SET NOT NULL") {
		t.Fatal("migration must set tenant_id NOT NULL after backfill")
	}
}

// TestMigration0256FKPresent asserts the FK to metaldocs.tenants is added.
func TestMigration0256FKPresent(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile("0256_templates_template_version_tenant_id.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	sql := string(body)

	if !strings.Contains(sql, "fk_templates_template_version_tenant") {
		t.Fatal("migration must add FK constraint fk_templates_template_version_tenant")
	}
	if !strings.Contains(sql, "REFERENCES metaldocs.tenants(id)") {
		t.Fatal("FK must reference metaldocs.tenants(id)")
	}
}

// TestMigration0256NoUniqueIndexChange asserts the migration deliberately does
// NOT touch the existing unique index/constraint (unlike 0253): template_id is a
// globally-unique uuid FK so the (template_id, version_number) uniqueness domain
// is already per-tenant. Guards against a future copy-paste re-introducing the
// 0253 step-4 index swap, which would be a no-op risk here.
func TestMigration0256NoUniqueIndexChange(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile("0256_templates_template_version_tenant_id.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	sql := string(body)

	if strings.Contains(sql, "DROP INDEX") {
		t.Fatal("migration must not drop any index — the existing unique domain is already per-tenant")
	}
	if strings.Contains(sql, "CREATE UNIQUE INDEX") {
		t.Fatal("migration must not create a new unique index — no key is widened by tenant_id")
	}
}

// TestMigration0256RLSEnabled asserts RLS is enabled, forced, and the
// tenant_isolation policy uses the canonical GUC-based USING clause from 0237.
func TestMigration0256RLSEnabled(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile("0256_templates_template_version_tenant_id.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	sql := string(body)

	if !strings.Contains(sql, "ENABLE ROW LEVEL SECURITY") {
		t.Fatal("migration must ENABLE ROW LEVEL SECURITY on templates_template_version")
	}
	if !strings.Contains(sql, "FORCE ROW LEVEL SECURITY") {
		t.Fatal("migration must FORCE ROW LEVEL SECURITY on templates_template_version")
	}
	if !strings.Contains(sql, "CREATE POLICY tenant_isolation ON public.templates_template_version") {
		t.Fatal("migration must create tenant_isolation policy on public.templates_template_version")
	}
	if !strings.Contains(sql, "current_setting('metaldocs.tenant_id', true)") {
		t.Fatal("tenant_isolation policy must key on metaldocs.tenant_id GUC")
	}
	if !strings.Contains(sql, "DROP POLICY IF EXISTS tenant_isolation ON public.templates_template_version") {
		t.Fatal("policy creation must be preceded by DROP POLICY IF EXISTS for idempotency")
	}
}

// TestMigration0256BackfillOrphanCheck asserts the migration raises loudly rather
// than silently skip version rows whose template_id has no matching tenant_id.
func TestMigration0256BackfillOrphanCheck(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile("0256_templates_template_version_tenant_id.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	sql := string(body)

	if !strings.Contains(sql, "RAISE EXCEPTION") {
		t.Fatal("migration must RAISE EXCEPTION when orphan rows are detected before backfill")
	}
}

// TestMigration0256TriggerBranchResolvesTenant asserts the CREATE OR REPLACE of
// enforce_capability_asserted changes the templates_template_version branch to
// resolve the real NEW.tenant_id (RLS parity), not NULL. SECURITY-CRITICAL: this
// is the one branch that may change; everything else is verbatim.
func TestMigration0256TriggerBranchResolvesTenant(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile("0256_templates_template_version_tenant_id.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	sql := string(body)

	if !strings.Contains(sql, "CREATE OR REPLACE FUNCTION public.enforce_capability_asserted()") {
		t.Fatal("migration must CREATE OR REPLACE the enforce_capability_asserted trigger function")
	}
	// The version branch must now resolve the real tenant_id. Locate the branch
	// header and assert its assignment is NEW.tenant_id, not NULL.
	const branchHeader = "WHEN TG_TABLE_NAME = 'templates_template_version' THEN"
	bi := strings.Index(sql, branchHeader)
	if bi < 0 {
		t.Fatal("trigger function must contain the templates_template_version CASE branch")
	}
	// The next ELSE/WHEN bounds the branch body.
	rest := sql[bi:]
	end := strings.Index(rest[len(branchHeader):], "WHEN TG_TABLE_NAME")
	if end < 0 {
		end = strings.Index(rest[len(branchHeader):], "ELSE")
	}
	branchBody := rest[:len(branchHeader)+end]
	if !strings.Contains(branchBody, "v_tenant_id     := NEW.tenant_id;") {
		t.Fatal("templates_template_version branch must resolve v_tenant_id := NEW.tenant_id (RLS parity)")
	}
	if strings.Contains(branchBody, "v_tenant_id     := NULL;") {
		t.Fatal("templates_template_version branch must no longer set v_tenant_id := NULL")
	}
}

// TestMigration0256BackfillBypassesCapTrigger asserts the backfill disables the
// capability trigger around the UPDATE (the table is gated by
// trg_require_cap_asserted for UPDATE; a raw migration UPDATE would otherwise be
// rejected) and re-enables it after — the canonical 0230 pattern.
func TestMigration0256BackfillBypassesCapTrigger(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile("0256_templates_template_version_tenant_id.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	sql := string(body)

	if !strings.Contains(sql, "DISABLE TRIGGER trg_require_cap_asserted") {
		t.Fatal("backfill must DISABLE TRIGGER trg_require_cap_asserted around the UPDATE")
	}
	if !strings.Contains(sql, "ENABLE TRIGGER trg_require_cap_asserted") {
		t.Fatal("backfill must re-ENABLE TRIGGER trg_require_cap_asserted after the UPDATE")
	}
	disIdx := strings.Index(sql, "DISABLE TRIGGER trg_require_cap_asserted")
	updIdx := strings.Index(sql, "UPDATE public.templates_template_version v")
	enIdx := strings.Index(sql, "ENABLE TRIGGER trg_require_cap_asserted")
	if !(disIdx < updIdx && updIdx < enIdx) {
		t.Fatal("trigger must be disabled before, and re-enabled after, the backfill UPDATE")
	}
}
