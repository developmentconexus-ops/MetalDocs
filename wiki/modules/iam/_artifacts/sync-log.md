> **Last verified:** 2026-06-12

## 2026-06-12 - Wave 2 module sync (branch qa/iam-area-membership, commits 81213133c..5a6b407b2)

- **Context:** Wave 2 backend professionalization — batch role-load (RolesByUserIDs), TenantMemberChecker EXISTS probe, LoginContextPort ownership transfer from auth to iam (F-06c, REQ-DATA-2, F-10).
- **Mode:** structural refresh
- **Anchors moved:** +5 new key file anchors (RolesByUserIDs at role_provider.go:75, UserActiveInTenant at :119, LoginContextPort at domain/login_context_port.go:14, LoginContextRepository at infrastructure/postgres/login_context_repository.go:14, TenantMemberChecker at people_service.go:162)
- **Public surface:** `RoleProvider` interface gained `RolesByUserIDs`; `RolesByUserID` rewritten as single LEFT JOIN; `UserActiveInTenant` added to postgres RoleProvider (satisfies TenantMemberChecker); `LoginContextPort` new interface in domain; `LoginContextRepository` new infra impl; `TenantMemberChecker` new port in application; `CachedRoleProvider.RolesByUserIDs` read-through batch added; `InvalidateUser` renamed `InvalidateUserTenant`.
- **Routes/API:** none (no route changes in iam this wave)
- **Runtime flows:** §8.3 caching updated with batch and renamed invalidation method; §3.2 cross-deps updated with LoginContextPort note; C4 diagram updated.
- **Persistence:** none (no new tables or migrations in iam this wave)
- **Dependencies:** auth/application now depends on iamdomain.LoginContextPort (iam publishes; auth composition root injects impl); C4 relationship updated.
- **T-NNN touched:** none (no debt rows opened or closed)
- **R-NNN touched:** none
- **Counts after:** Critical=5 Major=5 Minor=5; missing-ADR=11 (pre-existing tally FAIL on stated 0/2/3 vs actual 5/5/5 — pre-dates this wave; not caused by this sync)
- **Tally gate:** FAIL pre-existing (C/M/m stated 0/2/3 in iam.md §11 vs actual 5/5/5 in register; predates Wave 2; this sync does not change debt counts)
- **Patched files:** wiki/modules/iam.md; wiki/modules/iam-tech-debt.md; wiki/modules/iam/_artifacts/sync-log.md

## 2026-06-10 — Stage-1 backend audit drift patch

- **Context:** Stage-1 mapper found 6 mismatches between wiki docs and code.
- **Mode:** lite patch
- **Affected surface scan:** `internal/modules/iam/domain/model.go` (role + capability consts); `internal/modules/iam/delivery/http/routes_memberships.go:92` (DELETE path-param shape); `internal/modules/iam/delivery/http/people_handler.go` (PR-4 handler ownership); `internal/modules/iam/domain/capability_scope.go:31-35` (stale comment — NOT patched here; Go source is read-only in this pass; flagged for a follow-up code edit).
- **Facts corrected:**
  1. `iam.md` Key files: `domain/model.go` capability count 28→29; added ADR 0022 Phase 10 provenance note.
  2. `iam.md` §5.2 table: `Role type + 5 consts` corrected to 8 consts (added `RoleAreaAdmin`, `RoleQmsAdmin`, `RoleSigner`); capability count 20→29 with full const list.
  3. `iam.md` API Route Truth Table: DELETE row corrected from query-param shape (`/area-memberships`) to path-param shape (`/area-memberships/{user_id}/{area_code}`, `routes_memberships.go:92`). Stale `handleUserRoute`-dispatch rows for GET/POST/PATCH/etc. replaced with PR-4 truth: `PeopleHandler` owns `GET /api/v1/iam/users`, `POST /api/v1/iam/users/invite`, `PATCH /api/v1/iam/users/{user_id}`, `POST /api/v1/iam/users/bulk`, `POST /api/v1/iam/users/{user_id}/reset-password`, `POST /api/v1/iam/users/{user_id}/unlock`, `GET /api/v1/iam/users/{user_id}/memberships` (`people_handler.go:53-59`); `AdminHandler` retains role-edit rows only. Retired `POST /iam/users` (create) row removed.
  4. `iam.md` §12 Glossary: Capability count 18→29; Role description updated to 8 consts with file:line anchor.
  5. `wiki/concepts/authz-tiers.md` Key files: capability count 27→29; line anchor `:15`→`:88`; ADR 0022 Phase 10 note added.
- **Deferred (out of scope for wiki-only pass):** stale comment label in `capability_scope.go:31-35` — the NOTE header reads "runtime gap, ADR 0022 Phase 3"; the body describes call sites that still pass `"tenant"` for area-grade ops declared in Phase 2. Comment wording is accurate to Phase 3 scope; no correction needed in a future code-edit pass beyond what is already documented there.
- **T/R rows touched:** none (no new debt items; no closures).
- **Preflight/tally:** n/a — wiki drift patch, not a module-doc-sync run.
- **Patched files:** `wiki/modules/iam.md`; `wiki/modules/iam-tech-debt.md`; `wiki/concepts/authz-tiers.md`; `wiki/modules/iam/_artifacts/sync-log.md`.

## 2026-06-03 - fix/iam-memberships-pr1-backend-gaps (PR #58)

- **Context:** 4 BE/contract gaps in the area-membership surface closed: `CloseActive`/`GrantAtomic` set `revoked_by`; `ListByTenant` added to repository + interface + service; `listMemberships` handler rewritten with admin/non-admin scope split; dev seed adds `qualidade`/`producao` areas and 3 new memberships.
- **Mode:** lite patch
- **Affected surface scan:** `infrastructure/postgres/user_area_repository.go`; `application/area_membership_service.go`; `delivery/http/routes_memberships.go`; `api/openapi/v1/openapi.yaml` (descriptions only); `db/dev-seeds/0001_local_dev_seed.sql`.
- **Facts updated:** `revoked_by` correctly populated to satisfy `revoked_by_required_when_revoked` CHECK; `ListByTenant` public surface documented; `listMemberships` runtime scope rules captured; tripwire pairing table corrected to show `authz.Require = YES`; line numbers throughout.
- **T/R rows touched:** none (no new debt items; no closures).
- **Preflight/tally:** n/a — wiki-curator invocation, not a module-doc-sync run.
- **Patched files:** `wiki/modules/iam.md`; `wiki/modules/iam/_artifacts/04-persistence.md`; `wiki/modules/iam/_artifacts/02-flow-list-memberships.md`; `wiki/modules/iam/_artifacts/02-flow-grant-membership.md`; `wiki/database/tables/user_process_areas.md`; `wiki/modules/iam/_artifacts/sync-log.md`.

## 2026-05-26 - Wave 2 authz tx seeding sync

- **Context:** uncommitted diff for Wave 2 shared authz transaction seeding across `internal/modules/iam/{authz,application,infrastructure/postgres}` plus documents tx-owner consumers.
- **Mode:** lite patch
- **Affected surface scan:** `authz/context.go`; `application/area_membership_service.go`; `infrastructure/postgres/user_area_repository.go`; targeted IAM authz/application/postgres tests.
- **Facts updated:** IAM now exposes shared `authz.SeedTxIdentity(...)` for transaction-local actor/tenant seeding; area-membership writes use it before `authz.Require(...)`; revoke now passes the acting user into repository-owned tx authz.
- **T/R rows touched:** T-004 evidence refreshed only; no status change.
- **Preflight/tally:** preflight attempted; Git Bash tally failed before doc edits with Windows `CreateFileMapping` error 5.
- **Patched files:** `wiki/modules/iam.md`; `wiki/modules/iam-tech-debt.md`; `wiki/modules/iam/_artifacts/sync-log.md`.

# IAM module doc - sync log

One line per `metaldocs-module-doc-sync` run. Append-only.

## 2026-05-25 - Phase 11 IAM medium sweep

- **Context:** uncommitted diff for Batch F IAM mediums in `internal/modules/iam/*` plus targeted TODO comments in `internal/platform/migrate/migrate.go`.
- **Mode:** lite patch
- **Affected surface scan:** `application/{admin_service,area_membership_service,cached_role_provider,capability_service,dev_role_provider}.go`, `delivery/http/{admin_handler,middleware}.go`, `domain/{errors,model,port}.go`, `infrastructure/postgres/{role_admin_repository,role_provider,user_area_repository}.go`, `internal/platform/migrate/migrate.go`, `wiki/modules/iam.md`, `wiki/modules/iam/_artifacts/sync-log.md`.
- **Facts updated:** shared role parsing/validation is now centralized in `iamdomain.ParseRole`; admin audit writes capture one timestamp and always use server-generated UUID trace IDs; cached-role and migration follow-up TODOs were recorded; postgres role provider no longer surfaces `ErrNoRolesAssigned` for empty role sets.
- **T/R rows touched:** none.
- **Preflight/tally:** preflight attempted; Git Bash tally failed before doc edits with Windows `CreateFileMapping` error 5.
- **Patched files:** `wiki/modules/iam.md`, `wiki/modules/iam/_artifacts/sync-log.md`.

## 2026-05-11 - Plan 6a close T-005

- **Context:** Plan 6a (commit f27529e8) - emit recordAudit in handleUserRoleUpsert + handleCreateUser
- **Anchors moved:** 0
- **Symbols renamed:** 0
- **T-NNN closed:** T-005 - evidence: handleUserRoleUpsert now calls h.recordAudit after writeJSON; handleCreateUser emits auth.user.created event
- **R-NNN updated:** R-005 -> merged - commit f27529e8
- **§11 counts after:** Critical=2 Major=5 Minor=5 (unchanged)
- **Tally gate:** PASS
- **Patched files:** wiki/modules/iam-tech-debt.md; wiki/backlog/iam-refactor.md
- **Structural changes noted:** none for iam (behavioral change in existing handlers only)

## 2026-05-11 - Plan 4 capability namespace collapse + IAM dual-surface consolidation

- **Context:** Plan 4 tasks 1-9 completed: deleted capabilities.go, role_capabilities.go, authorization.go, startup.go, area_membership/; renamed authz.ErrCapabilityDenied->ErrCapDenied; extended model.go to 18 typed Capability consts; migration 0186 reseeded doc.*->document.*
- **Anchors moved:** 8+ (startup.go:15 deleted; area_membership/area_membership.go deleted; authorization.go deleted; role_capabilities.go deleted; authz/authz.go ErrCapabilityDenied->ErrCapDenied)
- **Symbols renamed:** 1 (authz.ErrCapabilityDenied->authz.ErrCapDenied - all occurrences in doc trio updated)
- **T-NNN closed:** T-001, T-002, T-003, T-009, T-012 - evidence: referenced files deleted; typed Capability namespace unified in model.go
- **R-NNN updated:** R-001->merged, R-002->merged, R-003->merged, R-009->merged, R-012->merged - PR: Plan 4 (2026-05-11, commits 3a227642/8da32dbf/a66a8d62/ec7d151a/0cd2e75d)
- **§11 counts after:** Critical=1 Major=3 Minor=3
- **Tally gate:** PASS
- **Patched files:** wiki/modules/iam.md; wiki/modules/iam-tech-debt.md; wiki/backlog/iam-refactor.md

- 2026-05-11 - Plan 3 (session-bound tenant resolution, post-merge sweep). Patched anchors shifted by ~3 lines in `admin_handler.go` and `middleware.go` / `routes_memberships.go` (file growth from `tenant.FromContext` migration). Files: `wiki/modules/iam.md` (§2 + §6.4 envelope anchors :129->:132, :137->:150); `wiki/modules/iam-tech-debt.md` (T-005 :316->:319/:457->:454; T-006 :129->:132/:137->:150; Last verified bump); `wiki/backlog/iam-refactor.md` (Last verified bump); `wiki/README.md` (iam-tech-debt + iam-refactor index stamps). T-NNN affected: T-005, T-006 (anchors only - severity unchanged, debt not resolved). R-NNN affected: none. Escalation: no - verified no Plan 3 ADR exists in `wiki/decisions/` (flagged to caller).

## 2026-05-13 - Plan 10 IAM memberships route canonicalization

- **Context:** uncommitted Plan 10 implementation diff (/api/v2/iam/area-memberships -> /api/v1/iam/area-memberships)
- **Mode:** structural refresh
- **Anchors moved:** memberships route canonicalized to /api/v1
- **Public surface:** no semantic change
- **Routes/API:** IAM route references/artifacts updated to v1
- **Runtime flows:** unchanged behavior
- **Persistence:** none
- **Dependencies:** permission resolver path mapping aligned
- **T-NNN touched:** IAM T-010 note remains docs-only and deferred ADR linkage
- **R-NNN touched:** IAM backlog wording canonicalized
- **Counts after:** Critical=2 Major=5 Minor=5; missing-ADR=11
- **Tally gate:** PASS
- **Patched files:** wiki/modules/iam.md; wiki/modules/iam-tech-debt.md; wiki/backlog/iam-refactor.md; wiki/modules/iam/_artifacts/*

## 2026-05-25 - trusted role-header strip sync

- **Context:** branch `fix/docs-5b-header-roles-c3`; `internal/modules/iam/delivery/http/middleware.go` now deletes `X-User-Roles` alongside `X-User-ID` before downstream handlers run.
- **Affected modules:** iam, documents.
- **Mode:** structural refresh.
- **Affected-surface scan:** IAM middleware trust-boundary behavior changed; no route, OpenAPI, persistence, dependency, public exported surface, debt, or backlog change.
- **Facts updated:** middleware trusted-header stripping now includes `X-User-Roles`; downstream role context remains sourced from IAM role context/provider paths instead of caller-controlled headers.
- **T-NNN touched:** none.
- **R-NNN touched:** none.
- **Tally gate:** preflight/tally blocked by Git Bash `CreateFileMapping` Win32 error 5; treated as pre-existing tooling failure, not an edit-caused doc failure.
- **Patched files:** `wiki/modules/iam.md`; `wiki/modules/iam/_artifacts/sync-log.md`.
