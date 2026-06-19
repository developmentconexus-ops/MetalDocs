# Feature F5.7 — role-admin `iam_users` upsert drops `tenant_id` (re-audit Major #2)

> **Milestone:** 5 — HS-5 remediation · **Feature:** `f5.7-role-admin-tenant-id`
> **Status:** **APPROVED 2026-06-19** — scope inherited from `milestone.md` F5.7 row (operator
> approved milestone spec 2026-06-16). Real defect, surgical fix. No architecture decision (no ADR).

## What the milestone asked

Fix `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:61-68` and `:109-116` —
the `iam_users` upsert must include `tenant_id`; identify the correct tenant source in the call path
and pass it through. Closes re-audit 2026-06-16 **Major #2**.

## Investigation — confirmed real defect

`metaldocs.iam_users` carries a **`tenant_id uuid NOT NULL`** column with a **sentinel default**:

```sql
tenant_id uuid DEFAULT 'ffffffff-ffff-ffff-ffff-ffffffffffff'::uuid NOT NULL   -- 0001_current_schema.sql:1423
```

Both production upserts omit `tenant_id` from the INSERT column list:

```sql
INSERT INTO metaldocs.iam_users (user_id, display_name, is_active, updated_at)   -- :61-68, :109-116
VALUES ($1, $2, TRUE, NOW()) ...
```

So a **brand-new** user created via the role-admin path gets `tenant_id = 'ffff…ffff'` (the sentinel),
**not** the actual tenant — even though the caller already has the real tenant in hand:

| Caller | Tenant passed | Source |
|--------|---------------|--------|
| `auth/application/service.go:209` (bootstrap admin) | `tenant.DevTenantID` | real |
| `auth/application/service.go:615` (`ReplaceUserRoles`) | `fields.tenantID` | real |

The `tenant_id` is threaded all the way to `UpsertUserAndAssignRoleTx(... tenantID ...)` and is already
used correctly by the sibling `iam_user_roles` INSERT (`$2::uuid`) and the DELETE — it is **only**
dropped at the `iam_users` INSERT. The reference inserter `internal/test/e2e_seed.go:509` already
includes `tenant_id`, confirming the correct shape.

**Blast radius of the sentinel.** `iam_users` has `UNIQUE (tenant_id, user_id)` and a partial active
unique index, and the ADR-0031 `TenantUserReader` port resolves *tenant → member user_ids* off this
table to scope `auth_identities` reads. A row stamped with the sentinel tenant is invisible to its
real tenant's membership/display-name reads and mis-bucketed under the placeholder tenant — a
tenant-isolation correctness defect, not cosmetic.

## Consumer contract (what the fixed producer must satisfy)

1. A `iam_users` row **created** via `UpsertUserAndAssignRole(Tx)` / `ReplaceUserRoles(Tx)` carries the
   **caller-supplied `tenant_id`**, never the sentinel default.
2. The sibling `iam_user_roles` write and the authz gate are **unchanged** (already correct).
3. On `ON CONFLICT (user_id)` the row's `tenant_id` is set to the supplied tenant (`EXCLUDED.tenant_id`)
   — this both keeps an existing correct row correct and repairs any pre-existing sentinel row, with no
   new behavior for the common same-tenant re-assign path.

## Non-goals

- No change to `iam_user_roles` writes, the DELETE, or the authz `Require`/seed path.
- No DB migration, no schema change, no change to the sentinel default itself (kept for back-compat).
- No API/DTO change; no change to the `tenant_id` value the callers compute.
- No backfill of historical sentinel rows (out of scope; the on-conflict update repairs lazily on next
  upsert). Recorded as a bounded note, not a defer of F5.7.
- The global `PRIMARY KEY (user_id)` (cross-tenant uniqueness of `user_id`) is pre-existing and out of
  scope — not introduced or altered here.

## Validation Gate

1. **TDD (failing-first):** a sqlmock test asserts the `iam_users` INSERT receives **three** args
   ending in `testTenant` (`"alice", "Alice", testTenant`) — fails against HEAD (drops tenant), passes
   after fix. Applies to both `UpsertUserAndAssignRole` and `ReplaceUserRoles`.
2. **Existing exact-match tests updated** to the new arg list and pass
   (`TestUpsertUserAndAssignRole_PassesTenantID`, `TestReplaceUserRoles_DeleteThenInsert_PersistsSingleRole`).
3. **Build** — `go build ./...` clean.
4. **Tests** — `go test -count=1 ./internal/modules/iam/... ./internal/modules/auth/...` all green.
5. **Grep** — no remaining `INSERT INTO metaldocs.iam_users (user_id, display_name, is_active`
   (tenant-less column list) in `internal/` outside tests.
