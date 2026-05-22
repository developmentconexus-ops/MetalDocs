package unit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"

	"metaldocs/internal/platform/ratelimit"
)

// noUser simulates a route where IAM middleware did NOT run (empty user id).
var noUser = func(r *http.Request) string { return "" }

func mustRatelimitConfig(t *testing.T, quota int, cidrs []netip.Prefix) ratelimit.Config {
	t.Helper()
	cfg, err := ratelimit.NewConfig(map[ratelimit.RouteKey]int{ratelimit.RouteExportPDF: quota})
	if err != nil {
		t.Fatalf("NewConfig: %v", err)
	}
	cfg.TrustedProxyCIDRs = cidrs
	return cfg
}

// TestIPFallback_MisorderedRoute_TwoXFFBucketedSeparately verifies that when
// userExtractor returns "" (misordered middleware — no IAM) and the upstream
// is a trusted proxy, two distinct X-Forwarded-For clients are rate-limited
// in separate buckets.
func TestIPFallback_MisorderedRoute_TwoXFFBucketedSeparately(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	trustedCIDR, _ := netip.ParsePrefix("10.0.0.0/24")
	cfg := mustRatelimitConfig(t, 1, []netip.Prefix{trustedCIDR})
	mw := ratelimit.New(ctx, cfg)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })
	h := mw.Limit(ratelimit.RouteExportPDF, noUser, next)

	// Client A (XFF 198.51.100.7) burns its 1 token.
	rrA := httptest.NewRecorder()
	reqA := httptest.NewRequest("POST", "/x", nil)
	reqA.RemoteAddr = "10.0.0.5:12345" // trusted proxy
	reqA.Header.Set("X-Forwarded-For", "198.51.100.7")
	h.ServeHTTP(rrA, reqA)
	if rrA.Code != 204 {
		t.Fatalf("client A first: want 204, got %d", rrA.Code)
	}

	// Client A is now throttled.
	rrA2 := httptest.NewRecorder()
	reqA2 := httptest.NewRequest("POST", "/x", nil)
	reqA2.RemoteAddr = "10.0.0.5:12345"
	reqA2.Header.Set("X-Forwarded-For", "198.51.100.7")
	h.ServeHTTP(rrA2, reqA2)
	if rrA2.Code != http.StatusTooManyRequests {
		t.Fatalf("client A second: want 429, got %d", rrA2.Code)
	}

	// Client B (XFF 198.51.100.8) has its own bucket — still allowed.
	rrB := httptest.NewRecorder()
	reqB := httptest.NewRequest("POST", "/x", nil)
	reqB.RemoteAddr = "10.0.0.5:12346"
	reqB.Header.Set("X-Forwarded-For", "198.51.100.8")
	h.ServeHTTP(rrB, reqB)
	if rrB.Code != 204 {
		t.Fatalf("client B: want 204 (separate bucket), got %d", rrB.Code)
	}

	// Two distinct IP keys in map.
	if got := mw.Size(); got != 2 {
		t.Fatalf("want 2 limiter entries (one per client IP), got %d", got)
	}
}

// TestIPFallback_UntrustedProxy_BucketsByRemoteAddr verifies that when the
// upstream is NOT a trusted proxy, X-Forwarded-For is ignored and all requests
// from the same RemoteAddr share one bucket — a forged XFF cannot split buckets.
func TestIPFallback_UntrustedProxy_BucketsByRemoteAddr(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// No trusted CIDRs → RemoteAddr always used.
	cfg := mustRatelimitConfig(t, 1, nil)
	mw := ratelimit.New(ctx, cfg)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })
	h := mw.Limit(ratelimit.RouteExportPDF, noUser, next)

	// First request from same RemoteAddr — allowed.
	rr1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("POST", "/x", nil)
	req1.RemoteAddr = "203.0.113.5:9999"
	req1.Header.Set("X-Forwarded-For", "1.2.3.4") // forged — should be ignored
	h.ServeHTTP(rr1, req1)
	if rr1.Code != 204 {
		t.Fatalf("first: want 204, got %d", rr1.Code)
	}

	// Second request — same RemoteAddr, different forged XFF — same bucket → 429.
	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("POST", "/x", nil)
	req2.RemoteAddr = "203.0.113.5:9999"
	req2.Header.Set("X-Forwarded-For", "5.6.7.8") // still ignored
	h.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusTooManyRequests {
		t.Fatalf("second: want 429 (same RemoteAddr bucket), got %d", rr2.Code)
	}

	if got := mw.Size(); got != 1 {
		t.Fatalf("want 1 limiter entry (RemoteAddr bucket), got %d", got)
	}
}

// TestIPFallback_UnparseableIP_FailClosed verifies that when RemoteAddr is
// unparseable and no XFF is available, the request is denied (fail-closed),
// not bypassed (the original H2 fail-open behaviour).
func TestIPFallback_UnparseableIP_FailClosed(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	cfg := mustRatelimitConfig(t, 60, nil)
	mw := ratelimit.New(ctx, cfg)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })
	h := mw.Limit(ratelimit.RouteExportPDF, noUser, next)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/x", nil)
	req.RemoteAddr = "not-an-ip" // unparseable
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("unparseable IP: want 429 (fail-closed), got %d (H2 regression — was fail-open)", rr.Code)
	}
}

// TestIPFallback_CapEnforced_FailClosed verifies that when the limiter map
// reaches MaxEntries, new IP keys are denied fail-closed (not bypassed).
func TestIPFallback_CapEnforced_FailClosed(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	cfg, err := ratelimit.NewConfig(map[ratelimit.RouteKey]int{ratelimit.RouteExportPDF: 60})
	if err != nil {
		t.Fatalf("NewConfig: %v", err)
	}
	cfg.MaxEntries = 2
	mw := ratelimit.New(ctx, cfg)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })
	h := mw.Limit(ratelimit.RouteExportPDF, noUser, next)

	// Fill the map: two distinct IPs succeed.
	for i, ip := range []string{"1.0.0.1", "1.0.0.2"} {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/x", nil)
		req.RemoteAddr = ip + ":1234"
		h.ServeHTTP(rr, req)
		if rr.Code != 204 {
			t.Fatalf("fill[%d]: want 204, got %d", i, rr.Code)
		}
	}

	// Third distinct IP — map full → fail-closed 429.
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/x", nil)
	req.RemoteAddr = "1.0.0.3:1234"
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("overflow: want 429 (cap fail-closed), got %d", rr.Code)
	}
}
