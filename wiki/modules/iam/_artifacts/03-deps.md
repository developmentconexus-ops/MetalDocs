# IAM module cross-dependencies (`internal/modules/iam`)

## 1. Imports OUT

| Imported package | First seen in (file:line) | Symbols used | Purpose (1 line) |
|---|---|---|---|
| `metaldocs/internal/modules/audit/domain` | `internal/modules/iam/delivery/http/admin_handler.go:12` (`iam/delivery/http`) | `Writer`, `Reader`, `Event`, `ListEventsQuery` | IAM admin HTTP handler reads/writes audit events. |
| `metaldocs/internal/modules/auth/domain` | `internal/modules/iam/delivery/http/admin_handler.go:13` (`iam/delivery/http`) | `ManagedUser`, `OnlineUser`, `UpdateUserParams`, `ErrPasswordPolicy`, `ErrUserAlreadyExists`, `ErrIdentityNotFound`, `CurrentUserFromContext` | IAM HTTP layer interoperates with auth user/domain models and auth-context checks. |
| `metaldocs/internal/platform/authn` | `internal/modules/iam/delivery/http/admin_handler.go:16` (`iam/delivery/http`) | `UserIDFromContext` | IAM HTTP handlers resolve authenticated actor/user from request context. |
| `metaldocs/internal/platform/httpresponse` | `internal/modules/iam/delivery/http/middleware.go:11` (`iam/delivery/http`) | `WriteJSON` | IAM middleware writes structured API error envelopes. |
| `metaldocs/internal/platform/tenant` | `internal/modules/iam/delivery/http/admin_handler.go:17` (`iam/delivery/http`) | `DevTenantID` | IAM HTTP handlers/middleware default tenant when request header is absent. |

## 2. Imports IN

| Importer package | File:line of import | Symbols used | Notes |
|---|---|---|---|
| `metaldocs/apps/api/cmd/metaldocs-api` | `apps/api/cmd/metaldocs-api/main.go:42` | `iamapp.NewCapabilityService`, `iamapp.NewCachedRoleProvider`, `iamapp.NewAdminService`, `iamapp.NewAreaMembershipService`, `iamdelivery.NewMiddleware`, `iamdelivery.NewAdminHandler`, `iamdelivery.NewMembershipHandler`, `iampg.NewUserAreaRepository` | `? from iam/application`, `iam/delivery/http`, `iam/infrastructure/postgres` |
| `metaldocs/apps/api/cmd/metaldocs-api` | `apps/api/cmd/metaldocs-api/permissions.go:8` | `iamdelivery.PermissionResolver`, `iamdomain.Cap*` constants | `? from iam/delivery/http`, `iam/domain` |
| `metaldocs/apps/api/internal/wiring` | `apps/api/internal/wiring/documents.go:7` | `iamapp.CapabilityService`, `iamdomain.Capability` | `? from iam/application`, `iam/domain`; adapter returns documents `CapabilityChecker` |
| `metaldocs/internal/modules/documents/application` | `internal/modules/documents/application/fillin_authz.go:9` | `authz.WithCapCache`, `authz.Require`, `iamdomain.Capability` | `? from iam/authz`, `iam/domain` |
| `metaldocs/internal/modules/documents/approval/application` | `internal/modules/documents/approval/application/cancel_service.go:12` | `authz.WithCapCache`, `authz.Require`, `authz.BypassSystem` | `? from iam/authz` |
| `metaldocs/internal/modules/documents/approval/http` | `internal/modules/documents/approval/http/errors.go:17` | `authz.ErrCapabilityDenied`, `iamdomain.UserIDFromContext` | `? from iam/authz`, `iam/domain` |
| `metaldocs/internal/modules/documents/delivery/http` | `internal/modules/documents/delivery/http/handler.go:17` | `iamapp.ErrAccessDenied`, `iamapp.ErrCapabilityDenied`, `iamdomain.WithAuthContext`, `iamdomain.UserIDFromContext`, `iamdomain.RolesFromContext` | `? from iam/application`, `iam/domain` |
| `metaldocs/internal/modules/documents/http` | `internal/modules/documents/http/view_handler.go:9` | `authz.ErrCapabilityDenied`, `iamdomain.UserIDFromContext` | `? from iam/authz`, `iam/domain` |
| `metaldocs/internal/modules/templates_v2/delivery/http` | `internal/modules/templates_v2/delivery/http/routes_lifecycle.go:8` | `iamdomain.RolesFromContext`, `iamdomain.UserIDFromContext` | `? from iam/domain` |
| `metaldocs/internal/modules/auth/application` | `internal/modules/auth/application/service.go:18` | `iamdomain.Role` | `? from iam/domain` |
| `metaldocs/internal/modules/auth/delivery/http` | `internal/modules/auth/delivery/http/middleware.go:10` | `iamdomain.WithAuthContext`, `iamdomain.Role` | `? from iam/domain` |
| `metaldocs/internal/modules/auth/domain` | `internal/modules/auth/domain/model.go:6` | `iamdomain.Role` | `? from iam/domain` |
| `metaldocs/internal/modules/auth/infrastructure/memory` | `internal/modules/auth/infrastructure/memory/repository.go:10` | `iamdomain.Role` | `? from iam/domain` |
| `metaldocs/internal/platform/bootstrap` | `internal/platform/bootstrap/api.go:19` | `iamdomain.RoleProvider`, `iampg.NewRoleProvider`, `iampg.NewRoleAdminRepository` | `? from iam/domain`, `iam/infrastructure/postgres` |
| `metaldocs/internal/platform/authn` | `internal/platform/authn/context.go:7` | `iamdomain.WithAuthContext`, `iamdomain.Role` | `? from iam/domain` |
| `metaldocs/internal/platform/security` | `internal/platform/security/ratelimit.go:13` | `iamdomain.UserIDFromContext` | `? from iam/domain` |
| `metaldocs/internal/testsupport/http` | `internal/testsupport/http/auth_request.go:9` | `iamdomain.WithAuthContext` | `? from iam/domain` |

Confirmed module-path check: no directory `internal/modules/approval/` exists in this workspace; IAM imports were verified under `internal/modules/documents/approval/...`.

## 3. DI / wiring touchpoints

| Site | File:line | What is wired |
|---|---|---|
| API main | `apps/api/cmd/metaldocs-api/main.go:163` | `iamapp.NewCapabilityService(deps.SQLDB)` |
| API main | `apps/api/cmd/metaldocs-api/main.go:165` | `iamapp.NewCachedRoleProvider(deps.RoleProvider, authn.CacheTTL())` |
| API main | `apps/api/cmd/metaldocs-api/main.go:173` | `iamdelivery.NewMiddleware(capabilityService, cachedProvider, authn.Enabled(), authCfg.LegacyHeaderEnabled)` |
| API main | `apps/api/cmd/metaldocs-api/main.go:181` | `iamapp.NewAdminService(deps.RoleAdminRepo, cachedProvider)` |
| API main | `apps/api/cmd/metaldocs-api/main.go:182` | `iamdelivery.NewAdminHandler(iamAdminService, authService, deps.AuditWriter)` |
| API main | `apps/api/cmd/metaldocs-api/main.go:217` | `iampg.NewUserAreaRepository(deps.SQLDB)` (argument to IAM area-membership service constructor) |
| API main | `apps/api/cmd/metaldocs-api/main.go:217` | `iamapp.NewAreaMembershipService(iampg.NewUserAreaRepository(deps.SQLDB), nil)` |
| API main | `apps/api/cmd/metaldocs-api/main.go:219` | `iamdelivery.NewMembershipHandler(membershipService)` |

## 4. Configuration surface

| Name | Read at (file:line) | Required? | Default |
|---|---|---|---|
| `APP_ENV` | `internal/modules/iam/application/startup.go:62` | No | Behavior branch only (`development` check before returning DB write error). |
| `DATABASE_URL` | `internal/modules/iam/integration_test.go:28` | Test-only | Falls back to `METALDOCS_DATABASE_URL` when empty. |
| `METALDOCS_DATABASE_URL` | `internal/modules/iam/integration_test.go:30` | Test-only | Used when `DATABASE_URL` is empty. |
| `DATABASE_URL` | `internal/modules/iam/domain/role_capabilities_integration_test.go:19` | Test-only | Falls back to `METALDOCS_DATABASE_URL` when empty. |
| `METALDOCS_DATABASE_URL` | `internal/modules/iam/domain/role_capabilities_integration_test.go:21` | Test-only | Used when `DATABASE_URL` is empty. |

String scan results in repository: no occurrences of `IAM_AUTHZ_ENFORCED` and no occurrences of `IAM_AUTHZ_LEGACY_HEADER`.

## 5. Test surface

| Test file | Subject (file under test) | Kind (unit / integration / e2e) |
|---|---|---|
| `internal/modules/iam/integration_test.go` | `internal/modules/iam` module integration path | integration |
| `internal/modules/iam/application/area_membership_test.go` | `internal/modules/iam/application/area_membership_service.go` | unit |
| `internal/modules/iam/application/authorization_bench_test.go` | `internal/modules/iam/application/authorization.go` | unit (benchmark) |
| `internal/modules/iam/application/authorization_test.go` | `internal/modules/iam/application/authorization.go` | unit |
| `internal/modules/iam/area_membership/area_membership_test.go` | `internal/modules/iam/area_membership/area_membership.go` | unit |
| `internal/modules/iam/authz/authz_bypass_test.go` | `internal/modules/iam/authz/authz.go` | unit |
| `internal/modules/iam/authz/authz_test.go` | `internal/modules/iam/authz/authz.go` | unit |
| `internal/modules/iam/authz/context_test.go` | `internal/modules/iam/authz/context.go` | unit |
| `internal/modules/iam/delivery/http/middleware_test.go` | `internal/modules/iam/delivery/http/middleware.go` | unit |
| `internal/modules/iam/delivery/http/routes_memberships_contract_test.go` | `internal/modules/iam/delivery/http/routes_memberships.go` | integration (HTTP contract) |
| `internal/modules/iam/domain/role_capabilities_integration_test.go` | `internal/modules/iam/domain/role_capabilities.go` | integration |
| `internal/modules/iam/domain/role_capabilities_test.go` | `internal/modules/iam/domain/role_capabilities.go` | unit |
| `internal/modules/iam/infrastructure/postgres/role_admin_repository_test.go` | `internal/modules/iam/infrastructure/postgres/role_admin_repository.go` | integration |
| `internal/modules/iam/infrastructure/postgres/role_provider_test.go` | `internal/modules/iam/infrastructure/postgres/role_provider.go` | integration |
| `tests/integration/iam/capability_service_test.go` | IAM capability service integration behavior | integration |
| `tests/integration/iam/migration_0170_test.go` | IAM migration 0170 integration behavior | integration |
| `tests/integration/iam/tenant_isolation_test.go` | IAM tenant isolation integration behavior | integration |
