package config

import (
	"testing"
)

func TestLoadFanoutConfig(t *testing.T) {
	tests := []struct {
		name      string
		urlEnv    string
		tokenEnv  string
		wantURL   string
		wantToken string
		wantErr   bool
	}{
		{name: "both unset", urlEnv: "", tokenEnv: "", wantURL: "", wantToken: ""},
		{name: "url and token both set", urlEnv: "http://fanout", tokenEnv: "tok", wantURL: "http://fanout", wantToken: "tok"},
		{name: "url set token missing", urlEnv: "http://fanout", tokenEnv: "", wantErr: true},
		{name: "token set url missing", urlEnv: "", tokenEnv: "tok", wantURL: "", wantToken: "tok"},
		{name: "whitespace trimmed", urlEnv: "  http://fanout  ", tokenEnv: "  tok  ", wantURL: "http://fanout", wantToken: "tok"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("METALDOCS_FANOUT_URL", tt.urlEnv)
			t.Setenv("METALDOCS_DOCX_RENDERER_SERVICE_TOKEN", tt.tokenEnv)
			cfg, err := LoadFanoutConfig()
			if (err != nil) != tt.wantErr {
				t.Fatalf("LoadFanoutConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if cfg.URL != tt.wantURL {
					t.Errorf("URL = %q, want %q", cfg.URL, tt.wantURL)
				}
				if cfg.ServiceToken != tt.wantToken {
					t.Errorf("ServiceToken = %q, want %q", cfg.ServiceToken, tt.wantToken)
				}
			}
		})
	}
}
