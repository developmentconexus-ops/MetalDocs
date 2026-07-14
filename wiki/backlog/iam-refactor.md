# Refactor Backlog — iam

> One row = one PR. Pulled from [`wiki/modules/iam-tech-debt.md`](../modules/iam-tech-debt.md). Rows without a `debt_id` are blocked from grooming.

**Last verified:** 2026-06-21 (verify-and-archive sweep — solved rows pruned; see _cleanup-2026-06-21.md)

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
| R-004 | Attach `enforce_capability_asserted` trigger to IAM-owned mutating tables (`iam_user_roles`, `user_process_areas`, `iam_users`) — new migration; first wire `authz.Require` in the corresponding repo methods | T-004 | L | Major | R-001 | — | merged (partial) | Plan 5 (2026-05-11): `iam_user_roles` + `user_process_areas` done; `iam_users` INSERT residual |
| R-008 | Add `CachedRoleProvider.InvalidateGroup(groupID)` + wire it into the (future) group-write site; cover with unit test | T-008 | S | Minor | (deferred until group writes exist) | — | open | — |
| R-010 | Move `iamdomain.Role` to a neutral package (e.g. `internal/platform/iam-types`) so neither IAM nor auth depends on the other for the enum | T-010 | M | Minor | — | — | merged | ARC-06 direct commit 2026-07-02 (no PR); `internal/platform/iamtypes` new pkg, `iam/domain.Role` aliased, auth imports `iamtypes` not `iam/domain` for Role — see `wiki/modules/iam-tech-debt.md` T-010 |
| R-011 | Author ADR for "tenant-scoping rule" — every IAM table has `tenant_id`, every repo method filters by it; cite Group B fix as origin | T-011 | XS | Minor | — | — | open | — |
| R-013 | Add Go doc comments to the public surface in IAM (cap-prefix consts, `domain/port.go` interfaces, `delivery/http/admin_handler.go` request structs) — bulk-doc PR | maint:doc-cleanup | M | Minor | — | — | open | — |

## Notes

- When a row merges: bump `status` to `merged`, link PR, remove the linked `T-NNN` row from the register in the same commit.
