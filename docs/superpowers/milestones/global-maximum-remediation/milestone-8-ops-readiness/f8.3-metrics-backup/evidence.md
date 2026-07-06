# Feature F8.3 — Evidence

> **Milestone:** 8 — Ops Readiness  ·  **Feature:** `f8.3-metrics-backup`  ·  **Closed:** 2026-07-06
> **Contract:** `spec.md` (Prometheus `/metrics` outside `/api/v1` + live log↔trace↔header
> correlation + executed backup/restore runbook).

## What was implemented

- **`/metrics` (Prometheus)** — `internal/platform/observability/prometheus.go` adds a dedicated
  registry with `metaldocs_http_requests_total`, `metaldocs_http_request_errors_total`,
  `metaldocs_http_request_duration_seconds` (all `{route,method}`), plus Go runtime + process
  collectors and DB-pool gauges. Vec increments reuse the route/method/duration already computed in
  `Wrap` (no double-count with the JSON registry). Served from an **infra-port `rootMux`** wired in
  `main.go` ahead of the middleware chain, so `GET /metrics` bypasses auth/iam/obs/rate-limit while
  everything else falls through to the normal chain. JSON `/api/v1/metrics` is untouched.
- **Correlation** — `HTTPObservability.Wrap` now emits the resolved trace id as the `X-Trace-Id`
  response header (`http.go`, commit `6bbb7aa8`), set before the downstream handler. Resolution order
  unchanged: active OTel span → inbound `X-Trace-Id` → requesttrace. This closes the
  log↔trace↔response-header triple at the root cause (a header echo), not a patch.
- **Backup runbook** — `wiki/runbooks/backup-restore.md` wraps the existing PowerShell scripts and was
  executed end-to-end (backup → scratch-DB restore → row-count validation). Surfaced + fixed a real
  backup-infra gap (see Verification): the dedicated backup role needs `BYPASSRLS` + `public`-schema
  grants to dump the M7 FORCE-RLS tables; `scripts/sql/create-backup-role.sql` + runbook now codify it
  (commit `b15a4480`). `ops/DEPLOY.md` pg_dump line points at the runbook.
- Producer matches consumer contract: scraper gets exposition-format families per gate; JSON pin
  consumers see an unchanged `/api/v1/metrics`; on-call gets a top-to-bottom-executable runbook.
- Commits: `1bbb59be` (/metrics + prometheus), `6bbb7aa8` (X-Trace-Id header + test),
  `b15a4480` (RLS-capable backup role + runbook), `74a08e61` (runbook + DEPLOY.md).

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — /metrics bypasses auth | `go test ./apps/api/cmd/metaldocs-api/ -run Metrics` | `ok ...metaldocs-api 4.697s` — `TestMetricsEndpoint_BypassesAuthChain_REQF83` (GET /metrics unauth→200, POST→401 falls through, self-count guard) | real |
| TDD — X-Trace-Id header | `go test ./internal/platform/observability/...` | `ok ...observability 7.133s` incl. `TestHTTPObservability_Wrap_SetsTraceIDResponseHeader` + inbound-echo test | real |
| /metrics serves 4 binding groups (live) | `docker exec metaldocs-api wget -qO- :8081/metrics \| grep '# TYPE'` | present: `metaldocs_http_requests_total` (counter), `metaldocs_http_request_errors_total` (counter), `metaldocs_http_request_duration_seconds` (histogram), `go_goroutines` (gauge), `process_cpu_seconds_total` (counter); 85 `metaldocs_*` samples | **real** (containerized api) |
| No /api/v1 contract change | `git diff 218e2d12 -- api/openapi` | empty (no bytes changed) | real |
| **Correlation (live)** | one-off api container `OTEL_TRACES_EXPORTER=console`, drive `GET /health/ready`, wait for BatchSpanProcessor flush, match ids | request `e8dc2036c85c2fd0762b2dac91da7a79`: response `X-Trace-Id: e8dc2036…` == slog `trace_id: e8dc2036…` (path /health/ready) == console span `TraceID: e8dc2036…` (SpanID 767322e3de6284bf, url.path /health/ready) — **one id across all three** | **real** |
| Runbook end-to-end | backup (non-superuser BYPASSRLS role) → scratch-DB restore → row-count validate | dump 535 KB, sha256 `464583575e099c73…80ca7`, restore 70 relations; validate `iam_users` 127==127 exact; `audit_events` restored 1780 vs source-now 1788 (+8 = live audit-append drift during dump → consistent point-in-time snapshot, not data loss) | **real** (see deviations) |
| Existing observability tests green | `go test ./internal/platform/observability/...` | `ok 7.133s` | real |

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| /metrics serves exposition format with the 4 binding groups | yes | Go test + live grep (5 families incl. Go/process) |
| No /api/v1 contract change | yes | `git diff` on `api/openapi` empty; JSON pin tests green |
| Correlation: span id == slog trace_id == X-Trace-Id header | yes | live triple `e8dc2036…` verbatim |
| Runbook executes end-to-end | yes | backup→restore→validate PASS (deviations labeled) |
| Existing observability tests green | yes | observability suite ok |

## Review disposition

- Spec-compliance review: PASS — `/metrics` mounted outside `/api/v1`; openapi.yaml unchanged; no
  Grafana/collector/worker-jobs endpoints; no new backup implementation (documented + proved existing
  scripts). Initial review found `/metrics` behind the auth chain (401) — fixed at root cause via the
  infra-port `rootMux` (dispatched agent), re-verified live 200 unauth.
- Code-quality review: PASS — Prometheus vecs reuse computed labels (no double-count); header set
  before body write; backup-role fix keeps the role non-superuser/SELECT-only (BYPASSRLS bypasses RLS
  reads only, grants no write/DDL).

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| **Deviation, not defer — host PG client tools absent.** Backup/restore drive ran pg_dump/pg_restore/psql via a `postgres:16` client container against the DB, not host binaries. | Same client/server major; proves the runbook logic; the runbook already instructs installing matching client tools on the ops host. | Ops host provisioning owns installing PG16 client tools per runbook prereq |
| Fleet scrape topology (worker/jobs `/metrics`) | Rabbit hole per spec; single-node v1 scrapes api only | First multi-node deploy (runbook note) |
