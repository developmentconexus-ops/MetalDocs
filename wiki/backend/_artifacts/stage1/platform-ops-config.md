# Stage-1 Audit Artifact — Platform Ops: Config, Observability, Feature Flags

> **Area:** `platform-ops-config`
> **Code paths:** `internal/platform/config`, `internal/platform/observability`, `internal/platform/featureflags`
> **Produced:** 2026-06-10
> **Author:** Stage-1 mapping agent (read-only)

---

## 1. Identity & purpose

These three packages form the **startup spine and runtime observability surface** of the MetalDocs backend.

`internal/platform/config` is a pure 12-factor config layer: eight standalone `Load*` functions parse environment variables, validate inputs, fail-fast on missing secrets, and return typed structs. Every other platform package and all three binaries derive their runtime values from this layer exclusively.

`internal/platform/observability` provides the HTTP middleware that wraps the entire request mux, emitting structured JSON access logs and accumulating in-process RED metrics per route. It also owns the health/readiness probe endpoints and the `RuntimeStatusProvider` abstraction, which describes the live state of infrastructure dependencies (Postgres, Gotenberg) to the readiness probe and the `/api/v1/metrics` endpoint.

`internal/platform/featureflags` exposes a single HTTP handler — `GET /api/v1/feature-flags` — that returns server-controlled rollout percentages to the browser client. It holds no state of its own; it reads a `FeatureFlagsConfig` struct loaded at startup by the config layer.

---

## 2. File inventory

### `internal/platform/config` (9 files)

| File | Role |
|---|---|
| `attachments.go` | `LoadAttachmentsConfig()`: storage provider selection (memory/local/MinIO), HMAC signing secret (≥32 bytes required), download TTL, MinIO connection credentials, auto-bucket flag. Also defines `StorageProvider`, `AppEnv` types and `parseBoolEnv` helper. |
| `cors.go` | `LoadCORSConfig()`: CORS enabled/origins/methods/headers/credentials/max-age. Validates wildcard origin with credentials is rejected. Exports `splitCSV` and `normalizeUpper` helpers. |
| `feature_flags.go` | `LoadFeatureFlagsConfig()`: single rollout percentage key `METALDOCS_MDDM_NATIVE_EXPORT_ROLLOUT_PCT` (0–100). Exports `ErrInvalidPercentage` sentinel. |
| `feature_flags_test.go` | Unit test: out-of-range percentages (-1, 101) return `ErrInvalidPercentage`. |
| `gotenberg.go` | `LoadGotenbergConfig()`: optional absolute http(s) URL for Gotenberg. Empty env → disabled. URL validation blocks ftp/other schemes. |
| `gotenberg_test.go` | Unit tests: disabled when empty, enabled with valid URL, rejects ftp scheme. |
| `jobs.go` | `LoadJobsConfig()`: River jobs enabled flag, River schema, temporal queue max-worker count (default 10). Imports `github.com/riverqueue/river` for `river.QueueConfig`. |
| `postgres.go` | `LoadPostgresConfig()`: accepts `DATABASE_URL` (full DSN) or `PGHOST/PGPORT/PGDATABASE/PGUSER/PGPASSWORD/PGSSLMODE` (default sslmode=require). Validates scheme. |
| `ratelimit.go` | `LoadRateLimitConfig()`: window seconds (default 60), max requests (default 120), delegates to `LoadTrustedProxyCIDRs` for IP source. |
| `repository.go` | `RepositoryMode()`: `METALDOCS_REPOSITORY` → `"memory"` or `"postgres"` (default `"memory"`). |
| `trusted_proxy.go` | `LoadTrustedProxyCIDRs()` / `ParseTrustedProxyCIDRs()`: comma-separated CIDR list → `[]netip.Prefix`. Empty → nil (fail-closed: no upstream trusted). |
| `worker.go` | `LoadWorkerConfig()`: poll interval (default 10s), batch size (default 25), review reminder days (default 14), run-once flag, max attempts (default 5), retry base/max seconds (default 10/300). Cross-field validation: `RetryMaxSeconds >= RetryBaseSeconds`. |

### `internal/platform/observability` (5 files, including `.gitkeep`)

| File | Role |
|---|---|
| `.gitkeep` | Empty placeholder (predates the package files). |
| `http.go` | `HTTPObservability`: HTTP middleware (`Wrap`) + metrics handler. Emits `slog` JSON access log per request. Maintains per-route `routeMetrics` ring-buffer (200 samples) for p50/p95/p99 percentiles. `normalizeRoute` extracts parameterized route labels for 5 hardcoded path patterns. |
| `health.go` | `HealthHandler`: registers `/api/v1/health/live`, `/api/v1/health/ready`, `/healthz`. Delegates to `RuntimeStatusProvider`; if provider is nil, returns a static "live/ready" response. |
| `runtime.go` | `RuntimeStatusProvider` interface + two concrete implementations: `StaticRuntimeStatusProvider` (memory-mode; hardcoded up status) and `PostgresRuntimeStatusProvider` (postgres-mode; pings DB; queries `metaldocs.auth_identities`, `metaldocs.auth_sessions`, `metaldocs.outbox_events` for runtime metrics). Pluggable `DependencyCheck` slice runs under a 2-second timeout each. |
| `runtime_test.go` | Unit tests: static provider degrades on dependency failure; postgres provider shares a readiness deadline across checks; runtime metrics omit failed query sections gracefully. |

### `internal/platform/featureflags` (2 files)

| File | Role |
|---|---|
| `handler.go` | `Handler`: `GET /api/v1/feature-flags` returns JSON `{"MDDM_NATIVE_EXPORT_ROLLOUT_PCT": <int>}`. `Cache-Control: no-store`. 405 on non-GET. |
| `handler_test.go` | Unit test: 200 + correct JSON key present. |

---

## 3. Public surface

### `internal/platform/config` — exported types and functions consumed elsewhere

| Symbol | Consumed by |
|---|---|
| `RepositoryMode() (string, error)` | `apps/api/cmd/metaldocs-api/main.go:152` |
| `RepositoryMemory`, `RepositoryPostgres` constants | `internal/platform/bootstrap/api.go`, `apps/api/cmd/metaldocs-api/main.go`, tests |
| `LoadCORSConfig() (CORSConfig, error)` | `apps/api/cmd/metaldocs-api/main.go:163`, `internal/platform/security/cors.go` |
| `LoadRateLimitConfig() (RateLimitConfig, error)` | `apps/api/cmd/metaldocs-api/main.go:159`, `internal/platform/security/ratelimit.go` |
| `LoadAttachmentsConfig() (AttachmentsConfig, error)` | `apps/api/cmd/metaldocs-api/main.go:167`, `internal/platform/bootstrap/api.go`, `internal/platform/storage/minio/store.go` |
| `StorageProvider`, `StorageProviderMemory/Local/MinIO` | `internal/platform/bootstrap/api.go:89` |
| `AppEnv`, `AppEnvLocal/Dev/Staging/Production` | `internal/platform/storage/minio/store.go` (conditional S3 URL logic) |
| `LoadGotenbergConfig() (GotenbergConfig, error)` | `internal/platform/bootstrap/api.go:62` |
| `LoadPostgresConfig() (PostgresConfig, error)` | `internal/platform/bootstrap/api.go:79` |
| `LoadJobsConfig() (JobsConfig, error)` | `apps/api/cmd/metaldocs-api/main.go:434`, `apps/jobs/cmd/metaldocs-jobs/main.go:26` |
| `LoadWorkerConfig() (WorkerConfig, error)` | `apps/worker/cmd/metaldocs-worker/main.go:55`, `internal/platform/bootstrap/worker.go` |
| `LoadTrustedProxyCIDRs() / ParseTrustedProxyCIDRs()` | `internal/platform/config/ratelimit.go:42` (indirect), `tests/unit/trusted_proxy_test.go` |
| `LoadFeatureFlagsConfig() (FeatureFlagsConfig, error)` | `apps/api/cmd/metaldocs-api/main.go:175` |
| `ErrInvalidPercentage` | `internal/platform/config/feature_flags_test.go` |

### `internal/platform/observability` — exported types and functions

| Symbol | Consumed by |
|---|---|
| `NewHTTPObservability(...RuntimeStatusProvider) *HTTPObservability` | `apps/api/cmd/metaldocs-api/main.go:275` |
| `HTTPObservability.Wrap(next http.Handler) http.Handler` | `apps/api/cmd/metaldocs-api/main.go:598` (wraps `rateLimiter.Wrap(mux)`) |
| `HTTPObservability.MetricsHandler() http.Handler` | `apps/api/cmd/metaldocs-api/main.go:572` (registered at `/api/v1/metrics`) |
| `NewHealthHandler(RuntimeStatusProvider) *HealthHandler` | `internal/platform/bootstrap/api.go:118` (implicit via `StatusProvider`) |
| `HealthHandler.RegisterRoutes(mux)` | `apps/api/cmd/metaldocs-api/main.go` (via bootstrap) |
| `RuntimeStatusProvider` interface | `internal/platform/bootstrap/api.go:45` (`APIDependencies.StatusProvider`) |
| `NewStaticRuntimeStatusProvider(...)` | `internal/platform/bootstrap/api.go:152` (memory mode) |
| `NewPostgresRuntimeStatusProvider(db, ...)` | `internal/platform/bootstrap/api.go:118` (postgres mode) |
| `DependencyCheck` struct | `internal/platform/bootstrap/api.go:187` (`gotenbergHealthCheck` factory) |
| `DependencyCheckResult` struct | `internal/platform/bootstrap/api.go:188` (returned by check closures) |

### `internal/platform/featureflags` — exported types and functions

| Symbol | Consumed by |
|---|---|
| `NewHandler(cfg config.FeatureFlagsConfig) *Handler` | `apps/api/cmd/metaldocs-api/main.go:274` |
| `Handler.RegisterRoutes(mux *http.ServeMux)` | `apps/api/cmd/metaldocs-api/main.go` (implicit via handler usage) |

### HTTP routes registered

| Method | Path | Auth | Notes |
|---|---|---|---|
| `GET` | `/api/v1/feature-flags` | None (public) | Returns `FeatureFlagsConfig` as JSON; `Cache-Control: no-store`. Registered via `Handler.RegisterRoutes`. |
| `GET` | `/api/v1/health/live` | None (public) | Liveness probe — always 200 when process is up. Alias: `/healthz`. |
| `GET` | `/api/v1/health/ready` | None (public) | Readiness probe — 200 or 503; checks DB ping + registered `DependencyCheck` items. |
| `GET` | `/api/v1/metrics` | `CapMetricsView` (tier-1) | In-process RED metrics + runtime stats. Route registered in `main.go:572`. |

---

## 4. Logic flows

### Flow 1: API startup config loading sequence

All `Load*` functions are called sequentially in `apps/api/cmd/metaldocs-api/main.go` before any module wiring. Failure at any step calls `log.Fatalf` — startup aborts.

1. `config.RepositoryMode()` reads `METALDOCS_REPOSITORY` → validates to `"memory"` or `"postgres"`. (`config/repository.go:14`)
2. `requirePostgresRepositoryMode(repoMode)` asserts postgres-only for the API binary. (`main.go:677`) If mode is `"memory"`, startup aborts with an explicit message.
3. `config.LoadRateLimitConfig()` reads `METALDOCS_RATE_LIMIT_*` and delegates to `LoadTrustedProxyCIDRs()` for the IP-trust CIDR list. (`config/ratelimit.go:21`, `config/trusted_proxy.go:13`)
4. `config.LoadCORSConfig()` reads `METALDOCS_CORS_*`, validates wildcard+credentials conflict. (`config/cors.go:19`)
5. `config.LoadAttachmentsConfig()` reads `APP_ENV`, `METALDOCS_STORAGE_PROVIDER`, `METALDOCS_ATTACHMENTS_*`, `METALDOCS_MINIO_*`. Requires `METALDOCS_ATTACHMENTS_SIGNING_SECRET` ≥ 32 bytes unconditionally. (`config/attachments.go:42`)
6. `authn.LoadRuntimeConfig()` reads `METALDOCS_AUTH_*`, `APP_ENV`, `METALDOCS_BOOTSTRAP_ADMIN_*`. (`platform/authn/config.go:37`)
7. `config.LoadFeatureFlagsConfig()` reads `METALDOCS_MDDM_NATIVE_EXPORT_ROLLOUT_PCT`. (`config/feature_flags.go:22`, `main.go:175`)
8. `bootstrap.BuildAPIDependencies(ctx, repoMode, attachmentsCfg)` calls `config.LoadGotenbergConfig()` and `config.LoadPostgresConfig()` internally. (`bootstrap/api.go:61-82`)
9. `config.LoadJobsConfig()` reads `METALDOCS_JOBS_*`. (`config/jobs.go:19`, `main.go:434`)

All error paths: fatal — no partial startup.

### Flow 2: HTTP request observability path

Every request entering the API mux passes through `HTTPObservability.Wrap` before it reaches any handler.

1. Middleware reads `X-Trace-Id` header; validates via `requesttrace.Normalize` (printable ASCII, ≤128 chars). (`observability/http.go:61-63`)
2. If header is absent or invalid, generates a new UUID via `requesttrace.Resolve`. (`requesttrace/context.go:34`)
3. Stores trace ID in request context with `requesttrace.WithTraceID`. (`observability/http.go:65`, `requesttrace/context.go:12`)
4. Delegates to `next.ServeHTTP(sw, r)` where `sw` is a `statusWriter` that captures the response status code. (`observability/http.go:68-69`)
5. After the inner handler returns: computes `elapsedMs`, classifies `isError = status >= 400`. (`observability/http.go:74-79`)
6. Calls `o.getMetric(route, method)` — double-checked locking; creates `routeMetrics` on first encounter (ring buffer, cap 200). (`observability/http.go:156-176`)
7. Atomically increments `requests`, optionally `errors`, `durationMs`; appends to ring buffer for percentile tracking. (`observability/http.go:82-87`)
8. Extracts `userID` from auth context (falls back to `"anonymous"`); extracts `documentID`/`profileCode` from path. (`observability/http.go:89-93`)
9. Emits `slog.Info("http_request", ...)` with 9 structured fields to stdout. (`observability/http.go:95-106`)

Chain order in production: `cors.Wrap(originProtection.Wrap(authMiddleware.Wrap(iamMiddleware.Wrap(presenceBump.Wrap(httpObs.Wrap(rateLimiter.Wrap(mux)))))))` — `httpObs.Wrap` sits **inside** authN, so unauthenticated rejects do not appear in RED metrics. (`main.go:598-602`)

### Flow 3: Readiness probe with Gotenberg dependency check

1. `GET /api/v1/health/ready` → `HealthHandler.handleReady(w, r)`. (`observability/health.go:31`)
2. Calls `provider.Ready(r.Context())`. For postgres mode this is `PostgresRuntimeStatusProvider.Ready`. (`observability/runtime.go:113`)
3. Opens a 3-second context deadline for the entire Ready call. (`observability/runtime.go:117`)
4. Pings `p.db` with `PingContext(readyCtx)`. On failure: status=degraded, code=503, check[0].status=down. (`observability/runtime.go:133-137`)
5. Iterates `p.checks` via `applyDependencyChecks`; each check runs under its own 2-second sub-deadline. (`observability/runtime.go:286-292`)
6. For Gotenberg: the check closure makes a real `GET {url}/health` request with a 2-second `http.Client.Timeout`. (`bootstrap/api.go:197-209`) Non-200 → error.
7. If any check is not `"up"` or `"skipped"`, the outer status flips to degraded (503). (`observability/runtime.go:315-320`)
8. Response JSON: `{"status":"ready"|"degraded","checks":[...]}`. (`observability/runtime.go:146-149`)

### Flow 4: Feature flag request lifecycle

1. At startup, `config.LoadFeatureFlagsConfig()` reads `METALDOCS_MDDM_NATIVE_EXPORT_ROLLOUT_PCT`. Default 0. Validates 0–100. (`config/feature_flags.go:22-36`)
2. `featureflags.NewHandler(featureFlagsCfg)` captures the config as an immutable struct. (`featureflags/handler.go:18`)
3. `h.RegisterRoutes(mux)` registers `handleFunc` on the standard mux at `/api/v1/feature-flags`. (`featureflags/handler.go:23`)
4. On `GET /api/v1/feature-flags`: sets `Content-Type: application/json`, `Cache-Control: no-store`, encodes `{"MDDM_NATIVE_EXPORT_ROLLOUT_PCT": <int>}`. (`featureflags/handler.go:31-40`)
5. Config is static for the lifetime of the process; there is no hot-reload mechanism.

### Flow 5: Worker config and runtime metrics flow

1. `apps/worker/cmd/metaldocs-worker/main.go` calls `config.LoadWorkerConfig()` at startup. (`worker/main.go:55`)
2. Passes `WorkerConfig` to `bootstrap.BuildWorkerDependencies` then to `workerapp.NewService`. (`worker/main.go:59,65`)
3. Worker loop tick-driven by `time.NewTicker(PollIntervalSeconds)`. (`worker/main.go:106`)
4. On each outbox event processed, the worker logs `trace_id` using the `TraceID` field from the outbox row — the same `trace_id` inserted by the API handler at publish time (`messaging/outbox/postgres/publisher.go:54`). This provides manual trace correlation across api→worker hops but is not W3C `traceparent` propagation.
5. `/api/v1/metrics` queries `metaldocs.auth_identities`, `metaldocs.auth_sessions`, and `metaldocs.outbox_events` inline on each request under 3-second per-query timeouts. (`observability/runtime.go:183-212`)

---

## 5. Dependencies

### Outbound imports

**`internal/platform/config`**

| Import | Reason |
|---|---|
| `os`, `strconv`, `strings`, `fmt`, `net/url`, `net/netip` | Standard library; env var parsing and CIDR/URL validation |
| `github.com/riverqueue/river` (`river.QueueConfig`) | `config/jobs.go:9` — `JobsConfig.Queues` typed map uses River's config struct |

**`internal/platform/observability`**

| Import | Reason |
|---|---|
| `log/slog`, `os` | Structured JSON logging to stdout |
| `sync`, `sync/atomic` | Concurrent metric counter updates and ring-buffer locking |
| `database/sql` | `PostgresRuntimeStatusProvider` holds `*sql.DB` for readiness ping and metric queries |
| `metaldocs/internal/modules/auth/domain` | `authdomain.CurrentUserFromContext` to extract `user_id` for access logs (`http.go:90`) |
| `metaldocs/internal/platform/requesttrace` | Trace ID propagation in context |

Note: `observability` imports `internal/modules/auth/domain` — a platform package importing a module domain. This is a layering exception; see section 10.

**`internal/platform/featureflags`**

| Import | Reason |
|---|---|
| `encoding/json`, `net/http` | JSON response encoding |
| `metaldocs/internal/platform/config` | Reads `FeatureFlagsConfig` struct |

### Inbound — who imports these packages (verified with grep)

**`internal/platform/config`** is imported by 23 files:
- `apps/api/cmd/metaldocs-api/main.go` — primary consumer of all 8 Load functions
- `apps/worker/cmd/metaldocs-worker/main.go` — `LoadWorkerConfig`
- `apps/jobs/cmd/metaldocs-jobs/main.go` — `LoadJobsConfig`
- `internal/platform/bootstrap/api.go`, `bootstrap/worker.go`, `bootstrap/jobs.go` — dependency assembly
- `internal/platform/security/ratelimit.go`, `security/cors.go` — config struct consumers
- `internal/platform/authn/config.go` — `RepositoryMemory` and `AttachmentsConfig` (partially)
- `internal/platform/storage/minio/store.go` — `AttachmentsConfig`, `AppEnv`
- `internal/platform/featureflags/handler.go` and `handler_test.go`
- `apps/api/cmd/metaldocs-e2e-seed/main.go`
- Various unit test files under `tests/unit/`

**`internal/platform/observability`** is imported by 2 files:
- `apps/api/cmd/metaldocs-api/main.go` — `NewHTTPObservability`, `MetricsHandler`
- `internal/platform/bootstrap/api.go` — `NewStaticRuntimeStatusProvider`, `NewPostgresRuntimeStatusProvider`, `DependencyCheck`

**`internal/platform/featureflags`** is imported by 2 files:
- `apps/api/cmd/metaldocs-api/main.go` — `NewHandler`
- `internal/platform/featureflags/handler_test.go`

---

## 6. Persistence

`internal/platform/config` — **stateless**. Pure env-var parsing; no DB, no file I/O beyond `os.Getenv`.

`internal/platform/featureflags` — **stateless**. Config loaded at startup, no writes.

`internal/platform/observability` — **in-process state only** (the `byKey` metric map + ring buffers). Not persisted; resets on restart.

`PostgresRuntimeStatusProvider.RuntimeMetrics` queries three tables on every `/api/v1/metrics` GET:
- `metaldocs.auth_identities` — COUNT by active/inactive/locked state
- `metaldocs.auth_sessions` — COUNT by active/expired/revoked state
- `metaldocs.outbox_events` — COUNT by claimable/pending/dead-lettered state

`PostgresRuntimeStatusProvider.Ready` runs `db.PingContext` on `GET /api/v1/health/ready`.

No migrations are owned by this area.

---

## 7. Config & environment

Full catalog of every environment variable in scope:

### `internal/platform/config`

| Variable | Default | Required | Validation | Consumer |
|---|---|---|---|---|
| `METALDOCS_REPOSITORY` | `"memory"` | No | `"memory"` or `"postgres"` | `RepositoryMode()` |
| `METALDOCS_CORS_ENABLED` | `false` | No | `"true"` / other | `LoadCORSConfig()` |
| `METALDOCS_CORS_ALLOWED_ORIGINS` | (empty) | No | CSV | `LoadCORSConfig()` |
| `METALDOCS_CORS_ALLOWED_METHODS` | `GET,POST,PUT,OPTIONS` | No | CSV, uppercased | `LoadCORSConfig()` |
| `METALDOCS_CORS_ALLOWED_HEADERS` | `Content-Type,X-Trace-Id` | No | CSV | `LoadCORSConfig()` |
| `METALDOCS_CORS_EXPOSED_HEADERS` | (empty) | No | CSV | `LoadCORSConfig()` |
| `METALDOCS_CORS_ALLOW_CREDENTIALS` | `false` | No | `"true"` / other; rejects `*` origin when true | `LoadCORSConfig()` |
| `METALDOCS_CORS_MAX_AGE_SECONDS` | `300` | No | integer ≥ 0 | `LoadCORSConfig()` |
| `METALDOCS_RATE_LIMIT_ENABLED` | `false` | No | `"true"` / other | `LoadRateLimitConfig()` |
| `METALDOCS_RATE_LIMIT_WINDOW_SECONDS` | `60` | No | integer > 0 | `LoadRateLimitConfig()` |
| `METALDOCS_RATE_LIMIT_MAX_REQUESTS` | `120` | No | integer > 0 | `LoadRateLimitConfig()` |
| `METALDOCS_TRUSTED_PROXY_CIDRS` | (empty → fail-closed) | No | comma-separated CIDR list; invalid entry = error | `LoadTrustedProxyCIDRs()` |
| `APP_ENV` | `"local"` | No | any string; guards auth-disabled and MinIO URL logic | `LoadAttachmentsConfig()`, `authn.LoadRuntimeConfig()` |
| `METALDOCS_STORAGE_PROVIDER` | `"local"` | No | `"memory"`, `"local"`, `"minio"` | `LoadAttachmentsConfig()` |
| `METALDOCS_ATTACHMENTS_ROOT` | `"non_git/attachments"` | No | string | `LoadAttachmentsConfig()` |
| `METALDOCS_ATTACHMENTS_SIGNING_SECRET` | — | **Yes** (all providers) | ≥ 32 bytes | `LoadAttachmentsConfig()` |
| `METALDOCS_ATTACHMENTS_DOWNLOAD_TTL_SECONDS` | `300` | No | integer ≥ 30 | `LoadAttachmentsConfig()` |
| `METALDOCS_MINIO_ENDPOINT` | — | Yes (MinIO only) | non-empty | `LoadAttachmentsConfig()` |
| `METALDOCS_MINIO_PUBLIC_ENDPOINT` | = `METALDOCS_MINIO_ENDPOINT` | No | non-empty | `LoadAttachmentsConfig()` |
| `METALDOCS_MINIO_ACCESS_KEY` | — | Yes (MinIO only) | non-empty | `LoadAttachmentsConfig()` |
| `METALDOCS_MINIO_SECRET_KEY` | — | Yes (MinIO only) | non-empty | `LoadAttachmentsConfig()` |
| `METALDOCS_MINIO_BUCKET` | — | Yes (MinIO only) | non-empty | `LoadAttachmentsConfig()` |
| `METALDOCS_MINIO_USE_SSL` | `false` | No | `"true"` / `"1"` / other | `LoadAttachmentsConfig()` |
| `METALDOCS_MINIO_AUTO_CREATE_BUCKET` | `false` | No | `"true"` / `"1"` / other | `LoadAttachmentsConfig()` |
| `METALDOCS_GOTENBERG_URL` | (empty → disabled) | No | absolute `http(s)` URL | `LoadGotenbergConfig()` |
| `DATABASE_URL` | — | Yes (postgres mode; or PG* set) | postgres/postgresql scheme | `LoadPostgresConfig()` |
| `PGHOST` | — | Yes (if `DATABASE_URL` absent) | non-empty | `LoadPostgresConfig()` |
| `PGPORT` | `5432` | No | any | `LoadPostgresConfig()` |
| `PGDATABASE` | — | Yes (if `DATABASE_URL` absent) | non-empty | `LoadPostgresConfig()` |
| `PGUSER` | — | Yes (if `DATABASE_URL` absent) | non-empty | `LoadPostgresConfig()` |
| `PGPASSWORD` | — | Yes (if `DATABASE_URL` absent) | non-empty | `LoadPostgresConfig()` |
| `PGSSLMODE` | `"require"` | No | any string | `LoadPostgresConfig()` |
| `METALDOCS_JOBS_ENABLED` | `true` | No | `"true"` / `"1"` / other | `LoadJobsConfig()` |
| `METALDOCS_JOBS_RIVER_SCHEMA` | `""` | No | string | `LoadJobsConfig()` |
| `METALDOCS_JOBS_TEMPORAL_MAX_WORKERS` | `10` | No | integer ≥ 1 | `LoadJobsConfig()` |
| `METALDOCS_WORKER_POLL_INTERVAL_SECONDS` | `10` | No | integer ≥ 1 | `LoadWorkerConfig()` |
| `METALDOCS_WORKER_BATCH_SIZE` | `25` | No | integer ≥ 1 | `LoadWorkerConfig()` |
| `METALDOCS_WORKER_REVIEW_REMINDER_DAYS` | `14` | No | integer ≥ 1 | `LoadWorkerConfig()` |
| `METALDOCS_WORKER_RUN_ONCE` | `false` | No | `"true"` / `"1"` / other | `LoadWorkerConfig()` |
| `METALDOCS_WORKER_MAX_ATTEMPTS` | `5` | No | integer ≥ 1 | `LoadWorkerConfig()` |
| `METALDOCS_WORKER_RETRY_BASE_SECONDS` | `10` | No | integer ≥ 1 | `LoadWorkerConfig()` |
| `METALDOCS_WORKER_RETRY_MAX_SECONDS` | `300` | No | integer ≥ `RetryBaseSeconds` | `LoadWorkerConfig()` |
| `METALDOCS_MDDM_NATIVE_EXPORT_ROLLOUT_PCT` | `0` | No | integer 0–100 | `LoadFeatureFlagsConfig()` |

Note: `METALDOCS_AUTH_*`, `METALDOCS_AUTHZ_CACHE_TTL_SECONDS`, and `METALDOCS_BOOTSTRAP_ADMIN_*` are parsed by `internal/platform/authn/config.go`, not by `internal/platform/config`. They are adjacent but out of scope for this area.

---

## 8. Concurrency & async

**`internal/platform/config`** — no goroutines; pure synchronous env parsing.

**`internal/platform/featureflags`** — no goroutines; config is read-only after `NewHandler`.

**`internal/platform/observability`**:

- `HTTPObservability.byKey` map is guarded by `sync.RWMutex`. Hot path takes `RLock`; new-route creation uses double-checked locking under `Lock`. (`observability/http.go:156-176`)
- Per-route counters (`requests`, `errors`, `durationMs`) are updated via `sync/atomic`. (`observability/http.go:82-86`)
- The 200-sample ring buffer inside each `routeMetrics` is protected by `sync.Mutex` (separate from the map lock, avoiding contention between routes). (`observability/http.go:266-275`)
- No goroutines are spawned; all work is done inline in the request goroutine.
- `PostgresRuntimeStatusProvider.Ready` and `RuntimeMetrics` each call `queryRuntimeMetric` with `context.WithTimeout` (3 seconds for the outer call, 3 seconds for each sub-query). No goroutines; sequential DB calls. (`observability/runtime.go:183-210`)
- `applyDependencyChecks` creates a new `context.WithTimeout(ctx, 2*time.Second)` per dependency check, but checks run sequentially, not in parallel. (`observability/runtime.go:291`) This means total blocking time for N checks is up to 2N seconds, bounded by the outer 3-second readiness context.

---

## 9. Error handling & observability

**`internal/platform/config`**: All errors are returned from `Load*` functions as descriptive `fmt.Errorf` strings. No logging, no sentinel wrapping beyond `ErrInvalidPercentage`. Callers are expected to `log.Fatalf` on error (and do: `main.go:153-177`).

**`internal/platform/featureflags`**: `json.NewEncoder(w).Encode(...)` error is silently discarded with `_ = ...`. (`featureflags/handler.go:38`) This is acceptable for a simple struct with no error path, but leaves write failures invisible.

**`internal/platform/observability`**:
- Access logs: `slog.Info("http_request", ...)` to stdout. Uses `log/slog` with `NewJSONHandler(os.Stdout, ...)`. (`observability/http.go:53`) No OpenTelemetry exporter is configured anywhere in this package or the wider `internal/platform` area — confirmed by absence of `go.opentelemetry.io/*` in `go.mod` and no OTel imports in any platform file.
- Metrics: in-process only. There is no Prometheus exposition endpoint and no OTLP exporter. The `GET /api/v1/metrics` endpoint serves the in-process counters as JSON, gated by `CapMetricsView`.
- Trace context: `requesttrace` uses a custom `X-Trace-Id` header (not W3C `traceparent`). The trace ID propagates through the HTTP access log and through the outbox `trace_id` column (via `messaging.Event.TraceID`), but there is no structured span creation, no sampling, and no export to any tracing backend.
- `writeJSON` in `health.go:40` discards `json.Encode` errors. (`observability/health.go:43`)
- `truncateReadinessError` caps error strings at 160 characters to prevent response bloat from verbose DB errors. (`observability/runtime.go:264-270`)
- RFC 9457 is not used by this area; health/readiness/metrics responses use ad hoc JSON shapes, not `problem+json`.

---

## 10. Legacy / duplication / smell flags

- **`parseBoolEnv` duplicated across packages.** `internal/platform/config/attachments.go:108` and `internal/platform/authn/config.go:213` define identical `parseBoolEnv(name string, defaultValue bool) bool` functions. Similarly, `splitCSV` is defined in both `internal/platform/config/cors.go:63` and `internal/platform/authn/config.go:222`. These are package-private helpers that should live in one shared location or be inlined, not copied. Git history shows `authn/config.go` predates the `config` package restructure. Confirmed by grep (`func parseBoolEnv` appears in both files). *Smell: duplication / inconsistent internal factoring.*

- **`normalizeRoute` hardcodes a partial route set.** `internal/platform/observability/http.go:178-208` contains 7 hardcoded prefix/suffix string checks covering only `/api/v1/documents/`, `/api/v1/document-profiles/`, `/api/v1/workflow/documents/`, and three `/api/v1/iam/users/` sub-paths. All other parameterized routes (templates, taxonomy, controlled-documents, search, approval, etc.) log the full path including raw IDs, which inflates the cardinality of the `byKey` metric map and leaks IDs into structured logs. As routes are added, this function will silently miss new parameterized paths. *Smell: incomplete normalization / log cardinality drift / maintenance burden.*

- **`observability` imports a domain module (`auth/domain`).** `internal/platform/observability/http.go:15` imports `metaldocs/internal/modules/auth/domain` to read `CurrentUserFromContext`. This violates the platform-is-domain-free rule stated in `backend-target-architecture.md` REQ-TOP-2. The dependency inversion is achievable via a context-key interface or a trivial accessor func at the platform boundary. *Smell: layering violation — platform package imports module domain.*

- **Feature flag config lives in two separate packages.** `internal/platform/config/feature_flags.go` defines `FeatureFlagsConfig` and `LoadFeatureFlagsConfig`, while `internal/platform/featureflags/handler.go` imports it. The config struct is owned in `config`; the handler in `featureflags`. This is the correct split, but the JSON response key `"MDDM_NATIVE_EXPORT_ROLLOUT_PCT"` is a string literal in the handler (`featureflags/handler.go:29`) with no tie to the env var name or any typed constant — a future rename of the env var would not be caught at compile time. *Smell: loose coupling between env var name and wire-format key (RF-8 lifecycle gap).*

- **`RepositoryMemory` mode is a dead path in production.** `config/repository.go:9-24` defines `RepositoryMemory = "memory"` as a valid mode, and `bootstrap/api.go:127-153` supports it. However, `apps/api/cmd/metaldocs-api/main.go:677-680` calls `requirePostgresRepositoryMode` which fatal-exits if mode is not `"postgres"`. Memory mode is only exercised in `bootstrap/api_test.go` and `bootstrap/worker_test.go`. The constant and its handling path are kept alive solely by tests; production cannot reach them. Git history shows `repository.go` was introduced at project bootstrap (`af8f8b43c`). *Smell: dead production path; REQ-TOP-3 adjacent — a permanent test-only mode should be documented as such.*

- **Dependency checks in `applyDependencyChecks` are sequential, not parallel.** `observability/runtime.go:286-323` runs each `DependencyCheck.Check` sequentially under individual 2-second timeouts. With N checks, worst-case blocking on `/api/v1/health/ready` is 2N seconds + the outer 3-second Postgres ping, not 3 seconds total. Currently only Gotenberg is registered (1 check), so this is not a production risk today, but the pattern is misleading — the outer deadline comment implies all checks share the budget, while the actual implementation is additive. (`runtime_test.go:30-60` tests the outer deadline, which passes because the test uses `nil` db and only measures the outer context expiry on DB nil-check, not the sequential check overhead.) *Smell: misleading concurrency contract / future-scaling hazard.*

- **`METALDOCS_ATTACHMENTS_SIGNING_SECRET` is required unconditionally, even for `StorageProviderMemory`.** `config/attachments.go:63-68` requires the secret for all providers, including `"memory"`, which writes to local disk and never uses HMAC-signed URLs. This forces developers and CI to provide a secret value even when running in memory mode. The enforcement is security-conservative but operationally inconvenient without documentation. *Smell: over-strict validation for non-production storage modes; no comment explains the invariant.*

- **`.gitkeep` in `internal/platform/observability/`** (`observability/.gitkeep`) is a residual scaffold artifact. The package now has real files. *Smell: harmless noise; REQ-TOP-3 adjacent empty-scaffold cleanup.*

- **No OpenTelemetry anywhere in the platform.** Zero OTel imports exist in `internal/platform` or `apps/`. The observability package uses a custom `X-Trace-Id` header, in-process JSON counters, and `slog` to stdout. This is a confirmed finding that validates RF-1: exporter wiring is missing, cross-service trace propagation does not use W3C `traceparent`, and no trace/span data is exported to any backend. *Smell: observability gap vs REQ-OBS-3 target; RF-1 open.*

---

## 11. Wiki drift

No existing wiki doc covers `internal/platform/config`, `internal/platform/observability`, or `internal/platform/featureflags` directly. Section D4 of `wiki/architecture/backend-blueprint.md` grades config as ✅ and states "fail-fast on missing secrets" — this is accurate. Section D2 grades observability as 🟡 and flags "exporter wiring, cross-service trace, readiness-probe depth" as unverified — this audit confirms all three are missing (no OTel, custom trace ID only, readiness checks are sequential not budget-shared).

One claim in the blueprint is now slightly inaccurate:

> `wiki/architecture/backend-blueprint.md` line 207–208: "We have: `platform/observability` (metrics, tracing, structured logging)…"

The word "tracing" is misleading. The package propagates a `trace_id` string value via a custom header and context key, but there is no distributed tracing instrumentation, no span creation, and no trace exporter. A reader could interpret "tracing" as OTel-compatible distributed tracing; the code provides only request-ID correlation.

---

## 12. Open questions

- **[runtime-unverified]** Does `PostgresRuntimeStatusProvider.RuntimeMetrics` execute its three queries against `metaldocs.auth_identities`, `metaldocs.auth_sessions`, and `metaldocs.outbox_events` within the correct schema on a live Postgres instance? The query SQL uses the unqualified table names prefixed with `metaldocs.` — the schema search path on the connection pool is set elsewhere (`platform/db/postgres/connect.go`). Correctness depends on the pool's `search_path` or explicit schema qualification. Docker is currently down; verification requires a live instance.

- **[runtime-unverified]** `LoadRateLimitConfig()` embeds `TrustedProxyCIDRs` from `LoadTrustedProxyCIDRs()`. The rate-limit middleware is expected to use these CIDRs to decide whether to trust `X-Forwarded-For` for IP-keyed buckets. The actual honoring logic is in `internal/platform/ratelimit/middleware.go` (outside scope of this area). Whether the CIDR list flows through correctly to the rate-limiter's IP extraction has not been verified in this audit.

- **[design question]** The feature-flag config is loaded once at startup with no hot-reload. If `METALDOCS_MDDM_NATIVE_EXPORT_ROLLOUT_PCT` needs to change in production, a process restart is required. The handler's `Cache-Control: no-store` prevents browser caching, but the server-side value is frozen. Whether a restart-based rollout lifecycle is acceptable for production flag management is undocumented (RF-8).

- **[design question]** `WorkerConfig.ReviewReminderDays` is defined in `config/worker.go:11` and loaded by the worker binary, but the worker binary (`apps/worker/cmd/metaldocs-worker/main.go`) does not reference `ReviewReminderDays` directly — it passes the full `WorkerConfig` to `bootstrap.BuildWorkerDependencies`. Whether this field reaches the review-reminder logic downstream and whether the default of 14 days is the production value has not been verified in this audit.
