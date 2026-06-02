# Module: iam

> Living architecture doc. Shape: Arc42 + C4 + ADR cross-links.

**Last verified:** 2026-06-02 (PR-2: added `CapSessionManage`; broadened `CapAuditRead` grants to qms_admin/area_admin/approver via migration 0210; registry size 27 → 28; ADR 0019 documents the read-only-by-design naming exception. Prior verification line preserved below.)

**Prior verification — 2026-06-02 (PR-1):** T-005 (audit emission) + T-006 (RFC 9457 envelope) closed; §5.3 + §6.2 + §6.3 + §6.4 cap columns refreshed against `apps/api/cmd/metaldocs-api/permissions.go:108-123` (ADR 0016 view-grade split). | **Owner:** unassigned | **Status:** active | **Maturity:** L2

**Prior verification — 2026-06-01:** Approval route admin PR-3 spike: confirmed `GET /api/v1/iam/roles` does **not** exist today — role catalogue is hard-coded in `internal/modules/iam/domain/model.go:10-16` (`approver`, `author`, `editor`, `system_admin`, `viewer`). Proposed endpoint shape `{roles:[{code,label}]}` gated by `CapMembershipView` is documented in [ADR 0018 §"IAM roles source"](../decisions/0018-approval-route-lifecycle.md); implementation deferred to PR-4 of approval route admin work or its own micro-PR. Frontend hard-codes the same list at `frontend/apps/web/src/features/approval/pages/RouteAdminPage.tsx:10` as `STAGE_ROLES`.

**Prior verification — 2026-06-01:** ADR 0016: added view-grade caps `metrics.view`, `membership.view`, `user.view`, `taxonomy.view`; registry now 24 consts; prior: P2 consolidation: §3/§5 C4 fragments tagged as module-scoped with pointer to canonical diagrams; added Failure modes section; 2026-05-26 Wave 2 authz tx seeding sync) | **Owner:** unassigned | **Status:** active (partial contract-first; defense-in-depth now two-layer on IAM writes) | **Maturity:** L2

> **Key files:**
> - `internal/modules/iam/application/capability_service.go:31` â€” tier-1 `CanDo` (DB-backed EXISTS over 4 branches)
> - `internal/modules/iam/authz/authz.go:44` â€” tier-2 `Require` (in-tx area check); system_admin bypass `:58`
> - `internal/modules/iam/authz/context.go:13` â€” `ErrActorContextMissing` / `ErrTenantContextMissing`; `MustActorID` `:21`, `MustTenantID` `:34`
> - `internal/modules/iam/delivery/http/middleware.go:31` â€” `NewMiddleware`; `Wrap` `:49`; strips trusted identity headers `X-User-ID` and `X-User-Roles`; tenant resolution: prefers `tenant.FromContext`, falls back to `X-Tenant-ID` header then `DevTenantID` (legacy-header mode only)
> - `internal/modules/iam/delivery/http/middleware.go:77` â€” `tenant.FromContext` call (primary tenant source post-Plan 3)
> - `internal/modules/iam/delivery/http/admin_handler.go:82` â€” admin routes registration; `handleAdminOverview` reads `tenant.FromContext` at `:109`
> - `internal/modules/iam/delivery/http/routes_memberships.go:30` â€” area-membership routes; `tenantIDFromRequest` delegates to `tenant.FromContext` at `:146`
> - `internal/modules/iam/application/admin_service.go:23` â€” `UpsertUserAndAssignRole`
> - `internal/modules/iam/application/area_membership_service.go:49` â€” `Grant` use case
> - `internal/modules/iam/application/cached_role_provider.go:18` â€” TTL-cached role provider
> - `internal/modules/iam/infrastructure/postgres/role_provider.go:19` â€” tenant-scoped `RolesByUserID`
> - `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:34` â€” `UpsertUserAndAssignRole` (BEGIN-authz.Require-DELETE-INSERT); `ReplaceUserRoles` at `:76` â€” both now call `authz.Require(CapUserManage)` (Plan 5 closed T-004 partial)
> - `internal/modules/iam/infrastructure/postgres/user_area_repository.go:52` â€” `Insert` (tier-2 `authz.Require(CapMembershipManage)`); `CloseActive` at `:84`; `GrantAtomic` at `:109` â€” all three write methods now have tier-2 enforcement (Plan 5 closed T-004 partial)
> - `internal/modules/iam/domain/model.go:13` â€” single typed `Capability` namespace (28 consts; ADR 0019 (PR-2) added `CapSessionManage` and broadened `CapAuditRead` grants; ADR 0016 added `CapMetricsView`, `CapMembershipView`, `CapUserView`, `CapTaxonomyView`; Plan 5 added `CapControlledDocumentObsolete` + `CapControlledDocumentSupersede`; Plan 4 closed T-001; size locked by `TestCapabilityRegistrySize`)
> - `db/migrations/0218_iam_caps_audit_session_pr2.sql` â€” PR-2 grants: `audit.read` → {qms_admin, area_admin, approver}; `session.manage` → {system_admin}; ADR 0019
> - `apps/api/cmd/metaldocs-api/permissions.go:54,196` â€” `(method,path)â†’Cap*` resolver
> - `apps/api/internal/wiring/documents.go:24` â€” `NewCapabilityChecker` adapter (J2 fix)
> - `migrations/0142b_role_capabilities_v2_enforce.sql:67-179` â€” original `enforce_capability_asserted()` function (approval tables only; superseded by 0188)
> - `migrations/0188_tripwire_extend.sql:18` â€” Plan 5: extended `enforce_capability_asserted()` covering 12 tables; trigger attachments at `:186-233`
> - `internal/platform/tenant/const.go:4` â€” `DevTenantID` sentinel

---

## 1. Introduction & Goals

IAM owns identity-derived authorization for MetalDocs: it answers "can user X perform capability C in tenant T (and area A, when area-scoped)?". It does NOT authenticate (login/JWT) â€” that lives in `internal/modules/auth/`. IAM publishes the role catalogue, capability catalogue, roleâ†”capability map, tenant-scoped role assignments, area-scoped membership grants, HTTP middleware that enforces tier-1, in-transaction helpers that enforce tier-2, and an admin HTTP surface for managing users + roles + memberships. The Postgres tripwire trigger that backs both tiers at the database layer is owned conceptually by IAM (migration 0142b) even though it attaches to approval tables.

### 1.1 Requirements overview

- **Tenant-scoped capability checks at the HTTP edge** â€” drivers: regulated QMS (ISO 9001) needs per-tenant role boundaries; source: `wiki/decisions/0007-two-tier-authz.md`.
- **Area-scoped grants for QMS process areas** â€” drivers: ISO segregation per process area; source: `wiki/concepts/iso-segregation.md`, `wiki/decisions/0007-two-tier-authz.md`.
- **DB-layer enforcement floor** â€” drivers: bug J2 (`permissiveAuthzChecker` returned nil); source: ADR 0007 amendment (J2) + `migrations/0142b_role_capabilities_v2_enforce.sql`.
- **Group-derived grants** â€” drivers: scale role admin to teams; source: migration 0163.
- **Tenant isolation by row** â€” drivers: multi-tenant data hygiene; source: Group B fix (audit 2026-05-03 B5/B6), migration 0162.

### 1.2 Quality Goals

| Rank | Goal | How verified |
|---|---|---|
| 1 | Correctness (no path bypasses authz) | `internal/modules/iam/authz/authz_bypass_test.go`; `tests/integration/iam/capability_service_test.go`; CI lint `authz-call-present` + `tripwire-pairing` (`scripts/api-lint/code_rules.go`) |
| 2 | Tenant isolation | `tests/integration/iam/tenant_isolation_test.go`; unique constraint `ux_iam_users_tenant_user`; `RolesByUserID` filters `tenant_id` (`role_provider.go:19`) |

### 1.3 Stakeholders

| Role | Expectation |
|---|---|
| End user | Cannot perform an op the role does not grant; gets a structured error |
| Operator (system_admin) | Can assign/replace user roles; can grant/revoke area memberships; mutations are audit-traceable |
| Developer | One way to ask "can this user?" at the edge (`CapabilityService.CanDo`) and one way to enforce inside a tx (`authz.Require`); typed errors for missing context |
| Auditor (ISO) | Per-area grant table with `effective_from`/`effective_to`; segregation enforced; capability matrix versioned |

---

## 2. Architecture Constraints

- Language: Go 1.25; stdlib `net/http`, `database/sql`, `chi`-free.
- Persistence: Postgres; tables under schemas `metaldocs.*` (admin) and `public.*` (process-area + governance).
- Authz model: two-tier per ADR 0007; system_admin bypass on both tiers.
- DB enforcement floor: `metaldocs.asserted_caps` GUC + `enforce_capability_asserted` trigger. Plan 5 migration 0188 expanded coverage to 10 tables: `approval_instances`, `approval_signoffs`, `iam_user_roles`, `user_process_areas`, `documents`, `controlled_documents`, `cd_sequence_counters`, `document_profiles`, `document_process_areas`, `document_families`, `templates_template`, `templates_template_version`.
- IAM is NOT under oapi-codegen yet (ADR 0012 documents the partial rollout). Membership routes have no `operationId`; admin POST `/api/v1/iam/users/{userId}/roles` has request/response schemas (`api/openapi/v1/openapi.yaml:5043,5054`) but no codegen stub.
- Error envelope: IAM emits RFC 9457 Problem responses via `internal/platform/problem/problem.go` (`problem.Write` sets `Content-Type: application/problem+json`); admin handler and membership handler both route through `problem.New(...)`. Matches `wiki/architecture/api-design-system.md`. T-006 closed (Plan 7, 2026-05-12).
- Tenant sentinel: `DevTenantID = "ffffffff-ffff-ffff-ffff-ffffffffffff"` (`internal/platform/tenant/const.go:4`). Primary tenant source is `tenant.FromContext` (injected by auth middleware from `auth_sessions.tenant_id`). IAM middleware falls back to `X-Tenant-ID` header only in legacy-header mode; it strips trusted identity headers `X-User-ID` and `X-User-Roles` before downstream handlers. See `wiki/architecture/tenant-context.md` for the full pattern.

---

## 3. System Scope & Context — module-scoped (C4 Level 1)

> System-level context lives in [`wiki/diagrams/c4-context.md`](../diagrams/c4-context.md). The diagram below is **module-scoped**: it shows iam's relationship with tier-2 consumers (documents/approval/templates), auth, audit, and the platform packages that read its capability/tenant context.

```mermaid
C4Context
    title System Context — iam (module-scoped)
    Person(actor, "End user / admin", "HTTP client")
    System_Boundary(b1, "MetalDocs API") {
        System(iam, "iam", "Capabilities, roles, memberships, authz")
        System(docs, "documents/approval/templates", "Tier-1 + tier-2 consumers")
        System(auth, "auth", "Login + session; owns ManagedUser")
        System(audit, "audit", "Event sink")
        System(platform, "platform/{authn,httpresponse,tenant,bootstrap,security}", "Cross-cutting")
    }
    SystemDb(db, "Postgres", "iam_users, iam_user_roles, iam_groups*, role_capabilities, user_process_areas, governance_events")
    Rel(actor, iam, "HTTP /api/v1/iam/* + /api/v1/iam/*")
    Rel(iam, docs, "Tier-1 middleware wraps all routes")
    Rel(docs, iam, "Tier-2 authz.Require in approval/finalize tx")
    Rel(auth, iam, "Imports iamdomain.Role / WithAuthContext")
    Rel(iam, audit, "AdminHandler.recordAudit â†’ audit.Writer")
    Rel(iam, platform, "tenant.DevTenantID, authn.UserIDFromContext, httpresponse.WriteJSON")
    Rel(iam, db, "SQL: roles, memberships, capability check")
```

### 3.1 Business Context

Auditors verify three things per controlled document: (a) only authorised roles act on it (tier-1), (b) per ISO 9001 the actor's role is bound to the document's process area (tier-2), (c) the submitter cannot sign-off their own work (SoD, enforced in approval, not here). IAM owns (a) and (b). The roleâ†”capability matrix is the contract operators rely on; migrations 0165 (reseed) and 0169 (process-area backfill) define the current 40-row matrix.

### 3.2 Technical Context

Inbound HTTP routes (own): see Â§5.3. Inbound Go imports: 17 importers (`apps/api/cmd/metaldocs-api`, `apps/api/internal/wiring`, `internal/modules/documents/**` including `documents/approval`, `internal/modules/templates/delivery/http`, `internal/modules/auth/**`, `internal/platform/{bootstrap,authn,security}`, `internal/testsupport/http`). See `_artifacts/03-deps.md` Â§2 for full table.

Outbound Go imports: 5 packages â€” `internal/modules/audit/domain`, `internal/modules/auth/domain`, `internal/platform/authn`, `internal/platform/httpresponse`, `internal/platform/tenant`.

Outbound DB writes (owned): `iam_users`, `iam_user_roles`, `iam_groups`, `iam_group_members`, `iam_group_roles`, `role_capabilities` (read-only at runtime; reseeded by migrations), `user_process_areas`. Reads (not owned): `governance_events`, `document_process_areas` (taxonomy, via FK).

---

## 4. Solution Strategy

- **Two distinct authz tiers**, not a unification gap â€” driver: `wiki/decisions/0007-two-tier-authz.md`. Tier-1 reads `iam_user_roles`; tier-2 reads `user_process_areas`. Shared `role_capabilities`.
- **DB tripwire as enforcer of last resort** â€” driver: bug J2 + ADR 0007 codegen-rejected amendment. Originally backed approval tables only (0142b); Plan 5 (migration 0188) extended to 12 tables across all regulated modules.
- **DELETE-then-INSERT for role replacement** â€” driver: idempotent semantics under unique `(tenant_id, user_id)` constraint added in 0166; chosen over upsert to keep `assigned_at` truthful.
- **Tenant-scoped role provider with TTL cache** â€” driver: hot path on every request; `CachedRoleProvider` wraps the Postgres impl with key `userID|tenantID` and explicit `InvalidateUser` on admin writes (`application/admin_service.go:42`).

---

## 5. Building Block View — module-scoped (C4 Level 2)

> System-level container topology lives in [`wiki/diagrams/c4-container-backend.md`](../diagrams/c4-container-backend.md). The diagram below decomposes the internal Go packages of iam (middleware/handlers/services/role provider cache/authz package).

### 5.1 Whitebox — iam

```mermaid
C4Container
    title Container View — iam (module-internal packages)
    Container(mw, "HTTP middleware", "Go stdlib", "Wrap + tier-1 enforcement")
    Container(adminH, "AdminHandler", "Go stdlib mux", "/api/v1/iam/users/* + /api/v1/iam/admin/overview")
    Container(memH, "MembershipHandler", "Go stdlib mux", "/api/v1/iam/area-memberships")
    Container(adminSvc, "AdminService", "Go", "Role upsert/replace orchestration")
    Container(memSvc, "AreaMembershipService", "Go", "Grant/Revoke/ListActive over area memberships")
    Container(capSvc, "CapabilityService", "Go", "Tier-1 CanDo (DB EXISTS)")
    Container(authz, "authz pkg", "Go", "Tier-2 Require + GUC context helpers")
    Container(roleProv, "CachedRoleProvider", "Go", "TTL-cached wrapper over Postgres RoleProvider")
    ContainerDb(db, "Postgres", "Postgres", "iam_users, iam_user_roles, iam_groups*, role_capabilities, user_process_areas, governance_events")
    Rel(adminH, adminSvc, "")
    Rel(memH, memSvc, "")
    Rel(mw, capSvc, "CanDo on every request")
    Rel(adminSvc, db, "tx: DELETE+INSERT iam_user_roles")
    Rel(memSvc, db, "tx: UPDATE+INSERT user_process_areas")
    Rel(capSvc, db, "SELECT EXISTS across 4 branches")
    Rel(authz, db, "SELECT system_admin or user_process_areas JOIN role_capabilities")
    Rel(roleProv, db, "RolesByUserID")
```

### 5.2 Public surface (by file)

Full table in `_artifacts/01-surface.md` (129 exported symbols). High-level grouping:

| File | Exports |
|---|---|
| `domain/model.go` | `Role` type + 5 consts (`RoleApprover`, `RoleAuthor`, `RoleEditor`, `RoleSystemAdmin`, `RoleViewer`); `Capability` type + 20 typed consts (`CapDocument{View,Create,Edit,Submit,Signoff}`, `CapWorkflow{Review,Approve}`, `CapTemplate{View,Create,Edit,Submit,Approve,Publish}`, `CapControlledDocumentCreate`, `CapControlledDocumentObsolete`, `CapControlledDocumentSupersede`, `CapTaxonomyManage`, `CapMembershipManage`, `CapRouteManage`, `CapUserManage`) â€” Plan 4 collapsed the dual namespace; Plan 5 added `CapControlledDocumentObsolete`/`CapControlledDocumentSupersede` |
| `domain/context.go` | `WithAuthContext`, `UserIDFromContext`, `RolesFromContext` |
| `domain/errors.go` | `ErrUserNotFound`, `ErrUserInactive` |
| `domain/port.go` | `RoleProvider`, `RoleAdminRepository` interfaces (all methods tenant-scoped) |
| `domain/user_area.go` | `UserProcessArea`, `(m).IsActive(now)` |
| `application/capability_service.go` | `ErrCapabilityDenied`, `CapabilityService`, `NewCapabilityService`, `CanDo` |
| `application/admin_service.go` | `RoleCacheInvalidator`, `AdminService`, `NewAdminService`, `UpsertUserAndAssignRole`, `ReplaceUserRoles` |
| `application/area_membership_service.go` | `ErrMembershipNotFound`, `ErrUnknownRole`, `UserAreaWriteRepository`, `MembershipGovernanceLogger`, `AreaMembershipService`, `NewAreaMembershipService`, `ListActive`, `Grant`, `Revoke` |
| `application/cached_role_provider.go` | `CachedRoleProvider`, `NewCachedRoleProvider`, `RolesByUserID`, `InvalidateUser`, `InvalidateAll` |
| `application/dev_role_provider.go` | `DevRoleProvider`, `NewDevRoleProvider`, `RolesByUserID` (memory-mode) |
| `authz/authz.go` | `ErrCapDenied` (struct, carries capability/area fields), `WithCapCache`, `Require`, `BypassSystem` |
| `authz/context.go` | `ErrActorContextMissing`, `ErrTenantContextMissing`, `MustActorID`, `MustTenantID` |
| `delivery/http/middleware.go` | `PermissionResolver`, `Middleware`, `NewMiddleware`, `WithPermissionResolver`, `Wrap` |
| `delivery/http/admin_handler.go` | `UserAdminService` iface, `AdminHandler`, `Upsert/Replace/Create/Update/ResetPassword` request structs, `NewAdminHandler`, `WithAuditReader`, `RegisterRoutes` |
| `delivery/http/routes_memberships.go` | `MembershipHandler`, `NewMembershipHandler`, `RegisterRoutes` |
| `infrastructure/memory/role_admin_repository.go` | `RoleAdminRepository`, `Newâ€¦`, `HasAnyRole`, `UpsertUserAndAssignRole`, `ReplaceUserRoles` (dev-only) |
| `infrastructure/postgres/role_admin_repository.go` | `RoleAdminRepository`, `Newâ€¦`, `HasAnyRole`, `UpsertUserAndAssignRole`, `ReplaceUserRoles` |
| `infrastructure/postgres/role_provider.go` | `RoleProvider`, `Newâ€¦`, `RolesByUserID` |
| `infrastructure/postgres/user_area_repository.go` | `UserAreaRepository`, `Newâ€¦`, `ListActive`, `Insert`, `CloseActive`, `GrantAtomic`, `GetActiveByUserAndArea` |

### 5.3 HTTP operations

| Method | Path | Handler | Tier-1 capability |
|---|---|---|---|
| GET | `/api/v1/iam/area-memberships` | `MembershipHandler.listMemberships` (`routes_memberships.go:30`) | `membership.view` |
| POST | `/api/v1/iam/area-memberships` | `MembershipHandler.grantMembership` (`:31`) | `membership.manage` |
| DELETE | `/api/v1/iam/area-memberships` | `MembershipHandler.revokeMembership` (`:32`) | `membership.manage` |
| GET | `/api/v1/iam/admin/overview` | `AdminHandler.handleAdminOverview` (`admin_handler.go:85`) | `user.view` |
| GET | `/api/v1/iam/users` | `AdminHandler.handleListUsers` (`:88`) | `user.view` |
| POST | `/api/v1/iam/users` | `AdminHandler.handleCreateUser` (`:90`) | `user.manage` |
| POST | `/api/v1/iam/users/{userId}/roles` | `AdminHandler.handleUserRoleUpsert` (`:196`) | `user.manage` |
| PUT | `/api/v1/iam/users/{userId}/roles` | `AdminHandler.handleReplaceUserRoles` (`:198`) | `user.manage` |
| POST | `/api/v1/iam/users/{userId}/reset-password` | `AdminHandler.handleResetPassword` (`:206`) | `user.manage` |
| POST | `/api/v1/iam/users/{userId}/unlock` | `AdminHandler.handleUnlockUser` (`:210`) | `user.manage` |
| PATCH | `/api/v1/iam/users/{userId}` | `AdminHandler.handlePatchUser` (`:214`) | `user.manage` |

Post-ADR-0016 split (GET = view-grade, writes = manage-grade) verified against `apps/api/cmd/metaldocs-api/permissions.go:108-123`.

## API Route Truth Table (Plan 8 Baseline)

| Method | Path | Runtime owner (file:line) | Handler method | Spec path | operationId | Codegen method | Status | Notes |
|---|---|---|---|---|---|---|---|---|
| GET | `/api/v1/iam/area-memberships` | `internal/modules/iam/delivery/http/routes_memberships.go:31` | `listMemberships` | â€” | â€” | â€” | Spec missing | Runtime mounted route has no OpenAPI path yet. |
| POST | `/api/v1/iam/area-memberships` | `internal/modules/iam/delivery/http/routes_memberships.go:32` | `grantMembership` | â€” | â€” | â€” | Spec missing | Runtime mounted route has no OpenAPI path yet. |
| DELETE | `/api/v1/iam/area-memberships` | `internal/modules/iam/delivery/http/routes_memberships.go:33` | `revokeMembership` | â€” | â€” | â€” | Spec missing | Runtime mounted route has no OpenAPI path yet. |
| GET | `/api/v1/iam/users` | `internal/modules/iam/delivery/http/admin_handler.go:84` | `handleUsers` (`GET` branch) | `/iam/users` | â€” | â€” | Aligned | Spec server is `/api/v1`; operationId not defined. |
| POST | `/api/v1/iam/users` | `internal/modules/iam/delivery/http/admin_handler.go:84` | `handleUsers` (`POST` branch) | `/iam/users` | â€” | â€” | Aligned | Spec server is `/api/v1`; operationId not defined. |
| PATCH | `/api/v1/iam/users/{userId}` | `internal/modules/iam/delivery/http/admin_handler.go:85` | `handleUserRoute` -> `handlePatchUser` | `/iam/users/{userId}` | â€” | â€” | Aligned | Routed through path suffix dispatcher; operationId not defined. |
| POST | `/api/v1/iam/users/{userId}/reset-password` | `internal/modules/iam/delivery/http/admin_handler.go:85` | `handleUserRoute` -> `handleResetPassword` | `/iam/users/{userId}/reset-password` | â€” | â€” | Aligned | Routed through path suffix dispatcher; operationId not defined. |
| POST | `/api/v1/iam/users/{userId}/unlock` | `internal/modules/iam/delivery/http/admin_handler.go:85` | `handleUserRoute` -> `handleUnlockUser` | `/iam/users/{userId}/unlock` | â€” | â€” | Aligned | Routed through path suffix dispatcher; operationId not defined. |
| POST | `/api/v1/iam/users/{userId}/roles` | `internal/modules/iam/delivery/http/admin_handler.go:85` | `handleUserRoute` -> `handleUserRoleUpsert` | `/iam/users/{userId}/roles` | â€” | â€” | Aligned | Routed through path suffix dispatcher; operationId not defined. |
| PUT | `/api/v1/iam/users/{userId}/roles` | `internal/modules/iam/delivery/http/admin_handler.go:85` | `handleUserRoute` -> `handleReplaceUserRoles` | `/iam/users/{userId}/roles` | â€” | â€” | Aligned | Routed through path suffix dispatcher; operationId not defined. |
| GET | `/api/v1/iam/admin/overview` | `internal/modules/iam/delivery/http/admin_handler.go:86` | `handleAdminOverview` | `/iam/admin/overview` | â€” | â€” | Aligned | Spec server is `/api/v1`; operationId not defined. |
| GET | `/api/v1/iam/roles` | â€” | â€” | â€” | â€” | â€” | **Proposed (not implemented)** | Spike result from approval route admin PR-3. Catalogue lives in `internal/modules/iam/domain/model.go:10-16`. Proposal shape `{roles:[{code,label}]}` gated by `CapMembershipView` — see [ADR 0018](../decisions/0018-approval-route-lifecycle.md) §"IAM roles source". Implementation deferred to PR-4 or micro-PR. |

- Module contract status: Contracted
- Owner: leandro

Permission resolver: `apps/api/cmd/metaldocs-api/permissions.go:54,196`. None of these ops is wired through oapi-codegen; only `POST .../roles` has request+response schema components in `openapi.yaml`.

---

## 6. Runtime View

### 6.1 listAreaMemberships â€” GET /api/v1/iam/area-memberships

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant MW as Middleware.Wrap
    participant CS as CapabilityService
    participant H as MembershipHandler
    participant S as AreaMembershipService
    participant R as UserAreaRepository
    participant DB as Postgres
    C->>MW: GET /api/v1/iam/area-memberships
    MW->>CS: CanDo(userID, tenantID, "membership.manage")
    CS->>DB: SELECT EXISTS(...) over iam_user_roles + iam_group_*
    DB-->>CS: allowed=true|false
    alt denied
        CS-->>MW: ErrCapabilityDenied
        MW-->>C: 403 {error:{code,message,trace_id}}
    else allowed
        MW->>H: listMemberships
        H->>S: ListActive(userID, tenantID)
        S->>R: ListActive(ctx, userID, tenantID, now)
        R->>DB: SELECT user_process_areas WHERE effective_to IS NULL
        DB-->>R: rows
        R-->>S: []UserProcessArea
        S-->>H: items
        H-->>C: 200 {"items":[...]}
    end
```

Read-only. Tripwire pairing: N/A. State transitions: none.

### 6.2 grantAreaMembership â€” POST /api/v1/iam/area-memberships

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant MW as Middleware.Wrap
    participant H as MembershipHandler
    participant S as AreaMembershipService
    participant R as UserAreaRepository
    participant DB as Postgres
    C->>MW: POST body {userId,areaCode,role,grantedBy}
    MW->>MW: tier-1 CanDo("membership.manage")
    MW->>H: grantMembership
    H->>S: Grant(userID, tenantID, areaCode, role, grantedBy)
    S->>R: GetActiveByUserAndArea
    R-->>S: existing | nil
    alt prior active
        S->>R: GrantAtomic(old, new) [tx: UPDATE close + INSERT new]
    else first grant
        S->>R: Insert(new)
    end
    R-->>S: nil
    S->>S: governance logger (nil in prod) â€” skipped
    S-->>H: nil
    H-->>C: 201 {userId,tenantId,areaCode,role}
```

State transitions:

| Entity | From | To | Trigger | Capability (tier-1) |
|---|---|---|---|---|
| `public.user_process_areas` row | active (`effective_to IS NULL`) | closed (`effective_to = newMembership.effective_from`) + new active row | `POST /api/v1/iam/area-memberships` via `GrantAtomic` | `membership.manage` |
| `public.user_process_areas` row | absent | new active row | `POST /api/v1/iam/area-memberships` via `Insert` | `membership.manage` |

Tripwire pairing: **active** â€” `trg_require_cap_asserted` attached to `user_process_areas` by migration 0188 (Plan 5). Tier-2 `authz.Require(CapMembershipManage)` is called inside each write method (`Insert` at `user_area_repository.go:59`, `CloseActive` at `:91`, `GrantAtomic` at `:118`). T-004 (IAM write paths lack tier-2) partially closed for the user_area side.

### 6.3 upsertUserRole â€” POST /api/v1/iam/users/{userId}/roles

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant MW as Middleware.Wrap
    participant H as AdminHandler
    participant S as AdminService
    participant R as RoleAdminRepository (pg)
    participant Cache as CachedRoleProvider
    participant DB as Postgres
    C->>MW: POST {userId, role}
    MW->>MW: tier-1 CanDo("user.manage")
    MW->>H: handleUserRoleUpsert
    H->>S: UpsertUserAndAssignRole(userID, displayName, tenantID, role, assignedBy)
    S->>R: same
    R->>DB: tx: INSERT iam_users ON CONFLICT UPDATE; DELETE iam_user_roles; INSERT iam_user_roles; COMMIT
    DB-->>R: ok
    R-->>S: nil
    S->>Cache: InvalidateUser(userID) [post-commit]
    S-->>H: nil
    H-->>C: 200 {role}
```

State transitions:

| Entity | From | To | Trigger | Capability (tier-1) |
|---|---|---|---|---|
| `metaldocs.iam_user_roles` row for `(tenant_id, user_id)` | any role or none | requested role | `POST /api/v1/iam/users/{userId}/roles` | `user.manage` |

Tripwire pairing: **active** â€” `trg_require_cap_asserted` attached to `iam_user_roles` by migration 0188 (Plan 5). Tier-2 `authz.Require(CapUserManage)` is called inside `UpsertUserAndAssignRole` (`role_admin_repository.go:40`) and `ReplaceUserRoles` (`:82`). Audit log emission: `iam.user.role.upserted` event recorded post-write at `admin_handler.go:369` â€” T-005 closed (Plan 6a, 2026-05-11).

### 6.4 Failure modes (current envelope)

| Condition | HTTP | Body |
|---|---|---|
| Tier-1 capability denied | 403 | `{error:{code:"forbidden",message:...,trace_id}}` (`middleware.go:132`) |
| Tier-1 no capability mapped (`guarded=false`) | passes through | (no enforcement) |
| Validation error (admin) | 400 | `{code,message}` (`routes_memberships.go:150` analog in admin) |
| Tier-2 `authz.Require` denied | error returned to caller | typed `authz.ErrCapDenied` (`authz/authz.go:11`) |
| GUC missing in tier-2 | error returned | `authz.ErrActorContextMissing` / `ErrTenantContextMissing` |

RFC 9457 Problem envelope is now the IAM response shape: admin + membership handlers call `problem.Write(w, problem.New(...))` (`internal/platform/problem/problem.go:76-87` sets `Content-Type: application/problem+json` and serializes `Type`/`Title`/`Status`/`Detail`/`Instance`/`Code`/`Errors` per RFC 9457). T-006 closed (Plan 7, 2026-05-12).

---

## 7. Deployment View

- Single Go binary `apps/api/cmd/metaldocs-api` (port `:8081`).
- IAM constructors wired in `apps/api/cmd/metaldocs-api/main.go:163-219`: `NewCapabilityService(SQLDB)`, `NewCachedRoleProvider(provider, authn.CacheTTL())`, `NewMiddleware(...)`, `NewAdminService(...)`, `NewAdminHandler(...)`, `NewMembershipHandler(NewAreaMembershipService(NewUserAreaRepository(SQLDB), nil))`.
- The membership service's `MembershipGovernanceLogger` argument is wired with **`nil`** (`main.go:217`) â€” governance log emission is dead in this binary. Recorded as T-007.
- Migrations applied externally (forward-only). `application/startup.go` was deleted in Plan 4 (its only role was `CheckRoleCapabilitiesVersion`, closed T-012). No boot-time drift check remains.
- No env-var flags for IAM (e.g. `IAM_AUTHZ_ENFORCED` does NOT exist; the `enabled` boolean reaches `NewMiddleware` via `authn.Enabled()` â€” artifact 03 Â§3).

---

## 8. Cross-cutting Concepts

### 8.1 Authentication & Authorization

- Authentication: `auth` module owns it; IAM consumes `auth.domain.ManagedUser` only from `AdminHandler`.
- Tier-1 (edge): `CapabilityService.CanDo`. Resolver: `apps/api/cmd/metaldocs-api/permissions.go`.
- Tier-2 (in-tx): `authz.Require` â€” consumed BY `documents/**` (5 import sites), `controlled-documents` (changeStatus), `taxonomy` (FamilyRepository.Create/Update, ProfileRepository.Create/Update, AreaRepository.Create/Update), `templates` (CreateTemplate, SubmitForReview, Review, Approve, PublishTemplateVersion, ArchiveTemplate), AND NOW also by IAM itself (`role_admin_repository.go`, `user_area_repository.go`). See `_artifacts/03-deps.md` for full import table.
- Tier-3 (DB enforcer): `enforce_capability_asserted` trigger on 12 tables (expanded by Plan 5, migration 0188). Reads `metaldocs.asserted_caps` GUC.
- system_admin: bypasses tier-1 (`capability_service.go:33-45`) and tier-2 (`authz/authz.go:58`).

### 8.2 Tenant scoping

Every IAM-owned table has `tenant_id` (`iam_users` since 0130, `iam_user_roles` since 0162, `iam_groups*` since 0163, `user_process_areas` since 0125). Unique constraints are tenant-scoped (e.g. `ux_iam_users_tenant_user`, partial-unique active-membership index `ux_user_process_areas_single_active`). `DevTenantID` sentinel is the only legal value in single-tenant dev.

### 8.3 Caching

`CachedRoleProvider` (`application/cached_role_provider.go:18`) wraps the postgres provider with TTL cache keyed by `userID|tenantID`. `InvalidateUser(userID)` runs after every `AdminService` write (`admin_service.go:42`). No `InvalidateGroup` exists â€” group-membership writes do not invalidate (T-008).

### 8.4 Cross-deps (consumers)

- `internal/modules/documents/application/fillin_authz.go:9` â€” tier-2 + `iamdomain.Capability`
- `internal/modules/documents/approval/application/cancel_service.go:12` â€” tier-2 + `BypassSystem`
- `internal/modules/documents/delivery/http/handler.go:17` â€” `iamapp.ErrCapabilityDenied` (sentinel from `application/capability_service.go`); `authz.ErrCapDenied` (struct from `authz/authz.go`, carries capability/area â€” T-009 closed by Plan 4)
- `internal/modules/templates/delivery/http/routes_lifecycle.go:8` â€” `iamdomain.RolesFromContext`
- `internal/modules/auth/{application,delivery,domain,infrastructure}` â€” 4 sites import `iamdomain.Role` (circular concern; auth shouldn't depend on iam's role enum if iam can depend on auth â€” see T-010)
- `internal/platform/{bootstrap,authn,security,testsupport}` â€” 4 sites use IAM context helpers

### 8.5 Concurrency / Transactions

- Tier-1 path: no tx (single `db.QueryRowContext`).
- Tier-2 path: requires `*sql.Tx` argument (`authz.Require(ctx, tx, cap, area)`).
- IAM mutations:
  - `RoleAdminRepository.UpsertUserAndAssignRole` opens its own tx (`role_admin_repository.go:35` `db.BeginTx`); calls `authz.Require(CapUserManage)` at `:40`.
  - `RoleAdminRepository.ReplaceUserRoles` opens its own tx (`:77`); calls `authz.Require(CapUserManage)` at `:82`.
  - `UserAreaRepository.Insert` opens its own tx (`user_area_repository.go:53`); calls `authz.Require(CapMembershipManage)` at `:59`.
  - `UserAreaRepository.CloseActive` opens its own tx (`:85`); calls `authz.Require(CapMembershipManage)` at `:91`.
  - `UserAreaRepository.GrantAtomic` opens its own tx (`:110`); calls `authz.Require(CapMembershipManage)` at `:118`.

### 8.6 Error namespace

`iamapp.ErrCapabilityDenied` (`application/capability_service.go:10`) â€” sentinel `var`, suitable for `errors.Is`. `authz.ErrCapDenied` (`authz/authz.go:11`) â€” typed struct carrying capability + area context. Names no longer collide (Plan 4, T-009 closed).

---

## 9. Architecture Decisions

| Decision | Link / Status |
|---|---|
| Two-tier authz (CapabilityService + authz.Require) | [`wiki/decisions/0007-two-tier-authz.md`](../decisions/0007-two-tier-authz.md) |
| DB tripwire (`enforce_capability_asserted`) as enforcement floor | ADR 0007 â€” codegen-rejected amendment (2026-05-10) |
| `document.create` via `CapabilityChecker` adapter | ADR 0007 â€” J2 amendment (2026-05-05) |
| `tenant_id` per IAM table (Group B) | ADR 0007 references migration 0162; tenant-isolation rule is implicit. No standalone ADR. **missing-ADR** â†’ T-011 |
| Collapse dual capability namespaces â€” typed `iamdomain.Capability` wins; `capabilities.go` deleted; DB reseeded to `document.*` | Plan 4 (2026-05-11). No standalone ADR â€” **ADR-TODO** per Plan 13. |
| Delete `AuthorizationService` (third authz surface); Plan 5 wired tier-2 per module (IAM repos, controlled-documents, taxonomy, templates, documents) | Plan 4 (2026-05-11). No standalone ADR â€” **ADR-TODO** per Plan 13. |
| Delete `area_membership/` Go wrapper; canonical write surface is `UserAreaRepository.GrantAtomic`; SECURITY DEFINER SQL funcs stay for e2e seed | Plan 4 (2026-05-11). No standalone ADR â€” **ADR-TODO** per Plan 13. |
| Role list endpoint (`GET /api/v1/iam/roles`) shape + cap gate proposal (deferred) | [`wiki/decisions/0018-approval-route-lifecycle.md`](../decisions/0018-approval-route-lifecycle.md) §"IAM roles source" |

---

## 10. Quality Requirements

| Goal | Scenario | Pass criteria |
|---|---|---|
| Correctness â€” tier-1 denial | Authn'd user without `user.manage` POSTs `/api/v1/iam/users/{id}/roles` | 403 `forbidden`; no row in `iam_user_roles` |
| Correctness â€” tier-2 GUC missing | Caller forgets `SET LOCAL metaldocs.actor_id` before `authz.Require` | `ErrActorContextMissing` returned; tx not advanced |
| Tenant isolation | User has `author` in tenant A; `RolesByUserID(user, B)` | empty slice (verified by `tests/integration/iam/tenant_isolation_test.go`) |
| system_admin bypass | system_admin user calls any guarded route | tier-1 short-circuit at `capability_service.go:33-45`; tier-2 short-circuit at `authz/authz.go:58` |

---

## 11. Risks & Technical Debt

Pointer-only. Body in [`wiki/modules/iam-tech-debt.md`](iam-tech-debt.md). Severity rubric (concrete triggers) lives in the register, not here.

Summary counts (open + partial only; closed items retained in register with CLOSED markers):
- Critical: 0
- Major: 2 (T-004 partial, T-007)
- Minor: 3 (T-008, T-010, T-011)
- Decisions without ADR link: 11

Top 3 (by severity, then by blast-radius):
1. T-004 â€” `iam_users` INSERT inside `UpsertUserAndAssignRole`/`ReplaceUserRoles` still tier-1 only; no dedicated tripwire on `metaldocs.iam_users`. Residual after Plan 5 closed the `iam_user_roles`/`user_process_areas` halves. Major (partially closed).
2. T-007 â€” `MembershipGovernanceLogger` wired with `nil` in `main.go:217`; grant/revoke produce no governance log on the application-service path while the SECURITY DEFINER path still does. Major.
3. T-010 â€” `auth` module imports `iamdomain.Role` while IAM admin handler imports `auth/domain.ManagedUser`; non-circular today but blocks future structural moves of `admin_handler`. Minor.

Refactor backlog: [`wiki/backlog/iam-refactor.md`](../backlog/iam-refactor.md).

---

## 12. Glossary

| Term | Definition |
|---|---|
| Capability | Fine-grained permission string (e.g. `document.view`) consumed by tier-1 CanDo. Typed as `iamdomain.Capability`; 18 consts in `model.go` (namespace `document.*` / `template.*`). |
| Role | Named bundle of capabilities; 5 canonical (`viewer`, `editor`, `author`, `approver`, `system_admin`) + 3 area-only (`signer`, `area_admin`, `qms_admin`). |
| Tier-1 | Edge / HTTP middleware authz check using `CapabilityService.CanDo`. |
| Tier-2 | In-tx area-scoped authz check using `authz.Require(ctx, tx, cap, areaCode)`. |
| Tripwire | DB-side `enforce_capability_asserted()` trigger that rejects mutating rows on guarded tables when `metaldocs.asserted_caps` GUC does not contain the required capability. Migration 0188 (Plan 5) extended coverage from 2 tables to 12: `approval_instances`, `approval_signoffs`, `iam_user_roles`, `user_process_areas`, `documents`, `controlled_documents`, `cd_sequence_counters`, `document_profiles`, `document_process_areas`, `document_families`, `templates_template`, `templates_template_version`. |
| GUC | Grand Unified Configuration â€” Postgres session/local setting. IAM reads `metaldocs.actor_id`, `metaldocs.tenant_id`, `metaldocs.asserted_caps`. |
| Area membership | Row in `public.user_process_areas` granting a user a role within a process area for an `effective_from`â†’`effective_to` window. |
| SECURITY DEFINER | Postgres function attribute that executes with the function owner's privileges, reading `metaldocs.actor_id` GUC. Used by `metaldocs.grant_area_membership` / `revoke_area_membership` (migration 0137) â€” called only by e2e seed + integration tests; the Go `area_membership/` wrapper was deleted in Plan 4. |

---

## Failure modes

| Failure | Symptom | Detection | Response |
|---|---|---|---|
| Postgres unavailable for `CapabilityService.CanDo` | All non-public routes 500 (tier-1 fails closed) | Middleware logs; `/healthz` | Restore Postgres; cache miss cannot fall through — fail-closed by design |
| Tier-2 `authz.Require` ctx missing actor / tenant GUC | `ErrActorContextMissing` / `ErrTenantContextMissing` thrown | Caller did not set `metaldocs.actor_id` / `metaldocs.tenant_id` before in-tx capability check | Wrap call site in `SeedTxIdentity`; never call `Require` without GUC seeded |
| Tier-3 Postgres tripwire abort on iam write | `UpsertUserAndAssignRole` / `ReplaceUserRoles` INSERT rejected | `RAISE` from migration 0188 trigger; caller sees 500 mapped to RFC 9457 | Bypassed `authz.Require(CapUserManage)` — fix-forward; never disable tripwire |
| `CachedRoleProvider` stale after admin role change | User keeps old roles until TTL expires | Admin handler calls `InvalidateUser(userID)` after every `assignRole`/`replaceRoles` (`application/admin_service.go:42`) | If invalidation forgotten on new write path, add it; TTL bounds blast radius |
| Spoofed `X-User-Roles` / `X-User-ID` headers | Caller attempts role escalation via headers | Closed 2026-05-25 — `Middleware.Wrap` now strips both before downstream (`Trusted role-header strip sync`) | Regression test on middleware; never restore the legacy fallback |
| Membership double-grant for `(user, area)` | UNIQUE violation on `user_process_areas` insert | `Insert` / `GrantAtomic` returns pq 23505 | Caller refetches `ListActive`; surfaces as 409 `iam.membership_exists` |
| Membership revoke race (active row already closed) | `CloseActive` returns 0 rows | `ErrMembershipNotFound` from service | Caller refetches and surfaces NO-OP to UI |
| Capability matrix drift from migrations 0165 / 0169 | Operator expects capability that no longer maps to role | Tier-1 returns 403 unexpectedly | Audit matrix vs migration order; never hand-edit `role_capabilities` rows |
| `system_admin` bypass triggered unintentionally | `authz.BypassSystem` skips tier-2 checks for system_admin | Audit log shows admin acting on tier-2 path | Expected for system_admin; restrict who holds the role |
| Tenant context default fallback (`tenant.DevTenantID`) leaks into prod | Multi-tenant data crosses boundary | `bootstrap` config check at startup | Production must reject `DevTenantID`; flag deploys that set `AllowDevTenantFallback=true` |

## Cross-links

- ADRs: [`decisions/0007-two-tier-authz.md`](../decisions/0007-two-tier-authz.md), [`decisions/0012-contract-first-api.md`](../decisions/0012-contract-first-api.md)
- Concepts: [`concepts/authz-tiers.md`](../concepts/authz-tiers.md), [`concepts/iso-segregation.md`](../concepts/iso-segregation.md)
- Architecture: [`architecture/api-design-system.md`](../architecture/api-design-system.md), [`architecture/api-contract.md`](../architecture/api-contract.md), [`architecture/tenant-context.md`](../architecture/tenant-context.md)
- Modules: [`modules/approval.md`](approval.md) (tier-2 consumer), [`modules/documents.md`](documents.md) (tier-2 consumer), [`modules/templates.md`](templates.md) (predecessor frontend doc), [`modules/templates.md`](templates.md) (Arc42 backend doc â€” consumer of `template.*` capability namespace; T-001 closed Plan 4), [`modules/auth.md`](auth.md) (bidirectional dep: auth imports `iamdomain`; `iam.AdminHandler` imports `authdomain.ManagedUser/OnlineUser/UpdateUserParams`)
- See also: [`modules/audit.md`](audit.md) â€” iam `AdminHandler.recordAudit` calls `audit.Writer.Record` for role/user admin ops including `handleUserRoleUpsert` (`admin_handler.go:369`, T-005 closed Plan 6a)
- See also: [`modules/controlled-documents.md Â§8.1`](controlled-documents.md#81-authentication--authorization) â€” controlled-documents now has tier-2 `authz.Require` for Create + changeStatus and tier-3 DB tripwire on `controlled_documents` + `cd_sequence_counters`
- Backlog: [`backlog/iam-refactor.md`](../backlog/iam-refactor.md)
- Tech debt: [`iam-tech-debt.md`](iam-tech-debt.md)
- Source artifacts: [`iam/_artifacts/00-context.md`](iam/_artifacts/00-context.md) through [`05-industry.md`](iam/_artifacts/05-industry.md)

## Changelog

- 2026-05-25 - Phase 11 IAM medium sweep: shared `ParseRole` validation now backs admin-role parsing and area-membership grants, admin audit events reuse a single captured `time.Now()` and always stamp server-generated UUID trace IDs, role caches and migration follow-up TODOs were refreshed, and the postgres role provider now returns an empty role slice instead of `ErrNoRolesAssigned`.
- 2026-05-25 - Trusted role-header strip sync: `Middleware.Wrap` now deletes both `X-User-ID` and `X-User-Roles` before downstream handling, keeping role data sourced from IAM context/role providers instead of caller-controlled headers.
- 2026-05-11 â€” Plan 5 (tier-2 authz.Require + Postgres tripwire expansion): `role_admin_repository.go` `UpsertUserAndAssignRole`/`ReplaceUserRoles` now call `authz.Require(CapUserManage)`; `user_area_repository.go` `Insert`/`CloseActive`/`GrantAtomic` call `authz.Require(CapMembershipManage)`. Import aliased to `iamdomain`. `domain/model.go` gained `CapControlledDocumentObsolete`/`CapControlledDocumentSupersede`. Migration 0188 attaches `trg_require_cap_asserted` to 10 new tables. Â§2 DB floor, Â§6.2/Â§6.3 tripwire-pairing, Â§8.1 tiers, Â§8.5 concurrency, Â§9, Â§11, Â§12 all updated.
- 2026-05-11 â€” Plan 4 (capability namespace collapse + IAM dual-surface consolidation): collapsed `capabilities.go` + `model.go` into single typed `Capability` namespace (18 consts); deleted `area_membership/`, `application/authorization.go`, `domain/role_capabilities.go`, `application/startup.go`; renamed `authz.ErrCapabilityDenied` â†’ `authz.ErrCapDenied`; Â§5.1 C4 updated; Â§5.2 file table pruned; Â§5.4 removed; Â§8.5/Â§8.6 updated; Â§9/Â§10/Â§11/Â§12 refreshed. Closed T-001/T-002/T-003/T-009/T-012.
- 2026-05-11 â€” Plan 3 (module sweep): IAM middleware now calls `tenant.FromContext` as primary tenant source; legacy fallback to `X-Tenant-ID` header preserved under `legacyHeader` mode; `handleAdminOverview` + `tenantIDFromRequest` updated; Key files, Â§2, cross-links updated.
- 2026-05-10 â€” initial publish; supersedes retired `iam-rbac.md` stub. Author: Claude (Opus 4.7) under metaldocs-module-doc skill.
