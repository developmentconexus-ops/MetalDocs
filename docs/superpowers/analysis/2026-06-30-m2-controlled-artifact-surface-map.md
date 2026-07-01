# M2 surface map — controlled-artifact shared view layer

**Date:** 2026-06-30
**Purpose:** Ground-truth inventory of the document-side components to be generalized + the template-side data sources a template adapter must wrap, with a normalized-view-model gap analysis. Feeds M2 tasks T7–T14 of the template↔document parity plan.
**Method:** read-only investigation of current source (branch `feat/template-document-parity`).

---

## Component inventory (document side — sources to generalize)

| Component | path | Props | Data sources |
|---|---|---|---|
| `DocumentDetailLayout` (tabbed shell) | `frontend/apps/web/src/features/documents/pages/DocumentDetailLayout.tsx` | none (zero-prop) | router only; tabs via `NavLink` ("Documento" index, "Distribuição" child). Child pages read `useParams<{documentId}>`. |
| `DocumentPublishedPage` (detail view) | `frontend/apps/web/src/features/documents/pages/DocumentPublishedPage.tsx` | reads `useParams` | `useDocumentDetailQuery`, `useDocumentRevisionHistoryQuery`, `useApprovalInstanceQuery`, `useDistributionSummaryQuery`, `useAreasQuery`, `useProfilesQuery` |
| `DocumentHero` | `frontend/apps/web/src/features/documents/components/DocumentHero.tsx` | `{breadcrumbItems[], docCard, badges, title, subtitle, actions}` — all ReactNode slots (pure layout) | none (slots) |
| `EditorMetaSidebar` | `frontend/apps/web/src/features/documents/components/EditorMetaSidebar.tsx` | `{open,onToggle,loading?,code?,profileLabel?,areaLabel?,visibilityLabel?,fileSizeBytes?,pageCount?,history?,approvalChain?,documentStatus?}` (dumb) | parent (`DocumentEditorPage`) maps queries → props. "Próximos aprovadores" section renders only when `documentStatus==='under_review'` && approvalChain present. |
| `SignoffDetailPage` (approval route) | `frontend/apps/web/src/features/approval/pages/SignoffDetailPage.tsx` | reads `useParams`; local `tab` state | `useActiveDocumentContextQuery` → `GET /controlled-documents/{id}/active-document` |
| `ControlledDocumentDetailPanel` (decision panel) | `frontend/apps/web/src/features/approval/components/ControlledDocumentDetailPanel.tsx` | `{documentId,approvalState,contentHash,revisionVersion,lockedByInstanceId?,lockedByActor?,lockAcquiredAt?,effectiveFrom?,effectiveTo?,publishedDocumentId?,autoOpenSignoff?,initialSignoffDecision?,beforeDecision?}` | `getInstance(documentId)` (direct async, not hook) → `GET /documents/{id}/approval-instance`; ETag seeded into `etagCache` for If-Match writes |
| `ReviewDocumentCanvas` (read-only canvas) | `frontend/apps/web/src/features/approval/components/ReviewDocumentCanvas.tsx` | `{documentId,currentRevisionId,status,approverDisplay}` | renders blob read-only A4 |

**Document decision actions** (driven by `TRANSITION_POLICY[status].actions` in ControlledDocumentDetailPanel):
- Submeter → `submit` `POST /documents/{id}/submit`
- Assinar → `SignoffDialog` → `signoff` `POST /documents/{id}/signoff` (ETag If-Match)
- Cancelar instância → `cancel` `POST /documents/{id}/cancel`
- Publicar/Agendar → `publish` `POST /documents/{id}/publish` or `schedule-publish`

## Document data layer

| Hook | API fn | Endpoint | Key return fields |
|---|---|---|---|
| `useDocumentDetailQuery` | `getDocument` | `GET /documents/{id}` | id, name, status, code, revision_number, revision_version, current_revision_id, created_by, controlled_document_id, profile_code_snapshot, process_area_code_snapshot, revision_title, current_revision_file_size_bytes, current_revision_page_count, template_version_id, created_at |
| `useApprovalInstanceQuery` | `getApprovalInstance` | `GET /documents/{id}/approval-instance` | id, status, submitted_by/at, completed_at, stages[]{id,stage_index,label,status,signoffs[]{actor_user_id,decision,signed_at}}, etag |
| `useDocumentRevisionHistoryQuery` | `getDocumentRevisionHistory` | `GET /documents/{id}/revision-history` | items[]{document_id,revision_number,revision_title,status,created_at,is_current} |
| `useDistributionSummaryQuery` | `getDistributionSummary` | `GET /documents/{id}/distribution` | total_targets |
| `useActiveDocumentContextQuery` | `getActiveDocumentContext` | `GET /controlled-documents/{id}/active-document` | document_id, approval_state, content_hash, revision_version, published_document_id, approval_instance_id |

## Template data layer (what a template adapter has to work with)

- `getTemplate(id)` `GET /templates/{id}` → `{template: TemplateDTO, latest_version: VersionDTO}` (in `body.data`). File `features/templates/api/templates.ts:119`.
- `getVersion(templateId, n)` — single version by number. **No list-versions query/endpoint. No template audit/history query.**
- Lifecycle: `submitForReview` `POST .../versions/{n}/submit`; `reviewVersion` `POST .../review`; `approveVersion` `POST .../approve`. All return `VersionDTO` from `body.data.version`.
- **No `createNextVersion` client fn exists** — T13 must add it → `POST /api/v1/templates/{id}/versions` (backend `CreateNextVersion` in `create.go`).
- **`TemplateDTO`**: id, tenant_id, doc_type_code|null, key, name, description?, latest_version, latest_revision_number, published_version_id|null, published_version_number?, current_revision_number?, created_by, created_at, archived_at|null. **No code, no area, no profile, no fileSize/pageCount.**
- **`VersionDTO`**: id, template_id, version_number, revision_number?, status('draft'|'under_review'|'approved'|'published'|'obsolete'), docx_storage_key|null, content_hash|null, metadata_schema, placeholder_schema, author_id, pending_reviewer_role|null, pending_approver_role|null, reviewer_id, reviewed_at, approver_id, approved_at, submitted_at, published_at, obsoleted_at, lock_version?, created_at.

## Shared primitives (reuse, don't reinvent)

- `StatusPill` `components/ui/StatusPill.tsx` — `{status, className?}`; handles under_review (label "Em Revisão", `pill pill-review`).
- `CodeChip` `components/ui/CodeChip.tsx` — `{children, className?}`.
- `formatRevisionCode` `lib/labels/revisionCode.ts` → `REV{nn}` (zero-safe); re-exported via `features/documents/lib/documentDetailMeta.ts`.
- `VersionBadge` `features/shared/components/editor-chrome`.
- `TabBar` `components/ui/TabBar.tsx` — `{tabs[]{key,label,count?},activeKey,onTabChange,ariaLabel?}` roving-tab a11y, **state-driven**. **Neither shell uses it today** (DocumentDetailLayout=NavLink routing; SignoffDetailPage=local state). Candidate to standardize the shared shell on.

## NORMALIZED VIEW-MODEL — gaps (drives T7 design)

The view-model fields must be **optional/nullable**; the shared shell must **conditionally render** sections so templates render a reduced surface gracefully.

| Field | Document source | Template source | Gap |
|---|---|---|---|
| `kind` | adapter-injected `"document"` | adapter-injected `"template"` | no DTO field (both) |
| `id` | `id` | `TemplateDTO.id` | — |
| `code` | `code` | **none** (`key` is a slug, not a regulated code) | **GAP (template)** |
| `title` | `name` | `TemplateDTO.name` | — |
| `status` | `status` | `VersionDTO.status` (version-level; adapter picks which version) | semantic: template status is per-version |
| `versionNumber` | `revision_version` (OCC counter — not display) | `VersionDTO.version_number` (sequential) | naming/semantic mismatch |
| `revisionLabel` | `formatRevisionCode(revision_number)` | `formatRevisionCode(VersionDTO.revision_number)` | same formatter ✓ |
| `hero.breadcrumb` | area + code | none (no area) | **GAP (template)** |
| `hero.badges` | CodeChip + vigente + profile type | none | **GAP (template)** |
| `hero.subtitle` | status-derived string | derive from VersionDTO.published_at/approved_at (no formatter yet) | partial |
| `meta.profile/area` | snapshots via taxonomy queries | `doc_type_code` loosely ~ profile; **no area** | **GAP (template area)** |
| `meta.visibilityLabel` | not on DocumentDetailResponse (editor-session prop only) | n/a | partial (doc too) |
| `meta.fileSize` | `current_revision_file_size_bytes` | none | **GAP (template)** |
| `meta.pageCount` | `current_revision_page_count` | none | **GAP (template)** |
| `meta.dates` | created_at + completed_at (vigente desde); next-review hardcoded "—" | VersionDTO created_at/published_at/approved_at | doc next-review is known backlog gap |
| `approvalChain[]` | `ApprovalInstance.stages[].signoffs[]` | **none** — only `pending_reviewer_role`/`pending_approver_role` on VersionDTO | **GAP (template): no instance/signoff model** |
| `lineage[]` | `revision-history` endpoint | **no list query** (latest_version gives max; sequential fetch only) | **GAP (template)** |
| `tabs[]` | "Documento"+"Distribuição" (router) | no template shell exists | **GAP**; standardize on `TabBar`? |
| `actions` | `TRANSITION_POLICY[status]` (submit/signoff/cancel/publish, ETag If-Match) | map `VersionDTO.status` via `canSubmit/canReview/canApprove/canPublish`; `VersionActionPanel` is editor-embedded; no detail-panel | adapter maps; **ETag signoff pattern has no template analogue** |

### Design implications for T7+
1. `ArtifactViewModel` fields nullable; shell renders sections only when present. Template detail = reduced (no approval-chain section, no distribution tab, no fileSize/pageCount KPIs unless added).
2. `ArtifactApprovalScreen` must stay presentational; the **document approval backend (instance/stages/signoffs + ETag) vs template (reviewer→approver role on the version row) diverge hard** — the adapter normalizes both into `ArtifactActionSet` (available decisions + submit handlers). Do NOT leak ETag/instance specifics into the shared screen.
3. Template lineage: either add a backend list endpoint (contract change — weigh in T12/T13) or build lineage from `getVersion(1..latest_version)` sequential fetch. Prefer the smaller move first; surface as a decision.
4. `code` for templates: decide display fallback (`key`/`name`) — no regulated code exists. Carry as an explicit adapter decision.
5. Stale comment to fix: `features/templates/lib/canActOnVersion.ts:84-85` still says approve "spawns next draft in-tx" — false after M1·T2. Clean up (M3 docs or inline).
