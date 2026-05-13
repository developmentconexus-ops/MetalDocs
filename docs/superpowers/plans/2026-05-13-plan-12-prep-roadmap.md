# Plan 12 Prep Roadmap Correction Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Correct roadmap state before Plan 12 screen work starts.

**Architecture:** This is a documentation-only prep slice. It updates the roadmap so prerequisite evidence is not contradictory and Plan 12 is open until the seven screen PRs actually land.

**Tech Stack:** Markdown docs, Git.

---

## File Structure

- Modify: `wiki/backlog/roadmap.md`
  - Update `Last verified` to reflect this Plan 12 prep correction.
  - Correct the Plan 11 execution-order table row to match the Plan 11 body evidence.
  - Correct the Plan 12 body status from done to open/not started and link the Plan 12 spec.
- Create: `docs/superpowers/plans/2026-05-13-plan-12-prep-roadmap.md`
  - Records this implementation plan and its verification steps.

## Task 1: Correct Roadmap Status

**Files:**
- Modify: `wiki/backlog/roadmap.md`
- Verify: `git diff -- wiki/backlog/roadmap.md`

- [ ] **Step 1: Confirm current contradictory status**

Run:

```powershell
rg -n "Plan 11|Plan 12|Last verified|Status:" wiki/backlog/roadmap.md
```

Expected:

- The execution table shows Plan 11 as `pending`.
- The Plan 11 body shows `Status: done 2026-05-13`.
- The Plan 12 body shows `Status: done 2026-05-13`.

- [ ] **Step 2: Update roadmap with the minimal correction**

Change only these facts:

```markdown
> **Last verified:** 2026-05-13 (Plan 12 prep correction)
```

```markdown
| P1 | 11 | Editor frontend stabilization (parallel to Plan 4) | 8 commits | done 2026-05-13 |
```

```markdown
- **Status:** open 2026-05-13. Roadmap correction: this Plan was previously stamped `done` with unrelated commit evidence; implementation has not started. Execution guide: `docs/superpowers/specs/2026-05-13-plan-12-screens.md`.
```

- [ ] **Step 3: Verify the roadmap diff**

Run:

```powershell
git diff -- wiki/backlog/roadmap.md
```

Expected:

- Only `wiki/backlog/roadmap.md` changes.
- Plan 11 table status matches the Plan 11 body.
- Plan 12 body is open/not started and references the Plan 12 spec.

## Task 2: Verify And Commit Prep Slice

**Files:**
- Modify: `docs/superpowers/plans/2026-05-13-plan-12-prep-roadmap.md`
- Modify: `wiki/backlog/roadmap.md`
- Verify: Git status and Markdown text checks.

- [ ] **Step 1: Check for incomplete placeholder markers in the new plan**

Run:

```powershell
rg -n "T[B]D|fill[ ]in|implement[ ]later" docs/superpowers/plans/2026-05-13-plan-12-prep-roadmap.md
```

Expected:

- Exit code `1`, no matches.

- [ ] **Step 2: Check final working tree**

Run:

```powershell
git status --short
```

Expected:

- `docs/superpowers/plans/2026-05-13-plan-12-prep-roadmap.md`
- `wiki/backlog/roadmap.md`

- [ ] **Step 3: Commit the prep slice**

Run:

```powershell
git add -- docs/superpowers/plans/2026-05-13-plan-12-prep-roadmap.md wiki/backlog/roadmap.md
git commit -m "docs: correct plan 12 roadmap status"
```

Expected:

- Commit succeeds with both documentation files included.
