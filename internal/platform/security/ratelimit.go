package security

import (
	"encoding/json"
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

type windowCounter struct {
	windowStart time.Time
	count       int
}

type RateLimiter struct {
	enabled           bool
	window            time.Duration
	maxRequests       int
	now               func() time.Time
	mu                sync.Mutex
	byIdentity        map[string]windowCounter
	trustedProxyCIDRs []netip.Prefix
}

func NewRateLimiter(cfg config.RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		enabled:           cfg.Enabled,
		window:            time.Duration(cfg.WindowSeconds) * time.Second,
		maxRequests:       cfg.MaxRequests,
		now:               time.Now,
		byIdentity:        map[string]windowCounter{},
		trustedProxyCIDRs: cfg.TrustedProxyCIDRs,
	}
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

	current, ok := r.byIdentity[identity]
	if !ok || now.Sub(current.windowStart) >= r.window {
		r.byIdentity[identity] = windowCounter{
			windowStart: now,
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
	r.byIdentity[identity] = current
	return true, "0"
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
