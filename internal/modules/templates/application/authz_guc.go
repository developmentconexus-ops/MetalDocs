package application

import (
	"context"
	"database/sql"
)

// setAuthzGUC sets the transaction-local context required by authz.Require.
func setAuthzGUC(ctx context.Context, tx *sql.Tx, tenantID, actorID string) error {
	if _, err := tx.ExecContext(ctx, "SELECT set_config('metaldocs.tenant_id', $1, true)", tenantID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "SELECT set_config('metaldocs.actor_id', $1, true)", actorID); err != nil {
		return err
	}
	return nil
}
