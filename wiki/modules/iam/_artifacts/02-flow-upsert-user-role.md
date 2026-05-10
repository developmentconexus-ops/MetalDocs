# Data-flow trace — upsertUserRole

Operation: `upsertUserRole`  
HTTP: `POST /api/v1/iam/users/{userId}/roles`  
Module: `internal/modules/iam`

## 1. Entry point

| Layer | Symbol | File:line |
|---|---|---|
| OpenAPI op | `(unclear: no operationId on this POST block)` | `api/openapi/v1/openapi.yaml:2642` `api/openapi/v1/openapi.yaml:2643` |
| Generated server stub | `n/a — hand-written stdlib mux` | `internal/modules/iam/delivery/http/admin_handler.go:82` |
| Handler | `(*AdminHandler).handleUserRoleUpsert` | `internal/modules/iam/delivery/http/admin_handler.go:316` |

## 2. Call chain

1. `internal/modules/iam/delivery/http/admin_handler.go:189` `(*AdminHandler).handleUserRoute` — parses `/api/v1/iam/users/{userId}/...` and dispatches POST `roles` to upsert handler.  
   → calls: `internal/modules/iam/delivery/http/admin_handler.go:196` `(*AdminHandler).handleUserRoleUpsert`
2. `internal/modules/iam/delivery/http/admin_handler.go:316` `(*AdminHandler).handleUserRoleUpsert` — validates method/body/role, resolves `assignedBy` and tenant, invokes admin service.  
   → calls: `internal/modules/iam/delivery/http/admin_handler.go:350` `(*AdminHandler).service.UpsertUserAndAssignRole`
3. `internal/modules/iam/application/admin_service.go:23` `(*AdminService).UpsertUserAndAssignRole` — trims inputs/defaults and calls role admin repository.  
   → calls: `internal/modules/iam/application/admin_service.go:38` `RoleAdminRepository.UpsertUserAndAssignRole`
4. `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:33` `(*RoleAdminRepository).UpsertUserAndAssignRole` — creates SQL transaction boundary.  
   → calls: `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:34` `db.BeginTx`
5. `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:50` `tx.ExecContext` — deletes existing role row for `(tenant_id,user_id)` from `iam_user_roles`.  
   → calls: `database/sql` `(*Tx).ExecContext` (via `tx.ExecContext`)
6. `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:56` `tx.ExecContext` — inserts replacement role row into `iam_user_roles`.  
   → calls: `database/sql` `(*Tx).ExecContext` (via `tx.ExecContext`)
7. `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:63` `tx.Commit` — commits role mutation transaction.  
   → calls: `database/sql` `(*Tx).Commit`
8. `internal/modules/iam/application/admin_service.go:42` `RoleCacheInvalidator.InvalidateUser` — invalidates cached roles for user after successful repository call.  
   → calls: `internal/modules/iam/application/cached_role_provider.go:65` `(*CachedRoleProvider).InvalidateUser`

Authz capability gate on route path:
- `apps/api/cmd/metaldocs-api/permissions.go:54` maps `POST /api/v1/iam/users/*/roles` to `iamdomain.CapUserManage`.
- `internal/modules/iam/delivery/http/middleware.go:61` resolves capability from resolver.
- `internal/modules/iam/delivery/http/middleware.go:83` enforces with `m.caps.CanDo(...)`.

Idempotency interaction:
- `(not found)` for this route in handler chain files above.

## 3. State changes

| Entity | From | To | Trigger | Capability required |
|---|---|---|---|---|
| `metaldocs.iam_user_roles` row for `(tenant_id,user_id)` | existing row or none | replaced with one inserted row for requested role | `POST /api/v1/iam/users/{userId}/roles` | `user.manage` (`iamdomain.CapUserManage`) |

Anchors: delete `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:51`; insert `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:57`; capability mapping `apps/api/cmd/metaldocs-api/permissions.go:54`.

## 4. SQL touched

| File:line | Verb | Table(s) | Auth-area arg (if any) |
|---|---|---|---|
| `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:46` | INSERT .. ON CONFLICT UPDATE | `metaldocs.iam_users` | none |
| `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:50` | DELETE | `metaldocs.iam_user_roles` | none |
| `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:56` | INSERT | `metaldocs.iam_user_roles` | none |

Tripwire pairing (`authz.Require(...)` before mutating SQL on same tx):
- `authz.Require` anchor: `(not found)` in `internal/modules/iam/delivery/http/admin_handler.go`, `internal/modules/iam/application/admin_service.go`, `internal/modules/iam/infrastructure/postgres/role_admin_repository.go`.
- Mutating SQL anchors: `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:50` and `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:56`.
- Pairing status for this repo call: `N/A`.

Trigger check on `iam_user_roles` for `enforce_capability_asserted`:
- `migrations/*` contains `enforce_capability_asserted` function/trigger references for approval tables (`migrations/0142b_role_capabilities_v2_enforce.sql:201`, `:207`).
- `(not found)` trigger reference targeting `metaldocs.iam_user_roles` in migrations search results.
- Tripwire pairing annotation: `N/A (no trigger on iam_user_roles)`.

## 5. Response shape

- 2xx schema ref: `#/components/schemas/UpsertUserRoleResponse` (`api/openapi/v1/openapi.yaml:2663`; schema defined at `api/openapi/v1/openapi.yaml:5054`).
- Request schema ref: `#/components/schemas/UpsertUserRoleRequest` (`api/openapi/v1/openapi.yaml:2656`; schema defined at `api/openapi/v1/openapi.yaml:5043`).
- Error responses declared on this POST op: `400`, `401`, `403` (all `ApiErrorEnvelope` at `api/openapi/v1/openapi.yaml:2667`, `api/openapi/v1/openapi.yaml:2673`, `api/openapi/v1/openapi.yaml:2679`; schema component `api/openapi/v1/openapi.yaml:5210`).
- Problem `type` URI: `(not found)` in this OpenAPI op block.

## 6. Cross-references

- Idempotency: no (`not found` in handler/service/repository chain files for this op).
- Pagination: no (single upsert response map at `internal/modules/iam/delivery/http/admin_handler.go:355`).
- Audit log emission: no in this handler path (`handleUserRoleUpsert` has no `recordAudit` call between `:316` and `:360`).
  - Audit sink implementation exists at `internal/modules/audit/infrastructure/postgres/writer.go:20` (`INSERT INTO metaldocs.audit_events` at `:22`), invoked by `recordAudit` via `h.audit.Record(...)` at `internal/modules/iam/delivery/http/admin_handler.go:457`.
  - Emission timing for this op: `(not found)` because this op does not call `recordAudit`; repository transaction commits at `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:63`.
