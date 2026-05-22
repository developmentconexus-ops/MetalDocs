package authn

import (
	"strings"
	"testing"
)

func clearAuthEnv(t *testing.T) {
	t.Helper()
	vars := []string{
		"APP_ENV",
		"METALDOCS_AUTH_ENABLED",
		"METALDOCS_AUTH_SESSION_SECRET",
		"METALDOCS_AUTH_SESSION_COOKIE_NAME",
		"METALDOCS_AUTH_SESSION_TTL_HOURS",
		"METALDOCS_AUTH_PASSWORD_MIN_LENGTH",
		"METALDOCS_AUTH_LOGIN_MAX_FAILED_ATTEMPTS",
		"METALDOCS_AUTH_LOGIN_LOCK_MINUTES",
		"METALDOCS_BOOTSTRAP_ADMIN_ENABLED",
		"METALDOCS_BOOTSTRAP_ADMIN_PASSWORD",
		"METALDOCS_TRUSTED_PROXY_CIDRS",
	}
	for _, v := range vars {
		t.Setenv(v, "")
	}
}

func TestLoadRuntimeConfig_AuthDisabledRequiresLocalAppEnv(t *testing.T) {
	cases := []struct {
		name      string
		appEnv    string
		setAppEnv bool
		wantErr   bool
	}{
		{name: "local permits auth disabled", appEnv: "local", setAppEnv: true, wantErr: false},
		{name: "production rejects auth disabled", appEnv: "production", setAppEnv: true, wantErr: true},
		{name: "unset rejects auth disabled", setAppEnv: false, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			clearAuthEnv(t)
			if tc.setAppEnv {
				t.Setenv("APP_ENV", tc.appEnv)
			}
			t.Setenv("METALDOCS_AUTH_ENABLED", "false")
			t.Setenv("METALDOCS_BOOTSTRAP_ADMIN_ENABLED", "false")

			_, err := LoadRuntimeConfig()
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !strings.Contains(err.Error(), "METALDOCS_AUTH_ENABLED=false only permitted when APP_ENV=local") {
					t.Fatalf("unexpected error message: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestLoadRuntimeConfig_AuthEnabledIgnoresAppEnv(t *testing.T) {
	cases := []struct {
		name      string
		appEnv    string
		setAppEnv bool
	}{
		{name: "local", appEnv: "local", setAppEnv: true},
		{name: "production", appEnv: "production", setAppEnv: true},
		{name: "unset", setAppEnv: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			clearAuthEnv(t)
			if tc.setAppEnv {
				t.Setenv("APP_ENV", tc.appEnv)
			}
			t.Setenv("METALDOCS_AUTH_ENABLED", "true")
			t.Setenv("METALDOCS_AUTH_SESSION_SECRET", "test-secret-value")
			t.Setenv("METALDOCS_BOOTSTRAP_ADMIN_ENABLED", "false")

			if _, err := LoadRuntimeConfig(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
