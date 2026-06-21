# Program: Frontend Screen Completion

> **Governing spec:** `./mission.md`
> **Status:** Planning
> **Owner / operator:** leandrotca

Finish every remaining/partial frontend screen to Professional-SaaS, industry-grade quality, matching the Grade-A backend (HEAD `d477e9f0`) and the canonical frontend architecture rules. Screens that are blocked on missing backend endpoints (Distribuição fanout, Notifications) get those endpoints built full-stack to the Grade-A bar (D1). **Terminal acceptance:** an independent `mission-validator` judges a per-screen completion re-audit — every in-scope screen real-API + redesign-tokens + both reviewer agents APPROVE, zero mock/dead-route/stub remaining, new backend `api-lint -strict` = 0 with all 6 CI guards green. Evidence base: `./discovery-brief.md`.

## Milestones

| # | Milestone | Objective (one line) | Status | Gate result |
|---|-----------|----------------------|--------|-------------|
| 0 | `milestone-0-truth-reset` | Honest routed app: 1 index route, no dead stubs, correct tracker, per-screen DoD written | planned | — |
| 1 | `milestone-1-dashboard-real-data` | Home screen renders 100% live data (kill MOCK_STATS/MOCK_ACTIVITY) | planned | — |
| 2 | `milestone-2-distribuicao` | Distribuição backed by a Grade-A fanout/coverage endpoint; live coverage, no illustrative data | planned | — |
| 3 | `milestone-3-notifications` | Notifications center real end-to-end (new backend + wired screen) | planned | — |
| 4 | `milestone-4-publicado-obsoleto` | Publicado "em breve" gaps closed + Documento Obsoleto variant built | planned | — |
| 5 | `milestone-5-signoff-taxonomy` | Detalhe Signoff screen built + Taxonomy Admin restyled to tokens (net-new/polish, last) | planned | — |

Status vocabulary: `planned` → `in-progress` → `passed` (operator-approved) / `blocked` (hard-stop open). The **Gate result** column links the milestone-validator's verdict (`qa/milestone-qa.md`); `passed` requires a validator **PASS** *and* operator HS-1 approval.

## Hard-stops raised

| When | HS id | What | Resolution |
|------|-------|------|------------|
| | | | |

## Program close-out / reconciliation

Fill in only when the last milestone (M5) has passed **and** the terminal gate is green:

- [ ] Every planned feature (M0..M5) has a complete evidence row.
- [ ] Zero unplanned scope (anything added is recorded with rationale).
- [ ] Every bounded defer has a written trigger and an owner.
- [ ] **Terminal acceptance — dispatch `mission-validator`** (`.claude/agents/mission-validator.md`): after M5 passes its milestone-validator + HS-1, the closing session runs the §8 per-screen re-audit fan-out (main session) and dispatches `mission-validator` to judge it + write `qa/mission-validation.md`. PASS required before sign-off; on FAIL → HS-5 micro-milestone.
- [ ] Terminal acceptance passed — link `qa/mission-validation.md`.
- [ ] Operator sign-off: <date / name>
