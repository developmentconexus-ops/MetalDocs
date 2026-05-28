# Module #4 — `internal/modules/iam` — Review Findings

**Date:** 2026-05-22
**Reviewers:** go-reviewer, security-reviewer, silent-failure-hunter, type-design-analyzer, database-reviewer (all Sonnet 4.6)
**LoC:** ~3377 across 28 files
**Raw:** 9C / 18H / 20M / 12L → deduped: **8C / 15H / 18M / 11L**

## Scope

All hand-written Go under `internal/modules/iam/`: `application/`, `authz/`, `delivery/http/`, `domain/`, `infrastructure/{memory,postgres}/`, `integration_test.go`.
Migrations: `0002`, `0125`, `0130`, `0136`, `0162`.
Excluded: generated `api/api.gen.go` (none in this module).

---

## Critical

### C1 — `delivery/http/middleware.go:68,81-82` — Legacy `X-User-Id` delete + re-read on same canonical header

```go
r.Header.Del("X-User-ID")          // line 68 — strips X-User-Id (same canonical key)
// ...
userID = r.Header.Get("X-User-Id") // line 82 — reads back the header just deleted
```

Go's `textproto.CanonicalMIMEHeaderKey` normalises both `X-User-ID` and `X-User-Id` to `X-User-Id`. `Del` on line 68 removes the exact key read on line 82. The `legacyHeader` branch therefore always returns `""` — a dead code bypass. But the code path still exists, was wired as a constructor bool, and the auth module's mirror bug was confirmed live. Either this was a working bypass that got accidentally disabled by the `Del`, or the Del is the intended hardening and the re-read is dead code. Either way: the legacyHeader branch must be removed — it is an ambiguous auth boundary.

Recommend: delete lines 81-83 (`if userID == "" && m.legacyHeader { ... }`) and the `legacyHeader` field from the struct and constructor. `Del` on line 68 stays as a hardening measure.

**Fix branch:** `fix/iam-4-middleware-c1-c3-c4-c7-c8` (land second)

---

### C2 — `infrastructure/postgres/role_admin_repository.go:113-121` — `ReplaceUserRoles` silently keeps only last role

```go
var lastRole string
for _, role := range roles {
    if code := strings.TrimSpace(string(role)); code != "" {
        lastRole = code
    }
}
// only lastRole written to DB
```

Function signature accepts `[]domain.Role`; the handler validates and returns the full slice in its response. A caller sending `["editor","system_admin"]` sees both in the response but only `"system_admin"` is stored. An attacker with `system_admin` rights can send roles in any order to ensure the highest-privilege role lands last. The in-memory repo (`memory/role_admin_repository.go`) inserts all roles correctly — the two implementations diverge on the interface contract.

Recommend: enforce single-role at the handler (`len(roles) > 1 → 400`) and change the signature to `role domain.Role`, matching the actual schema constraint. Never silently truncate authz writes.

**Fix branch:** `fix/iam-4-replace-roles-c2` (land third)

---

### C3 — `delivery/http/middleware.go:70-73` — Nil resolver is fail-open (every route passes unauthenticated)

```go
if m.resolver == nil {
    next.ServeHTTP(w, r)
    return
}
```

If `WithPermissionResolver` is never called, the middleware short-circuits and every request reaches the handler with no capability check, no role resolution, no identity in context. Default-allow on an authz boundary.

Recommend: fail closed — return `500 INTERNAL_ERROR` via `problem.Write` when `resolver == nil` and middleware is enabled. Or `panic("iam: permission resolver not configured")` at `NewMiddleware` construction time.

**Fix branch:** `fix/iam-4-middleware-c1-c3-c4-c7-c8`

---

### C4 — `delivery/http/middleware.go:74-78` — `VisibilitySessionRequired` treated identically to `VisibilityPublic`

```go
capability, visibility := m.resolver(r.Method, r.URL.Path)
if visibility != VisibilityPermissionGuarded {
    next.ServeHTTP(w, r)
    return
}
```

Both `VisibilityPublic` and `VisibilitySessionRequired` fall into the `!= VisibilityPermissionGuarded` branch and pass through unconditionally. No session presence check is performed for `SessionRequired` routes — an unauthenticated request is indistinguishable from an authenticated one at this layer.

Recommend: add branch before the unconditional pass-through:
```go
if visibility == VisibilitySessionRequired {
    if userID, _ := domain.UserIDFromContext(r.Context()); userID == "" {
        _ = problem.Write(w, problem.New(http.StatusUnauthorized, "UNAUTHENTICATED", "session required"))
        return
    }
}
```

**Fix branch:** `fix/iam-4-middleware-c1-c3-c4-c7-c8`

---

### C5 — `infrastructure/postgres/user_area_repository.go:103` — `CloseActive` discards `RowsAffected`, silent no-op on missing row

```go
if _, err := tx.ExecContext(ctx, q, userID, tenantID, areaCode, effectiveTo); err != nil {
    return fmt.Errorf("close active user process area: %w", err)
}
return tx.Commit()
```

If the active membership row doesn't exist (already closed, wrong tenant, concurrent revoke), zero rows are updated, the transaction commits, and the caller's `Revoke` path records a successful audit event for a no-op. `GrantAtomic` at lines 143-149 correctly checks `RowsAffected` — `CloseActive` does not.

Recommend: capture result and return `domain.ErrMembershipNotFound` on `rowsAffected == 0`, matching the pattern at lines 143-149.

**Fix branch:** `fix/iam-4-area-repo-c5-h4` (land fourth)

---

### C6 — `infrastructure/postgres/role_provider.go:20-28` — `checkUserSQL` missing `tenant_id` filter (cross-tenant auth bypass)

```sql
SELECT is_active FROM metaldocs.iam_users WHERE user_id = $1
```

Migration `0130` added `tenant_id` and `deactivated_at` to `iam_users` for per-tenant deactivation, but `checkUserSQL` still reads the global `is_active` column with no `tenant_id` predicate. A user deactivated in tenant A but active in tenant B passes this check for any tenant's query, and vice-versa. Cross-tenant auth bypass.

Recommend: replace with `WHERE user_id = $1 AND tenant_id = $2::uuid AND deactivated_at IS NULL` and pass `tenantID` (already in scope in `RolesByUserID`) as the second parameter.

**Fix branch:** `fix/iam-4-role-provider-c6-h5` (land first — contained to one file, highest blast-radius)

---

### C7 — `delivery/http/middleware.go:91-95` — Tenant identity falls back to client-supplied header, then `DevTenantID`

```go
tenantID, err := tenant.FromContext(r.Context())
if err != nil {
    tenantID = strings.TrimSpace(r.Header.Get("X-Tenant-ID"))
    if tenantID == "" {
        tenantID = tenant.DevTenantID
    }
}
```

Three fallback tiers: session context → client header → hardcoded dev UUID. Any authenticated user can inject a `X-Tenant-ID` header to scope their capability check against a different tenant's role table. The `DevTenantID` fallback means a misconfigured request silently becomes a dev-tenant request rather than a 401.

Recommend: fail on missing context — `return 401/500` if `tenant.FromContext` fails. Tenant identity must come from the verified session only.

**Fix branch:** `fix/iam-4-middleware-c1-c3-c4-c7-c8`

---

### C8 — `delivery/http/middleware.go:97-99` — Nil `caps` silently skips capability check

```go
if m.caps != nil {
    if err := m.caps.CanDo(r.Context(), userID, tenantID, string(capability)); err != nil {
        // return 403
    }
}
```

When `m.caps` is nil the entire tier-1 capability check is skipped. Any request with a valid session reaches the handler regardless of permissions. Default-allow fallback on an authz boundary.

Recommend: treat nil `caps` as misconfiguration — `panic` at `NewMiddleware` or return `500` at runtime before the check. Never silently open access on nil.

**Fix branch:** `fix/iam-4-middleware-c1-c3-c4-c7-c8`

---

## High

### H1 — `authz/authz.go:87-88` — Capability query error returned unwrapped, loses diagnostic context

```go
err = tx.QueryRowContext(...).Scan(&granted)
if err != nil {
    return err   // bare return — no fmt.Errorf wrapping
}
```

Every other error path in `Require` wraps with `fmt.Errorf("authz: ...: %w", err)`. This bare `return err` makes a DB failure on the capability scan indistinguishable from other authz errors in logs.

Recommend: `return fmt.Errorf("authz: capability check: %w", err)`.

---

### H2 — `delivery/http/admin_handler.go:171` — `_ = json.Unmarshal` silently discards payload parse failure

```go
_ = json.Unmarshal([]byte(item.PayloadJSON), &payload)
```

A corrupt `PayloadJSON` produces an empty `map[string]any{}` in the response with no log. Silent data loss in the admin audit overview.

Recommend: log parse failure with event ID — `log.Printf("iam admin: unmarshal payload event %s: %v", item.ID, err)`. Do not return an error; partial payload degradation is acceptable, silence is not.

---

### H3 — `delivery/http/admin_handler.go:461-484` — `recordAudit` has two silent early returns before write

```go
tenantID, err := tenant.FromContext(r.Context())
if err != nil {
    return   // no log
}
payloadJSON, err := json.Marshal(payload)
if err != nil {
    return   // no log
}
```

If context is mis-wired or payload is un-marshallable, the audit write is skipped entirely with no trace in logs. Compliance and forensic failure for an authz-mutation audit trail.

Recommend: add `log.Printf("iam audit: missing tenant for %s: %v", action, err)` and `log.Printf("iam audit: marshal payload for %s: %v", action, err)` respectively.

---

### H4 — `infrastructure/postgres/user_area_repository.go:28,99,126,181` — `tenant_id::text = $2` casts column, destroys index use

All four queries in `user_area_repository.go` use `tenant_id::text = $2` where `$2` is a `string`. The column is `UUID NOT NULL`. Casting the stored column to text forces full table scans — the partial unique index `ux_user_process_areas_one_active` and any tenant-scoped index are unusable. `GrantAtomic`'s INSERT correctly uses `$2::uuid`.

Recommend: change all `tenant_id::text = $2` predicates to `tenant_id = $2::uuid` (cast the parameter, not the column), matching the pattern already used in `Insert:67`.

**Fix branch:** `fix/iam-4-area-repo-c5-h4`

---

### H5 — `infrastructure/postgres/role_provider.go:27` — `err == sql.ErrNoRows` direct equality

```go
if err == sql.ErrNoRows {
```

Same pattern as 4 sites in the `auth` module review. `user_area_repository.go:192` in the same module uses `errors.Is` correctly.

Recommend: `errors.Is(err, sql.ErrNoRows)`.

**Fix branch:** `fix/iam-4-role-provider-c6-h5`

---

### H6 — `authz/authz.go:99-102` — `BypassSystem` exported with no caller identity check

```go
func BypassSystem(ctx context.Context, tx *sql.Tx) error {
    _, err := tx.ExecContext(ctx, "SELECT set_config('metaldocs.bypass_authz', 'scheduler', true)")
    return err
}
```

Any code importing `authz` can suppress the Postgres authz tripwire with no type-level proof of authority. A developer adding a new feature can call `BypassSystem` accidentally and the DB-level trigger is silenced at runtime.

Recommend: move to an unexported function with a single exported bridge accessible only from the scheduler infrastructure package, or introduce an opaque `BypassToken` with a private constructor.

**Fix branch:** `fix/iam-4-authz-h6-h10` (land fifth)

---

### H7 — `delivery/http/routes_memberships.go:42-45` — BOLA: any authenticated user can read any other user's memberships

```go
userID := strings.TrimSpace(r.URL.Query().Get("userId"))
if userID == "" {
    userID = strings.TrimSpace(authenticatedActor(r))
}
```

No check that the caller is the queried user or holds `CapMembershipManage`. Any authenticated user can enumerate area memberships for any other tenant user by supplying `?userId=<victim>`.

Recommend: after resolving `userID`, check `userID == authenticatedActor(r) || caps.CanDo(ctx, actorID, tenantID, CapMembershipManage) == nil`; return `403` otherwise.

**Fix branch:** `fix/iam-4-bola-membership-h7` (land sixth)

---

### H8 — `delivery/http/admin_handler.go:347-349,386-388` — Caller-supplied `assignedBy` poisons audit trail

```go
assignedBy := strings.TrimSpace(req.AssignedBy)
if assignedBy == "" {
    assignedBy = authenticatedActor(r)
}
```

A `system_admin` caller can write any string into the audit trail attribution for role grants and replacements.

Recommend: remove `AssignedBy` from `UpsertUserRoleRequest` and `ReplaceUserRolesRequest`; always derive from `authenticatedActor(r)`.

---

### H9 — `delivery/http/admin_handler.go:121` — `ListOnlineUsers` has no tenant filter (cross-tenant presence leak)

```go
onlineUsers, err := h.authService.ListOnlineUsers(r.Context(), activeSince)
```

No `tenantID` parameter. In multi-tenant deployment this returns online users across all tenants to any caller reaching this endpoint.

Recommend: add `tenantID` as parameter to `ListOnlineUsers`; pass from context (already resolved at line 110).

---

### H10 — `authz/authz.go:104-130` — `capCache` key excludes actor+tenant, cross-actor cache hit possible

```go
func cacheKey(capability, areaCode string) string {
    return capability + "\x00" + areaCode
}
```

Cache is request-scoped but key encodes neither `actorID` nor `tenantID`. If a context is re-used across actor identity transitions in a single request (impersonation path, middleware ordering issue), a grant for actor A is returned as a hit for actor B.

Recommend: include actorID and tenantID in the key: `actorID + "\x00" + tenantID + "\x00" + capability + "\x00" + areaCode`.

**Fix branch:** `fix/iam-4-authz-h6-h10`

---

### H11 — `0162_iam_user_roles_tenant_id.sql` — Sentinel DEFAULT `'ffffffff-...'` never dropped

Migration adds `tenant_id UUID NOT NULL DEFAULT 'ffffffff-ffff-ffff-ffff-ffffffffffff'`. Migration `0130` documented a "Phase 5: drop DEFAULT" step for `iam_users`; no equivalent migration exists for `iam_user_roles`. Any INSERT that omits `tenant_id` silently assigns the sentinel UUID instead of failing.

Recommend: add a migration `ALTER TABLE metaldocs.iam_user_roles ALTER COLUMN tenant_id DROP DEFAULT` now that all INSERT paths supply `tenant_id` explicitly.

---

### H12 — `iam_user_roles` missing composite index on `(tenant_id, user_id)`

`0162` adds `tenant_id` but no index. Every `RolesByUserID` call queries `WHERE user_id = $1 AND tenant_id = $2::uuid`. The PK `(user_id, role_code)` does not cover the `tenant_id` filter.

Recommend: add `CREATE INDEX IF NOT EXISTS idx_iam_user_roles_tenant_user ON metaldocs.iam_user_roles (tenant_id, user_id)`.

---

### H13 — `application/cached_role_provider.go:44-58` — TOCTOU: invalidate + in-flight write-back race

Read-then-write with no lock between them. Two goroutines both find a stale entry, both call `base.RolesByUserID`, and the second write-back after `InvalidateUser` fires restores stale roles until TTL expiry.

Recommend: use `golang.org/x/sync/singleflight` for the miss path to collapse concurrent fetches and prevent post-invalidate write-back.

---

### H14 — `infrastructure/postgres/user_area_repository.go:69` — `Insert` surfaces raw Postgres constraint violation, no domain sentinel

Plain INSERT without `ON CONFLICT`. Unique constraint violation surfaces as a raw `pq.Error` with error code `23505` rather than a structured domain sentinel.

Recommend: catch `pq.Error.Code == "23505"` and return `domain.ErrMembershipAlreadyActive`.

---

### H15 — `user_process_areas` missing index for time-range queries in `ListActive`

`ListActive` queries `effective_from <= $3 AND (effective_to IS NULL OR effective_to > $3)`. The partial unique index `ux_user_process_areas_one_active` only covers `WHERE effective_to IS NULL` — unusable for historical range queries.

Recommend: `CREATE INDEX IF NOT EXISTS ix_user_process_areas_user_tenant ON user_process_areas (user_id, tenant_id, effective_from DESC)`. Fix C2's cast issue first so the column is indexable.

---

## Medium

| ID | Location | Finding |
|----|----------|---------|
| M1 | `cached_role_provider.go:57` | No TTL eviction sweep — `items` map grows without bound in multi-tenant deployments |
| M2 | `dev_role_provider.go:22` | Ignores `tenantID` — cross-tenant role leak in dev/staging; needs build-tag guard or startup assertion |
| M3 | `admin_handler.go:474` | Dual `time.Now()` for `ID` and `OccurredAt` — non-unique collision risk; use UUID + single `now` capture |
| M4 | `cached_role_provider.go` | Cache not invalidated on `ReplaceUserRoles`/`UpsertUserAndAssignRole` — stale roles for up to TTL after demotion |
| M5 | `0125:42` | `ix_governance_events_resource` lacks `tenant_id` — cross-tenant resource collision in range queries |
| M6 | `0162` | Migration not wrapped in transaction — partial-apply risk on interrupted DDL |
| M7 | `role_provider.go:60-63` | Empty-roles case returns `domain.ErrUserNotFound` → 404; should be `ErrNoRoleAssigned` → 403 or 400 |
| M8 | `0002:12` | `iam_user_roles` FK `ON DELETE CASCADE` from `iam_users` — hard-delete silently purges all roles; add no-delete trigger matching `user_process_areas` pattern |
| M9 | `domain/port.go:11` | `RoleAdminRepository` god-interface: mixes bootstrap read (`HasAnyRole`) with lifecycle writes; split into `RoleBootstrapQuery` + `RoleAdminRepository` |
| M10 | `admin_handler.go:38-65`, `area_membership_service.go:55` | Three copies of role validation switch — add `domain.ParseRole(s string) (Role, error)` as single source |
| M11 | `admin_handler.go` (28 sites) | `_ = problem.Write(...)` discards write errors throughout; make `problem.Write` return void or log errors |
| M12 | `application/capability_service.go:13` | Takes `*sql.DB` directly — breaks hexagonal port; extract `CapabilityChecker` interface or move to infra layer |
| M13 | `authz.go:132-157` | `appendAssertedCap` O(n²): full JSON unmarshal+marshal+set_config per `Require` call in a single request |
| M14 | `area_membership_service.go:56-108` | Duplicated `UserProcessArea` construction block in both branches of existing-membership check |
| M15 | `admin_service.go:38` | `UpsertUserAndAssignRole` has no role-value validation in service layer; relies solely on handler validation |
| M16 | `admin_handler.go:191-216` | Manual `strings.Split(path, "/")` router predates Go 1.22 pattern syntax used in `routes_memberships.go` |
| M17 | `middleware.go:34` | Unknown `Visibility` values fall through to "not guarded" silently; add default-close for unknown values |
| M18 | `0136` | FK targets non-PK unique index `ux_iam_users_tenant_user`; fragile if index is rebuilt — document intent |

---

## Low

| ID | Location | Finding |
|----|----------|---------|
| L1 | `domain/context.go` vs `authz/authz.go` | Context key style inconsistency: `type string` vs `struct{}` — standardise on `struct{}` |
| L2 | `admin_service.go:62` | `len(roles) == 0` returns `ErrUserNotFound` (404) instead of `ErrInvalidArgument` (400) |
| L3 | `role_provider.go:55` | DB `roleCode` cast to `domain.Role` without validating against known constants — add `domain.ParseRole` guard at infrastructure boundary |
| L4 | `authz/authz.go:99` | `BypassSystem` error returned unwrapped — add `fmt.Errorf("authz: set bypass: %w", err)` |
| L5 | `0125` | `user_process_areas` PK excludes `tenant_id` — cross-tenant collision possible with same `(user_id, area_code, effective_from)` |
| L6 | `0003` | `chk_iam_user_roles_role_code` hardcodes role values — verify `0166` updated constraint; document two-table role taxonomy |
| L7 | `0002` | `assigned_by TEXT` nullable, no FK on `iam_user_roles` — add `NOT NULL` + sentinel for existing NULLs |
| L8 | `admin_handler.go:502-505` | `parseRoles` conflates empty-input with invalid-role — return distinct error strings |
| L9 | Admin endpoints | No visible rate limiting on mutation routes — confirm gateway-level throttle is in place |
| L10 | `domain/context.go:12` | `WithAuthContext` accepts empty `userID` silently — add guard or typed `UserID` parameter (pending C1 type refactor) |
| L11 | Type design (all boundaries) | `UserID`, `TenantID`, `AreaCode` are naked `string` throughout port interfaces — typed IDs would make transposition a compile error; large refactor, correct bar for an authz module |

---

## G3 Handoff — 8 Criticals

All Criticals require owner + ETA + fix-branch before cursor advances.

| ID | File:line | Severity | Owner | ETA | Fix branch | Status |
|----|-----------|----------|-------|-----|------------|--------|
| 4-C6 | `role_provider.go:20-28` no tenant_id filter | Critical | leandrotca | TBC | `fix/iam-4-role-provider-c6-h5` | Backlog (land first) |
| 4-C1 | `middleware.go:68,81-82` legacy header dead-or-live bypass | Critical | leandrotca | TBC | `fix/iam-4-middleware-c1-c3-c4-c7-c8` | Backlog (land second) |
| 4-C3 | `middleware.go:70-73` nil resolver fail-open | Critical | leandrotca | TBC | `fix/iam-4-middleware-c1-c3-c4-c7-c8` | Backlog |
| 4-C4 | `middleware.go:74-78` SessionRequired no session check | Critical | leandrotca | TBC | `fix/iam-4-middleware-c1-c3-c4-c7-c8` | Backlog |
| 4-C7 | `middleware.go:91-95` tenant_id falls back to header + DevTenantID | Critical | leandrotca | TBC | `fix/iam-4-middleware-c1-c3-c4-c7-c8` | Backlog |
| 4-C8 | `middleware.go:97-99` nil caps skips capability check | Critical | leandrotca | TBC | `fix/iam-4-middleware-c1-c3-c4-c7-c8` | Backlog |
| 4-C2 | `role_admin_repository.go:113-121` last-role-wins silent privilege escalation | Critical | leandrotca | TBC | `fix/iam-4-replace-roles-c2` | Backlog (land third) |
| 4-C5 | `user_area_repository.go:103` CloseActive no RowsAffected | Critical | leandrotca | TBC | `fix/iam-4-area-repo-c5-h4` | Backlog (land fourth) |

**Fix branch land order:** `fix/iam-4-role-provider-c6-h5` → `fix/iam-4-middleware-c1-c3-c4-c7-c8` → `fix/iam-4-replace-roles-c2` → `fix/iam-4-area-repo-c5-h4` → `fix/iam-4-authz-h6-h10` → `fix/iam-4-bola-membership-h7`

### Cascade notes

- `fix/iam-4-middleware-c1-c3-c4-c7-c8`: C1+C3+C4+C7+C8 all in `middleware.go` — single coordinated rewrite; cascades H10 (tenant in cache key), M17 (Visibility default-close)
- `fix/iam-4-replace-roles-c2`: cascades H2 (`_ = json.Unmarshal` in same handler file), H8 (remove `assignedBy` from request), M10 (role validation centralisation)
- `fix/iam-4-area-repo-c5-h4`: cascades H4 (tenant_id cast fix, same file), H14 (domain sentinel for constraint), H15 (index migration)
- `fix/iam-4-authz-h6-h10`: cascades L4 (BypassSystem error wrap), M13 (appendAssertedCap O(n²))
- `fix/iam-4-bola-membership-h7`: standalone

---

## Module Assessment

IAM is significantly better-structured than `auth` (hexagonal discipline, `GrantAtomic` correctness, `defer tx.Rollback()` pattern consistent, `authz.Require` is default-deny). The critical count (8) matches `auth`, but the character differs: `auth` had credential-pipeline exploits; IAM's criticals are mainly middleware misconfiguration hazards (fail-open defaults) and one cross-tenant auth bypass (`checkUserSQL`). The middleware cluster (C1/C3/C4/C7/C8) is the highest-priority bundle — five distinct ways the authz middleware can be open-passed.
