# Unit 2.2 (G2) — request_changes on approval stages — Evidence

**Date:** 2026-07-11
**Branch:** claude/friendly-darwin-ef6a11 (worktree)
**Spec:** docs/superpowers/specs/2026-07-10-review-approval-workflow-model.md — rule R3; §4 gap G2
**Budget:** ≤150k tokens

## Scope (locked)
Relax `ErrVerdictWrongStageKind` so that on **approval-kind** stages the verdict
`request_changes` is allowed (comment-only, unsigned); `ready` remains forbidden on
approval stages. Enforced in BOTH layers (service pre-check + `domain.NewVerdict`).
No DB change (migration 0286 has no verdict×stage CHECK). Signing kernel's signed-reject
(`signature_meaning='rejection'`) untouched.

## Design decisions (pre-implementation orientation)
- Owning module: `documents/approval` (nested kernel, ADR 0072). No cross-module edge added.
- Invariants touched: contract-first (OpenAPI is route truth); RFC 9457 problem+json; authz
  unchanged (verdict path already `authz.Require(CapApprovalReview)` in-tx). H-PRE-1 N/A (no new
  lock-holding tx; no new authz-recording read).
- Exactly TWO stage kinds exist (`review`, `approval` — domain/route.go:39-40). Post-relax the
  only forbidden verdict case is `ready`-on-approval → new sentinel `ErrVerdictReadyOnApprovalStage`.
  `ErrVerdictWrongStageKind` retained as the defensive default for an unknown/empty stage kind.
- request_changes on an approval stage reuses the EXISTING request_changes branch (under_review→draft
  revert). Verified correct: document stays `under_review` through approval stages (submit sets
  draft→under_review; decision_service signoff-reject from an approval stage does the identical
  under_review→draft revert with no explicit freeze-thaw). Instance goes terminal (changes_requested);
  stale frozen hash on the terminal instance is irrelevant — mirrors decision_service reject precedent.
  No stage-status mutation added (matches existing review-stage request_changes semantics).
- Contract impact: endpoint can now return 422 (business-rule rejection of `ready`-on-approval),
  matching the freeze-effective-date / reason-for-change 422 precedent. Description updated;
  '422' response added; api.gen.go regenerated via oapi-codegen (never hand-edited).

## Slices
1. domain guard matrix (`NewVerdict`) + domain unit tests
2. service pre-check mirror (`RecordVerdict`) + integration tests (testdb)
3. contract-first OpenAPI + HTTP error mapping

## Dispatch ledger
| Slice | Agent (model) | Role | Verdict |
|---|---|---|---|
| 1 | sonnet general-purpose (ac71c086) | implement | GREEN (domain tests) |
| 1 | cavecrew-reviewer (sonnet, aab164bb) | review | 0🔴 1🟡 → corrective applied (negative test + stale-doc fix), re-green |
| 1 | orchestrator | verify+commit | build+domain tests green; committed |
| 2 | sonnet general-purpose (slice2-impl) | implement | service pre-check switch + integration test edits + fixture SeedWithCaps wrap |
| 2b | sonnet general-purpose (af721105) | implement | app-layer fake unit test (integration harness pre-broken) — PASS |
| 2 | cavecrew-reviewer (sonnet, a2656 1a) | review | 0🔴 1🟡 → app unit test didn't isolate service guard (domain backstop raises same sentinel). Corrective: acting actor made ineligible so guard must fire before eligibility; re-green |
| 3 | sonnet general-purpose (slice3-impl) | implement | openapi 422 + errors.go codes + regenerated api.gen.go + errors_test cases |
| 3 | cavecrew-reviewer (sonnet, a6ce03af) | review | 1🟡 → ErrVerdictWrongStageKind (unreachable internal-state) mapped 422 but sibling precedent reserves 500. Corrective: remapped to 500 + `internal.verdict_wrong_stage_kind` code + generic title; re-green. Confirmed (a) openapi 422 via shared $ref, (b) errors.Is used, (c) api.gen.go codegen-shaped only |
| all | orchestrator | verify+commit | L0 green; touched-pkg tests green; committed slices 2/2b/3 |

## Verification ladder
- **L0 — PASS** (with pre-existing vendor caveat):
  - `go build ./...` → FAILS repo-wide on a **pre-existing** vendor gap (missing `go.opentelemetry.io/proto/otlp/logs/v1`, `.../collector/logs/v1`, `github.com/redis/go-redis/v9/internal/maintnotifications/logs` under vendor/). Reproduces on untouched base; independent of unit 2.2.
  - `GOFLAGS=-mod=mod go build ./...` → **EXIT 0** (real signal on this change; build-flag only, no code change).
  - `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` → **0 violation(s)**.
  - `check-module-boundaries.ps1` → **[module-boundaries] OK**.
  - api.gen.go: two consecutive `go generate` produce identical md5 (deterministic pure codegen, not hand-edited); module rebuilds clean.
- **L1 — PASS (unit) / DEFERRED (integration)**:
  - `go test ./internal/modules/documents/approval/{domain,application,http,http/contracts}/...` → **all ok** (domain guard matrix, service pre-check fake test, http 422/500 mapping).
  - Integration suite (`tests/integration/approval`) NOT runnable here — see bounded defer #2. Edited integration tests reviewed correct-on-paper.
- **L2 — ASK (contract changed)**: review-verdict endpoint gained a `422` response + description reword. Live :80-stack exercise is owned by the hub; flagged for hub coordination, not run in this worktree.

## Bounded defers
1. **Vendor gap (pre-existing, repo-wide).** `go build ./...`/`go test ./...` fail under `-mod=vendor` due to three missing vendored packages (otlp logs, redis maintnotifications). Reproduces on the untouched base commit; NOT introduced by unit 2.2. Local verification used `GOFLAGS=-mod=mod`. Fix = `go mod vendor` refresh — out of this unit's locked scope; flag to hub.
2. **Integration-harness identity/capability breakage (pre-existing).** `tests/integration/approval` fixtures build contexts with `context.Background()`, so `metaldocs.actor_id` GUC is unseeded and authz hard-fails (`authz: metaldocs.actor_id GUC not set`); the same failure reproduces on untouched committed `freeze_integration_test.go`. Fixtures would additionally need per-actor capability grants. Slice 2's service guard is instead proven by the application-layer fake unit test (`TestRecordVerdict_ReadyOnApprovalStageRejected`, isolated via an ineligible actor). Repairing the integration harness is broader than unit 2.2; flag to hub.
3. **L2 live contract exercise.** Deferred to hub (owns the :80 container stack) — see L2 above.
