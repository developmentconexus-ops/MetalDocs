# Feature F6.3 — Spec (structured reason-for-change)

> **Milestone:** 6 — eQMS periodic review/expiry + structured reason-for-change  ·  **Folder:** `f6.3-reason-for-change`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-07-04 / Leandro (gate 93cd6114 + D4 validation-contract 00c4ec8f are the standing approval; design decisions locked below)

> Contract pinned in `../validation-contract.md` §5 (binding, HS-7). This file records the consumer
> contract, interview-resolved design decisions, non-goals, and the Validation Gate.

## Interview record (fail-closed gate)

Contract discovered by the committed gate (93cd6114) + D4 (`../validation-contract.md` §5). Residual
design points resolved below.

| # | Question | Answer |
|---|----------|--------|
| 1 | Structured field vs reuse of free-text `revision_title`? | **New distinct field** `reason_for_change`. `revision_title` (title of the revision) stays as-is; reason-for-change is the *why*, a separate structured field (runtime-verified: `revision_title` is written at `submit_service.go:188,194` — F6.3 must NOT overload it). |
| 2 | Is `reason_category` in scope, and its enum? | **Optional** `reason_category` enum, lean fixed set: `content`, `corrective`, `regulatory`, `periodic_review`, `administrative`. DB CHECK enforces the set; nullable. |
| 3 | Where is it captured + carried to the audit trail? | Threaded into `SubmitRequest` (`submit_service.go:33-43`); persisted on `public.documents` by extending the existing draft→under_review UPDATE (`:185`, NOT a second write, NOT `revision_title`); carried into the audit trail **in the business tx** via the existing `approval_submitted` governance-event payload (`emitter.Emit`, `:208-229` — add `reason_for_change`/`reason_category` to `payloadMap`). |
| 4 | Required or optional at the API? | **Required for REV≥1** (friendly 422 problem+json if missing); REV 0 (initial creation) follows the existing `revision_title` default convention — reason optional at REV 0. Nullable in DB for legacy rows (expand/contract). |

## Consumer contract (FIRST — before any producer)

- **Consumers:**
  - The **FE revision-submit form** — consumes a new `reason_for_change` (+ optional `reason_category`)
    field on the submit-revision request schema.
  - The **audit reader** — observes one `approval_submitted` audit event whose payload carries the
    structured reason (+ category) for every REV≥1 submit (21 CFR Part 11 attributable change reason).
- **Contract:** exactly `../validation-contract.md` §5 — the request field(s), the persist-on-row +
  in-tx audit-payload capture, REV≥1-required / legacy-nullable, no reuse of `revision_title`.
- **Source of truth for the contract:** `../validation-contract.md` §5 + `api/openapi/v1/openapi.yaml`
  submit-revision request schema (once edited) + the generated `SubmitRequest`.

## What this feature implements

Add `reason_for_change` (+ optional `reason_category` enum) to the submit-revision request (contract
-first), persist it on `public.documents` at submit (extending the existing UPDATE, not overloading
`revision_title`), and carry it into the audit trail in the business tx via the existing governance
-event emitter. Required at the API for REV≥1; nullable in DB for legacy.

## Non-goals (mandatory)

- **Periodic review/expiry + the `document.review` capability** — that is F6.2 (shares the migration only).
- **Backfilling `reason_for_change` onto legacy rows** — intent can't be reconstructed; legacy stays NULL.
- **Overloading `revision_title`** — the title is not the reason; forbidden by acceptance.
- **A separate reason-capture path on the CD service** — F6.3 lands on the governed documents/approval submit path only.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| Generated `SubmitRequest` carries `reason_for_change` (+ `reason_category`); `oasdiff` green | `oasdiff` M1 gate + pin test `TestSubmitRequestReasonField` | real |
| Submitting a REV≥1 revision with a reason persists it on `public.documents` | `go test -run TestSubmitPersistsReason ./internal/modules/documents/approval/... -tags integration` | real (testdb) |
| The submit emits **one** audit event whose payload carries `reason_for_change` (+ category) | `go test -run TestSubmitReasonOnAuditTrail ./internal/modules/documents/approval/... -tags integration` | real (testdb) |
| REV≥1 submit **without** the field is rejected at the API (422 problem+json) | `go test -run TestSubmitReasonRequiredRev1 ./internal/modules/documents/approval/...` | real |
| No code path writes the structured reason into `revision_title` | `grep` census + code review | real |
| `reason_category` outside the enum rejected by DB CHECK | `go test -run TestReasonCategoryCheck ./internal/modules/documents/... -tags integration` | real (testdb) |
| Live drive: submit a revision with a structured reason → audit event shows it | live drive proof in `evidence.md` | real (live) |

> TDD: failing test first. testdb factory; targeted `-run` only.

## ADR needed?

- [x] Covered by the **single F6.2 ADR** (review model + reason-for-change together) — the gate §9
  scoped one ADR for both eQMS additions. No separate F6.3 ADR. Link on creation: `wiki/decisions/00NN-document-periodic-review-and-reason-for-change.md`.
