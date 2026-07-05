# Feature F6.2 — periodic review/expiry + capability-gated review workflow

> **Milestone:** 6 — eQMS periodic review/expiry + structured reason-for-change  ·  **Folder:** `f6.2-periodic-review`
> **Status:** Implementing

## Source

- Milestone spec row (F6.2): review/expiry model + scheduled surfacing (River) + capability-gated
  review workflow, contract-first. Accept: contract + generated code + FE consumer; scheduled-job
  proof; live drive of a due-review cycle.
- Governing-spec reference: finding 14 (dimension 6 product gap) — ISO 9001 §7.5.3; `../validation-contract.md` §2–§4.

## Plan

Executed via `superpowers:subagent-driven-development` (fresh subagent per task; sonnet
implement/review, haiku mechanical; main reviews + commits). TDD: failing test first. **Ordered
tasks** (T0/T1 are the shared foundation for F6.2 **and** F6.3):

- **T0 — contract-first openapi (shared).** Edit `api/openapi/v1/openapi.yaml`: (a) [F6.3] `reason_for_change` + optional `reason_category` enum on the submit-revision request; (b) [F6.2] optional `effective_to` + `review_due_at` on the schedule/publish request; (c) [F6.2] `review_due_at`/`effective_from`/`effective_to`/`last_reviewed_at` on the document response DTO(s); (d) [F6.2] mark-reviewed op `POST /documents/{documentId}/review` (body: next `review_due_at`, optional `effective_to`, `revision_version` OCC); (e) [F6.2] review-due list filter param. **Propose the diff for main review BEFORE regen (HS-7).** Then `go generate` BE + FE types. Pin tests for the new shapes.
- **T1 — migration (shared).** `db/migrations/0NNN_document_review_and_reason.sql` on `public.documents`: `+ review_due_at timestamptz NULL`, `+ last_reviewed_at timestamptz NULL`, `+ reason_for_change text NULL`, `+ reason_category text NULL`; CHECKs: `effective_to > effective_from`, review-due sanity, `reason_category` enum. Provision on the testdb template. DB-CHECK integration proof (TDD: seed bad rows → rejected).
- **T2 — capability `document.review` (10 touchpoints).** const + `validCapabilities`; `ScopeTenant` classify; **registry `TestCapabilityRegistrySize` 34→35**; seed grants (`db/reference-data/...`); tier-1 route map (`permissions.go`); guard tests green; **regen M2 tripwire arms + drift check green** (arm includes `document.review`); negative tripwire proof.
- **T3 — documents published read-port.** `ReviewDueReader.ListDueForReview(ctx, tx, now, limit)` on the documents module (`domain/port.go` + repository query), tenant-scoped. Integration proof.
- **T4 — River surfacer + write-port.** documents write-port to flag due docs (idempotent — set a surfaced/overdue marker via `ON CONFLICT`/idempotent UPDATE); River periodic job `document-review-surfacer` (queue `maintenance`, 1h, `RunOnStart:false`) in a new `internal/modules/jobs/document_review_surfacer` package + register in `metaldocs-jobs` main + the periodic-jobs list; per-run `SeedTxIdentity` (M3). Idempotency + tenant-isolation integration proof.
- **T5 — mark-reviewed workflow.** documents application service `MarkReviewed` (tier-2 `authz.Require(CapDocumentReview,"tenant")` after `SeedTxIdentity`; route through M4 `CanTransitionDocumentStatus`/unified fn; OCC `revision_version` CAS; set `last_reviewed_at` + next `review_due_at`); http handler + route wiring; wire `effective_to`/initial `review_due_at` set on the schedule/publish path (`publish_service.go`). Authz + mark-reviewed integration proofs.
- **T6 — response mapping + list filter.** Map review fields into the document response DTO; add the review-due list filter to the documents read service. Pin test + FE type consumer.
- **T7 — ADR + wiki.** ADR `wiki/decisions/00NN-document-periodic-review-and-reason-for-change.md` (Accepted); refresh `wiki/modules/documents.md`, `wiki/modules/jobs.md`, `wiki/modules/controlled-documents.md` (cross-link), `wiki/database/tables/documents.md`. (wiki-curator after code.)

Files touched (indicative): `api/openapi/v1/openapi.yaml` + generated; `db/migrations/`; `internal/modules/iam/domain/{model.go,capability_scope.go,model_test.go}`; `db/reference-data/`; `apps/api/cmd/metaldocs-api/permissions.go`; M2 generated tripwire arms; `internal/modules/documents/{domain/port.go,repository/,application/}`; `internal/modules/documents/approval/application/publish_service.go`; `internal/modules/jobs/document_review_surfacer/`; `apps/jobs/cmd/metaldocs-jobs/main.go`; `internal/modules/jobs/maintenance/periodic.go`; `wiki/`.

Test strategy: testdb factory for every integration proof; targeted `-run` only (no 20-min suite);
negative authz (tripwire fires); idempotency (twice→once); tenant isolation (cross-tenant 0 rows).

## Execution notes

<filled during subagent-driven-development>
