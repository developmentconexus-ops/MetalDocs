package application

import (
	"context"
	"database/sql"

	platformdb "metaldocs/internal/platform/db"
	platformtenant "metaldocs/internal/platform/tenant"
)

func newTxRunner(database *sql.DB) platformdb.TxRunner {
	return platformdb.NewTxRunner(database)
}

// authzCtx builds a context carrying the tenant+actor identity the TxRunner
// chokepoint (internal/platform/db/runner.go, M3 F3.1) auto-seeds from ctx on
// every Do/DoReadOnly. Tests that assert against a mocked set_config exec
// (matching the chokepoint's SQL) must drive the call through a context
// carrying the same tenant/actor the mock expects.
func authzCtx(tenantID, actorID string) context.Context {
	ctx := platformtenant.WithTenantID(context.Background(), tenantID)
	return platformtenant.WithActorID(ctx, actorID)
}
