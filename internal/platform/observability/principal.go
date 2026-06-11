package observability

import (
	"context"
	"sync/atomic"
)

// principalSlot is a mutable holder installed into the request context by the
// observability middleware BEFORE authentication runs. Because the middleware
// chain places access logging outside authn (REQ-MW-4), the authenticated
// principal is not in the outer context when the log line is written; the
// authn middleware reports it outward through this slot instead
// (observability stays free of module imports per REQ-TOP-2).
type principalSlot struct {
	v atomic.Value // string
}

type principalSlotKey struct{}

func withPrincipalSlot(ctx context.Context) context.Context {
	return context.WithValue(ctx, principalSlotKey{}, &principalSlot{})
}

// SetPrincipal records the authenticated principal id for access-log
// attribution. Called by the authn middleware once a session resolves; a
// no-op when the observability middleware is not installed or id is empty.
func SetPrincipal(ctx context.Context, userID string) {
	if userID == "" {
		return
	}
	if s, ok := ctx.Value(principalSlotKey{}).(*principalSlot); ok {
		s.v.Store(userID)
	}
}

func principalFromContext(ctx context.Context) string {
	if s, ok := ctx.Value(principalSlotKey{}).(*principalSlot); ok {
		if v, ok := s.v.Load().(string); ok {
			return v
		}
	}
	return ""
}
