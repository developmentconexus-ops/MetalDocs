# Feature F4 — `review-stage-runtime-verdicts` — spec.md

> **Milestone:** 2b — Approval Kernel Backend  ·  **Feature:** `f4-review-stage-runtime-verdicts`
> **Status:** Approved for implementation

## Interview record (self-directed, no operator ambiguity requiring a stop)

| # | Question | Answer |
|---|----------|--------|
| 1 | Does `cancel` already require a non-empty reason? | Yes — `contracts.CancelRequest.Validate()` + `CancelService.ErrReasonRequired` already enforce this end-to-end. F4's only cancel-reason gap is DB persistence of `cancel_reason` (column exists from F1/migration 0286, SELECTed already, never written). |
| 2 | Which document-lifecycle graph governs `request_changes` (under_review→draft)? | `internal/modules/documents/domain/state.go`'s `CanTransitionDocumentStatus` — the real, wired graph used by `decision_service.go`/`cancel_service.go`/`submit_service.go`. It **already** has `StateUnderReview→StateDraft` (reused today by reject + cancel rollback). No new edge needed. |
| 3 | Is `internal/modules/documents/approval/domain/state.go` (a separate `legalTransitions` map, 8-state "Spec 2" graph, `StateRejected` as its own state) relevant to F4? | No — confirmed via grep that zero application/http files import it; only its own `state_test.go` and wiki docs reference it. Dead/unwired code, out of F4 scope (not touched, per "don't refactor adjacent code"). |
| 4 | Design spec §2.3 says "re-submit re-enters the SAME instance" — does F4 implement this? | **No — bounded defer (see below).** `submit_service.go` (444 lines, read in full) unconditionally creates a brand-new instance + stage rows on every call; it has zero logic to detect/reuse an existing `changes_requested` instance for the document. Plan.md's F4 file list never lists `submit_service.go`. Implementing same-instance reentry is a materially larger change (new detect-and-reuse branch in the submit path, snapshot/pool-reuse semantics) than plan.md scoped for this feature. Deferred with a written trigger below rather than silently dropped or silently scope-crept. |

## Consumer contract

A caller with `approval.review` in the stage's area, acting on a `review`-kind stage of an
`in_progress` instance, can `POST` a verdict:

- `ready` — stage-level "approve equivalent" for review-kind stages. Counted via
  `domain.EvaluateQuorum` exactly like a signoff approval. On quorum reached: stage/instance advance
  exactly as `RecordSignoff` does today (reuse, not fork).
- `request_changes` — stage-level "reject equivalent". On any `request_changes` verdict (no quorum
  needed to reject, mirrors today's signoff-reject semantics): instance status → `changes_requested`
  (new, non-terminal, distinct from `in_progress`); document status → `draft` via the already-legal
  `under_review→draft` edge; author must revise and resubmit (resubmit creates a **new** instance —
  bounded defer, not same-instance reentry, per Interview #4).

Endpoint: `POST /api/v1/approval/instances/{instance_id}/stages/{stage_id}/review-verdict` (real path
convention confirmed via `api/openapi/v1/openapi.yaml` grep: `/approval/instances/{instance_id}/...`,
NOT plan.md's shorthand `/approval-instances/{instanceId}/...`).

Mirrors `decision_service.go`'s `RecordSignoff` pattern exactly: off-tx actor-name lookup (H-PRE-1) →
in-tx `LoadInstance` (FOR UPDATE) → OCC via `ExpectedRevisionVersion` → area-code resolution →
`authz.Require(CapApprovalReview, areaCode)` (tier-2; tier-1 row added to `permissions.go`, closing F3's
documented defer) → eligibility (`domain.CheckEligibility`) → SoD (`domain.CheckSoD`) → verdict
construction + idempotent-replay-aware insert → quorum/reject evaluation → stage/instance transition
via repo methods only → document-status transition (friendly check + DB-level OCC) → governance-event
emission → lifecycle-event enqueue.

`Cancel(reason)` gains a caller (`CancelService.CancelInstance`) that persists `reason` to
`approval_instances.cancel_reason` (currently discarded by the domain method, only reaching the
governance event) — closes the one genuine cancel-reason gap.

`Instance.SkipStage`/`ErrCannotSkipLastStage` deleted outright (design spec §2.4, W11 — no product
surface, violates route immutability). Every call site removed, not stubbed.

## Non-goals

- Same-instance resubmit reentry (`submit_service.go` unchanged) — **bounded defer**, see below.
- Freeze boundary / content hash gating (F5).
- Signature-meaning fixed vocabulary / SoD unification (F7).
- SLA / visibility / worklist scope=oversee (F8).
- Delegation (F9).
- `approval_signoffs_decision_check` / `Signoff.Decision`/`NewSignoff` widening — **not needed**:
  verdicts get their own new table/column shape distinct from signoffs (decision below), not a widened
  reuse of the `approval_signoffs` row shape (only the *counting function* `EvaluateQuorum` is reused,
  not the storage row).
- `ux_approval_instances_active_document_id` partial-unique-index widening — deliberately **not**
  touched (widening it to include `changes_requested` without also fixing `submit_service.go`'s reuse
  logic would make resubmit fail outright — a regression, not a fix. See bounded defer.)

## Bounded defers (declared before code, per HS-6 discipline)

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|--------------------------|
| Resubmit after `request_changes` creates a **new** approval instance rather than reactivating the same one (design spec §2.3 partial gap) | `submit_service.go` has zero reuse-detection logic today; adding it is a materially larger change than plan.md scoped for F4 (new branch: detect existing `changes_requested` instance for document, reuse its id/stage snapshot vs. insert fresh) — plan.md's F4 file list never lists this file | Before M2c (FE) exposes resubmit UX, or during F5 (freeze-boundary work, which also touches the submit path) — whichever comes first. Requires: (a) `submit_service.go` gains a lookup-existing-`changes_requested`-instance-for-document branch, (b) decide whether `ux_approval_instances_active_document_id` should then be widened to `status IN ('in_progress','changes_requested')` once reuse logic exists |

## Validation Gate

| Item | Test / proof |
|------|--------------|
| `InstanceStatus` gains `changes_requested` (5th value, non-terminal) | domain unit test |
| `SkipStage`/`ErrCannotSkipLastStage` deleted; zero call sites remain | `grep -rn 'SkipStage\|ErrCannotSkipLastStage'` → zero matches outside git history |
| Migration 0288 (0286/0287 already taken; F3 used reference-data, no migration): `approval_instances_status_check` widened to include `changes_requested`; `approval_instances.cancel_reason` writable | migration applies clean against baseline; testdb-factory integration test round-trips `cancel_reason` |
| New `review_verdict_service.go`: `ready` reuses `EvaluateQuorum` approvals path; `request_changes` transitions instance→`changes_requested`, document→`draft` | unit + testdb integration tests, both verdicts |
| Tier-1 row added for `review-verdict` route, `CapApprovalReview` tier-2 check | `TestPermissionResolver`/`TestRouteCoverage` updated; api-lint drift/parity lints pass |
| `Cancel(reason)` persists to `cancel_reason` column | testdb integration test asserts column value post-cancel |
| No regression | full `go test ./...` + integration suite for approval/iam packages green |
| ADR: freeze boundary + review layer + choke-point concurrency | **NOT this feature** — F5's due ADR per milestone's 4-ADR ledger; F4 introduces the review layer's runtime verdict mechanics only, the ADR is written when F5's freeze-boundary decision is also settled (both ADRs can be one document if scope overlaps at F5 time) |

## Approval

Approved by: self (autonomous milestone execution, no operator ambiguity blocking) — 2026-07-07.
