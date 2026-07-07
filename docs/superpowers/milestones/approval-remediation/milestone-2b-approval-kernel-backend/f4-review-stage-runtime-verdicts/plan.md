# Feature F4 — plan.md

Base: `docs/superpowers/plans/2026-07-07-m2b-approval-kernel-backend.md` F4 section, corrected per
spec.md's Interview record:

- Migration file is `0288_review_verdicts_changes_requested.sql` (0286/0287 taken; F3 used
  reference-data only, no migration consumed 0288 until now).
- Route path: `POST /api/v1/approval/instances/{instance_id}/stages/{stage_id}/review-verdict`
  (real convention, not plan.md's `/approval-instances/{instanceId}/...` shorthand).
- No new edge needed in any `legalTransitions` map — `documents/domain/state.go`'s
  `CanTransitionDocumentStatus` (the real, wired transition fn) already legalizes
  `under_review→draft`. Do not touch `internal/modules/documents/approval/domain/state.go` (dead,
  unwired file) — out of scope.
- Do NOT touch `submit_service.go` — same-instance resubmit reentry is a bounded defer (spec.md).
- Do NOT widen `ux_approval_instances_active_document_id` or `approval_signoffs_decision_check` —
  verdicts get their own table, not a reuse of `approval_signoffs` rows (only `EvaluateQuorum`, the
  counting function, is reused).

## Tasks

1. **Contract.** `api/openapi/v1/openapi.yaml`: add `review-verdict` POST route + request body
   `{verdict: ready|request_changes, comment?: string}` (comment required when
   `request_changes` — enforce in contract `Validate()`, mirror `CancelRequest` pattern). Cancel body
   already has required `reason` (no change needed — confirmed already wired). Regenerate via repo's
   oapi-codegen make target.

2. **Domain (TDD, failing test first):**
   - `domain/instance.go`: add `InstanceChangesRequested InstanceStatus = "changes_requested"` (5th
     value, non-terminal). Delete `SkipStage` method + `ErrCannotSkipLastStage` sentinel entirely
     (`domain/errors.go`). `Cancel(reason string)`: stop discarding — return/store reason for the
     caller to persist (mirror however `RejectHere` already surfaces its reason field).
   - New `domain/review_verdict.go` (mirrors `domain/signoff.go`'s shape): `Verdict` value object,
     `VerdictReady`/`VerdictRequestChanges` consts, `NewVerdict(...)` validating `request_changes`
     requires non-empty comment, wrong-stage-kind rejection (stage must be `StageKindReview`).
   - New sentinel errors as needed (`ErrVerdictWrongStageKind`, `ErrVerdictCommentRequired`) in
     `domain/errors.go`.
   - Domain unit tests: `instance_test.go` additions (status value, SkipStage gone), new
     `review_verdict_test.go`.

3. **Migration `0288_review_verdicts_changes_requested.sql`** (verify exact table/column names
   against `db/baseline/0001_current_schema.sql` before writing SQL, per standing constraint):
   - New table `approval_review_verdicts` (mirror `approval_signoffs`' column shape: id, tenant_id,
     instance_id, stage_id, actor_id, verdict, comment, created_at — adapt exact columns to the real
     `approval_signoffs` schema for consistency).
   - `ALTER TABLE approval_instances ADD CONSTRAINT approval_instances_status_check` widened (drop +
     recreate, or `ALTER ... DROP CONSTRAINT ... ADD CONSTRAINT ...`) to include
     `'changes_requested'`.
   - Confirm `cancel_reason` column already exists (F1/0286) — no schema change needed for it, only
     the write-path wiring in step 4.
   - RLS: `ENABLE`/`FORCE ROW LEVEL SECURITY` + tenant_isolation policy on the new table, copying the
     exact policy shape from `db/migrations/0285_approval_signoffs_rls.sql`.

4. **Application service** `application/review_verdict_service.go` (mirror `decision_service.go`'s
   `RecordSignoff` structure exactly): off-tx actor-name lookup → in-tx `LoadInstance` FOR UPDATE →
   OCC via `ExpectedRevisionVersion` → area-code resolution → `authz.Require(CapApprovalReview,
   areaCode)` → eligibility (`domain.CheckEligibility`) → SoD (`domain.CheckSoD`, author cannot
   verdict own doc) → stage-kind check (must be `StageKindReview`) → verdict insert (idempotent
   replay-aware, mirror signoff's pattern) →
   - `ready`: `EvaluateQuorum` with this verdict counted in the "approvals" bucket exactly like
     signoff-approve; on quorum reached, advance stage/instance exactly as `RecordSignoff` does.
   - `request_changes`: instance → `InstanceChangesRequested`; document →
     `docsdomain.CanTransitionDocumentStatus(DocStatusUnderReview, DocStatusDraft)` (already legal) +
     DB-level OCC `WHERE` guard (mirror reject path in `decision_service.go`).
   - Governance-event emission + lifecycle-event enqueue (mirror decision_service.go's calls exactly).

5. **Cancel-reason persistence.** `cancel_service.go`: after `instance.Cancel(reason)`, the
   repository's cancel-write method must now also `SET cancel_reason = $N` (currently only status +
   governance event). Add the column write to whichever repository method performs the cancel UPDATE
   (`infrastructure/postgres_approval_repository.go`).

6. **HTTP layer.** `http/review_verdict_handler.go` + `http/contracts/review_verdict.go` (mirror
   `signoff_handler.go`/`contracts/signoff.go` shapes). Register route in
   `internal/modules/documents/approval/http/router.go`. Add tier-1 row in
   `apps/api/cmd/metaldocs-api/permissions.go` for the new route → `CapApprovalReview` (closes F3's
   documented defer). Update `permissions_test.go` (`TestPermissionResolver`/`TestRouteCoverage`).

7. **Integration tests (testdb factory, REQUIRED framework):**
   `tests/integration/approval/review_verdict_integration_test.go` — ready verdict advances quorum;
   request_changes transitions instance+document; comment required for request_changes; wrong stage
   kind rejected; SoD blocks self-verdict; concurrent verdicts on same stage serialize (OCC); cancel
   persists reason to DB column (round-trip read).

8. **Grep-zero check:** `SkipStage`/`ErrCannotSkipLastStage` — zero matches outside git history.

9. **Full verification sweep:** `go build ./...`; `go build -tags integration ./...`; targeted +
   broad `go test`; both authz lints (`api-lint -strict`); `go test ./scripts/api-lint/...`.

10. **evidence.md**, then commit `feat(approval): F4 review verdicts + changes_requested + SkipStage
    deleted + cancel reason (W11/W13, spec §2)`. NOT pushed.

## ADR note

Spec.md flags: the "freeze boundary + review layer + choke-point concurrency" ADR (2 of 4 due for the
milestone) is F5's, not F4's — F4 only introduces the review-verdict *mechanics*; the ADR is authored
once F5's freeze-boundary decision is also settled, per milestone's ADR ledger. Not a gap in F4.
