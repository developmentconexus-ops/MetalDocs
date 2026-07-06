# Feature F8.3 — Prometheus /metrics + correlation + backup runbook

> **Milestone:** 8 — Ops Readiness  ·  **Folder:** `f8.3-metrics-backup`
> **Status:** Planning

## Source

- Milestone spec row F8.3 (`../milestone.md`) + validation-contract §3 (binding).

## Plan

**Task A — /metrics endpoint (implementer subagent, sonnet, TDD)**
Files: `internal/platform/observability/` (new prometheus.go + test), `apps/api/cmd/metaldocs-api/main.go`
(mount `/metrics`), `go.mod` (promote client_golang to direct).
1. Failing test: scrape test handler, assert 4 binding family groups present
   (`metaldocs_http_requests_total`, `metaldocs_http_request_errors_total`,
   `metaldocs_http_request_duration_seconds` histogram, `go_*`/`process_*`, db pool gauges).
2. Implement: custom `prometheus.Collector` reading the existing `HTTPObservability` per-route
   counters (no double instrumentation) + native histogram observed in the same Wrap path where
   duration is already recorded; registry with runtime/process collectors + DBPool gauge collector;
   `promhttp.HandlerFor` mounted at `/metrics` (outside /api/v1; NOT in openapi.yaml).
3. Verify gateway/nginx config does not proxy `/metrics` publicly; report disposition (no edit unless
   it does — then smallest deny rule).
4. Existing observability tests green; `go build ./...`.

**Task B — backup/restore runbook (implementer subagent, sonnet, doc)**
Files: `wiki/runbooks/backup-restore.md` (new), `ops/DEPLOY.md` pg_dump line pointer (coordinate with
F8.1 Task C rewrite — B runs after C lands to avoid conflict), wiki index entry.
Content per contract §3.4: prerequisites (dedicated role, refuses metaldocs_app), backup, restore
(scratch DB, never live), validation gate, recovery guidance, scrape-topology trigger note.

**Task C — live proofs (main session)**
- Correlation drive: api with `OTEL_TRACES_EXPORTER=console`, one authenticated request, capture span
  + slog line + X-Trace-Id → one id (contract §3.3).
- `curl /metrics` grep families.
- Execute runbook end-to-end: backup → scratch-DB restore → gate PASS; transcripts.

Order: A → (B after F8.1 Task C) → C. Reviews after A and B.

## Execution notes

(filled during execution)
