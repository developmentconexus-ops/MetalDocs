# Program: Grade-A Architecture Remediation

> **Governing spec:** `docs/superpowers/specs/2026-06-14-grade-a-architecture-remediation-design.md`
> **Status:** In progress (M0, M1 passed 2026-06-14; test-infra-rebaseline micro-task before M2)
> **Owner / operator:** leandrotca.work (operator) + backend agent (Opus 4.8)

Take the backend's three formerly-C audit dimensions (module-boundaries/DDD, contract/API,
composition/observability) to **Grade A−/A**, and fully close the **H-D class** (handler/contract
field drift; tri-source route drift) and the **H-G class** (cross-module reach-without-a-port +
hardcoded domain state) — not just the instances. Every fix carries evidence; symptom-patching is a
hard-stop. **Terminal acceptance:** the M5 independent multi-agent re-audit passes the §6 pass bar
(3 dimensions ≥ A−, 0 new Critical/Major, H-D and H-G classes at 0) and the operator signs off Grade A.

## Milestones

| # | Milestone | Objective (one line) | Status | Gate result |
|---|-----------|----------------------|--------|-------------|
| 0 | `milestone-0-docs-destaling` | One unambiguous progression surface; stale docs stop polluting agent context | passed | [PASS](milestone-0-docs-destaling/qa/milestone-qa.md) |
| 1 | `milestone-1-reach-a-blockers` | Close all 4 Grade-A blockers + the error-contract (bare-405) tail | passed | [PASS](milestone-1-reach-a-blockers/qa/milestone-qa.md) |
| 2 | `milestone-2-contract-tail` | Eliminate handler-emits-undeclared-field drift (H-D), one FE regen | planned | — |
| 3 | `milestone-3-mechanical-quality` | Harden code-quality + persistence; dead-surface deletes, tx-hazard hoist | planned | — |
| 4 | `milestone-4-systemic-ports` | Close H-G class via shared ports (UserDisplayNameReader, TemplateVersionStateReader) | planned | — |
| 5 | `milestone-5-independent-re-audit` | Prove Grade A by independent fresh multi-agent re-audit (authoritative) | planned | — |

Status vocabulary: `planned` → `in-progress` → `passed` (operator-approved) / `blocked` (hard-stop open).

## Hard-stops raised

| When | HS id | What | Resolution |
|------|-------|------|------------|
| 2026-06-14 | HS-1 | M0 close gate — validator PASS presented to operator | **Approved** by operator 2026-06-14; M0 → passed, M1 opened |
| 2026-06-14 | HS-1 | M1 close gate — validator PASS (C1–C7) presented to operator | **Approved** by operator 2026-06-14 (option 2); M1 → passed. Condition: run a bounded **test-infra-rebaseline** micro-task before M2 to discharge the F1.3 AC5 full-HTTP-E2E defer. |
| (carry-forward) | HS-2 | F0.1 watch — FE eigenpal `file:` path defer | Schedule before any FE `pnpm install` / M2 start |

## Program close-out / reconciliation

Fill in only when the last milestone has passed:

- [ ] Every planned feature (M0–M4) has a complete evidence row.
- [ ] Zero unplanned scope merged; anything added is recorded with rationale.
- [ ] Every bounded defer has a written trigger and an owner.
- [ ] M5 re-audit passed the §6 pass bar — link the evidence.
- [ ] Forward roadmap (F0.3) reflects the executed program and any deferred triggers.
- [ ] Operator sign-off: <date / name>
