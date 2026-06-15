# Feature F2.3 — Evidence

> **Milestone:** 2  ·  **Feature:** `f2.3-templates-envelope`  ·  **Closed:** 2026-06-14
> **Contract:** `spec.md` (scope = contract-only; align contract to live emit; envelope direction from FE consumer).

## What was implemented

Closed the bare-array-vs-envelope drift on `GET /templates`: the handler emits an envelope
`{ data: { templates: [...] }, meta: { limit, offset } }` and the FE consumes that envelope, but the
OpenAPI 200 was declared as a bare `[TemplateDTO]`. Producer/consumer already agreed on the envelope;
the contract is now corrected to match:

- **`api/openapi/v1/openapi.yaml`** — (a) added schema `ListTemplatesResponse` = `object{ data: object{ templates: [TemplateDTO] (required) }, meta: object{ limit:int, offset:int (required) } }`, top-level `required: [data, meta]`; (b) `GET /templates` 200 schema changed from `{ type: array, items: TemplateDTO }` → `$ref: ListTemplatesResponse`. `TemplateDTO` reused unchanged (item shape already correct — `toTemplateResponse` emits the same 14 keys, verified).
- **`internal/modules/templates/api/api.gen.go`** — regenerated (`go generate`): new `ListTemplatesResponse` struct (`Data struct{ Templates []TemplateDTO }`, `Meta struct{ Limit int; Offset int }`); `ListTemplates200JSONResponse` retyped from `[]TemplateDTO` → `ListTemplatesResponse`. (Base64 churn = embedded-spec blob re-gzip, expected.)
- **No handler change.** `listTemplates` (`routes_query.go:67-75`) already emits the envelope via raw `writeJSON`; the strict-server binding `routes_generated.go:15-17` delegates to it and does not use the strict typed response, so the server compiles and the runtime emit is unchanged.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — red then green | `python -c "...['/templates']['get']['responses']['200']...schema is \$ref ListTemplatesResponse; data.templates->TemplateDTO; meta requires limit,offset"` | RED: `200 is not the envelope ref (got 'array')` → GREEN: `200=ListTemplatesResponse; data.templates->TemplateDTO; meta requires limit,offset` | real |
| Contract gate | `go run ./scripts/api-lint/ -strict api/openapi/v1/openapi.yaml .` | **0** violations (envelope keys snake-safe; op already `x-pagination-exempt`) | real |
| Static build | `go build ./...` | exit 0 | — |
| Static vet | `go vet ./internal/modules/templates/...` | exit 0 | — |
| Regen correctness | `go generate ./internal/modules/templates/api/...` + read | `ListTemplatesResponse{ Data struct{ Templates []TemplateDTO }; Meta struct{ Limit int; Offset int } }`; `ListTemplates200JSONResponse` = `ListTemplatesResponse`; only `openapi.yaml`+`templates/api/api.gen.go` changed | real |
| Targeted regression | `go test ./internal/modules/templates/... -p 2` | all `ok` (6 pkgs; delivery/http 2.726s) — `routes_query_test`/`routes_contract_test` still green | real |
| Runtime emit proof (Docker/API up `:8081`) | cookie login, `GET /api/v1/templates?limit=5` | `200`; top-level keys `['data','meta']`; `data` keys `['templates']`; `meta` = `{'limit':5,'offset':0}`; item[0] = the 14 `TemplateDTO` keys | real |
| Focused-slice precheck (H-D) | `listTemplates` top-level emit keys + `toTemplateResponse` keys vs `ListTemplatesResponse`/`TemplateDTO` | top-level `{data, meta}` both declared; `data.templates` declared; item keys = `TemplateDTO` (14, exact) → **0 emitted-but-undeclared fields** on `/templates` | real |

**Field-truth table (tri-source) — all agree on the envelope:** runtime `{data:{templates:[14-key item]}, meta:{limit,offset}}` ↔ spec `ListTemplatesResponse{data.templates:[TemplateDTO], meta:{limit,offset}}` ↔ gen Go `ListTemplatesResponse{Data.Templates []TemplateDTO, Meta.{Limit,Offset int}}` ↔ FE consumer `body.data.templates` + `body.meta.{limit,offset}` (`features/templates/api/templates.ts:131,141-142`) — *FE generated type gains `ListTemplatesResponse` at the milestone-batched `gen:api` (HS-2-gated); the FE hand-parse already matches the shape.*

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| OpenAPI `/templates` 200 is the envelope (`$ref: ListTemplatesResponse`) requiring `data.templates:[TemplateDTO]` + `meta.{limit,offset}`; YAML parses | yes | TDD green row |
| Generated server type is the envelope (not `[]TemplateDTO`); build clean | yes | regen + static build rows |
| `api-lint -strict` 0 violations | yes | contract-gate row |
| Field-truth table agrees across runtime ↔ spec ↔ codegen ↔ FE | yes (FE gen deferred to milestone regen; hand-parse already matches) | field-truth table |
| FE generated `ListTemplatesResponse` appears; `tsc` 0 | **deferred** — milestone-level, single `gen:api` after F2.1–F2.3, HS-2-gated | per spec (milestone gate, not feature) |
| `/templates` emits no field outside the declared envelope | yes | focused-slice precheck row |

## Review disposition

- **Spec-compliance + code-quality review** (independent `caveman:cavecrew-reviewer` subagent, read-only, over the 2-file diff): **NO FINDINGS.** Verified the new schema matches the runtime emit exactly (`data`/`meta`/`templates`/`limit`/`offset`, all snake_case), required lists correct (`[data,meta]`, `[templates]`, `[limit,offset]`), generated struct `Data.Templates []TemplateDTO` + `Meta.{Limit,Offset int}`, `ListTemplates200JSONResponse` retyped to the envelope, **no handler/FE/behavior change** (handler still delegates to the raw `listTemplates` emit). No scope creep.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| FE `gen:api` + `tsc 0`, and FE hand-parse removal (`features/templates/api/templates.ts` `body.data.templates`/`body.meta.*` → generated `ListTemplatesResponse`) | Milestone-level single regen (F2.1–F2.3 batched); HS-2 (FE eigenpal `file:` path) must clear first; after regen the generated type describes the real shape | Milestone M2 close gate; trigger = HS-2 resolved → run single `gen:api`, then switch FE off hand-parse, `tsc 0` |
| Undeclared query params (request-side drift): `GET /templates` declares `parameters: []` but runtime reads `limit`/`offset`/`doc_type`; runtime `limit` max **200** vs canon max **100** | Different audit class from M2's emit-undeclared-**response-field** (H-D); request-contract, not this milestone | A request-contract milestone, or M3 |
| `listTemplates`/`toTemplateResponse` raw-map → emit-from-generated-type | H-D is *declared-closed* here; hard-prevent re-drift = serialize from gen type (M3 raw-map pattern) | Next touch of `routes_query.go`/`routes_create.go` serialization, or M3 |
