//go:build integration

package migrations_test

import (
	"os"
	"strings"
	"testing"
)

// TestMigration0255_FileExists verifies the migration file is present.
func TestMigration0255_FileExists(t *testing.T) {
	if _, err := os.Stat("../../../db/migrations/0255_template_version_tenant_consistency_trigger.sql"); err != nil {
		t.Fatalf("expected migration file 0255_template_version_tenant_consistency_trigger.sql: %v", err)
	}
}

// TestMigration0255_FileContent asserts the migration contains the required
// structural elements: the trigger function, its DROP+CREATE binding, the GUC
// read, the parent-JOIN tenant consistency check, and the schema_migrations
// insert.
func TestMigration0255_FileContent(t *testing.T) {
	data, err := os.ReadFile("../../../db/migrations/0255_template_version_tenant_consistency_trigger.sql")
	if err != nil {
		t.Fatalf("read migration 0255: %v", err)
	}
	content := string(data)

	checks := []struct {
		label string
		want  string
	}{
		{
			label: "trigger function CREATE OR REPLACE",
			want:  "CREATE OR REPLACE FUNCTION public.enforce_template_version_tenant_consistent()",
		},
		{
			label: "GUC read for asserting tenant",
			want:  "current_setting('metaldocs.tenant_id', true)",
		},
		{
			label: "parent template JOIN for tenant resolution",
			want:  "FROM public.templates_template",
		},
		{
			label: "cross-tenant check DISTINCT FROM",
			want:  "IS DISTINCT FROM",
		},
		{
			label: "check_violation ERRCODE on cross-tenant write",
			want:  "check_violation",
		},
		{
			label: "DROP TRIGGER IF EXISTS for idempotency",
			want:  "DROP TRIGGER IF EXISTS trg_template_version_tenant_consistent",
		},
		{
			label: "CREATE TRIGGER binding",
			want:  "CREATE TRIGGER trg_template_version_tenant_consistent",
		},
		{
			label: "trigger targets templates_template_version",
			want:  "ON public.templates_template_version",
		},
		{
			label: "BEFORE INSERT OR UPDATE",
			want:  "BEFORE INSERT OR UPDATE",
		},
		{
			label: "schema_migrations insert",
			want:  "INSERT INTO public.schema_migrations",
		},
		{
			label: "migration version 0255",
			want:  "'0255'",
		},
		{
			label: "wrapped in transaction",
			want:  "BEGIN;",
		},
		{
			label: "transaction committed",
			want:  "COMMIT;",
		},
	}

	for _, c := range checks {
		if !strings.Contains(content, c.want) {
			t.Errorf("migration 0255: missing %s — expected to find %q in file", c.label, c.want)
		}
	}
}
