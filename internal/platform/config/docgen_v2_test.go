package config

import "testing"

func TestLoadDocgenV2ConfigDisabledWhenURLMissing(t *testing.T) {
	t.Setenv("METALDOCS_DOCGEN_V2_URL", "")
	t.Setenv("METALDOCS_DOCGEN_V2_SERVICE_TOKEN", "")
	t.Setenv("METALDOCS_DOCGEN_V2_ALLOWED_HOSTS", "")

	cfg, err := LoadDocgenV2Config()
	if err != nil {
		t.Fatalf("LoadDocgenV2Config() error = %v", err)
	}
	if cfg.Enabled {
		t.Fatalf("expected disabled config when URL is missing")
	}
}

func TestLoadDocgenV2ConfigRejectsEmptyServiceToken(t *testing.T) {
	t.Setenv("METALDOCS_DOCGEN_V2_URL", "https://docgen.example.test")
	t.Setenv("METALDOCS_DOCGEN_V2_SERVICE_TOKEN", " ")
	t.Setenv("METALDOCS_DOCGEN_V2_ALLOWED_HOSTS", "docgen.example.test")

	_, err := LoadDocgenV2Config()
	if err == nil {
		t.Fatalf("expected empty service token error")
	}
}

func TestLoadDocgenV2ConfigRejectsNonHTTPURL(t *testing.T) {
	t.Setenv("METALDOCS_DOCGEN_V2_URL", "file:///etc/passwd")
	t.Setenv("METALDOCS_DOCGEN_V2_SERVICE_TOKEN", "token")
	t.Setenv("METALDOCS_DOCGEN_V2_ALLOWED_HOSTS", "docgen.example.test")

	_, err := LoadDocgenV2Config()
	if err == nil {
		t.Fatalf("expected non-http URL error")
	}
}

func TestLoadDocgenV2ConfigRejectsMissingAllowlist(t *testing.T) {
	t.Setenv("METALDOCS_DOCGEN_V2_URL", "https://docgen.example.test")
	t.Setenv("METALDOCS_DOCGEN_V2_SERVICE_TOKEN", "token")
	t.Setenv("METALDOCS_DOCGEN_V2_ALLOWED_HOSTS", "")

	_, err := LoadDocgenV2Config()
	if err == nil {
		t.Fatalf("expected missing allowlist error")
	}
}

func TestLoadDocgenV2ConfigRejectsHostOutsideAllowlist(t *testing.T) {
	t.Setenv("METALDOCS_DOCGEN_V2_URL", "https://metadata.google.internal")
	t.Setenv("METALDOCS_DOCGEN_V2_SERVICE_TOKEN", "token")
	t.Setenv("METALDOCS_DOCGEN_V2_ALLOWED_HOSTS", "docgen.example.test")

	_, err := LoadDocgenV2Config()
	if err == nil {
		t.Fatalf("expected allowlist error")
	}
}

func TestLoadDocgenV2ConfigAcceptsAllowlistedHost(t *testing.T) {
	t.Setenv("METALDOCS_DOCGEN_V2_URL", "https://docgen.example.test:8443")
	t.Setenv("METALDOCS_DOCGEN_V2_SERVICE_TOKEN", " token ")
	t.Setenv("METALDOCS_DOCGEN_V2_ALLOWED_HOSTS", "docgen.example.test")
	t.Setenv("METALDOCS_DOCGEN_V2_REQUEST_TIMEOUT_SECONDS", "15")

	cfg, err := LoadDocgenV2Config()
	if err != nil {
		t.Fatalf("LoadDocgenV2Config() error = %v", err)
	}
	if !cfg.Enabled {
		t.Fatalf("expected enabled config")
	}
	if cfg.APIURL != "https://docgen.example.test:8443" {
		t.Fatalf("APIURL = %q", cfg.APIURL)
	}
	if cfg.ServiceToken != "token" {
		t.Fatalf("ServiceToken = %q", cfg.ServiceToken)
	}
	if cfg.RequestTimeoutSeconds != 15 {
		t.Fatalf("RequestTimeoutSeconds = %d", cfg.RequestTimeoutSeconds)
	}
	if len(cfg.AllowedHosts) != 1 || cfg.AllowedHosts[0] != "docgen.example.test" {
		t.Fatalf("AllowedHosts = %#v", cfg.AllowedHosts)
	}
}
