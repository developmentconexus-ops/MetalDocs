package ratelimit

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

const (
	defaultSweepInterval     = time.Minute
	defaultIdleThreshold     = 2 * time.Minute
	defaultMaxLimiterEntries = 100_000
)

type limiterEntry struct {
	lim        *rate.Limiter
	lastAccess atomic.Int64 // unix nano; updated on every Allow() call
}

// memoryStore is the default in-memory Store backend. It preserves, bit for
// bit, the original fixed-window-via-token-bucket semantics: one
// rate.Limiter per bucket key, LoadOrStore-raced insertion, a maxEntries cap
// with fail-closed behavior on overflow, and a background sweeper that
// evicts idle entries.
//
// Known limitation (the defect M8/F8.2 fixes): each process owns its own
// limiters map, so N replicas of metaldocs-api each enforce the full quota
// independently — the effective fleet-wide quota is N*quota. This backend is
// only correct for single-replica deployments; see redisStore for the
// multi-replica-safe backend, and the METALDOCS_MULTI_REPLICA startup guard
// in store_config.go that refuses to boot this backend when multi-replica
// mode is declared.
type memoryStore struct {
	sweepInterval time.Duration
	idleThreshold time.Duration
	maxEntries    int
	now           func() time.Time
	logger        *slog.Logger
	limiters      sync.Map     // key -> *limiterEntry
	size          atomic.Int64 // limiter cardinality for fail-closed cap enforcement
	wg            sync.WaitGroup
}

// newMemoryStore constructs the in-memory store with a background sweeper
// that evicts limiter entries idle for longer than idleThreshold. The
// sweeper exits when ctx is cancelled. Pass a long-lived (app-wide) context
// — never a per-request context.
func newMemoryStore(ctx context.Context, sweepInterval, idleThreshold time.Duration, maxEntries int, logger *slog.Logger) *memoryStore {
	if sweepInterval <= 0 {
		sweepInterval = defaultSweepInterval
	}
	if idleThreshold <= 0 {
		idleThreshold = defaultIdleThreshold
	}
	if maxEntries <= 0 {
		maxEntries = defaultMaxLimiterEntries
	}
	if logger == nil {
		logger = slog.Default()
	}
	s := &memoryStore{
		sweepInterval: sweepInterval,
		idleThreshold: idleThreshold,
		maxEntries:    maxEntries,
		now:           time.Now,
		logger:        logger,
	}
	s.wg.Add(1)
	go s.sweepLoop(ctx)
	return s
}

// Wait blocks until the sweeper goroutine has exited. Intended for tests
// asserting the goroutine terminates on context cancel.
func (s *memoryStore) Wait() { s.wg.Wait() }

// WithClock injects a deterministic time source. Intended for tests; must be
// called before the store sees traffic.
func (s *memoryStore) WithClock(now func() time.Time) {
	if now != nil {
		s.now = now
	}
}

// SweepNow runs one sweep pass synchronously against the current clock.
// Intended for tests that want deterministic eviction timing.
func (s *memoryStore) SweepNow() { s.sweep() }

// Size returns the current limiter-entry count. Intended for tests.
func (s *memoryStore) Size() int {
	n := 0
	s.limiters.Range(func(_, _ any) bool {
		n++
		return true
	})
	return n
}

// Close is a no-op beyond letting the sweeper goroutine be reaped by ctx
// cancellation (the caller-provided ctx passed to newMemoryStore owns that
// lifecycle) — present to satisfy the Store interface.
func (s *memoryStore) Close() {}

func (s *memoryStore) sweepLoop(ctx context.Context) {
	defer s.wg.Done()
	t := time.NewTicker(s.sweepInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			s.sweep()
		}
	}
}

func (s *memoryStore) sweep() {
	cutoff := s.now().Add(-s.idleThreshold).UnixNano()
	evicted := 0
	remaining := 0
	s.limiters.Range(func(k, v any) bool {
		e, ok := v.(*limiterEntry)
		if !ok {
			s.limiters.Delete(k)
			s.size.Add(-1)
			evicted++
			return true
		}
		if e.lastAccess.Load() < cutoff {
			s.limiters.Delete(k)
			s.size.Add(-1)
			evicted++
		} else {
			remaining++
		}
		return true
	})
	if evicted > 0 {
		s.logger.Warn("ratelimit: swept idle limiters",
			"evicted", evicted,
			"remaining", remaining,
		)
	}
}

// Allow implements Store. err is always nil for the in-memory backend — it
// has no external failure mode. A false allowed with retryAfter == 0 means
// "map full, fail closed" (mirrors the pre-refactor loadOrInsert-cap
// contract); the caller (Middleware.Limit) distinguishes this from a normal
// quota-exceeded rejection only via logging, both produce the same 429.
func (s *memoryStore) Allow(_ context.Context, key string, quota int, window time.Duration) (bool, time.Duration, error) {
	interval := window / time.Duration(quota)
	nowNS := s.now().UnixNano()

	entry, ok := s.loadOrInsert(key, interval, quota, nowNS)
	if !ok {
		// Limiter map cap reached. Mirrors security.RateLimiter fail-closed
		// cap contract from C2.
		s.logger.Warn("ratelimit: limiter map full, denying new key",
			"key", key,
			"max_entries", s.maxEntries,
		)
		return false, 60 * time.Second, nil
	}
	entry.lastAccess.Store(nowNS)

	reservation := entry.lim.Reserve()
	if !reservation.OK() {
		return false, 60 * time.Second, nil
	}
	if d := reservation.Delay(); d > 0 {
		reservation.Cancel()
		// +1s to match the pre-refactor int(d.Seconds())+1 rounding: callers
		// convert retryAfter back to whole seconds via int(retryAfter.Seconds()),
		// so pad here to preserve the exact prior Retry-After value.
		return false, d + time.Second, nil
	}
	return true, 0, nil
}

// loadOrInsert returns the limiter entry for lk, allocating one on miss while
// enforcing the maxEntries cap. Returns (entry, true) on success;
// (nil, false) when inserting would exceed the cap (caller must fail-closed).
func (s *memoryStore) loadOrInsert(lk string, interval time.Duration, quota int, nowNS int64) (*limiterEntry, bool) {
	if v, ok := s.limiters.Load(lk); ok {
		return v.(*limiterEntry), true
	}
	fresh := &limiterEntry{lim: rate.NewLimiter(rate.Every(interval), quota)}
	fresh.lastAccess.Store(nowNS)
	actual, loaded := s.limiters.LoadOrStore(lk, fresh)
	if loaded {
		return actual.(*limiterEntry), true
	}
	if n := s.size.Add(1); n > int64(s.maxEntries) {
		// Lost the cap race — roll back the insert.
		s.limiters.Delete(lk)
		s.size.Add(-1)
		return nil, false
	}
	return fresh, true
}
