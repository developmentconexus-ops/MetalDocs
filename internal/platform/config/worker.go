package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// WorkerConfig holds the async outbox worker configuration read from
// environment variables at startup.
type WorkerConfig struct {
	PollIntervalSeconds int
	BatchSize           int
	RunOnce             bool
	MaxAttempts         int
	// RetryMaxSeconds must stay >= RetryBaseSeconds so backoff growth remains monotonic.
	RetryBaseSeconds int
	RetryMaxSeconds  int
}

// LoadWorkerConfig reads the worker configuration from environment
// variables, applying defaults for anything unset.
func LoadWorkerConfig() (WorkerConfig, error) {
	cfg := WorkerConfig{
		PollIntervalSeconds: 10,
		BatchSize:           25,
		MaxAttempts:         5,
		RetryBaseSeconds:    10,
		RetryMaxSeconds:     300,
	}

	var err error
	if cfg.PollIntervalSeconds, err = parseWorkerIntEnv("METALDOCS_WORKER_POLL_INTERVAL_SECONDS", cfg.PollIntervalSeconds, 1); err != nil {
		return WorkerConfig{}, err
	}
	if cfg.BatchSize, err = parseWorkerIntEnv("METALDOCS_WORKER_BATCH_SIZE", cfg.BatchSize, 1); err != nil {
		return WorkerConfig{}, err
	}
	if raw := strings.TrimSpace(os.Getenv("METALDOCS_WORKER_RUN_ONCE")); raw != "" {
		cfg.RunOnce = strings.EqualFold(raw, "true") || raw == "1"
	}
	if cfg.MaxAttempts, err = parseWorkerIntEnv("METALDOCS_WORKER_MAX_ATTEMPTS", cfg.MaxAttempts, 1); err != nil {
		return WorkerConfig{}, err
	}
	if cfg.RetryBaseSeconds, err = parseWorkerIntEnv("METALDOCS_WORKER_RETRY_BASE_SECONDS", cfg.RetryBaseSeconds, 1); err != nil {
		return WorkerConfig{}, err
	}
	// RetryMaxSeconds' floor is RetryBaseSeconds (already resolved above), not a
	// fixed constant, so backoff growth stays monotonic.
	if cfg.RetryMaxSeconds, err = parseWorkerIntEnv("METALDOCS_WORKER_RETRY_MAX_SECONDS", cfg.RetryMaxSeconds, cfg.RetryBaseSeconds); err != nil {
		return WorkerConfig{}, err
	}

	return cfg, nil
}

// parseWorkerIntEnv reads an integer environment variable, returning
// defaultValue when unset. A present-but-unparsable value, or one below min,
// yields "invalid <name>" — the single error shape for every worker int setting.
func parseWorkerIntEnv(name string, defaultValue, min int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return defaultValue, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < min {
		return 0, fmt.Errorf("invalid %s", name)
	}
	return value, nil
}
