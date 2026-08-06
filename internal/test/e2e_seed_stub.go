//go:build !integration && !production

package test

import (
	"context"
	"database/sql"

	"metaldocs/internal/platform/httprouter"
)

func RegisterE2EHandlers(_ httprouter.Muxer, _ *sql.DB, _ func(context.Context) error) {}
