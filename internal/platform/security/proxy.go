package security

import (
	"net"
	"net/http"
	"net/netip"
	"strings"
)

// IsTrustedRemote reports whether r.RemoteAddr originates from one of the
// configured trusted proxy CIDRs. When cidrs is empty (the default), the
// function returns false so upstream-supplied headers (X-Forwarded-Proto,
// X-Forwarded-For) are never honored.
func IsTrustedRemote(r *http.Request, cidrs []netip.Prefix) bool {
	if r == nil || len(cidrs) == 0 {
		return false
	}
	addr, ok := remoteAddrToNetip(r.RemoteAddr)
	if !ok {
		return false
	}
	return prefixesContain(cidrs, addr)
}

// ClientIP returns the best-effort client IP for r. When IsTrustedRemote is
// true, the leftmost entry of X-Forwarded-For is parsed and returned;
// otherwise the parsed r.RemoteAddr is returned. The zero netip.Addr signals
// "no parseable IP" — callers should treat that as unknown identity.
func ClientIP(r *http.Request, cidrs []netip.Prefix) netip.Addr {
	if r == nil {
		return netip.Addr{}
	}
	if IsTrustedRemote(r, cidrs) {
		if first := leftmostForwardedFor(r.Header.Get("X-Forwarded-For")); first != "" {
			if addr, err := netip.ParseAddr(first); err == nil {
				return addr.Unmap()
			}
		}
	}
	addr, _ := remoteAddrToNetip(r.RemoteAddr)
	return addr
}

func leftmostForwardedFor(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}
	if idx := strings.Index(header, ","); idx >= 0 {
		return strings.TrimSpace(header[:idx])
	}
	return header
}

func remoteAddrToNetip(remoteAddr string) (netip.Addr, bool) {
	remoteAddr = strings.TrimSpace(remoteAddr)
	if remoteAddr == "" {
		return netip.Addr{}, false
	}
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	host = strings.Trim(host, "[]")
	addr, err := netip.ParseAddr(host)
	if err != nil {
		return netip.Addr{}, false
	}
	return addr.Unmap(), true
}

func prefixesContain(cidrs []netip.Prefix, addr netip.Addr) bool {
	addr = addr.Unmap()
	for _, prefix := range cidrs {
		if prefix.Contains(addr) {
			return true
		}
	}
	return false
}
