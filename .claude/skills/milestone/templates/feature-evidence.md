# Feature <id> — Evidence

> **Milestone:** <n>  ·  **Feature:** `f<n>.x-<slug>`  ·  **Closed:** <date>
> A feature is closed only when every row below is filled with real output —
> not "done" / "green" / "looks good".

## What was implemented

Bullet the actual changes, by outcome. Link the commits (`<sha> <subject>`).

## Verification

| Check | Command / action | Result (evidence) |
|-------|------------------|-------------------|
| Static (build/lint/types) | `<cmd>` | <pass + key line> |
| Targeted test | `<cmd>` | <n passed> |
| Runtime proof (observable changes only) | <what was exercised + how> | <observed behavior / response shape> |

> Observable changes must be runtime-verified here by us — never deferred to the operator.

## Acceptance vs milestone spec

Restate this feature's "what to validate" from `milestone.md` and mark each met/not-met
with the evidence above.

| Acceptance criterion (from milestone.md) | Met? | Evidence |
|------------------------------------------|------|----------|
| <criterion> | yes/no | <link to row above> |

## Review disposition

- Spec-compliance review: <verdict + how findings were resolved>
- Code-quality review: <verdict + how findings were resolved, by root-cause family>

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| <none, or item> | | |
