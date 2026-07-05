package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	documentsdomain "metaldocs/internal/modules/documents/domain"
)

// ReviewDueReaderPG is the documents-owned Postgres adapter for the
// ReviewDueReader published read-port (M6 F6.2 §4.1). Consumed by the jobs
// module's River periodic surfacer (T4) — the only documents-owned surface it
// may call for review-due documents; never raw SQL on public.documents.
type ReviewDueReaderPG struct {
	db *sql.DB
}

// NewReviewDueReaderPG builds the adapter over the documents pool. Wired at
// the composition root and injected into the jobs module's surfacer.
func NewReviewDueReaderPG(db *sql.DB) *ReviewDueReaderPG {
	return &ReviewDueReaderPG{db: db}
}

var _ documentsdomain.ReviewDueReader = (*ReviewDueReaderPG)(nil)

// ListDueForReview returns published, currently-effective documents whose
// review_due_at <= now, ordered by review_due_at ascending, capped at limit.
//
// Tenant scoping is enforced by RLS (public.documents FORCE ROW LEVEL SECURITY
// + the tenant_isolation policy keyed on the tx-local metaldocs.tenant_id
// GUC, db/baseline/0001_current_schema.sql:4840) — this query carries no
// explicit tenant_id predicate, mirroring the read-port's contract that the
// caller has already seeded tenant identity on tx (M3 backstop). This
// intentionally differs from same-module sibling queries in this package
// (e.g. loadDocumentArea) that ALSO filter tenant_id explicitly as a
// belt-and-suspenders local check; ListDueForReview is the one published
// cross-module read surface for this projection, and the T3 contract's fixed
// signature (ctx, tx, now, limit) carries no tenantID parameter to filter on
// — RLS is the sole enforcement point here, by design.
//
// "Currently-effective" = effective_from <= now (already effective) AND
// (effective_to IS NULL OR effective_to > now) (not yet expired) — the F6.2
// review/expiry model (validation-contract.md §2).
func (r *ReviewDueReaderPG) ListDueForReview(ctx context.Context, tx *sql.Tx, now time.Time, limit int) ([]documentsdomain.ReviewDueView, error) {
	if limit <= 0 {
		limit = 25
	}

	rows, err := tx.QueryContext(ctx, `
SELECT id::text, tenant_id::text, code, name, status, review_due_at
  FROM public.documents
 WHERE status = 'published'
   AND review_due_at IS NOT NULL
   AND review_due_at <= $1
   AND effective_from IS NOT NULL
   AND effective_from <= $1
   AND (effective_to IS NULL OR effective_to > $1)
 ORDER BY review_due_at ASC
 LIMIT $2`,
		now, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list due for review: query: %w", err)
	}
	defer rows.Close()

	var out []documentsdomain.ReviewDueView
	for rows.Next() {
		var v documentsdomain.ReviewDueView
		var reviewDueAt time.Time
		if err := rows.Scan(&v.ID, &v.TenantID, &v.Code, &v.Name, &v.Status, &reviewDueAt); err != nil {
			return nil, fmt.Errorf("list due for review: scan: %w", err)
		}
		v.ReviewDueAt = &reviewDueAt
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list due for review: rows: %w", err)
	}
	return out, nil
}
