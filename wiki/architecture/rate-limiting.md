# Architecture: Rate Limiting

> **Last verified:** 2026-05-22 (commit H2 fix)
> **Scope:** the two parallel rate-limiter implementations currently live in the API, what each one keys on, which routes each one protects, and the bounded-memory contract both honor (review finding **C2**).
> **Out of scope:** circuit-breakers on downstream calls, infrastructure-layer (CDN / WAF) limiting, per-tenant quota enforcement, billing-driven throttling. Merging the two limiters is intentionally **deferred** to a separate refactor — see "Known duplication" below.
> **Key files:**
> - `internal/platform/security/ratelimit.go`     global per-identity envelope limiter (`security.RateLimiter`)
> - `internal/platform/ratelimit/middleware.go`   per-route token-bucket limiter (`ratelimit.Middleware`)
> - `internal/platform/ratelimit/config.go`       `RouteKey` catalog + `DefaultConfig` quotas
> - `internal/platform/config/ratelimit.go`       `RateLimitConfig` (env-driven, drives the global limiter)
> - `apps/api/cmd/metaldocs-api/main.go`          wires both limiters into the middleware chain and the documents module
> - `internal/modules/documents/delivery/http/handler.go`        per-route `rl.Limit(...)` wiring for autosave
> - `internal/modules/documents/delivery/http/export_handler.go` per-route `rl.Limit(...)` wiring for PDF export

---

## 1. Two limiters, two different jobs

MetalDocs ships two distinct rate-limit middlewares. They are **not** redundant — they answer different questions:

| Limiter | Scope | Keyed on | Algorithm | Quota source |
|---|---|---|---|---|
| `security.RateLimiter` (`Wrap`) | **Global envelope** — every non-health request | `user:<UserID>` if authenticated, else `ip:<ClientIP>` (CIDR-aware, see `wiki/architecture/trusted-proxy.md`) | Fixed window (`WindowSeconds` + `MaxRequests`) | `METALDOCS_RATELIMIT_*` env (`config.LoadRateLimitConfig`) |
| `ratelimit.Middleware` (`Limit`) | **Per-route**, only on a small set of expensive write paths | `userExtractor(*http.Request) string` → `user:<id>` if set, else `ip:<ClientIP>` via trusted-proxy resolution (same CIDR config as global limiter — H2 fix) | Token bucket (`golang.org/x/time/rate`) — quota req/min, burst = quota | `ratelimit.Config.Quotas` keyed by `RouteKey` (`DefaultConfig`) |

The global limiter is the cheap blanket guard against abusive clients. The per-route limiter is the targeted guard on operations that are individually expensive (storage uploads, PDF rendering) where a small number of legitimate-looking calls can still be damaging.

## 2. Where each one is wired

### 2.1 `security.RateLimiter` — global envelope

Constructed once at startup and wrapped around the entire mux:

```
apps/api/cmd/metaldocs-api/main.go:209
    rateLimiter := security.NewRateLimiter(rateCfg)

apps/api/cmd/metaldocs-api/main.go:471
    handler := cors.Wrap(
        originProtection.Wrap(
            authMiddleware.Wrap(
                iamMiddleware.Wrap(
                    httpObs.Wrap(
                        rateLimiter.Wrap(mux))))))
```

- Health probes (`/api/v1/health/live`, `/api/v1/health/ready`) are skipped via `shouldSkipRateLimit`.
- Identity preference order: authenticated user → IAM-context user → trusted-proxy-resolved client IP → `ip:unknown`.
- On rejection: `HTTP 429` with `Retry-After` (seconds until window rolls) and the standard `RATE_LIMITED` envelope body.

### 2.2 `ratelimit.Middleware` — per-route token bucket

**Status:** wired in code and exercised by tests, **not yet active in production**. Production startup currently registers documents routes via `docMod.RegisterRoutes(mux)` ([apps/api/cmd/metaldocs-api/main.go:380](../../apps/api/cmd/metaldocs-api/main.go)) which intentionally passes a nil limiter through `buildLegacyMux` ([internal/modules/documents/module.go:146-156](../../internal/modules/documents/module.go)). Activating the per-route limiter is a two-line wiring change:

```
apps/api/cmd/metaldocs-api/main.go (proposed)
    rl := ratelimit.New(ctx, ratelimit.DefaultConfig())
    docMod.RegisterRoutesWithRateLimit(mux, rl, userFn)
```

`buildLegacyMux` then fans out to per-handler `rl.Limit(routeKey, userFn, h)` for the routes below. Every other documents route is registered as a plain `mux.HandleFunc` and is only covered by the global envelope:

| Route | Method | `RouteKey` | Default quota (req/min) | Callsite |
|---|---|---|---|---|
| `/api/v1/documents/{id}/autosave/presign` | `POST` | `RouteAutosavePresign` | 60 | [`handler.go:145-148`](../../internal/modules/documents/delivery/http/handler.go) |
| `/api/v1/documents/{id}/autosave/commit`  | `POST` | `RouteAutosaveCommit`  | 30 | [`handler.go:149-152`](../../internal/modules/documents/delivery/http/handler.go) |
| `/api/v1/documents/{id}/export/pdf`       | `POST` | `RouteExportPDF`       | 20 | [`export_handler.go:44-48`](../../internal/modules/documents/delivery/http/export_handler.go) |

`RouteUploadsPresign` and `RouteDocumentsRender` exist in the catalog (`config.go:12-16`) but have no `rl.Limit` callsite yet — they are reserved for future endpoints. Adding a new per-route limit means (a) declaring a `RouteKey` constant, (b) giving it a `Quotas` entry in `DefaultConfig`, and (c) wrapping the handler with `rl.Limit(...)`.

Activating the per-route limiter in production is intentionally out of scope for the C2 patch: the C2 fix bounds the memory of an existing-but-dormant code path so it is safe to activate, but actually activating it changes behavior (clients can now see 429s on autosave / export) and belongs to a follow-up change.

## 3. Bounded-memory contract (finding C2)

Both limiters historically kept an unbounded per-identity map and would never evict — every distinct IP or user seen since process start consumed a permanent slot. The fix:

### 3.1 `security.RateLimiter`

- Every call to `allow()` invokes `sweepExpiredLocked(now)` inside the existing mutex before deciding. Entries with `windowStart + window < now` are dropped; a single `slog.Warn` per non-empty sweep batch carries `evicted` + `remaining`.
- Hard cap: `defaultMaxRateLimitEntries = 100_000` (overridable in tests via `WithMaxEntries`). On overflow new identities are **fail-closed** denied with `HTTP 429` + `Retry-After`, and a `slog.Warn` records `map_size` + `max_entries`. Existing identities inside the cap continue to be evaluated.
- Test helpers: `WithClock(func() time.Time)`, `WithLogger(*slog.Logger)`, `Size() int`.

### 3.2 `ratelimit.Middleware`

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

### 3.3 Regression tests

| Test | Asserts |
|---|---|
| [`TestRateLimiterSweepsExpiredWindowEntries`](../../tests/unit/rate_limit_eviction_test.go) | Global limiter: 2_000 distinct identities collapse to 1 after window elapse + one subsequent admission. |
| [`TestRateLimiterFailsClosedOnMapOverflow`](../../tests/unit/rate_limit_eviction_test.go) | Global limiter: with `WithMaxEntries(8)`, the 9th distinct identity returns 429 with `Retry-After`; existing identities still admitted. |
| [`TestSweep_EvictsIdleEntries`](../../internal/platform/ratelimit/eviction_test.go) | Per-route limiter: 10_000 unique identities collapse to 0 after deterministic-clock sweep. |
| [`TestSweeper_ExitsOnContextCancel`](../../internal/platform/ratelimit/eviction_test.go) | Per-route limiter: sweeper goroutine returns within 1s of `cancel()`, `runtime.NumGoroutine()` returns to baseline. |
| [`TestLimit_LimiterReusedAcrossRequests`](../../internal/platform/ratelimit/eviction_test.go) | Per-route limiter: quota=1, second request on the same key is 429 (proves limiter persisted, not freshly allocated). |
| [`TestLimit_RaceUnderConcurrentFirstHit`](../../internal/platform/ratelimit/eviction_test.go) | Per-route limiter: 64 goroutines racing 4 distinct keys yield exactly 4 surviving entries. |
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

## 4. Known duplication

Both limiters solve overlapping problems with different algorithms and different config shapes. The C2 fix kept them separate on purpose — merging them is a larger surgery (algorithm choice, config migration, route catalog rationalization, removal of one of the two wiring points) and is intentionally **out of scope** for the C2 patch.

Pick-one direction for the future refactor (recorded here so the next change has a starting position, not a binding decision):

- `ratelimit.Middleware` is the better long-term substrate: token-bucket gives smoother behavior under burst, per-route quotas are first-class, sweeper lifecycle is observable, and the route catalog already exists.
- The global envelope's value is its identity-resolution path (IAM user → trusted-proxy IP), which can be lifted out as a `userExtractor` and reused.
- Net direction: collapse the global envelope into a default-route `ratelimit.Middleware` chain entry that uses the same identity extractor, then retire `security.RateLimiter`.

Track this under a follow-up refactor task. Do not silently merge the two during unrelated work.
