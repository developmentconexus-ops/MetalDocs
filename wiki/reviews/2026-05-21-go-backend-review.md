# Go Backend Review — 2026-05-21

**Initiative:** Multi-session ECC review of the MetalDocs Go backend.
**Started:** 2026-05-21
**Mode:** Append-only. One row per module. Findings live in `2026-05-21-go-backend-review/<module>.md`.
**Cursor:** see `MEMORY.md` → `project_go_backend_review` for next-up module.

## Scope

- In: all Go code under `apps/api/` and `internal/` (hand-written).
- Out: generated code (`*/api/api.gen.go`, `frontend/apps/web/src/lib/api-types/`), vendored editor (`frontend/apps/web/eigenpal/`), wiki/spec docs unless code contradicts them.
- Tests reviewed alongside owning module.

## Severity Convention

| Level    | Meaning                          | Action                                |
|----------|----------------------------------|---------------------------------------|
| Critical | Security flaw or data-loss risk  | Open issue + spawn fix task at once   |
| High     | Bug or correctness gap           | Track in module backlog               |
| Medium   | Maintainability / design smell   | Note for opportunistic refactor       |
| Low      | Style / minor                    | Note only                             |

## Per-Session Loop

1. Read `MEMORY.md` → `project_go_backend_review` → confirm cursor.
2. Pick next `Pending` module in table below.
3. Run `metaldocs-module-doc` skill to confirm module surface.
4. Spawn ECC agents in parallel (single message, multiple Agent blocks):
   - `ecc:go-reviewer` (idioms, errors, concurrency)
   - `ecc:security-reviewer` (OWASP, auth boundary)
   - `ecc:database-reviewer` (only if module touches Postgres)
   - `ecc:silent-failure-hunter` (swallowed errors)
   - `ecc:type-design-analyzer` (invariants, encapsulation)
5. Consolidate digests → append to `2026-05-21-go-backend-review/<module>.md`.
6. Update tracker row: status, severity counts, date, link.
7. Update cursor in `MEMORY.md`.
8. Commit: `docs(review): <module> findings`.

## Status Legend

- `Pending` — not started
- `In Progress` — agents dispatched, findings being consolidated
- `Done` — findings committed, tracker row updated
- `Skipped` — explicit decision, reason in findings file

## Module Tracker

| #  | Module                                          | Status  | Critical | High | Medium | Low | Reviewer | Date | Findings |
|----|-------------------------------------------------|---------|----------|------|--------|-----|----------|------|----------|
| 1  | `apps/api/cmd/metaldocs-api`                    | Done    | 4        | 5    | 6      | 6   | go-reviewer, security-reviewer, silent-failure-hunter, type-design-analyzer | 2026-05-21 | [cmd-metaldocs-api.md](2026-05-21-go-backend-review/cmd-metaldocs-api.md) |
| 2a | `platform/{authn,security,idempotency,ratelimit,tenant,problem,httpresponse}` | Done | 5 | 11 | 11 | 8 | go-reviewer, security-reviewer, silent-failure-hunter, type-design-analyzer, database-reviewer | 2026-05-21 | [platform-2a-security.md](2026-05-21-go-backend-review/platform-2a-security.md) |
| 2b | `platform/{db,migrate,bootstrap,objectstore,storage,messaging,servicebus,jobs,worker}` | Done | 10 | 24 | 16 | 8 | go-reviewer, security-reviewer, silent-failure-hunter, type-design-analyzer, database-reviewer | 2026-05-22 | [platform-2b-data-infra.md](2026-05-21-go-backend-review/platform-2b-data-infra.md) |
| 2c | `platform/{config,observability,cache,featureflags,formval,httpclient,pagination,docgenv2,render}` | Done | 1 | 14 | 25 | 18 | go-reviewer, security-reviewer, silent-failure-hunter, type-design-analyzer, database-reviewer | 2026-05-22 | [platform-2c-support-observability.md](2026-05-21-go-backend-review/platform-2c-support-observability.md) |
| 3  | `internal/modules/auth`                         | Done    | 8        | 18   | 22     | 11  | go-reviewer, security-reviewer, silent-failure-hunter, type-design-analyzer, database-reviewer | 2026-05-22 | [auth-3.md](2026-05-21-go-backend-review/auth-3.md) |
| 4  | `internal/modules/iam`                          | Done    | 8        | 15   | 18     | 11  | go-reviewer, security-reviewer, silent-failure-hunter, type-design-analyzer, database-reviewer | 2026-05-22 | [iam-4.md](2026-05-21-go-backend-review/iam-4.md) |
| 5a | `documents/{domain,application,repository}` (~9K LoC) | Done    | 9  | 12   | 18     | 10  | go-reviewer, security-reviewer, silent-failure-hunter, type-design-analyzer, database-reviewer | 2026-05-22 | [documents-5a.md](2026-05-21-go-backend-review/documents-5a.md) |
| 5b | `documents/{delivery,http,jobs}` (~4K LoC)      | Done    | 6        | 19   | 22     | 12  | go-reviewer, security-reviewer, silent-failure-hunter, type-design-analyzer, database-reviewer | 2026-05-22 | [documents-5b.md](2026-05-21-go-backend-review/documents-5b.md) |
| 5c | `documents/approval/{domain,application}` (~13K LoC) | Done    | 12 | 20   | 21     | 13  | go-reviewer, security-reviewer, silent-failure-hunter, type-design-analyzer, database-reviewer | 2026-05-22 | [documents-5c.md](2026-05-21-go-backend-review/documents-5c.md) |
| 5d | `documents/approval/{http,repository,infrastructure,jobs}` (~7K LoC) | Done | 11 | 20 | 15 | 11 | go-reviewer, security-reviewer, silent-failure-hunter, type-design-analyzer, database-reviewer | 2026-05-22 | [documents-5d.md](2026-05-21-go-backend-review/documents-5d.md) |
| 6  | `internal/modules/controlleddocuments`          | Done    | 9        | 13   | 16     | 9   | go-reviewer, security-reviewer, silent-failure-hunter, type-design-analyzer, database-reviewer | 2026-05-22 | [controlleddocuments-6.md](2026-05-21-go-backend-review/controlleddocuments-6.md) |
| 7  | `internal/modules/taxonomy`                     | Done    | 5        | 14   | 13     | 8   | go-reviewer, security-reviewer, silent-failure-hunter, type-design-analyzer, database-reviewer | 2026-05-22 | [taxonomy-7.md](2026-05-21-go-backend-review/taxonomy-7.md) |
| 8  | `internal/modules/templates`                    | Pending | -        | -    | -      | -   | -        | -    | -        |
| 9  | `internal/modules/audit`                        | Pending | -        | -    | -      | -   | -        | -    | -        |
| 10 | `internal/modules/render`                       | Pending | -        | -    | -      | -   | -        | -    | -        |
| 11 | `internal/modules/search`                       | Pending | -        | -    | -      | -   | -        | -    | -        |
| 12 | `internal/modules/jobs`                         | Pending | -        | -    | -      | -   | -        | -    | -        |
| 13 | Shared infra (`internal/test`, `objectstore`, `docgenv2`, fanout) | Pending | - | - | - | - | - | - | - |

## Notes

- Module order revised from initial plan: `approval` module does not exist in repo. Added `render`, `search`, `jobs` (real modules under `internal/modules/`).
- Platform row split into #2a/#2b/#2c on 2026-05-21 — 25 packages / ~5200 LoC too big for a single agent pass. Grouped by concern: security boundary (#2a), data + infra (#2b), support + observability (#2c).
- Documents row split into #5a/#5b/#5c/#5d on 2026-05-22 — ~33K LoC hand-written (largest module). #5a = core domain+application+repository; #5b = delivery+http+jobs; #5c = approval business logic; #5d = approval delivery+infra.

## Critical Backlog (G3 handoff)

Per plan §6 G3: each Critical needs owner + ETA + reserved fix-branch before cursor advances. ETA = TBC means owner to set on branch-cut. Land in stated order — `fix/migrate-2b-c6-c7` first (silent data bomb, standalone, smallest).

### Module #2a (carried open)

| ID | File:line | Severity | Owner | ETA | Fix branch | Status |
|----|-----------|----------|-------|-----|-----------|--------|
| 2a-C3 | idempotency two-phase (in progress on branch) | Critical | leandrotca | TBC | `fix/h11-idempotency-schema-v2` | WIP |
| 2a-C4 | idempotency two-phase (in progress on branch) | Critical | leandrotca | TBC | `fix/h11-idempotency-schema-v2` | WIP |

### Module #2b (10 Criticals, 4 fix branches reserved 2026-05-22)

| ID | File:line | Severity | Owner | ETA | Fix branch | Status |
|----|-----------|----------|-------|-----|-----------|--------|
| 2b-C6 | `migrations/0042_*` duplicate prefix | Critical | leandrotca | TBC | `fix/migrate-2b-c6-c7` | Backlog (land first) |
| 2b-C7 | `migrations/0130_*` duplicate prefix | Critical | leandrotca | TBC | `fix/migrate-2b-c6-c7` | Backlog (land first) |
| 2b-C1 | `internal/platform/migrate/migrate.go:74-78` swallow | Critical | leandrotca | TBC | `fix/migrate-2b-c1-c5-c8-c9` | Backlog |
| 2b-C5 | `internal/platform/migrate/migrate.go:24-69` no advisory lock | Critical | leandrotca | TBC | `fix/migrate-2b-c1-c5-c8-c9` | Backlog |
| 2b-C8 | `migrations/0176_pdf_dispatch_outbox.sql:1-24` no tx | Critical | leandrotca | TBC | `fix/migrate-2b-c1-c5-c8-c9` | Backlog |
| 2b-C9 | `migrations/0111_docx_v2_exports.sql:1,6` unqualified FK + no tx | Critical | leandrotca | TBC | `fix/migrate-2b-c1-c5-c8-c9` | Backlog |
| 2b-C2 | `internal/platform/storage/local/store.go:20,31,40` path containment | Critical | leandrotca | TBC | `fix/storage-2b-c2` | Backlog |
| 2b-C3 | `internal/platform/config/docgen_v2.go:20-32` SSRF | Critical | leandrotca | TBC | `fix/docgen-2b-c3-c4` | Backlog |
| 2b-C4 | `internal/platform/config/docgen_v2.go:24` empty token | Critical | leandrotca | TBC | `fix/docgen-2b-c3-c4` | Backlog |
| 2b-C10 | `internal/platform/messaging/events.go:7-18` typed boundary | Critical | leandrotca | TBC | `fix/messaging-2b-c10` | Backlog (cascades H14, H15, H19) |

### Module #2c (1 Critical, 1 fix branch reserved 2026-05-22)

| ID | File:line | Severity | Owner | ETA | Fix branch | Status |
|----|-----------|----------|-------|-----|-----------|--------|
| 2c-C1 | `internal/platform/pagination/cursor.go:37-43` HMAC-less cursor, anchor tamperable | Critical | leandrotca | TBC | `fix/cursor-2c-c1` | Backlog (cascades H7, H8, H9, M26, M27, L15, L16 — coordinated cursor rewrite) |

### Module #3 (8 Criticals, 4 fix branches reserved 2026-05-22)

| ID | File:line | Severity | Owner | ETA | Fix branch | Status |
|----|-----------|----------|-------|-----|-----------|--------|
| 3-C2 | `internal/modules/auth/delivery/http/middleware.go:59` `X-User-Id` legacy header → full auth bypass | Critical | leandrotca | TBC | `fix/auth-3-c2` | Backlog (land first) |
| 3-C1 | `internal/modules/auth/application/service.go:147` swallowed `RecordFailedLogin` → unbounded brute-force | Critical | leandrotca | TBC | `fix/auth-3-c1-c3-c6` | Backlog |
| 3-C3 | `internal/modules/auth/infrastructure/postgres/repository.go:163` + `service.go:140-148` lockout TOCTOU | Critical | leandrotca | TBC | `fix/auth-3-c1-c3-c6` | Backlog |
| 3-C6 | `application/service.go:534-538` + `Config.SessionSecret` empty-secret HMAC | Critical | leandrotca | TBC | `fix/auth-3-c1-c3-c6` | Backlog |
| 3-C4 | `infrastructure/postgres/repository.go:120-131` `RevokeSession`/`TouchSession` no `RowsAffected` | Critical | leandrotca | TBC | `fix/auth-3-c4-c5-c8` | Backlog |
| 3-C5 | `auth_sessions` schema missing `tenant_id` (verify migration before fix) | Critical | leandrotca | TBC | `fix/auth-3-c4-c5-c8` | Backlog (verify) |
| 3-C8 | `infrastructure/postgres/repository.go:77,380,396,411` `err == sql.ErrNoRows` direct equality | Critical | leandrotca | TBC | `fix/auth-3-c4-c5-c8` | Backlog |
| 3-C7 | `domain/model.go:14` `PasswordHash string` vs plaintext indistinguishable | Critical | leandrotca | TBC | `fix/auth-3-c7-types` | Backlog (cascades H12, H13, H15, M15, M16) |

### Module #4 (8 Criticals, 6 fix branches reserved 2026-05-22)

| ID | File:line | Severity | Owner | ETA | Fix branch | Status |
|----|-----------|----------|-------|-----|------------|--------|
| 4-C6 | `infrastructure/postgres/role_provider.go:20-28` checkUserSQL no tenant_id → cross-tenant auth bypass | Critical | leandrotca | TBC | `fix/iam-4-role-provider-c6-h5` | Backlog (land first) |
| 4-C1 | `delivery/http/middleware.go:68,81-82` legacy X-User-Id delete+re-read on same canonical header | Critical | leandrotca | TBC | `fix/iam-4-middleware-c1-c3-c4-c7-c8` | Backlog (land second) |
| 4-C3 | `delivery/http/middleware.go:70-73` nil resolver fail-open → all routes pass unauthenticated | Critical | leandrotca | TBC | `fix/iam-4-middleware-c1-c3-c4-c7-c8` | Backlog |
| 4-C4 | `delivery/http/middleware.go:74-78` VisibilitySessionRequired no session check | Critical | leandrotca | TBC | `fix/iam-4-middleware-c1-c3-c4-c7-c8` | Backlog |
| 4-C7 | `delivery/http/middleware.go:91-95` tenant_id fallback to client header then DevTenantID | Critical | leandrotca | TBC | `fix/iam-4-middleware-c1-c3-c4-c7-c8` | Backlog |
| 4-C8 | `delivery/http/middleware.go:97-99` nil caps silently skips capability check | Critical | leandrotca | TBC | `fix/iam-4-middleware-c1-c3-c4-c7-c8` | Backlog |
| 4-C2 | `infrastructure/postgres/role_admin_repository.go:113-121` ReplaceUserRoles last-role-wins silent privilege escalation | Critical | leandrotca | TBC | `fix/iam-4-replace-roles-c2` | Backlog (land third) |
| 4-C5 | `infrastructure/postgres/user_area_repository.go:103` CloseActive no RowsAffected → silent revoke no-op | Critical | leandrotca | TBC | `fix/iam-4-area-repo-c5-h4` | Backlog (land fourth) |

### Module #5d (11 Criticals, 6 fix branches reserved 2026-05-22)

| ID | File:line | Severity | Owner | ETA | Fix branch | Status |
|----|-----------|----------|-------|-----|------------|--------|
| 5d-C1 | `http/route_admin_handler.go:165` no authz + raw SQL in ListRoutesHandler | Critical | leandrotca | TBC | `fix/approval-5d-list-routes-c1` | Backlog (land first) |
| 5d-C3 | `http/doc_approval_handler.go:97` idempotency replay before eligibility check | Critical | leandrotca | TBC | `fix/approval-5d-replay-eligibility-c3` | Backlog (land second) |
| 5d-C8 | `repository/postgres_approval_repository.go:504` loadSignoffsForInstance no tenant predicate | Critical | leandrotca | TBC | `fix/approval-5d-tenant-isolation-c8-c9` | Backlog |
| 5d-C9 | `repository/postgres_approval_repository.go:421` loadStageInstances no tenant predicate | Critical | leandrotca | TBC | `fix/approval-5d-tenant-isolation-c8-c9` | Backlog |
| 5d-C7 | `http/publish_handler.go:47` + `signoff_handler.go:35` parseIfMatch discarded | Critical | leandrotca | TBC | `fix/approval-5d-occ-c7` | Backlog |
| 5d-C11 | `http/contracts/signoff.go:6` Decision bare string, Validate skippable | Critical | leandrotca | TBC | `fix/approval-5d-occ-c7` | Backlog |
| 5d-C2 | `repository/postgres_approval_repository.go:441` skipReason never scanned | Critical | leandrotca | TBC | `fix/approval-5d-repo-c2-c10` | Backlog |
| 5d-C4 | `http/supersede_handler.go:53` raw SQL + TOCTOU in SupersedeHandler | Critical | leandrotca | TBC | `fix/approval-5d-repo-c2-c10` | Backlog |
| 5d-C10 | `repository/postgres_approval_repository.go:134` InsertSignoff wrong ON CONFLICT target | Critical | leandrotca | TBC | `fix/approval-5d-repo-c2-c10` | Backlog |
| 5d-C5 | `http/errors.go:214` WriteJSON fallback marshal error discarded | Critical | leandrotca | TBC | `fix/approval-5d-writejson-c5-c6` | Backlog |
| 5d-C6 | `http/publish_handler.go:50` readSvc nil-check missing → panic | Critical | leandrotca | TBC | `fix/approval-5d-writejson-c5-c6` | Backlog |

### Module #5c (12 Criticals, 6 fix branches reserved 2026-05-22)

| ID | File:line | Severity | Owner | ETA | Fix branch | Status |
|----|-----------|----------|-------|-----|------------|--------|
| 5c-C1 | `application/idempotency.go:38` canonicalize error discarded → idempotency collapse | Critical | leandrotca | TBC | `fix/approval-5c-idempotency-c1` | Backlog (land first) |
| 5c-C6 | `application/cancel_service.go:87` BypassAuthz public flag → authz bypass | Critical | leandrotca | TBC | `fix/approval-5c-authz-bypass-c6-c7` | Backlog (land second) |
| 5c-C7 | `application/read_service.go:38` LoadInstance no authz check → IDOR | Critical | leandrotca | TBC | `fix/approval-5c-authz-bypass-c6-c7` | Backlog |
| 5c-C8 | `application/decision_service.go:96` hash from caller-supplied data → integrity violation | Critical | leandrotca | TBC | `fix/approval-5c-hash-integrity-c8` | Backlog (land third) |
| 5c-C2 | `application/decision_service.go:163` governance event rolled back → audit trail lost | Critical | leandrotca | TBC | `fix/approval-5c-audit-trail-c2-c4` | Backlog |
| 5c-C3 | `application/decision_service.go:420` PDF dispatch error silently discarded | Critical | leandrotca | TBC | `fix/approval-5c-audit-trail-c2-c4` | Backlog |
| 5c-C4 | `application/route_admin_service.go:112,202,263` json.Marshal discarded in 3 event payloads | Critical | leandrotca | TBC | `fix/approval-5c-audit-trail-c2-c4` | Backlog |
| 5c-C9 | `application/obsolete_service.go:110` approval_instances UPDATE missing tenant_id | Critical | leandrotca | TBC | `fix/approval-5c-tenant-isolation-c9-c10-c11` | Backlog |
| 5c-C10 | `application/decision_service.go:309` approve-path UPDATE no RowsAffected → phantom approval | Critical | leandrotca | TBC | `fix/approval-5c-tenant-isolation-c9-c10-c11` | Backlog |
| 5c-C11 | `application/decision_service.go:362` reject-path UPDATE no RowsAffected → phantom rejection | Critical | leandrotca | TBC | `fix/approval-5c-tenant-isolation-c9-c10-c11` | Backlog |
| 5c-C12 | `application/read_service.go:163,231` inbox/count bypass transaction+GUC → RLS violation | Critical | leandrotca | TBC | `fix/approval-5c-rls-bypass-c12` | Backlog |
| 5c-C5 | `application/services.go:22` e2etest imported in production RealClock | Critical | leandrotca | TBC | `fix/approval-5c-production-clock-c5` | Backlog |

### Module #5b (6 Criticals, 4 fix branches reserved 2026-05-22)

| ID | File:line | Severity | Owner | ETA | Fix branch | Status |
|----|-----------|----------|-------|-----|------------|--------|
| 5b-C3 | `delivery/http/handler.go:1165,1187` X-User-Roles header trusted → privilege escalation | Critical | leandrotca | TBC | `fix/docs-5b-header-roles-c3` | Backlog (land first) |
| 5b-C4 | `http/pdf_webhook_handler.go:83` tenant_id from body → cross-tenant PDF overwrite | Critical | leandrotca | TBC | `fix/docs-5b-webhook-tenant-c4` | Backlog (land second) |
| 5b-C1 | `delivery/http/handler.go:559-564` content-hash DB error silently discarded | Critical | leandrotca | TBC | `fix/docs-5b-finalize-c1-c2` | Backlog |
| 5b-C2 | `delivery/http/handler.go:618-622` err.Error() in JSON response → info leak + JSON injection | Critical | leandrotca | TBC | `fix/docs-5b-finalize-c1-c2` | Backlog |
| 5b-C5 | `repository/repository.go:1035` unbounded DELETE in DeleteExpiredPending | Critical | leandrotca | TBC | `fix/docs-5b-sweeper-unbounded-c5-c6` | Backlog |
| 5b-C6 | `repository/repository.go:667` unbounded UPDATE in ExpireStaleSessions | Critical | leandrotca | TBC | `fix/docs-5b-sweeper-unbounded-c5-c6` | Backlog |

### Module #7 (5 Criticals, 5 fix branches reserved 2026-05-22)

| ID | File:line | Severity | Owner | ETA | Fix branch | Status |
|----|-----------|----------|-------|-----|------------|--------|
| 7-C1 | `delivery/http/routes_*.go` — 6 read handlers (list/get areas, profiles, families) no authz.Require gate | Critical | leandrotca | TBC | `fix/taxonomy-7-authz-reads-c1` | Backlog (land first) |
| 7-C2 | `infrastructure/repository.go` + `family_repository.go` — all repo read methods missing setAuthzGUC → RLS unset for SELECTs | Critical | leandrotca | TBC | `fix/taxonomy-7-authz-reads-c1` | Backlog (same branch as C1) |
| 7-C3 | `infrastructure/template_version_checker.go:13` — no tenant_id predicate → cross-tenant IDOR on template version lookup | Critical | leandrotca | TBC | `fix/taxonomy-7-tenant-isolation-c3` | Backlog (land second) |
| 7-C4 | `infrastructure/family_repository.go:48` — FamilyUpdate lost-update race: GetByCode outside transaction, no FOR UPDATE | Critical | leandrotca | TBC | `fix/taxonomy-7-toctou-c4-c5` | Backlog (land third) |
| 7-C5 | `application/area_service.go:67` + `profile_service.go:77,115` — SetParent/SetDefaultTemplate/Archive read-then-write races, no FOR UPDATE | Critical | leandrotca | TBC | `fix/taxonomy-7-toctou-c4-c5` | Backlog |

### Module #6 (9 Criticals, 6 fix branches reserved 2026-05-22)

| ID | File:line | Severity | Owner | ETA | Fix branch | Status |
|----|-----------|----------|-------|-----|------------|--------|
| 6-C1 | `delivery/http/routes.go:245` GetActiveDocument no visibility/authz gate → IDOR | Critical | leandrotca | TBC | `fix/cddocs-6-authz-idor-c1-c2-c3` | Backlog (land first) |
| 6-C2 | `application/service.go:405-449` changeStatus GetByID without CanRead → restricted docs unprotected | Critical | leandrotca | TBC | `fix/cddocs-6-authz-idor-c1-c2-c3` | Backlog |
| 6-C3 | `application/service.go:357-364` PreviewCode/PeekSeq no authz → sequence counters enumerable | Critical | leandrotca | TBC | `fix/cddocs-6-authz-idor-c1-c2-c3` | Backlog |
| 6-C4 | `application/service.go:420` changeStatus missing setAuthzGUC → RLS context unset | Critical | leandrotca | TBC | `fix/cddocs-6-authz-guc-c4` | Backlog (land second) |
| 6-C5 | `application/service.go:405-429` TOCTOU race in changeStatus: GetByID outside transaction | Critical | leandrotca | TBC | `fix/cddocs-6-toctou-c5` | Backlog (land third) |
| 6-C6 | `application/service.go:165,253,434` json.Marshal errors discarded → nil governance event payloads | Critical | leandrotca | TBC | `fix/cddocs-6-audit-trail-c6` | Backlog |
| 6-C7 | `infrastructure/repository.go:549` GetTemplateVersionState no tenant_id → cross-tenant leakage | Critical | leandrotca | TBC | `fix/cddocs-6-tenant-isolation-c7` | Backlog (may need migration) |
| 6-C8 | `application/migration.go:30-35` unbounded backfill SELECT → full-table scan, autovacuum contention | Critical | leandrotca | TBC | `fix/cddocs-6-migration-c8-c9` | Backlog |
| 6-C9 | `application/migration.go:61-68` ON CONFLICT re-link + skipped counter inverted → silent re-link | Critical | leandrotca | TBC | `fix/cddocs-6-migration-c8-c9` | Backlog |

### Module #5a (9 Criticals, 3 fix branches reserved 2026-05-22)

| ID | File:line | Severity | Owner | ETA | Fix branch | Status |
|----|-----------|----------|-------|-----|------------|--------|
| 5a-C1 | `repository/repository.go:1381,1405` MarkArchived/Unarchive no RowsAffected → silent success on cross-tenant docID | Critical | leandrotca | TBC | `fix/docs-5a-rows-affected-c1-c2` | Backlog (land first) |
| 5a-C2 | `repository/snapshot_repository.go:46,103,113,148,159` 5 write methods no RowsAffected → silent freeze/docx/pdf loss | Critical | leandrotca | TBC | `fix/docs-5a-rows-affected-c1-c2` | Backlog |
| 5a-C9 | `domain/values_hash.go:18` json.Marshal discarded → hash incorrect → freeze idempotency broken | Critical | leandrotca | TBC | `fix/docs-5a-values-hash-c9` | Backlog (land second) |
| 5a-C3 | `repository/repository.go:1054,1083` CreateCheckpoint/ListCheckpoints no tenant_id → IDOR | Critical | leandrotca | TBC | `fix/docs-5a-tenant-isolation-c3-c4-c5-c6-c7-c8` | Backlog (land third; requires migration C7 first) |
| 5a-C4 | `repository/repository.go:1176` RestoreCheckpoint no tenant_id → cross-tenant restore | Critical | leandrotca | TBC | `fix/docs-5a-tenant-isolation-c3-c4-c5-c6-c7-c8` | Backlog |
| 5a-C5 | `repository/repository.go:777` GetPendingForCommit bare pendingID no tenant scope | Critical | leandrotca | TBC | `fix/docs-5a-tenant-isolation-c3-c4-c5-c6-c7-c8` | Backlog |
| 5a-C6 | `repository/repository.go:592` HeartbeatSession no tenant_id → cross-tenant session keepalive | Critical | leandrotca | TBC | `fix/docs-5a-tenant-isolation-c3-c4-c5-c6-c7-c8` | Backlog |
| 5a-C7 | `editor_sessions` schema missing tenant_id column (migration required) | Critical | leandrotca | TBC | `fix/docs-5a-tenant-isolation-c3-c4-c5-c6-c7-c8` | Backlog |
| 5a-C8 | `repository/repository.go:1147` GetRevision no tenant_id scope | Critical | leandrotca | TBC | `fix/docs-5a-tenant-isolation-c3-c4-c5-c6-c7-c8` | Backlog |
