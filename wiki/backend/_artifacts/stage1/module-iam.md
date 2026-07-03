# Stage-1 Audit Artifact — module-iam

> **Generated:** 2026-06-10 | **Branch:** qa/iam-area-membership | **Repo root:** MetalDocs
>
> Read-only evidence pass. Every claim is anchored to a file:line or tagged
> `[runtime-unverified]` where only a live run can confirm.

---

## 1. Identity & purpose

The `iam` module is the authorization backbone of MetalDocs. It answers "can user X perform capability C in tenant T (and area A when area-scoped)?" via two runtime enforcement tiers (HTTP edge + in-transaction) and a Postgres trigger as a DB-layer floor. It owns: the typed capability registry (29 consts), the canonical 8-role catalog, the tenant-scoped role assignment tables (`iam_user_roles`, `iam_groups*`), the area-scoped membership table (`user_process_areas`), the HTTP admin surface for user and role management, the area-membership grant/revoke surface, and the real-time presence (WebSocket + snapshot) sub-system.

IAM does **not** own authentication (login, sessions, password hashing) — those live in `internal/modules/auth/`. IAM consumes `auth/domain` types (`ManagedUser`, `OnlineUser`) for the admin overview, and `auth/application.Service` for the People-tab orchestration. The Postgres tripwire trigger (`enforce_capability_asserted`) conceptually belongs to IAM but is attached to 12 tables across all regulated modules via migration 0188.

Every IAM-owned table is multi-tenant: each row carries `tenant_id` and every repository method filters by it. The `system_admin` role bypasses both tier-1 and tier-2 enforcement. Area-grade capabilities require the actor to hold an active `user_process_areas` row for the target area whose role grants that capability; `system_admin` bypasses this via the existing tier-2 SQL short-circuit.

---

## 2. File inventory

### `internal/modules/iam/domain/` — core types, catalog, and errors

| File | Role |
|---|---|
| `model.go` | `Role` type + 8 consts; `Capability` type + 29 typed consts; `validCapabilities` guard set; `validRoles`/`areaRoles` sets; `ParseRole`, `IsValidRole`, `IsAreaRole`, `AreaRoles`, `AllCapabilities`, `MustCapability` |
| `catalog.go` | `RoleDescriptor`, `CapabilityDescriptor`, `RoleCapabilityLink` types; `canonicalRoles` (8 descriptors); `capabilityDescriptions` map (pt-BR); `CanonicalRoles()`, `CapabilityCatalog()`, `capabilityCategory()` |
| `capability_scope.go` | `CapabilityScope` type + `ScopeTenant`/`ScopeArea` consts; `capabilityScopes` map classifying all 29 caps; `ScopeOf`, `IsAreaGrade` |
| `port.go` | `RoleProvider` interface; `RoleAdminRepository` interface |
| `user_area.go` | `UserProcessArea` struct; `IsActive(now)` method |
| `observability.go` | `PlanTier`, `SeatUsage`, `StorageUsage`, `CountWindows`, `UsageSnapshot`, `RoleCount`, `KpiSnapshot`, `TenantPlan` domain types |
| `observability_port.go` | `ObservabilityRepository` interface (8 methods); `ErrTenantPlanNotFound` |
| `context.go` | `WithAuthContext`, `UserIDFromContext`, `RolesFromContext` — context key helpers |
| `errors.go` | `ErrUserNotFound`, `ErrUserInactive`, `ErrNoRolesAssigned`, `ErrInvalidRole` |
| `catalog_test.go` | Unit tests for `CapabilityCatalog`, registry size lock, role catalog |
| `capability_scope_test.go` | `TestEveryCapabilityClassified`, `TestAreaGradeCapabilitySet`, `TestScopeOf` |
| `model_test.go` | `ParseRole`, `IsAreaRole` unit tests |
| `.gitkeep` | Empty marker |

### `internal/modules/iam/authz/` — tier-2 enforcement package

| File | Role |
|---|---|
| `authz.go` | `SystemAdminExistsSQL` shared const; `ErrCapDenied` struct; `capCache`/`assertedCache` structs; `WithCapCache`, `Require`, `WithBackgroundBypass`, `BypassSystem`, `setBypassGUC`; in-tx GUC append/load for `metaldocs.asserted_caps` |
| `context.go` | `ErrActorContextMissing`, `ErrTenantContextMissing`, `MustActorID`, `MustTenantID`, `SeedTxIdentity` |
| `bypass_audit.go` | `BypassKind`, `BypassEvent`, `BypassAuditSink` interface; `bypassAuditSink` atomic pointer; `SetBypassAuditSink`, `recordBypass`, `softActorID`, `softTenantID` |
| `authz_test.go` | Unit tests for `Require`, cache, GUC seeding |
| `authz_bypass_test.go` | Tests for `BypassSystem` fail-closed and audit paths |
| `context_test.go` | Tests for `MustActorID`, `MustTenantID`, `SeedTxIdentity` |

### `internal/modules/iam/application/` — services

| File | Role |
|---|---|
| `capability_service.go` | `ErrCapabilityDenied` sentinel; `CapabilityService` + `NewCapabilityService`; `CanDo` (4-branch UNION EXISTS); `IsSystemAdmin`; `CapsByUserID` |
| `admin_service.go` | `RoleCacheInvalidator` interface; `AdminService` + `NewAdminService`; `UpsertUserAndAssignRole`, `ReplaceUserRoles` (validate + delegate to repo + invalidate cache) |
| `area_membership_service.go` | `ErrMembershipNotFound`, `ErrUnknownRole`, `ErrMembershipExists` sentinels; `UserAreaWriteRepository` interface (8 methods); `MembershipGovernanceLogger` interface; `AreaMembershipService` + `NewAreaMembershipService`; `WithRoleCacheInvalidator`; `ListActive`, `ListByTenant`, `DirectoryScope`, `ListByTenantInManagedAreas`, `Grant`, `Revoke`; `buildMembership` helper |
| `cached_role_provider.go` | `CachedRoleProvider` + `NewCachedRoleProvider`; background TTL sweep goroutine; `RolesByUserID`, `InvalidateUserTenant`; `cloneRoles` |
| `observability_service.go` | `MfaCoveragePctReader` interface; `ObservabilityService` + `NewObservabilityService`; `GetUsage`, `GetKpi`; exported window constants (`DormantThresholdDays`, `FailedLoginsWindowSec`, etc.) |
| `people_service.go` | `ErrPeopleValidation`, `ErrAreaUnknown`, `ErrUserNotInTenant`, `ErrCursorExpired` sentinels; `ListedUser`, `ListFilters`, `ListResult`, `BulkOutcome`, `InviteInput`, `PatchInput`; `PeopleService` + `NewPeopleService`; `Invite`, `PatchAtomic`, `BulkAction`, `ListFiltered`, `VerifyUserInTenant`, `ListMemberships`; `generateTempPassword` (crypto/rand 16-char); `applyPeopleFilters` (in-memory) |
| `dev_role_provider.go` | `DevRoleProvider` + `NewDevRoleProvider` — in-memory provider for dev/memory mode |
| `area_membership_test.go` | Unit tests for `AreaMembershipService` |
| `capability_service_test.go` | Unit tests for `CapabilityService.CanDo` |
| `observability_service_test.go` | Unit tests for `ObservabilityService` |
| `people_service_test.go` | Unit tests for `PeopleService` |
| `.gitkeep` | Empty marker |

### `internal/modules/iam/delivery/http/` — HTTP handlers

| File | Role |
|---|---|
| `middleware.go` | `Visibility` enum; `PermissionResolver` func type; `Middleware` + `NewMiddleware`; `Wrap` — strips `X-User-ID`/`X-User-Roles`, resolves capability via resolver, enforces tier-1 `CanDo`, enriches context with roles |
| `admin_handler.go` | `UserAdminService`, `AuditEventLister`, `KpiReader`, `PresenceReader` interfaces; `AdminHandler` + `NewAdminHandler`; `RegisterRoutes`; `handleAdminOverview` (concurrent errgroup: KPI + presence + audit); `handleUserRoleUpsert`, `handleReplaceUserRoles`; `recordAudit`, `writeProblem`; `UpsertUserRoleRequest`, `ReplaceUserRolesRequest` |
| `people_handler.go` | `PeopleHandler` + `NewPeopleHandler`; `RegisterRoutes` (7 typed mux patterns); `handleListUsers`, `handleInvite`, `handlePatch`, `handleResetPassword`, `handleUnlock`, `handleBulk`, `handleListMemberships`; `guardUserInTenant`; `toManagedUserCore` |
| `routes_memberships.go` | `MembershipUserTenantVerifier` interface; `MembershipHandler` + `NewMembershipHandler`; `RegisterRoutes` (GET/POST/DELETE); `listMemberships` (directory scope via `DirectoryScope`), `grantMembership`, `revokeMembership`; `toMembershipDTO`; `guardMembershipUserInTenant`; `isSelf`; `recordMembershipAudit`; `writeMembershipError` |
| `routes_roles_caps.go` | `RoleCapabilitiesReader` interface; `RolesCapsHandler` + `NewRolesCapsHandler`; `RegisterRoutes` (`GET /roles`, `GET /capabilities`, `GET /role-capabilities`); `buildRoleCatalog`, `buildCapabilityCatalog` (process-start cache) |
| `observability_handler.go` | `ObservabilityHandler` + `NewObservabilityHandler`; `RegisterRoutes` (`GET /usage`, `GET /kpi`); `usageToJSON`, `kpiToJSON` |
| `sessions_handler.go` | `SessionAdmin` interface; `SessionsHandler` + `NewSessionsHandler`; `RegisterRoutes` (`/auth/sessions`, `/auth/sessions/`); list + find + revoke single + revoke all for user |
| `middleware_test.go` | Unit tests for `Middleware.Wrap` |
| `middleware_problem_test.go` | Tests for RFC 9457 problem shape from middleware |
| `admin_handler_test.go` | Tests for admin handler routes |
| `routes_memberships_contract_test.go` | Contract tests for membership routes |
| `routes_roles_caps_test.go` | Tests for roles/caps handler |
| `sessions_handler_test.go` | Tests for sessions handler |
| `.gitkeep` | Empty marker |

### `internal/modules/iam/infrastructure/` — data layer

| Sub-package | File | Role |
|---|---|---|
| `postgres` | `role_provider.go` | `RoleProvider` + `RolesByUserID` (2-query: check user active, then fetch roles) |
| `postgres` | `role_admin_repository.go` | `RoleAdminRepository`; `HasAnyRole`; `UpsertUserAndAssignRole` (BEGIN → SeedTxIdentity → Require(CapUserManage) → DELETE + INSERT → COMMIT); `ReplaceUserRoles`, `ReplaceUserRolesTx` |
| `postgres` | `user_area_repository.go` | `UserAreaRepository`; `ListActive`, `ListByTenant`, `MembershipDirectoryScope` (shared `SystemAdminExistsSQL`), `ListByTenantInManagedAreas` (SQL subquery JOIN for R3), `Insert`, `CloseActive`, `GrantAtomic`; `GetActiveByUserAndArea`; `scanUserProcessArea`, `rollbackTx`, `grantedByActor` |
| `postgres` | `role_capabilities_repository.go` | `RoleCapabilitiesRepository`; `ListRoleCapabilities` — reads global `role_capabilities` matrix |
| `postgres` | `observability_repository.go` | `ObservabilityRepository`; 8 read-only tenant-scoped queries (seats, storage stub −1, audit event counts, active users, plan, locked accounts, failed logins, dormant users, role distribution) |
| `postgres` | `area_catalog_reader.go` | `ProcessAreaCatalog`; `AreaCodeExists` — validates area code via `document_process_areas` |
| `postgres` | `role_provider_test.go`, `role_admin_repository_test.go`, `area_catalog_reader_test.go` | Unit/integration tests |
| `memory` | `role_admin_repository.go` | `RoleAdminRepository` (in-memory) — dev/memory mode; `HasAnyRole`, `UpsertUserAndAssignRole`, `ReplaceUserRoles` |
| `memory` | `role_admin_repository_test.go` | Tests for in-memory repository |

### `internal/modules/iam/presence/` — real-time presence sub-system

| File | Role |
|---|---|
| `model.go` | Constants (`IdleAfter`, `OfflineAfter`, `BumpDebounce`, `HeartbeatInterval`, `ClientIdleTimeout`, `TickInterval`); `Status` type; `Item`, `Event` structs; `ClassifyStatus` |
| `repository.go` | `Repository` interface; `PostgresRepository` + `NewPostgresRepository`; `BumpLastSeen`, `Snapshot`, `TenantsWithRecentPresence` |
| `hub.go` | `Hub` + `NewHub`; `Run` (ticker goroutine), `RunHeartbeat`; `Subscribe`; `tickAll`/`tickRoom`; `diff` (join/leave/status-change events); `Conn`, `room` types; `broadcast` (drops slow consumers) |
| `handler.go` | `Handler` + `NewHandler`; `RegisterRoutes` (`GET /presence/snapshot`, `GET /presence/stream`); `handleSnapshot`; `handleStream` (WebSocket upgrade, initial snapshot, subscribe, read/write pump with idle timer) |
| `middleware.go` | `BumpMiddleware` + `NewBumpMiddleware`; `Wrap` (debounced last_seen_at bump per authenticated request); `StartCleanup` (eviction goroutine) |
| `presence_test.go` | Unit tests for presence hub, diff, and middleware |

### `internal/modules/iam/api/` — generated contract types

| File | Role |
|---|---|
| `api.gen.go` | oapi-codegen generated types for IAM's portion of the OpenAPI spec: `ManagedUserCore`, `AreaMembership`, `UserInviteRequest/Response`, `ListUsersResponse`, `ListMembershipsResponse`, `UserRole` enum, `RoleDescriptor`, `CapabilityDescriptor`, `ListRolesResponse`, `ListCapabilitiesResponse`, `RoleCapabilityLink`, `ListRoleCapabilitiesResponse`, and all other IAM contract types |
| `cfg.yaml` | oapi-codegen configuration for the `iamapi` package |
| `gen.go` | `//go:generate` directive |

### `internal/modules/iam/`

| File | Role |
|---|---|
| `integration_test.go` | Module-level integration test stub |

---

## 3. Public surface

### Exported types and functions consumed by other modules/packages

**`domain` package (imported by auth, documents, templates, taxonomy, platform)**

| Symbol | Consumed by |
|---|---|
| `Role`, `RoleApprover … RoleViewer` (8 consts) | `auth/domain`, `auth/application`, `auth/delivery/http/middleware`, `auth/infrastructure/memory`; `templates/delivery/http` |
| `Capability`, `Cap*` (29 typed consts) | `documents/application/fillin_authz.go`, `documents/application/ports.go`, `documents/delivery/http/handler.go`, `controlleddocuments/application/service.go`, `taxonomy/infrastructure`, `templates/application`, `apps/api/cmd/metaldocs-api/permissions.go` |
| `IsAreaGrade`, `ScopeOf` | `scripts/api-lint/registry_rules.go` |
| `WithAuthContext`, `UserIDFromContext`, `RolesFromContext` | `auth/delivery/http/middleware.go`, `platform/authn/context.go`, `templates/delivery/http` |
| `UserProcessArea`, `IsActive` | `documents/application/fillin_authz.go`, `controlleddocuments/application/service.go` |
| `ErrUserNotFound`, `ErrInvalidRole` | `documents/application`, `templates/application` |
| `IsValidCapability`, `IsValidRole`, `ParseRole`, `AreaRoles` | `apps/api/cmd/metaldocs-api/permissions.go` |
| `CanonicalRoles`, `CapabilityCatalog` | `delivery/http/routes_roles_caps.go` |

**`authz` package (imported by all modules that have tier-2 checks)**

| Symbol | Consumed by |
|---|---|
| `Require(ctx, tx, capability, areaCode)` | `iam/infrastructure/postgres/role_admin_repository.go`, `iam/infrastructure/postgres/user_area_repository.go`, `documents/repository`, `controlleddocuments/infrastructure/repository.go`, `taxonomy/infrastructure`, `templates/application/lifecycle.go` |
| `ErrCapDenied` (struct, carries capability/area/actor) | `documents/delivery/http/handler.go`, `iam/delivery/http/routes_memberships.go` |
| `ErrActorContextMissing`, `ErrTenantContextMissing` | All tier-2 callers |
| `SeedTxIdentity` | All tier-2 callers that own their own tx |
| `WithCapCache`, `WithBackgroundBypass`, `BypassSystem` | `documents/approval/application/scheduler_service.go`, `documents/jobs/session_sweeper.go`, `documents/jobs/orphan_pending_sweeper.go`, `jobs/stuck_instance_watchdog/job.go` |
| `SetBypassAuditSink` | `apps/api/cmd/metaldocs-api/main.go` |
| `SystemAdminExistsSQL` | `infrastructure/postgres/user_area_repository.go` (DRY reuse) |

**`application` package**

| Symbol | Consumed by |
|---|---|
| `CapabilityService`, `NewCapabilityService`, `CanDo`, `ErrCapabilityDenied` | `delivery/http/middleware.go`, `apps/api/cmd/metaldocs-api/main.go`, `apps/api/internal/wiring/documents.go` |
| `AdminService`, `NewAdminService`, `UpsertUserAndAssignRole`, `ReplaceUserRoles` | `delivery/http/admin_handler.go`, `main.go` |
| `AreaMembershipService`, `NewAreaMembershipService`, `Grant`, `Revoke`, `ListActive`, `ListByTenant` | `delivery/http/routes_memberships.go`, `delivery/http/people_handler.go`, `main.go` |
| `PeopleService`, `NewPeopleService`, `Invite`, `PatchAtomic`, `BulkAction`, `ListFiltered`, `VerifyUserInTenant` | `delivery/http/people_handler.go`, `delivery/http/routes_memberships.go` (as `MembershipUserTenantVerifier`) |
| `CachedRoleProvider`, `NewCachedRoleProvider`, `InvalidateUserTenant` | `main.go`, wiring |
| `ObservabilityService`, `NewObservabilityService`, `GetUsage`, `GetKpi` | `delivery/http/observability_handler.go`, `admin_handler.go` (via `KpiReader`) |
| `ErrMembershipExists`, `ErrMembershipNotFound`, `ErrUserNotInTenant`, `ErrCursorExpired` | `delivery/http` handlers |

### HTTP routes (PEP bindings)

| Method | Path | Handler | Tier-1 capability | Notes |
|---|---|---|---|---|
| GET | `/api/v1/iam/area-memberships` | `MembershipHandler.listMemberships` (`routes_memberships.go:96`) | `membership.view` | Directory scope resolved via `DirectoryScope` at data layer (system_admin → tenant-wide; area_admin → managed areas SQL subquery; others → self-only) |
| POST | `/api/v1/iam/area-memberships` | `MembershipHandler.grantMembership` (`routes_memberships.go:173`) | `membership.manage` | Self-grant 403; tier-2 area-scoped in `Insert`/`GrantAtomic`; 409 on duplicate same-role |
| DELETE | `/api/v1/iam/area-memberships/{user_id}/{area_code}` | `MembershipHandler.revokeMembership` (`routes_memberships.go:249`) | `membership.manage` | Path params; cross-tenant 404 guard; tier-2 area-scoped in `CloseActive` |
| GET | `/api/v1/iam/admin/overview` | `AdminHandler.handleAdminOverview` (`admin_handler.go:151`) | `user.view` | Concurrent errgroup: KPI + presence + recent audit events |
| GET | `/api/v1/iam/users` | `PeopleHandler.handleListUsers` (`people_handler.go:64`) | `user.view` | In-memory filter + cursor pagination over `auth.ListUsers` |
| POST | `/api/v1/iam/users/invite` | `PeopleHandler.handleInvite` (`people_handler.go:123`) | `user.manage` | Generates temp password (crypto/rand); audit excludes password |
| PATCH | `/api/v1/iam/users/{user_id}` | `PeopleHandler.handlePatch` (`people_handler.go:182`) | `user.manage` | Cross-tenant 404 guard; metadata + optional role replace |
| POST | `/api/v1/iam/users/bulk` | `PeopleHandler.handleBulk` (`people_handler.go:292`) | `user.manage` | activate/deactivate/unlock; force-logout → 501 (deferred) |
| POST | `/api/v1/iam/users/{user_id}/reset-password` | `PeopleHandler.handleResetPassword` (`people_handler.go:237`) | `user.manage` | Cross-tenant 404 guard |
| POST | `/api/v1/iam/users/{user_id}/unlock` | `PeopleHandler.handleUnlock` (`people_handler.go:269`) | `user.manage` | Cross-tenant 404 guard |
| GET | `/api/v1/iam/users/{user_id}/memberships` | `PeopleHandler.handleListMemberships` (`people_handler.go:356`) | `membership.view` | Cross-tenant 404 guard |
| POST | `/api/v1/iam/users/{user_id}/roles` | `AdminHandler.handleUserRoleUpsertTyped` (`admin_handler.go:126`) | `user.manage` | |
| PUT | `/api/v1/iam/users/{user_id}/roles` | `AdminHandler.handleReplaceUserRolesTyped` (`admin_handler.go:135`) | `user.manage` | |
| GET | `/api/v1/iam/roles` | `RolesCapsHandler.listRoles` (`routes_roles_caps.go:52`) | `membership.view` | Returns cached `CanonicalRoles()` catalog |
| GET | `/api/v1/iam/capabilities` | `RolesCapsHandler.listCapabilities` (`routes_roles_caps.go:60`) | `membership.view` | Returns cached `CapabilityCatalog()` |
| GET | `/api/v1/iam/role-capabilities` | `RolesCapsHandler.listRoleCapabilities` (`routes_roles_caps.go:68`) | `membership.view` | DB read of `role_capabilities` matrix |
| GET | `/api/v1/iam/usage` | `ObservabilityHandler.handleUsage` (`observability_handler.go:30`) | `metrics.view` | |
| GET | `/api/v1/iam/kpi` | `ObservabilityHandler.handleKpi` (`observability_handler.go:48`) | `metrics.view` | |
| GET | `/api/v1/iam/presence/snapshot` | `presence.Handler.handleSnapshot` (`presence/handler.go:69`) | `user.view` | |
| GET | `/api/v1/iam/presence/stream` | `presence.Handler.handleStream` (`presence/handler.go:86`) | `user.view` | WebSocket upgrade; idle timeout 90s |
| GET | `/api/v1/auth/sessions` | `SessionsHandler.handleSessions` (`sessions_handler.go:55`) | `session.manage` | |
| DELETE | `/api/v1/auth/sessions/{id}` | `SessionsHandler.handleSessionByID` | `session.manage` | Also POST `.../revoke-all` |

---

## 4. Logic flows

### 4.1 Tier-1 capability enforcement — `Middleware.Wrap`

1. **Strip headers** (`middleware.go:59-60`): `X-User-ID` and `X-User-Roles` deleted from the inbound request unconditionally — prevents header-spoofed role escalation.
2. **Nil resolver guard** (`middleware.go:63-66`): if `PermissionResolver` is nil, return 500; fail-closed, never pass unauthenticated.
3. **Resolve route** (`middleware.go:67`): call `m.resolver(method, path)` → returns `(Capability, Visibility)`.
4. **Public routes** (`middleware.go:77-79`): `VisibilityPublic` passes through immediately.
5. **Extract userID** (`middleware.go:85-89`): from `iamdomain.UserIDFromContext` — populated by the auth middleware upstream; 401 if absent.
6. **Extract tenantID** (`middleware.go:94-98`): from `tenant.FromContext` — populated by auth middleware from `auth_sessions.tenant_id`; 401 if absent. Legacy `X-Tenant-ID` fallback and `DevTenantID` fallback have been removed.
7. **Session-only routes** (`middleware.go:102-109`): `VisibilitySessionRequired` calls `resolveRoles` and passes through without a capability check.
8. **Capability check** (`middleware.go:119-122`): `m.caps.CanDo(ctx, userID, tenantID, capability)` — a single `SELECT EXISTS(UNION ALL)` over 4 branches; 403 on denial.
9. **Enrich context** (`middleware.go:124-128`): `resolveRoles` adds `iamdomain.WithAuthContext(rctx, userID, roles)` for downstream handlers.

### 4.2 Tier-2 enforcement inside a transaction — `authz.Require`

1. **Read GUCs** (`authz.go:77-84`): `MustActorID` and `MustTenantID` read `metaldocs.actor_id` and `metaldocs.tenant_id` from the live transaction via `SELECT current_setting(..., true)`. Return typed errors on empty.
2. **Cache short-circuit** (`authz.go:85-87`): if `(actor, tenant, cap, area)` was already granted in this context + tx pointer, skip DB and append to `asserted_caps`.
3. **System-admin bypass** (`authz.go:90-112`): `SELECT` + `SystemAdminExistsSQL` (direct role OR group-derived, `$1=actor $2=tenant`). On match, emit `BypassEvent{Kind:BypassKindSystemAdmin}` to the bypass audit sink (in-tx, fail-closed if the write fails), store in cache, append to `asserted_caps`, return nil.
4. **Area-grant check** (`authz.go:115-133`): `SELECT EXISTS(role_capabilities rc JOIN user_process_areas upa)` with `$1=capability $2=areaCode $3=actorID $4=tenantID`. `$2='tenant'` skips the area filter (tenant-grade caps). Returns `ErrCapDenied{Capability, AreaCode, ActorID}` on denial.
5. **Append to GUC** (`authz.go:135-136`): on grant, store in per-tx granted cache and append `{"cap":…,"area":…}` to `metaldocs.asserted_caps` JSON array via transaction-local `set_config(..., true)`.
6. **Tripwire fires** (DB-side): the `trg_require_cap_asserted` trigger on the mutating table reads `metaldocs.asserted_caps` and raises `ErrCapabilityNotAsserted` if the required cap is absent. This is the enforcement floor that cannot be circumvented if the trigger is attached and `session_replication_role` is not `replica`.

### 4.3 Grant area membership — `POST /api/v1/iam/area-memberships`

1. **Tier-1** (`middleware.go`): `CanDo(membership.manage)` — held by `area_admin` and `system_admin`.
2. **Self-grant block** (`routes_memberships.go:201-204`): `isSelf(ctx, userID)` compares authenticated actor with target; 403 if equal.
3. **Cross-tenant 404 guard** (`routes_memberships.go:211-213`): `guardMembershipUserInTenant` → `PeopleService.VerifyUserInTenant` → `auth.ListUsers(tenantID)` linear scan; 404 on miss.
4. **Role-change detection** (`area_membership_service.go:113-122`): `GetActiveByUserAndArea` fetches any existing active row. If found with same role → `ErrMembershipExists` → 409. If found with different role → `GrantAtomic`.
5. **Tier-2 enforcement** (repository, `user_area_repository.go:191`): `authz.SeedTxIdentity` seeds `actor_id`/`tenant_id` GUCs on the tx, then `authz.Require(membership.manage, areaCode)` — area-scoped since ADR 0022 Phase 3. System-admin bypasses; area_admin passes only within their managed areas.
6. **Write** (`user_area_repository.go:177-229`): `Insert` runs `INSERT public.user_process_areas`. `GrantAtomic` first asserts `oldMembership.AreaCode == newMembership.AreaCode` (hard invariant), then `UPDATE SET effective_to = newFrom, revoked_by = actorID` (satisfies `revoked_by_required_when_revoked` CHECK), then `INSERT` new row; COMMIT.
7. **Tripwire** (DB): `trg_require_cap_asserted` fires on `user_process_areas` INSERT/UPDATE; confirms `asserted_caps` contains `membership.manage`.
8. **Cache invalidation** (`area_membership_service.go:128`): `s.invalidate(userID, tenantID)` → `CachedRoleProvider.InvalidateUserTenant` — evicts the user's cache entry immediately.
9. **Governance log** (`area_membership_service.go:129-133`): only if `s.logger != nil`; wired as `nil` in production (`main.go:325`).
10. **Audit** (`routes_memberships.go:233-244`): `recordMembershipAudit` emits `iam.area_membership.granted` event; log-and-continue on failure.
11. **Response** (`routes_memberships.go:240-245`): 201 `{user_id, tenant_id, area_code, role}`.

### 4.4 Tenant-scoped role assignment — `POST /api/v1/iam/users/{user_id}/roles`

1. **Tier-1** (`middleware.go`): `CanDo(user.manage)` — held by `system_admin` only.
2. **Decode and validate** (`admin_handler.go:304-319`): `ParseRole` validates canonical 8-role enum.
3. **Delegate to service** (`admin_handler.go:330`): `AdminService.UpsertUserAndAssignRole` → validates, delegates to `RoleAdminRepository.UpsertUserAndAssignRole`.
4. **Tier-2 enforcement** (`role_admin_repository.go:43-48`): `authz.SeedTxIdentity` seeds GUCs; `authz.Require(user.manage, "tenant")` — tenant-grade cap.
5. **DELETE-then-INSERT** (`role_admin_repository.go:50-75`): `UPSERT iam_users ON CONFLICT DO UPDATE`; `DELETE iam_user_roles WHERE tenant_id=$1 AND user_id=$2`; `INSERT iam_user_roles`. All in one tx; COMMIT.
6. **Tripwire**: `trg_require_cap_asserted` on `iam_user_roles` confirms `user.manage` in `asserted_caps`.
7. **Cache invalidation** (`admin_service.go:48-50`): `s.invalidator.InvalidateUserTenant(userID, tenantID)`.
8. **Response + audit** (`admin_handler.go:335-343`): 200 `{user_id, role, display_name}`; `recordAudit` emits `iam.user.role.upserted`.

### 4.5 Presence stream — `GET /api/v1/iam/presence/stream`

1. **Tier-1** (`middleware.go`): `CanDo(user.view)`.
2. **Tenant extraction** (`presence/handler.go:92`): `tenant.FromContext`.
3. **Initial snapshot** (`presence/handler.go:111-116`): `repo.Snapshot(ctx, tenantID, now)` — SELECT of active `iam_users` within `OfflineAfter` (5 min), classify `online`/`idle`.
4. **Send initial state** (`presence/handler.go:118-120`): `{type:"snapshot", presence:[…]}` over the WebSocket.
5. **Subscribe** (`presence/handler.go:122`): `hub.Subscribe(tenantID, items)` — seeds `room.prev` so the first tick produces only genuine deltas.
6. **Hub tick** (`presence/hub.go:184-215`): every `TickInterval` (15s), calls `repo.Snapshot` per room, calls `diff(prev, curr)`, broadcasts join/leave/online/idle events to all `Conn`s.
7. **Client reader goroutine** (`presence/handler.go:128-136`): discards frames but enables ping-induced read deadlines; cancels `ctx` on any read error (drives graceful disconnect).
8. **Idle timer** (`presence/handler.go:138-163`): `ClientIdleTimeout` (90s) closes the connection on silence.
9. **Slow consumer protection** (`presence/hub.go:285-291`): if a conn's outbound buffer is full, `hub` drops the conn with `go c.Close()`.

---

## 5. Dependencies

### 5.1 Outbound imports (what IAM imports)

| Package | Why |
|---|---|
| `internal/modules/audit/domain` | `auditdomain.Writer`, `auditdomain.Event` — admin handler and membership handler emit audit events |
| `internal/modules/auth/application` | `authapp.Service` — `PeopleService` depends on `auth.Service` for user CRUD (`CreateUserWithInput`, `UpdateUser`, `AdminResetPassword`, `UnlockUser`, `ListUsers`) |
| `internal/modules/auth/domain` | `authdomain.ManagedUser`, `authdomain.OnlineUser`, `authdomain.UpdateUserParams`, `authdomain.CreateUserInput`, `authdomain.Session`; error sentinels `ErrUserAlreadyExists`, `ErrIdentityNotFound`, `ErrPasswordPolicy` |
| `internal/modules/auth/infrastructure/postgres` | `authpg.Repository`, `authpg.SessionAdminQuery`, `authpg.SessionListItem` — SessionsHandler depends on this directly (cross-module infra import) |
| `internal/platform/authn` | `authn.UserIDFromContext`, `authn.CacheTTL` |
| `internal/platform/httpresponse` | `httpresponse.WriteJSON` — aliased as `writeJSON` in the delivery package |
| `internal/platform/problem` | `problem.New`, `problem.Write` — RFC 9457 envelope |
| `internal/platform/tenant` | `tenant.FromContext` — primary tenant source |
| `github.com/google/uuid` | UUID generation for audit event IDs |
| `golang.org/x/sync/errgroup` | Concurrent reads in `handleAdminOverview` |
| `github.com/coder/websocket` | WebSocket upgrade, writer, read in presence handler |

### 5.2 Inbound imports (who imports IAM) — verified by grep

| Importer | What it imports |
|---|---|
| `apps/api/cmd/metaldocs-api/main.go` | All IAM packages for DI wiring |
| `apps/api/internal/wiring/documents.go` | `iamapp.CapabilityService` via `NewCapabilityChecker` adapter (J2 fix) |
| `internal/modules/documents/application/fillin_authz.go` | `iamdomain.Capability`, `authz.Require`, `authz.SeedTxIdentity`, `authz.WithBackgroundBypass` |
| `internal/modules/documents/application/view_service.go` | `iamdomain.Capability`, `authz.Require` |
| `internal/modules/documents/application/reconstruct_service.go` | `authz.Require` |
| `internal/modules/documents/repository/repository.go` | `authz.Require`, `iamdomain.Capability` |
| `internal/modules/documents/delivery/http/handler.go` | `iamapp.ErrCapabilityDenied`, `authz.ErrCapDenied` |
| `internal/modules/documents/http/fillin_handler.go` | `iamdomain.UserIDFromContext` |
| `internal/modules/documents/http/view_handler.go` | `iamdomain.UserIDFromContext` |
| `internal/modules/documents/jobs/session_sweeper.go` | `authz.WithBackgroundBypass`, `authz.BypassSystem` |
| `internal/modules/documents/jobs/orphan_pending_sweeper.go` | `authz.WithBackgroundBypass`, `authz.BypassSystem` |
| `internal/modules/documents/approval/application/` (5 files) | `authz.Require`, `authz.SeedTxIdentity`, `authz.BypassSystem`, `iamdomain.Capability` |
| `internal/modules/documents/approval/http/` | `iamdomain.RolesFromContext`, `iamapp.ErrCapabilityDenied` |
| `internal/modules/controlleddocuments/application/service.go` | `authz.Require`, `iamdomain.Capability` |
| `internal/modules/controlleddocuments/infrastructure/repository.go` | `authz.Require`, `iamdomain.Capability` |
| `internal/modules/taxonomy/infrastructure/` | `authz.Require`, `authz.SeedTxIdentity`, `iamdomain.Capability` |
| `internal/modules/templates/application/lifecycle.go` | `authz.Require`, `iamdomain.Capability` |
| `internal/modules/templates/delivery/http/` | `iamdomain.RolesFromContext`, `iamapp.CapabilityService` |
| `internal/modules/auth/application/service.go` | `iamdomain.Role` |
| `internal/modules/auth/domain/model.go` | `iamdomain.Role` |
| `internal/modules/auth/delivery/http/middleware.go` | `iamdomain.WithAuthContext`, `iamdomain.Role` |
| `internal/modules/auth/infrastructure/memory/repository.go` | `iamdomain.Role` |
| `internal/modules/search/delivery/http/handler.go` | `iamdomain.UserIDFromContext` |
| `internal/platform/bootstrap/api.go` | `iamdomain.Role`, `iamapp.DevRoleProvider` |
| `internal/platform/authn/context.go` | `iamdomain.UserIDFromContext` (through the authn bridge) |
| `internal/platform/security/ratelimit.go` | `iamdomain.UserIDFromContext` |
| `internal/jobs/stuck_instance_watchdog/job.go` | `authz.WithBackgroundBypass`, `authz.BypassSystem` |
| `scripts/api-lint/registry_rules.go` | `iamdomain.IsAreaGrade`, parses `model.go` const→value map |

---

## 6. Persistence

### Tables owned (write)

| Table | Schema | Key columns | Notes |
|---|---|---|---|
| `iam_users` | `metaldocs` | `user_id`, `tenant_id`, `display_name`, `is_active`, `deactivated_at`, `last_seen_at` | UPSERT by `UpsertUserAndAssignRole`/`ReplaceUserRoles`; bumped by `presence.BumpLastSeen`. Unique on `(tenant_id, user_id)` |
| `iam_user_roles` | `metaldocs` | `user_id`, `tenant_id`, `role_code`, `assigned_at`, `assigned_by` | DELETE + INSERT (role replacement). Tripwire: `trg_require_cap_asserted` (migration 0188). Unique: `ux_iam_users_tenant_user` |
| `iam_groups` | `metaldocs` | `id`, `tenant_id` | Group entity. Added migration 0163 |
| `iam_group_members` | `metaldocs` | `group_id`, `user_id`, `tenant_id` | Group membership |
| `iam_group_roles` | `metaldocs` | `group_id`, `role` | Group role assignments. No runtime write surface in IAM module today (T-008) |
| `user_process_areas` | `public` | `user_id`, `tenant_id`, `area_code`, `role`, `effective_from`, `effective_to`, `granted_by`, `revoked_by` | Area memberships; `effective_to IS NULL` = active. Tripwire: `trg_require_cap_asserted` (migration 0188). Partial unique: `ux_user_process_areas_single_active` |

### Tables read (not owned)

| Table | Schema | Query purpose |
|---|---|---|
| `role_capabilities` | `metaldocs` | Capability check in `CanDo` (4-branch), `authz.Require` (area grant), `MembershipDirectoryScope`, `ListByTenantInManagedAreas`, `ListRoleCapabilities` |
| `audit_events` | `metaldocs` | Observability: action-prefix count, active-user count, failed-login count |
| `auth_identities` | `metaldocs` | Observability: locked accounts (`CountLockedAccounts`), dormant users (`CountDormantUsers`), failed logins |
| `document_process_areas` | `metaldocs` | Area catalog reader: `AreaCodeExists` validation |
| `tenant_plans` | `metaldocs` | `GetTenantPlan` for Usage card |
| `auth_sessions` | `metaldocs` | `SessionsHandler` via `authpg.Repository` |

### Migration files (key IAM-related)

| Migration | Change |
|---|---|
| `0130_*` | Added `tenant_id` to `iam_users` |
| `0162_*` | Added `tenant_id` to `iam_user_roles`; unique constraint |
| `0163_*` | Added `iam_groups*` tables with `tenant_id` |
| `0165_*` | Reseeded `role_capabilities` to canonical namespace |
| `0166_*` | Unique `(tenant_id, user_id)` on `iam_user_roles` |
| `0188_tripwire_extend.sql` | Extended `enforce_capability_asserted()` to 12 tables including `iam_user_roles` and `user_process_areas` |
| `0210_*` | Capability namespace (`doc.*` → `document.*`) |
| `0217_*` | View-grade capabilities (`CapMetricsView`, `CapMembershipView`, `CapUserView`, `CapTaxonomyView`) |
| `0218_iam_caps_audit_session_pr2.sql` | `audit.read` grants; `session.manage` grant |
| `0219_iam_users_last_login_context.sql` | Last-login context columns on `iam_users` |
| `0220_iam_users_last_seen_at.sql` | `last_seen_at` column for presence |
| `0221_tenant_plans.sql` | `metaldocs.tenant_plans` table |
| `0225_authz_p2_document_lifecycle_grants.sql` | Seeded 4 previously-unseeded write caps |
| `0227_authz_p10_merge_redundant_caps.sql` | Merged 4 redundant caps; registry 33→29 |
| `0228_authz_p11_reserve_tenant_area_code.sql` | Reserved `area_code='tenant'` in `user_process_areas`; CHECK rejects it |
| `0229_authz_p12_rename_document_lifecycle_caps.sql` | `doc.*` → `document.*` cap-value rename |
| `0230_authz_decommission_reviewer_role.sql` | Decommissioned `reviewer` role |
| `db/reference-data/0001_product_reference_data.sql` | `role_capabilities` baseline seed (94 rows as of Phase 7) |

### Query patterns

- **Read-only (no tx):** `CanDo`, `RolesByUserID`, `ListActive`, `ListByTenant`, `MembershipDirectoryScope`, `ListByTenantInManagedAreas`, `GetActiveByUserAndArea`, all observability queries — plain `db.QueryRowContext` or `db.QueryContext`.
- **Mutations (own tx):** `UpsertUserAndAssignRole`, `ReplaceUserRoles`, `Insert`, `CloseActive`, `GrantAtomic` — each calls `db.BeginTx`, seeds GUCs, calls `authz.Require`, runs DML, COMMIT; defer Rollback guards.
- **Optional filter pattern:** `ListByTenant` uses `($n = '' OR col = $n)` for static parameterised optional filters (safe; injection-proof; no dynamic SQL).
- **Pagination in Go:** `PeopleService.ListFiltered` loads all users via `auth.ListUsers`, applies filters in Go, and does cursor pagination in-memory (known O(N) smell, documented in service package comment as PR-5/PR-11 work).

---

## 7. Config & environment

IAM consumes no standalone env vars of its own. All IAM behaviour is driven by the shared `authn.Enabled()` boolean (passed to `NewMiddleware` as `enabled bool`) and the `authn.CacheTTL()` duration (passed to `NewCachedRoleProvider`). Both are read from the platform `authn` config package which wraps the bootstrap env layer.

`DevTenantID = "ffffffff-ffff-ffff-ffff-ffffffffffff"` (`internal/platform/tenant/const.go:4`) is the sentinel used in single-tenant dev mode. The middleware's legacy fallback to `X-Tenant-ID` header or `DevTenantID` has been removed; primary tenant source is `tenant.FromContext` only.

The presence sub-system uses no env config; its constants (`TickInterval`, `BumpDebounce`, `ClientIdleTimeout`, `HeartbeatInterval`) are package-level constants in `presence/model.go`.

---

## 8. Concurrency & async

### Goroutines

| Goroutine | Location | Lifecycle |
|---|---|---|
| Role cache TTL sweep | `application/cached_role_provider.go:36-54` | Started by `NewCachedRoleProvider`; terminates on `ctx.Done()` |
| `BumpMiddleware` debounced write | `presence/middleware.go:92-98` | Per authenticated request (if outside debounce window); detached `context.Background()` with 2s timeout; fire-and-forget |
| `BumpMiddleware` cleanup sweep | `presence/middleware.go:117-129` | Started by `StartCleanup(ctx)`; terminates on `ctx.Done()`; evicts stale `lastBump` entries every `BumpDebounce*5` |
| Presence hub ticker | `presence/hub.go:112-125` | Started by `Hub.Run(ctx)`; terminates on `ctx.Done()` |
| Presence hub heartbeat | `presence/hub.go:237-247` | Started by `Hub.RunHeartbeat(ctx)`; terminates on `ctx.Done()` |
| Presence handler WS reader | `presence/handler.go:128-136` | Per WebSocket connection; cancels parent ctx on read error |
| Admin overview concurrent reads | `admin_handler.go:175-220` | `errgroup.WithContext` within request scope; 3 goroutines (KPI + presence + audit); scoped to the request lifetime |

### Channels

- `Conn.out` (`presence/hub.go:52`): buffered `chan Event` (capacity 32) per WebSocket connection. The hub drops slow consumers rather than blocking.
- `Hub.started`, `Hub.stopped` (`presence/hub.go:28-29`): unbuffered signalling channels for `Run` lifecycle.

### Outbox / async writes

None. All IAM mutations are synchronous, inline with the HTTP request transaction. There is no outbox, job queue, or deferred write path in this module.

### In-memory shared state under locks

- `CachedRoleProvider.items` (`application/cached_role_provider.go:22`): `sync.RWMutex`-protected map.
- `BumpMiddleware.lastBump` (`presence/middleware.go:28`): `sync.Mutex`-protected map.
- `Hub.rooms` (`presence/hub.go:28`): `sync.Mutex`-protected map of rooms; each `room` has its own `sync.Mutex` for its `conns` and `prev` maps.
- `bypassAuditSink` (`authz/bypass_audit.go:50`): `atomic.Pointer[BypassAuditSink]` — set once at startup, read on every bypass.

---

## 9. Error handling & observability

### Error patterns

- **Sentinel errors** (`errors.New`): `ErrCapabilityDenied` (`application/capability_service.go:12`), `ErrMembershipNotFound`, `ErrUnknownRole`, `ErrMembershipExists` (`area_membership_service.go:13-20`), `ErrUserNotInTenant`, `ErrCursorExpired`, `ErrPeopleValidation`, `ErrAreaUnknown` (`people_service.go:36-51`), `ErrActorContextMissing`, `ErrTenantContextMissing` (`authz/context.go:14,18`), `ErrBypassNotBackground` (`authz/authz.go:145`), `ErrTenantPlanNotFound` (`domain/observability_port.go:58`), domain errors in `domain/errors.go`.
- **Typed struct error** (`errors.As`): `authz.ErrCapDenied{Capability, AreaCode, ActorID}` (`authz/authz.go:36-44`) — carries context for audit; mapped to 403 in `writeMembershipError` via `errors.As`.
- **Wrapped errors** (`fmt.Errorf("...: %w", err)`): all repository and service methods wrap errors with context; callers use `errors.Is` against sentinels.
- **`ErrNoRolesAssigned`** (`domain/errors.go:9`): returned by `RoleProvider` when DB query succeeds but the user has no role rows; `Middleware.Wrap` maps this to 403.

### RFC 9457 Problem responses

All HTTP error responses go through `problem.Write(w, problem.New(status, code, message))` (`internal/platform/problem/problem.go`). This sets `Content-Type: application/problem+json` and serializes `type/title/status/detail/instance/code/errors` per RFC 9457. No module-local `writeAPIError` wrapper exists any longer (T-006 closed, 2026-05-12).

### Logging

- **`log.Printf`**: used in delivery handlers for non-critical path errors (audit write failures, list failures, verify-user failures) — `admin_handler.go:221`, `routes_memberships.go:131`, `people_handler.go:97`.
- **`slog`** (`log/slog`): used in `admin_handler.go` (warn on write failure), `observability_handler.go` (Error for service failures), `user_area_repository.go` (warn on zero-rows mutations), `presence/hub.go` (warn on snapshot failure, slow consumer drop).
- **No structured tracing spans** are emitted by IAM. The `X-Trace-Id` header is threaded into audit event `TraceID` fields as a pass-through correlation ID.
- **No metrics** (Prometheus/OpenTelemetry) are emitted by IAM. Observability is provided exclusively through the audit trail and the KPI/usage endpoints.

### Audit emission points

| Event action | Emitter | Notes |
|---|---|---|
| `iam.user.role.upserted` | `admin_handler.go:340` | After successful `handleUserRoleUpsert` |
| `iam.user.roles.replaced` | `admin_handler.go:379` | After successful `handleReplaceUserRoles` |
| `iam.user.invited` | `people_handler.go:167` | Excludes temp password |
| `iam.user.updated` | `people_handler.go:226` | Metadata + role changes (no password material) |
| `iam.user.bulk.<action>` | `people_handler.go:334` | Per successful user in bulk op |
| `auth.user.password_reset` | `people_handler.go:259` | |
| `auth.user.unlocked` | `people_handler.go:286` | |
| `iam.area_membership.granted` | `routes_memberships.go:233` | |
| `iam.area_membership.revoked` | `routes_memberships.go:287` | |
| `authz.bypass.system_admin` | `authz/authz.go:101` (via `bypassAuditSink`) | In-tx; fail-closed |
| `authz.bypass.background` | `authz/authz.go:182` (via `bypassAuditSink`) | In-tx |

All audit writes are log-and-continue after the mutating write commits (except bypass audit, which is in-tx and fail-closed).

---

## 10. Legacy / duplication / smell flags

- **FLAG-01 (medium): `PeopleService.VerifyUserInTenant` is an O(N) full-scan guard** — `people_service.go:585-600`: `VerifyUserInTenant` calls `auth.ListUsers(tenantID)` (full table load) and linearly scans for `userID`. This pattern is repeated in `ListFiltered` (`people_service.go:518`) and in `BulkAction` (`people_service.go:457`). At scale, every guarded mutation (patch, reset-password, unlock, membership verify) pays the cost of loading the entire tenant's user list. The service package comment at line 1 documents it as a PR-5/PR-11 debt item. `RF-XXX` (no specific RF ID in the target architecture register for this pattern).

- **FLAG-02 (medium): `PeopleService.ListFiltered` applies all filters in Go, not SQL** — `people_service.go:617-648` (`applyPeopleFilters`): filters for `is_active`, `role`, `area_code`, `q` are applied post-load in Go after `auth.ListUsers` returns the full tenant user list. Cursor pagination is similarly in-memory (`decodeCursorIndex` `people_service.go:668`). The package comment acknowledges this is a PR-5/PR-11 deferral; acceptable for small datasets, degrades linearly with tenant size.

- **FLAG-03 (major): `MembershipGovernanceLogger` wired as `nil` in production** — `apps/api/cmd/metaldocs-api/main.go:325`: `NewAreaMembershipService(..., nil)`. The `AreaMembershipService.Grant` and `Revoke` methods nil-check before calling (`area_membership_service.go:129,148,182,196`), so grant/revoke produce no governance log at the application-service layer. The SECURITY DEFINER SQL path (now deleted) previously wrote to `governance_events` automatically; the direct-DML path does not. Recorded as T-007 in `iam-tech-debt.md`.

- **FLAG-04 (low): `iam_users` INSERT inside `UpsertUserAndAssignRole`/`ReplaceUserRoles` has no dedicated tripwire** — `role_admin_repository.go:52-58,103-109`: the `INSERT INTO metaldocs.iam_users` (UPSERT) executes within the same tx as the `authz.Require(CapUserManage)` call, so the enclosing tx's capability check functionally covers it. However, the tripwire trigger (`trg_require_cap_asserted`) is not attached to `metaldocs.iam_users`; only `iam_user_roles` and `user_process_areas` are covered. The DB enforcement floor is absent for the user-record write. Residual from T-004 (partially closed by Plan 5).

- **FLAG-05 (low): Two inline TODO comments referencing migration 0002 cascade FK concern** — `role_admin_repository.go:62` and `:114`: `// TODO: migration 0002 uses cascading FKs from iam_user_roles; keep role replacement explicit and review any future hard-delete path carefully`. These are informational caution comments, not actionable bugs, but they signal a design concern around hard-delete semantics that has never been documented in an ADR.

- **FLAG-06 (low): Two inline TODO comments referencing migration 0136 non-PK unique key** — `user_area_repository.go:201` and `:346`: `// TODO: migration 0136 hangs these FKs off a non-PK unique key on iam_users; keep caller identity writes aligned with iam_users uniqueness`. Same nature as FLAG-05.

- **FLAG-07 (low): `CachedRoleProvider` has no max-size cap** — `application/cached_role_provider.go:22`: the comment explicitly acknowledges: "The cache has no max-size cap, so a very large distinct (user, tenant) working set could grow it between sweeps; acceptable at current scale." At multi-tenant SaaS scale, unbounded map growth under user churn is a latent OOM risk.

- **FLAG-08 (low): `sessions_handler.go` imports `auth/infrastructure/postgres` directly** — `sessions_handler.go:19`: `authpg "metaldocs/internal/modules/auth/infrastructure/postgres"`. This is a cross-module infra import from a delivery handler (IAM delivery importing auth infrastructure). The `SessionAdmin` interface defined in the same file (`sessions_handler.go:29`) abstracts the behaviour, but the type `authpg.SessionListItem` and `authpg.SessionAdminQuery` leak through the interface boundary, creating a compile-time dependency on the auth infra layer from IAM delivery.

- **FLAG-09 (low): `parseRoles` is defined in `admin_handler.go` but is not used by any current caller** — `admin_handler.go:418-435`: `parseRoles` returns `([]Role, bool)` and is a superset of `parseExactlyOneRole`. The only call site in this file is `parseExactlyOneRole` which invokes it internally. `parseRoles` is not exported and is not referenced anywhere else in the codebase; it exists only as a helper for a multi-role future that does not yet exist. It is dead code at the module level but is not entirely dead (called by `parseExactlyOneRole`), so it does not strictly qualify as dead — flagged as a reachable-but-single-consumer internal helper with no clear future consumer.

- **FLAG-10 (low): `RolesByUserID` in `postgres/role_provider.go` makes two round trips** — `role_provider.go:20-57`: first checks `iam_users WHERE user_id = $1 AND tenant_id = $2 AND deactivated_at IS NULL` (line 21-30), then queries `iam_user_roles` (line 34-57). These could be merged into a single JOIN query, especially given this is on the hot path of every authenticated request (the CachedRoleProvider absorbs most of the cost, but the two-query pattern fires on every cache miss).

- **FLAG-11 (info): `authenticatedActor` fallback returns `"system"` string** — `admin_handler.go:448-453`: when `authn.UserIDFromContext` has no actor, `authenticatedActor` returns the literal string `"system"`. This propagates into audit `ActorID` fields on background or unauthenticated paths. The same pattern appears in `authz/bypass_audit.go:89` (`softActorID` fallback). This is deliberate for background jobs but could produce misleading audit entries if an HTTP request ever reaches this path without a populated context.

---

## 11. Wiki drift

The following claims in `wiki/modules/iam.md` (last verified 2026-06-09) and `wiki/modules/iam-tech-debt.md` (last verified 2026-06-03) were cross-checked against the current code and are **accurate** on the checked-out `qa/iam-area-membership` branch. No drift was detected.

However, there are two specific areas where the existing wiki documentation contains stale line references that no longer match the current code:

1. **`wiki/modules/iam.md`, API Route Truth Table (lines 214-222)**: The route truth table at lines 215-221 still describes routes as being handled via the `handleUserRoute` suffix dispatcher (`admin_handler.go:85`) and `handleUsers` method. In the current code, `admin_handler.go:120-124` uses `RegisterRoutes` with typed Go 1.22 mux patterns and named methods `handleUserRoleUpsertTyped` / `handleReplaceUserRolesTyped`. The `handleUserRoute` suffix dispatcher was removed in PR-4. The People-tab endpoints listed in the truth table (list, invite, patch, bulk, reset, unlock, memberships) are now on `PeopleHandler` not `AdminHandler`, and `people_handler.go` exists — the wiki table correctly documents this in §5.3 but the older route truth table block retains the pre-PR-4 wiring description.

2. **`wiki/modules/iam.md`, Key files anchor `middleware.go:77`**: The key files block (line 23) references `internal/modules/iam/delivery/http/middleware.go:77` as the `tenant.FromContext` call. The current `middleware.go:94` is where `tenant.FromContext` is called; line 77 now contains the `VisibilitySessionRequired` branch check. The behaviour is correct but the line anchor is stale by approximately 17 lines.

3. **`wiki/concepts/authz-tiers.md`, Key files anchor `domain/model.go:15`**: States "27 total" capabilities. The current `model.go` has 29 typed consts in `validCapabilities` (registry minimized from 33 to 29 in ADR 0022 Phase 10). The count is stale.

4. **`wiki/modules/iam.md`, §5.2 Public surface "18 consts"**: Line 166 states "18 consts" for the capability namespace. The current `model.go` has 29 typed capability consts after the Phase 10 minimization and Phase 1-12 additions. This section of §5.2 has not been updated since before ADR 0022 Phase 1.

5. **`wiki/modules/iam.md`, DELETE route shape**: Line 214 (route truth table) shows the DELETE route as `DELETE /api/v1/iam/area-memberships` with query params `userId` + `areaCode`. The current `routes_memberships.go:92` registers `DELETE /api/v1/iam/area-memberships/{user_id}/{area_code}` with path params, changed in the Phase E2 hardening referenced in the `routes_memberships.go` comment at line 90 (`F-DELETE-SHAPE`). This is a meaningful contract drift.

---

## 12. Open questions

- **[runtime-unverified]** `T-007` — `MembershipGovernanceLogger` is wired as `nil` in production (`main.go:325`). It is unclear if any consumer downstream relies on governance log events for audit compliance or if the audit events emitted by `recordMembershipAudit` (`iam.area_membership.granted/revoked`) fully cover the ISO 9001 requirement for this write. The wiki notes `T-007` as open (major severity).

- **[runtime-unverified]** Presence hub multi-instance gap — `presence/hub.go:15`: the comment states "multi-instance scaling via Redis pub/sub is deferred". If more than one API process runs in production, presence state is not shared across instances (each hub is in-process only). This is documented as out-of-scope for PR-9 but is a production correctness concern for horizontally-scaled deployments.

- **[runtime-unverified]** `StorageUsedBytes` always returns `-1` (`infrastructure/postgres/observability_repository.go:41`). The FE renders `"—"` instead of a value. The resolution path (adding `tenant_id` to attachment tables or a materialized rollup) is documented as T-PR8-2 in the tech-debt register but has no migration or implementation.

- **[runtime-unverified]** `CountAuditEventsByActionPrefix` with prefix `"http.request."` always returns `0` because no audit events are emitted with this namespace (`observability_repository.go:56-84` comment, T-PR8-3). The API calls usage pane always shows 0.

- **[runtime-unverified]** The `document.create` and `document.edit` capabilities are classified `ScopeArea` in `capability_scope.go:39,40` (i.e. area-grade), and ADR 0022 Phase 7 aligned their tier-2 call sites to pass real area codes. However, the corresponding OpenAPI operations do not carry `x-authz-area` annotations (only `x-authz-skip-area` with documented exceptions per `permissions_authz_scope_test.go`). Whether the annotation-vs-enforcement gap is a known and accepted exception or a documentation debt requires confirmation against the ADR 0022 Phase 5 scope boundary.

- **[runtime-unverified]** `forceReleaseDocumentSession` resolves to tier-1 cap `membership.manage` while its tier-2 check uses `document.edit` (cross-tier cap divergence noted in ADR 0022 Phase 7 open follow-ups). The current IAM module does not control this; it is in the documents module, but the divergence is a known pre-existing defect surfaced during the ADR 0022 review.
  > RESOLVED in ADR 0022 Phase 11 F4 (see ADR 0022 §349-351). Tier-2 `ForceReleaseSession`/`ForceReleaseSessionTx` (`internal/modules/documents/repository/repository.go`) now asserts `CapMembershipManage`, matching tier-1. Regression-pinned by `TestTier1Tier2CapabilityCoherence_F4Sites` (`apps/api/cmd/metaldocs-api/permissions_test.go`).
