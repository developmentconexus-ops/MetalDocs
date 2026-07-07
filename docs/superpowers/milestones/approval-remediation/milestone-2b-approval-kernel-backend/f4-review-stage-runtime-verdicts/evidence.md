# Feature F4 — Evidence

> **Milestone:** 2b — Approval Kernel Backend  ·  **Feature:** `f4-review-stage-runtime-verdicts`  ·
> **Closed:** 2026-07-07
> **Contract:** `spec.md` / `plan.md`

## What was implemented

### Domain (`internal/modules/documents/approval/domain/`)
- `review_verdict.go` (new): `Verdict` type (`VerdictReady`/`VerdictRequestChanges`),
  `ErrVerdictWrongStageKind`, `ErrVerdictCommentRequired`, immutable `ReviewVerdict` value object
  (unexported fields + getters), `VerdictParams`, `NewVerdict(p VerdictParams) (*ReviewVerdict, error)`
  — validates all required fields, rejects non-review `StageKind`, requires a non-empty `Comment` for
  `request_changes`. `MarshalJSON` for API-shape serialization.
- `review_verdict_test.go` (new): 6 unit tests (ready/request_changes-with-comment/without-comment/
  wrong-stage-kind/invalid-verdict-value/table of 6 required-field cases).
- `instance.go`: added `InstanceChangesRequested InstanceStatus = "changes_requested"` (5th,
  non-terminal status). `Cancel(reason string)` now stores `inst.CancelReason = &reason` (previously
  discarded in-memory too, not just at the DB layer). Deleted `SkipStage` method and
  `ErrCannotSkipLastStage` outright (design spec §2.4, W11) — zero call sites remain.
- `instance_test.go`: added `TestInstanceCancelStoresReason`, `TestInstanceChangesRequestedIsNonTerminal`;
  removed no tests here (the `SkipStage`-specific test lived in `integration_test.go`, see below).
- `integration_test.go`: deleted `TestSkipStageDuringSigning` (R2-1) — the only other test exercising
  the deleted `SkipStage` method.

### Migration
- `db/migrations/0288_review_verdicts_changes_requested.sql` (new): widens
  `approval_instances_status_check` to add `'changes_requested'`; creates
  `approval_review_verdicts` table (id, approval_instance_id, stage_instance_id, actor_user_id text,
  actor_tenant_id uuid, verdict text CHECK IN `ready`/`request_changes`, comment, verdict_at,
  actor_display_name_snapshot, `UNIQUE(stage_instance_id, actor_user_id)`); RLS
  `ENABLE`+`FORCE`+`tenant_isolation` policy copied verbatim from `0285_approval_signoffs_rls.sql`'s
  idiom; ledger row inserted into `schema_migrations`.

### Repository (`internal/modules/documents/approval/infrastructure/`)
- `approval_repository.go`: `VerdictInsertResult{ID string; WasReplay bool}`; `ApprovalRepository`
  interface extended with `UpdateInstanceStatusWithReason(...)`, `InsertVerdict(...)`,
  `LoadStageVerdicts(...)`.
- `postgres_approval_repository.go`: `UpdateInstanceStatusWithReason` (OCC UPDATE, `cancel_reason = $3`
  added to SET clause — this is the concrete fix for the cancel-reason persistence gap);
  `InsertVerdict` (`ON CONFLICT (stage_instance_id, actor_user_id) DO NOTHING RETURNING id`, mirrors
  `InsertSignoff` exactly, reuses `ErrActorAlreadySigned` sentinel on field-mismatch replay-vs-conflict
  detection); `loadVerdictByStageActor` / `scanVerdict` (private helpers — `scanVerdict` hardcodes
  `StageKind: domain.StageKindReview` since the row itself carries no stage_kind column, safe because
  only review-kind stages can ever produce a verdict row); `LoadStageVerdicts` (`ORDER BY verdict_at ASC`).

### Application (`internal/modules/documents/approval/application/`)
- `review_verdict_service.go` (new, ~420 lines): `ReviewVerdictService.RecordVerdict` — mirrors
  `DecisionService.RecordSignoff` structurally: off-tx `LoadActorDisplayName` (H-PRE-1) → in-tx
  `LoadInstance` (FOR UPDATE) → OCC (`ExpectedRevisionVersion`) → `in_progress`-only guard →
  `docapp.LoadDocumentAreaCode` + `authz.Require(CapApprovalReview, areaCode)` → active-stage +
  stage-id match + `Kind != StageKindReview` → `ErrVerdictWrongStageKind` → `domain.CheckEligibility`
  (governance-event-emitting failure path) → SoD author-check (`instance.SubmittedBy ==
  req.ActorUserID` → `ErrAuthorCannotSign`) → `domain.NewVerdict` + `repo.InsertVerdict`
  (replay-aware) → verdict-kind switch:
  - `ready`: `LoadStageVerdicts` → `verdictsAsApprovals` adapter → `ApplyEligibilityDrift` +
    `domain.EvaluateQuorum` (both reused unchanged) → on `QuorumApprovedStage`: complete stage,
    `instance.AdvanceStage()`; if instance now `InstanceApproved`: unresolved-comments gate,
    `UpdateInstanceStatus`, `authz.Require(CapDocumentEdit)`,
    `docsdomain.CanTransitionDocumentStatus(UnderReview, Approved)` + OCC UPDATE to `approved`,
    lifecycle-enqueue `EventTypeDocumentApproved`; else activate next stage.
  - `request_changes`: `UpdateInstanceStatus(InstanceChangesRequested)` (no quorum needed — a single
    verdict collapses the instance), `SET LOCAL metaldocs.cancel_in_progress` GUC,
    `authz.Require(CapDocumentEdit)`, `CanTransitionDocumentStatus(UnderReview, Draft)` + OCC UPDATE to
    `draft`. No lifecycle event enqueued (non-terminal transition — no
    `EventTypeDocumentChangesRequested` const exists, matching spec.md's non-goals framing). **Post-review
    fix:** the first-pass implementation called `UpdateInstanceStatusWithReason(..., reason=Comment)`,
    writing the reviewer's comment into `approval_instances.cancel_reason` — a column reserved for actual
    cancellations. Caught in spec-compliance review: the comment is already durably recorded on the
    `approval_review_verdicts` row inserted earlier in the same tx, so writing it again into a
    differently-named column was redundant AND semantically misleading (a `changes_requested` instance
    would show a populated `cancel_reason`, implying it was cancelled). Fixed to call plain
    `UpdateInstanceStatus` (no reason param) for the request_changes path; `UpdateInstanceStatusWithReason`
    remains exclusively for `CancelInstance`.
  - Governance event emitted in both cases (`EventTypeReviewVerdictRecorded` /
    `EventTypeReviewChangesRequested`).
- `events.go`: two new `EventType` consts, `EventTypeReviewVerdictRecorded` /
  `EventTypeReviewChangesRequested`.
- `services.go`: `Services.ReviewVerdict *ReviewVerdictService` field; wired in `NewServices` and
  `WithLifecycleEnqueuer`.
- `cancel_service.go`: `CancelInstance` now calls `UpdateInstanceStatusWithReason(...,
  in.Reason)` instead of `UpdateInstanceStatus(...)` — closes the F4 cancel-reason persistence gap
  (spec.md Interview #1: the column existed since F1/migration 0286, was SELECTed, never written).
- `cancel_service_test.go`: `cancelFakeRepo` gained `receivedReason` field + an explicit
  `UpdateInstanceStatusWithReason` override (needed — the fake embeds a nil
  `infrastructure.ApprovalRepository`, so the new interface method would otherwise panic on a nil-embed
  call); `TestCancelInstance_HappyPath` asserts `receivedReason == "stakeholder withdrew"`.
- `decision_otel_test.go`: `stubApprovalRepo` gained 3 trivial stub methods
  (`UpdateInstanceStatusWithReason`, `InsertVerdict`, `LoadStageVerdicts`) to satisfy the widened
  `ApprovalRepository` interface.

### HTTP (`internal/modules/documents/approval/http/`)
- `contracts/review_verdict.go` (new): `Verdict` string type, `ReviewVerdictRequest{Verdict, Comment}`
  with `Validate()` (enum + comment-required-for-request_changes), `ReviewVerdictResponse{VerdictID,
  WasReplay, Outcome}`.
- `review_verdict_handler.go` (new): `ReviewVerdictHandler` mirrors `SignoffHandler` — tenant/actor
  extraction, required `Idempotency-Key`, `parseIfMatch` (OCC precondition), strict JSON decode +
  `Validate()`, verdict-enum switch, `reviewVerdictPayloadHash` (sha256 misuse-guard, mirrors
  `signoffPayloadHash`), self-managed replay via `idempStore.BeginStageReplay` (same pattern as
  signoffs — NOT the router-level idempotency middleware), `reviewVerdictOutcome` mapping
  (`approved` > `changes_requested` > `stage_completed` > `pending`).
- `handler.go`: `reviewVerdictService` interface, `Handler.reviewVerdictSvc` field, wired in
  `NewHandler`.
- `routes_generated.go` (hand-maintained adapter file): `RecordApprovalReviewVerdict` method
  delegating to `h.ReviewVerdictHandler`.

### Contract-first API (`api/openapi/v1/openapi.yaml` + generated)
- New path `POST /approval/instances/{instance_id}/stages/{stage_id}/review-verdict`
  (`operationId: recordApprovalReviewVerdict`, tag `approval`), full parameter set (`instance_id`,
  `stage_id`, `Idempotency-Key`, `If-Match`), request/response schemas `ReviewVerdictRequest` /
  `ReviewVerdictResponse`, standard 400/401/403/404/409/412/428/500 refs, `x-authz-area` block noting
  the `approval.review` capability.
- `internal/modules/documents/approval/api/api.gen.go` regenerated via `go generate
  ./internal/modules/documents/approval/api/...` (never hand-edited) — new types
  `ReviewVerdictRequest`/`ReviewVerdictRequestVerdict` (with `.Valid()`), `ReviewVerdictResponse`,
  `RecordApprovalReviewVerdictParams`, strict-server request/response object types, and the
  `ServerInterface.RecordApprovalReviewVerdict` method signature (which forced `routes_generated.go`'s
  adapter to exist — the generated interface fails to compile without a `Handler` method for every
  spec'd operation).

### AuthZ (tier-1 + tier-2)
- `apps/api/cmd/metaldocs-api/permissions.go`: new tier-1 row —
  `POST /api/v1/approval/instances/{id}/review-verdict → CapApprovalReview` (closes F3's documented
  defer: `CapApprovalReview` previously had no live route).
- Tier-2: `authz.Require(CapApprovalReview, areaCode)` inside `RecordVerdict` (verdict recording) and
  `authz.Require(CapDocumentEdit, areaCode)` before both document-status transitions (approve / revert
  to draft) — mirrors `RecordSignoff`'s existing pattern exactly, never a role check.
- `scripts/api-lint/tripwire-allowlist.txt`: added
  `internal/modules/documents/approval/infrastructure/postgres_approval_repository.go|UpdateInstanceStatusWithReason`
  immediately after the sibling `UpdateInstanceStatus` line — legitimate TRIPWIRE-PAIRING false
  positive (authz is enforced one layer up in the calling service, same pattern as the existing
  `UpdateInstanceStatus` entry).
- No `internal/platform/tripwire/arms.go` change — `approval_review_verdicts` is a new table absent
  from the `TripwireArms` registry; TRIPWIRE-ARM-DRIFT only checks tables already registered, so it is
  silently unchecked rather than failing. Documented as a bounded defer below (not silently skipped).

### Integration test
- `tests/integration/approval/review_verdict_integration_test.go` (new): builds a
  `reviewVerdictFixture` (tenant, author, reviewer, a document at `under_review` with
  `process_area_code_snapshot` stamped, an `in_progress` approval instance, and a raw-SQL-seeded
  active stage instance of a chosen `stage_kind` — no factory builder yet exists for a review-kind
  stage instance, so this follows `eligibility_test.go`'s raw-SQL FK-chain precedent) and calls
  `ReviewVerdictService.RecordVerdict` directly through the real Postgres-backed repository. Covers:
  `TestReviewVerdict_ReadyAdvancesQuorum`, `TestReviewVerdict_RequestChangesCollapsesInstanceAndDocument`,
  `TestReviewVerdict_RequestChangesRequiresComment`, `TestReviewVerdict_WrongStageKindRejected`,
  `TestReviewVerdict_SoDBlocksSelfVerdict`, `TestCancelInstance_ReasonPersists`.

## Verification

| Check | Command | Result |
|-------|---------|--------|
| Build | `go build ./...` | clean, exit 0 |
| Integration-tag build | `go build -tags integration ./...` | clean, exit 0 |
| Approval module suite | `go test ./internal/modules/documents/approval/...` | all subpackages PASS (domain/application/http/contracts/infrastructure/idempotency/signature/jobs) |
| Permissions/tier-1 suite | `go test ./apps/api/cmd/metaldocs-api/...` | PASS — `TestRouteCoverage` auto-discovered the new tier-1 row with no test-file edits needed |
| api-lint unit suite | `go test ./scripts/api-lint/...` | PASS (after the `UpdateInstanceStatusWithReason` allow-list addition) |
| api-lint strict (all lints incl. TRIPWIRE-ARM-DRIFT/PARITY, TRIPWIRE-PAIRING) | `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` | `0 violation(s)` |
| Full regression | `go test ./...` | all packages PASS (no `[no test files]` package failed; no regressions) |
| Grep-zero: `SkipStage`/`ErrCannotSkipLastStage` | `grep -rn 'SkipStage\|ErrCannotSkipLastStage' --include=*.go .` | zero matches |
| New integration test build/vet | `go vet -tags integration ./tests/integration/approval/...` | clean, exit 0 |
| New integration test **live run** | — | **not run** — see Bounded defers (no `DATABASE_URL`/`METALDOCS_DATABASE_URL` obtainable without reading `.env`, which is forbidden; same precedent as ADR 0027 line 290 and F3's own tripwire-integration note) |
| **Post-review re-verification** (after the `cancel_reason` overload fix) | `go build ./...`; `go build -tags integration ./...`; `go test -count=1 ./internal/modules/documents/approval/...`; `go test -count=1 ./...`; `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` | all clean/PASS, `0 violation(s)`, zero `FAIL` lines in full suite |

Grep-zero proof:
```
$ grep -rn "SkipStage\|ErrCannotSkipLastStage" --include=*.go .
(no matches)
```

## Acceptance vs spec Validation Gate

| Gate item | Met? | Evidence |
|-----------|------|----------|
| `InstanceStatus` gains `changes_requested` (5th value, non-terminal) | yes | `domain/instance.go` const + `TestInstanceChangesRequestedIsNonTerminal` |
| `SkipStage`/`ErrCannotSkipLastStage` deleted; zero call sites remain | yes | grep-zero proof above; `TestSkipStageDuringSigning` deleted alongside |
| Migration 0288: `approval_instances_status_check` widened; `cancel_reason` writable | yes (schema) / **partial** (live-run) | migration file written per exact spec (`0288_review_verdicts_changes_requested.sql`); SQL reviewed against `db/baseline/0001_current_schema.sql` naming; **not applied against a live/testdb database in this session** (see defer) |
| New `review_verdict_service.go`: `ready` reuses `EvaluateQuorum`; `request_changes` transitions instance→`changes_requested`, document→`draft` | yes (code + unit) / **not live-verified** | `review_verdict_service.go` written exactly as described; domain unit tests green; integration test written and compiles/vets clean but **not executed against Postgres** (see defer) |
| Tier-1 row + `CapApprovalReview` tier-2 check | yes | `permissions.go` new row; `TestRouteCoverage` green (auto-discovery); both api-lint lints `0 violation(s)` |
| `Cancel(reason)` persists to `cancel_reason` column | yes (code) / **not live-verified** | `cancel_service.go` fix + `cancelFakeRepo` unit assertion green; `TestCancelInstance_ReasonPersists` integration test written but not executed (see defer) |
| No regression | yes | full `go test ./...` green |
| ADR: freeze boundary + review layer + choke-point concurrency | not this feature | per spec.md — explicitly F5's due ADR, not F4's |

## Review disposition

- **`zeroContentHash` adapter hack**: `verdictsAsApprovals` constructs synthetic `domain.Signoff`
  values from `ReviewVerdict` rows purely so `domain.EvaluateQuorum` (which expects `[]Signoff`) can be
  reused unchanged. `ContentHash` is set to a 64-char placeholder (`"0"` × 64, verified length) since
  verdicts have no content-hash concept of their own (spec.md non-goals: no e-signature semantics) and
  only `ActorUserID` is actually read by the quorum-counting logic. This is function-reuse via an
  adapter, not storage-shape reuse — matches spec.md's explicit non-goal ("not a widened reuse of the
  `approval_signoffs` row shape").
- **No tripwire arm added for `approval_review_verdicts`**: confirmed via clean `go test
  ./scripts/api-lint/...` and clean `go run ./scripts/api-lint -strict ...` that TRIPWIRE-ARM-DRIFT does
  not fire — a table absent from the `TripwireArms` registry is unchecked by DRIFT, not flagged. This
  is a real coverage gap in the tripwire arm system (a mutating table can silently ship with no
  generated tripwire trigger), not something F4 introduced — it is a structural property of the
  arm-generation system predating this feature. Bounded-deferred below rather than silently left
  unmentioned.
- **SoD check implemented inline, not via `domain.CheckSoD`**: spec.md's consumer-contract description
  names `domain.CheckSoD` as the reused SoD primitive. The actual implementation checks
  `instance.SubmittedBy == req.ActorUserID` inline and returns `domain.ErrAuthorCannotSign` directly.
  This is functionally equivalent for the author-self-verdict case (the only SoD rule spec.md's
  Validation Gate exercises); `CheckSoD`'s second parameter (`priorSignoffs`, for
  actor-already-signed-in-an-earlier-stage detection) is not applicable here since `InsertVerdict`'s
  idempotent-replay/conflict distinction already handles the actor-already-recorded-a-verdict case at
  the same stage. Judgment call: reusing the exact `CheckSoD` function signature would have required
  either fabricating a `[]Signoff` slice from unrelated verdict rows (same shape problem as the quorum
  adapter, but for a check that doesn't need it) or a signature change to `CheckSoD` itself (out of
  scope — `domain/sod.go` untouched). Not a functional gap.
- **`EventTypeReviewVerdictRecorded`/`EventTypeReviewChangesRequested` naming**: chosen to read as
  `<domain>.<event>` consistent with the existing `EventType` string-const style
  (`review_verdict.recorded`, `review_verdict.changes_requested`), distinct from
  `EventTypeSignoffRecorded`/`EventTypeSignoffRejected` so the two verdict mechanisms remain
  audit-distinguishable even though they parallel each other structurally.
- **No lifecycle-event enqueue for `request_changes`**: confirmed no
  `EventTypeDocumentChangesRequested`-equivalent const exists in
  `internal/modules/documents/domain/notification_events.go` (only
  `Published/Superseded/Obsoleted/Approved/Rejected`) — matches spec.md's non-terminal framing of
  `changes_requested`. `WithLifecycleEnqueuer` is still wired on `ReviewVerdictService` for interface
  parity with the other services, in case a future terminal path needs it.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|--------------------------|
| Live DB run of migration 0288 + the new integration test (`tests/integration/approval/review_verdict_integration_test.go`) not executed in this session | No `DATABASE_URL`/`METALDOCS_DATABASE_URL` obtainable without reading `.env` (forbidden by standing rule); a local `metaldocs-postgres` Docker container is running but its password is only available via `.env` or container-env inspection, both blocked by the same rule. Code compiles and `go vet -tags integration` passes clean; the migration SQL was hand-verified against `db/baseline/0001_current_schema.sql` naming and mirrors `0285_approval_signoffs_rls.sql`'s exact RLS idiom | Run `.\scripts\start-api.ps1` (or the project's normal PowerShell-based local-dev bootstrap, which sources the DSN without printing it) then `go test -tags integration ./tests/integration/approval/... -run 'ReviewVerdict|CancelInstance_ReasonPersists' -v`; owner: whoever next has an authorized local Postgres session |
| `approval_review_verdicts` has no tripwire arm registered in `internal/platform/tripwire/arms.go` | TRIPWIRE-ARM-DRIFT only checks tables already present in the `TripwireArms` registry — a new table absent from it is unchecked, not flagged, so no lint forced this. Adding an arm is a structural change to the arm-generation system (new golden migration/render.go branch) that plan.md's F4 task list did not scope | Add before `approval_review_verdicts` carries data outside test/dev — a dedicated small task: register the table in `arms.go`, regenerate, confirm TRIPWIRE-ARM-PARITY picks it up |
| Same-instance resubmit reentry after `request_changes` (spec.md Interview #4 / Non-goals) | Carried forward unchanged from spec.md — `submit_service.go` has no reuse-detection logic; out of F4's file scope | Per spec.md: before M2c (FE) exposes resubmit UX, or during F5 (freeze-boundary work) |
| SoD check implemented inline rather than via `domain.CheckSoD` | See Review disposition — functionally equivalent for the exercised rule, but diverges from spec.md's literal consumer-contract wording | None required unless a future actor-already-signed-in-a-prior-stage SoD rule needs enforcing for verdicts specifically; revisit alongside F7 (SoD unification) |
