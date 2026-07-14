# Refactor Backlog - controlled-documents

> Actionable rows. One row = one PR. Pulled from [wiki/modules/controlled-documents-tech-debt.md](../modules/controlled-documents-tech-debt.md).

**Last verified:** 2026-06-21 (verify-and-archive sweep — solved rows pruned; see _cleanup-2026-06-21.md)

## Rows

| id | title | debt_id | effort | impact | blocked_by | owner | status | pr |
|---|---|---|---|---|---|---|---|---|
| R-005 | Adopt `SET LOCAL metaldocs.tenant_id` GUC + RLS policies for controlled-documents-owned tables | T-005 | L | major | - | - | open | - |
| R-009 | Replace `WithDocumentInitializer` setter with constructor injection (split controlled-documents module init into two phases or move port to a shared package) | T-009 | M | minor | - | - | open | - |
| R-012 | Add Go doc comments to remaining exported controlled-documents symbols | T-012 | M | minor | - | - | open | - |

## Notes

- R-005 is `L`: split into "GUC adoption per repository method" + "RLS policy migration" + "tenant ADR" before opening.
