package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	documentsdomain "metaldocs/internal/modules/documents/domain"
)

// ReviewSurfaceWriterPG is the documents-owned Postgres adapter for the
// ReviewSurfaceWriter published write-port (M6 F6.2 §4.2). Consumed by the
// jobs module's River periodic surfacer (T4) — the only documents-owned
// surface it may call to mark review-due documents surfaced; never raw SQL on
// public.documents.
type ReviewSurfaceWriterPG struct {
	db *sql.DB
}

// NewReviewSurfaceWriterPG builds the adapter over the documents pool. Wired
// at the composition root and injected into the jobs module's surfacer.
func NewReviewSurfaceWriterPG(db *sql.DB) *ReviewSurfaceWriterPG {
	return &ReviewSurfaceWriterPG{db: db}
}

var _ documentsdomain.ReviewSurfaceWriter = (*ReviewSurfaceWriterPG)(nil)

// MarkSurfaced idempotently marks every published, currently-effective,
// review-due document as surfaced for its current review_due_at cycle.
//
// The idempotency guard is the last WHERE predicate:
// (review_surfaced_at IS NULL OR review_surfaced_at < review_due_at). Once a
// document is flagged for the current cycle (review_surfaced_at >=
// review_due_at, since this UPDATE sets it to `now` and the eligibility
// predicate already required review_due_at <= now), a second run in the same
// cycle flips 0 rows for that document. When review_due_at is later advanced
// (mark-reviewed), review_surfaced_at < review_due_at becomes true again and
// the document re-surfaces on its next due cycle.
//
// Tenant scoping is enforced by RLS (public.documents FORCE ROW LEVEL
// SECURITY + the tenant_isolation policy keyed on the tx-local
// metaldocs.tenant_id GUC, db/baseline/0001_current_schema.sql:4840) exactly
// as ListDueForReview — no explicit tenant_id predicate. The scheduler
// surfacer runs this cross-tenant with the tenant GUC left unset, which the
// tenant_isolation policy's NULLIF(...) IS NULL branch treats as "all
// tenants" (mirrors stuck_instance_watchdog.listStuckInstances).
//
// The documents/UPDATE DB tripwire (enforce_capability_asserted) still fires
// for this UPDATE; the caller is responsible for the scheduler bypass
// (authz.BypassSystem under authz.WithBackgroundBypass) before calling this
// port — this adapter issues no authz calls itself.
func (w *ReviewSurfaceWriterPG) MarkSurfaced(ctx context.Context, tx *sql.Tx, now time.Time) ([]documentsdomain.SurfacedDoc, error) {
	rows, err := tx.QueryContext(ctx, `
UPDATE public.documents
   SET review_surfaced_at = $1
 WHERE review_due_at IS NOT NULL AND review_due_at <= $1
   AND status = 'published'
   AND effective_from IS NOT NULL AND effective_from <= $1
   AND (effective_to IS NULL OR effective_to > $1)
   AND (review_surfaced_at IS NULL OR review_surfaced_at < review_due_at)
RETURNING id::text, review_due_at`,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("mark surfaced: update: %w", err)
	}
	defer rows.Close()

	var out []documentsdomain.SurfacedDoc
	for rows.Next() {
		var d documentsdomain.SurfacedDoc
		var reviewDueAt time.Time
		if err := rows.Scan(&d.ID, &reviewDueAt); err != nil {
			return nil, fmt.Errorf("mark surfaced: scan: %w", err)
		}
		d.ReviewDueAt = &reviewDueAt
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("mark surfaced: rows: %w", err)
	}
	return out, nil
}
