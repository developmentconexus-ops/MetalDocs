# Program: Grade-A Architecture Remediation

> **Governing spec:** `docs/superpowers/specs/2026-06-14-grade-a-architecture-remediation-design.md`
> **Status:** In progress (M0, M1 passed 2026-06-14; test-infra-rebaseline micro-task done 2026-06-14; M2 passed 2026-06-14 — validator C1–C7 PASS; M3 passed 2026-06-15 — validator C1–C7 PASS, awaiting operator HS-1 to open M4)
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
| 2 | `milestone-2-contract-tail` | Eliminate handler-emits-undeclared-field drift (H-D), one FE regen | passed | [PASS](milestone-2-contract-tail/qa/milestone-qa.md) |
| 3 | `milestone-3-mechanical-quality` | Harden code-quality + persistence; dead-surface deletes, tx-hazard hoist | passed | [PASS](milestone-3-mechanical-quality/qa/milestone-qa.md) |
| 4 | `milestone-4-systemic-ports` | Close H-G class via shared ports (UserDisplayNameReader, TemplateVersionStateReader) | planned | — |
| 5 | `milestone-5-independent-re-audit` | Prove Grade A by independent fresh multi-agent re-audit (authoritative) | planned | — |

Status vocabulary: `planned` → `in-progress` → `passed` (operator-approved) / `blocked` (hard-stop open).

## Hard-stops raised

| When | HS id | What | Resolution |
|------|-------|------|------------|
| 2026-06-14 | HS-1 | M0 close gate — validator PASS presented to operator | **Approved** by operator 2026-06-14; M0 → passed, M1 opened |
| 2026-06-14 | HS-1 | M1 close gate — validator PASS (C1–C7) presented to operator | **Approved** by operator 2026-06-14 (option 2); M1 → passed. Condition: run a bounded **test-infra-rebaseline** micro-task before M2 to discharge the F1.3 AC5 full-HTTP-E2E defer. **Condition met 2026-06-14** — full HTTP `seed→finalize→signoff` E2E green, snapshot read-back `matches=t`; evidence `milestone-1-reach-a-blockers/test-infra-rebaseline/evidence.md`. M2 awaits operator go. |
| 2026-06-14 | HS-1 | M2 open gate — operator approved opening M2 ("Open M2 — spec it") | **Approved** by operator 2026-06-14; M2 → in-progress, `milestone.md` authored up front before any feature |
| (carry-forward) | HS-2 | F0.1 watch — FE eigenpal `file:` path defer | **Did not trip in M2.** The single `gen:api` ran via the present `openapi-typescript` binary only (no FE `pnpm install`), so the eigenpal `file:` path was never exercised. Defer still open for any future FE `pnpm install`. |
| 2026-06-14 | HS-1 | M2 close gate — `milestone-validator` C1–C7 **PASS** presented to operator | **Approved** by operator 2026-06-14 ("Open M3 — spec it"); M2 → passed. Verdict: `milestone-2-contract-tail/qa/milestone-qa.md` |
| 2026-06-14 | HS-1 | M3 open gate — operator approved opening M3 | **Approved** by operator 2026-06-14; M3 → in-progress, `milestone.md` authored up front (F3.1–F3.6, no execution detail) before any feature |
| 2026-06-14 | HS-6 | M3 spec stale — pre-F3.1 investigation found Wave 2.11 (`63f74368`) + 2.12 (H-6a) already did F3.1 (7 orphan deletes) and F3.2 (H-PRE-1 deadlock hoist) *before* the 06-13 audit that re-flagged them; F3.6 ("dead camelCase MarshalJSON") doesn't exist | **Replanned** by operator 2026-06-14 ("amend milestone.md, then execute"). `milestone.md` amended: F3.1/F3.2 → verify-already-done evidence rows; F3.6 struck (security-unsafe stale finding); real remainder = F3.3, F3.4, F3.5. Doc-drift (stale wiki/GitNexus refs to deleted symbols) handed to wiki-curator. |
| 2026-06-14 | HS-6 | F3.4 spec one-liner wrong — governing-spec `COUNT(*) OVER()` prescription assumed OFFSET, but the code is keyset/cursor pagination (a window on the cursor-filtered query counts only the post-cursor tail) | **Replanned** — operator delegated the engineering call ("what do you recommend as a Google senior engineer"); chose **Approach B (CTE single-query)**: count over the base-filtered set in a CTE *before* the cursor predicate. Recorded in `f3.4-…/spec.md` HS-6 reconciliation. |
| 2026-06-15 | HS-6 | F3.5 site set stale — milestone named 3 `DeleteObject` sites (`:537/:740/:331`); current tree has **1** (file is 710 lines; `:331` not a DeleteObject) | **Resolved under the row's own "or documented" clause** (not a full stop) — single real site `service.go:534` fixed; two stale-named sites documented non-existent. `f3.5-…/spec.md` + `evidence.md`. |
| 2026-06-15 | HS-1 | M3 close gate — `milestone-validator` C1–C7 **PASS** presented to operator | **Awaiting operator approval.** Verdict: `milestone-3-mechanical-quality/qa/milestone-qa.md`. M3 → passed (status flipped on the PASS). **No M4 open and no merge without explicit operator go.** |

## Program close-out / reconciliation

Fill in only when the last milestone has passed:

- [ ] Every planned feature (M0–M4) has a complete evidence row.
- [ ] Zero unplanned scope merged; anything added is recorded with rationale.
- [ ] Every bounded defer has a written trigger and an owner.
- [ ] M5 re-audit passed the §6 pass bar — link the evidence.
- [ ] Forward roadmap (F0.3) reflects the executed program and any deferred triggers.
- [ ] Operator sign-off: <date / name>
