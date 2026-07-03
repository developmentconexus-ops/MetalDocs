# Program: Global Maximum Remediation

> **Governing spec:** `./mission.md`
> **Status:** M0 + M1 passed (operator-approved HS-1, 2026-07-03). M2 not yet started — start in a fresh session. Commits local, not pushed.
> **Owner / operator:** Leandro

Convert every finding of the 2026-07-03 final architecture review (commit 778f494a) into shipped, gate-enforced remediation: enforcement automation for the hand-sync defect class, kernel correctness, async consolidation onto River, the two ISO-core eQMS product gaps, tenant lifecycle, ops readiness, and governance hygiene. Terminal acceptance: an independent re-run of the 10-dimension review reaching CONFIRMED on every in-scope dimension, plus every mission-installed CI gate green from clean state with negative proof — judged by `mission-validator`.

## Milestones

| # | Milestone | Objective (one line) | Status | Gate result |
|---|-----------|----------------------|--------|-------------|
| 0 | `milestone-0-versionref-contract` | Land the planned VersionRef nested-ref contract cutover + ADR 0065 | passed (2026-07-03) | [PASS](milestone-0-versionref-contract/qa/milestone-qa.md) |
| 1 | `milestone-1-contract-fe-gates` | oasdiff CI gate, nullable⇒required lint, blocking contract-sync, ESLint feature boundaries | passed (2026-07-03) | [PASS](milestone-1-contract-fe-gates/qa/milestone-qa.md) |
| 2 | `milestone-2-authz-enforcement-generation` | Tripwire arms generated from capability registry + CI drift check; cap-name divergences closed | planned | — |
| 3 | `milestone-3-tenancy-chokepoint` | TxRunner auto-seed GUCs; RLS backstop covers worker/jobs; ADR 0027 amended | planned | — |
| 4 | `milestone-4-versioning-kernel` | Unified 9-status state machine; publish race proven safe; concurrency idiom unified | planned | — |
| 5 | `milestone-5-async-river-consolidation` | Janitors + staging outbox onto River; lease scheduler retired; outbox retention; fanout ordering | planned | — |
| 6 | `milestone-6-eqms-review-reason` | Periodic review/expiry + structured reason-for-change (ISO 9001 §7.5.3 / Part 11) | planned | — |
| 7 | `milestone-7-tenant-lifecycle` | Tenant onboarding, export, erasure design (crypto-shredding ADR) | planned | — |
| 8 | `milestone-8-ops-readiness` | Dockerfiles, distributed rate limiter, metrics/trace-correlation/backup posture | planned | — |
| 9 | `milestone-9-governance-hygiene` | ADR hygiene, REQ-ID traceability gate, test policy, CLAUDE.md/wiki truth, structure renames | planned | — |

Status vocabulary: `planned` → `in-progress` → `passed` (operator-approved) / `blocked` (hard-stop open). The **Gate result** column links the milestone-validator's verdict (`<milestone>/qa/milestone-qa.md`); `passed` requires a validator **PASS** *and* operator HS-1 approval.

**Mission-specific execution rules (from mission.md D2/D4):** each milestone runs in a fresh session dispatched via `spawn_task` with a self-contained context brief opening with `/goal`; each milestone authors `validation-contract.md` BEFORE implementation; runtime-visible milestones (M0, M5–M8) close with a live/preview QA drive. M5/M6/M7 require a `developing-new-work` gate (and M5/M7 an ADR) before planning.

## Hard-stops raised

| When | HS id | What | Resolution |
|------|-------|------|------------|
| | | | |

## Program close-out / reconciliation

Fill in only when the last milestone has passed:

- [ ] Every planned feature has a complete evidence row.
- [ ] Zero unplanned scope (anything added is recorded with rationale).
- [ ] Every bounded defer has a written trigger and an owner.
- [ ] **Terminal acceptance: main session re-runs the 10-dimension review fan-out, then DISPATCH `mission-validator`** to judge the artifact + gate evidence against mission.md §8 — verdict to `qa/mission-validation.md`.
- [ ] Operator sign-off: <date / name>
