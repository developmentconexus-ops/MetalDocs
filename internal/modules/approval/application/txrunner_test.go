package application

import (
	"context"
	"database/sql"

	"metaldocs/internal/platform/db"
	platformtenant "metaldocs/internal/platform/tenant"
)

// newTxRunner wraps a test *sql.DB in the db.TxRunner port so service-method
// call sites compile after the H-1d *sql.DB -> db.TxRunner refactor. Named to
// avoid the local `db` var that shadows the db package inside test builders.
func newTxRunner(database *sql.DB) db.TxRunner { return db.NewTxRunner(database) }

// authzCtx builds a context carrying the tenant+actor identity the TxRunner
// chokepoint (internal/platform/db/runner.go, M3 F3.1) auto-seeds from ctx on
// every Do/DoReadOnly. Tests that assert against a mocked set_config exec
// (matching the chokepoint's SQL) must drive the call through a context
// carrying the same tenant/actor the mock expects.
func authzCtx(tenantID, actorID string) context.Context {
	ctx := platformtenant.WithTenantID(context.Background(), tenantID)
	return platformtenant.WithActorID(ctx, actorID)
}
