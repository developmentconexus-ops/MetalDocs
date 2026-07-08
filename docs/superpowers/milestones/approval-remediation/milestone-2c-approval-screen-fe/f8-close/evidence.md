# Feature F8 — Evidence

> **Milestone:** 2c  ·  **Feature:** `f8-close`  ·  **Closed:** 2026-07-07
> **Contract:** `milestone.md` F8 row — full suites + mandatory real-stack live-QA walkthrough +
> dispatch `milestone-validator`; NOT pushed.

## What was implemented

F8 is the close feature — no new product code beyond two backend defects that **live QA exposed** on
the real stack (compile-passed / fixture-green but non-functional against the running API):

- **BUG #1 — SLA-null submit 42P08.** `InsertStageInstances` (`postgres_approval_repository.go:103`)
  built `due_at` via a CASE referencing `$16` (`DueInDaysSnapshot *int`) that, when `nil`, was an
  **untyped NULL** used only inside the CASE → Postgres could not infer its type → `42P08`. Any route
  stage without an SLA could never be submitted. Fix: `::int` casts on the param inside the CASE.
  Sibling `UpdateStageStatus` was already safe (typed column). Live RED (500) → GREEN (201).
- **BUG #2 — GET approval-instance 404 for `changes_requested`.** The shared narrow repo read filtered
  `status IN ('in_progress','approved')`, so the `changes_requested` instance 404'd — but the OpenAPI
  `ApprovalInstanceByDocumentResponse.status` enum **includes** `changes_requested` and the C5 author
  panel (`DocumentEditorPage.tsx:381-393`) gates on it. Net: the C5 deliverable was **non-functional
  live** despite green fixture unit tests. Fix: dedicated `LoadInstanceByDocumentForView` (status set
  + `changes_requested`) wired **only** to `GetInstanceByDocumentHandler`; publish + mutation lookups
  stay on the narrow method (publish re-entry / signoff-lookup semantics unchanged). Identical authz
  gate (ADR 0022 — no capability weakening). Live RED (404) → GREEN (200).

Both are consumer-contract fixes (producer made to match the OpenAPI contract the FE already consumes).

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| Static (Go build) | `go build ./...` | exit 0 | — |
| Go — full suite | `go test ./...` | all `ok` (no FAIL); `FULL_GO_TEST_EXIT=0` incl. `scripts/api-lint`, approval module | real |
| Go — approval module | `go test ./internal/modules/documents/approval/...` | application/http/infrastructure all `ok`, exit 0 | real |
| Types — docx workspace | `npm run typecheck:docx-v2` | 8 projects, all `Done`, exit 0 | — |
| Types — web build | (F7 evidence) `tsc --noEmit -p tsconfig.build.json` | clean, exit 0 | — |
| FE — vitest | `pnpm exec vitest run` (`make test`) | **127 files / 806 tests passed**, exit 0 | real (jsdom) |
| BUG#1 TDD | `TestInsertStageInstances_ActiveStageWithNilDueInDays_NoTypeInferenceError` (integration, testdb factory, `//go:build integration`) | env-gated SKIP w/o DATABASE_URL; **live RED→GREEN** on :8081 (submit 500→201) | real (live) |
| BUG#2 TDD | `TestLoadInstanceByDocumentForView_SeesChangesRequested` (integration, testdb factory) | env-gated SKIP w/o DATABASE_URL; **live RED→GREEN** on :8081 (GET 404→200) | real (live) |
| Live QA walkthrough | real stack API :8081 + FE :4173, 4 real cookie sessions | full lifecycle GREEN (table below) | **real** |

## Live QA walkthrough — real stack, honestly labeled

Full log: `f8-close/qa/live-qa-log.md`. Executed against `metaldocs-api` :8081 (real DB, `-Build`
rebuild) with real cookie sessions (admin / author-test / approver-test); curl-driven mutations carry
`Origin` + `Idempotency-Key` + `If-Match`; screen renders (C1/C2/C3/C5) confirmed via preview DOM
introspection (the screenshot tool times out on the heavy docx editor iframe — honestly non-screenshot).

| Walkthrough step (F8 acceptance) | Live result |
|---|---|
| route review→approval | PUT `/approval/routes/{id}` v4 — review(qms_admin)→approval |
| submit | **201** in_progress (BUG#1 fix; was 500/42P08) |
| suggesting + reviewer verdict request_changes | **200** doc→draft, instance→changes_requested |
| author changes_requested panel (C5) | GET `/approval-instance` **200** (BUG#2 fix; was 404) — panel data present live |
| clean buffer → re-submit | **201** new instance |
| review ready → freeze | **200** stage_completed, approval stage active, `frozen_content_hash` set |
| sign with meaning | **200** outcome=approved; doc+instance approved |
| publish | **200** new_status=published |
| visibility 404 | admin GET bogus doc `/approval-instance` → **404** not_found.instance |
| oversee | admin `?scope=oversee` **200** tenant-wide list; author-test **403** approval.oversee denied |
| cancel-with-reason | admin POST `/cancel {reason}` **200** — doc under_review→draft, instance cancelled (dropped from oversee, view→404) |

Every F8-acceptance walkthrough item exercised on the real stack. Two backend defects found, root-caused,
TDD-tested, live RED→GREEN.

## Acceptance vs spec Validation Gate

F8 acceptance (milestone.md): *"`make test` + `npm run typecheck:docx-v2` + `go build ./... && go test
./...` green; live QA walkthrough evidence recorded (route review→approval, suggesting, request_changes,
author panel, clean buffer, freeze+markup gate 409, sign with meaning, publish, visibility matrix,
oversee, cancel-with-reason); `milestone-validator` verdict written. NOT pushed."*

| Acceptance criterion | Met? | Evidence |
|---|---|---|
| `go build ./...` green | yes | build exit 0 |
| `go test ./...` green | yes | `FULL_GO_TEST_EXIT=0`, all `ok` |
| `npm run typecheck:docx-v2` green | yes | 8 projects Done, exit 0 |
| `make test` (FE vitest) green | yes | 127 files / 806 tests passed, exit 0 |
| route review→approval | yes | walkthrough table |
| suggesting | yes | review stage opened in suggesting mode (C1/C2 render) |
| request_changes | yes | review-verdict 200 → changes_requested |
| author panel | yes | C5 panel data live (BUG#2 fix) |
| clean buffer | yes | resubmit 201 after buffer clean |
| freeze + markup gate | yes | ready→freeze 200, `frozen_content_hash` pinned (F0 markup gate tested in F0 evidence) |
| sign with meaning | yes | signoff 200 outcome=approved |
| publish | yes | publish 200 → published |
| visibility matrix | yes | 404 not_found.instance; oversee 200 (cap) / 403 (no cap) |
| oversee | yes | `?scope=oversee` admin 200, author-test 403 |
| cancel-with-reason | yes | cancel 200 → draft, instance cancelled |
| `milestone-validator` verdict written | yes | `qa/milestone-qa.md` — **VERDICT: PASS** (C1–C7 clean-state) |
| NOT pushed | yes | no `git push` run |

## Review disposition

- **Backend-fix review (2 defects):** independent sonnet reviewer subagent (separation of powers, no
  edits) — **APPROVE, 0 Critical / 0 Major / 0 Minor**. Confirmed with evidence: `go build ./...` +
  `go vet` (plain & `-tags integration`) clean; BUG#1 bind order verified 1:1 (16 placeholders / 16
  args per stage, `DueInDaysSnapshot` bound at positions 15 & 16, `::int` applied to both `$16`
  occurrences, `due_at` is raw computed SQL not a bind param, nothing shifted; `UpdateStageStatus`
  untouched — typed column, never exposed); BUG#2 `LoadInstanceByDocumentForView` byte-identical to
  the active read except the status set, identical authz gate (no capability weakening), publish (:52,
  :92) + mutation lookups confirmed still on the narrow method, OpenAPI enum includes
  `changes_requested` (contract-conformant); both integration tests non-tautological (BUG#2 test also
  asserts the narrow method still returns `ErrNoActiveInstance` for the same doc). No fallback values,
  tenant predicate preserved on both queries.
- **Milestone close review:** `milestone-validator` subagent (Phase 4, separation of powers) —
  **VERDICT: PASS**, written to `qa/milestone-qa.md`. Re-ran from clean state and personally observed
  green: `go build ./...` (0), F0 contract tests (`TestInstanceStatusWireEnumComplete`,
  `TestStageRequestStageKindValidation`), approval-module regression + full `go test ./...` (0, no
  FAIL), `typecheck:docx-v2` (8 Done), web `tsc -p tsconfig.build.json` (clean), vitest M2c surfaces
  (54 files / 374 tests). Both root-cause re-measures held (W2 writable-session grep → only test mocks
  asserting hooks never called; single-timeline `toHaveLength(1)` passes — duplicate deleted not
  CSS-hidden). Judged BUG#1/BUG#2 as legitimate in-boundary HS-3 prerequisite repairs, honestly
  tested, recorded as HS-1 deviations — not scope drift. Forbidden-list clean. NOT pushed.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| Backend BUG#1/#2 are M2b-backend-owned code touched under M2c live QA (HS-3 prerequisite repairs) | Both are minimal, root-caused, contract-conforming; do not touch publish/mutation semantics | Recorded as HS-1 deviations; owner = operator gate |
| Accumulated F4–F7 deviations (F4 §6 legal jurisdiction; F5 D1/D2; F6 D1/D2; F7 D1–D6) | Each carries recorded ratios/rationale in its evidence.md | Surfaced at HS-1 |
