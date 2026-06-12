package unit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"

	authdomain "metaldocs/internal/modules/auth/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/ratelimit"
)

// userIDExtractorForTest mirrors the composition-root userIDExtractor injected
// into GlobalEnvelopeWrap in main.go: authdomain.CurrentUser → iamdomain.UserID → "".
func userIDExtractorForTest(r *http.Request) string {
	if cu, ok := authdomain.CurrentUserFromContext(r.Context()); ok && strings.TrimSpace(cu.UserID) != "" {
		return strings.TrimSpace(cu.UserID)
	}
	if uid := strings.TrimSpace(iamdomain.UserIDFromContext(r.Context())); uid != "" {
		return uid
	}
	return ""
}

func newGlobalEnvelope(t *testing.T, quotaPerMin int, trustedCIDRs ...netip.Prefix) (*ratelimit.Middleware, func()) {
	t.Helper()
	cfg, err := ratelimit.NewConfig(map[ratelimit.RouteKey]int{ratelimit.RouteGlobalEnvelope: quotaPerMin})
	if err != nil {
		t.Fatalf("ratelimit.NewConfig: %v", err)
	}
	cfg.TrustedProxyCIDRs = trustedCIDRs
	ctx, cancel := context.WithCancel(context.Background())
	m := ratelimit.New(ctx, cfg)
	return m, cancel
}

func wrapWithEnvelope(m *ratelimit.Middleware, next http.Handler) http.Handler {
	return m.GlobalEnvelopeWrap(userIDExtractorForTest, next)
}

func TestRateLimiterBlocksWhenLimitExceeded(t *testing.T) {
	m, cancel := newGlobalEnvelope(t, 1)
	defer cancel()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/documents", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := wrapWithEnvelope(m, mux)

	req1 := httptest.NewRequest(http.MethodGet, "/api/v1/documents", nil)
	req1 = req1.WithContext(authdomain.WithCurrentUser(req1.Context(), authdomain.CurrentUser{UserID: "user-1"}))
	req1.RemoteAddr = "203.0.113.1:1234"
	rr1 := httptest.NewRecorder()
	h.ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Fatalf("expected first request 200, got %d", rr1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/documents", nil)
	req2 = req2.WithContext(authdomain.WithCurrentUser(req2.Context(), authdomain.CurrentUser{UserID: "user-1"}))
	req2.RemoteAddr = "203.0.113.1:1235"
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second request 429, got %d", rr2.Code)
	}
}

func TestRateLimiterIsolatedByIdentity(t *testing.T) {
	m, cancel := newGlobalEnvelope(t, 1)
	defer cancel()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/search/documents", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := wrapWithEnvelope(m, mux)

	req1 := httptest.NewRequest(http.MethodGet, "/api/v1/search/documents", nil)
	req1 = req1.WithContext(authdomain.WithCurrentUser(req1.Context(), authdomain.CurrentUser{UserID: "user-a"}))
	req1.RemoteAddr = "203.0.113.1:1234"
	rr1 := httptest.NewRecorder()
	h.ServeHTTP(rr1, req1)

	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/search/documents", nil)
	req2 = req2.WithContext(authdomain.WithCurrentUser(req2.Context(), authdomain.CurrentUser{UserID: "user-b"}))
	req2.RemoteAddr = "203.0.113.2:1234"
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req2)

	if rr1.Code != http.StatusOK || rr2.Code != http.StatusOK {
		t.Fatalf("expected both users allowed, got user-a=%d user-b=%d", rr1.Code, rr2.Code)
	}
}

// TestRateLimiterIsolatesAnonByXFFWhenBehindTrustedProxy proves that when the
// reverse-proxy IP is in the trusted CIDR, two distinct upstream clients are
// keyed separately by their leftmost X-Forwarded-For value instead of sharing
// the proxy's RemoteAddr bucket.
func TestRateLimiterIsolatesAnonByXFFWhenBehindTrustedProxy(t *testing.T) {
	cidr, err := netip.ParsePrefix("10.0.0.0/8")
	if err != nil {
		t.Fatalf("invalid cidr: %v", err)
	}
	m, cancel := newGlobalEnvelope(t, 1, cidr)
	defer cancel()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := wrapWithEnvelope(m, mux)

	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req1.RemoteAddr = "10.0.0.5:443"
	req1.Header.Set("X-Forwarded-For", "198.51.100.7")
	rr1 := httptest.NewRecorder()
	h.ServeHTTP(rr1, req1)

	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req2.RemoteAddr = "10.0.0.5:443"
	req2.Header.Set("X-Forwarded-For", "198.51.100.8")
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req2)

	if rr1.Code != http.StatusOK || rr2.Code != http.StatusOK {
		t.Fatalf("distinct XFF clients must be isolated when behind trusted proxy, got %d / %d", rr1.Code, rr2.Code)
	}
}

// TestRateLimiterIgnoresXFFFromUntrustedSource confirms that a directly
// reachable client cannot forge X-Forwarded-For to evade rate limiting:
// without a trusted upstream, RemoteAddr is the only IP source.
func TestRateLimiterIgnoresXFFFromUntrustedSource(t *testing.T) {
	m, cancel := newGlobalEnvelope(t, 1)
	defer cancel()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := wrapWithEnvelope(m, mux)

	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req1.RemoteAddr = "203.0.113.7:9000"
	req1.Header.Set("X-Forwarded-For", "198.51.100.7")
	rr1 := httptest.NewRecorder()
	h.ServeHTTP(rr1, req1)

	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req2.RemoteAddr = "203.0.113.7:9001"
	req2.Header.Set("X-Forwarded-For", "198.51.100.8")
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req2)

	if rr1.Code != http.StatusOK {
		t.Fatalf("first request must be allowed, got %d", rr1.Code)
	}
	if rr2.Code != http.StatusTooManyRequests {
		t.Fatalf("second request from same RemoteAddr must be rate-limited despite differing XFF, got %d", rr2.Code)
	}
}

func TestRateLimiterSkipsHealthOnly(t *testing.T) {
	m, cancel := newGlobalEnvelope(t, 1)
	defer cancel()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/health/live", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/api/v1/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := wrapWithEnvelope(m, mux)

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health/live", nil)
		req.RemoteAddr = "203.0.113.1:1234"
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("expected health 200, got %d", rr.Code)
		}
	}

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics", nil)
		req.RemoteAddr = "203.0.113.1:1234"
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if i == 0 && rr.Code != http.StatusOK {
			t.Fatalf("expected first metrics request 200, got %d", rr.Code)
		}
		if i == 1 && rr.Code != http.StatusTooManyRequests {
			t.Fatalf("expected second metrics request 429, got %d", rr.Code)
		}
	}
}
