package config

import (
	"os"
	"strconv"
	"strings"
)

type DocgenConfig struct {
	// immutable after Load().
	Enabled bool
	// immutable after Load().
	APIURL string
	// immutable after Load().
	RequestTimeoutSeconds int
}

func LoadDocgenConfig() (DocgenConfig, error) {
	appEnv := strings.TrimSpace(os.Getenv("APP_ENV"))
	apiURL := strings.TrimSpace(os.Getenv("METALDOCS_DOCGEN_API_URL"))
	if apiURL == "" && strings.EqualFold(appEnv, "local") {
		apiURL = "http://127.0.0.1:3001"
	}
	enabled := apiURL != ""

	timeoutSeconds := 10
	if raw := strings.TrimSpace(os.Getenv("METALDOCS_DOCGEN_REQUEST_TIMEOUT_SECONDS")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return DocgenConfig{}, err
		}
		timeoutSeconds = parsed
	}
	if timeoutSeconds <= 0 {
		timeoutSeconds = 30
	}

	return DocgenConfig{
		Enabled:               enabled,
		APIURL:                apiURL,
		RequestTimeoutSeconds: timeoutSeconds,
	}, nil
}
