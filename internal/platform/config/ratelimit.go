package config

import (
	"fmt"
	"net/netip"
	"os"
	"strconv"
	"strings"
)

type RateLimitConfig struct {
	Enabled       bool
	WindowSeconds int
	MaxRequests   int
	// TrustedProxyCIDRs governs whether X-Forwarded-For may be honored when
	// keying the rate-limit bucket for unauthenticated traffic. Empty means
	// no upstream is trusted, and RemoteAddr is the only IP source.
	TrustedProxyCIDRs []netip.Prefix
}

func LoadRateLimitConfig() (RateLimitConfig, error) {
	enabled := strings.EqualFold(strings.TrimSpace(os.Getenv("METALDOCS_RATE_LIMIT_ENABLED")), "true")

	window := 60
	if raw := strings.TrimSpace(os.Getenv("METALDOCS_RATE_LIMIT_WINDOW_SECONDS")); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n <= 0 {
			return RateLimitConfig{}, fmt.Errorf("invalid METALDOCS_RATE_LIMIT_WINDOW_SECONDS: %s", raw)
		}
		window = n
	}

	maxReq := 120
	if raw := strings.TrimSpace(os.Getenv("METALDOCS_RATE_LIMIT_MAX_REQUESTS")); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n <= 0 {
			return RateLimitConfig{}, fmt.Errorf("invalid METALDOCS_RATE_LIMIT_MAX_REQUESTS: %s", raw)
		}
		maxReq = n
	}

	trustedProxyCIDRs, err := LoadTrustedProxyCIDRs()
	if err != nil {
		return RateLimitConfig{}, err
	}

	return RateLimitConfig{
		Enabled:           enabled,
		WindowSeconds:     window,
		MaxRequests:       maxReq,
		TrustedProxyCIDRs: trustedProxyCIDRs,
	}, nil
}
