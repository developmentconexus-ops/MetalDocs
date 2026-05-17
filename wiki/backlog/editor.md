# Editor — Deferred Backlog

> **Last verified:** 2026-05-06
> **Scope:** Deferred implementation items for the `/documents/:documentID/edit` editor screen. The right `EditorMetaSidebar` ships with mock data while backend endpoints are designed.
> **Out of scope:** Bug fixes (see `bugs/`), shared editor primitive (`EditorChrome`).
> **Key files:**
> - `frontend/apps/web/src/features/documents/components/EditorMetaSidebar.tsx:12` — TODO block for `MOCK_META`
> - `frontend/apps/web/src/features/documents/components/EditorMetaSidebar.tsx:23` — TODO block for `MOCK_REVISIONS`
> - `frontend/apps/web/src/features/documents/components/EditorMetaSidebar.tsx:34` — TODO block for `MOCK_APPROVERS`
> - `frontend/apps/web/design-source/editor/IMPLEMENTATION.md` — Phase 3c open item

---

## Status

| Item | Priority | Depends on | Status |
|---|---|---|---|
| Metadados rows (Perfil, Área, Próx. revisão, Visibilidade) | High | Extend `GET /api/v1/documents/:id` response | Deferred |
| Vigência atual (approval date) | High | Extend `GET /api/v1/documents/:id` response | Deferred |
| Revisões timeline (full history list) | High | New endpoint `GET /api/v1/documents/:id/revisions` | Deferred |
| Próximos aprovadores (signoff list) | Medium | New endpoint `GET /api/v1/documents/:id/signoffs` | Deferred |

---

## Item 1 — Metadados rows

**File:** `frontend/apps/web/src/features/documents/components/EditorMetaSidebar.tsx:15-21` (`MOCK_META`)

**Current state:** Hardcoded array with 5 mock rows: Perfil "Procedimento Op.", Área "RH", Vigência atual "v4 · 15/04/2026", Próx. revisão "15/04/2027", Visibilidade "Toda empresa".

**What it should do:** Show real metadata for the document being edited. Drives compliance visibility (who can see it, when next review is due, which profile/area governs it).

**Backend work needed:**
- Extend `GET /api/v1/documents/:id` response payload with:
  - `ProfileName` — denormalized profile display name (already have `ProfileCode`)
  - `AreaName` — denormalized area name (already have `AreaCode` via profile join)
  - `ApprovedAt` — timestamp the current revision was approved (for "Vigência atual" row)
  - `NextReviewAt` — `ApprovedAt + Profile.ReviewIntervalDays` (computed; could be derived client-side if both fields land)
  - `Visibility` — visibility scope enum: `'company' | 'area' | 'restricted'` (depends on doc-profile governance — needs design)
- Profile/Area names are joins in `documents` table — should be fast.
- `Visibility` requires governance schema work; may land later than the rest.

**Frontend work:**
1. Update `lib/api-types/` via OpenAPI codegen after backend extends response.
2. Replace `MOCK_META` with derived rows from `doc` prop (already passed via `DocumentEditorPage`).
3. Format dates with locale-aware helper (probably extend existing util).
4. Remove TODO comment block.

---

## Item 2 — Revisões timeline

**File:** `frontend/apps/web/src/features/documents/components/EditorMetaSidebar.tsx:26-32` (`MOCK_REVISIONS`)

**Current state:** Hardcoded 5-item array (v5..v1) rendered via `TimelineRail`. The "v5 · agora" entry is marked `active`.

**What it should do:** Show full revision history of the document with subtitles (label or commit-style summary) and timestamps. Active entry = `CurrentRevisionID`.

**Backend work needed:**
- New endpoint: `GET /api/v1/documents/:id/revisions`
- Returns: `[{ ID, VersionNum, Label?, CreatedAt, CreatedBy, IsCurrent }]` ordered by `VersionNum desc`.
- `Label` = optional rev-message (could come from `Checkpoint.Label` if revisions tie 1:1 to checkpoints, or a separate `revision_label` column).
- Pagination not needed — typical doc has <50 revisions.

**Frontend work:**
1. Codegen types after endpoint ships.
2. Add TanStack Query hook `useDocumentRevisions(documentID)` in `features/documents/queries/`.
3. Map response → `TimelineRail` items: `{ id, title: 'v${VersionNum}', subtitle: Label ?? CreatedBy, aside: formatDate(CreatedAt), active: IsCurrent }`.
4. Loading / error states (skeleton on the timeline, error toast).
5. Remove TODO comment block.

**Open question:** Should clicking a non-active timeline row navigate to a read-only view of that revision? Out-of-scope for first wire-up; defer until checkpoints UX is rethought.

---

## Item 3 — Próximos aprovadores

**File:** `frontend/apps/web/src/features/documents/components/EditorMetaSidebar.tsx:37-40` (`MOCK_APPROVERS`)

**Current state:** Hardcoded 2-item array. First marked `'next'` (highlighted badge), second `'wait'`.

**What it should do:** Show the active signoff sequence for the current revision when status is `under_review`. Empty when doc is in `draft` or `published`.

**Backend work needed:**
- Endpoint exists for approval module but not surfaced via doc API:
  - Likely already returnable: signoff records with `ActorID`, `Role`, `Status`, sequence order.
- Either:
  - (a) Add `Signoffs` array to `GET /api/v1/documents/:id` response (preferred — one fewer round trip), or
  - (b) Standalone `GET /api/v1/documents/:id/signoffs` returning the sequence.
- Need actor display names — backend should denormalize `ActorName`.

**Frontend work:**
1. Codegen types.
2. Add hook (or read from extended doc response).
3. Map signoffs → approver rows: status `next` for first pending, `wait` for subsequent pending. Hide approved/rejected entries (or render with different badge — design TBD).
4. Hide entire section when no active signoffs (e.g. doc in `draft`).
5. Remove TODO comment block.

**Open question:** Show completed approvers too (for context)? Defer until approval module redesign.

---

## Cross-cutting notes

- All three items share a single TanStack Query refetch invalidation key per document. When backend ships any of these, plumb the query key consistently with existing `useDocumentSession` patterns.
- Sidebar visibility persists via `localStorage` key `editor-sidebar-open`. No backend dependency.
- The `templatePlugin` mode gating in `MetalDocsEditor` (skipped for `document-edit`/`readonly`) is intentional — keeps eigenpal sidebar empty so the canvas centers when there are no comments. If a future feature needs template-style annotations in the document editor, lift the gate back and use CSS to hide the chip class instead.

## Integration Audit (2026-05-17)

Scope: Documents Editor review comments, revision sidebar, clean release semantics, and Plan 12 baseline for Template Editor parity.

Required gates:

- `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-system-runnable.ps1 -TargetRoute /api/v1/documents` -> PASS
- `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-module-contract-sync.ps1 -Module documents` -> FAIL (`[MISSING] feature API wrapper presence - frontend/apps/web/src/features/documents/api/documentsV2.ts`)

Gate decision:

- Classification: `runtime-contract-prereq` (script result: `shared contract prerequisite`)
- Stop status: implementation blocked until contract gate is repaired and rerun as PASS

Route truth evidence snapshot:

- OpenAPI comments paths: `api/openapi/v1/openapi.yaml:3723` and `api/openapi/v1/openapi.yaml:3741`
- Generated backend surface: `internal/modules/documents/api/api.gen.go`
- Runtime route wiring: `internal/modules/documents/delivery/http/handler.go:120` and `internal/modules/documents/delivery/http/handler.go:155`
- Active frontend wrapper currently in use: `frontend/apps/web/src/features/documents/api/documents.ts`

| Item | Source | Runtime/API reality | Frontend reality | Classification | Action |
|---|---|---|---|---|---|
| Documents module contract gate | `scripts/check-module-contract-sync.ps1 -Module documents` | Gate expects wrapper `frontend/apps/web/src/features/documents/api/documentsV2.ts` | Screen uses `frontend/apps/web/src/features/documents/api/documents.ts` | runtime-contract-prereq | Stop screen implementation and repair/align module contract gate expectations first |
| Inline document comments load | `DocumentEditorPage`, `useDocumentComments`, OpenAPI comments paths | GET comments route exists in runtime + OpenAPI + generated backend | Wrapper exists in `documents.ts` | blocked by prerequisite gate | Re-audit after gate repair |
| Add/reply/resolve/delete comments | document comments API + Eigenpal callbacks | POST/PATCH/DELETE comment routes exist | Page-level behavior still needs full lifecycle audit | blocked by prerequisite gate | Re-audit after gate repair |
| Reviewer adds comments during review | approved lifecycle spec | Not verified yet under current gate failure | Not verified yet under current gate failure | blocked by prerequisite gate | Re-audit after gate repair; if readonly commenting unsupported, classify `Eigenpal adapter prerequisite` |
| Unresolved comments block approval | approved lifecycle spec | No server-side enforcement verified in this run | No enforcement verified in this run | backend/API prerequisite (pending) | Keep as prerequisite candidate; confirm after gate repair |
| Template parity baseline | lifecycle spec + template backlog | Template comments capability still unknown in this run | No fake comments UI allowed | defer/prerequisite (pending) | Keep deferred until backend/database/API capability is proven |

### Prerequisite to unblock Task 1

- Fix the documents module contract-sync failure (wrapper truth drift between expected `documentsV2.ts` and actual `documents.ts`), then rerun:
  - `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-system-runnable.ps1 -TargetRoute /api/v1/documents`
  - `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-module-contract-sync.ps1 -Module documents`

### Next step after prerequisite passes

- Resume this audit and classify every visible comment/revision/sidebar/action item to completion before any code implementation.
