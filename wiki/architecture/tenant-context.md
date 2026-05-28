# Architecture: Session-Bound Tenant Context

> **Last verified:** 2026-05-21 (spec-review convergence sync)
> **Freeze verification note (2026-05-21):** Terminology and ownership framing were re-checked during spec-review cleanup; runtime/source-of-truth claims in this doc were not expanded in this pass.
> **Scope:** `internal/platform/tenant` package; how tenant identity flows from login through every request handler; `AllowDevTenantFallback` flag; IAM legacy fallback pattern.
> **Out of scope:** per-tenant IAM role assignment (see `wiki/modules/iam.md   8.2`); row-level Postgres isolation via GUC/RLS (tracked per-module as tech debt).
> **Key files:**
> - `internal/platform/tenant/context.go:18`     `WithTenantID` (auth middleware injects here)
> - `internal/platform/tenant/context.go:24`     `FromContext` (all handlers read from here)
> - `internal/platform/tenant/context.go:14`     `ErrTenantMissing` (returned when context has no tenant)
> - `internal/platform/tenant/const.go:4`     `DevTenantID` sentinel (`ffffffff-ffff-ffff-ffff-ffffffffffff`)
> - `internal/modules/auth/application/service.go:172`     `resolveLoginTenant` (binds tenant at login time)
> - `internal/modules/auth/application/service.go:37`     `AllowDevTenantFallback` config flag
> - `internal/modules/auth/delivery/http/middleware.go:83-88`     injects `tenant.WithTenantID`; strips `X-Tenant-ID` header
> - `internal/modules/auth/domain/model.go:26`     `Session.TenantID` (persisted in `auth_sessions`)
> - `internal/modules/auth/domain/model.go:95`     `CurrentUser.TenantID` / `CurrentUser.TenantName`
> - `internal/modules/auth/domain/errors.go:18`     `ErrTenantNotPermitted`
> - `internal/modules/auth/domain/errors.go:21`     `ErrTenantClaimRequired`
> - `internal/modules/auth/infrastructure/postgres/repository.go:115`     `GetTenantByID`
> - `db/migrations/0214_tenants_master_table.sql`     canonical `metaldocs.tenants` master table + `auth_sessions.tenant_id` UUID conversion
> - `migrations/0184_auth_sessions_tenant_id.sql`     adds `tenant_id` column to `auth_sessions`
> - `migrations/0185_revoke_ambiguous_sessions.sql`     revokes pre-migration sessions without a tenant binding

---

## 1. Problem this solves

Before Plan 3, every module read the `X-Tenant-ID` HTTP request header to determine which tenant's data to read/write. This created a header-forgery vector: any authenticated user could set `X-Tenant-ID` to any UUID and access another tenant's data. It also scattered the trust decision across 10+ handler files.

Plan 3 moves tenant binding to **login time**:

1. At login, the auth service resolves which tenant the user belongs to (`resolveLoginTenant`).
2. That tenant ID is stored in `auth_sessions.tenant_id` (migration 0184; UUID-aligned in migration 0214).
3. On every subsequent request, the auth middleware reads the session's `tenant_id` and injects it into the request context via `tenant.WithTenantID`.
4. The middleware **strips** the `X-Tenant-ID` header so no downstream handler can read it.
5. All handlers call `tenant.FromContext`     they get the session-verified tenant or an error.

---

## 2. Package: `internal/platform/tenant`

### 2.1 `WithTenantID`

```go
func WithTenantID(ctx context.Context, tenantID string) context.Context
```

Returns a child context carrying `tenantID`. The **only production caller** is `auth/delivery/http/middleware.go` (after `ResolveSession` succeeds). Test helpers may also call it directly.

### 2.2 `FromContext`

```go
func FromContext(ctx context.Context) (string, error)
```

Extracts the authenticated tenant ID. Returns `ErrTenantMissing` when:
- The context has no tenant value (middleware was bypassed, or auth is disabled).
- The stored value is empty or whitespace.

**Handlers MUST treat `ErrTenantMissing` as an internal-server-error invariant violation**, not a 400 validation error     it means the auth middleware did not run or the session was corrupt.

### 2.3 `ErrTenantMissing`

Sentinel error (`errors.New("tenant: not present in context")`). No caller should expose this string to clients.

### 2.4 `DevTenantID`

```go
const DevTenantID = "ffffffff-ffff-ffff-ffff-ffffffffffff"
```

Compile-time UUID used **only** when `AllowDevTenantFallback=true` and the user has no IAM roles (local dev login). Never used as a fallback in handler paths after Plan 3.

---

## 3. Login-time tenant resolution (`resolveLoginTenant`)

`application/service.go:172`:

```go
func (s *Service) resolveLoginTenant(ctx context.Context, userID, claimedTenantID string) (string, error)
```

Rules (in order):

| Condition | Result |
|---|---|
| `claimedTenantID` non-empty AND user has a role in that tenant | `claimedTenantID` accepted |
| `claimedTenantID` non-empty but user has NO role there | `ErrTenantNotPermitted`     403 `AUTH_TENANT_FORBIDDEN` |
| `claimedTenantID` empty AND user has exactly 1 tenant | that tenant |
| `claimedTenantID` empty AND user has 0 tenants AND `AllowDevTenantFallback=true` | `DevTenantID` |
| `claimedTenantID` empty AND user has multiple tenants | `ErrTenantClaimRequired`     403 `AUTH_TENANT_REQUIRED` |

The selected tenant ID is stored in `auth_sessions.tenant_id` (migration 0184; UUID-aligned in 0214). `GetUserTenants` reads `iam_user_roles` to determine which tenants a user has any role in (`domain/port.go:23`). `GetTenantByID` resolves the session-bound tenant to a canonical `metaldocs.tenants` row so auth responses can expose `tenantName` alongside `tenantId`.

### 3.1 `AllowDevTenantFallback`

Config flag on `authapp.Config` (`service.go:37`). When `true`, login succeeds for users with no IAM rows (dev bootstrapping before roles are seeded). **Default: `false`** (prod-safe). Set to `true` only in local dev / tests.

---

## 4. Middleware injection (`auth/delivery/http/middleware.go:83-88`)

After `ResolveSession` returns a `CurrentUser`:

```go
ctx := authdomain.WithCurrentUser(r.Context(), currentUser)
ctx = iamdomain.WithAuthContext(ctx, currentUser.UserID, currentUser.Roles)
ctx = platformtenant.WithTenantID(ctx, currentUser.TenantID)
r2 := r.WithContext(ctx)
r2.Header = r2.Header.Clone()
r2.Header.Del("X-Tenant-ID")
next.ServeHTTP(w, r2)
```

Key points:
- `WithTenantID` writes the session's tenant into the request context.
- `r2.Header.Del("X-Tenant-ID")` strips the header so no downstream code can read it     even if it was present on the inbound request.
- All three context writes happen in one block; if `ResolveSession` fails, none of them run.

---

## 5. IAM legacy fallback pattern

IAM's HTTP middleware (`iam/delivery/http/middleware.go:77-83`) has a deliberate two-step fallback for the transition period:

```go
tenantID, err := tenant.FromContext(r.Context())
if err != nil {
    tenantID = strings.TrimSpace(r.Header.Get("X-Tenant-ID"))
    if tenantID == "" {
        tenantID = tenant.DevTenantID
    }
}
```

This fallback activates **only** when `tenant.FromContext` returns `ErrTenantMissing`     i.e., when auth middleware did not run (legacy-header mode where `LegacyHeaderEnabled=true` and `X-User-Id` header bypassed the auth middleware entirely). In normal production flows (auth enabled, no `LegacyHeaderEnabled`), `FromContext` always succeeds, and the header fallback is never reached.

The fallback is preserved deliberately: removing it would break `LegacyHeaderEnabled` dev flows before that bypass is fully retired (see `wiki/modules/auth-tech-debt.md` T-001).

---

## 6. Affected modules (Plan 3 sweep)

All of the following now call `tenant.FromContext` (or a thin wrapper over it) instead of reading `X-Tenant-ID` directly:

| Module | File | Helper |
|---|---|---|
| IAM | `iam/delivery/http/middleware.go:77` | direct `tenant.FromContext` (with legacy fallback) |
| IAM | `iam/delivery/http/admin_handler.go:109` | direct `tenant.FromContext` |
| IAM | `iam/delivery/http/routes_memberships.go:146` | `tenantIDFromRequest`     `tenant.FromContext` |
| controlled-documents | `internal/modules/controlleddocuments/delivery/http/routes.go:488` | `tenantIDFromRequest`     `tenant.FromContext` |
| controlled-documents | `internal/modules/controlleddocuments/delivery/http/handler.go:50` | `injectTenant` middleware     `tenant.FromContext` |
| templates | `templates/delivery/http/handler.go:83` | `tenantIDFromReq`     `tenant.FromContext` |
| taxonomy | `taxonomy/delivery/http/routes_profiles.go:230` | `tenantIDFromRequest`     `tenant.FromContext` |
| documents | `documents/delivery/http/handler.go` | `tenant.FromContext` |
| documents/approval | `documents/approval/http/handler.go` | `tenant.FromContext` |
| documents | `documents/http/fillin_handler.go` | `tenant.FromContext` |
| documents | `documents/http/placeholder_options_handler.go` | `tenant.FromContext` |
| documents | `documents/http/reconstruct_handler.go` | `tenant.FromContext` |

---

## 7. Migrations

| Migration | Effect |
|---|---|
| `0184_auth_sessions_tenant_id.sql` | Adds `tenant_id UUID NOT NULL DEFAULT '...'` column to `metaldocs.auth_sessions` |
| `0185_revoke_ambiguous_sessions.sql` | Marks all pre-existing sessions as revoked (they carry no valid `tenant_id`) |
| `0214_tenants_master_table.sql` | Creates `metaldocs.tenants`, converts `auth_sessions.tenant_id` to UUID on upgrade paths, and adds auth/IAM tenant FKs |

---

## 8. Cross-links

- `wiki/modules/auth.md`     full auth module doc;   6.1 login sequence,   6.2 resolve-session sequence,   8.7 config
- `wiki/modules/iam.md   8.2`     tenant scoping in IAM tables
- `wiki/modules/controlled-documents.md   8.7`     controlled-documents tenant scoping
- `wiki/modules/taxonomy.md   8`     taxonomy tenant scoping (T-001 resolved)
- `wiki/modules/templates-tech-debt.md#T-003`     templates header-trust (T-003 resolved)
- `wiki/concepts/authz-tiers.md`     how tenant context feeds into tier-1 and tier-2 authz
- `wiki/backlog/roadmap.md`     Plan 3 entry

