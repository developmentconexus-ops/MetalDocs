# Editor Sidebar Governed Revision Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the editor right sidebar mocks with real governed revision, visibility, and approval-chain data while preserving MetalDocs domain boundaries: `documents` owns governed revision history, `registry` owns visibility, `approval` owns live signoff tracking, and `document_revisions` stays technical/autosave-only.

**Architecture:** Implement this in slices that respect the truth hierarchy. First repair the shared `approval-instance` contract so frontend does not consume runtime-only shapes. Then add governed revision metadata (`revision_title`) and governed revision-history reads in `documents`, normalize registry visibility consumption, and only then wire the sidebar through TanStack Query. Finish with architecture/code-review gating, browser E2E, screenshots, and doc sync.

**Tech Stack:** Go 1.25, Postgres migrations, OpenAPI 3.0.3 + oapi-codegen, openapi-typescript, React, TanStack Query v5, Vitest, PowerShell scripts, Codex Browser.

---

## File Structure

### Backend / database

- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/api/openapi/v1/openapi.yaml`
  - Add/repair `approval-instance`, `finalize`, and `revision-history` contract shapes.
- Create: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/migrations/0200_documents_revision_title.sql`
  - Add `documents.revision_title` with forward-only migration.
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/internal/modules/documents/api/api.gen.go`
  - Regenerated documents contract.
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/internal/modules/documents/approval/api/api.gen.go`
  - Regenerated approval contract if approval-instance tag coverage is adjusted.
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/internal/modules/documents/delivery/http/handler.go`
  - Parse `revisionTitle` on finalize, expose `revision-history` endpoint, include `RevisionTitle` in detail response.
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/internal/modules/documents/application/service.go`
  - Governed revision-history service entrypoint if needed.
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/internal/modules/documents/application/ports.go`
  - Add repo contract for revision history if needed.
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/internal/modules/documents/repository/repository.go`
  - Persist `revision_title`; query governed revision history by `controlled_document_id`.
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/internal/modules/documents/approval/http/contracts/instance_read.go`
  - If needed before codegen convergence, align with real approval-instance payload semantics.
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/internal/modules/documents/approval/http/get_instance_handler.go`
  - Keep mapper aligned with contract/runtime truth.

### Frontend

- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/frontend/apps/web/src/lib/api-types/index.d.ts`
  - Regenerated frontend types.
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/frontend/apps/web/src/lib/queryKeys.ts`
  - Add `documents.revisionHistory(id)`.
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/frontend/apps/web/src/features/documents/api/documents.ts`
  - Use generated approval-instance/document/revision-history types; send `revisionTitle` on finalize.
- Create: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/frontend/apps/web/src/features/documents/queries/useDocumentRevisionHistoryQuery.ts`
  - Query hook for governed revision history.
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/frontend/apps/web/src/features/documents/queries/useApprovalInstanceQuery.ts`
  - Shift to generated types after contract repair.
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/frontend/apps/web/src/features/registry/types.ts`
  - Remove legacy omission of `visibility` or replace with generated type use.
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/frontend/apps/web/src/features/registry/api/controlledDocuments.ts`
  - Normalize controlled-document detail to generated contract.
- Create: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/frontend/apps/web/src/features/registry/queries/useControlledDocumentDetailQuery.ts`
  - Dedicated detail query for visibility if not already present.
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/frontend/apps/web/src/features/documents/components/EditorMetaSidebar.tsx`
  - Replace mocks with real data and status-gated rendering.
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx`
  - Fetch and pass real sidebar data; finalize dialog/body includes `revisionTitle`.
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/frontend/apps/web/src/features/documents/lib/documentDetailMeta.ts`
  - Add governed revision formatting helpers like `formatRevisionCode` and sidebar display helpers.

### Tests

- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/internal/modules/documents/repository/repository_commit_upload_integration_test.go`
  - Ensure autosave does not become sidebar history source.
- Create: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/internal/modules/documents/repository/repository_revision_history_integration_test.go`
  - Governed revision-history query coverage.
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/internal/modules/documents/delivery/http/handler_test.go`
  - Finalize validation + revision-history route tests.
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/frontend/apps/web/src/features/documents/__tests__/documents.test.ts`
  - Generated wrapper coverage for finalize / approval-instance / revision-history.
- Create: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/frontend/apps/web/src/features/documents/queries/__tests__/useDocumentRevisionHistoryQuery.test.tsx`
  - Query hook coverage.
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/frontend/apps/web/src/features/documents/pages/DocumentEditorPage.test.tsx`
  - Sidebar truth-by-status coverage.
- Create: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/frontend/apps/web/src/features/documents/components/EditorMetaSidebar.test.tsx`
  - Pure/sidebar rendering semantics if page-level tests become noisy.

### Docs

- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/wiki/modules/documents.md`
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/wiki/modules/documents-tech-debt.md`
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/wiki/modules/approval.md`
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/wiki/modules/registry.md`
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/wiki/backlog/editor.md`
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/frontend/apps/web/design-source/editor/NOTES.md`

---

### Task 1: Repair `approval-instance` Shared Contract

**Files:**
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/api/openapi/v1/openapi.yaml`
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/internal/modules/documents/approval/http/contracts/instance_read.go`
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/internal/modules/documents/approval/http/get_instance_handler.go`
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/internal/modules/documents/approval/api/api.gen.go`
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/frontend/apps/web/src/lib/api-types/index.d.ts`
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/frontend/apps/web/src/features/documents/api/documents.ts`
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/frontend/apps/web/src/features/documents/queries/useApprovalInstanceQuery.ts`
- Test: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/frontend/apps/web/src/features/documents/__tests__/documents.test.ts`

- [ ] **Step 1: Write the failing frontend wrapper test for typed approval-instance payload**

```ts
it('reads approval-instance with runtime-aligned statuses and signoff payload', async () => {
  vi.spyOn(global, 'fetch').mockResolvedValue(
    new Response(JSON.stringify({
      id: 'inst-1',
      document_id: 'doc-1',
      route_id: 'route-1',
      tenant_id: 'tenant-1',
      status: 'in_progress',
      submitted_by: 'user-1',
      submitted_at: '2026-05-18T12:00:00Z',
      stages: [
        {
          id: 'stage-1',
          stage_index: 1,
          label: 'Qualidade',
          status: 'active',
          signoffs: [],
        },
        {
          id: 'stage-2',
          stage_index: 2,
          label: 'Diretoria',
          status: 'pending',
          signoffs: [
            {
              id: 'signoff-1',
              actor_user_id: 'Maria Souza',
              decision: 'approve',
              signature_method: 'password_reauth',
              signed_at: '2026-05-18T12:30:00Z',
            },
          ],
        },
      ],
      etag: '\"v1\"',
    }), { status: 200, headers: { 'Content-Type': 'application/json' } }),
  as any);

  const result = await getApprovalInstance('doc-1');
  expect(result.status).toBe('in_progress');
  expect(result.stages[0]?.status).toBe('active');
  expect(result.stages[1]?.signoffs[0]?.actor_user_id).toBe('Maria Souza');
});
```

- [ ] **Step 2: Run the failing test**

Run: `cd frontend/apps/web; pnpm.cmd vitest run src/features/documents/__tests__/documents.test.ts -t "reads approval-instance with runtime-aligned statuses and signoff payload"`
Expected: FAIL because current generated/frontend wrapper contract does not type the `200` payload correctly or manual types drift from runtime statuses.

- [ ] **Step 3: Add the OpenAPI response schema for `GET /api/v1/documents/{id}/approval-instance`**

```yaml
/api/v1/documents/{id}/approval-instance:
  get:
    tags: [approval]
    operationId: getApprovalInstanceByDocument
    summary: Get active approval instance for a document
    parameters:
      - in: path
        name: id
        required: true
        schema: { type: string, format: uuid }
    responses:
      '200':
        description: ok
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ApprovalInstanceByDocumentResponse'
      '404':
        description: no_active_instance
        content:
          application/problem+json:
            schema:
              $ref: '#/components/schemas/Problem'
```

- [ ] **Step 4: Regenerate contracts and inspect generated shapes**

Run:

```powershell
npx @redocly/cli lint api/openapi/v1/openapi.yaml
$env:GOFLAGS = '-mod=mod'
go generate ./internal/modules/documents/approval/api/...
cd frontend/apps/web
pnpm gen:api
```

Expected: PASS; generated files change in approval API package and frontend `index.d.ts` now includes a typed `200` response for `getApprovalInstanceByDocument`.

- [ ] **Step 5: Update frontend wrapper and query hook to generated types**

```ts
export type ApprovalInstanceResponse =
  components['schemas']['ApprovalInstanceByDocumentResponse'];

export async function getApprovalInstance(documentId: string): Promise<ApprovalInstanceResponse> {
  return apiFetch<ApprovalInstanceResponse>(`/api/v1/documents/${documentId}/approval-instance`);
}
```

- [ ] **Step 6: Re-run focused verification**

Run:

```powershell
cd frontend/apps/web
pnpm.cmd vitest run src/features/documents/__tests__/documents.test.ts -t "reads approval-instance with runtime-aligned statuses and signoff payload"
pnpm.cmd tsc --noEmit -p tsconfig.build.json
cd ../..
go test ./internal/modules/documents/approval/http ./internal/modules/documents/approval/api -count=1
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add api/openapi/v1/openapi.yaml internal/modules/documents/approval/api/api.gen.go internal/modules/documents/approval/http/contracts/instance_read.go internal/modules/documents/approval/http/get_instance_handler.go frontend/apps/web/src/lib/api-types/index.d.ts frontend/apps/web/src/features/documents/api/documents.ts frontend/apps/web/src/features/documents/queries/useApprovalInstanceQuery.ts frontend/apps/web/src/features/documents/__tests__/documents.test.ts
git commit -m "fix(approval): align approval-instance contract for sidebar consumers"
```

### Task 2: Add `documents.revision_title` Database Foundation

**Files:**
- Create: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/migrations/0200_documents_revision_title.sql`
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/wiki/modules/documents.md`
- Test: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/internal/modules/documents/repository/repository_revision_history_integration_test.go`

- [ ] **Step 1: Write the failing integration test asserting governed revision title persists on documents rows**

```go
func TestGovernedRevisionTitleColumnExistsAndCanBeRead(t *testing.T) {
    db := openTestDB(t)
    ctx := context.Background()

    var title sql.NullString
    err := db.QueryRowContext(ctx, `
        SELECT revision_title
        FROM documents
        WHERE id = $1::uuid
    `, mustSeedDocumentWithRevisionTitle(t, db, "Correção de procedimento")).Scan(&title)
    if err != nil {
        t.Fatalf("scan revision_title: %v", err)
    }
    if !title.Valid || title.String != "Correção de procedimento" {
        t.Fatalf("revision_title = %#v", title)
    }
}
```

- [ ] **Step 2: Run the failing test**

Run: `go test ./internal/modules/documents/repository -run TestGovernedRevisionTitleColumnExistsAndCanBeRead -count=1`
Expected: FAIL because `revision_title` column does not exist yet.

- [ ] **Step 3: Add the forward-only migration**

```sql
ALTER TABLE public.documents
  ADD COLUMN IF NOT EXISTS revision_title text;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0200', 'add documents.revision_title for governed revision metadata')
ON CONFLICT (version) DO NOTHING;
```

- [ ] **Step 4: Apply DB verification gates**

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-db-dictionary-coverage.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-system-runnable.ps1 -TargetRoute /api/v1/documents
```

Expected: PASS; no baseline or startup contradiction introduced.

- [ ] **Step 5: Re-run the integration test**

Run: `go test ./internal/modules/documents/repository -run TestGovernedRevisionTitleColumnExistsAndCanBeRead -count=1`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add migrations/0200_documents_revision_title.sql wiki/modules/documents.md internal/modules/documents/repository/repository_revision_history_integration_test.go
git commit -m "feat(documents): add governed revision title column"
```

### Task 3: Extend Finalize To Require And Persist `revisionTitle`

**Files:**
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/api/openapi/v1/openapi.yaml`
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/internal/modules/documents/api/api.gen.go`
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/internal/modules/documents/delivery/http/handler.go`
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/internal/modules/documents/repository/repository.go`
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/frontend/apps/web/src/lib/api-types/index.d.ts`
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/frontend/apps/web/src/features/documents/api/documents.ts`
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx`
- Test: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/internal/modules/documents/delivery/http/handler_test.go`
- Test: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/frontend/apps/web/src/features/documents/__tests__/documents.test.ts`

- [ ] **Step 1: Write the failing backend test for finalize without `revisionTitle`**

```go
func TestFinalizeDocument_RequiresRevisionTitle(t *testing.T) {
    req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc-1/finalize", strings.NewReader(`{}`))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Idempotency-Key", "11111111-1111-4111-8111-111111111111")
    rr := httptest.NewRecorder()

    newTestHandler(t).ServeHTTP(rr, req)

    if rr.Code != http.StatusBadRequest {
        t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
    }
}
```

- [ ] **Step 2: Write the failing frontend wrapper test for finalize body**

```ts
it('sends revisionTitle when finalizing document', async () => {
  vi.spyOn(globalThis.crypto, 'randomUUID').mockReturnValue('11111111-1111-4111-8111-111111111111');
  const fetchSpy = vi.spyOn(global, 'fetch').mockResolvedValue(
    new Response(JSON.stringify({ instanceId: 'inst-1' }), { status: 201, headers: { 'Content-Type': 'application/json' } }),
  as any);

  await finalizeDocument('doc-1', { revisionTitle: 'Ajuste de escopo operacional' });

  const [, init] = fetchSpy.mock.calls[0] ?? [];
  expect(JSON.parse(String(init?.body))).toEqual({ revisionTitle: 'Ajuste de escopo operacional' });
});
```

- [ ] **Step 3: Run the failing tests**

Run:

```powershell
go test ./internal/modules/documents/delivery/http -run TestFinalizeDocument_RequiresRevisionTitle -count=1
cd frontend/apps/web
pnpm.cmd vitest run src/features/documents/__tests__/documents.test.ts -t "sends revisionTitle when finalizing document"
```

Expected: FAIL.

- [ ] **Step 4: Extend contract, handler, and wrapper**

```yaml
requestBody:
  required: true
  content:
    application/json:
      schema:
        type: object
        required: [revisionTitle]
        properties:
          revisionTitle:
            type: string
            minLength: 1
            maxLength: 200
```

```ts
export async function finalizeDocument(id: string, body: { revisionTitle: string }): Promise<FinalizeDocumentResult> {
  return apiFetch<FinalizeDocumentResult>(`/api/v1/documents/${id}/finalize`, {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify(body),
    idempotencyKey: crypto.randomUUID(),
  });
}
```

- [ ] **Step 5: Re-run focused verification**

Run:

```powershell
npx @redocly/cli lint api/openapi/v1/openapi.yaml
$env:GOFLAGS = '-mod=mod'
go generate ./internal/modules/documents/api/...
go test ./internal/modules/documents/delivery/http ./internal/modules/documents/repository -count=1
cd frontend/apps/web
pnpm gen:api
pnpm.cmd vitest run src/features/documents/__tests__/documents.test.ts
pnpm.cmd tsc --noEmit -p tsconfig.build.json
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add api/openapi/v1/openapi.yaml internal/modules/documents/api/api.gen.go internal/modules/documents/delivery/http/handler.go internal/modules/documents/repository/repository.go frontend/apps/web/src/lib/api-types/index.d.ts frontend/apps/web/src/features/documents/api/documents.ts frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx internal/modules/documents/delivery/http/handler_test.go frontend/apps/web/src/features/documents/__tests__/documents.test.ts
git commit -m "feat(documents): require governed revision title on finalize"
```

### Task 4: Add Governed Revision-History Read Model

**Files:**
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/api/openapi/v1/openapi.yaml`
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/internal/modules/documents/api/api.gen.go`
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/internal/modules/documents/application/ports.go`
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/internal/modules/documents/application/service.go`
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/internal/modules/documents/repository/repository.go`
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/internal/modules/documents/delivery/http/handler.go`
- Create: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/internal/modules/documents/repository/repository_revision_history_integration_test.go`
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/internal/modules/documents/delivery/http/handler_test.go`

- [ ] **Step 1: Write the failing integration test asserting governed history uses `documents`, not `document_revisions`**

```go
func TestListRevisionHistory_ReturnsGovernedDocumentsNotAutosaveRows(t *testing.T) {
    repo, db := openRepositoryTestHarness(t)
    ctx := context.Background()

    docID := seedControlledDocumentLineage(t, db, []seedGovernedRevision{
        {DocumentID: "doc-1", RevisionNumber: 1, RevisionTitle: "Primeira versão", Status: "published"},
        {DocumentID: "doc-2", RevisionNumber: 2, RevisionTitle: "Ajuste operacional", Status: "draft"},
    })
    seedTechnicalAutosaveRows(t, db, "doc-2", 3)

    items, err := repo.ListRevisionHistory(ctx, tenantID, docID)
    if err != nil {
        t.Fatalf("ListRevisionHistory: %v", err)
    }
    if len(items) != 2 {
        t.Fatalf("len(items) = %d, want 2", len(items))
    }
}
```

- [ ] **Step 2: Run the failing test**

Run: `go test ./internal/modules/documents/repository -run TestListRevisionHistory_ReturnsGovernedDocumentsNotAutosaveRows -count=1`
Expected: FAIL because no governed history API exists yet.

- [ ] **Step 3: Add contract and repository query**

```yaml
/api/v1/documents/{id}/revision-history:
  get:
    tags: [documents]
    operationId: getDocumentRevisionHistory
    responses:
      '200':
        description: ok
        content:
          application/json:
            schema:
              type: object
              required: [items]
              properties:
                items:
                  type: array
                  items:
                    $ref: '#/components/schemas/DocumentRevisionHistoryItem'
```

```go
const q = `
WITH anchor AS (
  SELECT controlled_document_id, id AS current_document_id
  FROM documents
  WHERE tenant_id = $1::uuid AND id = $2::uuid
)
SELECT d.id::text, d.revision_number, COALESCE(d.revision_title, ''), d.status,
       d.created_at, (d.id = a.current_document_id) AS is_current
FROM anchor a
JOIN documents d ON d.controlled_document_id = a.controlled_document_id
ORDER BY d.revision_number DESC
`
```

- [ ] **Step 4: Expose service and HTTP handler**

```go
type RevisionHistoryItem struct {
    DocumentID string
    RevisionVersion int
    RevisionTitle string
    Status string
    CreatedAt time.Time
    IsCurrent bool
}
```

- [ ] **Step 5: Re-run focused verification**

Run:

```powershell
$env:GOFLAGS = '-mod=mod'
go generate ./internal/modules/documents/api/...
go test ./internal/modules/documents/repository ./internal/modules/documents/delivery/http -count=1
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add api/openapi/v1/openapi.yaml internal/modules/documents/api/api.gen.go internal/modules/documents/application/ports.go internal/modules/documents/application/service.go internal/modules/documents/repository/repository.go internal/modules/documents/delivery/http/handler.go internal/modules/documents/repository/repository_revision_history_integration_test.go internal/modules/documents/delivery/http/handler_test.go
git commit -m "feat(documents): add governed revision history read model"
```

### Task 5: Normalize Registry Visibility Consumption

**Files:**
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/frontend/apps/web/src/features/registry/types.ts`
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/frontend/apps/web/src/features/registry/api/controlledDocuments.ts`
- Create: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/frontend/apps/web/src/features/registry/queries/useControlledDocumentDetailQuery.ts`
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/frontend/apps/web/src/lib/queryKeys.ts`

- [ ] **Step 1: Write the failing test asserting controlled-document detail exposes visibility**

```ts
it('reads controlled document detail including visibility', async () => {
  vi.spyOn(global, 'fetch').mockResolvedValue(
    new Response(JSON.stringify({
      id: 'cd-1',
      tenantId: 'tenant-1',
      profileCode: 'POP',
      processAreaCode: 'RH',
      code: 'POP-RH-001',
      title: 'Procedimento RH',
      ownerUserId: 'user-1',
      status: 'active',
      visibility: { scope: 'restricted', areaCodes: ['RH'], userIds: [] },
      createdAt: '2026-05-18T12:00:00Z',
      updatedAt: '2026-05-18T12:00:00Z',
    }), { status: 200, headers: { 'Content-Type': 'application/json' } }),
  as any);

  const result = await fetchControlledDocument('cd-1');
  expect(result.visibility.scope).toBe('restricted');
});
```

- [ ] **Step 2: Run the failing test**

Run: `cd frontend/apps/web; pnpm.cmd vitest run src/features/registry/__tests__/controlledDocuments.test.ts -t "reads controlled document detail including visibility"`
Expected: FAIL because current legacy type omits `visibility`.

- [ ] **Step 3: Normalize the wrapper to generated contract and add detail query hook**

```ts
export type ControlledDocument = components['schemas']['ControlledDocumentResponse'];
```

```ts
export function useControlledDocumentDetailQuery(id: string) {
  return useQuery({
    queryKey: QK.controlledDocuments.detail(id),
    queryFn: () => fetchControlledDocument(id),
    enabled: Boolean(id),
    staleTime: 5 * 60 * 1000,
  });
}
```

- [ ] **Step 4: Re-run focused verification**

Run:

```powershell
cd frontend/apps/web
pnpm.cmd vitest run src/features/registry/__tests__/controlledDocuments.test.ts
pnpm.cmd tsc --noEmit -p tsconfig.build.json
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/apps/web/src/features/registry/types.ts frontend/apps/web/src/features/registry/api/controlledDocuments.ts frontend/apps/web/src/features/registry/queries/useControlledDocumentDetailQuery.ts frontend/apps/web/src/lib/queryKeys.ts frontend/apps/web/src/features/registry/__tests__/controlledDocuments.test.ts
git commit -m "refactor(registry): expose controlled document visibility to sidebar"
```

### Task 6: Wire Governed Sidebar Data In The Editor

**Files:**
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/frontend/apps/web/src/features/documents/components/EditorMetaSidebar.tsx`
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx`
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/frontend/apps/web/src/features/documents/lib/documentDetailMeta.ts`
- Create: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/frontend/apps/web/src/features/documents/queries/useDocumentRevisionHistoryQuery.ts`
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/frontend/apps/web/src/features/documents/api/documents.ts`

- [ ] **Step 1: Write the failing sidebar rendering test for real metadata and governed history**

```tsx
it('renders real governed metadata and hides approval chain in draft', async () => {
  render(
    <EditorMetaSidebar
      open
      onToggle={() => {}}
      code="POP-RH-001"
      profileLabel="Procedimento Operacional"
      areaLabel="Recursos Humanos"
      visibilityLabel="Restrito à área RH"
      history={[
        { documentId: 'doc-2', revisionVersion: 2, revisionCode: 'REV01', revisionTitle: 'Ajuste operacional', status: 'draft', createdAt: '2026-05-18T12:00:00Z', isCurrent: true },
        { documentId: 'doc-1', revisionVersion: 1, revisionCode: 'REV00', revisionTitle: 'Primeira versão', status: 'published', createdAt: '2026-05-10T12:00:00Z', isCurrent: false },
      ]}
      approvalChain={null}
      documentStatus="draft"
    />,
  );

  expect(screen.getByText('Procedimento Operacional')).toBeInTheDocument();
  expect(screen.getByText('Recursos Humanos')).toBeInTheDocument();
  expect(screen.getByText('Restrito à área RH')).toBeInTheDocument();
  expect(screen.getByText('REV01')).toBeInTheDocument();
  expect(screen.getByText('Ajuste operacional')).toBeInTheDocument();
  expect(screen.queryByText('Próximos aprovadores')).not.toBeInTheDocument();
});
```

- [ ] **Step 2: Run the failing test**

Run: `cd frontend/apps/web; pnpm.cmd vitest run src/features/documents/components/EditorMetaSidebar.test.tsx -t "renders real governed metadata and hides approval chain in draft"`
Expected: FAIL because sidebar props and rendering are still mock-based.

- [ ] **Step 3: Add revision-history query hook and helpers**

```ts
export function formatRevisionCode(revisionVersion: number): string {
  return `REV${String(Math.max(revisionVersion - 1, 0)).padStart(2, '0')}`;
}
```

```ts
export function useDocumentRevisionHistoryQuery(id: string) {
  return useQuery({
    queryKey: QK.documents.revisionHistory(id),
    queryFn: () => getDocumentRevisionHistory(id),
    enabled: Boolean(id),
  });
}
```

- [ ] **Step 4: Replace sidebar mocks with real props and status gates**

```tsx
{documentStatus === 'under_review' && approvalChain ? (
  <section className={styles.section}>
    <div className={styles.sectionHeader}>Próximos aprovadores</div>
    {approvalChain.stages.map((stage) => (
      <div key={stage.id} className={styles.approverRow}>
        <span className={styles.approverName}>{stage.label}</span>
        <span className={styles.approverRole}>{stage.status}</span>
      </div>
    ))}
  </section>
) : null}
```

- [ ] **Step 5: Re-run focused verification**

Run:

```powershell
cd frontend/apps/web
pnpm.cmd vitest run src/features/documents/components/EditorMetaSidebar.test.tsx src/features/documents/pages/DocumentEditorPage.test.tsx
pnpm.cmd tsc --noEmit -p tsconfig.build.json
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add frontend/apps/web/src/features/documents/components/EditorMetaSidebar.tsx frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx frontend/apps/web/src/features/documents/lib/documentDetailMeta.ts frontend/apps/web/src/features/documents/queries/useDocumentRevisionHistoryQuery.ts frontend/apps/web/src/features/documents/api/documents.ts frontend/apps/web/src/features/documents/pages/DocumentEditorPage.test.tsx frontend/apps/web/src/features/documents/components/EditorMetaSidebar.test.tsx frontend/apps/web/src/lib/queryKeys.ts
git commit -m "feat(documents): wire governed editor sidebar data"
```

### Task 7: Sync Module Docs And Screen Memory

**Files:**
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/wiki/modules/documents.md`
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/wiki/modules/documents-tech-debt.md`
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/wiki/modules/approval.md`
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/wiki/modules/registry.md`
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/wiki/backlog/editor.md`
- Modify: `C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/frontend/apps/web/design-source/editor/NOTES.md`

- [ ] **Step 1: Update module docs with concrete implementation truth**

```md
- Documents now distinguish governed revision metadata (`documents.revision_title`, `documents.revision_number`) from technical autosave lineage in `document_revisions`.
- Editor sidebar consumes governed revision history from `GET /api/v1/documents/{id}/revision-history`.
- Approval chain on the editor sidebar is sourced from the repaired `approval-instance` contract and is only rendered for `under_review`.
- Controlled-document visibility is resolved through registry contract truth, not a documents-local duplicate field.
```

- [ ] **Step 2: Run docs hygiene check**

Run: `git diff --check`
Expected: PASS.

- [ ] **Step 3: Commit**

```bash
git add wiki/modules/documents.md wiki/modules/documents-tech-debt.md wiki/modules/approval.md wiki/modules/registry.md wiki/backlog/editor.md frontend/apps/web/design-source/editor/NOTES.md
git commit -m "docs(editor): sync governed sidebar implementation"
```

### Task 8: Architecture/Rules Review Gate With Subagent

**Files:**
- Review only; no product file required unless findings produce fixes.

- [ ] **Step 1: Dispatch review subagent against implementation diff and architecture rules**

Use a fresh review subagent with this prompt:

```text
Run a professional code review of the editor governed sidebar implementation.
Compare the implementation against:
- docs/superpowers/specs/2026-05-18-editor-sidebar-governed-revision-design.md
- AGENTS.md rules
- wiki/architecture/backend-api-structure.md
- wiki/architecture/api-contract.md
- wiki/architecture/api-design-system.md
- wiki/architecture/frontend-structure.md
- wiki/modules/documents.md
- wiki/modules/approval.md
- wiki/modules/registry.md
- wiki/database/tables/document_revisions.md
Focus on bugs, contract drift, architectural violations, wrong domain boundaries, missing tests, and any place where technical `document_revisions` leaked into governed revision history.
Ignore the right-sidebar visual style unless it causes semantic defects.
Return findings ordered by severity with file references.
```

- [ ] **Step 2: If findings exist, fix them in new commits before proceeding**

Run: address each finding with the same discipline as earlier tasks.
Expected: no unresolved architecture or contract findings remain.

- [ ] **Step 3: Re-run the review subagent until green**

Expected: final review returns no material issues in the agreed scope.

- [ ] **Step 4: Commit any review-driven fixes**

```bash
git add <affected files>
git commit -m "fix(editor): address governed sidebar review findings"
```

### Task 9: Full Browser E2E Validation Loop

**Files:**
- Evidence only unless bugs are found; screenshots captured through Browser workflow.

- [ ] **Step 1: Start from a clean locally running app/API**

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-system-runnable.ps1 -TargetRoute /api/v1/documents
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-module-contract-sync.ps1 -Module documents
```

Expected: PASS.

- [ ] **Step 2: Browser flow A — draft editor sidebar truth**

Use Browser to:
- open a draft document editor
- expand the right sidebar
- verify:
  - real `Código`
  - real `Perfil`
  - real `Área`
  - real `Visibilidade`
  - governed history entries show `REV00/REV01...`
  - current draft row shows `Draft`
  - approval chain section is absent
- capture screenshot of expanded sidebar

Expected: all values come from real API-backed data; no mock text remains.

- [ ] **Step 3: Browser flow B — finalize requires revision title**

Use Browser to:
- type content in a draft
- trigger submit/finalize without a revision title
- verify validation prevents submission
- provide `revisionTitle`
- submit successfully
- capture screenshot of validation state and successful transition

Expected: submission is blocked without title and succeeds with title.

- [ ] **Step 4: Browser flow C — under review approval chain truth**

Use Browser to:
- reopen the document now in `under_review`
- expand sidebar
- verify approval chain section is present
- verify the full chain renders: prior/completed, active, waiting
- verify names/statuses align with API response, not placeholder text
- capture screenshot of the approval chain

Expected: sidebar reflects the real approval instance and full chain semantics.

- [ ] **Step 5: Browser flow D — published lineage continuity**

Use Browser and/or seeded data to inspect a lineage where at least one published revision and one newer draft exist.
Verify:
- published revision remains `REV00 · Publicada` (or equivalent current published number)
- newer draft appears as `REV01 · Draft`
- history ordering is newest first
- titles/reasons match governed revision titles
- no autosave-level noise appears
- capture screenshot

Expected: governed history semantics match the spec exactly.

- [ ] **Step 6: Browser/API spot-checks**

Use browser devtools/network or equivalent Browser inspection to verify:
- `GET /api/v1/documents/{id}` returns `RevisionTitle`
- `GET /api/v1/documents/{id}/revision-history` is called and returns governed rows
- `GET /api/v1/documents/{id}/approval-instance` returns typed payload used by UI
- no sidebar section is populated from mocks

Expected: runtime/API behavior matches the contract and UI.

- [ ] **Step 7: If E2E finds defects, loop back to the relevant task slice**

Rule:
- do not stop at “mostly works”
- fix, retest, recommit, and rerun Browser validation until screenshots and runtime evidence match the spec 100%

- [ ] **Step 8: Final verification command bundle**

Run:

```powershell
npx @redocly/cli lint api/openapi/v1/openapi.yaml
$env:GOFLAGS = '-mod=mod'
go build ./...
go test ./internal/modules/documents/... ./internal/modules/documents/approval/... ./internal/modules/registry/... -count=1
cd frontend/apps/web
pnpm gen:api
pnpm.cmd tsc --noEmit -p tsconfig.build.json
pnpm test -- src/features/documents/__tests__/documents.test.ts src/features/documents/queries/__tests__/useDocumentRevisionHistoryQuery.test.tsx src/features/documents/components/EditorMetaSidebar.test.tsx src/features/documents/pages/DocumentEditorPage.test.tsx src/features/registry/__tests__/controlledDocuments.test.ts
```

Expected: PASS.

- [ ] **Step 9: Final commit for any last-mile fixes from E2E**

```bash
git add <affected files>
git commit -m "fix(editor): close governed sidebar e2e validation gaps"
```

## Spec Coverage Check

- Governed revision vs technical revision: Tasks 2, 3, 4, 6, 8, 9
- `revision_title` on `documents`: Tasks 2 and 3
- `revision_title` required at finalize and frozen there: Task 3
- History sourced from governed `documents` lineage: Task 4
- Approval chain shown as full chain, only when real: Tasks 1 and 6
- Visibility from registry contract: Task 5
- Taxonomy code -> name resolution in sidebar: Task 6
- Shared contract prerequisite for `approval-instance`: Task 1
- E2E proof with screenshots and API/runtime confirmation: Task 9
- Module/doc sync after implementation: Task 7
- Review subagent loop against skills/rules/architecture: Task 8

## Placeholder Scan

- No `TODO`, `TBD`, “similar to”, or “implement later” placeholders remain in the tasks.
- Every task has exact files, exact commands, and concrete expected behavior.

## Type Consistency Check

- Governed revision title is consistently named `revision_title` in DB, `RevisionTitle` in backend document detail payload, and `revisionTitle` in frontend request/response usage.
- Governed history is consistently named `revision-history` as a documents read model.
- Approval-instance statuses are consistently aligned to runtime values `in_progress`, `active`, `passed`, `failed`, `cancelled` after Task 1.
