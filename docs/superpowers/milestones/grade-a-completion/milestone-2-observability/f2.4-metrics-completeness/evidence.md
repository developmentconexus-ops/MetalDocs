# Feature F2.4 — Evidence

> **Milestone:** 2 — Composition / observability  ·  **Feature:** `f2.4-metrics-completeness`
> **Closed:** 2026-06-16
> **Status:** CLOSED — all spec rows met

## Validation Gate results

| Row | Criterion | Command / proof | Result | Fixture vs real |
|-----|-----------|-----------------|--------|-----------------|
| 1 | `"db_pool"` key present when `SetDBPool` called | `go test ./internal/platform/observability/ -run TestMetricsHandler_IncludesDBPoolStats` → PASS | PASS | fixture |
| 2 | `"db_pool"` key absent when `SetDBPool` not called | `go test ./internal/platform/observability/ -run TestMetricsHandler_NoDBPoolKey_WhenNotWired` → PASS | PASS | fixture |
| 3 | Adapter maps `sql.DBStats` fields to snake_case | `go test ./internal/platform/db/postgres/ -run TestSQLDBPoolStatsAdapter_Maps` → PASS | PASS | fixture |
| 4 | "Prometheus" comment removed | `grep -c '"Prometheus scrape endpoint"' apps/api/cmd/metaldocs-api/permissions.go` → 0 | PASS | static |
| 5 | Existing metrics keys unchanged | All observability package tests pass (regression) | PASS | fixture |
| 6 | Whole-repo regression | `go test ./...` exits 0; no FAIL lines | PASS | fixture |
| 7 | Runtime proof — `db_pool` in live `/api/v1/metrics` | See §Real-provider below | PASS | **real-provider** |

## Real-provider evidence

API rebuilt with F2.4 changes, started via `.\scripts\start-api.ps1 -Build`. Fresh login with session cookie. `GET /api/v1/metrics` response:

```json
{
  "db_pool": {
    "idle": 2,
    "in_use": 0,
    "max_idle_closed": 0,
    "max_idle_time_closed": 0,
    "max_lifetime_closed": 0,
    "max_open_connections": 25,
    "open_connections": 2,
    "wait_count": 0,
    "wait_duration_ms": 0
  },
  "items": [
    {
      "route": "/api/v1/auth/login",
      "method": "POST",
      "requests": 5,
      "errors": 0,
      "duration_total_ms": 1079,
      "avg_duration_ms": 215
    }
  ],
  "runtime": { "repositoryStatus": "up", "repositoryMode": "postgres" },
  "scheduler": { "jobs": {} }
}
```

**Key evidence:**
- `"db_pool"` top-level key present (spec C1)
- `"open_connections": 2` — non-zero, proving live `*sql.DB.Stats()` call at request time (spec C6)
- `"idle": 2` — pool has idle connections (startup migrations completed)
- `"max_open_connections": 25` — pool config visible (spec C6)
- All snake_case keys: `in_use`, `idle`, `open_connections`, `wait_count` all present (spec C6)
- `"items"`, `"runtime"`, `"scheduler"` keys still present — no regression to existing payload (spec C4)

## Comment fix verified

```
Select-String -Path apps/api/cmd/metaldocs-api/permissions.go -Pattern "Prometheus scrape endpoint"
→ (no output)
```

Line 95 now reads: `// Metrics — JSON scrape endpoint; read-only.`

## Commits

| Commit | Description |
|--------|-------------|
| `5d7d2083` | docs(f2.4): spec.md + plan.md — db_pool metrics + Prometheus comment fix |
| `99cd68e8` | test(f2.4): failing tests for db_pool metrics key and pool stats adapter [TDD red] |
| `750902ce` | feat(observability): DBPoolStatsProvider interface + SetDBPool setter (F2.4) |
| `7e4dc036` | feat(postgres): SQLDBPoolStatsAdapter — DBPoolStats() for metrics endpoint (F2.4) |
| `6f98440c` | feat(api): wire db_pool stats into /api/v1/metrics; fix JSON comment (F2.4) |

## Defers

None. All spec rows closed with evidence.
