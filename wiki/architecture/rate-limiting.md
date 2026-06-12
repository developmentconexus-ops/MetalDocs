# Architecture: Rate Limiting

> **Last verified:** 2026-06-12 (Wave 2.8: `security.RateLimiter` deleted; `platform/ratelimit` fully active in prod chain + documents routes — F-05/D-04)
> **Scope:** the single `platform/ratelimit.Middleware` that now covers all rate-limiting: global envelope + per-route quota on expensive write paths.
> **Out of scope:** circuit-breakers on downstream calls, infrastructure-layer (CDN / WAF) limiting, per-tenant quota enforcement, billing-driven throttling.
> **Key files:**
> - `internal/platform/ratelimit/middleware.go`   token-bucket limiter (`ratelimit.Middleware`) — global envelope + per-route
> - `internal/platform/ratelimit/config.go`       `RouteKey` catalog + `DefaultConfig` quotas
> - `apps/api/cmd/metaldocs-api/main.go`          wires global envelope + pre-auth login limiter into the chain; documents routes via `RegisterRoutesWithRateLimit`
> - `internal/modules/documents/delivery/http/handler.go`        per-route `rl.Limit(...)` wiring for autosave
> - `internal/modules/documents/delivery/http/export_handler.go` per-route `rl.Limit(...)` wiring for PDF export

---

## 1. One limiter, two operating modes

MetalDocs now uses a single `ratelimit.Middleware` for all rate-limiting. The same instance runs in two modes:

| Mode | Scope | Keyed on | Algorithm | Quota source |
|---|---|---|---|---|
| **Global envelope** (`GlobalEnvelopeWrap`) | Every non-health request — the outer mux wrapper at the `rate_limit` chain slot | `userExtractor(*http.Request)` → `user:<UserID>` if authenticated, else `ip:<ClientIP>` via trusted-proxy resolution (CIDR-aware) | Token bucket (`golang.org/x/time/rate`) — 120 req/min, burst = 120 | `RouteGlobalEnvelope` quota in `DefaultConfig()` |
| **Per-route** (`Limit`) | Small set of expensive write paths in the documents module | Same `userExtractor` → user → IP fallback | Token bucket — per-route quota, burst = quota | Per-`RouteKey` quotas in `DefaultConfig()` |

The global envelope is the cheap blanket guard against abusive clients. The per-route limiter is the targeted guard on operations that are individually expensive (storage uploads, PDF rendering) where a small number of legitimate-looking calls can still be damaging.

**Dead env vars (Wave 2.8):** `METALDOCS_RATE_LIMIT_ENABLED`, `METALDOCS_RATE_LIMIT_WINDOW_SECONDS`, `METALDOCS_RATE_LIMIT_MAX_REQUESTS` were parsed exclusively by the deleted `security.RateLimiter`. They are now dead configuration. Operators may remove them from `.env` files; the binary ignores them. The `config.LoadRateLimitConfig()` function and `config.RateLimitConfig` type remain in `internal/platform/config/ratelimit.go` but are no longer called at startup.

## 2. Where each one is wired

### 2.1 Global envelope — `GlobalEnvelopeWrap`

Constructed once at startup, passed into the `rate_limit` chain slot at the outermost mux wrapper position (after `presence_bump`, before the mux itself):

```
apps/api/cmd/metaldocs-api/main.go
    globalRateCfg := ratelimit.DefaultConfig()          // includes RouteGlobalEnvelope: 120
    globalRateCfg.TrustedProxyCIDRs = authCfg.TrustedProxyCIDRs
    globalLimiter := ratelimit.New(ctx, globalRateCfg)

    handler := buildChain(mux, apiChain(
        ...,
        func(next http.Handler) http.Handler { return globalLimiter.GlobalEnvelopeWrap(userIDExtractor, next) },
    ))
```

- Health probes (`/api/v1/health/live`, `/api/v1/health/ready`) are skipped inside `GlobalEnvelopeWrap`.
- Identity preference order: `authdomain.CurrentUserFromContext` → `iamdomain.UserIDFromContext` → trusted-proxy-resolved client IP → fail-closed 429.
- On rejection: `HTTP 429` with `Retry-After` (seconds until token refills) and an RFC 9457 `application/problem+json` body, code `RATE_LIMITED`.

### 2.2 Per-route quota — `Limit` on documents routes

**Status: active in production (Wave 2.8).** Documents routes are registered via `docMod.RegisterRoutesWithRateLimit(mux, globalLimiter, userIDExtractor)`, which fans out to per-handler `rl.Limit(routeKey, userFn, h)` for the routes below. Every other documents route is registered as a plain `mux.HandleFunc` and is only covered by the global envelope:

| Route | Method | `RouteKey` | Default quota (req/min) | Callsite |
|---|---|---|---|---|
| `/api/v1/documents/{id}/autosave/presign` | `POST` | `RouteAutosavePresign` | 60 | [`handler.go:157-160`](../../internal/modules/documents/delivery/http/handler.go) |
| `/api/v1/documents/{id}/autosave/commit`  | `POST` | `RouteAutosaveCommit`  | 30 | [`handler.go:161-164`](../../internal/modules/documents/delivery/http/handler.go) |
| `/api/v1/documents/{id}/export/pdf`       | `POST` | `RouteExportPDF`       | 20 | [`export_handler.go:44-48`](../../internal/modules/documents/delivery/http/export_handler.go) |

`RouteUploadsPresign` and `RouteDocumentsRender` exist in the catalog (`config.go`) but have no `rl.Limit` callsite yet — they are reserved for future endpoints. Adding a new per-route limit means (a) declaring a `RouteKey` constant, (b) giving it a `Quotas` entry in `DefaultConfig`, and (c) wrapping the handler with `rl.Limit(...)`.

## 3. Bounded-memory contract (finding C2)

The limiter keeps per-identity `*rate.Limiter` entries in a `sync.Map`. The fix:

### 3.1 `ratelimit.Middleware`

- Each `*rate.Limiter` is wrapped in `limiterEntry { lim *rate.Limiter; lastAccess atomic.Int64 }`. `lastAccess` is updated lock-free on every admission.
- `New(ctx, cfg)` starts a `time.NewTicker(SweepInterval)` background goroutine. On each tick it scans `m.limiters` (a `sync.Map`) and deletes entries whose `lastAccess` is older than `IdleThreshold`; a single `slog.Warn` per non-empty sweep batch carries the count and remaining size.
- The sweeper goroutine exits on `<-ctx.Done()`. A `sync.WaitGroup` exposes `Wait()` so callers (and tests) can prove no goroutine leak.
- Defaults: `SweepInterval = time.Minute`, `IdleThreshold = 2 * time.Minute`. Override via `Config.SweepInterval` / `Config.IdleThreshold`.
- Race-safe first-hit: `m.limiters.Load` first, only `LoadOrStore` on miss — fixes the eager `rate.NewLimiter` allocation defect (review **M5**).
- Hard cap: `Config.MaxEntries` (default `100_000`). New keys past cap denied fail-closed 429; atomic counter keeps cardinality consistent with sweep decrement. Mirrors `security.RateLimiter` contract.

#### IP-fallback contract (H2 fix)

`Limit` no longer bypasses when `userExtractor` returns `""`. Identity resolution:

1. `userExtractor(r) != ""` → key `<route>:user:<id>`
2. `userExtractor(r) == ""` → log `DebugContext` + resolve `security.ClientIP(r, Config.TrustedProxyCIDRs)` → key `<route>:ip:<addr>`
3. ClientIP invalid (unparseable RemoteAddr, no XFF) → fail-closed `429`

`Config.TrustedProxyCIDRs` is the **single source of truth** — same env var `METALDOCS_TRUSTED_PROXY_CIDRS` loaded via `config.LoadTrustedProxyCIDRs()` as the global limiter. Set at startup, passed into `ratelimit.Config` before constructing the middleware. When empty (default): XFF headers ignored, fallback keys on parsed `RemoteAddr`.

The `user:` / `ip:` literal prefixes namespace the two keyspaces to prevent a user id that happens to look like an IP from colliding with the IP bucket.

### 3.2 Regression tests

| Test | Asserts |
|---|---|
| [`TestRateLimiterSweepsExpiredWindowEntries`](../../tests/unit/rate_limit_eviction_test.go) | Global envelope: 2_000 distinct identities drop to 0 after idle-clock sweep; one new admission creates exactly 1 entry. |
| [`TestRateLimiterFailsClosedOnMapOverflow`](../../tests/unit/rate_limit_eviction_test.go) | Global envelope: with `MaxEntries=8`, the 9th distinct identity returns 429 with `Retry-After`; existing identities still admitted. |
| [`TestRateLimiterBlocksWhenLimitExceeded`](../../tests/unit/rate_limit_middleware_test.go) | Global envelope: quota=1, second request from same user is 429. |
| [`TestRateLimiterIsolatedByIdentity`](../../tests/unit/rate_limit_middleware_test.go) | Global envelope: two distinct users each get their own bucket. |
| [`TestRateLimiterSkipsHealthOnly`](../../tests/unit/rate_limit_middleware_test.go) | Global envelope: `/api/v1/health/live` skipped; `/api/v1/metrics` counted. |
| [`TestRateLimiterIsolatesAnonByXFFWhenBehindTrustedProxy`](../../tests/unit/rate_limit_middleware_test.go) | Global envelope: trusted proxy → two distinct XFF clients bucket separately. |
| [`TestRateLimiterIgnoresXFFFromUntrustedSource`](../../tests/unit/rate_limit_middleware_test.go) | Global envelope: untrusted proxy → XFF ignored; RemoteAddr used; forged header cannot split buckets. |
| [`TestSweep_EvictsIdleEntries`](../../internal/platform/ratelimit/eviction_test.go) | Middleware: 10_000 unique identities collapse to 0 after deterministic-clock sweep. |
| [`TestSweeper_ExitsOnContextCancel`](../../internal/platform/ratelimit/eviction_test.go) | Middleware: sweeper goroutine returns within 1s of `cancel()`, `runtime.NumGoroutine()` returns to baseline. |
| [`TestLimit_LimiterReusedAcrossRequests`](../../internal/platform/ratelimit/eviction_test.go) | Middleware: quota=1, second request on the same key is 429 (proves limiter persisted, not freshly allocated). |
| [`TestLimit_RaceUnderConcurrentFirstHit`](../../internal/platform/ratelimit/eviction_test.go) | Middleware: 64 goroutines racing 4 distinct keys yield exactly 4 surviving entries. |
| [`TestIPFallback_MisorderedRoute_TwoXFFBucketedSeparately`](../../tests/unit/ratelimit_ip_fallback_test.go) | H2 fix: misordered route (no IAM), trusted proxy → two distinct XFF clients bucket separately. |
| [`TestIPFallback_UntrustedProxy_BucketsByRemoteAddr`](../../tests/unit/ratelimit_ip_fallback_test.go) | H2 fix: untrusted proxy → XFF ignored, forged header cannot split buckets; RemoteAddr used. |
| [`TestIPFallback_UnparseableIP_FailClosed`](../../tests/unit/ratelimit_ip_fallback_test.go) | H2 fix: unparseable RemoteAddr + no XFF → 429 fail-closed (was fail-open). |
| [`TestIPFallback_CapEnforced_FailClosed`](../../tests/unit/ratelimit_ip_fallback_test.go) | H2 fix: IP keys respect MaxEntries cap; overflow denied 429, not bypassed. |

Verification commands:

```
go vet ./...
go test -race ./internal/platform/...
go test -race -run TestRateLimit -count 100 ./tests/unit/...
```

## 4. Prior duplication (resolved — Wave 2.8)

Wave 2.8 (F-05/D-04) deleted `internal/platform/security/ratelimit.go` and collapsed the two-limiter design into `platform/ratelimit.Middleware`. The chosen direction (recorded in the pre-Wave §4 as "pick-one"):

- `ratelimit.Middleware` retained: token-bucket, per-route quotas, observable sweeper lifecycle, domain-free (REQ-TOP-2 compliant).
- Identity-resolution path (`authdomain.CurrentUser → iamdomain.UserID → client IP`) lifted into a `userIDExtractor` closure injected at the composition root (`main.go`) — no domain imports needed inside the platform package.
- `RouteGlobalEnvelope` added to `DefaultConfig()` at 120 req/min (same as old fixed-window default of 120 req / 60s), replacing the blanket `Wrap()` slot in the middleware chain.

Dead env vars: `METALDOCS_RATE_LIMIT_ENABLED`, `METALDOCS_RATE_LIMIT_WINDOW_SECONDS`, `METALDOCS_RATE_LIMIT_MAX_REQUESTS` — operators may remove them. The `config.LoadRateLimitConfig()` function and `config.RateLimitConfig` type are retained in `internal/platform/config/ratelimit.go` but are no longer called at startup (dead config, not dead code — removing the type is a separate cleanup if desired).
