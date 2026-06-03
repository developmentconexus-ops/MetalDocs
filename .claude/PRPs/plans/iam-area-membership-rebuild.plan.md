# Plan: IAM Area Membership Admin — Full Production Rebuild

## Summary
Replace the legacy `/admin/memberships` MVP stub with a production-grade Memberships surface that matches the IAM Admin Center bar (PR-1..PR-12b). Contract-first via OpenAPI + oapi-codegen, TanStack Query, CSS Modules + design tokens, ADR 0016 view/manage cap split, RFC 9457 errors, audit + governance emission, vitest + Go integration tests.

## User Story
As a **tenant admin (system_admin) or area_admin**, I want a Memberships surface inside Admin Center that lets me search by user OR by area, see all active grants with their effective windows, and grant/revoke with proper validation and audit trail, so that I can run ISO-segregated area access without leaving the admin console or guessing role caps.

## Problem → Solution
**Current state (post-PR #55):** orphaned `/admin/memberships` route running an inline-styled stub; hand-rolled `membershipApi.ts`; no OpenAPI schemas (skeleton only); UI offers free-text User ID filter and nothing else; `MembershipGovernanceLogger` wired with `nil` (T-007 open); no audit emission on grant/revoke; mixed PT/EN copy with mojibake; no tests; not integrated into the 6-tab Admin Center IA.

**Desired state:** 7th tab in Admin Center (`overview | people | roles | audit | sessions | usage | memberships`) using design-system primitives, TanStack Query hooks, OpenAPI-codegen types, ADR 0016 cap-gated manage UI, RFC 9457 error handling, grant/revoke emit both audit events and governance log entries, full test coverage (vitest unit + integration + Go contract), wiki bumped + T-007 closed.

## Metadata
- **Complexity**: Large → split into 6 mergeable PRs
- **Source PRD**: N/A (free-form rebuild request, anchored to QA findings from `qa/iam-area-membership` PR #55)
- **PRD Phase**: N/A
- **Estimated Files**: ~30 created/changed across BE + FE + spec + wiki + tests
- **Branch strategy**: stack per-PR onto `feat/iam-memberships-rebuild` cut from `main` after `qa/iam-area-membership` merges; each PR is independently mergeable.

---

## Phase 0 Decision — Tab Integration (REQUIRED before PR-2)

| Option | Description | Pros | Cons |
|---|---|---|---|
| **A (Recommended)** | Fold Memberships into `AdminCenterPage` as the 7th tab. Route becomes `/admin/memberships` BUT mounted inside the `AdminCenterPage` outlet alongside `/admin/overview`, `/admin/people`, etc. | One IA, consistent chrome, matches Admin Center precedent, full-feature surface (user + area cross-cut). | Larger PR-2 (page rebuild + IA reshape). |
| B | Keep deep-link admin route at `/admin/memberships` outside Admin Center; surface "Manage memberships" CTA from People tab `UserDetailDrawer` and Roles tab. | Smaller PR-2. | Two IAs (orphan route + drawer-embedded mini-grid), copy split, harder to discover. |

**Default = A.** Plan below assumes A. If user picks B, drop PR-5 area-context view and demote PR-2 to "render in standalone shell".

**Halt point:** do not start PR-2 until user confirms A or B in writing.

---

## UX Design

### Before (PR #55 baseline)
```
┌────────────────────────────────────────────────────────┐
│  Memberships de Area                  [+ Conceder acesso] │  ← orphan page
│  ┌──────────────────────────────────────┐  [Buscar]      │
│  │ Filtrar por User ID                  │                │
│  └──────────────────────────────────────┘                │
│                                                            │
│  user_id    area  papel    desde       ate   acoes        │
│  approver   rh    approver 2026-05-31    -   [Revogar]    │
│                                                            │
│  (inline styles, mixed PT/EN, mojibake, no tenant ctx)    │
└────────────────────────────────────────────────────────┘
```

### After (Option A)
```
┌────────────────────────────────────────────────────────┐
│  IAM › Central de administracao                                │
│  ──────────────────────────────────────────────────────────── │
│  [Visao geral] [Pessoas] [Funcoes] [Auditoria]                │
│  [Sessoes]    [Consumo]  [Memberships ←active]                │
│                                                                │
│  ┌─────────────────┐  Search: [______________]  Area: [▾]      │
│  │ Total ativos    │                          Funcao: [▾]      │
│  │      37         │                          [+ Conceder]     │
│  └─────────────────┘                                            │
│                                                                 │
│  Usuario       Area       Funcao    Desde       Concedido por  │
│  ─────────────────────────────────────────────────────────────  │
│  approver      RH (Rec…)  approver  2026-05-31  admin   [•••] │
│  author-test   QMS        author    2026-04-12  admin   [•••] │
│  …                                                              │
│                              [< prev | 1 / 3 | next >]          │
│                                                                 │
│  (CSS Modules + tokens, sonner toast, ADR 0016 cap-gated [•••])│
└────────────────────────────────────────────────────────┘
```

### Interaction Changes
| Touchpoint | Before | After | Notes |
|---|---|---|---|
| Discovery | Hidden orphan route | 7th tab in Admin Center | matches existing nav |
| Filter | Free-text User ID only | Search-as-you-type (user OR area) + Area + Role selects | TanStack Query keyed on params |
| Grant | `window.confirm` blocker, dialog with bare select | Design-system `Dialog` + `Select` + `Input`, validation messages | RFC 9457 problem mapping |
| Revoke | `window.confirm("Revogar acesso…")` | `Dialog` confirm with tenant + area + user context | no native prompts |
| Empty state | text "Nenhum membership encontrado." | Empty-state component with CTA when `membership.manage` | reuses Admin Center pattern |
| Error state | red text dump | `ErrorBanner` component (existing `components/ErrorBanner.tsx`) | structured |
| Feedback | none | sonner toast on grant/revoke success/failure | matches PeopleTab pattern |
| Loading | "Carregando..." text | Skeleton rows in `AreaMembershipsTable` | matches `UsersDirectory` |
| Pagination | none | cursor + limit per PR-12b cursor contract | reuse `decodeCursor` shape |
| Sort | none | per-column sort (user, area, role, granted-at) | URL searchParams |
| Cap-gate | route gate on `membership.view`; manage UI shown to all viewers (FIXED in PR #55) | Same FIXED gate + per-row [•••] hidden when no `membership.manage` | defense-in-depth, server still enforces |
| Tenant context | invisible | breadcrumb shows tenant name (from session) | matches Admin Center toolbar |

---

## Mandatory Reading

| Priority | File | Lines | Why |
|---|---|---|---|
| P0 | `CLAUDE.md` | all | read order, gates, hard-stop rule, evidence rule |
| P0 | `wiki/quality/qa-operating-system.md` | all | 7-gate loop, classification, severity |
| P0 | `wiki/quality/screen-qa-checklist.md` | all | per-screen QA bar |
| P0 | `wiki/modules/iam.md` | all | route table §5.3, capability registry, T-004/T-007 status, ADR refs |
| P0 | `wiki/modules/iam-tech-debt.md` | T-007, T-008, T-011 rows | T-007 closure in scope |
| P0 | `wiki/decisions/0016-view-grade-capabilities.md` | all | view/manage cap split (canonical) |
| P0 | `wiki/decisions/0012-contract-first-api.md` | all | OpenAPI + oapi-codegen ladder |
| P0 | `wiki/architecture/frontend-structure.md` | all | feature-sliced layout, CSS Modules + tokens, TanStack Query, generated types |
| P0 | `wiki/architecture/api-contract.md` + `api-design-system.md` | all | RFC 9457 + spec conventions |
| P0 | `frontend/apps/web/src/features/iam/pages/AdminCenterPage.tsx` | 1-92 | tab IA + `useHasCapability` + `WorkspaceViewFrame` |
| P0 | `frontend/apps/web/src/features/iam/tabs/PeopleTab.tsx` | 1-200 | tab pattern, searchParams, mutations, toast, dialogs |
| P0 | `frontend/apps/web/src/features/iam/tabs/PeopleTab.route.tsx` | all | tab+drawer outlet pattern |
| P0 | `frontend/apps/web/src/features/iam/queries/useUserMembershipsQuery.ts` | all | exact `useQuery + api.GET` shape to mirror |
| P0 | `frontend/apps/web/src/features/iam/components/UsersDirectory.tsx` + `.module.css` | all | table pattern with sort + skeleton |
| P0 | `frontend/apps/web/src/features/iam/components/InviteUserDialog.tsx` + `.module.css` | all | dialog form pattern with `useMutation` + sonner |
| P0 | `frontend/apps/web/src/features/iam/components/UserMembershipsTable.tsx` | all | existing memberships subview to reconcile/reuse |
| P0 | `frontend/apps/web/src/features/iam/routes.tsx` | all | how `admin/*` is composed; how to add 7th tab |
| P0 | `frontend/apps/web/src/features/shell/components/AppShell.tsx` | 1-78 | capability gate inheritance, `requiresAnyCapability` semantics |
| P0 | `frontend/apps/web/src/features/shell/components/AppShell.test.tsx` | 60-145 | acceptance test pattern for cap-gate redirects |
| P0 | `frontend/apps/web/src/lib/queryKeys.ts` | 67-84 | add `QK.iam.memberships*` keys here |
| P0 | `internal/modules/iam/delivery/http/routes_memberships.go` | all | current handler (DTO already added in PR #55); needs operationId comments, query-param shape for DELETE, error envelope alignment, governance + audit hooks |
| P0 | `internal/modules/iam/delivery/http/people_handler.go` | all | how PeopleHandler emits audit + RFC 9457 + tenant guard — mirror exactly for memberships |
| P0 | `internal/modules/iam/application/area_membership_service.go` | all | `Grant` / `Revoke` / `ListActive` service surface; governance logger interface |
| P0 | `apps/api/cmd/metaldocs-api/main.go` | 217 ± 20 | T-007 wiring site (`MembershipGovernanceLogger=nil`) |
| P0 | `apps/api/cmd/metaldocs-api/permissions.go` | all | add routeRules for the area-membership ops (currently mapped via fallthrough; PR-1 nails it down like PR-7b did for PeopleHandler) |
| P0 | `api/openapi/v1/openapi.yaml` | 2103-2146 | current skeleton spec — must flesh out with schemas, query params, error responses |
| P1 | `internal/modules/iam/infrastructure/postgres/user_area_repository.go` | 52-150 | tier-2 enforcement + tx pattern for Insert/Close/GrantAtomic |
| P1 | `internal/modules/iam/authz/authz.go` | all | tier-2 Require shape (already wired into repo writes) |
| P1 | `internal/platform/problem/problem.go` | all | RFC 9457 envelope builder |
| P1 | `internal/modules/audit/domain/writer.go` (or equivalent) | all | audit emission interface |
| P1 | `internal/modules/iam/application/area_membership_governance_logger.go` (NEW or existing iface) | all | locate the iface definition + any non-nil impl |
| P2 | `internal/modules/iam/delivery/http/admin_handler.go` | 319-369 | role-upsert audit emission pattern → mirror for membership grant/revoke |
| P2 | `tests/integration/iam/*` | all | how Go integration tests are structured for handler + tenant isolation |

## External Documentation

| Topic | Source | Key Takeaway |
|---|---|---|
| oapi-codegen | https://github.com/oapi-codegen/oapi-codegen (already pinned in `tools/`) | regen FE types via project script (`scripts/generate-frontend-types.ps1` or equivalent); never hand-edit `lib/api-types/` |
| TanStack Query v5 | already in tree | use `staleTime` per resource hot-path; invalidate by exact key on mutation success |
| sonner | already imported via `RootProviders.tsx` | `toast.success` / `toast.error` after mutation settle |
| RFC 9457 | `wiki/architecture/api-design-system.md` | `application/problem+json` with `type/title/status/detail/instance/code/errors` |

**No additional external research needed** — feature uses fully established internal patterns.

---

## Patterns to Mirror

### NAMING_CONVENTION (FE features)
```tsx
// SOURCE: frontend/apps/web/src/features/iam/tabs/PeopleTab.tsx:60
export default function PeopleTab() { … }

// SOURCE: frontend/apps/web/src/features/iam/tabs/PeopleTab.route.tsx:4
export function Component() {
  return (<><PeopleTab /><Outlet /></>);
}
```
→ Use `MembershipsTab` + `MembershipsTab.route.tsx` + `MembershipsTab.module.css`.

### QUERY_HOOK
```ts
// SOURCE: frontend/apps/web/src/features/iam/queries/useUserMembershipsQuery.ts:7
export function useUserMembershipsQuery(userId: string) {
  return useQuery({
    queryKey: QK.iam.userMemberships(userId),
    queryFn: async () => {
      const { data, error } = await api.GET("/iam/users/{userId}/memberships", {
        params: { path: { userId } },
      });
      if (error) throw error;
      return data;
    },
    staleTime: STALE_30S,
    enabled: userId.length > 0,
  });
}
```
→ Build `useMembershipsQuery(params)`, `useGrantMembershipMutation()`, `useRevokeMembershipMutation()` in the same shape against `api.GET/POST/DELETE("/iam/area-memberships", …)` after codegen regenerates.

### MUTATION_PATTERN
```tsx
// SOURCE: frontend/apps/web/src/features/iam/mutations/useBulkUsersMutation.ts (referenced from PeopleTab.tsx:8)
// USE: mutation onSuccess → toast.success + queryClient.invalidateQueries({ queryKey: QK.iam.memberships.all })
```

### CAP_GATE
```tsx
// SOURCE: frontend/apps/web/src/features/iam/pages/AdminCenterPage.tsx:13-16
const canManageMembership = useHasCapability("membership.manage");
const canViewMembership = useHasCapability("membership.view");
```
→ Use `useHasCapability("membership.manage")` for `+ Conceder` / `[•••]` menu / `Revogar` row action visibility. Route stays gated at `membership.view` in `routes.tsx`.

### RFC_9457_ERROR (BE)
```go
// SOURCE: internal/modules/iam/delivery/http/people_handler.go (multiple sites)
_ = problem.Write(w, problem.New(http.StatusForbidden, "AUTH_FORBIDDEN", "Insufficient permissions"))
```
→ Already in `routes_memberships.go` post-PR #55; PR-1 adds explicit codes for `MEMBERSHIP_NOT_FOUND`, `UNKNOWN_ROLE`, `MEMBERSHIP_EXISTS` (409), `VALIDATION_ERROR` (400 cases — null/empty/bad role).

### AUDIT_EMISSION (BE)
```go
// SOURCE: internal/modules/iam/delivery/http/admin_handler.go:369 (handleUserRoleUpsert)
h.recordAudit(r.Context(), "iam.user.role.upserted", map[string]any{…})
```
→ Add `recordMembershipAudit` helper on `MembershipHandler` emitting `iam.area_membership.granted` + `iam.area_membership.revoked` with `{actor, tenant, target_user, area_code, role, trace_id}`.

### TENANT_GUARD (BE)
```go
// SOURCE: internal/modules/iam/delivery/http/people_handler.go::handleListMemberships
//   calls guardUserInTenant → 404 on cross-tenant
```
→ Add `guardMembershipUserInTenant` to `MembershipHandler` so cross-tenant probes return 404 (matches PR-12b PeopleHandler pattern).

### TIER_2_ENFORCE (BE repo)
```go
// SOURCE: internal/modules/iam/infrastructure/postgres/user_area_repository.go:59
//   authz.Require(ctx, tx, iamdomain.CapMembershipManage, areaCode)
```
→ Already wired (Plan 5). PR-3 governance-logger does NOT change this.

### TEST_STRUCTURE (FE)
```tsx
// SOURCE: frontend/apps/web/src/features/shell/components/AppShell.test.tsx:60
{ path: 'admin/memberships', handle: { requiresCapability: 'membership.view' } }
```
→ Add tests: `redirects /admin/memberships when only user.view`, `renders /admin/memberships when membership.view granted`, `hides + Conceder when only membership.view, shows it when membership.manage`.

### TEST_STRUCTURE (BE)
```go
// SOURCE: tests/integration/iam/people_handler_*.go (e.g. TestListMemberships_RejectsCrossTenantUserWith404 — PR-12b)
```
→ Mirror for `TestGrantMembership_EmitsAuditAndGovernance`, `TestRevokeMembership_RejectsCrossTenantUserWith404`, `TestListAreaMemberships_AreaScopedUnderTenantIsolation`, `TestGrantMembership_DuplicateReturns409`.

---

## Files to Change

| File | Action | Justification | PR |
|---|---|---|---|
| `api/openapi/v1/openapi.yaml` (lines 2103-2146 + new component schemas) | UPDATE | Flesh out skeleton spec: add component schemas (`AreaMembership`, `AreaMembershipListResponse`, `GrantAreaMembershipRequest`, `RevokeAreaMembershipRequest`), query params (`userId`, `areaCode`, `cursor`, `limit`), all error responses w/ `ApiErrorEnvelope`, server-relative path (`/iam/area-memberships` not `/api/v1/iam/area-memberships`) | PR-1 |
| `lib/api-types/index.d.ts` (regen output) | REGEN | Output of `corepack pnpm gen:api` (or repo script) — never hand-edit | PR-1 |
| `apps/api/cmd/metaldocs-api/permissions.go` | UPDATE | Explicit routeRules for GET/POST/DELETE `/iam/area-memberships` mapped to `CapMembershipView`/`CapMembershipManage` (close fallthrough hole; mirror PR-7b PeopleHandler fix) | PR-1 |
| `internal/modules/iam/delivery/http/routes_memberships.go` | UPDATE | Add `guardMembershipUserInTenant` (cross-tenant 404), `recordMembershipAudit` (calls `audit.Writer.Record`), explicit `MEMBERSHIP_EXISTS` 409 mapping, query-param shape match for DELETE (already query, document in spec), structured validation errors via `problem.New` with `errors[]` field | PR-1 |
| `internal/modules/iam/delivery/http/routes_memberships_test.go` | CREATE | Unit/integration for handler error envelope + audit emission stub | PR-1 |
| `tests/integration/iam/area_memberships_handler_test.go` | CREATE | `TestListAreaMemberships_ContractShape`, `TestGrantMembership_EmitsAuditAndGovernance`, `TestRevokeMembership_RejectsCrossTenantUserWith404`, `TestGrantMembership_DuplicateReturns409`, `TestListAreaMemberships_AreaScopedUnderTenantIsolation` | PR-1 |
| `internal/modules/iam/application/area_membership_governance_logger.go` (or existing) | CONFIRM | Locate iface; if missing-impl, add `postgresMembershipGovernanceLogger` writing to `governance_events` table (mirror what SECURITY DEFINER SQL path emits today) | PR-3 |
| `internal/modules/iam/infrastructure/postgres/governance_logger.go` | CREATE | Concrete logger writing `governance_events` row per grant/revoke; same actor/tenant/trace shape as SECURITY DEFINER path | PR-3 |
| `apps/api/cmd/metaldocs-api/main.go` (line 217) | UPDATE | Wire concrete logger: `NewAreaMembershipService(iampg.NewUserAreaRepository(deps.SQLDB), iampg.NewMembershipGovernanceLogger(deps.SQLDB))` | PR-3 |
| `tests/integration/iam/membership_governance_test.go` | CREATE | Asserts `governance_events` row written on both grant + revoke from app-service path; closes T-007 | PR-3 |
| `frontend/apps/web/src/lib/queryKeys.ts` | UPDATE | Add `QK.iam.memberships.list(params)`, `QK.iam.memberships.byArea(areaCode)`, `QK.iam.memberships.byUser(userId)`, `QK.iam.memberships.kpi()` | PR-2 |
| `frontend/apps/web/src/features/iam/queries/useMembershipsQuery.ts` | CREATE | TanStack Query hook against `api.GET("/iam/area-memberships", { params: { query: …}})` | PR-2 |
| `frontend/apps/web/src/features/iam/queries/useMembershipsKpiQuery.ts` | CREATE | Single KPI count (total active) for header strip | PR-2 |
| `frontend/apps/web/src/features/iam/mutations/useGrantMembershipMutation.ts` | CREATE | POST + toast + `invalidateQueries(QK.iam.memberships.list({}))` | PR-2 |
| `frontend/apps/web/src/features/iam/mutations/useRevokeMembershipMutation.ts` | CREATE | DELETE + toast + invalidate | PR-2 |
| `frontend/apps/web/src/features/iam/tabs/MembershipsTab.tsx` | CREATE | Container — owns searchParams, calls hooks, renders directory + filter bar + dialogs | PR-2 |
| `frontend/apps/web/src/features/iam/tabs/MembershipsTab.module.css` | CREATE | CSS Module + tokens | PR-2 |
| `frontend/apps/web/src/features/iam/tabs/MembershipsTab.route.tsx` | CREATE | Lazy route stub + outlet | PR-2 |
| `frontend/apps/web/src/features/iam/components/MembershipsDirectory.tsx` + `.module.css` | CREATE | Table primitive — sort, skeleton, empty state, per-row action menu (cap-gated) | PR-2 |
| `frontend/apps/web/src/features/iam/components/MembershipsFilterBar.tsx` + `.module.css` | CREATE | Search + Area select + Role select | PR-2 |
| `frontend/apps/web/src/features/iam/components/GrantMembershipDialog.tsx` + `.module.css` | CREATE | Design-system Dialog form (user picker, area select, role select, validation) | PR-2 |
| `frontend/apps/web/src/features/iam/components/RevokeMembershipDialog.tsx` + `.module.css` | CREATE | Confirm dialog with full context — replaces `window.confirm` | PR-2 |
| `frontend/apps/web/src/features/iam/components/MembershipKpiStrip.tsx` + `.module.css` | CREATE | Total-active KPI card matching `KpiStrip` pattern | PR-2 |
| `frontend/apps/web/src/features/iam/pages/AdminCenterPage.tsx` | UPDATE | Append 7th tab `{ path: "memberships", label: "Memberships", isVisible: canViewMembership }` | PR-2 |
| `frontend/apps/web/src/features/iam/routes.tsx` | UPDATE | Add `{ path: "memberships", handle: { requiresCapability: "membership.view" }, lazy: () => import("./tabs/MembershipsTab.route") }` inside the `admin/*` children; DELETE the orphan `admin/memberships` top-level route block | PR-2 |
| `frontend/apps/web/src/features/iam/AreaMembershipAdminPage.tsx` | DELETE | Replaced by `MembershipsTab` | PR-2 |
| `frontend/apps/web/src/features/iam/pages/AreaMembershipAdminRoutePage.tsx` | DELETE | Orphan route shim — no longer needed | PR-2 |
| `frontend/apps/web/src/features/iam/MembershipGrantDialog.tsx` | DELETE | Replaced by `GrantMembershipDialog` in `components/` | PR-2 |
| `frontend/apps/web/src/features/iam/membershipApi.ts` | DELETE | Replaced by codegen + TanStack hooks | PR-2 |
| `frontend/apps/web/src/features/shell/components/AppShell.test.tsx` | UPDATE | Add cases for the new tab path under the Admin Center; keep PR-12b redirect cases | PR-2 |
| `frontend/apps/web/src/features/iam/tabs/MembershipsTab.test.tsx` | CREATE | Integration: admin sees + Conceder + [•••]; viewer with membership.view does NOT; empty state; sonner toast on grant success | PR-2 |
| `frontend/apps/web/src/features/iam/queries/__tests__/useMembershipsQuery.test.tsx` | CREATE | Unit: param→queryKey mapping, error path | PR-2 |
| `frontend/apps/web/src/features/iam/tabs/MembershipsByAreaView.tsx` + `.module.css` | CREATE | Reverse pivot: "who is in process area X" — toggle inside MembershipsTab | PR-5 |
| `frontend/apps/web/src/features/iam/components/UserMembershipsTable.tsx` | UPDATE | Reuse via `MembershipsDirectory` underneath, OR add a "Manage" link that deep-links into MembershipsTab pre-filtered by user — pick during PR-2 review | PR-2 |
| `wiki/modules/iam.md` | UPDATE | §5.3 route-truth-table rows for the three ops (now with operationIds, Aligned, Contracted); bump `Last verified:`; mention 7-tab Admin Center IA | PR-2 + PR-3 |
| `wiki/modules/iam-tech-debt.md` | UPDATE | Mark T-007 CLOSED with resolution pointer | PR-3 |
| `wiki/decisions/0021-area-membership-governance-logger.md` (NEW ADR) | CREATE | Decision: governance logger writes to `governance_events` from app-service path; SECURITY DEFINER path remains for e2e seeding; both share same row schema | PR-3 |

## NOT Building
- ❌ Bulk grant/revoke (CSV import) — out of v1 scope; record in `iam-tech-debt.md` as `T-MEMB-1` after PR-2 lands.
- ❌ Per-membership edit (change role on an active grant) — current model is revoke-then-grant; surface as a single "Modify" dialog only if the user community asks (record as `T-MEMB-2`).
- ❌ Group-derived membership UI (`iam_group_*`) — group plumbing exists in DB (migration 0163) but no admin write surface; out of scope.
- ❌ Membership expiration scheduling (`effective_to` set in future) — current API does not accept it; defer.
- ❌ Audit Tab integration — Memberships actions will already land in the audit feed via PR-1's `recordMembershipAudit`; no new Audit Tab work needed.
- ❌ Tier-A platform-owner membership ops — Tier-B (tenant admin) only.
- ❌ `tenant_id` column rename or schema change — runtime is fine.
- ❌ Wholesale rewrite of `routes_memberships.go` to oapi-codegen-generated server stubs — IAM is still pre-codegen on the BE side per ADR 0012 partial rollout; mirror PeopleHandler's hand-rolled-with-spec-tagged pattern. Hard stop if reviewers push for stub-gen — record as future-ADR.

---

## Step-by-Step Tasks

### PR-1 — Contract truth + BE polish + audit/tenant-guard

#### Task 1.1 — Flesh out OpenAPI spec
- **ACTION**: Replace lines 2103-2146 of `api/openapi/v1/openapi.yaml` with a full spec: rename path to server-relative `/iam/area-memberships` (matches all other ops in the doc — current `/api/v1/iam/area-memberships` is the only outlier); add component schemas `AreaMembership`, `AreaMembershipListResponse`, `GrantAreaMembershipRequest`, `GrantAreaMembershipResponse`, `RevokeAreaMembershipRequest`; add query parameters (`userId`, `areaCode`, `role`, `cursor`, `limit`) to GET; add all error responses (400/403/404/409/500) referencing `ApiErrorEnvelope`.
- **IMPLEMENT**: Schemas mirror `membershipDTO` shape from PR #55 (`userId`, `tenantId`, `areaCode`, `role`, `effectiveFrom`, `effectiveTo`, `grantedBy`).
- **MIRROR**: shape of `/iam/users/{userId}/memberships` (openapi.yaml:415).
- **IMPORTS**: N/A — YAML.
- **GOTCHA**: Server base is `/api/v1`; do NOT double-prefix.
- **VALIDATE**: `corepack pnpm gen:api` (or `scripts/gen-frontend-types.*`) succeeds and `lib/api-types/index.d.ts` exports `paths["/iam/area-memberships"]`. `grep -n "areaMembership" frontend/apps/web/src/lib/api-types/index.d.ts` returns the new operationIds.

#### Task 1.2 — Permissions resolver routeRules
- **ACTION**: In `apps/api/cmd/metaldocs-api/permissions.go`, add explicit entries for the three ops so they never fall through to `VisibilitySessionRequired` (mirror PR-7b PeopleHandler fix).
- **IMPLEMENT**: GET → `CapMembershipView`, POST → `CapMembershipManage`, DELETE → `CapMembershipManage`.
- **MIRROR**: existing PeopleHandler routeRules lines (post-T-PR7B-3).
- **GOTCHA**: order matters — most-specific path first; add a `TestPermissionResolver_AreaMembershipRoutes` lock test.
- **VALIDATE**: `go test ./apps/api/cmd/metaldocs-api/... -run PermissionResolver_AreaMembership` green.

#### Task 1.3 — Handler hardening (audit + tenant guard + 409)
- **ACTION**: In `internal/modules/iam/delivery/http/routes_memberships.go`: add `guardMembershipUserInTenant`; on duplicate grant map `UNIQUE` violation to 409 `MEMBERSHIP_EXISTS`; add `recordMembershipAudit` helper called after successful grant + revoke; emit `iam.area_membership.granted` / `iam.area_membership.revoked` actions.
- **IMPLEMENT**: Use `audit.Writer.Record` (passed via constructor — extend `NewMembershipHandler` signature to accept it, wire in `main.go`).
- **MIRROR**: `PeopleHandler.handleListMemberships` cross-tenant guard + `AdminHandler.recordAudit` style.
- **IMPORTS**: `metaldocs/internal/modules/audit/domain` for `Writer`.
- **GOTCHA**: `recordAudit` MUST not fail the request — log-and-continue. Trace ID via `r.Context()` (already in `problem.New`).
- **VALIDATE**: integration tests in 1.4 pass; manual `curl` shows 409 on duplicate, 404 on cross-tenant probe.

#### Task 1.4 — Go integration tests
- **ACTION**: Create `tests/integration/iam/area_memberships_handler_test.go`.
- **IMPLEMENT**: 5 tests as listed under Files to Change. Use the existing test harness (`tests/integration/testenv`).
- **MIRROR**: `tests/integration/iam/people_handler_*.go` pattern.
- **VALIDATE**: `go test -race -run TestArea ./tests/integration/iam/...` green.

#### PR-1 acceptance evidence
- Commands: `corepack pnpm gen:api`; `go test -race ./internal/modules/iam/... ./tests/integration/iam/... ./apps/api/cmd/metaldocs-api/...`; `corepack pnpm tsc --noEmit -p tsconfig.build.json`.
- Runtime proof: `curl -i POST /iam/area-memberships` with valid + duplicate + bad-role + cross-tenant → 201 / 409 / 400 / 404 respectively, each with `Content-Type: application/problem+json` (except 201).
- Persisted: `SELECT * FROM audit_events WHERE action LIKE 'iam.area_membership.%'` shows two rows; `SELECT * FROM user_process_areas WHERE ...` confirms state.
- Wiki: bump `Last verified:` on `wiki/modules/iam.md`, refresh §5.3 route table with operationIds.
- **Branch**: `feat/iam-memberships-pr1-contract` → PR to `main`.

---

### PR-2 — Frontend rebuild + 7th tab IA

**Prereq:** Phase 0 decision = A confirmed.

#### Task 2.1 — Query keys + hooks
- **ACTION**: Add membership keys to `lib/queryKeys.ts`; create three hooks under `features/iam/queries/` and two mutations under `features/iam/mutations/`.
- **IMPLEMENT**: Mirror `useUserMembershipsQuery.ts` literally; mutations invalidate `QK.iam.memberships.list({})` on success.
- **MIRROR**: `useUsersQuery.ts`, `useBulkUsersMutation.ts`.
- **GOTCHA**: `useGrantMembershipMutation` and `useRevokeMembershipMutation` must surface RFC 9457 errors via `resolveErrorMessage`; surface to UI via sonner.
- **VALIDATE**: `corepack pnpm exec vitest run src/features/iam/queries/__tests__/useMembershipsQuery.test.tsx`.

#### Task 2.2 — Presentational components
- **ACTION**: Create `MembershipsDirectory`, `MembershipsFilterBar`, `GrantMembershipDialog`, `RevokeMembershipDialog`, `MembershipKpiStrip`.
- **IMPLEMENT**: Pure presentational — props in, events out. CSS Modules + tokens, no inline styles. Use `components/ui/{Dialog,Input,Select,Button,Table}` primitives.
- **MIRROR**: `UsersDirectory.tsx`, `InviteUserDialog.tsx`, `KpiCard.tsx`, `UsersFilterBar.tsx`.
- **GOTCHA**: ROLES list MUST come from a single source — extract `IAM_AREA_ROLES` const in `features/iam/constants.ts` (mirror existing `IamRole` union); never re-declare in dialog.
- **VALIDATE**: Storybook-less; verify by snapshot test on `MembershipsTab.test.tsx`.

#### Task 2.3 — Container `MembershipsTab`
- **ACTION**: Container owns `searchParams` (q, areaCode, role, sort, dir, cursor, limit), calls `useMembershipsQuery`, `useMembershipsKpiQuery`, the two mutations, manages dialog open state.
- **IMPLEMENT**: Mirror `PeopleTab.tsx` literally — VALID_ROLE/VALID_AREA enums, `readEnum` helper, sonner toast on success/failure, cap-gated `+ Conceder` button via `useHasCapability("membership.manage")`.
- **MIRROR**: `PeopleTab.tsx`.
- **GOTCHA**: Pagination via cursor — backend already returns `nextCursor` if PR-1 added it; if not, gate this behind a config and degrade to limit-only paging until PR-4.
- **VALIDATE**: Manual Preview-driven QA per `wiki/quality/screen-qa-checklist.md`.

#### Task 2.4 — Route + AdminCenter IA
- **ACTION**: Add `memberships` tab to `AdminCenterPage.tsx`; add child route inside the `admin/*` children block in `routes.tsx`; DELETE orphan top-level `admin/memberships` route block; DELETE legacy `AreaMembershipAdminPage.tsx`, `AreaMembershipAdminRoutePage.tsx`, `MembershipGrantDialog.tsx`, `membershipApi.ts`.
- **IMPLEMENT**: Capability gate matches `roles` tab → `requiresCapability: "membership.view"`.
- **MIRROR**: how `roles` is added in `routes.tsx:31-35`.
- **GOTCHA**: AppShell.test.tsx had test cases pinning the orphan path — update to match new IA. Ensure deep-link `/admin/memberships` still resolves (now under AdminCenter parent).
- **VALIDATE**: `corepack pnpm exec vitest run src/features/shell/components/AppShell.test.tsx`.

#### Task 2.5 — FE tests
- **ACTION**: Add `MembershipsTab.test.tsx` (integration), `useMembershipsQuery.test.tsx` (unit).
- **IMPLEMENT**: Use the existing render util + msw or `setupServer` if present; otherwise mock `api.GET`/`api.POST`.
- **MIRROR**: existing `queries/__tests__/` tests.
- **VALIDATE**: `corepack pnpm exec vitest run src/features/iam/{tabs,queries}/`.

#### PR-2 acceptance evidence
- Commands: `corepack pnpm tsc --noEmit -p tsconfig.build.json`; `corepack pnpm exec vitest run`; `pnpm exec eslint --max-warnings 0 src/features/iam`.
- Preview runtime proof (per `screen-qa-checklist`): initial load, happy-path grant persists (refetched list shows new row), revoke persists, validation error on bad role, server 409 surfaces as toast, network failure recoverable, refresh re-entry preserves URL state, admin sees `+ Conceder`, viewer-only role doesn't, area-scoping returns correct rows under tenant isolation, no console errors, no mojibake.
- Persisted: `SELECT * FROM user_process_areas WHERE tenant_id = … AND user_id = …` reflects the UI grant/revoke.
- Wiki: bump `Last verified:` on `iam.md` with PR-2 close-out note.
- **Branch**: `feat/iam-memberships-pr2-frontend` → PR to `main`.

---

### PR-3 — Close T-007: real `MembershipGovernanceLogger`

#### Task 3.1 — Implement governance logger
- **ACTION**: Create `internal/modules/iam/infrastructure/postgres/governance_logger.go` implementing the existing `MembershipGovernanceLogger` interface.
- **IMPLEMENT**: Insert into `governance_events` (same table the SECURITY DEFINER funcs write to today — confirm column shape via migration grep). Emit one row per grant + per revoke with `{actor, tenant, type='area_membership.granted'|'.revoked', target_user, area_code, role, occurred_at, trace_id}`.
- **MIRROR**: existing audit writer pattern (`audit/infrastructure/postgres/writer.go`).
- **GOTCHA**: Reuse the same row schema the SECURITY DEFINER funcs write so downstream consumers (auditors, reports) don't see two formats. Verify by reading `migrations/0137_*.sql` (grant_area_membership func definition).
- **VALIDATE**: `go test ./internal/modules/iam/infrastructure/postgres/... -run Governance` green.

#### Task 3.2 — Wire into main
- **ACTION**: Replace `main.go:217` `nil` with `iampg.NewMembershipGovernanceLogger(deps.SQLDB)`.
- **IMPLEMENT**: One-line change.
- **GOTCHA**: Make the logger nil-safe at the service callsite still — if startup wiring fails, fall back to noop and log a `slog.Warn`. Do NOT crash boot.
- **VALIDATE**: Boot the API; grant a membership via PR-2 UI; `SELECT * FROM governance_events ORDER BY occurred_at DESC LIMIT 5` shows the new row.

#### Task 3.3 — Integration test
- **ACTION**: Create `tests/integration/iam/membership_governance_test.go`.
- **IMPLEMENT**: Grant via the application service path; assert one `governance_events` row written. Revoke; assert second row. Ensure SECURITY DEFINER path is NOT triggered (the test calls the Go API, not the SQL func).
- **VALIDATE**: `go test -race -run Governance ./tests/integration/iam/...`.

#### Task 3.4 — ADR + wiki + tech-debt close
- **ACTION**: Create `wiki/decisions/0021-area-membership-governance-logger.md`; mark T-007 CLOSED in `iam-tech-debt.md` with resolution pointer.
- **IMPLEMENT**: ADR explains: app-service path now emits parity rows; SECURITY DEFINER path retained for e2e seed; both share schema; if they diverge, the seed path is canonical (auditor view).
- **VALIDATE**: `wiki-curator` agent dispatched; `Last verified:` on `iam.md` bumped.

#### PR-3 acceptance evidence
- Commands: `go test ./...`; manual `curl` grant → `SELECT FROM governance_events`.
- Persisted proof: row count delta = +1 per grant, +1 per revoke; both rows share schema with SECURITY DEFINER-emitted rows from `migrations/0137`.
- Wiki: T-007 closed; ADR 0021 published; `iam.md` §11 summary count decrement (Major 2 → 1).
- **Branch**: `feat/iam-memberships-pr3-governance` → PR to `main`.

---

### PR-4 — Cursor pagination + KPI parity (optional polish)

#### Task 4.1 — Server-side cursor
- **ACTION**: Extend `AreaMembershipService.ListActive` and the handler to accept `cursor` + `limit`; return `{items, nextCursor}` envelope. Mirror PR-12b cursor shape (410 `CURSOR_EXPIRED` on stale anchor).
- **VALIDATE**: integration test for cursor round-trip + stale-cursor 410.

#### Task 4.2 — KPI endpoint
- **ACTION**: Add `GET /iam/area-memberships/kpi` returning total-active + last-7d-grants/-revokes for the current tenant.
- **IMPLEMENT**: Tier-1 cap `membership.view`.
- **VALIDATE**: `useMembershipsKpiQuery` lights up the strip in the UI.

#### PR-4 acceptance evidence
- KPI strip renders with real numbers; pagination "next" loads next page; stale-cursor banner appears with refresh CTA.
- **Branch**: `feat/iam-memberships-pr4-pagination` → PR to `main`.

---

### PR-5 — By-Area pivot view (optional Phase-A enrichment)

#### Task 5.1 — Reverse pivot
- **ACTION**: Add `MembershipsByAreaView` — toggle inside the tab: "By user" (default) ⇄ "By area" (group rows by `areaCode`, show users-per-area).
- **IMPLEMENT**: Reuse `useMembershipsQuery` with `groupBy=area` param OR client-side group (decide during implementation based on dataset size — area count is bounded by taxonomy so client-side group is fine).
- **VALIDATE**: Preview QA: switch view → correct grouping, manage actions still cap-gated, deep-link state preserved in URL.

#### PR-5 acceptance evidence
- Preview screenshots of both views; URL state proof; no extra network calls when toggling (cache hit).
- **Branch**: `feat/iam-memberships-pr5-by-area` → PR to `main`.

---

### PR-6 — Close-out hardening (mirrors PR-12b style)

#### Task 6.1 — Quality-gate sweep
- **ACTION**: Run full QA loop: vitest, tsc, eslint, `go test -race ./...`, Preview-driven `screen-qa-checklist`. Run `code-reviewer`, `security-reviewer`, `accessibility` reviewers in parallel; root-cause-by-family fix loop.
- **VALIDATE**: classify every finding (severity + family); fix or HARD-STOP each.

#### Task 6.2 — Wiki sync
- **ACTION**: Dispatch `wiki-curator`; refresh `Last verified:`; ensure §5.3 route table fully aligned with the OpenAPI spec; update Failure modes table.

#### Task 6.3 — Final PR
- **ACTION**: Roll up findings + fixes into close-out PR. Match PR-12b format.

#### PR-6 acceptance evidence
- Same as IAM Admin Center close-out (PR-12b): N findings closed with severity tags + family, evidence per finding (commands + Preview + persisted), wiki bumped, no orphan TODO.
- **Branch**: `feat/iam-memberships-pr6-closeout` → PR to `main`.

---

## Testing Strategy

### Unit Tests
| Test | Input | Expected | Edge? |
|---|---|---|---|
| `useMembershipsQuery` keyless | `{}` | `QK.iam.memberships.list({})` | – |
| `useMembershipsQuery` byUser | `{userId: 'u1'}` | key + GET with `?userId=u1` | – |
| `useGrantMembershipMutation` error | server returns 409 problem | `toast.error` with mapped message | yes |
| `GrantMembershipDialog` validation | empty user | dialog stays open, inline error | yes |
| `MembershipsDirectory` empty | `items=[]` | empty-state component | yes |
| `MembershipsTab` cap-gate | user caps `["membership.view"]` | no `+ Conceder` button rendered | yes |
| `MembershipsTab` cap-gate | user caps include `membership.manage` | `+ Conceder` rendered | – |
| `AppShell` redirect | only `user.view` | navigates to `/` | yes |

### Integration Tests (FE)
- `MembershipsTab.test.tsx`: render with admin caps → grant flow → toast + list refetched.
- `MembershipsTab.test.tsx`: render with view-only caps → no manage UI, table read-only.

### Integration Tests (BE)
- `TestListAreaMemberships_ContractShape`: response matches OpenAPI schema (lowerCamel).
- `TestGrantMembership_EmitsAuditAndGovernance`: 1 row in `audit_events` + 1 row in `governance_events`.
- `TestRevokeMembership_RejectsCrossTenantUserWith404`: PR-12b parity.
- `TestGrantMembership_DuplicateReturns409`: UNIQUE violation maps to 409 `MEMBERSHIP_EXISTS`.
- `TestListAreaMemberships_AreaScopedUnderTenantIsolation`: tenant A's memberships invisible from tenant B context.
- `TestPermissionResolver_AreaMembershipRoutes`: lock routeRules.

### Edge Cases Checklist
- [ ] Empty result set → empty state
- [ ] Pagination boundary (cursor expired) → 410 banner
- [ ] Network failure mid-grant → toast + dialog stays open
- [ ] Concurrent grant of same (user, area) → 409 (second client)
- [ ] Revoke of already-revoked → 404 with NO-OP toast
- [ ] Cross-tenant probe → 404 (not 403; per PR-12b)
- [ ] User has membership.view but not .manage → manage UI hidden, route still accessible
- [ ] Refresh during pending mutation → query invalidation on remount
- [ ] Unicode in `areaCode` / `userId` → URL encoding correct
- [ ] Stale cache after grant → invalidation hits
- [ ] Native browser refresh on `/admin/memberships?q=foo` → state preserved
- [ ] Logout mid-page → redirected to login

---

## Validation Commands

### Static Analysis
```powershell
cd frontend\apps\web
corepack pnpm tsc --noEmit -p tsconfig.build.json
corepack pnpm exec eslint --max-warnings 0 src/features/iam
```
EXPECT: zero errors / zero warnings.

### Unit + Integration Tests (FE)
```powershell
corepack pnpm exec vitest run src/features/iam src/features/shell/components/AppShell.test.tsx
```
EXPECT: all green.

### Backend Tests
```bash
go test -race ./internal/modules/iam/... ./tests/integration/iam/... ./apps/api/cmd/metaldocs-api/...
```
EXPECT: all green.

### Contract Regen
```powershell
corepack pnpm gen:api   # or repo's script
```
EXPECT: `frontend/apps/web/src/lib/api-types/index.d.ts` updated; git diff shows only the three new operations.

### Browser Validation
```powershell
.\scripts\start-api.ps1 -Build
corepack pnpm --filter @metaldocs/web dev -- --port 4173 --strictPort
```
Drive `screen-qa-checklist.md` via Preview tools (`preview_*`). NOT Playwright. NOT Chrome MCP.

### Manual Validation Checklist (per PR)
- [ ] tsc green
- [ ] vitest green (FE)
- [ ] go test green (BE)
- [ ] eslint clean
- [ ] OpenAPI spec lint (`spectral lint api/openapi/v1/openapi.yaml` if available)
- [ ] Preview QA full pass per `screen-qa-checklist`
- [ ] `wiki/modules/iam.md` Last verified bumped
- [ ] PR description matches PR-12b format

---

## Acceptance Criteria

### Per-PR
- All static + test gates green.
- Evidence recorded in PR description (commands + Preview + persisted/API).
- No HIGH/CRITICAL findings open at PR merge.
- Wiki bumped where code truth changed.

### Feature Close-Out (after PR-6)
- 7-tab Admin Center IA includes Memberships.
- `useState`-only legacy stub deleted, no inline styles, no hand-rolled types.
- T-007 closed in `iam-tech-debt.md`.
- Audit + governance rows written on every grant + revoke from the application-service path.
- `wiki/modules/iam.md` §5.3 lists three ops as Aligned + Contracted with operationIds.
- ADR 0021 published.
- Per-screen QA report (template below) attached to the close-out PR.

---

## Per-Screen QA Report (close-out template)

```markdown
# QA Report — Area Membership Admin Rebuild (feat/iam-memberships-rebuild)
- Route: /admin/memberships (Admin Center tab 7)
- Page: features/iam/tabs/MembershipsTab.tsx
- Owning module: iam
- QA class: screen + authz (admin + area scoping)
- Acting roles: system_admin (allow); approver, area_admin (view-only); viewer (redirect)

## Gate results
- Gate 0 scope: pass — IA decision recorded; runtime/startup truth confirmed
- Gate 1 impl: tsc green; vitest green; go test green
- Gate 3 product QA: full screen-qa-checklist pass via Preview
- Gate 5 regression: cross-tab smoke pass (People + Roles + Audit)

## Findings (severity-ordered)
| # | Severity | Family | Finding | Disposition |
|---|---|---|---|---|

## Evidence
- Commands: <list>
- Runtime: <Preview screenshots + snapshots>
- Persisted/API: <SELECT + curl outputs>

## Hard-stops / Defers
- <none>, or <link>
```

---

## Risks

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| `gen:api` script differs / unknown invocation | Medium | Med | PR-1 Task 1.1 includes "locate the script" verification; if absent, document command and add to repo scripts. |
| Existing OpenAPI path `/api/v1/iam/area-memberships` (with double prefix) is referenced elsewhere | Low | Med | Grep before renaming; `lib/api-types/` is auto-generated so only the spec entry matters. |
| `governance_events` row schema mismatch between SECURITY DEFINER funcs and new Go logger | Med | High | PR-3 Task 3.1 explicitly inspects migration 0137 to copy the exact column shape before writing. Integration test asserts shape parity. |
| Cap-gate viewer-vs-manage regressions in Admin Center cap chain | Low | Med | AppShell.test.tsx already locks redirect cases; PR-2 Task 2.4 updates them. |
| `corepack pnpm` script names differ across PRs | Low | Low | `package.json scripts` survey in PR-1; documented in this plan. |
| Hidden consumer of legacy `membershipApi.ts` | Low | Med | grep before deletion in PR-2 Task 2.4. |
| Backend handler change to `recordMembershipAudit` breaks existing callers | Very low | Med | The handler is only mounted by `MembershipHandler.RegisterRoutes`; no other Go callers. |

---

## Notes
- This rebuild deliberately stays inside the existing tier-1/tier-2/tripwire authz framework — no cross-module authz model change → no hard-stop expected.
- Branch `qa/iam-area-membership` (PR #55) is the prerequisite baseline. Merge that first, then cut `feat/iam-memberships-rebuild` from `main`.
- The rebuild does NOT introduce new architectural decisions beyond ADR 0021 (governance logger wiring). All other decisions reuse existing ADRs (0007, 0012, 0016).
- If the user picks Phase 0 Option B, drop PR-5 entirely and demote PR-2 Task 2.4 to "render in standalone shell".
- All PRs target `main`. Use the IAM Admin Center close-out (PR-12b) as the merge-quality bar.

## Confidence Score
**8 / 10** — single-pass implementable. Two unknowns:
1. Exact `gen:api` script invocation (verifiable in 1 grep).
2. `governance_events` table column shape (verifiable in 1 migration read).
Both resolved inside PR-1 Task 1.1 / PR-3 Task 3.1 respectively. Everything else is direct mirroring of existing IAM Admin Center patterns.
