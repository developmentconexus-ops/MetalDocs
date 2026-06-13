package infrastructure

import (
	"context"
	"database/sql"
	"fmt"

	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/tenant"
)

const NoActor = ""

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
	actorID := iamdomain.UserIDFromContext(ctx)
	if actorID == NoActor {
		return fmt.Errorf("taxonomy: actor context missing")
	}
	return authz.SeedTxIdentity(ctx, tx, tenantID, actorID)
}
