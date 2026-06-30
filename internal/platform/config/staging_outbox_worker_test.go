package config

import (
	"testing"
)

func TestLoadStagingOutboxWorkerConfig_Defaults(t *testing.T) {
	t.Setenv("METALDOCS_STAGING_OUTBOX_POLL_INTERVAL_SECONDS", "")
	t.Setenv("METALDOCS_STAGING_OUTBOX_BATCH_SIZE", "")
	t.Setenv("METALDOCS_STAGING_OUTBOX_MAX_ATTEMPTS", "")
	t.Setenv("METALDOCS_STAGING_OUTBOX_STALE_AFTER_SECONDS", "")

	cfg, err := LoadStagingOutboxWorkerConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PollIntervalSeconds != 5 {
		t.Errorf("PollIntervalSeconds = %d, want 5", cfg.PollIntervalSeconds)
	}
	if cfg.BatchSize != 10 {
		t.Errorf("BatchSize = %d, want 10", cfg.BatchSize)
	}
	if cfg.MaxAttempts != 5 {
		t.Errorf("MaxAttempts = %d, want 5", cfg.MaxAttempts)
	}
	if cfg.StaleAfterSeconds != 300 {
		t.Errorf("StaleAfterSeconds = %d, want 300", cfg.StaleAfterSeconds)
	}
}

func TestLoadStagingOutboxWorkerConfig_EnvOverrides(t *testing.T) {
	t.Setenv("METALDOCS_STAGING_OUTBOX_POLL_INTERVAL_SECONDS", "15")
	t.Setenv("METALDOCS_STAGING_OUTBOX_BATCH_SIZE", "20")
	t.Setenv("METALDOCS_STAGING_OUTBOX_MAX_ATTEMPTS", "3")
	t.Setenv("METALDOCS_STAGING_OUTBOX_STALE_AFTER_SECONDS", "600")

	cfg, err := LoadStagingOutboxWorkerConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PollIntervalSeconds != 15 {
		t.Errorf("PollIntervalSeconds = %d, want 15", cfg.PollIntervalSeconds)
	}
	if cfg.BatchSize != 20 {
		t.Errorf("BatchSize = %d, want 20", cfg.BatchSize)
	}
	if cfg.MaxAttempts != 3 {
		t.Errorf("MaxAttempts = %d, want 3", cfg.MaxAttempts)
	}
	if cfg.StaleAfterSeconds != 600 {
		t.Errorf("StaleAfterSeconds = %d, want 600", cfg.StaleAfterSeconds)
	}
}

func TestLoadStagingOutboxWorkerConfig_InvalidPollInterval(t *testing.T) {
	t.Setenv("METALDOCS_STAGING_OUTBOX_POLL_INTERVAL_SECONDS", "notanumber")
	_, err := LoadStagingOutboxWorkerConfig()
	if err == nil {
		t.Fatal("expected error for invalid METALDOCS_STAGING_OUTBOX_POLL_INTERVAL_SECONDS")
	}
}

func TestLoadStagingOutboxWorkerConfig_ZeroRejected(t *testing.T) {
	for _, tt := range []struct {
		env string
		val string
	}{
		{"METALDOCS_STAGING_OUTBOX_POLL_INTERVAL_SECONDS", "0"},
		{"METALDOCS_STAGING_OUTBOX_BATCH_SIZE", "0"},
		{"METALDOCS_STAGING_OUTBOX_MAX_ATTEMPTS", "0"},
		{"METALDOCS_STAGING_OUTBOX_STALE_AFTER_SECONDS", "0"},
	} {
		t.Run(tt.env, func(t *testing.T) {
			t.Setenv(tt.env, tt.val)
			if _, err := LoadStagingOutboxWorkerConfig(); err == nil {
				t.Fatalf("expected error for %s=0", tt.env)
			}
		})
	}
}

func TestLoadStagingOutboxWorkerConfig_WhitespaceTrimmed(t *testing.T) {
	t.Setenv("METALDOCS_STAGING_OUTBOX_POLL_INTERVAL_SECONDS", "  7  ")
	cfg, err := LoadStagingOutboxWorkerConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PollIntervalSeconds != 7 {
		t.Errorf("PollIntervalSeconds = %d, want 7", cfg.PollIntervalSeconds)
	}
}
