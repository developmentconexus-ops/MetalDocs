//go:build integration

package bootstrap_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/platform/bootstrap"
	"metaldocs/internal/platform/config"
	"metaldocs/tests/integration/testdb"
)

// buildTestDeps is a helper that calls BuildAPIDependencies with a real postgres
// test DB (skipped when DATABASE_URL is not set). Memory mode was removed in 2.13.
//
// BuildAPIDependencies now boot-fatals under an unsafe DB identity (A6.1,
// postgres.AssertSafeIdentity) and the ambient DSN these tests would
// otherwise inherit is superuser + BYPASSRLS in dev -- exactly the posture
// the gate exists to refuse. Point it at the dedicated metaldocs_runtime
// role instead so these tests keep exercising Gotenberg-check wiring rather
// than the identity gate itself (that gate has its own coverage in
// db_identity_test.go).
func buildTestDeps(t *testing.T, attachmentsCfg config.AttachmentsConfig) bootstrap.APIDependencies {
	t.Helper()
	t.Setenv("METALDOCS_DATABASE_URL", testdb.RuntimeRoleDSN(t))
	deps, err := bootstrap.BuildAPIDependencies(context.Background(), config.RepositoryPostgres, attachmentsCfg)
	if err != nil {
		t.Fatalf("BuildAPIDependencies() error = %v", err)
	}
	t.Cleanup(deps.Cleanup)
	return deps
}

func TestBuildAPIDependenciesIncludesGotenbergCheckWhenHealthy(t *testing.T) {
	t.Setenv("METALDOCS_GOTENBERG_URL", "")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Setenv("METALDOCS_GOTENBERG_URL", server.URL)

	deps := buildTestDeps(t, config.AttachmentsConfig{Provider: config.StorageProviderMemory})

	statusCode, payload := deps.StatusProvider.Ready(context.Background())
	if statusCode != http.StatusOK {
		t.Fatalf("Ready() statusCode = %d, want %d", statusCode, http.StatusOK)
	}

	check := findCheck(t, payload, "gotenberg")
	if got := check["status"]; got != "up" {
		t.Fatalf("gotenberg status = %v, want up", got)
	}
	if got := check["detail"]; got != server.URL {
		t.Fatalf("gotenberg detail = %v, want %q", got, server.URL)
	}
}

func TestBuildAPIDependenciesMarksGotenbergSkippedWhenNotConfigured(t *testing.T) {
	t.Setenv("METALDOCS_GOTENBERG_URL", "")

	deps := buildTestDeps(t, config.AttachmentsConfig{Provider: config.StorageProviderMemory})

	statusCode, payload := deps.StatusProvider.Ready(context.Background())
	if statusCode != http.StatusOK {
		t.Fatalf("Ready() statusCode = %d, want %d", statusCode, http.StatusOK)
	}

	check := findCheck(t, payload, "gotenberg")
	if got := check["status"]; got != "skipped" {
		t.Fatalf("gotenberg status = %v, want skipped", got)
	}
	if got := check["detail"]; got != "gotenberg not configured" {
		t.Fatalf("gotenberg detail = %v, want %q", got, "gotenberg not configured")
	}
}

func TestBuildAPIDependenciesMarksGotenbergDownWhenHealthCheckFails(t *testing.T) {
	t.Setenv("METALDOCS_GOTENBERG_URL", "")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	t.Setenv("METALDOCS_GOTENBERG_URL", server.URL)

	deps := buildTestDeps(t, config.AttachmentsConfig{Provider: config.StorageProviderMemory})

	statusCode, payload := deps.StatusProvider.Ready(context.Background())
	if statusCode != http.StatusServiceUnavailable {
		t.Fatalf("Ready() statusCode = %d, want %d", statusCode, http.StatusServiceUnavailable)
	}
	if got := payload["status"]; got != "degraded" {
		t.Fatalf("Ready() status = %v, want degraded", got)
	}

	check := findCheck(t, payload, "gotenberg")
	if got := check["status"]; got != "down" {
		t.Fatalf("gotenberg status = %v, want down", got)
	}
	if got := check["detail"]; got != "runtime: gotenberg unhealthy: status 503" {
		t.Fatalf("gotenberg detail = %v, want %q", got, "runtime: gotenberg unhealthy: status 503")
	}
}

func TestBuildMinioClientsFailsWhenPublicEndpointIsInvalid(t *testing.T) {
	_, _, _, err := bootstrap.BuildMinioClients(config.AttachmentsConfig{
		Provider:            config.StorageProviderMinIO,
		MinIOEndpoint:       "localhost:9000",
		MinIOPublicEndpoint: "://bad endpoint",
		MinIOAccessKey:      "minioadmin",
		MinIOSecretKey:      "minioadmin",
		MinIOBucket:         "metaldocs",
	})
	if err == nil {
		t.Fatal("expected public endpoint init error")
	}
}

func findCheck(t *testing.T, payload map[string]any, name string) map[string]any {
	t.Helper()

	checks, ok := payload["checks"].([]map[string]any)
	if !ok {
		t.Fatalf("checks payload type = %T, want []map[string]any", payload["checks"])
	}
	for _, check := range checks {
		if check["name"] == name {
			return check
		}
	}
	t.Fatalf("check %q not found in %#v", name, checks)
	return nil
}
