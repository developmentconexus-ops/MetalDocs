package unit

import (
	"os"
	"testing"

	"metaldocs/internal/platform/config"
)

func TestLoadAttachmentsConfigRequiresSecretForMinIO(t *testing.T) {
	t.Setenv("APP_ENV", "local")
	t.Setenv("METALDOCS_STORAGE_PROVIDER", "minio")
	t.Setenv("METALDOCS_MINIO_ENDPOINT", "minio:9000")
	t.Setenv("METALDOCS_MINIO_ACCESS_KEY", "minioadmin")
	t.Setenv("METALDOCS_MINIO_SECRET_KEY", "secret")
	t.Setenv("METALDOCS_MINIO_BUCKET", "metaldocs-attachments")
	_ = os.Unsetenv("METALDOCS_ATTACHMENTS_SIGNING_SECRET")

	_, err := config.LoadAttachmentsConfig()
	if err == nil {
		t.Fatalf("expected error when minio storage has no signing secret")
	}
}

func TestLoadAttachmentsConfigRequiresSecretForLocalProvider(t *testing.T) {
	t.Setenv("APP_ENV", "local")
	t.Setenv("METALDOCS_STORAGE_PROVIDER", "local")
	_ = os.Unsetenv("METALDOCS_ATTACHMENTS_SIGNING_SECRET")

	_, err := config.LoadAttachmentsConfig()
	if err == nil {
		t.Fatalf("expected error when local storage has no signing secret")
	}
}

func TestLoadAttachmentsConfigRejectsShortSecret(t *testing.T) {
	t.Setenv("APP_ENV", "local")
	t.Setenv("METALDOCS_STORAGE_PROVIDER", "local")
	// 31 bytes — one short of the NIST SP 800-107 / FIPS 198-1 minimum.
	t.Setenv("METALDOCS_ATTACHMENTS_SIGNING_SECRET", "0123456789012345678901234567890")

	_, err := config.LoadAttachmentsConfig()
	if err == nil {
		t.Fatalf("expected error when signing secret is shorter than 32 bytes")
	}
}

func TestLoadAttachmentsConfigAcceptsThirtyTwoByteSecret(t *testing.T) {
	t.Setenv("APP_ENV", "local")
	t.Setenv("METALDOCS_STORAGE_PROVIDER", "local")
	t.Setenv("METALDOCS_ATTACHMENTS_SIGNING_SECRET", "01234567890123456789012345678901")

	cfg, err := config.LoadAttachmentsConfig()
	if err != nil {
		t.Fatalf("expected 32-byte secret to be accepted, got: %v", err)
	}
	if len(cfg.DownloadSecret) != 32 {
		t.Fatalf("expected DownloadSecret length 32, got %d", len(cfg.DownloadSecret))
	}
}
