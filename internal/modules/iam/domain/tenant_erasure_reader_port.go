package domain

import "context"

// TenantErasureReader is the narrow read surface cross-module consumers use
// to learn whether a tenant has been erased (metaldocs.tenants.erased_at IS
// NOT NULL) without reaching into iam's tables directly. IAM owns
// metaldocs.tenants; a caller that must decide whether to withhold that
// tenant's data at read time depends on this port instead of issuing raw SQL
// (module-boundary rule: cross-module access goes through a module's
// application service or published Go interface, never another module's
// repository, SQL, or domain internals).
//
// The implementation reads from the connection pool — never from a caller's
// lock-holding transaction (H-PRE-1) — and mirrors the narrow-port pattern of
// UserDisplayNameReader / TenantUserReader (M4/F4.1, F4.5) rather than
// TenantRepository.TenantLookupTx, which is tx-scoped for the onboarding
// existence check and treats "no row" as a benign not-found. This port is
// used to gate PII disclosure, so it deliberately does not offer that
// lenient shape: see IsErased's own doc for why "not found" fails closed here
// instead.
type TenantErasureReader interface {
	// IsErased reports whether tenantID's metaldocs.tenants row has a
	// non-null erased_at. Unlike TenantRepository.TenantLookupTx, a tenantID
	// that resolves to no row is NOT treated as "not erased" — it is returned
	// as an error. This is a read-path integrity gate (deciding whether to
	// serve stored audit payload content), not a best-effort existence check:
	// an audit event's tenant_id is always populated by domain.NewEvent and
	// should always resolve to a real tenant row, so a miss here is an
	// anomaly the caller must fail closed on rather than silently treat as
	// "safe to serve".
	IsErased(ctx context.Context, tenantID string) (bool, error)
}
