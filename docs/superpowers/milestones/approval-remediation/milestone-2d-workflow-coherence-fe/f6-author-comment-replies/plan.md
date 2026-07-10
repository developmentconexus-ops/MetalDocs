# F2d.6 Author Comment Replies — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement
> this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** In the `author-waiting` workspace mode, give the author a sidebar panel to read reviewer
comment threads and reply to / resolve them — without ever editing document content.

**Architecture:** FE-only, zero backend. A new presentational `AuthorCommentsPanel` renders comment
threads (root + replies) with a reply composer and resolve/reopen actions, delegating all writes to the
already-wired `useDocumentComments` hook (`reply`, `resolve`, `reopen`). It is threaded into
`WorkspaceSidebar.contextualPanel` for `author-waiting` mode in `DocumentWorkspacePage` — the same slot
F2d.5 S2b used for `RequestedChangesPanel`. Reply/resolve gate on the author's already-held
`CapDocumentView` (system-impact gate GREEN, `docs/superpowers/analysis/2026-07-09-f6-author-comment-replies-system-impact.md`).

**Tech Stack:** React 18, react-router-dom, TanStack Query, vitest + Testing Library, CSS modules (Wine
tokens). Reused: `useDocumentComments`/`useDocumentCommentsQuery` (`features/documents/hooks/editor`,
`features/documents/queries`), `EditorComment` type (`@metaldocs/editor-ui`, type-only),
`WorkspaceSidebar.contextualPanel`, `deriveWorkspaceMode` `author-waiting` branch.

**Locked constraints (from the gate, §10):**
1. Zero backend diff — reuse `reply`/`resolve`/`reopen` verbatim (`git status`/grep gate).
2. No content editing — reply + resolve on comment threads ONLY; never expose document-body editing.
3. Reuse canonical comment primitives — no new hook, no bespoke fetch/mutation.
4. Mode-gated surface — panel appears in `author-waiting`; no new eligibility derivation outside
   `deriveWorkspaceMode`.

**Reopen trigger (flagged, NOT built):** a "content-readonly / comments-interactive" editor mode reusing
`@metaldocs/editor-ui`'s native thread UI would retire this sidebar panel — deferred because it needs an
external workspace-package capability and the design brief §9.1 ratified the sidebar panel.

**Key anchors (verified 2026-07-09):**
- `useDocumentComments(documentID, authorDisplay)` → `{ comments, loading, loadError, add, resolve,
  reopen, reply, remove, retry, setComments }` (`features/documents/hooks/editor/useDocumentComments.ts:11`).
  `reply(replyC: EditorComment, parent: EditorComment)`; `resolve(c)`; `reopen(c)`.
- Reply serialization: `createComment(documentID, { library_comment_id: reply.id, parent_library_id:
  parent.id, author_display, content: reply.body })` (`queries/useDocumentCommentsQuery.ts:85-90`).
- `EditorComment` shape (from `rowToEditorComment`, `useDocumentCommentsQuery.ts:15-24`): `{ id: number,
  parentId?: number, author: string, createdAt?: string, body: unknown (ProseMirror node[]), resolved:
  boolean }`.
- `DocumentCommentContentNode` = free-form `{ [key: string]: unknown }` (`lib/api-types/index.d.ts:2736`);
  `content: DocumentCommentContentNode[]`. Body text lives in ProseMirror `text` leaves.
- `DocumentWorkspacePage.tsx`: `commentsHook` already built (`commentsHook.comments` line ~280,
  `commentsHook.reply` line ~285); `canUseComments = docStatus === 'draft' || 'under_review'` (line 219);
  `contextualPanel` populated only for `author-changes-requested` (lines 207-217); `author-waiting`
  canvas = read-only `DocumentShell` (lines 295-304); `mode` from `deriveWorkspaceMode`; `currentUser`
  available for author display.
- `WorkspaceSidebar.tsx`: `WorkspaceSidebarProps.contextualPanel?: ReactNode` (line 37), rendered between
  timeline and footer (line 122).
- `deriveWorkspaceMode` author-waiting: `viewer?.is_author === true && docStatus === 'under_review'`
  (`features/approval/lib/workspaceMode.ts:70-72`).
- Fixture idiom: `makeComment(overrides)` + `QueryClientProvider` wrapper `retry:false` + mocked `fetch`
  (`hooks/editor/__tests__/useDocumentComments.add.test.tsx:15-34`,
  `.load.test.tsx:37-49` for thread/parent shape); `RequestedChangesPanel.test.tsx:18-26` for panel-props
  idiom.

---

## File Structure

- **Create:** `frontend/apps/web/src/features/documents/components/workspace/AuthorCommentsPanel.tsx`
  — presentational thread panel (root+replies, reply composer, resolve/reopen). Owns no query/mutation
  state; all writes via injected callbacks.
- **Create:** `frontend/apps/web/src/features/documents/components/workspace/AuthorCommentsPanel.module.css`
  — Wine-token styles, mirroring `RequestedChangesPanel.module.css` card idiom.
- **Create:** `frontend/apps/web/src/features/documents/components/workspace/AuthorCommentsPanel.test.tsx`
  — unit tests (render threads with text, reply composer submits, resolve/reopen, empty state).
- **Create:** `frontend/apps/web/src/features/documents/components/workspace/commentText.ts`
  — pure `commentText(body: unknown): string` extractor (walk ProseMirror nodes, concat `text` leaves).
- **Create:** `frontend/apps/web/src/features/documents/components/workspace/commentText.test.ts`
  — extractor unit tests.
- **Modify:** `frontend/apps/web/src/features/documents/pages/DocumentWorkspacePage.tsx` — thread
  `AuthorCommentsPanel` into `contextualPanel` for `author-waiting`.
- **Modify:** `frontend/apps/web/src/features/documents/pages/DocumentWorkspacePage.test.tsx` — author-waiting
  renders the panel; other modes don't.

---

## Task 1: `commentText` body-text extractor

**Files:**
- Create: `frontend/apps/web/src/features/documents/components/workspace/commentText.ts`
- Test: `frontend/apps/web/src/features/documents/components/workspace/commentText.test.ts`

- [ ] **Step 1: Write the failing test**

```ts
import { describe, expect, it } from 'vitest';
import { commentText } from './commentText';

describe('commentText', () => {
  it('extracts text from a paragraph of text leaves', () => {
    const body = [{ type: 'paragraph', content: [{ type: 'text', text: 'Corrigir a seção 3.' }] }];
    expect(commentText(body)).toBe('Corrigir a seção 3.');
  });

  it('concatenates nested + multi-node content', () => {
    const body = [
      { type: 'paragraph', content: [{ type: 'text', text: 'Olá ' }, { type: 'text', text: 'João' }] },
      { type: 'paragraph', content: [{ type: 'text', text: 'segunda linha' }] },
    ];
    expect(commentText(body)).toBe('Olá João segunda linha');
  });

  it('returns empty string for a body with no text leaves (fail-closed, no crash)', () => {
    expect(commentText([{ type: 'horizontalRule' }])).toBe('');
    expect(commentText([])).toBe('');
    expect(commentText(undefined)).toBe('');
    expect(commentText('not-an-array')).toBe('');
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run (from `frontend/apps/web`): `npx vitest run src/features/documents/components/workspace/commentText.test.ts`
Expected: FAIL — `commentText` not defined.

- [ ] **Step 3: Write minimal implementation**

```ts
// F2d.6 — extract display text from an opaque comment body (ProseMirror node[]).
// The API stores comment content as free-form DocumentCommentContentNode[]
// (lib/api-types: `{ [key: string]: unknown }`); text lives in `text` leaves.
// Fail-closed: unknown/empty shapes yield '' (never throw) — no-fallback friendly.
export function commentText(body: unknown): string {
  const walk = (node: unknown): string => {
    if (!node || typeof node !== 'object') return '';
    const n = node as { text?: unknown; content?: unknown };
    if (typeof n.text === 'string') return n.text;
    if (Array.isArray(n.content)) return n.content.map(walk).join('');
    return '';
  };
  if (!Array.isArray(body)) return '';
  return body.map(walk).join(' ').trim();
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `npx vitest run src/features/documents/components/workspace/commentText.test.ts`
Expected: PASS (3 tests).

- [ ] **Step 5: Commit**

```bash
git add frontend/apps/web/src/features/documents/components/workspace/commentText.ts frontend/apps/web/src/features/documents/components/workspace/commentText.test.ts
git commit -m "feat(approval-fe): F2d.6 — commentText body extractor"
```

---

## Task 2: `AuthorCommentsPanel` component

**Files:**
- Create: `frontend/apps/web/src/features/documents/components/workspace/AuthorCommentsPanel.tsx`
- Create: `frontend/apps/web/src/features/documents/components/workspace/AuthorCommentsPanel.module.css`
- Test: `frontend/apps/web/src/features/documents/components/workspace/AuthorCommentsPanel.test.tsx`

**Contract:** Pure presentational. Groups `comments` into threads (root = `parentId == null`; replies =
`parentId === root.id`, in `id` order). For each root thread: author + `commentText(body)` + each reply
(author + text); a resolve button (if `!root.resolved`) or reopen (if `root.resolved`); and a reply
composer (textarea + "Responder" submit, disabled when empty). Submitting calls `onReply(text, root)` and
clears the textarea. All writes delegated via callbacks — the panel never calls a mutation directly and
never edits document content (constraint 2). Empty state when no comments.

**Props:**
```ts
export type AuthorCommentsPanelProps = {
  comments: EditorComment[];
  onReply: (text: string, parent: EditorComment) => void;
  onResolve: (comment: EditorComment) => void;
  onReopen: (comment: EditorComment) => void;
};
```

- [ ] **Step 1: Write the failing test**

```tsx
import { fireEvent, render, screen, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { EditorComment } from '@metaldocs/editor-ui';
import { AuthorCommentsPanel } from './AuthorCommentsPanel';

const txt = (t: string) => [{ type: 'paragraph', content: [{ type: 'text', text: t }] }];
const root: EditorComment = { id: 1, author: 'Revisor', body: txt('Ajuste a introdução'), resolved: false };
const reply: EditorComment = { id: 2, parentId: 1, author: 'Autor', body: txt('Feito'), resolved: false };

function renderPanel(comments: EditorComment[], over: Partial<Record<'onReply' | 'onResolve' | 'onReopen', ReturnType<typeof vi.fn>>> = {}) {
  const onReply = over.onReply ?? vi.fn();
  const onResolve = over.onResolve ?? vi.fn();
  const onReopen = over.onReopen ?? vi.fn();
  render(<AuthorCommentsPanel comments={comments} onReply={onReply} onResolve={onResolve} onReopen={onReopen} />);
  return { onReply, onResolve, onReopen };
}

describe('AuthorCommentsPanel', () => {
  it('renders the reviewer thread text and its reply', () => {
    renderPanel([root, reply]);
    expect(screen.getByText('Ajuste a introdução')).toBeInTheDocument();
    expect(screen.getByText('Feito')).toBeInTheDocument();
    expect(screen.getByText('Revisor')).toBeInTheDocument();
  });

  it('reply composer submits text against the root comment then clears', () => {
    const { onReply } = renderPanel([root]);
    const box = screen.getByLabelText(/responder ao comentário/i);
    fireEvent.change(box, { target: { value: 'Vou corrigir' } });
    fireEvent.click(screen.getByRole('button', { name: /responder/i }));
    expect(onReply).toHaveBeenCalledWith('Vou corrigir', root);
    expect((box as HTMLTextAreaElement).value).toBe('');
  });

  it('does not submit an empty/whitespace reply', () => {
    const { onReply } = renderPanel([root]);
    fireEvent.change(screen.getByLabelText(/responder ao comentário/i), { target: { value: '   ' } });
    fireEvent.click(screen.getByRole('button', { name: /responder/i }));
    expect(onReply).not.toHaveBeenCalled();
  });

  it('resolve button calls onResolve for an unresolved thread', () => {
    const { onResolve } = renderPanel([root]);
    fireEvent.click(screen.getByRole('button', { name: /resolver/i }));
    expect(onResolve).toHaveBeenCalledWith(root);
  });

  it('resolved thread shows reopen, calls onReopen', () => {
    const { onReopen } = renderPanel([{ ...root, resolved: true }]);
    fireEvent.click(screen.getByRole('button', { name: /reabrir/i }));
    expect(onReopen).toHaveBeenCalled();
  });

  it('renders an empty state when there are no comments', () => {
    renderPanel([]);
    expect(screen.getByText(/nenhum comentário/i)).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `npx vitest run src/features/documents/components/workspace/AuthorCommentsPanel.test.tsx`
Expected: FAIL — module not found / `AuthorCommentsPanel` undefined.

- [ ] **Step 3: Write the implementation**

`AuthorCommentsPanel.tsx`:
```tsx
import { useState } from 'react';
import type { EditorComment } from '@metaldocs/editor-ui';
import { commentText } from './commentText';
import styles from './AuthorCommentsPanel.module.css';

export type AuthorCommentsPanelProps = {
  comments: EditorComment[];
  onReply: (text: string, parent: EditorComment) => void;
  onResolve: (comment: EditorComment) => void;
  onReopen: (comment: EditorComment) => void;
};

/**
 * AuthorCommentsPanel — F2d.6. Author-waiting sidebar surface (WorkspaceSidebar
 * contextualPanel slot). Read reviewer comment threads and reply / resolve —
 * NEVER edits document content. Pure presentational: every write is delegated
 * via callbacks (onReply/onResolve/onReopen); the panel owns only local composer
 * text. Reuses the existing useDocumentComments mutations at the call site.
 */
export function AuthorCommentsPanel({
  comments,
  onReply,
  onResolve,
  onReopen,
}: AuthorCommentsPanelProps): React.ReactElement {
  const roots = comments.filter((c) => c.parentId == null);
  const repliesOf = (rootId: EditorComment['id']) =>
    comments.filter((c) => c.parentId === rootId).sort((a, b) => Number(a.id) - Number(b.id));

  return (
    <aside className={styles.root} aria-label="Comentários da revisão">
      <div className={styles.header}>
        <span className={styles.title}>Comentários da revisão</span>
      </div>
      {roots.length === 0 ? (
        <p className={styles.emptyState}>Nenhum comentário nesta revisão.</p>
      ) : (
        <ul className={styles.list}>
          {roots.map((root) => (
            <CommentThread
              key={root.id}
              root={root}
              replies={repliesOf(root.id)}
              onReply={onReply}
              onResolve={onResolve}
              onReopen={onReopen}
            />
          ))}
        </ul>
      )}
    </aside>
  );
}

function CommentThread({
  root,
  replies,
  onReply,
  onResolve,
  onReopen,
}: {
  root: EditorComment;
  replies: EditorComment[];
  onReply: (text: string, parent: EditorComment) => void;
  onResolve: (comment: EditorComment) => void;
  onReopen: (comment: EditorComment) => void;
}): React.ReactElement {
  const [draft, setDraft] = useState('');

  const submit = () => {
    const text = draft.trim();
    if (!text) return;
    onReply(text, root);
    setDraft('');
  };

  return (
    <li className={styles.card}>
      <div className={styles.cardMeta}>
        <span className={styles.cardAuthor}>{root.author}</span>
        {root.resolved ? <span className={styles.resolvedTag}>Resolvido</span> : null}
      </div>
      <p className={styles.cardBody}>{commentText(root.body)}</p>

      {replies.length > 0 ? (
        <ul className={styles.replies}>
          {replies.map((reply) => (
            <li key={reply.id} className={styles.reply}>
              <span className={styles.cardAuthor}>{reply.author}</span>
              <p className={styles.cardBody}>{commentText(reply.body)}</p>
            </li>
          ))}
        </ul>
      ) : null}

      <div className={styles.cardActions}>
        {root.resolved ? (
          <button type="button" className={styles.actionBtn} onClick={() => onReopen(root)}>
            Reabrir
          </button>
        ) : (
          <button type="button" className={styles.actionBtn} onClick={() => onResolve(root)}>
            Resolver
          </button>
        )}
      </div>

      <div className={styles.composer}>
        <label className={styles.srOnly} htmlFor={`reply-${root.id}`}>
          Responder ao comentário de {root.author}
        </label>
        <textarea
          id={`reply-${root.id}`}
          className={styles.textarea}
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          rows={2}
          placeholder="Responder…"
        />
        <button
          type="button"
          className={`${styles.actionBtn} ${styles.replyBtn}`}
          onClick={submit}
          disabled={draft.trim().length === 0}
        >
          Responder
        </button>
      </div>
    </li>
  );
}
```

`AuthorCommentsPanel.module.css` — reuse the Wine tokens present in `RequestedChangesPanel.module.css`
(open that file and mirror `--sp-*`, `--r-*`, `--text`, `--text-muted`, `--border`, `--surface`,
`--accent` usage). Provide classes: `root, header, title, list, card, cardMeta, cardAuthor, cardBody,
resolvedTag, replies, reply, cardActions, actionBtn, replyBtn, composer, textarea, emptyState, srOnly`.
`srOnly` = visually-hidden label (`position:absolute;width:1px;height:1px;overflow:hidden;clip:rect(0 0 0 0)`).
Do NOT invent new token names — if a needed token isn't in `RequestedChangesPanel.module.css` or
`DocumentWorkspacePage.module.css`, reuse the closest existing one.

- [ ] **Step 4: Run test to verify it passes**

Run: `npx vitest run src/features/documents/components/workspace/AuthorCommentsPanel.test.tsx`
Expected: PASS (6 tests). `npx tsc --noEmit -p .` → 0 errors.

- [ ] **Step 5: Commit**

```bash
git add frontend/apps/web/src/features/documents/components/workspace/AuthorCommentsPanel.tsx frontend/apps/web/src/features/documents/components/workspace/AuthorCommentsPanel.module.css frontend/apps/web/src/features/documents/components/workspace/AuthorCommentsPanel.test.tsx
git commit -m "feat(approval-fe): F2d.6 — AuthorCommentsPanel (thread + reply composer + resolve/reopen)"
```

---

## Task 3: Wire `AuthorCommentsPanel` into `DocumentWorkspacePage` author-waiting

**Files:**
- Modify: `frontend/apps/web/src/features/documents/pages/DocumentWorkspacePage.tsx`
- Test: `frontend/apps/web/src/features/documents/pages/DocumentWorkspacePage.test.tsx`

**Behavior:** In `author-waiting` mode, set `contextualPanel` to an `<AuthorCommentsPanel>` fed by the
existing `commentsHook`. Reply builds an `EditorComment` from the composer text (fresh id = max existing
`id` + 1; body = a single text paragraph; author = current user display; `parentId` = root id) and calls
`commentsHook.reply(replyComment, root)`. Resolve/reopen call `commentsHook.resolve/reopen`. No change to
the canvas branch (stays read-only `DocumentShell`) — constraint 2. Other modes' `contextualPanel`
unchanged (author-changes-requested keeps `RequestedChangesPanel`).

- [ ] **Step 1: Write the failing test**

FIRST read `DocumentWorkspacePage.test.tsx` in full and REUSE its fixture/mocking idiom (mode fixtures,
query stubs). Stub `AuthorCommentsPanel` like sibling child stubs, exposing the reply hook so the test can
assert wiring:

```tsx
vi.mock('../components/workspace/AuthorCommentsPanel', () => ({
  AuthorCommentsPanel: ({ comments, onReply }: { comments: unknown[]; onReply: (t: string, p: unknown) => void }) => (
    <div data-testid="author-comments-panel">
      <span>threads:{comments.length}</span>
      <button onClick={() => onReply('resposta', { id: 1 })}>reply-probe</button>
    </div>
  ),
}));
```

Add tests (adapt fixtures to the file's helpers):
```tsx
it('author-waiting: renders the AuthorCommentsPanel in the sidebar', () => {
  // fixture: doc.status='under_review', viewer.is_author=true → mode 'author-waiting'
  expect(screen.getByTestId('author-comments-panel')).toBeInTheDocument();
});

it('author-waiting: canvas stays the read-only docx shell, no content editing', () => {
  // same fixture — assert the read-only DocumentShell canvas renders (existing testid),
  // and the editing EditorCanvas is NOT mounted.
  expect(screen.queryByTestId('editor-canvas')).not.toBeInTheDocument(); // adapt to the file's editing-canvas testid
});

it('observing (non-author under_review): no AuthorCommentsPanel', () => {
  // fixture: doc.status='under_review', viewer.is_author=false → mode 'observing'
  expect(screen.queryByTestId('author-comments-panel')).not.toBeInTheDocument();
});

it('author-changes-requested: keeps RequestedChangesPanel, not AuthorCommentsPanel', () => {
  // existing author-changes-requested fixture
  expect(screen.queryByTestId('author-comments-panel')).not.toBeInTheDocument();
});
```
(Adapt testids/fixtures to whatever the file actually uses. The binding intent: AuthorCommentsPanel present
ONLY in author-waiting; canvas remains read-only there; author-changes-requested keeps its own panel.)

- [ ] **Step 2: Run test to verify it fails**

Run: `npx vitest run src/features/documents/pages/DocumentWorkspacePage.test.tsx`
Expected: the new author-waiting-panel tests FAIL; pre-existing PASS.

- [ ] **Step 3: Implement the wiring**

In `DocumentWorkspacePage.tsx`:
- Import: `import { AuthorCommentsPanel } from '../components/workspace/AuthorCommentsPanel';`
- Add a reply-builder near the comment wiring (reuse `commentsHook` + current user display already in scope
  as `currentUser?.displayName ?? ''` — match the exact expression the file already uses for `author`):
```tsx
// F2d.6 — build a reply EditorComment from composer text (fresh monotonic id;
// single text paragraph body) and delegate to the existing reply mutation.
const handleAuthorReply = (text: string, parent: EditorComment) => {
  const nextId = commentsHook.comments.reduce((max, c) => Math.max(max, Number(c.id) || 0), 0) + 1;
  const reply: EditorComment = {
    id: nextId,
    parentId: Number(parent.id),
    author: currentUser?.displayName ?? '',
    body: [{ type: 'paragraph', content: [{ type: 'text', text }] }],
    resolved: false,
  };
  void commentsHook.reply(reply, parent);
};
```
- Where `contextualPanel` is computed (the block at lines ~207-217 that sets it for
  `author-changes-requested`), add the `author-waiting` branch — keep the existing branch intact:
```tsx
const contextualPanel =
  mode === 'author-changes-requested'
    ? (/* existing RequestedChangesPanel JSX — unchanged */)
    : mode === 'author-waiting'
    ? (
        <AuthorCommentsPanel
          comments={commentsHook.comments}
          onReply={handleAuthorReply}
          onResolve={(c) => void commentsHook.resolve(c)}
          onReopen={(c) => void commentsHook.reopen(c)}
        />
      )
    : null;
```
(If the existing code expresses `contextualPanel` differently — e.g. an `if`/`let` — match that structure;
the invariant is: author-changes-requested branch untouched, new author-waiting branch added, all other
modes still `null`.)
- Ensure `EditorComment` is imported as a type in this file (it already imports the type per the anchor
  map — `DocumentWorkspacePage.tsx:3`).

- [ ] **Step 4: Run tests to verify they pass**

Run: `npx vitest run src/features/documents/pages/DocumentWorkspacePage.test.tsx`
Expected: PASS (pre-existing + new). `npx tsc --noEmit -p .` → 0 errors.

- [ ] **Step 5: Commit**

```bash
git add frontend/apps/web/src/features/documents/pages/DocumentWorkspacePage.tsx frontend/apps/web/src/features/documents/pages/DocumentWorkspacePage.test.tsx
git commit -m "feat(approval-fe): F2d.6 — wire AuthorCommentsPanel into author-waiting sidebar"
```

---

## Task 4: Full gates + independent review + evidence.md

**Files:**
- Create: `docs/superpowers/milestones/approval-remediation/milestone-2d-workflow-coherence-fe/f6-author-comment-replies/evidence.md`

- [ ] **Step 1: Run the full gate battery** (from `frontend/apps/web`)
  - `npx tsc --noEmit -p .` → 0 errors.
  - `npx vitest run src/features/documents` → all green.
  - `npx vitest run src/features/approval` → green except the documented pre-existing
    `ApprovalCockpitPage ?decision=reject` fail (F2d.7 territory).
  - Zero-backend gate (from repo root): `git diff --name-only <feature-base> HEAD` shows ONLY
    `frontend/apps/web/src/features/documents` files; no backend/contract/migration file.
  - No-new-editor-value-import gate: `AuthorCommentsPanel` imports `EditorComment` as **type-only**
    (`import type`), so it adds no static editor-ui runtime edge — confirm via grep.

- [ ] **Step 2: Independent review** — dispatch a `caveman:cavecrew-reviewer` over the whole F2d.6 diff
  (base → HEAD). Focus: constraint compliance (zero backend, no content editing, reused primitives,
  mode-gated), fail-closed `commentText`, reply id/body correctness, no assertion hollowing, a11y (labeled
  composer, aria-label on panel). Fix any 🔴/🟡 with a follow-up implementer + re-review before closing.

- [ ] **Step 3: Write `evidence.md`** — mirror the F2d.5b evidence shape: per-task commits, gate outcomes
  (commands + results), the zero-backend proof, independent-review disposition, and the flagged reopen
  trigger (editor-native comments mode). Mark F2d.6 FEATURE COMPLETE. Note NOT pushed.

- [ ] **Step 4: Commit evidence + milestone row DONE marker**

```bash
git add docs/superpowers/milestones/approval-remediation/milestone-2d-workflow-coherence-fe/f6-author-comment-replies/evidence.md docs/superpowers/milestones/approval-remediation/milestone-2d-workflow-coherence-fe/milestone.md
git commit -m "docs(approval): F2d.6 evidence.md + milestone row DONE"
```

---

## Self-Review (writing-plans)

- **Spec coverage:** brief §9.1 (author reply + resolve in author-waiting) → Task 2 (panel) + Task 3
  (wiring). Milestone-row test spec (reply composer visible in author-waiting, no content editing) →
  Task 3 tests. Optional real-DB authz test = bounded defer (asserts already-covered backend behavior;
  system-impact gate §8) — noted, not a blocker.
- **Placeholder scan:** none — every step carries concrete code or an exact command.
- **Type consistency:** `EditorComment` shape (`id: number`, `parentId?: number`, `body: unknown`,
  `resolved: boolean`) used identically in Tasks 1-3; `onReply(text, parent)` signature matches between
  the panel (Task 2) and the wiring builder (Task 3); reply `body` node shape matches the `commentText`
  extractor's expectations (Task 1). `commentText`/`AuthorCommentsPanel` names stable across tasks.
- **Constraint check:** zero backend (Task 4 gate), no content editing (canvas untouched, Task 3),
  reused mutations (Task 3 uses `commentsHook.reply/resolve/reopen`), mode-gated (Task 3 keys on
  `mode === 'author-waiting'`, no new eligibility derivation).
