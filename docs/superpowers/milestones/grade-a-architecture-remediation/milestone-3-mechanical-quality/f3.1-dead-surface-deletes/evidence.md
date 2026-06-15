# Feature F3.1 — Evidence (verify-already-done)

> **Milestone:** 3  ·  **Feature:** `f3.1-dead-surface-deletes`  ·  **Closed:** 2026-06-14
> **Contract:** `milestone.md` F3.1 row (amended). No code to write — the 7 orphans were already
> deleted in **Wave 2.11** (`63f74368`, 2026-06-12) and **Wave 2.12** (H-6a), *before* the 2026-06-13
> audit that re-flagged them. This row records the zero-caller proof against the **current** tree.

## What was implemented

Nothing. F3.1 is a verify-already-done row. The dead-surface deletes were performed in prior waves;
this feature confirms they remain absent in the current tree and that no resurrected reference exists.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| Zero-caller proof — 6 named symbols | `grep -nE 'CutoverService\|CompositionConfig\|func \(s \*AreaService\) SetParent\|resolvePermissionFallback\|ReviewReminderDays\|SnapshotFromTemplate' **/*.go` | **No matches found** (0 `.go` references) | real |
| Zero-caller proof — `SetParent` (broad) | `grep -n 'SetParent' **/*.go` | **No matches found** | real |
| `coverage_boost_test.go` not a delete target | Glob + read | File present at `internal/modules/documents/approval/application/coverage_boost_test.go`; exercises only **live** production symbols (RealClock, NewServices, SubmitRevisionForReview, RecordSignoff, PublishApproved, … 16 sections); CutoverService section removed in Wave 2.11 | real |

> The `grep` is over the working tree at 2026-06-14. The 2026-06-11 dead-code-hygiene artifact's
> file:line anchors are stale (the symbols no longer exist at those lines because the files/methods
> were deleted). Independent of the subagent investigation, the two greps above were run directly.

## Acceptance vs spec Validation Gate

| Acceptance criterion (from milestone.md F3.1) | Met? | Evidence |
|-------------------------------------|------|----------|
| Zero-caller proof recorded for all 7 named symbols (0 `.go` matches) | yes | rows 1–2 |
| `coverage_boost_test.go` confirmed live (exercises only production symbols) | yes | row 3 |
| Build green / no broken imports | yes | nothing was changed in this feature; the prior-wave deletes already shipped with `go build ./...` green (Wave 2.11 close-out) |

## Review disposition

- Spec-compliance review: n/a — no code change. The reconciliation itself was reviewed at the HS-6
  operator gate ("amend milestone.md, then execute").
- Code-quality review: n/a — no code change.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| Stale wiki/GitNexus refs to the deleted symbols (e.g. `wiki/reviews/2026-05-21-*/taxonomy-7.md` SetParent, stage1 module-taxonomy.md, GitNexus index) | Doc/index drift only — code is correct; no runtime impact | wiki-curator dispatch + GitNexus re-`analyze`; owner: backend agent, next wiki sync |
