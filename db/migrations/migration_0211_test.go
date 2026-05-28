package migrations_test

import (
	"os"
	"strings"
	"testing"
)

func TestMigration0211BackfillsThenTightensTenantInvariant(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile("0211_editor_sessions_tenant_id.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	sql := string(body)

	if strings.Contains(sql, "ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL DEFAULT 'ffffffff-ffff-ffff-ffff-ffffffffffff'") {
		t.Fatal("migration still uses sentinel default for tenant_id")
	}
	if !strings.Contains(sql, "SET tenant_id = d.tenant_id") || !strings.Contains(sql, "FROM public.documents d") {
		t.Fatal("migration must backfill tenant_id from public.documents")
	}
	if !strings.Contains(sql, "es.tenant_id IS NULL OR es.tenant_id = 'ffffffff-ffff-ffff-ffff-ffffffffffff'") {
		t.Fatal("migration must repair both null and sentinel tenant_id rows")
	}
	if !strings.Contains(sql, "ALTER COLUMN tenant_id DROP DEFAULT") {
		t.Fatal("migration must drop any tenant_id default after backfill")
	}
	if !strings.Contains(sql, "ALTER COLUMN tenant_id SET NOT NULL") {
		t.Fatal("migration must restore tenant_id NOT NULL after backfill")
	}
	if !strings.Contains(sql, "idx_editor_sessions_tenant_id") {
		t.Fatal("migration must keep the tenant_id index")
	}
	if !strings.Contains(sql, "schema_migrations") || !strings.Contains(sql, "'0211'") {
		t.Fatal("migration must record schema_migrations version 0211")
	}
}
