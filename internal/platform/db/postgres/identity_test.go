//go:build integration
// +build integration

package postgres_test

import (
	"context"
	"testing"

	"metaldocs/internal/platform/db/postgres"
	"metaldocs/tests/integration/testdb"
)

// TestAssertSafeIdentity_RefusesAmbientUnsafeIdentity is the A6.1 RED proof.
// The ambient test DSN (METALDOCS_DATABASE_URL/DATABASE_URL) is, in every
// environment this repo currently ships, the Postgres image's bootstrap
// SUPERUSER (deploy/compose/docker-compose.yml passes ${POSTGRES_USER} as
// PGUSER for api/worker/jobs and as the admin identity testdb itself connects
// with). AssertSafeIdentity must refuse it.
func TestAssertSafeIdentity_RefusesAmbientUnsafeIdentity(t *testing.T) {
	db, _ := testdb.Open(t)

	status, err := postgres.QueryIdentityStatus(context.Background(), db)
	if err != nil {
		t.Fatalf("QueryIdentityStatus: %v", err)
	}
	if !status.Unsafe() {
		t.Skipf("ambient test identity %q is already safe (rolsuper=%t rolbypassrls=%t) -- this environment does not reproduce the A6.1 vulnerability under test", status.RoleName, status.Superuser, status.BypassRLS)
	}

	if err := postgres.AssertSafeIdentity(context.Background(), db); err == nil {
		t.Fatalf("AssertSafeIdentity(%q) = nil, want a refusal: rolsuper=%t rolbypassrls=%t defeats RLS and REVOKE-based hardening", status.RoleName, status.Superuser, status.BypassRLS)
	}
}

// TestAssertSafeIdentity_AcceptsDedicatedRuntimeRole is the GREEN proof: a
// genuinely safe identity (metaldocs_runtime, created by A6.1's
// db/grants/0000_identity_roles.sql and granted privileges by
// db/grants/0001_role_grants.sql) must pass.
func TestAssertSafeIdentity_AcceptsDedicatedRuntimeRole(t *testing.T) {
	_, dbName := testdb.Open(t)
	db := testdb.OpenAsRuntimeRole(t, dbName)

	status, err := postgres.QueryIdentityStatus(context.Background(), db)
	if err != nil {
		t.Fatalf("QueryIdentityStatus: %v", err)
	}
	if status.Unsafe() {
		t.Fatalf("metaldocs_runtime is unsafe (rolsuper=%t rolbypassrls=%t) -- grants provisioning regressed", status.Superuser, status.BypassRLS)
	}

	if err := postgres.AssertSafeIdentity(context.Background(), db); err != nil {
		t.Fatalf("AssertSafeIdentity(metaldocs_runtime) = %v, want nil", err)
	}
}
