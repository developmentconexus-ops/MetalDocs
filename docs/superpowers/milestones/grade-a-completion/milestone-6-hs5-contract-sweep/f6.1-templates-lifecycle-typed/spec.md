# Feature F6.1 — Templates lifecycle typed responses + OpenAPI 200 schemas

> **Milestone:** 6 — HS-5 Contract Sweep  ·  **Folder:** `f6.1-templates-lifecycle-typed`
> **Status:** Approved 2026-06-19 — code change may begin.
> **Approved before code:** 2026-06-19 / leandrotca.work (via M6 operator scope confirmation 2026-06-19; F6.5 decision confirmed close-in-scope).

## Interview record (B1.5)

| # | Question | Answer / source |
|---|----------|----------------|
| 1 | Wire output changing? | **No.** Existing handler tests (`routes_lifecycle_test.go:30,213,254,312,362,471`) already decode the `data.{version,template,approval_config,next_draft}` shape. F6.1 is structural retype + spec-declaration of the existing shape — byte-identical JSON. |
| 2 | Generated types already exist for these ops? | Partially: `ApproveTemplateVersionResponse` already declared (`api.gen.go:59`, used to be unconsumed by handler). Submit/review/archive/upsert lack 200-body schemas in `openapi.yaml:1485-1614` — `description: ok` with no `content`. F6.1 declares them + regens. |
| 3 | Reused schema for submit + review (same shape `{data.version: VersionDTO}`)? | Yes. Both responses are identical — `{data: {version: VersionDTO}}`. Define one shared schema `TemplateVersionEnvelope` in `components.schemas`, reference from both. |
| 4 | Approve response — already-declared shape correct? | Yes — `ApproveTemplateVersionResponse` (`api.gen.go:59`) matches the existing handler envelope exactly: `{data: {version: VersionDTO, next_draft?: {id, version_number}}}`. Handler just needs to construct the generated type instead of the map literal. |
| 5 | `archiveTemplate` shape? | `{data: {template: TemplateDTO}}` — declare new schema `ArchiveTemplateResponse`. |
| 6 | `upsertTemplateApprovalConfig` shape? | `{data: {approval_config: {template_id: uuid, reviewer_role?: string, approver_role: string}}}` — declare `UpsertTemplateApprovalConfigResponse` with a nested `TemplateApprovalConfig` schema. |
| 7 | Will codegen regen touch other modules? | No. `cfg.yaml` uses `include-tags: [templates]`; only `templates/api/api.gen.go` regenerates. |
| 8 | Vendor-mode regen flag? | `GOFLAGS=-mod=mod go generate ./internal/modules/templates/api/...` (per `wiki/references/oapi-codegen.md:18`). |
| 9 | Does this regress F5.1 (templates literal) or F5.4 (templates create/autosave typed)? | No. F5.1 is in `template_version_reader.go` (infrastructure), untouched. F5.4 is in `routes_create.go`, untouched. F6.1 only touches `routes_lifecycle.go` + OpenAPI schemas + codegen output. |
| 10 | HS-2 risk — does this imply codegen-first redesign across modules? | No. F6.1 is bounded to 5 declared schemas + 1 module regen — within the existing codegen posture for templates (already strict-server in `cfg.yaml`). |

## Consumer contract (FIRST — before any producer)

**Consumers:**
- FE TanStack-Query callers via the templates `api.gen.ts` codegen (regenerated under F6.6).
- Existing Go handler tests (`routes_lifecycle_test.go`) that decode the wire JSON envelope by struct.

**Contract — wire-identical to today, now declared in OpenAPI + emitted via generated types.**

| Op (`operationId`) | Path / method | 200 body schema (NEW or aligned) | Handler emits |
|-------------------|---------------|-----------------------------------|----------------|
| `submitTemplateVersion` | `POST /templates/{id}/versions/{n}/submit` | NEW `TemplateVersionEnvelope`: `{data: {version: VersionDTO}}` | `templatesapi.TemplateVersionEnvelope{Data: {Version: dto}}` |
| `reviewTemplateVersion` | `POST /templates/{id}/versions/{n}/review` | NEW `TemplateVersionEnvelope` (shared) | same |
| `approveTemplateVersion` | `POST /templates/{id}/versions/{n}/approve` | ALIGN — already-declared `ApproveTemplateVersionResponse` | `templatesapi.ApproveTemplateVersionResponse{Data: {...}}` |
| `archiveTemplate` | `POST /templates/{id}/archive` | NEW `ArchiveTemplateResponse`: `{data: {template: TemplateDTO}}` | `templatesapi.ArchiveTemplateResponse{Data: {Template: dto}}` |
| `upsertTemplateApprovalConfig` | `PUT /templates/{id}/approval-config` | NEW `UpsertTemplateApprovalConfigResponse`: `{data: {approval_config: TemplateApprovalConfig}}`, where `TemplateApprovalConfig: {template_id: uuid, reviewer_role?: string, approver_role: string}` | `templatesapi.UpsertTemplateApprovalConfigResponse{Data: {...}}` |

Status code = **200** for all five (matches today). Error responses (400/401/403/404/409/500) unchanged.

**Source of truth for the contract:** the existing handler emit at `routes_lifecycle.go:46,100,164,196,239` and the existing handler-test decoders (`routes_lifecycle_test.go:30,213,254,312,362,471`).

## What this feature implements

1. **OpenAPI schema declarations** (`api/openapi/v1/openapi.yaml`):
   - Add `components.schemas.TemplateVersionEnvelope`, `ArchiveTemplateResponse`, `UpsertTemplateApprovalConfigResponse`, `TemplateApprovalConfig`.
   - Add `200.content.application/json.schema: $ref` to submit / review / archive / upsertApprovalConfig responses.
   - Align approve response (already has `$ref: '#/components/schemas/ApproveTemplateVersionResponse'` — no change needed; verify only).
2. **BE codegen regen:** `GOFLAGS=-mod=mod go generate ./internal/modules/templates/api/...` → commit updated `api.gen.go`.
3. **Handler retype** (`internal/modules/templates/delivery/http/routes_lifecycle.go`):
   - `submitForReview` (:46): build and writeJSON `templatesapi.TemplateVersionEnvelope`.
   - `review` (:100): same.
   - `approve` (:164): build `templatesapi.ApproveTemplateVersionResponse`, including the `NextDraft` pointer (nil on reject path, populated on publish path — matches existing logic at :158–162).
   - `archiveTemplate` (:196): `templatesapi.ArchiveTemplateResponse`.
   - `upsertApprovalConfig` (:239): `templatesapi.UpsertTemplateApprovalConfigResponse`.
4. **Type conversions** to handle where generated nested types differ from raw Go fields:
   - `cfg.TemplateID` (Go string) → generated `openapi_types.UUID` if generated as UUID. (Confirm at regen; same pattern as F5.4 `toAPITemplateDTO`.)
   - `cfg.ReviewerRole` (`*string`) → generated `*string` (1:1).

## Non-goals (mandatory)

- No change to wire JSON keys/values (byte-identical to today).
- No change to status codes (200 for all five; error envelopes untouched).
- No change to authz logic, capability strings, or service-layer commands.
- No FE codegen regen in F6.1 (that's F6.6).
- No touching `routes_query.go` (F6.2), `routes_create.go` (F5.4, done), `routes_autosave.go`, `routes_catalog.go` (F6.4), `routes_schema.go` (F6.4).
- No retype of error response paths (`writeErr` / `writeMappedErr`) — already typed via `httpresponse`.
- No new product capabilities.
- No structural refactor of the 5 handlers beyond the `writeJSON` payload swap.

## Validation Gate (concrete — approved before code)

| # | Acceptance criterion | Named test / proof command | Real vs fixture |
|---|----------------------|----------------------------|------------------|
| 1 | All 5 handlers serialize the generated `*Response`/envelope type — zero `map[string]any` literal remaining in `routes_lifecycle.go`. | `grep -nE 'map\[string\]any' internal/modules/templates/delivery/http/routes_lifecycle.go` → 0 hits | real (grep) |
| 2 | OpenAPI declares 200 body schemas for all 5 ops. | `grep -nA2 'operationId: (submitTemplateVersion\|reviewTemplateVersion\|approveTemplateVersion\|archiveTemplate\|upsertTemplateApprovalConfig)' -A30 api/openapi/v1/openapi.yaml` shows `schema: $ref` for each | real (grep) |
| 3 | BE codegen is fresh — no uncommitted diff after regen. | `GOFLAGS=-mod=mod go generate ./internal/modules/templates/api/...` then `git diff --exit-code internal/modules/templates/api/api.gen.go` → exit 0 | real |
| 4 | Build + existing tests green from clean. | `go build ./...` → clean; `go test -count=1 ./internal/modules/templates/...` → 0 FAIL; whole-repo `go test -count=1 ./...` → 0 FAIL | real |
| 5 | Wire-shape parity — `TestSubmitForReview_Happy`, `TestReview_Accept_Happy`, `TestApprove_Accept_Happy` (next_draft populated), `TestApprove_Reject_NextDraftNull`, `TestArchiveTemplate_Happy`, `TestUpsertApprovalConfig_Happy` all pass without modification of their decode structs. | `go test -count=1 -run 'TestSubmitForReview\|TestReview_Accept\|TestApprove_\|TestArchiveTemplate_\|TestUpsertApprovalConfig_' ./internal/modules/templates/delivery/http/...` → green | real |
| 6 | NEW typed-shape contract test — decodes the lifecycle 200 bodies with `json.Decoder + DisallowUnknownFields` into the generated `*Response` types and asserts non-zero typed fields round-trip. | `go test -count=1 -run 'TestLifecycle_TypedResponseShape' ./internal/modules/templates/delivery/http/...` → green (TDD: failing first, then implement) | real |

## ADR needed?

- [x] No durable decision — F6.1 follows the contract-first ADR 0012 posture already in force for templates; no new design.
