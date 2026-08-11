//go:build integration
// +build integration

package testdb

import (
	"context"

	tenant "metaldocs/internal/platform/tenant"
)

// AuthzCtx builds a context carrying the tenant+actor identity the TxRunner
// chokepoint (internal/platform/db/runner.go seedTxIdentityFromContext)
// auto-seeds into the metaldocs.tenant_id / metaldocs.actor_id GUCs on every
// Do call. Integration tests that drive a service-under-test through a
// TxRunner-backed call MUST pass a context built with this helper (or an
// equivalent seeded ctx) — an unseeded context.Background() or a context
// carrying the wrong key pair (e.g. iamdomain.WithAuthContext) leaves the
// GUCs unset and the SUT's own authz.Require call fails closed with
// "authz: metaldocs.actor_id GUC not set on transaction".
//
// Centralizes the package-local authzCtx helpers duplicated in
// controlleddocuments/application/txrunner_test_helper_test.go and
// templates/application/txrunner_test_helper_test.go — same semantics.
func AuthzCtx(tenantID, actorID string) context.Context {
	return tenant.WithActorID(tenant.WithTenantID(context.Background(), tenantID), actorID)
}
