//go:build integration

package migrations_test

import (
	"context"
	"os"
	"testing"

	"metaldocs/tests/integration/testdb"
)

func TestMigration0169_FileExists(t *testing.T) {
	if _, err := os.Stat("../../../migrations/0169_role_capabilities_process_areas.sql"); err != nil {
		t.Fatalf("expected migration file 0169_role_capabilities_process_areas.sql: %v", err)
	}
}

func TestMigration0169_ProcessAreaRoleCapabilitiesSeeded(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)

	checks := []struct {
		role       string
		capability string
	}{
		{"signer", "doc.view"},
		{"signer", "doc.signoff"},
		{"area_admin", "doc.view"},
		{"area_admin", "doc.create"},
		{"area_admin", "doc.edit"},
		{"area_admin", "doc.submit"},
		{"area_admin", "doc.signoff"},
		{"area_admin", "membership.manage"},
		{"qms_admin", "doc.view"},
		{"qms_admin", "doc.create"},
		{"qms_admin", "doc.edit"},
		{"qms_admin", "doc.submit"},
		{"qms_admin", "doc.signoff"},
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
				  FROM metaldocs.role_capabilities
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
