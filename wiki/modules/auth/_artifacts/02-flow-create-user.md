# Admin User Creation Data Flow (`CreateUser`)

- Last verified date: 2026-05-10
- Status: Research

## 1. Entry point
- Entry point in auth HTTP layer: **n/a** for this operation. In the IAM HTTP layer, `POST /api/v1/iam/users` dispatches to `handleCreateUser` (`internal/modules/iam/delivery/http/admin_handler.go:83`, `internal/modules/iam/delivery/http/admin_handler.go:88`, `internal/modules/iam/delivery/http/admin_handler.go:92`, `internal/modules/iam/delivery/http/admin_handler.go:93`).
- Exact call site to auth application service: `h.authService.CreateUser(...)` in `handleCreateUser` (`internal/modules/iam/delivery/http/admin_handler.go:259`, `internal/modules/iam/delivery/http/admin_handler.go:280`).

## 2. Call chain
- `handleCreateUser` parses request, roles, tenant, actor, then calls `authService.CreateUser` (`internal/modules/iam/delivery/http/admin_handler.go:265`, `internal/modules/iam/delivery/http/admin_handler.go:270`, `internal/modules/iam/delivery/http/admin_handler.go:275`, `internal/modules/iam/delivery/http/admin_handler.go:279`, `internal/modules/iam/delivery/http/admin_handler.go:280`).
- `Service.CreateUser` entry is at `internal/modules/auth/application/service.go:279`.
- `Service.CreateUser` calls `validatePassword(password)` (`internal/modules/auth/application/service.go:297`) then `hashPassword(password)` (`internal/modules/auth/application/service.go:300`).
- `hashPassword` uses `bcrypt.GenerateFromPassword(..., bcrypt.DefaultCost)` (`internal/modules/auth/application/service.go:431`, `internal/modules/auth/application/service.go:432`).
- `Service.CreateUser` calls `s.repo.CreateUser(...)` (`internal/modules/auth/application/service.go:305`).
- Postgres `Repository.CreateUser` starts its own DB transaction with `BeginTx`, inserts `metaldocs.auth_identities`, then commits (`internal/modules/auth/infrastructure/postgres/repository.go:151`, `internal/modules/auth/infrastructure/postgres/repository.go:152`, `internal/modules/auth/infrastructure/postgres/repository.go:163`, `internal/modules/auth/infrastructure/postgres/repository.go:166`, `internal/modules/auth/infrastructure/postgres/repository.go:170`).
- `Service.CreateUser` then calls `s.roleAdmin.ReplaceUserRoles(...)` (`internal/modules/auth/application/service.go:325`).
- `RoleAdminRepository.ReplaceUserRoles` also starts its own transaction with `BeginTx`, upserts `metaldocs.iam_users`, deletes prior `metaldocs.iam_user_roles`, optionally inserts one new `metaldocs.iam_user_roles`, and commits (`internal/modules/iam/infrastructure/postgres/role_admin_repository.go:72`, `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:73`, `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:80`, `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:85`, `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:90`, `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:106`, `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:112`).
- Transaction boundary fact: `repo.CreateUser` and `ReplaceUserRoles` execute in distinct transactions (each function independently calls `BeginTx` and commits) and there is no shared outer transaction in `Service.CreateUser` (`internal/modules/auth/application/service.go:279`, `internal/modules/auth/application/service.go:305`, `internal/modules/auth/application/service.go:325`, `internal/modules/auth/infrastructure/postgres/repository.go:152`, `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:73`).

## 3. State changes
- `metaldocs.auth_identities`: one row inserted by auth repository `CreateUser` (`internal/modules/auth/infrastructure/postgres/repository.go:163`, `internal/modules/auth/infrastructure/postgres/repository.go:166`).
- `metaldocs.iam_users`: upserted during role replacement (`internal/modules/iam/infrastructure/postgres/role_admin_repository.go:80`, `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:85`).
- `metaldocs.iam_user_roles`: prior rows deleted for `(tenant_id, user_id)` and then one row inserted when `lastRole` is non-empty (`internal/modules/iam/infrastructure/postgres/role_admin_repository.go:90`, `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:95`, `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:101`, `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:106`).
- Migration context for FK coupling/decoupling:
- `0021` defines `auth_identities.user_id` referencing `iam_users(user_id)` (`migrations/0021_init_auth_identities_and_sessions.sql:1`, `migrations/0021_init_auth_identities_and_sessions.sql:2`).
- `0036` drops FK from `auth_identities` to `iam_users` and rewires `auth_sessions` FK to `auth_identities` (`migrations/0036_decouple_auth_identity_from_iam_user_tables.sql:39`, `migrations/0036_decouple_auth_identity_from_iam_user_tables.sql:40`, `migrations/0036_decouple_auth_identity_from_iam_user_tables.sql:82`, `migrations/0036_decouple_auth_identity_from_iam_user_tables.sql:85`).

## 4. SQL touched
- Auth write:
- `INSERT INTO metaldocs.auth_identities (...)` (`internal/modules/auth/infrastructure/postgres/repository.go:163`).
- IAM role-admin writes inside `ReplaceUserRoles`:
- `INSERT INTO metaldocs.iam_users (...) ... ON CONFLICT ...` (`internal/modules/iam/infrastructure/postgres/role_admin_repository.go:80`, `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:83`).
- `DELETE FROM metaldocs.iam_user_roles WHERE tenant_id = $1::uuid AND user_id = $2` (`internal/modules/iam/infrastructure/postgres/role_admin_repository.go:90`).
- `INSERT INTO metaldocs.iam_user_roles (...)` (`internal/modules/iam/infrastructure/postgres/role_admin_repository.go:106`).
- Tripwire scope: N/A for this auth/iam admin flow (no tripwire-specific logic referenced in these functions).

## 5. Response shape
- On success, `handleCreateUser` returns HTTP `201 Created` with JSON body `{"userId": strings.TrimSpace(defaultString(req.UserID, req.Username))}` (`internal/modules/iam/delivery/http/admin_handler.go:284`).
- Error responses are delegated to `h.writeAuthError(...)` from this handler path (`internal/modules/iam/delivery/http/admin_handler.go:281`).

## 6. Cross-refs
- Idempotency: no explicit idempotency-key handling in this flow (`internal/modules/iam/delivery/http/admin_handler.go:259`, `internal/modules/iam/delivery/http/admin_handler.go:285`, `internal/modules/auth/application/service.go:279`, `internal/modules/auth/application/service.go:326`).
- Pagination: no pagination in create-user flow (`internal/modules/iam/delivery/http/admin_handler.go:259`, `internal/modules/iam/delivery/http/admin_handler.go:285`).
- Audit log emission for create-user variant: no `h.recordAudit(...)` call inside `handleCreateUser` (`internal/modules/iam/delivery/http/admin_handler.go:259`, `internal/modules/iam/delivery/http/admin_handler.go:285`).
- Audit comparison point: `handleReplaceUserRoles` does emit `h.recordAudit(...)` (`internal/modules/iam/delivery/http/admin_handler.go:398`).
- Two-transaction non-atomicity: factual; auth identity insert transaction and IAM role replacement transaction are separate and not wrapped by one transaction (`internal/modules/auth/infrastructure/postgres/repository.go:152`, `internal/modules/auth/infrastructure/postgres/repository.go:170`, `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:73`, `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:112`, `internal/modules/auth/application/service.go:305`, `internal/modules/auth/application/service.go:325`).
