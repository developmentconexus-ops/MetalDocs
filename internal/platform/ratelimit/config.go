package ratelimit

import (
	"fmt"
	"net/netip"
	"time"
)

// Per-route quotas from spec §Rate limits. Values are requests-per-minute
// per user. Routes not listed here are unlimited.
//
// Envvar overrides: METALDOCS_RLIMIT_<ROUTE_KEY> (e.g. EXPORT_PDF=30).
type RouteKey string

const (
	RouteUploadsPresign  RouteKey = "uploads_presign"
	RouteAutosavePresign RouteKey = "autosave_presign"
	RouteAutosaveCommit  RouteKey = "autosave_commit"
	RouteDocumentsRender RouteKey = "documents_render"
	RouteExportPDF       RouteKey = "export_pdf"
)

type Config struct {
	quotas map[RouteKey]int // req/min; validated >= 1 by NewConfig

	// Optional sweeper schedule. Zero values fall back to 1m sweep and 2m
	// idle threshold; tests use small values to exercise eviction quickly.
	SweepInterval time.Duration
	IdleThreshold time.Duration

	// MaxEntries caps the limiter map. Zero falls back to
	// defaultMaxLimiterEntries. New keys past the cap fail-closed (429),
	// matching the global limiter's contract from C2.
	MaxEntries int

	// TrustedProxyCIDRs is the single source of truth for upstream proxies
	// trusted to set X-Forwarded-For. Loaded once at startup from
	// METALDOCS_TRUSTED_PROXY_CIDRS via config.LoadTrustedProxyCIDRs.
	// Drives the IP-fallback path when userExtractor returns "" (H2).
	TrustedProxyCIDRs []netip.Prefix
}

// NewConfig validates that every quota value is >= 1 and returns an immutable
// Config. Returns an error for any zero or negative value.
func NewConfig(q map[RouteKey]int) (Config, error) {
	for k, v := range q {
		if v < 1 {
			return Config{}, fmt.Errorf("ratelimit: quota for %q must be >= 1, got %d", k, v)
		}
	}
	copied := make(map[RouteKey]int, len(q))
	for k, v := range q {
		copied[k] = v
	}
	return Config{quotas: copied}, nil
}

// QuotaFor returns the configured quota (req/min) for a route key and whether
// a quota exists for that key.
func (c Config) QuotaFor(k RouteKey) (int, bool) {
	v, ok := c.quotas[k]
	return v, ok
}

// DefaultConfig returns a validated Config with the default per-route quotas.
// Panics only if the built-in defaults are invalid (static; cannot happen at runtime).
func DefaultConfig() Config {
	cfg, err := NewConfig(map[RouteKey]int{
		RouteUploadsPresign:  60,
		RouteAutosavePresign: 60,
		RouteAutosaveCommit:  30,
		RouteDocumentsRender: 30,
		RouteExportPDF:       20,
	})
	if err != nil {
		panic("ratelimit: DefaultConfig has invalid static defaults: " + err.Error())
	}
	return cfg
}
