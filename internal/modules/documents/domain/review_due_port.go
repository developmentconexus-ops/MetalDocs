package domain

import (
	"context"
	"time"

	"metaldocs/internal/platform/db"
)

// ReviewDueView is the documents-owned projection of a document due for
// periodic review (M6 F6.2 §4.1). Minimal by design — only what the jobs
// module's River surfacer (T4) and its evidence need: enough to identify the
// document and report why it surfaced.
type ReviewDueView struct {
	// ID is the document id.
	ID string
	// TenantID is the owning tenant. Present for evidence/logging even though
	// the read itself is tenant-scoped by the caller's tx GUC (RLS), not by an
	// explicit predicate in the query.
	TenantID string
	// Code is the document's business code (public.documents.code, NULLABLE —
	// no NOT NULL constraint on the column, db/baseline/0001_current_schema.sql
	// documents.code). nil when the row has no code assigned.
	Code *string
	// Name is the document's display name.
	Name string
	// Status is the document's status at read time (owner vocabulary,
	// DocumentStatus values). Always "published" for rows this port returns.
	Status string
	// ReviewDueAt is the instant the document became (or will become) due for
	// review. Never nil for rows this port returns (the query requires
	// review_due_at IS NOT NULL).
	ReviewDueAt *time.Time
}

// ReviewDueReader is the documents-owned published read-port for periodic
// review surfacing (M6 F6.2 §4.1, the cross-module boundary). The `jobs`
// module's River periodic surfacer calls THIS port — never raw SQL against
// public.documents — to find documents due for review. The interface lives in
// the owner's domain package; the Postgres adapter lives in the documents
// repository and is wired at the composition root (mirrors ActiveInstanceReader,
// active_instance_port.go).
type ReviewDueReader interface {
	// ListDueForReview returns published, currently-effective documents whose
	// review_due_at <= now, ordered by review_due_at ascending, capped at
	// limit. Runs in the caller-provided tx: tenant scoping is enforced by an
	// EXPLICIT tenant_id predicate in the query (primary), matching the
	// module convention for tenant-scoped documents queries (e.g.
	// active_instance_reader.go's `WHERE tenant_id = $1::uuid`); the
	// public.documents RLS FORCE policy on the tx's metaldocs.tenant_id GUC is
	// the backstop. The caller is still responsible for seeding tenant
	// identity (SeedTxIdentity/SeedTxTenant, M3) before calling this port, so
	// the RLS backstop is live.
	ListDueForReview(ctx context.Context, tx db.Tx, tenantID string, now time.Time, limit int) ([]ReviewDueView, error)

	// ListTenantsWithDueReviews returns the DISTINCT tenant_ids (as text) that
	// currently have at least one review-due document (the same "due-core"
	// eligibility as ListDueForReview). This is a system-level, cross-tenant
	// enumeration READ — called by the jobs surfacer under authz.BypassSystem
	// with NO tenant GUC seeded, exactly like
	// stuck_instance_watchdog.listStuckInstances. It exists so the surfacer
	// can iterate tenants and seed each one (authz.SeedTxTenant) before its
	// tenant-scoped MarkSurfaced write, per validation-contract.md §4.2/§4.3 —
	// reads are safe unseeded; only the write must be seeded.
	ListTenantsWithDueReviews(ctx context.Context, tx db.Tx, now time.Time) ([]string, error)
}

// NoopReviewDueReader is the fail-closed default: it reports no due documents
// and no tenants with due documents. Used as the nil-guard default at wiring
// and in unit tests that do not exercise the review-due surfacer.
type NoopReviewDueReader struct{}

func (NoopReviewDueReader) ListDueForReview(context.Context, db.Tx, string, time.Time, int) ([]ReviewDueView, error) {
	return nil, nil
}

func (NoopReviewDueReader) ListTenantsWithDueReviews(context.Context, db.Tx, time.Time) ([]string, error) {
	return nil, nil
}
