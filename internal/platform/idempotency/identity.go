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
// This is the ONLY place that decision should be made. Grep
// `idempotency.TenantActorFromContext` for every call site that has adopted
// it; a hand-rolled `tenant.FromContext(ctx)` + discarded-error closure
// living outside this file, anywhere in the tree, is the exact class of bug
// this function exists to make structurally unrepresentable — the fix is to
// delete that closure and call this function, not to patch its error
// handling in place.
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
