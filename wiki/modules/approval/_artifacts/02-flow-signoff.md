## Task

For operation `SignoffByDocument` (HTTP `POST /api/v1/documents/{id}/signoffs`) in module `internal/modules/documents/approval`, produce an artifact at `wiki/modules/approval/_artifacts/02-flow-signoff.md` tracing the call end-to-end.

### 1. Entry point

| Layer | Symbol | File:line |
|---|---|---|
| OpenAPI op | NOT FOUND (`/api/v1/documents/{id}/signoffs`) | `api/openapi/spec2.yaml` searched via `rg -n "documents/\\{id\\}/signoff|documents/\\{id\\}/signoffs"` |
| Generated server stub | NOT FOUND (`ServerInterface` method for `/api/v1/documents/{id}/signoffs`) | `internal/**/api/*.gen.go` searched via `rg -n "SignoffByDocument|signoff"` |
| Handler | `Handler.SignoffByDocumentHandler` | `internal/modules/documents/approval/http/doc_approval_handler.go:51` |

### 2. Call chain

1. `internal/modules/documents/approval/http/doc_approval_handler.go:51` `SignoffByDocumentHandler` — HTTP entrypoint for document-level signoff (`/signoff` in code comment).
   -> calls: `internal/modules/documents/approval/http/doc_approval_handler.go:68` `contracts.Decode`
2. `internal/modules/documents/approval/http/doc_approval_handler.go:68` `contracts.Decode` — request bind/parse into `docSignoffRequest`; decision/password checks run in same handler.
   -> calls: `internal/modules/documents/approval/http/doc_approval_handler.go:88` `h.idempStore.CheckReplay`
3. `internal/modules/documents/approval/http/doc_approval_handler.go:88` `idempStore.CheckReplay` — HTTP-layer idempotency lookup before loading instance.
   -> calls: `internal/modules/documents/approval/infrastructure/postgres_signoff_idemp_store.go:25` `PostgresSignoffIdempStore.CheckReplay`
4. `internal/modules/documents/approval/infrastructure/postgres_signoff_idemp_store.go:29` `s.inner.CheckReplay` — idempotency store backend lookup.
   -> calls: `internal/modules/documents/approval/http/doc_approval_handler.go:99` `h.readSvc.LoadActiveInstanceByDocument`
5. `internal/modules/documents/approval/application/read_service.go:63` `ReadService.LoadActiveInstanceByDocument` — opens read-only tx and loads active instance for document.
   -> calls: `internal/modules/documents/approval/application/read_service.go:64` `db.BeginTx`
6. `internal/modules/documents/approval/application/read_service.go:70` `s.repo.LoadActiveInstanceByDocument` — repo read by document id.
   -> calls: `internal/modules/documents/approval/repository/postgres_approval_repository.go:265` `LoadActiveInstanceByDocument`
7. `internal/modules/documents/approval/http/doc_approval_handler.go:115` `h.decisionSvc.RecordSignoff` — enters core write flow.
   -> calls: `internal/modules/documents/approval/application/decision_service.go:88` `RecordSignoff`
8. `internal/modules/documents/approval/application/decision_service.go:106` `db.BeginTx` — write transaction boundary.
   -> calls: `internal/modules/documents/approval/application/decision_service.go:111` `setAuthzGUC`
9. `internal/modules/documents/approval/application/authz_guc.go:11` `setAuthzGUC` — writes transaction-local GUCs for authz context.
   -> calls: `internal/modules/documents/approval/application/authz_guc.go:12` `set_config('metaldocs.tenant_id', ...)`; `internal/modules/documents/approval/application/authz_guc.go:15` `set_config('metaldocs.actor_id', ...)`
10. `internal/modules/documents/approval/application/decision_service.go:117` `s.repo.LoadInstance` — loads approval instance; stage rows loaded with `FOR UPDATE`.
   -> calls: `internal/modules/documents/approval/repository/postgres_approval_repository.go:229` `LoadInstance`
11. `internal/modules/documents/approval/repository/postgres_approval_repository.go:314` `loadStageInstances` query comment and `:316` `FOR UPDATE` — lock stage rows.
   -> calls: `internal/modules/documents/approval/application/decision_service.go:136` `loadDocumentAreaCode`
12. `internal/modules/documents/approval/application/submit_service.go:276` `loadDocumentAreaCode` — loads document area used by authz.
   -> calls: `internal/modules/documents/approval/application/decision_service.go:141` `authz.Require`
13. `internal/modules/documents/approval/application/decision_service.go:141` `authz.Require(ctx, tx, "doc.signoff", areaCode)` — authorization gate.
   -> calls: `internal/modules/documents/approval/application/decision_service.go:159` `domain.CheckEligibility`
14. `internal/modules/documents/approval/application/decision_service.go:159` `domain.CheckEligibility` — eligibility gate against `EligibleActorIDs` snapshot.
   -> calls: `internal/modules/documents/approval/application/decision_service.go:178` `domain.CheckSoD`
15. `internal/modules/documents/approval/application/decision_service.go:178` `domain.CheckSoD` — separation-of-duties check.
   -> calls: `internal/modules/documents/approval/application/decision_service.go:217` `s.repo.InsertSignoff`
16. `internal/modules/documents/approval/repository/postgres_approval_repository.go:114` `InsertSignoff` — insert into `approval_signoffs` with `ON CONFLICT` replay detection.
   -> calls: `internal/modules/documents/approval/repository/postgres_approval_repository.go:127` `INSERT INTO approval_signoffs`; `:132` `ON CONFLICT (approval_instance_id, actor_user_id) DO NOTHING`; `:159` `LoadSignoffByActor` field-compare replay
17. `migrations/0135_approval_instances.sql:125` `trg_signoff_tenant_consistent` + `:151` `trg_signoff_sod`; `migrations/0180_signoff_eligibility_trigger.sql:26` `enforce_signoff_eligibility_trg` — `BEFORE INSERT ON approval_signoffs` tripwire trigger pairing with step 16 insert.
   -> calls: `internal/modules/documents/approval/application/decision_service.go:234` `loadStageSignoffs`
18. `internal/modules/documents/approval/application/decision_service.go:234` `loadStageSignoffs` and `:251` `domain.EvaluateQuorum` — quorum evaluation.
   -> calls: approval/rejection branches below
19. `internal/modules/documents/approval/application/decision_service.go:259` `case domain.QuorumApprovedStage` — approval branch.
   -> calls: `:275` `UpdateInstanceStatus(... InstanceApproved ...)`; `:284` `freezeInvoker.Freeze`; `:293` `UPDATE documents ... status='approved'`
20. `internal/modules/documents/approval/application/decision_service.go:322` `case domain.QuorumRejectedStage` — rejection branch.
   -> calls: `:325` `UpdateInstanceStatus(... InstanceRejected ...)`; `:334` `set_config('metaldocs.cancel_in_progress', ...)`; `:343` `UPDATE documents ... status='draft'`
21. `internal/modules/documents/approval/application/decision_service.go:380` `s.emitter.Emit` — governance event emit in same tx.
   -> calls: `internal/modules/documents/approval/application/events.go:39` `sqlEmitter.Emit`; `:35` `INSERT INTO governance_events`
22. `internal/modules/documents/approval/application/decision_service.go:387` `s.pdfOutbox.Enqueue` — PDF transactional outbox enqueue.
   -> calls: `internal/modules/render/fanout/pdf_outbox_repository.go:25` `Enqueue`; `:33` `INSERT INTO metaldocs.pdf_dispatch_outbox`; `:35` `ON CONFLICT (tenant_id, revision_id) DO NOTHING`
23. `internal/modules/documents/approval/application/decision_service.go:394` `tx.Commit` — commit write tx.
   -> calls: `internal/modules/documents/approval/http/doc_approval_handler.go:133` `h.idempStore.RecordReplay`
24. `internal/modules/documents/approval/infrastructure/postgres_signoff_idemp_store.go:42` `RecordReplay` — HTTP-layer idempotency write-back after service returns.
   -> calls: `internal/modules/documents/approval/infrastructure/postgres_signoff_idemp_store.go:47` `s.inner.RecordReplay`

### 3. State changes

| Entity | From | To | Trigger | Capability required |
|---|---|---|---|---|
| `approval_stage_instances.status` | `active` | `completed` | quorum approved stage | `doc.signoff` (`internal/modules/documents/approval/application/decision_service.go:141`, `:259`) |
| `approval_instances.status` | `in_progress` | `approved` | all stages complete in approval branch | `doc.signoff` (`internal/modules/documents/approval/application/decision_service.go:141`, `:275`) |
| `documents.status` | `under_review` | `approved` | approval branch document update | `doc.signoff` (`internal/modules/documents/approval/application/decision_service.go:141`, `:293`) |
| `approval_stage_instances.status` | `active` | `rejected_here` | quorum rejected stage | `doc.signoff` (`internal/modules/documents/approval/application/decision_service.go:141`, `:322`) |
| `approval_instances.status` | `in_progress` | `rejected` | rejection branch | `doc.signoff` (`internal/modules/documents/approval/application/decision_service.go:141`, `:325`) |
| `documents.status` | `under_review` | `draft` | rejection branch document update + cancel GUC | `doc.signoff` (`internal/modules/documents/approval/application/decision_service.go:141`, `:334`, `:343`) |

### 4. SQL touched

| File:line | Verb | Table(s) | Auth-area arg (if any) |
|---|---|---|---|
| `internal/modules/documents/approval/application/authz_guc.go:12` | SELECT set_config | tx GUC `metaldocs.tenant_id` | n/a |
| `internal/modules/documents/approval/application/authz_guc.go:15` | SELECT set_config | tx GUC `metaldocs.actor_id` | n/a |
| `internal/modules/documents/approval/repository/postgres_approval_repository.go:314` | SELECT ... FOR UPDATE | `approval_stage_instances` | n/a |
| `internal/modules/documents/approval/application/submit_service.go:280` | SELECT | `documents` | feeds `authz.Require(..., areaCode)` |
| `internal/modules/documents/approval/application/decision_service.go:141` | authz.Require | IAM authz function call in tx | capability `doc.signoff`, area `areaCode` |
| `internal/modules/documents/approval/repository/postgres_approval_repository.go:127` | INSERT | `approval_signoffs` | paired after `authz.Require` |
| `internal/modules/documents/approval/repository/postgres_approval_repository.go:132` | ON CONFLICT DO NOTHING | `approval_signoffs` | replay path |
| `internal/modules/documents/approval/repository/postgres_approval_repository.go:182` | SELECT | `approval_signoffs`, `approval_instances` | replay field-compare load |
| `internal/modules/documents/approval/application/decision_service.go:275` | UPDATE | `approval_instances` | approval branch |
| `internal/modules/documents/approval/application/decision_service.go:293` | UPDATE | `documents` | approval branch |
| `internal/modules/documents/approval/application/decision_service.go:325` | UPDATE | `approval_instances` | rejection branch |
| `internal/modules/documents/approval/application/decision_service.go:334` | SELECT set_config | tx GUC `metaldocs.cancel_in_progress` | rejection branch cancel GUC |
| `internal/modules/documents/approval/application/decision_service.go:343` | UPDATE | `documents` | rejection branch |
| `internal/modules/documents/approval/application/events.go:35` | INSERT | `governance_events` | n/a |
| `internal/modules/render/fanout/pdf_outbox_repository.go:33` | INSERT | `metaldocs.pdf_dispatch_outbox` | n/a |
| `internal/modules/render/fanout/pdf_outbox_repository.go:35` | ON CONFLICT DO NOTHING | `metaldocs.pdf_dispatch_outbox` | outbox idempotency |

Tripwire pairing (`approval_signoffs` INSERT):
- app-layer authz before insert: `internal/modules/documents/approval/application/decision_service.go:141` `authz.Require(..., "doc.signoff", areaCode)` before `internal/modules/documents/approval/repository/postgres_approval_repository.go:127` `INSERT INTO approval_signoffs`.
- DB tripwires on INSERT: `migrations/0135_approval_instances.sql:125` (`trg_signoff_tenant_consistent`), `migrations/0135_approval_instances.sql:151` (`trg_signoff_sod`), `migrations/0180_signoff_eligibility_trigger.sql:26` (`enforce_signoff_eligibility_trg`), all `BEFORE INSERT ON approval_signoffs`.

### 5. Response shape

- 2xx schema ref for `POST /api/v1/documents/{id}/signoffs`: NOT FOUND in searched OpenAPI path set (`api/openapi/spec2.yaml` contains `POST /approval/instances/{instance_id}/stages/{stage_id}/signoffs` at `api/openapi/spec2.yaml:73`).
- Error responses declared on `POST /api/v1/documents/{id}/signoffs`: NOT FOUND in OpenAPI for this path.
- Handler runtime success response for document route: `contracts.SignoffResponse` via `WriteJSON(..., http.StatusOK, ...)` at `internal/modules/documents/approval/http/doc_approval_handler.go:92` and `:136`.

### 6. Cross-references

- Idempotency: yes.
- HTTP-layer idempotency: `CheckReplay` before service write (`internal/modules/documents/approval/http/doc_approval_handler.go:88`), `RecordReplay` after success (`internal/modules/documents/approval/http/doc_approval_handler.go:133`), store impl `internal/modules/documents/approval/infrastructure/postgres_signoff_idemp_store.go:25` and `:42`.
- Repository-layer idempotency: `approval_signoffs` `ON CONFLICT ... DO NOTHING` + field-compare replay (`internal/modules/documents/approval/repository/postgres_approval_repository.go:132`, `:159`, `:168`).
- Two-layer idempotency present: HTTP idempStore + repository ON CONFLICT field-compare replay.
- GUC writes documented: authz GUCs (`metaldocs.tenant_id`, `metaldocs.actor_id`) at `internal/modules/documents/approval/application/authz_guc.go:12`, `:15`; cancel GUC (`metaldocs.cancel_in_progress`) at `internal/modules/documents/approval/application/decision_service.go:334`.
- Transactional outbox pattern: enqueue inside same tx before commit (`internal/modules/documents/approval/application/decision_service.go:387`, `:394`; implementation `internal/modules/render/fanout/pdf_outbox_repository.go:25`, `:33`).
- Tripwire pairing explicitly present on `approval_signoffs` INSERT via BEFORE INSERT triggers (`migrations/0135_approval_instances.sql:125`, `:151`; `migrations/0180_signoff_eligibility_trigger.sql:26`).
- Pagination: no cursor in this operation path.
- Audit log emission: yes, governance event insert through emitter (`internal/modules/documents/approval/application/decision_service.go:380`; `internal/modules/documents/approval/application/events.go:35`).

Additional required step checks from task list:
- Step 10 approval branch "cancel GUC": NOT FOUND in approval branch of `RecordSignoff`; `set_config('metaldocs.cancel_in_progress', ...)` is in rejection branch only (`internal/modules/documents/approval/application/decision_service.go:334`).
