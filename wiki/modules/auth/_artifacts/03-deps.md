# Auth Module Cross-Dependencies

Module: `internal/modules/auth`
Artifact: `wiki/modules/auth/_artifacts/03-deps.md`

## 1. Imports OUT

| Imported package | First seen in (file:line) | Symbols used | Purpose (1 line) |
|---|---|---|---|
| `internal/modules/iam/domain` | `internal/modules/auth/application/service.go:18` | `RoleProvider`, `RoleAdminRepository`, `RoleSystemAdmin`, `ErrUserNotFound`, `Role`, `ErrUserInactive`, `WithAuthContext` | IAM roles/types/context for auth session/user workflows |
| `internal/platform/tenant` | `internal/modules/auth/application/service.go:19` | `DevTenantID` | Fallback/default tenant id for auth flows |
| `internal/platform/httpresponse` | `internal/modules/auth/delivery/http/handler.go:13` | `WriteJSON` | HTTP JSON response writer for auth handlers |

## 2. Imports IN

| Importer package | File:line of import | Symbols used | Notes |
|---|---|---|---|
| `apps/api/cmd/metaldocs-api` | `apps/api/cmd/metaldocs-api/main.go:40` | `authapp.NewService` | API composition root wires auth service |
| `apps/api/cmd/metaldocs-api` | `apps/api/cmd/metaldocs-api/main.go:41` | `authdelivery.NewHandler`, `authdelivery.NewMiddleware` | API composition root wires auth HTTP delivery |
| `apps/api/cmd/metaldocs-api` | `apps/api/cmd/metaldocs-api/permissions.go:7` | `authdelivery.PublicPathChecker` | Public-path checker type for auth middleware |
| `apps/api/cmd/metaldocs-e2e-seed` | `apps/api/cmd/metaldocs-e2e-seed/main.go:10` | `authapp.NewService`, `*authapp.Service` | E2E seed binary builds and uses auth service |
| `apps/api/cmd/metaldocs-e2e-seed` | `apps/api/cmd/metaldocs-e2e-seed/main.go:11` | `authdomain.UpdateUserParams` | E2E seed updates auth user fields |
| `internal/platform/authn` | `internal/platform/authn/config.go:10` | `authapp.Config` | Runtime auth config builder returns auth config |
| `internal/platform/bootstrap` | `internal/platform/bootstrap/api.go:16` | `authdomain.Repository` | Dependency container type includes auth repo interface |
| `internal/platform/bootstrap` | `internal/platform/bootstrap/api.go:17` | `authmemory.NewRepository` | In-memory auth repository wiring |
| `internal/platform/bootstrap` | `internal/platform/bootstrap/api.go:18` | `authpg.NewRepository` | Postgres auth repository wiring |
| `internal/platform/observability` | `internal/platform/observability/http.go:15` | `authdomain.CurrentUserFromContext` | Reads current user from auth context for logs |
| `internal/platform/security` | `internal/platform/security/ratelimit.go:12` | `authdomain.CurrentUserFromContext` | Uses auth user id for rate-limit identity |
| `internal/modules/iam/delivery/http` | `internal/modules/iam/delivery/http/admin_handler.go:13` | `authdomain.ManagedUser`, `authdomain.OnlineUser`, `authdomain.UpdateUserParams`, `ErrPasswordPolicy`, `ErrUserAlreadyExists`, `ErrIdentityNotFound` | IAM admin handler contracts and error mapping to auth types |
| `internal/modules/iam/delivery/http` | `internal/modules/iam/delivery/http/middleware.go:8` | `authdomain.CurrentUserFromContext` | IAM middleware checks whether auth already populated current user |

## 3. DI / wiring touchpoints

| Site | File:line | What is wired |
|---|---|---|
| `internal/platform/bootstrap` | `internal/platform/bootstrap/api.go:74` | `authpg.NewRepository(db)` for postgres mode |
| `internal/platform/bootstrap` | `internal/platform/bootstrap/api.go:113` | `authmemory.NewRepository()` for memory mode |
| `apps/api/cmd/metaldocs-api` | `apps/api/cmd/metaldocs-api/main.go:148` | `authapp.NewService(deps.AuthRepo, deps.RoleProvider, deps.RoleAdminRepo, authCfg)` |
| `apps/api/cmd/metaldocs-api` | `apps/api/cmd/metaldocs-api/main.go:158` | `authdelivery.NewHandler(authService)` |
| `apps/api/cmd/metaldocs-api` | `apps/api/cmd/metaldocs-api/main.go:171` | `authdelivery.NewMiddleware(authService, authCfg, authn.Enabled()).WithPublicPathChecker(...)` |
| `apps/api/cmd/metaldocs-e2e-seed` | `apps/api/cmd/metaldocs-e2e-seed/main.go:58` | `authapp.NewService(deps.AuthRepo, deps.RoleProvider, deps.RoleAdminRepo, authCfg)` |

## 4. Configuration surface

| Name | Read at (file:line) | Required? | Default |
|---|---|---|---|
| `SessionCookieName` | `internal/platform/authn/config.go:101` | No | `metaldocs_session` when env empty |
| `SessionTTL` | `internal/platform/authn/config.go:102` | No | `12h` (`METALDOCS_AUTH_SESSION_TTL_HOURS` fallback) |
| `SessionSecret` | `internal/platform/authn/config.go:103` | Yes when `authn.Enabled()==true` (`internal/platform/authn/config.go:42`) | none |
| `PasswordMinLength` | `internal/platform/authn/config.go:104` | No | `8` |
| `LoginMaxFailedAttempts` | `internal/platform/authn/config.go:105` | No | `5` |
| `LoginLockDuration` | `internal/platform/authn/config.go:106` | No | `15m` |
| `LegacyHeaderEnabled` | `internal/platform/authn/config.go:107` | No | `false` |
| `OriginProtection` | `internal/platform/authn/config.go:116` | No | `authn.Enabled()` |
| `TrustedOrigins` | `internal/platform/authn/config.go:115` | No | `nil`/empty |
| `BootstrapAdminEnabled` | `internal/platform/authn/config.go:108` | No | `APP_ENV==local` |
| `BootstrapAdminUserID` | `internal/platform/authn/config.go:109` | No | `admin-local` |
| `BootstrapAdminUsername` | `internal/platform/authn/config.go:110` | No | `admin` |
| `BootstrapAdminEmail` | `internal/platform/authn/config.go:111` | No | empty string |
| `BootstrapAdminPassword` | `internal/platform/authn/config.go:112` | Yes when `BootstrapAdminEnabled==true` (`internal/platform/authn/config.go:119`) | none |
| `BootstrapAdminName` | `internal/platform/authn/config.go:113` | No | `Administrator` |
| `CookieSecure` | `internal/platform/authn/config.go:114` | No | `APP_ENV!=local` |

## 5. Test surface

| Test file | Subject (file under test) | Kind (unit / integration / e2e) |
|---|---|---|
| `internal/modules/auth/delivery/http/middleware_test.go` | `internal/modules/auth/delivery/http/middleware.go` | unit |
| `internal/modules/auth/infrastructure/postgres/repository_test.go` | `internal/modules/auth/infrastructure/postgres/repository.go` | integration (DB-backed) |
| `tests/unit/auth_password_change_flow_test.go` | auth flow across `application` + `delivery/http` + `infrastructure/memory` | unit |
| `tests/unit/auth_login_policy_test.go` | `internal/modules/auth/application/service.go` (login/lock policy) | unit |

Integration scan for `tests/integration/**/auth*`: none found.

# OUT: 3, # IN: 13, # DI: 6
