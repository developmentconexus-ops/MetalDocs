package unit

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"

	"metaldocs/internal/platform/security"
)

func mustPrefix(t *testing.T, raw string) netip.Prefix {
	t.Helper()
	p, err := netip.ParsePrefix(raw)
	if err != nil {
		t.Fatalf("invalid prefix %q: %v", raw, err)
	}
	return p
}

func TestOriginProtectionAllowsTrustedOriginWithSessionCookie(t *testing.T) {
	protector := security.NewOriginProtection(security.OriginProtectionConfig{
		Enabled:           true,
		SessionCookieName: "metaldocs_session",
		TrustedOrigins:    []string{"http://127.0.0.1:4173"},
	})

	handler := protector.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:8080/api/v1/auth/change-password", nil)
	req.AddCookie(&http.Cookie{Name: "metaldocs_session", Value: "session-token"})
	req.Header.Set("Origin", "http://127.0.0.1:4173")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected trusted origin request to pass, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestOriginProtectionBlocksUnsafeRequestWithoutOrigin(t *testing.T) {
	protector := security.NewOriginProtection(security.OriginProtectionConfig{
		Enabled:           true,
		SessionCookieName: "metaldocs_session",
		TrustedOrigins:    []string{"http://127.0.0.1:4173"},
	})

	handler := protector.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:8080/api/v1/auth/change-password", nil)
	req.AddCookie(&http.Cookie{Name: "metaldocs_session", Value: "session-token"})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected missing origin to be rejected, got %d body=%s", rr.Code, rr.Body.String())
	}
}

// TestOriginProtectionBlocksSpoofedXForwardedProtoFromUntrustedSource confirms
// the C1 regression cannot recur: a directly reachable client (RemoteAddr
// outside any trusted CIDR) cannot upgrade the request scheme via the header
// to match an https:// trusted origin.
func TestOriginProtectionBlocksSpoofedXForwardedProtoFromUntrustedSource(t *testing.T) {
	protector := security.NewOriginProtection(security.OriginProtectionConfig{
		Enabled:           true,
		SessionCookieName: "metaldocs_session",
		TrustedOrigins:    nil,
		TrustedProxyCIDRs: []netip.Prefix{mustPrefix(t, "10.0.0.0/8")},
	})

	handler := protector.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodPost, "http://internal-host:8080/api/v1/auth/change-password", nil)
	req.RemoteAddr = "203.0.113.7:5555"
	req.AddCookie(&http.Cookie{Name: "metaldocs_session", Value: "session-token"})
	req.Header.Set("Origin", "https://internal-host:8080")
	req.Header.Set("X-Forwarded-Proto", "https")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("spoofed X-Forwarded-Proto from untrusted source must be rejected, got %d body=%s", rr.Code, rr.Body.String())
	}
}

// TestOriginProtectionHonorsXForwardedProtoFromTrustedProxy verifies the
// legitimate reverse-proxy path: when the upstream IP is inside the trusted
// CIDR, the forwarded scheme is honored so HTTPS terminating in front of the
// API still matches same-origin against the https://host Origin.
func TestOriginProtectionHonorsXForwardedProtoFromTrustedProxy(t *testing.T) {
	protector := security.NewOriginProtection(security.OriginProtectionConfig{
		Enabled:           true,
		SessionCookieName: "metaldocs_session",
		TrustedOrigins:    nil,
		TrustedProxyCIDRs: []netip.Prefix{mustPrefix(t, "10.0.0.0/8")},
	})

	handler := protector.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodPost, "http://internal-host:8080/api/v1/auth/change-password", nil)
	req.RemoteAddr = "10.0.0.5:443"
	req.AddCookie(&http.Cookie{Name: "metaldocs_session", Value: "session-token"})
	req.Header.Set("Origin", "https://internal-host:8080")
	req.Header.Set("X-Forwarded-Proto", "https")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("trusted-proxy X-Forwarded-Proto must be honored, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestOriginProtectionAllowsGetWithoutOrigin(t *testing.T) {
	protector := security.NewOriginProtection(security.OriginProtectionConfig{
		Enabled:           true,
		SessionCookieName: "metaldocs_session",
		TrustedOrigins:    []string{"http://127.0.0.1:4173"},
	})

	handler := protector.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:8080/api/v1/auth/me", nil)
	req.AddCookie(&http.Cookie{Name: "metaldocs_session", Value: "session-token"})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected safe method to bypass origin protection, got %d body=%s", rr.Code, rr.Body.String())
	}
}
