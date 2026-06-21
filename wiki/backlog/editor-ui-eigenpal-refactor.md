# Refactor Backlog — editor-ui-eigenpal

> Actionable rows. One row = one PR. Pulled from `wiki/modules/editor-ui-eigenpal-tech-debt.md`.

**Last verified:** 2026-06-21 (verify-and-archive sweep — solved rows pruned; see _cleanup-2026-06-21.md)

## Rows

| id | title | debt_id | effort | impact | blocked_by | owner | status | pr |
|---|---|---|---|---|---|---|---|---|
| R-007 | Write ADR for `templatePlugin` mode gating rule | T-007 | S | Minor | — | — | open | — |
| R-008 | Write ADR for wrapper-only consumption boundary (`@eigenpal/docx-js-editor` only via `@metaldocs/editor-ui`) | T-008 | S | Minor | — | — | open | — |
| R-010 | Bump eigenpal package to next fork build once upstream PR series lands | maint:dep-bump | M | Major | R-001 | — | open | — |

## Notes

- R-007 and R-008 are sibling ADR PRs; R-008 implicitly closes T-002's "missing-ADR" subclaim once the wrapper-only rule has a decision record.
