package infrastructure

import (
	"context"
	"database/sql"
	"fmt"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/tenant"
)

const NoActor = ""

func setAuthzGUC(ctx context.Context, tx *sql.Tx) error {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("taxonomy: tenant context: %w", err)
	}
	actorID := iamdomain.UserIDFromContext(ctx)
	if actorID == NoActor {
		return fmt.Errorf("taxonomy: actor context missing")
	}
	if _, err := tx.ExecContext(ctx, "SELECT set_config('metaldocs.tenant_id', $1, true)", tenantID); err != nil {
		return fmt.Errorf("taxonomy: set tenant_id GUC: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "SELECT set_config('metaldocs.actor_id', $1, true)", actorID); err != nil {
		return fmt.Errorf("taxonomy: set actor_id GUC: %w", err)
	}
	return nil
}
