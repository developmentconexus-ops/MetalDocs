# Refactor Backlog — audit

> Actionable rows. One row = one PR. Pulled from `wiki/modules/audit-tech-debt.md`.

**Last verified:** 2026-06-21 (verify-and-archive sweep — solved rows pruned; see _cleanup-2026-06-21.md)

## Rows

| id | title | debt_id | effort | impact | blocked_by | owner | status | pr |
|---|---|---|---|---|---|---|---|---|
| R-003 | Add retention policy: monthly partition + pg_cron purge job + ADR | T-003 | L | major | — | — | merged (goroutine only; pg_cron + partitioning deferred) | Plan 6a (2026-05-11, commit b5b077b7 + main.go) |
| R-006 | Switch event id from timestamp string to ULID/UUIDv7 generator | T-006 | S | minor | — | — | open | — |
| R-009 | Make SELECT/INSERT grants explicit in migrations; document role split | T-009 | S | minor | — | — | open | — |
| R-010 | Cap audit payload size (Postgres CHECK + app-side guard with truncation policy) | T-010 | S | minor | — | — | open | — |
| R-011 | Author ADR for append-only-by-grant + port/adapter shape + Writer-equals-Reader instance | T-011 | S | minor | — | — | open | — |
| R-012 | Add Go doc comments to all exported audit symbols | T-012 | S | minor | — | — | open | — |

## Notes

- R-003 is `L`: split into "partition migration" PR + "pg_cron purge job" PR + "retention ADR" PR before opening.
