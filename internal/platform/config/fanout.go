package config

import (
	"fmt"
	"os"
	"strings"
)

// FanoutConfig holds the eigenpal/fanout renderer connection settings.
type FanoutConfig struct {
	URL          string
	ServiceToken string
}

// LoadFanoutConfig reads METALDOCS_FANOUT_URL and METALDOCS_DOCX_RENDERER_SERVICE_TOKEN.
// If URL is non-empty and ServiceToken is empty, an error is returned because
// the renderer requires authentication.
func LoadFanoutConfig() (FanoutConfig, error) {
	url := strings.TrimSpace(os.Getenv("METALDOCS_FANOUT_URL"))
	token := strings.TrimSpace(os.Getenv("METALDOCS_DOCX_RENDERER_SERVICE_TOKEN"))
	if url != "" && token == "" {
		return FanoutConfig{}, fmt.Errorf("METALDOCS_DOCX_RENDERER_SERVICE_TOKEN is required when METALDOCS_FANOUT_URL is set")
	}
	return FanoutConfig{URL: url, ServiceToken: token}, nil
}
