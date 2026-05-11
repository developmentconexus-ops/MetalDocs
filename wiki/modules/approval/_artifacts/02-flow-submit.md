# Data-flow Trace — POST /api/v2/documents/{id}/finalize (submit revision for review)

## 1. Entry point

### Entry A (documents-module primary path)
| Layer | Symbol | File:line |
|---|---|---|
| OpenAPI op | `(operationId missing for this path)` | `api/openapi/v1/openapi.yaml:3251` |
| Generated server stub | `ServerInterface.PostApiV2DocumentsIdFinalize` | `internal/modules/documents/api/api.gen.go:757` |
| Handler | `(*Handler).finalizeDocument` | `internal/modules/documents/delivery/http/handler.go:316` |

### Entry B (approval-module direct entry)
| Layer | Symbol | File:line |
|---|---|---|
| OpenAPI op | `(operationId missing in spec2 submit path)` | `api/openapi/spec2.yaml:33` |
| Generated server stub | `n/a — approval module uses net/http mux route` | `internal/modules/documents/approval/http/router.go:8` |
| Handler | `(*Handler).SubmitHandler` | `internal/modules/documents/approval/http/submit_handler.go:14` |

## 2. Call chain

1. `internal/modules/documents/delivery/http/handler.go:316` `(*Handler).finalizeDocument` — documents-module entry for `POST /api/v2/documents/{id}/finalize`.
   ? calls: `internal/modules/documents/delivery/http/handler.go:403` `submitSvc.SubmitRevisionForReview`

2. `internal/modules/documents/approval/http/submit_handler.go:14` `(*Handler).SubmitHandler` — approval-module direct entry for `POST /api/v2/documents/{id}/submit`.
   ? calls: `internal/modules/documents/approval/http/submit_handler.go:50` `submitSvc.SubmitRevisionForReview`

3. `internal/modules/documents/approval/application/submit_service.go:43` `(*SubmitService).SubmitRevisionForReview` — core submit flow.
   ? calls: `internal/modules/documents/approval/application/services.go:61` `ValidateEventPayload`

4. `internal/modules/documents/approval/application/services.go:61` `ValidateEventPayload` — validates payload values for float64 rejection.
   ? returns to: `submit_service.go:45`

5. `internal/modules/documents/approval/application/submit_service.go:50` `ComputeContentHash` — computes canonical content hash.
   ? calls: `internal/modules/documents/approval/application/content_hash.go:35` `ComputeContentHash`

6. `internal/modules/documents/approval/application/submit_service.go:61` `ComputeIdempotencyKey` — derives idempotency key from actor/document/server time.
   ? calls: `internal/modules/documents/approval/application/idempotency.go:25` `ComputeIdempotencyKey`

7. `internal/modules/documents/approval/application/submit_service.go:68` `db.BeginTx` — transaction boundary starts.
   ? calls: `database/sql` tx begin

8. `internal/modules/documents/approval/application/submit_service.go:73` `authz.WithCapCache` — capability cache attached to context.
   ? calls: `internal/modules/iam/authz/authz.go:28` `WithCapCache`

9. `internal/modules/documents/approval/application/submit_service.go:75` `setAuthzGUC` — sets `metaldocs.tenant_id` + `metaldocs.actor_id` on tx.
   ? calls: `internal/modules/documents/approval/application/authz_guc.go:11` `setAuthzGUC`

10. `internal/modules/documents/approval/application/submit_service.go:80` `loadDocumentAreaCode` — loads area code from `documents` with `controlled_documents` fallback and tenant fallback.
    ? calls: `internal/modules/documents/approval/application/submit_service.go:276` `loadDocumentAreaCode`

11. `internal/modules/documents/approval/application/submit_service.go:85` `authz.Require(ctx, tx, "doc.submit", areaCode)` — capability gate on tx.
    ? calls: `internal/modules/iam/authz/authz.go:44` `Require`

12. `internal/modules/documents/approval/application/submit_service.go:91` `s.loadRoute` — loads route + stages from catalogue tables.
    ? calls: `internal/modules/documents/approval/application/submit_service.go:226` `(*SubmitService).loadRoute`

13. `internal/modules/documents/approval/application/submit_service.go:98` `route.Validate` — validates route structural invariants.
    ? calls: `internal/modules/documents/approval/domain/route.go:48` `(Route).Validate`

14. `internal/modules/documents/approval/application/submit_service.go:121` `repo.InsertInstance` — inserts `approval_instances` row.
    ? calls: `internal/modules/documents/approval/repository/postgres_approval_repository.go:32` `(*postgresApprovalRepository).InsertInstance`

15. `internal/modules/documents/approval/application/submit_service.go:140` `resolveEligibleActors` (per stage) — loads stage eligible users at query time.
    ? calls: `internal/modules/documents/approval/application/submit_service.go:299` `resolveEligibleActors`

16. `internal/modules/documents/approval/application/submit_service.go:162` `repo.InsertStageInstances` — bulk inserts stage snapshots.
    ? calls: `internal/modules/documents/approval/repository/postgres_approval_repository.go:56` `(*postgresApprovalRepository).InsertStageInstances`

17. `internal/modules/documents/approval/application/submit_service.go:168` `tx.ExecContext(UPDATE documents...)` — transition `documents.status` draft ? under_review and increments `revision_version` with OCC predicate `revision_version = $3`.
    ? on zero rows: `internal/modules/documents/approval/application/submit_service.go:184` returns `repository.ErrStaleRevision`

18. `internal/modules/documents/approval/application/submit_service.go:208` `s.emitter.Emit` — emits governance event with `EventType: "approval_submitted"`.
    ? calls: `internal/modules/documents/approval/application/events.go:39` `(*sqlEmitter).Emit`

19. `internal/modules/documents/approval/application/submit_service.go:214` `tx.Commit` — commits all writes.
    ? calls: `database/sql` tx commit

## 3. State changes

| Entity | From | To | Trigger | Capability required |
|---|---|---|---|---|
| `documents.status` | `draft` | `under_review` | `SubmitRevisionForReview` update statement | `doc.submit` (`submit_service.go:85`) |
| `approval_instances.status` | `pending` (operation statement) | `in_review` (operation statement) | `InsertInstance` writes `domain.InstanceInProgress` snapshot (`submit_service.go:113`, `postgres_approval_repository.go:33`) | `doc.submit` (`submit_service.go:85`) |

## 4. SQL touched

| File:line | Verb | Table(s) | Auth-area arg (if any) |
|---|---|---|---|
| `internal/modules/documents/delivery/http/handler.go:345` | SELECT | `documents` | none |
| `internal/modules/documents/delivery/http/handler.go:362` | SELECT | `controlled_documents` | none |
| `internal/modules/documents/delivery/http/handler.go:378` | SELECT | `approval_routes` | none |
| `internal/modules/documents/delivery/http/handler.go:397` | SELECT | `document_revisions` | none |
| `internal/modules/documents/approval/application/authz_guc.go:12` | SELECT set_config | session GUC | tenant_id |
| `internal/modules/documents/approval/application/authz_guc.go:15` | SELECT set_config | session GUC | actor_id |
| `internal/modules/documents/approval/application/submit_service.go:278` | SELECT | `documents` + `controlled_documents` | none |
| `internal/modules/iam/authz/authz.go:60` | SELECT EXISTS | `metaldocs.iam_user_roles` | actor/tenant from tx GUC context |
| `internal/modules/iam/authz/authz.go:75` | SELECT EXISTS | `metaldocs.role_capabilities` + `metaldocs.user_process_areas` | `capability='doc.submit'`, `areaCode` argument |
| `internal/modules/documents/approval/application/submit_service.go:228` | SELECT | `approval_routes` | none |
| `internal/modules/documents/approval/application/submit_service.go:241` | SELECT | `approval_route_stages` | none |
| `internal/modules/documents/approval/repository/postgres_approval_repository.go:33` | INSERT | `approval_instances` | none |
| `internal/modules/documents/approval/application/submit_service.go:301` | SELECT | `metaldocs.user_process_areas` | stage `areaCode`, `requiredRole`, window `effective_from <= now()` and `(effective_to IS NULL OR effective_to > now())` |
| `internal/modules/documents/approval/repository/postgres_approval_repository.go:104` | INSERT | `approval_stage_instances` | none |
| `internal/modules/documents/approval/application/submit_service.go:168` | UPDATE | `documents` | OCC `revision_version = $3` |
| `internal/modules/documents/approval/application/events.go:44` | INSERT | `governance_events` | none |

Tripwire pairing (same tx):
- `authz.Require("doc.submit", areaCode)` at `internal/modules/documents/approval/application/submit_service.go:85`
- `INSERT approval_instances` at `internal/modules/documents/approval/repository/postgres_approval_repository.go:33`
- Pairing status: `OK` (Require occurs before INSERT in `SubmitRevisionForReview` call order).

## 5. Response shape

- Finalize path (`/api/v2/documents/{id}/finalize`):
- Declared responses: `200`, `409` at `api/openapi/v1/openapi.yaml:3255`.
- 2xx schema ref: `(unclear: no response content schema declared on this path in openapi.yaml)`.
- Problem `type` URI: `(unclear: v1 finalize path does not declare Problem type URI fields)`.

- Direct submit path (`/documents/{id}/submit` in spec2):
- 2xx schema ref: `#/components/schemas/SubmitResponse` at `api/openapi/spec2.yaml:59` and schema at `api/openapi/spec2.yaml:767`.
- Error responses declared: `400`, `403`, `409`, `412`, `428`, `500` at `api/openapi/spec2.yaml:60-71`.
- Problem `type` URI: `(unclear: ErrorResponse schema uses {error{code,message,details},request_id}; no URI-typed problem field)` at `api/openapi/spec2.yaml:596`.

## 6. Cross-references

- Idempotency: `yes`.
- Derivation path: `internal/modules/documents/approval/application/idempotency.go:25`.
- Persistence column on submit instance insert: `approval_instances.idempotency_key` at `internal/modules/documents/approval/repository/postgres_approval_repository.go:36-37`.

- Pagination: `no` for this operation.

- Audit log / governance event emission: `yes`.
- Service emit call: `internal/modules/documents/approval/application/submit_service.go:208`.
- Sink insert path: `internal/modules/documents/approval/application/events.go:34` and `events.go:44` (`governance_events`).
