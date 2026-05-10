### 1. Entry point
| Layer | Symbol | File:line |
|---|---|---|
| OpenAPI op | operationId `finalizeDocument` | (unclear: `operationId: finalizeDocument` not found in `api/openapi/v1/openapi.yaml`; path exists at `api/openapi/v1/openapi.yaml:3251`) |
| Generated server stub | `PostApiV2DocumentsIdFinalize` | (unclear: `api/api.gen.go` not found; generated stub found at `internal/modules/documents/api/api.gen.go:1215`) |
| Handler | `Handler.finalizeDocument` | `internal/modules/documents/delivery/http/handler.go:316` |

### 2. Call chain
1. `Handler.finalizeDocument` receives POST and validates role/ownership via `h.authorizeDocumentScope` (`internal/modules/documents/delivery/http/handler.go:316`, `internal/modules/documents/delivery/http/handler.go:336`, `internal/modules/documents/delivery/http/handler.go:869`).
2. Tier-1 check in handler scope: `hasAnyRole(r, roleAdmin, roleDocumentFiller)` (`internal/modules/documents/delivery/http/handler.go:870`) with role strings `system_admin`/`document_filler` (`internal/modules/documents/delivery/http/handler.go:26`, `internal/modules/documents/delivery/http/handler.go:28`).
3. Handler loads draft guard data from `documents` (`internal/modules/documents/delivery/http/handler.go:344`).
4. Handler resolves approval route and content hash, then calls `h.submitSvc.SubmitRevisionForReview(...)` (`internal/modules/documents/delivery/http/handler.go:403`).
5. `SubmitService.SubmitRevisionForReview` starts tx with `db.BeginTx` (`internal/modules/documents/approval/application/submit_service.go:43`, `internal/modules/documents/approval/application/submit_service.go:68`).
6. `setAuthzGUC(ctx, tx, tenantID, actorID)` called (`internal/modules/documents/approval/application/submit_service.go:75`) implemented at `internal/modules/documents/approval/application/authz_guc.go:11`.
7. Area code resolved by `loadDocumentAreaCode` (`internal/modules/documents/approval/application/submit_service.go:80`, `internal/modules/documents/approval/application/submit_service.go:276`).
8. In-tx authz check: `authz.Require(ctx, tx, "doc.submit", areaCode)` (`internal/modules/documents/approval/application/submit_service.go:85`).
9. Approval instance insert: `s.repo.InsertInstance` (`internal/modules/documents/approval/application/submit_service.go:121`) -> `INSERT INTO approval_instances` (`internal/modules/documents/approval/repository/postgres_approval_repository.go:34`).
10. Stage instances materialized: `s.repo.InsertStageInstances` (`internal/modules/documents/approval/application/submit_service.go:162`) -> `INSERT INTO approval_stage_instances ... eligible_actor_ids` (`internal/modules/documents/approval/repository/postgres_approval_repository.go:97`).
11. Document transition executed in same tx: `UPDATE documents SET status='under_review' ... WHERE status='draft'` (`internal/modules/documents/approval/application/submit_service.go:168`).
12. Governance event emitted: `s.emitter.Emit` (`internal/modules/documents/approval/application/submit_service.go:208`) -> `INSERT INTO governance_events` (`internal/modules/documents/approval/application/events.go:35`).
13. Tx commit: `tx.Commit()` (`internal/modules/documents/approval/application/submit_service.go:214`).
14. Post-commit audit-only call: `h.svc.Finalize(...)` (`internal/modules/documents/delivery/http/handler.go:416`) -> `UpdateDocumentStatus(... draft -> finalized ...)` (`internal/modules/documents/application/service.go:753`).
15. CapabilityChecker impl injection (module wiring context): `capabilityServiceAdapter` in `apps/api/internal/wiring/documents.go:14`, returned by `NewCapabilityChecker` at `apps/api/internal/wiring/documents.go:24`; finalize submit path uses `SubmitSvc` wiring from `apps/api/cmd/metaldocs-api/main.go:316` and `internal/modules/documents/module.go:61`.
16. Idempotency files check: `internal/modules/documents/approval/application/idempotency.go` and `internal/modules/documents/approval/infrastructure/postgres_signoff_idemp_store.go` are signoff-oriented; not called by `finalizeDocument` flow (unclear: no call sites in this handler/submit path).
17. Trigger note: `enforce_snapshot_on_submit_trg` exists on `documents` (`migrations/0152_placeholder_fillin_columns.sql:47`) and fires on `UPDATE documents` in step 11; not on `approval_instances` insert.

### 3. State changes
| Entity | From | To | Trigger | Capability required |
|---|---|---|---|---|
| Document | `draft` | `under_review` | `UPDATE documents ... status='under_review'` in submit tx (`internal/modules/documents/approval/application/submit_service.go:168`) | `doc.submit` via `authz.Require` (`internal/modules/documents/approval/application/submit_service.go:85`) |
| ApprovalInstance | (none) | `in_progress` (created) | `INSERT INTO approval_instances` (`internal/modules/documents/approval/repository/postgres_approval_repository.go:34`) | `doc.submit` via `authz.Require` (`internal/modules/documents/approval/application/submit_service.go:85`) |

### 4. SQL touched
| file:line | verb | table | paired with authz.Require? (yes/no + anchor) |
|---|---|---|---|
| `internal/modules/documents/approval/application/submit_service.go:168` | UPDATE | `documents` | yes (`authz.Require("doc.submit", areaCode)` at `internal/modules/documents/approval/application/submit_service.go:85`) |
| `internal/modules/documents/approval/repository/postgres_approval_repository.go:34` | INSERT | `approval_instances` | yes (`internal/modules/documents/approval/application/submit_service.go:85`) |
| (unclear: not touched in finalize flow) | INSERT | `approval_signoffs` | no in this flow (signoff insert exists at `internal/modules/documents/approval/repository/postgres_approval_repository.go:127`) |
| `internal/modules/documents/approval/application/events.go:35` | INSERT | `governance_events` | yes (same tx after `authz.Require`, anchor `internal/modules/documents/approval/application/submit_service.go:85`) |
| (unclear: not touched in finalize flow) | INSERT | `metaldocs.idempotency_keys` | no in this flow (`INSERT` exists at `internal/platform/idempotency/postgres_store.go:69`) |

Postgres tripwire note (0142b):
- `approval_instances` pairing: trigger installed `migrations/0142b_role_capabilities_v2_enforce.sql:201` with required cap `'doc.submit'` in function logic (`migrations/0142b_role_capabilities_v2_enforce.sql:84`, `migrations/0142b_role_capabilities_v2_enforce.sql:85`) and app-level `authz.Require` present at `internal/modules/documents/approval/application/submit_service.go:85`.
- `approval_signoffs` pairing: trigger installed `migrations/0142b_role_capabilities_v2_enforce.sql:207`; this finalize flow does not execute signoff INSERT.

### 5. Response shape
- 201 body struct definition: (unclear: no named struct; handler writes anonymous `map[string]string{"instanceId": result.InstanceID}` at `internal/modules/documents/delivery/http/handler.go:420`).
- Error responses (trigger + return statements):
  - 400: profile missing -> `http.Error(w, "{\"error\":\"profile not found\"}", http.StatusBadRequest)` (`internal/modules/documents/delivery/http/handler.go:370`).
  - 401: (unclear: no explicit 401 return in `finalizeDocument` path).
  - 403: role/ownership denial -> `httpErr(w, http.StatusForbidden, "forbidden")` (`internal/modules/documents/delivery/http/handler.go:871`, `internal/modules/documents/delivery/http/handler.go:887`).
  - 404: via mapped downstream errors -> `status, msg := mapErr(err); httpErr(w, status, msg)` (`internal/modules/documents/delivery/http/handler.go:412`) and `mapErr` has 404 branches (`internal/modules/documents/delivery/http/handler.go:966`, `internal/modules/documents/delivery/http/handler.go:968`, `internal/modules/documents/delivery/http/handler.go:970`, `internal/modules/documents/delivery/http/handler.go:972`, `internal/modules/documents/delivery/http/handler.go:1000`).
  - 409:
    - draft precondition fail -> `httpErr(w, http.StatusConflict, "document not in draft state")` (`internal/modules/documents/delivery/http/handler.go:351`).
    - route missing -> `httpErr(w, http.StatusConflict, "no active approval route for profile "+profileCode)` (`internal/modules/documents/delivery/http/handler.go:387`).
    - mapped conflict from submit service -> `status, msg := mapErr(err); httpErr(w, status, msg)` (`internal/modules/documents/delivery/http/handler.go:412`) with conflict mappings in `mapErr` (`internal/modules/documents/delivery/http/handler.go:988`, `internal/modules/documents/delivery/http/handler.go:990`, `internal/modules/documents/delivery/http/handler.go:992`, `internal/modules/documents/delivery/http/handler.go:994`, `internal/modules/documents/delivery/http/handler.go:996`, `internal/modules/documents/delivery/http/handler.go:998`, `internal/modules/documents/delivery/http/handler.go:1003`, `internal/modules/documents/delivery/http/handler.go:1005`).
  - 500:
    - query errors -> `httpErr(w, http.StatusInternalServerError, "internal error")` (`internal/modules/documents/delivery/http/handler.go:353`, `internal/modules/documents/delivery/http/handler.go:365`, `internal/modules/documents/delivery/http/handler.go:389`).
    - mapped default -> `return http.StatusInternalServerError, "internal_error"` (`internal/modules/documents/delivery/http/handler.go:1009`).

### 6. Cross-references
- Idempotency: no (for `finalizeDocument` flow). Submit path computes internal deterministic key via `ComputeIdempotencyKey` (`internal/modules/documents/approval/application/submit_service.go:61`, `internal/modules/documents/approval/application/idempotency.go:20`), but does not use `Idempotency-Key` header nor `metaldocs.idempotency_keys` store in this path.
- Pagination: n/a.
- Audit log emission: yes. `EventEmitter` insert into `governance_events` via `s.emitter.Emit` (`internal/modules/documents/approval/application/submit_service.go:208`) and SQL at `internal/modules/documents/approval/application/events.go:35`.
- Capability namespace: `doc.*` string in tx authz call; exact string: `"doc.submit"` (`internal/modules/documents/approval/application/submit_service.go:85`).
