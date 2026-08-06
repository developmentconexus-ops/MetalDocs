//go:build !integration || production

package main

import (
	"database/sql"

	"metaldocs/internal/platform/httprouter"
)

// e2ePublisher returns nil in every build without the e2e handlers.
//
// The tag is the EXACT complement of the real file's. internal/test's existing
// pair uses `!integration && !production`, which leaves a hole: a literal
// `-tags production` build satisfies neither file and would not compile. That
// hole is latent and pre-existing; this pair must not inherit it.
func e2ePublisher(_ *sql.DB) httprouter.SurfacePublisher { return nil }
