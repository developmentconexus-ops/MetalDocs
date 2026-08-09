package config

import "testing"

// These tests pin the METALDOCS_DATABASE_URL / DATABASE_URL precedence rule
// documented at the point of decision in LoadPostgresConfig: when both are
// set, METALDOCS_DATABASE_URL wins because it is this repo's project-prefixed
// convention and must agree with tests/integration/testdb.DSN()'s own
// precedence, not silently diverge from it.

func TestLoadPostgresConfigUsesMetaldocsDatabaseURLWhenOnlyItIsSet(t *testing.T) {
	t.Setenv("METALDOCS_DATABASE_URL", "postgres://user:pass@metaldocs-host:5432/metaldocsdb")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("PGHOST", "")
	t.Setenv("PGPORT", "")
	t.Setenv("PGDATABASE", "")
	t.Setenv("PGUSER", "")
	t.Setenv("PGPASSWORD", "")

	cfg, err := LoadPostgresConfig()
	if err != nil {
		t.Fatalf("LoadPostgresConfig error: %v", err)
	}
	if cfg.DSN != "postgres://user:pass@metaldocs-host:5432/metaldocsdb" {
		t.Fatalf("DSN = %q, want the METALDOCS_DATABASE_URL value", cfg.DSN)
	}
}

func TestLoadPostgresConfigUsesDatabaseURLWhenOnlyItIsSet(t *testing.T) {
	t.Setenv("METALDOCS_DATABASE_URL", "")
	t.Setenv("DATABASE_URL", "postgres://user:pass@generic-host:5432/generic")
	t.Setenv("PGHOST", "")
	t.Setenv("PGPORT", "")
	t.Setenv("PGDATABASE", "")
	t.Setenv("PGUSER", "")
	t.Setenv("PGPASSWORD", "")

	cfg, err := LoadPostgresConfig()
	if err != nil {
		t.Fatalf("LoadPostgresConfig error: %v", err)
	}
	if cfg.DSN != "postgres://user:pass@generic-host:5432/generic" {
		t.Fatalf("DSN = %q, want the DATABASE_URL value", cfg.DSN)
	}
}

func TestLoadPostgresConfigPrefersMetaldocsDatabaseURLWhenBothSet(t *testing.T) {
	t.Setenv("METALDOCS_DATABASE_URL", "postgres://user:pass@metaldocs-host:5432/metaldocsdb")
	t.Setenv("DATABASE_URL", "postgres://user:pass@generic-host:5432/generic")
	t.Setenv("PGHOST", "")
	t.Setenv("PGPORT", "")
	t.Setenv("PGDATABASE", "")
	t.Setenv("PGUSER", "")
	t.Setenv("PGPASSWORD", "")

	cfg, err := LoadPostgresConfig()
	if err != nil {
		t.Fatalf("LoadPostgresConfig error: %v", err)
	}
	if cfg.DSN != "postgres://user:pass@metaldocs-host:5432/metaldocsdb" {
		t.Fatalf("DSN = %q, want METALDOCS_DATABASE_URL to win when both are set", cfg.DSN)
	}
}
