//go:build integration

package bootstrap_test

import (
	"context"
	"net/http"
	"testing"

	"metaldocs/internal/platform/bootstrap"
	"metaldocs/internal/platform/config"
	"metaldocs/internal/platform/db/postgres"
	"metaldocs/internal/platform/observability"
	"metaldocs/tests/integration/testdb"
)

// TestBuildAPIDependencies_RefusesUnsafeDBIdentity is the A6.1 boot-fatal RED
// proof at the actual composition root: BuildAPIDependencies must return a
// non-nil error -- and construct nothing usable -- when the ambient DSN
// (METALDOCS_DATABASE_URL/DATABASE_URL, the same env var precedence
// config.LoadPostgresConfig and testdb.DSN both read) resolves to a
// SUPERUSER or BYPASSRLS identity. Before A6.1, this call quietly succeeded
// under the dev compose's superuser PGUSER (deploy/compose/docker-compose.yml
// passes ${POSTGRES_USER}), which is exactly issue #88's finding.
func TestBuildAPIDependencies_RefusesUnsafeDBIdentity(t *testing.T) {
	db, _ := testdb.Open(t)
	status, err := postgres.QueryIdentityStatus(context.Background(), db)
	if err != nil {
		t.Fatalf("QueryIdentityStatus: %v", err)
	}
	if !status.Unsafe() {
		t.Skipf("ambient test identity %q is already safe (rolsuper=%t rolbypassrls=%t) -- this environment does not reproduce the A6.1 vulnerability under test", status.RoleName, status.Superuser, status.BypassRLS)
	}

	deps, err := bootstrap.BuildAPIDependencies(context.Background(), config.RepositoryPostgres, config.AttachmentsConfig{Provider: config.StorageProviderMemory})
	if err == nil {
		deps.Cleanup()
		t.Fatalf("BuildAPIDependencies() succeeded under unsafe identity %q (rolsuper=%t rolbypassrls=%t) -- must refuse to boot (A6.1)", status.RoleName, status.Superuser, status.BypassRLS)
	}
}

// TestDBIdentityHealthCheck_ReadinessFailsClosedOnUnsafeIdentity proves the
// second half of A6.1's acceptance bar: "readiness impossible on an unsafe DB
// identity". BuildAPIDependencies already refuses to construct a
// StatusProvider under an unsafe identity, so this test drives
// DBIdentityHealthCheck directly (the same DependencyCheck BuildAPIDependencies
// wires in) to prove Ready() reports not-ready -- never 200/"ready", and
// never merely "degraded while still serving" -- for the identity that
// matters after boot: one that changed post-construction.
func TestDBIdentityHealthCheck_ReadinessFailsClosedOnUnsafeIdentity(t *testing.T) {
	db, _ := testdb.Open(t)
	status, err := postgres.QueryIdentityStatus(context.Background(), db)
	if err != nil {
		t.Fatalf("QueryIdentityStatus: %v", err)
	}
	if !status.Unsafe() {
		t.Skipf("ambient test identity %q is already safe (rolsuper=%t rolbypassrls=%t) -- this environment does not reproduce the A6.1 vulnerability under test", status.RoleName, status.Superuser, status.BypassRLS)
	}

	provider := observability.NewPostgresRuntimeStatusProvider(db, config.RepositoryPostgres, string(config.StorageProviderMemory), false, bootstrap.DBIdentityHealthCheck(db))
	code, payload := provider.Ready(context.Background())
	if code == http.StatusOK {
		t.Fatalf("Ready() = %d under unsafe identity %q, want != 200: %#v", code, status.RoleName, payload)
	}
	if payload["status"] == "ready" {
		t.Fatalf("Ready() status = %q under unsafe identity %q, want anything but \"ready\": %#v", payload["status"], status.RoleName, payload)
	}
}

// TestDBIdentityHealthCheck_ReadyUnderSafeRuntimeRole is the readiness-side
// GREEN proof: a genuinely safe identity (metaldocs_runtime) must report
// ready.
func TestDBIdentityHealthCheck_ReadyUnderSafeRuntimeRole(t *testing.T) {
	_, dbName := testdb.Open(t)
	db := testdb.OpenAsRuntimeRole(t, dbName)

	provider := observability.NewPostgresRuntimeStatusProvider(db, config.RepositoryPostgres, string(config.StorageProviderMemory), false, bootstrap.DBIdentityHealthCheck(db))
	code, payload := provider.Ready(context.Background())
	if code != http.StatusOK {
		t.Fatalf("Ready() = %d under safe identity metaldocs_runtime, want 200: %#v", code, payload)
	}
	if payload["status"] != "ready" {
		t.Fatalf("Ready() status = %q under safe identity metaldocs_runtime, want \"ready\": %#v", payload["status"], payload)
	}
}
