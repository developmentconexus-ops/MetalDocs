package security

import (
	"net/http"
	"net/netip"
	"net/url"
	"strings"

	"metaldocs/internal/platform/problem"
)

// OriginProtectionConfig configures OriginProtection.
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

// OriginProtection blocks cross-site cookie-authenticated state-changing
// requests (CSRF) by validating the Origin/Referer against same-origin or a
// trusted-origins allowlist.
type OriginProtection struct {
	enabled           bool
	cookieName        string
	trustedOrigins    map[string]struct{}
	trustedProxyCIDRs []netip.Prefix
}

// NewOriginProtection constructs an OriginProtection from cfg.
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

// Wrap returns next wrapped with same-origin enforcement for cookie-carrying
// state-changing requests, rejecting requests whose Origin/Referer fails the
// same-origin-or-trusted-origins check.
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
			writeOriginError(w, r)
			return
		}

		if refererOrigin := originFromReferer(r.Header.Get("Referer")); refererOrigin != "" {
			if p.isAllowedOrigin(r, refererOrigin) {
				next.ServeHTTP(w, r)
				return
			}
			writeOriginError(w, r)
			return
		}

		writeOriginError(w, r)
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

func writeOriginError(w http.ResponseWriter, r *http.Request) {
	problem.Respond(w, r, problem.New(http.StatusForbidden, problem.CodePermissionOriginForbidden, "Cross-site session request blocked"))
}
