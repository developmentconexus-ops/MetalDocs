# Module #5d — documents/approval/{http,repository,infrastructure,jobs}

**Reviewed:** 2026-05-22
**Reviewers:** ecc:go-reviewer, ecc:security-reviewer, ecc:silent-failure-hunter, ecc:type-design-analyzer, ecc:database-reviewer
**Scope:** `internal/modules/documents/approval/http/`, `internal/modules/documents/approval/repository/`, `internal/modules/documents/approval/infrastructure/`, `internal/modules/documents/approval/jobs/`

---

## Summary

| Severity | Count |
|----------|-------|
| Critical | 11    |
| High     | 20    |
| Medium   | 15    |
| Low      | 11    |

---

## Critical

### 5d-C1 — `ListRoutesHandler` executes raw SQL in HTTP handler with no authz gate → any authenticated user enumerates all routes + stage configurations

**File:** `internal/modules/documents/approval/http/route_admin_handler.go:165-292`
**Fix branch:** `fix/approval-5d-list-routes-c1` (land first)

Two `QueryContext` calls execute directly in the HTTP handler, bypassing `ApprovalRepository`. No capability check guards this handler — any authenticated user in the tenant can enumerate all approval routes including `required_role`, `required_capability`, `area_code`, and `drift_policy` for every stage of every route. This reveals the full authorization model to unprivileged actors.

Fix: (1) add `authz.Require(CapManageRoutes)` before the first query; (2) move route+stage loading into `ApprovalRepository.ListRoutes`; (3) replace the string-joined `IN ($1::uuid, $2::uuid, ...)` placeholder construction with a single JOIN query.

---

### 5d-C2 — `skipReason` variable declared and checked but never passed to `rows.Scan` → `SkipReason` always empty string

**File:** `internal/modules/documents/approval/repository/postgres_approval_repository.go:441`
**Fix branch:** `fix/approval-5d-repo-c2-c10`

```go
var skipReason sql.NullString   // declared at line 438
// ... rows.Scan(...) at line 441 — skipReason NOT in the Scan args
if skipReason.Valid {           // line 466 — always false, zero value
```

The SELECT does not include `skip_reason`; `rows.Scan` does not receive `&skipReason`. `StageInstance.SkipReason` is always `""` regardless of DB content — silently hiding why a stage was skipped from the API response and audit trail.

Fix: add `skip_reason` to the SELECT column list and `&skipReason` to the `rows.Scan` call.

---

### 5d-C3 — Idempotency replay in `SignoffByDocumentHandler` returns before eligibility check → revoked actor replays signoff

**File:** `internal/modules/documents/approval/http/doc_approval_handler.go:58-152`
**Fix branch:** `fix/approval-5d-replay-eligibility-c3` (land second)

The idempotency store check runs at lines 97-107, before the active instance is loaded and before `CheckEligibility` fires. A replay hit returns `was_replay: true` with the cached outcome without re-verifying the actor is still an eligible actor for that stage. If an actor's eligibility was revoked after their initial signoff, a client retry using the same `Idempotency-Key` succeeds and returns an approval outcome without the revocation being enforced.

Fix: move the idempotency `CheckReplay` call to after instance load and eligibility verification, or re-validate eligibility on replay before returning the cached outcome.

---

### 5d-C4 — `SupersedeHandler` raw SQL for `revision_version` in HTTP handler + TOCTOU race

**File:** `internal/modules/documents/approval/http/supersede_handler.go:53-58`
**Fix branch:** `fix/approval-5d-repo-c2-c10`

`SELECT revision_version FROM documents WHERE id=$1 AND tenant_id=$2` executed directly in the HTTP handler. The fetched `priorRevisionVersion` is passed to `PublishSuperseding` without any locking — between the SELECT and the downstream UPDATE another transaction can change the document's state. Also: no nil guard on `h.db` here while all other raw-DB handlers check for nil.

Fix: move the revision_version lookup into the same transaction as the supersede update in the service/repository layer; add `if h.db == nil` guard.

---

### 5d-C5 — `WriteJSON` fallback `json.Marshal` error discarded → blank 500 body on double-marshal failure

**File:** `internal/modules/documents/approval/http/errors.go:214`
**Fix branch:** `fix/approval-5d-writejson-c5-c6`

```go
payload, _ = json.Marshal(fallback)
```

If both the primary and fallback marshals fail, `payload` is nil and `w.Write(nil)` sends an empty body with status 500. The client receives a structureless empty response with no `Content-Type`.

Fix: use a hardcoded byte literal as the ultimate fallback:
```go
w.Write([]byte(`{"status":500,"title":"internal error"}`))
```

---

### 5d-C6 — `PublishHandler` uses `h.readSvc` without nil guard → nil-pointer panic on misconfigured wiring

**File:** `internal/modules/documents/approval/http/publish_handler.go:50-54`
**Fix branch:** `fix/approval-5d-writejson-c5-c6`

Every other handler performs an explicit nil guard; `PublishHandler` does not. A misconfigured `Handler` produces a panic, not a 500 response.

Fix: add `if h.readSvc == nil { WriteError(w, errors.New("read service not configured")); return }` before line 50.

---

### 5d-C7 — `parseIfMatch` result discarded in `PublishHandler` and `SignoffHandler` → OCC not enforced

**Files:** `internal/modules/documents/approval/http/publish_handler.go:45-48`, `signoff_handler.go:33-36`
**Fix branch:** `fix/approval-5d-occ-c7`

Both handlers parse `If-Match` for format validation then discard the version with `if _, err := parseIfMatch(...)`. The publish and signoff operations proceed without version pinning — a client operating on a stale instance is never rejected at the HTTP layer.

Fix: capture the parsed version and thread it through to `PublishRequest` / `SignoffRequest` as `ExpectedRevisionVersion`, and enforce it in the service layer.

---

### 5d-C8 — `loadSignoffsForInstance` no tenant predicate → cross-tenant signoff leakage

**File:** `internal/modules/documents/approval/repository/postgres_approval_repository.go:497-546`
**Fix branch:** `fix/approval-5d-tenant-isolation-c8-c9`

```sql
WHERE approval_instance_id = $1
```

No tenant gate. A caller passing a cross-tenant `instanceID` (obtained via another path) receives all signoffs for that instance including actor user IDs, decisions, content hashes, and signature payloads.

Fix: join through `approval_instances` to enforce tenant:
```sql
WHERE s.approval_instance_id = $1
  AND EXISTS (SELECT 1 FROM approval_instances ai WHERE ai.id = $1 AND ai.tenant_id = $2)
```
Add `tenantID string` to the `loadSignoffsForInstance` signature.

---

### 5d-C9 — `loadStageInstances` no tenant predicate → cross-tenant stage leakage

**File:** `internal/modules/documents/approval/repository/postgres_approval_repository.go:411-423`
**Fix branch:** `fix/approval-5d-tenant-isolation-c8-c9`

Same pattern as C8 — `WHERE approval_instance_id = $1` only. Leaks `eligible_actor_ids` (JSONB array of user IDs), stage policies, and drift configuration for any instance whose ID is guessable.

Fix: same subquery tenant check; add `tenantID` to `loadStageInstances` signature.

---

### 5d-C10 — `InsertSignoff` ON CONFLICT targets wrong constraint pair → cross-stage double-sign silently accepted or returns wrong error

**File:** `internal/modules/documents/approval/repository/postgres_approval_repository.go:113-175`
**Fix branch:** `fix/approval-5d-repo-c2-c10`

`ON CONFLICT (approval_instance_id, actor_user_id) DO NOTHING` — conflict fires on instance+actor. An actor who signed stage 1 and then signs stage 2 of the same instance hits the conflict, falls into the replay branch, compares stage IDs (different), and returns `ErrActorAlreadySigned` — which is the wrong error (they haven't yet signed stage 2). Also: if the tighter `approval_signoffs_stage_instance_id_actor_user_id_key` constraint fires first (ordering is non-deterministic), `MapPgError` maps it to `ErrActorAlreadySigned` but the ON CONFLICT clause did not cover it, bypassing the replay check entirely.

Fix: change the conflict target to `ON CONFLICT (stage_instance_id, actor_user_id)` matching the tighter business rule; align the error mapping in `errors.go` to match.

---

### 5d-C11 — `SignoffRequest.Decision` is bare `string`; `Validate()` call is not enforced by the type → invalid decision flows into domain

**File:** `internal/modules/documents/approval/http/contracts/signoff.go:6`
**Fix branch:** `fix/approval-5d-occ-c7`

Nothing prevents `SignoffRequest{Decision: "maybe"}` from being constructed and passed to the service if `Validate()` is ever skipped. The type carries no invariant.

Fix: define `type Decision string` with constants `DecisionApprove`/`DecisionReject`; use `Decision` on the field. The `switch` in `Validate()` then exhausts the domain. The handler should parse and validate in one step, never holding a raw string past the contract boundary.

---

## High

### 5d-H1 — ETag hardcoded as `"v1"` in both instance handlers → cache invalidation broken

**Files:** `internal/modules/documents/approval/http/get_instance_handler.go:44`, `doc_approval_handler.go:44`

Fix: derive from `inst.RevisionVersion`: `fmt.Sprintf("\"v%d\"", inst.RevisionVersion)`.

---

### 5d-H2 — `CancelByDocumentHandler` passes `ExpectedRevisionVersion: 0` hardcoded → OCC disabled for doc-scoped cancel

**File:** `internal/modules/documents/approval/http/doc_approval_handler.go:197-200`

`CancelHandler` (instance-scoped) correctly parses If-Match; `CancelByDocumentHandler` always sends revision 0. If the service enforces OCC, every doc-cancel fails. If ignored, the field is misleading dead weight.

Fix: parse and forward If-Match, matching the instance-scoped pattern.

---

### 5d-H3 — `resolveEligibleActorNames` N+1: one `QueryRowContext` per actor per stage

**File:** `internal/modules/documents/approval/http/get_instance_handler.go:120-135`

3 stages × 5 eligible actors = 15 sequential DB round-trips per GET. With large eligible sets this is also an amplification DoS vector if route configuration is not constrained.

Fix: batch with `WHERE user_id = ANY($1::text[])` in one query; add a hard cap on eligible actors per stage in route validation.

---

### 5d-H4 — Package-level `var` function seams for cancel/obsolete/publish/supersede → race under `t.Parallel`

**File:** `internal/modules/documents/approval/http/handler.go`

`cancelInstance`, `markObsolete`, `publishApproved`, `schedulePublish`, `publishSuperseding` are package-level vars mutated in tests. Concurrent tests under `-race` will race on these writes.

Fix: inject as interface fields on `Handler` (consistent with `submitSvc`, `decisionSvc`, `readSvc`). Remove the package-level vars.

---

### 5d-H5 — In-process rate limiter in `PasswordReauthProvider` not shared across replicas → 5×N attempts per window

**File:** `internal/modules/documents/approval/infrastructure/signature/password_reauth.go`

`maxFailures=5` enforced per-process. With N replicas an attacker gets 5N attempts per 60-second window.

Fix: document as known per-instance limitation; track issue to back with Redis counter or Postgres row + advisory lock.

---

### 5d-H6 — `ScheduledPublishWorker.Service` and `DB` public fields, no nil guard in `Work`

**File:** `internal/modules/documents/approval/jobs/scheduled_publish_job.go:19-27`

Nil `Service` or `DB` produces a panic caught by River as an unhandled job failure, not a structured error with retry semantics.

Fix: add constructor `NewScheduledPublishWorker(svc, db)` that panics on nil; make fields unexported; add nil guard in `Work`.

---

### 5d-H7 — `repository/errors.go:55` — wrong default sentinel for unrecognised unique violation

**File:** `internal/modules/documents/approval/repository/errors.go:55`

Any unknown `23505` constraint maps to `ErrActorAlreadySigned` — wrong HTTP 409 code and wrong user-facing message for future constraints.

Fix: return `fmt.Errorf("%w: unique constraint %q", ErrUnknownDB, pgErr.ConstraintName)` in the default branch.

---

### 5d-H8 — `RowsAffected()` error discarded in `UpdateStageStatus` and `UpdateInstanceStatus`

**File:** `internal/modules/documents/approval/repository/postgres_approval_repository.go:567, 589`

`n, _ := res.RowsAffected()` — driver error makes `n == 0` → returns domain sentinel instead of DB error.

Fix:
```go
n, err := res.RowsAffected()
if err != nil { return fmt.Errorf("rows affected: %w", err) }
```

---

### 5d-H9 — `InsertInstance` and `InsertStageInstances` RowsAffected not checked → silent zero-row insert

**File:** `internal/modules/documents/approval/repository/postgres_approval_repository.go:34-55, 106`

`_, err := tx.ExecContext(...)` result discarded. A zero-row insert (trigger suppression, deferred constraint, bug) returns no error but starts an approval workflow with no persisted record.

Fix: check `affected == 1` for `InsertInstance`; `affected == int64(len(stages))` for `InsertStageInstances`.

---

### 5d-H10 — `postgres_signoff_idemp_store.go:46` — `json.Marshal` error discarded → nil body stored → unmarshal error on replay treated as miss

**File:** `internal/modules/documents/approval/infrastructure/postgres_signoff_idemp_store.go:46`

If marshal fails, `RecordReplay` stores null body. `CheckReplay` then calls `json.Unmarshal(nil, ...)` → error → treated as miss → signoff re-executes.

Fix:
```go
body, err := json.Marshal(map[string]string{"outcome": outcome})
if err != nil { return fmt.Errorf("marshal idempotency body: %w", err) }
```

---

### 5d-H11 — `password_reauth.go:101` — `json.Marshal` discarded in `Sign` → nil `SignaturePayload`

**File:** `internal/modules/documents/approval/infrastructure/signature/password_reauth.go:101`

`payload, _ := json.Marshal(...)` — nil payload written to `SignatureResult.Payload` → null signature payload in DB → downstream audit and replay code encounters nil.

Fix: return the marshal error wrapped with context.

---

### 5d-H12 — `CheckReplay` passes empty string as `requestBody` → potential idempotency key bypass

**File:** `internal/modules/documents/approval/infrastructure/postgres_signoff_idemp_store.go:29`

`s.inner.CheckReplay(ctx, tenantID, actorID, idempKey, "")` — fifth arg is `requestBody`. If the platform store uses this in key hashing, all signoff idempotency keys are computed without the request body component, colliding across different signoff requests with the same key.

Fix: verify whether the platform `Store.CheckReplay` uses `requestBody`; if so, pass the canonical serialized request body.

---

### 5d-H13 — `SignoffHandler` (instance-scoped) has no idempotency store check

**File:** `internal/modules/documents/approval/http/signoff_handler.go:52-66`

`SignoffByDocumentHandler` has idempotency replay protection; `SignoffHandler` relies only on the DB-level `ON CONFLICT DO NOTHING`. No `WasReplay` field in the response. Inconsistent protection across equivalent endpoints.

Fix: apply the same `CheckReplay` / `RecordReplay` pattern to `SignoffHandler`.

---

### 5d-H14 — `LoadActiveInstanceByDocument` non-deterministic `ORDER BY` → missing tie-breaker

**File:** `internal/modules/documents/approval/repository/postgres_approval_repository.go:276-312`

`ORDER BY ai.submitted_at DESC LIMIT 1` — ties on sub-millisecond submits are non-deterministic.

Fix: add `ai.id DESC` as secondary sort key. Add partial index `ON approval_instances(document_id, tenant_id) WHERE status IN ('in_progress','approved')`.

---

### 5d-H15 — `UpdateRouteHandler` If-Match parsed but discarded → concurrent route updates not detected

**File:** `internal/modules/documents/approval/http/route_admin_handler.go:62-112`

Same pattern as `PublishHandler` — parsed version never forwarded to `UpdateRouteInput`.

Fix: pass to `UpdateRouteInput.ExpectedVersion`; enforce in `RouteAdminService.UpdateRoute`.

---

### 5d-H16 — `handler.go:39-44` — duplicate exported + unexported error var pairs

**File:** `internal/modules/documents/approval/http/handler.go:39-44`

`ErrIfMatchRequired`/`errIfMatchRequired` and `ErrIfMatchMalformed`/`errIfMatchMalformed` — four vars for two sentinels. Unexported aliases serve no purpose, create maintenance trap.

Fix: delete `errIfMatchRequired` and `errIfMatchMalformed`; use exported names throughout.

---

### 5d-H17 — `LoadCurrentPublishedHead` `FOR UPDATE` without confirmed index → may lock wrong row

**File:** `internal/modules/documents/approval/repository/postgres_approval_repository.go:340-385`

`ORDER BY revision_number DESC LIMIT 1 FOR UPDATE` — without a covering index, Postgres resolves ORDER BY via sort before applying LIMIT, which can lock rows that are then discarded. Confirm with `EXPLAIN ANALYZE` that the plan is `Index Scan Backward`.

---

### 5d-H18 — `ScheduledPublishArgs` has no `UniqueOpts` → duplicate job enqueueing possible

**File:** `internal/modules/documents/approval/jobs/scheduled_publish_args.go`

River supports `UniqueOpts` to prevent duplicate enqueues. Without it, a retry storm or double-call to `EnqueueScheduledPublishTx` creates duplicate jobs for the same `(document_id, schedule_generation)`.

Fix: implement `InsertOpts()` returning `river.InsertOpts{UniqueOpts: river.UniqueOpts{ByArgs: true}}`.

---

### 5d-H19 — `bcrypt.Cost` not enforced against minimum → stale low-cost hashes accepted silently

**File:** `internal/modules/documents/approval/infrastructure/signature/password_reauth.go:88`

`bcrypt.Cost(hash)` is called but cost is only stored in `SignatureResult` — never compared to a minimum. A hash at cost 10 is accepted.

Fix: check `cost >= MinAcceptedBcryptCost` (define as 12); reject or re-hash on-the-fly for legacy values.

---

### 5d-H20 — `cancel_handler_test.go:88` constructs `&Handler{}` zero value instead of configured handler

**File:** `internal/modules/documents/approval/http/cancel_handler_test.go:88`

Tests pass a zero-value `Handler` to the mux — service wiring is never verified. Happy-path tests pass only because `cancelInstance` package-level var is replaced entirely.

Fix: pass the configured `Handler` instance, or (better) remove package-level vars and inject the service (H4).

---

## Medium

### 5d-M1 — `_ = fmt.Sprintf` in `password_reauth.go` keeps dead `fmt` import alive

**File:** `internal/modules/documents/approval/infrastructure/signature/password_reauth.go:178`

Fix: remove the line and the import.

---

### 5d-M2 — `contracts/route.go` — `Quorum` and `DriftPolicy` as bare strings

**File:** `internal/modules/documents/approval/http/contracts/route.go:7-13`

Valid sets are finite and stable (`"any_1_of"`, `"all_of"`, `"m_of_n"`, `"reduce_quorum"`, `"fail_stage"`, `"keep_snapshot"`). A typo in a DB seed or admin form is invisible until a runtime switch.

Fix: `type QuorumKind string`, `type DriftPolicy string` with exported constants.

---

### 5d-M3 — `contracts/instance_read.go` — Status/Decision/SignatureMethod bare strings in API responses

**File:** `internal/modules/documents/approval/http/contracts/instance_read.go:8`

Fix: define `type InstanceStatus string`, `type SignoffDecision string`, `type SignatureMethod string` with constants.

---

### 5d-M4 — `contracts/strictjson.go` — `ErrDuplicateKey` exported but never produced

**File:** `internal/modules/documents/approval/http/contracts/strictjson.go:44-55`

`// TODO: duplicate key detection` — a caller sending `{"decision":"approve","decision":"reject"}` has the last key win silently.

Fix: implement token-level duplicate detection or remove `ErrDuplicateKey` from the exported taxonomy.

---

### 5d-M5 — `contracts/route.go` — `ProfileCode`/`RequiredRole`/`AreaCode` have no max-length or pattern validation

**File:** `internal/modules/documents/approval/http/contracts/route.go:23-31`

Only non-empty checked. Overlength or non-slug values propagate to DB.

Fix: add `validateSlug(field, value, maxLen int)` helper enforcing e.g. lowercase alphanumeric + underscore, max 64 chars.

---

### 5d-M6 — `looksLikeValidationError` heuristic string-match → misclassifies future dependency errors

**File:** `internal/modules/documents/approval/http/errors.go:234-243`

`strings.Contains(err.Error(), " is required")` or `" must be "` — any future error message from a dependency containing these substrings produces a wrong 400.

Fix: introduce typed `ErrValidation` sentinel; wrap all contract validation errors with it; replace heuristic with `errors.Is(err, ErrValidation)`.

---

### 5d-M7 — `MapErrorToResponse` two-level nested switch → hard to maintain as error types grow

**File:** `internal/modules/documents/approval/http/errors.go:65-201`

Outer `errors.Is` + inner `errors.As` nested switch. A new error type wrapping an outer sentinel hits the wrong branch silently.

Fix: flatten into a single ordered chain of `errors.Is`/`errors.As` checks.

---

### 5d-M8 — `LoadActiveInstanceByDocument` uses `QueryRowContext` with `ORDER BY ... LIMIT 1` — suppresses `rows.Err()`

**File:** `internal/modules/documents/approval/repository/postgres_approval_repository.go:276`

Minor hygiene — `QueryRowContext` with `LIMIT 1` is correct but suppresses driver-level errors that `QueryContext` + `rows.Err()` would surface.

---

### 5d-M9 — `loadStageInstances` and `loadSignoffsForInstance` scan errors not wrapped with context

**Files:** `internal/modules/documents/approval/repository/postgres_approval_repository.go:447-448, 522-524`

`return nil, err` with no `fmt.Errorf("scan stage instance row for instance %s: %w", ...)` context.

Fix: wrap all scan errors with row identity context.

---

### 5d-M10 — `PublishRequest` is a hollow struct misleadingly shaped as a JSON body contract

**File:** `internal/modules/documents/approval/http/contracts/publish.go:8-10`

Only header-sourced fields, no JSON tags, never decoded from body. Reads as a body contract to maintainers.

Fix: remove or rename to `PublishHeaders`; document as header-binding struct.

---

### 5d-M11 — `NewRouteResponse`/`NewVersion` field uses `int` with `omitempty` → version 0 silently dropped

**File:** `internal/modules/documents/approval/http/contracts/route.go:96-106`

`NewVersion int \`json:"new_version,omitempty"\`` — `omitempty` on `int` omits zero, which may be a valid version.

Fix: use `*int` so `nil` means absent and `0` is preserved.

---

### 5d-M12 — `IdempotencyKey` fields have no JSON tags → looks like a body field but is header-sourced

**Files:** Multiple contract structs (`SubmitRequest`, `SignoffRequest`, `CancelRequest`, etc.)

Fix: add comment `// set by handler from X-Idempotency-Key header, not decoded from JSON body` or move to a separate non-JSON struct.

---

### 5d-M13 — `ScheduledPublishJob` opens its own transaction from pool — must verify full idempotency for River at-least-once

**File:** `internal/modules/documents/approval/jobs/scheduled_publish_job.go`

If `RunScheduledPublishJob` commits but River crashes before marking the job complete, the job reruns. Verify `ExpectedRevisionVersion` + `ScheduleGeneration` together form a complete idempotency guard for all mutations.

---

### 5d-M14 — `approval_signoffs(approval_instance_id)` index not confirmed → seq scan on every signoff load

**File:** `internal/modules/documents/approval/repository/postgres_approval_repository.go:497`

FK column on `approval_signoffs` — Postgres does not auto-create FK indexes. Unindexed FK causes seq scan on `loadSignoffsForInstance` and `LoadSignoffByActor` JOIN, both on the hot approval decision path.

Fix: confirm `CREATE INDEX ON approval_signoffs(approval_instance_id)` exists in migrations.

---

### 5d-M15 — `approval_stage_instances(approval_instance_id)` index not confirmed → seq scan in `UpdateStageStatus` JOIN

**File:** `internal/modules/documents/approval/repository/postgres_approval_repository.go:551`

Same issue — unindexed FK on `approval_stage_instances` causes full table scan when the UPDATE joins through to `approval_instances` to assert tenant.

Fix: confirm `CREATE INDEX ON approval_stage_instances(approval_instance_id)` exists.

---

## Low

### 5d-L1 — `cancelInstance` and peer package-level vars kept alive by `_ = fmt.Sprintf`-style sentinel in test files

**Files:** `internal/modules/documents/approval/http/cancel_handler.go:14-19`

Symptom of H4. Anti-pattern note only.

---

### 5d-L2 — `contracts/route.go` stage `Order` error message uses slice index not `stage.Order`

**File:** `internal/modules/documents/approval/http/contracts/route.go:119-138`

Error says `"stages[%d].order must be %d"` using `i` — confusing when order is wrong. Include actual `stage.Order` value.

---

### 5d-L3 — `contracts_test.go` uses flat `t.Fatalf` — no subtests → single failure masks all subsequent cases

**File:** `internal/modules/documents/approval/http/contracts/contracts_test.go`

Fix: convert each validation case to `t.Run` subtest.

---

### 5d-L4 — `contracts/signoff.go` — `SignoffResponse.Outcome` bare string

**File:** `internal/modules/documents/approval/http/contracts/signoff.go:30-35`

Fix: define `type SignoffOutcome string` with constants for `"stage_advanced"`, `"instance_approved"`, etc.

---

### 5d-L5 — `repository/approval_repository.go` — all IDs bare `string` on interface methods → transposition undetectable

**File:** `internal/modules/documents/approval/repository/approval_repository.go:38-46`

Five-parameter functions with `tenantID, instanceID, actorUserID, docID, stageID` all as `string`. A transposition compiles silently.

Fix: introduce `type TenantID string`, `type InstanceID string`, `type ActorID string` newtypes.

---

### 5d-L6 — `infrastructure/signature/provider.go` — `SignRequest.ContentHash` is bare string with no SHA-256 format constraint

**File:** `internal/modules/documents/approval/infrastructure/signature/provider.go:14-18`

Same constraint validated in `contracts/submit.go` (64 hex chars) but not expressed at the `SignRequest` type level.

Fix: `type SHA256Hex string` with constructor validating format.

---

### 5d-L7 — `handler.go:64-76` — `NewHandler` assigns individual service fields from bundle then both paths exist

**File:** `internal/modules/documents/approval/http/handler.go:64-76`

`h.submitSvc = services.Submit` AND `h.services.Submit` both reachable — two code paths for same value. Submit handler checks both at lines 44-46.

Fix: pick one (bundle or individual fields) and remove the redundancy.

---

### 5d-L8 — `WriteError` silently ignores `problem.Write` failure before fallback

**File:** `internal/modules/documents/approval/http/errors.go:204-208`

If `problem.Write` fails, fallback `WriteJSON` is called; that failure is also dropped. Log before fallback.

---

### 5d-L9 — `postgres_approval_repository.go:232` — `LoadInstance` accepts `*sql.Tx` but call sites pass `*sql.DB`

**File:** `internal/modules/documents/approval/repository/postgres_approval_repository.go:232`

Interface-contract inconsistency. Callers passing `*sql.DB` must wrap in transaction or there is a compile-time type mismatch hidden in the application layer. Audit all call sites.

---

### 5d-L10 — Dates in `InstanceResponse` as bare `string` (`SubmittedAt`, `CompletedAt *string`)

**File:** `internal/modules/documents/approval/http/contracts/instance_read.go:10-11`

Fix: use `time.Time` / `*time.Time`; JSON encoder handles RFC3339 automatically.

---

### 5d-L11 — `password_reauth_test.go` and `registry_test.go` lack edge-case tests for rate limit window reset

**File:** `internal/modules/documents/approval/infrastructure/signature/`

The 60-second window reset is critical for rate limiter correctness; no test verifies the sweep/reset path.

---

## Critical Backlog — Fix Branches

| ID | File:line | Severity | Owner | ETA | Fix branch | Status |
|----|-----------|----------|-------|-----|------------|--------|
| 5d-C1 | `http/route_admin_handler.go:165` no authz gate + raw SQL in ListRoutesHandler | Critical | leandrotca | TBC | `fix/approval-5d-list-routes-c1` | Backlog (land first) |
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
