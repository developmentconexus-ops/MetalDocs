package authz

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// ErrActorContextMissing indicates the metaldocs.actor_id GUC was not set on the
// current transaction. Callers must SET LOCAL metaldocs.actor_id = '<userID>'
// before invoking authz functions.
var ErrActorContextMissing = errors.New("authz: metaldocs.actor_id GUC not set on transaction")

// ErrTenantContextMissing indicates the metaldocs.tenant_id GUC was not set on
// the current transaction.
var ErrTenantContextMissing = errors.New("authz: metaldocs.tenant_id GUC not set on transaction")

// readSoftGUC reads a transaction-local custom GUC via current_setting(name, true)
// and normalizes the "not seeded" case to an empty string.
//
// PostgreSQL subtlety this exists to absorb: a *placeholder* custom GUC
// (metaldocs.*) that was never introduced on the connection makes
// current_setting(..., true) return SQL **NULL**, not ”. Scanning that NULL into
// a plain Go string fails at the driver ("converting NULL to string is
// unsupported") — which on a cold, never-seeded pooled connection turns a clean
// "identity missing" into an opaque driver error. Scanning into sql.NullString and
// collapsing NULL to "" gives every caller one unambiguous unset signal (NULL and
// ” both mean "not seeded"). This is the same idiom loadAssertedCaps already uses;
// centralizing it here removes the three-way split across the GUC readers.
//
// query is always a compile-time constant literal at the call site (no injection
// surface); it is passed through verbatim so the emitted SQL stays byte-identical
// to the historical statements the sqlmock tests match on.
func readSoftGUC(ctx context.Context, tx *sql.Tx, query string) (string, error) {
	var v sql.NullString
	if err := tx.QueryRowContext(ctx, query).Scan(&v); err != nil {
		return "", err
	}
	if !v.Valid {
		return "", nil
	}
	return v.String, nil
}

// MustActorID returns the metaldocs.actor_id GUC value for the given transaction.
// Returns ErrActorContextMissing if the GUC is unset (SQL NULL) or empty.
func MustActorID(ctx context.Context, tx *sql.Tx) (string, error) {
	v, err := readSoftGUC(ctx, tx, "SELECT current_setting('metaldocs.actor_id', true)")
	if err != nil {
		return "", fmt.Errorf("read actor_id GUC: %w", err)
	}
	if v == "" {
		return "", ErrActorContextMissing
	}
	return v, nil
}

// MustTenantID returns the metaldocs.tenant_id GUC value for the given transaction.
// Returns ErrTenantContextMissing if the GUC is unset (SQL NULL) or empty.
func MustTenantID(ctx context.Context, tx *sql.Tx) (string, error) {
	v, err := readSoftGUC(ctx, tx, "SELECT current_setting('metaldocs.tenant_id', true)")
	if err != nil {
		return "", fmt.Errorf("read tenant_id GUC: %w", err)
	}
	if v == "" {
		return "", ErrTenantContextMissing
	}
	return v, nil
}

// SeedTxIdentity sets the transaction-local actor and tenant GUCs required by
// Require. Both values must be non-empty.
func SeedTxIdentity(ctx context.Context, tx *sql.Tx, tenantID, actorID string) error {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return ErrTenantContextMissing
	}
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return ErrActorContextMissing
	}

	if _, err := tx.ExecContext(ctx, `
SELECT
	set_config('metaldocs.tenant_id', $1, true),
	set_config('metaldocs.actor_id', $2, true)
`, tenantID, actorID); err != nil {
		return fmt.Errorf("seed authz tx identity: %w", err)
	}
	return nil
}

// SeedTxTenant sets ONLY the transaction-local metaldocs.tenant_id GUC — the
// minimum seed that engages the FORCE ROW LEVEL SECURITY backstop (M3 F3.2 /
// validation-contract.md §2.1). It sets and requires NO actor: RLS reads
// only metaldocs.tenant_id, and async system work (worker/jobs processing
// txs) has no human actor to seed — SeedTxIdentity's actor-required
// contract is the wrong primitive for that seam. SeedTxTenant is additive
// and orthogonal to authz.BypassSystem: the RLS backstop (this) and the
// write-tripwire bypass (BypassSystem) are separate gates and both must
// still be called where each is needed.
//
// tenantID must be non-empty (trimmed); the GUC is seeded tx-local via
// set_config(..., true) — never session-level, so a pooled connection never
// leaks a tenant identity to the next borrower.
func SeedTxTenant(ctx context.Context, tx *sql.Tx, tenantID string) error {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return ErrTenantContextMissing
	}

	if _, err := tx.ExecContext(ctx, `SELECT set_config('metaldocs.tenant_id', $1, true)`, tenantID); err != nil {
		return fmt.Errorf("seed authz tx tenant: %w", err)
	}
	return nil
}
