package authz

import (
	"context"
	"database/sql"
	"sync/atomic"
)

// BypassKind classifies an authorization bypass for the audit trail.
type BypassKind string

const (
	// BypassKindSystemAdmin is the tier-2 short-circuit in Require: an actor holding
	// system_admin is granted any capability in any area without the area check.
	BypassKindSystemAdmin BypassKind = "system_admin"
	// BypassKindBackground is the BypassSystem scheduler/job bridge: a background
	// task disables tier-2 area enforcement for the remainder of its transaction.
	BypassKindBackground BypassKind = "background"
)

// BypassEvent is the audit record for a single authorization bypass (ADR 0022
// Phase 11, F8). It carries the same axes as a normal grant audit (actor, tenant,
// resource, context) so a bypass is visible at the same fidelity in the audit.read
// surface.
type BypassEvent struct {
	Kind       BypassKind
	ActorID    string // system_admin user id; "system" for background bridges
	TenantID   string // tenant uuid (text); "" for cross-tenant background sweeps
	Capability string // the capability granted (system_admin short-circuit only)
	AreaCode   string // the area arg of the short-circuited check (system_admin only)
}

// BypassAuditSink records authorization bypasses into the audit pipe. It is set
// once at the composition root via SetBypassAuditSink so this low-level package
// stays a leaf — the concrete audit Writer is adapted at the root, never imported
// here. RecordBypass writes inside the caller's transaction (tx) at the same
// atomicity as the in-tx normal-grant audit (documentsAuditAdapter.WriteTx): if it
// returns an error the bypassed operation is rolled back, so a bypass can never
// commit unaudited.
type BypassAuditSink interface {
	RecordBypass(ctx context.Context, tx *sql.Tx, ev BypassEvent) error
}

// bypassAuditSink is the process-wide sink, normally set once before any
// request/job goroutine runs. An unset (nil) sink disables bypass auditing (tools
// / unit tests that never wire it) and makes the emit paths add ZERO SQL, so
// sqlmock-based authz tests are unaffected. Stored as an atomic pointer so a test
// that re-sets the sink (with t.Cleanup) can never data-race the hot-path reads in
// recordBypass / BypassSystem under `go test -race`.
var bypassAuditSink atomic.Pointer[BypassAuditSink]

// SetBypassAuditSink installs the bypass audit sink. Call once at startup before
// serving (production) or with t.Cleanup reset in tests; passing nil disables
// bypass auditing.
func SetBypassAuditSink(sink BypassAuditSink) {
	if sink == nil {
		bypassAuditSink.Store(nil)
		return
	}
	bypassAuditSink.Store(&sink)
}

// currentBypassSink returns the installed sink, or nil if auditing is off.
func currentBypassSink() BypassAuditSink {
	if p := bypassAuditSink.Load(); p != nil {
		return *p
	}
	return nil
}

// recordBypass emits a bypass audit event through the configured sink, propagating
// any error into the caller's tx (fail-closed/atomic — matches the in-tx grant
// audit convention). A nil sink is a no-op with no SQL.
func recordBypass(ctx context.Context, tx *sql.Tx, ev BypassEvent) error {
	sink := currentBypassSink()
	if sink == nil {
		return nil
	}
	return sink.RecordBypass(ctx, tx, ev)
}

// softActorID / softTenantID read the transaction-local identity GUCs without
// failing when unset (unlike MustActorID/MustTenantID). Used by the background
// bypass bridge where identity may not be seeded (cross-tenant system sweeps) —
// those are attributed to "system" / "". The GUC names are compile-time constants
// (no injection surface), inlined as literals so the queries are identical to
// MustActorID's (keeps the sqlmock-based tests matching one statement shape).
func softActorID(ctx context.Context, tx *sql.Tx) string {
	return softGUC(ctx, tx, "SELECT current_setting('metaldocs.actor_id', true)", "system")
}

func softTenantID(ctx context.Context, tx *sql.Tx) string {
	return softGUC(ctx, tx, "SELECT current_setting('metaldocs.tenant_id', true)", "")
}

func softGUC(ctx context.Context, tx *sql.Tx, query, def string) string {
	var v string
	if err := tx.QueryRowContext(ctx, query).Scan(&v); err != nil {
		return def
	}
	if v == "" {
		return def
	}
	return v
}
