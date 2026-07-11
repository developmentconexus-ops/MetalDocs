//go:build integration

package migrations_test

import (
	"context"
	"os"
	"testing"

	"metaldocs/tests/integration/testdb"
)

func TestMigration0169_FileExists(t *testing.T) {
	if _, err := os.Stat("../../../archive/migrations/0169_role_capabilities_process_areas.sql"); err != nil {
		t.Fatalf("expected migration file 0169_role_capabilities_process_areas.sql: %v", err)
	}
}

func TestMigration0169_ProcessAreaRoleCapabilitiesSeeded(t *testing.T) {
	ctx := context.Background()
	db, schema := testdb.Open(t)

	// doc.* was renamed to document.* (confirmed against
	// db/reference-data/0001_product_reference_data.sql, the live
	// role_capabilities seed source of truth: e.g. line 124-125 seeds
	// ('signer', 'document.signoff') / ('signer', 'document.view'), not doc.*).
	checks := []struct {
		role       string
		capability string
	}{
		{"signer", "document.view"},
		{"signer", "document.signoff"},
		{"area_admin", "document.view"},
		{"area_admin", "document.create"},
		{"area_admin", "document.edit"},
		{"area_admin", "document.submit"},
		{"area_admin", "document.signoff"},
		{"area_admin", "membership.manage"},
		{"qms_admin", "document.view"},
		{"qms_admin", "document.create"},
		{"qms_admin", "document.edit"},
		{"qms_admin", "document.submit"},
		{"qms_admin", "document.signoff"},
		{"qms_admin", "template.view"},
		{"qms_admin", "template.approve"},
		{"qms_admin", "template.publish"},
		{"qms_admin", "route.manage"},
		{"qms_admin", "taxonomy.manage"},
	}

	for _, c := range checks {
		var found bool
		err := db.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT 1
				  FROM `+testdb.Qualified(schema, "role_capabilities")+`
				 WHERE role = $1
				   AND capability = $2
			)`, c.role, c.capability).Scan(&found)
		if err != nil {
			t.Fatalf("query role_capabilities role=%s capability=%s: %v", c.role, c.capability, err)
		}
		if !found {
			t.Fatalf("missing capability %s for role %s", c.capability, c.role)
		}
	}
}
