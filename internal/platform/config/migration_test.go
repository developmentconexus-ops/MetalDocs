package config

import (
	"testing"
)

func TestLoadMigrationConfig(t *testing.T) {
	tests := []struct {
		name     string
		skipEnv  string
		dirEnv   string
		wantSkip bool
		wantDir  string
	}{
		{name: "defaults", skipEnv: "", dirEnv: "", wantSkip: false, wantDir: "db/migrations"},
		{name: "skip true lowercase", skipEnv: "true", wantSkip: true, wantDir: "db/migrations"},
		{name: "skip TRUE uppercase", skipEnv: "TRUE", wantSkip: true, wantDir: "db/migrations"},
		{name: "skip True mixed", skipEnv: "True", wantSkip: true, wantDir: "db/migrations"},
		{name: "skip false", skipEnv: "false", wantSkip: false, wantDir: "db/migrations"},
		{name: "skip 1 is not true", skipEnv: "1", wantSkip: false, wantDir: "db/migrations"},
		{name: "custom dir", skipEnv: "", dirEnv: "custom/migrations", wantSkip: false, wantDir: "custom/migrations"},
		{name: "dir whitespace trimmed", skipEnv: "", dirEnv: "  my/path  ", wantSkip: false, wantDir: "my/path"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("METALDOCS_SKIP_STARTUP_MIGRATIONS", tt.skipEnv)
			t.Setenv("METALDOCS_MIGRATIONS_DIR", tt.dirEnv)
			cfg, err := LoadMigrationConfig()
			if err != nil {
				t.Fatalf("LoadMigrationConfig() unexpected error: %v", err)
			}
			if cfg.Skip != tt.wantSkip {
				t.Errorf("Skip = %v, want %v", cfg.Skip, tt.wantSkip)
			}
			if cfg.Dir != tt.wantDir {
				t.Errorf("Dir = %q, want %q", cfg.Dir, tt.wantDir)
			}
		})
	}
}
