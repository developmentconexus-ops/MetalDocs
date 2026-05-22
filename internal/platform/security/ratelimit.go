package security

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"sync"
	"time"

	authdomain "metaldocs/internal/modules/auth/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/config"
)

// defaultMaxRateLimitEntries caps the in-memory identity map to bound memory
// against rotating-IP DoS. Sweep evicts entries whose window has expired on
// every allow() call; if the map still hits the cap (sustained attack with
// distinct identities inside one window), new identities are denied
// fail-closed instead of inflating the map until OOM.
const defaultMaxRateLimitEntries = 100_000

type windowCounter struct {
	windowStart time.Time
	lastSeen    time.Time
	count       int
}

type RateLimiter struct {
	enabled           bool
	window            time.Duration
	maxRequests       int
	maxEntries        int
	now               func() time.Time
	logger            *slog.Logger
	mu                sync.Mutex
	byIdentity        map[string]windowCounter
	trustedProxyCIDRs []netip.Prefix
}

func NewRateLimiter(cfg config.RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		enabled:           cfg.Enabled,
		window:            time.Duration(cfg.WindowSeconds) * time.Second,
		maxRequests:       cfg.MaxRequests,
		maxEntries:        defaultMaxRateLimitEntries,
		now:               time.Now,
		logger:            slog.Default(),
		byIdentity:        map[string]windowCounter{},
		trustedProxyCIDRs: cfg.TrustedProxyCIDRs,
	}
}

// WithMaxEntries overrides the in-memory identity cap. Intended for tests
// that exercise the fail-closed overflow path with small maps.
func (r *RateLimiter) WithMaxEntries(n int) *RateLimiter {
	if n > 0 {
		r.maxEntries = n
	}
	return r
}

// WithClock injects a deterministic time source. Intended for tests.
func (r *RateLimiter) WithClock(now func() time.Time) *RateLimiter {
	if now != nil {
		r.now = now
	}
	return r
}

// WithLogger replaces the slog.Logger used for eviction / overflow warnings.
func (r *RateLimiter) WithLogger(l *slog.Logger) *RateLimiter {
	if l != nil {
		r.logger = l
	}
	return r
}

// Size returns the current identity-map size. Intended for tests.
func (r *RateLimiter) Size() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.byIdentity)
}

func (r *RateLimiter) Wrap(next http.Handler) http.Handler {
	if !r.enabled {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if shouldSkipRateLimit(req.URL.Path) {
			next.ServeHTTP(w, req)
			return
		}

		identity := r.requestIdentity(req)
		allowed, retryAfter := r.allow(identity)
		if !allowed {
			w.Header().Set("Retry-After", retryAfter)
			writeAPIError(w, http.StatusTooManyRequests, "RATE_LIMITED", "Too many requests", requestTraceID(req))
			return
		}

		next.ServeHTTP(w, req)
	})
}

func (r *RateLimiter) allow(identity string) (bool, string) {
	now := r.now().UTC()

	r.mu.Lock()
	defer r.mu.Unlock()

	r.sweepExpiredLocked(now)

	current, ok := r.byIdentity[identity]
	if !ok {
		if len(r.byIdentity) >= r.maxEntries {
			r.logger.Warn("ratelimit: identity map full, denying new identity",
				"map_size", len(r.byIdentity),
				"max_entries", r.maxEntries,
			)
			return false, strconvSecondsCeil(r.window)
		}
		r.byIdentity[identity] = windowCounter{
			windowStart: now,
			lastSeen:    now,
			count:       1,
		}
		return true, "0"
	}

	if now.Sub(current.windowStart) >= r.window {
		r.byIdentity[identity] = windowCounter{
			windowStart: now,
			lastSeen:    now,
			count:       1,
		}
		return true, "0"
	}

	if current.count >= r.maxRequests {
		retryAfter := current.windowStart.Add(r.window).Sub(now)
		if retryAfter < 0 {
			retryAfter = 0
		}
		return false, strconvSecondsCeil(retryAfter)
	}

	current.count++
	current.lastSeen = now
	r.byIdentity[identity] = current
	return true, "0"
}

// sweepExpiredLocked evicts entries whose window has elapsed. Caller must hold r.mu.
// Per-call sweep is O(n) but n is bounded by maxEntries (~100k); acceptable cost.
func (r *RateLimiter) sweepExpiredLocked(now time.Time) {
	evicted := 0
	for k, v := range r.byIdentity {
		if now.Sub(v.windowStart) >= r.window {
			delete(r.byIdentity, k)
			evicted++
		}
	}
	if evicted > 0 {
		r.logger.Warn("ratelimit: swept expired identities",
			"evicted", evicted,
			"remaining", len(r.byIdentity),
		)
	}
}

func shouldSkipRateLimit(path string) bool {
	return path == "/api/v1/health/live" || path == "/api/v1/health/ready"
}

func (r *RateLimiter) requestIdentity(req *http.Request) string {
	if currentUser, ok := authdomain.CurrentUserFromContext(req.Context()); ok && strings.TrimSpace(currentUser.UserID) != "" {
		return "user:" + strings.TrimSpace(currentUser.UserID)
	}
	if userID := strings.TrimSpace(iamdomain.UserIDFromContext(req.Context())); userID != "" {
		return "user:" + userID
	}
	if addr := ClientIP(req, r.trustedProxyCIDRs); addr.IsValid() {
		return "ip:" + addr.String()
	}
	return "ip:unknown"
}

func requestTraceID(r *http.Request) string {
	if traceID := strings.TrimSpace(r.Header.Get("X-Trace-Id")); traceID != "" {
		return traceID
	}
	return "trace-local"
}

func writeAPIError(w http.ResponseWriter, status int, code, message, traceID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]any{
			"code":     code,
			"message":  message,
			"details":  map[string]any{},
			"trace_id": traceID,
		},
	})
}

func strconvSecondsCeil(d time.Duration) string {
	sec := int(d / time.Second)
	if d%time.Second != 0 {
		sec++
	}
	if sec < 0 {
		sec = 0
	}
	return strconv.Itoa(sec)
}
