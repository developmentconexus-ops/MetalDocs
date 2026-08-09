package ratelimit

import (
	"fmt"
	"net/netip"
	"time"
)

// RouteKey identifies a rate-limited route for per-route quota lookup.
//
// Per-route quotas from spec §Rate limits. Values are requests-per-minute
// per user. Routes not listed here are unlimited.
//
// Quotas are compiled-in, not env-overridable: build a map via NewConfig, or
// use DefaultConfig for the spec defaults; main.go selects/wires the Config
// for each Middleware instance. The store *backend* (in-memory vs Redis) is
// the one env-controlled knob — see StoreConfig / LoadStoreConfig in
// store_config.go for METALDOCS_RATELIMIT_STORE and related vars.
type RouteKey string

// Route key values, keyed to their spec-defined per-route quotas.
const (
	// RouteAuthLogin is the pre-auth IP-keyed login limit (REQ-MW-5); it is
	// enforced before authentication, so its limiter instance always keys by
	// client IP (the user extractor returns "").
	RouteAuthLogin RouteKey = "auth_login"

	// RouteGlobalEnvelope is the post-authn blanket limiter that replaces the
	// legacy security.RateLimiter (F-05/D-04, Wave 2.8). It applies to every
	// non-health request and keys on authenticated user → trusted-proxy-resolved
	// IP, matching the old limiter's identity precedence. Default: 120 req/min.
	RouteGlobalEnvelope RouteKey = "global_envelope"

	RouteUploadsPresign  RouteKey = "uploads_presign"
	RouteAutosavePresign RouteKey = "autosave_presign"
	RouteAutosaveCommit  RouteKey = "autosave_commit"
	RouteDocumentsRender RouteKey = "documents_render"
	RouteExportPDF       RouteKey = "export_pdf"
)

// Config holds the rate-limit middleware's per-route quotas plus sweeper and
// store configuration.
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

	// Store, when non-nil, is used by New as the backing Store instead of
	// constructing a fresh in-memory store. This lets callers (main.go, or
	// tests) share one Store instance — e.g. one Redis-backed store — across
	// multiple Middleware instances. When nil, New builds a private
	// in-memory store per the SweepInterval/IdleThreshold/MaxEntries fields
	// above (unchanged default behavior).
	Store Store
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
		RouteGlobalEnvelope:  120,
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
