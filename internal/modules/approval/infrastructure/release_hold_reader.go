package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	approvaldomain "metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/platform/db"
)

// ReleaseHoldReaderPG is the approval-owned Postgres adapter for the
// ReleaseHoldReader published read-port (ADR 0085 Stage C, W2). Consumed by the
// jobs module's River periodic release_hold_reconciler — the only
// approval-owned surface it may call for stuck release generations; never raw
// SQL on public.release_generations.
type ReleaseHoldReaderPG struct {
	db *sql.DB
}

// NewReleaseHoldReaderPG builds the adapter over the approval pool. Wired at
// the composition root and injected into the jobs module's reconciler.
func NewReleaseHoldReaderPG(db *sql.DB) *ReleaseHoldReaderPG {
	return &ReleaseHoldReaderPG{db: db}
}

var _ approvaldomain.ReleaseHoldReader = (*ReleaseHoldReaderPG)(nil)

// releaseHoldStuckCorePredicate is the single source of truth for "is this
// release generation stuck in readiness hold right now" (mirrors
// slaOverdueCorePredicate's single-source discipline): both
// ListTenantsWithStuckHolds and ListStuckHolds reference this exact fragment,
// so "stuck" cannot drift between the enumeration and the per-tenant read.
//
// Binds `now` as $1 (referenced four times) and the threshold in SECONDS as $2.
// Every call site aliases public.release_generations `g` and joins its document
// `d`.
//
// Conjuncts 1-3 are the ADR's sweep definition (unreleased, old enough, not
// evaluated recently). Conjuncts 4-6 are what make the alert TRUE: without
// them the sweep would fire forever on generations the coordinator would only
// ever no-op on, and the honest signal would drown in permanent noise.
//
//  1. released_at IS NULL      — the release never happened.
//  2. created_at older than the threshold — a generation created seconds ago is
//     not stuck, it is in flight.
//  3. never evaluated, or last evaluated before the threshold — the "lost
//     fact / dead-lettered consumer" and "dead timer" cases respectively.
//  4. the document is still in a releasable status ('approved','scheduled') —
//     EvaluateRelease no-ops on anything else (domain/release.go), so such a
//     generation is not WAITING for anything and must not be alerted on.
//  5. no future plan date — a generation legitimately parked on
//     `awaiting_effective_date` with its timer armed months out is NOT stuck;
//     only a plan whose date has already passed points at a dead timer.
//  6. the generation is still the document's freeze head — a superseded
//     generation (document returned to draft and re-approved) is permanently
//     unreleasable by design; mirrors the isFreezeHeadTx conjunct.
const releaseHoldStuckCorePredicate = `g.released_at IS NULL
   AND g.created_at < $1::timestamptz - ($2 * interval '1 second')
   AND (g.last_evaluated_at IS NULL
        OR g.last_evaluated_at < $1::timestamptz - ($2 * interval '1 second'))
   AND d.status IN ('approved','scheduled')
   AND (d.planned_effective_from IS NULL OR d.planned_effective_from <= $1::timestamptz)
   AND NOT EXISTS (
       SELECT 1
         FROM public.release_generations newer
        WHERE newer.tenant_id       = g.tenant_id
          AND newer.document_id     = g.document_id
          AND newer.generation_seq  > g.generation_seq)`

// ListStuckHolds returns one tenant's stuck release generations, oldest first,
// capped at limit.
//
// Tenant scoping is an EXPLICIT g.tenant_id predicate (primary); the
// release_generations FORCE RLS policy keyed on the tx-local
// metaldocs.tenant_id GUC is the BACKSTOP. The caller must seed tenant identity
// (SeedTxTenant) before calling this port.
//
// Index note: with the tenant bound, the (tenant_id, last_evaluated_at)
// WHERE released_at IS NULL partial index (ix_release_generations_open,
// migration 0310 — created for exactly this sweep) is a prefix match.
func (r *ReleaseHoldReaderPG) ListStuckHolds(ctx context.Context, tx db.Tx, tenantID string, now time.Time, stuckAfter time.Duration, limit int) ([]approvaldomain.StuckReleaseHoldView, error) {
	if limit <= 0 {
		limit = 25
	}

	rows, err := tx.QueryContext(ctx, `
SELECT g.id::text, g.tenant_id::text, g.document_id::text, g.approval_instance_id::text,
       COALESCE(g.hold_reason, ''), COALESCE(g.hold_detail, ''),
       g.created_at, g.last_evaluated_at
  FROM public.release_generations g
  JOIN public.documents d ON d.id = g.document_id AND d.tenant_id = g.tenant_id
 WHERE g.tenant_id = $3::uuid
   AND `+releaseHoldStuckCorePredicate+`
 ORDER BY g.created_at ASC
 LIMIT $4`,
		now.UTC(), stuckAfterSeconds(stuckAfter), tenantID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list stuck release holds: query: %w", err)
	}
	defer rows.Close()

	var out []approvaldomain.StuckReleaseHoldView
	for rows.Next() {
		var v approvaldomain.StuckReleaseHoldView
		var lastEvaluated sql.NullTime
		if err := rows.Scan(&v.GenerationID, &v.TenantID, &v.DocumentID, &v.ApprovalInstanceID,
			&v.HoldReason, &v.HoldDetail, &v.CreatedAt, &lastEvaluated); err != nil {
			return nil, fmt.Errorf("list stuck release holds: scan: %w", err)
		}
		if lastEvaluated.Valid {
			t := lastEvaluated.Time
			v.LastEvaluatedAt = &t
		}
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list stuck release holds: rows: %w", err)
	}
	return out, nil
}

// ListTenantsWithStuckHolds returns the DISTINCT tenant_ids (as text) with at
// least one generation matching releaseHoldStuckCorePredicate. System-level,
// cross-tenant enumeration read: the caller runs it under authz.BypassSystem
// with NO tenant GUC seeded (mirrors ListTenantsWithOverdueStages).
func (r *ReleaseHoldReaderPG) ListTenantsWithStuckHolds(ctx context.Context, tx db.Tx, now time.Time, stuckAfter time.Duration) ([]string, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT DISTINCT g.tenant_id::text
  FROM public.release_generations g
  JOIN public.documents d ON d.id = g.document_id AND d.tenant_id = g.tenant_id
 WHERE `+releaseHoldStuckCorePredicate,
		now.UTC(), stuckAfterSeconds(stuckAfter),
	)
	if err != nil {
		return nil, fmt.Errorf("list tenants with stuck release holds: query: %w", err)
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var tenantID string
		if err := rows.Scan(&tenantID); err != nil {
			return nil, fmt.Errorf("list tenants with stuck release holds: scan: %w", err)
		}
		out = append(out, tenantID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list tenants with stuck release holds: rows: %w", err)
	}
	return out, nil
}

// stuckAfterSeconds converts the threshold to whole seconds for the interval
// arithmetic. A non-positive threshold would make every open generation "stuck"
// the instant it is created, so it fails closed onto the caller-independent
// floor of one second.
func stuckAfterSeconds(stuckAfter time.Duration) int64 {
	if stuckAfter <= 0 {
		return 1
	}
	return int64(stuckAfter / time.Second)
}
