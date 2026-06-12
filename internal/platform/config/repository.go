package config

import (
	"fmt"
	"os"
	"strings"
)

const (
	RepositoryPostgres = "postgres"
)

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
