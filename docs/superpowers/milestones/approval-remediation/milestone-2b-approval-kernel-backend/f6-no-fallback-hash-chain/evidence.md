# Feature F6 — `no-fallback-hash-chain` — evidence.md

Ref: `spec.md`, `plan.md` (this folder). Parent HEAD before this feature: `5739c76b` (F5 freeze boundary).

## Implementation summary

Removed the two production COALESCE-over-hash expressions identified in spec.md Interview #1/#7 and
replaced them with explicit, status-scoped reads that never substitute a value:

1. **Repository rename** (`internal/modules/documents/approval/infrastructure/approval_repository.go`,
   `postgres_approval_repository.go`, `errors.go`): `LoadActiveDocumentContentHash(ctx, tx, tenantID,
   documentID)` → `LoadFrozenContentHash(ctx, tx, tenantID, instanceID)`. New body is a bare
   `SELECT frozen_content_hash FROM approval_instances WHERE id=$1 AND tenant_id=$2` — no COALESCE, no
   `document_revisions` subquery. Missing row or NULL pin → `ErrNoActiveContentHash` (existing sentinel,
   reused, doc comment updated).

2. **Signoff call site** (`decision_service.go`, `RecordSignoff`): now calls
   `s.repo.LoadFrozenContentHash(ctx, tx, req.TenantID, instance.ID)` (was
   `LoadActiveDocumentContentHash(ctx, tx, req.TenantID, instance.DocumentID)`). Error mapping unchanged
   (`ErrNoActiveContentHash` → `ErrContentHashMismatch` → HTTP 412, same as before).

3. **Publish fail-closed guard** (`publish_service.go`, NEW check, both `PublishApproved` and
   `SchedulePublish`, immediately after the existing `instance.Status != domain.InstanceApproved` check):
   ```go
   if instance.FrozenContentHash == nil {
       return ErrContentHashMismatch
   }
   ```
   Additive — no currently-passing real flow reaches `InstanceApproved` without a pin (F4/F5 stage-kind
   gating), so this only detects an otherwise-silent impossible state. Per plan.md Task 3's decision, wraps
   as `application.ErrContentHashMismatch` directly (matches decision_service.go's existing pattern; no new
   `http/errors.go` case needed).

4. **Display path split** (`internal/modules/documents/infrastructure/active_instance_reader.go`,
   `ActiveInstanceForControlledDocument`): deleted the COALESCE from the main FULL OUTER JOIN projection
   (and dropped `content_hash_at_submit` from its subquery entirely). Hash resolution is now two explicit
   steps run after the projection resolves `DocumentID`/`Status`:
   - Step 1: the existing `under_review` in-progress-instance lookup query gained one column
     (`frozen_content_hash`) — same round-trip, same `WHERE` predicate, no new query. If non-NULL, that's
     `view.ContentHash`.
   - Step 2 (fallthrough): if no instance, or the instance is not yet frozen, or the document isn't
     `under_review` at all — a separate `SELECT content_hash FROM document_revisions WHERE document_id=$1
     ORDER BY created_at DESC LIMIT 1` supplies the head-revision hash.

5. **Parity test** (`tests/integration/controlleddocuments/active_instance_parity_test.go`):
   `rawGetActiveInstance`'s baseline updated to replay the same two-step logic (no longer replaying the old
   defect verbatim). Added two new sub-tests: `under_review_with_frozen_instance` (pin is authoritative) and
   `under_review_not_yet_frozen_falls_through_to_head_revision` (correct fallthrough, not a regression of the
   old defect).

6. **testdb factory** (`tests/integration/testdb/factory.go`): `NewApprovalInstance` previously never set
   `frozen_content_hash` (column stayed permanently NULL for every fixture). Added `WithFrozenContentHash
   (hash string)` option (empty string forces NULL — the not-yet-frozen fixture case) and a default:
   any fixture seeded with `status == "approved"` and no explicit override gets a default frozen hash
   (`repeat('a', 64)`), matching the real invariant that an approved instance always has a pin by then.
   `ApprovalInstance` struct gained a `FrozenContentHash *string` field. This was necessary because the new
   publish fail-closed guard (item 3) would otherwise break every existing real-DB integration test that
   seeds an `"approved"` instance and calls `PublishApproved`/`SchedulePublish` expecting success
   (`publish_no_autoversion_integration_test.go`, `publish_race_integration_test.go`,
   `schedule_publish_review_fields_integration_test.go`) — fixed via the factory default with zero edits
   needed to those three test files.

## Test fixtures updated to keep compiling/passing under the renamed method + new guards

- `decision_service_test.go`: `fakeDecisionRepo.LoadActiveDocumentContentHash` → `LoadFrozenContentHash`
  (reads `r.instance.FrozenContentHash`); `buildSingleStageInstance` now sets a non-nil pin; added
  `validContentHashPtr` package var (Go can't take `&` of a `const`); new test
  `TestRecordSignoff_NullFrozenHash_FailsClosed`.
- `phase5_integration_test.go`: `phase5Repo.LoadActiveDocumentContentHash` → `LoadFrozenContentHash`; both
  `inProgressInstance` builders and the mid-flow `approvedInstance` fixture (Step 3, PublishApproved) now
  carry a non-nil `FrozenContentHash`.
- `decision_otel_test.go`, `submit_service_test.go`: method renamed on their stub/panic fakes (no behavior
  change).
- `service_invariants_test.go`: `buildAllCompletedInstance`, `buildTwoStageInstance`,
  `TestPublishApproved_OCC_StaleRevision`, `TestSchedulePublish_OCC_StaleRevision` fixtures given a non-nil
  `FrozenContentHash` — these tests assert a different invariant (no-active-stage, OCC staleness) and would
  otherwise now fail earlier on the new hash check instead of exercising their actual assertion.
- `publish_service_test.go`: all 9 `Status: domain.InstanceApproved` fixture blocks given
  `FrozenContentHash: &validContentHashPtr`; added
  `TestPublishApproved_NullFrozenHash_FailsClosed` and `TestSchedulePublish_NullFrozenHash_FailsClosed`
  (new, TDD-first fail-closed regression guards).
- `internal/test/e2e_seed.go`: comments only — clarifies the seeded `documents.content_hash_at_submit` value
  is no longer read by the real signoff/publish path post-F6.

## Verification sweep

| Command | Result |
|---|---|
| `go build ./...` | clean, no output |
| `go build -tags integration ./...` | clean, no output |
| `go vet -tags integration ./...` | Same 5 import-cycle/package-path findings as on parent HEAD `5739c76b` pre-F6 (confirmed via `git stash`/`git stash pop` diff-free comparison) — pre-existing, not a regression, unrelated to F6 |
| `go test -count=1 ./internal/modules/documents/approval/...` | all packages `ok` |
| `go test -count=1 ./internal/modules/documents/...` | all packages `ok` |
| `go test -count=1 ./...` | zero `FAIL` lines across full suite (141-line log, tail confirmed `ok`) |
| `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` | `0 violation(s)` |
| `go test ./scripts/api-lint/...` | `ok` |
| `grep -rn "COALESCE(d.content_hash_at_submit\|COALESCE(active.content_hash_at_submit" internal/` | zero hits |

## Review disposition (judgment calls)

1. **Publish error mapping** — returned `ErrContentHashMismatch` directly from `publish_service.go` (same
   package, already has this sentinel) rather than a raw `infrastructure.ErrNoActiveContentHash` requiring a
   new `http/errors.go` case. Matches `decision_service.go`'s existing pattern exactly. Per plan.md Task 3's
   own pre-recorded decision.
2. **`SchedulePublish` also guarded** — plan.md's Task 3 literal text only names `PublishApproved`, but
   spec.md's Consumer contract and the locked principle ("publish is the final link in the canonical hash
   chain") apply to both publish-family transitions off an `InstanceApproved` instance. Applied the same
   guard to `SchedulePublish` for consistency; flagged here as a scope note rather than silently expanding
   without a rationale.
3. **testdb factory default-seed** — not explicitly specified in plan.md, but required to avoid the new
   publish guard silently breaking three pre-existing real-DB integration tests. Chose a default
   (`status == "approved"` → auto-seed a valid hash unless overridden) over editing all three test files
   individually, since it encodes the actual domain invariant ("approved implies frozen") at the fixture
   level rather than patching each call site.
4. **`documents.content_hash_at_submit` column** — left in place per spec.md's Non-goals (schema-minimization
   is a separate, larger decision). Confirmed via grep it has zero remaining readers in the touched read path
   after this change (only `documents/application/service.go`'s write path remains, out of scope).
5. **`view.ContentHash` fallthrough correctness** — the not-yet-frozen fallthrough intentionally returns the
   SAME value the old (defective) COALESCE would have returned for a genuinely-unfrozen in-progress instance
   — this is not a regression of the fixed defect, because the defect was specifically the COALESCE firing
   for an ALREADY-frozen instance (silently overridden by a later head-revision hash). Verified via the new
   `under_review_not_yet_frozen_falls_through_to_head_revision` parity sub-test.

## Bounded defers (unchanged from spec.md, restated for closure)

| Defer | Rerun command | Owner |
|---|---|---|
| Live-DB run of `active_instance_parity_test.go`'s new/updated sub-tests and the three publish-family integration tests that now rely on the factory's auto-seeded pin | `go test -tags integration -run TestActiveInstanceReader_ParityWithRawGetActiveInstance -v ./tests/integration/controlleddocuments/...` and `go test -tags integration -run "TestPublish|TestSchedulePublish" -v ./internal/modules/documents/approval/application/...` against a running local Postgres | Next session with live DB access, before this feature is considered fully closed at the milestone level |
| `documents.content_hash_at_submit` column left in place with one fewer production reader | N/A — future dedicated cleanup task once all cross-module readers/writers are grep-confirmed absent | Unassigned, low priority (F10 cleanup candidate) |

## Scope confirmation

Touched exactly the files named in plan.md's task list plus the testdb factory (judgment call #3 above,
required to keep pre-existing integration tests from silently breaking under the new guard) and the fixture
files needed to keep the touched production code compiling/passing. No adjacent refactor, no unrelated file
touched. No `.env` read. No live DB probed.
