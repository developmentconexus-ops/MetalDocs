package domain

import (
	"context"
	"database/sql"
	"time"
)

// OverdueStageView is the approval-owned projection of a stage instance whose
// due_at has passed (F8, spec.md §4). Minimal by design — enough for the
// jobs module's River SLA surfacer to identify and log/count what it found;
// mirrors documents.ReviewDueView's shape and purpose but for the DISTINCT
// per-approval-stage SLA concept (not the per-document periodic eQMS review
// document_review_surfacer already covers — F8 Interview #2).
type OverdueStageView struct {
	// StageInstanceID is the approval_stage_instances.id.
	StageInstanceID string
	// ApprovalInstanceID is the owning approval_instances.id.
	ApprovalInstanceID string
	// TenantID is the owning tenant. Present for evidence/logging even though
	// the read itself is tenant-scoped by an explicit predicate.
	TenantID string
	// DueAt is the instant this stage became overdue. Never nil for rows this
	// port returns (the query requires due_at IS NOT NULL AND due_at <= now).
	DueAt *time.Time
}

// SurfacedStage is the approval-owned projection of a stage instance newly
// marked surfaced by one MarkStagesOverdueSurfaced run (F8).
//
// It carries the subject and the eligible actors so the `jobs` surfacer can
// enqueue an overdue notification WITHOUT reading approval's tables: resolving
// who is eligible stays inside the module that owns eligibility (invariant 6).
type SurfacedStage struct {
	// StageInstanceID is the approval_stage_instances.id.
	StageInstanceID string
	// TenantID is the owning tenant, carried so the caller can build a
	// single-tenant event without a second lookup.
	TenantID string
	// SubjectKind is "document" or "template" (the M3 kernel discrimination).
	SubjectKind string
	// SubjectID is the subject's key.
	SubjectID string
	// EligibleActorIDs are the people this stage is waiting on — the snapshot
	// taken at submit, not a live recomputation.
	EligibleActorIDs []string
	// DueAt is the due_at value that made this stage eligible (never nil for
	// rows this port returns).
	DueAt *time.Time
}

// SLAOverdueReader is the approval-owned published read-port for the F8 SLA
// surfacer (the cross-module boundary the `jobs` module's River periodic
// approval_sla_surfacer calls — never raw SQL against
// public.approval_stage_instances). The interface lives in the owner's
// domain package; the Postgres adapter lives in the approval infrastructure
// repository and is wired at the composition root — mirrors
// documents.ReviewDueReader (review_due_port.go). Deliberately a DISTINCT
// port/mechanism from documents' review-due surfacing: an approval stage's
// due_at (this instance's SLA clock, activated per-stage) is a different
// concept from a published document's periodic review_due_at (F8 Interview
// #2) — they must never be conflated onto one job or one column.
type SLAOverdueReader interface {
	// ListOverdueStages returns active stage instances whose due_at <= now,
	// not yet surfaced for their current due_at cycle (sla_surfaced_at IS
	// NULL OR sla_surfaced_at < due_at), ordered by due_at ascending, capped
	// at limit. Runs in the caller-provided tx: tenant scoping is enforced by
	// an EXPLICIT tenant_id predicate in the query (primary); the
	// approval_stage_instances RLS FORCE policy on the tx's
	// metaldocs.tenant_id GUC is the backstop. The caller is responsible for
	// seeding tenant identity (SeedTxTenant) before calling this port.
	ListOverdueStages(ctx context.Context, tx *sql.Tx, tenantID string, now time.Time, limit int) ([]OverdueStageView, error)

	// ListTenantsWithOverdueStages returns the DISTINCT tenant_ids (as text)
	// that currently have at least one overdue, not-yet-surfaced active stage
	// instance (the same eligibility as ListOverdueStages). Called by the
	// jobs surfacer under authz.BypassSystem with NO tenant GUC seeded —
	// mirrors documents.ReviewDueReader.ListTenantsWithDueReviews.
	ListTenantsWithOverdueStages(ctx context.Context, tx *sql.Tx, now time.Time) ([]string, error)
}

// NoopSLAOverdueReader is the fail-closed default: it reports no overdue
// stages and no tenants with overdue stages.
type NoopSLAOverdueReader struct{}

func (NoopSLAOverdueReader) ListOverdueStages(context.Context, *sql.Tx, string, time.Time, int) ([]OverdueStageView, error) {
	return nil, nil
}

func (NoopSLAOverdueReader) ListTenantsWithOverdueStages(context.Context, *sql.Tx, time.Time) ([]string, error) {
	return nil, nil
}

// SLASurfaceWriter is the approval-owned published write-port for idempotent
// overdue-stage surfacing (F8, the cross-module boundary). Mirrors
// documents.ReviewSurfaceWriter.
type SLASurfaceWriter interface {
	// MarkStagesOverdueSurfaced idempotently marks every active, overdue
	// stage instance as surfaced for its current due_at cycle by setting
	// sla_surfaced_at = now (the caller-provided now). A stage already
	// surfaced for its current cycle (sla_surfaced_at >= due_at) is left
	// untouched — running this twice for the same cycle surfaces nothing the
	// second time. Alert-only (ADR 0068): this NEVER mutates status or
	// due_at, only the sla_surfaced_at marker. The returned rows carry the
	// subject (SubjectKind/SubjectID) and the eligible actors
	// (EligibleActorIDs) of each newly surfaced stage, so the caller can
	// enqueue an overdue notification without a second read against
	// approval's tables.
	//
	// Runs in the caller-provided tx. Tenant scoping mirrors
	// ListOverdueStages: an EXPLICIT tenant_id predicate in the UPDATE's
	// WHERE (primary); RLS FORCE on the seeded GUC is the backstop. A
	// scheduler caller must set authz.BypassSystem (requires
	// authz.WithBackgroundBypass on ctx) before calling this port — mirrors
	// ReviewSurfaceWriter.MarkSurfaced.
	MarkStagesOverdueSurfaced(ctx context.Context, tx *sql.Tx, tenantID string, now time.Time) ([]SurfacedStage, error)
}

// NoopSLASurfaceWriter is the fail-closed default: it surfaces nothing.
type NoopSLASurfaceWriter struct{}

func (NoopSLASurfaceWriter) MarkStagesOverdueSurfaced(context.Context, *sql.Tx, string, time.Time) ([]SurfacedStage, error) {
	return nil, nil
}
