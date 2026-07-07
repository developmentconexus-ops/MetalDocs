# F3 — Cockpit = editor shell (C1): kill standalone canvas + writable session

> **Milestone:** M2c approval-screen-fe · **Consumer:** the approval cockpit route
> (`approvals/:documentId`) and the author editor page, which must both mount the **same**
> editor canvas; F4 (sidebar IA) then wires the real approval sidebar into the cockpit frame.
> **Status:** Approved — 2026-07-07. Approval line below.

## Problem (runtime truth today)

`ReviewDocumentCanvas` (`features/approval/components/ReviewDocumentCanvas.tsx`) mounts
`MetalDocsEditor mode="review"` **and holds a writer session + autosave**
(`useDocumentSession({ enabled: status === 'under_review' })` at line 35, `useDocumentAutosave`
at 45) — reviewers can write new document revisions during approval (the **W2 writable-session
vector**). It duplicates, near-verbatim, the buffer-fetch + editor-mount logic in
`DocumentEditorPage.tsx:109-154,519-536`. The approval canvas is a cramped A4 box inside a tab,
not the real editor cockpit (violates **C1: the cockpit IS the editor shell**).

## Consumer contract (what downstream requires, defined before producer)

### 1. `DocumentShell` — the shared editor-canvas region (new)

`features/documents/components/DocumentShell.tsx`. Presentational + buffer-owning; **no writer
session, no autosave of its own** — autosave is an injected callback (absent → no persistence).

```tsx
export interface DocumentShellProps {
  documentId: string;
  currentRevisionId: string | null;          // shell fetches the buffer from the signed URL
  editorMode: 'document-edit' | 'readonly' | 'review';
  editorRef?: React.Ref<MetalDocsEditorRef>;
  author: string;
  comments?: EditorComment[];
  onCommentsChange?: (c: EditorComment[]) => void;
  onCommentAdd?: (c: EditorComment) => void;
  onCommentResolve?: (c: EditorComment) => void;
  onCommentDelete?: (c: EditorComment) => void;
  onCommentReply?: (reply: EditorComment, parent: EditorComment) => void;
  onAutoSave?: (buf: ArrayBuffer) => void | Promise<void>;   // ABSENT ⇒ editor never persists
  onChange?: () => void;
  onTrackedChangesChange?: (c: TrackedChange[]) => void;      // F1 surface; F4 consumes
  chrome?: { center?: React.ReactNode; right?: React.ReactNode }; // optional EditorChrome header
  notice?: React.ReactNode;                                  // optional banner above the editor
}
```

- **Owns:** the buffer state machine (`undefined` loading / `null` error / `ArrayBuffer` ready)
  fetched from `signedRevisionURL(documentId, currentRevisionId)` (the exact logic lifted from
  `DocumentEditorPage`/`ReviewDocumentCanvas` — single source now), the loading/error render,
  and the `MetalDocsEditor` mount. When `chrome` is supplied it wraps the editor in
  `EditorChrome center/right`; when omitted it renders the editor bare (cockpit case — the
  cockpit frame already shows the doc header).
- **Never** imports `useDocumentSession` or `useDocumentAutosave`. Persistence is the caller's
  concern, injected via `onAutoSave`.

### 2. Author page unchanged (zero visual regression)

`DocumentEditorPage.tsx` delegates its `<main class=canvas>…EditorChrome…MetalDocsEditor…</main>`
region to `DocumentShell`, passing `editorMode={canEditContent ? 'document-edit' : 'readonly'}`,
`chrome={{ center: <code+title+status>, right: <autosave+submit> }}`, `onAutoSave={handleSave}`,
its comment handlers, `onChange={handleEditorChange}`, and `editorRef`. The rail (back button) and
`ArtifactMetaSidebar` stay owned by the page. **`DocumentEditorPage.test.tsx` stays green.**

### 3. `ApprovalCockpitPage` — cockpit mounts the shell (renamed from `SignoffDetailPage`)

`features/approval/pages/ApprovalCockpitPage.tsx`. Keeps the `ArtifactApprovalScreen` two-pane
frame + the existing decision sidebar (`decisionExtras`, decision model, dialogs) **untouched
in F3** (F4 rebuilds the sidebar). Its `main` slot renders `DocumentShell` (no `chrome` — the
frame owns the header) in the **resolved editor mode**, replacing `ReviewDocumentCanvas` + the A4
box. The Comentários tab stays.

**Mode resolution (from the approval instance DTO):**

| Active stage (`status === 'active'`) | Current user | `editorMode` |
|---|---|---|
| `stage_kind === 'review'` | eligible actor (`user_id` ∈ active stage `actors`, actor `status` active/waiting) | `'review'` (suggesting) |
| `stage_kind === 'approval'` | any | `'readonly'` (viewing) |
| review or approval | non-eligible / oversee observer | `'readonly'` |
| no active stage / instance absent | — | `'readonly'` |

- **W2 fix (non-negotiable):** the approval feature mounts **no** `useDocumentSession` and **no**
  `useDocumentAutosave` after this feature. `DocumentShell` in `'review'` mode is passed **no**
  `onAutoSave` → suggestions are client-side (surfaced via `onTrackedChangesChange`, consumed by
  F4) + comments persist through the comments API; the reviewer never writes a document revision.
  Verification gate: `Grep 'useDocumentSession|useDocumentAutosave' frontend/apps/web/src/features/approval` → **zero**.
- **Signoff unchanged:** approval-stage `signOff` still reads `content_hash`/`revision_version`
  from `useControlledDocumentActiveDocumentQuery` (already in the page). The old
  `flushSave`→`canvasRef` pre-signoff save path is **deleted** with the canvas (nothing to flush —
  there is no writable session).

### 4. Visibility 404 → standard not-found screen

When the instance/document read returns the visibility 404 (`not_found.instance_not_visible`, or
the existing `NOT_FOUND` on the document detail), the cockpit renders the **standard inline
not-found screen** (`role="alert"`, PT-BR "Documento não encontrado.") — **not** an error toast.
(Matches today's `docQuery.isError` alert branch at `SignoffDetailPage.tsx:135-141`; F3 keeps that
shape and ensures the visibility-scoped 404 lands there, not on a toast.)

## Non-goals

- **Approval sidebar rebuild** (StageContextHeader / single timeline / IntegrityDisclosure /
  DecisionFooter / SuggestionList) — that is **F4**. F3 leaves `ArtifactApprovalScreen` +
  `DocumentApprovalExtras` + the duplicated "Fluxo de aprovação" band exactly as they are.
- **Rendering tracked-change suggestions in the UI** — F3 only wires `onTrackedChangesChange`
  through; F4's `SuggestionList` renders them.
- **Durable suggestion persistence** — server-authoritative suggestion-resolution gate remains the
  program-level HS-2 bounded defer; F3 review mode is a live-session suggesting affordance feeding
  the verdict + comments (which do persist).
- **Extracting the outer page frame / rail / meta-sidebar into the shell** — the shell is the
  canvas region only; each page keeps its own outer frame (author rail+meta sidebar; cockpit
  `ArtifactApprovalScreen`). Deliberate: the frames genuinely differ; the canvas is the real dup.
- **Changing `MetalDocsEditor` internals or the `mode` semantics** — reuse the existing
  `document-edit`/`readonly`/`review` modes.

## Interview record (B1.5) — resolved design questions

| # | Question | Finding (runtime truth, file:line) | Decision |
|---|----------|-----------------------------------|----------|
| 1 | Shell = outer page frame, or canvas region only? | Author frame = rail + `ArtifactMetaSidebar` (`DocumentEditorPage.tsx:461-595`); cockpit frame = `ArtifactApprovalScreen` two-pane (`ArtifactApprovalScreen.tsx:114-208`). The **only** near-verbatim duplication is buffer-fetch + `MetalDocsEditor` mount (`DocumentEditorPage.tsx:109-154,519-536` ≈ `ReviewDocumentCanvas.tsx:59-143`). | Shell = **canvas region** (chrome slot + buffer + editor). Frames stay page-owned. Global-max: extract the real dup, don't force a shared frame the two screens don't share. |
| 2 | Does the cockpit drop `ArtifactApprovalScreen`? | F4 explicitly removes the duplicated flow band "at `ArtifactApprovalScreen.tsx:176,196` usage in cockpit" — i.e. the frame is still in use through F4. | **Keep** `ArtifactApprovalScreen` in F3; embed the shell in its `main` slot. F4 owns sidebar changes. |
| 3 | How does `'review'` mode persist without a writer session? | `ReviewDocumentCanvas` persisted via autosave (the W2 vector). Comments persist via the comments API independently (`useDocumentComments`). Suggestions are client-side (F1 track-changes surface). | Review mode: **no autosave**, comments persist via comments API, suggestions surface via `onTrackedChangesChange`. Durable suggestion persistence = HS-2 bounded defer. |
| 4 | How is actor eligibility determined FE-side? | No eligibility field in the DTO; server enforces (`signoff.not_eligible` 403). Active stage `actors[].user_id` + `actor.status` (active/waiting) is the visible signal (`api-types` `ApprovalStageActorResponse`). | Eligible ⇔ current user id ∈ active stage `actors` with actor `status` ∈ {active, waiting}. Errs to `readonly` when unknown (fail-safe: no writable affordance without positive eligibility). |
| 5 | Where does the active stage / stage_kind come from? | `ApprovalInstance…Response.stages[].status` (`active`) + `stages[].stage_kind` ('review'\|'approval', optional) (`api-types:3338-3357`). Loaded via `getInstance(documentId)` through the adapter (`useDocumentApprovalArtifact`). | Resolve mode from `instance.stages.find(s => s.status==='active')`. `stage_kind` absent ⇒ treat as `approval` (readonly) — fail-safe. |
| 6 | Visibility 404 surface today? | `docQuery.isError` → inline `role="alert"` (`SignoffDetailPage.tsx:135-141`); no toast. `not_found.instance` message exists (`errorMessages.ts:111`). | Keep inline not-found screen; ensure the visibility-scoped 404 lands there, not a toast. |

## Validation Gate

- **New/renamed test `ApprovalCockpitPage.test.tsx`** (house pattern: `MemoryRouter` +
  `Route path="/approvals/:documentId"`, `QueryClient({ retry:false })`, spy the api fns, mock
  `@metaldocs/editor-ui` `MetalDocsEditor`, mock the session/autosave hooks to **assert they are
  never called** in the approval feature):
  - **approval stage** → `DocumentShell` mounts `MetalDocsEditor` with `mode="readonly"`; assert
    `useDocumentSession`/`useDocumentAutosave` mocks **not called** (spy call-count 0).
  - **review stage + eligible actor** → editor `mode="review"`; still **no** session/autosave.
  - **non-eligible / oversee** on a review stage → `mode="readonly"`.
  - **visibility 404** (`not_found.instance_not_visible` / document `NOT_FOUND`) → the inline
    not-found `role="alert"` screen renders; **no** toast spy fired.
- **`DocumentShell.test.tsx`** — buffer state machine (loading → ready → error), `onAutoSave`
  absent ⇒ editor mounts and no persistence path is exercised; `chrome` present ⇒ header renders,
  absent ⇒ bare editor.
- **Regression:** `DocumentEditorPage.test.tsx` stays green (author page unchanged behavior).
- `ReviewDocumentCanvas.tsx` **and** `ReviewDocumentCanvas.test.tsx` deleted.
- `Grep 'useDocumentSession|useDocumentAutosave' frontend/apps/web/src/features/approval` → **zero**.
- `vitest run` (targeted files) PASS; package typecheck clean.

## Approval

- **Contract approved:** 2026-07-07 (main session, per ratified master plan §F3 + design spec §8
  C1/W2). Consumer = the cockpit route + the author editor page; operator holds the HS-1 close gate.
