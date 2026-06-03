# Plan: ADR 0022 Phase 4 — Area-scoped membership directory listing

## Summary
Replace the `isMembershipDirectoryAdmin` `RoleSystemAdmin` string gate in the
`GET /api/v1/iam/area-memberships` handler with capability/area-aware directory
scoping resolved at the data layer. `system_admin` keeps the full tenant
directory; `area_admin` sees memberships only in areas they manage
(`membership.manage`-granting role in `user_process_areas`), filtered IN SQL;
everyone else stays self-only. The handler stops pivoting on a role-name literal.

## User Story
As an `area_admin`, I want the membership directory to show every membership in
the areas I manage (not just my own row), so that I can administer my areas
without `system_admin` rights, while never seeing memberships outside my areas.

## Problem → Solution
Today the directory gate is `role == RoleSystemAdmin`: a binary system-admin /
self-only split (ADR 0021 violation: pivots on role name; ADR 0022 R2 leaves
`area_admin` over-restricted to self-only). → Three-way scope resolved at the
data layer: tenant-wide (system_admin bypass) / managed-areas (SQL-filtered) /
self-only.

## Metadata
- **Complexity**: Medium
- **Source PRD**: `wiki/decisions/0022-authz-capability-coherence.md` (Phase 4)
- **PRD Phase**: Phase 4 — Directory/list area-scoping
- **Estimated Files**: 6 (3 prod, 3 test) + ADR

---

## UX Design
Internal authz change — no user-facing UX transformation. The Admin Center
Memberships tab simply returns more rows for an `area_admin` (their managed
areas instead of only their own row).

---

## Mandatory Reading

| Priority | File | Lines | Why |
|---|---|---|---|
| P0 | `internal/modules/iam/delivery/http/routes_memberships.go` | 93-147, 368-379 | listMemberships + the gate being replaced |
| P0 | `internal/modules/iam/authz/authz.go` | 90-113 | the role_capabilities↔user_process_areas join + system_admin bypass to mirror in SQL |
| P0 | `internal/modules/iam/infrastructure/postgres/user_area_repository.go` | 53-87 | ListByTenant SQL to extend |
| P0 | `internal/modules/iam/application/area_membership_service.go` | 23-30, 56-62 | UserAreaWriteRepository interface + ListByTenant service method |
| P1 | `tests/unit/iam_memberships/area_memberships_handler_test.go` | 44-206, 261-330 | in-memory fake + harness + the 3 scope tests to extend |
| P1 | `internal/modules/iam/delivery/http/routes_memberships_contract_test.go` | 18-42 | fakeUserAreaWriteRepository that must satisfy the widened interface |
| P2 | `tests/integration/iam/membership_area_scope_test.go` | all | real-DB seeding pattern (GUC bypass) for the integration test |

## External Documentation
No external research needed — feature uses established internal patterns
(repository + service + hand-rolled handler; tier-2 SQL cap join already exists
in `authz.Require`).

---

## Patterns to Mirror

### CAP_AREA_JOIN (the managed-areas derivation — mirror as a SELECT DISTINCT)
// SOURCE: internal/modules/iam/authz/authz.go:91-102
```go
SELECT EXISTS (
  SELECT 1
  FROM metaldocs.role_capabilities rc
  JOIN metaldocs.user_process_areas upa
    ON upa.role = rc.role
   AND upa.tenant_id = $4::uuid
   AND upa.user_id   = $3
   AND upa.effective_to IS NULL
  WHERE rc.capability = $1
    AND ($2 = 'tenant' OR upa.area_code = $2)
)
```

### SYSTEM_ADMIN_BYPASS (tenant-wide detection at data layer — mirror the EXISTS)
// SOURCE: internal/modules/iam/authz/authz.go:66-82
```go
SELECT EXISTS (
  SELECT 1 FROM metaldocs.iam_user_roles ur
   WHERE ur.user_id=$1 AND ur.tenant_id=$2::uuid AND ur.role_code='system_admin'
  UNION ALL
  SELECT 1 FROM metaldocs.iam_group_members gm
    JOIN metaldocs.iam_groups g ON g.id=gm.group_id
    JOIN metaldocs.iam_group_roles gr ON gr.group_id=gm.group_id
   WHERE gm.user_id=$1 AND gm.tenant_id=$2::uuid AND g.tenant_id=$2::uuid AND gr.role='system_admin'
)
```

### REPOSITORY_LIST (static, injection-safe optional filters)
// SOURCE: internal/modules/iam/infrastructure/postgres/user_area_repository.go:57-87
```go
WHERE tenant_id = $1::uuid
  AND effective_from <= $2
  AND (effective_to IS NULL OR effective_to > $2)
  AND ($3 = '' OR user_id   = $3)
  AND ($4 = '' OR area_code = $4)
  AND ($5 = '' OR role      = $5)
```

### SERVICE_PASSTHROUGH
// SOURCE: internal/modules/iam/application/area_membership_service.go:60-62
```go
func (s *AreaMembershipService) ListByTenant(ctx, tenantID, userID, areaCode, role string) ([]domain.UserProcessArea, error) {
	return s.repo.ListByTenant(ctx, tenantID, userID, areaCode, role, s.nowFn())
}
```

### TEST_STRUCTURE (in-memory fake + httptest harness)
// SOURCE: tests/unit/iam_memberships/area_memberships_handler_test.go:182-206
(harness wires NewAreaMembershipService(memAreaRepo, nil) → NewMembershipHandler)

---

## Files to Change

| File | Action | Justification |
|---|---|---|
| `internal/modules/iam/application/area_membership_service.go` | UPDATE | widen `UserAreaWriteRepository` with `MembershipDirectoryScope` + `ListByTenantInManagedAreas`; add 2 service passthroughs |
| `internal/modules/iam/infrastructure/postgres/user_area_repository.go` | UPDATE | implement both methods in SQL (R3 data-layer filter) |
| `internal/modules/iam/delivery/http/routes_memberships.go` | UPDATE | rewrite listMemberships scope branch; delete `isMembershipDirectoryAdmin` + `RolesFromContext` import |
| `tests/unit/iam_memberships/area_memberships_handler_test.go` | UPDATE | extend memAreaRepo with new methods + scope config; add area_admin managed-area test |
| `internal/modules/iam/delivery/http/routes_memberships_contract_test.go` | UPDATE | add stub methods to fakeUserAreaWriteRepository |
| `tests/integration/iam/membership_area_scope_test.go` | UPDATE | add real-DB area_admin directory test (managed vs other area) |
| `wiki/decisions/0022-authz-capability-coherence.md` | UPDATE | mark Phase 4 complete + status line |

## NOT Building
- No OpenAPI shape change (listAreaMemberships contract unchanged).
- No new managed-areas table/service (YAGNI — `user_process_areas` + `role_capabilities` already express it).
- No change to grant/revoke (Phase 3).
- No `authz-call-present` lint activation (Phase 5).
- No array-param plumbing — managed-area restriction is a SQL subquery JOIN, not a passed `[]string` ANY() param (avoids pgx-stdlib array-binding gotcha and is unambiguously "filtered at the data layer").

---

## Step-by-Step Tasks

### Task 1: Widen the repository interface + service passthroughs
- **ACTION**: In `area_membership_service.go`, add to `UserAreaWriteRepository`:
  `MembershipDirectoryScope(ctx, tenantID, actorID string) (tenantWide bool, hasManagedAreas bool, err error)`
  and `ListByTenantInManagedAreas(ctx, tenantID, userID, areaCode, role, actorID, capability string, now time.Time) ([]domain.UserProcessArea, error)`.
  Add service methods `DirectoryScope(ctx, tenantID, actorID)` and
  `ListByTenantInManagedAreas(ctx, tenantID, userID, areaCode, role, actorID, capability string)` passing `s.nowFn()`.
- **MIRROR**: SERVICE_PASSTHROUGH.
- **GOTCHA**: keep `capability` a param (handler passes `string(domain.CapMembershipManage)`) so the service stays cap-agnostic.
- **VALIDATE**: `go build ./internal/modules/iam/...`.

### Task 2: Implement Postgres methods (R3 — SQL filter)
- **ACTION**: In `user_area_repository.go` add `MembershipDirectoryScope` (one query, two EXISTS columns: system_admin bypass + cap-area join) and `ListByTenantInManagedAreas` (ListByTenant SQL + `AND area_code IN (<cap-area subquery on $6 actor, $7 capability>)`).
- **MIRROR**: SYSTEM_ADMIN_BYPASS, CAP_AREA_JOIN, REPOSITORY_LIST.
- **IMPORTS**: none new.
- **GOTCHA**: membership rows are `public.user_process_areas`; `role_capabilities` is `metaldocs.role_capabilities`; the `metaldocs.user_process_areas` view maps to the same base table — use `public.user_process_areas` in the subquery for consistency with the rest of this repo. The cap-area subquery must NOT apply the `$2='tenant'` area branch (we want ALL the actor's managed areas).
- **VALIDATE**: `go build ./...`.

### Task 3: Rewrite the handler scope branch
- **ACTION**: In `listMemberships`, replace the `if !isMembershipDirectoryAdmin(...)` block with:
  resolve `actor` via `authn.UserIDFromContext` (403 if empty); call
  `h.svc.DirectoryScope(ctx, tenantID, actor)`; switch:
  - `tenantWide` → keep current tenant-wide path (optional userId/area/role filters, cross-tenant guard when userId set).
  - `hasManagedAreas` → call `ListByTenantInManagedAreas(...)` (R3); still honor userId/area/role filters (passed through) + cross-tenant guard when userId set.
  - else (self-only) → enforce userId==actor-or-empty (403 on mismatch), set userId=actor, use existing ListByTenant.
  Delete `isMembershipDirectoryAdmin` and drop the now-unused `iamdomain.RolesFromContext`/`RoleSystemAdmin` usage. Keep `iamdomain` import (still used by DTO).
- **MIRROR**: existing listMemberships structure.
- **GOTCHA**: `authenticatedActor` returns `"system"` fallback — do NOT use it for the scope query; use `authn.UserIDFromContext` directly so an unauthenticated caller is rejected, not scoped to a "system" actor.
- **VALIDATE**: `go build ./...`; `grep RoleSystemAdmin routes_memberships.go` → no match.

### Task 4: Unit tests (fakes + scope coverage)
- **ACTION**: Extend `memAreaRepo` with `tenantWide map[string]bool` and `managedAreas map[string][]string` + the two new methods. `MembershipDirectoryScope` reads the maps; `ListByTenantInManagedAreas` filters in-mem rows to `managedAreas[actor]` (intersect with optional filters). Update `userReq`/harness so an area_admin actor is configured. Add `TestListAreaMemberships_AreaAdminSeesManagedAreasOnly` (rows in managed area returned, rows in other area excluded). Keep existing system_admin tenant-wide + viewer self-only tests green (configure `tenantWide[adminID]=true`).
- **MIRROR**: TEST_STRUCTURE.
- **GOTCHA**: existing tests rely on the in-memory repo; the system_admin path now flows through `DirectoryScope` → fake must report `tenantWide=true` for `adminID`. Viewer reports both false → self-only branch.
- **VALIDATE**: `go test ./tests/unit/iam_memberships/... -count=1`.

### Task 5: Contract-test fake stubs
- **ACTION**: Add `MembershipDirectoryScope` (return false,false,nil) and `ListByTenantInManagedAreas` (return nil,nil) to `fakeUserAreaWriteRepository`.
- **VALIDATE**: `go test ./internal/modules/iam/delivery/http/... -count=1`.

### Task 6: Integration test (real DB, R3 proof)
- **ACTION**: Add `TestMembershipDirectory_AreaAdminScopedInSQL` to `tests/integration/iam/membership_area_scope_test.go`: seed an area_admin with `membership.manage` in area A (via role grant) and memberships of other users in area A and area B; assert `ListByTenantInManagedAreas` returns area-A rows only, never area-B. Assert system_admin path (`MembershipDirectoryScope` → tenantWide) sees all.
- **MIRROR**: existing GUC-bypass seeding in the same file.
- **VALIDATE**: `go test -tags=integration ./tests/integration/iam/... -run MembershipDirectory -count=1` (DB-gated; record if env unavailable).

### Task 7: ADR + gates
- **ACTION**: Mark Phase 4 ✅ COMPLETE in ADR 0022 (status line + Phase 4 section), bump `Last verified`. Run all gates.
- **VALIDATE**: gates below.

---

## Testing Strategy

### Unit Tests
| Test | Input | Expected | Edge? |
|---|---|---|---|
| SystemAdminSeesTenantWideDirectory | adminID, tenantWide=true, no filter | all tenant rows | no |
| AreaAdminSeesManagedAreasOnly | area-admin, managedAreas=[QMS] | QMS rows only, RH excluded | yes (BOLA guard) |
| NonAdminIsSelfScoped | viewer, both false | own row only | yes |
| NonAdminCannotTargetOther | viewer, userId=other | 403 | yes |
| AreaScopedUnderTenantIsolation | admin, cross-tenant target | only same-tenant row | yes |

### Edge Cases Checklist
- [x] area_admin requests an unmanaged area via `areaCode` filter → empty (intersect, no leak)
- [x] actor with zero managed areas + not system_admin → self-only
- [x] cross-tenant isolation preserved (tenant_id bound in all queries)
- [x] unauthenticated (no actor) → 403

---

## Validation Commands
```powershell
go build ./...
go test ./internal/modules/iam/... ./tests/unit/iam_memberships/... -count=1
go run ./scripts/api-lint api/openapi/v1/openapi.yaml .
npx @redocly/cli lint api/openapi/v1/openapi.yaml
```
EXPECT: build clean; iam + memberships tests pass; api-lint violation count unchanged vs Phase-3 baseline (no new); redocly valid.

---

## Acceptance Criteria
- [ ] `RoleSystemAdmin` literal gone from routes_memberships.go
- [ ] area_admin managed-area filter is in SQL (R3)
- [ ] system_admin tenant-wide; area_admin managed-only; viewer self-only — all tested
- [ ] cross-tenant isolation preserved
- [ ] no OpenAPI shape change
- [ ] all gates green

## Risks
| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| pgx array binding for ANY() | n/a | — | avoided: SQL subquery JOIN, no array param |
| widened interface breaks other fakes | medium | low | update both fakes (Task 4, 5) — only 2 implementors |
| schema mismatch public vs metaldocs | low | med | use public.user_process_areas + metaldocs.role_capabilities (verified in baseline) |
| integration DB env unavailable | medium | low | unit + contract tests cover branching; record DB gate status |

## Notes
- R3 satisfied by the in-SQL subquery JOIN (ADR explicitly allows "pass the managed-area set / join").
- system_admin tenant-wide detection moves from handler role-string to data-layer EXISTS (same bypass authz.Require already uses) — handler no longer names a role (ADR 0021).
- Confidence: 8/10 single-pass (integration DB env is the only soft spot).
