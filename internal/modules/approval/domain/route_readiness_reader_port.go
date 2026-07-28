package domain

import (
	"context"

	"metaldocs/internal/platform/db"
)

// RouteReadinessReader is the narrow read surface cross-module consumers use to
// ask "does an active approval route exist for this subject?" without reaching
// across the module boundary into approval's owned base table (ADR-0039 D1/D3(b);
// same H-G class as controlleddocumentsdomain.CDFieldReader and
// taxonomydomain.AreaCatalogReader). approval owns approval_routes; callers that
// need route readiness depend on this port instead of issuing raw SQL against
// the table.
//
// Readiness — not resolution. Consumers get a boolean/set, never a route id or
// any route internals; route resolution stays inside approval (submit path,
// ApprovalRouteResolver). The two methods differ deliberately in shape because
// their call sites differ:
//
//   - ActiveRouteSubjectKeys is a bulk catalog read run off-tx (admin listings,
//     annotation of a profile catalog). The adapter binds its own pool, so the
//     caller needs no executor of its own.
//   - HasActiveRoute takes a db.DB executor so the same stateless method serves
//     an in-tx caller: pass the caller's *sql.Tx and the check runs inside the
//     caller's existing transaction (*sql.Tx satisfies db.DB structurally —
//     internal/platform/db/tx.go). This is the creation-gate call site
//     (controlleddocuments atomic create), which must observe the same snapshot
//     as the INSERT it guards.
//
// Both SELECTs are plain and non-recording, and MUST stay non-recording even
// when run in-tx (HS-PRE-1: never call an authz-recording read inside a
// lock-holding tx).
type RouteReadinessReader interface {
	// ActiveRouteSubjectKeys returns the set of approval_routes.subject_key
	// values that have at least one active route for (tenantID, subjectKind).
	// An empty (non-nil) map means "no subject in this tenant is route-ready";
	// it is never an error. Off-tx / pool-backed.
	ActiveRouteSubjectKeys(ctx context.Context, tenantID, subjectKind string) (map[string]struct{}, error)

	// HasActiveRoute reports whether an active approval route exists for
	// (tenantID, subjectKind, subjectKey). exec is the caller's executor — pass
	// a *sql.Tx to run the check inside the caller's transaction, or a *sql.DB
	// pool for an off-tx check.
	HasActiveRoute(ctx context.Context, exec db.DB, tenantID, subjectKind, subjectKey string) (bool, error)
}

// NoopRouteReadinessReader is the explicit null-object for callers that never
// exercise a route-readiness read (tests, paths where approval governance is
// out of scope). It satisfies RouteReadinessReader and always reports "no route"
// with nil errors.
//
// Note the asymmetry with the production adapter: a Noop reader reports every
// subject as NOT route-ready, so a consumer that gates on HasActiveRoute fails
// CLOSED under Noop. That is the intended default — never wire this into a
// composition root that must enforce the gate.
type NoopRouteReadinessReader struct{}

// ActiveRouteSubjectKeys always returns an empty set with a nil error.
func (NoopRouteReadinessReader) ActiveRouteSubjectKeys(_ context.Context, _, _ string) (map[string]struct{}, error) {
	return map[string]struct{}{}, nil
}

// HasActiveRoute always returns (false, nil).
func (NoopRouteReadinessReader) HasActiveRoute(_ context.Context, _ db.DB, _, _, _ string) (bool, error) {
	return false, nil
}
