package config

import "os"

type GotenbergConfig struct {
	// immutable after Load().
	Enabled bool
	// immutable after Load().
	URL string
}

func LoadGotenbergConfig() GotenbergConfig {
	url := os.Getenv("METALDOCS_GOTENBERG_URL")
	if url == "" {
		return GotenbergConfig{}
	}
	return GotenbergConfig{
		Enabled: true,
		URL:     url,
	}
}
