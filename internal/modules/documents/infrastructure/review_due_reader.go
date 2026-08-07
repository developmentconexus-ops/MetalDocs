package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	documentsdomain "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/platform/db"
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

// dueCorePredicate is the single source of truth for "is this document due
// for periodic review right now" (validation-contract.md §4/D3 single-source
// requirement): published, has a review_due_at that has arrived, and is
// currently effective (already started, not yet expired). Every SQL site that
// needs the due-core eligibility check — ListDueForReview,
// ListTenantsWithDueReviews here, and ReviewSurfaceWriterPG.MarkSurfaced's
// eligibility clause — references this exact fragment so the definition of
// "due" cannot drift between reader and writer. It takes one positional
// parameter: the `now` instant, referenced 3 times as $1 — every consumer
// binds `now` as its FIRST query argument regardless of what other
// predicates (e.g. an explicit tenant_id = $2) are spliced around it, so
// this fragment's $1 numbering never needs to change.
const dueCorePredicate = `status = 'published'
   AND review_due_at IS NOT NULL
   AND review_due_at <= $1
   AND effective_from IS NOT NULL
   AND effective_from <= $1
   AND (effective_to IS NULL OR effective_to > $1)`

// ListDueForReview returns published, currently-effective documents whose
// review_due_at <= now, ordered by review_due_at ascending, capped at limit.
//
// Tenant scoping is enforced by an EXPLICIT tenant_id predicate in the WHERE
// (primary), matching every other tenant-scoped documents query in this
// package (e.g. active_instance_reader.go's `WHERE tenant_id = $1::uuid`).
// RLS (public.documents FORCE ROW LEVEL SECURITY + the tenant_isolation
// policy keyed on the tx-local metaldocs.tenant_id GUC,
// db/baseline/0001_current_schema.sql:4840) is the BACKSTOP — the caller is
// still responsible for seeding tenant identity on tx (M3) so the backstop is
// live, but isolation no longer depends solely on it.
//
// "Currently-effective" = effective_from <= now (already effective) AND
// (effective_to IS NULL OR effective_to > now) (not yet expired) — the F6.2
// review/expiry model (validation-contract.md §2).
func (r *ReviewDueReaderPG) ListDueForReview(ctx context.Context, tx db.Tx, tenantID string, now time.Time, limit int) ([]documentsdomain.ReviewDueView, error) {
	if limit <= 0 {
		limit = 25
	}

	rows, err := tx.QueryContext(ctx, `
SELECT id::text, tenant_id::text, code, name, status, review_due_at
  FROM public.documents
 WHERE tenant_id = $2::uuid
   AND `+dueCorePredicate+`
 ORDER BY review_due_at ASC
 LIMIT $3`,
		now, tenantID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list due for review: query: %w", err)
	}
	defer rows.Close()

	var out []documentsdomain.ReviewDueView
	for rows.Next() {
		var v documentsdomain.ReviewDueView
		var reviewDueAt time.Time
		var code sql.NullString // documents.code is nullable
		if err := rows.Scan(&v.ID, &v.TenantID, &code, &v.Name, &v.Status, &reviewDueAt); err != nil {
			return nil, fmt.Errorf("list due for review: scan: %w", err)
		}
		if code.Valid {
			v.Code = &code.String
		}
		v.ReviewDueAt = &reviewDueAt
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list due for review: rows: %w", err)
	}
	return out, nil
}

// ListTenantsWithDueReviews returns the DISTINCT tenant_ids (as text) that
// currently have at least one document matching the due-core predicate (the
// same eligibility ListDueForReview and MarkSurfaced use). This is a
// system-level, cross-tenant enumeration read: the caller runs it under
// authz.BypassSystem with NO tenant GUC seeded (mirrors
// stuck_instance_watchdog.listStuckInstances) so it can drive a
// per-tenant-seeded write loop — validation-contract.md §4.2/§4.3. Reads are
// safe unseeded; only the subsequent MarkSurfaced write per tenant must be
// seeded.
func (r *ReviewDueReaderPG) ListTenantsWithDueReviews(ctx context.Context, tx db.Tx, now time.Time) ([]string, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT DISTINCT tenant_id::text
  FROM public.documents
 WHERE `+dueCorePredicate,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("list tenants with due reviews: query: %w", err)
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var tenantID string
		if err := rows.Scan(&tenantID); err != nil {
			return nil, fmt.Errorf("list tenants with due reviews: scan: %w", err)
		}
		out = append(out, tenantID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list tenants with due reviews: rows: %w", err)
	}
	return out, nil
}
