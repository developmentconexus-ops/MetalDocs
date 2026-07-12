package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	approvaldomain "metaldocs/internal/modules/approval/domain"
)

// SLAOverdueReaderPG is the approval-owned Postgres adapter for the
// SLAOverdueReader published read-port (F8, spec.md §4/W4). Consumed by the
// jobs module's River periodic approval_sla_surfacer — the only
// approval-owned surface it may call for overdue stage instances; never raw
// SQL on public.approval_stage_instances.
type SLAOverdueReaderPG struct {
	db *sql.DB
}

// NewSLAOverdueReaderPG builds the adapter over the approval pool. Wired at
// the composition root and injected into the jobs module's SLA surfacer.
func NewSLAOverdueReaderPG(db *sql.DB) *SLAOverdueReaderPG {
	return &SLAOverdueReaderPG{db: db}
}

var _ approvaldomain.SLAOverdueReader = (*SLAOverdueReaderPG)(nil)

// slaOverdueCorePredicate is the single source of truth for "is this stage
// instance overdue and not yet surfaced for its current cycle right now"
// (mirrors documents' dueCorePredicate single-source discipline): the stage
// must be active, carry a due_at that has arrived, and not already be marked
// surfaced for THIS due_at (sla_surfaced_at IS NULL OR sla_surfaced_at <
// due_at). Both ListOverdueStages and MarkStagesOverdueSurfaced reference
// this exact fragment so "overdue" cannot drift between reader and writer.
// Binds `now` as $1 (referenced twice).
const slaOverdueCorePredicate = `status = 'active'
   AND due_at IS NOT NULL
   AND due_at <= $1
   AND (sla_surfaced_at IS NULL OR sla_surfaced_at < due_at)`

// ListOverdueStages returns active stage instances whose due_at <= now and
// not yet surfaced for their current cycle, ordered by due_at ascending,
// capped at limit.
//
// Tenant scoping is enforced by an EXPLICIT tenant_id predicate via a join to
// approval_instances (approval_stage_instances itself has no tenant_id
// column — tenancy is inherited through its parent approval_instances row,
// mirroring every other approval_stage_instances query in this package,
// e.g. UpdateStageStatus's `FROM approval_instances ai ... ai.tenant_id =
// $4`). RLS (approval_stage_instances FORCE ROW LEVEL SECURITY keyed on the
// tx-local metaldocs.tenant_id GUC) is the BACKSTOP; the caller is
// responsible for seeding tenant identity (SeedTxTenant) before calling this
// port.
func (r *SLAOverdueReaderPG) ListOverdueStages(ctx context.Context, tx *sql.Tx, tenantID string, now time.Time, limit int) ([]approvaldomain.OverdueStageView, error) {
	if limit <= 0 {
		limit = 25
	}

	rows, err := tx.QueryContext(ctx, `
SELECT asi.id::text, asi.approval_instance_id::text, ai.tenant_id::text, asi.due_at
  FROM public.approval_stage_instances asi
  JOIN public.approval_instances ai ON ai.id = asi.approval_instance_id
 WHERE ai.tenant_id = $2::uuid
   AND `+slaOverdueCorePredicate+`
 ORDER BY asi.due_at ASC
 LIMIT $3`,
		now, tenantID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list overdue stages: query: %w", err)
	}
	defer rows.Close()

	var out []approvaldomain.OverdueStageView
	for rows.Next() {
		var v approvaldomain.OverdueStageView
		var dueAt time.Time
		if err := rows.Scan(&v.StageInstanceID, &v.ApprovalInstanceID, &v.TenantID, &dueAt); err != nil {
			return nil, fmt.Errorf("list overdue stages: scan: %w", err)
		}
		v.DueAt = &dueAt
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list overdue stages: rows: %w", err)
	}
	return out, nil
}

// ListTenantsWithOverdueStages returns the DISTINCT tenant_ids (as text) that
// currently have at least one stage instance matching slaOverdueCorePredicate
// (the same eligibility ListOverdueStages and MarkStagesOverdueSurfaced use).
// System-level, cross-tenant enumeration read: the caller runs it under
// authz.BypassSystem with NO tenant GUC seeded (mirrors
// documents.ReviewDueReaderPG.ListTenantsWithDueReviews).
func (r *SLAOverdueReaderPG) ListTenantsWithOverdueStages(ctx context.Context, tx *sql.Tx, now time.Time) ([]string, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT DISTINCT ai.tenant_id::text
  FROM public.approval_stage_instances asi
  JOIN public.approval_instances ai ON ai.id = asi.approval_instance_id
 WHERE `+slaOverdueCorePredicate,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("list tenants with overdue stages: query: %w", err)
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var tenantID string
		if err := rows.Scan(&tenantID); err != nil {
			return nil, fmt.Errorf("list tenants with overdue stages: scan: %w", err)
		}
		out = append(out, tenantID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list tenants with overdue stages: rows: %w", err)
	}
	return out, nil
}
