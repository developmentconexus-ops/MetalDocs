package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// StagingOutboxWorkerConfig controls the poll/retry behaviour of
// StagingOutboxWorker (render/fanout). The defaults are the literal values that
// were previously hardcoded in NewStagingOutboxWorker so that existing
// deployments without any METALDOCS_STAGING_OUTBOX_* env-vars are unaffected.
type StagingOutboxWorkerConfig struct {
	PollIntervalSeconds int
	BatchSize           int
	MaxAttempts         int
	// StaleAfterSeconds is how long a claimed row may remain in 'processing'
	// before ResetStaleClaims reclaims it.
	StaleAfterSeconds int
}

// LoadStagingOutboxWorkerConfig reads METALDOCS_STAGING_OUTBOX_* env-vars,
// mirroring the LoadWorkerConfig pattern in worker.go. Defaults match the
// previous hardcoded literals (5s poll, batch 10, 5 attempts, 300s stale).
func LoadStagingOutboxWorkerConfig() (StagingOutboxWorkerConfig, error) {
	cfg := StagingOutboxWorkerConfig{
		PollIntervalSeconds: 5,
		BatchSize:           10,
		MaxAttempts:         5,
		StaleAfterSeconds:   300,
	}

	if raw := strings.TrimSpace(os.Getenv("METALDOCS_STAGING_OUTBOX_POLL_INTERVAL_SECONDS")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 {
			return StagingOutboxWorkerConfig{}, fmt.Errorf("invalid METALDOCS_STAGING_OUTBOX_POLL_INTERVAL_SECONDS")
		}
		cfg.PollIntervalSeconds = value
	}
	if raw := strings.TrimSpace(os.Getenv("METALDOCS_STAGING_OUTBOX_BATCH_SIZE")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 {
			return StagingOutboxWorkerConfig{}, fmt.Errorf("invalid METALDOCS_STAGING_OUTBOX_BATCH_SIZE")
		}
		cfg.BatchSize = value
	}
	if raw := strings.TrimSpace(os.Getenv("METALDOCS_STAGING_OUTBOX_MAX_ATTEMPTS")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 {
			return StagingOutboxWorkerConfig{}, fmt.Errorf("invalid METALDOCS_STAGING_OUTBOX_MAX_ATTEMPTS")
		}
		cfg.MaxAttempts = value
	}
	if raw := strings.TrimSpace(os.Getenv("METALDOCS_STAGING_OUTBOX_STALE_AFTER_SECONDS")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 {
			return StagingOutboxWorkerConfig{}, fmt.Errorf("invalid METALDOCS_STAGING_OUTBOX_STALE_AFTER_SECONDS")
		}
		cfg.StaleAfterSeconds = value
	}

	return cfg, nil
}
