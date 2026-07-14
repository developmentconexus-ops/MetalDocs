> **Last verified:** 2026-06-08 (Phase E1 casing big-bang: response field names updated to snake_case; prior: 2026-06-03)

## 1. Entry point

| Layer | Symbol | File:line |
|---|---|---|
| OpenAPI op | `grantAreaMembership` | `api/openapi/v1/openapi.yaml` (operationId confirmed present since PR-1 contract) |
| Generated server stub | n/a — hand-written stdlib mux | `internal/modules/iam/delivery/http/routes_memberships.go:88` |
| Handler | `(*MembershipHandler).grantMembership` | `internal/modules/iam/delivery/http/routes_memberships.go:149` |

## 2. Call chain

1. `internal/modules/iam/delivery/http/routes_memberships.go:149` `(*MembershipHandler).grantMembership` — decodes request body, validates fields, checks self-grant lockdown (`isSelf` at `:353`), and dispatches grant.
   calls: `internal/modules/iam/application/area_membership_service.go:64` `(*AreaMembershipService).Grant`
2. `internal/modules/iam/application/area_membership_service.go:64` `(*AreaMembershipService).Grant` — validates role and checks current active membership.
   calls: `internal/modules/iam/infrastructure/postgres/user_area_repository.go:264` `(*UserAreaRepository).GetActiveByUserAndArea`
3. `internal/modules/iam/application/area_membership_service.go:80` — when active membership exists with same role → returns `ErrMembershipExists` (409 at handler).
4. `internal/modules/iam/application/area_membership_service.go:84` — when active membership exists with different role → performs close+insert atomic grant.
   calls: `internal/modules/iam/infrastructure/postgres/user_area_repository.go:185` `(*UserAreaRepository).GrantAtomic`
5. `internal/modules/iam/infrastructure/postgres/user_area_repository.go:185` `(*UserAreaRepository).GrantAtomic` — opens SQL transaction; calls `authz.Require(CapMembershipManage)` at `:196`; UPDATE sets `effective_to` + `revoked_by` on old row (satisfies `revoked_by_required_when_revoked` CHECK); INSERT new row; commits.
6. `internal/modules/iam/application/area_membership_service.go:95` — when no active row → first grant via `Insert`.
   calls: `internal/modules/iam/infrastructure/postgres/user_area_repository.go:89` `(*UserAreaRepository).Insert`; calls `authz.Require(CapMembershipManage)` at `:100`.
7. `internal/modules/iam/delivery/http/routes_memberships.go:207` `recordMembershipAudit` — emits `iam.area_membership.granted` audit event (log-and-continue; failures do not roll back the committed grant).

Authz/tier-2 in call chain:
- `internal/modules/iam/infrastructure/postgres/user_area_repository.go:100` — `authz.Require(CapMembershipManage, "tenant")` in `Insert`
- `internal/modules/iam/infrastructure/postgres/user_area_repository.go:196` — `authz.Require(CapMembershipManage, "tenant")` in `GrantAtomic`

Idempotency interactions:
- Duplicate same-role grant on active row → `ErrMembershipExists` → 409 `MEMBERSHIP_EXISTS` (not idempotent; caller must revoke first)

## 3. State changes

| Entity | From | To | Trigger | Capability required |
|---|---|---|---|---|
| `user_process_areas` membership (same user+tenant+area, active row exists, same role) | active row | unchanged — error returned | `POST /api/v1/iam/area-memberships` | `membership.manage` |
| `user_process_areas` membership (same user+tenant+area, active row exists, different role) | prior row active (`effective_to IS NULL`) | prior row closed (`effective_to + revoked_by = actor`) and new active row inserted | `POST /api/v1/iam/area-memberships` via `GrantAtomic` | `membership.manage` |
| `user_process_areas` membership (no active row) | no active row | new active row inserted | `POST /api/v1/iam/area-memberships` via `Insert` | `membership.manage` |

## 4. SQL touched

| File:line | Verb | Table(s) | Auth-area arg |
|---|---|---|---|
| `internal/modules/iam/infrastructure/postgres/user_area_repository.go:264` | SELECT | `public.user_process_areas` | user+tenant+area+now filter |
| `internal/modules/iam/infrastructure/postgres/user_area_repository.go:200` | UPDATE (sets effective_to + revoked_by) | `public.user_process_areas` | n/a |
| `internal/modules/iam/infrastructure/postgres/user_area_repository.go:237` | INSERT | `public.user_process_areas` | n/a |
| `internal/modules/iam/infrastructure/postgres/user_area_repository.go:104` | INSERT | `public.user_process_areas` | n/a |

Tripwire pairing anchors:
- authz anchor: `internal/modules/iam/infrastructure/postgres/user_area_repository.go:100` (`Insert`), `:196` (`GrantAtomic`) — `authz.Require(CapMembershipManage, "tenant")` confirmed present.
- mutating SQL anchors: `:200` (UPDATE), `:237` (INSERT), `:104` (INSERT).
- tripwire scope check: `migrations/0142b_role_capabilities_v2_enforce.sql` + `migrations/0188_tripwire_extend.sql` — `trg_require_cap_asserted` attached to `user_process_areas` by migration 0188.
- pairing result: **active** — tier-2 `authz.Require` present in all write methods; tripwire attached by migration 0188.

## 5. Response shape

- 2xx schema ref: OpenAPI `grantAreaMembership` — `GrantAreaMembershipResponse` (HTTP 201 JSON `{user_id, tenant_id, area_code, role}`)
- Error responses (RFC 9457 Problem):
  - 400 `VALIDATION_ERROR` — missing/blank fields
  - 403 `AUTH_FORBIDDEN` — self-grant or insufficient manage scope
  - 404 `NOT_FOUND` — cross-tenant user probe
  - 409 `MEMBERSHIP_EXISTS` — duplicate same-role grant on active row
  - 422/400 `UNKNOWN_ROLE` — unrecognised role value
  - 500 `INTERNAL_ERROR` — unexpected failure
- Handler success path: `routes_memberships.go:214` returns 201.

## 6. Cross-references

- Idempotency: no (non-idempotent; same-role duplicate = 409; role-change = close+insert via `GrantAtomic`)
- Pagination: no
- Audit log emission: yes — `iam.area_membership.granted` emitted at `routes_memberships.go:207` after committed write; `MembershipGovernanceLogger` in application service wired as `nil` in production (`main.go`)
- Self-grant lockdown: `isSelf(ctx, userID)` at `routes_memberships.go:175` — `CapMembershipManage` holder cannot grant themselves additional area roles

Tier-1 authz middleware path for this route:
- Permission mapping: `apps/api/cmd/metaldocs-api/permissions.go` maps `POST /api/v1/iam/area-memberships` to `iamdomain.CapMembershipManage`.
- Enforcement call: `internal/modules/iam/delivery/http/middleware.go` resolves route capability and enforces with `CapabilityService.CanDo`.
