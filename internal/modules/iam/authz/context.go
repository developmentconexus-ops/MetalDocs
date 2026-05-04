package authz

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// ErrActorContextMissing indicates the metaldocs.actor_id GUC was not set on the
// current transaction. Callers must SET LOCAL metaldocs.actor_id = '<userID>'
// before invoking authz functions.
var ErrActorContextMissing = errors.New("authz: metaldocs.actor_id GUC not set on transaction")

// ErrTenantContextMissing indicates the metaldocs.tenant_id GUC was not set on
// the current transaction.
var ErrTenantContextMissing = errors.New("authz: metaldocs.tenant_id GUC not set on transaction")

// MustActorID returns the metaldocs.actor_id GUC value for the given transaction.
// Returns ErrActorContextMissing if the GUC is unset or empty.
func MustActorID(ctx context.Context, tx *sql.Tx) (string, error) {
	var v string
	if err := tx.QueryRowContext(ctx, "SELECT current_setting('metaldocs.actor_id', true)").Scan(&v); err != nil {
		return "", fmt.Errorf("read actor_id GUC: %w", err)
	}
	if v == "" {
		return "", ErrActorContextMissing
	}
	return v, nil
}

// MustTenantID returns the metaldocs.tenant_id GUC value for the given transaction.
// Returns ErrTenantContextMissing if the GUC is unset or empty.
func MustTenantID(ctx context.Context, tx *sql.Tx) (string, error) {
	var v string
	if err := tx.QueryRowContext(ctx, "SELECT current_setting('metaldocs.tenant_id', true)").Scan(&v); err != nil {
		return "", fmt.Errorf("read tenant_id GUC: %w", err)
	}
	if v == "" {
		return "", ErrTenantContextMissing
	}
	return v, nil
}
