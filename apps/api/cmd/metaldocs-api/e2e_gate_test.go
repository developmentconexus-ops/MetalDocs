package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Asserts that with METALDOCS_E2E unset the e2etest registrar is not invoked
// and the corresponding test-only paths are absent from the mux. Regression
// guard for the C1 finding: permissions.go does not enumerate /internal/test/*
// routes, so any unconditional mount would be treated as fully public by
// newPublicPathChecker, bypassing both authMiddleware and iamMiddleware.
func TestE2EHandlersGate_EnvUnset_DoesNotRegister(t *testing.T) {
	t.Setenv("METALDOCS_E2E", "")
	if e2eHandlersEnabled() {
		t.Fatal("e2eHandlersEnabled returned true for empty env")
	}

	mux := http.NewServeMux()
	called := false
	mountE2EHandlersIfEnabled(mux, func(*http.ServeMux) { called = true })
	if called {
		t.Fatal("register callback invoked with METALDOCS_E2E unset")
	}

	paths := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/internal/test/seed"},
		{http.MethodPost, "/internal/test/reset"},
		{http.MethodGet, "/internal/test/governance-events"},
		{http.MethodPost, "/internal/test/advance-clock"},
		{http.MethodPost, "/internal/test/trigger-scheduler-tick"},
	}
	for _, p := range paths {
		req := httptest.NewRequest(p.method, p.path, nil)
		if _, pattern := mux.Handler(req); pattern != "" {
			t.Fatalf("%s %s registered (pattern=%q); expected absent", p.method, p.path, pattern)
		}
	}
}

func TestE2EHandlersGate_EnvSet_CallsRegistrar(t *testing.T) {
	t.Setenv("METALDOCS_E2E", "1")
	if !e2eHandlersEnabled() {
		t.Fatal("e2eHandlersEnabled returned false for METALDOCS_E2E=1")
	}
	mux := http.NewServeMux()
	called := false
	mountE2EHandlersIfEnabled(mux, func(*http.ServeMux) { called = true })
	if !called {
		t.Fatal("register callback not invoked when METALDOCS_E2E=1")
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
