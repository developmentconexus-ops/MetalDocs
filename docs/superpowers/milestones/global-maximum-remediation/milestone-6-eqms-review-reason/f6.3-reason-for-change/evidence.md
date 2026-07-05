# Feature F6.3 — Evidence

> **Milestone:** 6  ·  **Feature:** `f6.3-reason-for-change`  ·  **Closed:** 2026-07-05
> **Contract:** `spec.md` (consumer contract + Validation Gate) + `../validation-contract.md` §5 (binding, HS-7).

## What was implemented

Structured reason-for-change captured at revision submit (draft→under_review), a distinct field
(never overloading `revision_title`), carried into the audit trail in the business tx. Producer
matches the `spec.md` consumer contract (submit-revision request field + `approval_submitted` audit
payload).

- **Capture (T8)** `approval/application/submit_service.go` — `SubmitRequest` +`ReasonForChange`
  +`ReasonCategory`; persisted by **extending the existing** draft→under_review UPDATE
  (`SET ... reason_for_change=$5, reason_category=$6`), not a second write; added to the
  `approval_submitted` governance-event `payloadMap` (audit in-tx). `normalizeReasonForChange` —
  required for REV≥1 (`ErrReasonForChangeRequired`, 422), optional at REV 0; `validReasonCategories`
  mirrors `ck_documents_reason_category`. 422 mappings `validation.reason_for_change_required` /
  `validation.reason_category_invalid`. Contract-first openapi + regenerated `SubmitRequest`. —
  `4e9ffbde`, `8ea0dc6c`.
- **Root-fix (T8b)** derive `revision_number` in-tx via new
  `LoadGovernedRevisionNumber(ctx, tx, tenantID, documentID)` — the REV≥1 gate (for both
  reason-for-change **and** pre-existing revision_title) never fired on the HTTP path because
  `submit_handler.go` built `SubmitRequest` with `RevisionNumber` always 0. Now derived after authz
  from the document row; `RevisionNumber` removed from `SubmitRequest`; content-hash input uses the
  derived value in-tx. — `06b1929f`.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — failing test first, then green | subagent TDD (reason capture + derived-revision root-fix) | red→green; suite green below | real |
| Static build | `go build ./...` | `BUILD_DONE=0` | — |
| Contract pin — `SubmitRequest` carries `reason_for_change` (+category) | `go test -run TestSubmitRequestReasonField ./internal/modules/documents/approval/http/contracts/...` | `ok ... 1.299s` | real |
| Persist on row (mock/unit) | `go test -run 'TestSubmitPersistsReason$\|TestSubmitPersistsReason_OptionalCategoryOmitted' ./internal/modules/documents/approval/application/...` | `ok ... 2.623s` | fixture (mock tx) |
| Audit payload carries reason (unit) | `go test -run 'TestSubmitReasonOnAuditTrail$' ./.../approval/application/...` | `ok` (same run) | fixture (mock emitter) |
| REV≥1 required → 422; REV 0 optional | `go test -run 'TestSubmitReasonRequiredRev1$\|TestSubmitReasonOptionalAtRev0$\|TestSubmitReasonRequiredRev1_DerivedFromDocumentRow_NoClientRevisionNumber' ./.../approval/application/...` | `ok` (derived-from-row path proven) | real (logic) |
| `reason_category` outside enum rejected (unit) | `go test -run TestSubmitReasonCategoryInvalidRejected ./.../approval/application/...` | `ok` | fixture |
| Persist on real DB | `TestSubmitPersistsReason_RealDB` (integration) | authored; **validator-run** (row is not HTTP-readable back — persisted in the same UPDATE proven live to transition the doc) | real (testdb) |
| Audit event on real DB | `TestSubmitReasonOnAuditTrail_RealDB` (integration) | authored; **validator-run** — the reason lands in `governance_events`, which has **no HTTP read surface**, so it is not live-observable | real (testdb) |
| `reason_category` DB CHECK | `TestDocumentReviewCheckConstraints` (integration, case 3) | authored; **validator-run** | real (testdb) |
| No path writes reason into `revision_title` | grep census + review | census clean (`reason_for_change` never assigned to a `revision_title` target) | real |
| **Live drive: submit-for-review carrying structured reason** | `.\scripts\start-api.ps1 -Build` → login → submit (see capture) | **GREEN (HTTP path)** — `SUBMIT_STATUS=201` with `reason_for_change`+`reason_category` in the body; `DOC_AFTER {status:under_review, rev_ver:1}` (draft→under_review, rev 0→1). Proves the field is accepted and drives the real business tx end-to-end. Reason **capture on the audit trail is not live-observable** (governance_events has no read endpoint) — that leg is validator-run integration. | real (live) |

> **Honest labeling.** Unit rows use mock tx/emitter (fixture). The live HTTP drive is the
> real-provider proof that the **structured reason field is accepted and drives the submit business
> tx end-to-end** (201, draft→under_review, rev bump). The two things the live drive **cannot**
> show — the reason persisted on the row and the reason on the `governance_events` audit payload —
> have no HTTP read surface; they are testdb-factory integration tests, authored during TDD and
> re-run by the `milestone-validator` from clean state (needs `DATABASE_URL`; `.env` is forbidden).

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| Generated `SubmitRequest` carries reason field(s); oasdiff green | **yes** | contract pin row |
| REV≥1 submit with reason persists on `public.documents` | **yes** (unit) / **validator-run** (real DB); live drive proved the same UPDATE transitions the doc | persist rows + live drive row |
| Submit emits **one** audit event carrying the reason | **yes** (unit) / **validator-run** (real DB — governance_events, no HTTP read surface) | audit rows |
| REV≥1 submit without reason → 422 problem+json | **yes** | required-Rev1 row (derived-from-row proven) |
| No code path writes reason into `revision_title` | **yes** | grep census row |
| `reason_category` outside enum rejected by DB CHECK | **validator-run** | CHECK row |
| Live drive: submit carrying structured reason drives the business tx | **yes (HTTP path)** — 201, draft→under_review, rev 0→1 | live drive row (GREEN) |

## Review disposition

- Spec-compliance review: subagent + main-session. The HTTP-path RevisionNumber-always-0 gap was
  caught in review (would have made the REV≥1 gate dead on the real handler) and root-fixed as T8b
  (derive in-tx) rather than symptom-patched — surfaced the governed content-hash consequence
  (HS-2-class), safe at v1 fresh re-baseline (no in-flight instances; no golden hash pins old value).
- Code-quality review: chosen `LoadGovernedRevisionNumber` over extending the 8-call-site
  `LoadDocumentAreaCode` (smaller blast radius); reviewed by root-cause family.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| Backfill `reason_for_change` on legacy rows | Intent unreconstructable; spec non-goal; legacy stays NULL | None — permanent by design |
