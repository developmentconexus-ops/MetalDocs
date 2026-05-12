# Refactor Backlog — iam

> One row = one PR. Pulled from [`wiki/modules/iam-tech-debt.md`](../modules/iam-tech-debt.md). Rows without a `debt_id` are blocked from grooming.

**Last verified:** 2026-05-12 (Plan 7)

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
| R-001 | Collapse capability namespaces to a single typed `iamdomain.Capability` enum (winner per `roadmap.md:17`); delete `domain/capabilities.go` and reseed `role_capabilities` to `document.*` | T-001 | M | Critical | — | — | merged | Plan 4 (2026-05-11, commit 3a227642) |
| R-002 | Pick one area-membership write path: delete `area_membership/` Go wrapper, centralise on `UserAreaRepository.GrantAtomic`; SECURITY DEFINER SQL funcs stay for e2e seed / integration tests | T-002 | M | Major | R-007 | — | merged | Plan 4 (2026-05-11, commit a66a8d62) |
| R-003 | Delete unwired `AuthorizationService` (third authz surface); Plan 5 wires tier-2 `authz.Require` per module instead | T-003 | M | Major | — | — | merged | Plan 4 (2026-05-11, commit 8da32dbf) |
| R-004 | Attach `enforce_capability_asserted` trigger to IAM-owned mutating tables (`iam_user_roles`, `user_process_areas`, `iam_users`) — new migration; first wire `authz.Require` in the corresponding repo methods | T-004 | L | Major | R-001 | — | merged (partial) | Plan 5 (2026-05-11): `iam_user_roles` + `user_process_areas` done; `iam_users` INSERT residual |
| R-005 | Emit `auditdomain.Writer.Record` from `handleUserRoleUpsert` (and any other admin op missing audit) — match the call pattern in handlers that already do | T-005 | S | Critical | — | — | merged | Plan 6a (2026-05-11, commit f27529e8) |
| R-006 | Migrate IAM error envelopes (middleware + membership handler) to RFC 9457 Problem+JSON; add `metaldocs.authz.forbidden` and `metaldocs.iam.*` Problem `type` URIs | T-006 | M | Major | — | — | merged | Plan 7 (2026-05-11, commit 1ecfe674) |
| R-007 | Implement and wire a real `MembershipGovernanceLogger` (Postgres `governance_events` writer) at `main.go:217` | T-007 | S | Major | — | — | open | — |
| R-008 | Add `CachedRoleProvider.InvalidateGroup(groupID)` + wire it into the (future) group-write site; cover with unit test | T-008 | S | Minor | (deferred until group writes exist) | — | open | — |
| R-009 | Rename `authz.ErrCapabilityDenied` → `authz.ErrCapDenied` (struct kept); `iamapp.ErrCapabilityDenied` sentinel preserved | T-009 | XS | Minor | — | — | merged | Plan 4 (2026-05-11, commit ec7d151a) |
| R-010 | Move `iamdomain.Role` to a neutral package (e.g. `internal/platform/iam-types`) so neither IAM nor auth depends on the other for the enum | T-010 | M | Minor | — | — | open | — |
| R-011 | Author ADR for "tenant-scoping rule" — every IAM table has `tenant_id`, every repo method filters by it; cite Group B fix as origin | T-011 | XS | Minor | — | — | open | — |
| R-012 | Delete in-process `RoleCapabilities` map + `RoleCapabilitiesVersion` + `CheckRoleCapabilitiesVersion`; DB `role_capabilities` is single source of truth | T-012 | M | Minor | R-001 | — | merged | Plan 4 (2026-05-11, commit 0cd2e75d) |
| R-013 | Add Go doc comments to the public surface in IAM (cap-prefix consts, `domain/port.go` interfaces, `delivery/http/admin_handler.go` request structs) — bulk-doc PR | maint:doc-cleanup | M | Minor | — | — | open | — |
| R-014 | Retire `wiki/modules/iam-rbac.md` (predecessor stub) — file deleted; inbound links repointed to `wiki/modules/iam.md`; README index updated | maint:doc-cleanup | XS | Minor | — | — | merged | (in-commit doc-only) |

## Notes

- `L`-effort rows: R-004 splits into (a) wire `authz.Require` in IAM repos, (b) write the trigger-attach migration, (c) update consumers / contract tests. Open as a tracking issue first.
- R-001 unblocks R-004 and R-012; pull from the top.
- R-014 was the only doc-only row; merged in the publish follow-up commit after user confirmation.
- When a row merges: bump `status` to `merged`, link PR, remove the linked `T-NNN` row from the register in the same commit.
