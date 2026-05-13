# Sync log — registry

> Append-only log of `metaldocs-module-doc-sync` runs against this module. Newest at top.

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
