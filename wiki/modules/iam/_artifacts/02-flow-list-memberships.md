> **Last verified:** 2026-06-08 (Phase E1 casing big-bang: membershipDTO response field names updated to snake_case; prior: 2026-06-03)

## 1. Entry point
| Layer | Symbol | File:line |
|---|---|---|
| OpenAPI op | `listAreaMemberships` | `api/openapi/v1/openapi.yaml` (userId/areaCode/role param descriptions refreshed; no shape change) |
| Generated server stub | n/a — hand-written stdlib mux | `internal/modules/iam/delivery/http/routes_memberships.go:87` |
| Handler | `(*MembershipHandler).listMemberships` | `internal/modules/iam/delivery/http/routes_memberships.go:93` |

## 2. Call chain
1. `apps/api/cmd/metaldocs-api/main.go` composed handler chain (`authMiddleware.Wrap(iamMiddleware.Wrap(...mux))`) — tier-1 IAM middleware runs before route dispatch then -> calls: `internal/modules/iam/delivery/http/middleware.go:49` `(*Middleware).Wrap`
2. `internal/modules/iam/delivery/http/middleware.go` `resolver(r.Method, r.URL.Path)` + `m.caps.CanDo(...)` gate — resolves `/api/v1/iam/area-memberships` to `membership.view` capability and enforces tier-1 authz then -> calls: `apps/api/cmd/metaldocs-api/permissions.go` `newPermissionResolver` mapping and `internal/modules/iam/application/capability_service.go:31` `(*CapabilityService).CanDo`
3. `internal/modules/iam/delivery/http/routes_memberships.go:93` `(*MembershipHandler).listMemberships` — reads query params `userId`/`areaCode`/`role`; checks `isMembershipDirectoryAdmin` (`:365`) to determine scope:
   - system_admin: uses caller-supplied filters directly
   - non-admin: forces `userID = authenticated actor`; rejects userId filters targeting other users with 403
4. When `userID != ""`: calls `guardMembershipUserInTenant` (`:297`) — cross-tenant probes return 404 (mirrors PeopleHandler pattern)
5. `internal/modules/iam/delivery/http/routes_memberships.go:135` calls `h.svc.ListByTenant(ctx, tenantID, userID, areaCode, role)` — application use case at `internal/modules/iam/application/area_membership_service.go:60`
6. `internal/modules/iam/application/area_membership_service.go:61` `repo.ListByTenant(ctx, tenantID, userID, areaCode, role, now)` — passes to `internal/modules/iam/infrastructure/postgres/user_area_repository.go:57` `(*UserAreaRepository).ListByTenant`
7. `internal/modules/iam/infrastructure/postgres/user_area_repository.go:69` `r.db.QueryContext(ctx, q, tenantID, now, userID, areaCode, role)` — SELECT active rows from `user_process_areas` with optional exact-match filters via `($n = '' OR col = $n)`; scans rows and returns response items (DB driver boundary).

## 3. State changes
none

## 4. SQL touched
| File:line | Verb | Table(s) | Auth-area arg |
|---|---|---|---|
| `internal/modules/iam/application/capability_service.go:31` | SELECT (EXISTS) | `metaldocs.iam_user_roles`, `metaldocs.iam_group_members`, `metaldocs.iam_group_roles`, `metaldocs.role_capabilities` | tenant-level (`tenantID`), capability from resolver (`membership.view`) |
| `internal/modules/iam/infrastructure/postgres/user_area_repository.go:57` | SELECT | `public.user_process_areas` | `tenant_id=$1` (required); optional exact-match filters for `user_id`, `area_code`, `role` |

Tripwire pairing: N/A (read)

## 5. Response shape
- Tier-1 gate: `membership.view` (held by every role per ADR 0016; directory scope further gated by `isMembershipDirectoryAdmin` inside the handler)
- 2xx schema ref: OpenAPI `listAreaMemberships` response — `AreaMembershipListResponse` (`{items: AreaMembershipRow[]}`)
- Concrete handler payload: `200` JSON `{"items": [membershipDTO...]}` from `routes_memberships.go:145`; `membershipDTO` fields: `user_id`, `tenant_id`, `area_code`, `role`, `effective_from`, `effective_to`, `granted_by`
- Error responses: RFC 9457 Problem via `problem.Write` (`routes_memberships.go:273`); middleware 403 via `{error:{code,message,trace_id}}`

## 6. Cross-references
- Idempotency: no
- Pagination: no
- Audit log emission: no for GET list flow (audit recorded only on grant/revoke paths)
- Admin scope: system_admin (role `RoleSystemAdmin`) gets tenant-wide directory; area_admin and others see only their own rows — see `isMembershipDirectoryAdmin` at `routes_memberships.go:365`
- Cross-tenant guard: `guardMembershipUserInTenant` at `routes_memberships.go:297` (returns 404 on cross-tenant probe when userId in scope)
