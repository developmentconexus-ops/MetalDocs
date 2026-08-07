package domain

import (
	"context"
	"time"

	"metaldocs/internal/platform/db"
)

// StuckReleaseHoldView is the approval-owned projection of ONE release
// generation that has been sitting in readiness hold past the reconciliation
// threshold (ADR 0085 §"Observability, attribution, reconciliation": "a River
// periodic maintenance job sweeps generations stuck in readiness hold beyond a
// threshold — lost fact, dead-lettered consumer, dead timer — and alerts").
//
// Minimal by design: enough for the jobs module's reconciler to name the stuck
// generation in a structured alert. Mirrors OverdueStageView's shape and
// purpose for the DISTINCT release-hold concept.
type StuckReleaseHoldView struct {
	// GenerationID is the public.release_generations.id — the alert's resource.
	GenerationID string
	// TenantID is the owning tenant.
	TenantID string
	// DocumentID is the document whose release is stuck.
	DocumentID string
	// ApprovalInstanceID is the approval instance the generation was keyed on
	// (the causal identity an operator needs to investigate the hold).
	ApprovalInstanceID string
	// HoldReason is the recorded readiness-hold reason, or "" when the
	// generation was NEVER evaluated (last_evaluated_at IS NULL) — the "lost
	// fact / dead-lettered consumer" case, which carries no reason at all and
	// is the most severe of the three.
	HoldReason string
	// HoldDetail is the free-text detail recorded alongside the reason ("" when
	// absent).
	HoldDetail string
	// CreatedAt is when the generation was created; the alert's age is measured
	// from here.
	CreatedAt time.Time
	// LastEvaluatedAt is the last coordinator evaluation, nil when never
	// evaluated.
	LastEvaluatedAt *time.Time
}

// ReleaseHoldReader is the approval-owned published read-port for the ADR 0085
// reconciliation sweep (the cross-module boundary the `jobs` module's River
// periodic release_hold_reconciler calls — never raw SQL against
// public.release_generations). The interface lives in the owner's domain
// package; the Postgres adapter lives in the approval infrastructure package
// and is wired at the composition root — mirrors SLAOverdueReader
// (sla_port.go).
//
// Read-only by construction: the sweep is ALERT-ONLY (ADR 0068 posture, cited
// verbatim by ADR 0085). There is no writer twin to this port, because
// recovery from a stuck hold is always "re-run the idempotent evaluation" — a
// coordinator call, never a reconciler mutation.
type ReleaseHoldReader interface {
	// ListStuckHolds returns the tenant's release generations that are still
	// unreleased, older than stuckAfter, not evaluated within stuckAfter, and
	// still ACTIONABLE (see the adapter's core predicate), oldest first, capped
	// at limit. Runs in the caller-provided tx: tenant scoping is enforced by
	// an EXPLICIT tenant_id predicate (primary); the release_generations RLS
	// FORCE policy on the tx-local metaldocs.tenant_id GUC is the backstop. The
	// caller is responsible for seeding tenant identity (SeedTxTenant) before
	// calling this port.
	ListStuckHolds(ctx context.Context, tx db.Tx, tenantID string, now time.Time, stuckAfter time.Duration, limit int) ([]StuckReleaseHoldView, error)

	// ListTenantsWithStuckHolds returns the DISTINCT tenant_ids (as text) that
	// currently have at least one stuck hold (the same eligibility as
	// ListStuckHolds). Called by the jobs reconciler under authz.BypassSystem
	// with NO tenant GUC seeded — mirrors
	// SLAOverdueReader.ListTenantsWithOverdueStages.
	ListTenantsWithStuckHolds(ctx context.Context, tx db.Tx, now time.Time, stuckAfter time.Duration) ([]string, error)
}

// NoopReleaseHoldReader is the fail-closed default: it reports no stuck holds
// and no tenants with stuck holds.
type NoopReleaseHoldReader struct{}

func (NoopReleaseHoldReader) ListStuckHolds(context.Context, db.Tx, string, time.Time, time.Duration, int) ([]StuckReleaseHoldView, error) {
	return nil, nil
}

func (NoopReleaseHoldReader) ListTenantsWithStuckHolds(context.Context, db.Tx, time.Time, time.Duration) ([]string, error) {
	return nil, nil
}
