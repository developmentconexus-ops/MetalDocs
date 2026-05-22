package ratelimit

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/netip"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"

	"metaldocs/internal/platform/security"
)

const (
	defaultSweepInterval     = time.Minute
	defaultIdleThreshold     = 2 * time.Minute
	defaultMaxLimiterEntries = 100_000
)

type limiterEntry struct {
	lim        *rate.Limiter
	lastAccess atomic.Int64 // unix nano; updated on every Limit() call
}

type Middleware struct {
	cfg           Config
	sweepInterval time.Duration
	idleThreshold time.Duration
	maxEntries    int
	trustedCIDRs  []netip.Prefix
	now           func() time.Time
	logger        *slog.Logger
	limiters      sync.Map     // key: "<route_key>:user:<user_id>" or "<route_key>:ip:<addr>" → *limiterEntry
	size          atomic.Int64 // limiter cardinality for fail-closed cap enforcement
	wg            sync.WaitGroup
}

// New constructs a rate-limit middleware with a background sweeper that evicts
// limiter entries idle for longer than the idle threshold. The sweeper exits
// when ctx is cancelled. Pass a long-lived (app-wide) context — never a
// per-request context.
func New(ctx context.Context, cfg Config) *Middleware {
	sweep := cfg.SweepInterval
	if sweep <= 0 {
		sweep = defaultSweepInterval
	}
	idle := cfg.IdleThreshold
	if idle <= 0 {
		idle = defaultIdleThreshold
	}
	maxN := cfg.MaxEntries
	if maxN <= 0 {
		maxN = defaultMaxLimiterEntries
	}
	m := &Middleware{
		cfg:           cfg,
		sweepInterval: sweep,
		idleThreshold: idle,
		maxEntries:    maxN,
		trustedCIDRs:  cfg.TrustedProxyCIDRs,
		now:           time.Now,
		logger:        slog.Default(),
	}
	m.wg.Add(1)
	go m.sweepLoop(ctx)
	return m
}

// Wait blocks until the sweeper goroutine has exited. Intended for tests
// asserting the goroutine terminates on context cancel.
func (m *Middleware) Wait() { m.wg.Wait() }

// WithClock injects a deterministic time source. Intended for tests; must be
// called before the middleware sees traffic.
func (m *Middleware) WithClock(now func() time.Time) *Middleware {
	if now != nil {
		m.now = now
	}
	return m
}

// WithLogger replaces the slog.Logger used for sweep / fallback / overflow
// records. Intended for tests that want to assert log content.
func (m *Middleware) WithLogger(l *slog.Logger) *Middleware {
	if l != nil {
		m.logger = l
	}
	return m
}

// SweepNow runs one sweep pass synchronously against the current clock.
// Intended for tests that want deterministic eviction timing.
func (m *Middleware) SweepNow() { m.sweep() }

// Size returns the current limiter-entry count. Intended for tests.
func (m *Middleware) Size() int {
	n := 0
	m.limiters.Range(func(_, _ any) bool {
		n++
		return true
	})
	return n
}

func (m *Middleware) sweepLoop(ctx context.Context) {
	defer m.wg.Done()
	t := time.NewTicker(m.sweepInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			m.sweep()
		}
	}
}

func (m *Middleware) sweep() {
	cutoff := m.now().Add(-m.idleThreshold).UnixNano()
	evicted := 0
	remaining := 0
	m.limiters.Range(func(k, v any) bool {
		e, ok := v.(*limiterEntry)
		if !ok {
			m.limiters.Delete(k)
			m.size.Add(-1)
			evicted++
			return true
		}
		if e.lastAccess.Load() < cutoff {
			m.limiters.Delete(k)
			m.size.Add(-1)
			evicted++
		} else {
			remaining++
		}
		return true
	})
	if evicted > 0 {
		m.logger.Warn("ratelimit: swept idle limiters",
			"evicted", evicted,
			"remaining", remaining,
		)
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
	interval := time.Minute / time.Duration(quota)
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
		nowNS := m.now().UnixNano()

		entry, ok := m.loadOrInsert(lk, interval, quota, nowNS)
		if !ok {
			// Limiter map cap reached. Mirrors security.RateLimiter
			// fail-closed cap contract from C2.
			m.logger.WarnContext(r.Context(), "ratelimit: limiter map full, denying new key",
				"route", string(key),
				"key", lk,
				"max_entries", m.maxEntries,
			)
			writeRateLimitError(w, quota, 60)
			return
		}
		entry.lastAccess.Store(nowNS)

		reservation := entry.lim.Reserve()
		if !reservation.OK() {
			writeRateLimitError(w, quota, 60)
			return
		}
		if d := reservation.Delay(); d > 0 {
			reservation.Cancel()
			writeRateLimitError(w, quota, int(d.Seconds())+1)
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

// loadOrInsert returns the limiter entry for lk, allocating one on miss while
// enforcing the maxEntries cap. Returns (entry, true) on success;
// (nil, false) when inserting would exceed the cap (caller must fail-closed).
func (m *Middleware) loadOrInsert(lk string, interval time.Duration, quota int, nowNS int64) (*limiterEntry, bool) {
	if v, ok := m.limiters.Load(lk); ok {
		return v.(*limiterEntry), true
	}
	fresh := &limiterEntry{lim: rate.NewLimiter(rate.Every(interval), quota)}
	fresh.lastAccess.Store(nowNS)
	actual, loaded := m.limiters.LoadOrStore(lk, fresh)
	if loaded {
		return actual.(*limiterEntry), true
	}
	if n := m.size.Add(1); n > int64(m.maxEntries) {
		// Lost the cap race — roll back the insert.
		m.limiters.Delete(lk)
		m.size.Add(-1)
		return nil, false
	}
	return fresh, true
}

func writeRateLimitError(w http.ResponseWriter, quota, retryAfterSec int) {
	w.Header().Set("content-type", "application/json")
	w.Header().Set("retry-after", strconv.Itoa(retryAfterSec))
	w.WriteHeader(http.StatusTooManyRequests)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error":               "rate_limited",
		"quota_per_minute":    quota,
		"retry_after_seconds": retryAfterSec,
	})
}
