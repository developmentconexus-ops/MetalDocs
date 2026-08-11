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
//
// DoReadOnly (a READ ONLY *sql.Tx variant for pure-read paths) was deleted in
// A5.2 (Lane E, issue #92): every production call site either needed the
// read-write path already (authz.Require's F8 bypass audit INSERTs reject a
// READ ONLY tx — ADR 0022 Phase 11, H-PRE-1) or gained nothing from the
// write-guard. G1 (commit 817abd59) had already migrated every caller that
// mixed DoReadOnly with authz.Require onto Do; this makes that class of bug
// unrepresentable instead of merely lint-guarded (scripts/api-lint's
// checkAuthzRequireRWTx guard is deleted alongside it — the call it detected
// can no longer be written). If a genuine pure-read, non-authz path wants a
// write-guard again, reintroduce it deliberately with the same MUST-NOT-mix
// docs, not by reverting this comment.
type TxRunner interface {
	Do(ctx context.Context, fn func(tx *sql.Tx) error) error
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

// do is the shared implementation behind Do. opts is passed directly to
// BeginTx; today Do always passes nil (default read-write transaction). The
// opts parameter stays (rather than being inlined away) because it is the
// seam a future genuine read-only path would reuse — see the TxRunner
// interface doc for why DoReadOnly itself was deleted, not just its callers
// migrated.
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
	tenantID, actorID, ok := ctxIdentity(ctx)
	if !ok {
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

// ctxIdentity extracts the tenant+actor identity pair from ctx. Absence is
// modeled as ok=false, not as an error: an identity-less ctx (system/janitor
// paths) legitimately gets no seed and NULL-permissive RLS applies.
func ctxIdentity(ctx context.Context) (tenantID, actorID string, ok bool) {
	tenantID, tenantErr := platformtenant.FromContext(ctx)
	if tenantErr != nil {
		return "", "", false
	}
	actorID, actorErr := platformtenant.ActorFromContext(ctx)
	if actorErr != nil {
		return "", "", false
	}
	return tenantID, actorID, true
}

// Do begins a transaction, runs fn, and finalizes it. A non-nil fn error rolls
// back and is returned unwrapped so callers retain errors.Is/As on domain
// sentinels. A panic inside fn rolls back and re-panics.
func (r *sqlTxRunner) Do(ctx context.Context, fn func(tx *sql.Tx) error) error {
	return r.do(ctx, nil, fn)
}
