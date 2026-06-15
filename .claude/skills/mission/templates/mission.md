# Mission: <Mission Name>

> **Status:** Drafting | Operator-approved | Executing (Milestone <n>) | Terminal-validating | Complete | Blocked
> **Date:** <date>  ·  **Branch of record:** <branch>
> **Type:** remediation | greenfield-build | enhancement | migration
> **Slug:** `<mission-slug>`  ·  **Owner / operator:** <name>
> **Evidence base:** `./discovery-brief.md`  ·  **Program index:** `./README.md`
> **Governs:** Milestones M0..M<n> below. Each milestone gets its own plan via the `milestone` skill,
> executed in a fresh session. This file is the **stable governing contract** — it says *what* the
> mission is and *what proves it done*; it contains **no execution detail**.

---

## 1. Problem / why now
The problem this mission exists to solve, stated against evidence (cite `discovery-brief.md`). Why it
matters and why now. No solution here — just the gap.

## 2. Goals / Non-Goals
**Goals** — what success delivers (bulleted, measurable where possible).
**Non-Goals** — what this mission explicitly does *not* do (YAGNI-ruthless; name the tempting things you
are deliberately excluding so scope can't creep into them later).

## 3. Locked decisions (operator-approved)
The load-bearing strategic choices, decided with the operator (Phase 2). These are binding.

| # | Decision | Value |
|---|----------|-------|
| D1 | Scope | <full vs bounded; what's in/out> |
| D2 | Execution | <e.g. one mission.md governs; per-milestone plans + inter-milestone operator gates; fresh session per milestone> |
| D3 | Proof of done | <how the mission proves it reached the bar — the terminal validation mechanism> |
| D4 | <key shape decision> | <value> |
| D5 | Sequencing | <ordering rationale — dependencies first, risk-isolating last> |
| … | | |

## 4. Discovery summary
2–4 sentences on what discovery found and the confidence in it (verified vs assumed). Link the brief:
`./discovery-brief.md`. The rest of this document stands on those findings — every milestone below
traces to one.

## 5. Work / requirement inventory
The concrete items the mission must address, each mapped to a milestone. For **remediation/migration**
use `file:line` + class/site. For **greenfield/enhancement** use requirements/capabilities. Every row
maps to a milestone in §7; anything found in discovery but *not* addressed appears here marked
**out-of-scope** with a reason.

| # | Item (site / requirement) | Class / kind | Milestone |
|---|---------------------------|--------------|-----------|
| 1 | <site or requirement> | <class> | M<n> / F<n>.x |
| … | | | |

## 6. Program architecture (by reference)
This mission executes via the `milestone` skill. The per-feature close-out loop, the per-feature
consumer-contract spec gate, and the per-milestone `milestone-validator` gate are defined there — **not
duplicated here**. See `.claude/skills/milestone/SKILL.md`. What this section adds is only the
program-scale shape:

```
Mission: <name>
└── Milestone (M0..M<n>)        ── each: features → milestone-validator gate → HS-1 operator gate
    └── Feature (Fx.y)          ── each: spec(consumer-contract) → plan → TDD → evidence
Terminal acceptance (§8)        ── independent mission-validator, after the last milestone passes
```

## 7. Milestones
The decomposition. Each milestone is a bounded slice **validatable in one pass**. Per feature: *what to
implement* (by outcome, not steps) and *what to validate* (objectively checkable). **No execution detail.**
Order dependencies first and risk-isolating work last (so late milestones can't regress the bar).

### M0 — <Title>
**Objective:** <one line — the coherent slice this milestone delivers, and the bar it moves>.

| Feature | What to implement | What to validate (acceptance) |
|---------|-------------------|-------------------------------|
| F0.1 <slug> | <outcome> | <objective criterion — a test, a route shape, a clean build, an observed runtime behavior> |
| F0.2 <slug> | | |

**Milestone gate:** <which canonical QA checklist(s) + what the close gate proves; root-cause criterion
if this milestone moves a quality bar>. Validated by the `milestone-validator` (separation of powers).

### M1 — <Title>
… (repeat per milestone)

## 8. ★ Terminal acceptance — definition of done (written up front)
The mission's binding definition-of-done. Authored **now**, before any milestone runs, and judged at the
end by the independent `mission-validator` (`.claude/agents/mission-validator.md`).

- **Pass bar (the mission shall be X):** <measurable end-state — e.g. "3 dimensions ≥ A−, 0 new
  Critical/Major, defect class C at 0", or "feature X live end-to-end with acceptance suite green">.
- **What to validate:** <the concrete checklist of conditions that together mean done>.
- **How to validate:** <the method — re-audit fan-out / full E2E run / acceptance suite / CI guard greps —
  with the exact commands or agent shape the validator will run>.
- **Who validates:** the independent `mission-validator` subagent. It judges and writes
  `qa/mission-validation.md` only — it never edits code or flips status.
- **On miss (HS-5):** the missed criteria become a bounded remediation micro-milestone, run through
  `milestone`, then the `mission-validator` is re-dispatched. The operator decides continue vs replan at
  each loop.

## 9. Hard-stop catalog
The hard-stops in force for this mission (defaults below; add mission-specific ones). State explicitly what
trips each.

| ID | Trigger | Action |
|----|---------|--------|
| HS-1 | Every milestone boundary | Operator review gate; no next milestone / no merge without approval |
| HS-2 | A fix implies redesign outside the assigned boundary | Stop; report the boundary + minimum prerequisite plan; no symptom-patch |
| HS-3 | A prerequisite boundary fails (build/runnable/auth/route/contract truth) | Repair (e.g. `runtime-contract-prereq`); rerun the checkpoint; resume |
| HS-4 | A `milestone-validator` returns FAIL | Open the named fix feature; re-run its lifecycle; re-dispatch the validator |
| HS-5 | The terminal `mission-validator` misses the §8 bar | Bounded remediation micro-milestone; re-dispatch; operator decides continue vs replan |
| HS-6 | Scope drift / off-plan discovery | Stop; surface the deviation; replan before continuing |

## 10. Constraints respected
House rules and invariants this mission must not violate (skill routing, DB/FE/BE boundaries, contract-first
regen order, no-merge-by-agent, any architecture invariants like advisory-lock hazards, relevant ADRs).

## 11. Execution model
One `mission.md` governs all milestones. Each milestone → its own plan via `milestone` (writing-plans where
installed), executed in a **fresh session**, subagent-driven. Operator gate between every milestone (HS-1);
**no merge by the agent**. Token discipline: parallel fan-out only where it pays (discovery, terminal
re-audit); everything else direct tools. Model policy: sonnet analysis/review, haiku mechanical, never fable
workers, ≤15 concurrent.

## 12. End-state / reconciliation
Fill only when the last milestone has passed and the terminal gate is green:
- [ ] Every planned feature (M0..M<n>) has a complete evidence row.
- [ ] Zero unplanned scope merged; anything added is recorded with rationale.
- [ ] Every bounded defer has a written trigger and an owner.
- [ ] Terminal acceptance (§8) passed — link `qa/mission-validation.md`.
- [ ] Operator sign-off: <date / name>
