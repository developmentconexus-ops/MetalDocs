# Feature F3 — Evidence

> **Milestone:** 1 · **Feature:** `f3-error-mapping` · **Closed:** 2026-07-06
> **Contract:** `spec.md` (map submit sentinels to typed RFC 9457; never 500).

## What was implemented
- `MapErrorToResponse` (`http/errors.go:154-170`) gains four sentinel arms, each → its finalize-era
  status + a typed dot-notation `problem.Code`:
  `ErrRevisionTitleRequired`→422, `ErrDocumentNotDraft`→409, `ErrProfileNotConfigured`→400,
  `ErrApprovalRouteMissing`→409. Inline comments cite the finalize-era mapping each mirrors.
- Existing reason/idempotency/if-match/stale mappings untouched.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| Unit — sentinel→status table | `go test -count=1 -run TestMapErrorToResponse ./internal/modules/documents/approval/http/` | `ok` — 4 new rows in `errors_test.go` `TestMapErrorToResponse` assert each sentinel → its status + code: revision-title→422/`validation.revision_title_required`, not-draft→409/`state.document_not_draft`, profile→400/`validation.profile_not_configured`, route-missing→409/`state.approval_route_missing` | **real** |
| Static | `go build ./...` | `EXIT 0` | — |
| Runtime — no-route sentinel | docker curl no-route `POST /submit` | **409** `state.approval_route_missing` (problem+json, not 500) | **real** (docker) |
| Runtime — duplicate submission | re-submit under_review doc | **409** `conflict.duplicate_submission` (not 500) | **real** (docker) |
| Runtime — REV≥1 empty title | F1 integration `TestSubmitRev1RequiresReason_RealDB` path | 422 typed | **real** (testdb) |

> No 500s observed across the live QA transcript (`scratch_qa/M1-live-qa.md`). Every domain sentinel
> resolves to RFC 9457 problem+json.

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| Each sentinel → its status + non-unknown code | yes | http unit table + live 409 rows |
| REV≥1 empty-title submit → 422 typed | yes | F1 integration |
| build + test green | yes | build EXIT 0; approval http `ok` |

## Review disposition
- Spec-compliance: PASS — statuses match the pre-removal finalize mappings exactly (same contract, new
  entrypoint); no existing mapping changed.
- Code-quality: PASS — arms are typed codes in the module-local taxonomy; provenance comments keep the
  finalize→submit continuity legible.
- Milestone-validator (C1/C5 non-blocking finding, 2026-07-06): F3 evidence originally overstated
  HTTP-layer coverage — the four-arm `MapErrorToResponse` unit table did not exist (2 arms rested on
  code-inspection only). **Closed** by adding the four rows to `errors_test.go` `TestMapErrorToResponse`
  (green) and correcting the evidence row above to cite the real test. No production-code change — the
  mapping arms were already correct and live-proven.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| none | | |
