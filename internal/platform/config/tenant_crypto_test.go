package config_test

import (
	"encoding/base64"
	"testing"

	"metaldocs/internal/platform/config"
)

func TestLoadTenantKEK_UnsetReturnsNotConfigured(t *testing.T) {
	t.Setenv("METALDOCS_TENANT_KEK", "")

	kek, configured, err := config.LoadTenantKEK()
	if err != nil {
		t.Fatalf("LoadTenantKEK() error = %v, want nil", err)
	}
	if configured {
		t.Fatalf("LoadTenantKEK() configured = true, want false when env unset")
	}
	if kek != nil {
		t.Fatalf("LoadTenantKEK() kek = %v, want nil when env unset", kek)
	}
}

func TestLoadTenantKEK_ValidBase64Returns32Bytes(t *testing.T) {
	raw := make([]byte, 32)
	for i := range raw {
		raw[i] = byte(i)
	}
	t.Setenv("METALDOCS_TENANT_KEK", base64.StdEncoding.EncodeToString(raw))

	kek, configured, err := config.LoadTenantKEK()
	if err != nil {
		t.Fatalf("LoadTenantKEK() error = %v, want nil", err)
	}
	if !configured {
		t.Fatalf("LoadTenantKEK() configured = false, want true")
	}
	if len(kek) != 32 {
		t.Fatalf("LoadTenantKEK() kek length = %d, want 32", len(kek))
	}
}

func TestLoadTenantKEK_InvalidBase64Errors(t *testing.T) {
	t.Setenv("METALDOCS_TENANT_KEK", "not-valid-base64!!!")

	if _, _, err := config.LoadTenantKEK(); err == nil {
		t.Fatalf("LoadTenantKEK() with invalid base64 = nil error, want failure")
	}
}

func TestLoadTenantKEK_WrongDecodedLengthErrors(t *testing.T) {
	raw := make([]byte, 16) // too short
	t.Setenv("METALDOCS_TENANT_KEK", base64.StdEncoding.EncodeToString(raw))

	if _, _, err := config.LoadTenantKEK(); err == nil {
		t.Fatalf("LoadTenantKEK() with a 16-byte decoded key = nil error, want failure")
	}
}
