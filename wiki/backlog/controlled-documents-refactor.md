# Refactor Backlog - controlled-documents

> Actionable rows. One row = one PR. Pulled from [wiki/modules/controlled-documents-tech-debt.md](../modules/controlled-documents-tech-debt.md).

**Last verified:** 2026-05-21 (backend platform freeze)

## Rows

| id | title | debt_id | effort | impact | blocked_by | owner | status | pr |
|---|---|---|---|---|---|---|---|---|
| R-001 | Verify + wire capability gates for Obsolete/Supersede (resolver + service-side guard) | T-001 | M | critical | - | - | merged | Plan 5 (2026-05-11) |
| R-002 | Emit governance event from `changeStatus` (active->obsolete, active->superseded) | T-002 | S | critical | - | - | merged | Plan 6a (2026-05-11, commit 5bb06964) |
| R-003 | Migrate controlled-documents error responses to RFC 9457 problem+json | T-003 | M | major | - | - | merged | Plan 7 (2026-05-11) |
| R-004 | Apply tier-3 tripwire to `controlled_documents` + `cd_sequence_counters`; pair `authz.Require` with each mutator | T-004 | L | major | R-001 | - | merged | Plan 5 (2026-05-11) |
| R-005 | Adopt `SET LOCAL metaldocs.tenant_id` GUC + RLS policies for controlled-documents-owned tables | T-005 | L | major | - | - | open | - |
| R-006 | Authz for `GetActiveDocument` (tenant header-source sub-issue resolved by Plan 3; remaining: add read-policy authz check) | T-006 | M | major | - | - | open | - |
| R-007 | Implement 422 `template_invalid` mapping OR drop spec branch | T-007 | S | major | - | - | merged | Plan 7 (2026-05-11, commit 395b0b24) |
| R-008 | Move controlled-documents audit emission to platform-owned `internal/audit` writer | T-008 | M | major | audit#R-001 | - | merged | Plan 6a (2026-05-11, commit 71a2dc53) |
| R-009 | Replace `WithDocumentInitializer` setter with constructor injection (split controlled-documents module init into two phases or move port to a shared package) | T-009 | M | minor | - | - | open | - |
| R-010 | Expose controlled-documents repository (or a read-only port) via `Module` so external wiring stops reaching into `infrastructure` | T-010 | S | minor | - | - | open | - |
| R-011 | Restructure OpenAPI tree under canonical controlled-documents paths only | T-011 | S | minor | - | - | open | - |
| R-012 | Add Go doc comments to remaining exported controlled-documents symbols | T-012 | M | minor | - | - | open | - |
| R-100 | Audit + remove residual references to legacy `profile_sequence_counters` (tests, fixtures, seed scripts) | maint:migration-cleanup | S | minor | - | - | verified closed | Plan 10 verification: runtime grep clean; table dropped by migration 0182 |

## Notes

- R-001 + R-002 are critical-path: both protect the QMS regulated lifecycle. Open before any other controlled-documents refactor.
- R-004 is blocked by R-001: wire authz capabilities first, then enforce at tier-3.
- R-005 is `L`: split into "GUC adoption per repository method" + "RLS policy migration" + "tenant ADR" before opening.
- R-008 is blocked by the audit module's `internal/audit` port adoption (see [wiki/modules/audit-tech-debt.md](../modules/audit-tech-debt.md) T-007 and [wiki/backlog/audit-refactor.md](audit-refactor.md)).
- R-100 verifies migration 0182 fully retired the legacy `profile_sequence_counters` table: code grep + test grep + fixture grep.
