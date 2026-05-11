# Plan 11 · Editor Frontend Stabilization

> **For agentic workers:** Execute task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking. Honor `CLAUDE.md` (metaldocs-frontend skill, no legacy paths). Apply `/simplify` discipline — ACL enforcement + chrome polish only.

**Goal:** Enforce the eigenpal Anti-Corruption Layer on the last consumer (`TemplateEditorPage`), align the stale wiring test with current gating, and stabilize `editor-chrome` (autosave 7-state, aria-live, tokens, baseline tests, JSDoc).

**Architecture:** Three workstreams. (A) `TemplateEditorPage` swaps direct `@eigenpal/docx-js-editor` imports for `MetalDocsEditor`/`MetalDocsEditorRef` from `@metaldocs/editor-ui`. (B) `templatePlugin.wiring.test.tsx` is rewritten to mirror current mode-gating (`template-draft` ⇒ plugin in; `document-edit` ⇒ plugin out). (C) `editor-chrome` widens `AutosaveState` to the 7-state union and surfaces `dirty`/`stale`/`session_lost` distinctly; adds `role="status" aria-live="polite"`; replaces magic px values with tokens where tokens exist; adds RTL coverage (`EditorChrome.test.tsx`); tightens JSDoc on `editorChromeStyles`/`pointer-events`/slot truthy-collapse; adds an eigenpal-version comment guard.

**Tech Stack:** React 18, TypeScript 5.4, Vitest 1.6 + `@testing-library/react`, CSS Modules, design tokens at `frontend/apps/web/src/styles/tokens.css`. No new deps.

**Prerequisite:** Plan 3 must be done. `vendor/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz` is present (verified 2026-05-11).

**Out of scope (do NOT touch this round):**
- editor-ui-eigenpal R-004..R-010 (dormant `createOutlinePlugin`, `mergefieldPlugin.ts` rename, `onLockLost` wiring, ADRs, doc cleanup, dep bump) — deferred.
- Eigenpal 0.3.x features or any new wrapper prop.
- Redesign of `EditorChrome` slot shape.
- Playwright/E2E.

**Push back if asked to:** merge Plan 12 screen work in; redesign slot API; add eigenpal features; promote tokens that do not already exist.

---

## File map

**Modify**
- `frontend/apps/web/src/features/templates/pages/TemplateEditorPage.tsx` — replace `DocxEditor`/`DocxEditorRef`/`createEmptyDocument` import with `MetalDocsEditor`/`MetalDocsEditorRef`; rewire ref + mount.
- `packages/editor-ui/test/templatePlugin.wiring.test.tsx` — rewrite gating assertions.
- `frontend/apps/web/src/features/shared/components/editor-chrome/parts/AutosaveStatus.tsx` — widen union, branch on new states, add `role="status" aria-live` (polite default, `assertive` on error).
- `frontend/apps/web/src/features/shared/components/editor-chrome/parts/AutosaveStatus.module.css` — labels for new states (no new colors needed beyond existing dot variants; reuse `dotIdle`/`dotError`).
- `frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx:184` — remove ternary collapse; pass `autosave.status` directly.
- `frontend/apps/web/src/features/shared/components/editor-chrome/EditorChrome.tsx` — JSDoc on `editorChromeStyles`, on each slot prop (truthy-collapse), on `center` slot (`pointer-events:none`, opt-in rule), and on the wrapper (eigenpal 0.2.0 version comment guard).
- `frontend/apps/web/src/features/shared/components/editor-chrome/EditorChrome.module.css` — replace magic px values with tokens where they exist; leave `// TODO:token` comments where no token maps.
- `frontend/apps/web/src/features/shared/components/editor-chrome/parts/VersionBadge.module.css` — replace `10.5px` with `var(--font-size-2xs)` (9.5 vs 10.5 — see Task C5 note), `2px 6px` → `var(--sp-1) var(--sp-2)` if tolerable; otherwise `// TODO:token`.
- `frontend/apps/web/src/features/shared/components/editor-chrome/parts/AutosaveStatus.module.css` — replace `12px` → `var(--font-size-sm)` (12.5 — see Task C5 note) or leave; `8px` / `14px` likely have no token — `// TODO:token`.

**Create**
- `frontend/apps/web/src/features/shared/components/editor-chrome/EditorChrome.test.tsx` — RTL specs for slot truthy-collapse, autosave 7 state branches, `aria-live` presence, `VersionBadge` passthrough.

**Do not modify**
- `packages/editor-ui/src/MetalDocsEditor.tsx` (mode gating unchanged).
- `packages/editor-ui/src/index.ts` (public surface unchanged for this plan).
- `frontend/apps/web/src/features/shared/components/editor-chrome/parts/VersionBadge.tsx`.

---

## Workstream A — `TemplateEditorPage` consumes `MetalDocsEditor`

### Task A1 · Audit current direct-eigenpal touchpoints in `TemplateEditorPage`

**Files:** read-only — `frontend/apps/web/src/features/templates/pages/TemplateEditorPage.tsx`

- [ ] **Step 1: List every `DocxEditor` / `DocxEditorRef` / `createEmptyDocument` / `@eigenpal/docx-js-editor` usage in the file.**

Run:
```powershell
Grep -n "eigenpal|DocxEditor|createEmptyDocument" -- frontend/apps/web/src/features/templates/pages/TemplateEditorPage.tsx
```
Confirmed reference points (as of 2026-05-11): line 1 (`styles.css`), line 4 (`DocxEditor`, `DocxEditorRef`), line 5 (`createEmptyDocument`), line 61 (`useRef<DocxEditorRef>`), line 66 (`createEmptyDocument()`), and the JSX `<DocxEditor ... ref={editorRef} />` mount site (`grep -n "<DocxEditor" frontend/apps/web/src/features/templates/pages/TemplateEditorPage.tsx`).

- [ ] **Step 2: Note every prop currently passed to `<DocxEditor>`.**

Read the JSX block and write down the prop list. The migration must keep behavior identical: same `mode`, same `documentBuffer`, same `author`, same comments/title handlers, same `onChange`/`onAutoSave` (whichever exists), same `renderTitleBarRight` (if used).

No commit on this step.

### Task A2 · Swap imports + ref type to `MetalDocsEditor`

**Files:** modify `frontend/apps/web/src/features/templates/pages/TemplateEditorPage.tsx`

- [ ] **Step 1: Replace direct eigenpal imports with the adapter import.**

Replace lines 1–5:
```ts
import '@eigenpal/docx-js-editor/styles.css';
import * as React from 'react';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { DocxEditor, type DocxEditorRef } from '@eigenpal/docx-js-editor/react';
import { createEmptyDocument } from '@eigenpal/docx-js-editor/core';
```
With:
```ts
import * as React from 'react';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { MetalDocsEditor, type MetalDocsEditorRef } from '@metaldocs/editor-ui';
```
Rationale: `MetalDocsEditor` already imports `'@eigenpal/docx-js-editor/styles.css'` inside the package (`packages/editor-ui/src/MetalDocsEditor.tsx:3`), so the top-level CSS import is redundant. The wrapper does not re-export `createEmptyDocument` — see Task A3 for the blank-doc replacement.

- [ ] **Step 2: Update the editor ref type.**

Find line 61 (`const editorRef = useRef<DocxEditorRef>(null);`) and change to:
```ts
const editorRef = useRef<MetalDocsEditorRef>(null);
```

- [ ] **Step 3: Audit every `editorRef.current.<member>` call site in this file.**

Run:
```powershell
Grep -n "editorRef" -- frontend/apps/web/src/features/templates/pages/TemplateEditorPage.tsx
```
The adapter ref surface is `{ getDocumentBuffer(): Promise<ArrayBuffer | null>; focus(): void }` (`packages/editor-ui/src/types.ts:29`). If any call site uses `editorRef.current.save()` or any other `DocxEditorRef` member that `MetalDocsEditorRef` does not expose, stop and surface the gap (do not silently change semantics). Switch `save()` calls to `getDocumentBuffer()` where they are equivalent (the wrapper's `getDocumentBuffer` delegates to `inner.save()` per `MetalDocsEditor.tsx:19-22`).

- [ ] **Step 4: Commit.**

```powershell
git add frontend/apps/web/src/features/templates/pages/TemplateEditorPage.tsx
git commit -m "refactor(templates): swap DocxEditor for MetalDocsEditor in TemplateEditorPage`n`nWorkstream A of Plan 11 — enforces the editor-ui ACL on the last direct-eigenpal consumer. Ref type narrows to MetalDocsEditorRef; styles.css import drops (wrapper handles it). Closes editor-ui-eigenpal T-002 first half."
```

### Task A3 · Replace the JSX mount `<DocxEditor>` with `<MetalDocsEditor>`

**Files:** modify `frontend/apps/web/src/features/templates/pages/TemplateEditorPage.tsx`

- [ ] **Step 1: Locate the JSX mount.**

Run:
```powershell
Grep -n "<DocxEditor" -- frontend/apps/web/src/features/templates/pages/TemplateEditorPage.tsx
```

- [ ] **Step 2: Replace `<DocxEditor ... />` with `<MetalDocsEditor mode="template-draft" ... />`.**

Keep every prop that is supported by `MetalDocsEditorProps` (see `packages/editor-ui/src/types.ts:7-27`): `documentBuffer`, `author`, `documentName`, `documentNameEditable`, `onDocumentNameChange`, `comments`, `onCommentsChange`, `onCommentAdd`, `onCommentResolve`, `onCommentDelete`, `onCommentReply`, `renderTitleBarRight`, `sidebarModel`, `externalPlugins`, `onAutoSave`, `showRuler`.

Notes during migration:
- `mode` must be `"template-draft"` (template authoring surface). The wrapper handles plugin gating (templatePlugin) + maps to eigenpal `editing` automatically.
- The current page wires autosave via `useTemplateAutosave` (`TemplateEditorPage.tsx:59`) — pass that hook's commit callback as `onAutoSave={(buf) => autosave.commit(buf)}` (or the exact contract the hook exports — verify by reading the hook). The wrapper applies the 1500ms debounce + `inFlightRef` guard internally; remove any locally implemented debounce timers that duplicate it.
- `editorPlugins` (existing `useMemo([filterTransactionGuard()])` at line 67) maps to `externalPlugins={editorPlugins}`. Confirm `filterTransactionGuard()` returns an `EditorPlugin` shape.
- `blankDoc` from `createEmptyDocument()` (line 66): if the page only used it as the initial `documentBuffer` fallback when `draft.buffer` is undefined, the simpler shape is to pass `documentBuffer={draft.buffer}` and let the wrapper/eigenpal handle empty state. If the page relies on a non-null `ArrayBuffer` always, keep `createEmptyDocument` by adding a re-export later (deferred — out of scope). For this plan, only swap if the page already tolerates `undefined` (check the JSX). If a hard dependency on `createEmptyDocument` exists, stop and flag — do not introduce a re-export of eigenpal `core` (would re-violate the ACL).

- [ ] **Step 3: Run typecheck.**

```powershell
cd frontend/apps/web; npx tsc -p tsconfig.json --noEmit
```
Expected: exit 0. If errors, fix the prop mismatch named in the diagnostic — do not relax types.

- [ ] **Step 4: Run the existing test suite for the templates feature (if any).**

```powershell
cd frontend/apps/web; npx vitest run src/features/templates
```
Expected: green or "no test files found" (acceptable — there are no co-located tests for `TemplateEditorPage`).

- [ ] **Step 5: Commit.**

```powershell
git add frontend/apps/web/src/features/templates/pages/TemplateEditorPage.tsx
git commit -m "refactor(templates): mount MetalDocsEditor in template-draft mode`n`nReplaces direct <DocxEditor> mount with the ACL wrapper. Autosave + plugin gating now flow through MetalDocsEditor; local autosave debounce stays in useTemplateAutosave (commit pipeline). Closes editor-ui-eigenpal T-002 / R-002."
```

### Task A4 · Manual smoke (template authoring)

**Files:** none (manual verification).

- [ ] **Step 1: Start the API.**

```powershell
.\scripts\start-api.ps1
```
Expected: API on `:8081`, no panic.

- [ ] **Step 2: Start the frontend dev server.**

```powershell
cd frontend/apps/web; npm run dev
```

- [ ] **Step 3: Log in and exercise a template editor session.**

Login at `/login` with `admin` / `AdminMetalDocs123!`. Navigate to a template version page. Confirm:
- Editor mounts; no `eigenpal` import errors in console.
- Typing in the canvas fires autosave after ~1500ms; the `AutosaveStatus` indicator transitions `saving → saved`.
- `templatePlugin` sidebar items render (the placeholder chips eigenpal injects in `template-draft` mode).
- Toolbar wine palette + version badge + autosave indicator render via `EditorChrome` slots (no regression).

If any branch fails, stop and diagnose before continuing — do not paper over with adapter changes.

- [ ] **Step 4: No commit (manual verification).**

---

## Workstream B — Stale wiring test rewrite

### Task B1 · Replace incorrect `document-edit` plugin assertion

**Files:** modify `packages/editor-ui/test/templatePlugin.wiring.test.tsx`

- [ ] **Step 1: Read the current test and identify the mismatch.**

Current line 29–34 asserts `document-edit` includes `templatePlugin`. Production at `packages/editor-ui/src/MetalDocsEditor.tsx:55-56` gates `templatePlugin` to `mode === 'template-draft'`.

- [ ] **Step 2: Rewrite the file.**

Replace the whole `describe` block contents with the four assertions below. Keep the existing `vi.mock` block (lines 8–26) unchanged.

```tsx
describe('template plugin wiring', () => {
  it('includes templatePlugin only when mode is template-draft', () => {
    render(<MetalDocsEditor mode="template-draft" author="u1" />);
    const host = screen.getByTestId('plugin-host');
    expect(host.getAttribute('data-plugins')).toBe('1');
    expect(host.getAttribute('data-plugin-names')).toContain('template');
  });

  it('omits templatePlugin in document-edit mode', () => {
    render(<MetalDocsEditor mode="document-edit" author="u1" />);
    const host = screen.getByTestId('plugin-host');
    expect(host.getAttribute('data-plugins')).toBe('0');
    expect(host.getAttribute('data-plugin-names') ?? '').not.toContain('template');
  });

  it('omits templatePlugin in readonly mode', () => {
    render(<MetalDocsEditor mode="readonly" author="u1" />);
    const host = screen.getByTestId('plugin-host');
    expect(host.getAttribute('data-plugins')).toBe('0');
  });

  it('adds the sidebar bridge plugin alongside templatePlugin in template-draft mode', () => {
    render(
      <MetalDocsEditor
        mode="template-draft"
        author="u1"
        sidebarModel={{
          used: ['a'],
          missing: [],
          orphans: [],
          bannerError: false,
          errorCategories: [],
        }}
      />
    );
    const host = screen.getByTestId('plugin-host');
    const names = host.getAttribute('data-plugin-names') ?? '';
    expect(host.getAttribute('data-plugins')).toBe('2');
    expect(names).toContain('template');
    expect(names).toContain('metaldocs-sidebar-model');
  });

  it('includes external plugins regardless of mode', () => {
    render(
      <MetalDocsEditor
        mode="document-edit"
        author="u1"
        externalPlugins={[{ id: 'custom', name: 'custom' } as never]}
      />
    );
    const host = screen.getByTestId('plugin-host');
    expect(host.getAttribute('data-plugins')).toBe('1');
    expect(host.getAttribute('data-plugin-names')).toContain('custom');
  });
});
```

- [ ] **Step 3: Run the test.**

```powershell
cd packages/editor-ui; npx vitest run test/templatePlugin.wiring.test.tsx
```
Expected: 5 passed.

- [ ] **Step 4: Run the full editor-ui suite.**

```powershell
cd packages/editor-ui; npx vitest run
```
Expected: all green.

- [ ] **Step 5: Commit.**

```powershell
git add packages/editor-ui/test/templatePlugin.wiring.test.tsx
git commit -m "test(editor-ui): align templatePlugin wiring test with template-draft gating`n`nRewrites assertions to match production gating in MetalDocsEditor.tsx:55-56 (templatePlugin only in template-draft mode). Closes editor-ui-eigenpal T-003 / R-003."
```

---

## Workstream C — Editor-chrome polish

### Task C1 · Widen `AutosaveState` to the 7-state union

**Files:**
- Modify: `frontend/apps/web/src/features/shared/components/editor-chrome/parts/AutosaveStatus.tsx`
- Modify: `frontend/apps/web/src/features/shared/components/editor-chrome/parts/AutosaveStatus.module.css`

- [ ] **Step 1: Read the source-of-truth union.**

`frontend/apps/web/src/features/documents/hooks/v2/useDocumentAutosave.ts:5`:
```ts
export type AutosaveStatus = 'idle' | 'dirty' | 'saving' | 'saved' | 'stale' | 'session_lost' | 'error';
```
The chrome component will mirror this exactly (renamed `AutosaveState` for backwards-compat with current `TemplateEditorPage` and `editor-chrome/index.ts` exports).

- [ ] **Step 2: Widen the union, add labels, add render branches.**

Edit `AutosaveStatus.tsx`. Replace the type and default labels:
```ts
export type AutosaveState =
  | 'idle' | 'dirty' | 'saving' | 'saved' | 'stale' | 'session_lost' | 'error';

const DEFAULT_LABELS: Record<AutosaveState, string> = {
  idle: 'Salvo',
  dirty: 'Editado',
  saving: 'Salvando…',
  saved: 'Salvo',
  stale: 'Atualização disponível',
  session_lost: 'Sessão perdida',
  error: 'Erro ao salvar',
};
```

Update `AutosaveStatusProps.labels` to `Partial<Record<AutosaveState, string>>` and merge the same way:
```ts
type AutosaveStatusProps = {
  status: AutosaveState;
  labels?: Partial<Record<AutosaveState, string>>;
  className?: string;
};
```

Replace the four-branch body with explicit branches for all seven states. Reuse existing dot CSS classes; add dot variants in CSS where new color semantics are needed:

```tsx
export function AutosaveStatus({ status, labels, className }: AutosaveStatusProps) {
  const lbl = { ...DEFAULT_LABELS, ...(labels ?? {}) };
  const isError = status === 'error' || status === 'session_lost';
  const isWarn = status === 'stale';
  const wrapperClass =
    `${styles.status}` +
    (isError ? ` ${styles.statusError}` : '') +
    (isWarn ? ` ${styles.statusWarn}` : '') +
    (className ? ` ${className}` : '');

  const ariaLive: 'polite' | 'assertive' = isError ? 'assertive' : 'polite';

  return (
    <span className={wrapperClass} role="status" aria-live={ariaLive}>
      {renderIcon(status)}
      {lbl[status]}
    </span>
  );
}

function renderIcon(status: AutosaveState) {
  switch (status) {
    case 'saving':
      return <span className={`${styles.dot} ${styles.dotSaving}`} aria-hidden="true" />;
    case 'saved':
      return <CheckIcon className={styles.check} />;
    case 'error':
    case 'session_lost':
      return <span className={`${styles.dot} ${styles.dotError}`} aria-hidden="true" />;
    case 'stale':
      return <span className={`${styles.dot} ${styles.dotWarn}`} aria-hidden="true" />;
    case 'dirty':
      return <span className={`${styles.dot} ${styles.dotDirty}`} aria-hidden="true" />;
    case 'idle':
    default:
      return <span className={`${styles.dot} ${styles.dotIdle}`} aria-hidden="true" />;
  }
}
```

(Keep `CheckIcon` as-is below.)

- [ ] **Step 3: Add the two new dot variants + warn label color to the CSS module.**

Edit `AutosaveStatus.module.css`. Append:
```css
.statusWarn {
  color: var(--warning);
}

.dotDirty {
  background: var(--text-muted);
}

.dotWarn {
  background: var(--warning);
}
```
No new tokens invented — `--warning` and `--text-muted` already exist in `frontend/apps/web/src/styles/tokens.css:28,40`.

- [ ] **Step 4: Typecheck.**

```powershell
cd frontend/apps/web; npx tsc -p tsconfig.json --noEmit
```
Expected: exit 0. Expect compile errors at `DocumentEditorPage.tsx:184` (ternary collapse mismatched) — those are fixed in Task C2.

- [ ] **Step 5: No commit yet — Task C2 lands together.**

### Task C2 · Remove the autosave ternary collapse at the consumer

**Files:** modify `frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx`

- [ ] **Step 1: Replace the ternary collapse (lines 184–188).**

Old:
```ts
const autosaveState: AutosaveState =
  autosave.status === 'saving' ? 'saving' :
  autosave.status === 'error' ? 'error' :
  autosave.status === 'saved' ? 'saved' :
  'idle';
```
New:
```ts
const autosaveState: AutosaveState = autosave.status;
```
Since `useDocumentAutosave.AutosaveStatus` and the new `AutosaveState` are structurally identical 7-value unions, this is a direct assignment. Keep the `AutosaveState` type alias import (otherwise inline `autosave.status` everywhere it is used downstream).

- [ ] **Step 2: Typecheck.**

```powershell
cd frontend/apps/web; npx tsc -p tsconfig.json --noEmit
```
Expected: exit 0.

- [ ] **Step 3: Commit C1 + C2 together.**

```powershell
git add frontend/apps/web/src/features/shared/components/editor-chrome/parts/AutosaveStatus.tsx frontend/apps/web/src/features/shared/components/editor-chrome/parts/AutosaveStatus.module.css frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx
git commit -m "feat(editor-chrome): widen AutosaveState to 7 states; drop consumer ternary collapse`n`nAutosaveStatus now mirrors useDocumentAutosave's union (dirty/stale/session_lost no longer rendered as 'Salvo'). Adds role=status + aria-live (assertive on error/session_lost, polite otherwise). DocumentEditorPage passes hook status through directly. Closes editor-chrome R-001 + R-002."
```

### Task C3 · Eigenpal version comment guard (R-003)

**Files:** modify `frontend/apps/web/src/features/shared/components/editor-chrome/EditorChrome.module.css`

- [ ] **Step 1: Add a top-of-file version-pin comment.**

Append above the existing block-comment header at line 1:
```css
/* --- Eigenpal coupling pin ---------------------------------------------
   The :global(.ep-root ...) overrides below target eigenpal data-testid
   attributes and SVG fills as shipped by
     vendor/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz
   On any eigenpal version bump:
     1. Re-run the manual smoke checklist in
        wiki/references/eigenpal-controlled-package.md
     2. Re-verify every selector matches at least one node in the
        rendered DOM of TemplateEditorPage + DocumentEditorPage.
     3. Update this pin comment to the new version.
   --------------------------------------------------------------------- */
```

- [ ] **Step 2: Commit.**

```powershell
git add frontend/apps/web/src/features/shared/components/editor-chrome/EditorChrome.module.css
git commit -m "docs(editor-chrome): pin eigenpal-version review trigger via top-of-file comment`n`nMinimal R-003 — any version bump that changes this comment triggers a review of the 17 :global overrides + SVG fills. Closes editor-chrome R-003."
```

### Task C4 · Baseline RTL tests for `EditorChrome` + `AutosaveStatus` + `VersionBadge`

**Files:** create `frontend/apps/web/src/features/shared/components/editor-chrome/EditorChrome.test.tsx`

- [ ] **Step 1: Confirm test runner config.**

The frontend uses `vitest` via `frontend/apps/web/package.json` (check the `test` script). Tests must run under jsdom; `@testing-library/react` + `@testing-library/jest-dom` are already devDeps.

- [ ] **Step 2: Write the test file.**

```tsx
import { describe, it, expect, afterEach } from 'vitest';
import { cleanup, render, screen } from '@testing-library/react';
import { EditorChrome, VersionBadge, AutosaveStatus, type AutosaveState } from './index';
import styles from './EditorChrome.module.css';
import autosaveStyles from './parts/AutosaveStatus.module.css';

afterEach(cleanup);

describe('EditorChrome', () => {
  it('renders only truthy slots', () => {
    render(
      <EditorChrome left={<span data-testid="L">L</span>} right={<span data-testid="R">R</span>}>
        <div data-testid="canvas" />
      </EditorChrome>
    );
    expect(screen.getByTestId('L')).toBeInTheDocument();
    expect(screen.getByTestId('R')).toBeInTheDocument();
    expect(screen.queryByTestId('canvas')).toBeInTheDocument();
    // center and alert are not provided ⇒ no overlay divs
    expect(document.querySelector(`.${styles.overlayCenter}`)).toBeNull();
    expect(document.querySelector(`.${styles.overlayAlert}`)).toBeNull();
  });

  it('collapses falsy slots (null, false, "")', () => {
    render(
      <EditorChrome left={null} center={false} right={''} alert={undefined}>
        <div data-testid="canvas" />
      </EditorChrome>
    );
    expect(document.querySelector(`.${styles.overlayLeft}`)).toBeNull();
    expect(document.querySelector(`.${styles.overlayCenter}`)).toBeNull();
    expect(document.querySelector(`.${styles.overlayRight}`)).toBeNull();
    expect(document.querySelector(`.${styles.overlayAlert}`)).toBeNull();
  });

  it('appends className to the wrapper', () => {
    const { container } = render(
      <EditorChrome className="extra"><div /></EditorChrome>
    );
    expect(container.firstChild).toHaveClass(styles.wrapper);
    expect(container.firstChild).toHaveClass('extra');
  });
});

describe('AutosaveStatus', () => {
  const cases: Array<{ status: AutosaveState; label: string; live: 'polite' | 'assertive' }> = [
    { status: 'idle',         label: 'Salvo',                  live: 'polite' },
    { status: 'dirty',        label: 'Editado',                live: 'polite' },
    { status: 'saving',       label: 'Salvando…',              live: 'polite' },
    { status: 'saved',        label: 'Salvo',                  live: 'polite' },
    { status: 'stale',        label: 'Atualização disponível', live: 'polite' },
    { status: 'session_lost', label: 'Sessão perdida',         live: 'assertive' },
    { status: 'error',        label: 'Erro ao salvar',         live: 'assertive' },
  ];

  for (const c of cases) {
    it(`renders ${c.status} with role=status, aria-live=${c.live}, label "${c.label}"`, () => {
      render(<AutosaveStatus status={c.status} />);
      const el = screen.getByRole('status');
      expect(el).toHaveAttribute('aria-live', c.live);
      expect(el).toHaveTextContent(c.label);
    });
  }

  it('error state applies error class', () => {
    render(<AutosaveStatus status="error" />);
    expect(screen.getByRole('status')).toHaveClass(autosaveStyles.statusError);
  });

  it('stale state applies warn class', () => {
    render(<AutosaveStatus status="stale" />);
    expect(screen.getByRole('status')).toHaveClass(autosaveStyles.statusWarn);
  });

  it('caller labels override defaults per-state', () => {
    render(<AutosaveStatus status="saving" labels={{ saving: 'Custom...' }} />);
    expect(screen.getByRole('status')).toHaveTextContent('Custom...');
  });
});

describe('VersionBadge', () => {
  it('renders children inside a span with the badge class', () => {
    const { container } = render(<VersionBadge>v5</VersionBadge>);
    expect(container.firstChild).toHaveTextContent('v5');
    expect(container.firstChild).toHaveProperty('tagName', 'SPAN');
  });

  it('appends className', () => {
    const { container } = render(<VersionBadge className="extra">REV05</VersionBadge>);
    expect(container.firstChild).toHaveClass('extra');
  });
});
```

- [ ] **Step 3: Run the new tests.**

```powershell
cd frontend/apps/web; npx vitest run src/features/shared/components/editor-chrome/EditorChrome.test.tsx
```
Expected: all green.

- [ ] **Step 4: Commit.**

```powershell
git add frontend/apps/web/src/features/shared/components/editor-chrome/EditorChrome.test.tsx
git commit -m "test(editor-chrome): baseline RTL coverage — slot collapse, autosave 7 states + aria-live, VersionBadge`n`nCloses editor-chrome R-004."
```

### Task C5 · Magic-px → tokens where mapping exists

**Files:**
- Modify: `frontend/apps/web/src/features/shared/components/editor-chrome/EditorChrome.module.css`
- Modify: `frontend/apps/web/src/features/shared/components/editor-chrome/parts/VersionBadge.module.css`
- Modify: `frontend/apps/web/src/features/shared/components/editor-chrome/parts/AutosaveStatus.module.css`

**Rule:** only swap a literal when an existing token at `frontend/apps/web/src/styles/tokens.css` matches exactly. If the literal does not match any token (e.g. `40px` overlay height, `26px` button height, `8px` dot, `14px` check, `2px 6px` badge padding, `10.5px` mono size, `1.2s` pulse animation), leave the value and add a `/* TODO:token */` inline comment next to it. Do NOT invent new tokens.

- [ ] **Step 1: Edit `EditorChrome.module.css`.**

Apply only these swaps (verify each against tokens.css before editing):
- Line `font-size: 13px;` in `.overlayCenter` → leave + `/* TODO:token 13px between sm(12.5) and md(14) */`.
- Line `font-size: 12.5px;` in `.overlayAlert` → `font-size: var(--font-size-sm);`.
- Line `font-size: 15px;` in `.docTitle` → leave + `/* TODO:token 15px between md(14) and lg(22) */`.
- Line `font-size: 12px;` in `.docMeta` → leave + `/* TODO:token 12px between xs(11) and sm(12.5) */`.
- Line `font-size: 12px;` in `.ghostBtn` and `.primaryBtn` → leave + `/* TODO:token */`.
- Hardcoded `color: #fff;` in `.primaryBtn` → `color: var(--text-on-brand);` (token exists, line 12 of tokens.css).
- `width: 26px; height: 26px;` in `.iconBtn` → leave + `/* TODO:token --btn-h-sm not defined */`.
- `border-radius: var(--r-1);` etc. — already token-driven; no change.
- Eigenpal `:global(...)` overrides — DO NOT TOUCH (T-003 is the eigenpal coupling pin; selector + value churn there is its own future plan).

- [ ] **Step 2: Edit `VersionBadge.module.css`.**

- `font-size: 10.5px;` → leave + `/* TODO:token 10.5px between 2xs(9.5) and xs(11) */`.
- `padding: 2px 6px;` → leave + `/* TODO:token sub-sp-1 values */`.
- `color: #fff;` → `color: var(--text-on-brand);`.

- [ ] **Step 3: Edit `AutosaveStatus.module.css`.**

- `font-size: 12px;` → leave + `/* TODO:token */` (see EditorChrome note above).
- `width: 8px; height: 8px;` for `.dot` → leave + `/* TODO:token */`.
- `width: 14px; height: 14px;` for `.check` → leave + `/* TODO:token */`.
- `animation: pulse 1.2s ease-in-out infinite;` → leave + `/* TODO:token motion duration */`.

- [ ] **Step 4: Visual smoke.**

Reload the running dev server (Task A4) and confirm `TemplateEditorPage` + `DocumentEditorPage` look identical to before the swap. If any color/spacing shifts visibly, revert the offending swap.

- [ ] **Step 5: Commit.**

```powershell
git add frontend/apps/web/src/features/shared/components/editor-chrome/EditorChrome.module.css frontend/apps/web/src/features/shared/components/editor-chrome/parts/VersionBadge.module.css frontend/apps/web/src/features/shared/components/editor-chrome/parts/AutosaveStatus.module.css
git commit -m "style(editor-chrome): swap #fff for --text-on-brand + token-gap TODOs`n`nLands the trivial token migrations (color literals matching --text-on-brand and the 12.5px → --font-size-sm swap). Every literal without a clean token mapping carries an inline TODO:token comment for future scope. Closes editor-chrome R-005 within the /simplify boundary."
```

### Task C6 · JSDoc tightening (R-006, R-007, R-009)

**Files:** modify `frontend/apps/web/src/features/shared/components/editor-chrome/EditorChrome.tsx`

- [ ] **Step 1: Expand the per-slot and `editorChromeStyles` JSDoc.**

Replace the props block + `editorChromeStyles` block with the version below. No behavior change.

```tsx
export type EditorChromeProps = {
  /**
   * Slot rendered top-left over eigenpal title bar (back btn, etc.).
   * Falsy values (`null`, `false`, `0`, `''`, `undefined`) suppress the
   * overlay — no wrapper div is rendered. Pass any truthy `ReactNode`
   * to mount the overlay.
   */
  left?: ReactNode;
  /**
   * Slot rendered centered over eigenpal title bar (title, badges, pill).
   * The overlay has `pointer-events: none` so clicks pass through to
   * eigenpal's title bar. Interactive children (buttons, dropdowns) MUST
   * opt back in with `pointer-events: auto` on their own root element,
   * otherwise mouse activation is silently lost (keyboard activation
   * still works — producing inconsistent behavior). See T-007.
   *
   * Truthy-collapse rule applies (see `left`).
   */
  center?: ReactNode;
  /**
   * Slot rendered top-right over eigenpal title bar (autosave, actions).
   * Truthy-collapse rule applies.
   */
  right?: ReactNode;
  /**
   * Optional alert banner rendered just below the 40px title bar.
   * Truthy-collapse rule applies.
   */
  alert?: ReactNode;
  /** The eigenpal editor instance (DocxEditor / MetalDocsEditor). */
  children: ReactNode;
  /** Extra class on the wrapper for page-specific tweaks. */
  className?: string;
};
```

And below the component, replace the `editorChromeStyles` re-export with:
```tsx
/**
 * Re-export of the EditorChrome CSS Module class record.
 *
 * Consumers reach button/text primitives by string property access
 * (e.g. `editorChromeStyles.iconBtn`, `.primaryBtn`, `.docTitle`).
 *
 * **Caveat:** The record is weakly typed — `typeof styles` is the inferred
 * CSS-Module shape, so a typo (`primarybtn`) yields `undefined` at runtime
 * without a TypeScript error unless `noUncheckedIndexedAccess` is enabled.
 * Treat it as an untyped string lookup. Known class keys:
 *   `iconBtn`, `ghostBtn`, `primaryBtn`, `docTitle`, `docSep`, `docMeta`,
 *   `wrapper`, `overlayLeft`, `overlayCenter`, `overlayRight`, `overlayAlert`.
 */
export const editorChromeStyles = styles;
```

- [ ] **Step 2: Typecheck.**

```powershell
cd frontend/apps/web; npx tsc -p tsconfig.json --noEmit
```
Expected: exit 0.

- [ ] **Step 3: Commit.**

```powershell
git add frontend/apps/web/src/features/shared/components/editor-chrome/EditorChrome.tsx
git commit -m "docs(editor-chrome): expand JSDoc on slot truthy-collapse, pointer-events, editorChromeStyles`n`nCloses editor-chrome R-006 (typing caveat documented), R-007 (pointer-events:none contract surfaced on the center slot), R-009 (truthy-collapse rule on every slot prop)."
```

### Task C7 · ADR-gap stub for R-008 (skip the heavy ADR, leave a marker)

**Note:** Per the prompt, R-008 (ADR 0013 — editor-chrome extraction + slot API) is one-liner cleanup; the full ADR belongs to Plan 13 (Doc-comment + ADR sweep, see `wiki/backlog/roadmap.md:144`). Do **not** author a new ADR file here. Instead, link the missing-ADR rows from the wiki to Plan 13.

- [ ] **Step 1: Append a follow-up note to the wiki backlog row.**

Edit `wiki/backlog/editor-chrome-refactor.md` row R-008 (the table cell `Sketch (one line)`) to append: ` Owner: Plan 13 (ADR sweep).` This keeps the row open but pins owner.

- [ ] **Step 2: Commit.**

```powershell
git add wiki/backlog/editor-chrome-refactor.md
git commit -m "docs(wiki): pin editor-chrome R-008 ADR to Plan 13 sweep`n`nKeeps R-008 open without authoring ADR 0013 inside the Plan 11 scope; the ADR sweep plan owns it."
```

---

## Verification

### Task V1 · Full test suite

- [ ] **Step 1: Run editor-ui tests.**

```powershell
cd packages/editor-ui; npx vitest run
```
Expected: all green.

- [ ] **Step 2: Run the web app tests.**

```powershell
cd frontend/apps/web; npx vitest run
```
Expected: all green (the new `EditorChrome.test.tsx` plus pre-existing suites).

- [ ] **Step 3: Run typecheck on both.**

```powershell
cd packages/editor-ui; npx tsc -p tsconfig.json --noEmit
cd ../../frontend/apps/web; npx tsc -p tsconfig.json --noEmit
```
Expected: exit 0 for both.

### Task V2 · Manual smoke (document editor)

- [ ] **Step 1: Open `DocumentEditorPage` in dev.**

Navigate to a published document instance editor. Confirm:
- Autosave indicator transitions visibly through `dirty` (after a keystroke), `saving`, `saved`. No longer collapses to `Salvo`.
- Force a session-lost path (e.g. open the same doc in a second tab to trigger force-release). The status indicator displays `Sessão perdida` with `aria-live="assertive"`.
- Sanity-test screen reader (NVDA on Windows): status changes are announced.

- [ ] **Step 2: Open `TemplateEditorPage`.**

Confirm same behavior as Task A4 still holds post chrome changes.

- [ ] **Step 3: No commit.**

### Task V3 · Wiki + roadmap closeout

- [ ] **Step 1: Dispatch the `wiki-curator` agent.**

Use the Agent tool to invoke `wiki-curator` with the brief: "Plan 11 (editor frontend stabilization) landed. Touched files: TemplateEditorPage.tsx, packages/editor-ui/test/templatePlugin.wiring.test.tsx, editor-chrome (parts/AutosaveStatus.tsx, AutosaveStatus.module.css, EditorChrome.tsx, EditorChrome.module.css, parts/VersionBadge.module.css, EditorChrome.test.tsx), DocumentEditorPage.tsx. Refresh Last verified stamps on wiki/modules/editor-ui-eigenpal.md, editor-ui-eigenpal-tech-debt.md, editor-chrome.md, editor-chrome-tech-debt.md, backlog/editor-ui-eigenpal-refactor.md, backlog/editor-chrome-refactor.md. Mark T-002, T-003 (editor-ui-eigenpal) and T-001..T-007 + T-009 (editor-chrome) as closed where the corresponding R-row landed. Do NOT close T-004..T-008 in editor-ui-eigenpal (deferred). Do NOT close T-008 in editor-chrome (Plan 13)."

- [ ] **Step 2: Update roadmap.**

Edit `wiki/backlog/roadmap.md`:
- Plan 11 row in the execution-order table: `Status: done YYYY-MM-DD` (today's ISO date).
- Plan 11 body section: append a `**PRs:**` line listing the PR(s).
- Bump the `> **Last verified:**` stamp at the top.

- [ ] **Step 3: Final commit.**

```powershell
git add wiki/backlog/roadmap.md
git commit -m "docs(roadmap): mark Plan 11 done`n`nEditor frontend stabilization landed — ACL enforced on TemplateEditorPage, wiring test aligned, editor-chrome autosave widened + a11y + tests + JSDoc."
```

---

## Closes (debt registers)

| Register | Row | Note |
|---|---|---|
| `wiki/modules/editor-ui-eigenpal-tech-debt.md` | T-002 | Closed in Workstream A |
| `wiki/modules/editor-ui-eigenpal-tech-debt.md` | T-003 | Closed in Workstream B |
| `wiki/backlog/editor-ui-eigenpal-refactor.md` | R-002, R-003 | Closed |
| `wiki/modules/editor-chrome-tech-debt.md` | T-001 | Closed in Task C1+C2 |
| `wiki/modules/editor-chrome-tech-debt.md` | T-002 | Closed in Task C1 (aria-live) |
| `wiki/modules/editor-chrome-tech-debt.md` | T-003 | Partially closed in Task C3 (comment-guard only; no automated selector test) |
| `wiki/modules/editor-chrome-tech-debt.md` | T-004 | Closed in Task C4 |
| `wiki/modules/editor-chrome-tech-debt.md` | T-005 | Partially closed in Task C5 (trivial swaps only; remaining gaps carry `TODO:token`) |
| `wiki/modules/editor-chrome-tech-debt.md` | T-006, T-007, T-009 | Closed in Task C6 (JSDoc-only) |
| `wiki/modules/editor-chrome-tech-debt.md` | T-008 | Deferred to Plan 13 |
| `wiki/backlog/editor-chrome-refactor.md` | R-001..R-007, R-009 | Closed |
| `wiki/backlog/editor-chrome-refactor.md` | R-008 | Pinned to Plan 13 |
| `wiki/backlog/editor-ui-eigenpal-refactor.md` | R-004..R-010 | **Deferred** — not in Plan 11 scope per /simplify discipline |

---

## Self-review checklist (run before handoff)

- [ ] Every step has runnable commands and complete code.
- [ ] No `TBD` / `TODO: implement` / "add error handling" placeholders (the `TODO:token` markers in CSS are intentional and explicitly documented).
- [ ] `AutosaveState` union in `AutosaveStatus.tsx` matches `useDocumentAutosave.AutosaveStatus` exactly.
- [ ] `MetalDocsEditorRef` surface (`getDocumentBuffer`, `focus`) is consistent with how `TemplateEditorPage` consumes the ref.
- [ ] Plan does not introduce new eigenpal exports, new wrapper props, or new tokens.
- [ ] Plan 13 ownership of R-008 is explicit.
- [ ] All commits are atomic and message-prefixed (`feat:` / `refactor:` / `test:` / `docs:` / `style:`).
