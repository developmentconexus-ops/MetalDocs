package config

import (
	"testing"
)

func TestLoadRetentionConfig(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		wantDays int
		wantErr  bool
	}{
		{name: "unset disables retention", envValue: "", wantDays: 0},
		{name: "zero disables retention", envValue: "0", wantDays: 0},
		{name: "positive value", envValue: "90", wantDays: 90},
		{name: "one day", envValue: "1", wantDays: 1},
		{name: "negative fails loud", envValue: "-1", wantErr: true},
		{name: "non-numeric fails loud", envValue: "thirty", wantErr: true},
		{name: "float fails loud", envValue: "90.5", wantErr: true},
		{name: "whitespace trimmed valid", envValue: " 30 ", wantDays: 30},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("AUDIT_RETENTION_DAYS", tt.envValue)
			cfg, err := LoadRetentionConfig()
			if (err != nil) != tt.wantErr {
				t.Fatalf("LoadRetentionConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && cfg.Days != tt.wantDays {
				t.Fatalf("Days = %d, want %d", cfg.Days, tt.wantDays)
			}
		})
	}
}
