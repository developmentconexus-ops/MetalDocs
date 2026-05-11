# Sync log — documents

> Append-only log of `metaldocs-module-doc-sync` runs against this module. Newest at top.

## 2026-05-11 · Plan 6a — close T-005

- **Context:** Plan 6a (commit 0e106ed9) · wrap RenameDocument UPDATE + audit.Write in single transaction
- **Anchors moved:** 0
- **Symbols renamed:** 0
- **T-NNN closed:** T-005 · evidence: Service.RenameDocument uses BeginTx when s.db != nil; UpdateDocumentNameTx added to Repository interface + postgres impl; WriteTx added to Audit interface; documentsV2AuditAdapter.WriteTx wired
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
