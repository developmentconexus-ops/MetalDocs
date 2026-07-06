# F9.1 — adr-hygiene (plan)

> Input: `spec.md` (approved). Executor: sonnet subagent(s), main session reviews.

## Plan

### Task 1 — governance rule (docs)
File: `wiki/standards/documentation-governance.md`. Add an "ADR status field" subsection: rule
(≤3 physical lines, ≤400 chars total, canonical vocabulary `Proposed | Accepted | Accepted (amended
YYYY-MM-DD by NNNN) | Superseded by NNNN | Deprecated | Historical` + optional one date/scope line +
optional one history-pointer line), the rationale (mega-status anti-pattern, review 778f494a:105),
and the repeatable sweep command (bash awk one-liner, committed verbatim in the doc).

### Task 2 — ADR 0022 split
- Create `wiki/decisions/0022-execution-history.md`: header explaining provenance (relocated from the
  0022 status field 2026-07-06), then the FULL former status/Last-verified changelog content,
  restructured into dated phase entries. Zero information loss.
- Rewrite 0022 status block to ≤3 lines: `Accepted (fully executed 2026-06-13)` + extends-pointer +
  `Execution history: [0022-execution-history.md]`. Keep `Last verified` as its own short field
  (trim its embedded changelog into the history doc too — same rule applies to the status *block*;
  Last-verified keeps date + one clause).

### Task 3 — 0027 / 0070 / 0015 splits (same pattern, proportional)
These are smaller: history content may fit as a short companion doc OR as a `## Status history`
section INSIDE the same ADR file (below the header block) — status field itself ≤3 lines/≤400 chars.
Prefer in-file section when history <½ page (avoid doc sprawl); 0022 alone clearly needs a companion.

### Task 4 — 0013 supersession research + stamp
- Research: read 0013 D-2/D-5 (allocation at INSERT; publish paths set CurrentRevisionNumber), 0052
  decision (manual revision creation), 0053 (frontend shell), current runtime (templates repository
  SQL + lifecycle.go). Determine: superseded, amended, or intact.
- Stamp per finding (spec interview row 5): expected `Accepted (amended by 0052 …)`; add reciprocal
  reference in 0052 (`Amends/Supersedes` line); update `index.md` rows (0013 + 0052 Superseded-by
  column); write the disposition (incl. why review's "Superseded" wording was refined) into evidence.

### Task 5 — sweep + gate
Run the sweep command from Task 1 over all ADRs → must be 0 violations. Capture output.
Run `git diff --stat` scope check. Fill `evidence.md`.

## Test strategy
Docs feature — the "tests" are the executable sweep (RED before Task 2/3 on 0022/0027/0070/0015,
GREEN after) and the content-preservation diff review. Capture the RED run before splitting (TDD
analog: failing check first).

## Files touched
`wiki/decisions/0022-*.md` (2 files), `0027-*.md`, `0070-*.md`, `0015-*.md`, `0013-*.md`,
`0052-*.md`, `index.md`, `wiki/standards/documentation-governance.md`, feature folder.
