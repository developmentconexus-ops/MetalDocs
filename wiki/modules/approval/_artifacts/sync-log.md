# Sync log - approval

## 2026-05-16 - controlled-document review route polish

- **Context:** uncommitted diff: approval inbox navigation target changed from `/controlled-documents/{controlled_document_id}` to `/controlled-documents/{controlled_document_id}`.
- **Mode:** lite patch
- **Anchors moved:** none
- **Public surface:** none
- **Routes/API:** frontend navigation spec/backlog references updated; approval HTTP API unchanged
- **Runtime flows:** review/open-document UI path now targets canonical controlled-document route
- **Persistence:** none
- **Dependencies:** none
- **T-NNN touched:** none
- **R-NNN touched:** none
- **Counts after:** Critical=2 Major=4 Minor=6; missing-ADR=10
- **Tally gate:** PASS preflight
- **Patched files:** `wiki/backlog/caixa-aprovacao.md`; `wiki/modules/approval/_artifacts/sync-log.md`

> Append-only log of `metaldocs-module-doc-sync` runs against this module. Newest at top.

## 2026-05-15 - D4 hard-cutover debt/backlog closure sync (0194)

- **Context:** Worker E wiki/docs lane closeout for approval linkage rename debt row and refactor backlog row.
- **Mode:** lite patch
- **Anchors moved:** debt + backlog status wording
- **Public surface:** unchanged
- **Routes/API:** unchanged runtime; docs keep `/api/v1/documents/*` references
- **Persistence:** T-008 marked closed; `approval_instances` linkage naming synchronized to `document_id` with migration 0194 evidence
- **T-NNN touched:** T-008 -> closed (2026-05-15)
- **R-NNN touched:** R-008 -> merged (migration 0194)
- **Counts after:** Critical=2 Major=4 Minor=6; missing-ADR=10
- **Tally gate:** pending
- **Patched files:** `wiki/modules/approval.md`; `wiki/modules/approval-tech-debt.md`; `wiki/backlog/approval-refactor.md`; `wiki/modules/approval/_artifacts/sync-log.md`
## 2026-05-14 - Plan 12.2 caixa-aprovacao screen reality-first sync

- **Context:** commits `a0a90f7e..3d9572cb` (design/spec, approvals screen implementation, review fixes, backlog/design notes sync)
- **Mode:** lite patch
- **Anchors moved:** none
- **Public surface:** none
- **Routes/API:** none (frontend now uses existing approval routes; no backend route/contract delta)
- **Runtime flows:** none
- **Persistence:** none
- **Dependencies:** none
- **T-NNN touched:** none
- **R-NNN touched:** none
- **Counts after:** Critical=2 Major=4 Minor=6; missing-ADR=10
- **Tally gate:** PASS
- **Patched files:** wiki/modules/approval.md; wiki/modules/approval/_artifacts/sync-log.md
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

## 2026-05-18 - approval unresolved-comments hardening sync

- **Context:** commits after `ac448cdc` closing review findings on approval conflict UX and editor comment-load persistence
- **Mode:** lite patch
- **Anchors moved:** none
- **Public surface:** documented 409 `approval.unresolved_comments` failure mode on final-stage signoff
- **Routes/API:** no route shape change; problem-code behavior now recorded in failure table
- **Runtime flows:** signoff dialog resolves mapped business conflicts inline; unknown approval codes still fall back to safe generic copy
- **Persistence:** none
- **Dependencies:** `wiki/concepts/error-ux.md` updated to reflect shared `resolveErrorMessage(code)` handling on approval dialog conflicts
- **T-NNN touched:** none
- **R-NNN touched:** none
- **Counts after:** Critical=2 Major=4 Minor=6; missing-ADR=10
- **Tally gate:** PASS
- **Patched files:** `wiki/modules/approval.md`; `wiki/modules/approval/_artifacts/sync-log.md`; `wiki/concepts/error-ux.md`
