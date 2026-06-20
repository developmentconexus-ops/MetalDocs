# Feature F8.2 — Evidence (metrics typed envelope + OpenAPI alignment)

> **Milestone:** 8  ·  **Feature:** `f8.2-metrics-typed-envelope`  ·  **Closed:** 2026-06-20
> **Contract:** `spec.md` (approved 2026-06-20). Plan: `plan.md`.
> **Commit:** recorded at commit time below.

## What was implemented

- **OpenAPI** ([`api/openapi/v1/openapi.yaml:5013`](../../../../../api/openapi/v1/openapi.yaml)) —
  `MetricsResponse` now declares `scheduler` and `db_pool` (`type: object, additionalProperties: true`,
  optional). Closes the **live divergence**: the handler has emitted both keys for releases, but the
  schema declared only `items`+`runtime`.
- **BE** ([`internal/platform/observability/http.go:175`](../../../../../internal/platform/observability/http.go)) —
  added platform-local exported `MetricsResponse{Items []metricItem; Runtime/Scheduler/DBPool map[string]any}`
  and `MetricsHandler` now builds and emits it via `httpresponse.WriteJSON`, replacing the top-level
  `payload := map[string]any{…}` response literal. Removed the now-unused `encoding/json` import.
- **FE** ([`frontend/apps/web/src/lib/api-types/index.d.ts`](../../../../../frontend/apps/web/src/lib/api-types/index.d.ts)) —
  regenerated via `npm run gen:api`; `MetricsResponse` gains the two optional keys (+8 lines).

**REQ-TOP-2 honored:** the envelope is platform-local (no module-generated import); the three dynamic
sub-objects stay `map[string]any` (operator decision 2026-06-20 — providers unchanged).
**Producer matches consumer contract:** ops/FE consumer expects `{ items, runtime, scheduler?, db_pool? }`;
the typed envelope emits exactly that, and the OpenAPI now declares every emitted top-level key.

## Wire-equivalence (one honest caveat)

The prior literal keyed `runtime`/`scheduler`/`db_pool` off **provider != nil**; the typed struct keys off
`omitempty` (map emptiness). For all wired providers (which return populated maps) the wire is identical —
proven by the **existing** parity-lock tests passing unmodified. The single untested edge: a wired provider
returning an **empty non-nil** map would now omit its key instead of emitting `{}`. Both decode to "no data"
under the OpenAPI optional `additionalProperties:true` object, so it is within contract. No production
provider returns an empty map.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| Wire-shape parity-lock (existing, unmodified) | `go test -count=1 ./internal/platform/observability/...` | `ok` (9.283s) — incl. `TestMetricsHandler_IncludesDBPoolStats`, `_NoDBPoolKey_WhenNotWired`, `_IncludesSchedulerCounters`, `_NoSchedulerKey_WhenNotWired` | real |
| New typed-envelope key-set test | `TestMetricsHandler_TypedEnvelope_KeySet` / `_ItemsAlwaysPresent` | PASS — all-wired body = `{items,runtime,scheduler,db_pool}`; no-provider body = `{items}` only | real |
| H-D red→green (top-level response literal removed) | `grep -nE 'map\[string\]any\{' internal/platform/observability/http.go` | **0 hits (exit 1)** (was 1: `payload := map[string]any{"items":…}`) | real |
| OpenAPI declares every emitted top-level key | `grep -n 'scheduler\|db_pool' …/openapi.yaml` | `:5024 scheduler`, `:5028 db_pool` under `MetricsResponse` | real |
| FE codegen regenerated & committed | `git diff --stat …/api-types/index.d.ts` | `1 file changed, 8 insertions(+)` — both keys optional | real |
| Static (build) | `GOFLAGS=-mod=mod go build ./...` | exit 0 | — |

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| No top-level response `map[string]any` literal in metrics handler | yes | grep 1→0 |
| OpenAPI declares every emitted top-level key | yes | `:5024`/`:5028` under `MetricsResponse` |
| FE codegen regenerated & committed | yes | `index.d.ts` +8 lines, both optional |
| Wire shape unchanged | yes | existing 4 parity-lock tests pass unmodified + new key-set test |

## Review disposition

- Spec-compliance review: PASS — divergence closed; REQ-TOP-2 respected (platform-local types, no module
  import); non-goals (provider interfaces, sub-object typing, scrape format) untouched.
- Code-quality review: PASS — conditional assignment preserved; dead `encoding/json` import removed;
  `omitempty` caveat documented; runtime stays in OpenAPI `required` (pre-existing, untouched).

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| None | | |
