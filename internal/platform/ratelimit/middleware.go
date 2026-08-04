package ratelimit

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/netip"
	"strconv"
	"time"

	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/security"
)

type Middleware struct {
	cfg          Config
	trustedCIDRs []netip.Prefix
	logger       *slog.Logger
	store        Store
}

// New constructs a rate-limit middleware. When cfg.Store is set, that Store
// is used directly (letting callers share one Store — e.g. one Redis
// connection — across multiple Middleware instances). Otherwise New builds a
// private in-memory Store (single-replica semantics; see memoryStore doc
// comment for the multi-replica caveat). Pass a long-lived (app-wide)
// context — never a per-request context — since the in-memory store's
// sweeper goroutine keys its lifetime off it.
func New(ctx context.Context, cfg Config) *Middleware {
	store := cfg.Store
	if store == nil {
		store = newMemoryStore(ctx, cfg.SweepInterval, cfg.IdleThreshold, cfg.MaxEntries, slog.Default())
	}
	return &Middleware{
		cfg:          cfg,
		trustedCIDRs: cfg.TrustedProxyCIDRs,
		logger:       slog.Default(),
		store:        store,
	}
}

// NewWithStore constructs a rate-limit middleware backed by an
// already-constructed Store — used to share one Store (e.g. one Redis
// client) across multiple Middleware instances (pre-auth login limiter +
// global envelope limiter), and by tests that want to exercise a specific
// backend (redisStore, or a second memoryStore instance) directly.
func NewWithStore(cfg Config, store Store) *Middleware {
	return &Middleware{
		cfg:          cfg,
		trustedCIDRs: cfg.TrustedProxyCIDRs,
		logger:       slog.Default(),
		store:        store,
	}
}

// Wait blocks until the underlying store's background goroutines (if any)
// have exited. Intended for tests asserting the in-memory sweeper's
// goroutine terminates on context cancel. No-op for stores without
// background work.
func (m *Middleware) Wait() {
	if ms, ok := m.store.(*memoryStore); ok {
		ms.Wait()
	}
}

// WithClock injects a deterministic time source into the underlying
// in-memory store. Intended for tests; must be called before the middleware
// sees traffic. No-op when the backing store is not the in-memory store.
func (m *Middleware) WithClock(now func() time.Time) *Middleware {
	if ms, ok := m.store.(*memoryStore); ok {
		ms.WithClock(now)
	}
	return m
}

// WithLogger replaces the slog.Logger used for fallback / overflow records
// logged directly by Middleware. Intended for tests that want to assert log
// content.
func (m *Middleware) WithLogger(l *slog.Logger) *Middleware {
	if l != nil {
		m.logger = l
	}
	return m
}

// SweepNow runs one sweep pass synchronously against the current clock, when
// the backing store is the in-memory store. Intended for tests that want
// deterministic eviction timing.
func (m *Middleware) SweepNow() {
	if ms, ok := m.store.(*memoryStore); ok {
		ms.SweepNow()
	}
}

// Size returns the current limiter-entry count of the backing in-memory
// store. Intended for tests. Returns 0 for non-memory backends.
func (m *Middleware) Size() int {
	if ms, ok := m.store.(*memoryStore); ok {
		return ms.Size()
	}
	return 0
}

// Close releases the backing store's resources (e.g. a Redis connection).
// Safe to call multiple times.
func (m *Middleware) Close() {
	if m.store != nil {
		m.store.Close()
	}
}

// Limit returns an http.Handler wrapper that enforces the quota for the
// given route. userExtractor pulls the subject id out of request ctx (the
// IAM middleware sets it before this middleware runs). When the user id is
// empty (route misordered or intentionally anonymous), the limiter falls
// back to an IP key resolved via the trusted-proxy CIDRs from Config —
// never bypasses (H2 fix).
func (m *Middleware) Limit(key RouteKey, userExtractor func(*http.Request) string, next http.Handler) http.Handler {
	quota, ok := m.cfg.QuotaFor(key)
	if !ok {
		return next // no quota configured
	}
	if quota <= 0 {
		// Defense-in-depth: NewConfig validates >= 1, but guard here so a
		// zero injected via reflection or test helpers degrades to no-limit
		// instead of panicking on integer divide.
		m.logger.Error("ratelimit: invalid quota, degrading to no-limit", "route", string(key), "quota", quota)
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lk, ok := m.bucketKey(r, key, userExtractor)
		if !ok {
			// Neither a user id nor a parseable client IP. Prior behavior
			// was a silent bypass (H2 fail-open). Now fail-closed.
			m.logger.WarnContext(r.Context(), "ratelimit: no identity and no parseable IP, fail-closed",
				"route", string(key),
				"remote_addr", r.RemoteAddr,
			)
			writeRateLimitError(w, quota, 60)
			return
		}

		allowed, retryAfter, err := m.store.Allow(r.Context(), lk, quota, time.Minute)
		if err != nil {
			// Store implementations are expected to fail open internally
			// (see redisStore) rather than return err for a backend outage.
			// A non-nil err here is unexpected; fail open rather than deny
			// traffic for a rate-limiter-internal problem — same
			// availability-over-strictness rationale as redisStore's
			// documented fail-open branch.
			m.logger.ErrorContext(r.Context(), "ratelimit: store error, failing open", "route", string(key), "err", err)
			next.ServeHTTP(w, r)
			return
		}
		if !allowed {
			retrySec := int(retryAfter.Seconds())
			if retrySec < 1 {
				retrySec = 1
			}
			writeRateLimitError(w, quota, retrySec)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// bucketKey resolves the limiter key for r. Preference: authenticated user
// id from userExtractor → trusted-proxy-resolved client IP. Returns
// (key, true) on success; ("", false) when both paths fail — caller must
// fail-closed. The literal "user:" / "ip:" prefixes namespace the two
// keyspaces so a user id matching an IP string cannot collide with the IP
// bucket.
func (m *Middleware) bucketKey(r *http.Request, key RouteKey, userExtractor func(*http.Request) string) (string, bool) {
	if user := userExtractor(r); user != "" {
		return string(key) + ":user:" + user, true
	}
	m.logger.DebugContext(r.Context(), "ratelimit: empty user id, falling back to IP key",
		"route", string(key),
		"remote_addr", r.RemoteAddr,
	)
	addr := security.ClientIP(r, m.trustedCIDRs)
	if !addr.IsValid() {
		return "", false
	}
	return string(key) + ":ip:" + addr.String(), true
}

// GlobalEnvelopeWrap returns an http.Handler wrapper that enforces the
// RouteGlobalEnvelope quota on every request except health probes
// (/api/v1/health/live, /api/v1/health/ready). It replaces the legacy
// security.RateLimiter (F-05/D-04, Wave 2.8).
//
// userExtractor must return the authenticated user id from the request context,
// or "" when the request is unauthenticated (e.g. a session-less path that
// slipped through authn). When "" is returned the limiter falls back to the
// trusted-proxy-resolved client IP — same identity precedence as the old
// security.RateLimiter.
//
// The method returns next unchanged when no quota is configured for
// RouteGlobalEnvelope (e.g. quota map was built without it).
func (m *Middleware) GlobalEnvelopeWrap(userExtractor func(*http.Request) string, next http.Handler) http.Handler {
	limited := m.Limit(RouteGlobalEnvelope, userExtractor, next)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/health/live", "/api/v1/health/ready":
			next.ServeHTTP(w, r)
			return
		}
		limited.ServeHTTP(w, r)
	})
}

func writeRateLimitError(w http.ResponseWriter, quota, retryAfterSec int) {
	// RFC 9457 (AD-2): one error shape across the API. The quota/retry detail
	// the legacy body carried is preserved via the standard Retry-After header
	// plus the human-readable detail string.
	w.Header().Set("Retry-After", strconv.Itoa(retryAfterSec))
	prob := problem.New(http.StatusTooManyRequests, problem.CodeRateLimitExceeded, "Too many requests").
		WithDetail(fmt.Sprintf("Rate limit of %d requests per minute exceeded; retry after %d seconds.", quota, retryAfterSec))
	_ = problem.Write(w, prob)
}
