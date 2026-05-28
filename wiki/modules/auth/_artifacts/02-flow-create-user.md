# Admin User Creation Data Flow (`CreateUser`)

- Last verified date: 2026-05-10
- Status: Research

## 1. Entry point
- Entry point in auth HTTP layer: **n/a** for this operation. In the IAM HTTP layer, `POST /api/v1/iam/users` dispatches to `handleCreateUser` (`internal/modules/iam/delivery/http/admin_handler.go:83`, `internal/modules/iam/delivery/http/admin_handler.go:88`, `internal/modules/iam/delivery/http/admin_handler.go:92`, `internal/modules/iam/delivery/http/admin_handler.go:93`).
- Exact call site to auth application service: `h.authService.CreateUser(...)` in `handleCreateUser` (`internal/modules/iam/delivery/http/admin_handler.go:259`, `internal/modules/iam/delivery/http/admin_handler.go:280`).

## 2. Call chain
- `handleCreateUser` parses request, roles, tenant, actor, then calls `authService.CreateUser` (`internal/modules/iam/delivery/http/admin_handler.go:265`, `internal/modules/iam/delivery/http/admin_handler.go:270`, `internal/modules/iam/delivery/http/admin_handler.go:275`, `internal/modules/iam/delivery/http/admin_handler.go:279`, `internal/modules/iam/delivery/http/admin_handler.go:280`).
- `Service.CreateUser` delegates to `CreateUserWithInput`, which normalizes the input, validates the password, and hashes it before persistence.
- `hashPassword` uses `bcrypt.GenerateFromPassword(..., bcrypt.DefaultCost)` (`internal/modules/auth/application/service.go:431`, `internal/modules/auth/application/service.go:432`).
- On the canonical postgres path, auth opens a transaction via `BeginTx`, calls `CreateUserTx`, then reuses the same transaction for IAM `ReplaceUserRolesTx` before commit.
- Fallback adapters still call `s.repo.CreateUser(...)` followed by `s.roleAdmin.ReplaceUserRoles(...)` sequentially when tx-aware interfaces are unavailable.

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
- Audit log emission for create-user variant is handled in IAM `handleCreateUser` after the auth service returns successfully.
- Audit comparison point: `handleReplaceUserRoles` does emit `h.recordAudit(...)` (`internal/modules/iam/delivery/http/admin_handler.go:398`).
- Transactionality note: the postgres shared-tx path is now atomic across auth identity insert and IAM role replacement; residual non-atomicity remains only for fallback adapters that lack tx-aware interfaces.
