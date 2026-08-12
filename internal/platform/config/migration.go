package config

import (
	"os"
	"strings"
)

// MigrationConfig holds startup-migration settings.
type MigrationConfig struct {
	Skip bool
	Dir  string
	// GrantsDir is the bootstrap grants stage (db/grants), which as of A6.1
	// contains both 0000_identity_roles.sql (role creation + ownership
	// transfer) and 0001_role_grants.sql (privilege grants). Applied only by
	// apps/dbprovision/cmd/metaldocs-dbprovision now, not at every
	// metaldocs-api startup -- see internal/platform/migrate.ApplyGrants.
	GrantsDir string
	// PrerequisitesDir is the superuser-only extensions stage (db/prerequisites,
	// e.g. CREATE EXTENSION pgcrypto/pg_trgm). Applied only by
	// metaldocs-dbprovision, connected as the bootstrap superuser, before
	// GrantsDir and Dir -- extensions require superuser and nothing else in
	// the provisioning sequence does.
	PrerequisitesDir string
}

// LoadMigrationConfig reads migration configuration from the environment.
// METALDOCS_SKIP_STARTUP_MIGRATIONS == "true" (case-insensitive, trimmed) skips migrations.
// METALDOCS_MIGRATIONS_DIR sets the directory; defaults to "db/migrations".
// METALDOCS_GRANTS_DIR sets the grants-stage directory; defaults to "db/grants".
// METALDOCS_PREREQUISITES_DIR sets the extensions-stage directory; defaults to "db/prerequisites".
func LoadMigrationConfig() (MigrationConfig, error) {
	skip := strings.EqualFold(strings.TrimSpace(os.Getenv("METALDOCS_SKIP_STARTUP_MIGRATIONS")), "true")
	dir := strings.TrimSpace(os.Getenv("METALDOCS_MIGRATIONS_DIR"))
	if dir == "" {
		dir = "db/migrations"
	}
	grantsDir := strings.TrimSpace(os.Getenv("METALDOCS_GRANTS_DIR"))
	if grantsDir == "" {
		grantsDir = "db/grants"
	}
	prerequisitesDir := strings.TrimSpace(os.Getenv("METALDOCS_PREREQUISITES_DIR"))
	if prerequisitesDir == "" {
		prerequisitesDir = "db/prerequisites"
	}
	return MigrationConfig{Skip: skip, Dir: dir, GrantsDir: grantsDir, PrerequisitesDir: prerequisitesDir}, nil
}
