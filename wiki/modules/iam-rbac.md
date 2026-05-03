# Module: iam-rbac

> **Last verified:** 2026-05-02
> **Scope:** Capabilities, roles, area-scoped grants, DB-backed authorization.
> **Out of scope:** Authentication mechanism (login, sessions) — see `wiki/references/local-dev-credentials.md`.
> **Key files:**
> - `internal/modules/iam/domain/capabilities.go:3` — 16 `Cap*` string constants
> - `internal/modules/iam/domain/model.go:5` — Role constants (viewer, editor, author, approver, system_admin)
> - `internal/modules/iam/domain/role_capabilities.go:5` — in-process role→capability map (RoleCapabilitiesVersion = 2)
> - `internal/modules/iam/application/capability_service.go:12` — `CapabilityService.CanDo` — DB-backed check
> - `internal/modules/iam/authz/authz.go:34` — `authz.Require` — area-scoped check with system_admin bypass
> - `internal/modules/iam/delivery/http/middleware.go:30` — `NewMiddleware` takes `*iamapp.CapabilityService`
> - `apps/api/cmd/metaldocs-api/permissions.go:12` — `newPermissionResolver` maps routes → `Cap*` constants

## Model

- **Capability**: fine-grained permission string (e.g. `doc.view`, `template.approve`). Defined in `capabilities.go`.
- **Role**: named bundle of capabilities. Five roles as of migration 0166: `viewer`, `editor`, `author`, `approver`, `system_admin`.
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

HTTP requests are gated by `Middleware.Wrap` (middleware.go:48). The `PermissionResolver` maps `(method, path)` → `(capability string, guarded bool)`. If guarded, `CapabilityService.CanDo` (capability_service.go:20) runs a single SQL EXISTS query across four branches:

1. Direct `iam_user_roles` with `role_code = 'system_admin'`
2. Group-derived `iam_group_roles` with `role = 'system_admin'`
3. Direct role → `role_capabilities` join for the required capability
4. Group-derived role → `role_capabilities` join for the required capability

Returning `allowed = true` on any branch grants access.

## authz.Require (area-scoped check)

`authz.Require` (authz.go:34) is the transactional authz used inside approval/signoff handlers. It checks:

1. `system_admin` bypass — queries `iam_user_roles` for `role_code = 'system_admin'` before any capability check (authz.go:40). If true, grants immediately.
2. Falls through to `user_process_areas` JOIN `role_capabilities` scoped by `area_code`.

The bypass ensures `system_admin` users can act on any area without needing a `user_process_areas` row.

## Schema changes (migrations 0162–0166)

| Migration | Change |
|---|---|
| 0162 | `tenant_id UUID` added to `metaldocs.iam_user_roles` (default `ffffffff-...`) |
| 0163 | `iam_groups`, `iam_group_members`, `iam_group_roles` tables created |
| 0164 | `visibility TEXT DEFAULT 'area'` added to `public.documents_v2` (values: public/area/restricted) |
| 0165 | `role_capabilities` truncated and reseeded with 40 rows covering the 5 canonical roles |
| 0166 | `admin` → `system_admin`, `reviewer` → `approver` renamed; UNIQUE(tenant_id, user_id) added |

## StaticAuthorizer removed

`internal/modules/iam/application/authorizer.go` and its test were deleted. `CapabilityService` (DB-backed) is the sole authorizer. `NewMiddleware` now takes `*iamapp.CapabilityService` instead of `iamdomain.Authorizer`.

## ISO segregation overlay

Independently of capabilities, the approval module enforces that the submitter of a document cannot signoff on it. See [concepts/iso-segregation.md](../concepts/iso-segregation.md).

## See also

- [vision/target-users.md](../vision/target-users.md)
- [concepts/iso-segregation.md](../concepts/iso-segregation.md)
- [modules/approval.md](approval.md)
- [references/local-dev-credentials.md](../references/local-dev-credentials.md)
