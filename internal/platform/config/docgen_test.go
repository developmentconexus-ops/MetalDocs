package config

import "testing"

func TestLoadDocgenConfigDefaultsToLocalDocgenEndpoint(t *testing.T) {
	t.Setenv("APP_ENV", "local")
	t.Setenv("METALDOCS_DOCGEN_API_URL", "")

	cfg, err := LoadDocgenConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !cfg.Enabled {
		t.Fatalf("expected enabled config in local env when URL is omitted")
	}
	if cfg.APIURL != "http://127.0.0.1:3001" {
		t.Fatalf("api url = %q, want %q", cfg.APIURL, "http://127.0.0.1:3001")
	}
}

func TestLoadDocgenConfigDisabledOutsideLocalWhenURLMissing(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("METALDOCS_DOCGEN_API_URL", "")

	cfg, err := LoadDocgenConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Enabled {
		t.Fatalf("expected disabled config when URL is missing outside local env")
	}
	if cfg.APIURL != "" {
		t.Fatalf("api url = %q, want empty", cfg.APIURL)
	}
}

func TestLoadDocgenConfigDefaultsTimeoutWhenConfiguredValueIsZeroOrNegative(t *testing.T) {
	t.Setenv("APP_ENV", "local")
	t.Setenv("METALDOCS_DOCGEN_API_URL", "http://127.0.0.1:3001")

	for _, raw := range []string{"0", "-5"} {
		t.Setenv("METALDOCS_DOCGEN_REQUEST_TIMEOUT_SECONDS", raw)

		cfg, err := LoadDocgenConfig()
		if err != nil {
			t.Fatalf("raw=%s expected no error, got %v", raw, err)
		}
		if cfg.RequestTimeoutSeconds != 30 {
			t.Fatalf("raw=%s timeout=%d, want 30", raw, cfg.RequestTimeoutSeconds)
		}
	}
}
