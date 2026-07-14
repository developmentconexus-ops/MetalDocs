# Stage-1 Audit Artifact — module-approval

**Area:** `internal/modules/documents/approval`
**Files:** 107 Go source files (production + test)
**Produced:** 2026-06-10 (Stage 1 — truth map only; no redesign proposals)
**Code read as of branch:** `qa/iam-area-membership`

---

## 1. Identity & Purpose

The approval module owns the ISO 9001 sign-off chain for controlled documents. It accepts submit requests from the documents module, snapshots the eligible-approver pool per stage at that moment (J1 rule), accumulates per-stage signoffs under configurable quorum policies (`any_1_of`, `all_of`, `m_of_n`), drives the approval instance through its lifecycle (in_progress → approved/rejected/cancelled), and emits the complete governance audit trail inside the same transaction as every state change.

The module lives under `internal/modules/documents/approval` rather than as a top-level sibling because, at initial design time, approval was tightly coupled to the documents domain: the submit entry-point is called from `documents/delivery/http` (via `finalizeDocument`), approval writes directly to the `documents` table, and the eligibility snapshot reads `metaldocs.user_process_areas`. The nesting reflects that originating coupling. As of the current codebase the approval package owns its own transport (`http/`), domain (`domain/`), application services (`application/`), repository (`repository/`), infrastructure adapters (`infrastructure/`), and River job worker (`jobs/`) — making it a self-contained subdomain that happens to share the `documents/` prefix.

Defense-in-depth authz is the defining feature: three Go-layer checks (tier-1 middleware, tier-2 `authz.Require` in-tx, tier-3 Postgres tripwire `enforce_capability_asserted`) plus a fourth layer on signoff (`enforce_signoff_eligibility_trg`). Every mutating operation emits a `governance_events` row in the same transaction, making the audit trail closed-loop by construction.

---

## 2. File Inventory

### subpackage: `api/`
| File | Role |
|---|---|
| `api/api.gen.go` | oapi-codegen generated server interface, wrapper, and all request/response types for the 16 approval routes; do not edit manually |
| `api/gen.go` | oapi-codegen generator directive (`//go:generate`) |

### subpackage: `application/`
| File | Role |
|---|---|
| `application/services.go` | `Services` composition root; `Clock` interface; `RealClock`; `ValidateEventPayload`; `ScheduledPublishEnqueuer` port interface; error `ErrContentHashMismatch` |
| `application/authz_guc.go` | `setAuthzGUC` — writes `metaldocs.tenant_id` + `metaldocs.actor_id` GUCs inside a tx; called first by every mutating service |
| `application/membership_tx.go` | `WithMembershipContext` — older tx helper that sets membership-writer role GUCs; now only internally referenced; not called by submit/decision path |
| `application/submit_service.go` | `SubmitService.SubmitRevisionForReview` — full submit flow: hash, idempotency key, authz, route load, stage instance creation, OCC doc UPDATE, governance event, commit |
| `application/submit_service_test.go` | Unit tests for `SubmitRevisionForReview` (7 tests) |
| `application/submit_eligible_actors_test.go` | Integration test — verifies eligible-actor resolution against a live DB fixture |
| `application/decision_service.go` | `DecisionService.RecordSignoff` — approve/reject flow; eligibility drift; quorum eval; freeze/pin; PDF outbox; governance event; deprecated post-commit dispatcher path still compiled |
| `application/decision_service_test.go` | 9 unit tests for signoff path |
| `application/decision_service_freeze_test.go` | Tests for freeze/pin invocation on final signoff |
| `application/decision_service_reauth_test.go` | Tests for password reauth signature integration |
| `application/cancel_service.go` | `CancelService.CancelInstance` + `SystemCancelInstance`; sets `metaldocs.cancel_in_progress` GUC; emits governance event |
| `application/cancel_service_test.go` | Unit tests for cancel path |
| `application/publish_service.go` | `PublishService.PublishApproved` (approved → published) and `SchedulePublish` (approved → scheduled); River enqueuer injection |
| `application/publish_service_test.go` | Unit tests for publish and schedule paths |
| `application/scheduler_service.go` | `SchedulerService.RunScheduledPublishJob` — River worker execution path; stale-generation no-op; supersede handling |
| `application/supersede_service.go` | `SupersedeService.PublishSuperseding` — atomic publish-as-superseder |
| `application/supersede_service_test.go` | Unit tests for supersede path |
| `application/obsolete_service.go` | `ObsoleteService.MarkObsolete` — published/superseded → obsolete |
| `application/obsolete_service_test.go` | Unit tests for obsolete path |
| `application/cutover_service.go` | `CutoverService.ValidateLegacyCutoverReady` — preflight for migration 0142 legacy cutover; **not wired in production composition root** (see §10) |
| `application/cutover_service_test.go` | Tests for cutover preflight |
| `application/read_service.go` | `ReadService` — `LoadInstance`, `LoadActiveInstanceByDocument`, `LoadActiveInstanceByDocumentForMutation`, `ListPendingForActor`, `ListInboxItems`, `CountPendingForActor` |
| `application/read_service_test.go` | Unit tests for read paths |
| `application/route_admin_service.go` | `RouteAdminService.Create/Update/Deactivate/List` with idempotency; governance events; FK/version guard helpers |
| `application/route_admin_service_test.go` | Unit tests for route admin operations |
| `application/route_admin_idemp.go` | `RouteAdminIdempStore` port; `RouteAdminReplayCommitter`; payload hash functions for create/update/deactivate |
| `application/events.go` | `EventEmitter` interface; `sqlEmitter`; `GovernanceEvent` struct; `EventType` constants; `MemoryEmitter` for tests |
| `application/events_test.go` | Unit tests for event emitter |
| `application/content_hash.go` | `ComputeContentHash` — canonical SHA-256 over tenantID+docID+revisionNumber+formData; `ErrFloatInFormData` guard |
| `application/content_hash_test.go` | Tests for hash computation |
| `application/idempotency.go` | `ComputeIdempotencyKey` — deterministic submit-path key derivation |
| `application/idempotency_test.go` | Tests for key derivation |
| `application/membership_tx_test.go` | Tests for `WithMembershipContext` (tx helper) |
| `application/coverage_boost_test.go` | Coverage-padding tests targeting ≥90% on the application package |
| `application/phase5_integration_test.go` | 2-test integration suite exercising Phase 5 service composition |

### subpackage: `domain/`
| File | Role |
|---|---|
| `domain/instance.go` | `Instance` aggregate root; `StageInstance`; `InstanceStatus`/`StageStatus` constants; `AdvanceStage`, `RejectHere`, `SkipStage`, `BumpRevisionVersion`, `Cancel` methods |
| `domain/instance_test.go` | Unit tests for instance state machine methods |
| `domain/signoff.go` | `Signoff` immutable value object (unexported fields + getters); `NewSignoff` constructor with SHA-256 hash validation; `MarshalJSON` |
| `domain/signoff_test.go` | Unit tests for signoff construction |
| `domain/sod.go` | `CheckSoD` — pure SoD rule (submitter ≠ approver, no re-sign across stages) |
| `domain/sod_test.go` | Unit tests for SoD rules |
| `domain/eligibility.go` | `CheckEligibility` — pure J1 rule against `EligibleActorIDs` snapshot |
| `domain/eligibility_test.go` | Unit tests for eligibility check |
| `domain/quorum.go` | `EvaluateQuorum` / `EvaluateQuorumResult` — handles `any_1_of`, `all_of`, `m_of_n`; `ComputeEffectiveDenominator` |
| `domain/quorum_test.go` | 8 unit tests covering all quorum variants |
| `domain/drift.go` | `ApplyEligibilityDrift` — pure drift policy computation (`reduce_quorum`, `fail_stage`, `keep_snapshot`) |
| `domain/drift_test.go` | Unit tests for drift policies |
| `domain/route.go` | `Route`, `Stage` structs; `QuorumPolicy`, `DriftPolicy` enum types; `Route.Validate` structural invariant |
| `domain/route_test.go` | Unit tests for route validation |
| `domain/state.go` | `DocState` and the 8-state Spec 2 lifecycle graph; `IsLegalTransition`; `StateFromString` (rejects legacy "finalized"/"archived") |
| `domain/state_test.go` | Unit tests for state machine edges |
| `domain/errors.go` | `ErrEmptyEligiblePool`, and other domain-level sentinels |
| `domain/integration_test.go` | Integration test for domain rules under live DB |

### subpackage: `http/`
| File | Role |
|---|---|
| `http/handler.go` | `Handler` struct + `NewHandler`; service interface definitions (`submitService`, `decisionService`, etc.); `parseIfMatch`; `instanceETag` |
| `http/handler_test.go` | Tests for handler construction and `parseIfMatch` |
| `http/router.go` | `RegisterRoutes` — mounts all 16 routes onto `*http.ServeMux` via `approvalapi.ServerInterfaceWrapper` |
| `http/router_test.go` | Tests for route registration |
| `http/routes_generated.go` | Thin adapters implementing `approvalapi.ServerInterface` by delegating to concrete handler methods (bridging codegen params to raw handler) |
| `http/errors.go` | `MapErrorToResponse`, `WriteError`, `WriteJSON`; 40+ typed `problem.Code` constants; `ValidationError` type |
| `http/errors_test.go` | Tests for error mapping |
| `http/signoff_handler.go` | `SignoffHandler` — stage-scoped signoff (POST `.../stages/{stage_id}/signoffs`); `ErrIdempotencyRequired`; `ErrContentHashMismatch` re-export |
| `http/signoff_handler_test.go` | Tests for stage-scoped signoff handler |
| `http/signoff_idempotency.go` | `signoffPayloadHash` — derives the replay fingerprint for document-scoped and stage-scoped signoff handlers |
| `http/doc_approval_handler.go` | `GetInstanceByDocumentHandler`, `SignoffByDocumentHandler`, `CancelByDocumentHandler`; `failReplaySlot` helper; uses legacy `"log"` package (see §10) |
| `http/submit_handler.go` | `SubmitHandler` — POST `documents/{id}/submit` |
| `http/submit_handler_test.go` | Tests for submit handler |
| `http/publish_handler.go` | `PublishHandler` and `SchedulePublishHandler` |
| `http/publish_handler_test.go` | Tests for publish handlers |
| `http/cancel_handler.go` | `CancelHandler` — instance-scoped cancel |
| `http/cancel_handler_test.go` | Tests for cancel handler |
| `http/get_instance_handler.go` | `GetInstanceHandler` — GET `.../instances/{instance_id}` |
| `http/get_instance_handler_test.go` | Tests |
| `http/inbox_handler.go` | `InboxHandler` — GET `/api/v1/approval/inbox`; `parseInboxLimit`/`parseInboxOffset` |
| `http/inbox_handler_test.go` | Tests |
| `http/obsolete_handler.go` | `ObsoleteHandler` |
| `http/obsolete_handler_test.go` | Tests |
| `http/supersede_handler.go` | `SupersedeHandler` |
| `http/supersede_handler_test.go` | Tests |
| `http/route_admin_handler.go` | `CreateRouteHandler`, `UpdateRouteHandler`, `DeactivateRouteHandler`, `ListRoutesHandler`; `mapInstanceResponse` |
| `http/route_admin_handler_test.go` | Tests for route admin handlers |
| `http/contracts/cancel.go` | `CancelRequest`/`CancelResponse` DTOs + Validate |
| `http/contracts/instance_read.go` | `ApprovalInstanceResponse` + stage/signoff view types |
| `http/contracts/publish.go` | `PublishRequest`/`SchedulePublishRequest` DTOs |
| `http/contracts/route.go` | `CreateRouteRequest`, `UpdateRouteRequest`, `DeactivateRouteRequest`, `ListRoutesResponse`, `RouteSummary`, `StageSummary` |
| `http/contracts/signoff.go` | `SignoffRequest`/`SignoffResponse` DTOs + Validate |
| `http/contracts/submit.go` | `SubmitRequest`/`SubmitResponse` DTOs |
| `http/contracts/supersede.go` | `SupersedeRequest`/`SupersedeResponse` DTOs |
| `http/contracts/obsolete.go` | `ObsoleteRequest`/`ObsoleteResponse` DTOs |
| `http/contracts/strictjson.go` | `Decode` — strict JSON decoder (unknown-field detection, Content-Type check, size limit, empty-body check); `ErrValidation`, `ErrContentType`, `ErrBodyTooLarge`, `ErrEmptyBody` sentinels |
| `http/contracts/errors.go` | Contract-level error sentinels used by validators |
| `http/contracts/contracts_test.go` | Tests for contract validation and `Decode` |

### subpackage: `infrastructure/`
| File | Role |
|---|---|
| `infrastructure/postgres_signoff_idemp_store.go` | `PostgresSignoffIdempStore` — Stripe-style idempotency for signoff (document-scoped and stage-scoped routes); wraps `internal/platform/idempotency.Store` |
| `infrastructure/postgres_signoff_idemp_store_test.go` | Tests for signoff idempotency store |
| `infrastructure/postgres_route_admin_idemp_store.go` | `PostgresRouteAdminIdempStore` — idempotency for Create/Update/Deactivate route admin operations; ~95% wrapper code duplicated from signoff store (T-014) |
| `infrastructure/signature/provider.go` | `Provider` interface; `Registry` — pluggable signature method dispatch |
| `infrastructure/signature/password_reauth.go` | `PasswordReauthProvider` — bcrypt verify + `InMemoryAuthFailureRateLimiter` (process-local; production wired with same; see §10) |
| `infrastructure/signature/password_reauth_test.go` | Tests for password reauth |
| `infrastructure/signature/registry_test.go` | Tests for provider registry |

### subpackage: `jobs/`
| File | Role |
|---|---|
| `jobs/scheduled_publish_args.go` | `ScheduledPublishArgs` — River job payload struct |
| `jobs/scheduled_publish_job.go` | `ScheduledPublishWorker` (River worker); `RiverScheduledPublishEnqueuer`; `NewWorkers`; `NewScheduledPublishEnqueuer` |
| `jobs/scheduled_publish_job_test.go` | Unit tests for scheduled publish worker |

### subpackage: `repository/`
| File | Role |
|---|---|
| `repository/approval_repository.go` | `ApprovalRepository` interface (16 methods); `Route`/`RouteStage` projection structs; `SignoffInsertResult`; `ScheduledPublishRow` |
| `repository/errors.go` | All repository error sentinels + `MapPgError` (maps pgconn error codes to domain errors) |
| `repository/postgres_approval_repository.go` | `postgresApprovalRepository` — all 16 interface methods; `listRoutesQuery` const; `scanRouteListRows`; `loadStageInstances`; `loadSignoffsForInstance` |
| `repository/postgres_approval_repository_test.go` | Integration tests for repository |

---

## 3. Public Surface

### Exported types / functions consumed outside this module

| Symbol | File:line | Consumers |
|---|---|---|
| `application.Services` | `application/services.go:25` | `apps/api/cmd/metaldocs-api/main.go:433` |
| `application.NewServices` | `application/services.go:53` | `main.go:433` |
| `application.SubmitRevisionForReview` | `application/submit_service.go:48` | `documents/delivery/http/handler.go` (via `approvalSubmitter` interface), `documents/module.go` |
| `application.SubmitRequest` / `SubmitResult` | `application/submit_service.go:29,42` | `documents/delivery/http/handler.go` |
| `application.RecordSignoff` | `application/decision_service.go:159` | `http/signoff_handler.go`, `http/doc_approval_handler.go` |
| `application.CancelInstance` / `SystemCancelInstance` | `application/cancel_service.go:44,50` | `http/cancel_handler.go`; `jobs/stuck_instance_watchdog/job.go` |
| `application.PublishApproved` | `application/publish_service.go:48` | `http/publish_handler.go` |
| `application.SchedulePublish` | `application/publish_service.go:190` | `http/publish_handler.go` |
| `application.RunScheduledPublishJob` | `application/scheduler_service.go:44` | `jobs/scheduled_publish_job.go` |
| `application.MarkObsolete` | `application/obsolete_service.go:42` | `http/obsolete_handler.go` |
| `application.PublishSuperseding` | `application/supersede_service.go:40` | `http/supersede_handler.go` |
| `application.RouteAdminService.Create/Update/Deactivate/List` | `application/route_admin_service.go` | `http/route_admin_handler.go` |
| `application.ComputeContentHash` | `application/content_hash.go:35` | `http/signoff_handler.go`, `http/doc_approval_handler.go` |
| `application.EventEmitter` | `application/events.go:34` | `main.go:432` |
| `application.NewSQLEmitter` | `application/events.go:42` | `main.go:432` |
| `application.ErrRevisionTitleRequired` | `application/submit_service.go:250` | `documents/delivery/http/handler.go:1173` |
| `domain.CheckSoD` | `domain/sod.go:15` | `application/decision_service.go:284` |
| `domain.CheckEligibility` | `domain/eligibility.go:11` | `application/decision_service.go:261`; `http/doc_approval_handler.go:168` |
| `domain.EvaluateQuorum` / `EvaluateQuorumResult` | `domain/quorum.go:41,45` | `application/decision_service.go:361` |
| `domain.ApplyEligibilityDrift` | `domain/drift.go:17` | `application/decision_service.go:357` |
| `domain.Instance` | `domain/instance.go:58` | `application/` services, `http/handler.go` |
| `domain.Route` + `Validate` | `domain/route.go:39` | `application/submit_service.go:112`; `route_admin_service.go:177` |
| `domain.IsLegalTransition` | `domain/state.go:60` | `application/obsolete_service.go`, `cancel_service.go` |
| `repository.NewPostgresApprovalRepository` | `repository/postgres_approval_repository.go:22` | `main.go:431` |
| `repository.ApprovalRepository` | `repository/approval_repository.go:60` | all application services |
| `repository.MapPgError` | `repository/errors.go:34` | `application/route_admin_service.go:193`; repository internal |
| `infrastructure.NewPostgresSignoffIdempStore` | `infrastructure/postgres_signoff_idemp_store.go:59` | `main.go:514` |
| `infrastructure.NewPostgresRouteAdminIdempStore` | `infrastructure/postgres_route_admin_idemp_store.go` | `main.go:515` |
| `signature.Provider` / `Registry` / `NewRegistry` | `infrastructure/signature/provider.go` | `application/decision_service.go:69`; `apps/api/cmd/metaldocs-api/reauth.go:44` |
| `signature.NewPasswordReauthProvider` | `infrastructure/signature/password_reauth.go:56` | `apps/api/cmd/metaldocs-api/reauth.go:46` |
| `signature.NewInMemoryAuthFailureRateLimiter` | `infrastructure/signature/password_reauth.go:125` | `apps/api/cmd/metaldocs-api/reauth.go:49` |
| `jobs.NewWorkers` | `jobs/scheduled_publish_job.go:70` | `apps/jobs/cmd/metaldocs-jobs/main.go` |
| `jobs.NewScheduledPublishEnqueuer` | `jobs/scheduled_publish_job.go:76` | `main.go:450` |
| `approvalhttp.NewHandler` | `http/handler.go:79` | `main.go:517` |
| `approvalhttp.Handler.RegisterRoutes` | `http/router.go:10` | `main.go:518` |

### HTTP Routes

All 16 routes are mounted via `approvalapi.ServerInterfaceWrapper` at `http/router.go:10-37`.

| Method | Path | Handler | Authz (tier-2 capability) |
|---|---|---|---|
| POST | `/api/v1/documents/{id}/submit` | `SubmitHandler` | `doc.submit` + `doc.edit` (area-grade) |
| POST | `/api/v1/documents/{id}/signoff` | `SignoffByDocumentHandler` | `doc.signoff` (area-grade) |
| POST | `/api/v1/approval/instances/{instance_id}/stages/{stage_id}/signoffs` | `SignoffHandler` | `doc.signoff` (area-grade) |
| POST | `/api/v1/documents/{id}/publish` | `PublishHandler` | `doc.publish` + `doc.edit` (area-grade) |
| POST | `/api/v1/documents/{id}/schedule-publish` | `SchedulePublishHandler` | `doc.publish` + `doc.edit` (area-grade) |
| POST | `/api/v1/documents/{id}/supersede` | `SupersedeHandler` | `doc.publish` + `doc.edit` (area-grade) |
| POST | `/api/v1/documents/{id}/obsolete` | `ObsoleteHandler` | `doc.obsolete` (area-grade) |
| POST | `/api/v1/documents/{id}/cancel` | `CancelByDocumentHandler` | `doc.edit` (area-grade) |
| POST | `/api/v1/approval/instances/{instance_id}/cancel` | `CancelHandler` | `doc.edit` (area-grade) |
| GET | `/api/v1/documents/{id}/approval-instance` | `GetInstanceByDocumentHandler` | `doc.view` (tier-1 only; coalesced to tenant-grade in service) |
| GET | `/api/v1/approval/instances/{instance_id}` | `GetInstanceHandler` | `doc.view` (tier-1 only; coalesced to tenant-grade) |
| GET | `/api/v1/approval/inbox` | `InboxHandler` | tier-1 only + JSONB `@>` actor scoping |
| POST | `/api/v1/approval/routes` | `CreateRouteHandler` | `route.admin` (tenant-grade) |
| PUT | `/api/v1/approval/routes/{id}` | `UpdateRouteHandler` | `route.admin` (tenant-grade) |
| POST | `/api/v1/approval/routes/{id}/deactivate` | `DeactivateRouteHandler` | `route.admin` (tenant-grade) |
| GET | `/api/v1/approval/routes` | `ListRoutesHandler` | `route.admin` (tenant-grade) |

**Note:** The deactivate route is registered as `POST .../deactivate` at `router.go:27`, not `DELETE .../routes/{id}`. The wiki's API Route Truth Table still lists it as `DELETE /api/v1/approval/routes/{id}` — see §11.

---

## 4. Logic Flows

### Flow 1: SubmitRevisionForReview

The submit path has two entry points: `documents/delivery/http` calls `SubmitRevisionForReview` from `finalizeDocument` (`documents/delivery/http/handler.go:575`), and the direct HTTP handler `SubmitHandler` (`http/submit_handler.go:14`) is also exposed.

**Steps** (anchored to `application/submit_service.go`):

1. **Payload validation** `:50` — `ValidateEventPayload` rejects any float64 in `ContentFormData` (JSON marshal default); returns `ErrFloatInFormData`.
2. **Content hash** `:55` — `ComputeContentHash({TenantID, DocumentID, RevisionNumber, FormData})` computes canonical SHA-256 via `content_hash.go:35`.
3. **Idempotency key** `:66` — `ComputeIdempotencyKey({ActorUserID, DocumentID, Timestamp})` deterministic key for UNIQUE `(document_id, idempotency_key)` guard.
4. **BEGIN TX** `:76` — `db.BeginTx`; `authz.WithCapCache(ctx)`.
5. **GUC write** `:83` — `setAuthzGUC(ctx, tx, TenantID, SubmittedBy)` sets `metaldocs.tenant_id` + `metaldocs.actor_id` via `set_config(..., true)`.
6. **Area resolution** `:90` — `docapp.LoadDocumentAreaCode` — shared resolver (ADR 0022 F7); fail-closes on missing area.
7. **Tier-2 authz** `:95,99` — `authz.Require(doc.submit)` then `authz.Require(doc.edit)`.
8. **Route load** `:105` — inline `loadRoute` (SQL inside tx) fetches `approval_routes` + `approval_route_stages`; returns `domain.Route`.
9. **Route validation** `:112` — `route.Validate()` checks dense stage ordering, quorum+M consistency, unique names.
10. **Instance creation** `:141` — `repo.InsertInstance(ctx, tx, inst)` — triggers `enforce_capability_asserted` BEFORE INSERT on `approval_instances`.
11. **Eligible actor snapshot** (loop `:160`) — per stage, `resolveEligibleActors` queries `metaldocs.user_process_areas` with `effective_from <= now()` predicate; first stage marked `active`, others `pending`.
12. **Bulk stage insert** `:182` — `repo.InsertStageInstances` — single multi-row VALUES; persists `eligible_actor_ids` as JSONB.
13. **OCC document UPDATE** `:188` — `UPDATE documents SET status='under_review' … WHERE revision_version=$3`; 0 rows affected → `ErrStaleRevision`.
14. **Governance event** `:213` — `emitter.Emit(ctx, tx, GovernanceEvent{EventType:"approval_submitted"})`.
15. **COMMIT** `:240`.

### Flow 2: RecordSignoff (approve/reject path)

Entry point: `SignoffByDocumentHandler` (`http/doc_approval_handler.go:87`) or `SignoffHandler` (`http/signoff_handler.go:20`).

**Steps** (anchored to `application/decision_service.go`):

1. **HTTP idempotency** — `h.idempStore.BeginDocumentReplay` (`http/doc_approval_handler.go:139`) or `BeginStageReplay`; replay fingerprint includes docID+decision+reason+content_hash but NOT server-resolved instanceID/stageID (ADR 0017; `signoff_idempotency.go:11`).
2. **Payload validation** `decision_service.go:161` — `ValidateEventPayload` on `SignaturePayload`.
3. **BEGIN TX** `:167`.
4. **GUC write** `:172` — `setAuthzGUC`.
5. **LoadInstance** `:178` — `repo.LoadInstance` with `SELECT … FOR UPDATE` on stage rows (J1 lock prevents concurrent re-snapshot).
6. **Instance terminal check** `:196` — `instance.Status != InstanceInProgress` → `ErrInstanceCompleted`.
7. **Area resolution** `:202` — `docapp.LoadDocumentAreaCode`.
8. **Tier-2 authz** `:207` — `authz.Require(doc.signoff)`.
9. **E-signature reauth** `:217` — `resolveSignaturePayload` — for `password_reauth` method, verifies bcrypt via `PasswordReauthProvider.Sign`; raw credential never stored.
10. **Content pin** `:230` — `loadActiveDocumentContentHash` mirrors the COALESCE in the active-document endpoint; compares against client-echoed `_content_hash`; fails closed on missing or null.
11. **Active stage check** `:249` — `instance.Active()` returns nil → `ErrNoActiveStage`.
12. **J1 eligibility** `:261` — `domain.CheckEligibility(actorID, stage.EligibleActorIDs)`. Not eligible → emits `signoff.rejected` governance event in a separate tx (`emitEligibilityRejection`) and returns `ErrActorNotEligible`.
13. **SoD check** `:284` — `domain.CheckSoD(submittedBy, actorID, priorSignoffs)` → `ErrAuthorCannotSign` or `ErrActorAlreadySigned`.
14. **Actor display name lookup** `:292` — inline SELECT from `metaldocs.iam_users`.
15. **InsertSignoff** `:324` — `repo.InsertSignoff` — `INSERT ON CONFLICT DO NOTHING RETURNING id`; no RETURNING → field-compare replay vs `ErrActorAlreadySigned`; tripwires `enforce_capability_asserted` and `enforce_signoff_eligibility_trg` fire here.
16. **Quorum evaluation** `:341` — `domain.ApplyEligibilityDrift` + `domain.EvaluateQuorum` for `QuorumApprovedStage`, `QuorumRejectedStage`, or `QuorumPending`.
17. **Approve branch** `:376` — stage → `completed`; `instance.AdvanceStage()`; if final: `enforce_capability_asserted` on doc UPDATE requires `doc.edit` (`authz.Require` `:407`); `pinInvoker.Pin` (in-tx); `UPDATE documents SET status='approved'`; PDF outbox enqueue `:554`.
18. **Reject branch** `:470` — stage → `rejected_here`; instance → `rejected`; `set_config('metaldocs.cancel_in_progress', instanceID, true)` `:484` (trigger whitelist); `authz.Require(doc.edit)` `:491`; `UPDATE documents SET status='draft', revision_version+=1`.
19. **Governance event** `:536` — `signoff_recorded` in same tx.
20. **COMMIT** `:561`.
21. **HTTP replay slot** — `replayHandle.Complete(outcome)`.

### Flow 3: ListInboxItems (read path)

Entry: `InboxHandler` (`http/inbox_handler.go:15`).

**Steps** (anchored to `application/read_service.go:237`):

1. BEGIN TX; `setAuthzGUC` `:253`.
2. Single JOIN query `:257` — `approval_instances JOIN approval_stage_instances (status='active') LEFT JOIN documents WHERE asi.eligible_actor_ids @> $2::jsonb AND status='in_progress'`; inline signoff-count subquery for `QuorumProgress`.
3. COMMIT.
4. Separate `CountPendingForActor` call `:326` (separate transaction — snapshot drift risk, T-005).
5. Handler writes `InboxResponse{Items, Total}`.

No tier-2 `authz.Require` call — tier-1 session + JSONB actor containment scoping only.

### Flow 4: RouteAdminService.Create (with idempotency)

Entry: `CreateRouteHandler` (`http/route_admin_handler.go:16`).

**Steps** (anchored to `application/route_admin_service.go`):

1. Handler calls `contracts.Decode`, `CreateRouteRequest.Validate` (`http/route_admin_handler.go:28-40`).
2. `computeCreateRoutePayloadHash(profileCode, name, stages)` → SHA-256 over canonical stage representation (`route_admin_idemp.go:41`).
3. `idempStore.BeginCreateReplay` — reads or creates `idempotency_keys` row; if replay exists, return cached `RouteID` immediately.
4. `createTx`: BEGIN TX; `setAuthzGUC`; `authz.Require(route.admin, "tenant")`.
5. `route.Validate()` `:178`.
6. `INSERT INTO approval_routes … RETURNING id` `:184`. FK violation on `routeProfileFKConstraint` → `ErrRouteProfileUnknown` (422); other FK → wrapped 500.
7. `insertRouteStages` — batched multi-row INSERT `:620`.
8. Governance event `EventType: "route.config.created"` `:221` — raw string literal, not an `EventType` constant (see §10).
9. COMMIT.
10. `committer.Complete(routeID, nil)` — persists replay envelope.

### Flow 5: RunScheduledPublishJob (River worker path)

Entry: `ScheduledPublishWorker.Work` (`jobs/scheduled_publish_job.go:33`) called by `metaldocs-jobs` River runtime.

**Steps** (anchored to `application/scheduler_service.go:44`):

1. `authz.WithBackgroundBypass(ctx)` — signals background origin (ADR 0022 Phase 7).
2. BEGIN TX; `authz.BypassSystem(ctx, tx)` — sets system GUCs.
3. `loadScheduledDocumentState` — `SELECT … FOR UPDATE` on `documents`.
4. Guard checks `:55-69`: document not found → no-op log; `scheduledJobMatchesState` — status='scheduled', generation match, revision match, effective_at match; if any fail → no-op log (stale/duplicate job).
5. Pre-effective-date check `:66` — if `clock.Now()` before `effective_from` → no-op log.
6. `publishScheduledDocumentTx`: if superseded target, `repo.LoadCurrentPublishedHead` + `repo.MarkSuperseded`; `UPDATE documents SET status='published'`; governance event; COMMIT.

---

## 5. Dependencies

### Outbound (imports from outside `approval/`)

| Target | Import path | Purpose |
|---|---|---|
| `iam/authz` | `metaldocs/internal/modules/iam/authz` | `authz.Require`, `authz.WithCapCache`, `authz.BypassSystem`, `authz.WithBackgroundBypass` |
| `iam/domain` | `metaldocs/internal/modules/iam/domain` | `UserIDFromContext`, capability constants (`CapDocumentSubmit`, `CapDocumentSignoff`, etc.) |
| `documents/application` | `metaldocs/internal/modules/documents/application` | `LoadDocumentAreaCode`, `PDFDispatchInvoker`, `ApproverContext`, `PinInvoker` |
| `documents/domain` | `metaldocs/internal/modules/documents/domain` | `ErrDocumentNotFound`, `ErrEffectiveDateMissing` |
| `platform/tenant` | `metaldocs/internal/platform/tenant` | `tenant.FromContext` — extracts tenant from session |
| `platform/idempotency` | `metaldocs/internal/platform/idempotency` | `Store`, `Key`, `ErrConflict`, `ReplayHandle` |
| `platform/problem` | `metaldocs/internal/platform/problem` | `problem.New`, `problem.Write`, `problem.Code` |
| `render/fanout` | (via `PDFOutboxEnqueuer` interface) | PDF outbox enqueue inside approval tx |
| `river` | `github.com/riverqueue/river` | River job worker and enqueuer |
| `pgx/v5/pgconn` | `github.com/jackc/pgx/v5/pgconn` | Postgres error code classification in `MapPgError` |
| `oapi-codegen/runtime` | `github.com/oapi-codegen/runtime` | Codegen server interface wrapper |
| `golang.org/x/crypto/bcrypt` | bcrypt | Password hash comparison in `PasswordReauthProvider` |

### Inbound (verified with grep)

| Consumer | Path | What it imports |
|---|---|---|
| `documents/delivery/http` | `internal/modules/documents/delivery/http/handler.go:18` | `application.SubmitRequest`, `SubmitResult`, `SubmitRevisionForReview` (via `approvalSubmitter` interface) |
| `documents/module.go` | `internal/modules/documents/module.go:10` | `approvalapp.SubmitRequest`, `SubmitResult` |
| `jobs/stuck_instance_watchdog` | `internal/modules/jobs/stuck_instance_watchdog/job.go:12` | `application.CancelService`, `CancelInput`, `SystemCancelInstance` |
| `apps/api/cmd/metaldocs-api` | `main.go:25-28`, `reauth.go:8` | `application`, `http`, `infrastructure`, `jobs`, `repository` packages |
| `apps/jobs/cmd/metaldocs-jobs` | (River worker registration) | `jobs.NewWorkers`, `jobs.NewScheduledPublishEnqueuer` |
| `iam/integration_test.go` | `internal/modules/iam/integration_test.go:16` | `repository.ApprovalRepository` (test only) |
| `documents/delivery/http` finalize wiring test | `internal/modules/documents/delivery/http/finalize_wiring_test.go` | `application.SubmitRevisionForReview` |

---

## 6. Persistence

### Tables owned by this module

| Table | Schema | Primary operations |
|---|---|---|
| `public.approval_routes` | `id, tenant_id, profile_code, name, version, created_at, created_by, active` | INSERT (create), UPDATE (update/deactivate), SELECT (submit, list) |
| `public.approval_route_stages` | `id, route_id, stage_order, name, required_role, required_capability, area_code, quorum, quorum_m, on_eligibility_drift` | INSERT (batched), DELETE+INSERT (update), SELECT (submit) |
| `public.approval_instances` | `id, tenant_id, document_id, route_id, route_version_snapshot, status, submitted_by, submitted_at, completed_at, content_hash_at_submit, idempotency_key` | INSERT (submit), UPDATE status (decision/cancel), SELECT FOR UPDATE (signoff) |
| `public.approval_stage_instances` | `id, approval_instance_id, stage_order, *_snapshot cols, eligible_actor_ids JSONB, effective_denominator, status, opened_at, completed_at, skip_reason` | Bulk INSERT (submit), UPDATE status (decision), SELECT FOR UPDATE (signoff/cancel) |
| `public.approval_signoffs` | `id, approval_instance_id, stage_instance_id, actor_user_id, actor_tenant_id, decision, comment, signed_at, signature_method, signature_payload JSONB, content_hash, actor_display_name_snapshot` | INSERT ON CONFLICT DO NOTHING (signoff), SELECT (quorum eval, SoD, replay) |
| `public.governance_events` | `id, tenant_id, event_type, actor_user_id, resource_type, resource_id, reason, payload_json JSONB, created_at` | INSERT in same tx as every state change |

### Cross-table writes (this module writes outside its owned tables)

| Table | Write operation | Service | Trigger |
|---|---|---|---|
| `public.documents` | `UPDATE status, revision_version` | `submit_service`, `decision_service`, `publish_service`, `cancel_service`, `scheduler_service`, `supersede_service`, `obsolete_service` | Every lifecycle transition |
| `metaldocs.idempotency_keys` | INSERT + UPDATE (via platform store) | `PostgresSignoffIdempStore`, `PostgresRouteAdminIdempStore` | Signoff and route admin mutations |
| `public.pdf_dispatch_outbox` (via interface) | INSERT ON CONFLICT DO NOTHING | `decision_service.go:554` (via `PDFOutboxEnqueuer`) | Final-stage approval |
| River `river_job` table | INSERT (via `RiverScheduledPublishEnqueuer`) | `publish_service.go:293` | Schedule-publish |

### Postgres triggers on approval-owned tables

| Trigger | Table | What it enforces |
|---|---|---|
| `enforce_capability_asserted` | `approval_instances`, `approval_signoffs` (BEFORE INSERT) | GUC `metaldocs.asserted_caps` must contain the required capability; aborts INSERT if missing |
| `enforce_signoff_eligibility_trg` | `approval_signoffs` (BEFORE INSERT) | `eligible_actor_ids @> actor_user_id` on the parent stage; backstop for J1 |
| `enforce_route_immutable` | `approval_route_stages` (BEFORE UPDATE/DELETE on route in use) | Raises `ErrRouteInUse` if the route is referenced by instances |
| Immutable signoff trigger | `approval_signoffs` (BEFORE UPDATE/DELETE) | Prevents modification of any signoff row (audit integrity) |
| `enforce_document_transition` | `public.documents` | Permits `under_review → draft` only when `metaldocs.cancel_in_progress` GUC matches |

### Migration files (forward-only; primary approval schema in archive; incremental in db/migrations)

Core schema is in `db/baseline/0001_current_schema.sql` (curated baseline). Incremental approval-related migrations include: `0212_idempotency_keys_two_phase_schema.sql`, `0216_approval_stage_instances_skip_reason.sql`, `0231_db_hardening_tripwire_and_dead_schema.sql`. Historical migrations in `archive/migrations/` include: `0134_approval_routes`, `0135_approval_instances`, `0138_grants_approval_tables`, `0142b_role_capabilities_v2_enforce` (tripwire), `0144_cancel_state` (cancel GUC), `0145_route_config_immutable_trigger`, `0146_approval_routes_active_column`, `0147_idempotency_keys`, `0180_signoff_eligibility_trigger`, `0194_approval_document_id_rename`, `0195_approval_validate_iam_fks`.

### Query patterns

- **OCC concurrency**: `WHERE revision_version = $N` on `documents` UPDATE (submit, publish, schedule, cancel); `WHERE version = $N` on `approval_routes` UPDATE (route admin).
- **FOR UPDATE locking**: `loadStageInstances` (`postgres_approval_repository.go:563`), `LoadCurrentPublishedHead` (`:382`), `GetDocumentRevisionVersion` (`:404`), `loadScheduledDocumentState` (`:177`).
- **Bulk INSERT**: `InsertStageInstances` builds multi-row VALUES string (`postgres_approval_repository.go:57-98`); `insertRouteStages` same pattern (`route_admin_service.go:620-651`).
- **JSONB containment**: `eligible_actor_ids @> $2::jsonb` in inbox/list queries.

---

## 7. Config & Environment

No `os.Getenv` calls exist inside the module itself. All environment reads are handled in the composition root.

| Env var | Purpose | Consumed at |
|---|---|---|
| `METALDOCS_FANOUT_URL` | Controls whether the freeze/pin path is active; absence forces a no-op freeze invoker | `apps/api/cmd/metaldocs-api/main.go:719` |
| `METALDOCS_JOBS_ENABLED` | Enables River scheduled-publish enqueuer | `main.go` (conditional River client construction) |
| `METALDOCS_JOBS_RIVER_SCHEMA` | River schema override | `internal/platform/config/jobs.go` |
| `METALDOCS_JOBS_TEMPORAL_MAX_WORKERS` | River temporal queue concurrency | `internal/platform/config/jobs.go` |
| `ENABLE_JOB_STUCK_INSTANCE_WATCHDOG` | Enables the stuck-instance cron | `main.go:532` |

The module's service composition takes concrete dependencies (db, repo, emitter, clock, optional enqueuer, optional idempotency stores, optional signature registry) from the composition root — no internal env reads.

---

## 8. Concurrency & Async

- **Transaction ownership**: every mutating service opens its own `db.BeginTx` and owns the lifecycle; repository methods accept `*sql.Tx` — they never open transactions themselves. The pattern is consistent across all 9 services.
- **FOR UPDATE locks**: `loadStageInstances` acquires `SELECT … FOR UPDATE` on all stage rows within the approval-instance load (`postgres_approval_repository.go:563`). This is the J1 lock preventing concurrent re-snapshot of eligible actors during an in-flight signoff.
- **Outbox enqueue in-tx**: `pdfOutbox.Enqueue` (via `PDFOutboxEnqueuer` interface) runs inside the final-stage signoff transaction — atomicity guarantee; consumer processes out-of-band (`decision_service.go:554`).
- **River job enqueue in-tx**: `scheduledPublishEnqueuer.EnqueueScheduledPublishTx` runs inside the `SchedulePublish` transaction (`publish_service.go:293`).
- **Background River worker**: `ScheduledPublishWorker.Work` runs in the `metaldocs-jobs` process, not the API process; communicates exclusively via `documents` table state + River job payload (`jobs/scheduled_publish_job.go:33`).
- **Stuck-instance watchdog**: `jobs/stuck_instance_watchdog` runs as a cron from the API process (`main.go:532`); calls `SystemCancelInstance` which bypasses tier-2 authz but still writes the cancel GUC.
- **Idempotency store**: `PostgresSignoffIdempStore` uses the platform `idempotency.Store` (Postgres-backed); 24h expiry on `idempotency_keys`. The `beginReplay` call is non-transactional relative to the signoff tx; the signoff tx begins after replay check.
- **In-memory rate limiter**: `InMemoryAuthFailureRateLimiter` uses `sync.Mutex`; process-local (no cross-replica coordination) (`infrastructure/signature/password_reauth.go:120`).

---

## 9. Error Handling & Observability

### Error patterns

- **Sentinel errors with `errors.Is`**: primary pattern throughout; see `repository/errors.go:11-24` for the full catalog.
- **Wrapped errors with `%w`**: service errors always wrap with operation context (e.g., `"submit: load route: %w"`).
- **`MapPgError`** (`repository/errors.go:34`): translates `pgconn.PgError` SQLSTATE codes to domain sentinels (23505→`ErrDuplicateSubmission`/`ErrActorAlreadySigned`/`ErrDuplicateRouteProfile`, 23503→`ErrFKViolation`, 42501→`ErrInsufficientPrivilege`, P0001→`ErrRouteInUse`).
- **Trigger-raised errors**: `enforce_route_immutable` raises `ErrRouteInUse:` prefix (SQLSTATE P0001); `enforce_capability_asserted` raises `ErrCapabilityNotAsserted`; both decoded via `MapPgError`.

### HTTP error mapping

`MapErrorToResponse` (`http/errors.go:83`) maps ~40 sentinel/type cases to RFC 9457 `problem.Code` constants. `WriteError` calls `problem.Write(w, prob)`. Server-side errors (5xx) are logged via `slog.Error` with code, status, and underlying error.

### RFC 9457 compliance

Fully adopted (T-001 closed Plan 7). All non-2xx responses use `application/problem+json` via `problem.Write`. The legacy `{error:{code,message}}` envelope is gone from this module.

### Logging

Standard `log/slog` only (`log/slog` package). Notable slog call sites:
- `slog.ErrorContext` for `committer.Fail` errors in route admin (`route_admin_service.go:141,271,418`).
- `slog.Error` in `WriteError` for 5xx responses (`http/errors.go:261`).
- `slog.InfoContext` for scheduler skip reasons (`scheduler_service.go:57,63,68`).
- Two `log.Printf("WARN …")` calls using the legacy `"log"` package in `doc_approval_handler.go:39,195` and `signoff_handler.go` (see §10).

### Observability

No metrics or distributed tracing instrumented. Per `wiki/architecture/system-overview.md` IP-007: observability stack not wired. `log/slog` is the only signal available.

---

## 10. Legacy / Duplication / Smell Flags

- **F-01 — Deprecated `PDFDispatchInvoker` path still compiled (T-004)**
  WHERE: `application/decision_service.go:43,60,64,566-573`
  WHY: `DecisionService` accepts a legacy post-commit `PDFDispatchInvoker` in `NewDecisionService`; the post-commit dispatch block `:566-573` runs when `pdfOutbox == nil`. Production wires `WithPDFOutbox` so the path is inactive, but the dead surface is constructible, keeping a zero-`IdempotencyKey` dispatch path alive for any caller that does not wire the outbox. `git log` shows this predates the outbox path (commit `fa59949ff`).

- **F-02 — `CutoverService` dead code (no production wiring)**
  WHERE: `application/cutover_service.go:1-80`
  WHY: `CutoverService.ValidateLegacyCutoverReady` is a one-time migration preflight for migration 0142, which was applied over a year ago (commit `fa59949ff`). The service is not exported from `Services`, not wired in `main.go`, and has no HTTP route. It is reachable only from test code (`coverage_boost_test.go:ValidateLegacyCutoverReady`). This is confirmed dead code post-cutover.

- **F-03 — Route admin event types use raw string literals instead of `EventType` constants**
  WHERE: `application/route_admin_service.go:223,375,520`
  WHY: `EventType("route.config.created")`, `"route.config.updated"`, `"route.config.deactivated"` are inline string literals; the `EventType` type and constants are defined in `application/events.go:12-18` for submit/signoff/cancel/publish paths but route admin was never migrated to typed constants. Inconsistency risks typos and makes catalogue enumeration incomplete.

- **F-04 — Two GUC helpers with overlapping responsibilities (T-011)**
  WHERE: `application/authz_guc.go:11` (`setAuthzGUC`) and `application/membership_tx.go:22` (`WithMembershipContext`)
  WHY: Both write `metaldocs.tenant_id` + `metaldocs.actor_id`; `WithMembershipContext` additionally sets `ROLE` and `verified_capability`. The submit/decision/cancel/publish paths all use `setAuthzGUC`; `WithMembershipContext` has no callers in this module's production path — it appears to be a predecessor pattern. Two helpers for the same GUC surface invite wrong-helper selection by future contributors.

- **F-05 — `listRoutesQuery` aliases `created_at AS updated_at` (no `updated_at` column)**
  WHERE: `repository/postgres_approval_repository.go:437`
  WHY: `approval_routes` has no `updated_at` column (confirmed in baseline schema line 1729-1738). The query returns `created_at` for both `created_at` and `updated_at` fields. The `Route` struct (`:38`) exposes `UpdatedAt time.Time`. This is silently wrong: consumers would receive the creation timestamp for `UpdatedAt` even after route updates. No migration has added an `updated_at` column.

- **F-06 — `InMemoryAuthFailureRateLimiter` wired in production (process-local, no cross-replica)**
  WHERE: `infrastructure/signature/password_reauth.go:118` (comment: "intended for tests/dev only"); `apps/api/cmd/metaldocs-api/reauth.go:49`
  WHY: The production composition root wires `NewInMemoryAuthFailureRateLimiter()`. The comment in the source explicitly flags this as dev/test-only. In a multi-replica deployment the brute-force limit would not be shared across replicas; each replica maintains independent counters. The `AuthFailureRateLimiter` interface exists and is pluggable, so a Postgres-backed or Redis-backed implementation can be substituted, but as-shipped production uses the in-memory variant.

- **F-07 — Idempotency store wrapper duplication (T-014)**
  WHERE: `infrastructure/postgres_signoff_idemp_store.go` and `infrastructure/postgres_route_admin_idemp_store.go`
  WHY: ~95% of the wrapper code (`ReplayHandle` adapter, `beginReplay` helper, JSON envelope marshal/unmarshal) is duplicated between the two stores. The platform `idempotency.Store` interface is identical; only the envelope struct differs. No generic helper exists yet; deferred to 3rd-store boundary by policy.

- **F-08 — `coverage_boost_test.go` as a separate file**
  WHERE: `application/coverage_boost_test.go:1`
  WHY: A coverage-padding file (goal comment: "push total coverage to ≥90%") creates a parallel test file with no co-location to the production code it exercises. While not incorrect, it signals that the primary test files did not organically reach coverage targets. Several tests in this file exercise implementation details (e.g., `scanSignoffs` via public API) rather than behavior, and they repeat setup from other test files.

- **F-09 — Legacy `"log"` package in two HTTP handler files**
  WHERE: `http/doc_approval_handler.go:7,39,195`; `http/signoff_handler.go:5,15` (import line)
  WHY: The module otherwise uses `log/slog` exclusively. Two warning-level log calls use `log.Printf("WARN …")`, bypassing the structured logging adapter and losing context propagation. This is inconsistent with the module's own observability stance.

- **F-10 — `listRoutesQuery` correlated subquery for `total_count` on every row**
  WHERE: `repository/postgres_approval_repository.go:439`
  WHY: `(SELECT COUNT(*) FROM approval_routes WHERE tenant_id = $1::uuid) AS total_count` is re-evaluated for every row returned by the join. For a small route catalogue this is harmless, but it is a correlated subquery in a loop — an N+1 pattern at the SQL level. The `Total` value is the same for all rows and is pulled from the first row by `scanRouteListRows`.

- **F-11 — `ListPendingForActor` opens a tx, runs instance IDs query, then calls `repo.LoadInstance` in a loop inside the same tx**
  WHERE: `application/read_service.go:172-232`
  WHY: The method performs an N+1 pattern: SELECT DISTINCT IDs, then `LoadInstance` (which itself does `loadStageInstances` + `loadSignoffsForInstance`) per ID. This was acceptable for small inbox counts but is a scalability smell. The `ListInboxItems` method (`read_service.go:237`) uses a single JOIN with a subquery — the two methods serve different consumers but represent inconsistent query strategies.

- **F-12 — `doc_approval_handler.go` line 168 performs eligibility check before `RecordSignoff`**
  WHERE: `http/doc_approval_handler.go:168` (`domain.CheckEligibility` call in HTTP layer)
  WHY: The eligibility check runs in the handler, before the service call, using the instance loaded in a separate preliminary read (`loadActiveInstanceByDocumentForMutation`). The service also performs the same `CheckEligibility` at step J1 inside its transaction. The pre-check in the handler is a short-circuit optimization but creates duplication of the domain rule invocation across two layers. If the active stage changes between the pre-check and the service's load, the service's in-tx check catches it regardless — making the HTTP-layer call partially redundant.

- **F-13 — Route deactivate registered as `POST .../deactivate` but OpenAPI spec and wiki show `DELETE .../routes/{id}`**
  WHERE: `http/router.go:27` (`POST /api/v1/approval/routes/{id}/deactivate`); wiki `approval.md` §5.3 table (row: `DELETE /api/v1/approval/routes/{id}`)
  WHY: The path and method in `router.go` are `POST …/deactivate` but the wiki's route truth table lists `DELETE /api/v1/approval/routes/{id}` with handler `DeactivateRouteHandler`. The comment at `router.go:25-26` explains the design choice (body+headers required), but the wiki table has not been updated. See §11.

---

## 11. Wiki Drift

The following claims in the existing wiki docs no longer match the code as read:

1. **`wiki/modules/approval.md` §5.3 Route Truth Table — Deactivate row method/path**
   - Doc (line ~214-216): `DELETE /api/v1/approval/routes/{id}` with `wrapper.DeactivateApprovalRoute → h.DeactivateRouteHandler`
   - Code reality: `http/router.go:27` registers `POST /api/v1/approval/routes/{id}/deactivate` — the verb is POST and the path has a `/deactivate` suffix, not a bare `{id}` DELETE.

2. **`wiki/modules/approval.md` §5.2 Public surface — `PasswordReauthProvider` path**
   - Doc (line 175): `internal/modules/documents/approval/infra/signature/password_reauth.go:53`
   - Code reality: the directory is `infrastructure/signature/`, not `infra/signature/`. The path `infra/signature/` does not exist in the filesystem.

3. **`wiki/modules/approval.md` §5.1 Container diagram — `infra1` label**
   - Doc (line 127): `Container(infra1, "infra/signature/", "Go", ...)`
   - Code reality: the physical path is `infrastructure/signature/`.

4. **`wiki/modules/approval-tech-debt.md` T-007 — package naming inconsistency**
   - Doc: "Two adjacent package directories with near-identical purpose … `internal/modules/documents/approval/infra/signature/`"
   - Code reality: only `infrastructure/signature/` exists. `infra/signature/` is a stale path reference in the debt entry, not a real directory.

5. **`wiki/modules/approval.md` §8.7 — `ErrActorAlreadySigned` mapped to SoD cross-stage**
   - Doc: `approval-tech-debt` and approval.md describe `domain.ErrActorAlreadySigned` as an SoD cross-stage duplicate
   - Code reality: `domain/sod.go:7` names the error `ErrActorAlreadySigned` but the HTTP error code is `sod.cross_stage_duplicate` (`http/errors.go:119-121`). The domain file has `var ErrAuthorCannotSign` (not `ErrSoDSubmitterCannotSign` as the wiki §6.2 failure table says). Wiki failure table row says `ErrSoDSubmitterCannotSign` — the actual sentinel is `domain.ErrAuthorCannotSign` (`domain/sod.go:6`).

---

## 12. Open Questions

1. **[runtime-unverified]** The `InMemoryAuthFailureRateLimiter` is confirmed wired in production (`reauth.go:49`) but the actual failure-count window behavior under concurrent load and across replicas cannot be verified without a running instance. If MetalDocs runs with multiple API replicas, brute-force protection is per-process, not global.

2. **[runtime-unverified]** `loadStageInstances` acquires `FOR UPDATE` on all stage rows during any `LoadInstance` call, including read paths (e.g., `ReadService.LoadInstance` opens a default read-write tx at `read_service.go:45`). Under concurrent signoff load, every inbox read and instance-get would contend on the same stage row locks. The functional impact (deadlock probability, latency) is unverifiable without observing live lock statistics.

3. **[runtime-unverified]** The `CutoverService` is dead code (F-02), but its `ValidateLegacyCutoverReady` still queries for `documents` with `status IN ('finalized','archived')`. Migration 0142 removed the compat window from the transition trigger. Whether any rows with those legacy statuses remain in the DB cannot be confirmed without a live query.

4. **Genuine unknown — `CutoverService` export boundary**: `CutoverService` is exported from the `application` package but receives no wiring in `Services` nor in `main.go`. It is only reachable from tests. It is unclear whether it was intentionally left as a public utility or should be unexported/deleted.

5. **Genuine unknown — `approval_routes.updated_at` absence**: `listRoutesQuery` aliases `created_at AS updated_at`. The `Route` struct carries `UpdatedAt time.Time`. No migration exists to add a real `updated_at` column. It is unclear whether consumers (frontend route admin) ever observe stale `updated_at` values, or whether the field is intentionally unused.
