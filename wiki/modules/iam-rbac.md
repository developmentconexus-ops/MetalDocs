# Module: iam-rbac

> **Last verified:** 2026-05-03
> **Scope:** Capabilities, roles, area-scoped grants, DB-backed authorization.
> **Out of scope:** Authentication mechanism (login, sessions) — see `wiki/references/local-dev-credentials.md`.
> **Key files:**
> - `internal/modules/iam/domain/port.go:6` — `RoleProvider` interface with `RolesByUserID(ctx, userID, tenantID)` (tenant scoping enforced)
> - `internal/modules/iam/domain/port.go:11` — `RoleAdminRepository` interface: `HasAnyRole`, `UpsertUserAndAssignRole`, `ReplaceUserRoles` — all include `tenantID` param
> - `internal/modules/iam/infrastructure/postgres/role_provider.go:19` — `RolesByUserID` impl; filters `iam_user_roles` by `tenant_id = $2::uuid`
> - `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:20` — `HasAnyRole` filters by `tenant_id`; `UpsertUserAndAssignRole` at :33, `ReplaceUserRoles` at :72 — both use DELETE-then-INSERT
> - `internal/modules/iam/domain/capabilities.go:3` — 16 `Cap*` string constants
> - `internal/modules/iam/domain/model.go:5` — Role constants (viewer, editor, author, approver, system_admin)
> - `internal/modules/iam/domain/role_capabilities.go:5` — in-process role→capability map (RoleCapabilitiesVersion = 2)
> - `internal/modules/iam/application/capability_service.go:12` — `CapabilityService` struct; `CanDo` at :31 — DB-backed tier-1 check
> - `internal/modules/iam/application/cached_role_provider.go:36` — `roleCacheKey(userID, tenantID)` — cache key format `userID|tenantID`
> - `internal/modules/iam/authz/authz.go:34` — `authz.Require` doc comment; func at :44 — area-scoped tier-2 check with system_admin bypass at :58
> - `internal/modules/iam/authz/context.go:13` — typed errors `ErrActorContextMissing` (:13) / `ErrTenantContextMissing` (:17)
> - `internal/modules/iam/authz/context.go:21` — `MustActorID` GUC helper (uses `missing_ok=true`); `MustTenantID` at :34
> - `internal/modules/iam/delivery/http/middleware.go:30` — `NewMiddleware` takes `*iamapp.CapabilityService`
> - `apps/api/cmd/metaldocs-api/permissions.go:12` — `newPermissionResolver` maps routes → `Cap*` constants
> - `internal/platform/tenant/const.go:4` — `DevTenantID = "ffffffff-ffff-ffff-ffff-ffffffffffff"` sentinel UUID for single-tenant dev/test mode (extracted from bootstrap/api.go by c4a7d9a9)

## Model

- **Capability**: fine-grained permission string (e.g. `doc.view`, `template.approve`). Defined in `capabilities.go`.
- **Role**: named bundle of capabilities. Five canonical roles as of migration 0166: `viewer`, `editor`, `author`, `approver`, `system_admin`. Three legacy process-area roles also exist in `user_process_areas`: `signer`, `area_admin`, `qms_admin` (capabilities backfilled by migration 0169).
- **Area grant** (`user_process_areas`): scopes a role to one or more areas. A user can be `author` in `RH` only.
- **Group**: `iam_groups` + `iam_group_members` + `iam_group_roles` — role grants can be made via group membership (added migration 0163). `CapabilityService.CanDo` checks both direct user roles and group-derived roles.

## Roles

| Role | Notes |
|---|---|
| `viewer` | Read-only access to docs and templates |
| `editor` | Can create/edit docs and view templates; registry.create included |
| `author` | Full doc authoring + template authoring/submission |
| `approver` | Doc authoring + doc/template signoff |
| `system_admin` | All capabilities; bypasses area-scoped check (see authz bypass below) |

Migration 0166 renamed `admin` → `system_admin` and `reviewer` → `approver` in `iam_user_roles`. UNIQUE(tenant_id, user_id) constraint added; one role per user per tenant.

## Capability matrix (migration 0165 — 40 rows)

| Capability | viewer | editor | author | approver | system_admin |
|---|:---:|:---:|:---:|:---:|:---:|
| `doc.view` | Y | Y | Y | Y | Y |
| `doc.create` | | Y | Y | Y | Y |
| `doc.edit` | | Y | Y | Y | Y |
| `doc.submit` | | | Y | Y | Y |
| `doc.signoff` | | | | Y | Y |
| `template.view` | Y | Y | Y | Y | Y |
| `template.create` | | | Y | | Y |
| `template.edit` | | Y | Y | | Y |
| `template.submit` | | | Y | | Y |
| `template.approve` | | | | Y | Y |
| `template.publish` | | | | | Y |
| `registry.create` | | Y | Y | | Y |
| `taxonomy.manage` | | | | | Y |
| `membership.manage` | | | | | Y |
| `route.manage` | | | | | Y |
| `user.manage` | | | | | Y |

## Auth flow

HTTP requests are gated by `Middleware.Wrap` (middleware.go:48). The `PermissionResolver` maps `(method, path)` → `(capability string, guarded bool)`. If guarded, `CapabilityService.CanDo` (capability_service.go:31) runs a single SQL EXISTS query across four branches:

1. Direct `iam_user_roles` with `role_code = 'system_admin'`
2. Group-derived `iam_group_roles` with `role = 'system_admin'`
3. Direct role → `role_capabilities` join for the required capability
4. Group-derived role → `role_capabilities` join for the required capability

Returning `allowed = true` on any branch grants access.

## authz.Require (area-scoped check)

`authz.Require` (authz.go:44) is the transactional authz used inside approval/signoff handlers. It checks:

1. `system_admin` bypass — queries `iam_user_roles` for `role_code = 'system_admin'` before any capability check (authz.go:59). If true, grants immediately.
2. Falls through to `user_process_areas` JOIN `role_capabilities` scoped by `area_code`.

The bypass ensures `system_admin` users can act on any area without needing a `user_process_areas` row.

## Schema changes (migrations 0162–0169)

| Migration | Change |
|---|---|
| 0162 | `tenant_id UUID` added to `metaldocs.iam_user_roles` (default `ffffffff-...`) |
| 0163 | `iam_groups`, `iam_group_members`, `iam_group_roles` tables created |
| 0164 | `visibility TEXT DEFAULT 'area'` added to `public.documents` (values: public/area/restricted; was mistakenly added to the now-dropped `public.documents_v2` — corrected by migration 0167) |
| 0165 | `role_capabilities` truncated and reseeded with 40 rows covering the 5 canonical roles |
| 0166 | `admin` → `system_admin`, `reviewer` → `approver` renamed; UNIQUE(tenant_id, user_id) added |
| 0169 | `role_capabilities` backfill for `signer`, `area_admin`, `qms_admin` (process-area roles left with zero caps by 0165's TRUNCATE); idempotent via `ON CONFLICT DO NOTHING` |
| 0170 | Dev `approver` seed user role corrected back to `approver` (file: `migrations/0170_dev_approver_role_correction.sql`); 0166 blanket rename had incorrectly promoted it to `system_admin` |

## StaticAuthorizer removed

`internal/modules/iam/application/authorizer.go` and its test were deleted. `CapabilityService` (DB-backed) is the sole authorizer. `NewMiddleware` now takes `*iamapp.CapabilityService` instead of `iamdomain.Authorizer`.

## ISO segregation overlay

Independently of capabilities, the approval module enforces that the submitter of a document cannot signoff on it. See [concepts/iso-segregation.md](../concepts/iso-segregation.md).

## See also

- [concepts/authz-tiers.md](../concepts/authz-tiers.md) — two-tier authz model quick reference
- [decisions/0007-two-tier-authz.md](../decisions/0007-two-tier-authz.md) — ADR rationale
- [vision/target-users.md](../vision/target-users.md)
- [concepts/iso-segregation.md](../concepts/iso-segregation.md)
- [modules/approval.md](approval.md)
- [references/local-dev-credentials.md](../references/local-dev-credentials.md)
