// Package fanoutinfra hosts the render fanout module's transaction-lifecycle
// infrastructure. Owning a *sql.DB transaction (BeginTx/Commit/Rollback) is
// only permitted in specific layers (infrastructure/repository/application/
// platform, per tools/cilint's txownership rule) — the fanout package itself
// is not one of them. This mirrors the same seeded-tx shape already used
// elsewhere in the codebase for background/job writes (e.g.
// internal/modules/notifications/infrastructure/fanout_worker.go,
// internal/platform/worker/pdf_job_runner.go): BeginTx, seed the tenant GUC,
// run the caller's work, commit/rollback.
package fanoutinfra

import (
	"context"
	"database/sql"
	"fmt"

	"metaldocs/internal/modules/iam/authz"
)

// RunSeededTx begins a transaction on db, seeds it with tenantID via
// authz.SeedTxTenant, runs fn, and commits on success (rolling back on any
// error). tenantID is taken as an explicit parameter rather than read from
// ctx: this helper is used from background/job code paths, which are not
// HTTP-shaped and carry no request-scoped tenant identity on ctx.
//
// label prefixes returned error messages so callers keep their existing
// error-naming convention (e.g. "pdf_dispatch_outbox: begin tx: ...").
func RunSeededTx(ctx context.Context, db *sql.DB, label, tenantID string, fn func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: begin tx: %w", label, err)
	}
	if err := authz.SeedTxTenant(ctx, tx, tenantID); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("%s: seed tenant: %w", label, err)
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%s: commit: %w", label, err)
	}
	return nil
}
