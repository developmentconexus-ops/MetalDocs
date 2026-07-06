# F9.4 — doc-truth (plan)

> Input: `spec.md` (approved). Executor: sonnet subagent + wiki-curator agent; main reviews.
> Runs after F9.3; final verification re-run after F9.5 lands (layout text must match final tree).

## Plan

### Task 1 — runtime extraction (read-only)
Capture: `ls internal/modules/`; `chain.go` link list; the janitor set from
`internal/platform/jobs/river/.../maintenance` (`PeriodicJobs()`) + retention + jobs-binary periodic
config; where idempotency actually executes (approval idemp stores + any other per-handler sites —
grep `idempotency_keys` consumers). Paste into evidence as the truth basis.

### Task 2 — CLAUDE.md corrections
- System Facts: module list −docs +tokens; add one-line approval-exception footnote (ADR ref from
  F9.5).
- Lifecycle line: replace hand-listed chain with the chain.go-truth list (or shorten + anchor
  `chain.go:25`); remove the idempotency link claim; state idempotency = per-handler where
  implemented.
- Janitor/scheduler sentence: rewrite per extracted truth (River periodic + leader election; watchdog
  alert-only ADR 0068; no lease-reaper if retired — verify).

### Task 3 — skill reference fix
`.claude/skills/developing-new-work/references/invariant-checklist.md` lifecycle line (~:56) same
correction; bump `Last verified`.

### Task 4 — enumerate mission-touched wiki docs
From M0–M9 milestone evidence + `git log --name-only <mission range> -- wiki/` (mission range =
first M0 commit..HEAD): unique wiki file list → evidence.md table.

### Task 5 — wiki-curator pass
Dispatch wiki-curator agent over the Task-4 list: refresh `Last verified` stamps where content was
verified, fix broken file:line anchors (post-F9.5 renames!), update wiki/README.md index entries if
needed. Capture its report.

### Task 6 — final verification (post-F9.5)
Re-run Task 1 extraction; confirm CLAUDE.md still exact (esp. after renames); evidence.md complete.

## Files touched
`CLAUDE.md`, `wiki/**` (stamps/anchors), `.claude/skills/developing-new-work/references/invariant-checklist.md`, feature folder.
