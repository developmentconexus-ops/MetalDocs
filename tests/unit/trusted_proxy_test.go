package unit

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"

	"metaldocs/internal/platform/config"
	"metaldocs/internal/platform/security"
)

func mustParsePrefixes(t *testing.T, raw ...string) []netip.Prefix {
	t.Helper()
	out := make([]netip.Prefix, 0, len(raw))
	for _, r := range raw {
		p, err := netip.ParsePrefix(r)
		if err != nil {
			t.Fatalf("invalid prefix %q: %v", r, err)
		}
		out = append(out, p)
	}
	return out
}

func TestIsTrustedRemote_HonorsCIDRMatch(t *testing.T) {
	cidrs := mustParsePrefixes(t, "10.0.0.0/8")
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	req.RemoteAddr = "10.1.2.3:54321"

	if !security.IsTrustedRemote(req, cidrs) {
		t.Fatal("expected 10.1.2.3 to be inside 10.0.0.0/8")
	}
}

func TestIsTrustedRemote_RejectsOutsideCIDR(t *testing.T) {
	cidrs := mustParsePrefixes(t, "10.0.0.0/8")
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	req.RemoteAddr = "203.0.113.7:443"

	if security.IsTrustedRemote(req, cidrs) {
		t.Fatal("expected 203.0.113.7 to be untrusted")
	}
}

func TestIsTrustedRemote_EmptyCIDRsFailsClosed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	req.RemoteAddr = "10.0.0.1:1234"

	if security.IsTrustedRemote(req, nil) {
		t.Fatal("empty CIDRs must never trust any remote")
	}
}

func TestIsTrustedRemote_IPv6Prefix(t *testing.T) {
	cidrs := mustParsePrefixes(t, "fd00::/8")
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	req.RemoteAddr = "[fd12:3456:789a::1]:5000"

	if !security.IsTrustedRemote(req, cidrs) {
		t.Fatal("expected IPv6 fd12::1 to be inside fd00::/8")
	}
}

func TestIsTrustedRemote_MalformedRemoteAddr(t *testing.T) {
	cidrs := mustParsePrefixes(t, "10.0.0.0/8")
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	req.RemoteAddr = "not-an-address"

	if security.IsTrustedRemote(req, cidrs) {
		t.Fatal("malformed remote must be treated as untrusted")
	}
}

func TestClientIP_TrustedProxyHonorsLeftmostXFF(t *testing.T) {
	cidrs := mustParsePrefixes(t, "10.0.0.0/8")
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	req.RemoteAddr = "10.0.0.5:80"
	req.Header.Set("X-Forwarded-For", "198.51.100.7, 10.0.0.5")

	addr := security.ClientIP(req, cidrs)
	if !addr.IsValid() || addr.String() != "198.51.100.7" {
		t.Fatalf("expected leftmost XFF 198.51.100.7, got %v", addr)
	}
}

func TestClientIP_UntrustedSourceIgnoresXFF(t *testing.T) {
	cidrs := mustParsePrefixes(t, "10.0.0.0/8")
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	req.RemoteAddr = "203.0.113.7:9000"
	req.Header.Set("X-Forwarded-For", "198.51.100.7")

	addr := security.ClientIP(req, cidrs)
	if !addr.IsValid() || addr.String() != "203.0.113.7" {
		t.Fatalf("expected RemoteAddr 203.0.113.7, got %v", addr)
	}
}

func TestClientIP_EmptyCIDRsAlwaysUseRemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	req.RemoteAddr = "10.0.0.5:80"
	req.Header.Set("X-Forwarded-For", "198.51.100.7")

	addr := security.ClientIP(req, nil)
	if !addr.IsValid() || addr.String() != "10.0.0.5" {
		t.Fatalf("expected RemoteAddr fallback when no CIDRs, got %v", addr)
	}
}

func TestClientIP_MultiHopXFFPicksLeftmost(t *testing.T) {
	cidrs := mustParsePrefixes(t, "10.0.0.0/8")
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	req.RemoteAddr = "10.0.0.5:80"
	req.Header.Set("X-Forwarded-For", "198.51.100.7, 10.9.9.9, 10.0.0.5")

	addr := security.ClientIP(req, cidrs)
	if !addr.IsValid() || addr.String() != "198.51.100.7" {
		t.Fatalf("expected leftmost XFF across multi-hop chain, got %v", addr)
	}
}

func TestClientIP_IPv6XFF(t *testing.T) {
	cidrs := mustParsePrefixes(t, "fd00::/8")
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	req.RemoteAddr = "[fd00::1]:443"
	req.Header.Set("X-Forwarded-For", "2001:db8::1")

	addr := security.ClientIP(req, cidrs)
	if !addr.IsValid() || addr.String() != "2001:db8::1" {
		t.Fatalf("expected IPv6 XFF preserved, got %v", addr)
	}
}

func TestParseTrustedProxyCIDRs_RejectsMalformed(t *testing.T) {
	if _, err := config.ParseTrustedProxyCIDRs("10.0.0.0/8,not-a-cidr"); err == nil {
		t.Fatal("expected error on malformed entry")
	}
}

func TestParseTrustedProxyCIDRs_ParsesMixedIPv4IPv6(t *testing.T) {
	got, err := config.ParseTrustedProxyCIDRs("10.0.0.0/8, fd00::/8 ,127.0.0.1/32")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 prefixes, got %d (%v)", len(got), got)
	}
}

func TestParseTrustedProxyCIDRs_EmptyReturnsNil(t *testing.T) {
	got, err := config.ParseTrustedProxyCIDRs("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}
