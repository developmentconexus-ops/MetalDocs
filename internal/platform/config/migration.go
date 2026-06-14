package config

import (
	"os"
	"strings"
)

// MigrationConfig holds startup-migration settings.
type MigrationConfig struct {
	Skip bool
	Dir  string
}

// LoadMigrationConfig reads migration configuration from the environment.
// METALDOCS_SKIP_STARTUP_MIGRATIONS == "true" (case-insensitive, trimmed) skips migrations.
// METALDOCS_MIGRATIONS_DIR sets the directory; defaults to "db/migrations".
func LoadMigrationConfig() (MigrationConfig, error) {
	skip := strings.EqualFold(strings.TrimSpace(os.Getenv("METALDOCS_SKIP_STARTUP_MIGRATIONS")), "true")
	dir := strings.TrimSpace(os.Getenv("METALDOCS_MIGRATIONS_DIR"))
	if dir == "" {
		dir = "db/migrations"
	}
	return MigrationConfig{Skip: skip, Dir: dir}, nil
}
