# Feature F2.3 — Plan (the "how")

> Spec: `spec.md` (Approved 2026-06-14, scope = contract-only, envelope direction). Judge against the spec, not this file.

## Approach

Contract-first, align spec → live runtime emit. Add one `ListTemplatesResponse` envelope schema and
point `GET /templates` 200 at it (replacing the bare array). Reuse `TemplateDTO` for items (already
correct). No handler change (the strict-server binding `routes_generated.go:15-17` delegates to the raw
`h.listTemplates`, which writes the envelope via `writeJSON` — it does not use the strict typed response,
so the server compiles regardless of the 200 type). No FE re-wire, no param change. FE `gen:api` is the
milestone-batched single regen (after F2.1–F2.3, HS-2-gated) — **not** in this feature.

## Steps (each with its verify)

1. **TDD red.** Run the OpenAPI envelope assertion: `/templates` get 200 resolves to an object with
   required `data.templates` (items → `TemplateDTO`) + `meta.{limit,offset}`.
   → verify: assertion **fails** (200 is currently `type: array`, items `TemplateDTO`). Record the red baseline.
2. **Edit `api/openapi/v1/openapi.yaml`:**
   a. Add schema under `components.schemas` (near `TemplateDTO`):
      ```yaml
      ListTemplatesResponse:
        type: object
        required: [data, meta]
        properties:
          data:
            type: object
            required: [templates]
            properties:
              templates:
                type: array
                items: { $ref: '#/components/schemas/TemplateDTO' }
          meta:
            type: object
            required: [limit, offset]
            properties:
              limit: { type: integer }
              offset: { type: integer }
      ```
   b. Change `paths./templates.get.responses.200.content.application/json.schema` from the bare
      `{ type: array, items: { $ref: TemplateDTO } }` to `{ $ref: '#/components/schemas/ListTemplatesResponse' }`.
   → verify: YAML parses; assertion now **passes**.
3. **Regen Go server type:** `go generate ./internal/modules/templates/api/...` (templates module codegen).
   → verify: `internal/modules/templates/api/*.gen.go` gains `ListTemplatesResponse` with `Data` + `Meta`
   (nested `Templates []TemplateDTO`, `Limit`/`Offset int`). `git diff` touches only the templates gen file
   (+ embedded-spec blob churn, expected). Confirm correct module: codegen config lives under the templates api pkg.
4. **Build + lint:** `go build ./...`; `go vet ./internal/modules/templates/...`; `go run ./scripts/api-lint/ -strict api/openapi/v1/openapi.yaml .`.
   → verify: build 0, vet 0, api-lint **0 violations** (keys snake-safe; op already `x-pagination-exempt`).
5. **Touched-pkg tests:** `go test ./internal/modules/templates/... -p 2`.
   → verify: ok (no behavior change; the existing `routes_contract_test`/`routes_query_test` confirm the
   handler still serves 200 with the envelope; regen didn't break compile/tests).
6. **Runtime emit proof (Docker up):** `.\scripts\start-api.ps1 -Build`, cookie login, `GET /api/v1/templates` authed.
   → verify: response body is the envelope `{ data: { templates: [...] }, meta: { limit, offset } }`. Record the actual JSON.
7. **Field-truth table** into `evidence.md`: runtime emit ↔ spec ↔ gen server type ↔ FE consumer all agree
   on the envelope.

## Out of this feature (per spec non-goals)

- FE single `gen:api` + `tsc 0`, and FE hand-parse removal (`api/templates.ts`) → milestone-batched, HS-2-gated.
- `listTemplates`/`toTemplateResponse` raw-map refactor → M3.
- Undeclared `limit`/`offset`/`doc_type` query params + `limit`-max-200-vs-100 → request-contract defer / M3.

## Rollback

Additive schema + a single `$ref` swap on one response. Revert = restore the bare-array 200 + drop
`ListTemplatesResponse` + re-`go generate`. No data, no migration; the FE hand-parses the envelope
already (independent of the generated type) → zero runtime blast radius beyond the regenerated type.
