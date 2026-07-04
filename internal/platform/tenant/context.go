package tenant

import (
	"context"
	"errors"
	"strings"
)

type ctxKey struct{}

// actorCtxKey is a distinct key type from ctxKey so the tenant and actor
// carriers cannot collide on the same context.Value slot.
type actorCtxKey struct{}

// ErrTenantMissing is returned by FromContext when no authenticated tenant has
// been bound to the request context. Production handler paths MUST treat this
// as an internal-server-error invariant violation, not a 400.
var ErrTenantMissing = errors.New("tenant: not present in context")

// ErrActorMissing is returned by ActorFromContext when no authenticated actor
// has been bound to the request context. Consumers (e.g. the TxRunner
// chokepoint) MUST treat "missing" as "no identity to seed" and no-op rather
// than error, since system/janitor paths legitimately carry no actor.
var ErrActorMissing = errors.New("tenant: actor not present in context")

// WithTenantID returns a child context carrying the supplied tenant id. The
// auth middleware is the only production caller.
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	if strings.TrimSpace(tenantID) == "" {
		panic("WithTenantID: empty tenantID")
	}
	return context.WithValue(ctx, ctxKey{}, tenantID)
}

// FromContext extracts the authenticated-session tenant id. Returns
// ErrTenantMissing when absent or empty.
func FromContext(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", ErrTenantMissing
	}
	v, ok := ctx.Value(ctxKey{}).(string)
	if !ok {
		return "", ErrTenantMissing
	}
	if strings.TrimSpace(v) == "" {
		return "", ErrTenantMissing
	}
	return v, nil
}

// WithActorID returns a child context carrying the supplied actor (user) id.
// The auth middleware is the only production caller. Unlike WithTenantID,
// this does not panic on empty — the platform/db chokepoint treats "present
// but empty" as "absent" (no-op seed) rather than an invariant violation, so
// there is no need to fail fast here.
func WithActorID(ctx context.Context, actorID string) context.Context {
	return context.WithValue(ctx, actorCtxKey{}, actorID)
}

// ActorFromContext extracts the authenticated-session actor id. Returns
// ErrActorMissing when absent or empty.
func ActorFromContext(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", ErrActorMissing
	}
	v, ok := ctx.Value(actorCtxKey{}).(string)
	if !ok {
		return "", ErrActorMissing
	}
	if strings.TrimSpace(v) == "" {
		return "", ErrActorMissing
	}
	return v, nil
}
