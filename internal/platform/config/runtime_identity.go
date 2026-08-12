package config

import (
	"errors"
	"os"
)

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

// LoadRuntimeIdentityConfig reads METALDOCS_RUNTIME_DB_PASSWORD and fails if
// it is unset. No substitute value, no environment sniffing to decide when a
// default is "safe" (PR #110 review, CodeRabbit): this feeds
// apps/dbprovision's unconditional rotateRuntimePassword call, which sets
// metaldocs_runtime's REAL login password on every run. db/grants/
// 0000_identity_roles.sql deliberately creates that role with no password so
// it fails closed until rotated (see that file's header); a silent fallback
// here to a literal committed in this repository -- published, since this
// repo went public 2026-08-07 -- would rotate the role straight past that
// protection and back to a value any reader of the repo already has, with
// provisioning exiting 0 and nothing logged. The caller decides what "unset"
// means for its environment; this function does not guess.
//
// Local/dev flows all already set this var explicitly and do not rely on a
// default here: deploy/compose/docker-compose.yml's db-provision service
// resolves ${METALDOCS_RUNTIME_DB_PASSWORD:-metaldocs_runtime_dev} at the
// compose layer, so the variable reaching this process is always set;
// scripts/start-api.ps1 loads the developer's .env (which
// .env.example ships with this var already set) into the process
// environment before invoking metaldocs-dbprovision.exe;
// provision_integration_test.go's runProvisionAgainst sets it itself before
// calling run(). None of them depend on this function ever substituting a
// value.
func LoadRuntimeIdentityConfig() (RuntimeIdentityConfig, error) {
	pw := os.Getenv("METALDOCS_RUNTIME_DB_PASSWORD")
	if pw == "" {
		return RuntimeIdentityConfig{}, errors.New(
			"METALDOCS_RUNTIME_DB_PASSWORD is not set: refusing to provision " +
				"metaldocs_runtime with no password rather than substitute a " +
				"value; set it explicitly (see .env.example)")
	}
	return RuntimeIdentityConfig{RuntimePassword: pw}, nil
}
