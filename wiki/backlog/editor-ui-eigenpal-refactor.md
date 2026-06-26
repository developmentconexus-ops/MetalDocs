# Refactor Backlog — editor-ui-eigenpal

> Actionable rows. One row = one PR. Pulled from `wiki/modules/editor-ui-eigenpal-tech-debt.md`.

**Last verified:** 2026-06-23 (eigenpal migration: R-010 reframed as npm version bump; R-008 package name updated)

## Rows

| id | title | debt_id | effort | impact | blocked_by | owner | status | pr |
|---|---|---|---|---|---|---|---|---|
| R-007 | Write ADR for `templatePlugin` mode gating rule | T-007 | S | Minor | — | — | open | — |
| R-008 | Write ADR for wrapper-only consumption boundary (`@eigenpal/docx-editor-react` only via `@metaldocs/editor-ui`) | T-008 | S | Minor | — | — | open | — |
| R-010 | Bump `@eigenpal/docx-editor-react` to next published release when available | maint:dep-bump | S | Minor | — | — | open | — |

## Notes

- R-007 and R-008 are sibling ADR PRs; R-008 implicitly closes T-002's "missing-ADR" subclaim once the wrapper-only rule has a decision record.
- R-010 unblocked by eigenpal v1.9.0 adoption (2026-06-23): no tarball to replace, simply bump the version pin in `package.json`.
