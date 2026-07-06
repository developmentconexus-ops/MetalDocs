# Feature F8.3 — Spec (Prometheus /metrics + correlation proof + backup runbook)

> **Milestone:** 8 — Ops Readiness  ·  **Folder:** `f8.3-metrics-backup`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-07-05 — operator delegated approval via the M8 `/goal` brief;
> binding shape `../validation-contract.md` §3 (committed `218e2d12`).

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | Prometheus endpoint or ADR-keep-JSON? | Brief prefers real endpoint; JSON `/api/v1/metrics` is contract-bound (OpenAPI-mirrored, pin-tested) so it STAYS; `/metrics` added alongside, outside the versioned API surface (contract §3.2 — binding). |
| 2 | (verified) Existing counters? | Per-route requests/errors/duration(+p50/p95/p99) in `observability.HTTPObservability` (`http.go:211–239`); runtime/scheduler/dbpool providers map-backed. `prometheus/client_golang` already indirect in go.mod. |
| 3 | (verified) Correlation code path? | slog `http_request` line carries `trace_id` (`http.go:157–167`); resolution: active OTel span → X-Trace-Id → requesttrace (`http.go:94–112`). Proof = live match of span id ↔ log line ↔ response header. |
| 4 | (verified) Backup tooling? | Scripts exist (`backup-postgres.ps1` etc. + `run-backup-restore-gate.ps1` with restore smoke + evidence JSON); missing piece = wiki runbook + executed end-to-end proof. Global-max judgment: document + prove existing sound scripts; do NOT rebuild. |
| 5 | Worker/jobs scrape? | OUT (milestone rabbit hole) — trigger recorded in runbook/ADR-0071-adjacent note: fleet scrape topology at first multi-node deploy. |

## Consumer contract (FIRST)

- **Consumers:** (a) a Prometheus scraper — GET `/metrics` on the api binary, Prometheus text
  exposition format 0.0.4, families per gate below; (b) on-call operator — runbook
  `wiki/runbooks/backup-restore.md` executable top-to-bottom; (c) existing JSON consumers/pin tests —
  `GET /api/v1/metrics` wire shape untouched (`http_typed_test.go`, `http_scheduler_test.go`,
  `http_dbpool_test.go` stay green).
- **Contract (binding minimum metric groups):** `metaldocs_http_requests_total{route,method}`;
  `metaldocs_http_request_errors_total{route,method}`; request-duration histogram
  `metaldocs_http_request_duration_seconds{route,method}`; Go runtime + process collectors; DB pool
  gauges from `DBPoolStats`. No tenant data. No double-counting with the JSON registry.
- **Source of truth:** validation-contract §3 + Prometheus exposition format spec.

## What this feature implements

(a) `/metrics` (promhttp, direct `client_golang` dependency) mounted on the api mux outside
`/api/v1`, gateway-exposure disposition recorded (not publicly proxied by default — verify nginx/
gateway config); (b) live log↔trace correlation proof with `OTEL_TRACES_EXPORTER=console` — identical
trace id in span dump, slog line, and `X-Trace-Id` response header; defect (if mismatch) fixed at the
resolution root cause; (c) `wiki/runbooks/backup-restore.md` wrapping the existing scripts, executed
end-to-end (backup → scratch-DB restore → validation gate PASS) with transcript; `ops/DEPLOY.md`
pg_dump line pointed at the runbook.

## Non-goals (mandatory)

- No Grafana/alerting/OTel-collector deployment; no worker/jobs metrics endpoints.
- No change to JSON `/api/v1/metrics` shape or its OpenAPI schema; no openapi.yaml edit at all.
- No new backup implementation; no offsite/S3 automation.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| /metrics serves exposition format with the 4 binding groups | Go test asserting families via `promhttp` test scrape + live `curl /metrics` grep | real |
| No /api/v1 contract change | openapi.yaml untouched (git diff empty for `api/openapi`); JSON pin tests green | real |
| Correlation | live drive: console span trace id == slog `trace_id` == `X-Trace-Id` header, captured verbatim | real |
| Runbook executes end-to-end | backup → restore to scratch DB → `run-backup-restore-gate.ps1` (or validate script) PASS, transcript + evidence JSON | real |
| Existing observability tests green | `go test ./internal/platform/observability/...` | real |

## ADR needed?

- [x] No separate ADR — `/metrics`-outside-contract is a contract §3.2 binding clause; scrape-topology
  trigger recorded in the runbook; rate-limit ADR 0071 covers the ops-scale decision family.
