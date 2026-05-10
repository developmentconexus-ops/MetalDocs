## 1. Entry point

| Layer | Symbol | File:line |
|---|---|---|
| OpenAPI op | `grantAreaMembership` | `api/openapi/v1/openapi.yaml` (unclear: no `operationId: grantAreaMembership` found) |
| Generated server stub | `ServerInterface.<Method>` | `n/a — hand-written stdlib mux` |
| Handler | `(*MembershipHandler).grantMembership` | `internal/modules/iam/delivery/http/routes_memberships.go:58` |

## 2. Call chain

1. `internal/modules/iam/delivery/http/routes_memberships.go:58` `(*MembershipHandler).grantMembership` — decodes request body, validates fields, and dispatches grant.
   ? calls: `internal/modules/iam/application/area_membership_service.go:49` `(*AreaMembershipService).Grant`
2. `internal/modules/iam/application/area_membership_service.go:49` `(*AreaMembershipService).Grant` — validates role and checks current active membership.
   ? calls: `internal/modules/iam/infrastructure/postgres/user_area_repository.go:154` `(*UserAreaRepository).GetActiveByUserAndArea`
3. `internal/modules/iam/application/area_membership_service.go:75` `(*AreaMembershipService).Grant` — when active membership exists, performs close+insert atomic grant.
   ? calls: `internal/modules/iam/infrastructure/postgres/user_area_repository.go:90` `(*UserAreaRepository).GrantAtomic`
4. `internal/modules/iam/infrastructure/postgres/user_area_repository.go:90` `(*UserAreaRepository).GrantAtomic` — opens SQL transaction, updates prior active row, inserts new row, commits.
   ? calls: `internal/modules/iam/infrastructure/postgres/user_area_repository.go:91` `(*sql.DB).BeginTx`; `internal/modules/iam/infrastructure/postgres/user_area_repository.go:108` `(*sql.Tx).ExecContext`; `internal/modules/iam/infrastructure/postgres/user_area_repository.go:134` `(*sql.Tx).ExecContext`; `internal/modules/iam/infrastructure/postgres/user_area_repository.go:148` `(*sql.Tx).Commit`
5. `internal/modules/iam/application/area_membership_service.go:79` `(*AreaMembershipService).Grant` — emits governance log when logger is configured.
   ? calls: `internal/modules/iam/application/area_membership_service.go:79` `MembershipGovernanceLogger.Log` (unclear: concrete implementation not wired for this route)

Authz/tier-2 in call chain:
- `internal/modules/iam/application/area_membership_service.go` `authz.Require` not found.
- `internal/modules/iam/infrastructure/postgres/user_area_repository.go` `authz.Require` not found.

Idempotency interactions:
- none found in handler/service/repository files above.

## 3. State changes

| Entity | From | To | Trigger | Capability required |
|---|---|---|---|---|
| `user_process_areas` membership (same user+tenant+area, active row exists) | prior row active (`effective_to IS NULL`) | prior row closed (`effective_to = newMembership.effective_from`) and new active row inserted | `POST /api/v2/iam/area-memberships` via `GrantAtomic` | tier-1 route capability `membership.manage` (`apps/api/cmd/metaldocs-api/permissions.go:196-197`) |
| `user_process_areas` membership (no active row) | no active row | new active row inserted | `POST /api/v2/iam/area-memberships` via `Insert` | tier-1 route capability `membership.manage` (`apps/api/cmd/metaldocs-api/permissions.go:196-197`) |

## 4. SQL touched

| File:line | Verb | Table(s) | Auth-area arg (if any) |
|---|---|---|---|
| `internal/modules/iam/infrastructure/postgres/user_area_repository.go:166` | SELECT | `user_process_areas` | n/a |
| `internal/modules/iam/infrastructure/postgres/user_area_repository.go:108` | UPDATE | `user_process_areas` | n/a |
| `internal/modules/iam/infrastructure/postgres/user_area_repository.go:134` | INSERT | `user_process_areas` | n/a |
| `internal/modules/iam/infrastructure/postgres/user_area_repository.go:58` | INSERT | `user_process_areas` | n/a |

Tripwire pairing anchors:
- authz anchor: `internal/modules/iam/application/area_membership_service.go` `authz.Require` not found; `internal/modules/iam/infrastructure/postgres/user_area_repository.go` `authz.Require` not found.
- mutating SQL anchors: `internal/modules/iam/infrastructure/postgres/user_area_repository.go:108`, `:134`, `:58`.
- tripwire scope check: `migrations/0142b_role_capabilities_v2_enforce.sql:201-209` attaches `enforce_capability_asserted` only to `public.approval_instances` and `public.approval_signoffs`; no attachment to `user_process_areas` found.
- pairing result: `N/A` (mutating table in this flow is `user_process_areas`, not tripwire-attached in migration 0142b).

## 5. Response shape

- 2xx schema ref: `(unclear: no OpenAPI declaration for POST /api/v2/iam/area-memberships in api/openapi/v1/openapi.yaml)`
- Error responses declared on op + Problem type URI: `(unclear: no OpenAPI declaration for this operation in api/openapi/v1/openapi.yaml)`
- Handler-emitted success body (code path): `internal/modules/iam/delivery/http/routes_memberships.go:88-93` returns HTTP `201` JSON object keys `userId`, `tenantId`, `areaCode`, `role`.

## 6. Cross-references

- Idempotency: no
- Pagination: no
- Audit log emission: yes, conditional via `MembershipGovernanceLogger.Log` at `internal/modules/iam/application/area_membership_service.go:79` and `:101`; runtime wiring for API route passes `nil` logger at `apps/api/cmd/metaldocs-api/main.go:217`.

Tier-1 authz middleware path for this route:
- Permission mapping: `apps/api/cmd/metaldocs-api/permissions.go:196-197` maps `/api/v2/iam/area-memberships` to `iamdomain.CapMembershipManage`.
- Middleware integration: `apps/api/cmd/metaldocs-api/main.go:170-174` wires resolver into IAM middleware.
- Enforcement call: `internal/modules/iam/delivery/http/middleware.go:61-63` resolves route capability and `:83-85` enforces with `CapabilityService.CanDo`.