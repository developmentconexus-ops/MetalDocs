# Design — Eigenpal Review Cockpit + `@eigenpal/docx-editor-react` 1.9 Adoption

> **Date:** 2026-06-23
> **Status:** Design (approved in brainstorm; awaiting spec review → writing-plans)
> **Owner:** MetalDocs maintainer
> **Binding constraint:** Grade-A only. No patch, no workaround, no fork of shared tested
> components. Follow existing MetalDocs coding + backend patterns (runtime/contract truth,
> read-only-tx authz discipline, presign/commit autosave shape, no-fork mandate HS-2).

## 1. Problem & intent

The Revisão (sign-off) cockpit at `/approvals/:documentId` (`SignoffDetailPage`) renders the
document as a **flat rendered PDF** (`useDocumentPdfStatus` → `GET /documents/{id}/view`). A PDF
cannot be reviewed *actively* — the approver cannot comment in place or suggest corrections.

Eigenpal's newly released editor (`@eigenpal/docx-editor-react@1.9.0`) natively provides
**suggesting / tracked-changes** mode and **threaded comments**, both serialized into canonical
OOXML. Replacing the cockpit's PDF panel with the editor in suggesting mode turns Revisão into a
real review surface: the approver reads, comments, and proposes tracked corrections without
destructively editing; the author later accepts/rejects those suggestions natively.

Two independent-but-related changes, executed as **one combined spec** (architecture is well
defined; scope is bounded):

- **A — Dependency migration:** retire the vendored fork, adopt the published 1.9 package.
- **B — Review cockpit:** render the editor (review/suggesting mode) instead of the PDF.

A is a prerequisite for B and is absorbed behind the existing wrapper, so they ship together.

## 2. Grounded facts (verified this session)

| Fact | Source |
|---|---|
| Current dep is a vendored fork tarball wired via `file:` in 3 package.json | `frontend/apps/web/package.json:17`, `packages/editor-ui/package.json:29`, `apps/docx-renderer/package.json:15` |
| New package: `@eigenpal/docx-editor-react@1.9.0`, Apache-2.0, on npm | github.com/eigenpal/docx-editor, npm |
| Suggesting mode: `mode="suggesting"`; tracked changes serialize to `w:ins`/`w:del`, round-trip with Word; accept/reject via `acceptChangeById`/`rejectChangeById`/`acceptAllChanges` from `@eigenpal/docx-editor-core`; `author` + auto timestamp attribution | docx-editor.dev/docs/1.x/guides/tracked-changes |
| Comments live in-OOXML (`w:comment`) **and** are fully controllable via `comments` prop + `onComment*` callbacks (`onCommentsChange`, `onCommentAdd`, `onCommentReply`, `onCommentResolve`, `onCommentDelete`) + `getComments()` | docx-editor.dev/docs/1.x/guides/comments |
| No two-file version-compare / diff engine exists in 1.9 (only in-document tracked changes + `extractTrackedChanges` util) | repo tree scan — no `compare`/`diff` source |
| The `MetalDocsEditor` wrapper already maps app modes → eigenpal modes (`readonly`→`viewing`, `document-edit`→`editing`) and feeds controlled comments | `packages/editor-ui/src/MetalDocsEditor.tsx:61,79-82` |
| Cockpit "Documento" tab today = `<iframe src={pdf.url}>` via `useDocumentPdfStatus(documentId, …)` | `frontend/apps/web/src/features/approval/pages/SignoffDetailPage.tsx:88-106` |
| Working-content save path: browser→MinIO presigned PUT, commit writes a `revisions` row, idempotent on `content_hash`, **If-Match on `revision_version`**, single-writer last-writer-wins | `wiki/diagrams/sequence-edit-autosave.md`; `handler.go:752` (`presignAutosave`), `:790` (`commitAutosave`) |
| Governed REV = `documents.revision_number` (zero-based → `REVxx`); advances on governance/publish, **not** on working saves | `wiki/database/tables/documents.md:112,205` |
| Reject → document returns to `draft` in the same tx; instance keeps `rejected` for audit; author reopens, edits, re-finalizes | `wiki/workflows/approval.md:84-90` |
| Edit session/lock today enabled only for `draft` | `frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx:73,343,463` |

## 3. Source-of-truth model

One **living document**, one content lineage. No parallel review copy is ever created.

- **CD** (`controlled_documents`) — catalog slot (code `DC-RH-001`). Identity; never changes.
- **Governed REV** (`documents.revision_number` → `REV00`, `REV01`…) — a controlled revision of
  the CD. Advances with the **governance lifecycle** (new revision cycle / publish). A reviewer
  adding suggestions does **not** mint a REV.
- **Working content** (`document_revisions` chain; head = `documents.current_revision_id`) —
  the editable docx, save-points linked by `parent_revision_id`. **Every save** (author *or*
  reviewer) advances this chain. This is the single source of truth for content.
- **Inside the working content**, Eigenpal owns the fine-grained edit history: tracked
  suggestions (`w:ins`/`w:del`) carry author + timestamp and round-trip on every save.

Persistence choice (locked: option **(a)**): keep the `document_revisions` save-point chain
(MetalDocs-owned linear working history + server-authoritative `content_hash`); Eigenpal owns the
redline *within* each save. No new revision system, no in-place blob overwrite.

### Comments vs suggestions split (deliberate)

- **Suggestions** → in-OOXML, ride in the saved working revision. No table.
- **Comments** → existing controlled `metaldocs` comments store (`GET/POST/PATCH/DELETE
  /documents/{id}/comments`), fed to 1.9 via the controlled `comments` prop; 1.9's in-OOXML
  comment writes are suppressed (controlled mode).

Rationale: the server needs **queryable** comments (inbox counts, notifications, the cockpit
Comentários tab). Parsing docx blobs for that is the wrong tool; a comment index is the senior
pattern even when the doc could carry them. Suggestions need no server queryability → leave them
in the document.

## 4. Architecture

### §A — Dependency migration (behind the wrapper)

- Replace `@eigenpal/docx-js-editor@0.2.0` (`file:` tarball) with
  `@eigenpal/docx-editor-react@1.9.0` from npm in all 3 package.json. Pin exact; capture
  lockfile integrity. Delete `third_party/eigenpal/*.tgz`. Add Apache-2.0 NOTICE attribution.
- **Anti-corruption layer holds the blast radius:** all import/API churn is absorbed inside
  `packages/editor-ui/src/MetalDocsEditor.tsx`. The wrapper's public props/contract stay stable;
  consumers (`DocumentEditorPage`, the new cockpit mount) are insulated.
- Extend the wrapper mode map: add `review → 'suggesting'` alongside the existing
  `readonly → 'viewing'` and `document-edit → 'editing'`.
- Verify the 1.9 comment `Comment` shape against the wrapper's controlled-comment adapter; adjust
  the adapter inside the wrapper only if the shape moved between 0.2 and 1.9.
- Going forward: dependency updates via Renovate/Dependabot PRs, not hand-vendored tarballs.

### §B — Review cockpit surface

- In `SignoffDetailPage`, the "Documento" tab swaps the PDF `<iframe>` for
  `<MetalDocsEditor mode="review" author={approverDisplay} documentBuffer={workingRevisionDocx}
  comments={…} onComment*={…} ref={editorRef} />`.
- Content source flips from the rendered final PDF to the **working revision docx** — the same
  `signed-url` fetch `DocumentEditorPage` already uses (`GET
  /documents/{id}/revisions/{revisionId}/signed-url` → S3 ArrayBuffer).
- The cockpit stops calling `useDocumentPdfStatus` / `/documents/{id}/view`. (The `/view` PDF
  pipeline and its pre-existing read-only-tx authz defect are no longer on the cockpit path; the
  pipeline itself is left untouched for other consumers.)
- The right decision panel (`ControlledDocumentDetailPanel` + `SignoffDialog`) is unchanged
  (no-fork). The approve/reject deep-link (`?decision=`) auto-open behavior is preserved.

### §C — Reviewer write boundary (the Grade-A core)

The reviewer's suggestions autosave into the working-content `revisions` chain while the document
is `under_review`, **reusing** the existing presign/commit pipeline + If-Match concurrency. Net-new:

1. **Authz + status gate.** The autosave presign/commit path today assumes author + `draft`.
   Extend it so the **assigned approver** of the active stage may write working content while
   status is `under_review`. New capability for "write working content as review actor"; gate on
   stage eligibility (`approval_stage_instances.eligible_actor_ids`), following ADR 0022. Authz
   recording stays off any read-only / lock-holding tx (advisory-lock-deadlock constraint).
2. **Session/lock.** The assigned approver holds the edit session/lock during review (the
   editor's existing session mechanism, extended from `draft`-only to also cover the review actor
   under `under_review`). Single-writer is preserved **by status**: `draft` → author writes;
   `under_review` → assigned approver writes; the other party is locked out.
3. **ADR.** Extending the working-content write boundary to the review actor is an authz-boundary
   change → record an ADR (root-cause, not symptom-patch).

The **reject → draft** transition and all approval/decision logic are **existing and untouched**.
This design only *adds* the eigenpal mount and the reviewer-write gate; it changes nothing in the
approval state machine.

### §D — Save-on-decision guarantee

On **approve OR reject**, the cockpit flushes a final working-content save (editor serialize →
presign → commit) and awaits its completion **before** recording the sign-off, guaranteeing the
reviewer's suggestions + comments are persisted at decision time. Orchestrated on the frontend:
the decision handler awaits `editorRef` save+commit, then proceeds to the existing `signoff(...)`
call. No change to the signoff endpoint itself.

### §E — Round-trip resolution

Reject → `draft` (existing). The author reopens the document in `DocumentEditorPage` (editing
mode) and 1.9 renders the reviewer's redlines with native accept/reject UI (`acceptChangeById` /
`rejectChangeById`), comments visible. The author accepts, edits, or rejects-with-explanation,
then re-finalizes → new approval round. **No new author-side work** — 1.9 supplies the accept/
reject surface; the existing editing-mode mount gets it for free once the package is upgraded.

### §F — "Mudanças vs vX" tab

Removed. The redline now lives inline in the suggesting-mode editor, and no two-file version-
compare engine exists in 1.9 — we will not fabricate one. The existing backlog row in
`wiki/backlog/detalhe-signoff.md` is updated to reflect that the diff is now satisfied by in-
document tracked changes. **Optional future:** a tracked-changes *summary list* (count + jump
list) via `extractTrackedChanges` — deferred with an explicit trigger, not built now.

## 5. Components & boundaries

| Unit | Responsibility | Depends on |
|---|---|---|
| `MetalDocsEditor` wrapper | Anti-corruption layer over 1.9; mode map incl. `review`; controlled-comment adapter | `@eigenpal/docx-editor-react@1.9` |
| Cockpit "Documento" panel | Mount editor in review mode against the working revision docx | wrapper, signed-url fetch, comments query |
| Cockpit decision orchestration | Flush save → await commit → record signoff | `editorRef`, existing `signoff(...)` |
| Reviewer-write authz gate (backend) | Allow assigned approver to presign/commit working content under `under_review` | autosave handlers, authz (ADR 0022), stage eligibility |
| Author resolution (existing editor) | Render redlines + accept/reject natively after reject→draft | wrapper editing mode, 1.9 |

## 6. Error handling & edge cases

- **Save-on-decision fails** (presign/commit error): block the sign-off, surface the existing
  autosave error UX, let the reviewer retry. Never record a decision over unsaved review output.
- **Concurrency:** If-Match on `revision_version` as today; single-writer by status prevents
  author/reviewer simultaneous writes.
- **Approver not eligible** for the stage: authz denies the working-content write (403), editor
  mounts read-only (`viewing`) as a safe degrade.
- **Empty/blank working revision:** editor mounts on the working docx; if absent, honest empty
  state (no fabricated content) — consistent with current honest-degrade behavior.
- **1.9 comment-shape drift** vs 0.2: contained to the wrapper adapter; caught by wrapper tests.

## 7. Testing

- **Backend:** authz/status gate — assigned approver may presign/commit working content under
  `under_review`; non-eligible actor denied; author denied while `under_review`; reject→draft
  unchanged (regression). Follow the canonical testdb fixture framework (test-framework hard gate).
- **Wrapper:** `review → 'suggesting'` mapping; controlled-comment adapter against the 1.9 shape.
- **Cockpit:** editor mounts in review mode against the working revision (not PDF); decision
  flushes save before `signoff`; save-failure blocks the decision.
- **Round-trip (integration):** submit → reviewer suggests + comments → reject → author sees
  redlines → accept/reject → re-finalize.
- **Gates:** `go build ./...`, `go test ./...`, `npm run build:docx-v2` / `test:docx-v2` /
  `typecheck:docx-v2`, frontend `make test`, `tsc --noEmit`.

## 8. Non-goals (YAGNI)

- No real-time co-editing (single-writer last-writer-wins stays).
- No two-file version-compare engine (does not exist in 1.9; not faked).
- No migration of comments into OOXML (stay in the queryable store).
- No change to the approval state machine, reject path, freeze, or signoff endpoint.
- No new governed REV minted on review saves.

## 9. Rollout

Single combined change set. Order: §A (upgrade behind wrapper, all existing editor flows green)
→ §C backend reviewer-write gate + ADR → §B cockpit mount → §D save-on-decision → §F tab removal
→ tests + gates. Verify existing `DocumentEditorPage` editing/readonly flows still pass after §A
before touching the cockpit.

## 10. Open items for the implementation plan

- Exact capability name + grant set for the reviewer working-content write (ADR draft).
- Whether the reviewer session reuses the identical lock row/heartbeat as the author session or a
  status-scoped variant (prefer identical mechanism, status-gated eligibility).
- Confirm 1.9 `Comment` interface field parity with the wrapper adapter at install time.
