//go:build !integration && !production

package test

import (
	"context"
	"database/sql"
	"net/http"
)

func RegisterE2EHandlers(_ *http.ServeMux, _ *sql.DB, _ func(context.Context) error) {}
