package domain

import (
	"context"
	"time"

	"metaldocs/internal/platform/db"
)

// SurfacedDoc is the documents-owned projection of a document that was newly
// marked surfaced by one MarkSurfaced run (M6 F6.2 §4.2). Minimal by design —
// enough for the jobs module's River surfacer (T4) to log/count what it
// touched this tick.
type SurfacedDoc struct {
	// ID is the document id.
	ID string
	// ReviewDueAt is the review_due_at value that made this document eligible
	// (never nil for rows this port returns).
	ReviewDueAt *time.Time
}

// ReviewSurfaceWriter is the documents-owned published write-port for
// idempotent review-due surfacing (M6 F6.2 §4.2, the cross-module boundary).
// The `jobs` module's River periodic surfacer calls THIS port — never raw SQL
// against public.documents — to mark due documents surfaced. The interface
// lives in the owner's domain package; the Postgres adapter lives in the
// documents repository and is wired at the composition root (mirrors
// ReviewDueReader, review_due_port.go).
type ReviewSurfaceWriter interface {
	// MarkSurfaced idempotently marks every published, currently-effective,
	// review-due document as surfaced for its current review_due_at cycle by
	// setting review_surfaced_at = now (the caller-provided now). A document
	// already surfaced for its current cycle (review_surfaced_at >=
	// review_due_at) is left untouched — running MarkSurfaced twice for the
	// same cycle surfaces nothing the second time. Advancing review_due_at
	// past the stored review_surfaced_at (e.g. via mark-reviewed) makes the
	// document eligible again on the next run.
	//
	// Runs in the caller-provided tx. Tenant scoping mirrors ListDueForReview:
	// enforced by an EXPLICIT tenant_id predicate in the UPDATE's WHERE
	// (primary), matching the module convention for tenant-scoped documents
	// queries; the public.documents RLS FORCE policy on the tx's
	// metaldocs.tenant_id GUC is the backstop. The caller still seeds the
	// tenant GUC (SeedTxTenant) so the RLS backstop is live. The
	// documents/UPDATE DB tripwire (enforce_capability_asserted) still fires
	// for this UPDATE; a scheduler caller must set the scheduler bypass
	// (authz.BypassSystem, requires authz.WithBackgroundBypass on ctx) before
	// calling this port — MarkSurfaced does not set the bypass itself, mirroring
	// the existing janitors' own-tx bypass calls rather than hiding it here.
	MarkSurfaced(ctx context.Context, tx db.Tx, tenantID string, now time.Time) ([]SurfacedDoc, error)
}

// NoopReviewSurfaceWriter is the fail-closed default: it surfaces nothing.
// Used as the nil-guard default at wiring and in unit tests that do not
// exercise the review-due surfacer.
type NoopReviewSurfaceWriter struct{}

// MarkSurfaced surfaces nothing and returns no rows. The Noop default
// performs no write.
func (NoopReviewSurfaceWriter) MarkSurfaced(context.Context, db.Tx, string, time.Time) ([]SurfacedDoc, error) {
	return nil, nil
}
