# F2d.5b `f5b-pdf-read-canvas` Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Official post-approval PDF as the workspace canvas for `approved/scheduled/published` docs (D1), and a genuinely lazy editor chunk so the PDF canvas never downloads TipTap (D2). FE-only — zero backend diff.

**Architecture:** Status-keyed canvas branch in `DocumentWorkspacePage` (`PdfCanvas` when `doc.status ∈ {approved, scheduled, published}`; docx read canvas otherwise). `PdfCanvas` reuses the existing `useDocumentPdfStatus` polling hook against `GET /documents/{id}/view` (backend untouched). D2 converts `DocumentShell`'s static `MetalDocsEditor` import to `React.lazy` + `Suspense`.

**Tech Stack:** React 18, react-router-dom, vitest + Testing Library, CSS modules (Wine tokens). Working dir for all commands: `frontend/apps/web`.

**Design authority:** `docs/superpowers/specs/2026-07-09-f5b-pdf-official-view-design.md`. Milestone row F2d.5b in `milestone-2d-workflow-coherence-fe/milestone.md`.

**Binding constraints:**
- STRICT TDD: failing test first in every task, run it RED before implementing.
- Zero backend diff — `git status` must show no change outside `frontend/` + this feature's docs.
- `viewableStatuses` (backend) serves ONLY `approved/scheduled/published`. `lifecycle` mode ALSO covers `superseded/obsolete` — those get NO PdfCanvas (view returns 404 for them); they keep the docx read canvas. Gate on status set, never on `mode === 'lifecycle'`.
- Type-only imports of `@metaldocs/editor-ui` are free (erased at build). Only VALUE imports matter for D2. Current value-imports: `DocumentShell.tsx:2` (the target), `templates/components/TemplateReviewCanvas.tsx:1` + `templates/pages/TemplateEditorPage.tsx:3` (OUT OF SCOPE — do not touch templates).
- Wine design tokens only; PT-BR copy; no `any`; no new dependencies.

---

### Task 1: `PdfCanvas` component (TDD)

**Files:**
- Create: `frontend/apps/web/src/features/documents/components/workspace/PdfCanvas.tsx`
- Create: `frontend/apps/web/src/features/documents/components/workspace/PdfCanvas.module.css`
- Test: `frontend/apps/web/src/features/documents/components/workspace/PdfCanvas.test.tsx`

- [ ] **Step 1: Write the failing test**

```tsx
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, expect, it, vi, beforeEach } from 'vitest';
import type { DocumentPdfStatus } from '../../hooks/editor/useDocumentPdfStatus';
import { PdfCanvas } from './PdfCanvas';

const hookState: { value: DocumentPdfStatus } = {
  value: { status: 'pending', retry: vi.fn() },
};

vi.mock('../../hooks/editor/useDocumentPdfStatus', () => ({
  useDocumentPdfStatus: () => hookState.value,
}));

describe('PdfCanvas', () => {
  beforeEach(() => {
    hookState.value = { status: 'pending', retry: vi.fn() };
  });

  it('ready: renders the official PDF embed with the presigned URL', () => {
    hookState.value = { status: 'ready', url: 'https://s3/final.pdf', retry: vi.fn() };
    render(<PdfCanvas documentId="doc-1" />);
    const frame = screen.getByTitle('Documento oficial (PDF)');
    expect(frame).toHaveAttribute('src', 'https://s3/final.pdf');
  });

  it('pending: renders the generating state, no embed', () => {
    render(<PdfCanvas documentId="doc-1" />);
    expect(screen.getByRole('status')).toHaveTextContent('Gerando o PDF oficial');
    expect(screen.queryByTitle('Documento oficial (PDF)')).not.toBeInTheDocument();
  });

  it('failed: renders the error alert and retry re-polls', () => {
    const retry = vi.fn();
    hookState.value = { status: 'failed', retry };
    render(<PdfCanvas documentId="doc-1" />);
    expect(screen.getByRole('alert')).toHaveTextContent('Não foi possível gerar o PDF');
    fireEvent.click(screen.getByRole('button', { name: 'Tentar novamente' }));
    expect(retry).toHaveBeenCalledTimes(1);
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run (from `frontend/apps/web`): `npx vitest run src/features/documents/components/workspace/PdfCanvas.test.tsx`
Expected: FAIL — `Cannot find module './PdfCanvas'` (or equivalent resolve error).

- [ ] **Step 3: Write minimal implementation**

`PdfCanvas.tsx`:

```tsx
import { useDocumentPdfStatus } from '../../hooks/editor/useDocumentPdfStatus';
import styles from './PdfCanvas.module.css';

/**
 * F2d.5b D1 — the official post-approval PDF canvas.
 *
 * Rendered by DocumentWorkspacePage ONLY for document statuses the backend
 * serves via GET /documents/:id/view (`approved`/`scheduled`/`published` —
 * view_service.go viewableStatuses). Reuses the useDocumentPdfStatus polling
 * hook (pending → 3s poll → ready/failed). The PDF is the official artifact;
 * in-approval viewing stays on the source canvas (ADR 0080 amendment,
 * design 2026-07-09-f5b-pdf-official-view-design.md).
 */
export function PdfCanvas({ documentId }: { documentId: string }) {
  const pdf = useDocumentPdfStatus(documentId, true);

  if (pdf.status === 'ready' && pdf.url) {
    return (
      <iframe
        className={styles.frame}
        title="Documento oficial (PDF)"
        src={pdf.url}
      />
    );
  }

  if (pdf.status === 'failed') {
    return (
      <div role="alert" className={styles.state}>
        <p className={styles.stateTitle}>Não foi possível gerar o PDF oficial.</p>
        <button type="button" className={styles.retry} onClick={pdf.retry}>
          Tentar novamente
        </button>
      </div>
    );
  }

  return (
    <div role="status" aria-live="polite" className={styles.state}>
      Gerando o PDF oficial…
    </div>
  );
}
```

`PdfCanvas.module.css` (match the workspace canvas idiom — implementer: mirror token usage from `DocumentWorkspacePage.module.css`, e.g. surface/border/text tokens already used there):

```css
.frame {
  width: 100%;
  height: 100%;
  min-height: 60vh;
  border: none;
  background: var(--surface-raised);
}

.state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--space-3);
  min-height: 40vh;
  color: var(--text-muted);
}

.stateTitle {
  margin: 0;
  color: var(--text-primary);
}

.retry {
  border: 1px solid var(--border-default);
  background: var(--surface-raised);
  color: var(--text-primary);
  padding: var(--space-2) var(--space-4);
  border-radius: var(--radius-md);
  cursor: pointer;
}
```

(If any variable name doesn't exist in the app's token set, use the exact tokens `DocumentWorkspacePage.module.css` uses — copy, don't invent.)

- [ ] **Step 4: Run test to verify it passes**

Run: `npx vitest run src/features/documents/components/workspace/PdfCanvas.test.tsx`
Expected: 3 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/apps/web/src/features/documents/components/workspace/PdfCanvas.tsx frontend/apps/web/src/features/documents/components/workspace/PdfCanvas.module.css frontend/apps/web/src/features/documents/components/workspace/PdfCanvas.test.tsx
git commit -m "feat(approval-fe): F2d.5b D1 — PdfCanvas (official post-approval PDF, pending/failed/ready)"
```

---

### Task 2: Status-keyed canvas branch in `DocumentWorkspacePage` (TDD)

**Files:**
- Modify: `frontend/apps/web/src/features/documents/pages/DocumentWorkspacePage.tsx` (canvas branch, lines ~235-295)
- Test: `frontend/apps/web/src/features/documents/pages/DocumentWorkspacePage.test.tsx` (extend)

- [ ] **Step 1: Write the failing tests**

Extend the existing test file. FIRST read it and reuse its existing fixture builders/mocks exactly (it already mocks queries and renders per-mode fixtures — follow the established pattern; do not build a parallel harness). Mock `PdfCanvas` the same way the file mocks other heavy children if it does so; otherwise mock `useDocumentPdfStatus` at module level with a `ready` state. Add:

```tsx
// New describe block — F2d.5b D1: official-PDF canvas is status-keyed.

it('published (lifecycle): renders PdfCanvas, not the docx read canvas', () => {
  // fixture: doc.status='published', no active instance stage, viewer non-author
  // (or author — mode 'lifecycle' either way)
  // render page
  expect(screen.getByTestId('pdf-canvas')).toBeInTheDocument();      // or getByTitle('Documento oficial (PDF)') if unmocked
  expect(screen.queryByTestId('document-shell')).not.toBeInTheDocument();
});

it('approved (lifecycle, author): renders PdfCanvas', () => {
  // fixture: doc.status='approved', viewer.is_author=true → mode 'lifecycle'
  expect(screen.getByTestId('pdf-canvas')).toBeInTheDocument();
});

it('superseded (lifecycle): keeps the docx read canvas — /view does not serve it', () => {
  // fixture: doc.status='superseded' → mode 'lifecycle' but NOT a viewable status
  expect(screen.queryByTestId('pdf-canvas')).not.toBeInTheDocument();
  expect(screen.getByTestId('document-shell')).toBeInTheDocument();
});

it('under_review (observing): still the docx read canvas, never PdfCanvas', () => {
  // existing observing fixture
  expect(screen.queryByTestId('pdf-canvas')).not.toBeInTheDocument();
  expect(screen.getByTestId('document-shell')).toBeInTheDocument();
});
```

(Test-ids: if the suite mocks `DocumentShell`/`PdfCanvas`, give the mocks `data-testid="document-shell"` / `data-testid="pdf-canvas"`, consistent with how the file already stubs children. The assertions above are the contract; adapt selectors to the file's existing idiom.)

- [ ] **Step 2: Run to verify the new tests fail**

Run: `npx vitest run src/features/documents/pages/DocumentWorkspacePage.test.tsx`
Expected: the 4 new tests FAIL (no PdfCanvas branch yet); all pre-existing tests still PASS.

- [ ] **Step 3: Implement the branch**

In `DocumentWorkspacePage.tsx`:

Add import (with the other workspace component imports, ~line 21):

```tsx
import { PdfCanvas } from '../components/workspace/PdfCanvas';
```

Add the status set near `deriveWorkspaceMode` usage (module scope, above the component):

```tsx
// F2d.5b D1 — statuses whose official PDF the backend serves via
// GET /documents/:id/view (view_service.go viewableStatuses). Keyed on
// status, NOT on mode: 'lifecycle' also covers superseded/obsolete, which
// /view does not serve — those keep the docx read canvas.
const OFFICIAL_PDF_STATUSES = new Set(['approved', 'scheduled', 'published']);
```

Change the canvas fallback branch (currently `: doc.current_revision_id ? <DocumentShell .../> : empty`) to check PDF first. The full final branch chain inside `<main className={styles.canvas}>`:

```tsx
{mode === 'approving' ? (
  /* unchanged approving branch */
) : mode === 'author-editing' || mode === 'author-changes-requested' ? (
  /* unchanged editing branch */
) : OFFICIAL_PDF_STATUSES.has(docStatus) ? (
  <PdfCanvas documentId={documentId} />
) : doc.current_revision_id ? (
  <DocumentShell
    documentId={documentId}
    currentRevisionId={doc.current_revision_id}
    editorMode="readonly"
    author={currentUser?.displayName ?? ''}
  />
) : (
  <div className={styles.canvasEmpty}>Este documento ainda não possui conteúdo para exibir.</div>
)}
```

Note: `approving`/`author-editing`/`author-changes-requested` modes cannot co-occur with an official-PDF status (approving implies an active stage ⇒ `under_review`; author-editing implies draft/no-instance), so branch order is safe; the PDF check sits before the generic read canvas only.

- [ ] **Step 4: Run to verify all pass**

Run: `npx vitest run src/features/documents/pages/DocumentWorkspacePage.test.tsx`
Expected: full file PASS (pre-existing + 4 new).

- [ ] **Step 5: Commit**

```bash
git add frontend/apps/web/src/features/documents/pages/DocumentWorkspacePage.tsx frontend/apps/web/src/features/documents/pages/DocumentWorkspacePage.test.tsx
git commit -m "feat(approval-fe): F2d.5b D1 — status-keyed official-PDF canvas in the workspace"
```

---

### Task 3: D2 — lazy `MetalDocsEditor` in `DocumentShell` (TDD via the runtime chunk assertion)

**Files:**
- Modify: `frontend/apps/web/src/features/documents/components/DocumentShell.tsx` (lines 1-2, 122-142)
- Test (create): `frontend/apps/web/src/features/documents/components/workspace/editorChunk.lazy.test.tsx`

- [ ] **Step 1: Write the failing runtime assertion**

The REAL assertion (the F2d.5 lesson — an ineffective lazy greens while saving zero bytes): mock `@metaldocs/editor-ui` with a module-evaluation flag; rendering the PDF canvas path must NOT evaluate the module; rendering a docx read path MUST (lazily).

```tsx
import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

// Evaluation flag: flips ONLY when the editor-ui module factory actually runs
// (i.e. someone value-imports it). Type-only imports never evaluate a module.
const evaluated = vi.hoisted(() => ({ current: false }));

vi.mock('@metaldocs/editor-ui', () => {
  evaluated.current = true;
  return {
    MetalDocsEditor: (props: { mode: string }) => (
      <div data-testid="metaldocs-editor" data-mode={props.mode} />
    ),
  };
});

// Minimal stub of the signed-URL + file fetch DocumentShell performs.
vi.mock('../../../lib/api', () => ({
  apiFetch: vi.fn(async () => ({ url: 'https://s3/signed.docx' })),
}));

describe('F2d.5b D2 — editor chunk is genuinely lazy', () => {
  it('importing DocumentShell does NOT evaluate @metaldocs/editor-ui', async () => {
    await import('../DocumentShell');
    expect(evaluated.current).toBe(false);
  });

  it('mounting DocumentShell DOES lazily evaluate it (docx read path)', async () => {
    const { DocumentShell } = await import('../DocumentShell');
    vi.stubGlobal('fetch', vi.fn(async () => ({
      ok: true,
      arrayBuffer: async () => new ArrayBuffer(8),
    })) as unknown as typeof fetch);

    render(
      <DocumentShell
        documentId="doc-1"
        currentRevisionId="rev-1"
        editorMode="readonly"
        author="A"
      />,
    );
    expect(await screen.findByTestId('metaldocs-editor')).toBeInTheDocument();
    expect(evaluated.current).toBe(true);
    vi.unstubAllGlobals();
  });
});
```

- [ ] **Step 2: Run to verify it fails**

Run: `npx vitest run src/features/documents/components/workspace/editorChunk.lazy.test.tsx`
Expected: FIRST test FAILS — `evaluated.current` is `true` because `DocumentShell.tsx:2` statically value-imports `MetalDocsEditor`. (Second test passes — that behavior must survive the change.)

- [ ] **Step 3: Implement the lazy split**

`DocumentShell.tsx` — replace lines 1-2:

```tsx
import { Suspense, lazy, useEffect, useState } from 'react';
import type { MetalDocsEditorRef, EditorComment, TrackedChange } from '@metaldocs/editor-ui';
```

Add below the imports (module scope):

```tsx
// F2d.5b D2 — the editor (TipTap/ProseMirror + docx deps) is its own lazy
// chunk. Docx read/edit paths fetch it on canvas mount; the official-PDF
// canvas (PdfCanvas) never does. Type imports above are erased at build and
// pull nothing.
const MetalDocsEditor = lazy(() =>
  import('@metaldocs/editor-ui').then((m) => ({ default: m.MetalDocsEditor })),
);
```

Wrap the editor mount (the `else` branch that builds `body`, line ~122) in `Suspense`, reusing the exact existing loading markup as fallback:

```tsx
  } else {
    body = (
      <Suspense
        fallback={
          <div role="status" aria-live="polite" className={styles.loading}>
            Carregando documento…
          </div>
        }
      >
        <MetalDocsEditor
          ref={editorRef}
          /* ...all existing props unchanged... */
          showRuler={false}
        />
      </Suspense>
    );
  }
```

No other changes — props, ref forwarding, and both `chrome`/bare returns stay identical.

- [ ] **Step 4: Run the assertion + every suite that mounts the editor**

Run:
```
npx vitest run src/features/documents/components/workspace/editorChunk.lazy.test.tsx src/features/documents/pages/DocumentEditorPage.test.tsx src/features/documents/pages/DocumentWorkspacePage.test.tsx src/features/documents/components/workspace/EditorCanvas.test.tsx
```
Expected: all PASS. Known risk: tests that queried the editor synchronously after render may now need `await screen.findBy…` (Suspense resolves in a microtask). Fix ONLY by awaiting the query — never by removing the lazy boundary. If a pre-existing test fails for unrelated reasons, apply the legacy-test rule (repair only contract/invariant guards; report anything deleted).

- [ ] **Step 5: Commit**

```bash
git add frontend/apps/web/src/features/documents/components/DocumentShell.tsx frontend/apps/web/src/features/documents/components/workspace/editorChunk.lazy.test.tsx
git commit -m "feat(approval-fe): F2d.5b D2 — MetalDocsEditor becomes a lazy chunk (PDF canvas never fetches it)"
```

---

### Task 4: Full gates + static-graph and no-backend evidence

**Files:** none created (evidence commands only; output goes to Task 5's evidence.md).

- [ ] **Step 1: Full workspace/documents suite**

Run: `npx vitest run src/features/documents src/features/approval`
Expected: PASS except the one pre-existing known failure `ApprovalCockpitPage ?decision=reject preselect` (documented in F2d.5 S3 evidence; cockpit retires in F2d.7). Any OTHER failure = investigate before proceeding.

- [ ] **Step 2: Typecheck**

Run: `npx tsc --noEmit`
Expected: 0 errors.

- [ ] **Step 3: Static-graph gate (grep)**

Run (repo root): `rg -n "^import \{[^}]*MetalDocsEditor" frontend/apps/web/src`
Expected: matches ONLY in `features/templates/components/TemplateReviewCanvas.tsx` and `features/templates/pages/TemplateEditorPage.tsx` (out-of-scope whitelist). ZERO value-imports under `features/documents/` or `features/approval/`.

- [ ] **Step 4: Zero-backend gate**

Run (repo root): `git status --porcelain`
Expected: every changed path starts with `frontend/` or `docs/superpowers/milestones/.../f5b-pdf-read-canvas/`. Any `internal/`, `api/`, `db/`, `apps/` path = STOP, revert it.

- [ ] **Step 5: Build-level chunk evidence (best-effort)**

Run (from `frontend/apps/web`): `npm run build` — then list `dist/assets` and identify the editor chunk (contains editor-ui/tiptap) as a SEPARATE file from the workspace page chunk. Record chunk names + sizes in evidence.
Known risk: pnpm junction drift may break vite build (memory: `fe-node_modules-junction-drift`). If build fails on that pre-existing environment issue: record as bounded defer with the Task 3 runtime assertion + Step 3 grep as the primary split evidence — do NOT attempt a node_modules repair inside this feature.

---

### Task 5: Independent review, evidence.md, close

- [ ] **Step 1: Dispatch cavecrew-reviewer (independent, sonnet)**

Scope: the full F2d.5b diff (Tasks 1-3 commits). Reviewer must verify import paths exist before flagging absence (S2a false-positive lesson). Checklist: status-keyed gate (not mode-keyed; superseded/obsolete excluded), hook reuse (no new fetch logic), lazy boundary real (no residual value-import), Suspense fallback accessible (`role="status"`), Wine tokens only, no `any`, tests assert behavior not implementation.

- [ ] **Step 2: Address findings** — fix real ones (new commit), reject false positives with evidence.

- [ ] **Step 3: Write evidence.md**

Create `docs/superpowers/milestones/approval-remediation/milestone-2d-workflow-coherence-fe/f5b-pdf-read-canvas/evidence.md`: per-task TDD record (RED→GREEN), gate outputs (suite counts, tsc, grep, git-status, build chunks or bounded defer), reviewer disposition, commits list. Follow the F2d.5 evidence.md shape.

- [ ] **Step 4: Commit evidence**

```bash
git add docs/superpowers/milestones/approval-remediation/milestone-2d-workflow-coherence-fe/f5b-pdf-read-canvas/evidence.md
git commit -m "docs(approval): F2d.5b evidence — D1 PdfCanvas + D2 lazy editor chunk closed"
```

---

## Self-review notes (done at plan time)

- Spec coverage: D1 → Tasks 1-2; D2 → Task 3; zero-backend → Task 4 Step 4; chunk assertion → Task 3 test + Task 4 Steps 3/5; ADR 0080 amendment already applied (commit 9657c0a0) — no task needed.
- `superseded/obsolete` trap covered explicitly (Task 2 test 3).
- Type consistency: `PdfCanvas({ documentId })`, `OFFICIAL_PDF_STATUSES`, `DocumentPdfStatus` all match across tasks.
- `useDocumentPdfStatus` signature verified against source (`(documentID: string, enabled: boolean)`); here always mounted with `enabled: true` because PdfCanvas itself only renders for viewable statuses.
