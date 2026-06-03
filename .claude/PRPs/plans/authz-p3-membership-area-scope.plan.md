# Plan: ADR 0022 Phase 3 — Membership Area-Scoping

## Summary
Close the original `area_admin` membership 403 at its root: thread the membership's real `areaCode` into the tier-2 `authz.Require` calls (replacing the literal `"tenant"`), delete the lone hardcoded `RoleSystemAdmin` handler gate, and surface tier-2 denial as 403. Authorization becomes tier-1 (route cap) + tier-2 (area) only; handlers keep business invariants (self-grant 403, cross-tenant 404).

## User Story
As an `area_admin` holding `membership.manage` in my process area, I want to grant/revoke memberships within that area, so that I can administer my own area without a `system_admin` and without being able to escalate across the tenant.

## Problem → Solution
Membership writes pass `"tenant"` to tier-2 (area filter OFF) AND gate the handler on `role == RoleSystemAdmin` → `area_admin` is blocked (BFLA over-restriction). → Pass real `areaCode` to tier-2 (area-scoped) AND delete the role gate. R2 co-dependency: both land in ONE PR (gate removal without area-scope = BOLA escalation).

## Metadata
- **Complexity**: Medium
- **Source PRD**: `wiki/decisions/0022-authz-capability-coherence.md` (Phase 3)
- **PRD Phase**: Phase 3 — Membership area-scoping
- **Estimated Files**: 6 (3 prod, 1 spec, 2 test) + ADR

---

## UX Design
Internal authz change — no user-facing UX transformation. Behavior delta: `area_admin` grant/revoke in managed area now 201/204 instead of 403; outside managed area now 403 from tier-2.

---

## Mandatory Reading
| Priority | File | Lines | Why |
|---|---|---|---|
| P0 | `internal/modules/iam/authz/authz.go` | 51-113 | tier-2 `Require`; `"tenant"` skips area; system_admin bypass at :64-88 runs BEFORE area query (R1 free) |
| P0 | `internal/modules/iam/infrastructure/postgres/user_area_repository.go` | 89-262 | 3 call sites; areaCode already in scope (`membership.AreaCode` / `areaCode` param / `newMembership.AreaCode`) |
| P0 | `internal/modules/iam/delivery/http/routes_memberships.go` | 149-291,374-388 | grant/revoke handlers; `canManageMembershipTarget` gate; `writeMembershipError` |
| P1 | `internal/modules/templates/delivery/http/errors.go` | 52-53 | `errors.As(err, new(iamauthz.ErrCapDenied))` → 403 pattern |
| P1 | `db/reference-data/0001_product_reference_data.sql` | 31 | `area_admin → membership.manage` seeded |
| P2 | `tests/unit/iam_memberships/area_memberships_handler_test.go` | all | in-mem repo (no authz); handler-level coverage |
| P2 | `tests/integration/iam/capability_service_test.go` | all | DSN+sql.Open real-DB harness, seed helpers |

---

## Patterns to Mirror

### CAP_DENIED_TO_403
// SOURCE: internal/modules/templates/delivery/http/errors.go:52
case errors.As(err, new(iamauthz.ErrCapDenied)):
    return http.StatusForbidden, "capability_denied"

### TIER2_AREA_CALL
// SOURCE: authz.go:51 — pass real area, "tenant" only to skip
authz.Require(ctx, tx, string(iamdomain.CapMembershipManage), membership.AreaCode)

### INTEGRATION_HARNESS
// SOURCE: tests/integration/iam/capability_service_test.go:19-30
func openDB(t *testing.T) *sql.DB { ... testdb.DSN(t) ... db.PingContext → Skipf if unreachable }

---

## Files to Change
| File | Action | Justification |
|---|---|---|
| `internal/modules/iam/infrastructure/postgres/user_area_repository.go` | UPDATE | 3× `"tenant"` → real areaCode |
| `internal/modules/iam/delivery/http/routes_memberships.go` | UPDATE | delete `canManageMembershipTarget` gate (grant+revoke); map `ErrCapDenied`→403 |
| `api/openapi/v1/openapi.yaml` | UPDATE | x-authz-area + x-authz-skip-area on grant/revoke |
| `tests/unit/iam_memberships/area_memberships_handler_test.go` | UPDATE | gate-removal + self-grant + cross-tenant assertions |
| `tests/integration/iam/membership_area_scope_test.go` | CREATE | real-DB area_admin scope + R1 system_admin bypass |
| `wiki/decisions/0022-authz-capability-coherence.md` | UPDATE | mark Phase 3 complete |

## NOT Building
- `isMembershipDirectoryAdmin` / list scoping (Phase 4)
- Any OpenAPI shape change (x-authz-* only)
- `authz-call-present` activation (Phase 5)
- New schema / managed-areas table

---

## Step-by-Step Tasks

### Task 1: Thread areaCode into tier-2
- **ACTION**: In `user_area_repository.go`, replace `"tenant"` at Insert:100, CloseActive:152, GrantAtomic:196.
- **IMPLEMENT**: Insert → `membership.AreaCode`; CloseActive → `areaCode`; GrantAtomic → `newMembership.AreaCode`.
- **GOTCHA**: oldMembership.AreaCode == newMembership.AreaCode (service builds both with same area) — use newMembership.AreaCode.
- **VALIDATE**: `go build ./...`.

### Task 2: Delete role gate + map ErrCapDenied→403
- **ACTION**: Remove `canManageMembershipTarget` calls in grantMembership + revokeMembership; delete the function. Add `ErrCapDenied`→403 case to `writeMembershipError`.
- **IMPLEMENT**: drop the `if !canManageMembershipTarget(...) { 403 }` blocks. KEEP `isSelf` 403 (grant) + `guardMembershipUserInTenant` 404. In `writeMembershipError` add `case errors.As(err, new(authz.ErrCapDenied)): 403 AUTH_FORBIDDEN`.
- **IMPORTS**: add `"metaldocs/internal/modules/iam/authz"`.
- **GOTCHA**: revoke has no self-check today and Phase 3 adds none (revoke self is legitimate). Only grant blocks self.
- **VALIDATE**: `go build ./...`; grep shows `canManageMembershipTarget` gone.

### Task 3: Spec annotation
- **ACTION**: Add to grantAreaMembership + revokeAreaMembership ops.
- **IMPLEMENT**: grant → `x-authz-area: {source: body, field: areaCode}`; revoke → `x-authz-area: {source: query, field: areaCode}`; both + `x-authz-skip-area: true` + `x-authz-skip-reason` documenting hand-rolled-handler exception.
- **GOTCHA**: x-authz-skip-area presence silences authz-call-present (code_rules.go:106); without it the rule fails (handler GrantAreaMembership not found). Net lint delta must be 0.
- **VALIDATE**: redocly lint valid; api-lint still 455.

### Task 4: Unit tests
- **ACTION**: Add handler-level tests proving gate removed.
- **IMPLEMENT**: area_admin (non-system role in ctx) grant for other target → reaches service → 201 (was 403). self-grant → 403. cross-tenant → 404. (Area enforcement itself = integration; in-mem repo has no authz.)
- **VALIDATE**: `go test ./tests/unit/iam_memberships/...`.

### Task 5: Integration tests (real DB)
- **ACTION**: CREATE `tests/integration/iam/membership_area_scope_test.go` (`//go:build integration`).
- **IMPLEMENT**: seed iam_users + actor area_admin user_process_areas row in area QMS. Drive `AreaMembershipService.Grant/Revoke` (real `UserAreaRepository`).
  - area_admin grant target in QMS → nil.
  - area_admin grant target in RH (no row) → `ErrCapDenied`.
  - R1: system_admin (iam_user_roles, NO upa row) grant in any area → nil.
  - revoke mirror in QMS.
- **GOTCHA**: authz reads `metaldocs.user_process_areas` + `metaldocs.role_capabilities`; seed there. Skip if DB unreachable.
- **VALIDATE**: `go test -tags=integration ./tests/integration/iam/...`.

### Task 6: ADR
- **ACTION**: Mark Phase 3 ✅ COMPLETE in `0022-authz-capability-coherence.md` with gates evidence.

---

## Validation Commands
```powershell
go build ./...
go test ./internal/modules/iam/... ./tests/unit/iam_memberships/... -count=1
go test -tags=integration ./tests/integration/iam/... -count=1
go run ./scripts/api-lint api/openapi/v1/openapi.yaml .   # EXPECT 455 unchanged
npx @redocly/cli lint api/openapi/v1/openapi.yaml          # EXPECT valid
```

## Acceptance Criteria
- [ ] 3 tier-2 calls pass real areaCode
- [ ] role gate deleted; isSelf + cross-tenant guards kept
- [ ] ErrCapDenied → 403
- [ ] spec annotated, no shape change, lint 455 unchanged
- [ ] unit + integration tests green (R1 explicit)
- [ ] ADR Phase 3 marked complete

## Risks
| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| ErrCapDenied surfaces as 500 not 403 | High if unmapped | area_admin denial wrong status | Task 2 mapping + integration assert |
| Integration DB unreachable in CI | Medium | tests skip | Skipf pattern (mirrors existing) |
| schema mismatch public vs metaldocs upa | Low | seed wrong table | seed where authz reads (metaldocs.*) |

## Notes
R1 (system_admin bypass) is satisfied for free: `authz.Require` system_admin EXISTS check at authz.go:64-88 returns before the area-filtered capability query — a missing per-area row cannot block it. Integration test asserts it explicitly per ADR amendment.
