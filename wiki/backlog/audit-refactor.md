# Refactor Backlog — audit

> Actionable rows. One row = one PR. Pulled from `wiki/modules/audit-tech-debt.md`.

**Last verified:** 2026-05-12 (Plan 7)

## Rows

| id | title | debt_id | effort | impact | blocked_by | owner | status | pr |
|---|---|---|---|---|---|---|---|---|
| R-001 | Gate GET /api/v1/audit/events behind capability check (audit.read) | T-001 | S | critical | — | — | merged | Plan 6a (2026-05-11, commits 0279546f + 6b34c277) |
| R-002 | Migrate audit error envelope to RFC 9457 problem+json | T-002 | S | major | — | — | merged | Plan 7 (2026-05-11, commit 2ca727d6) |
| R-003 | Add retention policy: monthly partition + pg_cron purge job + ADR | T-003 | L | major | — | — | merged (goroutine only; pg_cron + partitioning deferred) | Plan 6a (2026-05-11, commit b5b077b7 + main.go) |
| R-004 | Add tamper-evidence: row-hash chain (prev_hash, row_hash) + integrity-validate job | T-004 | L | critical | R-005 | — | open | — |
| R-005 | Surface Record errors: emit metric + structured log + optional outbox for retry | T-005 | M | major | — | — | merged | Plan 6a (2026-05-11, commit 1994bb84) |
| R-006 | Switch event id from timestamp string to ULID/UUIDv7 generator | T-006 | S | minor | — | — | open | — |
| R-007 | Add tenant_id column + ListEvents tenant filter; backfill via auth tenant cutover | T-007 | L | major | auth#R-008 | — | merged | Plan 6a (2026-05-11, commit b5b077b7 + migration 0190) |
| R-008 | Add operationId to /audit/events and re-mount via oapi-codegen | T-008 | M | minor | — | — | open | — |
| R-009 | Make SELECT/INSERT grants explicit in migrations; document role split | T-009 | S | minor | — | — | open | — |
| R-010 | Cap audit payload size (Postgres CHECK + app-side guard with truncation policy) | T-010 | S | minor | — | — | open | — |
| R-011 | Author ADR for append-only-by-grant + port/adapter shape + Writer-equals-Reader instance | T-011 | S | minor | — | — | open | — |
| R-012 | Add Go doc comments to all exported audit symbols | T-012 | S | minor | — | — | open | — |

## Notes

- R-001 is critical-path: open before any other audit work.
- R-004 depends on R-005 — surfacing errors is a prerequisite for an integrity-validate job to alarm meaningfully.
- R-007 is blocked by `wiki/backlog/auth-refactor.md#R-008` (tenant_id on auth tables) — multi-tenant must land coherently across audit + auth + iam.
- R-003 is `L`: split into "partition migration" PR + "pg_cron purge job" PR + "retention ADR" PR before opening.
