package tenant

import (
	"context"
	"errors"
	"strings"
)

type ctxKey struct{}

// ErrTenantMissing is returned by FromContext when no authenticated tenant has
// been bound to the request context. Production handler paths MUST treat this
// as an internal-server-error invariant violation, not a 400.
var ErrTenantMissing = errors.New("tenant: not present in context")

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
