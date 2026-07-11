# Unit 2.5 — Route Builder v2 — Evidence

**Session:** orchestrator (Opus), worktree `clever-wiles-94d525`, budget ≤250k.
**Scope:** FE-only profile-governed approval route builder v2 (consumes G1 `governance_class`, merged main).
**Design spec:** `docs/superpowers/specs/2026-07-11-route-builder-v2-design.md`.
**Workflow spec:** `docs/superpowers/specs/2026-07-10-review-approval-workflow-model.md` (R1–R5).

## P0 gate
`developing-new-work` **skipped** — not a new backend module/feature/invariant. FE screen rebuild on an
already-merged contract (G1 ran its own gate at unit 2.1). Rationale recorded per HARNESS §2 (P0 = new
feature/module only).

## Bounded deviations (for HS-1)
1. v2 mock `route_builder_mock_v2` absent from repo — workflow spec §mock is the design authority;
   v1 `design-source/route-admin` is visual base. Noted; not a blocker.

## Dispatch ledger
| Slice | Implementer (sonnet) | Reviewer(s) (independent) | Verdict | Commit |
|---|---|---|---|---|
| S0 | _pending_ | | | |
| S1 | _pending_ | | | |
| S2 | _pending_ | | | |
| S3 | _pending_ | | | |
| S4 (review/QA) | — | frontend-screen-reviewer + frontend-code-reviewer | | |

## Verification ladder (from clean state)
- L0 typecheck (`pnpm exec tsc --noEmit -p tsconfig.build.json`) + lint: _pending_
- L1 vitest: _pending_
- L3 browser QA :80 (rendered UI, no passwords, curl-only=FAIL): _pending_

## Commands + outcomes
_appended per slice_
