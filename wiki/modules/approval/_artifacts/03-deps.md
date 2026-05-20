# Cross-deps — `internal/modules/documents/approval`

Generated: 2026-05-20  
Module root: `internal/modules/documents/approval/`  
Composition root: `apps/api/cmd/metaldocs-api/main.go`

---

## 1. Imports OUT

Internal MetalDocs packages imported by this module (non-test files only). Self-imports (intra-approval sub-packages) are excluded.

| Imported package | First seen in (file:line) | Symbols used | Purpose (1 line) |
|---|---|---|---|
| `internal/modules/iam/authz` | `application/cancel_service.go:12` | `Require`, `RequireAll` | Row-level authz enforcement via Postgres GUC |
| `internal/modules/iam/domain` | `http/handler.go:15` | `User` (via alias `iamdomain`) | HTTP handler receives IAM user context |
| `internal/modules/documents/application` | `application/decision_service.go:13` | `PDFDispatchInvoker` (via alias `docapp`) | Adapter interface for document PDF dispatch |
| `internal/modules/documents/domain` | `http/errors.go:16` | `ErrDocumentNotFound` (via alias `v2dom`) | Maps document-domain errors to HTTP responses |
| `internal/platform/tenant` | `http/handler.go:16` | `FromContext` | Extracts tenant from request context |
| `internal/platform/idempotency` | `infrastructure/postgres_signoff_idemp_store.go:9` | `Store`, `Key` | Signoff idempotency store interface |
| `internal/test` | `application/services.go:8` | `E2EClockOffset` (via alias `e2etest`) | Applies E2E clock offset to `RealClock.Now()` in non-production builds |

---

## 2. Imports IN

External packages (outside `internal/modules/documents/approval/`) that import this module. Intra-module imports omitted.

| Importer package | File:line of import | Symbols used | Notes |
|---|---|---|---|
| `apps/api/cmd/metaldocs-api` | `main.go:24` | `NewServices`, `NewDecisionService`, `NewSQLEmitter`, `RealClock`, `PDFDispatchInvoker`, `FreezeInvoker`, `CancelInput`, `GovernanceEvent`, `SubmitRequest`, `ScheduledPublishEnqueuer` | Primary HTTP composition root; wires approval services and the transactional River enqueue seam |
| `apps/api/cmd/metaldocs-api` | `main.go:25` | `NewHandler`, `RegisterRoutes` | Registers approval HTTP routes on mux |
| `apps/api/cmd/metaldocs-api` | `main.go:26` | `NewPostgresSignoffIdempStore` | Constructs signoff idempotency store |
| `apps/api/cmd/metaldocs-api` | `main.go:27` | `NewPostgresApprovalRepository` | Constructs Postgres approval repo |
| `internal/modules/documents` | `module.go:9` | `SubmitRequest`, `SubmitResult` (via `approvalapp`) | `documents.Deps.SubmitSvc` interface references approval types |
| `internal/modules/documents/delivery/http` | `handler.go:15` | `SubmitRequest`, `SubmitResult` (via `approvalapp`) | Finalize handler calls `SubmitRevisionForReview` via interface |
| `apps/jobs/cmd/metaldocs-jobs` | `main.go:14` | `NewServices`, `NewSQLEmitter`, `RealClock`, `NewWorkers` | Dedicated jobs runtime owns scheduled publish execution through River workers |
| `internal/modules/jobs/stuck_instance_watchdog` | `job.go:10` | `CancelInstance`, `CancelInput`, `GovernanceEvent`, `Emitter` | Background job cancels stuck approval instances |
| `internal/modules/iam` | `integration_test.go:16` | `repository` package — table fixture setup | IAM integration test inserts approval fixture rows to test RLS caps |

---

## 3. DI / wiring touchpoints

| Site | File:line | What is wired |
|---|---|---|
| `apps/api/cmd/metaldocs-api/main.go` | `main.go:299` | `approvalrepo.NewPostgresApprovalRepository(deps.SQLDB)` → `approvalRepo` |
| `apps/api/cmd/metaldocs-api/main.go` | `main.go:300` | `approvalapp.NewSQLEmitter()` → `approvalEmitter` |
| `apps/api/cmd/metaldocs-api/main.go` | `main.go:301` | `approvalapp.NewServices(approvalRepo, approvalEmitter, approvalapp.RealClock{})` → `approvalServices` |
| `apps/api/cmd/metaldocs-api/main.go` | `main.go:313` | `approvalapp.NewDecisionService(approvalRepo, approvalEmitter, RealClock{}, effectiveFreezeInvoker, pdfDispatchAdapter).WithPDFOutbox(pdfOutboxRepo)` → overwrites `approvalServices.Decision` |
| `apps/api/cmd/metaldocs-api/main.go` | `main.go:316` | `docDeps.SubmitSvc = approvalServices.Submit` — injects Submit into documents module deps |
| `apps/api/cmd/metaldocs-api/main.go` | `main.go:330` | `approvalinfra.NewPostgresSignoffIdempStore(deps.SQLDB)` → `signoffIdempStore` |
| `apps/api/cmd/metaldocs-api/main.go` | `main.go:331` | `approvalhttp.NewHandler(approvalServices, deps.SQLDB, signoffIdempStore)` → `approvalHandler` |
| `apps/api/cmd/metaldocs-api/main.go` | `main.go:332` | `approvalHandler.RegisterRoutes(mux)` |
| `apps/api/cmd/metaldocs-api/main.go` | `main.go:307` | `bootstrap.MigrateRiverSchema(ctx, deps.SQLDB, jobsCfg.RiverSchema)` — startup River schema sync for scheduled publish |
| `apps/api/cmd/metaldocs-api/main.go` | `main.go:318` | `approvalServices.WithScheduledPublishEnqueuer(approvaljobs.NewScheduledPublishEnqueuer(riverBundle.Client))` — transactional enqueue owned by API |
| `apps/api/cmd/metaldocs-api/main.go` | `main.go:352` | `stuck_instance_watchdog.New(deps.SQLDB, approvalServices.Cancel, approvalEmitter)` — watchdog job |
| `apps/jobs/cmd/metaldocs-jobs/main.go` | `main.go:29` | `bootstrap.BuildJobsDependencies(... approvaljobs.NewWorkers(...))` — dedicated jobs host for scheduled publish execution |

---

## 4. Configuration surface

Env vars and Postgres GUC keys read by or in service of this module. No `os.Getenv` calls inside the module itself — env reads for this boundary now live in composition/runtime config (`main.go` + `internal/platform/config/jobs.go`).

| Name | Read at (file:line) | Required? | Default |
|---|---|---|---|
| `METALDOCS_FANOUT_URL` | `main.go:228` (configures `FreezeInvoker` passed into this module) | No | If unset, `noopFreezeInvoker` is passed; `slog.Warn` emitted; freeze step will fail at runtime |
| `METALDOCS_JOBS_ENABLED` | `internal/platform/config/jobs.go:27` | No | Defaults true; jobs runtime exits early only when explicitly disabled |
| `METALDOCS_JOBS_RIVER_SCHEMA` | `internal/platform/config/jobs.go:21` | No | Empty string uses River default schema |
| `METALDOCS_JOBS_TEMPORAL_MAX_WORKERS` | `internal/platform/config/jobs.go:31` | No | Defaults queue `temporal` to 10 workers |
| `ENABLE_JOB_STUCK_INSTANCE_WATCHDOG` | `main.go:348` | No | Job disabled if unset |
| Postgres GUC `metaldocs.tenant_id` | `application/authz_guc.go:12` | — | Set via `set_config` inside tx; not an env var |
| Postgres GUC `metaldocs.actor_id` | `application/authz_guc.go:15` | — | Set via `set_config` inside tx; not an env var |
| Postgres GUC `metaldocs.bypass_authz` | `application/cancel_service.go:90` | — | Set to `'system'` for system-initiated cancels |
| Postgres GUC `metaldocs.cancel_in_progress` | `application/cancel_service.go:111`, `application/decision_service.go:334` | — | Set to instance ID during cancel/decision tx |

---

## 5. Test surface

| Test file | Subject (file under test) | Kind |
|---|---|---|
| `application/cancel_service_test.go` | `cancel_service.go` | unit |
| `application/content_hash_test.go` | `content_hash.go` | unit |
| `application/coverage_boost_test.go` | multiple application services | unit |
| `application/cutover_service_test.go` | `cutover_service.go` | unit |
| `application/decision_service_freeze_test.go` | `decision_service.go` (freeze path) | unit |
| `application/decision_service_test.go` | `decision_service.go` | unit |
| `application/events_test.go` | `events.go` | unit |
| `application/idempotency_test.go` | `idempotency.go` | unit |
| `application/membership_tx_test.go` | `membership_tx.go` | unit |
| `application/obsolete_service_test.go` | `obsolete_service.go` | unit |
| `application/phase5_integration_test.go` | cross-service scenario (no real DB) | unit (in-process) |
| `application/publish_service_test.go` | `publish_service.go` | unit |
| `application/read_service_test.go` | `read_service.go` | unit |
| `application/route_admin_service_test.go` | `route_admin_service.go` | unit |
| `application/scheduler_test_helpers_test.go` | shared scheduler test rows helper | unit |
| `application/submit_eligible_actors_test.go` | submit eligibility logic | integration (`//go:build integration`) |
| `application/submit_service_test.go` | `submit_service.go` | unit |
| `application/supersede_service_test.go` | `supersede_service.go` | unit |
| `domain/drift_test.go` | `drift.go` | unit |
| `domain/eligibility_test.go` | `eligibility.go` | unit |
| `domain/instance_test.go` | `instance.go` | unit |
| `domain/integration_test.go` | domain type round-trips (JSON serialisation) | unit |
| `domain/quorum_test.go` | `quorum.go` | unit |
| `domain/route_test.go` | `route.go` | unit |
| `domain/signoff_test.go` | `signoff.go` | unit |
| `domain/sod_test.go` | `sod.go` | unit |
| `domain/state_test.go` | `state.go` | unit |
| `http/cancel_handler_test.go` | `cancel_handler.go` | unit |
| `http/contracts/contracts_test.go` | contracts serialisation | unit |
| `http/errors_test.go` | `errors.go` | unit |
| `http/get_instance_handler_test.go` | `get_instance_handler.go` | unit |
| `http/inbox_handler_test.go` | `inbox_handler.go` | unit |
| `http/obsolete_handler_test.go` | `obsolete_handler.go` | unit |
| `http/publish_handler_test.go` | `publish_handler.go` | unit |
| `http/route_admin_handler_test.go` | `route_admin_handler.go` | unit |
| `http/router_test.go` | `router.go` | unit |
| `http/signoff_handler_test.go` | `signoff_handler.go` | unit |
| `http/submit_handler_test.go` | `submit_handler.go` | unit |
| `http/supersede_handler_test.go` | `supersede_handler.go` | unit |
| `infra/signature/password_reauth_test.go` | `password_reauth.go` | unit |
| `infra/signature/registry_test.go` | signature provider registry | unit |
| `infrastructure/postgres_signoff_idemp_store_test.go` | `postgres_signoff_idemp_store.go` | unit (nil-DB guard) |
| `repository/postgres_approval_repository_test.go` | `postgres_approval_repository.go` | integration (requires Postgres) |
