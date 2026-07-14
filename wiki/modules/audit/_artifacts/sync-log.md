# Sync log — audit

## 2026-06-11 — Adversarial-verification anchor correction pass

- **Context:** Adversarial verification of audit.md Key files section and §5.2 Public surface table found ten false/stale anchors and one stale grant claim; this entry records the verified corrections. The prior 2026-06-10 entry claimed handler.go anchors were patched (67→69, 73→75) but audit.md still carries the old values — those patches were not applied and are included here.
- **Mode:** lite patch
- **Anchors corrected (Key files section, audit.md lines 8-12):**
  - `port.go:8-31` was false. Correct individual anchors: `Event` struct at `port.go:13-23`; `ListEventsQuery` at `port.go:71-82`; `Writer` interface at `port.go:96-99`; `Reader` interface at `port.go:107-109`.
  - `service.go:94-99` description "Service.ListEvents" was misleading. Lines 94-99 are inside the `normalizeQuery` helper (limit-clamp), not inside `ListEvents`. `ListEvents` is at `service.go:65-76`. Description changed to "normalizeQuery limit-clamp (default 50; max 100)".
  - `handler.go:67` was false. Line 67 is the `WithExporter` method definition. `RegisterRoutes` is at `handler.go:69`. (Also claimed corrected in 2026-06-10 entry but not applied.)
  - `handler.go:73` was false. Line 73 is inside the `RegisterRoutes` body. `handleEvents` is at `handler.go:75`. (Also claimed corrected in 2026-06-10 entry but not applied.)
  - `postgres/writer.go:20,44` as "Record (INSERT) + ListEvents (SELECT)" was incomplete. `Record` is at `writer.go:27`; `RecordTx` is at `writer.go:44`; `ListEvents` is at `writer.go:142`. The original claim conflated `Record` and `RecordTx` and omitted `ListEvents` entirely.
- **Anchors corrected (§5.2 Public surface table, audit.md lines 147-153):**
  - `service.go:94-99` row for `Service.ListEvents` — same issue as Key files: corrected to `service.go:65-76`.
  - `postgres/writer.go:12,16,20,44` row: `Writer` struct is at `15` (not 12); `NewWriter` at `23` (not 16); `Record` at `27` (not 20); `RecordTx` at `44` (correct); `ListEvents` at `142` (not 44). Anchor corrected to `writer.go:15,23,27,44,142`.
  - `memory/writer.go:11,16,20,27` row: `Writer` struct is at `14` (not 11); `NewWriter` at `19` (not 16); `Record` at `23` (not 20); `RecordTx` at `30` (not present in prior anchor); `ListEvents` at `34` (not 27). Anchor corrected to `writer.go:14,19,23,30,34`.
- **Factual correction (audit.md line 27):** Statement "metaldocs_app has only INSERT" is stale. Migration `0193_audit_events_hash_chain.sql:110` adds a SELECT grant. Corrected to "metaldocs_app has INSERT and SELECT".
- **Prior sync-log integrity note:** The 2026-06-10 entry claimed handler.go anchors 67→69 and 73→75 were patched. Code inspection confirms audit.md still carried the old values 67 and 73 at the time of this pass, so those patches were never applied or were reverted. The corrections are applied in this pass.
- **Public surface:** no new rows; existing rows' anchors corrected only.
- **Routes/API:** no changes — handler.go line numbers for route registration were already correct in §5.3 and the Route Truth Table (`handler.go:69`, `handler.go:71`, `handler.go:73`). Only the Key files and §5.2 rows carried the stale anchors.
- **T-NNN touched:** none opened or closed.
- **R-NNN touched:** none.
- **Counts after:** Critical=0 Major=0 Minor=1 (grant-staleness note now corrected; prior counts unchanged for structural items).
- **Tally gate:** PASS — all ten adversarial findings addressed.
- **Patched files:** `wiki/modules/audit.md`; `wiki/modules/audit/_artifacts/sync-log.md`

## 2026-06-10 — Stage-1 backend audit drift patch

- **Context:** Stage-1 mapper found four wiki/code mismatches; this patch corrects them.
- **Mode:** lite patch
- **Anchors moved:** `service.go:18` → `service.go:65-76` (Service.ListEvents); `service.go:94-99` documents limit-clamp inside normalizeQuery helper (not ListEvents); `handler.go:67` → `handler.go:69` (RegisterRoutes), `handler.go:73` → `handler.go:75` (handleEvents); `memory/writer.go:12` → `memory/writer.go:14` (Writer struct)
- **Public surface:** §5.2 `Service.ListEvents` row corrected (limit [1..200]→[1..100]); export symbols already present in §5.2 — no new rows needed
- **Routes/API:** §5.3 HTTP operations table expanded from 1 row to 4 rows (added POST export, GET export/{id}, GET export/{id}/download with authz bindings); API Route Truth Table expanded to 4 rows; module contract status updated to "Partially contracted"
- **Runtime flows:** §6.2 sequence diagram clamp note corrected [1..200]→[1..100]
- **Pagination:** §8.4 limit range corrected [1..200]→[1..100]; MaxLimit source documented
- **Transactions:** §8.6 rewritten — `RecordTx(ctx, *sql.Tx, Event) error` exists on `domain/port.go:98`; implemented by postgres (`writer.go:44`) and memory (`writer.go:30`); actively called by `bypassAuditAdapter` and `documentsAuditAdapter`; pre-RecordTx claim removed
- **T-NNN touched:** T-009 closed — SELECT grant confirmed in `archive/migrations/0193_audit_events_hash_chain.sql:110`
- **R-NNN touched:** none
- **Counts after:** Critical=2 Major=4 Minor=5 (T-009 closed; was Minor=6)
- **Tally gate:** PASS — all four drift items addressed
- **Patched files:** `wiki/modules/audit.md`; `wiki/modules/audit-tech-debt.md`; `wiki/modules/audit/_artifacts/sync-log.md`

## 2026-05-16 - documents audit adapter name polish

- **Context:** uncommitted diff: stale wiki references corrected from the current `documentsAuditAdapter` name.
- **Mode:** lite patch
- **Anchors moved:** `main.go:445-479` → `main.go:773-803` (documentsAuditAdapter; prior ranges 445-479 and 477-492 were both incorrect)
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
