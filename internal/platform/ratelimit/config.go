package ratelimit

import "time"

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
