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
			t.Setenv("METRICS_ADDR", "")
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

// TestLoadServerConfig_MetricsAddr covers the F-R1 dedicated metrics listener
// address: default :9090, valid ":port"/"host:port" forms, and rejection of
// malformed or out-of-range values (fail-fast at boot, same discipline as
// APP_PORT).
func TestLoadServerConfig_MetricsAddr(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		wantAddr string
		wantErr  bool
	}{
		{name: "default when unset", envValue: "", wantAddr: ":9090"},
		{name: "colon port", envValue: ":9191", wantAddr: ":9191"},
		{name: "host and port", envValue: "127.0.0.1:9090", wantAddr: "127.0.0.1:9090"},
		{name: "whitespace trimmed", envValue: " :9090 ", wantAddr: ":9090"},
		{name: "no colon invalid", envValue: "9090", wantErr: true},
		{name: "port 0 invalid", envValue: ":0", wantErr: true},
		{name: "port 65536 invalid", envValue: ":65536", wantErr: true},
		{name: "non-numeric port invalid", envValue: ":abc", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("APP_PORT", "")
			t.Setenv("METRICS_ADDR", tt.envValue)
			cfg, err := LoadServerConfig()
			if (err != nil) != tt.wantErr {
				t.Fatalf("LoadServerConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && cfg.MetricsAddr != tt.wantAddr {
				t.Fatalf("MetricsAddr = %q, want %q", cfg.MetricsAddr, tt.wantAddr)
			}
		})
	}
}
