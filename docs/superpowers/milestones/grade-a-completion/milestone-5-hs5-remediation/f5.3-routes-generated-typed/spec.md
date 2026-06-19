# Feature F5.3 — routes_generated.go typed responses (H-D sites)

> **Milestone:** 5 — HS-5 remediation  ·  **Feature:** `f5.3-routes-generated-typed`
> **Status:** Approved 2026-06-16 — code change may begin.

## Consumer contract

**Consumer:** the two HTTP operations in `internal/modules/templates/delivery/http/routes_generated.go`
that currently emit untyped `map[string]any` payloads:

1. `PresignTemplateDocxUploadUrl` and `PresignTemplateSchemaUploadUrl` (both via the shared
   `presignTemplateUpload` helper at line 128) — public 200 body for the two presign endpoints.
2. `PublishTemplateVersion` (line 238) — public 200 body for `POST /templates/{id}/versions/{n}/publish`.

**What they need:** to emit the spec-declared 200 body for each operation, sourced from the
generated response type in `templatesapi` (`api.gen.go`) — the single source of truth derived from
`api/openapi/v1/openapi.yaml`. They currently emit hand-rolled `map[string]any{…}` literals — an
H-D finding because the wire shape is not pinned to the generated contract and an undeclared
field can leak (and does, on publish).

**Required shape after this feature:**

| Op | New response value | Source |
|----|-------------------|--------|
| Presign Docx / Schema (helper) | the helper returns `(url, storageKey)`; each caller writes the **op-specific** generated 200 type (`templatesapi.PresignTemplateDocxUploadUrl200JSONResponse` / `PresignTemplateSchemaUploadUrl200JSONResponse`) with `{Url: &url, StorageKey: &key}`. | `api.gen.go:2735,3447` (identical shape: `Url *string`, `StorageKey *string`). |
| Publish | `templatesapi.PublishTemplateVersion200JSONResponse{PublishedVersionId: …, NextDraftId: …, NextDraftVersionNum: …}` — **three fields, exactly the OpenAPI declaration**. | `api.gen.go:3054-3058`; `openapi.yaml:1325-1335`. |

The publish swap **drops `published_version_number`** from the wire — an undeclared field per the
OpenAPI spec at `openapi.yaml:1331` (`required: [published_version_id, next_draft_id,
next_draft_version_num]`, no other properties). This is exactly the M1/F1.3 declared-fields-only
contract closing this leak; no consumer reads it (FE generated types `index.d.ts:2640` reference is
in `TemplateDTO`, unrelated). The presign swap is wire-identical (same field names, JSON value
identical) — encoding now goes through the generated type so a future schema change in
`openapi.yaml` propagates by re-generation.

## Interview record (B1.5 — resolved by investigation, no operator questions needed)

| Question | Resolution | Source |
|----------|-----------|--------|
| Which generated types match these endpoints? | `PresignTemplate{Docx,Schema}UploadUrl200JSONResponse{Url, StorageKey *string}`; `PublishTemplateVersion200JSONResponse{PublishedVersionId, NextDraftId string; NextDraftVersionNum int}` | grep of `api.gen.go` |
| Does the presign helper serve both operations? | Yes — `presignTemplateUpload` at line 105 is called by both `PresignTemplateDocxUploadUrl` (96) and `PresignTemplateSchemaUploadUrl` (101). Each owns a distinct generated type (with the same wire shape). | `routes_generated.go:96-103` |
| Strategy for the dual-use helper? | Refactor helper to return `(url, storageKey string, err error)`; each caller writes its own op-specific typed response. Smaller delta than inlining; pins each handler to its own contract. | design choice |
| Does dropping `published_version_number` break FE? | No — spec at `openapi.yaml:1331` declares only 3 fields; `index.d.ts:2640` hit is `TemplateDTO`, not the publish 200. | FE grep; OpenAPI |
| Pointer vs value for presign fields? | Generated type is `*string` (`omitempty`). Always-set values, so use `&local` — no nil emit. | `api.gen.go:2735-2738` |
| Tests asserting `published_version_number` in publish 200? | None — only the handler itself in `routes_generated.go`. Tests check `published_version_id` / `next_draft_*`. | `_test.go` grep below |

## Non-goals

- No OpenAPI/spec change (no `published_version_number` added, no field renames).
- No change to the Presign service inputs/outputs or the Publish service call (`h.svc.*`).
- No new HTTP route, no behavior change other than the one undeclared field removed on publish.
- No frontend change.

## Validation Gate

1. **H-D grep → 0 in this file:**
   `grep -n "map\[string\]any" internal/modules/templates/delivery/http/routes_generated.go`
   → 0 matches (or only in comments).
2. **Typed-response shape pin:** a test exercising `PresignTemplateDocxUploadUrl` and
   `PublishTemplateVersion` asserts the **exact JSON keys** of the 200 body match the spec
   (presign: `{url, storage_key}`; publish: `{published_version_id, next_draft_id,
   next_draft_version_num}` — and **not** `published_version_number`).
3. `go build ./...` clean.
4. `go test -count=1 ./internal/modules/templates/...` green (including any existing route tests
   asserting publish/presign response keys — verified to not depend on the dropped field).
