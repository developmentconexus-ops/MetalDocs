# Tech Debt Register — iam

> Companion to [`wiki/modules/iam.md`](iam.md). Debt only — no fix prescriptions. Fixes live in [`wiki/backlog/iam-refactor.md`](../backlog/iam-refactor.md).

**Last verified:** 2026-06-02 (PR-7b hardening closed 8 findings — see T-PR7B-1..8)

### T-PR7B-1 · CRITICAL — cross-tenant ATO via `handleResetPassword` — CLOSED 2026-06-02
- **Severity:** critical (closed)
- **Observation:** `PeopleHandler.handleResetPassword` delegated to `authSvc.AdminResetPassword` without verifying the target userID belonged to the caller's tenant; `auth_identities` has no tenant_id column so SQL writes were tenant-blind. Allowed cross-tenant account takeover.
- **Resolution:** Added `PeopleService.VerifyUserInTenant` (membership probe via `auth.ListUsers(tenantID)`) + `guardUserInTenant` helper on `PeopleHandler`. Returns 404 NOT 403 to avoid leaking existence in other tenants.

### T-PR7B-2 · CRITICAL — cross-tenant unlock — CLOSED 2026-06-02
- **Severity:** critical (closed)
- **Observation:** Same shape as T-PR7B-1 on `handleUnlock`.
- **Resolution:** Same guard.

### T-PR7B-3 · CRITICAL — PeopleHandler routes fell through to SessionRequired — CLOSED 2026-06-02
- **Severity:** critical (closed)
- **Observation:** `POST /iam/users/invite`, `POST /iam/users/bulk`, `GET /iam/users/{userId}/memberships` were not enumerated in `apps/api/cmd/metaldocs-api/permissions.go` routeRules. Fallback returned `VisibilitySessionRequired`, so any authenticated session (Viewer included) bypassed the capability gate.
- **Resolution:** Added explicit rules with `CapUserManage` (invite, bulk) and `CapMembershipView` (memberships). Regression test `TestPermissionResolver_PeopleHandlerRoutes` locks the contract.

### T-PR7B-4 · HIGH — `tests/unit` package failed to build — CLOSED 2026-06-02
- **Severity:** high (closed)
- **Observation:** Three compile errors after PR-2/PR-5 interface drift (NewService now returns 2 values, NewDevRoleProvider gained allowedTenantID, RoleCacheInvalidator gained InvalidateUserTenant). Reset/unlock tests also still pointed at the retired AdminHandler routes.
- **Resolution:** Stub method on `fakeInvalidator`, updated call sites, rerouted tests via PeopleHandler.

### T-PR7B-5 · HIGH — LIKE metacharacter injection on `action` filter — CLOSED 2026-06-02
- **Severity:** high (closed)
- **Observation:** `audit/infrastructure/postgres/writer.go` built `LIKE` patterns directly from user input with no `%` / `_` escaping. Intra-tenant over-matching.
- **Resolution:** New `internal/platform/sqlescape` package; applied `LikeEscape` + `ESCAPE '\'` clause to action prefix and free-text q paths.

### T-PR7B-6 · HIGH — missing index on `auth_identities.last_failed_login_at` — CLOSED 2026-06-02
- **Severity:** high (closed)
- **Observation:** Migration 0210 added the column but only indexed `locked_until`. `CountRecentFailedLoginsByUser` seq-scanned per `/security/signals` call.
- **Resolution:** Migration 0211 partial index `WHERE last_failed_login_at IS NOT NULL`.

### T-PR7B-7 · MEDIUM — `GetExportStatus` silent bypass on empty actor — CLOSED 2026-06-02
- **Severity:** medium (closed)
- **Observation:** Service skipped per-actor ownership check when actorID was empty; handler discarded the `ok` bool from `authn.UserIDFromContext`. Edge layer covers HTTP today but the path was a latent defense-in-depth gap.
- **Resolution:** Service fails-closed with `ErrActorRequired`; handler returns 401 when context has no userID.

### T-PR7B-8 · MEDIUM — silent page-1 reset on stale cursor — CLOSED 2026-06-02
- **Severity:** medium (closed)
- **Observation:** `decodeCursorIndex` returned 0 when the anchor user was no longer in the filtered set; caller silently restarted pagination with `hasMore=true`. Client appending without dedup got duplicate rows / could loop.
- **Resolution:** `decodeCursorIndex` now returns `(idx, found)`; service emits `ErrCursorExpired`; handler maps to 410 `CURSOR_EXPIRED`.

### T-PR8-1 · MEDIUM — `tenant_plans` mutation surface is Tier-A only (open) — 2026-06-02
- **Severity:** medium (open, by design)
- **Surface:** `internal/modules/iam/infrastructure/postgres/observability_repository.go::GetTenantPlan` (read-only); `db/migrations/0221_tenant_plans.sql` (table + backfill).
- **Observation:** PR-8 reads `metaldocs.tenant_plans` to render the Admin Center Usage card. Upgrade/downgrade, billing, overage flows belong to the **Tier-A platform-owner surface** and are intentionally NOT implemented at Tier-B (tenant admin). Tenant admins see their quota envelope; they cannot change it from the SaaS app.
- **Resolution path:** stand up a separate Tier-A admin surface (`/platform/...`) with `CapPlatformAdmin` cap; wire UPDATE / billing webhook ingestion there.

### T-PR8-2 · MEDIUM — `storage.usedBytes` returns -1 (open) — 2026-06-02
- **Severity:** medium (open)
- **Surface:** `internal/modules/iam/infrastructure/postgres/observability_repository.go::StorageUsedBytes`.
- **Observation:** Blob-layer byte aggregation is not yet tenant-scoped end-to-end (`document_attachments` has no `tenant_id`, requires JOIN via `documents`; revision artifacts spread across multiple tables). PR-8 returns `-1` so the FE renders "—" rather than fabricating a zero.
- **Resolution path:** add `tenant_id` columns or a per-tenant materialized rollup that sums `document_attachments.size_bytes`, `document_versions.file_size_bytes`, `document_images.byte_size`, `document_exports.size_bytes`.

### T-PR8-3 · LOW — `usage.apiCalls` counts zero (open) — 2026-06-02
- **Severity:** low (open)
- **Surface:** `internal/modules/iam/infrastructure/postgres/observability_repository.go::CountAuditEventsByActionPrefix` with `actionPrefix = "http.request."`.
- **Observation:** No audit action under the `http.request.*` namespace exists today, so the query always returns 0. Acceptable for v1 — the FE shows 0 across the 24h/7d/30d panes.
- **Resolution path:** either start emitting a coarse `http.request.<route>` audit row at the API edge, or back the count with a dedicated request-log table.

---

**Prior verified:** 2026-05-26 (Wave 2 authz tx seeding sync)

## Severity scale

Rubric (concrete triggers) lives in `.claude/skills/metaldocs-module-doc/templates/tech-debt-register.md`. Summary:

- **Critical** — authn/authz bypass · regulated audit-trail gap · multi-tenant data leak · data-loss path · contract violation downstream depends on · schema/version drift the boot check is supposed to catch but does not.
- **Major** — defense-in-depth single-point gap · governance/observability sink nil on regulated path · duplicated write surfaces with different semantics · documented contract not yet followed by this module · cross-module dep that blocks another module's clean refactor.
- **Minor** — symbol collision · missing doc comments · latent (no current consumer) · non-circular bidirectional dep · missing standalone ADR for already-enforced rule.

When triggers overlap: pick the highest matching tier and justify in the row's `Observation`.

## Items

### T-001 · Dual capability namespaces — CLOSED 2026-05-11 (Plan 4)
- **Severity:** critical (closed)
- **Surface:** `internal/modules/iam/domain/capabilities.go:4-19` (16 string consts `CapDocView`, `CapDocCreate`, …) and `internal/modules/iam/domain/model.go:16-20` (5 typed `Capability` consts `CapDocumentView`, `CapDocumentCreate`, `CapDocumentEdit`, `CapWorkflowReview`, `CapWorkflowApprove`)
- **Observation:** Two parallel capability namespaces exist with overlapping semantics. `domain/capabilities.go` defines `doc.view` / `doc.create` / `doc.edit`; `domain/model.go` defines `document.view` / `document.create` / `document.edit` plus `workflow.review` / `workflow.approve`. The `role_capabilities` DB table (migration 0165) seeds the `doc.*` / `template.*` / `registry.*` (legacy literal capability namespace for controlled-documents) / `taxonomy.*` / `membership.*` / `route.*` / `user.*` namespace from `capabilities.go`. The typed `Capability` constants from `model.go` are imported by `internal/modules/documents/application/fillin_authz.go:9` and `apps/api/internal/wiring/documents.go:7`.
- **Evidence:** `_artifacts/01-surface.md` rows 120-152; `_artifacts/03-deps.md` §2 (documents importers).
- **Linked backlog row:** `backlog/iam-refactor.md#R-001`
- **Linked ADR:** missing-ADR (no decision recording why two namespaces coexist)
- **Consumer cross-ref:** `wiki/modules/documents-tech-debt.md#t-008` — documents module straddles both namespaces; closure here unblocks documents R-008

### T-002 · Two area-membership write surfaces — CLOSED 2026-05-11 (Plan 4)
- **Severity:** major (closed)
- **Surface:** `internal/modules/iam/area_membership/area_membership.go:53,65,77` (free `Grant`/`Revoke`/`List` taking `*sql.Tx`, calling `metaldocs.grant_area_membership` / `revoke_area_membership` SECURITY DEFINER funcs) vs `internal/modules/iam/application/area_membership_service.go:49,108` (`AreaMembershipService.Grant`/`Revoke`) calling `UserAreaRepository.GrantAtomic` (`infrastructure/postgres/user_area_repository.go:90`) with direct DML
- **Observation:** Two implementations of the same use case exist with different semantics. The v2 HTTP route at `/api/v1/iam/area-memberships` (POST) uses the application-service + repo path (artifact 02-flow-grant-membership). The `area_membership/` package is wired into none of the routes registered in `main.go` (per artifact 03 §3 DI touchpoints) — its callers are not in the IAM module. SECURITY DEFINER funcs `metaldocs.grant_area_membership` reads `metaldocs.actor_id` GUC; the direct-DML path does not.
- **Evidence:** `_artifacts/01-surface.md` (both surfaces); `_artifacts/04-persistence.md` §5 (tripwire pairing rows for both); `_artifacts/03-deps.md` §3.
- **Linked backlog row:** `backlog/iam-refactor.md#R-002`
- **Linked ADR:** missing-ADR

### T-003 · `AuthorizationService` is a third authz surface, unused in production — CLOSED 2026-05-11 (Plan 4)
- **Severity:** major (closed)
- **Surface:** `internal/modules/iam/application/authorization.go:42` (`AuthorizationService`), `:49` (`NewAuthorizationService`), `:81` (`Check(ctx, userID, tenantID, capability, ResourceCtx) error`); plus `ErrSoDViolation` `:16`, `TemplateAuthorChecker` iface `:33`, `AccessPolicy` `:24`, `WithAuthzCache` `:74`
- **Observation:** `AuthorizationService` exposes resource-aware authz (`ResourceCtx`) with SoD probing — distinct from tier-1 (`CapabilityService.CanDo`) and tier-2 (`authz.Require`). It is not wired in `apps/api/cmd/metaldocs-api/main.go` (artifact 03 §3 lists 8 DI touchpoints; none constructs `AuthorizationService`). The benchmark (`application/authorization_bench_test.go`) and unit test (`authorization_test.go`) exercise it in isolation. Three authz surfaces compete; only two are live.
- **Evidence:** `_artifacts/01-surface.md` rows 68-79; `_artifacts/03-deps.md` §3 (no constructor call).
- **Linked backlog row:** `backlog/iam-refactor.md#R-003`
- **Linked ADR:** missing-ADR

### T-004 · IAM mutations have neither tier-2 nor tripwire enforcement — PARTIALLY CLOSED 2026-05-11 (Plan 5)
- **Severity:** major → **partially resolved** (residual: `iam_users` INSERT still tier-1 only)
- **Surface (resolved):** `infrastructure/postgres/role_admin_repository.go:34,76` (`UpsertUserAndAssignRole` at `:40`, `ReplaceUserRoles` at `:82` now call `authz.Require(CapUserManage)`); `infrastructure/postgres/user_area_repository.go:52,84,109` (`Insert` at `:59`, `CloseActive` at `:91`, `GrantAtomic` at `:118` now call `authz.Require(CapMembershipManage)`). `migrations/0188_tripwire_extend.sql` attaches `trg_require_cap_asserted` to `metaldocs.iam_user_roles` (line 187) and `metaldocs.user_process_areas` (line 192).
- **Surface (residual):** `iam_users` INSERT inside `UpsertUserAndAssignRole` (`role_admin_repository.go:50`) and `ReplaceUserRoles` (`:92`) is still not explicitly guarded at tier-2 on the `iam_users` table itself (no separate tripwire trigger on `metaldocs.iam_users`). The capability check on the enclosing tx covers it functionally, but DB-layer enforcement is absent.
- **Observation (original):** All IAM-owned mutating tables (`iam_user_roles`, `user_process_areas`, `iam_users`) were guarded by tier-1 middleware only. `authz.Require(...)` was called by none of these repository methods. The Postgres tripwire `enforce_capability_asserted` was attached to `public.approval_instances` and `public.approval_signoffs` only (`migrations/0142b_role_capabilities_v2_enforce.sql:200-209`), not to any IAM-owned table. The defense-in-depth pattern (IP-004 in `references/industry-patterns-index.md`) was therefore single-layer for IAM admin writes.
- **Evidence:** `_artifacts/04-persistence.md` §3, §5; `_artifacts/05-industry.md` §IP-004.
- **Linked backlog row:** `backlog/iam-refactor.md#R-004`
- **Linked ADR:** [`wiki/decisions/0007-two-tier-authz.md`](../decisions/0007-two-tier-authz.md) (documents tiers but does not address IAM-table coverage)

### T-005 · Admin role upsert does not emit audit events — CLOSED 2026-05-11 (Plan 6a)
- **Severity:** critical (closed)
- **Severity rationale:** triggers Critical rubric — "regulated audit-trail gap: a mutation on an ISO 9001 / QMS / regulated path is not written to the audit sink." Role assignment is the privileged op an external auditor inspects first; absence from the trail is a compliance break, not a service degradation.
- **Surface:** `internal/modules/iam/delivery/http/admin_handler.go:319` (`handleUserRoleUpsert`) and `:454` (`recordAudit`)
- **Observation:** `handleUserRoleUpsert` (POST `/api/v1/iam/users/{userId}/roles`) does not call `recordAudit` between request validation and response (artifact 02-flow-upsert-user-role §6). The audit sink is wired (`auditdomain.Writer` passed into `NewAdminHandler` at `main.go:182`; sink impl at `internal/modules/audit/infrastructure/postgres/writer.go:20`). Other admin ops do call `recordAudit`; this one does not.
- **Evidence:** `_artifacts/02-flow-upsert-user-role.md` §6; `_artifacts/03-deps.md` §1 (audit OUT edge).
- **Linked backlog row:** `backlog/iam-refactor.md#R-005`
- **Linked ADR:** missing-ADR (audit-emission policy not formalised)

### T-006 · IAM error envelope is not RFC 9457 — CLOSED 2026-05-12 (Plan 7)
- **Severity:** major (closed)
- **Surface (resolved):** `internal/modules/iam/delivery/http/middleware.go:73,78,87,98,101,103` — local `writeAPIError`/`apiErrorEnvelope` deleted; all error paths call `problem.Write(w, problem.New(...))` directly. `internal/modules/iam/delivery/http/routes_memberships.go:38,47,53,59,67,73,77,82,96,110,116,119,124,139,141,143` — `writeMembershipAPIError` and inline error sites use `problem.Write`. Two non-RFC-9457 shapes are gone.
- **Observation (original):** IAM used two non-9457 shapes: `{error:{code,message,details,trace_id}}` in middleware and `{code,message}` in membership routes. No `type` URI, no `title`, no `status`, no `errors[]`.
- **Evidence:** `_artifacts/05-industry.md` §IP-001.
- **Linked backlog row:** `backlog/iam-refactor.md#R-006` (merged Plan 7 2026-05-11, commit `1ecfe674`)
- **Linked ADR:** `wiki/architecture/api-design-system.md`

### T-007 · `MembershipGovernanceLogger` wired with `nil` in production
- **Severity:** major
- **Surface:** `apps/api/cmd/metaldocs-api/main.go:217` (`NewAreaMembershipService(iampg.NewUserAreaRepository(deps.SQLDB), nil)`); consumer at `internal/modules/iam/application/area_membership_service.go:79,101`
- **Observation:** The second argument to `NewAreaMembershipService` is the `MembershipGovernanceLogger`. In production wiring it is `nil`. The service nil-checks before calling (`area_membership_service.go:79`), so grant/revoke produce no governance log. For the SECURITY DEFINER path (`area_membership/area_membership.go`), governance events ARE written by the SQL function itself (artifact 04 §5 note). The two write paths therefore emit different governance trails.
- **Evidence:** main.go:217 (verified by main agent read); artifact 02-flow-grant-membership §6.
- **Linked backlog row:** `backlog/iam-refactor.md#R-007`
- **Linked ADR:** missing-ADR

### T-008 · `CachedRoleProvider` not invalidated on group membership writes
- **Severity:** minor
- **Surface:** `internal/modules/iam/application/cached_role_provider.go:65` (`InvalidateUser`); admin write site `application/admin_service.go:42`
- **Observation:** `CachedRoleProvider.RolesByUserID` is invalidated after `AdminService.UpsertUserAndAssignRole` and `ReplaceUserRoles`. There is no admin route or service method that mutates `iam_group_members` / `iam_group_roles` in the IAM module today, so no live invalidation gap exists. If such routes are added, group changes will not invalidate the role cache. Recorded as latent debt.
- **Evidence:** `_artifacts/01-surface.md` rows for `CachedRoleProvider`; `_artifacts/03-deps.md` (no group-write site).
- **Linked backlog row:** `backlog/iam-refactor.md#R-008`
- **Linked ADR:** missing-ADR

### T-009 · `ErrCapabilityDenied` exists in two packages with different shapes — CLOSED 2026-05-11 (Plan 4)
- **Severity:** minor (closed)
- **Surface:** `internal/modules/iam/application/capability_service.go:10` (sentinel `error` var) vs `internal/modules/iam/authz/authz.go:11` (struct type with `(e).Error()` method `:17`)
- **Observation:** Both packages export `ErrCapabilityDenied` under the same name with different shapes. Consumers must qualify: `iamapp.ErrCapabilityDenied` is a sentinel suitable for `errors.Is`; `authz.ErrCapabilityDenied` is a typed error carrying capability/area context. Confused naming has already surfaced in `internal/modules/documents/delivery/http/handler.go:17` which imports the sentinel variant (artifact 03 §2).
- **Evidence:** `_artifacts/01-surface.md` rows 85, 93-95; `_artifacts/03-deps.md` §2.
- **Linked backlog row:** `backlog/iam-refactor.md#R-009`
- **Linked ADR:** missing-ADR
- **Consumer cross-ref:** `wiki/modules/documents.md#81-authentication--authorization` — documents handler imports the sentinel variant; closure here resolves the straddle at `handler.go:17`

### T-010 · `auth` module imports `iam/domain.Role` — bidirectional dependency
- **Severity:** minor
- **Surface:** `internal/modules/auth/application/service.go:18`, `internal/modules/auth/delivery/http/middleware.go:10`, `internal/modules/auth/domain/model.go:6`, `internal/modules/auth/infrastructure/memory/repository.go:10` (all import `iamdomain.Role`); IAM imports `auth/domain` from `internal/modules/iam/delivery/http/admin_handler.go:13`
- **Observation:** IAM depends on auth (for `ManagedUser` types) AND auth depends on IAM (for `Role` enum). The dependency is non-circular today because IAM-→auth lives in `delivery/http/` (admin handler) and auth-→IAM lives in `domain` and lower. If admin_handler types ever migrate up, the cycle becomes hard. Documented for future structural moves.
- **Evidence:** `_artifacts/03-deps.md` §1 + §2.
- **Linked backlog row:** `backlog/iam-refactor.md#R-010`
- **Linked ADR:** missing-ADR

### T-011 · Tenant-scoping rule lacks standalone ADR
- **Severity:** minor
- **Surface:** Multiple — migrations 0130, 0162, 0163; repository code at `role_provider.go:19`, `role_admin_repository.go:20,33,72`
- **Observation:** The convention "every IAM-owned table carries `tenant_id` and every repo method filters by it" was enforced by Group B fix (audit 2026-05-03 B5/B6) but does not have a dedicated ADR. ADR 0007 references migration 0162 in its "Key files" but does not author the tenancy rule.
- **Evidence:** `_artifacts/04-persistence.md` §1 (all tables); `wiki/decisions/0007-two-tier-authz.md` Key files; `wiki/bugs/audit-2026-05-03.md` B5/B6.
- **Linked backlog row:** `backlog/iam-refactor.md#R-011`
- **Linked ADR:** missing-ADR

### T-012 · `RoleCapabilities` map duplicates `role_capabilities` table — CLOSED 2026-05-11 (Plan 4)
- **Severity:** minor (closed)
- **Surface:** `internal/modules/iam/domain/role_capabilities.go:3` (`RoleCapabilitiesVersion = 2`) and `:5` (`var RoleCapabilities map[Role][]Capability`); DB seed in migration 0165 (40 rows in `metaldocs.role_capabilities`); drift check in `application/startup.go:15` `CheckRoleCapabilitiesVersion`
- **Observation:** The role↔capability mapping exists in two places: an in-process Go map (`RoleCapabilities`, using the typed `Capability` namespace from T-001) and the DB `role_capabilities` table (using the string namespace from T-001). The boot-time drift check compares versions, but the data shapes are not directly compared. The in-process map is read by `AuthorizationService` (T-003); the DB table is read by `CapabilityService.CanDo` (`capability_service.go:31`). Two sources of truth.
- **Evidence:** `_artifacts/01-surface.md` rows 152-156; `capability_service.go:31` SQL joins `metaldocs.role_capabilities`.
- **Linked backlog row:** `backlog/iam-refactor.md#R-012`
- **Linked ADR:** missing-ADR

---

## Coverage stats (updated 2026-05-11 post-Plan 4)

- Public symbols undocumented: ~97 / ~107 (Plan 4 deleted authorization.go (~12), startup.go (1), role_capabilities.go (2), area_membership.go (~7) = ~22 symbols removed; 10 documented symbols unchanged). Collective gap addressed by R-013 in backlog.
- Operations missing C4 placement: 0 / 11 (all 11 in §5.3 + Container diagram)
- Cross-deps missing in §5/§8: 0 / 22 (5 OUT + 17 IN named in §8.4 / §3.2)
- State transitions missing in §6: 0 / 2 (grant_membership + upsert_user_role both traced)
- Decisions without ADR link: 7 occurrences — T-005, T-006, T-007, T-008, T-010, T-011 marked `missing-ADR`; T-004 partial-link (ADR 0007 exists but does not address IAM-table coverage). T-001/T-002/T-003/T-009/T-012 closed by Plan 4 (ADR-TODO stubs per Plan 13).
