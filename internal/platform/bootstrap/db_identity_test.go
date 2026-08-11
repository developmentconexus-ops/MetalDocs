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

// TestDBIdentityHealthCheck_ReadinessFlipsLiveOnRoleAttributeChange is the
// dynamic counterpart to the three static tests above. Those prove
// DBIdentityHealthCheck distinguishes an unsafe identity from a safe one --
// two DIFFERENT identities. They do not prove the "no caching" property
// api.go's DBIdentityHealthCheck doc comment asserts: that Ready() re-queries
// pg_roles on every call, so a role attribute changed after boot is caught
// without a restart. This test drives that live: it flips metaldocs_runtime
// SUPERUSER mid-test with ALTER ROLE (a real, not simulated, attribute
// change) on the SAME already-constructed provider/check, and asserts
// Ready() flips from ready to not-ready in response -- the only way that
// transition can happen is a fresh QueryIdentityStatus call per Ready()
// invocation, since nothing else about the check or its inputs changed.
func TestDBIdentityHealthCheck_ReadinessFlipsLiveOnRoleAttributeChange(t *testing.T) {
	adminDB, dbName := testdb.Open(t)
	runtimeDB := testdb.OpenAsRuntimeRole(t, dbName)

	// Cluster-global: roles are not per-database, so this MUST be reverted
	// even though runtimeDB/adminDB themselves are per-test-database handles
	// -- otherwise this flip would leak into every other test (in this
	// package or any other) that opens a connection as metaldocs_runtime
	// after this one runs, silently making AssertSafeIdentity's "safe
	// identity" fixture unsafe cluster-wide.
	t.Cleanup(func() {
		if _, err := adminDB.ExecContext(context.Background(), "ALTER ROLE metaldocs_runtime NOSUPERUSER"); err != nil {
			t.Errorf("cleanup: revert metaldocs_runtime to NOSUPERUSER: %v", err)
		}
	})

	provider := observability.NewPostgresRuntimeStatusProvider(runtimeDB, config.RepositoryPostgres, string(config.StorageProviderMemory), false, bootstrap.DBIdentityHealthCheck(runtimeDB))

	// Baseline: metaldocs_runtime starts safe, so readiness must be green
	// before the flip -- otherwise a later "not ready" assertion would prove
	// nothing about live re-querying (it could just always report unready).
	code, payload := provider.Ready(context.Background())
	if code != http.StatusOK {
		t.Fatalf("Ready() before flip = %d, want 200 (baseline safe identity): %#v", code, payload)
	}
	if payload["status"] != "ready" {
		t.Fatalf("Ready() status before flip = %q, want \"ready\": %#v", payload["status"], payload)
	}

	// Live mid-test attribute change -- the exact scenario DBIdentityHealthCheck's
	// doc comment names: "a role attribute changed (e.g. ALTER ROLE ...
	// SUPERUSER) while the process kept running".
	if _, err := adminDB.ExecContext(context.Background(), "ALTER ROLE metaldocs_runtime SUPERUSER"); err != nil {
		t.Fatalf("ALTER ROLE metaldocs_runtime SUPERUSER: %v", err)
	}

	code, payload = provider.Ready(context.Background())
	if code == http.StatusOK {
		t.Fatalf("Ready() after live SUPERUSER flip = %d, want != 200 -- DBIdentityHealthCheck must re-query, not cache, pg_roles: %#v", code, payload)
	}
	if payload["status"] == "ready" {
		t.Fatalf("Ready() status after live SUPERUSER flip = %q, want anything but \"ready\": %#v", payload["status"], payload)
	}
}
