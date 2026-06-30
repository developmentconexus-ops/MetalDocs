package config

import "testing"

// setMinioEnv sets the minimum set of env vars for a minio LoadAttachmentsConfig call
// and returns a cleanup func. Uses t.Setenv so cleanup is automatic via testing.T.
func setMinioEnv(t *testing.T, region string) {
	t.Helper()
	t.Setenv("METALDOCS_STORAGE_PROVIDER", "minio")
	t.Setenv("METALDOCS_ATTACHMENTS_SIGNING_SECRET", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	t.Setenv("METALDOCS_MINIO_ENDPOINT", "minio:9000")
	t.Setenv("METALDOCS_MINIO_ACCESS_KEY", "minioadmin")
	t.Setenv("METALDOCS_MINIO_SECRET_KEY", "minioadmin")
	t.Setenv("METALDOCS_MINIO_BUCKET", "metaldocs-attachments")
	t.Setenv("METALDOCS_MINIO_REGION", region)
}

func TestLoadAttachmentsConfig_MinIORegion_Default(t *testing.T) {
	setMinioEnv(t, "") // unset → default "us-east-1"

	cfg, err := LoadAttachmentsConfig()
	if err != nil {
		t.Fatalf("LoadAttachmentsConfig: %v", err)
	}
	if cfg.MinIORegion != "us-east-1" {
		t.Fatalf("MinIORegion = %q, want \"us-east-1\"", cfg.MinIORegion)
	}
}

func TestLoadAttachmentsConfig_MinIORegion_Override(t *testing.T) {
	setMinioEnv(t, "eu-west-1")

	cfg, err := LoadAttachmentsConfig()
	if err != nil {
		t.Fatalf("LoadAttachmentsConfig: %v", err)
	}
	if cfg.MinIORegion != "eu-west-1" {
		t.Fatalf("MinIORegion = %q, want \"eu-west-1\"", cfg.MinIORegion)
	}
}
