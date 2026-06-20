# Feature F8.2 — Spec (SEED — approval pending)

> **Milestone:** 8 — Grade-A Contract & Boundary Completion  ·  **Folder:** `f8.2-metrics-typed-envelope`
> **Status:** Drafting (seed from post-M7 re-audit Composition Major #5; **interview + approval pending**)
> **Approved before code:** PENDING — *no implementation begins until this line is filled (Phase 3, fresh session).*

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | Type the sub-objects (`runtime`/`scheduler`/`db_pool`) or keep dynamic? | **Keep dynamic** (operator decision 2026-06-20): declare them in OpenAPI as `additionalProperties:true`; type only the top-level envelope. Minimal provider churn. |
| 2 | Where does the typed envelope live (REQ-TOP-2)? | Platform-local in `internal/platform/observability/` — hand-rolled `MetricsResponse`/`MetricItem`; platform must not import a module's generated package. |

## Consumer contract (FIRST)

- **Consumer(s):** ops/metrics consumer of `GET /api/v1/metrics`; FE generated types.
- **Contract:** `200 application/json` `{ items: MetricItem[], runtime: object, scheduler?: object, db_pool?: object }`; `MetricItem` per `openapi.yaml:4991`. `runtime/scheduler/db_pool` are declared dynamic objects.
- **Source of truth:** OpenAPI `MetricsResponse` (`openapi.yaml:5013`) — **must be extended** to declare `scheduler`+`db_pool` (currently only `items`+`runtime` declared → live divergence).

## What this feature implements

1. OpenAPI: add `scheduler` + `db_pool` (`type: object, additionalProperties: true`) to `MetricsResponse`; regen FE types.
2. `observability/http.go:175-196` — define platform-local `MetricsResponse{Items []MetricItem; Runtime, Scheduler, DBPool map[string]any}` (the map fields are declared-dynamic, allowlist-legitimate) and emit it typed; remove the top-level `payload := map[string]any{…}` literal.

## Non-goals (mandatory)

- No change to provider interfaces (`SchedulerMetrics()/DBPoolStats()/RuntimeMetrics()` stay `map[string]any`).
- No fully-typed sub-object structs.
- No new metrics fields or scrape format change.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| No top-level response `map[string]any` literal in metrics handler | read `observability/http.go` MetricsHandler | real |
| OpenAPI declares every emitted top-level key | `grep -n 'scheduler\|db_pool' api/openapi/v1/openapi.yaml` under MetricsResponse | real |
| FE codegen regenerated & committed | `git diff --stat` on generated FE types | real |
| Wire shape unchanged | observability metrics handler test asserting key set | real |

## ADR needed?

- [x] No durable decision — skip (declared-dynamic is consistent with existing audit `Payload`/security `Evidence` allowlist).
