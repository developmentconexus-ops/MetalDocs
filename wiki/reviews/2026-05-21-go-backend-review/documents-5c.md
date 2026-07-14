# Module #5c — documents/approval/{domain,application}

**Reviewed:** 2026-05-22
**Reviewers:** ecc:go-reviewer, ecc:security-reviewer, ecc:silent-failure-hunter, ecc:type-design-analyzer, ecc:database-reviewer
**Scope:** `internal/modules/documents/approval/domain/`, `internal/modules/documents/approval/application/`

---

## Summary

| Severity | Count |
|----------|-------|
| Critical | 12    |
| High     | 20    |
| Medium   | 21    |
| Low      | 13    |

---

## Critical

### 5c-C1 — `canonicalize` error discarded → all failed keys collapse to same SHA256 → idempotency bucketing corrupted

**File:** `internal/modules/documents/approval/application/idempotency.go:38`
**Fix branch:** `fix/approval-5c-idempotency-c1` (land first — smallest scope)

```go
b, _ := canonicalize(m)
sum := sha256.Sum256(b)
```

Error blanked with `_`. If canonicalization fails, `b` is `nil`, and `sha256.Sum256(nil)` = the constant hash of an empty byte sequence (`e3b0c44...`). Every key derived from a failed canonicalization becomes the same hash — all such requests collapse to the same idempotency bucket and are treated as duplicate replays of each other.

Fix: make `ComputeIdempotencyKey` return `(string, error)` and propagate the canonicalize error.

---

### 5c-C2 — Governance event emitted inside transaction that is immediately rolled back → eligibility-rejection audit trail silently lost

**File:** `internal/modules/documents/approval/application/decision_service.go:163-173`
**Fix branch:** `fix/approval-5c-audit-trail-c2-c4`

```go
if err := domain.CheckEligibility(...); err != nil {
    _ = s.emitter.Emit(ctx, tx, GovernanceEvent{...}) // writes to governance_events inside tx
    _ = tx.Rollback()                                  // rolls it back
    return SignoffResult{}, err
}
```

The `sqlEmitter` writes to `governance_events` inside the transaction; `tx.Rollback()` rolls the event row back too. Eligibility violations — security-relevant events in a regulatory compliance system — are never durably recorded.

Fix: emit the eligibility-rejection event after rollback using a separate connection/transaction, or use an outbox pattern (separate table, separate tx).

---

### 5c-C3 — PDF dispatch error silently discarded on deprecated path → approval reported as success despite PDF failure

**File:** `internal/modules/documents/approval/application/decision_service.go:420-425`
**Fix branch:** `fix/approval-5c-audit-trail-c2-c4`

```go
_ = s.pdfDispatcher.Dispatch(dispatchCtx, pdfTenantID, pdfRevisionID)
```

On the deprecated path (`pdfOutbox == nil`), the dispatch error is thrown away. No log, no metric, no returned error. The caller is told the approval succeeded.

Fix: log at ERROR level with `tenantID`, `revisionID`, error message. The deprecated path must be at least observable.

---

### 5c-C4 — `json.Marshal` error discarded in 3 route-admin governance event payloads → empty audit trail for route create/update/deactivate

**File:** `internal/modules/documents/approval/application/route_admin_service.go:112-113, 202-203, 263-264`
**Fix branch:** `fix/approval-5c-audit-trail-c2-c4`

```go
payload, _ := json.Marshal(map[string]any{...})
```

Pattern appears in `Create`, `Update`, and `Deactivate`. If marshalling fails, `payload` is nil; the event is recorded with empty payload via `sqlEmitter`'s fallback to `{}`. Route ID, profile code, and stage count are silently lost from the audit trail.

Fix: check the error and return it, matching every other service's event construction pattern.

---

### 5c-C5 — `RealClock.Now()` calls `e2etest.E2EClockOffset()` in production binary → production timestamps potentially skewed

**File:** `internal/modules/documents/approval/application/services.go:22`
**Fix branch:** `fix/approval-5c-production-clock-c5`

```go
func (RealClock) Now() time.Time { return time.Now().UTC().Add(e2etest.E2EClockOffset()) }
```

Production clock unconditionally imports and calls a test helper. All governance event timestamps, idempotency key timestamps, `effective_from` guards, and `signed_at` values are skewed if the offset is non-zero. This is a test-code dependency in the production code path.

Fix: `RealClock.Now()` must return `time.Now().UTC()` with no offset. Inject a test clock via the `Clock` interface at construction time in tests.

---

### 5c-C6 — `CancelInput.BypassAuthz bool` on public input struct → any caller can bypass authorization

**File:** `internal/modules/documents/approval/application/cancel_service.go:87-100`
**Fix branch:** `fix/approval-5c-authz-bypass-c6-c7` (land second)

`BypassAuthz` is a public bool field. When set to `true`, the service applies `authz.BypassSystem` and skips the `workflow.instance.cancel` capability check entirely. No server-side validation limits who can assert this flag — any HTTP handler or caller constructing `CancelInput` can set it.

Fix: remove `BypassAuthz` from the public input struct. Provide two distinct entry points: `CancelInstance` (user callers) and `SystemCancelInstance` (scheduler/watchdog), where the bypass is implicit and non-callable from user-facing handlers.

---

### 5c-C7 — `LoadInstance` performs no authorization check → IDOR on full instance data

**File:** `internal/modules/documents/approval/application/read_service.go:38-60`
**Fix branch:** `fix/approval-5c-authz-bypass-c6-c7`

`LoadInstance` takes `actorID` as a parameter but never checks any capability before loading the instance — including `EligibleActorIDs`, `ContentHashAtSubmit`, stage snapshots, and full signoff history. Any actor knowing a valid `instanceID` for their tenant can retrieve the complete approval record.

Fix: gate on `authz.Require` with at minimum a read/view capability for the document's area code before loading.

---

### 5c-C8 — Content hash computed from caller-supplied `ContentFormData` → hash integrity violation

**File:** `internal/modules/documents/approval/application/decision_service.go:96-103`
**Fix branch:** `fix/approval-5c-hash-integrity-c8`

`RecordSignoff` computes the content hash from `req.ContentFormData` supplied by the caller. An actor can submit an arbitrary `ContentFormData` map that does not match the stored document content and produce a persisted hash proving they reviewed something they did not.

Fix: fetch the current document form data from the database server-side and compute the hash independently, ignoring any client-supplied form data for this purpose.

---

### 5c-C9 — `UPDATE approval_instances` in `MarkObsolete` missing `tenant_id` → cross-tenant instance cancellation

**File:** `internal/modules/documents/approval/application/obsolete_service.go:110-121`
**Fix branch:** `fix/approval-5c-tenant-isolation-c9-c10-c11`

```sql
UPDATE approval_instances
   SET status = 'cancelled', completed_at = now()
 WHERE document_id = $1 AND status = 'in_progress'
```

No `tenant_id` predicate. If two tenants share a `document_id` UUID (possible in test environments with fixed seeds or external ID minting), this silently cancels a different tenant's in-progress approval.

Fix: add `AND tenant_id = $2` with `req.TenantID`.

---

### 5c-C10 — `UPDATE documents SET status = 'approved'` RowsAffected not checked → phantom approval on concurrent transition

**File:** `internal/modules/documents/approval/application/decision_service.go:309-319`
**Fix branch:** `fix/approval-5c-tenant-isolation-c9-c10-c11`

`tx.ExecContext` result discarded with `_`. If the document is already in a non-`under_review` state (concurrent transition, retry), the UPDATE silently affects 0 rows but the service sets `result.InstanceApproved = true`, enqueues PDF generation, and commits — reporting approval of a document that was not actually transitioned.

Fix: check `RowsAffected()`; roll back with `repository.ErrStaleRevision` if `n == 0`.

---

### 5c-C11 — `UPDATE documents SET status = 'draft'` RowsAffected not checked on reject path → phantom rejection

**File:** `internal/modules/documents/approval/application/decision_service.go:362-374`
**Fix branch:** `fix/approval-5c-tenant-isolation-c9-c10-c11`

Same pattern as C10 on the rejection path.

Fix: identical fix.

---

### 5c-C12 — `ListInboxItems` and `CountPendingForActor` query `db *sql.DB` directly without GUC context → RLS bypass

**Files:** `internal/modules/documents/approval/application/read_service.go:163-199, 231-247`
**Fix branch:** `fix/approval-5c-rls-bypass-c12`

Both methods call `db.QueryContext` directly on the connection pool, bypassing the `setAuthzGUC` setup that every other read path performs inside a transaction. If RLS policies on `approval_instances` or `approval_stage_instances` key on `current_setting('metaldocs.tenant_id')`, these queries return unscoped results. Additionally `ListInboxItems` signoff count subquery lacks `AND s.actor_tenant_id = ai.tenant_id`.

Fix: open a transaction, call `setAuthzGUC`, execute inside the transaction, commit — consistent with `LoadInstance` and `ListPendingForActor`.

---

## High

### 5c-H1 — `ComputeEffectiveDenominator` called with snapshot as both args → drift policies never applied at decision time

**File:** `internal/modules/documents/approval/application/decision_service.go:245`

```go
effectiveDenominator := domain.ComputeEffectiveDenominator(*activeStage, activeStage.EligibleActorIDs)
```

Both arguments are `activeStage.EligibleActorIDs` (the snapshot). The function trivially returns the full snapshot count; `DriftReduceQuorum` and `DriftFailStage` policies are never exercised at decision time. Additionally, if both fallbacks exhaust to 0, the code sets `effectiveDenominator = 1`, meaning a fully-drifted stage with no eligible actors requires only 1 signoff from any actor — a silent privilege escalation.

Fix: resolve the current live eligible set from the DB at decision time and pass it as the second argument, or if drift is intentionally frozen at submit time, use `activeStage.EffectiveDenominator` directly and document why. Remove the `effectiveDenominator = 1` fallback — this should return a drift error.

---

### 5c-H2 — `StageInstanceID` match optional → stage substitution attack

**File:** `internal/modules/documents/approval/application/decision_service.go:156-159`

```go
if req.StageInstanceID != "" && activeStage.ID != req.StageInstanceID {
```

A client omitting `StageInstanceID` silently signs any currently-active stage without specifying which one. In multi-route flows, stage identity matters for audit trail and SoD cross-stage tracking.

Fix: require the field at the handler boundary; return 400 if absent.

---

### 5c-H3 — `approval_stage_instances` UPDATE in cancel missing tenant_id join

**File:** `internal/modules/documents/approval/application/cancel_service.go:127-136`

Stage instances updated by `approval_instance_id` only — no tenant scope join. Same cross-tenant risk as C9.

Fix: join through `approval_instances` to enforce tenant scope:
```sql
UPDATE approval_stage_instances asi
   SET status = 'cancelled'
  FROM approval_instances ai
 WHERE asi.approval_instance_id = ai.id AND ai.tenant_id = $2
   AND asi.approval_instance_id = $1 AND asi.status IN ('active','pending')
```

---

### 5c-H4 — `ValidateLegacyCutoverReady` queries without authz bypass / RLS context

**File:** `internal/modules/documents/approval/application/cutover_service.go:41-51`

`db.QueryRowContext` called directly on pool with no `authz.BypassSystem`. If RLS is active on `documents`, this returns 0 (wrong answer) when called without superuser or a GUC-set context, silently authorising the cutover.

Fix: execute inside a transaction with `authz.BypassSystem`, or via a `SECURITY DEFINER` SQL function. Document the access requirement.

---

### 5c-H5 — `approval_route_stages` query missing tenant_id scope

**File:** `internal/modules/documents/approval/application/submit_service.go:271-303`

`loadRoute` fetches stages with only `WHERE route_id = $1`. If a `route_id` UUID ever collides across tenants, this returns another tenant's stage configuration — driving the wrong eligible actors, quorum policy, and capabilities for the entire approval workflow.

Fix: join through `approval_routes` with `AND ar.tenant_id = $2`:
```sql
FROM approval_route_stages ars
JOIN approval_routes ar ON ar.id = ars.route_id AND ar.tenant_id = $2
WHERE ars.route_id = $1
```

---

### 5c-H6 — `loadStageSignoffs` no tenant_id scope

**File:** `internal/modules/documents/approval/application/decision_service.go:451-466`

```sql
SELECT ... FROM approval_signoffs WHERE stage_instance_id = $1
```

No tenant guard. Should join through `approval_stage_instances → approval_instances` to assert tenant.

---

### 5c-H7 — IAM user display name lookup no tenant_id scope

**File:** `internal/modules/documents/approval/application/decision_service.go:194`

```sql
SELECT display_name FROM metaldocs.iam_users WHERE user_id = $1
```

No `tenant_id` predicate. If `iam_users` is shared across tenants, returns the first matching user regardless of tenant.

Fix: add `AND tenant_id = $2` using `req.TenantID`.

---

### 5c-H8 — `tx.Rollback()` called after `tx.Commit()` fails in scheduler

**File:** `internal/modules/documents/approval/application/scheduler_service.go:158-159`

After `Commit()` returns an error the transaction is terminated. The subsequent `Rollback()` is a no-op or driver error and misleads readers into thinking rollback is meaningful here.

Fix: remove `_ = tx.Rollback()` on the commit error branch. Additionally: add `defer tx.Rollback()` immediately after `BeginTx` in `RunScheduledPublishJob` — a deferred rollback on an already-committed transaction is a safe no-op in `database/sql`.

---

### 5c-H9 — `RowsAffected()` error discarded in `submit_service.go`

**File:** `internal/modules/documents/approval/application/submit_service.go:197`

```go
if n, _ := res.RowsAffected(); n == 0 {
```

If the driver returns an error, `n == 0` triggers `ErrStaleRevision` regardless of whether the update succeeded.

Fix:
```go
n, err := res.RowsAffected()
if err != nil {
    _ = tx.Rollback()
    return SubmitResult{}, fmt.Errorf("submit: rows affected: %w", err)
}
```

---

### 5c-H10 — `json.Marshal` error discarded in cancel event payload

**File:** `internal/modules/documents/approval/application/cancel_service.go:164`

Same pattern as C4 — cancel governance event recorded with empty payload on marshal failure.

Fix: check and return the error.

---

### 5c-H11 — `sql.ErrNoRows` compared with `!=` instead of `errors.Is`

**File:** `internal/modules/documents/approval/application/decision_service.go:194`

```go
if err != nil && err != sql.ErrNoRows {
```

Inconsistent with rest of codebase; breaks silently if driver wraps `sql.ErrNoRows`.

Fix: `!errors.Is(err, sql.ErrNoRows)`

---

### 5c-H12 — Read-only service methods open read-write transactions

**File:** `internal/modules/documents/approval/application/read_service.go:38-60, 63-85`

`LoadInstance` and `LoadActiveInstanceByDocument` open full read-write transactions for single read queries. Adds lock overhead; inconsistent with `ListInboxItems` which queries directly.

Fix: use `&sql.TxOptions{ReadOnly: true}` or `db.QueryRowContext` directly.

---

### 5c-H13 — N+1 query: `LoadInstance` called per result row in `ListPendingForActor`

**File:** `internal/modules/documents/approval/application/read_service.go:88-148`

Query returns N IDs; then calls `LoadInstance` per ID (each = 2 SQL round-trips). 50 pending items = 100+ queries per request.

Fix: add `LoadInstanceBatch(ctx, tx, tenantID, ids)` using `WHERE id = ANY($1)`, or retire `ListPendingForActor` in favour of `ListInboxItems`-style single JOIN.

---

### 5c-H14 — Inbox reads missing capability check → any actor can enumerate pending approvals

**File:** `internal/modules/documents/approval/application/read_service.go:88-219`

`ListPendingForActor` and `ListInboxItems` filter by `eligible_actor_ids` JSONB containment but neither calls `authz.Require` before executing. An actor with a revoked role but a known `actorID` can still enumerate all pending instances for their tenant.

Fix: require a minimum `workflow.inbox.view` capability before executing the inbox query.

---

### 5c-H15 — Second `authz.Require(CapDocumentEdit)` placed after instance state is written

**File:** `internal/modules/documents/approval/application/decision_service.go:293-296`

`UpdateInstanceStatus` is called at line 288 (instance marked approved); then `authz.Require(CapDocumentEdit)` is called. If the edit capability check fails, the transaction rolls back — but the failure reveals to the caller via error type that their signoff did complete the quorum.

Fix: move both `authz.Require` calls to before any state-transition writes.

---

### 5c-H16 — Deactivated routes accepted for new submissions

**File:** `internal/modules/documents/approval/application/submit_service.go:261-303`

`loadRoute` queries `WHERE id = $1 AND tenant_id = $2` with no `AND active = TRUE`. A caller with a known deactivated route ID can submit a new approval instance against it.

Fix: add `AND active = TRUE` to the route SELECT in `loadRoute`.

---

### 5c-H17 — SoD check missing `actor_tenant_id` validation

**File:** `internal/modules/documents/approval/application/decision_service.go:431-447`

`loadPriorSignoffs` filters by `actor_tenant_id = $3`; `domain.CheckSoD` then compares only `ActorUserID`. If `loadPriorSignoffs` ever loses its tenant filter, SoD passes for cross-tenant actors with colliding user IDs. `CheckSoD` has no internal tenant guard.

Fix: add `tenantID string` parameter to `CheckSoD` and assert `s.ActorTenantID() == tenantID` before the user ID comparison.

---

### 5c-H18 — `ListInboxItems` signoff count subquery missing tenant_id

**File:** `internal/modules/documents/approval/application/read_service.go:163-199`

```sql
SELECT count(*) FROM approval_signoffs s
WHERE s.approval_instance_id = ai.id AND s.stage_instance_id = asi.id AND s.decision = 'approve'
```

No `AND s.actor_tenant_id = ai.tenant_id`. Cross-tenant signoffs could inflate/deflate the count.

Fix: add `AND s.actor_tenant_id = ai.tenant_id` inside the subquery.

---

### 5c-H19 — `ListPendingForActor` uses OFFSET pagination → full table scan per page

**File:** `internal/modules/documents/approval/application/read_service.go:104-114`

```sql
ORDER BY ai.id LIMIT $4 OFFSET $5
```

Page 10 at offset 250 scans 260 rows to discard 250.

Fix: keyset/cursor pagination — `WHERE ai.id > $last_seen_id`.

---

### 5c-H20 — `resolveEligibleActors` called per-stage in submit transaction with no LIMIT

**File:** `internal/modules/documents/approval/application/submit_service.go:329-360`

Unbounded query against `metaldocs.user_process_areas` per stage. For a role with thousands of members, each submit call can return an oversized in-memory slice. No guard against empty result silently creating a zero-eligible stage.

Fix: add `LIMIT 500` with an error if the limit is hit. Document RLS expectations for this table.

---

## Medium

### 5c-M1 — `effectiveDenominator` falls back to `1` on full drift → any signoff satisfies quorum

**File:** `internal/modules/documents/approval/application/decision_service.go:244-253`

Two sequential fallbacks:
```go
if effectiveDenominator == 0 { effectiveDenominator = len(activeStage.EligibleActorIDs) }
if effectiveDenominator == 0 { effectiveDenominator = 1 }
```
When the entire eligible pool is gone, denominator is set to 1 — meaning any single signoff satisfies quorum silently. Should fail the stage or return an error.

---

### 5c-M2 — Unknown `DriftPolicy` silently falls through to snapshot denominator with no Reason

**File:** `internal/modules/documents/approval/domain/drift.go:66`

```go
default:
    return DriftResult{EffectiveDenominator: len(stage.EligibleActorIDs), ForcedOutcome: QuorumPending}
```

Misconfigured policy (typo, undeployed new value) silently uses snapshot denominator. No `Reason` field set.

Fix: populate `Reason` with a diagnostic string describing the unknown-policy fallback.

---

### 5c-M3 — `publish_service.go:PublishApproved` missing OCC revision_version guard

**File:** `internal/modules/documents/approval/application/publish_service.go:96-117`

UPDATE uses only `AND status = 'approved'` as concurrency guard. Two concurrent `PublishApproved` calls on the same document will both see `status = 'approved'` before either commits. `SchedulePublish` correctly uses `AND revision_version = $5`; `PublishApproved` omits it.

Fix: add `AND revision_version = expectedRevision` sourced from the loaded instance.

---

### 5c-M4 — `process_area_code_snapshot` scanned into bare `string` → NULL scan error

**File:** `internal/modules/documents/approval/application/cancel_service.go:79-80`

If the column is NULL, scanning into `string` fails. `submit_service.go` handles this correctly with `loadDocumentAreaCode` helper using `sql.NullString`.

Fix: use the shared `loadDocumentAreaCode` helper.

---

### 5c-M5 — `QuorumMofN` switch lacks `default` → unknown quorum type stalls forever silently

**File:** `internal/modules/documents/approval/domain/quorum.go:81-95`

Unknown quorum type silently returns `QuorumPending` outside the switch — indistinguishable from "no votes yet".

Fix: add `default` case inside the switch that logs/panics or returns a named error.

---

### 5c-M6 — `EmptyEligibleSet` quorum fallback allows any signoff to satisfy

**File:** `internal/modules/documents/approval/domain/quorum.go:41-48`

```go
if len(stage.EligibleActorIDs) == 0 {
    approveCount = len(approvals)
    rejectCount = len(rejections)
}
```

Misconfigured route with empty `RequiredRole` creates a stage with no eligible actors; any actor who can post a signoff satisfies quorum.

Fix: `Route.Validate()` must reject stages with both `RequiredRole` and `RequiredCapability` empty.

---

### 5c-M7 — `LoadInstance` accepts `actorID` parameter but never uses it

**File:** `internal/modules/documents/approval/application/read_service.go:38`

Dead parameter silently implies actor-scoping that does not exist.

Fix: remove `actorID` from the signature, or implement the authz check (which should be done per C7).

---

### 5c-M8 — `scheduler_service.go` silent discards for missing document, stale job, pre-effective date

**Files:** `internal/modules/documents/approval/application/scheduler_service.go:54-69`

Three separate early-return `nil` paths with no log statement. River job completes without error — no retry, no trace.

Fix: log at INFO/WARN with document ID and specific reason before each early return.

---

### 5c-M9 — `decision_service.go` not wrapped with `authz.WithCapCache` → 3 redundant capability round-trips per signoff

**File:** `internal/modules/documents/approval/application/decision_service.go:91`

Multiple `authz.Require` calls (lines 144, 293, 357) with no `WithCapCache` wrapping — each hits the DB independently.

Fix: add `ctx = authz.WithCapCache(ctx)` before `setAuthzGUC`, matching `submit_service.go`.

---

### 5c-M10 — `setAuthzGUC` called after `LoadInstance` in publish/cancel services

**Files:** `internal/modules/documents/approval/application/publish_service.go:46-73`, `cancel_service.go:47-103`

`LoadInstance` executes before GUC context is set — queries run without RLS context.

Fix: move `setAuthzGUC` to immediately after `BeginTx`, before any repository call.

---

### 5c-M11 — `MarshalJSON` on `Signoff` omits `actor_display_name_snapshot`

**File:** `internal/modules/documents/approval/domain/signoff.go:115-128`

Field populated but excluded from JSON output → API consumers never see who signed.

Fix: add `"actor_display_name_snapshot": s.actorDisplayNameSnapshot` to the marshal map.

---

### 5c-M12 — `BumpRevisionVersion` allows equal value (no-op) but name implies forward progress

**File:** `internal/modules/documents/approval/domain/instance.go:182-188`

`next == current` silently succeeds. The DB trigger may also allow this — verify. If strictly monotonic is intended, change guard to `next <= current`.

Fix: rename to `SetRevisionVersion` with explicit semantics doc, or tighten the guard.

---

### 5c-M13 — Hand-rolled `itoa` to avoid importing `strconv` in domain package

**File:** `internal/modules/documents/approval/domain/drift.go:71-94`

`strconv` is standard library with no external dependencies — perfectly appropriate in domain code.

Fix: delete `itoa`, use `strconv.Itoa(delta)`.

---

### 5c-M14 — `GovernanceEvent.EventType` / `ResourceType` are bare `string` with no shared type

**File:** `internal/modules/documents/approval/application/events.go:11-19`

Event types scattered as string literals across 6 service files. A typo produces a governance record with unrecognised event type that audit queries silently miss.

Fix: define `type EventType string` with constants for all event types used in the module.

---

### 5c-M15 — `GovernanceEvent.OccurredAt` never included in the INSERT

**File:** `internal/modules/documents/approval/application/events.go:35-37`

Field present in struct but omitted from the SQL. DB defaults to `now()` rather than the service-supplied timestamp.

Fix: include `occurred_at` in the INSERT, or remove the field from the struct.

---

### 5c-M16 — `SignoffRequest.Decision` is bare `string` instead of `domain.Decision`

**File:** `internal/modules/documents/approval/application/decision_service.go:70-79`

Invalid values (e.g. `"abstain"`) pass through `ValidateEventPayload` and fail deep in the transaction after the GUC and locks are set.

Fix: change `SignoffRequest.Decision` to `domain.Decision`; validate at handler boundary.

---

### 5c-M17 — `ComputeContentHash` called with `DocumentID: req.InstanceID` — semantic field mismatch

**File:** `internal/modules/documents/approval/application/decision_service.go:98-104`

Field named `DocumentID` in `ContentHashInput` is used to hold an instance ID. Silent documentation/naming bug.

Fix: rename to `EntityID` with godoc, or add a dedicated `ComputeSignoffContentHash` function.

---

### 5c-M18 — `cancel_service.go:61-165` uses bare `tx.Rollback()` without `_ =`

**File:** `internal/modules/documents/approval/application/cancel_service.go`

14+ call sites with bare `tx.Rollback()`. Triggers `go vet` / `staticcheck` warnings.

Fix: replace all with `_ = tx.Rollback()`.

---

### 5c-M19 — `ErrFloatInPayload` and `ErrFloatInFormData` are duplicate sentinels for the same condition

**File:** `internal/modules/documents/approval/application/services.go:26`, `content_hash.go:15`

Two different sentinel errors for float-in-map condition. Callers must know which error to check.

Fix: consolidate to `ErrFloatInFormData`; have `ValidateEventPayload` return it.

---

### 5c-M20 — `newCancelService` constructor defined but never called in `NewServices`

**File:** `internal/modules/documents/approval/application/cancel_service.go:189`

`NewServices` constructs `CancelService` directly via struct literal, ignoring the dedicated constructor.

Fix: use `newCancelService(...)` in `NewServices`, or remove the dead constructor.

---

### 5c-M21 — `ListInboxItems` runs outside transaction while `ListPendingForActor` uses one — inconsistent read semantics

**File:** `internal/modules/documents/approval/application/read_service.go:63-85, 88-148`

`ListPendingForActor` uses `db.BeginTx`; `ListInboxItems` queries the pool directly. Concurrent writes could cause the inbox query to observe partially-committed state.

Fix: consistent — either both use `BeginTx(ReadOnly: true)` or both use pool directly. The GUC context fix (C12) forces the transaction path.

---

## Low

### 5c-L1 — `isTerminal()` on `Instance` defined but never called

**File:** `internal/modules/documents/approval/domain/instance.go:203`

Fix: use it in `Cancel()` or delete it.

---

### 5c-L2 — `snapshot` map in `ApplyEligibilityDrift` is dead allocation

**File:** `internal/modules/documents/approval/domain/drift.go:13-16`

Built from `stage.EligibleActorIDs` but never read — the function only reads `current[id]`.

Fix: delete the `snapshot` map and its construction loop.

---

### 5c-L3 — `state.go:44` error message concatenated from user input with `errors.New`

**File:** `internal/modules/documents/approval/domain/state.go:44`

```go
return "", errors.New("unknown document state: " + s)
```

Fix: `fmt.Errorf("unknown document state: %q", s)` — wrappable, quoted, no concatenation.

---

### 5c-L4 — `IsLegalTransition` defined but never called before SQL state updates

**File:** `internal/modules/documents/approval/domain/state.go:60-70`

The domain transition graph exists but application services bypass it, relying entirely on the DB trigger.

Fix: call `domain.IsLegalTransition(from, to)` in aggregate methods or services before issuing SQL, making the domain the first enforcement layer.

---

### 5c-L5 — `quorum_test.go` and `signoff_test.go` discard `NewSignoff` errors with `_`

**Files:** `internal/modules/documents/approval/domain/quorum_test.go:10`, `signoff_test.go`

If `NewSignoff` fails, `s` is nil and `*s` panics with an opaque nil-pointer instead of a test failure message.

Fix: `s, err := NewSignoff(...); if err != nil { t.Fatal(err) }` with `t.Helper()`.

---

### 5c-L6 — `signoff_test.go` assertions give no expected/got context

**File:** `internal/modules/documents/approval/domain/signoff_test.go:117-125`

```go
if s.ID() != "sig-1" { t.Error("ID") }
```

No diagnostic when test fails.

Fix: `t.Errorf("ID() = %q; want %q", s.ID(), "sig-1")`

---

### 5c-L7 — Route not found error loses `sql.ErrNoRows` sentinel (unwrapped)

**File:** `internal/modules/documents/approval/application/submit_service.go:266`

```go
return domain.Route{}, fmt.Errorf("route %s not found for tenant %s", routeID, tenantID)
```

Original `sql.ErrNoRows` not wrapped with `%w` — `errors.Is` callers won't match.

Fix: wrap with a named sentinel: `fmt.Errorf("route %s not found: %w", routeID, repository.ErrNotFound)`.

---

### 5c-L8 — `rows.Err()` returned without wrapping context in `ListInboxItems`

**File:** `internal/modules/documents/approval/application/read_service.go:219`

All other error returns wrap with `"list inbox: ..."`. This one doesn't.

Fix: `if err := rows.Err(); err != nil { return nil, fmt.Errorf("list inbox: rows: %w", err) }`

---

### 5c-L9 — `idempotency.go` accepts empty `ActorUserID`/`DocumentID`/`Decision` without error

**File:** `internal/modules/documents/approval/application/idempotency.go`

Empty required fields produce valid-but-colliding idempotency keys.

Fix: validate required fields before computing; return error if any are empty.

---

### 5c-L10 — `pdfDispatcher` deprecated but no `//Deprecated:` tag; no startup warning if both `pdfOutbox` and `pdfDispatcher` are nil

**File:** `internal/modules/documents/approval/application/decision_service.go:46-59`

If both are nil, PDF is silently skipped on approval with no log.

Fix: add startup check in `NewDecisionService`; add `//Deprecated:` godoc to the field.

---

### 5c-L11 — `EvaluateQuorum` takes full `StageInstance` but uses only 3 fields

**File:** `internal/modules/documents/approval/domain/quorum.go:33`

Leaky API; also takes `approvals []Signoff` and `rejections []Signoff` pre-split by caller while doing its own eligible filtering.

Fix: accept `eligibleActorIDs []string`, `policy QuorumPolicy`, `quorumM *int` and a single `[]Signoff`; own the split internally.

---

### 5c-L12 — `domain/sod.go` `CheckSoD` takes `[]Signoff` by value — missing tenant cross-check

**File:** `internal/modules/documents/approval/domain/sod.go:15`

Relies entirely on caller having pre-filtered to correct tenant. If the caller ever omits the filter, SoD passes silently.

Fix: add `tenantID string` parameter to `CheckSoD`; assert `s.ActorTenantID() == tenantID`.

---

### 5c-L13 — `scheduledDocumentState` and `scheduledPublishState` are duplicate structs with manual field-by-field copy

**File:** `internal/modules/documents/approval/application/scheduler_service.go:32-41, 86-94`

If a field is added to one and forgotten in the other, the scheduler operates on stale data silently.

Fix: merge into one struct or make one embed the other.

---

## Critical Backlog — Fix Branches

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
