package config

import (
	"os"
	"strings"
)

// MigrationConfig holds startup-migration settings.
type MigrationConfig struct {
	Skip bool
	Dir  string
	// GrantsDir is the bootstrap grants stage (db/grants). Unlike Dir it is
	// ledger-less and re-applied unconditionally at every startup, so an edit
	// to the privilege posture reaches long-lived volumes that never re-run
	// initdb. See internal/platform/migrate.ApplyGrants.
	GrantsDir string
}

// LoadMigrationConfig reads migration configuration from the environment.
// METALDOCS_SKIP_STARTUP_MIGRATIONS == "true" (case-insensitive, trimmed) skips migrations.
// METALDOCS_MIGRATIONS_DIR sets the directory; defaults to "db/migrations".
// METALDOCS_GRANTS_DIR sets the grants-stage directory; defaults to "db/grants".
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
	return MigrationConfig{Skip: skip, Dir: dir, GrantsDir: grantsDir}, nil
}
