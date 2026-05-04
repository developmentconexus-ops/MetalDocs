# Group E Sub-Plan 3 — Misc UI Hardening Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Three small consumer-facing UI fixes — inbox area filter from taxonomy (E7), document rename rollback on server error (E9), RegistryCreateDialog submit gated on auth readiness (E12).

**Architecture:** Defensive frontend hardening only. Zero backend. Rides shared `apiFetch`/`ApiError`/`resolveErrorMessage` from sub-plan 1. Three independent files in three independent features → fully parallel execution.

**Tech Stack:** React 18 + TypeScript + Vite, Zustand (auth.store), vitest, msw.

**Spec:** `docs/superpowers/specs/2026-05-04-group-e-misc-design.md` (commit `0581decd`).

---

## Model Routing

| Phase | Role | Model |
|---|---|---|
| 0 | Worktree + codex spec validate | sonnet (controller) → codex (validator) |
| 1a | E7 Inbox area filter + tests | sonnet (mechanical) |
| 1b | E9 rename rollback + tests | sonnet (mechanical) |
| 1c | E12 submit gate + tests | sonnet (mechanical) |
| 2 | Verify (vitest + tsc + lint + smoke + codex audit) | sonnet → **codex audit** |
| 3 | Audit doc + wiki-curator + finishing-a-development-branch | sonnet + wiki-curator agent |
| Phase reviews | between phases | **opus** |

Phase 1a ‖ 1b ‖ 1c run in parallel — three independent files, no shared symbols changed.

---

## File Structure

| File | Responsibility | Action |
|---|---|---|
| `frontend/apps/web/src/features/approval/pages/InboxPage.tsx` | Inbox page | Modify (E7) |
| `frontend/apps/web/src/features/approval/pages/InboxPage.test.tsx` | Inbox tests | Modify (add area-filter test) |
| `frontend/apps/web/src/features/documents/v2/DocumentEditorPage.tsx` | Editor page | Modify (E9 — `handleRename`) |
| `frontend/apps/web/src/features/documents/v2/DocumentEditorPage.test.tsx` | Editor tests | Modify (append rollback test, file scaffolded by sub-plan 2) |
| `frontend/apps/web/src/features/registry/RegistryCreateDialog.tsx` | Create dialog | Modify (E12) |
| `frontend/apps/web/src/features/registry/RegistryCreateDialog.test.tsx` | Dialog tests | Create |
| `wiki/bugs/audit-2026-05-03.md` | Audit closure | Modify |

---

## Phase 0 — Setup

### Task 0.1: Worktree

- [ ] **Step 1: Create worktree off main**

```bash
git worktree add -b phase-e-misc ../metaldocs-phase-e-misc main
cd ../metaldocs-phase-e-misc
```

- [ ] **Step 2: Verify clean baseline**

```bash
cd frontend/apps/web && npx vitest run 2>&1 | tail -5
```

Expected: existing tests pass.

- [ ] **Step 3: Confirm dependency on sub-plan 1**

This sub-plan imports `resolveErrorMessage` from `frontend/apps/web/src/lib/api/errorMessages.ts` and `ApiError` from `frontend/apps/web/src/lib/api/client.ts` (created in sub-plan 1). If those files do not exist yet, **stop** and merge sub-plan 1 first.

```bash
ls frontend/apps/web/src/lib/api/errorMessages.ts frontend/apps/web/src/lib/api/client.ts
```

Expected: both files exist.

### Task 0.2: Codex spec validation

- [ ] **Step 1: Dispatch codex:codex-rescue subagent**

Prompt:

> Validate spec at `docs/superpowers/specs/2026-05-04-group-e-misc-design.md`. Report PASS/FAIL on: (1) E7 fetchAreas signature matches `frontend/apps/web/src/features/taxonomy/api.ts`; (2) E9 rollback handles both `ApiError` (post sub-plan 1 migration) and legacy plain Error from current `renameDocument`; (3) E12 `useAuthStore` selector `(s) => s.user` returns nullable shape `{userId, displayName}`; (4) no shared symbols changed across the three files (parallel-safe). Cite file:line for every concern.

- [ ] **Step 2: Address PASS/FAIL findings inline if any.** No commit needed.

---

## Phase 1a — E7: Inbox Area Filter from Taxonomy

### Task 1a.1: Failing test

**Files:**
- Modify: `frontend/apps/web/src/features/approval/pages/InboxPage.test.tsx`

- [ ] **Step 1: Add failing test**

Append (or insert into existing `describe('InboxPage')` block):

```tsx
import * as taxonomyApi from '../../taxonomy/api';

it('populates area filter from fetchAreas, not hardcoded list', async () => {
  vi.spyOn(taxonomyApi, 'fetchAreas').mockResolvedValue([
    { code: 'OPS', name: 'Operações' } as any,
    { code: 'QA', name: 'Qualidade' } as any,
  ]);
  vi.spyOn(approvalApi, 'listInbox').mockResolvedValue({ items: [], total: 0 });

  render(<InboxPage />, { wrapper: BrowserRouter });

  await screen.findByRole('option', { name: /Todas as áreas/i });
  expect(screen.getByRole('option', { name: /OPS — Operações/ })).toBeInTheDocument();
  expect(screen.getByRole('option', { name: /QA — Qualidade/ })).toBeInTheDocument();
  expect(screen.queryByRole('option', { name: /JUR/ })).not.toBeInTheDocument();
});
```

If file lacks `BrowserRouter` wrapper or `approvalApi` import, add them at the top:

```tsx
import { BrowserRouter } from 'react-router-dom';
import * as approvalApi from '../api/approvalApi';
```

- [ ] **Step 2: Run, expect FAIL**

```bash
cd frontend/apps/web && npx vitest run src/features/approval/pages/InboxPage.test.tsx -t "populates area filter"
```

Expected: FAIL — hardcoded `JUR` option still rendered.

### Task 1a.2: Replace hardcoded list with fetchAreas

**Files:**
- Modify: `frontend/apps/web/src/features/approval/pages/InboxPage.tsx`

- [ ] **Step 1: Remove `AREA_OPTIONS` constant (line 8)**

Delete:
```tsx
const AREA_OPTIONS = ['', 'JUR', 'RH', 'FIN', 'TI', 'COM', 'ENG'];
```

- [ ] **Step 2: Add imports (top of file)**

```tsx
import { fetchAreas } from '../../taxonomy/api';
import type { ProcessArea } from '../../taxonomy/types';
```

- [ ] **Step 3: Add areas state + load effect after `setPage` declaration (~line 38)**

```tsx
const [areas, setAreas] = useState<ProcessArea[]>([]);

useEffect(() => {
  void fetchAreas().then(setAreas).catch(() => setAreas([]));
}, []);
```

- [ ] **Step 4: Replace area select JSX**

Find the existing `<select>` rendering area filter (look for `areaFilter` and the previous `AREA_OPTIONS.map`). Replace with:

```tsx
<select value={areaFilter} onChange={(e) => setAreaFilter(e.target.value)}>
  <option value="">Todas as áreas</option>
  {areas.map((a) => (
    <option key={a.code} value={a.code}>{a.code} — {a.name}</option>
  ))}
</select>
```

- [ ] **Step 5: Run test, expect PASS**

```bash
cd frontend/apps/web && npx vitest run src/features/approval/pages/InboxPage.test.tsx -t "populates area filter"
```

- [ ] **Step 6: Run full inbox test file**

```bash
cd frontend/apps/web && npx vitest run src/features/approval/pages/InboxPage.test.tsx
```

Expected: all PASS, no regressions.

- [ ] **Step 7: Commit**

```bash
git add frontend/apps/web/src/features/approval/pages/InboxPage.tsx frontend/apps/web/src/features/approval/pages/InboxPage.test.tsx
git commit -m "fix(inbox): area filter sourced from taxonomy fetchAreas, not hardcoded list (E7)"
```

---

## Phase 1b — E9: Rename Rollback on Server Error

### Task 1b.1: Failing test

**Files:**
- Modify: `frontend/apps/web/src/features/documents/v2/DocumentEditorPage.test.tsx`

- [ ] **Step 1: Append test**

```tsx
import { ApiError } from '../../../lib/api/client';
import * as documentsApi from './api/documentsV2';
import { toast } from 'sonner';

it('rolls back document name on rename failure and toasts error', async () => {
  server.use(
    http.get('/api/v2/documents/:id', () => HttpResponse.json({
      Status: 'draft', CurrentRevisionID: 'r1', Name: 'Original.docx', Code: 'C-001', RevisionVersion: 1,
    })),
    http.get('/api/v2/documents/:id/revisions/:rid/signed-url', () => HttpResponse.json({ url: 'blob:x' })),
    http.get('blob:x', () => new HttpResponse(new ArrayBuffer(8))),
  );

  const renameSpy = vi.spyOn(documentsApi, 'renameDocument')
    .mockRejectedValueOnce(new ApiError('not_found', 404, 'Document not found'));
  const toastSpy = vi.spyOn(toast, 'error');

  render(<DocumentEditorPage documentID="d1" onDone={() => {}} />);

  // wait for initial doc to load and title to render
  await screen.findByText(/Original/);

  // simulate the editor calling onChangeName via the rename callback
  // (the editor mock in sub-plan 2 exposes mode prop; for E9 we trigger handleRename via the title element if exposed,
  // OR we extract handleRename through the MetalDocsEditor mock's `onTitleChange` prop)
  // The cleanest path: have the test mock invoke the prop. The MetalDocsEditor mock is:
  //   MetalDocsEditor: (props) => <div data-testid="editor" data-mode={props.mode} onClick={() => props.onTitleChange?.('NewName')} />
  // Adjust the mock in this test file to also forward onTitleChange. Then click the editor.
  // For now (until editor onTitleChange wired), invoke via title input if present.

  // Programmatic invocation via the rename input (assuming title bar has a contentEditable or input):
  // Placeholder: trigger a synthetic event the editor uses.
  // If your editor surface for renaming is the doc title text node, replace with:
  //   const titleEl = await screen.findByText(/Original/);
  //   fireEvent.change(titleEl, { target: { value: 'NewName' } });
  // Use whichever path matches the actual rename UI. If unsure, expose a test-only `data-testid="rename-trigger"` on
  // the rename callback wiring during 1b.2 and click it.

  // Generic dispatch: fire a custom rename event the page wires up.
  fireEvent.click(screen.getByTestId('editor')); // requires mock to forward onTitleChange via onClick

  await waitFor(() => expect(renameSpy).toHaveBeenCalledWith('d1', 'NewName'));
  await waitFor(() => expect(screen.getByText(/Original/)).toBeInTheDocument()); // rolled back
  expect(toastSpy).toHaveBeenCalled();
  const msg = toastSpy.mock.calls[0]?.[0];
  expect(typeof msg).toBe('string');
});
```

If the existing `MetalDocsEditor` mock (from sub-plan 2 test file) does not forward `onTitleChange`, update it to:

```tsx
vi.mock('@metaldocs/editor-ui', () => ({
  MetalDocsEditor: (props: any) => (
    <div
      data-testid="editor"
      data-mode={props.mode}
      onClick={() => props.onTitleChange?.('NewName')}
    />
  ),
}));
```

- [ ] **Step 2: Run, expect FAIL**

```bash
cd frontend/apps/web && npx vitest run src/features/documents/v2/DocumentEditorPage.test.tsx -t "rolls back"
```

Expected: FAIL — current `handleRename` does not roll back.

### Task 1b.2: Implement rollback

**Files:**
- Modify: `frontend/apps/web/src/features/documents/v2/DocumentEditorPage.tsx`

- [ ] **Step 1: Add import (top of file)**

```tsx
import { resolveErrorMessage } from '../../../lib/api/errorMessages';
```

- [ ] **Step 2: Replace `handleRename` (line ~112)**

```tsx
const handleRename = useCallback((name: string) => {
  const prev = documentName;
  setDocumentName(name);
  void renameDocument(documentID, name).catch((err: unknown) => {
    setDocumentName(prev);
    const code = (err && typeof err === 'object' && 'code' in err)
      ? (err as { code?: string }).code
      : undefined;
    toast.error(resolveErrorMessage(code, 'Falha ao renomear documento.'));
  });
}, [documentID, documentName]);
```

- [ ] **Step 3: Verify `<MetalDocsEditor onTitleChange={handleRename}` is wired**

Find the `<MetalDocsEditor>` instantiation (around line 230). If `onTitleChange` is not already a prop, add:

```tsx
<MetalDocsEditor
  ref={editorRef}
  mode={isEditable ? 'document-edit' : 'readonly'}
  onTitleChange={handleRename}
  ...
/>
```

(If editor-ui already accepts `onTitleChange`, the prop is forwarded; if not, the project's existing rename surface — likely an overlay title input — is the actual hook. In that case, ensure that surface invokes `handleRename`.)

- [ ] **Step 4: Run test, expect PASS**

```bash
cd frontend/apps/web && npx vitest run src/features/documents/v2/DocumentEditorPage.test.tsx -t "rolls back"
```

- [ ] **Step 5: Run full editor test file**

```bash
cd frontend/apps/web && npx vitest run src/features/documents/v2/DocumentEditorPage.test.tsx
```

Expected: all PASS.

- [ ] **Step 6: Commit**

```bash
git add frontend/apps/web/src/features/documents/v2/DocumentEditorPage.tsx frontend/apps/web/src/features/documents/v2/DocumentEditorPage.test.tsx
git commit -m "fix(editor): rollback document name on rename failure with structured toast (E9)"
```

---

## Phase 1c — E12: RegistryCreateDialog Submit Gate

### Task 1c.1: Failing tests

**Files:**
- Create: `frontend/apps/web/src/features/registry/RegistryCreateDialog.test.tsx`

- [ ] **Step 1: Write the test file**

```tsx
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { setupServer } from 'msw/node';
import { http, HttpResponse } from 'msw';
import { RegistryCreateDialog } from './RegistryCreateDialog';

vi.mock('../../store/auth.store', () => ({
  useAuthStore: vi.fn(),
}));
vi.mock('../taxonomy/api', () => ({
  fetchProfiles: vi.fn().mockResolvedValue([]),
  fetchAreas: vi.fn().mockResolvedValue([]),
}));

import { useAuthStore } from '../../store/auth.store';

const server = setupServer();
beforeEach(() => server.listen());
afterEach(() => { server.resetHandlers(); server.close(); vi.clearAllMocks(); });

describe('RegistryCreateDialog E12 submit gate', () => {
  it('disables submit and shows placeholder while currentUser is null', () => {
    (useAuthStore as unknown as vi.Mock).mockImplementation((sel: any) => sel({ user: null }));
    render(<RegistryCreateDialog onClose={() => {}} onCreated={() => {}} />);

    expect(screen.getByRole('button', { name: /Aguardando/i })).toBeDisabled();
    expect(screen.getByDisplayValue('Aguardando autenticação...')).toBeInTheDocument();
  });

  it('enables submit when currentUser ready', () => {
    (useAuthStore as unknown as vi.Mock).mockImplementation((sel: any) =>
      sel({ user: { userId: 'u-1', displayName: 'Alice' } }));
    render(<RegistryCreateDialog onClose={() => {}} onCreated={() => {}} />);

    expect(screen.getByRole('button', { name: /^Criar$/i })).not.toBeDisabled();
    expect(screen.getByDisplayValue('Alice')).toBeInTheDocument();
  });

  it('falls back to userId when displayName missing', () => {
    (useAuthStore as unknown as vi.Mock).mockImplementation((sel: any) =>
      sel({ user: { userId: 'u-2', displayName: null } }));
    render(<RegistryCreateDialog onClose={() => {}} onCreated={() => {}} />);

    expect(screen.getByDisplayValue('u-2')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /^Criar$/i })).not.toBeDisabled();
  });
});
```

- [ ] **Step 2: Run, expect FAIL**

```bash
cd frontend/apps/web && npx vitest run src/features/registry/RegistryCreateDialog.test.tsx
```

Expected: FAIL — current dialog has no `Aguardando` text.

### Task 1c.2: Implement submit gate

**Files:**
- Modify: `frontend/apps/web/src/features/registry/RegistryCreateDialog.tsx`

- [ ] **Step 1: Derive `isAuthReady` (after `currentUser` declaration, line ~14)**

```tsx
const currentUser = useAuthStore((s) => s.user);
const isAuthReady = !!currentUser?.userId;
```

- [ ] **Step 2: Replace the "Autor" readonly input (lines ~123-128)**

```tsx
<input
  value={
    isAuthReady
      ? (currentUser!.displayName ?? currentUser!.userId)
      : 'Aguardando autenticação...'
  }
  readOnly
  style={{
    width: '100%', padding: '6px 8px', boxSizing: 'border-box',
    background: '#f5f5f5', color: isAuthReady ? '#666' : '#aaa', cursor: 'not-allowed',
  }}
/>
```

- [ ] **Step 3: Replace submit button (line ~194)**

```tsx
<button type="submit" disabled={saving || !isAuthReady} style={{ padding: '6px 14px' }}>
  {!isAuthReady ? 'Aguardando...' : saving ? 'Criando...' : 'Criar'}
</button>
```

- [ ] **Step 4: Run tests, expect PASS**

```bash
cd frontend/apps/web && npx vitest run src/features/registry/RegistryCreateDialog.test.tsx
```

- [ ] **Step 5: Commit**

```bash
git add frontend/apps/web/src/features/registry/RegistryCreateDialog.tsx frontend/apps/web/src/features/registry/RegistryCreateDialog.test.tsx
git commit -m "fix(registry): gate create-dialog submit on auth readiness, show placeholder for author (E12)"
```

---

## Phase 2 — Verify

### Task 2.1: Full frontend sweep

- [ ] **Step 1: Run all vitest**

```bash
cd frontend/apps/web && npx vitest run
```

Expected: all PASS.

- [ ] **Step 2: Typecheck**

```bash
cd frontend/apps/web && npx tsc --noEmit
```

Expected: no new errors.

- [ ] **Step 3: Lint**

```bash
cd frontend/apps/web && npm run lint
```

Expected: zero new warnings.

### Task 2.2: Smoke flows

Use `.\scripts\start-api.ps1` and dev frontend.

- [ ] **Smoke E7:** Open Inbox → area dropdown shows codes from `process_areas` table for the current tenant (not the JUR/RH/FIN/TI/COM/ENG defaults).
- [ ] **Smoke E9:** Open a draft, rename to a name the backend rejects (e.g. trigger validation error). Observe: name visibly reverts to previous + toast.
- [ ] **Smoke E12:** Open Create dialog under throttled network. Submit button shows "Aguardando..." and is disabled while auth resolves; flips to "Criar" once `currentUser` is ready.

### Task 2.3: Codex independent audit

- [ ] **Step 1: Dispatch codex:codex-rescue subagent**

Prompt:

> Independent audit of Group E sub-plan 3 (`docs/superpowers/specs/2026-05-04-group-e-misc-design.md`) implementation. For each of E7, E9, E12 produce PASS/FAIL with file:line evidence. Verify:
> (a) `AREA_OPTIONS` constant fully removed; no other hardcoded area lists remain in InboxPage
> (b) `handleRename` captures previous name before optimistic update; `.catch` restores it; toast uses `resolveErrorMessage`
> (c) `RegistryCreateDialog` submit disabled when `currentUser?.userId` falsy; placeholder text rendered in author field; existing happy path preserved (no extra renders)
> (d) No raw `fetch` introduced (use `apiFetch` if any new HTTP calls were added)
> (e) Tests cover the failure paths, not just happy paths

Required: 3/3 PASS before Phase 3.

---

## Phase 3 — Closure

### Task 3.1: Audit doc closure

**Files:**
- Modify: `wiki/bugs/audit-2026-05-03.md` (lines 277, 279, 282)

- [ ] **Step 1: Mark E7/E9/E12 closed**

Change status column to `fixed` and Fix-commit column to the matching SHAs from Phase 1a/1b/1c commits.

- [ ] **Step 2: Add E5 deferral note**

In Group E section (around line 275), append note (or update E5 row's status column to `deferred` and Fix-commit to a reference like `→ E-admin sub-plan`):

```
| E5 | `document-profiles` endpoints 404 (`/bundle`, `/schema`, `/governance`) — not registered on backend | 🟡 medium | deferred | → E-admin sub-plan (TBD) |
```

Also append below the table (or in cross-cutting follow-ups section):

```
- **E5 deferred — E-admin sub-plan TBD**: implement `/api/v2/taxonomy/profiles/{code}/{bundle,schema,governance}` plus sibling registry-explorer admin routes. ~10 handlers + supporting services. Out of scope for sub-plan 3 (consumer-facing UX only).
```

- [ ] **Step 3: Commit**

```bash
git add wiki/bugs/audit-2026-05-03.md
git commit -m "docs(audit): close E7/E9/E12 with fix SHAs; defer E5 to E-admin sub-plan"
```

### Task 3.2: Wiki-curator dispatch

- [ ] **Step 1: Dispatch wiki-curator agent**

Prompt:

> Update wiki for Group E sub-plan 3. Refresh `Last verified` stamps on any frontend module docs that reference InboxPage, DocumentEditorPage, or RegistryCreateDialog. If `wiki/concepts/error-ux.md` exists (created by sub-plan 1), append a note about the rename rollback pattern as a canonical example of optimistic-UI + ApiError-driven recovery. Update `wiki/README.md` index. Do not create new docs unless a clearly new module/concept emerged (none expected here).

### Task 3.3: Branch finishing

- [ ] **Step 1: Use superpowers:finishing-a-development-branch**

Default to option 2 (push + PR) unless user requests otherwise.

---

## Acceptance Criteria

- [ ] E7: Inbox area filter populated from `fetchAreas()`; "Todas as áreas" option present
- [ ] E7: hardcoded `AREA_OPTIONS` constant removed
- [ ] E9: rename failure restores previous name in UI
- [ ] E9: error toast uses `resolveErrorMessage(err.code, fallback)`
- [ ] E12: submit button disabled until `currentUser?.userId` present
- [ ] E12: "Autor" field shows "Aguardando autenticação..." while loading
- [ ] All vitest pass
- [ ] `npx tsc --noEmit` passes
- [ ] No new lint warnings
- [ ] Codex audit returns 3/3 PASS
- [ ] Audit doc updated, E7/E9/E12 closed with commit SHAs
- [ ] Audit doc adds follow-up entry: "E-admin sub-plan deferred (E5 — bundle/schema/governance handlers)"
- [ ] Wiki refreshed by wiki-curator
