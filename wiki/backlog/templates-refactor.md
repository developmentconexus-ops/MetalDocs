# Refactor Backlog — templates

> Actionable rows. One row = one PR. Pulled from `wiki/modules/templates-tech-debt.md`. Rows that lack a debt-id are blocked from grooming.

**Last verified:** 2026-06-21 (verify-and-archive sweep — solved rows pruned; see _cleanup-2026-06-21.md)

## Rows

| id | title | debt_id | effort | impact | blocked_by | owner | status | pr |
|---|---|---|---|---|---|---|---|---|
| R-007 | Introduce `Repository.WithTx` and wrap `PublishTemplateVersion` / `Approve` / `CreateTemplate` in a single `pgx.Tx`; emit `AuditObsoleted` for the obsolete side-effect | T-007 | M | Major | — | — | open | — |
| R-009 | Verify `internal/platform/idempotency` replay semantics on generated POST routes (`/templates`, `/publish`, `/submit`, `/review`, `/approve`) and classify remaining POST mutation surfaces | T-009 | M | Major | - | - | open (partial wrapper exists) | Plan 12.4 verified generated create path requires/sends `Idempotency-Key` and receives HTTP 201; same-key replay audit still pending |
| R-011 | Add cursor pagination (Plan 2 cursor primitive) to `ListTemplates`; default page size 50 | T-011 | S | Minor | — | — | open | — |
| R-014 | Add Go doc comments to every exported symbol under `internal/modules/templates/{domain,application,delivery,repository}/` | T-014 | S | Minor | — | — | open | — |
| R-100 | Retire predecessor frontend-heavy stub `wiki/modules/templates.md` (kebab) and repoint inbound links to `wiki/modules/templates.md` | maint:doc-cleanup | XS | Minor | — | — | open | — |
| R-101 | Correct/retire this row — original rename intent lost (source==target); re-derive or close | maint:migration-cleanup | M | Minor | R-006 | — | open | — |

## Notes

- 2026-05-17 product/API note: creator-scoped template-use `visibility`, `areas`, and `specific_areas` were removed from runtime/API selection behavior. The database columns remain inert compatibility fields until a coordinated baseline/reference-data cleanup is planned.
- R-101 deferred until `/api/v1/` flip + dir rename can be done atomically (touches frontend `lib/api-types/`); Plan 12.4 removed the R-006 contract coverage blocker.
