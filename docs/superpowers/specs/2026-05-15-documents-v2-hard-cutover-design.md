# Documents V2 Hard Cutover Design

> **Status:** approved for planning
> **Date:** 2026-05-15
> **Scope:** Rename the current Documents v2 product/runtime surface to plain Documents without compatibility aliases.

## 1. Goal

MetalDocs should no longer expose "Documents v2" as the current product concept. The current runtime module is already `documents`, the database table is already `public.documents`, and the public API is already `/api/v1/documents/*`; this change removes the remaining v2 naming from current runtime, contract, frontend, database object names, and wiki truth.

The result should feel like a professional SaaS surface:

- Operators click **Novo documento** and land on `/documents/new`.
- Editors open the DOCX editor at `/documents/:documentID/edit`.
- Published/detail pages keep `/documents/:documentId`.
- API operation names and generated types say `Document`, not `DocumentV2`.
- Current DB constraints/triggers use `documents_*` names, not `documents_v2_*`.
- Historical migrations and evidence logs stay historical; they are not rewritten to hide project history.

## 2. Non-Goals

- Do not rename `docgen-v2`, `DocgenV2Client`, `DOCGEN_V2_*`, or `METALDOCS_DOCGEN_V2_*`; that is a separate service version, not the Documents module label.
- Do not purge `V2` suffixes from templates, taxonomy, IAM, approval, or registry unless a generated Documents contract directly requires it.
- Do not add `/documents-v2/*` redirects, aliases, shims, or compatibility routes.
- Do not edit historical migrations to make old archaeology look new.
- Do not broaden this into the full documents contract migration unless a gate proves it is required.

## 3. Current Reality

The live public API for documents is already `/api/v1/documents/*`, but the current codebase still leaks v2 naming through:

- Frontend product routes: `/documents-v2/new`, `/documents-v2/:documentID`.
- Frontend workspace handles and old navigation helpers: `documents-v2`.
- OpenAPI operationIds: `listDocumentsV2`, `getDocumentV2`, `submitDocumentForApprovalV2`, and related generated symbols.
- Generated Go wrappers and hand-written generated-route adapters such as `PublishDocumentV2`.
- Go runtime strings: `documents_v2` log labels and `document_v2` audit/resource names.
- Curated baseline object names: indexes/triggers such as `ux_documents_v2_cd_active` and `trg_documents_v2_legal_transition`.
- Wiki and acceptance docs that still describe the current wizard/editor as `/documents-v2`.

Historical migrations intentionally contain `documents_v2` because they record the real W1 scaffold and cutover path. Per database policy, those files are evidence and should not be patched for cosmetic cleanup.

## 4. Canonical Route Model

Use plain Documents routes with no ambiguity:

| Purpose | Old route | New route |
|---|---|---|
| Library | `/documents` | unchanged |
| Published/detail view | `/documents/:documentId` | unchanged |
| Published distribution child | `/documents/:documentId/distribution` | unchanged |
| Novo Documento wizard | `/documents-v2/new` | `/documents/new` |
| DOCX editor | `/documents-v2/:documentID` | `/documents/:documentID/edit` |

The editor route deliberately uses `/edit`. A direct rename to `/documents/:documentID` would collide with the existing published/detail route. The new route still removes v2 from the product path while preserving the distinction between viewing the controlled/published document and editing a draft document instance.

Static routes such as `/documents/new`, `/documents/all`, `/documents/mine`, and `/documents/recent` must remain registered before dynamic `/documents/:documentId` routes.

## 5. Contract Model

The public paths stay `/api/v1/documents/*`. The contract rename is about operation names and generated symbols, not URL versioning.

Document-owned operationIds should drop the `V2` suffix:

- `listDocumentsV2` -> `listDocuments`
- `getDocumentV2` -> `getDocument`
- `renameDocumentV2` -> `renameDocument`
- `finalizeDocumentV2` -> `finalizeDocument`
- `archiveDocumentV2` -> `archiveDocument`
- `duplicateDocumentV2` -> `duplicateDocument`
- `documentStatsV2` -> `documentStats`
- session/autosave/checkpoint/comment/fill-in/view/reconstruct/approval document-scoped operations similarly drop `V2`.

Generated Go and frontend API types must be regenerated from OpenAPI after the spec rename. Any hand-written adapter methods that exist only to satisfy generated method names must be renamed in lockstep.

If codegen reveals unrelated module-wide `V2` symbols outside the Documents boundary, stop and classify the mismatch instead of silently broadening the task.

## 6. Database Model

Current database truth should use clean names for current objects:

- `ux_documents_v2_cd_active` -> `ux_documents_cd_active`
- `ux_documents_v2_cd_revision` -> `ux_documents_cd_revision`
- `trg_documents_v2_legal_transition` -> `trg_documents_legal_transition`
- `trg_documents_v2_revision_version_monotonic` -> `trg_documents_revision_version_monotonic`
- `approval_instances.document_v2_id` -> `document_id`, if current runtime/repository usage can be migrated safely in the same transaction and wiki/database dictionary can be updated.

Because local and deployed databases may already contain the old object names, this requires a forward post-baseline migration. The curated baseline should also be updated so fresh installs start clean.

Historical migrations such as `0121_documents_v2_*`, `0133_documents_v2_*`, `0164_documents_v2_visibility.sql`, and `0168_drop_documents_v2_orphan.sql` stay unchanged.

## 7. Frontend Model

Frontend code remains feature-sliced under `frontend/apps/web/src/features/documents/`.

Required frontend changes:

- Route definitions use `documents/new` and `documents/:documentID/edit`.
- Navigation from toolbar, registry, published page, wizard success, and library/sidebar flows points to the new routes.
- Workspace handle names become plain `documents` or a more specific non-v2 value such as `document-editor`.
- Tests assert the new paths.
- No redirect route for `/documents-v2/*` is added.

TanStack Query keys and API wrappers should only change where their names contain v2. Server-state semantics do not change.

## 8. Backend Runtime Model

Backend package names stay `internal/modules/documents`.

Required backend runtime changes:

- Update log labels from `documents_v2` to `documents`.
- Update audit/resource literals from `document_v2` to `document` when they describe the current document resource.
- Rename generated adapter methods in `internal/modules/documents/approval/http/routes_generated.go` after OpenAPI codegen changes.
- Keep route paths `/api/v1/documents/*`; no API path aliases are introduced.

## 9. Wiki and Docs Model

Wiki truth should describe the current product as Documents:

- `wiki/modules/documents.md`
- `wiki/modules/novo-documento-wizard.md`
- `wiki/architecture/frontend-structure.md`
- `wiki/architecture/api-contract.md`
- `wiki/architecture/data-model.md`
- `wiki/tests/system-acceptance-test.md`
- affected workflow/backlog pages
- database dictionary pages for affected tables/columns

Historical UAT logs, evidence files, and old runbooks can retain `documents_v2` when they clearly describe past behavior. Current operational docs, acceptance tests, and module docs should not.

## 10. Verification Strategy

Preflight gates before implementation:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-system-runnable.ps1 -TargetRoute /api/v1/documents
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-module-contract-sync.ps1 -Module documents
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-module-contract-sync.ps1 -Module approval
```

Implementation verification:

```powershell
npx @redocly/cli lint api/openapi/v1/openapi.yaml
$env:GOFLAGS = "-mod=mod"; go generate ./internal/modules/documents/api/...
$env:GOFLAGS = "-mod=mod"; go generate ./internal/modules/documents/approval/api/...
go test ./internal/modules/documents/... -count=1
go test ./apps/api/cmd/metaldocs-api -count=1
cd frontend/apps/web
pnpm gen:api
pnpm.cmd tsc --noEmit -p tsconfig.build.json
pnpm test
```

Runtime smoke:

- Start API and frontend through project scripts.
- Login as dev admin.
- Navigate to `/documents/new`.
- Create a blank-template document.
- Confirm navigation to `/documents/:documentID/edit`.
- Confirm `/documents-v2/new` and `/documents-v2/:id` are not registered as supported product routes.
- Confirm `/documents/:documentId` still opens the published/detail route.

Wiki gates:

```powershell
& 'C:\Program Files\Git\bin\bash.exe' .claude/skills/metaldocs-module-doc/scripts/tally_check.sh documents
& 'C:\Program Files\Git\bin\bash.exe' .claude/skills/metaldocs-module-doc/scripts/tally_check.sh novo-documento-wizard
& 'C:\Program Files\Git\bin\bash.exe' .claude/skills/metaldocs-module-doc/scripts/tally_check.sh approval
```

## 11. Risks

- Route collision risk is real if the editor is renamed to `/documents/:id`; the design avoids this with `/documents/:id/edit`.
- OpenAPI codegen may expose unrelated `V2` suffixes from other modules. Those are not part of this scope unless required for Documents compilation.
- Database column rename `approval_instances.document_v2_id` is more invasive than index/trigger renames because application SQL and dictionary pages must move together. If the gate fails, treat it as a database prerequisite and repair before continuing.
- Existing dirty working-tree changes from the Novo Documento repair should not be reverted or mixed accidentally.

## 12. Success Criteria

- No current runtime/frontend/wiki acceptance path refers to `/documents-v2`.
- Documents OpenAPI/generated symbols no longer use `DocumentV2`/`DocumentsV2` naming.
- Current DB baseline and live migration path use clean Documents names for current objects.
- Novo Documento still creates a Registry slot + Documents draft through the real atomic workflow.
- Published/detail route and editor route both work without ambiguity.
