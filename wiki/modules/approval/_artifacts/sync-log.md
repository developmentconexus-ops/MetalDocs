# Sync log - approval

> Append-only log of `metaldocs-module-doc-sync` runs against this module. Newest at top.

## 2026-05-13 - Plan 10 approval column rename + v1 route canonicalization

- **Context:** uncommitted Plan 10 implementation diff (document_id rename migration 0194, constraints validation 0195, route prefix sweep)
- **Mode:** structural refresh
- **Anchors moved:** approval_instances.document_v2_id -> document_id; /api/v2/approval* -> /api/v1/approval*
- **Public surface:** repository error mapping updated for renamed unique index names
- **Routes/API:** approval endpoints and doc-scoped approval routes reflected as /api/v1
- **Runtime flows:** unchanged logic; canonical path updates only
- **Persistence:** index/constraint rename alignment reflected
- **Dependencies:** permission resolver + handler route references aligned
- **T-NNN touched:** T-007/T-008/T-009/T-011 evidence updates
- **R-NNN touched:** R-007..R-009/R-011 alignment updates
- **Counts after:** Critical=2 Major=4 Minor=6; missing-ADR=10
- **Tally gate:** PASS
- **Patched files:** wiki/modules/approval.md; wiki/modules/approval-tech-debt.md; wiki/backlog/approval-refactor.md; wiki/modules/approval/_artifacts/*
