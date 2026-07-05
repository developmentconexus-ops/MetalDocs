# Feature F6.2 — Spec (periodic review/expiry + capability-gated review workflow)

> **Milestone:** 6 — eQMS periodic review/expiry + structured reason-for-change  ·  **Folder:** `f6.2-periodic-review`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-07-04 / Leandro (operator-approved gate 93cd6114 + committed D4 validation-contract 00c4ec8f are the standing approval; design decisions locked below)

> This is the feature's **contract**, written and approved **before any code**. The concrete expected
> behaviors are pinned in `../validation-contract.md` §2–§4 (binding, section-by-section, HS-7). This
> file records the consumer contract, the interview-resolved design decisions, non-goals, and the
> Validation Gate.

## Interview record (fail-closed gate)

The consumer contract was discovered by the committed `developing-new-work` gate
(`../../../analysis/2026-07-04-m6-eqms-review-reason-system-impact.md`, 93cd6114) + the D4
`../validation-contract.md`. The residual genuinely-underspecified design points are resolved below
(sensible global-maximum defaults, recorded — not guessed silently).

| # | Question | Answer |
|---|----------|--------|
| 1 | Is the review-cycle a hardcoded interval (e.g. "review every 12 months") or client-supplied data? | **Client-supplied data.** ISO 9001 §7.5.3 does not mandate a period; hardcoding one is policy-in-code. Global-max: `review_due_at` / `effective_from` / `effective_to` are supplied via the contract (at schedule/publish and at mark-reviewed), so review-cycle policy lives in tenant data, not Go. Mark-reviewed carries the **next** `review_due_at`. |
| 2 | HTTP shape of the mark-reviewed op? | `POST /documents/{documentId}/review` — a discrete governance action (like publish), not a generic PATCH; body carries the next `review_due_at` (+ optional `effective_to`) and the OCC `revision_version`. |
| 3 | Where are `effective_from`/`effective_to`/`review_due_at` first set? | `effective_from` stays set on the existing schedule/publish path (`publish_service.go:282`). `effective_to` (expiry) + initial `review_due_at` are set on the same schedule/publish request — new **optional** request fields on the existing schedule/publish op (contract-first). No new column family. |
| 4 | Which statuses may be mark-reviewed? | Only a live/effective published revision — routed through the **M4 unified transition function**; a fresh-status guard consistent with the DB. Mark-reviewed updates review dates, it is **not** a `status` transition (status stays `published`); it sets `last_reviewed_at` + next `review_due_at` under the OCC CAS. (No new `status` value.) |
| 5 | `reason_category` enum for the surfacer/review? | N/A to F6.2 — `reason_category` belongs to F6.3. |
| 6 | Surfacer queue + interval? | River periodic job, queue `maintenance` (reuse the M5 janitor queue), `PeriodicInterval(1*time.Hour)`, `ID:"document-review-surfacer"`, `RunOnStart:false`. |
| 7 | What does "surface" mean — a new column, a notification, or a read projection? | **Idempotent flag on the row**: the surfacer marks due docs (e.g. sets a `review_overdue` boolean / a surfaced-at stamp) via the documents write-port, idempotent by construction; the FE review-due **list filter** reads them. No notification side effect in M6 (escalation deferred, contract §8). Running twice → surfaced once. |

## Consumer contract (FIRST — before any producer)

- **Consumers:**
  - The **FE documents view** — consumes new response DTO fields (`review_due_at`, `effective_from`,
    `effective_to`, `last_reviewed_at`) + a **review-due list filter** to show/act on due documents.
  - The **River surfacer job** (`metaldocs-jobs`) — consumes a new **documents published read-port**
    `ReviewDueReader.ListDueForReview(ctx, tx, now, limit)` (never raw SQL on `public.documents`).
  - The **mark-reviewed HTTP handler** — consumes a new documents application service that sets
    `last_reviewed_at` + next `review_due_at` under `authz.Require(CapDocumentReview)` + M4 transition.
- **Contract:** exactly `../validation-contract.md` §2 (columns + CHECKs), §3 (capability 10-touchpoint
  table + registry 34→35 + M2-generated arm + negative proof), §4 (read-port signature + River
  surfacer idempotency + tenant isolation + mark-reviewed via M4). Wire shapes are contract-first in
  `api/openapi/v1/openapi.yaml`, then regenerated.
- **Source of truth for the contract:** `../validation-contract.md` (D4, committed 00c4ec8f) +
  `api/openapi/v1/openapi.yaml` (once edited) + the M2 generated tripwire arms + the M4
  `CanTransitionDocumentStatus` fn.

## What this feature implements

Wire the review/expiry model on `public.documents` (reuse `effective_from`/`effective_to`; add
`review_due_at`/`last_reviewed_at` + DB CHECKs); add the `document.review` capability (all 10
touchpoints, registry 34→35, ADR, M2-generated tripwire arm); a River periodic **surfacer** reading
due docs through a new documents published read-port (idempotent, tenant-seeded); a contract-first
**mark-reviewed** workflow (tier-1+tier-2 authz, M4-routed) setting `last_reviewed_at` + next
`review_due_at`; and the contract-first response fields + review-due list filter with a named FE
consumer.

## Non-goals (mandatory)

- **Structured reason-for-change** — that is F6.3 (shares the migration only).
- **Training acknowledgment / obligated-reader attestation** — `distribution` owns it; out of finding scope.
- **Notification/escalation on overdue** — M6 surfaces + gates; escalation is a bounded defer (contract §8).
- **A new `status` value** — mark-reviewed updates review dates, not the lifecycle status; no 11th state.
- **A hardcoded review-cycle interval** — dates are client-supplied data (interview #1).
- **Duplicate `published_at`/`effective_date`/`expiry` columns** — reuse `effective_from`/`effective_to`.
- **Reworking `controlleddocuments` lifecycle** — untouched.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| Migration applies; DB CHECK rejects `effective_to ≤ effective_from` and bad `review_due_at` | `go test -run TestDocumentReviewCheckConstraints ./internal/modules/documents/... -tags integration` | real (testdb) |
| `document.review` reachable only with the capability; withheld → denied; UPDATE without asserting the cap trips the tripwire (P0001) | `go test -run TestMarkReviewedAuthz ./internal/modules/documents/... -tags integration` | real (testdb) |
| Registry size 34→35; M2 tripwire-arm drift check green; arm includes `document.review` | `go test -run TestCapabilityRegistrySize ./internal/modules/iam/...` + M2 drift lint | real |
| Surfacer flags a `review_due_at ≤ now()` doc on a tick; idempotent (twice→once); tenant-isolated (tenant-B doc not surfaced under tenant-A) | `go test -run TestReviewSurfacer ./internal/modules/jobs/... -tags integration` | real (testdb) |
| Read-port `ListDueForReview` used by the surfacer; no `documents` table SQL in the jobs module | `grep` census + `go test -run TestReviewDueReader ./internal/modules/documents/... -tags integration` | real |
| mark-reviewed sets `last_reviewed_at` + next `review_due_at`, routed through M4 transition fn, OCC CAS | `go test -run TestMarkReviewed ./internal/modules/documents/... -tags integration` | real (testdb) |
| Contract: new response fields + review-due filter + mark-reviewed op in openapi; `oasdiff` green; FE type present | `oasdiff` M1 gate + pin test `TestDocumentResponseReviewFields` | real |
| Live drive: set review-due → surfacer flags → capability-gated mark-reviewed | `.\scripts\start-api.ps1 -Build` + captured network/DB proof in `evidence.md` | real (live) |

> TDD: failing test first, then implement to green. testdb factory for every integration proof;
> targeted `-run` only (no full 20-min suite).

## ADR needed?

- [x] Durable decision made → **one ADR** (review-cycle model + effective-vs-publish semantics +
  `document.review` capability + surfacing-via-River), required by the `model_test.go:96` capability-bump
  convention. Author under `wiki/decisions/` with F6.2, link here on creation: `wiki/decisions/00NN-document-periodic-review-and-reason-for-change.md`.
