# Subagent prompt — Phase 4: Persistence map

You are a research-only Codex subagent. Output FACTS only.

## Task

For module at `internal/modules/auth/`, produce an artifact at `wiki/modules/auth/_artifacts/04-persistence.md` mapping its Postgres persistence surface.

### 1. Tables owned

Tables created by this module's migrations (or the global migration set, scoped by table-name prefix / ownership comment).

| Table | Created in (migration filename) | Notes |
|---|---|---|
| `metaldocs.auth_identities` | `migrations/0021_init_auth_identities_and_sessions.sql:1-13` | Created with `user_id` FK to `metaldocs.iam_users` in 0021, then decoupled from `iam_users` FK and extended with `display_name` + `is_active` in 0036 (`migrations/0036_decouple_auth_identity_from_iam_user_tables.sql:1-3,16-19,21-42`). |
| `metaldocs.auth_sessions` | `migrations/0021_init_auth_identities_and_sessions.sql:18-27` | Created with `user_id` FK to `metaldocs.iam_users` in 0021, then FK to `iam_users` dropped and replaced with FK to `auth_identities(user_id)` in 0036 (`migrations/0036_decouple_auth_identity_from_iam_user_tables.sql:44-65,67-88`). |

| Column | Type | Constraints (NOT NULL, FK, default) |
|---|---|---|
| `user_id` | `TEXT` | `PRIMARY KEY`; FK to `metaldocs.iam_users(user_id)` with `ON DELETE CASCADE` in 0021 (`migrations/0021_init_auth_identities_and_sessions.sql:2`); FK dropped in 0036 (`migrations/0036_decouple_auth_identity_from_iam_user_tables.sql:21-42`). |
| `username` | `TEXT` | `NOT NULL` (`migrations/0021_init_auth_identities_and_sessions.sql:3`). |
| `email` | `TEXT` | nullable (`migrations/0021_init_auth_identities_and_sessions.sql:4`). |
| `password_hash` | `TEXT` | `NOT NULL` (`migrations/0021_init_auth_identities_and_sessions.sql:5`). |
| `password_algo` | `TEXT` | `NOT NULL` (`migrations/0021_init_auth_identities_and_sessions.sql:6`). |
| `must_change_password` | `BOOLEAN` | `NOT NULL DEFAULT FALSE` (`migrations/0021_init_auth_identities_and_sessions.sql:7`). |
| `last_login_at` | `TIMESTAMPTZ` | nullable (`migrations/0021_init_auth_identities_and_sessions.sql:8`). |
| `failed_login_attempts` | `INT` | `NOT NULL DEFAULT 0` (`migrations/0021_init_auth_identities_and_sessions.sql:9`). |
| `locked_until` | `TIMESTAMPTZ` | nullable (`migrations/0021_init_auth_identities_and_sessions.sql:10`). |
| `created_at` | `TIMESTAMPTZ` | `NOT NULL DEFAULT NOW()` (`migrations/0021_init_auth_identities_and_sessions.sql:11`). |
| `updated_at` | `TIMESTAMPTZ` | `NOT NULL DEFAULT NOW()` (`migrations/0021_init_auth_identities_and_sessions.sql:12`). |
| `display_name` | `TEXT` | Added in 0036 (`migrations/0036_decouple_auth_identity_from_iam_user_tables.sql:2`), set `NOT NULL` in 0036 (`migrations/0036_decouple_auth_identity_from_iam_user_tables.sql:17`). |
| `is_active` | `BOOLEAN` | Added in 0036 (`migrations/0036_decouple_auth_identity_from_iam_user_tables.sql:3`), set `NOT NULL DEFAULT TRUE` in 0036 (`migrations/0036_decouple_auth_identity_from_iam_user_tables.sql:18-19`). |

| Column | Type | Constraints (NOT NULL, FK, default) |
|---|---|---|
| `session_id` | `TEXT` | `PRIMARY KEY` (`migrations/0021_init_auth_identities_and_sessions.sql:19`). |
| `user_id` | `TEXT` | `NOT NULL`; FK to `metaldocs.iam_users(user_id)` with `ON DELETE CASCADE` in 0021 (`migrations/0021_init_auth_identities_and_sessions.sql:20`); FK to `iam_users` dropped in 0036 (`migrations/0036_decouple_auth_identity_from_iam_user_tables.sql:44-65`); FK `fk_auth_sessions_identity` to `metaldocs.auth_identities(user_id)` with `ON DELETE CASCADE` added in 0036 (`migrations/0036_decouple_auth_identity_from_iam_user_tables.sql:82-86`). |
| `created_at` | `TIMESTAMPTZ` | `NOT NULL DEFAULT NOW()` (`migrations/0021_init_auth_identities_and_sessions.sql:21`). |
| `expires_at` | `TIMESTAMPTZ` | `NOT NULL` (`migrations/0021_init_auth_identities_and_sessions.sql:22`). |
| `revoked_at` | `TIMESTAMPTZ` | nullable (`migrations/0021_init_auth_identities_and_sessions.sql:23`). |
| `ip_address` | `TEXT` | nullable (`migrations/0021_init_auth_identities_and_sessions.sql:24`). |
| `user_agent` | `TEXT` | nullable (`migrations/0021_init_auth_identities_and_sessions.sql:25`). |
| `last_seen_at` | `TIMESTAMPTZ` | `NOT NULL DEFAULT NOW()` (`migrations/0021_init_auth_identities_and_sessions.sql:26`). |

Note on IAM table ownership and auth writes:
- `metaldocs.iam_users` is IAM-owned (`migrations/0002_init_iam_rbac.sql`, not expanded here).
- In `internal/modules/auth/infrastructure/postgres/repository.go`, `CreateUser` does not execute SQL against `metaldocs.iam_users`; it inserts only into `metaldocs.auth_identities` (`internal/modules/auth/infrastructure/postgres/repository.go:163-164`).

### 2. Tables read or written but NOT owned

Foreign tables this module touches.

| Table | Owner module | Read / Write | Operations using it |
|---|---|---|---|
| `metaldocs.iam_users` | `iam` | none in `repository.go` | no SQL reference found in `internal/modules/auth/infrastructure/postgres/repository.go` (full file reviewed). |
| `metaldocs.iam_user_roles` | `iam` | none in `repository.go` | no SQL reference found in `internal/modules/auth/infrastructure/postgres/repository.go` (full file reviewed). |

Context in auth application layer:
- `Service.CreateUser` calls `roleAdmin.ReplaceUserRoles(...)` (`internal/modules/auth/application/service.go:325`).
- `Service.BootstrapLocalAdmin` calls `roleAdmin.UpsertUserAndAssignRole(...)` (`internal/modules/auth/application/service.go:90-97`).
- These calls are outside `repository.go` and do not expose table names in this file.

### 3. Triggers, GUCs, functions

List Postgres triggers, functions, and GUCs (`SET LOCAL ...`) the module installs or relies on.

| Object | Kind (trigger / function / GUC) | File:line | Purpose |
|---|---|---|---|
| `public.enforce_capability_asserted()` | function | `migrations/0142b_role_capabilities_v2_enforce.sql:67-179` | Tripwire function for capability assertions (defined in migration). |
| `trg_require_cap_asserted_instances` | trigger | `migrations/0142b_role_capabilities_v2_enforce.sql:200-203` | Attached to `public.approval_instances` before insert. |
| `trg_require_cap_asserted_signoffs` | trigger | `migrations/0142b_role_capabilities_v2_enforce.sql:206-209` | Attached to `public.approval_signoffs` before insert. |
| `metaldocs.asserted_caps` | GUC read by function | `migrations/0142b_role_capabilities_v2_enforce.sql:52,139` | Capability assertion payload read by tripwire function. |
| `metaldocs.bypass_authz` | GUC read by function | `migrations/0142b_role_capabilities_v2_enforce.sql:53,97` | Bypass token read by tripwire function. |

Auth-table attachment check:
- No trigger attachment to `auth_identities` or `auth_sessions` is present in `migrations/0142b_role_capabilities_v2_enforce.sql`; trigger DDL targets only `public.approval_instances` and `public.approval_signoffs` (`migrations/0142b_role_capabilities_v2_enforce.sql:200-209`).

`metaldocs.actor_id` set in auth module:
- No matches for `metaldocs.actor_id` or `SET LOCAL` under `internal/modules/auth/` via `rg -n "metaldocs\.actor_id|SET LOCAL|set local|SET\s+metaldocs\.actor_id" internal/modules/auth/`.

`tenant_id` on auth tables:
- `tenant_id` column is not present in `auth_identities`/`auth_sessions` DDL shown in 0021/0036 (`migrations/0021_init_auth_identities_and_sessions.sql:1-30`; `migrations/0036_decouple_auth_identity_from_iam_user_tables.sql:1-88`).

### 4. Indexes

| Index | Table | Columns | Unique? | Purpose |
|---|---|---|---|---|
| `auth_identities` primary key (implicit index) | `metaldocs.auth_identities` | `user_id` | Yes | Primary key (`migrations/0021_init_auth_identities_and_sessions.sql:2`). |
| `uq_auth_identities_username_ci` | `metaldocs.auth_identities` | `LOWER(username)` | Yes | Case-insensitive username uniqueness (`migrations/0021_init_auth_identities_and_sessions.sql:15`). |
| `uq_auth_identities_email_ci` | `metaldocs.auth_identities` | `LOWER(email)` with predicate `WHERE email IS NOT NULL` | Yes | Case-insensitive email uniqueness for non-null emails (`migrations/0021_init_auth_identities_and_sessions.sql:16`). |
| `auth_sessions` primary key (implicit index) | `metaldocs.auth_sessions` | `session_id` | Yes | Primary key (`migrations/0021_init_auth_identities_and_sessions.sql:19`). |
| `idx_auth_sessions_user_id` | `metaldocs.auth_sessions` | `user_id` | No | Lookup by user (`migrations/0021_init_auth_identities_and_sessions.sql:29`). |
| `idx_auth_sessions_active` | `metaldocs.auth_sessions` | `user_id, expires_at DESC` with predicate `WHERE revoked_at IS NULL` | No | Active-session lookup path (`migrations/0021_init_auth_identities_and_sessions.sql:30`). |

No additional `CREATE INDEX`/`ALTER ... ADD CONSTRAINT UNIQUE` on these two tables was found in `migrations/0036_decouple_auth_identity_from_iam_user_tables.sql` or `migrations/0159_seed_dev_approver_user.sql`.

### 5. Tripwire pairing audit

For every repo method that mutates a table owned by this module:

| Method (file:line) | Authz.Require called? | Cap + area arg | SQL verb | Table |
|---|---|---|---|---|
| `CreateSession` (`internal/modules/auth/infrastructure/postgres/repository.go:43`) | NO | n/a | INSERT | `metaldocs.auth_sessions` |
| `TouchSession` (`internal/modules/auth/infrastructure/postgres/repository.go:80`) | NO | n/a | UPDATE | `metaldocs.auth_sessions` |
| `RevokeSession` (`internal/modules/auth/infrastructure/postgres/repository.go:93`) | NO | n/a | UPDATE | `metaldocs.auth_sessions` |
| `RevokeSessionsByUserID` (`internal/modules/auth/infrastructure/postgres/repository.go:106`) | NO | n/a | UPDATE | `metaldocs.auth_sessions` |
| `RecordSuccessfulLogin` (`internal/modules/auth/infrastructure/postgres/repository.go:120`) | NO | n/a | UPDATE | `metaldocs.auth_identities` |
| `RecordFailedLogin` (`internal/modules/auth/infrastructure/postgres/repository.go:136`) | NO | n/a | UPDATE | `metaldocs.auth_identities` |
| `CreateUser` (`internal/modules/auth/infrastructure/postgres/repository.go:151`) | NO | n/a | INSERT | `metaldocs.auth_identities` |
| `UpdateUser` (`internal/modules/auth/infrastructure/postgres/repository.go:252`) | NO | n/a | UPDATE | `metaldocs.auth_identities` |
| `BootstrapAdmin` (`internal/modules/auth/infrastructure/postgres/repository.go:298`) | NO | n/a | INSERT / UPDATE (`ON CONFLICT DO UPDATE`) | `metaldocs.auth_identities` |

Classification (requested scope):
- All rows above are `OUT-OF-SCOPE` for tripwire violations (auth tables explicitly outside tripwire scope per task instruction).

Method existence check:
- All requested methods were found in `internal/modules/auth/infrastructure/postgres/repository.go` at the lines listed above.

### 6. Migration history

Chronological list of migrations affecting this module.

| Order | Filename | Verb summary | Date (from filename or commit) |
|---|---|---|---|
| 1 | `0021_init_auth_identities_and_sessions.sql` | Creates `auth_identities` and `auth_sessions`; creates auth indexes (`migrations/0021_init_auth_identities_and_sessions.sql:1-30`). | `0021` (sequence in filename) |
| 2 | `0022_grant_auth_runtime_privileges.sql` | Grants `SELECT, INSERT, UPDATE` on `auth_identities` and `auth_sessions` to `metaldocs_app` (`migrations/0022_grant_auth_runtime_privileges.sql:1-2`). | `0022` |
| 3 | `0036_decouple_auth_identity_from_iam_user_tables.sql` | Adds columns to `auth_identities`; backfills; enforces not-null/default; drops FKs to `iam_users`; adds FK from `auth_sessions.user_id` to `auth_identities.user_id` (`migrations/0036_decouple_auth_identity_from_iam_user_tables.sql:1-88`). | `0036` |
| 4 | `0159_seed_dev_approver_user.sql` | Inserts seed row into `auth_identities` (plus IAM seed rows in `iam_users` and `iam_user_roles`) (`migrations/0159_seed_dev_approver_user.sql:11-27`). | `0159` |

`rg -l "auth_identities|auth_sessions" migrations/` returned:
- `migrations/0021_init_auth_identities_and_sessions.sql`
- `migrations/0022_grant_auth_runtime_privileges.sql`
- `migrations/0036_decouple_auth_identity_from_iam_user_tables.sql`
- `migrations/0159_seed_dev_approver_user.sql`

# tables owned: 2 · # tripwire violations: 0 (expected) · # migrations: 4
