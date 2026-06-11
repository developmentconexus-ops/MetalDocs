# Sync log - controlled-documents

> Append-only log of `metaldocs-module-doc-sync` runs against this module. Newest at top.

## 2026-06-10 — Stage-1 backend audit drift patch

- **Context:** Stage-1 mapper artifact `wiki/backend/_artifacts/stage1/module-controlled-documents.md` (written 2026-06-10) identified 5 drift categories between existing wiki docs and the current codebase on branch `qa/iam-area-membership`.
- **Mode:** lite patch (surgical corrections only; no restructure)
- **Affected modules:** controlled-documents only
- **Affected-surface scan:** file:line anchors, removed symbols, closed tech-debt statuses, visibility grant tables, tripwire trigger migration reference, RFC 9457 gap status
- **Changes applied:**
  1. `wiki/modules/controlled-documents.md`: fixed `service.go:293` anchor for `Obsolete`/`Supersede` → `service.go:451`/`service.go:455`; added `service.go:389` for `PreviewCode` and `service.go:584` for `CreateRevision`; removed stale `application/migration.go:13 BackfillLegacyDocuments` row from section 5.2; updated coverage stats from 90 → 94; updated section 7 deployment note; updated `Last verified` stamp to `2026-06-11` — adversarial verification confirmed stamp present and correct on 2026-06-11 (line 5).
  2. `wiki/modules/controlled-documents/_artifacts/01-surface.md`: replaced all `internal/modules/registry/` path prefixes with `internal/modules/controlleddocuments/`; removed `application/migration.go` file-tree entry, `application/migration_integration_test.go` file-tree entry, and `BackfillLegacyDocuments` public-surface table row; updated stale line-number anchors for `Obsolete` (now line 57, `service.go:451`) and `Supersede` (now line 58, `service.go:455`) — adversarial verification confirmed correct anchors present on 2026-06-11.
  3. `wiki/modules/controlled-documents/_artifacts/03-deps.md`: replaced all `internal/modules/registry/` path prefixes with `internal/modules/controlleddocuments/`; updated header title and note comment. **Residual drift (not applied):** auto-generated header metadata (`<!-- date: 2026-05-10 -->` on line 3) was not updated to reflect the 2026-06-10 edit date declared in the patch note on line 4 — generator metadata and patch note are now inconsistent.
  4. `wiki/modules/controlled-documents/_artifacts/04-persistence.md`: updated module name and path; added `controlled_document_area_grants` and `controlled_document_user_grants` to section 1; updated section 3 triggers to reference migration 0231 (not 0142b) and documented GUC writes; updated section 5 tripwire pairing audit to reflect T-004 closure (`Create` at `repository.go:341`, `CreateTx` at `repository.go:362`); appended 6 missing migrations (0188, 0198, 0203, 0210, 0225, 0231) to section 6; bumped Last verified.
  5. `wiki/modules/controlled-documents/_artifacts/05-industry.md`: updated title; marked IP-001 APPLIED (Plan 7 / T-003 + T-007 closed); updated IP-004 to reflect T-001 + T-004 closed with remaining T-006/T-005 gaps; updated IP-005 path references; updated IP-006 migration count; updated IP-008 GUC note; replaced all stale "Registry state" labels with "Module state"; updated summary table.
  6. `wiki/modules/controlled-documents-tech-debt.md`: bumped Last verified to `2026-06-11` (applied successfully; adversarial verification confirmed correct value on 2026-06-11).
- **T-NNN touched:** T-003 (confirmed closed in artifacts); T-004 (confirmed closed in artifacts); T-007 (confirmed closed in artifacts)
- **R-NNN touched:** none
- **Counts after:** Critical=2 Major=6 Minor=4; public symbols undocumented=79/94
- **Tally gate:** READ-ONLY — no source changes; drift patch only
- **Patched files:** `wiki/modules/controlled-documents.md`; `wiki/modules/controlled-documents-tech-debt.md`; `wiki/modules/controlled-documents/_artifacts/01-surface.md`; `wiki/modules/controlled-documents/_artifacts/03-deps.md`; `wiki/modules/controlled-documents/_artifacts/04-persistence.md`; `wiki/modules/controlled-documents/_artifacts/05-industry.md`; `wiki/modules/controlled-documents/_artifacts/sync-log.md`
- **Known residual drift after this run:** (1) `wiki/modules/controlled-documents/_artifacts/03-deps.md` generator-metadata comment (`<!-- date: 2026-05-10 -->` on line 3) was not updated to reflect the 2026-06-10 edit date declared in the patch note on line 4 — generator metadata and patch note are inconsistent.

## 2026-05-25 - backend medium quality-bar sync

- **Context:** uncommitted diff on `fix/phase10-controlleddocs` for Worker 10A, scoped to `internal/modules/controlleddocuments/`
- **Mode:** lite patch
- **Affected modules:** controlled-documents only
- **Affected-surface scan:** public domain/repository interfaces changed (`DBTX`, clone-template/document-ref constructors); application/repository error handling and warning logging changed; routes, OpenAPI/codegen, DB schema/migrations, persistence tables, and runtime route ownership unchanged
- **Public surface:** added `NewCloneTemplateRequest`, `NewDocumentRef`, `ErrCloneTemplateNameRequired`, `ErrDocumentRefIDRequired`; repository Tx methods now accept `controlleddocumentsdomain.DBTX`
- **Routes/API:** none
- **Runtime flows:** no route or lifecycle flow changes; error context/logging only
- **Persistence:** no schema/query-shape changes beyond error wrapping and existing RowsAffected checks
- **Dependencies:** no module dependency changes
- **T-NNN touched:** T-012 count refreshed for constructor additions; no debt row closed
- **R-NNN touched:** none
- **Counts after:** Critical=2 Major=6 Minor=4; public symbols undocumented=79/94
- **Tally gate:** BLOCKED by Git Bash/Windows access error during preflight (`CreateFileMapping ... Win32 error 5`)
- **Patched files:** `wiki/modules/controlled-documents.md`; `wiki/modules/controlled-documents-tech-debt.md`; `wiki/modules/controlled-documents/_artifacts/sync-log.md`

## 2026-05-21 - generated boundary mount + freeze workflow sync

- **Context:** uncommitted diff in this thread (controlled-documents runtime mount canonicalization + freeze workflow docs/skills hardening)
- **Mode:** lite patch
- **Anchors moved:** module title + Last verified stamp + API Route Truth runtime-owner references
- **Public surface:** no exported symbol additions/removals; route ownership expression updated to generated boundary dispatch
- **Routes/API:** runtime route mounting documented as `controlleddocumentsapi.HandlerWithOptions` dispatch (replacing stale per-route wrapper references in route-truth table)
- **Runtime flows:** no behavioral flow rewrites; idempotency error normalization note added to changelog
- **Persistence:** none
- **Dependencies:** none
- **T-NNN touched:** none
- **R-NNN touched:** none
- **Counts after:** Critical=2 Major=6 Minor=4; missing-ADR=9
- **Tally gate:** PASS
- **Patched files:** `wiki/modules/controlled-documents.md`; `wiki/modules/controlled-documents/_artifacts/sync-log.md`

## 2026-05-16 - release V2 frontend route polish

- **Context:** uncommitted diff: controlled-document route polish in frontend
- **Mode:** lite patch
- **Anchors moved:** `pages/ControlledDocumentsPage.tsx`
- **Public surface:** frontend route/path labels only; backend API unchanged
- **Routes/API:** SPA route references updated to `/controlled-documents`; HTTP API unchanged
- **Runtime flows:** approval inbox and editor return navigation target `/controlled-documents`
- **Persistence:** none
- **Dependencies:** controlled-documents frontend dependency artifact updated
- **T-NNN touched:** none
- **R-NNN touched:** none
- **Counts after:** Critical=2 Major=6 Minor=4; missing-ADR=9
- **Tally gate:** PASS preflight
- **Patched files:** `wiki/modules/controlled-documents/_artifacts/03-deps.md`; `wiki/modules/controlled-documents/_artifacts/sync-log.md`

## 2026-05-15 - Novo Documento atomic create runtime repair

- **Context:** uncommitted runtime repair for `POST /api/v1/controlled-documents`
- **Mode:** lite patch
- **Anchors moved:** `Last verified` stamp only
- **Public surface:** `RegistryService.Create` primes `metaldocs.tenant_id` and `metaldocs.actor_id`, then asserts `registry.create` inside the caller transaction
- **Routes/API:** no path/schema change; route truth remained `POST /api/v1/controlled-documents`
- **Runtime flows:** atomic create flow updated with in-tx authz GUC + `registry.create`
- **Persistence:** `cd_sequence_counters` and `controlled_documents` writes documented as protected by in-tx capability assertion
- **Dependencies:** documents dependency clarified as `document.create` plus `document.edit` during `CreateDocumentTx`
- **T-NNN touched:** none
- **R-NNN touched:** none
- **Counts after:** Critical=2 Major=6 Minor=4; missing-ADR=9
- **Tally gate:** PASS
- **Patched files:** `wiki/modules/controlled-documents.md`; `wiki/modules/controlled-documents/_artifacts/02-flow-atomic-create.md`; `wiki/modules/controlled-documents/_artifacts/sync-log.md`

## 2026-05-15 - DB foundation startup and migration policy sync

- **Context:** git range `b940d0d66fe44fa7fd6877d222d89e1bfae1eccd..cf123c5a` (DB foundation implementation and verification hardening)
- **Mode:** structural refresh
- **Anchors moved:** 2 (`Last verified` stamp in module doc; startup maintenance wording in public surface summary)
- **Public surface:** removed stale `RunStartupMigrations` reference from `_artifacts/01-surface.md`; preserved `RunLegacyMaintenance` / `BackfillLegacyDocuments` as recovery-only capability
- **Routes/API:** none
- **Runtime flows:** none
- **Persistence:** none (policy confirmation only; no controlled-documents table shape changes)
- **Dependencies:** removed stale composition-root startup hook reference from `_artifacts/03-deps.md`; startup no longer documents maintenance invocation at API boot
- **T-NNN touched:** none
- **R-NNN touched:** none
- **Counts after:** Critical=2 Major=6 Minor=4; missing-ADR=9
- **Tally gate:** PASS
- **Patched files:** `wiki/modules/controlled-documents.md`; `wiki/modules/controlled-documents/_artifacts/01-surface.md`; `wiki/modules/controlled-documents/_artifacts/03-deps.md`; `wiki/modules/controlled-documents/_artifacts/sync-log.md`
