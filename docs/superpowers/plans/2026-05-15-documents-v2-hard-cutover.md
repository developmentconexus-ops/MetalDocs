# Documents V2 Hard Cutover Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove the current Documents v2 naming from runtime, contracts, frontend routes, database current-object names, and wiki truth, with no `/documents-v2/*` compatibility aliases.

**Architecture:** Keep the live API paths on `/api/v1/documents/*`, but rename document-owned operationIds and generated symbols from `*V2` to plain `Document` names. Frontend uses `/documents/new` for creation and `/documents/:documentID/edit` for the DOCX editor so the existing `/documents/:documentId` published/detail route remains unambiguous. Database work uses one forward migration for current object names plus curated-baseline sync; historical migrations remain evidence and are not rewritten.

**Tech Stack:** Go 1.25, stdlib `net/http` mux, oapi-codegen v2, OpenAPI 3.0.3, React Router, TanStack Query, Vitest, PostgreSQL curated baseline + forward migrations, MetalDocs module wiki sync.

---

## Hard Rules For Execution

- Do not start implementation until the preflight gates in Task 0 pass.
- If any runtime/contract/db startup gate fails, stop feature work and switch to `runtime-contract-prereq`.
- Do not create `/documents-v2/*` redirects, aliases, route shims, or compatibility helpers.
- Do not edit historical migrations such as `0103_docx_v2_documents.sql`, `0121_documents_v2_link_template_version.sql`, `0133_documents_v2_transition_trigger.sql`, `0164_documents_v2_visibility.sql`, or `0168_drop_documents_v2_orphan.sql`.
- Do not rename `docgen-v2`, `DocgenV2Client`, `DOCGEN_V2_*`, or `METALDOCS_DOCGEN_V2_*`.
- Do not purge `V2` suffixes from templates, taxonomy, IAM, registry, or non-document approval surfaces unless a Documents compile/codegen gate proves it is required.
- Preserve unrelated dirty working-tree changes. If files from the previous Novo Documento repair are still dirty, do not revert or restyle them.

## Parallel Agent Strategy

After Task 0, run the discovery pack in parallel. After discovery, implementation can run in parallel only with the file ownership below:

| Lane | Owner | Files |
|---|---|---|
| Contract/codegen | Agent A | `api/openapi/v1/openapi.yaml`, `api/openapi/v1/partials/documents.yaml`, `internal/modules/documents/api/*`, `internal/modules/documents/approval/api/*`, `internal/modules/documents/approval/http/routes_generated.go`, `scripts/check-module-contract-sync.ps1` |
| Database | Agent B | `migrations/0202_documents_current_object_names.sql`, `db/baseline/0001_current_schema.sql`, `wiki/database/**` |
| Frontend | Agent C | `frontend/apps/web/src/**` except generated `lib/api-types/index.d.ts` |
| Backend runtime strings | Agent D | `apps/api/cmd/metaldocs-api/main.go`, `internal/modules/documents/**` excluding generated API dirs and DB-owned files |
| Wiki/docs | Agent E | `wiki/**`, `docs/**`, excluding the implementation plan itself |

Shared files are sequential. In particular:

- `api/openapi/v1/openapi.yaml` and generated API/type files are owned by Agent A only.
- `frontend/apps/web/src/lib/api-types/index.d.ts` is generated after Agent A's OpenAPI change and before frontend verification.
- Wiki sync happens after code truth lands.

---

## File Structure Map

### Contract/API

- Modify `api/openapi/v1/openapi.yaml`: remove `V2` suffix from document-owned operationIds and document-scoped approval operationIds.
- Modify `api/openapi/v1/partials/documents.yaml`: keep partial spec aligned if still consumed by docs/scripts.
- Regenerate `internal/modules/documents/api/api.gen.go`: generated only.
- Regenerate `internal/modules/documents/approval/api/api.gen.go`: generated only.
- Modify `internal/modules/documents/approval/http/routes_generated.go`: rename adapter method names to match regenerated approval API names.
- Modify `scripts/check-module-contract-sync.ps1`: update documents/approval backend pattern expectations where they reference renamed document-owned generated symbols.

### Database

- Create `migrations/0202_documents_current_object_names.sql`: forward-only, idempotent object rename for current index/trigger names.
- Modify `db/baseline/0001_current_schema.sql`: fresh baseline uses `ux_documents_cd_*` and `trg_documents_*` names.
- Modify database wiki pages if they mention stale current names.

### Frontend

- Modify `frontend/apps/web/src/features/documents/routes.tsx`: routes become `documents/new` and `documents/:documentID/edit`; workspace handle becomes `document-editor` or `documents`.
- Modify `frontend/apps/web/src/features/shell/components/AppToolbar.tsx`: new doc navigates to `/documents/new`.
- Modify `frontend/apps/web/src/features/shell/components/AppShell.tsx`: toolbar hide check uses the new handle.
- Modify `frontend/apps/web/src/features/documents/pages/NewDocumentWizardPage.tsx`: cancel path and success navigation use `/documents` and `/documents/:id/edit`.
- Modify `frontend/apps/web/src/features/documents/pages/NewDocumentWizardPage.test.tsx`: assertion expects `/documents/doc-xyz/edit`.
- Modify `frontend/apps/web/src/features/documents/pages/DocumentPublishedPage.tsx`: edit/view action navigates to `/documents/:id/edit`.
- Modify `frontend/apps/web/src/features/registry/pages/RegistryV2Page.tsx` and `frontend/apps/web/src/features/registry/RegistryListPage.tsx`: editor/open/new navigation uses `/documents/...`.
- Modify `frontend/apps/web/src/features/documents/components/LibrarySidebar.tsx`: new wizard path uses `/documents/new`.
- Modify stale current-route references under `frontend/apps/web/src/components/**` only when they are still compiled/current; do not broaden into legacy shell cleanup unless tests compile against them.

### Backend Runtime Strings

- Modify `apps/api/cmd/metaldocs-api/main.go`: rename `documentsV2AuditAdapter` to `documentsAuditAdapter`, `newDocumentsV2AuditAdapter` to `newDocumentsAuditAdapter`, log string to `documents audit write failed`, and comments to `documents`.
- Modify `internal/modules/documents/delivery/http/handler.go`: log labels from `documents_v2 finalize ...` to `documents finalize ...`.
- Modify `internal/modules/documents/approval/application/events_test.go`: expected `ResourceType` becomes `document`.
- Modify any document-module comments that describe current runtime as `documents_v2`; preserve docgen-v2 references.

### Wiki/Docs

- Modify current-truth docs only: `wiki/modules/documents.md`, `wiki/modules/novo-documento-wizard.md`, `wiki/architecture/frontend-structure.md`, `wiki/architecture/api-contract.md`, `wiki/tests/system-acceptance-test.md`, `wiki/workflows/user-onboarding.md`, `wiki/workflows/approval.md`, `wiki/workflows/freeze-and-fanout.md`, relevant backlog pages.
- Preserve historical UAT logs/evidence/runbooks unless they are current operational docs.
- Append sync-log entries for `documents`, `novo-documento-wizard`, and `approval`.

---

## Task 0: Mandatory Preflight Gates

**Files:**
- Read only: repository state and gate scripts.
- No edits.

- [ ] **Step 1: Confirm active branch and dirty state**

Run:

```powershell
git branch --show-current
git status --short
```

Expected:

- Branch is known.
- Dirty files from the previous Novo Documento repair may exist; record them.
- If unrelated user edits directly conflict with files in this plan, stop and ask for direction.

- [ ] **Step 2: Run runtime route gate**

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-system-runnable.ps1 -TargetRoute /api/v1/documents
```

Expected:

```text
PASS login-endpoint
PASS login-session
PASS auth-me
PASS target-route - GET /api/v1/documents returned HTTP 200
```

- [ ] **Step 3: Run documents contract gate**

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-module-contract-sync.ps1 -Module documents
```

Expected: exit 0. If it reports drift, switch to `runtime-contract-prereq` before continuing.

- [ ] **Step 4: Run approval contract gate**

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-module-contract-sync.ps1 -Module approval
```

Expected: exit 0. If it reports drift, switch to `runtime-contract-prereq` before continuing.

- [ ] **Step 5: Record current latest migration**

Run:

```powershell
Get-ChildItem migrations -Filter '*.sql' | Sort-Object Name | Select-Object -Last 5 Name
```

Expected latest before this work: `0201_fix_system_blank_placeholder_schema.sql`. Use `0202_documents_current_object_names.sql` for this plan unless another migration landed first.

---

## Task D1: Parallel Discovery - Contract And Generated Symbol Surface

**Parallel-safe:** yes, read-only.

**Files:**
- Read: `api/openapi/v1/openapi.yaml`
- Read: `api/openapi/v1/partials/documents.yaml`
- Read: `internal/modules/documents/api/api.gen.go`
- Read: `internal/modules/documents/approval/api/api.gen.go`
- Read: `internal/modules/documents/approval/http/routes_generated.go`
- Read: `scripts/check-module-contract-sync.ps1`

- [ ] **Step 1: Build the document operation rename table**

Run:

```powershell
rg -n "operationId: .*Document.*V2|operationId: listDocumentsV2|operationId: documentStatsV2|operationId: .*SessionV2|operationId: .*AutosaveV2|operationId: .*CheckpointV2|operationId: .*CommentV2|operationId: .*Placeholder.*V2" api/openapi/v1/openapi.yaml api/openapi/v1/partials/documents.yaml
```

Expected: list only documents-owned or document-scoped operations. Do not include templates/taxonomy/registry.

- [ ] **Step 2: Identify generated method names that must change**

Run:

```powershell
rg -n "DocumentsV2|DocumentV2|ListDocumentsV2|SubmitDocumentForApprovalV2|PublishDocumentV2|RecordDocumentSignoffV2|CancelDocumentApprovalV2|GetApprovalInstanceByDocumentV2" internal/modules/documents/api/api.gen.go internal/modules/documents/approval/api/api.gen.go internal/modules/documents/approval/http/routes_generated.go
```

Expected: generated methods and hand-written adapters that correspond to operationIds in Step 1.

- [ ] **Step 3: Report exact rename table**

Output a compact table like:

```markdown
| Old operationId | New operationId | Owning tag | Adapter method? |
|---|---|---|---|
| listDocumentsV2 | listDocuments | documents | generated only |
| publishDocumentV2 | publishDocument | approval | routes_generated.go |
```

Do not edit files in this discovery task.

---

## Task D2: Parallel Discovery - Database Current Object Surface

**Parallel-safe:** yes, read-only.

**Files:**
- Read: `db/baseline/0001_current_schema.sql`
- Read: `migrations/*.sql`
- Read: `wiki/database/tables/*.md`
- Read: `wiki/modules/documents/_artifacts/04-persistence.md`
- Read: `wiki/modules/approval*.md`

- [ ] **Step 1: Confirm approval column rename is already landed**

Run:

```powershell
rg -n "document_v2_id|document_id" db/baseline/0001_current_schema.sql migrations/0194_approval_document_id_rename.sql internal/modules/documents/approval/repository/postgres_approval_repository.go wiki/database/tables/approval_instances.md
```

Expected:

- Baseline and repository use `document_id`.
- Migration `0194_approval_document_id_rename.sql` exists.
- Do not create another column rename migration.

- [ ] **Step 2: Find current DB object names that still leak documents_v2**

Run:

```powershell
rg -n "ux_documents_v2|trg_documents_v2" db/baseline/0001_current_schema.sql migrations wiki/database wiki/modules/documents.md wiki/modules/documents/_artifacts/04-persistence.md
```

Expected current-object candidates:

- `ux_documents_v2_cd_active`
- `ux_documents_v2_cd_revision`
- `trg_documents_v2_legal_transition`
- `trg_documents_v2_revision_version_monotonic`

Historical migrations may appear; record them as out of scope.

- [ ] **Step 3: Report DB rename table**

Output:

```markdown
| Old object | New object | Needs forward migration | Needs baseline sync |
|---|---|---|---|
| ux_documents_v2_cd_active | ux_documents_cd_active | yes | yes |
```

Do not edit files in this discovery task.

---

## Task D3: Parallel Discovery - Frontend Route And Navigation Surface

**Parallel-safe:** yes, read-only.

**Files:**
- Read: `frontend/apps/web/src/features/documents/routes.tsx`
- Read: `frontend/apps/web/src/features/shell/components/AppToolbar.tsx`
- Read: `frontend/apps/web/src/features/shell/components/AppShell.tsx`
- Read: `frontend/apps/web/src/features/documents/pages/NewDocumentWizardPage.tsx`
- Read: `frontend/apps/web/src/features/documents/pages/DocumentPublishedPage.tsx`
- Read: `frontend/apps/web/src/features/registry/**`
- Read: `frontend/apps/web/src/components/**`

- [ ] **Step 1: Find current frontend v2 routes**

Run:

```powershell
rg -n "documents-v2|workspaceView.*documents-v2|Documents v2|/documents-v2" frontend/apps/web/src --glob '!**/node_modules/**'
```

Expected: current route declarations, navigation callsites, tests, and legacy shell references.

- [ ] **Step 2: Classify frontend files**

Classify each hit as:

- `current runtime`: route/navigation/test used by app today
- `legacy compiled shell`: still compiled but not central to current route path
- `historical/stale comment`: can be updated only if it describes current behavior

- [ ] **Step 3: Report exact frontend ownership**

Output:

```markdown
| File | Current v2 reference | New value | Classification |
|---|---|---|---|
| frontend/apps/web/src/features/documents/routes.tsx | documents-v2/new | documents/new | current runtime |
```

Do not edit files in this discovery task.

---

## Task D4: Parallel Discovery - Wiki Current Truth Surface

**Parallel-safe:** yes, read-only.

**Files:**
- Read: `wiki/modules/documents.md`
- Read: `wiki/modules/novo-documento-wizard.md`
- Read: `wiki/architecture/frontend-structure.md`
- Read: `wiki/architecture/api-contract.md`
- Read: `wiki/tests/system-acceptance-test.md`
- Read: current workflow/backlog docs

- [ ] **Step 1: Find current wiki/docs v2 references**

Run:

```powershell
rg -n "documents-v2|/api/v2/documents|documents_v2|document_v2|DocumentV2|DocumentsV2" wiki/modules/documents.md wiki/modules/novo-documento-wizard.md wiki/architecture/frontend-structure.md wiki/architecture/api-contract.md wiki/architecture/data-model.md wiki/tests/system-acceptance-test.md wiki/workflows/user-onboarding.md wiki/workflows/approval.md wiki/workflows/freeze-and-fanout.md wiki/backlog/documents-refactor.md wiki/backlog/novo-documento.md wiki/backlog/editor.md wiki/backlog/library-screen.md wiki/backlog/documento-publicado.md wiki/backlog/contract-first-followups.md wiki/modules/approval.md wiki/modules/approval-tech-debt.md wiki/backlog/approval-refactor.md
```

- [ ] **Step 2: Classify each doc reference**

Use:

- `current truth`: must be updated.
- `historical exception`: leave unchanged only if the doc is clearly historical.
- `debt row closed`: update status rather than delete evidence.

- [ ] **Step 3: Report wiki sync target list**

Output exact files that need patching after code truth lands.

---

## Task 1: Contract Rename And Codegen

**Depends on:** Tasks D1-D4 complete.

**Parallel-safe:** no. Owns shared contract and generated files.

**Files:**
- Modify: `api/openapi/v1/openapi.yaml`
- Modify: `api/openapi/v1/partials/documents.yaml`
- Generate: `internal/modules/documents/api/api.gen.go`
- Generate: `internal/modules/documents/approval/api/api.gen.go`
- Modify: `internal/modules/documents/approval/http/routes_generated.go`
- Modify: `scripts/check-module-contract-sync.ps1`

- [ ] **Step 1: Write the failing contract assertions**

Run before editing:

```powershell
rg -n "operationId: listDocumentsV2|operationId: getDocumentV2|operationId: publishDocumentV2|operationId: recordDocumentSignoffV2" api/openapi/v1/openapi.yaml
```

Expected before implementation: matches found. This is the RED evidence for the contract rename.

- [ ] **Step 2: Rename documents-tag operationIds in OpenAPI**

In `api/openapi/v1/openapi.yaml`, apply these replacements exactly:

```text
listDocumentsV2 -> listDocuments
getDocumentV2 -> getDocument
renameDocumentV2 -> renameDocument
finalizeDocumentV2 -> finalizeDocument
archiveDocumentV2 -> archiveDocument
duplicateDocumentV2 -> duplicateDocument
documentStatsV2 -> documentStats
acquireDocumentSessionV2 -> acquireDocumentSession
heartbeatDocumentSessionV2 -> heartbeatDocumentSession
releaseDocumentSessionV2 -> releaseDocumentSession
forceReleaseDocumentSessionV2 -> forceReleaseDocumentSession
presignDocumentAutosaveV2 -> presignDocumentAutosave
commitDocumentAutosaveV2 -> commitDocumentAutosave
listDocumentCheckpointsV2 -> listDocumentCheckpoints
createDocumentCheckpointV2 -> createDocumentCheckpoint
restoreDocumentCheckpointV2 -> restoreDocumentCheckpoint
getDocumentRevisionUrlV2 -> getDocumentRevisionUrl
listDocumentCommentsV2 -> listDocumentComments
createDocumentCommentV2 -> createDocumentComment
updateDocumentCommentV2 -> updateDocumentComment
deleteDocumentCommentV2 -> deleteDocumentComment
getDocumentFillInSchemaV2 -> getDocumentFillInSchema
listDocumentPlaceholderValuesV2 -> listDocumentPlaceholderValues
putDocumentPlaceholderValueV2 -> putDocumentPlaceholderValue
viewDocumentV2 -> viewDocument
reconstructDocumentV2 -> reconstructDocument
getDocumentPlaceholderOptionsV2 -> getDocumentPlaceholderOptions
```

Also rename document-scoped approval operationIds only:

```text
submitDocumentForApprovalV2 -> submitDocumentForApproval
publishDocumentV2 -> publishDocument
scheduleDocumentPublishV2 -> scheduleDocumentPublish
supersedeDocumentV2 -> supersedeDocument
obsoleteDocumentV2 -> obsoleteDocument
getApprovalInstanceByDocumentV2 -> getApprovalInstanceByDocument
recordDocumentSignoffV2 -> recordDocumentSignoff
cancelDocumentApprovalV2 -> cancelDocumentApproval
```

Leave non-document approval operations with `V2` for this plan:

```text
recordApprovalStageSignoffV2
cancelApprovalInstanceV2
getApprovalInstanceV2
listApprovalInboxV2
createApprovalRouteV2
listApprovalRoutesV2
updateApprovalRouteV2
deactivateApprovalRouteV2
```

- [ ] **Step 3: Keep the documents partial spec aligned**

If `api/openapi/v1/partials/documents.yaml` still contains these operationIds, replace:

```text
listDocumentsV2 -> listDocuments
createDocumentV2 -> createDocument
```

Do not change templates partials in this task.

- [ ] **Step 4: Verify OpenAPI RED is gone**

Run:

```powershell
rg -n "operationId: .*Document.*V2|operationId: listDocumentsV2|operationId: documentStatsV2|operationId: .*SessionV2|operationId: .*AutosaveV2|operationId: .*CheckpointV2|operationId: .*CommentV2|operationId: .*Placeholder.*V2" api/openapi/v1/openapi.yaml api/openapi/v1/partials/documents.yaml
```

Expected: no document-owned hits. If hits are unrelated non-document modules, classify before editing.

- [ ] **Step 5: Regenerate backend API packages**

Run:

```powershell
$env:GOFLAGS = "-mod=mod"
go generate ./internal/modules/documents/api/...
go generate ./internal/modules/documents/approval/api/...
```

Expected: generated files update with renamed methods/types.

- [ ] **Step 6: Update approval generated-route adapters**

In `internal/modules/documents/approval/http/routes_generated.go`, rename only methods that correspond to the document-scoped operationIds changed above:

```go
func (h *Handler) SubmitDocumentForApproval(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.SubmitHandler(w, r)
}

func (h *Handler) PublishDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.PublishHandler(w, r)
}

func (h *Handler) ScheduleDocumentPublish(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.SchedulePublishHandler(w, r)
}

func (h *Handler) SupersedeDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.SupersedeHandler(w, r)
}

func (h *Handler) ObsoleteDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.ObsoleteHandler(w, r)
}

func (h *Handler) GetApprovalInstanceByDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.GetInstanceByDocumentHandler(w, r)
}

func (h *Handler) RecordDocumentSignoff(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.SignoffByDocumentHandler(w, r)
}

func (h *Handler) CancelDocumentApproval(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	h.CancelByDocumentHandler(w, r)
}
```

Leave non-document methods unchanged:

```go
RecordApprovalStageSignoffV2
CancelApprovalInstanceV2
GetApprovalInstanceV2
ListApprovalInboxV2
CreateApprovalRouteV2
UpdateApprovalRouteV2
DeactivateApprovalRouteV2
ListApprovalRoutesV2
```

- [ ] **Step 7: Update contract sync script expectations**

In `scripts/check-module-contract-sync.ps1`, update only documents-related expected symbols.

If the `approval` config still uses broad non-document patterns:

```powershell
BackendPatterns = @('ListApprovalInboxV2', 'ListApprovalRoutesV2', 'CreateApprovalRouteV2')
```

leave it unchanged because those operations remain non-document approval scope.

If the documents default config or future explicit documents config references `ListDocumentsV2`, change it to `ListDocuments`.

- [ ] **Step 8: Run contract/codegen verification**

Run:

```powershell
npx @redocly/cli lint api/openapi/v1/openapi.yaml
go test ./internal/modules/documents/approval/http -count=1
go test ./internal/modules/documents/delivery/http -count=1
```

Expected: exit 0. If Redocly reports pre-existing lint outside changed paths, record it and run the project's accepted lint gate if one exists.

---

## Task 2: Database Current Object Rename

**Depends on:** Task D2 complete.

**Parallel-safe:** yes with Tasks 1, 3, 4. Owns DB files only.

**Files:**
- Create: `migrations/0202_documents_current_object_names.sql`
- Modify: `db/baseline/0001_current_schema.sql`
- Modify: database wiki pages if current object names are documented

- [ ] **Step 1: Write migration file**

Create `migrations/0202_documents_current_object_names.sql`:

```sql
-- 0202_documents_current_object_names.sql
-- Rename current Documents indexes/triggers from migration-era documents_v2 names.

DO $$
BEGIN
  IF to_regclass('public.ux_documents_v2_cd_active') IS NOT NULL
     AND to_regclass('public.ux_documents_cd_active') IS NULL THEN
    ALTER INDEX public.ux_documents_v2_cd_active RENAME TO ux_documents_cd_active;
  END IF;

  IF to_regclass('public.ux_documents_v2_cd_revision') IS NOT NULL
     AND to_regclass('public.ux_documents_cd_revision') IS NULL THEN
    ALTER INDEX public.ux_documents_v2_cd_revision RENAME TO ux_documents_cd_revision;
  END IF;
END $$;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM pg_trigger
    WHERE tgname = 'trg_documents_v2_legal_transition'
      AND tgrelid = 'public.documents'::regclass
  ) AND NOT EXISTS (
    SELECT 1
    FROM pg_trigger
    WHERE tgname = 'trg_documents_legal_transition'
      AND tgrelid = 'public.documents'::regclass
  ) THEN
    ALTER TRIGGER trg_documents_v2_legal_transition
      ON public.documents
      RENAME TO trg_documents_legal_transition;
  END IF;

  IF EXISTS (
    SELECT 1
    FROM pg_trigger
    WHERE tgname = 'trg_documents_v2_revision_version_monotonic'
      AND tgrelid = 'public.documents'::regclass
  ) AND NOT EXISTS (
    SELECT 1
    FROM pg_trigger
    WHERE tgname = 'trg_documents_revision_version_monotonic'
      AND tgrelid = 'public.documents'::regclass
  ) THEN
    ALTER TRIGGER trg_documents_v2_revision_version_monotonic
      ON public.documents
      RENAME TO trg_documents_revision_version_monotonic;
  END IF;
END $$;

INSERT INTO public.schema_migrations (version, name)
VALUES ('0202', 'rename current documents indexes and triggers')
ON CONFLICT (version) DO NOTHING;
```

- [ ] **Step 2: Update curated baseline object names**

In `db/baseline/0001_current_schema.sql`, replace current-object names:

```text
ux_documents_v2_cd_active -> ux_documents_cd_active
ux_documents_v2_cd_revision -> ux_documents_cd_revision
trg_documents_v2_legal_transition -> trg_documents_legal_transition
trg_documents_v2_revision_version_monotonic -> trg_documents_revision_version_monotonic
```

Do not edit historical migration comments in `migrations/`.

- [ ] **Step 3: Verify DB naming grep**

Run:

```powershell
rg -n "ux_documents_v2|trg_documents_v2" db/baseline/0001_current_schema.sql migrations/0202_documents_current_object_names.sql wiki/database wiki/modules/documents.md wiki/modules/documents/_artifacts/04-persistence.md
```

Expected: no hits in baseline or current wiki after wiki sync; migration `0202` may contain old names intentionally as source names.

- [ ] **Step 4: Run DB dictionary and bootstrap gates**

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-db-dictionary-coverage.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-bootstrap-baseline.ps1 -WithDevSeed
```

Expected: exit 0. If bootstrap is too expensive in the active session, run at least dictionary coverage now and mark bootstrap required before completion.

---

## Task 3: Frontend Route Hard Cutover

**Depends on:** Task D3 complete.

**Parallel-safe:** yes with Tasks 1, 2, 4. Do not touch generated frontend API types.

**Files:**
- Modify: `frontend/apps/web/src/features/documents/routes.tsx`
- Modify: `frontend/apps/web/src/features/shell/components/AppToolbar.tsx`
- Modify: `frontend/apps/web/src/features/shell/components/AppShell.tsx`
- Modify: `frontend/apps/web/src/features/documents/pages/NewDocumentWizardPage.tsx`
- Modify: `frontend/apps/web/src/features/documents/pages/NewDocumentWizardPage.test.tsx`
- Modify: `frontend/apps/web/src/features/documents/pages/DocumentPublishedPage.tsx`
- Modify: `frontend/apps/web/src/features/registry/pages/RegistryV2Page.tsx`
- Modify: `frontend/apps/web/src/features/registry/RegistryListPage.tsx`
- Modify: `frontend/apps/web/src/features/documents/components/LibrarySidebar.tsx`
- Modify compiled current references under `frontend/apps/web/src/components/**` if needed for typecheck

- [ ] **Step 1: Update tests first for new success navigation**

In `frontend/apps/web/src/features/documents/pages/NewDocumentWizardPage.test.tsx`, change:

```ts
it('navigates to /documents-v2/<document.id> on success', async () => {
```

to:

```ts
it('navigates to /documents/<document.id>/edit on success', async () => {
```

Change expected navigation:

```ts
expect(mockNavigate).toHaveBeenCalledWith('/documents/doc-xyz/edit');
```

Run RED:

```powershell
cmd /c "cd frontend\apps\web && npm test -- NewDocumentWizardPage"
```

Expected before production edit: test fails because page still navigates to `/documents-v2/doc-xyz`.

- [ ] **Step 2: Update route declarations**

In `frontend/apps/web/src/features/documents/routes.tsx`, replace the two v2 route entries with:

```tsx
  {
    path: "documents/new",
    handle: { workspaceView: "document-editor" },
    lazy: () => import("./pages/NewDocumentWizardPage").then((m) => ({ Component: m.NewDocumentWizardPage })),
  },
  {
    path: "documents/:documentID/edit",
    handle: { workspaceView: "document-editor" },
    lazy: () => import("./pages/DocumentEditorRoutePage"),
  },
```

Keep these entries before the dynamic `documents/:documentId` published/detail route, or move the published/detail route below them, so `documents/new` and `documents/:documentID/edit` are not swallowed by detail routing.

- [ ] **Step 3: Update toolbar and shell handle**

In `frontend/apps/web/src/features/shell/components/AppToolbar.tsx`, change:

```tsx
onClick={() => navigate('/documents-v2/new')}
```

to:

```tsx
onClick={() => navigate('/documents/new')}
```

In `frontend/apps/web/src/features/shell/components/AppShell.tsx`, change:

```ts
(m) => (m.handle as RouteHandle | undefined)?.workspaceView === 'documents-v2',
```

to:

```ts
(m) => (m.handle as RouteHandle | undefined)?.workspaceView === 'document-editor',
```

- [ ] **Step 4: Update wizard navigation**

In `frontend/apps/web/src/features/documents/pages/NewDocumentWizardPage.tsx`, change cancel:

```ts
navigate('/documents-v2');
```

to:

```ts
navigate('/documents');
```

Change success:

```ts
navigate(`/documents-v2/${result.document.id}`);
```

to:

```ts
navigate(`/documents/${result.document.id}/edit`);
```

- [ ] **Step 5: Update published page editor action**

In `frontend/apps/web/src/features/documents/pages/DocumentPublishedPage.tsx`, change:

```ts
const handleView = () => navigate(`/documents-v2/${documentId}`);
```

to:

```ts
const handleView = () => navigate(`/documents/${documentId}/edit`);
```

- [ ] **Step 6: Update registry navigation**

In `frontend/apps/web/src/features/registry/pages/RegistryV2Page.tsx`, change:

```tsx
onOpenDocumentEditor={(docId) => navigate(`/documents-v2/${docId}`)}
```

to:

```tsx
onOpenDocumentEditor={(docId) => navigate(`/documents/${docId}/edit`)}
```

In `frontend/apps/web/src/features/registry/RegistryListPage.tsx`, change:

```tsx
onClick={() => navigate('/documents-v2/new')}
```

to:

```tsx
onClick={() => navigate('/documents/new')}
```

- [ ] **Step 7: Update library/sidebar navigation**

In `frontend/apps/web/src/features/documents/components/LibrarySidebar.tsx`, replace `/documents-v2/new` with `/documents/new`.

If `frontend/apps/web/src/components/DocumentWorkspaceShell.tsx` still compiles and contains `"documents-v2"`, rename that workspace value to `"document-editor"` and label to `"Documents"` only if TypeScript reports it as current. Do not refactor the legacy shell otherwise.

- [ ] **Step 8: Verify frontend grep**

Run:

```powershell
rg -n "documents-v2|Documents v2|workspaceView.*documents-v2|/documents-v2" frontend/apps/web/src --glob '!**/node_modules/**'
```

Expected: no current runtime hits. Historical comments in compiled source should be updated or classified.

- [ ] **Step 9: Run frontend tests**

Run:

```powershell
cmd /c "cd frontend\apps\web && npm test -- NewDocumentWizardPage"
```

Expected: pass.

Full frontend verification runs in Task 8 after generated API types are refreshed.

---

## Task 4: Backend Runtime String Cleanup

**Depends on:** Task D1 complete.

**Parallel-safe:** yes with Tasks 1, 2, 3. Avoid generated API files.

**Files:**
- Modify: `apps/api/cmd/metaldocs-api/main.go`
- Modify: `internal/modules/documents/delivery/http/handler.go`
- Modify: `internal/modules/documents/approval/application/events_test.go`
- Modify: `internal/platform/docgenv2/templates_v2_reader.go` comments only if they incorrectly say `documents_v2/application`

- [ ] **Step 1: Update test expectation first**

In `internal/modules/documents/approval/application/events_test.go`, change:

```go
ResourceType: "document_v2",
```

to:

```go
ResourceType: "document",
```

Run RED:

```powershell
go test ./internal/modules/documents/approval/application -run Test -count=1
```

Expected before production edit: failing assertion if runtime still emits `document_v2`.

- [ ] **Step 2: Rename audit adapter type and constructor**

In `apps/api/cmd/metaldocs-api/main.go`, replace:

```go
type documentsV2AuditAdapter struct {
	writer auditdomain.Writer
}

func newDocumentsV2AuditAdapter(writer auditdomain.Writer) *documentsV2AuditAdapter {
	return &documentsV2AuditAdapter{writer: writer}
}

func (a *documentsV2AuditAdapter) WriteTx(...)
func (a *documentsV2AuditAdapter) Write(...)
```

with:

```go
type documentsAuditAdapter struct {
	writer auditdomain.Writer
}

func newDocumentsAuditAdapter(writer auditdomain.Writer) *documentsAuditAdapter {
	return &documentsAuditAdapter{writer: writer}
}

func (a *documentsAuditAdapter) WriteTx(...)
func (a *documentsAuditAdapter) Write(...)
```

Also replace callsites:

```go
newDocumentsV2AuditAdapter(...)
```

with:

```go
newDocumentsAuditAdapter(...)
```

- [ ] **Step 3: Rename log labels and comments**

Replace current-runtime log labels:

```go
log.Printf("documents_v2 audit write failed: %v", err)
log.Printf("documents_v2 finalize audit-only error: %v", err)
log.Printf("documents_v2 finalize idempotency record error: %v", err)
```

with:

```go
log.Printf("documents audit write failed: %v", err)
log.Printf("documents finalize audit-only error: %v", err)
log.Printf("documents finalize idempotency record error: %v", err)
```

In `apps/api/cmd/metaldocs-api/main.go`, change the current-runtime comment:

```go
// profileDefaultsAdapter bridges taxonomy ProfileRepository → documents_v2 ProfileDefaultTemplateReader.
```

to:

```go
// profileDefaultsAdapter bridges taxonomy ProfileRepository to documents ProfileDefaultTemplateReader.
```

- [ ] **Step 4: Keep docgen-v2 comments intact unless they point to documents_v2**

In `internal/platform/docgenv2/templates_v2_reader.go`, if the comment says:

```go
// TemplatesV2TemplateReader implements documents_v2/application.TemplateReader
```

change only the target module reference:

```go
// TemplatesV2TemplateReader implements documents/application.TemplateReader.
```

Do not rename `TemplatesV2TemplateReader`; it refers to template storage naming and is out of scope.

- [ ] **Step 5: Run backend runtime tests**

Run:

```powershell
gofmt -w apps/api/cmd/metaldocs-api/main.go internal/modules/documents/delivery/http/handler.go internal/modules/documents/approval/application/events_test.go internal/platform/docgenv2/templates_v2_reader.go
go test ./internal/modules/documents/approval/application -count=1
go test ./apps/api/cmd/metaldocs-api -count=1
```

Expected: pass.

---

## Task 5: Frontend Generated Types And Contract Consumers

**Depends on:** Task 1 complete.

**Parallel-safe:** no; owns generated frontend API types.

**Files:**
- Generate: `frontend/apps/web/src/lib/api-types/index.d.ts`
- Modify: frontend API wrappers only if renamed operation types break compilation

- [ ] **Step 1: Regenerate frontend API types**

Run:

```powershell
cmd /c "cd frontend\apps\web && pnpm gen:api"
```

Expected: `frontend/apps/web/src/lib/api-types/index.d.ts` updates generated operation/type names.

- [ ] **Step 2: Compile to reveal consumer drift**

Run:

```powershell
cmd /c "cd frontend\apps\web && pnpm.cmd tsc --noEmit -p tsconfig.build.json"
```

Expected after Task 1 + Task 3: pass or actionable references to renamed operation type keys.

- [ ] **Step 3: Repair generated type references only if needed**

If TypeScript fails on operation keys such as:

```ts
operations['listDocumentsV2']
operations['getDocumentV2']
```

rename to the generated plain key:

```ts
operations['listDocuments']
operations['getDocument']
```

Do not hand-write replacement API types. Use `components`, `operations`, or `paths` from generated `lib/api-types`.

---

## Task 6: Wiki And Docs Sync

**Depends on:** Tasks 1-5 complete.

**Parallel-safe:** yes after code truth lands. Owns wiki/docs only.

**Files:**
- Modify: `wiki/modules/documents.md`
- Modify: `wiki/modules/documents/_artifacts/04-persistence.md`
- Modify: `wiki/modules/documents/_artifacts/sync-log.md`
- Modify: `wiki/modules/novo-documento-wizard.md`
- Modify: `wiki/modules/novo-documento-wizard/_artifacts/sync-log.md`
- Modify: `wiki/modules/approval.md`
- Modify: `wiki/modules/approval-tech-debt.md`
- Modify: `wiki/backlog/approval-refactor.md`
- Modify: `wiki/architecture/frontend-structure.md`
- Modify: `wiki/architecture/api-contract.md`
- Modify: current workflow/backlog docs identified by D4

- [ ] **Step 1: Run wiki sync preflight**

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .claude/skills/metaldocs-module-doc-sync/scripts/wiki_sync_preflight.ps1 -Module documents
powershell -NoProfile -ExecutionPolicy Bypass -File .claude/skills/metaldocs-module-doc-sync/scripts/wiki_sync_preflight.ps1 -Module novo-documento-wizard
powershell -NoProfile -ExecutionPolicy Bypass -File .claude/skills/metaldocs-module-doc-sync/scripts/wiki_sync_preflight.ps1 -Module approval
```

Expected: each reports `readyForSync: true`.

- [ ] **Step 2: Update documents module truth**

In `wiki/modules/documents.md`:

- Update `Last verified` to this cutover date/context.
- Replace generated stub examples `ListDocumentsV2`/`DocumentV2` with plain names.
- Replace current one-active approval note from `document_v2_id` to `document_id`.
- Preserve historical note that `public.documents_v2` was dropped by migration 0168.
- Update route table operationIds to plain names for documents-owned operations.

- [ ] **Step 3: Update Novo Documento wizard truth**

In `wiki/modules/novo-documento-wizard.md`:

- Scope route becomes `/documents/new`.
- Success/edit route becomes `/documents/:documentID/edit`.
- Runtime smoke references use `/documents/new`.
- Keep API create endpoint as `/api/v1/controlled-documents`.

- [ ] **Step 4: Update approval truth for column rename status**

In `wiki/modules/approval-tech-debt.md`, close T-008 if code/baseline/migration confirm it is done:

```markdown
### T-008 · `approval_instances.document_v2_id` column retains `_v2` suffix post-cutover — CLOSED 2026-05-15
- **Severity:** minor (closed)
- **Surface (resolved):** `migrations/0194_approval_document_id_rename.sql`; `db/baseline/0001_current_schema.sql`; `internal/modules/documents/approval/repository/postgres_approval_repository.go`
- **Resolution:** Current schema and repository use `approval_instances.document_id`; the old `document_v2_id` name remains only in historical migrations.
```

In `wiki/backlog/approval-refactor.md`, mark R-008 `merged` or `done` with migration `0194` evidence, not this plan's migration.

- [ ] **Step 5: Update frontend architecture and acceptance docs**

In `wiki/architecture/frontend-structure.md`, replace route example:

```ts
{ path: "documents/new", lazy: () => import("./pages/NewDocumentWizardPage") },
{ path: "documents/:documentID/edit", lazy: () => import("./pages/DocumentEditorRoutePage") },
```

In `wiki/tests/system-acceptance-test.md`, update E12 route from `/documents-v2/new` to `/documents/new`.

In `wiki/workflows/user-onboarding.md`, update wizard route to `/documents/new`.

- [ ] **Step 6: Update current backlog endpoint references**

For current backlog pages, replace `/api/v2/documents` examples with `/api/v1/documents` unless the page is explicitly historical. At minimum:

```text
wiki/backlog/editor.md
wiki/backlog/library-screen.md
wiki/backlog/documento-publicado.md
wiki/backlog/contract-first-followups.md
wiki/workflows/approval.md
wiki/workflows/freeze-and-fanout.md
```

- [ ] **Step 7: Append sync log entries**

Prepend compact entries to:

```text
wiki/modules/documents/_artifacts/sync-log.md
wiki/modules/novo-documento-wizard/_artifacts/sync-log.md
wiki/modules/approval/_artifacts/sync-log.md
```

Use the template shape:

```markdown
## 2026-05-15 - Documents v2 hard cutover

- **Context:** docs/specs/2026-05-15-documents-v2-hard-cutover-design.md plus implementation diff
- **Mode:** lite patch
- **Anchors moved:** route/path and generated-symbol references
- **Public surface:** Documents operationIds/routes renamed to plain Documents naming
- **Routes/API:** `/api/v1/documents/*` unchanged; frontend wizard/editor routes changed to `/documents/new` and `/documents/:documentID/edit`
- **Runtime flows:** Novo Documento create still uses Registry atomic create and opens Documents editor
- **Persistence:** current index/trigger names moved from `documents_v2` to `documents`
- **Dependencies:** approval document-scoped generated adapter names updated
- **T-NNN touched:** approval T-008 closed if not already marked closed
- **R-NNN touched:** approval R-008 status updated if applicable
- **Counts after:** use tally gate output
- **Tally gate:** PASS after rerun
- **Patched files:** list patched wiki files
```

- [ ] **Step 8: Run wiki tally gates**

Run:

```powershell
& 'C:\Program Files\Git\bin\bash.exe' .claude/skills/metaldocs-module-doc/scripts/tally_check.sh documents
& 'C:\Program Files\Git\bin\bash.exe' .claude/skills/metaldocs-module-doc/scripts/tally_check.sh novo-documento-wizard
& 'C:\Program Files\Git\bin\bash.exe' .claude/skills/metaldocs-module-doc/scripts/tally_check.sh approval
```

Expected: PASS. If warnings are pre-existing, record them.

---

## Task 7: Cross-Surface No-V2 Current Truth Audit

**Depends on:** Tasks 1-6 complete.

**Parallel-safe:** no. This is an integration gate.

**Files:** read-only unless this audit reveals missed current-truth references.

- [ ] **Step 1: Audit frontend runtime**

Run:

```powershell
rg -n "documents-v2|Documents v2|workspaceView.*documents-v2|/documents-v2" frontend/apps/web/src --glob '!**/node_modules/**'
```

Expected: no hits in current source.

- [ ] **Step 2: Audit backend current runtime**

Run:

```powershell
rg -n "documents_v2|document_v2|DocumentV2|DocumentsV2" apps internal --glob '!**/api.gen.go'
```

Expected:

- No hits in current Documents runtime strings.
- Allowed hits only for `docgen-v2` service names or `TemplatesV2*` storage readers if classified out of scope.

- [ ] **Step 3: Audit generated document symbols**

Run:

```powershell
rg -n "DocumentsV2|DocumentV2|ListDocumentsV2|PublishDocumentV2|RecordDocumentSignoffV2|CancelDocumentApprovalV2|GetApprovalInstanceByDocumentV2" internal/modules/documents internal/modules/documents/approval frontend/apps/web/src/lib/api-types/index.d.ts
```

Expected: no hits for document-owned/document-scoped generated symbols. Non-document approval `V2` symbols may remain only if the regex catches them; classify explicitly.

- [ ] **Step 4: Audit database current truth**

Run:

```powershell
rg -n "ux_documents_v2|trg_documents_v2|document_v2_id" db/baseline/0001_current_schema.sql wiki/database wiki/modules/documents.md wiki/modules/approval.md wiki/modules/approval-tech-debt.md wiki/backlog/approval-refactor.md
```

Expected: no current-truth hits. Historical migration mentions are allowed outside this command.

- [ ] **Step 5: Audit wiki current routes**

Run:

```powershell
rg -n "documents-v2|/api/v2/documents|DocumentV2|DocumentsV2" wiki/modules/documents.md wiki/modules/novo-documento-wizard.md wiki/architecture/frontend-structure.md wiki/architecture/api-contract.md wiki/tests/system-acceptance-test.md wiki/workflows/user-onboarding.md wiki/workflows/approval.md wiki/workflows/freeze-and-fanout.md wiki/backlog/editor.md wiki/backlog/library-screen.md wiki/backlog/documento-publicado.md wiki/backlog/contract-first-followups.md
```

Expected: no current-truth hits. `public.documents_v2` historical notes may remain in `wiki/architecture/data-model.md` and `wiki/README.md`.

---

## Task 8: Full Verification And Runtime Smoke

**Depends on:** Tasks 1-7 complete.

**Parallel-safe:** no.

**Files:** no edits unless verification reveals a defect.

- [ ] **Step 1: Rerun mandatory gates**

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-system-runnable.ps1 -TargetRoute /api/v1/documents
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-module-contract-sync.ps1 -Module documents
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-module-contract-sync.ps1 -Module approval
```

Expected: all pass.

- [ ] **Step 2: Run backend verification**

Run:

```powershell
npx @redocly/cli lint api/openapi/v1/openapi.yaml
$env:GOFLAGS = "-mod=mod"; go generate ./internal/modules/documents/api/...
$env:GOFLAGS = "-mod=mod"; go generate ./internal/modules/documents/approval/api/...
go test ./internal/modules/documents/... -count=1
go test ./apps/api/cmd/metaldocs-api -count=1
```

Expected: exit 0 and no codegen drift after generation.

- [ ] **Step 3: Run DB verification**

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-db-dictionary-coverage.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-bootstrap-baseline.ps1 -WithDevSeed
```

Expected: pass.

- [ ] **Step 4: Run frontend verification**

Run:

```powershell
cmd /c "cd frontend\apps\web && pnpm gen:api"
cmd /c "cd frontend\apps\web && pnpm.cmd tsc --noEmit -p tsconfig.build.json"
cmd /c "cd frontend\apps\web && pnpm test"
```

Expected: pass. If full `pnpm test` is too slow, run the targeted route/wizard/editor tests first, but do not claim completion without either full test pass or an explicit documented limitation.

- [ ] **Step 5: Start runtime using project scripts**

Use the project-supported startup path:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/start-api.ps1 -NoWorker
cmd /c "cd frontend\apps\web && npm run dev"
```

If `scripts/start-api.ps1` fails due pre-existing local process issues, classify as `runtime-contract-prereq` and repair startup truth before continuing.

- [ ] **Step 6: Browser smoke**

Using Chrome CDP or Playwright:

1. Open `http://127.0.0.1:4173/login`.
2. Login with dev admin credentials from `wiki/references/local-dev-credentials.md`.
3. Navigate to `http://127.0.0.1:4173/documents/new`.
4. Create a blank-template document.
5. Confirm the app navigates to `/documents/<document-id>/edit`.
6. Confirm editor shell loads with real document ID.
7. Navigate to `/documents/<document-id>` and confirm the published/detail route is distinct from editor.
8. Navigate to `/documents-v2/new` and `/documents-v2/<document-id>` and confirm they are not supported product routes.

Capture screenshot evidence under `tmp/`.

- [ ] **Step 7: Final changed-file scope summary**

Run:

```powershell
git status --short
git diff --stat
```

Summarize by lane:

- Contract/codegen
- Database
- Frontend
- Backend runtime strings
- Wiki/docs
- Generated files
- Deferred/out-of-scope historical references

---

## Task 9: Implementation Review Handoff

**Depends on:** Task 8 complete.

**Parallel-safe:** no.

- [ ] **Step 1: Request code review**

Use `superpowers:requesting-code-review` with the design spec and this plan as review inputs.

Ask reviewers to focus on:

- no compatibility aliases accidentally added
- route collision between `/documents/:documentId` and `/documents/:documentID/edit`
- OpenAPI/codegen rename completeness
- DB migration idempotency and baseline sync
- no broad purge of unrelated `V2` module/service naming
- wiki current truth vs historical exceptions

- [ ] **Step 2: Address review using receiving-code-review**

If review returns findings, use `superpowers:receiving-code-review` before changing code.

- [ ] **Step 3: Completion report**

Final response must include:

- What changed in contract/API.
- What changed in DB.
- What changed in frontend routes/navigation.
- What changed in backend runtime strings.
- What wiki/docs were synced.
- Deferred/out-of-scope `V2` references with rationale.
- Verification evidence, including browser smoke screenshot paths.
