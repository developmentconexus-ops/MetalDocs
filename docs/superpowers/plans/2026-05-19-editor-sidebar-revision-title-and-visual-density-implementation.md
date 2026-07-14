# Editor Sidebar Revision Title and Visual Density Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `REV00` use an automatic governed title and improve the editor sidebar revision-history presentation/density without weakening later-revision title governance.

**Architecture:** Backend `documents` remains authoritative for the `REV00` default title and conditional finalize validation. Frontend mirrors the `REV00` rule only to skip unnecessary modal friction, while the sidebar renders governed history as revision code, title, and date with compact collapsible presentation. No data is sourced from `document_revisions` for business history.

**Tech Stack:** Go 1.25, Postgres-backed documents repository, React, TanStack Query v5, CSS Modules, Vitest, PowerShell verification, Codex Browser.

---

## File Structure

- Modify: `internal/modules/documents/approval/application/submit_service.go`
  - Apply canonical `REV00` default title at finalize/submission when `revision_number = 0`.
  - Preserve required title validation for `revision_number > 0`.
- Modify: `internal/modules/documents/approval/application/submit_service_test.go`
  - Cover `REV00` default title and `REV01+` required-title behavior.
- Modify: `internal/modules/documents/delivery/http/handler_test.go`
  - Ensure the HTTP finalize surface accepts missing/blank title only for the first governed revision through service behavior.
- Modify: `frontend/apps/web/src/features/documents/lib/documentDetailMeta.ts`
  - Add small helpers/constants for default revision title and revision-history row projection if useful.
- Modify: `frontend/apps/web/src/features/documents/lib/documentDetailMeta.test.ts`
  - Cover revision row text/title fallback and no status text.
- Modify: `frontend/apps/web/src/features/documents/components/EditorMetaSidebar.tsx`
  - Render revision rows as `REVxx`, title, and date.
  - Add collapse/expand for histories longer than 3 items.
  - Keep approval section gated to `under_review`.
- Modify: `frontend/apps/web/src/features/documents/components/EditorMetaSidebar.module.css`
  - Improve sidebar visual density, section rhythm, revision cards, and blank-area appearance.
- Modify: `frontend/apps/web/src/features/documents/components/EditorMetaSidebar.test.tsx`
  - Cover status-free revision rows, draft no-approvers, and collapse/expand threshold.
- Modify: `frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx`
  - Skip revision-title dialog for `REV00` and submit with no user-entered title.
  - Keep title dialog for `REV01+`.
- Modify: `frontend/apps/web/src/features/documents/pages/DocumentEditorPage.test.tsx`
  - Cover `REV00` no manual title path and `REV01+` required title path.
- Modify: `wiki/modules/documents.md`
  - Sync the changed `REV00` title semantics and sidebar row semantics after implementation.
- Modify: `wiki/modules/documents-tech-debt.md`
  - Record no new debt or explicitly defer rich long-history filtering if needed.

---

### Task 1: Backend `REV00` Default Title

**Files:**
- Modify: `internal/modules/documents/approval/application/submit_service.go`
- Modify: `internal/modules/documents/approval/application/submit_service_test.go`

- [ ] **Step 1: Write failing service tests**

Add or update tests in `submit_service_test.go` so the service behavior is explicit:

```go
func TestSubmitForApproval_DefaultsRevisionTitleForFirstGovernedRevision(t *testing.T) {
    svc, repo := newSubmitServiceTestHarness(t)
    repo.document = documentForSubmit{
        ID:             "doc-1",
        RevisionNumber: 0,
        Status:         "draft",
    }

    err := svc.SubmitForApproval(context.Background(), SubmitCommand{
        TenantID:       "tenant-1",
        DocumentID:     "doc-1",
        SubmittedBy:    "admin",
        RevisionTitle:  "",
        IdempotencyKey: "idem-1",
    })

    if err != nil {
        t.Fatalf("SubmitForApproval returned error: %v", err)
    }
    if repo.submittedRevisionTitle != "Criacao do documento" {
        t.Fatalf("revision title = %q, want %q", repo.submittedRevisionTitle, "Criacao do documento")
    }
}

func TestSubmitForApproval_RequiresRevisionTitleAfterFirstGovernedRevision(t *testing.T) {
    svc, repo := newSubmitServiceTestHarness(t)
    repo.document = documentForSubmit{
        ID:             "doc-1",
        RevisionNumber: 1,
        Status:         "draft",
    }

    err := svc.SubmitForApproval(context.Background(), SubmitCommand{
        TenantID:       "tenant-1",
        DocumentID:     "doc-1",
        SubmittedBy:    "admin",
        RevisionTitle:  "   ",
        IdempotencyKey: "idem-1",
    })

    if !errors.Is(err, ErrRevisionTitleRequired) {
        t.Fatalf("error = %v, want ErrRevisionTitleRequired", err)
    }
}
```

If the existing test harness uses different names, preserve the existing helper style and assert the same two behaviors.

- [ ] **Step 2: Run the focused failing test**

Run:

```powershell
go test ./internal/modules/documents/approval/application -run "RevisionTitle|SubmitForApproval" -count=1
```

Expected: FAIL because blank `revisionTitle` is currently rejected unconditionally.

- [ ] **Step 3: Implement conditional default**

In `submit_service.go`, centralize the title rule in a small helper:

```go
const defaultInitialRevisionTitle = "Criacao do documento"

func normalizeRevisionTitle(revisionNumber int, title string) (string, error) {
    trimmed := strings.TrimSpace(title)
    if trimmed != "" {
        return trimmed, nil
    }
    if revisionNumber == 0 {
        return defaultInitialRevisionTitle, nil
    }
    return "", ErrRevisionTitleRequired
}
```

Use the returned title in the existing submit/finalize transaction before persisting/freezing `documents.revision_title`.

- [ ] **Step 4: Run the focused backend tests**

Run:

```powershell
go test ./internal/modules/documents/approval/application -run "RevisionTitle|SubmitForApproval" -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit backend slice**

Run:

```powershell
git add internal/modules/documents/approval/application/submit_service.go internal/modules/documents/approval/application/submit_service_test.go
git commit -m "fix(documents): default initial governed revision title"
```

---

### Task 2: Sidebar Revision Row Semantics

**Files:**
- Modify: `frontend/apps/web/src/features/documents/lib/documentDetailMeta.ts`
- Modify: `frontend/apps/web/src/features/documents/lib/documentDetailMeta.test.ts`
- Modify: `frontend/apps/web/src/features/documents/components/EditorMetaSidebar.tsx`
- Modify: `frontend/apps/web/src/features/documents/components/EditorMetaSidebar.test.tsx`

- [ ] **Step 1: Write failing helper/component tests**

Add tests proving:

```tsx
it('renders revision code, title, and date without workflow status text', () => {
  render(
    <EditorMetaSidebar
      open
      onToggle={() => undefined}
      history={[{
        documentId: 'doc-1',
        revisionCode: 'REV00',
        revisionTitle: 'Criacao do documento',
        status: 'draft',
        createdAt: '2026-05-19T10:00:00-03:00',
        isCurrent: true,
      }]}
    />,
  );

  expect(screen.getByText('REV00')).toBeInTheDocument();
  expect(screen.getByText('Criacao do documento')).toBeInTheDocument();
  expect(screen.getByText('19/05/2026')).toBeInTheDocument();
  expect(screen.queryByText(/Draft|Em revis/i)).not.toBeInTheDocument();
});

it('collapses long governed histories and can expand them', async () => {
  const history = Array.from({ length: 5 }, (_, index) => ({
    documentId: `doc-${index}`,
    revisionCode: `REV0${index}`,
    revisionTitle: `Revision ${index}`,
    status: 'approved',
    createdAt: '2026-05-19T10:00:00-03:00',
    isCurrent: index === 4,
  }));

  render(<EditorMetaSidebar open onToggle={() => undefined} history={history} />);

  expect(screen.getByText('REV04')).toBeInTheDocument();
  expect(screen.queryByText('REV00')).not.toBeInTheDocument();

  await userEvent.click(screen.getByRole('button', { name: /ver todas as revis/i }));
  expect(screen.getByText('REV00')).toBeInTheDocument();
});
```

- [ ] **Step 2: Run the focused failing tests**

Run:

```powershell
cd frontend/apps/web
pnpm.cmd vitest run src/features/documents/components/EditorMetaSidebar.test.tsx src/features/documents/lib/documentDetailMeta.test.ts
```

Expected: FAIL because current revision rows combine code/status and no collapse affordance exists.

- [ ] **Step 3: Implement status-free revision rows**

Update `EditorMetaSidebar.tsx` to render a local revision list instead of feeding the generic `TimelineRail` with status-bearing titles:

```tsx
const MAX_COLLAPSED_HISTORY_ITEMS = 3;
const [historyExpanded, setHistoryExpanded] = useState(false);
const visibleHistory = historyExpanded || history.length <= MAX_COLLAPSED_HISTORY_ITEMS
  ? history
  : history.filter((item) => item.isCurrent).concat(history.filter((item) => !item.isCurrent).slice(0, MAX_COLLAPSED_HISTORY_ITEMS - 1));
```

Render each row with:

```tsx
<div className={`${styles.revisionRow} ${item.isCurrent ? styles.revisionRowCurrent : ''}`} key={item.documentId}>
  <div className={styles.revisionMarker} aria-hidden="true" />
  <div className={styles.revisionBody}>
    <span className={styles.revisionCode}>{item.revisionCode}</span>
    <span className={styles.revisionTitle}>{item.revisionTitle || 'Criacao do documento'}</span>
  </div>
  <time className={styles.revisionDate} dateTime={item.createdAt}>{formatShortDate(item.createdAt)}</time>
</div>
```

Add an expand/collapse button only when `history.length > 3`.

- [ ] **Step 4: Run the focused frontend tests**

Run:

```powershell
cd frontend/apps/web
pnpm.cmd vitest run src/features/documents/components/EditorMetaSidebar.test.tsx src/features/documents/lib/documentDetailMeta.test.ts
```

Expected: PASS.

- [ ] **Step 5: Commit frontend sidebar semantics**

Run:

```powershell
git add frontend/apps/web/src/features/documents/lib/documentDetailMeta.ts frontend/apps/web/src/features/documents/lib/documentDetailMeta.test.ts frontend/apps/web/src/features/documents/components/EditorMetaSidebar.tsx frontend/apps/web/src/features/documents/components/EditorMetaSidebar.test.tsx
git commit -m "feat(editor): render compact governed revision history"
```

---

### Task 3: `REV00` Frontend Submit Flow

**Files:**
- Modify: `frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx`
- Modify: `frontend/apps/web/src/features/documents/pages/DocumentEditorPage.test.tsx`

- [ ] **Step 1: Write failing page tests**

Add page tests proving:

```tsx
it('submits REV00 without opening revision title validation', async () => {
  mockDocumentDetail({ RevisionVersion: 0, Status: 'draft', RevisionTitle: null });
  render(<DocumentEditorPage documentID="doc-1" onDone={() => undefined} />);

  await userEvent.click(await screen.findByRole('button', { name: /submeter para revis/i }));

  expect(screen.queryByRole('dialog', { name: /titulo da revis/i })).not.toBeInTheDocument();
  expect(finalizeDocument).toHaveBeenCalledWith('doc-1', expect.objectContaining({
    revisionTitle: undefined,
  }));
});

it('keeps requiring a manual revision title after REV00', async () => {
  mockDocumentDetail({ RevisionVersion: 1, Status: 'draft', RevisionTitle: null });
  render(<DocumentEditorPage documentID="doc-1" onDone={() => undefined} />);

  await userEvent.click(await screen.findByRole('button', { name: /submeter para revis/i }));
  await userEvent.click(screen.getByRole('button', { name: /confirmar submiss/i }));

  expect(screen.getByRole('alert')).toHaveTextContent(/titulo da revis/i);
  expect(finalizeDocument).not.toHaveBeenCalled();
});
```

Adjust mock helper names to the existing test structure.

- [ ] **Step 2: Run the focused failing page test**

Run:

```powershell
cd frontend/apps/web
pnpm.cmd vitest run src/features/documents/pages/DocumentEditorPage.test.tsx -t "REV00|manual revision title"
```

Expected: FAIL because current flow always opens the title dialog.

- [ ] **Step 3: Implement conditional dialog skip**

Update `handleFinalize`:

```ts
function handleFinalize() {
  if ((doc?.RevisionVersion ?? 0) === 0) {
    void submitForReview(undefined);
    return;
  }
  setRevisionTitleInput(doc?.RevisionTitle ?? '');
  setRevisionTitleError(null);
  setRevisionTitleDialogOpen(true);
}
```

Update `submitForReview` to accept `revisionTitle?: string` and only include `revisionTitle` in the request body when a non-empty title is provided:

```ts
const body = revisionTitle ? { revisionTitle } : {};
await finalizeDocument(documentID, body);
```

Keep the dirty-editor save/flush behavior before the mutation unchanged.

- [ ] **Step 4: Run focused page tests**

Run:

```powershell
cd frontend/apps/web
pnpm.cmd vitest run src/features/documents/pages/DocumentEditorPage.test.tsx
```

Expected: PASS.

- [ ] **Step 5: Commit frontend submit flow**

Run:

```powershell
git add frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx frontend/apps/web/src/features/documents/pages/DocumentEditorPage.test.tsx
git commit -m "fix(editor): skip manual revision title for initial revision"
```

---

### Task 4: Sidebar Visual Density

**Files:**
- Modify: `frontend/apps/web/src/features/documents/components/EditorMetaSidebar.module.css`
- Modify: `frontend/apps/web/src/features/documents/components/EditorMetaSidebar.tsx`
- Modify: `frontend/apps/web/src/features/documents/components/EditorMetaSidebar.test.tsx`

- [ ] **Step 1: Add a test for dense sidebar structure**

In `EditorMetaSidebar.test.tsx`, assert semantic structure rather than pixels:

```tsx
it('uses compact section structure for metadata and revision history', () => {
  render(<EditorMetaSidebar open onToggle={() => undefined} code="POP-001" history={[]} />);

  expect(screen.getByRole('complementary', { name: /metadados do documento/i })).toBeInTheDocument();
  expect(screen.getByText('Metadados')).toBeInTheDocument();
  expect(screen.getByText('Revisoes')).toBeInTheDocument();
});
```

- [ ] **Step 2: Implement CSS density improvements**

Update CSS to:

```css
.sidebar {
  width: 280px;
  flex: 0 0 280px;
  min-height: 0;
  background:
    linear-gradient(180deg, rgba(255,255,255,0.92), rgba(250,247,244,0.96));
  border-left: 1px solid var(--border);
  overflow-y: auto;
  padding: var(--sp-4);
}

.section {
  padding: var(--sp-3) 0;
}

.section:first-child {
  padding-top: 0;
}

.revisionList {
  display: flex;
  flex-direction: column;
  gap: var(--sp-2);
}
```

Use repo tokens and existing wine palette; avoid inline styles.

- [ ] **Step 3: Run component tests and typecheck**

Run:

```powershell
cd frontend/apps/web
pnpm.cmd vitest run src/features/documents/components/EditorMetaSidebar.test.tsx
pnpm.cmd tsc --noEmit -p tsconfig.build.json
```

Expected: PASS.

- [ ] **Step 4: Browser screenshot check**

With API and frontend running, open:

```text
http://localhost:4173/documents/81d2703a-d701-4c55-a834-363e2ecf8c16/edit
```

Expected visual:

- sidebar is compact and visually bounded
- metadata displays real values
- history row displays `REV00`, `Criacao do documento`, and date
- draft has no approvers

Save screenshot:

```text
docs/superpowers/artifacts/screenshots/2026-05-19-editor-sidebar-draft-density.png
```

- [ ] **Step 5: Commit visual density slice**

Run:

```powershell
git add frontend/apps/web/src/features/documents/components/EditorMetaSidebar.module.css frontend/apps/web/src/features/documents/components/EditorMetaSidebar.tsx frontend/apps/web/src/features/documents/components/EditorMetaSidebar.test.tsx docs/superpowers/artifacts/screenshots/2026-05-19-editor-sidebar-draft-density.png
git commit -m "style(editor): tighten sidebar visual density"
```

---

### Task 5: Wiki Sync and Verification

**Files:**
- Modify: `wiki/modules/documents.md`
- Modify: `wiki/modules/documents-tech-debt.md`

- [ ] **Step 1: Update module wiki memory**

Add a dated entry to `wiki/modules/documents.md`:

```md
- 2026-05-19 - Editor sidebar amendment: `REV00` now uses the canonical initial governed title `Criacao do documento`; later governed revisions still require `revisionTitle` at formal submission. The editor sidebar renders governed revision rows as code/title/date without inline workflow status and collapses long histories. `document_revisions` remains technical/autosave-only.
```

If no new debt remains, add a short note to `wiki/modules/documents-tech-debt.md` that rich history search/filtering remains deferred rather than opening a new debt row.

- [ ] **Step 2: Run focused backend verification**

Run:

```powershell
go test ./internal/modules/documents/approval/application ./internal/modules/documents/delivery/http ./internal/modules/documents/repository -count=1
```

Expected: PASS.

- [ ] **Step 3: Run focused frontend verification**

Run:

```powershell
cd frontend/apps/web
pnpm.cmd vitest run src/features/documents/components/EditorMetaSidebar.test.tsx src/features/documents/pages/DocumentEditorPage.test.tsx src/features/documents/lib/documentDetailMeta.test.ts
pnpm.cmd tsc --noEmit -p tsconfig.build.json
```

Expected: PASS.

- [ ] **Step 4: Commit docs and final verification evidence**

Run:

```powershell
git add wiki/modules/documents.md wiki/modules/documents-tech-debt.md
git commit -m "docs(documents): sync editor sidebar revision semantics"
```

---

## Self-Review Checklist

- Spec coverage:
  - `REV00` automatic title: Task 1 and Task 3.
  - `REV01+` title required: Task 1 and Task 3.
  - Sidebar row without status: Task 2.
  - Compact/collapsible sidebar: Task 2 and Task 4.
  - Draft no approvers: covered by existing sidebar gate and Task 2 tests.
  - Wiki sync: Task 5.
- Placeholder scan:
  - No `TBD`, `TODO`, or unbounded “implement later” steps.
- Type consistency:
  - Existing frontend contract uses `RevisionVersion`; backend domain uses `revision_number`.
  - Existing API wrapper `finalizeDocument` remains the mutation boundary.
  - Existing `EditorSidebarRevisionItem` remains the sidebar input shape.

