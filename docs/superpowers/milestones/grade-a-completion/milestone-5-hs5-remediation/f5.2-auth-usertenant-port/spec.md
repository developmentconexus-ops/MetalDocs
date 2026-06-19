# Feature F5.2 — auth UserTenantReader port (H-G site #1)

> **Milestone:** 5 — HS-5 remediation  ·  **Feature:** `f5.2-auth-usertenant-port`
> **Status:** Approved 2026-06-16 — code change may begin.

## Consumer contract

**Consumer:** `Repository.GetUserTenants` in
`internal/modules/auth/infrastructure/postgres/repository.go:101`, ultimately serving
`auth/application/service.go:345` (tenant resolution during authenticate).

**What it needs:** the distinct tenant IDs a user holds any role in — `[]string`, sorted, no
duplicates. It currently obtains this by querying `metaldocs.iam_user_roles` **directly**
(`SELECT DISTINCT tenant_id::text FROM metaldocs.iam_user_roles WHERE user_id = $1 ORDER BY tenant_id`).
That table is owned by the **iam** module; a direct read of it from `auth/infrastructure` is the
H-G boundary violation (site #1, the one remaining `FROM metaldocs.iam_user_roles` outside `iam/`).

**Required shape after this feature:** `GetUserTenants` delegates to a narrow **iam-owned read
port** — `iamdomain.UserTenantReader` — exactly as the security module already resolves tenant
membership through `iamdomain.TenantUserReader` (M4/F4.5, ADR 0031). The port's method:

```go
// UserTenantReader — iam-owned. Inverse of TenantUserReader (user→tenants vs tenant→users).
UserTenantIDs(ctx context.Context, userID string) ([]string, error)
```

Returns the **distinct, sorted** tenant IDs the user has roles in; empty (non-nil) slice when the
user has no roles. The Postgres impl reproduces the replaced query **byte-for-semantics**
(DISTINCT, same WHERE, same ORDER BY) and reads the **connection pool**, never a caller tx
(H-PRE-1). A `NoopUserTenantReader` null-object returns an empty slice.

## Interview record (B1.5 — resolved by investigation, no operator questions needed)

| Question | Resolution | Source |
|----------|-----------|--------|
| Port name + shape? | `UserTenantReader.UserTenantIDs(ctx, userID) ([]string, error)` | milestone.md F5.2 row; inverse of existing `TenantUserReader` |
| Where does the port live? | `internal/modules/iam/domain/user_tenant_reader_port.go` (+ Noop) | mirrors `tenant_user_reader_port.go` |
| Postgres impl location? | `internal/modules/iam/infrastructure/postgres/user_tenant_repository.go` | mirrors `tenant_user_repository.go` |
| Exact SQL (parity)? | `SELECT DISTINCT tenant_id::text FROM metaldocs.iam_user_roles WHERE user_id = $1 ORDER BY tenant_id` — identical to the replaced query | `repository.go:102-107` |
| Off-tx (H-PRE-1)? | Yes — pool read only (`*sql.DB`), no caller tx | quality goal #4 |
| How is it wired? | `bootstrap/api.go:80` passes `iampg.NewUserTenantRepository(db)` into `authpg.NewRepository`; nil→Noop default | mirrors security wiring |
| 6 auth-pg test callers of `NewRepository(db)`? | none exercise postgres `GetUserTenants` (only memory repo tests do) → pass `nil` (→ Noop), no behavior change | grep of `*_test.go` |
| Import cycle? | none — `iam/domain` does not import `auth` | grep |

## HS-2 boundary check

This adds **one narrow read port** mirroring four existing iam ports (display-name, tenant-user,
admin-role, mfa). It is **not** a shared IAM-API redesign. If the work were to require changing the
iam role/membership write model or the auth↔iam contract surface, **HS-2 fires — stop and report.**
It does not: the change is additive and local.

## Non-goals

- No change to the memory auth repository (`infrastructure/memory`) — it keeps its own
  `GetUserTenants`/`SeedUserTenants` for tests.
- No new product capability, no HTTP/route/frontend change, no schema/migration change.
- `auth/domain/port.go`'s `Repository.GetUserTenants` signature is unchanged (still on the auth
  repo interface) — only its Postgres implementation changes.
- Do not alter the `tenant_id` default or any `iam_user_roles` semantics.

## Validation Gate

1. **H-G grep → 0:**
   `grep -rn "FROM metaldocs\.iam_user_roles" --include="*.go" internal/modules/ | grep -v "internal/modules/iam/" | grep -v "_test\.go"`
   → 0 matches.
2. **Parity — unit:** a white-box test pins the new port's query (contains `DISTINCT`,
   `FROM metaldocs.iam_user_roles`, `WHERE user_id = $1`, `ORDER BY tenant_id`).
3. **Parity — live:** `//go:build integration` test mirroring
   `tenant_user_repository_integration_test.go`: a user with two roles in tenant A + one in tenant
   B → `UserTenantIDs` = `[A, B]` (distinct, sorted); another user's tenant excluded; unknown user
   → empty non-nil slice. Run live if a DB is reachable; otherwise recorded as skipped (labeled).
4. `go build ./...` clean.
5. `go test -count=1 ./internal/modules/auth/... ./internal/modules/iam/...` green.
