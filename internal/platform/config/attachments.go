package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Storage provider values for METALDOCS_STORAGE_PROVIDER.
const (
	StorageProviderMemory StorageProvider = "memory"
	StorageProviderLocal  StorageProvider = "local"
	StorageProviderMinIO  StorageProvider = "minio"
)

// StorageProvider identifies which attachment storage backend is active.
type StorageProvider string

// AppEnv identifies the deployment environment (local, dev, staging, production).
type AppEnv string

// Deployment environment values for APP_ENV.
const (
	AppEnvLocal      AppEnv = "local"
	AppEnvDev        AppEnv = "dev"
	AppEnvStaging    AppEnv = "staging"
	AppEnvProduction AppEnv = "production"
)

// AttachmentsConfig holds attachment storage configuration read from
// environment variables at startup.
type AttachmentsConfig struct {
	Provider              StorageProvider
	AppEnv                AppEnv
	RootPath              string
	DownloadSecret        string
	DownloadTTLSeconds    int
	MinIOEndpoint         string
	MinIOPublicEndpoint   string
	MinIOAccessKey        string
	MinIOSecretKey        string
	MinIOBucket           string
	MinIORegion           string
	MinIOUseSSL           bool
	MinIOAutoCreateBucket bool
}

// LoadAttachmentsConfig reads attachment storage config from environment variables.
func LoadAttachmentsConfig() (AttachmentsConfig, error) {
	appEnv := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	if appEnv == "" {
		appEnv = string(AppEnvLocal)
	}

	provider := StorageProvider(strings.ToLower(strings.TrimSpace(os.Getenv("METALDOCS_STORAGE_PROVIDER"))))
	if provider == "" {
		provider = StorageProviderLocal
	}
	switch provider {
	case StorageProviderMemory, StorageProviderLocal, StorageProviderMinIO:
	default:
		return AttachmentsConfig{}, fmt.Errorf("invalid METALDOCS_STORAGE_PROVIDER: %s", provider)
	}

	root := strings.TrimSpace(os.Getenv("METALDOCS_ATTACHMENTS_ROOT"))
	if root == "" {
		root = "non_git/attachments"
	}

	secret := strings.TrimSpace(os.Getenv("METALDOCS_ATTACHMENTS_SIGNING_SECRET"))
	if secret == "" {
		return AttachmentsConfig{}, fmt.Errorf("METALDOCS_ATTACHMENTS_SIGNING_SECRET is required for provider %s", provider)
	}
	if len(secret) < 32 {
		return AttachmentsConfig{}, fmt.Errorf("METALDOCS_ATTACHMENTS_SIGNING_SECRET must be at least 32 bytes, got %d", len(secret))
	}

	ttlSeconds := 300
	if raw := strings.TrimSpace(os.Getenv("METALDOCS_ATTACHMENTS_DOWNLOAD_TTL_SECONDS")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 30 {
			return AttachmentsConfig{}, fmt.Errorf("invalid METALDOCS_ATTACHMENTS_DOWNLOAD_TTL_SECONDS")
		}
		ttlSeconds = parsed
	}

	cfg := AttachmentsConfig{
		Provider:           provider,
		AppEnv:             AppEnv(appEnv),
		RootPath:           root,
		DownloadSecret:     secret,
		DownloadTTLSeconds: ttlSeconds,
	}

	if provider == StorageProviderMinIO {
		if err := applyMinIOConfig(&cfg); err != nil {
			return AttachmentsConfig{}, err
		}
	}

	return cfg, nil
}

// applyMinIOConfig reads the METALDOCS_MINIO_* environment variables into cfg
// and validates that the required fields are present.
func applyMinIOConfig(cfg *AttachmentsConfig) error {
	cfg.MinIOEndpoint = strings.TrimSpace(os.Getenv("METALDOCS_MINIO_ENDPOINT"))
	cfg.MinIOPublicEndpoint = strings.TrimSpace(os.Getenv("METALDOCS_MINIO_PUBLIC_ENDPOINT"))
	cfg.MinIOAccessKey = strings.TrimSpace(os.Getenv("METALDOCS_MINIO_ACCESS_KEY"))
	cfg.MinIOSecretKey = os.Getenv("METALDOCS_MINIO_SECRET_KEY")
	cfg.MinIOBucket = strings.TrimSpace(os.Getenv("METALDOCS_MINIO_BUCKET"))
	cfg.MinIORegion = strings.TrimSpace(os.Getenv("METALDOCS_MINIO_REGION"))
	if cfg.MinIORegion == "" {
		cfg.MinIORegion = "us-east-1"
	}
	cfg.MinIOUseSSL = parseBoolEnv("METALDOCS_MINIO_USE_SSL", false)
	cfg.MinIOAutoCreateBucket = parseBoolEnv("METALDOCS_MINIO_AUTO_CREATE_BUCKET", false)

	if cfg.MinIOEndpoint == "" || cfg.MinIOAccessKey == "" || cfg.MinIOSecretKey == "" || cfg.MinIOBucket == "" {
		return fmt.Errorf("minio config missing: set METALDOCS_MINIO_ENDPOINT/METALDOCS_MINIO_ACCESS_KEY/METALDOCS_MINIO_SECRET_KEY/METALDOCS_MINIO_BUCKET")
	}
	if cfg.MinIOPublicEndpoint == "" {
		cfg.MinIOPublicEndpoint = cfg.MinIOEndpoint
	}
	return nil
}

// ParseBoolEnv reads an environment variable and interprets it as a boolean.
// True values: "1", "true", "yes", "on" (case-insensitive).
// False values: "0", "false", "no", "off" (case-insensitive).
// Empty or unset returns defaultValue.
func ParseBoolEnv(name string, defaultValue bool) bool {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv(name)))
	switch raw {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return defaultValue
	}
}

func parseBoolEnv(name string, defaultValue bool) bool {
	return ParseBoolEnv(name, defaultValue)
}
