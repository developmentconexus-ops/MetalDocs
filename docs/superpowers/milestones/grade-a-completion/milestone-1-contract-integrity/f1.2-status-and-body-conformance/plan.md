# Feature F1.2 — Status & Body Conformance — Plan

> **Milestone:** 1 — Contract / API integrity  ·  **Folder:** `f1.2-status-and-body-conformance`
> **Status:** Planning
> **Engine:** inline (`superpowers:writing-plans`-equivalent structure).

## Source

- **Milestone spec row (`milestone.md:32`):**
  - Implement: `renameDocument` → 200 no-content (A2, handler.go:519); `templates.createNextVersion`
    201 → 200 (A4, routes_create.go:36, H-D); `presignTemplateAutosave` 201 → 200
    (A5, routes_autosave.go:42, H-D).
  - Validate: handler tests assert status code + body shape equal the OpenAPI declaration for all
    three routes; H-D grep on the three sites returns 0.
- **Spec overrides (see `spec.md` interview record):**
  - **A4 keeps 201** (HS-6 deviation; canonical REST for genuine resource create) — mission §5 text
    "201→200" overridden.
  - **A4 / A5 OpenAPI amendments:** A4 declares `201` + flat `TemplateVersion` body; A5 declares
    `200` + flat `TemplatePresignAutosaveResponse` body. No `{data:{...}}` envelope on either.
  - **A2** = 200 empty body (no OpenAPI change).
- **Governing-spec reference:** mission §5 A2 / A4 / A5; `wiki/references/oapi-codegen.md`;
  CLAUDE.md §3 backend-API skill routing.

## Architectural scope decisions (consumer-contract-first)

1. **Authoring `TemplateVersion` schema component is in-scope** even though several non-A4 endpoints
   (`getTemplateVersion`, `commitTemplateAutosave`) also emit `map[string]any{data:{version:...}}`
   and would benefit from adopting it. Reasons:
   - The schema is a prerequisite for the A4 amendment; authoring it is irreducible.
   - It is **introduced** in F1.2; **adoption** by other handlers is **explicitly out of scope** —
     F1.4 (A6 class — `map[string]any` sweep) inherits the adoption. Non-adoption does not regress
     anything: those handlers continue to emit their current undeclared shape.
2. **`{data:{...}}` envelope is not touched on endpoints F1.2 does not own.** A4 and A5 drop it
   because they are the named sites; every other handler keeps it until F1.3/F1.4/F2.x picks it up.
   This avoids cross-feature scope drift (HS-6).
3. **No `domain.*` model changes.** Domain stays; only wire shape changes.
4. **No handler-test renames or unrelated rewrites.** Existing `TestPresignAutosave_Happy` is
   updated in place (status `201→200`, drop `data` indirection) — that test moves from documenting
   the bug to documenting the fix; no parallel new test for the happy path.

## Plan

### Phase A — TDD red (failing tests, no production code)

A1. **`internal/modules/documents/delivery/http/handler_rename_test.go`** (NEW)
   - Add `renameDocumentResult error` field on `fakeSvc` (or extend existing harness — re-use the
     pattern from `handler_checkpoints_test.go`).
   - `TestRenameDocument_TypedResponseShape`:
     - Happy path: PATCH `/api/v1/documents/{id}` with valid body → expect `rr.Code == 200`,
       `rr.Body.Len() == 0`, **no** `Content-Type: application/json` header (or assert absence of
       any body at all), and assert `Content-Length: 0` if Go sets it.
     - Also asserts `Content-Type` does **not** start with `application/json`.
   - **Expected red:** body contains JSON (current `httpresponse.WriteJSON(w, 200, doc)`); assertion
     on empty body fails.

A2. **`internal/modules/templates/delivery/http/routes_create_test.go`** (NEW)
   - `TestCreateNextVersion_TypedResponseShape`:
     - Build a `Handler` with a `fakeSvc` returning a deterministic `*domain.TemplateVersion`.
     - POST `/api/v1/templates/{id}/versions` → expect `rr.Code == 201`, JSON body is a single object
       (not a `data` wrapper), assert exact snake_case key set:
       `{id, template_id, version_number, revision_number, status, docx_storage_key, content_hash,
         metadata_schema, placeholder_schema, author_id, pending_reviewer_role,
         pending_approver_role, reviewer_id, approver_id, submitted_at, reviewed_at, approved_at,
         published_at, obsoleted_at, lock_version, created_at}`.
     - Assert `data` key is **absent** at the top level.
     - Reject any PascalCase key.
   - **Expected red:** current handler emits `{data:{version:{…}}}` — top-level `data` key present
     → assertion fails.

A3. **`internal/modules/templates/delivery/http/routes_autosave_test.go`** (UPDATE existing
    `TestPresignAutosave_Happy` at `:76`)
   - Change expected status: `201 → 200`.
   - Change decode struct: drop the outer `Data` indirection — decode the body directly into
     `{ UploadURL, StorageKey, ExpiresAt string }`.
   - Add an assertion that the response is a single flat object (top-level `data` key absent).
   - **Expected red:** current handler emits `201` + `{data:{...}}` → status assertion fails first.

A3a. **`internal/modules/templates/delivery/http/routes_autosave_test.go`** (UPDATE existing
     `TestCommitAutosave_Happy`)
   - Change decode struct: drop `Data.Version` indirection — decode flat `TemplateVersion`
     (same key set as V2).
   - Assert top-level `data` key absent; assert exact snake_case key set on the response.
   - **Expected red:** current handler emits `{data:{version:{...}}}` → flat decode misses fields.

A3b. **`internal/modules/templates/delivery/http/routes_query_test.go`** (NEW or extend existing
     file — first check whether the file exists; if not, scaffold using the autosave-test pattern)
   - `TestGetTemplateVersion_TypedResponseShape`:
     - Build a `Handler` with a `fakeSvc` returning a deterministic `*domain.TemplateVersion`.
     - GET `/api/v1/templates/{id}/versions/{n}` → expect `200`, flat body, exact key set,
       reject top-level `data`.
   - **Expected red:** current handler emits `{data:{version:{…}}}` → flat decode misses fields.

A4. **FE adapter tests** — locate or create:
   - Check `frontend/apps/web/src/features/templates/api/__tests__/` (or sibling). If a
     `templates.test.ts` exists, add cases there; otherwise add a small new test file scoped to
     the adapter.
   - Three cases:
     - `presignAutosave`: mock fetch → flat `{ upload_url, storage_key, expires_at }`; assert
       adapter returns it.
     - `commitAutosave`: mock fetch → flat `TemplateVersion`; assert adapter returns `VersionDTO`
       with all keys populated.
     - `getVersion`: same as `commitAutosave`.
   - **Expected red:** current adapters read `body.data.…` → return `undefined` for all fields.

→ Run `go test ./internal/modules/documents/delivery/http/ ./internal/modules/templates/delivery/http/`
  and the FE test command (e.g. `pnpm --filter web test -- templates`). Confirm A1-A4 all FAIL.
  **Record the red output in `evidence.md` TDD section.**

### Phase B — OpenAPI amendments + codegen

B1. **Add `TemplatePresignAutosaveResponse` schema** under `components.schemas` in
    `api/openapi/v1/openapi.yaml` (sibling to `ApproveTemplateVersionResponse`, ~line 5044):
    ```yaml
    TemplatePresignAutosaveResponse:
      type: object
      required: [upload_url, storage_key, expires_at]
      properties:
        upload_url:  { type: string, format: uri }
        storage_key: { type: string }
        expires_at:  { type: string, format: date-time }
    ```

B2. **Add `TemplateVersion` schema** under `components.schemas` (keys mirror Go
    `templates/domain.TemplateVersion` as wire-serialized today; reference inline `oneOf` /
    `nullable` only if needed for nullable timestamps — keep it flat and tolerant of optional
    fields):
    ```yaml
    TemplateVersion:
      type: object
      required:
        - id
        - template_id
        - version_number
        - revision_number
        - status
        - docx_storage_key
        - content_hash
        - lock_version
        - created_at
      properties:
        id:                    { type: string, format: uuid }
        template_id:           { type: string, format: uuid }
        version_number:        { type: integer }
        revision_number:       { type: integer }
        status:                { type: string }
        docx_storage_key:      { type: string }
        content_hash:          { type: string }
        metadata_schema:       { type: object, additionalProperties: true, nullable: true }
        placeholder_schema:    { type: object, additionalProperties: true, nullable: true }
        author_id:             { type: string, format: uuid, nullable: true }
        pending_reviewer_role: { type: string, nullable: true }
        pending_approver_role: { type: string, nullable: true }
        reviewer_id:           { type: string, format: uuid, nullable: true }
        approver_id:           { type: string, format: uuid, nullable: true }
        submitted_at:          { type: string, format: date-time, nullable: true }
        reviewed_at:           { type: string, format: date-time, nullable: true }
        approved_at:           { type: string, format: date-time, nullable: true }
        published_at:          { type: string, format: date-time, nullable: true }
        obsoleted_at:          { type: string, format: date-time, nullable: true }
        lock_version:          { type: integer }
        created_at:            { type: string, format: date-time }
    ```
   - The exact field-set + nullability is verified against `toVersionResponse`
     (`routes_create.go:66-94`) at write time; the plan above is the seed.

B3. **Amend `createTemplateVersion`** (~line 1346) — change response from `'200': { description: ok }`
    to:
    ```yaml
    '201':
      description: created
      content:
        application/json:
          schema: { $ref: '#/components/schemas/TemplateVersion' }
    ```
    (Drop the bare `'200'` line.)

B4. **Amend `presignTemplateAutosave`** (~line 1428) — change response to:
    ```yaml
    '200':
      description: ok
      content:
        application/json:
          schema: { $ref: '#/components/schemas/TemplatePresignAutosaveResponse' }
    ```

B4a. **Amend `commitTemplateAutosave`** (~line 1449) — change response to:
    ```yaml
    '200':
      description: ok
      content:
        application/json:
          schema: { $ref: '#/components/schemas/TemplateVersion' }
    ```

B4b. **Amend `getTemplateVersion`** (~line 1181) — change response to:
    ```yaml
    '200':
      description: ok
      content:
        application/json:
          schema: { $ref: '#/components/schemas/TemplateVersion' }
    ```

B5. **`renameDocument` unchanged** in OpenAPI — already `200` no-schema.

B6. **Backend codegen:**
    ```
    GOFLAGS=-mod=mod go generate ./internal/modules/templates/api/...
    GOFLAGS=-mod=mod go generate ./internal/modules/documents/api/...
    ```
   - Expect `api.gen.go` diffs only in `templates`: adds typed `TemplateVersion` model,
     `TemplatePresignAutosaveResponse` model. `documents` regen should be a no-op (no OpenAPI
     change to documents endpoints).
   - **Audit the diff** — any unexpected change means HS-3 codegen drift; repair before continuing.

B7. **FE codegen:** `pnpm --filter web gen:api`
   - Expect diffs in `frontend/apps/web/src/lib/api-types/index.d.ts` only: `createTemplateVersion`
     201 response, `presignTemplateAutosave` flat schema, new component schemas
     `TemplateVersion` + `TemplatePresignAutosaveResponse`.

### Phase C — Implementation (green)

C1. **`documents/delivery/http/handler.go` `renameDocument` (line 521 in current source):**
    - Remove `doc, err := h.svc.GetDocument(...)` and the `httpresponse.WriteJSON(w, http.StatusOK, doc)`.
    - Re-check whether removing `GetDocument` removes any necessary error path. **If it does**
      (e.g. error needed to translate stale-state), keep the call but discard `doc` and write
      `w.WriteHeader(http.StatusOK)` after. **If not** (rename success implies doc exists), drop
      the call entirely.
    - Final emit: `w.WriteHeader(http.StatusOK)` only; no `Content-Type` write; no body bytes.

C2. **`templates/delivery/http/routes_create.go` `createNextVersion`:**
    - Add a flat mapper `toAPITemplateVersion(v *domain.TemplateVersion) templatesapi.TemplateVersion`
      next to `toVersionResponse`. The new mapper returns the generated typed struct (snake-case
      JSON via the generated tags), not `map[string]any`.
    - Replace `writeJSON(w, http.StatusCreated, map[string]any{"data":{...}})` with the typed emit
      (via `writeJSON` if it accepts `any`, otherwise via the canonical generated-response helper —
      check whether the templates module has a `WriteTemplateVersion` helper from oapi-codegen
      strict-server mode; if yes, prefer it).
    - Status remains `201`.
    - Leave `toVersionResponse` (legacy `map[string]any`) in place — F1.4 will sweep it.

C3. **`templates/delivery/http/routes_autosave.go` `presignAutosave`:**
    - Replace `writeJSON(w, http.StatusCreated, map[string]any{"data":{...}})` with the typed
      `TemplatePresignAutosaveResponse` emit.
    - Status flips `201 → 200`.

C3a. **`templates/delivery/http/routes_autosave.go` `commitAutosave`:**
    - Replace `writeJSON(w, http.StatusOK, map[string]any{"data":{"version": toVersionResponse(v)}})`
      with the typed `TemplateVersion` emit (use the same mapper added in C2 —
      `toAPITemplateVersion`).
    - Status remains `200`.

C3b. **`templates/delivery/http/routes_query.go` `getVersion`:**
    - Replace `writeJSON(w, http.StatusOK, map[string]any{"data":{"version": toVersionResponse(v)}})`
      with the typed `TemplateVersion` emit (same `toAPITemplateVersion` mapper).
    - Status remains `200`.

C4. **FE adapter edits** in `frontend/apps/web/src/features/templates/api/templates.ts`:
   - `presignAutosave` (`:161-168`):
     ```ts
     export async function presignAutosave(
       templateId: string,
       versionNum: number,
     ): Promise<{ upload_url: string; storage_key: string; expires_at: string }> {
       return apiFetch<{ upload_url: string; storage_key: string; expires_at: string }>(
         `/api/v1/templates/${templateId}/versions/${versionNum}/autosave/presign`,
         { method: 'POST' },
       );
     }
     ```
   - `commitAutosave` (`:171-181`): drop the `body.data.version` indirection — decode
     `apiFetch<VersionDTO>` directly; return value unchanged.
   - `getVersion` (`:154-159`): drop the `body.data.version` indirection — decode
     `apiFetch<VersionDTO>` directly; return value unchanged.

C5. **Run the Phase A tests** — expect A1, A2, A3, A3a, A3b, A4 all GREEN.

### Phase D — Verification (V5–V10)

D1. **Codegen drift gate (V5):** re-run B6 + B7; `git status` shows only intended files dirty.
D2. **H-D grep (V6):**
   ```
   rg "map\[string\]any" internal/modules/templates/delivery/http/routes_create.go:36 \
                         internal/modules/templates/delivery/http/routes_autosave.go:42 \
                         internal/modules/templates/delivery/http/routes_autosave.go:90 \
                         internal/modules/templates/delivery/http/routes_query.go:160
   rg "WriteJSON\([^)]*,\s*doc\)" internal/modules/documents/delivery/http/handler.go
   ```
   All five F1.2 sites: 0 matches. Other handlers (createTemplate, lifecycle, query-list, …) keep
   their envelope — expected; F1.3 / future feature owns them.
D3. **Build + test (V7):** `go build ./...` exit 0; `go test ./...` no `FAIL`.
D4. **FE build (V8):** `pnpm --filter web typecheck`, `pnpm --filter web build`, `pnpm --filter web test`.
D5. **Module regression (V9):** `go test ./internal/modules/documents/... ./internal/modules/templates/...`.
D6. **M0 regression (V10):** `go test ./internal/modules/iam/... ./internal/modules/auth/...`.

### Phase E — Wiki + ADR

E1. **ADR:** create `wiki/decisions/0023-flat-typed-responses-and-presign-status.md`
   (number to be verified — find the next free ADR id at write time). Captures:
   - Modern endpoints return flat typed bodies; legacy `{data:{...}}` envelope is sunset per-endpoint
     as it is touched by a feature (never globally in one PR).
   - Presign endpoints return `200 + body` (industry-standard, S3/GCS parity); status `201` is
     reserved for endpoints that materialize a real server-side resource (e.g. `createTemplateVersion`).
   - Cross-references: F1.1 evidence (DocumentCheckpoint flat precedent); F1.4 plan
     (envelope sunset path).

E2. **Wiki stamp refresh:**
   - `wiki/architecture/api-contract.md` → bump `Last verified` if it references either route.
   - `wiki/references/oapi-codegen.md` → only if its examples list `createTemplateVersion` /
     `presignTemplateAutosave` (likely no).
   - `wiki/decisions/0012-contract-first-api.md` → cross-link the new ADR if relevant.
   - Run `metaldocs-module-doc-sync` skill discipline (named change context, affected modules =
     templates, documents; surfaces = wire shape of A2/A4/A5).

### Phase F — Evidence + commit

F1. Fill `evidence.md` (template): commands + real output for V1–V10, fixture-vs-real labels,
   review/QA disposition, bounded defers (none expected), HS watch.
F2. Commit (CLAUDE.md §5.0 standing authorization). No push.

## Test strategy summary

| Layer | Test home | Real vs fixture |
|-------|-----------|-----------------|
| Backend handler shape (V1–V3) | `*_test.go` next to each handler; existing harness pattern (`handler_checkpoints_test.go` for documents; `routes_autosave_test.go` for templates) | fixture (httptest + fake svc) |
| FE adapter shape (V4) | `templates/api/__tests__/templates.test.ts` (locate or create) | fixture (mocked fetch) |
| Codegen drift (V5) | command output diff | real |
| H-D grep (V6) | `rg` | real |
| Backend + FE build / tests (V7, V8, V9, V10) | `go test ./...`, `pnpm --filter web …` | real |

## Ordering rules

- TDD strict: Phase A before Phase B/C. Implementation does not begin until A1–A4 fail.
- OpenAPI amend → backend codegen → handler change → FE codegen → FE adapter → re-run tests.
  Strictly contract-first (mission §10).
- Per-site commit boundary not enforced (the three sites share the OpenAPI/codegen step); single
  F1.2 commit with all three is fine.

## Hard-stop watch

- **HS-2** — trips if the `TemplateVersion` schema authoring forces a shape that conflicts with
  other endpoints (e.g. `getTemplateVersion` emits a field the schema marks `required` but absent).
  Mitigation: schema fields with uncertain mandatoriness are marked optional (no `required:`); we
  re-add `required:` later as F1.4 / F2 hardens them.
- **HS-3** — trips if codegen produces unexpected drift (a parallel OpenAPI edit landed before this
  feature). Repair via `runtime-contract-prereq`; re-run B6 / B7.
- **HS-6** — already noted: A4 status code deviates from mission §5 text. Recorded in `spec.md`
  interview + `evidence.md`. Validator may flag — operator pre-approved.

## Execution notes

_Filled at TDD time (Phase 3.6) — model choices, deviations, questions answered._
