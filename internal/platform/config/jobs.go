package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/riverqueue/river"
)

type JobsConfig struct {
	Enabled     bool
	RiverSchema string
	Queues      map[string]river.QueueConfig
}

func LoadJobsConfig() (JobsConfig, error) {
	cfg := JobsConfig{
		Enabled:     true,
		RiverSchema: strings.TrimSpace(os.Getenv("METALDOCS_JOBS_RIVER_SCHEMA")),
		Queues: map[string]river.QueueConfig{
			"temporal": {MaxWorkers: 10},
		},
	}

	if raw := strings.TrimSpace(os.Getenv("METALDOCS_JOBS_ENABLED")); raw != "" {
		cfg.Enabled = strings.EqualFold(raw, "true") || raw == "1"
	}

	if raw := strings.TrimSpace(os.Getenv("METALDOCS_JOBS_TEMPORAL_MAX_WORKERS")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 {
			return JobsConfig{}, fmt.Errorf("invalid METALDOCS_JOBS_TEMPORAL_MAX_WORKERS")
		}
		cfg.Queues["temporal"] = river.QueueConfig{MaxWorkers: value}
	}

	return cfg, nil
}
