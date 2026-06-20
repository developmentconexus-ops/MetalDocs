# Feature F8.2 — Plan

> Engine: inline (superpowers:writing-plans structure). Spec: `./spec.md` (approved 2026-06-20).
> Contract-first regen order: OpenAPI → FE codegen. BE stays hand-rolled platform-local (REQ-TOP-2).

## Files touched

| File | Change |
|------|--------|
| `api/openapi/v1/openapi.yaml` | Under `MetricsResponse.properties`, add `scheduler` and `db_pool` as `type: object, additionalProperties: true` (optional — not added to `required`). Closes the live divergence (handler already emits them). |
| `frontend/apps/web/src/lib/api-types/index.d.ts` | Regenerated via `npm run gen:api` (openapi-typescript). Commit the diff. |
| `internal/platform/observability/http.go` | Add exported `MetricsResponse` struct (`Items []metricItem` + `Runtime/Scheduler/DBPool map[string]any` with `omitempty`); build it conditionally in `MetricsHandler`; emit via `httpresponse.WriteJSON`; remove the `payload := map[string]any{…}` literal. |
| `internal/platform/observability/http_typed_test.go` | NEW — assert the all-providers-wired key set is exactly `{items, runtime, scheduler, db_pool}` and `items` is present with no providers. |

## Test strategy

- **Class:** platform handler-unit (`package observability`, `httptest` + stub providers) — the canonical
  pattern already in `http_scheduler_test.go` / `http_dbpool_test.go`.
- The **existing** `TestMetricsHandler_IncludesDBPoolStats`, `_NoDBPoolKey_WhenNotWired`,
  `_IncludesSchedulerCounters`, `_NoSchedulerKey_WhenNotWired` decode the body into `map[string]any` and
  assert key presence/absence — they are the **wire-shape parity-lock** and must pass **unmodified**.
- New `http_typed_test.go`: parity-lock for the envelope key-set (`items` unconditional; runtime/scheduler/
  db_pool present iff wired). Its teeth = catching an accidental field-name change or a dropped `items`.
- **red→green:** the OpenAPI divergence — `grep scheduler|db_pool` under `MetricsResponse` returns 0 before,
  2 after; and the BE top-level `map[string]any` response literal (1 → 0).

## Task order

1. OpenAPI: add `scheduler` + `db_pool` to `MetricsResponse`. Grep to confirm.
2. `npm run gen:api` (frontend/apps/web) → regenerate `index.d.ts`. Inspect diff (MetricsResponse gains the
   two optional keys).
3. `http.go`: define `MetricsResponse` struct; rebuild `MetricsHandler` conditionally; emit typed; drop literal.
4. Write `http_typed_test.go`. Run observability tests (incl. existing four) → green.
5. `go build ./...`; `go test -count=1 ./internal/platform/observability/...`; FE typecheck if cheap.
6. Evidence + commit.

## Risk / rollback

- Honest edge (Q3): empty non-nil provider map now omits its key instead of emitting `{}` — within the
  OpenAPI optional-object contract; not exercised by any test or production provider. Documented in evidence.
- `runtime` stays in OpenAPI `required` (pre-existing); F8.2 adds no fields and does not touch it.
- Rollback = `git checkout` the 4 files. FE regen is deterministic from the yaml.
