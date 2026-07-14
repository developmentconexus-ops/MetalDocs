# Refactor Backlog — auth

> Actionable rows. One row = one PR. Pulled from `wiki/modules/auth-tech-debt.md`.

**Last verified:** 2026-06-21 (verify-and-archive sweep — solved rows pruned; see _cleanup-2026-06-21.md)

## Rows

| id | title | debt_id | effort | impact | blocked_by | owner | status | pr |
|---|---|---|---|---|---|---|---|---|
| R-007 | Extract auth↔iam shared identity contract into platform package | T-007 | L | minor | — | — | open | — |
| R-008 | Add tenant_id to auth_identities with backfill (auth_sessions.tenant_id already added by Plan 3 / migration 0184) | T-008 | M | minor | — | — | partial (sessions done 2026-05-11) | — |
| R-009 | Distinguish malformed-cookie from no-session in Logout error path | T-009 | XS | minor | — | — | open | — |
| R-010 | Author ADR for session-cookie + bcrypt + lockout policy | T-010 | S | minor | — | — | open | — |
| R-011 | Add Go doc comments to all exported auth symbols | T-011 | M | minor | — | — | open | — |

## Notes

- R-007 + R-008 are `L`: split into reader-extraction PR + writer-migration PR before opening.
