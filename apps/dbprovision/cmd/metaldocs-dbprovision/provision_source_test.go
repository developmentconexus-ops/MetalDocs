package main

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"
)

// TestIdentityRolesSQLDoesNotSetPasswords proves the fail-closed provisioning
// boundary without inspecting pg_authid. Roles are cluster-global, so a live
// database assertion is contaminated by an earlier provisioning run (or a
// parallel integration package) that may already have rotated the role. The
// property at this stage is instead about the committed CREATE ROLE source:
// no role declaration may bake a credential before the explicit rotation
// stage. This source proof is deterministic and still goes RED if a literal
// PASSWORD clause is reintroduced into the role declarations.
func TestIdentityRolesSQLDoesNotSetPasswords(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate source test")
	}
	root := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(root)
		if parent == root {
			t.Fatal("find repository root")
		}
		root = parent
	}

	sqlBytes, err := os.ReadFile(filepath.Join(root, "db", "grants", "0000_identity_roles.sql"))
	if err != nil {
		t.Fatalf("read identity-role grants: %v", err)
	}

	roleStatements := regexp.MustCompile(`(?im)^\s*EXECUTE\s+'CREATE ROLE (metaldocs_(?:owner|runtime|ci))\b[^;]*';`).FindAllStringSubmatch(string(sqlBytes), -1)
	if len(roleStatements) != 3 {
		t.Fatalf("found %d identity-role CREATE ROLE statements, want 3", len(roleStatements))
	}
	passwordClause := regexp.MustCompile(`(?i)\bPASSWORD\b`)
	for _, statement := range roleStatements {
		if passwordClause.MatchString(statement[0]) {
			t.Errorf("identity role %s has a PASSWORD clause in 0000_identity_roles.sql: %s", statement[1], statement[0])
		}
	}
}
