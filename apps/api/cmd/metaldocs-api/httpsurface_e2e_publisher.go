//go:build integration && !production

package main

import (
	"context"
	"database/sql"

	e2etest "metaldocs/internal/test"

	"metaldocs/internal/platform/httprouter"
)

// e2ePublisher returns the e2e scaffolding publisher in builds that actually
// contain the handlers. See the stub for the complement.
func e2ePublisher() httprouter.SurfacePublisher { return newE2EPublisher() }

// e2ePublisherImpl is the httprouter.SurfacePublisher for the internal-e2e
// surface (api/openapi/internal-e2e.yaml). Mount is total (§4): it always
// registers all four routes; RegisterE2EHandlers' handlers answer 501
// themselves when the database or scheduler-tick dependency is absent, so a
// nil-wired publisher is a valid, non-panicking object, not a partial mount.
//
// newE2EPublisher takes no dependencies today — nothing in main() calls
// Mount on this publisher yet (that wiring is Task 15's, which assembles the
// `publishers` slice and threads deps.SQLDB / the scheduler tick through).
// Until then this is an unconsumed but fully-compiling artifact whose
// existence is what makes `e2ePublisher() != nil` a meaningful build-tag
// signal ahead of the wiring that uses it.
type e2ePublisherImpl struct {
	db               *sql.DB
	runSchedulerTick func(context.Context) error
}

func newE2EPublisher() *e2ePublisherImpl {
	return &e2ePublisherImpl{}
}

func (p *e2ePublisherImpl) Name() string { return "internal-e2e" }
func (p *e2ePublisherImpl) Tag() string  { return "internal-e2e" }

func (p *e2ePublisherImpl) Mount(mux httprouter.Muxer) {
	e2etest.RegisterE2EHandlers(mux, p.db, p.runSchedulerTick)
}
