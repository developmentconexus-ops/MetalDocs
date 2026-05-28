package security

import (
	"net/http"
	"net/netip"
	"net/url"
	"strings"
)

const traceLocalOrigin = "trace-local"

type OriginProtectionConfig struct {
	Enabled           bool
	SessionCookieName string
	TrustedOrigins    []string
	// TrustedProxyCIDRs lists upstream CIDRs whose X-Forwarded-Proto header
	// may be honored when computing the same-origin comparison scheme.
	// Empty (default) means no upstream is trusted — directly reachable
	// clients can never spoof the request scheme via a forwarded header.
	TrustedProxyCIDRs []netip.Prefix
}

type OriginProtection struct {
	enabled           bool
	cookieName        string
	trustedOrigins    map[string]struct{}
	trustedProxyCIDRs []netip.Prefix
}

func NewOriginProtection(cfg OriginProtectionConfig) *OriginProtection {
	trusted := make(map[string]struct{}, len(cfg.TrustedOrigins))
	for _, origin := range cfg.TrustedOrigins {
		normalized := normalizeOrigin(origin)
		if normalized != "" {
			trusted[normalized] = struct{}{}
		}
	}

	return &OriginProtection{
		enabled:           cfg.Enabled,
		cookieName:        strings.TrimSpace(cfg.SessionCookieName),
		trustedOrigins:    trusted,
		trustedProxyCIDRs: cfg.TrustedProxyCIDRs,
	}
}

func (p *OriginProtection) Wrap(next http.Handler) http.Handler {
	if !p.enabled || p.cookieName == "" {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !requiresOriginProtection(r) {
			next.ServeHTTP(w, r)
			return
		}

		if _, err := r.Cookie(p.cookieName); err != nil {
			next.ServeHTTP(w, r)
			return
		}

		if origin := normalizeOrigin(r.Header.Get("Origin")); origin != "" {
			if p.isAllowedOrigin(r, origin) {
				next.ServeHTTP(w, r)
				return
			}
			writeOriginError(w)
			return
		}

		if refererOrigin := originFromReferer(r.Header.Get("Referer")); refererOrigin != "" {
			if p.isAllowedOrigin(r, refererOrigin) {
				next.ServeHTTP(w, r)
				return
			}
			writeOriginError(w)
			return
		}

		writeOriginError(w)
	})
}

func requiresOriginProtection(r *http.Request) bool {
	if r == nil {
		return false
	}
	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func (p *OriginProtection) isAllowedOrigin(r *http.Request, origin string) bool {
	if origin == p.sameOrigin(r) {
		return true
	}
	_, ok := p.trustedOrigins[origin]
	return ok
}

// sameOrigin reconstructs the canonical Origin string for r. The request
// scheme is taken from r.TLS by default; X-Forwarded-Proto is honored only
// when the immediate upstream is in TrustedProxyCIDRs. This prevents a
// directly reachable client from spoofing https:// via a forged header to
// satisfy the trusted-origins allowlist.
func (p *OriginProtection) sameOrigin(r *http.Request) string {
	if r == nil {
		return ""
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if IsTrustedRemote(r, p.trustedProxyCIDRs) {
		if forwardedProto := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); forwardedProto != "" {
			scheme = strings.ToLower(forwardedProto)
		}
	}
	host := strings.TrimSpace(r.Host)
	if host == "" {
		return ""
	}
	return normalizeOrigin(scheme + "://" + host)
}

func originFromReferer(referer string) string {
	ref := strings.TrimSpace(referer)
	if ref == "" {
		return ""
	}
	parsed, err := url.Parse(ref)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}
	return normalizeOrigin(parsed.Scheme + "://" + parsed.Host)
}

func normalizeOrigin(origin string) string {
	value := strings.TrimSpace(origin)
	if value == "" {
		return ""
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}
	return strings.ToLower(parsed.Scheme) + "://" + strings.ToLower(parsed.Host)
}

func writeOriginError(w http.ResponseWriter) {
	writeAPIError(w, http.StatusForbidden, "AUTH_INVALID_ORIGIN", "Cross-site session request blocked", traceLocalOrigin)
}
