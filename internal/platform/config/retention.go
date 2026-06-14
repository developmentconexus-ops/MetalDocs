package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// RetentionConfig holds audit-log retention settings.
type RetentionConfig struct {
	// Days is the number of days to retain audit events.
	// 0 means retention is disabled (default).
	Days int
}

// LoadRetentionConfig reads AUDIT_RETENTION_DAYS.
// Empty value -> Days 0, no error (retention disabled).
// Present and unparseable, or negative -> error (typos fail loud instead of
// silently disabling retention).
func LoadRetentionConfig() (RetentionConfig, error) {
	raw := strings.TrimSpace(os.Getenv("AUDIT_RETENTION_DAYS"))
	if raw == "" {
		return RetentionConfig{Days: 0}, nil
	}
	days, err := strconv.Atoi(raw)
	if err != nil || days < 0 {
		return RetentionConfig{}, fmt.Errorf("invalid AUDIT_RETENTION_DAYS value: %q", raw)
	}
	return RetentionConfig{Days: days}, nil
}
