package infrastructure

import (
	"context"
	"database/sql"
	"fmt"

	"metaldocs/internal/modules/iam/authz"
	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/tenant"
)

// setAuthzGUC resolves the tenant/actor identity from ctx and seeds the
// per-transaction authz GUCs via the canonical authz.SeedTxIdentity. taxonomy
// keeps this thin ctx-resolving wrapper (its repositories call it at 24 sites)
// instead of duplicating the resolution at each one; the GUC SQL itself is no
// longer copy-pasted (H-1a).
func setAuthzGUC(ctx context.Context, tx *sql.Tx) error {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("taxonomy: tenant context: %w", err)
	}
	// A3.3: the local NoActor == "" sentinel is gone — absence is now the
	// canonical accessor's explicit failure (which also rejects a
	// whitespace-only actor, which the "" comparison let through).
	actorID, err := authn.RequireUserID(ctx)
	if err != nil {
		return fmt.Errorf("taxonomy: actor context missing: %w", err)
	}
	return authz.SeedTxIdentity(ctx, tx, tenantID, actorID)
}
