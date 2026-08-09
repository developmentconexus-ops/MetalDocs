//go:build !integration && !production

package test

import (
	"context"
	"database/sql"

	"metaldocs/internal/platform/httprouter"
)

// RegisterE2EHandlers is a no-op stub for builds that exclude the e2e seed
// handlers (neither the integration nor the production tag is set); the real
// registration lives in the //go:build integration variant of this file.
func RegisterE2EHandlers(_ httprouter.Muxer, _ *sql.DB, _ func(context.Context) error) {}
