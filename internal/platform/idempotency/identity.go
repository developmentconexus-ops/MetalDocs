package idempotency

import (
	"context"

	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/tenant"
)

// TenantActorFromContext is the single sanctioned actorFromCtx implementation
// for idempotency.Require (see middleware.go's Require docstring). It exists
// because three call sites (taxonomy, templates, approval's idempotentHandler
// path — see #90/A3.5) had each hand-copied the same two-line extraction,
// and one of the three copies discarded tenant.FromContext's error instead
// of propagating it, silently scoping the replay key to tenantID="" for an
// authenticated-but-tenant-less request (A3.3-deferred defect, PR #108
// review thread, commit 66cfb664).
//
// Both halves fail closed, symmetrically: a blank tenant is exactly as
// dangerous as a blank actor (Require's own docstring already explains why
// for actor — one shared "" slot every failed-identity caller collides
// into), so neither may be silently substituted with a zero value.
//
// This is the ONLY place that decision should be made. A hand-rolled
// actorFromCtx-shaped closure (func(context.Context) (string, string,
// error)) living outside this file, anywhere in the tree, that calls
// tenant.FromContext directly is the exact class of bug this function exists
// to replace — and it is enforced, not just conventional: the
// idempotency-identity-scope-guard check (scripts/check-idempotency-identity-scope.sh,
// registered in tools/verify/registry.go, runs in ci.yml:verify) fails CI on
// any such closure outside this file. That check is a text/regex scan with
// documented gaps (method receivers, named result lists — see its header),
// not a compiler-enforced guarantee, so "structurally unrepresentable" would
// overstate what it does; read its header for exactly what it can and
// cannot catch. The fix for a flagged closure is still to delete it and
// call this function, not to patch its error handling in place.
func TenantActorFromContext(ctx context.Context) (string, string, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return "", "", err
	}
	actorID, err := authn.RequireUserID(ctx)
	if err != nil {
		return "", "", err
	}
	return tenantID, actorID, nil
}
