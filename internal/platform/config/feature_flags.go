package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

// FeatureFlagsConfig holds server-controlled feature flag values read from
// environment variables at startup.
type FeatureFlagsConfig struct {
	// MDDMNativeExportRolloutPercent is the percentage (0–100) of users for
	// whom the client-side MDDM DOCX export path is active.
	// Env: METALDOCS_MDDM_NATIVE_EXPORT_ROLLOUT_PCT (default 0)
	MDDMNativeExportRolloutPercent int
}

// ErrInvalidPercentage is returned when a rollout percentage env var is
// outside the valid 0-100 range.
var ErrInvalidPercentage = errors.New("invalid percentage")

// LoadFeatureFlagsConfig reads feature flag config from environment variables.
func LoadFeatureFlagsConfig() (FeatureFlagsConfig, error) {
	pct := 0
	if raw := strings.TrimSpace(os.Getenv("METALDOCS_MDDM_NATIVE_EXPORT_ROLLOUT_PCT")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return FeatureFlagsConfig{}, err
		}
		if parsed < 0 || parsed > 100 {
			return FeatureFlagsConfig{}, ErrInvalidPercentage
		}
		pct = parsed
	}
	return FeatureFlagsConfig{
		MDDMNativeExportRolloutPercent: pct,
	}, nil
}
