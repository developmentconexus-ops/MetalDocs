package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// DocgenV2Config holds connection settings for the docgen-v2 service.
type DocgenV2Config struct {
	Enabled               bool
	APIURL                string
	AllowedHosts          []string
	ServiceToken          string
	RequestTimeoutSeconds int
}

// LoadDocgenV2Config reads docgen-v2 config from environment variables.
// The service is considered disabled when METALDOCS_DOCGEN_V2_URL is empty.
func LoadDocgenV2Config() (DocgenV2Config, error) {
	apiURL := strings.TrimSpace(os.Getenv("METALDOCS_DOCGEN_V2_URL"))
	if apiURL == "" {
		return DocgenV2Config{}, nil
	}
	token := strings.TrimSpace(os.Getenv("METALDOCS_DOCGEN_V2_SERVICE_TOKEN"))
	if token == "" {
		return DocgenV2Config{}, fmt.Errorf("METALDOCS_DOCGEN_V2_SERVICE_TOKEN required when docgen-v2 is enabled")
	}
	timeoutSeconds := 10
	if raw := strings.TrimSpace(os.Getenv("METALDOCS_DOCGEN_V2_REQUEST_TIMEOUT_SECONDS")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			timeoutSeconds = parsed
		}
	}

	allowedHosts := parseDocgenV2AllowedHosts(os.Getenv("METALDOCS_DOCGEN_V2_ALLOWED_HOSTS"))
	if len(allowedHosts) == 0 {
		return DocgenV2Config{}, fmt.Errorf("METALDOCS_DOCGEN_V2_ALLOWED_HOSTS required when docgen-v2 is enabled")
	}
	if err := validateDocgenV2URL(apiURL, allowedHosts); err != nil {
		return DocgenV2Config{}, err
	}

	return DocgenV2Config{
		Enabled:               true,
		APIURL:                apiURL,
		AllowedHosts:          allowedHosts,
		ServiceToken:          token,
		RequestTimeoutSeconds: timeoutSeconds,
	}, nil
}

func validateDocgenV2URL(rawURL string, allowedHosts []string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("METALDOCS_DOCGEN_V2_URL must be an absolute http(s) URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("METALDOCS_DOCGEN_V2_URL must use http or https")
	}
	if !docgenV2HostAllowed(parsed, allowedHosts) {
		return fmt.Errorf("METALDOCS_DOCGEN_V2_URL host %q is not in METALDOCS_DOCGEN_V2_ALLOWED_HOSTS", parsed.Host)
	}
	return nil
}

func parseDocgenV2AllowedHosts(raw string) []string {
	parts := strings.Split(raw, ",")
	hosts := make([]string, 0, len(parts))
	for _, part := range parts {
		host := strings.ToLower(strings.TrimSpace(part))
		if host != "" {
			hosts = append(hosts, host)
		}
	}
	return hosts
}

func docgenV2HostAllowed(parsed *url.URL, allowedHosts []string) bool {
	hostWithPort := strings.ToLower(parsed.Host)
	hostOnly := strings.ToLower(parsed.Hostname())
	for _, allowed := range allowedHosts {
		if allowed == hostWithPort || allowed == hostOnly {
			return true
		}
	}
	return false
}
