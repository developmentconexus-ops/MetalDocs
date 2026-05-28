package config

import (
	"errors"
	"testing"
)

func TestLoadFeatureFlagsConfigRejectsOutOfRangePercentages(t *testing.T) {
	for _, raw := range []string{"-1", "101"} {
		t.Setenv("METALDOCS_MDDM_NATIVE_EXPORT_ROLLOUT_PCT", raw)

		_, err := LoadFeatureFlagsConfig()
		if !errors.Is(err, ErrInvalidPercentage) {
			t.Fatalf("raw=%s expected ErrInvalidPercentage, got %v", raw, err)
		}
	}
}
