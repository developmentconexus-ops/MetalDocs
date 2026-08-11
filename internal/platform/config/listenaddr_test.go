package config

import "testing"

// TestLoadListenAddr covers the A7.1 shared infra-port listen address loader
// used by worker/jobs (WORKER_METRICS_ADDR, JOBS_METRICS_ADDR) — same
// validation discipline LoadServerConfig already applies to METRICS_ADDR,
// factored out so a second/third binary does not re-implement it.
func TestLoadListenAddr(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		wantAddr string
		wantErr  bool
	}{
		{name: "default when unset", envValue: "", wantAddr: ":9091"},
		{name: "colon port", envValue: ":9191", wantAddr: ":9191"},
		{name: "host and port", envValue: "127.0.0.1:9091", wantAddr: "127.0.0.1:9091"},
		{name: "whitespace trimmed", envValue: " :9091 ", wantAddr: ":9091"},
		{name: "no colon invalid", envValue: "9091", wantErr: true},
		{name: "port 0 invalid", envValue: ":0", wantErr: true},
		{name: "port 65536 invalid", envValue: ":65536", wantErr: true},
		{name: "non-numeric port invalid", envValue: ":abc", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("WORKER_METRICS_ADDR_TEST", tt.envValue)
			addr, err := LoadListenAddr("WORKER_METRICS_ADDR_TEST", ":9091")
			if (err != nil) != tt.wantErr {
				t.Fatalf("LoadListenAddr() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && addr != tt.wantAddr {
				t.Fatalf("addr = %q, want %q", addr, tt.wantAddr)
			}
		})
	}
}
