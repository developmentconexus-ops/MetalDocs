package config

import (
	"fmt"
	"net/netip"
	"os"
	"strings"
)

// LoadTrustedProxyCIDRs parses METALDOCS_TRUSTED_PROXY_CIDRS (comma-separated
// CIDR list) into a slice of netip.Prefix. Empty env returns a nil slice,
// which downstream consumers must treat as "no upstream trusted" (fail-closed).
func LoadTrustedProxyCIDRs() ([]netip.Prefix, error) {
	return ParseTrustedProxyCIDRs(strings.TrimSpace(os.Getenv("METALDOCS_TRUSTED_PROXY_CIDRS")))
}

// ParseTrustedProxyCIDRs is exposed for tests and config validation. It mirrors
// the semantics of LoadTrustedProxyCIDRs but accepts the raw value directly.
func ParseTrustedProxyCIDRs(raw string) ([]netip.Prefix, error) {
	if raw == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	out := make([]netip.Prefix, 0, len(parts))
	for _, p := range parts {
		token := strings.TrimSpace(p)
		if token == "" {
			continue
		}
		prefix, err := netip.ParsePrefix(token)
		if err != nil {
			return nil, fmt.Errorf("invalid METALDOCS_TRUSTED_PROXY_CIDRS entry %q: %w", token, err)
		}
		out = append(out, prefix)
	}
	return out, nil
}
