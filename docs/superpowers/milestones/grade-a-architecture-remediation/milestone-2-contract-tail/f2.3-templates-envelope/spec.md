# Feature F2.3 — Spec

> **Milestone:** 2 — Contract Tail (H-D class)  ·  **Folder:** `f2.3-templates-envelope`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-14 / leandrotca.work (operator) — scope "Contract-only (kill H-D)", same cadence as F2.1/F2.2; direction (envelope) decided contract-first from the FE consumer.

> This is the feature's **contract**, written and approved **before any code**. It defines what the
> feature must do and how it will be proven — not how it will be built (that is `plan.md`). The
> milestone-validator judges the feature against *this* file (C1).

## Interview record (fail-closed gate)

| # | Question (resolved by investigation, not operator round-trip) | Resolution |
|---|----------|------------|
| 1 | `GET /templates` 200 is declared as a **bare array** `[TemplateDTO]`, but the handler emits an **envelope** `{ data: { templates: [...] }, meta: { limit, offset } }` and the FE consumes that envelope (`api/templates.ts:131` `body.data.templates`, `:141-142` `body.meta.limit/offset`). Which is the contract — bare array or envelope? | **Envelope.** Consumer-contract-first: the real FE consumer reads `data.templates` + `meta.{limit,offset}`; the runtime emits exactly that. The generated bare-array type is **unused** (FE hand-parses *because* the contract was wrong). Align the contract to the envelope. |
| 2 | Is the runtime envelope itself the defect against the design-system pagination canon (cursor-first, `page.{next_cursor,has_more}`, limit max 100, `PAGINATION-DRIFT` BLOCKING)? If so the fix would be a runtime+FE redesign (HS-2), not a contract edit. | **No — not a redesign.** `GET /templates` already carries `x-pagination-exempt: true` with a **permanent bounded-list** reason ("full per-tenant template catalog… a small bounded configuration set returned in one shot; a permanent bounded-list exemption, **not** a deferred cursor migration"). The endpoint is intentionally a bounded offset list; the `{data,meta}` offset envelope is its legitimate **permanent** shape. No cursor migration is owed. The drift is purely that the 200 schema was never updated from bare-array to the real envelope. Contract-only, in-scope. |

**Why no operator round-trip:** both forks resolved by reading the spec + FE + design-system canon. Direction is determined (envelope, from the consumer); the only canon concern (non-cursor) is already a sanctioned permanent exemption on this op. No remaining ambiguity → fail-closed gate clears without an interview, mirroring F2.1/F2.2.

## Consumer contract (FIRST — before any producer)

- **Consumer(s):** the FE templates-list path. **A real runtime field-consumer exists** and hand-parses the envelope today: `frontend/apps/web/src/features/templates/api/templates.ts:131` reads `body?.data?.templates`; `:141-142` read `body?.meta?.limit` / `body?.meta?.offset`. The hand-parse exists *because* the generated FE type (a bare `TemplateDTO[]`) does not describe the real response — i.e. the FE is already compensating for this exact contract drift. Declaring the envelope makes the generated type honest.
- **Contract — the exact shape the runtime already emits** (read from the producer's live output, `internal/modules/templates/delivery/http/routes_query.go:67-75`, `listTemplates`):
  ```json
  {
    "data": { "templates": [ TemplateDTO, ... ] },
    "meta": { "limit": <int>, "offset": <int> }
  }
  ```
  - `data` (object, required) → `templates` (array of `TemplateDTO`, required).
  - `meta` (object, required) → `limit` (integer, required), `offset` (integer, required).
  - **Item shape is unchanged and already correct:** `toTemplateResponse` (`routes_create.go:43-64`) emits exactly the 14 keys `TemplateDTO` declares (`id, tenant_id, doc_type_code, key, name, description, latest_version, latest_revision_number, published_version_id, published_version_number, current_revision_number, created_by, created_at, archived_at`) — verified key-for-key. **No nested item drift**; F2.3 reuses `TemplateDTO` for the items and touches only the top-level envelope.
- **Source of truth for the contract:** the live handler emit (`routes_query.go:67-75`) + the existing `TemplateDTO` schema (already correct). The OpenAPI `GET /templates` 200 response (currently bare array, `openapi.yaml`) is the surface being corrected to match that emit.

## What this feature implements

Add a `ListTemplatesResponse` schema to `api/openapi/v1/openapi.yaml` describing the envelope the
`/templates` handler already emits — `{ data: { templates: [TemplateDTO] }, meta: { limit, offset } }`,
with `data`, `data.templates`, `meta`, `meta.limit`, `meta.offset` all required — and change the
`GET /templates` 200 response from the bare `array`/`TemplateDTO` to `$ref: ListTemplatesResponse`.
Regenerate the server types (`oapi-codegen`); FE types are batched into the milestone's single
`gen:api`. After this feature, the `/templates` response shape is tri-source-declared (runtime ↔ spec ↔
generated server type ↔ generated FE type) and the bare-array drift is closed.

## Non-goals (mandatory)

- **No FE re-wire / hand-parse removal in this feature.** `features/templates/api/templates.ts`'s manual `body.data.templates` / `body.meta.*` parsing stays as-is. After the milestone's single `gen:api`, the generated `ListTemplatesResponse` type will describe the real shape, enabling a later switch off hand-parsing — **milestone-batched / future FE cleanup**, not feature scope. (Same discipline as F2.1/F2.2.) Recorded as a defer.
- **No handler-serialization refactor.** `listTemplates` keeps its `map[string]any` envelope and `toTemplateResponse` raw map. Emitting from the generated type is the raw-map→generated-type pattern owned by **M3**; doing it here is scope drift. *(Re-drift risk recorded as a defer, not closed here.)*
- **No request-side param fix.** `GET /templates` declares **no** query parameters in OpenAPI (`parameters: []`) yet the runtime reads `limit`/`offset`/`doc_type`. That undeclared-**request-param** drift is a *different* audit class from M2's emit-undeclared-**response-field** (H-D) objective. **Out of F2.3 scope**; recorded as a bounded defer.
- **No pagination-model change.** `/templates` stays a permanent bounded offset list (`x-pagination-exempt: true`, existing reason). No cursor migration, no `limit`-max change (the runtime's max-200 vs canon's max-100 is a pre-existing request-side question, tracked with the param defer — not this feature).
- **No other route or schema touched.** Only the `GET /templates` 200 response + one new `ListTemplatesResponse` schema. `TemplateDTO` is reused unchanged. The FE `gen:api` runs **once** for F2.1–F2.3 (not in this feature).

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| OpenAPI `GET /templates` 200 is the envelope (`$ref: ListTemplatesResponse`); `ListTemplatesResponse` requires `data.templates: [TemplateDTO]` + `meta.{limit,offset}`; YAML parses | `python -c "...; g=s['paths']['/templates']['get']['responses']['200']...; assert resolves to object with data+meta; assert data.templates.items ref == TemplateDTO; assert meta requires limit,offset"` → exit 0 | real |
| Generated server type for the 200 response is the envelope (not `[]TemplateDTO`); build clean | `oapi-codegen` regen (canonical command) → `go build ./...` exit 0; grep generated `ListTemplatesResponse` struct has `Data` + `Meta` | real |
| `api-lint -strict` reports **0** violations (envelope keys snake-safe; op already pagination-exempt) | `go run ./scripts/api-lint/ -strict api/openapi/v1/openapi.yaml .` → 0 violations | real |
| Field-truth table: runtime emit ↔ spec ↔ generated server type ↔ FE consumer all agree on `{data:{templates},meta:{limit,offset}}` | route/field-truth table in `evidence.md`; runtime `GET /api/v1/templates` (Docker up, authed) returns the envelope with `data.templates[]` + `meta.{limit,offset}` | real |
| FE generated `ListTemplatesResponse` type appears; FE `tsc` 0 after the **milestone** single regen | `pnpm --filter web gen:api` (milestone-batched) + `pnpm --filter web tsc` → 0 | real (milestone-level) |
| Focused-slice precheck: `/templates` emits no field outside the declared envelope (top-level keys `data`,`meta` only; item keys = `TemplateDTO`) | diff `listTemplates` emitted top-level keys + `toTemplateResponse` keys vs `ListTemplatesResponse`/`TemplateDTO` → only the envelope wrapper was the gap, now closed | real |

> TDD note: the OpenAPI envelope assertion is the failing-first check (fails before the schema edit —
> 200 is a bare array — passes after). Runtime emit proof is observed live (Docker up, authed). FE `tsc`
> 0 is a **milestone-level** gate (single `gen:api` after F2.1–F2.3).

## ADR needed?

- [x] No durable decision — skip. This aligns the contract to an existing runtime emit + sanctioned permanent-exempt pagination posture; no architectural choice. (The deferred handler-serialization-refactor, FE hand-parse removal, and undeclared-query-param drift are tracked below, not ADR-grade.)

## Deferred observations (recorded, not closed here — none block F2.3)

- **FE hand-parse removal:** after the milestone single `gen:api`, `features/templates/api/templates.ts`'s manual `body.data.templates` / `body.meta.*` parsing can switch to the generated `ListTemplatesResponse` type. Trigger: milestone M2 FE regen step (with `tsc 0`), or a later FE cleanup.
- **Undeclared query params (request-side drift):** `GET /templates` declares `parameters: []` but the runtime reads `limit`, `offset`, `doc_type` (`routes_query.go:28-49`). Different audit class from M2's response-field H-D. Also note the runtime `limit` max is **200** vs design-system canon **100** (`api-design-system.md` §4). Trigger: a request-contract milestone, or M3.
- **Re-drift risk:** `listTemplates` hand-builds the `map[string]any` envelope and `toTemplateResponse` the item map; declaring the shapes does not force serialization from the generated type, so a future field could drift again. Root-cause-hard fix = emit from generated type (M3 raw-map→generated-type pattern). Trigger: next touch of `routes_query.go`/`routes_create.go` serialization, or M3.
