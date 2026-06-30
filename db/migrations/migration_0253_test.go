package migrations_test

import (
	"os"
	"strings"
	"testing"
)

// TestMigration0253SelfRegistersInLedger locks the boot-safety + idempotency
// invariants for the document_exports tenant_id migration (mirrors 0250/0252
// convention). The runner does not auto-insert; the file must self-register in
// public.schema_migrations.
func TestMigration0253SelfRegistersInLedger(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile("0253_document_exports_tenant_id.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	sql := string(body)

	if !strings.Contains(sql, "BEGIN;") || !strings.Contains(sql, "COMMIT;") {
		t.Fatal("migration must be wrapped in BEGIN/COMMIT")
	}
	// Ledger self-registration: present, correct 4-digit version, idempotent.
	if !strings.Contains(sql, "INSERT INTO public.schema_migrations") {
		t.Fatal("migration must insert into public.schema_migrations")
	}
	if !strings.Contains(sql, "'0253'") {
		t.Fatal("migration must record schema_migrations version '0253'")
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

// TestMigration0253ColumnPresent asserts the migration adds tenant_id with NOT
// NULL and the ADD COLUMN uses IF NOT EXISTS for idempotency.
func TestMigration0253ColumnPresent(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile("0253_document_exports_tenant_id.sql")
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

// TestMigration0253FKPresent asserts the FK to metaldocs.tenants is added.
func TestMigration0253FKPresent(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile("0253_document_exports_tenant_id.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	sql := string(body)

	if !strings.Contains(sql, "fk_document_exports_tenant") {
		t.Fatal("migration must add FK constraint fk_document_exports_tenant")
	}
	if !strings.Contains(sql, "REFERENCES metaldocs.tenants(id)") {
		t.Fatal("FK must reference metaldocs.tenants(id)")
	}
}

// TestMigration0253UniqueIndexPresent asserts the tenant-scoped unique index
// is created and the old per-document index is dropped.
func TestMigration0253UniqueIndexPresent(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile("0253_document_exports_tenant_id.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	sql := string(body)

	if !strings.Contains(sql, "CREATE UNIQUE INDEX IF NOT EXISTS uq_document_exports_tenant_doc_hash") {
		t.Fatal("migration must create uq_document_exports_tenant_doc_hash unique index with IF NOT EXISTS")
	}
	if !strings.Contains(sql, "(tenant_id, document_id, composite_hash)") {
		t.Fatal("unique index must cover (tenant_id, document_id, composite_hash)")
	}
	if !strings.Contains(sql, "DROP INDEX IF EXISTS public.document_exports_doc_hash_uidx") {
		t.Fatal("migration must drop the old document_exports_doc_hash_uidx index")
	}
}

// TestMigration0253RLSEnabled asserts RLS is enabled, forced, and the
// tenant_isolation policy uses the canonical GUC-based USING clause from 0237.
func TestMigration0253RLSEnabled(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile("0253_document_exports_tenant_id.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	sql := string(body)

	if !strings.Contains(sql, "ENABLE ROW LEVEL SECURITY") {
		t.Fatal("migration must ENABLE ROW LEVEL SECURITY on document_exports")
	}
	if !strings.Contains(sql, "FORCE ROW LEVEL SECURITY") {
		t.Fatal("migration must FORCE ROW LEVEL SECURITY on document_exports")
	}
	if !strings.Contains(sql, "CREATE POLICY tenant_isolation ON public.document_exports") {
		t.Fatal("migration must create tenant_isolation policy on public.document_exports")
	}
	if !strings.Contains(sql, "current_setting('metaldocs.tenant_id', true)") {
		t.Fatal("tenant_isolation policy must key on metaldocs.tenant_id GUC")
	}
	if !strings.Contains(sql, "DROP POLICY IF EXISTS tenant_isolation ON public.document_exports") {
		t.Fatal("policy creation must be preceded by DROP POLICY IF EXISTS for idempotency")
	}
}

// TestMigration0253BackfillOrphanCheck asserts the migration will raise loudly
// rather than silently skip rows whose document_id has no matching tenant_id.
func TestMigration0253BackfillOrphanCheck(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile("0253_document_exports_tenant_id.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	sql := string(body)

	if !strings.Contains(sql, "RAISE EXCEPTION") {
		t.Fatal("migration must RAISE EXCEPTION when orphan rows are detected before backfill")
	}
}
