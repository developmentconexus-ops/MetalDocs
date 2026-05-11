# Refactor Backlog — iam

> One row = one PR. Pulled from [`wiki/modules/iam-tech-debt.md`](../modules/iam-tech-debt.md). Rows without a `debt_id` are blocked from grooming.

**Last verified:** 2026-05-11

## Schema

| Field | Required | Notes |
|---|---|---|
| `id` | yes | `R-001`, `R-002`, … |
| `title` | yes | imperative, scoped to a single PR |
| `debt_id` | yes | `T-NNN` from tech-debt register |
| `effort` | yes | XS · S · M · L (≥L = split first) |
| `impact` | yes | Critical · Major · Minor (mirror debt severity) |
| `blocked_by` | optional | other row id or external ticket |
| `owner` | optional | github handle |
| `status` | yes | open · in-progress · merged · cancelled |
| `pr` | optional | PR URL once opened |

## Rows

| id | title | debt_id | effort | impact | blocked_by | owner | status | pr |
|---|---|---|---|---|---|---|---|---|
| R-001 | Collapse capability namespaces to a single `Capability` typed enum sourced from `domain/capabilities.go`; remove the `Cap*` typed consts in `domain/model.go` | T-001 | M | Critical | — | — | open | — |
| R-002 | Pick one area-membership write path: route the v2 HTTP service through `area_membership/` (SECURITY DEFINER) OR delete `area_membership/` and centralise on `UserAreaRepository.GrantAtomic` with explicit governance-event INSERT | T-002 | M | Major | R-007 | — | open | — |
| R-003 | Wire `AuthorizationService` into the consumer that needs SoD (documents/approval) or delete it; resolve `ErrSoDViolation` ownership (IAM vs approval) | T-003 | M | Major | — | — | open | — |
| R-004 | Attach `enforce_capability_asserted` trigger to IAM-owned mutating tables (`iam_user_roles`, `user_process_areas`, `iam_users`) — new migration; first wire `authz.Require` in the corresponding repo methods | T-004 | L | Major | R-001 | — | open | — |
| R-005 | Emit `auditdomain.Writer.Record` from `handleUserRoleUpsert` (and any other admin op missing audit) — match the call pattern in handlers that already do | T-005 | S | Critical | — | — | open | — |
| R-006 | Migrate IAM error envelopes (middleware + membership handler) to RFC 9457 Problem+JSON; add `metaldocs.authz.forbidden` and `metaldocs.iam.*` Problem `type` URIs | T-006 | M | Major | — | — | open | — |
| R-007 | Implement and wire a real `MembershipGovernanceLogger` (Postgres `governance_events` writer) at `main.go:217` | T-007 | S | Major | — | — | open | — |
| R-008 | Add `CachedRoleProvider.InvalidateGroup(groupID)` + wire it into the (future) group-write site; cover with unit test | T-008 | S | Minor | (deferred until group writes exist) | — | open | — |
| R-009 | Rename one of the `ErrCapabilityDenied` symbols (e.g. `authz.ErrCapDenied` retains struct shape; `iamapp.ErrCapabilityDenied` stays sentinel) | T-009 | XS | Minor | — | — | open | — |
| R-010 | Move `iamdomain.Role` to a neutral package (e.g. `internal/platform/iam-types`) so neither IAM nor auth depends on the other for the enum | T-010 | M | Minor | — | — | open | — |
| R-011 | Author ADR for "tenant-scoping rule" — every IAM table has `tenant_id`, every repo method filters by it; cite Group B fix as origin | T-011 | XS | Minor | — | — | open | — |
| R-012 | Add a CI check that compares the in-process `RoleCapabilities` map against the seeded DB rows (not just count) | T-012 | M | Minor | R-001 | — | open | — |
| R-013 | Add Go doc comments to the public surface in IAM (cap-prefix consts, `domain/port.go` interfaces, `delivery/http/admin_handler.go` request structs) — bulk-doc PR | maint:doc-cleanup | M | Minor | — | — | open | — |
| R-014 | Retire `wiki/modules/iam-rbac.md` (predecessor stub) — file deleted; inbound links repointed to `wiki/modules/iam.md`; README index updated | maint:doc-cleanup | XS | Minor | — | — | merged | (in-commit doc-only) |

## Notes

- `L`-effort rows: R-004 splits into (a) wire `authz.Require` in IAM repos, (b) write the trigger-attach migration, (c) update consumers / contract tests. Open as a tracking issue first.
- R-001 unblocks R-004 and R-012; pull from the top.
- R-014 was the only doc-only row; merged in the publish follow-up commit after user confirmation.
- When a row merges: bump `status` to `merged`, link PR, remove the linked `T-NNN` row from the register in the same commit.
