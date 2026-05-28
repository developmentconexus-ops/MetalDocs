package migrations_test

import (
	"os"
	"strings"
	"testing"
)

func TestMigration0213ConvertsTemplatesTenantColumnsToUUID(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile("0213_templates_tenant_id_uuid.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	sql := string(body)

	if !strings.Contains(sql, "public.templates_template") || !strings.Contains(sql, "ALTER COLUMN tenant_id TYPE UUID") {
		t.Fatal("migration must convert public.templates_template.tenant_id to UUID")
	}
	if !strings.Contains(sql, "public.templates_audit_log") || !strings.Contains(sql, "USING tenant_id::uuid") {
		t.Fatal("migration must convert public.templates_audit_log.tenant_id to UUID using an explicit cast")
	}
	if !strings.Contains(sql, "backfill required for public.templates_template.tenant_id") {
		t.Fatal("migration must fail with an explicit templates_template backfill prerequisite when legacy tenant IDs are not UUID-shaped")
	}
	if !strings.Contains(sql, "backfill required for public.templates_audit_log.tenant_id") {
		t.Fatal("migration must fail with an explicit templates_audit_log backfill prerequisite when legacy tenant IDs are not UUID-shaped")
	}
	if !strings.Contains(sql, "schema_migrations") || !strings.Contains(sql, "'0213'") {
		t.Fatal("migration must record schema_migrations version 0213")
	}
}
