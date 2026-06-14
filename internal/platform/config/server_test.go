package config

import (
	"testing"
)

func TestLoadServerConfig(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		wantAddr string
		wantErr  bool
	}{
		{name: "default when unset", envValue: "", wantAddr: ":8080"},
		{name: "valid port 8081", envValue: "8081", wantAddr: ":8081"},
		{name: "valid port 1", envValue: "1", wantAddr: ":1"},
		{name: "valid port 65535", envValue: "65535", wantAddr: ":65535"},
		{name: "port 0 invalid", envValue: "0", wantErr: true},
		{name: "port 65536 invalid", envValue: "65536", wantErr: true},
		{name: "negative port", envValue: "-1", wantErr: true},
		{name: "non-numeric", envValue: "abc", wantErr: true},
		{name: "whitespace trimmed valid", envValue: " 9090 ", wantAddr: ":9090"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("APP_PORT", tt.envValue)
			cfg, err := LoadServerConfig()
			if (err != nil) != tt.wantErr {
				t.Fatalf("LoadServerConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && cfg.Addr != tt.wantAddr {
				t.Fatalf("Addr = %q, want %q", cfg.Addr, tt.wantAddr)
			}
		})
	}
}
