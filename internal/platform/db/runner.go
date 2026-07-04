package db

import (
	"context"
	"database/sql"
	"fmt"

	platformtenant "metaldocs/internal/platform/tenant"
)

// TxRunner owns the begin/commit/rollback lifecycle of a database transaction
// so application services depend on this port instead of the concrete *sql.DB
// connection pool. Do begins a transaction, invokes fn with it, commits when fn
// returns nil, and rolls back (best-effort) when fn returns an error or panics.
//
// The callback receives the live *sql.Tx by design. Exposing the concrete
// transaction here is a deliberate, bounded concession: the authorization layer
// (authz.Require / authz.SeedTxIdentity) and the catalogue readers are keyed on
// *sql.Tx, so the unit of work runs against it directly. What this port removes
// is the application layer's dependency on the *sql.DB pool and its ownership of
// commit/rollback — those now live in infrastructure, and a service can no
// longer leak a transaction or forget to commit.
type TxRunner interface {
	Do(ctx context.Context, fn func(tx *sql.Tx) error) error
	// DoReadOnly begins a READ ONLY transaction for pure read paths. The flag is
	// a write-guard that documents intent and prevents accidental DML.
	//
	// MUST NOT be used for any path that calls authz.Require: the F8 bypass audit
	// INSERTs on the system_admin/bypass short-circuit (ADR 0022 Phase 11), which
	// a READ ONLY tx rejects. authz.Require enforces this with a read-only-tx
	// guard (authz.ErrReadOnlyTx). SeedTxIdentity's SET LOCAL GUCs are RO-safe,
	// but the Require grant path is not — use Do for authz-gated reads.
	DoReadOnly(ctx context.Context, fn func(tx *sql.Tx) error) error
}

// sqlTxRunner is the production TxRunner backed by *sql.DB.
type sqlTxRunner struct {
	db *sql.DB
}

// NewTxRunner returns the production TxRunner backed by database. It panics on a
// nil pool because that can only be a wiring error, never a runtime condition.
func NewTxRunner(database *sql.DB) TxRunner {
	if database == nil {
		panic("db: NewTxRunner requires a non-nil *sql.DB")
	}
	return &sqlTxRunner{db: database}
}

// do is the shared implementation for Do and DoReadOnly. opts is passed directly
// to BeginTx; nil opts produce a default (read-write) transaction.
func (r *sqlTxRunner) do(ctx context.Context, opts *sql.TxOptions, fn func(tx *sql.Tx) error) (err error) {
	tx, beginErr := r.db.BeginTx(ctx, opts)
	if beginErr != nil {
		return fmt.Errorf("db: begin tx: %w", beginErr)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()
	if seedErr := seedTxIdentityFromContext(ctx, tx); seedErr != nil {
		_ = tx.Rollback()
		return fmt.Errorf("db: seed tx identity: %w", seedErr)
	}
	if err = fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("db: commit tx: %w", err)
	}
	return nil
}

// seedTxIdentityFromContext is the TxRunner chokepoint auto-seed (F3.1 /
// M3 tenancy-chokepoint contract §1.2). It reads the tenant+actor identity
// carried on ctx by the platform/tenant package (set once by the API authn
// middleware) and, ONLY when BOTH are present and non-empty, seeds
// metaldocs.tenant_id + metaldocs.actor_id as tx-local set_config calls —
// engaging the FORCE ROW LEVEL SECURITY backstop for this transaction.
//
// When either is absent (system/janitor paths with no request identity),
// this is a deliberate no-op: the tx runs GUC-unset, which is the existing
// NULL-permissive RLS escape hatch and MUST NOT be disturbed.
//
// This intentionally duplicates authz.SeedTxIdentity's SQL rather than
// calling it: internal/platform/db MUST NOT import internal/modules/iam
// (module-boundary rule — platform packages never depend on module
// packages). The seed is a SET LOCAL config write only, issued before fn
// runs and before any lock fn might take (H-PRE-1) — it is not an
// authz-recording read and no authz.Require is added here.
func seedTxIdentityFromContext(ctx context.Context, tx *sql.Tx) error {
	tenantID, tenantErr := platformtenant.FromContext(ctx)
	if tenantErr != nil {
		return nil // no-op: identity-less ctx, NULL-permissive RLS applies
	}
	actorID, actorErr := platformtenant.ActorFromContext(ctx)
	if actorErr != nil {
		return nil // no-op: identity-less ctx, NULL-permissive RLS applies
	}

	if _, err := tx.ExecContext(ctx, `
SELECT
	set_config('metaldocs.tenant_id', $1, true),
	set_config('metaldocs.actor_id', $2, true)
`, tenantID, actorID); err != nil {
		return err
	}
	return nil
}

// Do begins a transaction, runs fn, and finalizes it. A non-nil fn error rolls
// back and is returned unwrapped so callers retain errors.Is/As on domain
// sentinels. A panic inside fn rolls back and re-panics.
func (r *sqlTxRunner) Do(ctx context.Context, fn func(tx *sql.Tx) error) error {
	return r.do(ctx, nil, fn)
}

// DoReadOnly begins a READ ONLY transaction, runs fn, and finalizes it. A non-nil
// fn error rolls back and is returned unwrapped so callers retain errors.Is/As on
// domain sentinels. A panic inside fn rolls back and re-panics.
func (r *sqlTxRunner) DoReadOnly(ctx context.Context, fn func(tx *sql.Tx) error) error {
	return r.do(ctx, &sql.TxOptions{ReadOnly: true}, fn)
}
