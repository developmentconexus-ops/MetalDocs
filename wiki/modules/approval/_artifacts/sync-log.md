# Sync log - approval

## 2026-05-25 - approval 5c + 5d medium sweep sync

- **Context:** current uncommitted Phase 11 medium-sweep diff scoped to `internal/modules/documents/approval/**`.
- **Mode:** structural refresh
- **Anchors moved:** none.
- **Public surface:** `ReadService.LoadInstance` no longer accepts `actorID`; `application.SignoffRequest.Decision` is now typed as `domain.Decision`; approval contract DTOs now use typed quorum/drift/status/decision/signature aliases and `RouteResponse.NewVersion` is a pointer field.
- **Routes/API:** no route or OpenAPI ownership change; HTTP contract validation is stricter and now wraps a typed `contracts.ErrValidation` sentinel.
- **Runtime flows:** signoff seeds `authz.WithCapCache` before authz GUC setup; publish/cancel set authz context before repository reads; scheduler logs all three nil-return skip reasons; doc-signoff handlers parse boundary decisions into `domain.Decision`.
- **Persistence:** governance events now persist `occurred_at`; active-instance lookup orders by `submitted_at DESC, id DESC`; repository scan failures now include approval-instance context; cancel area snapshot scan now tolerates SQL NULL with `sql.NullString`.
- **Dependencies:** none beyond existing approval/http/authz/repository surfaces.
- **T-NNN touched:** none.
- **R-NNN touched:** none.
- **Counts after:** unchanged from prior row (Critical=2 Major=4 Minor=6; missing-ADR=10).
- **Tally gate:** FAIL pre-existing/environment: module-doc-sync preflight Git Bash step still aborts with `CreateFileMapping ... Win32 error 5`.
- **Patched files:** `wiki/modules/approval.md`; `wiki/modules/approval-tech-debt.md`; `wiki/backlog/approval-refactor.md`; `wiki/modules/approval/_artifacts/sync-log.md`.

## 2026-05-25 - residual Phase 7 highs sync (5c-H12 / 5d-H5 / 5d-H6)

- **Context:** current uncommitted residual-highs hardening diff scoped to `internal/modules/documents/approval/**`.
- **Mode:** lite patch
- **Anchors moved:** none.
- **Public surface:** `jobs.ScheduledPublishWorker` dependencies are now constructor-injected/private (`NewScheduledPublishWorker`), preserving `Work` nil guards.
- **Routes/API:** none.
- **Runtime flows:** unchanged business behavior; worker still delegates to `SchedulerService.RunScheduledPublishJob`.
- **Persistence:** none.
- **Dependencies:** `infrastructure/signature.PasswordReauthProvider` now requires an explicit `AuthFailureRateLimiter` dependency and fails closed when unconfigured; in-memory limiter kept as explicit local implementation (`NewInMemoryAuthFailureRateLimiter`) for test/dev usage.
- **Residual IDs touched:** `5c-H12` documented as intentional (read-service keeps read-write tx because repository reads may lock with `FOR UPDATE`); `5d-H5` mitigated with explicit fail-closed limiter dependency; `5d-H6` closed with private fields + constructor.
- **T-NNN touched:** none.
- **R-NNN touched:** none.
- **Counts after:** unchanged from prior row (Critical=2 Major=4 Minor=6; missing-ADR=10).
- **Tally gate:** FAIL pre-existing/environment: Git Bash unavailable in this environment (`CreateFileMapping ... Win32 error 5`) during module-doc-sync preflight.
- **Patched files:** `internal/modules/documents/approval/application/read_service.go`; `internal/modules/documents/approval/infrastructure/signature/password_reauth.go`; `internal/modules/documents/approval/infrastructure/signature/password_reauth_test.go`; `internal/modules/documents/approval/jobs/scheduled_publish_job.go`; `internal/modules/documents/approval/jobs/scheduled_publish_job_test.go`; `wiki/modules/approval/_artifacts/sync-log.md`.

## 2026-05-25 - approval 5c high hardening sync

- **Context:** current uncommitted Worker 7F approval application/repository diff.
- **Mode:** structural refresh
- **Anchors moved:** none.
- **Public surface:** no exported symbols added or removed; `RecordSignoff` now requires non-empty `StageInstanceID`.
- **Routes/API:** no HTTP route, OpenAPI, or generated API change.
- **Runtime flows:** signoff now applies live eligibility drift for non-`keep_snapshot` policies; cancel and cutover now use stronger transaction/authz context.
- **Persistence:** tenant-scoped cancel stage UPDATE; tenant-scoped route-stage/signoff/display-name reads; checked `RowsAffected`/`json.Marshal` error paths; scheduler rollback now owned by transaction defer.
- **Dependencies:** `cutover_service.go` now imports `iam/authz` for system-bypassed RLS context.
- **T-NNN touched:** T-006 closed with cancel/cutover authz/RLS evidence.
- **R-NNN touched:** R-006 closed.
- **Counts after:** Critical=2 Major=4 (T-004/T-005 open; T-003/T-006 closed) Minor=6; missing-ADR=10.
- **Tally gate:** FAIL pre-existing/environment: Git Bash in preflight failed with Win32 error 5 before wiki edits.
- **Patched files:** `wiki/modules/approval.md`; `wiki/modules/approval-tech-debt.md`; `wiki/backlog/approval-refactor.md`; `wiki/modules/approval/_artifacts/sync-log.md`.

## 2026-05-21 - approval wrapper-mounted routing sync

- **Context:** uncommitted Phase 3 backend-platform-freeze slice to move approval runtime mounting from raw per-route handler registration to generated `approvalapi.ServerInterfaceWrapper`.
- **Mode:** lite patch
- **Anchors moved:** none
- **Public surface:** unchanged route set (16 operations); mount style changed.
- **Routes/API:** `internal/modules/documents/approval/http/router.go` now mounts generated wrapper methods for every approval/doc-action route.
- **Runtime flows:** unchanged handler/business behavior; existing adapter methods in `routes_generated.go` continue delegating to concrete handlers.
- **Persistence:** none.
- **Dependencies:** none.
- **T-NNN touched:** none.
- **R-NNN touched:** none.
- **Counts after:** Critical=2 Major=4 Minor=6; missing-ADR=10 (pre-existing).
- **Tally gate:** pending.
- **Patched files:** `internal/modules/documents/approval/http/router.go`; `wiki/modules/approval.md`; `wiki/modules/approval/_artifacts/sync-log.md`.

## 2026-05-20 - deep QA wiki relocation sync

- **Context:** relocate the new documents+approval deep-QA companion artifacts from `docs/superpowers/` into a stable wiki reference folder with cleaner names.
- **Mode:** lite patch
- **Anchors moved:** deep-QA reference paths now point to `wiki/references/documents-approval-deep-qa/`
- **Public surface:** none
- **Routes/API:** none
- **Runtime flows:** none
- **Persistence:** none
- **Dependencies:** approval module wiki now points to the wiki-owned deep-QA reference set (`runbook.md`, `fixtures.md`, `matrix.md`)
- **T-NNN touched:** none
- **R-NNN touched:** none
- **Counts after:** Critical=2 Major=4 Minor=6; missing-ADR=10
- **Tally gate:** pending
- **Patched files:** `wiki/modules/approval.md`; `wiki/modules/approval/_artifacts/sync-log.md`; `wiki/references/local-dev-startup.md`; `wiki/references/README.md`; `wiki/references/documents-approval-deep-qa/*`

## 2026-05-20 - scheduled publish River cutover sync

- **Context:** current uncommitted scheduled publish River migration and dedicated `metaldocs-jobs` runtime cutover.
- **Mode:** structural refresh
- **Anchors moved:** scheduled publish runtime ownership moved from API cron lane to `metaldocs-jobs`
- **Public surface:** documented `RunScheduledPublishJob`, `ScheduledPublishJobInput`, `ScheduledPublishEnqueuer`, and the River worker/enqueuer package
- **Routes/API:** no approval HTTP route shape change; `POST /api/v1/documents/{id}/schedule-publish` now enqueues one River `scheduled_publish_cutover` job in the same transaction
- **Runtime flows:** scheduled publish execution now lives in the dedicated jobs host; stale jobs no-op when state, revision, effective time, or `schedule_generation` drifted
- **Persistence:** documented `documents.schedule_generation` as the invalidation fence carried in the job payload
- **Dependencies:** API now owns River schema migration plus transactional enqueue wiring; `apps/jobs/cmd/metaldocs-jobs` owns worker execution; legacy `effective_date_publisher` ownership removed
- **T-NNN touched:** none
- **R-NNN touched:** none
- **Counts after:** Critical=2 Major=4 Minor=6; missing-ADR=10
- **Tally gate:** PASS preflight
- **Patched files:** `wiki/modules/approval.md`; `wiki/modules/approval/_artifacts/01-surface.md`; `wiki/modules/approval/_artifacts/03-deps.md`; `wiki/modules/approval/_artifacts/sync-log.md`

## 2026-05-16 - controlled-document review route polish

- **Context:** uncommitted diff: approval inbox navigation target changed from `/controlled-documents/{controlled_document_id}` to `/controlled-documents/{controlled_document_id}`.
- **Mode:** lite patch
- **Anchors moved:** none
- **Public surface:** none
- **Routes/API:** frontend navigation spec/backlog references updated; approval HTTP API unchanged
- **Runtime flows:** review/open-document UI path now targets canonical controlled-document route
- **Persistence:** none
- **Dependencies:** none
- **T-NNN touched:** none
- **R-NNN touched:** none
- **Counts after:** Critical=2 Major=4 Minor=6; missing-ADR=10
- **Tally gate:** PASS preflight
- **Patched files:** `wiki/backlog/caixa-aprovacao.md`; `wiki/modules/approval/_artifacts/sync-log.md`

> Append-only log of `metaldocs-module-doc-sync` runs against this module. Newest at top.

## 2026-05-15 - D4 hard-cutover debt/backlog closure sync (0194)

- **Context:** Worker E wiki/docs lane closeout for approval linkage rename debt row and refactor backlog row.
- **Mode:** lite patch
- **Anchors moved:** debt + backlog status wording
- **Public surface:** unchanged
- **Routes/API:** unchanged runtime; docs keep `/api/v1/documents/*` references
- **Persistence:** T-008 marked closed; `approval_instances` linkage naming synchronized to `document_id` with migration 0194 evidence
- **T-NNN touched:** T-008 -> closed (2026-05-15)
- **R-NNN touched:** R-008 -> merged (migration 0194)
- **Counts after:** Critical=2 Major=4 Minor=6; missing-ADR=10
- **Tally gate:** pending
- **Patched files:** `wiki/modules/approval.md`; `wiki/modules/approval-tech-debt.md`; `wiki/backlog/approval-refactor.md`; `wiki/modules/approval/_artifacts/sync-log.md`
## 2026-05-14 - Plan 12.2 caixa-aprovacao screen reality-first sync

- **Context:** commits `a0a90f7e..3d9572cb` (design/spec, approvals screen implementation, review fixes, backlog/design notes sync)
- **Mode:** lite patch
- **Anchors moved:** none
- **Public surface:** none
- **Routes/API:** none (frontend now uses existing approval routes; no backend route/contract delta)
- **Runtime flows:** none
- **Persistence:** none
- **Dependencies:** none
- **T-NNN touched:** none
- **R-NNN touched:** none
- **Counts after:** Critical=2 Major=4 Minor=6; missing-ADR=10
- **Tally gate:** PASS
- **Patched files:** wiki/modules/approval.md; wiki/modules/approval/_artifacts/sync-log.md
## 2026-05-13 - Plan 10 approval column rename + v1 route canonicalization

- **Context:** uncommitted Plan 10 implementation diff (document_id rename migration 0194, constraints validation 0195, route prefix sweep)
- **Mode:** structural refresh
- **Anchors moved:** approval_instances.document_v2_id -> document_id; /api/v2/approval* -> /api/v1/approval*
- **Public surface:** repository error mapping updated for renamed unique index names
- **Routes/API:** approval endpoints and doc-scoped approval routes reflected as /api/v1
- **Runtime flows:** unchanged logic; canonical path updates only
- **Persistence:** index/constraint rename alignment reflected
- **Dependencies:** permission resolver + handler route references aligned
- **T-NNN touched:** T-007/T-008/T-009/T-011 evidence updates
- **R-NNN touched:** R-007..R-009/R-011 alignment updates
- **Counts after:** Critical=2 Major=4 Minor=6; missing-ADR=10
- **Tally gate:** PASS
- **Patched files:** wiki/modules/approval.md; wiki/modules/approval-tech-debt.md; wiki/backlog/approval-refactor.md; wiki/modules/approval/_artifacts/*

## 2026-05-18 - approval unresolved-comments hardening sync

- **Context:** commits after `ac448cdc` closing review findings on approval conflict UX and editor comment-load persistence
- **Mode:** lite patch
- **Anchors moved:** none
- **Public surface:** documented 409 `approval.unresolved_comments` failure mode on final-stage signoff
- **Routes/API:** no route shape change; problem-code behavior now recorded in failure table
- **Runtime flows:** signoff dialog resolves mapped business conflicts inline; unknown approval codes still fall back to safe generic copy
- **Persistence:** none
- **Dependencies:** `wiki/concepts/error-ux.md` updated to reflect shared `resolveErrorMessage(code)` handling on approval dialog conflicts
- **T-NNN touched:** none
- **R-NNN touched:** none
- **Counts after:** Critical=2 Major=4 Minor=6; missing-ADR=10
- **Tally gate:** PASS
- **Patched files:** `wiki/modules/approval.md`; `wiki/modules/approval/_artifacts/sync-log.md`; `wiki/concepts/error-ux.md`
