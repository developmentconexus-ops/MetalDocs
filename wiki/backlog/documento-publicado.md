# Backlog: Documento Publicado screen (`/documents/:id`)

> Last updated: 2026-06-23 (M4/F4.1: PDF download + Cobertura denominator + Páginas + Tamanho WIRED; Próxima revisão / Classificação / Documentos relacionados / Comentários are explicit defers-with-trigger; coverage numerator parked ADR-0042)

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

**Frontend (2026-05-29):** mock `PLACEHOLDER_RELATED` array removed. **DEFER (M4/F4.1, 2026-06-23):** section renders an honest "não disponível" empty state. **Trigger:** a related-documents relationship model + read endpoint. Re-wire grid once both land.

---

### CommentsCard — display-side architecture

Editor comments (`GET /api/v1/documents/:id/comments`, content is ProseMirror JSON `unknown[]`) exist but are scoped to the document editor. The published page needs a separate "discussion" model or a read-only ProseMirror renderer.

Decide:
1. Reuse editor comments with `extractPlainText` util for display
2. Separate `display_comments` table with plain-text content
3. Reply threading UX (parent_library_id is present but UX not designed)

**Frontend (2026-05-29):** mock `PLACEHOLDER_COMMENTS` array + reply-box shell removed. **DEFER (M4/F4.1, 2026-06-23):** section renders an honest "não disponível" empty state. **Trigger:** a decided display-comments architecture (reuse editor comments via `extractPlainText`, or a separate `display_comments` model). Re-wire once decided.

---

### ~~PDF download~~ — RESOLVED 2026-06-23 (M4/F4.1)

**Resolved.** The endpoint is `POST /api/v1/documents/:id/export/pdf` (operationId `exportDocumentPDF`,
returns `{ signed_url, size_bytes, cached, … }`, rate-limited 20/min) — it already shipped; the old
backlog line (`GET …/pdf`, "no endpoint") was stale. "Baixar PDF" now calls the existing `exportPDF`
client (`features/documents/api/exports.ts`), shows a pending state, opens the `signed_url`, and handles
429 — mirroring `ExportMenu.handlePDF`.

---

### Coverage KPI — RESOLVED-partial 2026-06-23 (M4/F4.1); numerator parked (ADR-0042)

**Denominator wired.** "Cobertura" KPI + side card now render the obligated-audience count from
`GET /documents/:id/distribution` (`DistributionSummaryResponse.total_targets`) via the M2
`useDistributionSummaryQuery` hook; EM_DASH on error (never a fabricated 0). **Trigger (remaining):** the
read **numerator** (% lido) is parked per **ADR-0042** — when a read-tracking API lands, replace the
"leitura em acompanhamento (ADR-0042)" label with the real percentage. Until then the card shows the
denominator + parked label only — **no fabricated %**.

---

### AuditCard / ISO seal — values_hash

`GET /api/v1/documents/:id` does not return `values_hash`. AuditCard + ISOSeal cut from this screen pending backend field addition.

---

### KPI: Próxima revisão — DEFER (no backend field)

**Trigger:** a review-due-date field on the document or controlled-document model. None exists today.
The KPI + "Próxima revisão" fact render an honest em-dash ("sem data de revisão definida"), not a
placeholder. Wire when the field lands.

### Classificação (confidentiality) — DEFER (no backend field)

**Trigger:** a confidentiality/classification field on `DocumentResponse`. None exists today — and the
classification taxonomy itself (public/internal/restricted?) is an unmade governance decision, not just
a missing column. The "Classificação" fact renders an honest em-dash. Wire when the field + taxonomy land.

### ~~KPI: Páginas / Tamanho~~ — RESOLVED 2026-06-23 (M4/F4.1)

**Resolved.** The old claim ("no page count or file size field in API response") was **stale**.
`DocumentResponse` already returns `current_revision_page_count` (`int ≥ 1 | null`) and
`current_revision_file_size_bytes` (`int64 ≥ 0 | null`) — verified at openapi.yaml:4832/4837 and generated
FE types `lib/api-types/index.d.ts:2540-2541`. Both are now wired on the published screen via the existing
`useDocumentDetailQuery` data, formatted by `formatPageCount` / `formatFileSize` (`lib/documentDetailMeta.ts`,
binary units + pt-BR, em-dash on null).

---

### `documentDetailMeta` formatter audit (drift hotfix 2026-05-10)

`DocumentPublishedPage.tsx:11` imported `formatPublishedAt` / `formatSignedAt` / `formatShortDate` from `lib/documentDetailMeta.ts`, but commit `bf2cd5fe` shipped only `resolveProfileLabel`, `resolveAreaLabel`, `SIGNOFF_STATUS_META`. Module-eval SyntaxError blanked `/documents/:id`. Hotfix added the 3 formatters as Intl pt-BR pure helpers (long / short-numeric / datetime, em-dash fallback on null/Invalid Date).

Followups:
- audit other `wiki/modules/documents.md` Key files anchors for similar drift (claimed but unshipped)
- add Vitest coverage for the 3 formatters (timezone-stable inputs, null/undefined/invalid)
- consider promoting Intl formatters to `lib/utils/formatDate.ts` if reused outside this feature
- dispatch `wiki-curator` to re-stamp `Last verified` and trim claims that don't match code

Cause: wiki documented planned exports as shipped without verification. Drift policy violated.
