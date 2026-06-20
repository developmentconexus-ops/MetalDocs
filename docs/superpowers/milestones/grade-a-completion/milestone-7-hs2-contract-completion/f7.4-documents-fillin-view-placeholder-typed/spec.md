# Feature F7.4 — Documents typed 200 bodies (codegen-first + 1 off-spec hand-rolled)

> **Milestone:** 7 — HS-2 Contract Completion  ·  **Folder:** `f7.4-documents-fillin-view-placeholder-typed`
> **Status:** Approved 2026-06-20 — code change may begin.
> **Approved before code:** 2026-06-20 / leandrotca.work — inherited from the M7 Phase-2 operator
> approval (commit `45a03fa6`) + the HS-6 scope extension (operator "convert it", 2026-06-20) folding
> `pdf_webhook_handler.go:113` into F7.4.

## Interview record (B1.5)

| # | Question | Answer / source |
|---|----------|----------------|
| 1 | Is documents codegen or pre-codegen? | **Codegen** — `internal/modules/documents/api/{cfg.yaml,gen.go,api.gen.go}`; `include-tags: [documents]`; `models + std-http-server + strict-server + embedded-spec`. Handlers use the generated **models** (e.g. `documentsapi.DocumentListResponse`) over a hand-rolled mux — **not** the strict `ServerInterface`. Adding 200 schemas adds models only; no routing rewire (HS-2 respected). |
| 2 | Which response-literal sites? | 5: `fillin_handler.go:58` (getDocumentFillInSchema), `:116` (putDocumentPlaceholderValue); `placeholder_options_handler.go:67,74` (getDocumentPlaceholderOptions, 2 branches); `view_handler.go:46-51` (viewDocument); **+ HS-6** `pdf_webhook_handler.go:113`. |
| 3 | Do the 4 codegen ops already exist in OpenAPI? | Yes — all 4 have operationIds + `tags: [documents]` but their `200` is `description: ok` with **no content schema** → no generated body type → handlers fall back to `map[string]any`. Fix = declare the 200 content schemas, regen. |
| 4 | fill-in-schema wire shape? | `{data: {placeholder_schema: [Placeholder...]}}`. `phs` is `[]templatesdomain.Placeholder` (17-field nested domain type incl `VisibilityCondition`, two `any` fields). |
| 5 | placeholder-options wire shape? | `{options: [...]}` — **polymorphic**: PHSelect branch emits `[]map[string]string` (`{value, display_name}`), PHUser branch emits `[]UserOptionView` (`{user_id, display_name}`). No single item type. |
| 6 | put wire shape? | `{placeholder_id: string, updated_at: <RFC3339 string>}`. |
| 7 | view wire shape? | `{pdf_status: string}` always; `signed_url` + `pdf_url` only when `PDFStatus=="ready" && SignedURL!=""`. |
| 8 | How to model the rich/polymorphic payloads without disproportionate codegen? | **Typed envelope, opaque items.** Declare the *envelope* precisely (generated named struct) but declare `placeholder_schema` items and `options` items as `items: {}` → oapi-codegen emits `[]interface{}`. The handler boxes the existing typed values (`templatesdomain.Placeholder`, `[]map[string]string`, `UserOptionView`) into the slice — each marshals via its **own** json tags, so wire is byte-identical with **zero** conversion structs and no drift. This kills the `map[string]any` **envelope** literal (the actual H-D defect) without faithfully re-modeling a 17-field nested domain type — proportionate to the operator's typed-body-parity appetite. The items were untyped on the wire before, so FE codegen does not regress. |
| 9 | put `updated_at` parity with generated `time.Time`? | Generated `UpdatedAt time.Time` marshals RFC3339**Nano**; the prior wire was `Format(time.RFC3339)` (second precision). Assign `time.Now().UTC().Truncate(time.Second)` so nanos=0 → marshals `…Z` seconds-only = byte-identical (same mitigation as F7.1 audit `OccurredAt`). |
| 10 | pdf_webhook (HS-6) — codegen or hand-rolled? | **Hand-rolled.** The route is **unwired** (`RegisterRoutes` uncalled, H-1e `:26`) and **deliberately off the OpenAPI spec** (Phase C wont-fix, `:120-121`). Contract-first codegen would contradict the documented wont-fix. Per operator "convert it": a hand-rolled typed struct (ADR 0012 posture), **no OpenAPI change**, wire-identical `{document_id, final_pdf_s3_key}`. |
| 11 | HS-2 risk — pipeline standup / routing rewire? | **No.** Existing pipeline; only new schemas + regen + emit generated models. No `ServerInterface`/`NewStrictHandler` wiring. |

## Consumer contract (FIRST — before any producer)

**Consumers:** FE documents callers + the OpenAPI documents schemas (source of truth for the 4 codegen ops).

| Op | Path / method | Status | Body type emitted | Wire keys (unchanged) |
|----|---------------|--------|-------------------|------------------------|
| getDocumentFillInSchema | `GET /documents/{id}/fill-in-schema` | 200 | `documentsapi.DocumentFillInSchemaResponse` (envelope; `placeholder_schema []interface{}`) | `{data:{placeholder_schema:[…]}}` |
| putDocumentPlaceholderValue | `PUT /documents/{id}/placeholders/{pid}` | 200 | `documentsapi.PutPlaceholderValueResponse` | `{placeholder_id, updated_at}` |
| getDocumentPlaceholderOptions | `GET /documents/{id}/placeholder-options/{pid}` | 200 | `documentsapi.PlaceholderOptionsResponse` (`options []interface{}`) | `{options:[…]}` |
| viewDocument | `GET /documents/{id}/view` | 200 | `documentsapi.ViewDocumentResponse` | `{pdf_status[, signed_url, pdf_url]}` |
| HandlePDFComplete (off-spec, unwired) | `POST /documents/{id}/pdf-complete` | 200 | hand-rolled `pdfCompleteResponse` | `{document_id, final_pdf_s3_key}` |

## What this feature implements

1. **Contract-first**: add 4 component schemas under `components/schemas` + `$ref` them from the 4 ops'
   `200.content.application/json.schema` (envelope precise; rich/polymorphic items as `items: {}`).
2. `GOFLAGS=-mod=mod go generate ./internal/modules/documents/api/...` → commit the regenerated
   `api.gen.go` (no uncommitted diff after regen).
3. Emit the generated models at the 4 sites; box existing typed values into the `[]interface{}` item
   slices (no `map[string]any` constructed in handler files).
4. **pdf_webhook (HS-6)**: hand-rolled `pdfCompleteResponse` struct + swap `:113`; no OpenAPI change.

## Non-goals (mandatory)

- No change to wire keys/values (byte-identical, incl. `updated_at` RFC3339 second-precision, the
  conditional `signed_url`/`pdf_url` presence, and empty `placeholder_schema`/`options` → `[]`).
- **No faithful re-modeling** of `Placeholder`/`VisibilityCondition`/`UserOptionView` into OpenAPI —
  items stay opaque (`items: {}`). Over-modeling them would be gold-plating beyond the H-D defect.
- No routing rewire through generated `ServerInterface`/`NewStrictHandler`; no new pipeline (HS-2).
- No OpenAPI entry for the off-spec pdf_webhook route (Phase C wont-fix preserved).
- No change to service/domain logic, authz, or error mapping.

## Validation Gate (concrete — approved before code)

| # | Acceptance criterion | Named test / proof command | Real vs fixture |
|---|----------------------|----------------------------|------------------|
| 1 | OpenAPI declares 200 JSON schemas for all 4 codegen ops | `grep` the 4 `$ref`s under the ops' 200 | real |
| 2 | BE codegen fresh — regen yields no uncommitted diff | `GOFLAGS=-mod=mod go generate ./internal/modules/documents/api/...` then `git diff --exit-code internal/modules/documents/api/api.gen.go` | real |
| 3 | Zero response-literal `map[string]any` (incl. via `writeFillInJSON`) in the 4 files | `grep -nE 'map\[string\]any' fillin_handler.go placeholder_options_handler.go view_handler.go pdf_webhook_handler.go` → 0 | real — **red→green: 5 response literals → 0** |
| 4 | Handlers emit the generated models / hand-rolled struct | build + the wire-parity tests below | real |
| 5 | Wire JSON unchanged | NEW per-route wire tests assert envelope keys + byte-parity of representative payloads (incl. `updated_at` second precision, empty → `[]`, conditional view keys, pdf_webhook shape) | real |
| 6 | Build + existing documents tests green | `go build ./...`; `go test -count=1 ./internal/modules/documents/...` → 0 FAIL | real |

## ADR needed?

- [x] No new durable decision — F7.4 follows contract-first (ADR 0012) for the 4 codegen ops and the
  ADR-0012 hand-rolled posture for the off-spec pdf_webhook route. The "typed envelope, opaque items"
  modeling choice is a proportionality call recorded here + in `evidence.md`, not an architecture change.
