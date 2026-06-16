# Feature F2.4 — Spec

> **Milestone:** 2 — Composition / observability  ·  **Folder:** `f2.4-metrics-completeness`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-16 — leandrotca.work — approved 2026-06-16.

> This is the feature's **contract**, written and approved **before any code**. It defines what the
> feature must do and how it will be proven — not how it will be built (that is `plan.md`). The
> milestone-validator judges the feature against *this* file (C1).

## Interview record (fail-closed gate)

Inline interview — milestone.md D4 fully constrains both changes; no ambiguity requiring external
clarification.

| # | Question | Answer |
|---|----------|--------|
| 1 | Which `sql.DBStats` fields to expose and under what key? | All 9 stdlib fields (`MaxOpenConnections`, `OpenConnections`, `InUse`, `Idle`, `WaitCount`, `WaitDuration`, `MaxIdleClosed`, `MaxIdleTimeClosed`, `MaxLifetimeClosed`) under top-level key `"db_pool"`. SREs scraping this endpoint need the full pool picture — partial exposure creates misleading dashboards. All fields are safe to expose (no credentials, no PII). |
| 2 | What comment replaces the "Prometheus scrape endpoint" string? | `// Metrics — JSON scrape endpoint; read-only.` — matches reality (the payload is `application/json`, not Prometheus text-format). No other change to the permissions.go entry. |
| 3 | Injection pattern — pass `*sql.DB` directly to `SetDBPool`, or wrap in a provider interface? | Provider interface (`DBPoolStatsProvider`), same pattern as `SchedulerMetricsProvider`. Avoids importing `database/sql` into the observability package via a concrete type if it ever changes; stays consistent with the established pattern and keeps the setter nil-safe for tests. |

## Consumer contract (FIRST — before any producer)

- **Consumer(s):**
  - **Primary — SRE scraping `/api/v1/metrics`** who needs DB connection pool stats (`InUse`, `Idle`,
    `WaitCount`) to detect pool exhaustion and connection churn.
  - **Secondary — code reader** who sees `permissions.go:95` and should not be misled into thinking
    the endpoint emits Prometheus text format.

- **Contract:**
  1. `GET /api/v1/metrics` response JSON includes a top-level `"db_pool"` object when a `*sql.DB`
     is wired. The object contains at minimum the keys `in_use`, `idle`, `open_connections`,
     `wait_count` (snake_case). The full set of keys mirrors `sql.DBStats` fields.
  2. When no `*sql.DB` is wired (e.g. nil `deps.SQLDB` in test/standalone mode), the `"db_pool"`
     key is **absent** (not null, not `{}`). Same nil-guard pattern as `"scheduler"`.
  3. The comment at `apps/api/cmd/metaldocs-api/permissions.go:95` reads
     `// Metrics — JSON scrape endpoint; read-only.` — no other change to that file.
  4. Existing `"items"`, `"runtime"`, and `"scheduler"` keys in the metrics payload are **unchanged**.
     This feature only adds `"db_pool"`.
  5. The `HTTPObservability` struct gains a `dbPool DBPoolStatsProvider` field and a
     `SetDBPool(p DBPoolStatsProvider)` setter — same set-once pattern as `SetSchedulerMetrics`.
  6. `*sql.DB` implements `DBPoolStatsProvider` via a thin adapter (not by adding methods to
     `*sql.DB`). The adapter wraps `db.Stats()` and maps fields to snake_case map keys.

- **Source of truth for the contract:**
  - Metrics handler → [`internal/platform/observability/http.go:163`](../../../../../internal/platform/observability/http.go)
  - Comment fix → [`apps/api/cmd/metaldocs-api/permissions.go:95`](../../../../../apps/api/cmd/metaldocs-api/permissions.go)
  - DB pool wiring → [`apps/api/cmd/metaldocs-api/main.go`](../../../../../apps/api/cmd/metaldocs-api/main.go) (after `httpObs` creation)

## What this feature implements

Two scoped changes:

1. **`internal/platform/observability/http.go`**
   - Add `DBPoolStatsProvider` interface: `DBPoolStats() map[string]any`.
   - Add `dbPool DBPoolStatsProvider` field on `HTTPObservability`.
   - Add `SetDBPool(p DBPoolStatsProvider)` setter.
   - Add nil-guard in `MetricsHandler`: `if o.dbPool != nil { payload["db_pool"] = o.dbPool.DBPoolStats() }`.

2. **`internal/platform/db/postgres/`** (or a thin `observability/` subpackage — either is fine)
   - Add `SQLDBPoolStatsAdapter` (or equivalent) wrapping `*sql.DB` and implementing `DBPoolStatsProvider`.
   - `DBPoolStats()` calls `db.Stats()` and returns a `map[string]any` with snake_case keys.

3. **`apps/api/cmd/metaldocs-api/main.go`**
   - After `httpObs` is created (line ~277), call `httpObs.SetDBPool(postgres.NewPoolStatsAdapter(deps.SQLDB))` if `deps.SQLDB != nil`.

4. **`apps/api/cmd/metaldocs-api/permissions.go:95`**
   - Change comment from `// Metrics — Prometheus scrape endpoint; read-only.` to
     `// Metrics — JSON scrape endpoint; read-only.`.

5. **New test files:**
   - `internal/platform/observability/http_dbpool_test.go` — asserts `"db_pool"` key present when
     wired and absent when not wired; asserts `"in_use"` / `"idle"` keys exist in the pool map.
   - `internal/platform/db/postgres/pool_stats_test.go` — asserts adapter maps `sql.DBStats` fields
     to correct snake_case keys.

## Non-goals (mandatory)

- **No Prometheus/OpenMetrics format.** The endpoint stays `application/json` (existing shape). No
  exporter, no format migration, no `/metrics` rename. HS-2 trigger.
- **No other comment changes.** Only the one `"Prometheus"` comment at `permissions.go:95`.
- **No pool tuning.** This feature reads pool stats; it does not change pool configuration.
- **No new endpoint.** All stats flow through the existing `/api/v1/metrics` route.
- **No M3/M4 scope.** Any code-quality or module-port findings stay in their milestones.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| 1. `"db_pool"` key present when `SetDBPool` called with non-nil adapter | `go test ./internal/platform/observability/ -run TestMetricsHandler_IncludesDBPoolStats -count=1` — stub adapter, assert `"db_pool"` key and `"in_use"` sub-key in JSON response | fixture |
| 2. `"db_pool"` key absent when `SetDBPool` not called | `go test ./internal/platform/observability/ -run TestMetricsHandler_NoDBPoolKey_WhenNotWired -count=1` — no `SetDBPool`, assert `"db_pool"` absent | fixture |
| 3. Adapter maps `sql.DBStats` fields to snake_case keys correctly | `go test ./internal/platform/db/postgres/ -run TestSQLDBPoolStatsAdapter_Maps -count=1` — construct `sql.DBStats` with known values, assert `"in_use"`, `"idle"`, `"open_connections"`, `"wait_count"` present with correct values | fixture |
| 4. Comment fix verified | `grep -c '"Prometheus scrape endpoint"' apps/api/cmd/metaldocs-api/permissions.go` returns `0` | static grep |
| 5. Existing metrics keys unchanged | `go test ./internal/platform/observability/ -run TestMetricsHandler -count=1` — all existing tests pass (regression) | fixture |
| 6. Whole-repo regression | `go test ./...` exits 0; no FAIL lines | fixture |
| 7. Runtime proof — `curl /api/v1/metrics` payload includes `"db_pool"` with live pool values | Start API; `curl -s ... /api/v1/metrics | jq .db_pool`; paste output with non-zero `open_connections` or `idle` into `evidence.md` | **real-provider** |

## ADR needed?

- [ ] No durable decision required — this feature adds pool stats to an existing endpoint via the
  same pattern already established by F2.2 (`SchedulerMetricsProvider`). No new architectural
  choice.
