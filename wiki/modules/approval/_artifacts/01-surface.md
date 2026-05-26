# Phase 1 — Surface scan (approval module)

> **Note:** Codex sandbox blocked file access. Main agent ran scan via Grep/Read. Facts only.

Module path: `internal/modules/documents/approval/`

## 1. File tree

```
internal/modules/documents/approval/
├── application/                          (use-case orchestration; 16 .go + 16 _test.go)
│   ├── authz_guc.go                      — `setAuthzGUC` (SET LOCAL metaldocs.tenant_id / actor_id)
│   ├── cancel_service.go                 — CancelService (cancel approval instance + revert doc to draft)
│   ├── content_hash.go                   — ComputeContentHash (signoff payload canonicalization)
│   ├── cutover_service.go                — CutoverService (legacy-state purge guard; ErrLegacyDocumentsRemain)
│   ├── decision_service.go               — DecisionService.RecordSignoff (eligibility + SoD + quorum + freeze)
│   ├── events.go                         — GovernanceEvent + EventEmitter (sqlEmitter + MemoryEmitter)
│   ├── idempotency.go                    — ComputeIdempotencyKey (input-hash key for replay store)
│   ├── membership_tx.go                  — WithMembershipContext (sets actor/tenant GUC pre-tx)
│   ├── obsolete_service.go               — ObsoleteService.MarkObsolete (published|superseded → obsolete)
│   ├── publish_service.go                — PublishService.PublishApproved + SchedulePublish
│   ├── read_service.go                   — ReadService (LoadInstance, ListInbox, CountPending)
│   ├── route_admin_service.go            — RouteAdminService (Create/Update/Deactivate routes)
│   ├── scheduler_service.go              — SchedulerService.RunScheduledPublishJob
│   ├── services.go                       — Services struct (composition root for module); Clock; ValidateEventPayload
│   ├── submit_service.go                 — SubmitService.SubmitRevisionForReview (draft → under_review + instance create)
│   └── supersede_service.go              — SupersedeService.PublishSuperseding (publish-as-superseder atomic)
├── jobs/                                 (dedicated River worker package for scheduled publish)
│   ├── scheduled_publish_args.go         — ScheduledPublishArgs (River payload; kind `scheduled_publish_cutover`)
│   └── scheduled_publish_job.go          — ScheduledPublishWorker + RiverScheduledPublishEnqueuer + NewWorkers
├── domain/                               (pure rules; 8 .go + 9 _test.go)
│   ├── drift.go                          — ApplyEligibilityDrift (DriftPolicy reconciliation)
│   ├── eligibility.go                    — CheckEligibility + ErrActorNotEligible (J1 sentinel)
│   ├── instance.go                       — Instance, StageInstance, InstanceStatus, StageStatus, AdvanceStage, RejectHere, SkipStage, Cancel, BumpRevisionVersion
│   ├── quorum.go                         — QuorumOutcome, ComputeEffectiveDenominator, EvaluateQuorum
│   ├── route.go                          — Route, Stage, QuorumPolicy, DriftPolicy; Route.Validate
│   ├── signoff.go                        — Signoff (private fields w/ getters), SignoffParams, NewSignoff, MarshalJSON
│   ├── sod.go                            — CheckSoD pure rule (`ErrSoDSubmitterCannotSign` etc.)
│   └── state.go                          — DocState enum, IsLegalTransition, StateFromString, ErrLegacyStateRejected
├── http/                                 (HTTP delivery; 16 .go + 16 _test.go inc. contracts)
│   ├── cancel_handler.go                 — CancelHandler
│   ├── doc_approval_handler.go           — GetInstanceByDocumentHandler, SignoffByDocumentHandler (idempotency replay), CancelByDocumentHandler
│   ├── errors.go                         — MapErrorToResponse, WriteError, WriteJSON, looksLikeValidationError (E4 gap site)
│   ├── get_instance_handler.go           — GetInstanceHandler + mapInstanceResponse
│   ├── handler.go                        — Handler struct + injected service ifaces; signoffIdempStore iface
│   ├── inbox_handler.go                  — InboxHandler + parseInboxLimit/Offset
│   ├── obsolete_handler.go               — ObsoleteHandler
│   ├── publish_handler.go                — PublishHandler + SchedulePublishHandler
│   ├── route_admin_handler.go            — CreateRouteHandler, UpdateRouteHandler, DeactivateRouteHandler, ListRoutesHandler
│   ├── router.go                         — RegisterRoutes (15 routes onto *http.ServeMux)
│   ├── signoff_handler.go                — SignoffHandler (instance/stage path) + signoffOutcome
│   ├── submit_handler.go                 — SubmitHandler (POST /api/v1/documents/{id}/submit)
│   ├── supersede_handler.go              — SupersedeHandler
│   └── contracts/                        — request/response DTOs + Validate(): cancel, errors, instance_read, obsolete, publish, route, signoff, strictjson, submit, supersede
├── infra/signature/                      (pluggable signature providers)
│   ├── password_reauth.go                — PasswordReauthProvider (rate limit + janitor)
│   └── provider.go                       — Provider iface, Registry, ErrUnknownSignatureMethod
├── infrastructure/                       (one shim file, naming inconsistent with infra/ above)
│   └── postgres_signoff_idemp_store.go   — PostgresSignoffIdempStore (CheckReplay/RecordReplay against metaldocs.idempotency_keys)
└── repository/                           (Postgres reads + writes; 3 .go)
    ├── approval_repository.go            — ApprovalRepository iface, SignoffInsertResult, ScheduledPublishRow
    ├── errors.go                         — MapPgError, MapHints (pq error code translation)
    └── postgres_approval_repository.go   — postgresApprovalRepository (NewPostgresApprovalRepository); 11 methods
```

Test files: 43 total `_test.go` files across application/domain/http/repository/infra. Largest concentration: `application/` (16) and `domain/` (9).

## 2. Public surface (top-level exported symbols only)

### domain/

| File:line | Kind | Name | Signature / receiver | Doc first line |
|---|---|---|---|---|
| `drift.go:4` | type | `DriftResult` | struct | (undocumented) |
| `drift.go:12` | func | `ApplyEligibilityDrift` | `func(stage StageInstance, currentEligible []string) DriftResult` | (undocumented) |
| `eligibility.go:7` | var | `ErrActorNotEligible` | `error` | sentinel for J1 |
| `eligibility.go:11` | func | `CheckEligibility` | `func(actorUserID string, eligibleActorIDs []string) error` | (undocumented) |
| `instance.go:16` | type | `InstanceStatus` | `string` | (undocumented) |
| `instance.go:26` | type | `StageStatus` | `string` | (undocumented) |
| `instance.go:37` | type | `StageInstance` | struct | (undocumented) |
| `instance.go:58` | type | `Instance` | struct | (undocumented) |
| `instance.go:75` | method | `(*Instance).Active` | `() *StageInstance` | (undocumented) |
| `instance.go:86` | method | `(*Instance).AdvanceStage` | `() error` | (undocumented) |
| `instance.go:118` | method | `(*Instance).RejectHere` | `(reason string) error` | (undocumented) |
| `instance.go:141` | method | `(*Instance).SkipStage` | `(reason string) error` | (undocumented) |
| `instance.go:182` | method | `(*Instance).BumpRevisionVersion` | `(next int) error` | (undocumented) |
| `instance.go:191` | method | `(*Instance).Cancel` | `(reason string) error` | (undocumented) |
| `quorum.go:4` | type | `QuorumOutcome` | `string` | (undocumented) |
| `quorum.go:14` | func | `ComputeEffectiveDenominator` | `(StageInstance, []string) int` | (undocumented) |
| `quorum.go:33` | func | `EvaluateQuorum` | `(StageInstance, []Signoff, []Signoff, int) QuorumOutcome` | (undocumented) |
| `route.go:9` | type | `QuorumPolicy` | `string` | (undocumented) |
| `route.go:18` | type | `DriftPolicy` | `string` | (undocumented) |
| `route.go:27` | type | `Stage` | struct | (undocumented) |
| `route.go:39` | type | `Route` | struct | (undocumented) |
| `route.go:48` | method | `Route.Validate` | `() error` | (undocumented) |
| `signoff.go:14` | type | `Decision` | `string` | (undocumented) |
| `signoff.go:22` | type | `Signoff` | struct (private fields) | (undocumented) |
| `signoff.go:38..49` | methods | getters (`ID/ApprovalInstanceID/StageInstanceID/ActorUserID/ActorTenantID/Decision/Comment/SignedAt/SignatureMethod/SignaturePayload/ContentHash/ActorDisplayNameSnapshot`) | `() T` | (undocumented) |
| `signoff.go:52` | type | `SignoffParams` | struct | (undocumented) |
| `signoff.go:69` | func | `NewSignoff` | `(SignoffParams) (*Signoff, error)` | (undocumented) |
| `signoff.go:115` | method | `(*Signoff).MarshalJSON` | `() ([]byte, error)` | (undocumented) |
| `sod.go:15` | func | `CheckSoD` | `(authorUserID, actorUserID string, prior []Signoff) error` | (undocumented) |
| `state.go:6` | var | `ErrLegacyStateRejected` | `error` | (undocumented) |
| `state.go:9` | type | `DocState` | `string` | (undocumented) |
| `state.go:23` | func | `AllStates` | `() []DocState` | (undocumented) |
| `state.go:35` | func | `StateFromString` | `(string) (DocState, error)` | (undocumented) |
| `state.go:61` | func | `IsLegalTransition` | `(from, to DocState) bool` | (undocumented) |

### application/

| File:line | Kind | Name | Signature | Doc first line |
|---|---|---|---|---|
| `cancel_service.go:17` | type | `CancelService` | struct | (undocumented) |
| `cancel_service.go:24` | var | `ErrReasonRequired` | error | (undocumented) |
| `cancel_service.go:27` | type | `CancelInput` | struct | (undocumented) |
| `cancel_service.go:40` | type | `CancelResult` | struct | (undocumented) |
| `cancel_service.go:47` | method | `(*CancelService).CancelInstance` | `(ctx, *sql.DB, CancelInput) (CancelResult, error)` | (undocumented) |
| `content_hash.go:15` | var | `ErrFloatInFormData` | error | (undocumented) |
| `content_hash.go:18` | type | `ContentHashInput` | struct | (undocumented) |
| `content_hash.go:35` | func | `ComputeContentHash` | `(ContentHashInput) (string, error)` | (undocumented) |
| `cutover_service.go:14` | var | `ErrLegacyDocumentsRemain` | error | (undocumented) |
| `cutover_service.go:22` | type | `CutoverService` | struct | (undocumented) |
| `cutover_service.go:28` | func | `NewCutoverService` | `(EventEmitter, Clock) *CutoverService` | (undocumented) |
| `cutover_service.go:39` | method | `(*CutoverService).ValidateLegacyCutoverReady` | `(ctx, *sql.DB) error` | (undocumented) |
| `decision_service.go:19` | iface | `FreezeInvoker` | one method | (undocumented) |
| `decision_service.go:23` | iface | `PDFDispatchInvoker` | one method | (undocumented) |
| `decision_service.go:28` | iface | `PDFOutboxEnqueuer` | one method | (undocumented) |
| `decision_service.go:33` | type | `DecisionService` | struct | (undocumented) |
| `decision_service.go:43` | func | `NewDecisionService` | `(repo, EventEmitter, FreezeInvoker, PDFDispatchInvoker, Clock) *DecisionService` | (undocumented) |
| `decision_service.go:60` | method | `(*DecisionService).WithPDFOutbox` | `(PDFOutboxEnqueuer) *DecisionService` | (undocumented) |
| `decision_service.go:66` | type | `SignoffRequest` | struct (`Decision domain.Decision`) | (undocumented) |
| `decision_service.go:80` | type | `SignoffResult` | struct | (undocumented) |
| `decision_service.go:88` | method | `(*DecisionService).RecordSignoff` | `(ctx, *sql.DB, SignoffRequest) (SignoffResult, error)` | (undocumented) |
| `events.go:11` | type | `EventType` | `string` constants (`approval.instance_cancelled`, `document_published`, `publish_scheduled`, `signoff.rejected`, `signoff_recorded`) | (undocumented) |
| `events.go:19` | type | `GovernanceEvent` | struct (`EventType EventType`, `OccurredAt time.Time`) | (undocumented) |
| `events.go:32` | iface | `EventEmitter` | `Emit(ctx, *sql.Tx, GovernanceEvent) error` | (undocumented) |
| `events.go:32` | func | `NewSQLEmitter` | `() EventEmitter` | (undocumented) |
| `events.go:52` | type | `MemoryEmitter` | struct (test seam) | (undocumented) |
| `idempotency.go:12` | type | `IdempotencyInput` | struct | (undocumented) |
| `idempotency.go:25` | func | `ComputeIdempotencyKey` | `(IdempotencyInput) string` | (undocumented) |
| `membership_tx.go:10` | var | `ErrNoActor` | error | (undocumented) |
| `membership_tx.go:22` | func | `WithMembershipContext` | `(ctx, *sql.Tx, tenant, actor) error` | sets GUC for tier-2 authz |
| `obsolete_service.go:15` | type | `ObsoleteService` | struct | (undocumented) |
| `obsolete_service.go:23` | var | `ErrInvalidObsoleteSource` | error | (undocumented) |
| `obsolete_service.go:26` | type | `MarkObsoleteRequest` | struct | (undocumented) |
| `obsolete_service.go:35` | type | `MarkObsoleteResult` | struct | (undocumented) |
| `obsolete_service.go:42` | method | `(*ObsoleteService).MarkObsolete` | `(ctx, *sql.DB, MarkObsoleteRequest) (MarkObsoleteResult, error)` | (undocumented) |
| `publish_service.go:17` | type | `PublishService` | struct | (undocumented) |
| `publish_service.go:25` | var | `ErrInstanceNotApproved` | error | (undocumented) |
| `publish_service.go:28` | type | `PublishRequest` | struct | (undocumented) |
| `publish_service.go:35` | type | `PublishResult` | struct | (undocumented) |
| `publish_service.go:44` | method | `(*PublishService).PublishApproved` | `(ctx, *sql.DB, PublishRequest) (PublishResult, error)` | (undocumented) |
| `publish_service.go:146` | var | `ErrEffectiveDateInPast` | error | (undocumented) |
| `publish_service.go:149` | type | `SchedulePublishRequest` | struct | (undocumented) |
| `publish_service.go:157` | type | `SchedulePublishResult` | struct | (undocumented) |
| `publish_service.go:166` | method | `(*PublishService).SchedulePublish` | `(ctx, *sql.DB, SchedulePublishRequest) (SchedulePublishResult, error)` | (undocumented) |
| `read_service.go:16` | type | `InboxView` | struct | (undocumented) |
| `read_service.go:29` | type | `ReadService` | struct | (undocumented) |
| `read_service.go:38` | method | `(*ReadService).LoadInstance` | `(ctx, *sql.DB, tenant, id) (*Instance, error)` | reads actor from session context |
| `read_service.go:63` | method | `(*ReadService).LoadActiveInstanceByDocument` | `(ctx, *sql.DB, tenant, doc) (*Instance, error)` | (undocumented) |
| `read_service.go:88` | method | `(*ReadService).ListPendingForActor` | `(ctx, *sql.DB, tenant, actor, area, limit, offset) ([]Instance, error)` | (undocumented) |
| `read_service.go:153` | method | `(*ReadService).ListInboxItems` | `(ctx, *sql.DB, tenant, actor, area, limit, offset) ([]InboxView, error)` | (undocumented) |
| `read_service.go:224` | method | `(*ReadService).CountPendingForActor` | `(ctx, *sql.DB, tenant, actor, area) (int, error)` | (undocumented) |
| `route_admin_service.go:16` | type | `RouteAdminService` | struct | (undocumented) |
| `route_admin_service.go:23` | var | `ErrRouteNotFound` | error | (undocumented) |
| `route_admin_service.go:26..67` | types/methods | `CreateRouteInput/Result`, `UpdateRouteInput/Result`, `DeactivateRouteInput/Result`, `Create`, `Update`, `Deactivate` | per file | (undocumented) |
| `scheduler_service.go:14` | type | `SchedulerService` | struct | (undocumented) |
| `scheduler_service.go:34` | method | `(*SchedulerService).RunScheduledPublishJob` | `(ctx, *sql.DB, ScheduledPublishJobInput) error` | dedicated jobs-runtime path; stale, mismatched, or early jobs no-op before publish |
| `services.go:12` | iface | `Clock` | `Now() time.Time` | (undocumented) |
| `services.go:17` | type | `RealClock` | struct | (undocumented) |
| `services.go:29` | type | `Services` | struct (composition root) | (undocumented) |
| `services.go:39` | type | `ScheduledPublishJobInput` | struct | River scheduled-publish payload boundary shared by API enqueue + jobs runtime |
| `services.go:47` | iface | `ScheduledPublishEnqueuer` | `EnqueueScheduledPublishTx(context.Context, *sql.Tx, ScheduledPublishJobInput) error` | transactional enqueue seam |
| `services.go:43` | func | `NewServices` | `(repo, EventEmitter, Clock) *Services` | (undocumented) |
| `services.go:61` | method | `(*Services).WithScheduledPublishEnqueuer` | `(ScheduledPublishEnqueuer) *Services` | wires the River enqueue seam into PublishService |
| `services.go:61` | func | `ValidateEventPayload` | `(map[string]any) error` | now returns `ErrFloatInFormData` for float-bearing payloads |
| `submit_service.go:19` | type | `SubmitService` | struct | (undocumented) |
| `submit_service.go:26` | type | `SubmitRequest` | struct | (undocumented) |
| `submit_service.go:36` | type | `SubmitResult` | struct | (undocumented) |
| `submit_service.go:43` | method | `(*SubmitService).SubmitRevisionForReview` | `(ctx, *sql.DB, SubmitRequest) (SubmitResult, error)` | (undocumented) |
| `supersede_service.go:14` | type | `SupersedeService` | struct | (undocumented) |
| `supersede_service.go:21` | type | `SupersedeRequest` | struct | (undocumented) |
| `supersede_service.go:31` | type | `SupersedeResult` | struct | (undocumented) |
| `supersede_service.go:40` | method | `(*SupersedeService).PublishSuperseding` | `(ctx, *sql.DB, SupersedeRequest) (SupersedeResult, error)` | (undocumented) |

### http/ (handler methods + helpers)

| File:line | Kind | Name |
|---|---|---|
| `handler.go:19,23,27,35,51` | private ifaces | `submitService`, `decisionService`, `readService`, `routeAdminService`, `signoffIdempStore` (note: comment line at :49 is "signoffIdempStore backs idempotent replay") |
| `handler.go:56` | type | `Handler` |
| `handler.go:66` | func | `NewHandler(*application.Services, *sql.DB, signoffIdempStore) *Handler` |
| `cancel_handler.go:21` | method | `(*Handler).CancelHandler` |
| `doc_approval_handler.go:16,51,146` | methods | `GetInstanceByDocumentHandler`, `SignoffByDocumentHandler`, `CancelByDocumentHandler` |
| `errors.go:22,147,153,181` | funcs | `MapErrorToResponse`, `WriteError`, `WriteJSON`, `looksLikeValidationError` |
| `get_instance_handler.go:14` | method | `GetInstanceHandler` |
| `inbox_handler.go:15` | method | `InboxHandler` |
| `obsolete_handler.go:21` | method | `ObsoleteHandler` |
| `publish_handler.go:30,69` | methods | `PublishHandler`, `SchedulePublishHandler` |
| `route_admin_handler.go:16,59,108,144` | methods | `CreateRouteHandler`, `UpdateRouteHandler`, `DeactivateRouteHandler`, `ListRoutesHandler` |
| `router.go:6` | method | `(*Handler).RegisterRoutes(*http.ServeMux)` |
| `signoff_handler.go:18` | method | `SignoffHandler` |
| `submit_handler.go:14` | method | `SubmitHandler` |
| `supersede_handler.go:21` | method | `SupersedeHandler` |
| `contracts/*.go` | DTO types | `CancelRequest/Response`, `ErrorResponse/Body`, `InstanceResponse`, `StageInstance`, `SignoffRecord`, `InboxItem`, `InboxResponse`, `ObsoleteRequest/Response`, `PublishRequest`, `SchedulePublishRequest`, `PublishResponse`, `StageRequest`, `CreateRouteRequest`, `UpdateRouteRequest`, `RouteResponse`, `StageResponse`, `ListStageItem`, `ListRouteItem`, `SignoffRequest/Response`, `SubmitRequest/Response`, `SupersedeRequest/Response`; helpers `Decode` (strictjson), `validateUUID`, `validateSHA256Hex`, `validateRequired`, `validateStages` |

### repository/

| File:line | Kind | Name |
|---|---|---|
| `approval_repository.go:12` | type | `SignoffInsertResult` (doc commented) |
| `approval_repository.go:18` | type | `ScheduledPublishRow` (doc commented) |
| `approval_repository.go:28` | iface | `ApprovalRepository` (doc commented; scheduling/publish persistence surface) |
| `errors.go:28,34` | type/func | `MapHints`, `MapPgError` |
| `postgres_approval_repository.go:20` | func | `NewPostgresApprovalRepository(*sql.DB) ApprovalRepository` |
| `postgres_approval_repository.go:32..514` | methods (private receiver) | `InsertInstance`, `InsertStageInstances`, `InsertSignoff`, `LoadSignoffByActor`, `LoadInstance`, `LoadActiveInstanceByDocument`, `ValidateScheduledSupersedeTarget`, `LoadCurrentPublishedHeadForDocument`, `LoadCurrentPublishedHead`, `MarkSuperseded`, `UpdateStageStatus`, `UpdateInstanceStatus` (+ private helpers `loadStageInstances`, `loadSignoffsForInstance`) |

### infrastructure/

| File:line | Kind | Name |
|---|---|---|
| `postgres_signoff_idemp_store.go:14` | type | `PostgresSignoffIdempStore` |
| `postgres_signoff_idemp_store.go:18,25,42` | funcs | `NewPostgresSignoffIdempStore`, `CheckReplay`, `RecordReplay` |

### jobs/

| File:line | Kind | Name |
|---|---|---|
| `scheduled_publish_args.go:5` | type | `ScheduledPublishArgs` |
| `scheduled_publish_args.go:13` | method | `(ScheduledPublishArgs).Kind` |
| `scheduled_publish_job.go:11` | type | `ScheduledPublishWorker` |
| `scheduled_publish_job.go:27` | type | `RiverScheduledPublishEnqueuer` |
| `scheduled_publish_job.go:31` | method | `(*RiverScheduledPublishEnqueuer).EnqueueScheduledPublishTx` |
| `scheduled_publish_job.go:44` | func | `NewWorkers` |
| `scheduled_publish_job.go:52` | func | `NewScheduledPublishEnqueuer` |

### infra/signature/

| File:line | Kind | Name |
|---|---|---|
| `provider.go:11` | var | `ErrUnknownSignatureMethod` |
| `provider.go:14,22,30,36,41,46,51` | types/funcs | `SignRequest`, `SignatureResult`, `Provider` iface, `Registry`, `NewRegistry`, `Register`, `Get` |
| `password_reauth.go:20,25,43,53,65,67` | types/funcs | `IamUserReader` iface, `EventEmitterStub` iface, `PasswordReauthProvider`, `NewPasswordReauthProvider`, `Method`, `Sign` (+ private rate-limit helpers) |

## 3. HTTP operations

Source: `internal/modules/documents/approval/http/router.go:6-30`. All wired onto Go 1.22 `*http.ServeMux` with method-prefixed pattern. Composition root: `apps/api/cmd/metaldocs-api/main.go:331-332` (`approvalHandler := approvalhttp.NewHandler(approvalServices, deps.SQLDB, signoffIdempStore); approvalHandler.RegisterRoutes(mux)`).

| Method | Path | Handler | Source |
|---|---|---|---|
| POST | `/api/v1/documents/{id}/submit` | `Handler.SubmitHandler` | `submit_handler.go:14` |
| POST | `/api/v1/approval/instances/{instance_id}/stages/{stage_id}/signoffs` | `Handler.SignoffHandler` | `signoff_handler.go:18` |
| POST | `/api/v1/documents/{id}/publish` | `Handler.PublishHandler` | `publish_handler.go:30` |
| POST | `/api/v1/documents/{id}/schedule-publish` | `Handler.SchedulePublishHandler` | `publish_handler.go:69` |
| POST | `/api/v1/documents/{id}/supersede` | `Handler.SupersedeHandler` | `supersede_handler.go:21` |
| POST | `/api/v1/documents/{id}/obsolete` | `Handler.ObsoleteHandler` | `obsolete_handler.go:21` |
| POST | `/api/v1/approval/instances/{instance_id}/cancel` | `Handler.CancelHandler` | `cancel_handler.go:21` |
| GET | `/api/v1/approval/instances/{instance_id}` | `Handler.GetInstanceHandler` | `get_instance_handler.go:14` |
| GET | `/api/v1/documents/{id}/approval-instance` | `Handler.GetInstanceByDocumentHandler` | `doc_approval_handler.go:16` |
| GET | `/api/v1/approval/inbox` | `Handler.InboxHandler` | `inbox_handler.go:15` |
| POST | `/api/v1/documents/{id}/signoff` | `Handler.SignoffByDocumentHandler` | `doc_approval_handler.go:51` |
| POST | `/api/v1/documents/{id}/cancel` | `Handler.CancelByDocumentHandler` | `doc_approval_handler.go:146` |
| POST | `/api/v1/approval/routes` | `Handler.CreateRouteHandler` | `route_admin_handler.go:16` |
| PUT | `/api/v1/approval/routes/{id}` | `Handler.UpdateRouteHandler` | `route_admin_handler.go:59` |
| DELETE | `/api/v1/approval/routes/{id}` | `Handler.DeactivateRouteHandler` | `route_admin_handler.go:108` |
| GET | `/api/v1/approval/routes` | `Handler.ListRoutesHandler` | `route_admin_handler.go:144` |

Total: **16 HTTP operations** (6 v2 doc-action POSTs, 2 instance-scoped, 2 reads, 1 inbox, 1 doc-scoped signoff, 1 doc-scoped cancel, 4 route admin).

## 4. Migration list

Module owns no `migrations/` subdir. Migrations affecting approval-owned tables (top-level `migrations/`):

| Filename | Verb summary | Tables touched |
|---|---|---|
| `0016_init_workflow_approvals.sql` | CREATE | `approval_routes`, `approval_route_stages`, `approval_route_steps`, `approval_instances`, `approval_stage_instances`, `approval_signoffs` (initial schema) |
| `0017_grant_workflow_approvals_privileges.sql` | GRANT | (privileges to `metaldocs_app`) |
| `0134_approval_routes.sql` | ALTER | `approval_routes` (schema additions) |
| `0135_approval_instances.sql` | ALTER | `approval_instances` (schema additions) |
| `0138_grants_approval_tables.sql` | GRANT | further privilege grants |
| `0140_revision_version_and_inbox_index.sql` | ALTER + CREATE INDEX | `approval_instances` (revision_version), inbox query index |
| `0141_governance_events_dedupe_signoff_uniqueness.sql` | UNIQUE INDEX | `governance_events` (signoff dedupe) + `approval_signoffs` |
| `0142a_role_capabilities_v2_additive.sql` | INSERT | `role_capabilities` seed (additive) |
| `0142b_role_capabilities_v2_enforce.sql` | TRIGGER | attaches `enforce_capability_asserted` to `approval_instances` + `approval_signoffs` (lines 200-209); installs `metaldocs.asserted_caps` GUC enforcer |
| `0142b_down.sql` | DROP TRIGGER | rollback companion |
| `0144_cancel_state.sql` | ALTER + TRIGGER | adds cancel state columns + cancel-flow GUC bypass (`metaldocs.cancel_in_progress`) |
| `0145_route_config_immutable_trigger.sql` | TRIGGER | enforce route stages immutable post-activation |
| `0146_approval_routes_active_column.sql` | ALTER | `approval_routes.active` boolean |
| `0151_seed_dev_tenant_approval_data.sql` | INSERT | dev seed (approval routes + steps for dev tenant) |
| `0160_grant_metaldocs_app_schema_objects.sql` | GRANT | `metaldocs_app` SELECT/INSERT/UPDATE on `metaldocs.idempotency_keys` (signoff replay store) |
| `0167_documents_bridge_and_state_columns.sql` | ALTER | `documents` (state columns used by approval state-transition flows) — touched but not owned |
| `0173_signoff_actor_displayname_snapshot.sql` | ALTER | `approval_signoffs.actor_display_name_snapshot` |
| `0180_signoff_eligibility_trigger.sql` | TRIGGER | `enforce_signoff_eligibility_trg` BEFORE INSERT on `approval_signoffs`; J1 defense in depth |

**Total: 18 migrations.** Tables read but not owned: `documents`, `controlled_documents`, `user_process_areas`, `iam_users`, `metaldocs.idempotency_keys`, `governance_events`, `metaldocs.pdf_dispatch_outbox`. Persistence map (Phase 4) will detail.

---

**Summary:** exported symbols ≈ 110 (private getters + DTOs counted as families) · HTTP operations: 16 · migrations: 18.
