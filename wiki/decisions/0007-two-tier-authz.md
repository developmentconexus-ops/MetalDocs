# ADR 0007 — Two-Tier Authorization

> **Status:** accepted 2026-05-03
> **Last verified:** 2026-05-03
> **Scope:** Authorization boundary between HTTP middleware (tier 1) and in-transaction area checks (tier 2).
> **Out of scope:** Authentication; Role/capability table definitions — see `wiki/modules/iam-rbac.md`.
> **Key files:**
> - `internal/modules/iam/application/capability_service.go:31` — tier-1 `CanDo` implementation
> - `internal/modules/iam/authz/authz.go:44` — tier-2 `Require` implementation; system_admin bypass at :58
> - `internal/modules/iam/authz/context.go:13` — typed errors `ErrActorContextMissing` / `ErrTenantContextMissing`; GUC helpers `MustActorID` at :21, `MustTenantID` at :34
> - `internal/modules/iam/infrastructure/postgres/role_provider.go:19` — `RolesByUserID` filters by `tenant_id` (Group B B5/B6 fix)
> - `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:20` — `HasAnyRole`, `UpsertUserAndAssignRole`, `ReplaceUserRoles` all tenant-scoped (Group B fix)
> - `internal/platform/tenant/const.go:4` — `DevTenantID = "ffffffff-ffff-ffff-ffff-ffffffffffff"` sentinel UUID for single-tenant dev/test mode (extracted from bootstrap/api.go by c4a7d9a9)

## Context

MetalDocs has two authz services that each consult `role_capabilities`:

1. `CapabilityService.CanDo(ctx, userID, tenantID, capability)` — reads `iam_user_roles`. Used by HTTP middleware to gate route access.
2. `authz.Require(ctx, tx, capability, areaCode)` — reads `user_process_areas`. Used inside service-layer transactions to gate area-scoped writes (signoff, area-bound state changes).

The 2026-05-02 IAM unification plan attempted to consolidate both. It unified the middleware path (`StaticAuthorizer` → `CapabilityService`) but left `authz.Require` reading `user_process_areas`. This produced apparent dual systems and confused engineers.

## Decision

Treat the two services as **distinct tiers** with explicit responsibilities, not as a unification gap.

| Tier | Service | Table | Used by | Question answered |
|---|---|---|---|---|
| 1 — Tenant | `CapabilityService.CanDo` | `iam_user_roles` JOIN `role_capabilities` | HTTP middleware | "Can user X do `doc.create` in tenant T?" |
| 2 — Area | `authz.Require` | `user_process_areas` JOIN `role_capabilities` | Service layer (in-tx) | "Can user X sign for area QA-01?" |

`role_capabilities` is shared: both tiers map a role code to a capability set. Roles in `iam_user_roles` are tenant-scoped; roles in `user_process_areas` are area-scoped.

`system_admin` bypasses both tiers.

`authz.Require` accepts `areaCode = "tenant"` as a degenerate form: skips the area filter, behaves like a tier-1 check inside the transaction.

## Consequences

- **Positive:** matches the QMS domain (ISO 9001 segregation requires per-area approver grants); clear contract; no schema migration needed.
- **Negative:** IAM admin UI must distinguish "tenant role" assignment from "area membership" assignment — separate flows.
- **Open:** future area-membership service consolidation if area grants diverge from `user_process_areas`.

## References

- `internal/modules/iam/application/capability_service.go` — tier 1
- `internal/modules/iam/authz/authz.go` — tier 2
- `internal/modules/iam/authz/context.go` — typed GUC errors (Group B fix)
- `internal/modules/iam/infrastructure/postgres/role_provider.go` — tenant-scoped `RolesByUserID` (Group B fix)
- `internal/modules/iam/infrastructure/postgres/role_admin_repository.go` — tenant-scoped admin ops, DELETE-then-INSERT (Group B fix)
- Migration 0162 — added `tenant_id` to `iam_user_roles`
- Migration 0165 — reseeded `role_capabilities`
- Migration 0170 (`migrations/0170_dev_approver_role_correction.sql`) — corrects dev approver role after 0166 over-promotion
- Audit 2026-05-03 — bugs B1-B6 (`wiki/bugs/audit-2026-05-03.md` lines 111-136)
- Tests: `internal/modules/iam/infrastructure/postgres/role_admin_repository_test.go`, `tests/integration/iam/tenant_isolation_test.go`
