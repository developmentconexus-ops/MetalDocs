## 2026-05-19 - editor sidebar identification visual sync

- **Context:** uncommitted frontend diff for `EditorMetaSidebar` identity layout refinement.
- **Mode:** lite patch
- **Anchors moved:** none
- **Public surface:** frontend-only sidebar copy/style changed; no exported backend/API surface changed
- **Routes/API:** none; page count/file size intentionally classified as DB/API/TanStack follow-up prerequisite before UI exposure
- **Runtime flows:** editor sidebar still uses real document detail, governed revision history, and approval-chain data; no workflow behavior changed
- **Persistence:** none
- **Dependencies:** documents frontend remains feature-local under `frontend/apps/web/src/features/documents/components`
- **T-NNN touched:** none
- **R-NNN touched:** none
- **Counts after:** Critical=1 Major=7 Minor=4; missing-ADR=8
- **Tally gate:** PASS (severity count drift repaired; warnings remain for backlog rows T-007/T-010/T-011/T-012)
- **Patched files:** `wiki/modules/documents.md`; `wiki/modules/documents/_artifacts/sync-log.md`

## 2026-05-15 - D4 hard-cutover wiki/docs sync

## 2026-05-17 - active v2 reference memory sync

- **Context:** post-merge scan after documents/templates v2 name polish and template wizard completion; active wiki/backlog surfaces still referenced removed `documentsV2.ts`, `/api/v2/documents`, and historical route names.
- **Mode:** lite patch.
- **Affected-surface scan:** `rg` across active wiki/backlog/architecture/module docs; historical migrations, release inventory, runbooks, and test artifacts intentionally left unchanged.
- **Routes/API:** no runtime route change; documentation corrected active future endpoints to `/api/v1/documents/*`.
- **Runtime flows:** none.
- **Persistence:** none.
- **Debt/backlog:** editor backlog contract-gate blocker marked resolved after `scripts/check-module-contract-sync.ps1 -Module documents` passed; distribution backlog future endpoints canonicalized to v1.
- **Tally gate:** preflight PASS before edits; final tally recorded in session output.
- **Patched files:** `internal/modules/documents/application/fillin_service.go`; `wiki/backlog/editor.md`; `wiki/backlog/distribuicao.md`; `wiki/architecture/system-overview.md`; `wiki/modules/documents/_artifacts/sync-log.md`.

## 2026-05-16 - documents frontend V2 naming polish

- **Context:** uncommitted diff: documents frontend API wrappers renamed editor hooks folder renamed the editor hook folder to `hooks/editor`.
- **Mode:** lite patch
- **Anchors moved:** `features/documents/hooks/editor/*`
- **Public surface:** frontend-only import/file naming changed; document HTTP API unchanged
- **Routes/API:** no backend route/schema change
- **Runtime flows:** editor/autosave/comments/pdf hooks keep existing behavior under canonical path names
- **Persistence:** IndexedDB DB name changed from `metaldocs_docs`
- **Dependencies:** audit adapter reference corrected to `documentsAuditAdapter`
- **T-NNN touched:** none
- **R-NNN touched:** none
- **Counts after:** Critical=1 Major=5 Minor=4; missing-ADR=6
- **Tally gate:** PASS preflight
- **Patched files:** `wiki/modules/documents.md`; `wiki/modules/documents/_artifacts/sync-log.md`

- **Context:** Worker E wiki/docs lane cutover refresh for current route/API truth and approval linkage rename evidence.
- **Mode:** lite patch
- **Anchors moved:** route/API wording only
- **Public surface:** `/documents/new`, `/documents/:documentID/edit`, `/documents/:documentId`
- **Routes/API:** `/api/v1/documents/*` canonical in current-truth sections; stale `/api/v2/documents` removed from active docs wording
- **Persistence:** approval linkage wording reflects `document_id` naming (0194 evidence preserved via docs)
- **T-NNN touched:** none
- **R-NNN touched:** none
- **Counts after:** Critical=1 Major=5 Minor=4; missing-ADR=6
- **Tally gate:** pending
- **Patched files:** `wiki/modules/documents.md`; `wiki/modules/documents/_artifacts/04-persistence.md`; `wiki/modules/documents/_artifacts/sync-log.md`
# Sync log — documents

> Append-only log of `metaldocs-module-doc-sync` runs against this module. Newest at top.

## 2026-05-19 - Document revision artifact metadata runtime sync

- **Context:** commits `ccdf7e16`, `14cd2005`, and `f1cf7988` from the 2026-05-19 artifact metadata plan; migration 0206/0207, documents API contract, generated frontend types, EigenPal wrapper page count, autosave commit plumbing, and editor sidebar rendering.
- **Mode:** structural refresh
- **Anchors moved:** none
- **Public surface:** `DocumentDetailResponse` now exposes `currentRevisionFileSizeBytes`, `currentRevisionPageCount`, and `currentRevisionPageCountSource`; autosave commit response includes `file_size_bytes`, `page_count`, and `page_count_source`; `MetalDocsEditorRef` includes `getPageCount()`.
- **Routes/API:** `POST /api/v1/documents/{id}/autosave/commit` accepts `page_count`; `GET /api/v1/documents/{id}` reads current-head artifact metadata through `documents.current_revision_id -> document_revisions`.
- **Runtime flows:** editor autosave collects EigenPal page count client-side, backend computes saved DOCX size server-side, and sidebar renders pages/size from real API/runtime data only.
- **Persistence:** `public.document_revisions` carries technical artifact metadata (`file_size_bytes`, `page_count`, `page_count_source`); governed revision history remains sourced from `public.documents` lineage by `controlled_document_id`.
- **Dependencies:** EigenPal remains isolated behind `packages/editor-ui`; frontend documents feature consumes generated API contract via documents API wrapper/TanStack detail state.
- **T-NNN touched:** none
- **R-NNN touched:** none
- **Counts after:** Critical=1 Major=7 Minor=4; missing-ADR=8
- **Tally gate:** PASS before edits and after sync patch
- **Patched files:** `wiki/modules/documents.md`; `wiki/modules/documents/_artifacts/sync-log.md`; `wiki/database/tables/document_revisions.md` already updated in DB slice. `_artifacts/04-persistence.md` was not patched because it currently contains invalid UTF-8 bytes that block safe `apply_patch` editing.

### 2026-05-19 review follow-up

- **Context:** code-review follow-up for commits through `12431266`.
- **Changes:** `GET /api/v1/documents/{id}` now exposes governed `RevisionNumber`; editor submission title gating uses `RevisionNumber` instead of technical `RevisionVersion`; approval submission hashes use governed `RevisionNumber` instead of technical `RevisionVersion`; finalize profile prerequisite and invalid autosave page-count errors now use the API problem envelope; autosave request/response wrapper types derive from generated OpenAPI operations; baseline/migration truth for zero-based revision numbers no longer rewrites `updated_at`.
- **Verification:** focused documents frontend tests, frontend build typecheck, `go generate ./internal/modules/documents/api/...`, `go test -vet=off ./internal/modules/documents/... -count=1`, and DB dictionary coverage passed. Redocly lint remained blocked by local `npm ECOMPROMISED` before execution.

## 2026-05-15 - Novo Documento blank-template create tripwire repair

- **Context:** uncommitted runtime repair for Documents v2 initialization inside Registry atomic create
- **Mode:** lite patch
- **Anchors moved:** none
- **Public surface:** `CreateDocumentTx` behavior updated; it asserts `document.create` before INSERT and `document.edit` before initial pointer/snapshot UPDATEs
- **Routes/API:** none
- **Runtime flows:** atomic CD+draft create note updated for Plan 5 tripwire behavior
- **Persistence:** `documents` INSERT/UPDATE tripwire pairing updated for `CreateDocumentTx`
- **Dependencies:** Registry remains caller/owner of atomic create; Documents v2 remains row-initialization port provider
- **T-NNN touched:** none
- **R-NNN touched:** none
- **Counts after:** Critical=1 Major=5 Minor=4; missing-ADR=6
- **Tally gate:** PASS
- **Patched files:** `wiki/modules/documents.md`; `wiki/modules/documents/_artifacts/04-persistence.md`; `wiki/modules/documents/_artifacts/sync-log.md`

## 2026-05-11 · Plan 6a — close T-005

- **Context:** Plan 6a (commit 0e106ed9) · wrap RenameDocument UPDATE + audit.Write in single transaction
- **Anchors moved:** 0
- **Symbols renamed:** 0
- **T-NNN closed:** T-005 · evidence: Service.RenameDocument uses BeginTx when s.db != nil; UpdateDocumentNameTx added to Repository interface + postgres impl; WriteTx added to Audit interface; documentsAuditAdapter.WriteTx wired
- **R-NNN updated:** R-005 → merged · commit 0e106ed9
- **§11 counts after:** Critical=1 Major=5 Minor=4 (unchanged)
- **Tally gate:** PASS (pre-existing WARNs T-007/T-010 have no backlog row — latent by design)
- **Patched files:** wiki/modules/documents-tech-debt.md · wiki/backlog/documents-refactor.md
- **Structural changes noted (sweep needed):** WithDB builder added to Service; UpdateDocumentNameTx added to Repository interface; WriteTx added to Audit interface — §5 Key Files not yet updated

## 2026-05-11 · Plan 4 — capability namespace closure (documents T-008)

- **Context:** Plan 4 tasks 1-9 completed: submit_service.go:85 changed from `"doc.submit"` literal to `string(iamdomain.CapDocumentSubmit)`; ~24 handler/service files renamed `authz.ErrCapabilityDenied`→`authz.ErrCapDenied`; migration 0186 renamed `doc.submit`→`document.submit` in role_capabilities table
- **Anchors moved:** 0 (submit_service.go:85 line number unchanged — mechanical replacement only)
- **Symbols renamed:** 1 (authz.ErrCapabilityDenied→authz.ErrCapDenied — updated §8.1 sentinel note)
- **T-NNN closed:** T-008 · evidence: submit_service.go:85 now uses typed `iamdomain.CapDocumentSubmit` const; `doc.*` namespace straddle resolved
- **R-NNN updated:** R-008→merged · PR: Plan 4 (2026-05-11, commit 3a227642)
- **§11 counts after:** Critical=1 Major=5 Minor=3
- **Tally gate:** PASS
- **Patched files:** wiki/modules/documents.md · wiki/modules/documents-tech-debt.md · wiki/backlog/documents-refactor.md

## 2026-05-13 - Plan 10 documents route + approval linkage canonicalization

- **Context:** uncommitted Plan 10 implementation diff (/api/v2/documents* -> /api/v1/documents*; approval instance linkage column rename)
- **Mode:** structural refresh
- **Anchors moved:** documents endpoints and approval-related document routes to /api/v1
- **Public surface:** no new endpoints; canonical path updates only
- **Routes/API:** route truth/artifacts refreshed to v1 paths
- **Runtime flows:** unchanged behavior
- **Persistence:** approval_instances document_id rename reflected in debt narrative
- **Dependencies:** idempotency route templates and resolver mapping aligned
- **T-NNN touched:** documents debt references updated to v1 naming
- **R-NNN touched:** documents refactor rows remain linked; wording canonicalized
- **Counts after:** Critical=1 Major=5 Minor=4; missing-ADR=6
- **Tally gate:** PASS
- **Patched files:** wiki/modules/documents.md; wiki/modules/documents-tech-debt.md; wiki/backlog/documents-refactor.md; wiki/modules/documents/_artifacts/*
## 2026-05-18 - approval comments + editor error-state sync

- **Context:** post-review hardening for unresolved-comment approval enforcement and persistent editor comment-load error UX
- **Mode:** lite patch
- **Anchors moved:** none
- **Public surface:** no new routes; documented existing 409 `approval.unresolved_comments` behavior and editor-side retry banner expectations
- **Routes/API:** none
- **Runtime flows:** document editor now keeps comment-query failures visible with retry; approval-release prerequisite row closed in editor backlog
- **Persistence:** documents T-011 closed with evidence in approval decision service/tests
- **Dependencies:** `wiki/backlog/editor.md` and `wiki/concepts/error-ux.md` synced to the shipped behavior
- **T-NNN touched:** T-011 -> closed (2026-05-18)
- **R-NNN touched:** none
- **Counts after:** Critical=1 Major=5 Minor=3
- **Tally gate:** pending
- **Patched files:** `wiki/modules/documents.md`; `wiki/modules/documents-tech-debt.md`; `wiki/modules/documents/_artifacts/sync-log.md`; `wiki/backlog/editor.md`; `wiki/concepts/error-ux.md`
