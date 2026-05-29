# Backlog: Documento Publicado screen (`/documents/:id`)

> Last updated: 2026-05-29 (QA `qa/documents-detail`: area/profile snapshot fields resolved + mock related/comments replaced with honest em-breve)

---

## Phase 3c gaps — require backend changes

### ~~`DocumentResponse` missing area_code, profile_code, created_at~~ — RESOLVED 2026-05-29

**Resolved**. `GET /api/v1/documents/:id` already returns `ProcessAreaCodeSnapshot`, `ProfileCodeSnapshot`, and `CreatedAt` on `DocumentResponse` (verified against live payload for `c1bb2112-21ea-46fc-ac1f-719d04994d41`: `ProcessAreaCodeSnapshot: "rh"`, `ProfileCodeSnapshot: "po"`, `CreatedAt: "2026-05-28T20:01:07.859403-03:00"`). Wiki backlog claim was stale.

Frontend now wires both via `resolveAreaLabel` + `resolveProfileLabel` (existing `documentDetailMeta.ts`) → fills breadcrumb middle slot, DocCardMini header + type, hero `typeLabel` badge, and the Identificação e responsabilidade facts grid (`Tipo` + `Área`).

---

### `DocumentResponse` missing `created_by_display_name`

`created_by` returns the username (e.g. `admin-local`), not the user's display name. The owner banner should show the full display name (e.g. "Administrator").

Backend fix: add `created_by_display_name` (denormalized snapshot at publish time) to `GET /api/v1/documents/:id` response.

**Frontend interim:** If `created_by === currentUser.username`, substitute `currentUser.displayName`. Does not cover documents created by other users.

---

### `DocumentResponse` missing `controlled_document_id`

"Iniciar revisão" calls `POST /api/v2/controlled-documents/:cdId/revisions`. The page only has the document `id`, not `controlled_document_id`. Currently RBAC gate is role-only (admin/editor/qms_admin/area_admin). Full author-ownership gate needs this field.

Backend fix: expose `controlled_document_id` in `GET /api/v1/documents/:id` response.

**Frontend impact:** Replace role-only gate with `user.userId === doc.controlled_document_owner_id || hasRole(['admin', 'qms_admin'])`.

---

### "Iniciar revisão" mutation — `POST /api/v2/controlled-documents/:cdId/revisions`

Button is rendered but `aria-disabled`. Need to wire the mutation with:
- Confirmation dialog (destructive action)
- Loading/pending state on button
- Success: navigate to new revision draft
- Error: toast via `sonner`

Mutation fn to add to `features/documents/api/` or new `controlledDocumentsApi.ts`.

---

## Previously deferred

### VersionTimeline — revision list endpoint

No `GET /api/v1/documents/:id/revisions` list endpoint exists. Only single-revision URL (`/revisions/:rid/url`). 

Backend: implement list endpoint returning `[{ revision_id, version_num, created_by, created_at, summary? }]`.

**Frontend:** `DocumentVersionTimeline` already built and takes `VersionEntry[]`. Wire once endpoint exists. Remove `PLACEHOLDER_VERSIONS`.

---

### RelatedGrid — relationship model

No related-documents relationship model in the backend. Needs design + backend work before frontend.

**Frontend (2026-05-29):** mock `PLACEHOLDER_RELATED` array removed. Section now renders honest "em breve" empty state. Re-wire grid once endpoint + relationship model land.

---

### CommentsCard — display-side architecture

Editor comments (`GET /api/v1/documents/:id/comments`, content is ProseMirror JSON `unknown[]`) exist but are scoped to the document editor. The published page needs a separate "discussion" model or a read-only ProseMirror renderer.

Decide:
1. Reuse editor comments with `extractPlainText` util for display
2. Separate `display_comments` table with plain-text content
3. Reply threading UX (parent_library_id is present but UX not designed)

**Frontend (2026-05-29):** mock `PLACEHOLDER_COMMENTS` array + reply-box shell removed. Section now renders honest "em breve" empty state. Re-wire once display-side architecture is decided.

---

### PDF download — `GET /api/v1/documents/:id/pdf`

No PDF generation endpoint. "Baixar PDF" button is `aria-disabled` pending this.

---

### Coverage KPI — fanout read coverage API

Requires fanout/read-tracking API. No endpoint today. "Cobertura" KPI shows `—`.

---

### AuditCard / ISO seal — values_hash

`GET /api/v1/documents/:id` does not return `values_hash`. AuditCard + ISOSeal cut from this screen pending backend field addition.

---

### KPI: Próxima revisão

No review-due-date field in the document or controlled-document model.

---

### KPI: Páginas

No page count or file size field in API response.

---

### `documentDetailMeta` formatter audit (drift hotfix 2026-05-10)

`DocumentPublishedPage.tsx:11` imported `formatPublishedAt` / `formatSignedAt` / `formatShortDate` from `lib/documentDetailMeta.ts`, but commit `bf2cd5fe` shipped only `resolveProfileLabel`, `resolveAreaLabel`, `SIGNOFF_STATUS_META`. Module-eval SyntaxError blanked `/documents/:id`. Hotfix added the 3 formatters as Intl pt-BR pure helpers (long / short-numeric / datetime, em-dash fallback on null/Invalid Date).

Followups:
- audit other `wiki/modules/documents.md` Key files anchors for similar drift (claimed but unshipped)
- add Vitest coverage for the 3 formatters (timezone-stable inputs, null/undefined/invalid)
- consider promoting Intl formatters to `lib/utils/formatDate.ts` if reused outside this feature
- dispatch `wiki-curator` to re-stamp `Last verified` and trim claims that don't match code

Cause: wiki documented planned exports as shipped without verification. Drift policy violated.
