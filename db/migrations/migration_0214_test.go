package migrations_test

import (
	"os"
	"strings"
	"testing"
)

func TestMigration0214AddsTenantsMasterAndConvertsAuthSessionsTenantIDToUUID(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile("0214_tenants_master_table.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	sql := string(body)

	if !strings.Contains(sql, "CREATE TABLE IF NOT EXISTS metaldocs.tenants") {
		t.Fatal("migration must create the canonical metaldocs.tenants table")
	}
	if !strings.Contains(sql, "ALTER TABLE metaldocs.auth_sessions") || !strings.Contains(sql, "ALTER COLUMN tenant_id TYPE UUID") {
		t.Fatal("migration must convert metaldocs.auth_sessions.tenant_id to UUID")
	}
	if !strings.Contains(sql, "backfill required for metaldocs.auth_sessions.tenant_id") {
		t.Fatal("migration must fail with an explicit auth_sessions tenant backfill prerequisite when legacy values are not UUID-shaped")
	}
	if !strings.Contains(sql, "INSERT INTO metaldocs.tenants (id, name, slug)") {
		t.Fatal("migration must seed tenant master rows from existing tenant-scoped auth tables")
	}
	if !strings.Contains(sql, "fk_auth_sessions_tenant") || !strings.Contains(sql, "iam_user_roles_tenant_id_fkey") || !strings.Contains(sql, "iam_users_tenant_id_fkey") {
		t.Fatal("migration must attach tenant foreign keys for auth and IAM ownership tables")
	}
	if !strings.Contains(sql, "schema_migrations") || !strings.Contains(sql, "'0214'") {
		t.Fatal("migration must record schema_migrations version 0214")
	}
}
