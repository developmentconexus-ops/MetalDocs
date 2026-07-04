# Program: Global Maximum Remediation

> **Governing spec:** `./mission.md`
> **Status:** M0 + M1 + M2 passed (operator-approved HS-1, 2026-07-03). M3 milestone-validator **PASS** (2026-07-03) — negative-RLS proof **run GREEN for real** (validator executed it live itself after F3.5 retargeted it off the tripwired `documents`); **HS-1 operator gate pending**. Commits local, not pushed.
> **Owner / operator:** Leandro

Convert every finding of the 2026-07-03 final architecture review (commit 778f494a) into shipped, gate-enforced remediation: enforcement automation for the hand-sync defect class, kernel correctness, async consolidation onto River, the two ISO-core eQMS product gaps, tenant lifecycle, ops readiness, and governance hygiene. Terminal acceptance: an independent re-run of the 10-dimension review reaching CONFIRMED on every in-scope dimension, plus every mission-installed CI gate green from clean state with negative proof — judged by `mission-validator`.

## Milestones

| # | Milestone | Objective (one line) | Status | Gate result |
|---|-----------|----------------------|--------|-------------|
| 0 | `milestone-0-versionref-contract` | Land the planned VersionRef nested-ref contract cutover + ADR 0065 | passed (2026-07-03) | [PASS](milestone-0-versionref-contract/qa/milestone-qa.md) |
| 1 | `milestone-1-contract-fe-gates` | oasdiff CI gate, nullable⇒required lint, blocking contract-sync, ESLint feature boundaries | passed (2026-07-03) | [PASS](milestone-1-contract-fe-gates/qa/milestone-qa.md) |
| 2 | `milestone-2-authz-enforcement-generation` | Tripwire arms generated from capability registry + CI drift check; cap-name divergences closed | passed (2026-07-03) | [PASS](milestone-2-authz-enforcement-generation/qa/milestone-qa.md) |
| 3 | `milestone-3-tenancy-chokepoint` | TxRunner auto-seed GUCs; RLS backstop covers worker/jobs; ADR 0027 amended | validator PASS — HS-1 pending (2026-07-03) | [PASS](milestone-3-tenancy-chokepoint/qa/milestone-qa.md) |
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
| M3 close (2026-07-03) | HS-4 | milestone-validator FAIL, finding F-1: three durable records falsely claimed `idempotency_keys` has no `tenant_id` / RLS N/A (split-brain vs schema + ADR 0027 body) | Opened fix-feature F3.4 (idempotency-keys-rls-truth); corrected all 3 records + `async-tenant-tables.txt`; re-dispatched validator → **PASS**. Docs-only, RLS byte-identical. |
| M3 F3.4 (2026-07-03) | HS-7 | F-1 fix required editing the committed D4 validation-contract | Operator approved the re-open; corrected §0.3/§2.4/§4 in place + dated auditable erratum; acceptance bar unchanged. |
| M3 real-run (2026-07-03) | HS-4 | Operator required a REAL run of the F3.2 negative-RLS proof (no defer). Real run RED: proof targeted `documents`, whose M2 capability write-tripwire (P0001) fires before the RLS tenant policy — leak-before + 42501 assertions invalid. Core isolation subtest (0-row SELECT/UPDATE/DELETE under `SeedTxTenant`) passed. | **RESOLVED** — F3.5 (rls-proof-real-green) retargeted the proof to `metaldocs.notifications` (FORCE-RLS tenant table, real F3.2 async seed site, NOT tripwired); ran GREEN for real (validator re-ran it live itself); defer closed; validator re-dispatched → **PASS**. |

## Program close-out / reconciliation

Fill in only when the last milestone has passed:

- [ ] Every planned feature has a complete evidence row.
- [ ] Zero unplanned scope (anything added is recorded with rationale).
- [ ] Every bounded defer has a written trigger and an owner.
- [ ] **Terminal acceptance: main session re-runs the 10-dimension review fan-out, then DISPATCH `mission-validator`** to judge the artifact + gate evidence against mission.md §8 — verdict to `qa/mission-validation.md`.
- [ ] Operator sign-off: <date / name>
