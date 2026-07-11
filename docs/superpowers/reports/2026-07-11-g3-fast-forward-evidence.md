# Evidence — Unit 2.3 (G3 fast-forward "Aprovar já")

**Date:** 2026-07-11 · **Branch:** `claude/jovial-cerf-6f376c` (chip worktree) · **Spec:** R5 + §4 G3 of `docs/superpowers/specs/2026-07-10-review-approval-workflow-model.md` · **Gate:** `docs/superpowers/analysis/2026-07-11-g3-fast-forward-system-impact.md` (🟢 Green, feature, no new capability, no ADR-triggering deviation) · **Plan:** `docs/superpowers/reports/2026-07-11-g3-fast-forward-plan.md`

## Commit range

| Commit | Slice |
|---|---|
| 216c2bbf | docs: system-impact gate + plan |
| 02341ef5 | S1 — eligibility surfacing (`fast_forward_eligible` + `next_stage_id` in ReviewVerdictResponse; contract-first + regen) |
| 674682e8 | S2 — behavior-preserving extraction of `recordVerdictInTx` / `recordSignoffInTx` (zero test edits) |
| b6ff9795 | S3 — `FastForwardService.RecordFastForward` (one tx, two ledger writes, fail-closed sentinels) + contract path |
| e3730e5d | S4 — HTTP handler + distinct idempotency template + tier-1 cap row + integration suite |
| (this)   | S5 — wiki sync (approval.md §6.2b etc.) + evidence |

## Dispatch ledger (harness §4 obligations 1–6)

| Slice | Implementer | Reviewer(s) | Verdict |
|---|---|---|---|
| Investigation (bulk read) | sonnet investigator a834dc0ba0c1370fe | — | compressed map delivered |
| S1 | sonnet a...eca4 (aaeca4ddacdcf7cc6) | sonnet ad35783a68b982874 | ACCEPT + 1 MINOR (duplicate delegations query) — fixed inline by orchestrator (2-line reuse of in-scope `delegations` var; within §4.1 trivial-glue allowance), re-verified green |
| S2 | sonnet a227f868574d1bf5f | sonnet a4fb2fae72c3ee97f | ACCEPT, no findings |
| S3 | sonnet a30cb969a267dee31 | sonnet aea34fe056cd4ed2b | ACCEPT, no findings (501-stub deviation declared + accepted; replaced in S4) |
| S4 | sonnet a7a6ba489a0cce8f0 | sonnet a34695069e53c6291 | ACCEPT + 1 finding (silent fail-open idemp wiring) → remediation dispatched to sonnet a1ae885d98edbbdab (compile-time assert + nil-store 500 fail-closed + tests), re-verified green |
| S5 wiki | wiki-curator (sonnet) a9603747be0f2d8a7 | — (docs-only) | applied |

Task board: 5 tasks (S1–S5), sequential blockedBy chain, all completed at reviewed-green. No harness violations: main session wrote no production code beyond the 2-line reviewer-directed fix noted above.

## Done criteria vs delivery

1. **Eligibility per R5 (a)∧(b), in contract** — `ReviewVerdictResult/Response.fast_forward_eligible` + `next_stage_id`, computed in `recordVerdictInTx` after `AdvanceStage()`: next active stage approval-kind ∧ `domain.ResolveEligibleIdentity` (same composition as signoff, delegations included). Probe failure = hint miss, no event, no error. Spec-first + regen (never hand-edited api.gen.go).
2. **Two ledger entries, one tx, idempotent** — `POST /approval/instances/{instance_id}/stages/{stage_id}/fast-forward` → single `runner.Do`: verdict core (ready) → `ErrFastForwardStageNotCompleted` if quorum unsatisfied → `ErrFastForwardNotEligible` if no eligible now-active approval stage → signoff core (approve). Rows in `approval_review_verdicts` + `approval_signoffs` + 2 `governance_events`, never collapsed; any leg failure rolls back everything (integration-proven for not-eligible AND content-hash mismatch). Idempotency: dedicated `fastForwardRouteTemplate` in the signoff idemp store, handler replay short-circuit, fail-closed 500 when store unwired (reviewer-driven hardening). H-PRE-1: single off-tx `LoadActorDisplayName`; `authz.Require` in writable tx per module norm. Tier-1 `CapDocumentSignoff`; tier-2 in-tx `CapApprovalReview` + signoff cap (both cores).
3. **Freeze boundary unchanged** — `executeFreeze` call site untouched (verdict core, review→approval crossing); fast-forward signature leg validates client `content_hash` against the hash frozen earlier in the same tx. `ready` on approval-kind stage still rejected (`ErrVerdictReadyOnApprovalStage`) — G2 regression covered at unit + integration level.

## Verification ladder (from clean state, `go clean -testcache`)

| Gate | Command | Result |
|---|---|---|
| L0 build | `go build ./...` | clean |
| L0 contract lint | `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` | 0 violations |
| L0 boundaries | `.\scripts\check-module-boundaries.ps1` | OK |
| L1 unit | `go test ./...` | zero FAIL lines |
| L1 integration | `.\scripts\test-integration.ps1` (canonical, DATABASE_URL derived) | FAILs = exactly the 9 accepted RED: sla_surfacer ×4, controlleddocuments cross-tenant sequence ×1, scenarios ×3, tenantdata ×1 (E-PROD-1..5). **No NEW failures.** |
| Targeted | `test-integration.ps1 -Run 'TestFastForward|TestReviewVerdict'` | PASS (6 fast-forward cases + all verdict cases incl. 2 new eligibility cases) |

Integration cases: happy path (row counts both tables, 2 governance events, frozen hash set, instance/document approved), not-eligible atomic rollback, stage-not-completed, G2 regression, SoD author block, content-hash mismatch rollback. L2/L3 not run — feature unit; UI surface is unit 2.4.

## Defers / notes (bounded)

- **Replay-response caveat:** on idempotent replay of a plain review verdict, `fast_forward_eligible` returns false (handler short-circuits before service). Hint-only field; FE (unit 2.4) should treat as advisory.
- **Pre-existing quirk (NOT introduced, NOT fixed):** review-verdict handler reuses `stageSignoffRouteTemplate` for idempotency namespacing; fast-forward got its own template. Also `verdict_id`/`signoff_id` response fields are `""` class-wide (pre-existing pattern; fast-forward mirrors siblings).
- **T-016 (wiki tech-debt entry):** wiki-curator flagged the composed-write pattern as lacking a dedicated ADR. Gate §9 judged no ADR required (in-bounds feature of ratified spec R5). Carried for hub/HS-1 disposition.
- Post-tx eligibility-rejection governance event emission (pre-existing wrapper pattern) inherited unchanged by fast-forward wrapper.

## HS-1 items

- T-016 ADR-or-not disposition (above).
- FE consumer guidance for `fast_forward_eligible` replay semantics → feeds unit 2.4.
