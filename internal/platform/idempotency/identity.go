package idempotency

import (
	"context"

	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/tenant"
)

// tenantActorFromContext is the only identity resolution path used by
// idempotency.Require. The middleware owns this call; it is deliberately not
// a callback in Require's API, so each route cannot install a parallel
// resolver with subtly different missing-identity semantics.
//
// Both halves fail closed, symmetrically: a blank tenant is exactly as
// dangerous as a blank actor (Require's own docstring already explains why
// for actor — one shared "" slot every failed-identity caller collides
// into), so neither may be silently substituted with a zero value.
func tenantActorFromContext(ctx context.Context) (string, string, error) {
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
