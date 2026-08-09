package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

// GotenbergConfig holds Gotenberg (DOCX-to-PDF rendering) configuration read
// from environment variables at startup.
type GotenbergConfig struct {
	// immutable after Load().
	Enabled bool
	// immutable after Load().
	URL string
}

// LoadGotenbergConfig reads Gotenberg config from environment variables.
// Returns a disabled config (no error) when METALDOCS_GOTENBERG_URL is unset.
func LoadGotenbergConfig() (GotenbergConfig, error) {
	rawURL := strings.TrimSpace(os.Getenv("METALDOCS_GOTENBERG_URL"))
	if rawURL == "" {
		return GotenbergConfig{}, nil
	}
	if err := validateGotenbergURL(rawURL); err != nil {
		return GotenbergConfig{}, err
	}
	return GotenbergConfig{
		Enabled: true,
		URL:     rawURL,
	}, nil
}

func validateGotenbergURL(rawURL string) error {
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("METALDOCS_GOTENBERG_URL must be an absolute http(s) URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("METALDOCS_GOTENBERG_URL must use http or https")
	}
	return nil
}
