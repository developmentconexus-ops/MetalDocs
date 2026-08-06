package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Asserts that with METALDOCS_E2E unset the e2e publisher (when compiled in)
// is never mounted and the corresponding test-only paths are absent from the
// mux. Regression guard for the C1 finding: permissions.go does not
// enumerate /internal/test/* routes, so any unconditional mount would be
// treated as fully public by newPublicPathChecker, bypassing both
// authMiddleware and iamMiddleware.
//
// Task 15b replaced mountE2EHandlersIfEnabled's callback indirection with
// main.go's inline `useE2E := e2e != nil && e2eHandlersEnabled()`, mirrored
// here directly rather than through a dedicated helper function.
func TestE2EHandlersGate_EnvUnset_DoesNotRegister(t *testing.T) {
	t.Setenv("METALDOCS_E2E", "")
	if e2eHandlersEnabled() {
		t.Fatal("e2eHandlersEnabled returned true for empty env")
	}

	mux := http.NewServeMux()
	e2e := e2ePublisher(nil)
	if e2e != nil && e2eHandlersEnabled() { // mirrors main.go's useE2E
		e2e.Mount(mux)
	}

	paths := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/internal/test/seed"},
		{http.MethodPost, "/internal/test/reset"},
		{http.MethodGet, "/internal/test/governance-events"},
		{http.MethodPost, "/internal/test/trigger-scheduler-tick"},
	}
	for _, p := range paths {
		req := httptest.NewRequest(p.method, p.path, nil)
		if _, pattern := mux.Handler(req); pattern != "" {
			t.Fatalf("%s %s registered (pattern=%q); expected absent", p.method, p.path, pattern)
		}
	}
}

// TestE2EHandlersGate_EnvSet_Mounts proves the complement: when
// METALDOCS_E2E=1 AND the e2e handlers are compiled into this build
// (`-tags integration`), the publisher mounts its routes. In a build without
// the handlers (e2ePublisher returns nil, see httpsurface_e2e_publisher_stub.go)
// there is nothing to mount regardless of the env var, so the test skips
// rather than asserting a false positive.
func TestE2EHandlersGate_EnvSet_Mounts(t *testing.T) {
	t.Setenv("METALDOCS_E2E", "1")
	if !e2eHandlersEnabled() {
		t.Fatal("e2eHandlersEnabled returned false for METALDOCS_E2E=1")
	}

	e2e := e2ePublisher(nil)
	if e2e == nil {
		t.Skip("e2e handlers not compiled into this build (no -tags integration)")
	}

	mux := http.NewServeMux()
	if e2e != nil && e2eHandlersEnabled() { // mirrors main.go's useE2E
		e2e.Mount(mux)
	}
	req := httptest.NewRequest(http.MethodPost, "/internal/test/seed", nil)
	if _, pattern := mux.Handler(req); pattern == "" {
		t.Fatal("expected /internal/test/seed registered when e2e enabled and compiled in")
	}
}

func TestE2EHandlersGate_WhitespaceTrimmed(t *testing.T) {
	t.Setenv("METALDOCS_E2E", "  1  ")
	if !e2eHandlersEnabled() {
		t.Fatal("e2eHandlersEnabled did not trim whitespace around \"1\"")
	}
}

// Only the literal "1" enables the gate. Truthy lookalikes ("true", "yes", "0")
// must not enable test-only routes; mirrors internal/test/e2e_seed.go.
func TestE2EHandlersGate_OnlyOneEnables(t *testing.T) {
	cases := []string{"0", "true", "TRUE", "yes", "on", "1.0", " ", "false"}
	for _, v := range cases {
		v := v
		t.Run(v, func(t *testing.T) {
			t.Setenv("METALDOCS_E2E", v)
			if e2eHandlersEnabled() {
				t.Fatalf("METALDOCS_E2E=%q treated as enabled; only \"1\" must enable", v)
			}
		})
	}
}
