# Program: <Program Name>

> **Governing spec:** `<path to the design/spec doc that defines this program>`
> **Status:** Planning | In progress (Milestone <n>) | Complete
> **Owner / operator:** <name>

One-paragraph statement of the program's goal and its terminal acceptance
(what proves the whole program is done — e.g. an independent re-audit passing a
named bar).

## Milestones

| # | Milestone | Objective (one line) | Status | Gate result |
|---|-----------|----------------------|--------|-------------|
| 0 | `milestone-0-<slug>` | <objective> | planned / in-progress / passed | `qa/milestone-qa.md` |
| 1 | `milestone-1-<slug>` | <objective> | planned | — |
| … | | | | |

Status vocabulary: `planned` → `in-progress` → `passed` (operator-approved) /
`blocked` (hard-stop open). The **Gate result** column links the milestone-validator's
verdict (`qa/milestone-qa.md`); `passed` requires a validator **PASS** *and* operator HS-1 approval.

## Hard-stops raised

| When | HS id | What | Resolution |
|------|-------|------|------------|
| | | | |

## Program close-out / reconciliation

Fill in only when the last milestone has passed:

- [ ] Every planned feature has a complete evidence row.
- [ ] Zero unplanned scope (anything added is recorded with rationale).
- [ ] Every bounded defer has a written trigger and an owner.
- [ ] Terminal acceptance passed — link the evidence.
- [ ] Operator sign-off: <date / name>
