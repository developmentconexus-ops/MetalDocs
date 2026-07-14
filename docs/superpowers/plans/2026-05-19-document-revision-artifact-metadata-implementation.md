# Document Revision Artifact Metadata Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Persist DOCX artifact metadata (`file_size_bytes`, `page_count`, `page_count_source`) on technical autosave revisions, expose it through the documents API, and render it in the editor sidebar using real runtime data.

**Architecture:** Store artifact metadata on `public.document_revisions`, because it belongs to the saved DOCX artifact row, not to governed revision lineage. Extend autosave commit so the client supplies EigenPal page count through the MetalDocs editor wrapper, while the backend computes server-authoritative size during the same commit transaction and surfaces current-head metadata through document detail and autosave responses.

**Tech Stack:** PostgreSQL forward migration tail, Go backend (`documents` module + OpenAPI contract), `openapi-typescript` generated frontend types, React + TanStack Query, `packages/editor-ui` EigenPal ACL wrapper, Vitest, browser E2E.

---

## File Map

- Create: `db/migrations/0206_document_revision_artifact_metadata.sql`
- Create: `docs/superpowers/plans/2026-05-19-document-revision-artifact-metadata-implementation.md`
- Modify: `wiki/database/tables/document_revisions.md`
- Modify: `wiki/modules/documents.md`
- Modify: `api/openapi/v1/openapi.yaml`
- Modify: `internal/modules/documents/application/service.go`
- Modify: `internal/modules/documents/application/service_test.go`
- Modify: `internal/modules/documents/domain/model.go`
- Modify: `internal/modules/documents/repository/repository.go`
- Modify: `internal/modules/documents/repository/repository_commit_upload_integration_test.go`
- Modify: `internal/modules/documents/delivery/http/handler.go`
- Modify: `internal/modules/documents/delivery/http/handler_test.go`
- Modify: `internal/platform/objectstore/document_presigner.go` or `internal/platform/objectstore/document_presigner_export.go` only if the existing size/stat surface is insufficient
- Modify: `packages/editor-ui/src/types.ts`
- Modify: `packages/editor-ui/src/MetalDocsEditor.tsx`
- Modify: `packages/editor-ui/src/index.ts` only if export surface changes
- Modify: `frontend/apps/web/src/features/documents/api/documents.ts`
- Modify: `frontend/apps/web/src/features/documents/queries/useDocumentDetailQuery.ts` only if cache/write helpers are added
- Modify: `frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx`
- Modify: `frontend/apps/web/src/features/documents/components/EditorMetaSidebar.tsx`
- Modify: `frontend/apps/web/src/features/documents/components/EditorMetaSidebar.test.tsx`
- Modify: `frontend/apps/web/src/features/documents/__tests__/DocumentEditorPage.test.tsx`
- Modify: `frontend/apps/web/src/lib/api-types/index.d.ts` via `pnpm gen:api`

### Task 1: Persist Artifact Metadata In The Database

**Files:**
- Create: `db/migrations/0206_document_revision_artifact_metadata.sql`
- Modify: `wiki/database/tables/document_revisions.md`

- [ ] **Step 1: Write the migration test target and dictionary acceptance criteria**

Use the migration requirements from the approved spec:

```sql
ALTER TABLE public.document_revisions
  ADD COLUMN file_size_bytes bigint,
  ADD COLUMN page_count integer,
  ADD COLUMN page_count_source text;
```

Dictionary page must describe:

- `file_size_bytes`: server-authoritative size of the saved DOCX artifact for this technical revision row
- `page_count`: client- or server-derived rendered page count for this saved technical revision row
- `page_count_source`: provenance enum, initially `eigenpal_client`

- [ ] **Step 2: Add the forward migration**

Create `db/migrations/0206_document_revision_artifact_metadata.sql` with one-shot guarded behavior:

```sql
BEGIN;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM public.schema_migrations
    WHERE version = '0206'
  ) THEN
    ALTER TABLE public.document_revisions
      ADD COLUMN file_size_bytes bigint,
      ADD COLUMN page_count integer,
      ADD COLUMN page_count_source text;

    ALTER TABLE public.document_revisions
      ADD CONSTRAINT document_revisions_file_size_bytes_nonnegative
      CHECK (file_size_bytes IS NULL OR file_size_bytes >= 0);

    ALTER TABLE public.document_revisions
      ADD CONSTRAINT document_revisions_page_count_positive
      CHECK (page_count IS NULL OR page_count > 0);

    ALTER TABLE public.document_revisions
      ADD CONSTRAINT document_revisions_page_count_source_check
      CHECK (
        page_count_source IS NULL
        OR page_count_source IN ('eigenpal_client', 'server_renderer')
      );

    INSERT INTO public.schema_migrations (version, description)
    VALUES ('0206', 'add artifact metadata to document_revisions');
  END IF;
END
$$;

COMMIT;
```

- [ ] **Step 3: Update the dictionary page**

Add the new columns to [document_revisions.md](/C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/wiki/database/tables/document_revisions.md:1) and clarify again that `public.document_revisions` is technical/autosave storage, not governed history.

- [ ] **Step 4: Run DB-specific verification**

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-db-dictionary-coverage.ps1
```

Expected:

- PASS for dictionary coverage
- no missing entry for `public.document_revisions`

- [ ] **Step 5: Commit**

Run:

```bash
git add db/migrations/0206_document_revision_artifact_metadata.sql wiki/database/tables/document_revisions.md
git commit -m "feat(db): persist document revision artifact metadata"
```

### Task 2: Persist And Return Metadata In The Backend Autosave Path

**Files:**
- Modify: `internal/modules/documents/domain/model.go`
- Modify: `internal/modules/documents/application/service.go`
- Modify: `internal/modules/documents/application/service_test.go`
- Modify: `internal/modules/documents/repository/repository.go`
- Modify: `internal/modules/documents/repository/repository_commit_upload_integration_test.go`
- Modify: `internal/platform/objectstore/document_presigner.go` or `internal/platform/objectstore/document_presigner_export.go` only if needed

- [ ] **Step 1: Write the failing backend tests**

Add tests for:

- `CommitAutosave` forwards `page_count` to repository and computes size before commit
- repository `CommitUpload` persists `file_size_bytes`, `page_count`, `page_count_source`
- invalid page count is rejected before insert

Use concrete assertions like:

```go
func TestCommitAutosave_ForwardsArtifactMetadata(t *testing.T) {
    repo := &fakeRepo{}
    presigner := &fakePresigner{hash: strings.Repeat("a", 64), size: 1304}
    svc := application.NewService(repo, presigner, nil, nil)

    _, err := svc.CommitAutosave(ctx, application.CommitAutosaveCmd{
        TenantID:        "tenant_1",
        ActorUserID:     "user_1",
        DocumentID:      "doc_1",
        SessionID:       "sess_1",
        PendingUploadID: "pending_1",
        PageCount:       ptrInt(3),
    })
    if err != nil {
        t.Fatalf("CommitAutosave: %v", err)
    }
    if repo.commitPageCount == nil || *repo.commitPageCount != 3 {
        t.Fatalf("page count not forwarded")
    }
    if repo.commitFileSizeBytes != 1304 {
        t.Fatalf("file size = %d, want 1304", repo.commitFileSizeBytes)
    }
}
```

- [ ] **Step 2: Extend application/service inputs and outputs**

Update `CommitAutosaveCmd` and `CommitResult` aliases/struct plumbing to carry:

```go
type CommitAutosaveCmd struct {
    TenantID, ActorUserID, DocumentID, SessionID, PendingUploadID string
    FormDataSnapshot                                              json.RawMessage
    PageCount                                                     *int
}
```

Repository commit call should receive:

```go
CommitUpload(
    ctx context.Context,
    tenantID, sessionID, userID, docID, pendingID, serverComputedHash string,
    formDataSnapshot []byte,
    fileSizeBytes int64,
    pageCount *int,
    pageCountSource *string,
) (*CommitResult, error)
```

- [ ] **Step 3: Compute size server-side in the service**

Prefer the existing object-store stat surface. If `DocumentPresigner` already has `SizeObject`, call it after successful `HashObject` verification and before repository commit:

```go
fileSizeBytes, err := s.presigner.SizeObject(ctx, meta.StorageKey)
if err != nil {
    return nil, fmt.Errorf("size s3 object: %w", err)
}
```

Validation rule:

```go
if cmd.PageCount != nil && *cmd.PageCount <= 0 {
    return nil, domain.ErrInvalidPageCount
}
```

`page_count_source` for this slice is:

```go
source := "eigenpal_client"
```

- [ ] **Step 4: Persist metadata in the repository transaction**

Update the revision insert in [repository.go](/C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/internal/modules/documents/repository/repository.go:864):

```sql
INSERT INTO document_revisions
  (document_id, parent_revision_id, session_id, storage_key, content_hash, form_data_snapshot, file_size_bytes, page_count, page_count_source)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
RETURNING id::text, revision_num
```

Extend `CommitResult` to include:

```go
type CommitResult struct {
    RevisionID       string
    RevisionNum      int64
    AlreadyConsumed  bool
    FileSizeBytes    *int64
    PageCount        *int
    PageCountSource  *string
}
```

On idempotent replay, fetch and return the existing metadata from `document_revisions`.

- [ ] **Step 5: Run focused backend verification**

Run:

```powershell
go test ./internal/modules/documents/application/... ./internal/modules/documents/repository/... -count=1
```

Expected:

- PASS for autosave service tests
- PASS for repository integration/tests with new metadata columns

- [ ] **Step 6: Commit**

Run:

```bash
git add internal/modules/documents/domain/model.go internal/modules/documents/application/service.go internal/modules/documents/application/service_test.go internal/modules/documents/repository/repository.go internal/modules/documents/repository/repository_commit_upload_integration_test.go internal/platform/objectstore/document_presigner.go internal/platform/objectstore/document_presigner_export.go
git commit -m "feat(documents): persist autosave artifact metadata"
```

### Task 3: Extend The Documents HTTP And OpenAPI Contract

**Files:**
- Modify: `api/openapi/v1/openapi.yaml`
- Modify: `internal/modules/documents/delivery/http/handler.go`
- Modify: `internal/modules/documents/delivery/http/handler_test.go`
- Modify: `internal/modules/documents/api/api.gen.go` via codegen
- Modify: `frontend/apps/web/src/lib/api-types/index.d.ts` via `pnpm gen:api`

- [ ] **Step 1: Write the failing HTTP contract tests**

Add tests in [handler_test.go](/C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/internal/modules/documents/delivery/http/handler_test.go:1) asserting:

- request accepts `page_count`
- response returns `file_size_bytes`, `page_count`, `page_count_source`
- document detail returns current revision artifact metadata

Use assertions like:

```go
if got := out["page_count"]; got != float64(3) {
    t.Fatalf("page_count = %v", got)
}
```

- [ ] **Step 2: Update OpenAPI**

In `api/openapi/v1/openapi.yaml`:

- add `page_count` to `commitDocumentAutosave` request body
- extend its `200` response schema with `file_size_bytes`, `page_count`, `page_count_source`
- extend `DocumentDetailResponse` with:
  - `currentRevisionFileSizeBytes`
  - `currentRevisionPageCount`
  - `currentRevisionPageCountSource`

- [ ] **Step 3: Update the HTTP handler request/response mapping**

In [handler.go](/C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/internal/modules/documents/delivery/http/handler.go:766):

```go
var req struct {
    SessionID        string          `json:"session_id"`
    PendingUploadID  string          `json:"pending_upload_id"`
    FormDataSnapshot json.RawMessage `json:"form_data_snapshot"`
    PageCount        *int            `json:"page_count"`
}
```

Pass `PageCount` into `application.CommitAutosaveCmd`, and write response JSON with the three metadata fields.

Also extend `documentDetailResponse` plus `toDocumentDetailResponse` to map current revision metadata from `domain.Document`.

- [ ] **Step 4: Regenerate contract surfaces and run contract verification**

Run:

```powershell
npx @redocly/cli lint api/openapi/v1/openapi.yaml
$env:GOFLAGS = "-mod=mod"
go generate ./internal/modules/documents/api/...
cd frontend/apps/web
pnpm gen:api
pnpm.cmd tsc --noEmit -p tsconfig.build.json
```

Expected:

- OpenAPI lint passes
- generated backend surface updates cleanly
- generated frontend types include the new fields
- frontend build typecheck passes

- [ ] **Step 5: Commit**

Run:

```bash
git add api/openapi/v1/openapi.yaml internal/modules/documents/delivery/http/handler.go internal/modules/documents/delivery/http/handler_test.go internal/modules/documents/api/api.gen.go frontend/apps/web/src/lib/api-types/index.d.ts
git commit -m "feat(documents): expose artifact metadata in API"
```

### Task 4: Collect Page Count Through The Editor Wrapper And Render It In The Sidebar

**Files:**
- Modify: `packages/editor-ui/src/types.ts`
- Modify: `packages/editor-ui/src/MetalDocsEditor.tsx`
- Modify: `frontend/apps/web/src/features/documents/api/documents.ts`
- Modify: `frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx`
- Modify: `frontend/apps/web/src/features/documents/components/EditorMetaSidebar.tsx`
- Modify: `frontend/apps/web/src/features/documents/components/EditorMetaSidebar.test.tsx`
- Modify: `frontend/apps/web/src/features/documents/__tests__/DocumentEditorPage.test.tsx`

- [ ] **Step 1: Write the failing frontend tests**

Add tests covering:

- `MetalDocsEditorRef.getPageCount()` returns EigenPal total pages or `null`
- `DocumentEditorPage` sends `page_count` in autosave commit
- `EditorMetaSidebar` renders pages and size from real props without showing placeholders when data exists

- [ ] **Step 2: Extend the editor ACL wrapper**

Update [types.ts](/C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/packages/editor-ui/src/types.ts:34):

```ts
export interface MetalDocsEditorRef {
  getDocumentBuffer(): Promise<ArrayBuffer | null>;
  saveNow(): Promise<ArrayBuffer | null>;
  getPageCount(): number | null;
  focus(): void;
}
```

Update [MetalDocsEditor.tsx](/C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/packages/editor-ui/src/MetalDocsEditor.tsx:17):

```ts
getPageCount() {
  if (!inner.current) return null;
  const total = inner.current.getTotalPages();
  return Number.isInteger(total) && total > 0 ? total : null;
},
```

- [ ] **Step 3: Extend the frontend API wrapper and page save path**

Update [documents.ts](/C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/frontend/apps/web/src/features/documents/api/documents.ts:20):

```ts
export type CommitResult = {
  revision_id: string;
  revision_num: number;
  idempotent_replay?: boolean;
  file_size_bytes?: number | null;
  page_count?: number | null;
  page_count_source?: string | null;
};
```

Update `commitAutosave()` request body type to include `page_count?: number`.

In `DocumentEditorPage.tsx`, collect the page count before queue/commit boundary and update local/detail state from autosave response if useful for immediate sidebar freshness.

- [ ] **Step 4: Render the metadata in the sidebar**

Update `EditorMetaSidebar.tsx` props to accept:

```ts
fileSizeBytes?: number | null;
pageCount?: number | null;
```

Render a `Paginas` row in `Identificacao`, for example:

- `3 paginas · 1.3 KB`
- `1 pagina · 1.3 KB`

Use a small local formatter in the documents feature instead of ad hoc string concatenation in JSX.

- [ ] **Step 5: Run frontend verification**

Run:

```powershell
cd frontend/apps/web
pnpm.cmd vitest run src/features/documents/components/EditorMetaSidebar.test.tsx src/features/documents/__tests__/DocumentEditorPage.test.tsx
pnpm.cmd tsc --noEmit -p tsconfig.build.json
```

Expected:

- sidebar tests pass
- editor page tests pass with `page_count` assertions
- frontend typecheck passes

- [ ] **Step 6: Commit**

Run:

```bash
git add packages/editor-ui/src/types.ts packages/editor-ui/src/MetalDocsEditor.tsx frontend/apps/web/src/features/documents/api/documents.ts frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx frontend/apps/web/src/features/documents/components/EditorMetaSidebar.tsx frontend/apps/web/src/features/documents/components/EditorMetaSidebar.test.tsx frontend/apps/web/src/features/documents/__tests__/DocumentEditorPage.test.tsx
git commit -m "feat(editor): show persisted artifact metadata"
```

### Task 5: Sync Wiki Memory And Run End-To-End Verification

**Files:**
- Modify: `wiki/modules/documents.md`
- Modify: `wiki/database/tables/document_revisions.md` if not already complete from Task 1
- Capture evidence under: `docs/superpowers/artifacts/screenshots/`

- [ ] **Step 1: Update module docs**

Document in [documents.md](/C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/wiki/modules/documents.md:1):

- current artifact metadata now comes from `documents.current_revision_id -> document_revisions`
- `page_count_source = eigenpal_client` for this slice
- `document_revisions` remains technical/autosave storage only

- [ ] **Step 2: Run the required verification gates**

Run:

```powershell
go test ./internal/modules/documents/... -count=1
cd frontend/apps/web
pnpm.cmd tsc --noEmit -p tsconfig.build.json
pnpm test
```

Expected:

- documents backend tests pass
- frontend typecheck passes
- relevant Vitest coverage remains green

- [ ] **Step 3: Run browser E2E on the editor**

Using `browser:browser`, verify:

- draft editor shows real `Paginas` + size metadata
- data matches API/runtime truth
- no mocks are used
- governed revision history still shows `REV00`, `REV01`, etc.
- draft still hides approvers
- `under_review` still shows full approval chain
- no `document_revisions` technical rows appear as business history

Capture screenshots for the main evidence states.

- [ ] **Step 4: Commit wiki sync**

Run:

```bash
git add wiki/modules/documents.md wiki/database/tables/document_revisions.md docs/superpowers/artifacts/screenshots
git commit -m "docs(documents): sync artifact metadata runtime truth"
```

## Self-Review

- Spec coverage: DB persistence, autosave efficiency, API contract, wrapper collection, sidebar rendering, provenance, verification, and defer boundaries are all mapped to explicit tasks above.
- Placeholder scan: no `TODO`/`TBD`/“implement later” markers remain.
- Type consistency: canonical field names in the plan are `file_size_bytes`, `page_count`, `page_count_source` on autosave response and `currentRevisionFileSizeBytes`, `currentRevisionPageCount`, `currentRevisionPageCountSource` on detail response.

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-05-19-document-revision-artifact-metadata-implementation.md`. Two execution options:

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

**Which approach?**

