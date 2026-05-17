# Sync log — audit

## 2026-05-16 - documents audit adapter name polish

- **Context:** uncommitted diff: stale wiki references corrected from the current `documentsAuditAdapter` name.
- **Mode:** lite patch
- **Anchors moved:** `main.go:445-479` -> `main.go:477-492`
- **Public surface:** none
- **Routes/API:** none
- **Runtime flows:** none
- **Persistence:** none
- **Dependencies:** audit consumer dependency docs updated for documents adapter name
- **T-NNN touched:** none
- **R-NNN touched:** none
- **Counts after:** Critical=2 Major=4 Minor=6; missing-ADR=12
- **Tally gate:** PASS preflight
- **Patched files:** `wiki/modules/audit.md`; `wiki/modules/audit/_artifacts/03-deps.md`; `wiki/modules/audit/_artifacts/06-selfreview.md`; `wiki/modules/audit/_artifacts/sync-log.md`

> Append-only log of `metaldocs-module-doc-sync` runs against this module. Newest at top.

## 2026-05-11 · Plan 6a — close T-001 T-003 T-005 T-007

- **Context:** Plan 6a (commits 0279546f..71a2dc53) · gate audit endpoint, retention goroutine, fix fire-and-forget, tenant_id column
- **Anchors moved:** 0
- **Symbols renamed:** 0
- **T-NNN closed:** T-001 (commit 6b34c277), T-003 (main.go goroutine), T-005 (commit 1994bb84), T-007 (commit b5b077b7) · evidence: route gated + errors logged + retention job + tenant_id added
- **R-NNN updated:** R-001 → merged, R-003 → merged (goroutine), R-005 → merged, R-007 → merged · commits per row
- **§11 counts after:** Critical=2 Major=4 Minor=6 (unchanged — closed rows still counted)
- **Tally gate:** PASS
- **Patched files:** wiki/modules/audit-tech-debt.md · wiki/backlog/audit-refactor.md
- **Structural changes noted (sweep needed):** RecordTx added to Writer interface; TenantID field added to Event + ListEventsQuery — §5 Key Files anchors not yet updated
