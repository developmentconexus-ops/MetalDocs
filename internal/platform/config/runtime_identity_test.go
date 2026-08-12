package config

import "testing"

// These tests pin the fail-closed behavior ratified on PR #110 (issue #88 /
// A6.1, CodeRabbit review): LoadRuntimeIdentityConfig must never substitute a
// value for METALDOCS_RUNTIME_DB_PASSWORD. A silent default here would defeat
// db/grants/0000_identity_roles.sql's fail-closed CREATE ROLE (no password)
// one layer up in the same call chain.

func TestLoadRuntimeIdentityConfigFailsClosedWhenUnset(t *testing.T) {
	t.Setenv("METALDOCS_RUNTIME_DB_PASSWORD", "")

	cfg, err := LoadRuntimeIdentityConfig()
	if err == nil {
		t.Fatalf("LoadRuntimeIdentityConfig err = nil, want an error naming METALDOCS_RUNTIME_DB_PASSWORD; cfg = %+v", cfg)
	}
	if cfg != (RuntimeIdentityConfig{}) {
		t.Fatalf("LoadRuntimeIdentityConfig returned a non-zero config on error: %+v", cfg)
	}
}

func TestLoadRuntimeIdentityConfigUsesExplicitValueWhenSet(t *testing.T) {
	t.Setenv("METALDOCS_RUNTIME_DB_PASSWORD", "a-test-only-value")

	cfg, err := LoadRuntimeIdentityConfig()
	if err != nil {
		t.Fatalf("LoadRuntimeIdentityConfig error: %v", err)
	}
	if cfg.RuntimePassword != "a-test-only-value" {
		t.Fatalf("RuntimePassword = %q, want the explicitly set value", cfg.RuntimePassword)
	}
}
