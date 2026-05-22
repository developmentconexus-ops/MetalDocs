package ratelimit

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

const (
	defaultSweepInterval = time.Minute
	defaultIdleThreshold = 2 * time.Minute
)

type limiterEntry struct {
	lim        *rate.Limiter
	lastAccess atomic.Int64 // unix nano; updated on every Limit() call
}

type Middleware struct {
	cfg           Config
	sweepInterval time.Duration
	idleThreshold time.Duration
	now           func() time.Time
	logger        *slog.Logger
	limiters      sync.Map // key: "<route_key>:<user_id>" → *limiterEntry
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
	m := &Middleware{
		cfg:           cfg,
		sweepInterval: sweep,
		idleThreshold: idle,
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
			evicted++
			return true
		}
		if e.lastAccess.Load() < cutoff {
			m.limiters.Delete(k)
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
// IAM middleware sets it before this middleware runs).
func (m *Middleware) Limit(key RouteKey, userExtractor func(*http.Request) string, next http.Handler) http.Handler {
	quota, ok := m.cfg.Quotas[key]
	if !ok {
		return next // no quota configured
	}
	if quota <= 0 {
		// Defense-in-depth against zero / negative quota (would panic
		// inside rate.Every on integer divide).
		m.logger.Error("ratelimit: invalid quota, degrading to no-limit", "route", string(key), "quota", quota)
		return next
	}
	interval := time.Minute / time.Duration(quota)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := userExtractor(r)
		if user == "" {
			// No user id → bypass; IAM middleware should have rejected already.
			next.ServeHTTP(w, r)
			return
		}
		lk := string(key) + ":" + user
		nowNS := m.now().UnixNano()

		// Load-first, only allocate a new rate.Limiter on miss. Prior
		// implementation eagerly allocated a limiter for every request,
		// throwing it away when the key already existed.
		var entry *limiterEntry
		if v, ok := m.limiters.Load(lk); ok {
			entry = v.(*limiterEntry)
		} else {
			fresh := &limiterEntry{lim: rate.NewLimiter(rate.Every(interval), quota)}
			fresh.lastAccess.Store(nowNS)
			actual, _ := m.limiters.LoadOrStore(lk, fresh)
			entry = actual.(*limiterEntry)
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
