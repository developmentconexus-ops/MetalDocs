# Sync log — registry

## 2026-05-16 - release V2 frontend route polish

- **Context:** uncommitted diff: controlled-document registry route renamed from `/controlled-documents` to `/controlled-documents`; page file renamed the controlled-document page file to `ControlledDocumentsPage.tsx`.
- **Mode:** lite patch
- **Anchors moved:** `pages/ControlledDocumentsPage.tsx`
- **Public surface:** frontend route/path labels only; backend registry API unchanged
- **Routes/API:** SPA route references updated to `/controlled-documents`; HTTP API unchanged
- **Runtime flows:** approval inbox and editor return navigation now target `/controlled-documents`
- **Persistence:** none
- **Dependencies:** registry frontend dependency artifact updated
- **T-NNN touched:** none
- **R-NNN touched:** none
- **Counts after:** Critical=2 Major=6 Minor=4; missing-ADR=9
- **Tally gate:** PASS preflight
- **Patched files:** `wiki/modules/registry/_artifacts/03-deps.md`; `wiki/modules/registry/_artifacts/sync-log.md`

> Append-only log of `metaldocs-module-doc-sync` runs against this module. Newest at top.

## 2026-05-15 - Novo Documento atomic create runtime repair

- **Context:** uncommitted runtime repair for `POST /api/v1/controlled-documents`
- **Mode:** lite patch
- **Anchors moved:** `Last verified` stamp only
- **Public surface:** `RegistryService.Create` now primes `metaldocs.tenant_id` and `metaldocs.actor_id`, then asserts `registry.create` inside the caller transaction before sequence allocation and CD insert
- **Routes/API:** no path/schema change; route truth remains `POST /api/v1/controlled-documents`
- **Runtime flows:** atomic create flow updated to include in-tx authz GUC + `registry.create`
- **Persistence:** `cd_sequence_counters` and `controlled_documents` writes now documented as protected by in-tx capability assertion
- **Dependencies:** Documents v2 dependency clarified as `document.create` plus `document.edit` during `CreateDocumentTx`
- **T-NNN touched:** none
- **R-NNN touched:** none
- **Counts after:** Critical=2 Major=6 Minor=4; missing-ADR=9
- **Tally gate:** PASS
- **Patched files:** `wiki/modules/registry.md`; `wiki/modules/registry/_artifacts/02-flow-atomic-create.md`; `wiki/modules/registry/_artifacts/sync-log.md`

## 2026-05-15 - DB foundation startup and migration policy sync

- **Context:** git range `b940d0d66fe44fa7fd6877d222d89e1bfae1eccd..cf123c5a` (DB foundation implementation and verification hardening)
- **Mode:** structural refresh
- **Anchors moved:** 2 (`Last verified` stamp in module doc; startup maintenance wording in public surface summary)
- **Public surface:** removed stale `RunStartupMigrations` reference from `_artifacts/01-surface.md`; preserved `RunLegacyMaintenance` / `BackfillLegacyDocuments` as recovery-only capability
- **Routes/API:** none
- **Runtime flows:** none
- **Persistence:** none (policy confirmation only; no registry table shape changes)
- **Dependencies:** removed stale composition-root startup hook reference from `_artifacts/03-deps.md`; startup no longer documents registry maintenance invocation at API boot
- **T-NNN touched:** none
- **R-NNN touched:** none
- **Counts after:** Critical=2 Major=6 Minor=4; missing-ADR=9
- **Tally gate:** PASS
- **Patched files:** `wiki/modules/registry.md`; `wiki/modules/registry/_artifacts/01-surface.md`; `wiki/modules/registry/_artifacts/03-deps.md`; `wiki/modules/registry/_artifacts/sync-log.md`

## 2026-05-11 · Plan 6a — close T-002 T-008

- **Context:** Plan 6a (commits 5bb06964 + 71a2dc53) · emit governance event on Obsolete/Supersede; route registry audit through AuditGovernanceAdapter
- **Anchors moved:** 0
- **Symbols renamed:** 0
- **T-NNN closed:** T-002 (commit 5bb06964), T-008 (commit 71a2dc53) · evidence: govLogger.Log called after changeStatus commit; registry module now uses AuditGovernanceAdapter instead of borrowing taxonomy DBGovernanceLogger
- **R-NNN updated:** R-002 → merged, R-008 → merged · commits per row
- **§11 counts after:** Critical=2 Major=6 Minor=4 (unchanged)
- **Tally gate:** PASS
- **Patched files:** wiki/modules/registry-tech-debt.md · wiki/backlog/registry-refactor.md
- **Structural changes noted (sweep needed):** AuditWriter field added to Dependencies struct; auditdomain OUT-edge added to registry/module.go — §5 Key Files + §8 cross-deps not yet updated

## 2026-05-13 - Plan 10 registry route canonicalization

- **Context:** uncommitted Plan 10 implementation diff (/api/v2 controlled-documents to /api/v1)
- **Mode:** structural refresh
- **Anchors moved:** /api/v2/controlled-documents* -> /api/v1/controlled-documents*
- **Public surface:** no semantic API change; canonical path update
- **Routes/API:** route truth/artifacts refreshed to v1
- **Runtime flows:** unchanged behavior
- **Persistence:** none
- **Dependencies:** resolver + frontend references aligned to v1
- **T-NNN touched:** T-010 evidence cross-check maintained
- **R-NNN touched:** R-010/R-100 status alignment retained
- **Counts after:** Critical=2 Major=6 Minor=4; missing-ADR=9
- **Tally gate:** PASS
- **Patched files:** wiki/modules/registry.md; wiki/modules/registry-tech-debt.md; wiki/backlog/registry-refactor.md; wiki/modules/registry/_artifacts/*
