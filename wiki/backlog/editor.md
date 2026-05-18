# Editor — Deferred Backlog

> **Last verified:** 2026-05-18
> **Scope:** Historical backlog and audit memory for the `/documents/:documentID/edit` editor screen. The governed sidebar slice is now implemented with real runtime data; this page now tracks only the remaining editor defers and the audit history that led to the implementation.
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

## Governed Sidebar Sync (2026-05-18)

This supersedes the older "sidebar deferred" assumptions recorded later in this file.

- `Código`, `Perfil`, `Área`, and `Visibilidade` now render from real runtime data composed from document detail, taxonomy lookups, and controlled-document visibility from the registry contract.
- Governed history now comes from `GET /api/v1/documents/{id}/revision-history` and formats business revisions as `REV00`, `REV01`, and so on.
- The history source is `documents` lineage by `controlled_document_id`; `document_revisions` remains technical/autosave-only and must not appear as business history.
- The approval chain now comes from `GET /api/v1/documents/{id}/approval-instance` and is rendered only when the document is `under_review`.
- Remaining sidebar defers are limited to `Vigência atual` and `Próx. revisão`, because the editor route still lacks a truthful effective-version/effective-date contract for those rows.

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
- `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-module-contract-sync.ps1 -Module documents` -> PASS (wrapper truth aligned to `frontend/apps/web/src/features/documents/api/documents.ts` by `0e446607`)

Route truth evidence:

- OpenAPI comments paths: `api/openapi/v1/openapi.yaml:3723` and `api/openapi/v1/openapi.yaml:3741`
- Generated backend comments surface: `internal/modules/documents/api/api.gen.go` (`ListDocumentComments`, `CreateDocumentComment`, `UpdateDocumentComment`, `DeleteDocumentComment`)
- Runtime route wiring: `internal/modules/documents/delivery/http/handler.go:120-123` and `internal/modules/documents/delivery/http/handler.go:155-158`
- Frontend wrapper: `frontend/apps/web/src/features/documents/api/documents.ts`

| Item | Source | Runtime/API reality | Frontend reality | Classification | Action |
|---|---|---|---|---|---|
| Documents runtime + contract gates | `check-system-runnable`, `check-module-contract-sync` | Both required gates pass | N/A | implemented and aligned | Proceed with screen-level hardening scope |
| Editor mounts by status/session | `DocumentEditorPage` + `useDocumentSession` | Runtime supports writer/readonly/lost session modes | Editor mode toggles by `isEditable` (draft+writer only) | implemented and aligned | Keep |
| Comments list load | comments route + wrapper + hook | GET route exists and is wired end-to-end | `useDocumentComments` uses local `useEffect` state, not TanStack Query | implemented but legacy-wired | Normalize to TanStack Query hooks/query keys |
| Add comment | POST comments route | Route exists | Callback gated by `if (isEditable)` | screen-local integration fix | Split `canEditContent` vs `canComment` and verify behavior |
| Reply comment | POST comments route (`parent_library_id`) | Route exists | Callback gated by `if (isEditable)` | screen-local integration fix | Same permission split as add |
| Resolve/reopen comment | PATCH comments route (`done`) | Route exists; `resolved_at` semantics present | Callback gated by `if (isEditable)` | screen-local integration fix | Same permission split as add |
| Delete comment | DELETE comments route | Route exists | Callback gated by `if (isEditable)` | screen-local integration fix | Same permission split as add |
| Reviewer can comment in `under_review` | lifecycle design spec | API supports comments regardless of draft-only frontend guard | UI currently blocks comment mutations when not draft+writer | screen-local integration fix or Eigenpal adapter prerequisite | Implement permission split, then manual smoke; if readonly cannot comment at editor level, reclassify as Eigenpal prerequisite |
| Rejected->draft keeps comments visible | lifecycle design spec + `document_comments` persistence | Comments are document-scoped and persisted | Hook loads all comments for document | implemented and aligned | Keep; document lifecycle rule in notes/wiki |
| Released output remains clean (no active comments) | lifecycle design spec | Editor comments are separate from released PDF rendering pipeline | Published screen currently has placeholder "discussion comments" section | defer | Plan 12.6 published-screen cleanup must avoid surfacing active editor comments as released content |
| Unresolved comments block approval/release | lifecycle design spec | Final approval now fails server-side with `approval.unresolved_comments` while active comments remain unresolved | Signoff dialog maps the conflict inline; editor keeps comment-load failures visible with persistent retry UI | implemented and aligned | Keep rule server-owned; mirror the mapped conflict in approval UX only |
| Sidebar metadata rows (Perfil/Área/Vigência/Próx revisão/Visibilidade) | `EditorMetaSidebar` + backlog | No complete single response shape currently wired for all rows | Mock values marked TODO | missing backend capability | Keep deferred exactly as backlog row 1 |
| Sidebar revisions timeline | `EditorMetaSidebar` + backlog | No `/documents/:id/revisions` list wired for editor sidebar | Mock timeline | missing backend capability | Keep deferred exactly as backlog row 2 |
| Sidebar next approvers | `EditorMetaSidebar` + backlog | No dedicated editor-side signoffs payload wired for this section | Mock approvers list | missing backend capability | Keep deferred exactly as backlog row 3 |
| Orphan comment partitioning | `useDocumentComments` + contract doc | Backend stores metadata; anchor ownership is DOCX-side | Hook currently returns empty `orphans` | implemented but legacy-wired | Keep as follow-up; do not block current lifecycle hardening |
| Template editor comments parity | lifecycle design + template backlog | No template comments persistence/API contract in current truth | No real template comments UI | missing backend capability | Prerequisite for template comments; never fake UI |

### Ready for implementation

- Document comments query normalization to TanStack Query.
- Document editor permission split: content editability vs commentability (draft/review/rejected).
- Notes/worksheet alignment with active-review-feedback lifecycle.

### Prerequisites

- Template comments database/backend/OpenAPI/generated/frontend surfaces before template comments UI parity.
- Potential Eigenpal adapter prerequisite if readonly mode cannot create review comments.

### Deferred

- Sidebar metadata/revisions/approvers sections that still depend on missing backend capability.
- Published-view discussion/comments surfaces until real product and backend behavior is defined for released views.
- Generic cross-module comments platform.

## approval-blocking-unresolved-comments

Resolved 2026-05-18. Final approval now checks unresolved active comments server-side and returns 409 `approval.unresolved_comments`. The approval dialog resolves that code inline, and the document editor no longer hides comment-query failures behind a toast-only state.

Remaining follow-up:

- Apply the same lifecycle rule to templates only after template comments exist.

### Verification needed next

- `cd frontend/apps/web; pnpm.cmd tsc --noEmit -p tsconfig.build.json`
- `cd frontend/apps/web; pnpm test -- src/features/documents/hooks/editor/__tests__/useDocumentComments.load.test.tsx src/features/documents/hooks/editor/__tests__/useDocumentComments.add.test.tsx src/features/documents/hooks/editor/__tests__/useDocumentComments.orphan.test.tsx src/features/documents/pages/DocumentEditorPage.test.tsx`
- Manual/Navegador smoke on `/documents/:documentID/edit` and template route for parity/defer checks.

## Integration Audit (2026-05-18)

Scope: editor screen re-audit after autosave/runtime fixes and approval unresolved-comments hardening, with emphasis on API contract truth, TanStack Query usage, frontend feature placement, and persistence-backed behaviors.

Evidence used:

- Screen route/page: `frontend/apps/web/src/features/documents/pages/DocumentEditorRoutePage.tsx`, `frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx`
- Editor hooks/wrappers: `useDocumentAutosave.ts`, `useDocumentComments.ts`, `useDocumentSession.ts`, `api/documents.ts`, `queries/useDocumentCommentsQuery.ts`
- Query key ownership: `frontend/apps/web/src/lib/queryKeys.ts`
- Runtime/handler truth: `internal/modules/documents/delivery/http/handler.go`, `internal/modules/documents/application/service.go`, `internal/modules/documents/repository/repository.go`, `internal/modules/documents/approval/application/decision_service.go`
- Contract truth: `api/openapi/v1/openapi.yaml`

| Item | Source | Runtime/API reality | Frontend reality | Classification | Action |
|---|---|---|---|---|---|
| Editor route + screen ownership | route page + frontend structure | `/documents/:documentID/edit` is owned by `features/documents` and mounted through `documentsRoutes` | `DocumentEditorRoutePage` delegates cleanly to `DocumentEditorPage` | implemented and aligned | Keep |
| Toolbar identity block (code, title, revision, status) | `GET /api/v1/documents/{id}` + `DocumentEditorPage` | Runtime handler, OpenAPI, generated backend surface, generated frontend types, and screen now agree on a typed detail payload; `FormDataJSON` is emitted as embedded JSON rather than Go `[]byte` base64 | `DocumentEditorPage` loads and refetches detail through `useDocumentDetailQuery` using the generated `DocumentDetailResponse` type | implemented and aligned | Keep |
| DOCX artifact load for current revision | signed revision URL route + page blob fetch | Runtime provides `/revisions/{rid}/url` and the editor can fetch the artifact bytes | Screen handles signed-url fetch + blob load correctly and shows inline error if artifact fetch fails | implemented and aligned | Keep |
| Writer/readonly session lease | session acquire/heartbeat/release handlers + `useDocumentSession` | Runtime supports single-writer lease with readonly fallback and best-effort release semantics | Dedicated hook models lease lifecycle directly; this is screen state, not a normal TanStack query | implemented and aligned | Keep custom hook boundary |
| Autosave debounce + commit persistence | autosave presign/commit handlers + repository/session invariants | Runtime path is now working: presign + upload + commit create new revision lineage and advance the base ack | `useDocumentAutosave` has crash-recovery, debounce, explicit flush gating, and local status state | implemented and aligned | Keep |
| Submit for review CTA (`Submeter para revisão`) | OpenAPI finalize contract + runtime `finalizeDocument` handler | Contract and runtime now align on required `Idempotency-Key` plus `201 { instanceId }` response semantics | `finalizeDocument()` now sends `idempotencyKey` and consumes the response body correctly, though it still lives in the legacy handwritten wrapper slice | implemented and aligned | Keep behavior; normalize this wrapper to generated frontend types only when the broader documents API slice is migrated |
| Comments list + CRUD | comments OpenAPI/runtime/generated surface + comments query hook | Runtime handler, OpenAPI, generated backend surface, generated frontend types, and query wrapper now agree on named comments schemas for list/create/update | Screen uses TanStack Query + `QK.documents.comments(id)` backed by generated comments types; the only remaining adaptation is the explicit editor-package content bridge between Eigenpal `Comment['content']` and open JSON payloads | implemented and aligned | Keep |
| Comment load failure visibility | `useDocumentComments` + `DocumentEditorPage` | Query failure is a real runtime possibility with persisted review-state implications | Screen now keeps the failure visible with inline retry instead of toast-only UX | implemented and aligned | Keep |
| Approval blocked by unresolved comments | approval decision service + shared error mapping | Final approval now fails server-side with `approval.unresolved_comments` while active comments remain unresolved | Screen-side UX aligns: editor shows persistent comment-load failure, approval dialog maps the business conflict inline | implemented and aligned | Keep server-owned rule |
| Non-draft PDF polling state | editor screen boundary vs published/view boundary | PDF availability belongs to published/view flows, not the editor | Unused editor-side polling was removed; the editor no longer owns background PDF readiness checks | implemented and aligned | Keep PDF status/download concerns out of `/documents/:documentID/edit`; wire them only on the published/view screen |
| Sidebar metadata rows (`Código`, `Perfil`, `Área`, `Visibilidade`) | governed sidebar implementation | Runtime truth now comes from document detail, taxonomy, and controlled-document visibility | Sidebar is runtime-backed | implemented and aligned | Keep |
| Sidebar governed revisions timeline | governed sidebar implementation | Runtime truth now comes from `GET /api/v1/documents/{id}/revision-history` | Sidebar is runtime-backed and excludes `document_revisions` autosave rows | implemented and aligned | Keep |
| Sidebar approval chain | governed sidebar implementation | Runtime truth now comes from `GET /api/v1/documents/{id}/approval-instance` | Sidebar renders the full chain only in `under_review` | implemented and aligned | Keep |
| Released output remains clean | lifecycle design + documents/approval persistence | Runtime approval/comments model keeps active comments out of clean released output | Published-view cleanup is still a separate product slice, not this editor screen | defer | Keep deferred until published-screen work |
| Template editor comments parity | lifecycle design + template backend truth | No template-owned comments capability exists yet | No real parity path can be claimed from the editor screen | missing backend capability | Keep deferred until template comments exist |

### Ready for implementation

- Comments wrapper normalization to generated frontend types.
### Prerequisites

- Template comments database/backend/OpenAPI/generated/frontend surfaces before parity claims.

### Deferred

- `Vigência atual` and `Próx. revisão` rows until the editor has a truthful effective-version contract.
- Published-view discussion/comment surfaces until the released-view product boundary is defined.
- Generic cross-module comments platform.

### Verification needed next

- `cd frontend/apps/web; pnpm.cmd tsc --noEmit -p tsconfig.build.json`
- `cd frontend/apps/web; pnpm.cmd vitest run src/features/documents/pages/DocumentEditorPage.test.tsx src/features/documents/hooks/editor/__tests__/useDocumentComments.load.test.tsx src/features/approval/components/SignoffDialog.test.tsx`
- Finalize prerequisite repair verified on 2026-05-18: `api/openapi/v1/openapi.yaml`, generated frontend types, `frontend/apps/web/src/features/documents/api/documents.ts`, and `internal/modules/documents/delivery/http/handler.go` now agree on required `Idempotency-Key` and `201 { instanceId }`

## Integration Audit (2026-05-18 Sidebar)

Scope: right `EditorMetaSidebar` on `/documents/:documentID/edit`, with emphasis on replacing mock rows using real `documents`, `registry`, `taxonomy`, and `approval` truth without leaking draft-incorrect workflow semantics into the editor.

Evidence used:

- Current sidebar implementation: `frontend/apps/web/src/features/documents/components/EditorMetaSidebar.tsx`
- Editor screen wiring: `frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx`
- Documents detail contract/runtime: `frontend/apps/web/src/features/documents/api/documents.ts`, `frontend/apps/web/src/lib/api-types/index.d.ts` (`DocumentDetailResponse`), `wiki/modules/documents.md`
- Taxonomy lookup surfaces: `frontend/apps/web/src/features/taxonomy/api/taxonomy.ts`, `frontend/apps/web/src/features/taxonomy/queries/useProfilesQuery.ts`, `frontend/apps/web/src/features/documents/queries/useAreasQuery.ts`
- Registry/controlled-document truth: `frontend/apps/web/src/features/registry/api/controlledDocuments.ts`, `frontend/apps/web/src/features/registry/types.ts`, `frontend/apps/web/src/lib/api-types/index.d.ts` (`ControlledDocumentVisibility`, controlled-document schema), `wiki/modules/registry.md`
- Approval-instance runtime + contract truth: `internal/modules/documents/approval/http/doc_approval_handler.go`, `internal/modules/documents/approval/http/get_instance_handler.go`, `internal/modules/documents/approval/http/contracts/instance_read.go`, `frontend/apps/web/src/features/documents/api/documents.ts`, `frontend/apps/web/src/lib/api-types/index.d.ts`

| Item | Source | Runtime/API reality | Frontend reality | Classification | Action |
|---|---|---|---|---|---|
| Sidebar shell (collapse toggle, persisted open state, right-rail container) | `EditorMetaSidebar`, `DocumentEditorPage` | Pure screen-local UI; no backend dependency | Already wired with `editor-sidebar-open` localStorage and toggle button | implemented and aligned | Keep |
| Section heading `Metadados` | design + component | Static shell label only | Already rendered correctly | implemented and aligned | Keep |
| `Código` row | `DocumentDetailResponse.Code` | `GET /api/v1/documents/{id}` already returns `Code` | Sidebar already receives `code` prop from `DocumentEditorPage` | implemented and aligned | Keep |
| `Perfil` display | `DocumentDetailResponse.ProfileCodeSnapshot` + taxonomy profiles list | Runtime already returns the profile code snapshot; taxonomy API can resolve code -> human name | Sidebar is still mock-only | screen-local integration fix | Resolve `ProfileCodeSnapshot` through `useProfilesQuery()` and render the real name with code fallback; do not invent a backend extension first |
| `Área` display | `DocumentDetailResponse.ProcessAreaCodeSnapshot` + taxonomy areas list | Runtime already returns the process-area code snapshot; taxonomy API can resolve code -> human name | Sidebar is still mock-only | screen-local integration fix | Resolve `ProcessAreaCodeSnapshot` through `useAreasQuery()` and render the real name with code fallback |
| `Visibilidade` row | `DocumentDetailResponse.ControlledDocumentID` + `GET /api/v1/controlled-documents/{id}` | Registry contract already has `visibility`; controlled-document identity is already present on document detail | Registry frontend wrapper is legacy-manual and omits `visibility` from `ControlledDocument`; sidebar is mock-only | implemented but legacy-wired | Normalize registry wrapper/types to generated contract, then show real visibility from the controlled document instead of adding a new documents field |
| `Vigência atual` row in draft editor | design + documents/registry module semantics | Controlled Document owns identity; document revisions own approval/publish lifecycle. For a draft editor, the draft under edit is not necessarily the currently effective published version. Current editor/detail contracts do not expose a truthful "effective version + effective date" pair for this row | Mock row incorrectly suggests the draft itself is already the current effective version | defer | Remove or hide this row in draft-focused first implementation; only reintroduce after product semantics for "current effective version from inside draft editor" are explicitly defined |
| `Próx. revisão` row | profile governance + effective-version semantics | Taxonomy profile exposes `reviewIntervalDays`, but computing the next review date also needs the effective approval/publish anchor for the currently valid revision. That anchor is not available on the editor detail contract | Mock row invents a due date | missing backend capability | Do not derive from draft timestamps. Add only after a truthful effective-date field/source exists |
| `Revisões` timeline | sidebar design + module wiki | No runtime list endpoint currently exposes revision lineage for this editor sidebar. `signedRevisionURL` exists for a specific revision file, but not the historical list needed for `TimelineRail` | Sidebar uses hardcoded revisions | missing backend capability | Keep deferred until a real revisions-list route or equivalent runtime surface exists |
| `Próximos aprovadores` section semantics | approval lifecycle + user requirement | Active approval instance only exists while a document is in review. Draft documents should not show approvers as if a route were already active | Current mock section shows approvers even in draft | screen-local integration fix | Render this section only when the document is `under_review` and an active approval instance exists; hide it in `draft`, `rejected`, `approved`, `published`, `obsolete`, etc. |
| `Próximos aprovadores` data source | `GET /api/v1/documents/{id}/approval-instance` runtime | Runtime endpoint exists and returns stage/signoff data with human display-name fallback in the handler mapper | Frontend wrapper is manual, the generated contract has no typed 200 body, and the handwritten status unions are wrong for runtime values (`in_progress`, `active`, `passed`, `failed`, `cancelled`) | shared contract prerequisite | Repair the approval-instance OpenAPI response schema and generated frontend types before using this endpoint as a professional sidebar contract |
| Approver badges (`próximo` / `aguarda`) | design + approval runtime | Once the approval-instance contract is fixed, the first active stage is enough to derive "próximo" vs later pending stages "aguarda" | Current mock badges are not connected to real stage status | screen-local integration fix | After contract repair, map active stage -> `próximo`, later pending stages -> `aguarda`, and ignore completed/rejected stages for this sidebar slice |

### Ready for implementation

- Sidebar shell stays as-is.
- Real profile and area labels can be wired now from document snapshot codes + taxonomy queries.
- Real visibility can be wired from the controlled document after local registry wrapper/type cleanup.
- Draft-incorrect workflow sections should be hidden instead of mocked.

### Prerequisites

- Approval-instance contract repair: runtime already responds, but OpenAPI/generated frontend types do not yet model the payload, and the current handwritten wrapper has status drift.

### Deferred

- `Vigência atual` until product semantics distinguish "draft being edited" from "currently effective published revision" inside this screen.
- `Próx. revisão` until an effective-date source exists.
- `Revisões` timeline until a real revisions-list capability exists.

### Verification needed next

- Browser E2E on `/documents/:documentID/edit` covering draft sidebar truth, finalize `revisionTitle`, under-review approval chain, and governed history ordering.
- Runtime/API spot-check that `revision-history` is sourced from governed `documents` lineage and not from technical `document_revisions`.
