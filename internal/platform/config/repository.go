package config

import (
	"fmt"
	"os"
	"strings"
)

// Repository mode values for METALDOCS_REPOSITORY.
const (
	RepositoryPostgres = "postgres"
)

// RepositoryMode reads METALDOCS_REPOSITORY and returns the configured
// repository mode, defaulting to and validating against RepositoryPostgres
// (the only supported mode).
func RepositoryMode() (string, error) {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("METALDOCS_REPOSITORY")))
	if mode == "" {
		mode = RepositoryPostgres
	}
	if mode != RepositoryPostgres {
		return "", fmt.Errorf("invalid METALDOCS_REPOSITORY: %q (only %q is supported)", mode, RepositoryPostgres)
	}
	return mode, nil
}
