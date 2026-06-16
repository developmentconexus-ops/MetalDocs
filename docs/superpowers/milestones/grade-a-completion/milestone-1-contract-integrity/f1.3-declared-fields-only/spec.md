# Feature F1.3 — createTemplate declared-fields-only — Spec

> **Milestone:** 1 (Contract / API integrity) · **Program:** grade-a-completion
> **Folder:** `f1.3-declared-fields-only`
> **Closes:** A3 (H-D class) — `createTemplate` leaks undeclared top-level `id` / `version_id` and
> emits `map[string]any` instead of typed `TemplateDTO` / `VersionDTO`.

## Interview record (contract read from source)

| Q | A | Source |
|---|---|--------|
| What does the handler currently emit? | `map[string]any{"id": res.Template.ID, "version_id": res.Version.ID, "data": map[string]any{"template": toTemplateResponse(res.Template), "version": toVersionResponse(res.Version)}}` — two undeclared top-level keys + legacy untyped mappers | `routes_generated.go:64-71` |
| What does the OpenAPI schema declare? | `CreateTemplateResponse` → `{data: {template: TemplateDTO, version: VersionDTO}}` — no `id`, no `version_id` at top level | `api/openapi/v1/openapi.yaml:5060-5081`; `api.gen.go:75-81` |
| Is the `data` envelope declared? | **Yes.** The `{data: {template, version}}` wrapper is the declared shape. Only the two top-level extra keys are undeclared. | `CreateTemplateResponse` schema |
| Does a typed `VersionDTO` mapper exist? | **Yes** — `toAPIVersionDTO` in `routes_mapping.go:18` (added F1.2 / ADR 0035). | `routes_mapping.go:18` |
| Does a typed `TemplateDTO` mapper exist? | **No** — needs to be added. Pattern: mirror `toAPIVersionDTO`; add to `routes_mapping.go`. | — |
| Can `toVersionResponse` be retired? | **No** — 5 callers in lifecycle/query/schema routes (`routes_lifecycle.go:43,92,140`, `routes_query.go:132`, `routes_schema.go:65`). Those are F1.4 territory. | grep above |
| Can `toTemplateResponse` be retired? | **No** — 3 callers in query/lifecycle routes (`routes_query.go:64,131`, `routes_lifecycle.go:178`). F1.4 territory. | grep above |
| Approach pre-decided? | Yes — milestone.md F1.3 row: "drop undeclared `id`/`version_id`"; operator approved design 2026-06-15. No hard-stop required before implementing. | operator approval 2026-06-15 |

## Invariant under change (contract)

`createTemplate` MUST return HTTP 201 + JSON body that is a **strict subset** of `CreateTemplateResponse`:
```json
{"data": {"template": <TemplateDTO>, "version": <VersionDTO>}}
```
No `id`, no `version_id`, no `data.data` nesting. The values in `data.template` and `data.version`
are byte-for-byte equivalent to what `TemplateDTO`/`VersionDTO` serialises — same as
`toTemplateResponse`/`toVersionResponse` for the fields they share, plus the typed UTC timestamps
and UUID-validated `id` fields the typed structs enforce.

## What this implements

1. **Add** `toAPITemplateDTO(t *domain.Template) (templatesapi.TemplateDTO, error)` to
   `routes_mapping.go`, following the `toAPIVersionDTO` pattern (same file). Maps every `TemplateDTO`
   field from `domain.Template`; uses `uuid.Parse` for `Id`/`TenantId`; preserves optional
   pointer fields (`ArchivedAt`, `Description`, `DocTypeCode`, `PublishedVersionId`,
   `PublishedVersionNumber`, `CurrentRevisionNumber`).

2. **Hoist** both mapper calls in `CreateTemplate` (`routes_generated.go`) **before** `writeJSON`:
   call `toAPITemplateDTO(res.Template)` and `toAPIVersionDTO(res.Version)`; on error write 500;
   otherwise assemble `templatesapi.CreateTemplateResponse{Data: {Template: tplDTO, Version: vDTO}}`
   and `writeJSON(w, http.StatusCreated, resp)`. The undeclared `id`/`version_id` keys disappear
   automatically (they were only in the old `map[string]any` literal).

3. **TDD test** `TestCreateTemplate_TypedResponseShape` in `routes_typed_response_test.go`:
   - Decodes the 201 body into a strict `templatesapi.CreateTemplateResponse` (no extra fields)
   - Asserts `id` / `version_id` absent at top level (raw-decode check)
   - Asserts `data.template.id` and `data.version.version_number` present (smoke fields)
   - RED before hoist (old body fails the strict-no-extra-keys check); GREEN after.

4. **Update** `TestCreateTemplate_Happy` (`routes_create_test.go`) to add assertion that top-level
   `id` and `version_id` are absent (the existing test does not assert their absence).

## Non-goals (HS-6 scope guard)

- No retirement of `toVersionResponse` or `toTemplateResponse` (each has callers in F1.4 scope).
- No OpenAPI schema change — `CreateTemplateResponse` already declares the correct shape.
- No other `createTemplate` behaviour change (status 201, authz, validation, service call — all untouched).
- No touch to other routes/handlers.

## Validation Gate

| # | Criterion | Named proof | Real vs fixture |
|---|-----------|-------------|-----------------|
| V1 | `id` and `version_id` absent at top level of 201 response | `TestCreateTemplate_TypedResponseShape` RED before hoist, PASS after | real (logic) |
| V2 | `data.template` conforms to `TemplateDTO` shape; `data.version` conforms to `VersionDTO` shape | strict-decode in `TestCreateTemplate_TypedResponseShape` (no `json.Decoder.DisallowUnknownFields` needed — typed struct decode is sufficient) | real (logic) |
| V3 | `TestCreateTemplate_Happy` still passes; top-level absence assertion added | `go test ./internal/modules/templates/delivery/http/... -run TestCreateTemplate` | real |
| V4 | H-D grep clean on the site: `grep -n '"id"\|"version_id"' routes_generated.go` inside the `writeJSON` call → 0 matches | structural | real |
| V5 | Build + vet + full suite green | `go build ./...` exit 0; `go vet ./internal/modules/templates/...`; `go test ./...` 0 FAIL | real |

## ADR needed?

No. `toAPITemplateDTO` follows the identical pattern established by `toAPIVersionDTO` (ADR 0035 adoption
ledger). No new durable decision. Wiki stamp bumped at M1 close by `wiki-curator`.

## Approval

Design presented to operator 2026-06-15; approved ("Proceed"). Consumer contract read from
`api.gen.go:75-81` and `routes_generated.go:64-71`. No implementation began before this spec.
