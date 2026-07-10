# Feature F2d.6 — Evidence (author comment replies in `author-waiting`)

**Intent:** Surface the author's reply-to / resolve affordance on reviewer comment threads in the
`author-waiting` workspace mode (design brief §9.1) — the author responds to review comments during
approval WITHOUT editing document content.

**System-impact gate:** 🟢 Green — `docs/superpowers/analysis/2026-07-09-f6-author-comment-replies-system-impact.md`.
FE-only surfacing of an already-authorized capability; no invariant touched; zero backend / contract /
migration change. Locked constraints: (1) zero backend diff, (2) no content editing, (3) reuse canonical
comment primitives, (4) mode-gated surface (only `deriveWorkspaceMode`'s `author-waiting`).

Base SHA: `ce005dc3` (plan). Design source: brief §9.1 + the ratified sidebar-panel decision (external
editor-native thread UI deferred — see Reopen trigger).

---

## Task 1 — `commentText` extractor · CLOSED
- Commit: `33e4eb3d`.
- `components/workspace/commentText.ts` — pure ProseMirror-node → plain-string walker. Recurses
  `content[]`, collects `text` leaves, joins roots with a space, trims. Non-array / non-object input →
  `''` (fail-safe empty, never throws on malformed body).
- **Why it exists:** `RequestedChangesPanel` renders author + resolve button only — NO thread body text.
  The `author-waiting` canvas is a frozen read-only source view, so the author needs the reviewer's
  actual comment text on-screen to reply meaningfully. The `EditorComment.body` is opaque ProseMirror
  (`DocumentCommentContentNode` is free-form); this extractor is the single place that flattens it.
- TDD: RED first. 3 tests — nested paragraph/text, multi-root join, malformed/empty → `''`.
  `vitest run …/commentText.test.ts` **3/3 PASS**.

## Task 2 — `AuthorCommentsPanel` (presentational) · CLOSED
- Commit: `53610f3f`.
- `components/workspace/AuthorCommentsPanel.tsx` (+ `.module.css`, `.test.tsx`). Pure presentational —
  every write delegated via `onReply` / `onResolve` / `onReopen` callbacks; the panel owns only local
  composer draft text. Groups roots (`parentId == null`) with their replies (`parentId === root.id`,
  sorted by id). Renders thread text via `commentText`, a resolve/reopen toggle per thread, and a
  per-thread reply composer (sr-only `<label htmlFor={reply-${root.id}}>`, submit guards on
  `draft.trim()`, clears on submit). `import type { EditorComment }` — editor-ui import is **type-only**
  (erased at compile; zero runtime edge, no TipTap bytes pulled).
- CSS mirrors `RequestedChangesPanel.module.css` Wine tokens; `replyBtn` uses `--brand` (verified
  `tokens.css:8` = `#6b1f2a`).
- **Separate component, not a `RequestedChangesPanel` fork** — different mode (`author-waiting` vs
  `author-changes-requested`), different affordances (reply composer + thread text vs resolve-only).
  Judged a legitimately distinct surface, not duplication.
- TDD: RED first. 6 tests — thread+reply text render; composer submits `(text, root)` then clears;
  whitespace-only reply does NOT submit; resolve calls `onResolve(root)`; reopen calls
  `onReopen({...root, resolved:true})` (argument asserted — tightened in `88ec9188`); empty state.
  `vitest run …/AuthorCommentsPanel.test.tsx` **6/6 PASS**.

## Task 3 — wire into `DocumentWorkspacePage` (`author-waiting` branch) · CLOSED
- Commit: `6239a9cf`.
- `DocumentWorkspacePage.tsx`: `handleAuthorReply(text, parent)` builds an optimistic `EditorComment`
  (next-id from the current thread max, `parentId = Number(parent.id)`, body wrapped as a single
  paragraph/text node) and calls `commentsHook.reply(reply, parent)`. Persisted display name comes from
  the hook's `authorDisplay` param, not the optimistic object (see review §🟡). New `author-waiting`
  branch in the `contextualPanel` chain renders `<AuthorCommentsPanel comments={commentsHook.comments}
  onReply={handleAuthorReply} onResolve={…resolve} onReopen={…reopen} />`; all other modes' panels
  unchanged. **Canvas branch untouched** — `author-waiting` keeps the read-only `DocumentShell`; the
  panel adds reply/resolve, never a content-edit surface (locked constraint 2 held).
- Reuses the canonical primitives verbatim: `commentsHook.reply` = `createComment` with
  `parent_library_id`; `resolve`/`reopen` = `patchComment` `{ done: true|false }` (locked constraint 3).
- TDD: RED first. `DocumentWorkspacePage.test.tsx` — `author-waiting` renders the comments panel;
  reply/resolve callbacks fire the hook; non-`author-waiting` modes do NOT render it. Suite green.

---

## Feature gates (Task 4, self-verified from clean state)
- **tsc:** `npx tsc --noEmit -p .` → **0 errors**.
- **Zero-backend gate:** `git diff --name-only ce005dc3 HEAD` = **7 files, ALL under
  `frontend/apps/web/src/features/documents`** (2 new workspace TSX + 1 CSS + 2 tests + commentText
  +test, workspace page + its test). No backend, no OpenAPI, no contract, no migration, no non-FE file.
- **Type-only editor-ui edge:** `grep editor-ui AuthorCommentsPanel.tsx` → `import type { EditorComment }`
  only. The new panel pulls zero TipTap runtime bytes.
- **documents suite:** `vitest run src/features/documents` → **274/274 PASS** (41 files).
- **approval suite:** `vitest run src/features/approval` → **163/164** — the sole failure is the
  documented pre-existing `ApprovalCockpitPage.test.tsx:342` (`?decision=reject`), which F2d.6 touches no
  file of (git diff confirms no approval/cockpit file in the feature). Cockpit retires in F2d.7.
  **Not a regression.**

## Bounded defer — real-DB authz corroboration
The milestone row's "Real-DB authz test (author can reply/resolve on own doc; non-`CapDocumentView`
actor cannot)" is a **bounded defer**, per the system-impact analysis §8 (marked *optional
corroboration*). Grounds: the feature adds **zero backend behavior** — reply is `createComment` with a
parent, resolve is `updateComment` with `Done`, both already gating on `authorizeDocumentScope →
CapDocumentView` (targeted-verified at `documents/delivery/http/handler.go:1083,1116,1201-1206` during
the gate), which the author already holds on their own document. A new real-DB test would assert
already-covered, unchanged backend authz — no new code path to guard. The FE surface is fully covered by
the vitest component/page tests above.

## Independent reviews (cavecrew-reviewer)
- Task 1 (commentText): clean.
- Task 2 (AuthorCommentsPanel): clean.
- Task 3 (workspace wiring): clean.
- **Whole-feature final review (`ce005dc3..HEAD`): "Feature is clean and ships. All 4 locked
  constraints held."** Two low-severity notes, disposed:
  - 🔵 nit — reopen test asserted `toHaveBeenCalled()` without the argument. **Folded** (`88ec9188`):
    now `toHaveBeenCalledWith({ ...root, resolved: true })`, mirroring the resolve test.
  - 🟡 risk — `handleAuthorReply` author `currentUser?.displayName ?? ''` yields a blank author if
    `currentUser` is missing. **Rejected with reasoning:** the optimistic `EditorComment.author` is
    render-only and corrected on refetch — the PERSISTED display name comes from the hook's
    `authorDisplay` param, not this object. `?? ''` is also the file-wide established idiom (the
    `DocumentShell` `author=` prop uses the identical expression). Fabricating a placeholder (`'Unknown'`)
    would inject a fake identity — a **no-fallback-principle violation** — and diverge from convention.
    No change; documented here.

---
## F2d.6 — FEATURE COMPLETE
Task 1 (`33e4eb3d`) · Task 2 (`53610f3f`) · Task 3 (`6239a9cf`) · nit (`88ec9188`). The author can now
read reviewer comment threads and reply / resolve them in `author-waiting` mode without any content-edit
surface. Reuses `createComment`(reply) / `updateComment`(resolve) verbatim; zero backend diff; editor-ui
import type-only. **NOT pushed.**

**Reopen trigger (ratified refusal, not a defer):** a "content-readonly / comments-interactive" editor
mode that reuses `@metaldocs/editor-ui`'s native thread UI in-canvas (instead of a sidebar panel) —
deferred because it needs a new external-package editor capability. The sidebar-panel surface was the
ratified brief §9.1 decision; revisit if the editor package gains a read-only-content thread mode.

Next: F2d.7 (cockpit retirement) → F2d.8 (UI-driven live QA close).
