# Feature <id> — Evidence

> **Milestone:** <n>  ·  **Feature:** `f<n>.x-<slug>`  ·  **Closed:** <date>
> **Contract:** `spec.md` (consumer contract + Validation Gate this proves against).
> A feature is closed only when every row below is filled with **real, honestly-labeled** output —
> not "done" / "green" / "looks good", and not a fixture passed off as the real provider.

## What was implemented

Bullet the actual changes, by outcome. Confirm the **producer matches the consumer contract** in
`spec.md` (not the reverse). Link the commits (`<sha> <subject>`).

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — failing test first, then green | `<test name / cmd>` | <red → green, key line> | real / fixture |
| Static (build/lint/types) | `<cmd>` | <pass + key line> | — |
| Targeted test | `<cmd>` | <n passed> | real / fixture |
| Runtime proof (observable changes only) | <what was exercised + how> | <observed behavior / response shape> | real / fixture |

> Observable changes must be runtime-verified here by us — never deferred to the operator.
> Fixture-only proof is labeled as such; it is not end-to-end proof of the real provider. A
> suite-level "all green" without per-criterion mapping below is **not** acceptance.

## Acceptance vs spec Validation Gate

Restate this feature's Validation Gate from `spec.md` (which traces to its `milestone.md` row) and
mark each criterion met/not-met against the evidence above.

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| <criterion> | yes/no | <link to row above> |

## Review disposition

- Spec-compliance review: <verdict + how findings were resolved>
- Code-quality review: <verdict + how findings were resolved, by root-cause family>

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| <none, or item> | | |
