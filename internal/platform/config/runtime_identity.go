package config

import "os"

// runtimeDBPasswordDevDefault mirrors the non-secret dev/CI fixture password
// db/grants/0000_identity_roles.sql bakes into `CREATE ROLE metaldocs_runtime
// ... PASSWORD`. It exists so a fresh environment with no override still
// provisions a role and a credential that agree with each other.
const runtimeDBPasswordDevDefault = "metaldocs_runtime_dev"

// RuntimeIdentityConfig carries the single credential source apps/dbprovision
// applies to metaldocs_runtime during provisioning. Every long-running server
// (metaldocs-api/worker/jobs) reads the SAME env var directly as its own
// PGPASSWORD (see deploy/compose/docker-compose.yml), so "the password
// provisioning set on the role" and "the password serving processes
// authenticate with" cannot drift apart into two different sources of truth
// (issue #88 / A6.1 review finding C2).
type RuntimeIdentityConfig struct {
	// RuntimePassword is the metaldocs_runtime login password to provision.
	RuntimePassword string
}

// LoadRuntimeIdentityConfig reads METALDOCS_RUNTIME_DB_PASSWORD, falling back
// to the db/grants/0000_identity_roles.sql dev-fixture default when unset --
// the same precedence tests/integration/testdb.OpenAsRuntimeRole already uses
// to connect as this role.
func LoadRuntimeIdentityConfig() RuntimeIdentityConfig {
	pw := os.Getenv("METALDOCS_RUNTIME_DB_PASSWORD")
	if pw == "" {
		pw = runtimeDBPasswordDevDefault
	}
	return RuntimeIdentityConfig{RuntimePassword: pw}
}
