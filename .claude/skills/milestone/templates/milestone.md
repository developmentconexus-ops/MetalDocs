# Milestone <n> — <Title>

> **Program:** <program-slug>  ·  **Governing spec:** `<path>`
> **Status:** Spec (drafting) | Spec approved | Executing | Validating | Passed | Blocked
> **Authored:** <date> — *before any feature in this milestone began.*

> This file is a **spec**, authored up front. It says **what** this milestone is,
> **which features** it contains, **what each feature implements**, and **what gets
> validated**. It contains **no execution steps** — the "how" of each feature lives in
> that feature's `plan.md`. The end-of-milestone QA (`qa/milestone-qa.md`) validates
> the milestone against *this* document.

## Objective

What this milestone delivers and why it is a coherent slice. If it advances a quality
bar (a grade, a closed defect class, a contract cleaned up), name the bar and the
exact criterion that proves it moved.

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F<n>.1 | `f<n>.1-<slug>` | <the change, by outcome — not steps> | <the observable/testable acceptance criteria> |
| F<n>.2 | `f<n>.2-<slug>` | | |
| … | | | |

For each feature, "what to validate" must be **objectively checkable** — a test that
passes, a route that responds with the contracted shape, a build that is clean, a
runtime behavior observed. Avoid "works" / "looks right".

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers — it judges and
writes `qa/milestone-qa.md`; the main session flips status only on its PASS), following the binding
C1–C7 checklist in `.claude/skills/milestone/references/milestone-end-validation.md`. What that gate
enforces for this milestone:

1. **Per-feature acceptance** — every feature above meets its declared "what to validate", and each
   feature's **consumer contract** (`spec.md`) was honored (producer matches consumer).
2. **Workflow-class QA checklist** — <name the canonical checklist(s) for this milestone's
   work, e.g. backend-api / screen / workflow-async / release close-out>.
3. **Regression** — previously-completed milestones still pass their gates.
4. **Quality-bar / root-cause check** (if applicable) — <the bar> is re-measured and the
   root cause is confirmed **fixed, not symptom-patched**.
5. **No unplanned scope** — anything implemented beyond this list is recorded with rationale.

## Dependencies & constraints

- Depends on: <prior milestones / external prerequisites>.
- Architectural constraints respected: <e.g. no migrations, reads stay live, advisory-lock
  hazard rules, contract-first regen order>.

## Applicable hard-stops

List the hard-stops in force for this milestone (default catalog HS-1..HS-6; add any
program-specific ones). State explicitly what would trip each one here.
