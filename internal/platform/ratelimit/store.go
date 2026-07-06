package ratelimit

import (
	"context"
	"time"
)

// Store decides whether a request identified by key is allowed under the
// given per-minute quota, and owns the counting state backing that decision.
// Implementations must be safe for concurrent use.
//
// quota is expressed in requests-per-minute (matches Config.quotas). window
// is always time.Minute today; it is passed explicitly so a Store
// implementation is not hard-coded to a single period.
//
// allowed reports whether the request may proceed. retryAfter is the
// caller-facing Retry-After duration when allowed is false (>= 1s); it is
// unspecified (and ignored) when allowed is true. err is non-nil only for a
// backend failure (e.g. Redis unreachable); Store implementations decide
// their own fail-open/fail-closed policy internally and should prefer
// returning (true, 0, nil) with a logged warning over surfacing err when the
// implementation intends to fail open — err is reserved for the rare case a
// caller must react to (currently unused by the in-memory store; used
// diagnostically by the Redis store, which fails open internally and never
// returns err from Allow).
type Store interface {
	Allow(ctx context.Context, key string, quota int, window time.Duration) (allowed bool, retryAfter time.Duration, err error)

	// Close releases any background resources (sweeper goroutines, network
	// connections). Safe to call multiple times.
	Close()
}
