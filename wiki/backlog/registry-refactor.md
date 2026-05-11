# Refactor Backlog — registry

> Actionable rows. One row = one PR. Pulled from [wiki/modules/registry-tech-debt.md](../modules/registry-tech-debt.md).

**Last verified:** 2026-05-11

## Rows

| id | title | debt_id | effort | impact | blocked_by | owner | status | pr |
|---|---|---|---|---|---|---|---|---|
| R-001 | Verify + wire capability gates for Obsolete/Supersede (resolver + service-side guard) | T-001 | M | critical | — | — | open | — |
| R-002 | Emit governance event from `changeStatus` (active→obsolete, active→superseded) | T-002 | S | critical | — | — | open | — |
| R-003 | Migrate registry error responses to RFC 9457 problem+json | T-003 | M | major | — | — | open | — |
| R-004 | Apply tier-3 tripwire to `controlled_documents` + `cd_sequence_counters`; pair `authz.Require` with each mutator | T-004 | L | major | R-001 | — | open | — |
| R-005 | Adopt `SET LOCAL metaldocs.tenant_id` GUC + RLS policies for registry-owned tables | T-005 | L | major | — | — | open | — |
| R-006 | Authz for `GetActiveDocument` (tenant header-source sub-issue resolved by Plan 3; remaining: add read-policy authz check) | T-006 | M | major | — | — | open | — |
| R-007 | Implement 422 `template_invalid` mapping OR drop spec branch | T-007 | S | major | — | — | open | — |
| R-008 | Move registry audit emission to platform-owned `internal/audit` writer | T-008 | M | major | audit#R-001 | — | open | — |
| R-009 | Replace `WithDocumentInitializer` setter with constructor injection (split registry module init into two phases or move port to a shared package) | T-009 | M | minor | — | — | open | — |
| R-010 | Expose registry repository (or a read-only port) via `Module` so external wiring stops reaching into `infrastructure` | T-010 | S | minor | — | — | open | — |
| R-011 | Restructure OpenAPI tree under `api/openapi/v2/partials/registry.yaml` (or move routes back to `/api/v1/`) | T-011 | S | minor | — | — | open | — |
| R-012 | Add Go doc comments to remaining 79 exported registry symbols | T-012 | M | minor | — | — | open | — |
| R-100 | Audit + remove residual references to legacy `profile_sequence_counters` (tests, fixtures, seed scripts) | maint:migration-cleanup | S | minor | — | — | open | — |

## Notes

- R-001 + R-002 are critical-path: both protect the QMS regulated lifecycle. Open before any other registry refactor.
- R-004 is blocked by R-001 — wire authz capabilities first, then enforce at tier-3.
- R-005 is `L`: split into "GUC adoption per repository method" + "RLS policy migration" + "tenant ADR" before opening.
- R-008 is blocked by audit module's `internal/audit` port adoption (see `wiki/modules/audit-tech-debt.md` T-007 / `wiki/backlog/audit-refactor.md`).
- R-100 (maint:migration-cleanup) verifies migration 0182 fully retired the legacy `profile_sequence_counters` table — code grep + test grep + fixture grep.
