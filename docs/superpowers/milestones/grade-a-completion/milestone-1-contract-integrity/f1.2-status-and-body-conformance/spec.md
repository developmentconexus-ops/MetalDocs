# Feature F1.2 — Spec

> **Milestone:** 1 — Contract / API integrity  ·  **Folder:** `f1.2-status-and-body-conformance`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-15 / leandrotca.work — A4 HS-6 deviation (keep 201, amend OpenAPI) confirmed by operator: "best for SaaS industry-grade and best logic, for future term".

> This is the feature's **contract**, written and approved **before any code**. It defines what the
> feature must do and how it will be proven — not how it will be built (that is `plan.md`). The
> milestone-validator judges the feature against *this* file (C1).

## Interview record (fail-closed gate)

Engine: inline (B1.5 dialog, one question at a time). `superpowers:brainstorming` skill present but
the consumer contracts for all three sites were objectively discoverable from OpenAPI + FE codegen +
FE adapter source, so the interview centered on resolving three contradictions between mission §5
framing and consumer reality.

| # | Question | Answer |
|---|----------|--------|
| 1 | A5 `presignTemplateAutosave` — OpenAPI declares 200 no-schema; FE (`templates.ts:161-168`) reads `body.data.{upload_url,storage_key,expires_at}` and uses `upload_url` to PUT the docx. Spec or handler wrong? (a) Amend OpenAPI with response schema + flat-typed body + flip 201→200; (b) drop body and redesign FE flow (HS-2). | **(a) flat typed schema, no `{data:{...}}` envelope.** Matches F1.1 precedent (`DocumentCheckpoint` flat). F1.4 inherits clean baseline (no nested wrapper types in codegen). Industry norm for presign endpoints (S3/GCS/Azure): 200 + body. |
| 2 | A4 `createNextVersion` — no FE consumer in repo; OpenAPI says 200 no-schema; handler emits 201 + `{data:{version:...}}`. (a) Drop body, 200 no-content per spec literal; (b) keep 201 + amend OpenAPI with flat `TemplateVersion` body (genuine resource create). | **(b) 201 + flat `TemplateVersion` body, amend OpenAPI.** Endpoint materializes a real DB row → 201 is canonical REST. Mission §5 text "201→200" is a misread of the H-D class — the H-D class is `map[string]any`, not the status code. **HS-6 deviation from mission §5 text; flagged in evidence.md.** |
| 3 | A2 `renameDocument` — OpenAPI says 200 no-schema; FE `Promise<void>`. 200 empty body or 204? | **200 empty body** — satisfies OpenAPI literal; no spec status-code change; FE indifferent. |
| 4 | A5 flat-schema change breaks FE adapter (`templates.ts:165-168` reads `body.data...`). Owner of the FE adapter edit? | **F1.2 owns the FE adapter edit + the FE test update same commit.** Deferral leaves FE broken at runtime (`importTemplateDocx` → upload fails). |
| 5 | Codegen regen commands — verify or trust skill to surface at TDD time? | **Verified now.** Backend: `GOFLAGS=-mod=mod go generate ./internal/modules/templates/api/... ./internal/modules/documents/api/...` (per `wiki/references/oapi-codegen.md`). FE: `pnpm --filter web gen:api` (script in `frontend/apps/web/package.json` → `openapi-typescript ../../../api/openapi/v1/openapi.yaml -o src/lib/api-types/index.d.ts`). |
| 6 | Wiki/module-doc sync — in F1.2 scope or defer to milestone close? | **In scope.** CLAUDE.md §2 drift policy: change code referenced by wiki doc → bump `Last verified` stamp same change. No accumulation. |
| 7 | (2026-06-15, mid-flight) Operator instruction: "remove legacy fallback and adapt to a standard in our code and pattern." HS-6 surfaced — strict F1.2 (3 sites) vs full templates-module sunset (~12 sites, would balloon into a milestone). Three sized options offered: P1 strict / P2-revised full module / P2-light templates-version cluster only. | **P2-light.** F1.2 absorbs `commitTemplateAutosave` (`routes_autosave.go:90`) + `getTemplateVersion` (`routes_query.go:160`) — every `toVersionResponse`-emitting site except `createTemplate` (F1.3 turf). Achieves "no half-typed `TemplateVersion` shape" without crossing into the broader templates-module envelope sunset (lifecycle / query-list / catalog / template-level / docx-upload). milestone.md F1.2 row broadened + F1.4 row shrunk same edit (operator-approved, recorded in milestone.md). Legacy mapper `toVersionResponse` stays alive for `createTemplate` only — F1.3 retires it. |

## Consumer contract (FIRST — before any producer)

**Five** handler sites, five consumer contracts (P2-light expansion — see Q7). Each producer is
rebuilt to match the consumer truth discovered above.

### Site A2 — `renameDocument` (`documents/delivery/http/handler.go:521`)

- **Consumer(s):** `frontend/apps/web/src/features/documents/api/documents.ts:33` (`renameDocument`,
  `Promise<void>`) + any future generated client that uses the OpenAPI declaration directly.
- **Contract:**
  - Method/path: `PATCH /api/v1/documents/{id}`
  - Status: `200 OK`
  - Body: **empty** (`Content-Length: 0`; no JSON envelope)
  - Headers: no `Content-Type` body header expected
- **Source of truth:** `api/openapi/v1/openapi.yaml:2384-2420` (operation `renameDocument`,
  declared `responses: 200: { description: ok }` — no schema).

### Site A4 — `createNextVersion` (`templates/delivery/http/routes_create.go:12`)

- **Consumer(s):** no FE consumer in repo today. Future consumers will read the OpenAPI-generated
  type `TemplateVersion` (already used by `getVersion` / `commitAutosave` flows).
- **Contract:**
  - Method/path: `POST /api/v1/templates/{id}/versions`
  - Status: `201 Created` (genuine resource materialized — new `template_versions` row)
  - Body: flat `TemplateVersion` JSON object (snake_case keys, matching the schema used by
    `getVersion` / `commitAutosave`):
    `{ id, template_id, version_number, revision_number, status, docx_storage_key, content_hash,
      metadata_schema, placeholder_schema, author_id, pending_reviewer_role, pending_approver_role,
      reviewer_id, approver_id, submitted_at, reviewed_at, approved_at, published_at, obsoleted_at,
      lock_version, created_at }`
  - **No `{data:{...}}` envelope.**
- **Source of truth:** OpenAPI `createTemplateVersion` (amended this feature) referencing component
  schema `TemplateVersion` (defined elsewhere in `openapi.yaml`; reuse — do not duplicate).

### Site A5 — `presignTemplateAutosave` (`templates/delivery/http/routes_autosave.go:12`)

- **Consumer(s):** `frontend/apps/web/src/features/templates/api/templates.ts:161-168`
  (`presignAutosave`, returns `{ upload_url, storage_key, expires_at }`); used by
  `importTemplateDocx` (same file `:194`) to PUT the docx to `upload_url`.
- **Contract:**
  - Method/path: `POST /api/v1/templates/{id}/versions/{n}/autosave/presign`
  - Status: `200 OK` (presign computes ephemeral token; no server-side resource materialized —
    industry-standard 200, S3/GCS/Azure parity)
  - Body: flat `TemplatePresignAutosaveResponse` JSON object:
    `{ upload_url: string (uri), storage_key: string, expires_at: string (date-time) }`
  - **No `{data:{...}}` envelope.**
- **Source of truth:** OpenAPI `presignTemplateAutosave` (amended this feature) referencing
  new component schema `TemplatePresignAutosaveResponse`.

### Site P2-A — `commitTemplateAutosave` (`templates/delivery/http/routes_autosave.go:90`)

- **Consumer(s):** `frontend/apps/web/src/features/templates/api/templates.ts:171-181`
  (`commitAutosave`, returns `VersionDTO`); used by `importTemplateDocx` after the upload to
  finalize the autosave revision.
- **Contract:**
  - Method/path: `POST /api/v1/templates/{id}/versions/{n}/autosave/commit`
  - Status: `200 OK` (no new resource created — autosave commits the in-flight revision in place)
  - Body: flat `TemplateVersion` (same component schema as A4)
  - **No `{data:{...}}` envelope.**
- **Source of truth:** OpenAPI `commitTemplateAutosave` (amended this feature) referencing
  `#/components/schemas/TemplateVersion`.

### Site P2-B — `getTemplateVersion` (`templates/delivery/http/routes_query.go:160`)

- **Consumer(s):** `frontend/apps/web/src/features/templates/api/templates.ts:154-159`
  (`getVersion`, returns `VersionDTO`); template editor + version inspector screens.
- **Contract:**
  - Method/path: `GET /api/v1/templates/{id}/versions/{n}`
  - Status: `200 OK`
  - Body: flat `TemplateVersion` (same component schema as A4)
  - **No `{data:{...}}` envelope.**
- **Source of truth:** OpenAPI `getTemplateVersion` (amended this feature) referencing
  `#/components/schemas/TemplateVersion`.

## What this feature implements

1. **OpenAPI amendments** (`api/openapi/v1/openapi.yaml`):
   - **NEW component schema** `TemplateVersion` (under `components.schemas`) — flat, snake-case,
     mirrors the wire shape produced by `toVersionResponse` today (verify field-set + nullability
     against `routes_create.go:66-94` at write time).
   - **NEW component schema** `TemplatePresignAutosaveResponse` — `{ upload_url, storage_key,
     expires_at }`.
   - `createTemplateVersion` (line ~1346): `201 + { schema: $ref TemplateVersion }` (status
     **stays 201** — genuine resource create; HS-6 deviation pre-approved).
   - `presignTemplateAutosave` (line ~1428): `200 + { schema: $ref TemplatePresignAutosaveResponse }`.
   - `commitTemplateAutosave` (line ~1449): `200 + { schema: $ref TemplateVersion }`.
   - `getTemplateVersion` (line ~1181): `200 + { schema: $ref TemplateVersion }`.
   - `renameDocument` (line ~2384): unchanged — already `200` no-schema. No amendment.
2. **Backend codegen regen** (`api.gen.go` for `templates` + `documents` modules): driven by amended
   OpenAPI, not hand-edited.
3. **Handler changes:**
   - `documents/delivery/http/handler.go:521` (`renameDocument`): drop the `httpresponse.WriteJSON(w, 200, doc)`
     → write `200 OK` with empty body via `w.WriteHeader(http.StatusOK)`. Drop the
     `h.svc.GetDocument` call (now unused on the happy path) **only if** no other side-effect
     depends on it — re-verify at implementation time; if it does, keep the call but discard `doc`.
   - `templates/delivery/http/routes_create.go:36` (`createNextVersion`): replace
     `writeJSON(w, 201, map[string]any{data:{version: toVersionResponse(v)}})` with the generated
     typed response (status `201`, flat `TemplateVersion`).
   - `templates/delivery/http/routes_autosave.go:42` (`presignTemplateAutosave`): replace
     `writeJSON(w, 201, map[string]any{data:{...}})` with the generated typed response (status `200`,
     flat `TemplatePresignAutosaveResponse`).
   - `templates/delivery/http/routes_autosave.go:90` (`commitTemplateAutosave`): replace
     `writeJSON(w, 200, map[string]any{data:{version: toVersionResponse(v)}})` with the generated
     typed response (status `200`, flat `TemplateVersion`).
   - `templates/delivery/http/routes_query.go:160` (`getVersion` / `GetTemplateVersion`): replace
     `writeJSON(w, 200, map[string]any{data:{version: toVersionResponse(v)}})` with the generated
     typed response (status `200`, flat `TemplateVersion`).
4. **FE codegen regen** (`frontend/apps/web/src/lib/api-types/index.d.ts`): driven by amended
   OpenAPI.
5. **FE adapter edits** (`frontend/apps/web/src/features/templates/api/templates.ts`):
   - `presignAutosave` (`:161-168`): read flat `{ upload_url, storage_key, expires_at }`; drop
     `body.data.…` indirection. Caller (`importTemplateDocx`) destructure unchanged.
   - `commitAutosave` (`:171-181`): read flat `TemplateVersion`; `body.data.version` → direct body.
     Return type `VersionDTO` unchanged; caller chains (`importTemplateDocx`) unchanged.
   - `getVersion` (`:154-159`): read flat `TemplateVersion`; `body.data.version` → direct body.
     Return type unchanged. Any screen consuming `getVersion` (template editor, version inspector)
     unaffected.
6. **Wiki stamp refresh** (CLAUDE.md §2): bump `Last verified` on any wiki doc referencing the
   three handlers / their routes — minimally `wiki/architecture/api-contract.md` and the touched
   module docs if they exist.

## Non-goals (mandatory)

- No changes to other route status codes or response bodies (F1.1/F1.3/F1.4 turf).
- No FE feature/UX changes beyond the one-line `presignAutosave` adapter edit and its test.
- No `templates/api/templates.ts` rework for endpoints other than `presignAutosave`,
  `commitAutosave`, `getVersion`. Listing, `getTemplate`, `createTemplate`, lifecycle adapters
  keep their current `body.data.…` indirection (F1.3 / future feature turf).
- No `domain.Document` / `domain.TemplateVersion` shape changes (these are domain models;
  wire shape is what changes).
- No idempotency / authz / route-mount changes.
- No `restoreCheckpoint` work (F1.4 turf — A6 class).
- No changes to `createTemplate` (`routes_generated.go:64`) — F1.3 turf (A3). The legacy
  `toVersionResponse` mapper stays alive because `createTemplate` still calls it; F1.3 retires it.
- No changes to `commitTemplate*` / `submit*` / `review*` / `approve*` / `publish*` / `getDocxURL`
  / `listTemplates` / `listVersions` / `placeholderCatalog` / `presignTemplateUpload` — these are
  templates-module endpoints with their own `{data:{...}}` envelope but are NOT in the
  P2-light cluster. Future feature owns them.
- No removal of `{data:{...}}` envelope from endpoints not named here.
- No schema/migration changes (mission Non-Goals).
- No regen of OpenAPI shape from memory (mission §10) — every amendment is grounded in a named
  consumer or industry-standard cite recorded above.

## Validation Gate (concrete — approved before code)

| # | Acceptance criterion | Named test / proof command | Real vs fixture |
|---|----------------------|----------------------------|-----------------|
| V1 | `PATCH /api/v1/documents/{id}` returns `200` with empty body (no `Content-Type: application/json`, `Content-Length: 0`) | `TestRenameDocument_TypedResponseShape` (NEW, `documents/delivery/http/handler_rename_test.go`) — asserts status 200, body length 0 | fixture |
| V2 | `POST /api/v1/templates/{id}/versions` returns `201` with flat `TemplateVersion` body (no `data` wrapper, snake_case keys = `id, template_id, version_number, …, created_at`) | `TestCreateNextVersion_TypedResponseShape` (NEW, `templates/delivery/http/routes_create_test.go`) — asserts 201, exact key set, rejects `data` envelope key | fixture |
| V3 | `POST /api/v1/templates/{id}/versions/{n}/autosave/presign` returns `200` with flat `TemplatePresignAutosaveResponse` (`upload_url`, `storage_key`, `expires_at` — no `data` wrapper) | `TestPresignAutosave_TypedResponseShape` (UPDATE existing `TestPresignAutosave_Happy` in `routes_autosave_test.go:76`) — asserts 200, exact key set | fixture |
| V3a | `POST /api/v1/templates/{id}/versions/{n}/autosave/commit` returns `200` with flat `TemplateVersion` body (no `data` wrapper, same snake_case key set as V2) | `TestCommitAutosave_TypedResponseShape` (UPDATE existing `TestCommitAutosave_Happy` in `routes_autosave_test.go`) — asserts 200, exact key set, rejects `data` top-level key | fixture |
| V3b | `GET /api/v1/templates/{id}/versions/{n}` returns `200` with flat `TemplateVersion` body (no `data` wrapper, same snake_case key set as V2) | `TestGetTemplateVersion_TypedResponseShape` (NEW in `routes_query_test.go`) — asserts 200, exact key set, rejects `data` top-level key | fixture |
| V4 | FE adapters `presignAutosave` / `commitAutosave` / `getVersion` in `frontend/apps/web/src/features/templates/api/templates.ts` read flat bodies; return shapes (`{upload_url,…}`, `VersionDTO`, `VersionDTO`) unchanged to callers | adapter tests in `templates.test.ts` (locate or create) — assert each adapter calls fetch and returns the flat shape | fixture |
| V5 | OpenAPI amendments lint-clean and codegen drift-free | `GOFLAGS=-mod=mod go generate ./internal/modules/templates/api/... ./internal/modules/documents/api/...` → `git diff --stat api.gen.go` shows only intended changes; `pnpm --filter web gen:api` → `git diff` shows only intended changes in `index.d.ts` | real (commands run) |
| V6 | H-D class grep on the five sites returns 0 (`map[string]any` removed; raw domain emit removed) | `rg "map\[string\]any" internal/modules/templates/delivery/http/routes_create.go internal/modules/templates/delivery/http/routes_autosave.go internal/modules/templates/delivery/http/routes_query.go` returns matches **only** outside `createNextVersion` / `commit*Autosave` / `getVersion` (i.e. the remaining out-of-scope handlers that keep their envelope); `rg "WriteJSON\([^)]*,\s*doc\)" internal/modules/documents/delivery/http/handler.go` → no matches for `renameDocument` site | real |
| V7 | Whole-repo build + tests green | `go build ./...` exit 0; `go test ./...` no `FAIL` | real |
| V8 | FE build + typecheck green | `pnpm --filter web build` exit 0; `pnpm --filter web typecheck` exit 0 (whichever script exists) | real |
| V9 | Documents + templates module regression | `go test ./internal/modules/documents/... ./internal/modules/templates/...` all `ok` | real |
| V10 | M0 authz/session corpus regression | `go test ./internal/modules/iam/... ./internal/modules/auth/...` (and any M0-named tests) all `ok` | real |

> TDD: write the failing test first (V1/V2/V3/V4), then implement to green. Fixture proof at the
> handler/adapter boundary is the right grain — the contract surface under test **is** the wire
> shape between handler and FE; repository/DB code is not touched.

## ADR needed?

- [x] Durable decision made → record an ADR.
  Rationale: F1.2 makes two durable architectural decisions:
  (1) presign endpoints return **200** (not 201) with flat typed bodies (industry-standard, S3/GCS
  parity); (2) modern endpoints drop the legacy `{data:{...}}` envelope in favor of flat typed
  responses. Both decisions will shape every future endpoint added to this codebase.
  ADR landed: [`wiki/decisions/0035-flat-typed-responses-and-presign-status.md`](../../../../wiki/decisions/0035-flat-typed-responses-and-presign-status.md) (2026-06-15).

---

**Approval line:**

`Approved: 2026-06-15 / leandrotca.work` — operator confirmed A4 deviation from mission §5 text; A5 flat schema; FE adapter same-commit; wiki stamps in scope.
