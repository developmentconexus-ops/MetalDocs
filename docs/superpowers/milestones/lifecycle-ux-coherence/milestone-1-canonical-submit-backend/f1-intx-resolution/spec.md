# F1 — In-tx server resolution of submit prerequisites

> **Milestone:** M1 · **Findings:** 1 · **Status:** spec approved (ADR 0073-ratified)
> **Approved:** 2026-07-06 — from ADR 0073 §2 (canonical /submit owns in-tx prereq resolution) + governing spec §1.3 sequence. No operator interview: contract fully pinned by ratified ADR.

## Consumer contract

**Consumer:** an author whose FE sends `/submit` with **no `route_id` and no `content_hash`**.
`SubmitService.SubmitRevisionForReview` must, **inside the submit transaction**, resolve any omitted
prerequisite from server-authoritative state, then run the existing route/hash/CAS flow unchanged:

1. **route_id** omitted → resolve the single active approval route for the document's controlled-document
   profile: `documents.controlled_document_id` → `CDFieldReader.ProfileCode` (tx-capable) → active
   `approval_routes` by profile (newest version). Empty profile → `docsdomain.ErrProfileNotConfigured`.
   No active route → `docsdomain.ErrApprovalRouteMissing`. (Mirrors `repository.go:1801-1817`.)
2. **content_hash** omitted → bind the head autosaved revision's `content_hash` (mirrors
   `repository.go:1819-1828`, COALESCE/no-rows tolerant → `""`). Fed into the existing
   `ComputeContentHash` wrapper exactly as finalize did (`{"_content_hash": headHash}`) — **no change to
   `content_hash_at_submit` semantics** (HS-2 boundary).
3. Explicit client `route_id`/`content_hash` keep their **exact prior semantics** (additive optionality).

All resolution reads are plain **non-recording SELECTs** on the caller's tx (HS-PRE-1). The narrow
`SubmitDefaultsResolver` port (ISP, ADR 0073 §2) carries the route/hash reads; profile stays in the
service via its wired `CDFieldReader`. `ApprovalRepository` is left untouched.

Capability model unchanged: `document.submit` + `document.edit` @area, in the writable submit tx (ADR 0022).

## Non-goals
- No new "submission tracking" surface; no FE work (M2).
- No idempotency-store change; no ETag change elsewhere (YAGNI §4).
- No change to REV≥1 title/reason gates (already governed in-tx).

## Validation Gate (integration, testdb factory)
- **REV0 empty-body**: fresh draft (`revision_version=0`, no client route/hash) → 201, `ETag "v1"`,
  approval instance created, document `under_review`.
- **REV≥1 empty title/reason**: → 422 typed (`ErrRevisionTitleRequired` / `ErrReasonForChangeRequired`).
- **Explicit route_id/content_hash**: honored (instance route = supplied route; hash from supplied value).
- **No active route**: omitted route + profile w/o active route → 409 `ErrApprovalRouteMissing`.
- **Replay**: same Idempotency-Key → same instance id, no duplicate row.
- Unit: resolver-nil + explicit route path still works (existing tests unaffected).
- `go build ./...` + `go test ./...` green.
