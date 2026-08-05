package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	approvaldomain "metaldocs/internal/modules/approval/domain"
)

// SLASurfaceWriterPG is the approval-owned Postgres adapter for the
// SLASurfaceWriter published write-port (F8, spec.md §4/W4). Consumed by the
// jobs module's River periodic approval_sla_surfacer — the only
// approval-owned surface it may call to mark overdue stage instances
// surfaced; never raw SQL on public.approval_stage_instances.
type SLASurfaceWriterPG struct {
	db *sql.DB
}

// NewSLASurfaceWriterPG builds the adapter over the approval pool. Wired at
// the composition root and injected into the jobs module's SLA surfacer.
func NewSLASurfaceWriterPG(db *sql.DB) *SLASurfaceWriterPG {
	return &SLASurfaceWriterPG{db: db}
}

var _ approvaldomain.SLASurfaceWriter = (*SLASurfaceWriterPG)(nil)

// MarkStagesOverdueSurfaced idempotently marks every active, overdue stage
// instance as surfaced for its current due_at cycle by setting
// sla_surfaced_at = now. Alert-only (ADR 0068): this NEVER touches status or
// due_at, only the sla_surfaced_at marker — mirrors
// stuck_instance_watchdog's alert-only contract and
// documents.ReviewSurfaceWriterPG.MarkSurfaced's idempotency shape.
//
// The idempotency guard is the last predicate in slaOverdueCorePredicate:
// (sla_surfaced_at IS NULL OR sla_surfaced_at < due_at). Once a stage is
// flagged for the current cycle, a second run in the same cycle flips 0 rows
// for that stage.
//
// Tenant scoping is enforced by an EXPLICIT tenant_id predicate via the join
// to approval_instances (approval_stage_instances has no tenant_id column of
// its own — mirrors ListOverdueStages and UpdateStageStatus). RLS is the
// backstop. Per the same contract as documents.ReviewSurfaceWriterPG, the
// caller MUST seed exactly one tenant on the tx (authz.BypassSystem then
// authz.SeedTxTenant(tenantID), under authz.WithBackgroundBypass) before
// calling this method — this adapter issues no authz calls itself.
func (w *SLASurfaceWriterPG) MarkStagesOverdueSurfaced(ctx context.Context, tx *sql.Tx, tenantID string, now time.Time) ([]approvaldomain.SurfacedStage, error) {
	rows, err := tx.QueryContext(ctx, `
WITH marked AS (
  UPDATE public.approval_stage_instances asi
     SET sla_surfaced_at = $1
    FROM public.approval_instances ai
   WHERE asi.approval_instance_id = ai.id
     AND ai.tenant_id = $2::uuid
     AND `+slaOverdueCorePredicate+`
  RETURNING asi.id, asi.approval_instance_id, asi.due_at, asi.eligible_actor_ids
)
SELECT m.id::text, ai.tenant_id::text, ai.subject_kind, ai.subject_key,
       m.eligible_actor_ids, m.due_at
  FROM marked m
  JOIN public.approval_instances ai ON ai.id = m.approval_instance_id`,
		now, tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("mark stages overdue surfaced: update: %w", err)
	}
	defer rows.Close()

	var out []approvaldomain.SurfacedStage
	for rows.Next() {
		var s approvaldomain.SurfacedStage
		var dueAt time.Time
		var eligibleJSON []byte
		if err := rows.Scan(&s.StageInstanceID, &s.TenantID, &s.SubjectKind, &s.SubjectID, &eligibleJSON, &dueAt); err != nil {
			return nil, fmt.Errorf("mark stages overdue surfaced: scan: %w", err)
		}
		if len(eligibleJSON) > 0 {
			if err := json.Unmarshal(eligibleJSON, &s.EligibleActorIDs); err != nil {
				return nil, fmt.Errorf("mark stages overdue surfaced: unmarshal eligible_actor_ids for stage %s: %w", s.StageInstanceID, err)
			}
		}
		s.DueAt = &dueAt
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("mark stages overdue surfaced: rows: %w", err)
	}
	return out, nil
}
