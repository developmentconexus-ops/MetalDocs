# Feature F6.3 — structured reason-for-change

> **Milestone:** 6 — eQMS periodic review/expiry + structured reason-for-change  ·  **Folder:** `f6.3-reason-for-change`
> **Status:** Implementing

## Source

- Milestone spec row (F6.3): structured reason-for-change at revision creation (contract field(s), not
  free text). Accept: contract + pin tests; revision-creation drive shows structured capture; audit
  trail carries it.
- Governing-spec reference: finding 14 (dimension 6 product gap) — 21 CFR Part 11; `../validation-contract.md` §5.

## Plan

Shared foundation T0 (openapi: reason_for_change/reason_category on submit-revision request) and T1
(migration: reason_for_change/reason_category columns + enum CHECK) are authored **jointly with F6.2**
(see `../f6.2-periodic-review/plan.md`) since both features add columns to `public.documents` in one
forward migration and edit one `openapi.yaml`. F6.3-specific:

- **T8 — capture + audit.** Thread `reason_for_change` (+ optional `reason_category`) from the submit
  handler into `SubmitRequest` (`submit_service.go:33-43`); persist on `public.documents` by extending
  the existing draft→under_review UPDATE (`:185-195`, add the columns to the `SET` — NOT a second
  write, NOT `revision_title`); add `reason_for_change`/`reason_category` to the `approval_submitted`
  governance-event `payloadMap` (`:208`) so the audit trail carries it **in the business tx**;
  enforce **required for REV≥1** at the handler/service (friendly 422 problem+json), optional at REV 0.

Files touched: `internal/modules/documents/approval/application/submit_service.go`; the submit http
handler; generated submit request DTO (via T0); `db/migrations/` (via T1).

Test strategy (TDD, testdb, targeted `-run`): pin test that generated `SubmitRequest` carries the
field; persist-on-row integration proof; **one** audit event whose payload carries the reason;
REV≥1-without-reason → 422; `reason_category` enum DB CHECK; grep census that `revision_title` is not
overloaded.

## Execution notes

<filled during subagent-driven-development>
