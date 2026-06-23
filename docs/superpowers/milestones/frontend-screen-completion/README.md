# Program: Frontend Screen Completion

> **Governing spec:** `./mission.md`
> **Status:** In progress — M0 + M1 + M2 + M3 + M4 passed (validator PASS + operator HS-1); **M5 next**
> **Owner / operator:** leandrotca

Finish every remaining/partial frontend screen to Professional-SaaS, industry-grade quality, matching the Grade-A backend (HEAD `d477e9f0`) and the canonical frontend architecture rules. Screens that are blocked on missing backend endpoints (Distribuição fanout, Notifications) get those endpoints built full-stack to the Grade-A bar (D1). **Terminal acceptance:** an independent `mission-validator` judges a per-screen completion re-audit — every in-scope screen real-API + redesign-tokens + both reviewer agents APPROVE, zero mock/dead-route/stub remaining, new backend `api-lint -strict` = 0 with all 6 CI guards green. Evidence base: `./discovery-brief.md`.

## Milestones

| # | Milestone | Objective (one line) | Status | Gate result |
|---|-----------|----------------------|--------|-------------|
| 0 | `milestone-0-truth-reset` | Honest routed app: 1 index route, no dead stubs, correct tracker, per-screen DoD written | passed | [PASS](milestone-0-truth-reset/qa/milestone-qa.md) |
| 1 | `milestone-1-dashboard-real-data` | Home screen renders 100% live data (kill MOCK_STATS/MOCK_ACTIVITY) | passed | [PASS](milestone-1-dashboard-real-data/qa/milestone-qa.md) |
| 2 | `milestone-2-distribuicao` | Distribuição coverage-scope (derive-on-read): Grade-A read-only endpoint serves the real obligated-reader set via two new owner-published views (CD + taxonomy) + distribution module; numerator honestly "tracking pending" (parked mission) | passed | [PASS](milestone-2-distribuicao/qa/milestone-qa.md) |
| 3 | `milestone-3-notifications` | Notifications center real end-to-end (new backend + lifecycle emitters + wired screen) | passed (validator PASS + HS-1 2026-06-23) | [PASS](milestone-3-notifications/qa/milestone-qa.md) |
| 4 | `milestone-4-publicado-obsoleto` | Publicado "em breve" gaps closed + Documento Obsoleto variant built | passed (validator PASS + HS-1 2026-06-23) | [PASS](milestone-4-publicado-obsoleto/qa/milestone-qa.md) |
| 5 | `milestone-5-signoff-taxonomy` | Detalhe Signoff screen built + Taxonomy Admin restyled to tokens (net-new/polish, last) | planned | — |

Status vocabulary: `planned` → `in-progress` → `passed` (operator-approved) / `blocked` (hard-stop open). The **Gate result** column links the milestone-validator's verdict (`qa/milestone-qa.md`); `passed` requires a validator **PASS** *and* operator HS-1 approval.

## Hard-stops raised

| When | HS id | What | Resolution |
|------|-------|------|------------|
| 2026-06-21 (M2/F2.1 mid-feature) | HS-6 | F2.1 spec promised per-recipient `area_code`/`source` + by-area coverage + company-scope obligated set from `metaldocs.v_cd_grantee`; recon revealed view doesn't carry those cols and is restricted-only by search-semantic contract (migration 0243). Distribution module can't raw-read CD/taxonomy base tables (ADR-0039). | Operator picked Option A (new owner-published sibling views) after evidence-based subagent recon. M2 re-decomposed: F2.1a (CD publishes `v_cd_obligated_readers`) + F2.1b (taxonomy publishes `v_process_area_name`) + F2.1c (distribution contract) + F2.2 + F2.3. Execution restarts in fresh `/milestone` session. |
| 2026-06-22 (M3/F3.3 pre-spec recon) | HS-2 + HS-6 | F3.3 recon revealed the locked `event_type → message` mapping does not match outbox runtime truth: `governance_events` is an **actor-centric audit log**, not a domain-event substrate — it records `actor_user_id` (who acted), never the recipient or business outcome. `signoff.rejected` fires only on **eligibility-failure** (not document rejection); `signoff_recorded` fires on **every** signoff with no terminal-outcome marker in payload. Building author-targeted notifications on top requires reverse-engineering recipients/outcomes via compensating views — a workaround the operator rejected. | **F3.2 (read surface) preserved** — table + endpoints are consumer-side, reusable. **F3.3 paused.** Operator directed root-cause fix: design a proper domain-event pattern (owning module emits outcome/recipient-bearing events), researched against industry/SaaS/OSS standards, applied as the codebase standard — *"redesign, not fully"*, incremental. **RESOLVED 2026-06-22:** research synthesized (token-lean, salvaged from the aborted deep-research run); operator picked **Option A — typed River-job domain events**; **ADR-0044** authored (supersedes ADR-0043 §6, eliminates both compensating views); **HS-2 partially lifted** (additive in-tx enqueue only; publish/approval semantics stay sacred); F3.3 spec re-scoped + **Approved (pre-code) 2026-06-22**. Contract type lands in `documents/domain` (module-boundaries-mandated). plan.md + TDD execute in a fresh session. |

## Program close-out / reconciliation

Fill in only when the last milestone (M5) has passed **and** the terminal gate is green:

- [ ] Every planned feature (M0..M5) has a complete evidence row.
- [ ] Zero unplanned scope (anything added is recorded with rationale).
- [ ] Every bounded defer has a written trigger and an owner.
- [ ] **Terminal acceptance — dispatch `mission-validator`** (`.claude/agents/mission-validator.md`): after M5 passes its milestone-validator + HS-1, the closing session runs the §8 per-screen re-audit fan-out (main session) and dispatches `mission-validator` to judge it + write `qa/mission-validation.md`. PASS required before sign-off; on FAIL → HS-5 micro-milestone.
- [ ] Terminal acceptance passed — link `qa/mission-validation.md`.
- [ ] Operator sign-off: <date / name>
