# Group B — Authz Cleanup Design

> **Status:** approved 2026-05-03
> **Scope:** Complete the authz unification started in `docs/superpowers/plans/2026-05-02-iam-rbac-unification.md`. Fix the 6 follow-up bugs B1-B6 from `wiki/bugs/audit-2026-05-03.md`.
> **Out of scope:** Document visibility filtering (deferred), IAM admin frontend UI flows.

---

## Why This Spec Exists

Yesterday's IAM unification plan (2026-05-02) unified the **HTTP middleware path** but left the **in-transaction path** (`authz.Require`) reading the legacy `user_process_areas` table. It also introduced new constraints (`UNIQUE(tenant_id, user_id)`, `tenant_id` column) without propagating them through the Go repository or provider layer. Six bugs surfaced during deep audit (B1-B6).

This spec formalises a **two-tier authz contract** instead of completing the unification, because the domain genuinely needs two tiers: tenant-level capabilities and area-scoped grants. Killing `user_process_areas` would conflate distinct concerns. Yesterday's "one model" goal was wrong about the requirements.

---

## Architecture: Two-Tier Authz Contract

| Tier | Table | Service | Question |
|---|---|---|---|
| 1. Tenant capability | `iam_user_roles` JOIN `role_capabilities` | `CapabilityService.CanDo(ctx, userID, tenantID, cap)` | "Can user X do `doc.create` in tenant T?" |
| 2. Area grant | `user_process_areas` JOIN `role_capabilities` | `authz.Require(ctx, tx, cap, areaCode)` | "Can user X sign for area QA-01?" |

**Invariant:** signoff/area-scoped actions require BOTH tiers pass. Tenant-only actions need only tier 1.

**system_admin bypass:** applies to BOTH tiers. Tier 1 bypass already in `CapabilityService.CanDo`. Tier 2 bypass already in `authz.Require` (added 2026-05-02).

**Boundary documentation:**
- ADR `wiki/decisions/0007-two-tier-authz.md` — formal record of the two-tier model
- Doc comments at `CapabilityService.CanDo` and `authz.Require` cross-referencing the ADR

---

## Per-Bug Fix Design

### B1 — ReplaceUserRoles 500 (ON CONFLICT mismatch + missing tenant_id)

**File:** `internal/modules/iam/infrastructure/postgres/role_admin_repository.go`

**Problem:**
- Migration 0166 added `UNIQUE(tenant_id, user_id)`. Repo `INSERT ... ON CONFLICT (user_id, role_code)` no longer matches the unique constraint.
- All write methods (`UpsertUserAndAssignRole`, `ReplaceUserRoles`) insert without `tenant_id`, relying on the default UUID — breaks any multi-tenant deployment.

**Fix:**
- Add `tenantID string` parameter to: `UpsertUserAndAssignRole`, `ReplaceUserRoles`, `HasAnyRole`.
- Recognise that `UNIQUE(tenant_id, user_id)` constraint enforces **one role row per user per tenant**. The existing `ReplaceUserRoles` accepts `[]domain.Role` but only the last one survives the unique constraint. Two options:
  - **Option A (preferred):** redefine `ReplaceUserRoles` semantics — accept slice but treat as set, write only the most-recent/highest-priority role; document this constraint at the interface.
  - **Option B:** shrink interface to `SetUserRole(ctx, userID, tenantID, role, assignedBy)` accepting a single role; deprecate `ReplaceUserRoles`.
- Implementation: inside tx, `DELETE FROM iam_user_roles WHERE tenant_id=$1 AND user_id=$2; INSERT ...`. No `ON CONFLICT` needed.
- All writes pass `tenant_id` explicitly. The `iam_users` upsert keeps its own `ON CONFLICT (user_id)` clause (different table, unchanged unique key).
- Memory impl mirrors signatures. Cascade callers: `auth/application/service.go`, `iam/delivery/http/admin_handler.go`, bootstrap.

**Decision needed:** Option A vs Option B. Default to A (less caller churn) unless the writing-plans phase finds B simpler.

**Test:** sqlmock unit (`TestReplaceUserRoles_PassesTenantID`, `TestReplaceUserRoles_DeleteThenInsert`, `TestReplaceUserRoles_OnlyLastRoleSurvives`) + integration (`TestReplaceUserRoles_TwoCallsSameTenantNoError`).

---

### B2 — Two authz models not synced (formalisation)

**Files:**
- New: `wiki/decisions/0007-two-tier-authz.md`
- Modify: `internal/modules/iam/application/capability_service.go` (doc comment)
- Modify: `internal/modules/iam/authz/authz.go` (doc comment)

**Problem:** No documented contract distinguishing the two services. Engineers conflate them. IAM admin assigning a tenant role does not grant area access; not communicated.

**Fix:** Write ADR. Add cross-reference doc comments. No code logic changes.

**Test:** wiki lint asserts ADR exists. No Go test.

---

### B3 — Dev approver user becomes system_admin

**File:** `migrations/0170_dev_approver_role_correction.sql` (new)

**Problem:** Migration 0159 seeded dev user `approver` with role `admin`. Migration 0166 blanket renamed `admin → system_admin`, catching the dev approver. Both `admin-local` and `approver` dev users are now `system_admin`. SoD never exercised.

**Fix:** New idempotent migration:
```sql
UPDATE metaldocs.iam_user_roles
   SET role_code = 'approver'
 WHERE user_id = 'approver'
   AND role_code = 'system_admin';
```

Wiki update: `wiki/references/dev-credentials.md` notes the corrected role.

**Test:** integration test seeds approver with `system_admin`, runs migration, asserts role is `approver`. Idempotency check (run twice).

---

### B4 — GUC strict-mode 42704

**Files:**
- Modify: `internal/modules/iam/authz/authz.go`

**Problem:** `current_setting('metaldocs.actor_id', false)` (strict) panics with PostgreSQL error 42704 when GUC is unset. Only approval/document modules set the GUC. Any caller outside those modules invoking `authz.Require` crashes the request.

**Fix:** New helpers:
```go
func MustActorID(ctx context.Context, tx *sql.Tx) (string, error)
func MustTenantID(ctx context.Context, tx *sql.Tx) (string, error)
```

Both use `current_setting(..., true)` (missing_ok). Empty string returns typed `ErrActorContextMissing` / `ErrTenantContextMissing`. `authz.Require` calls these helpers at the top, returns wrapped error early. Removes raw `current_setting('...', false)` calls from inline SQL — replace with `$N` parameter binds populated from helpers.

**Test:** sqlmock test asserts mocked empty result returns typed error. Integration test calls `Require` from a synthetic non-approval caller without setting GUC, expects typed error.

---

### B5 — RolesByUserID ignores tenant

**Files:**
- Modify: `internal/modules/iam/domain/port.go`
- Modify: `internal/modules/iam/infrastructure/postgres/role_provider.go`
- Modify: `internal/modules/iam/application/cached_role_provider.go`
- Modify: `internal/modules/iam/application/dev_role_provider.go`
- Modify: `internal/modules/iam/delivery/http/middleware.go:91`
- Modify: `internal/modules/iam/infrastructure/memory/role_provider.go` (if exists)

**Problem:** Signature `RolesByUserID(ctx, userID)` queries `iam_user_roles` without `tenant_id`. Cross-tenant role bleed: user with role in tenant A appears to have role in tenant B.

**Fix:** Signature change → `RolesByUserID(ctx, userID, tenantID string) ([]Role, error)`. SQL adds `AND tenant_id = $2::uuid`. All implementers and callers updated.

Cache key in `cached_role_provider.go` becomes `userID + "|" + tenantID`.

**Test:** sqlmock asserts `WHERE tenant_id` clause. Integration `TestRolesByUserID_RejectsCrossTenantBleed`.

---

### B6 — HasAnyRole ignores tenant

**Files:**
- Modify: `internal/modules/iam/domain/port.go`
- Modify: `internal/modules/iam/infrastructure/postgres/role_admin_repository.go`
- Modify: `internal/modules/iam/infrastructure/memory/role_admin_repository.go`
- Modify: bootstrap caller (find via grep `HasAnyRole(`)

**Problem:** `HasAnyRole(ctx, role)` returns true if ANY tenant has the role. Bootstrap admin seed for tenant T is skipped if some other tenant already has system_admin.

**Fix:** Signature → `HasAnyRole(ctx, role, tenantID string) (bool, error)`. SQL adds `WHERE tenant_id = $2::uuid`. Bootstrap caller passes current tenant.

**Test:** sqlmock asserts WHERE clause. Integration `TestHasAnyRole_TenantIsolation`.

---

## Rollout Plan

| Phase | Tasks | Parallelism | Model |
|---|---|---|---|
| 0 | Worktree, codex plan validate, wiki-curator verify | sequential | sonnet |
| 1 | B4 (helpers) ‖ B3 (migration) ‖ B2 (ADR) | parallel (no file overlap) | codex / sonnet / sonnet |
| 2 | B5 (RolesByUserID propagation) → B6 (HasAnyRole propagation) | sequential (port.go conflict) | codex |
| 3 | B1 (repo writes + tenant_id) | after B5/B6 | codex |
| 4 | Verify: `go test ./...`, integration tests, codex audit, smoke | sequential | sonnet → codex audit |
| 5 | Merge via `finishing-a-development-branch`, update audit doc, wiki-curator | sequential | sonnet |

**Phase review after each phase:** Opus.

---

## Testing Strategy

**Per-bug tests:** see "Per-Bug Fix Design" sections above.

**Cross-cutting:**
- `go test -mod=mod ./...` — full suite passes
- `go test -tags=integration ./tests/integration/iam/...` — cross-tenant + GUC tests
- Smoke: bootstrap dev DB, login as approver, verify role is `approver` (not `system_admin`), perform signoff successfully, verify SoD blocks self-approval
- Codex independent audit (Group A pattern) — per-bug PASS/FAIL with file:line evidence
- Wiki-curator agent: refresh stamps on `wiki/concepts/`, `wiki/modules/iam-*.md`, new ADR

**Coverage targets:** new code ≥80% line coverage. No new lint warnings.

---

## Acceptance Criteria

- [ ] B1: `ReplaceUserRoles` second call same tenant succeeds (no 500)
- [ ] B2: ADR exists, doc comments link to it from both services
- [ ] B3: dev `approver` user has `role_code = 'approver'` after fresh DB bootstrap
- [ ] B4: calling `authz.Require` outside approval module returns `ErrActorContextMissing`, not pq panic
- [ ] B5: user with role in tenant A returns empty roles when queried as tenant B
- [ ] B6: bootstrap admin runs for new tenant even if other tenants have system_admin
- [ ] All Go tests pass with `-mod=mod`
- [ ] Integration tests pass with `-tags=integration`
- [ ] Codex audit returns 6/6 PASS
- [ ] Smoke test: approver signoff flow completes without 403
- [ ] Audit doc updated, all 6 bugs marked fixed with commit SHAs

---

## Open Questions

None.

---

## References

- Audit: `wiki/bugs/audit-2026-05-03.md` (lines 111-136)
- Prior plan: `docs/superpowers/plans/2026-05-02-iam-rbac-unification.md`
- Group A spec: `docs/superpowers/plans/2026-05-03-group-a-blockers.md`
