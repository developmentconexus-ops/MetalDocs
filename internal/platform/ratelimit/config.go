package ratelimit

import (
    "net/netip"
    "time"
)

// Per-route quotas from spec §Rate limits. Values are requests-per-minute
// per user. Routes not listed here are unlimited.
//
// Envvar overrides: METALDOCS_RLIMIT_<ROUTE_KEY> (e.g. EXPORT_PDF=30).
type RouteKey string

const (
    RouteUploadsPresign   RouteKey = "uploads_presign"
    RouteAutosavePresign  RouteKey = "autosave_presign"
    RouteAutosaveCommit   RouteKey = "autosave_commit"
    RouteDocumentsRender  RouteKey = "documents_render"
    RouteExportPDF        RouteKey = "export_pdf"
)

type Config struct {
    Quotas map[RouteKey]int // req/min

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

func DefaultConfig() Config {
    return Config{
        Quotas: map[RouteKey]int{
            RouteUploadsPresign:  60,
            RouteAutosavePresign: 60,
            RouteAutosaveCommit:  30,
            RouteDocumentsRender: 30,
            RouteExportPDF:       20,
        },
    }
}
