# F1 — editor-ui adapter: track-changes + comment-marks API

> **Milestone:** M2c approval-screen-fe · **Consumer:** the approval **review sidebar** (F4) and the author **request-changes panel** (F6), which must list, accept, reject, and clear reviewer suggestions without ever touching a `@eigenpal`/`prosemirror` type.
> **Status:** Approved — 2026-07-07. Approval line below.

## Consumer contract (what downstream requires, defined before producer)

The review screen (F3/F4) and author panel (F6) consume `@metaldocs/editor-ui` only. They need a
**neutral, vendor-free** tracked-change surface on `MetalDocsEditorRef`:

1. **List** — `getTrackedChanges(): TrackedChange[]` where
   `TrackedChange = { revisionId: string; author: string; type: TrackedChangeType; excerpt: string }`.
   `revisionId` is a **string** across the wall even though the Word `w:id` is numeric (the wall
   never leaks the vendor's numeric-id contract). `excerpt` is the change's plain text for sidebar
   cards. `type` is a **MetalDocs-owned** literal union (`TrackedChangeType`) — not imported from the
   vendor — so consumers keep exhaustiveness checking in switch/case rendering.
2. **Resolve one** — `acceptChange(revisionId: string)` / `rejectChange(revisionId: string)`. A
   `replacement` (or coalesced) change is two-or-more vendor ids; one call must resolve **every** site
   (deletion half + insertion half + coalesced) so a change is never left half-resolved (integrity).
3. **Resolve all** — `acceptAllChanges()` / `rejectAllChanges()`.
4. **Comment mark** — `removeCommentMark(libraryCommentId: string)` (used when a thread is deleted).
5. **Change signal** — prop `onTrackedChangesChange?: (changes: TrackedChange[]) => void`, fired on
   edit, accept, and reject, so the sidebar re-renders without polling.

## Non-goals

- **Header/footer tracked changes.** Body view only — approval review content is body content
  (same body-only scoping precedent as `templatePlugin` coloring, wiki §12.5).
- **Real ProseMirror transaction proof.** The adapter unit test mocks the vendor at the module
  boundary (jsdom has no real PM); the end-to-end accept/reject-really-mutates-the-docx behavior is
  the vendor's own tested concern and is exercised in the F8 live-QA walkthrough.
- **A visible sidebar UI.** F1 is the ref API only; the sidebar that consumes it is F4.
- **Server-side suggestion state.** Suggestion resolution stays client-authoritative (eigenpal) +
  caught by the freeze hash chain — the server-authoritative gate is the F0/HS-2 bounded defer.

## Interview record (B1.5) — vendor surface feasibility

| # | Question | Finding (runtime truth, file:line) | Decision |
|---|----------|-----------------------------------|----------|
| 1 | Does the vendor expose accept/reject-by-id + extract? | Yes. `@eigenpal/docx-editor-core/prosemirror/commands`: `acceptChangeById(id)`, `rejectChangeById(id)`, `acceptAllChanges()`, `rejectAllChanges()`, `removeCommentMark(id)` (all curried `Command`s). `@eigenpal/docx-editor-core/prosemirror/utils/extractTrackedChanges`: `extractTrackedChanges(state) → { entries, commentToRevision }`. | Adapter is a thin passthrough. |
| 2 | What is the vendor entry shape? | `TrackedChangeEntry` (`dist/utils/comments.d.ts:44`): `type` (17-member union), `text`, `author`, `revisionId:number`, `insertionRevisionId?:number`, `coalescedRevisionIds?:number[]`. Replacement's insertion half carries a DIFFERENT id; coalesced ids map to the same conceptual change. | Map `text→excerpt`, `String(revisionId)`; resolve ALL ids on accept. |
| 3 | How does a command reach the live view without leaking a PM type past the wall? | `DocxEditorRef.getEditorRef().getView(): EditorView`; `.state` is a live getter. Vendor imports are legal INSIDE `MetalDocsEditor.tsx` (already vendor-coupled); the wall is the public surface (`types.ts`/`index.ts`), guarded by the public-surface test. | Confine vendor imports to the adapter; neutral types only on the barrel. |

## Validation Gate

- New test `src/trackChanges.test.tsx` (vitest): `getTrackedChanges()` returns neutral shapes with
  string `revisionId`; `acceptChange` on a `replacement` resolves ids `8,9,10` together;
  `acceptAllChanges` empties the set; `onTrackedChangesChange` fires with the post-accept set. RED
  first (methods absent → `TypeError`), GREEN after implement.
- `npm run test` (editor-ui) PASS incl. the existing `MetalDocsEditor.test.tsx` +
  `MetalDocsEditor.tokens.test.tsx` (no regression).
- `npm run typecheck` clean.
- ACL wall intact: grep `types.ts`/`index.ts` for `@eigenpal|prosemirror|EditorState|EditorView|Command|TrackedChangeEntry` → zero.

## Approval

- **Contract approved:** 2026-07-07 (main session, per ratified master plan §F1; consumer contract
  is the review-sidebar/author-panel need above; operator holds the HS-1 close gate).
