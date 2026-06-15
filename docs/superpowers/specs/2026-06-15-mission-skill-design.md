# /mission Skill — Design Spec

> **Status:** Approved (brainstorming gate, 2026-06-15) — ready for build
> **Date:** 2026-06-15
> **Branch of record:** main
> **Author:** backend agent (Opus 4.8) + operator (leandrotca.work)
> **Governs:** the build of the `/mission` skill (`.claude/skills/mission/`) + the `mission-validator` agent.
> **Builds on:** `.claude/skills/milestone/` (the execution layer this skill feeds) and the hand-built reference
> program `docs/superpowers/milestones/grade-a-architecture-remediation/` (the artifact `/mission` automates).
> **Build mechanism:** `anthropic-skills:skill-creator` (operator decision D5), not writing-plans.

---

## 1. Problem

`/milestone` runs a large program as a chain of validated milestones, but it **starts mid-stream**: its
Phase 1 ("Program init") says *"identify the governing spec; if none exists, stop and create one first."*
That front-end — research the problem, lock the strategy with the operator, decompose intent into a
**governing spec + program roadmap** — is done **by hand** today. The `grade-a-architecture-remediation`
program is the proof it works: a governing spec (`specs/2026-06-14-…-design.md`) standing on an
independent audit (the 2026-06-13 architecture audit — **evidence, not vibes**), a program README with a
milestone table + hard-stop ledger, and M0–M5 with an independent re-audit as terminal acceptance.

`/mission` automates that front-end: **intent → evidence-backed `mission.md` (governing spec with its own
up-front terminal validation) → scaffolded program tree → handoff to `/milestone`.** A mission decomposes
into milestones; milestones into features. Nothing is decomposed on assumptions, and the mission defines
what "done and correct" means **before** any milestone runs.

## 2. Goals / Non-Goals

**Goals**
- One skill that turns a raw intent into the exact Phase-1 inputs `/milestone` expects (governing spec +
  program README), plus a program-scale terminal gate, with **zero changes to `/milestone`**.
- Decomposition stands on a **written, cited evidence base** (`discovery-brief.md`) — never on vibes.
- The operator's strategy is captured as **locked decisions** through a fail-closed interview.
- The mission's **terminal definition-of-done is written up front** in `mission.md` and judged by an
  **independent** validator at the end (separation of powers, program scale).
- **Token-efficient by construction:** discovery sized to the mission; per-feature/per-milestone machinery
  is referenced from `/milestone`, never duplicated.

**Non-Goals**
- Not a replacement for `/milestone` — `/mission` plans and hands off; it does **not** execute milestones.
- No new product features inside the skill; no speculative configurability beyond the four mission types.
- The skill does **not** merge, and does **not** auto-flip program status — operator gates stay (HS-1).
- Does not re-implement brainstorming/writing-plans/skill-creator — it composes them.

## 3. Locked decisions (operator-approved, 2026-06-15)

| # | Decision | Value |
|---|----------|-------|
| D1 | Boundary | **Plan, then hand off.** `/mission` does research + decision-gate + writes `mission.md` + scaffolds the tree, then hands to `/milestone`. Each milestone still runs in its own fresh session with its own gates. |
| D2 | Terminal artifact | `/mission` emits **`mission.md`** stating the whole program **including its own terminal validation** — what the end-state shall be, what to validate, how to validate — defined **up front**. |
| D3 | Discovery | **Mandatory, adaptive depth, token-efficient.** Always a parallel-agent discovery phase producing a written `discovery-brief.md`; shape + agent count scale to mission type and size. |
| D4 | Mission type | **One skill, adaptive playbooks.** Detect/confirm `remediation \| greenfield-build \| enhancement \| migration`; adapt discovery, spec sections, validation — one shared `mission.md` template. |
| D5 | Terminal gate | **Independent validator (mission-scale).** `mission.md` defines the binding pass-bar up front; an independent `mission-validator` judges it at the end. Never "done" by self-assertion. |
| D6 | Build mechanism | Build the skill via **`skill-creator`**. |
| D7 | Layout | **Reuse the `/milestone` tree.** `mission.md` lives at the program root and *is* the governing spec `/milestone` Phase 1 looks for. Zero-friction seam. |

## 4. Skill architecture (files the build produces)

```
.claude/skills/mission/
  SKILL.md                       # the workflow: 6 phases, principles, seam to /milestone
  templates/                     # copy-and-fill at runtime
    mission.md                   # ★ governing spec + Terminal Acceptance (def-of-done up front)
    discovery-brief.md           # the cited evidence base
  references/
    mission-discovery.md         # adaptive discovery playbooks by mission type + token/model policy
    mission-decomposition.md     # intent → milestones → features heuristics + anti-patterns
    mission-terminal-acceptance.md  # how to author the pass-bar + the mission-validator charter
.claude/agents/
  mission-validator.md           # the independent program-scale terminal judge
```

`program-README.md` is **reused from** `.claude/skills/milestone/templates/program-README.md` (referenced,
not copied), keeping one source of truth for the program index.

### 4.1 Runtime output layout (what `/mission` produces for a user's program)
```
docs/superpowers/milestones/<mission-slug>/
  mission.md            # ★ governing spec + Terminal Acceptance
  discovery-brief.md    # cited evidence base
  README.md             # program index (from milestone's program-README template)
  milestone-0-<slug>/   # first milestone scaffolded; /milestone executes from here
    milestone.md
  qa/
    mission-validation.md  # mission-validator's terminal PASS/FAIL (written at the very end)
```
`README.md` "Governing spec:" → `./mission.md`. `/milestone` operates on this tree unchanged.

## 5. The flow (6 phases)

### Phase 0 — Intake & framing (cheap, no fan-out)
1. Capture the mission statement (the intent).
2. Classify the mission **type** (D4): `remediation | greenfield-build | enhancement | migration`. If
   ambiguous, ask the operator **one** question. Type selects the Phase-1 playbook.
3. Load house context cheaply: `CLAUDE.md`, `wiki/README.md`, `wiki/references/current-agent-handoff.md`,
   the skill-routing table, and relevant memories — so the plan respects house rules (skill routing,
   DB/FE/BE rules, hard-stops).
4. Pick `<mission-slug>`; verify it does not collide with an existing program tree.
5. Output a one-screen framing (intent · type · slug · success-in-one-sentence). **No file yet.**

### Phase 1 — Discovery (mandatory, adaptive, token-efficient)
1. Fan out a **small** set of parallel agents sized to the mission — default **3–6**, never a 42-agent
   blast unless the mission is genuinely huge and the operator opts in. Model policy: **sonnet** for
   analysis/judgement, **haiku** for mechanical sweeps, **never fable** for workers, **≤15 concurrent**.
2. Run the type playbook (`references/mission-discovery.md`):
   - **remediation** → audit current state: map affected modules, enumerate defects/debt with `file:line`,
     grade the touched dimensions (the grade-a 2026-06-13 audit shape, scoped down).
   - **greenfield-build** → requirements elaboration + prior-art/library research (Context7 / WebSearch) +
     integration-point map against the existing codebase.
   - **enhancement** → targeted impact scan: what exists, who consumes it, blast radius (GitNexus `impact`
     for genuinely high-risk symbols only).
   - **migration** → census of sites + from→to mapping + risk hotspots.
3. One **cheap skeptic pass** verifies the findings so the evidence base is not hallucinated; scale rigor
   to risk (skip only for trivially-verifiable findings, and say so).
4. Output **`discovery-brief.md`** — the cited evidence base. Every later claim in `mission.md` traces to it.

### Phase 2 — Decision interview (brainstorming gate, fail-closed)
1. Present **2–3 strategic approaches** with trade-offs + a recommendation (scope, sequencing, proof-of-done).
2. Interview the operator **one question at a time** on the load-bearing choices; lock them into a
   **Decisions table D1..Dn** in `mission.md` (the grade-a locked-decisions pattern).
3. **Fail closed:** never guess a strategic decision to keep moving.

### Phase 3 — Decompose & author `mission.md`
1. Decompose intent → **milestones → features** (`references/mission-decomposition.md`): each milestone a
   bounded slice validatable in one pass; each feature the atomic unit with *what to implement* + *what to
   validate*. Order dependencies first, risk-isolating work last (grade-a put systemic ports last "so it
   cannot regress the grade").
2. Author `mission.md` from the template (§6). Draft independent sections in parallel where it pays
   (milestone table / hard-stop catalog / terminal acceptance), main session reconciles.
3. **No execution detail** in `mission.md` milestones — that lives in each `milestone.md`/feature `spec.md`
   downstream. `mission.md` is a stable contract, not "whatever we end up doing".

### Phase 4 — Self-review + operator gate
1. Self-review: placeholder scan, internal consistency, scope (each milestone validatable in one pass),
   ambiguity, and **traceability** — every milestone/feature ↔ a discovery finding; every discovery finding
   ↔ a milestone **or** an explicit out-of-scope note.
2. Operator reviews `mission.md`. Iterate until approved.

### Phase 5 — Scaffold + handoff to `/milestone`
1. Scaffold the tree: `README.md` (milestone's program-README template), `milestone-0-<slug>/milestone.md`
   for the first milestone.
2. Commit (standing authorization, CLAUDE.md §5.0). Never push.
3. Hand off: invoke `/milestone` (its Phase 2 onward) — it sees the governing spec = `mission.md` and runs
   M0. The terminal `mission-validator` (§8) fires only **after the last milestone passes its own
   `milestone-validator`**.

## 6. `mission.md` template — section spec

13 sections. The standout is §9 (the operator's headline ask).

1. **Header** — status · date · branch · author/operator · mission type · slug · links (discovery-brief, README).
2. **Problem / why now** — evidence-cited problem statement.
3. **Goals / Non-Goals** — explicit, YAGNI-ruthless.
4. **Locked Decisions (D1..Dn)** — operator-approved strategic choices (Phase-2 output).
5. **Discovery summary** — short; links `discovery-brief.md`; states the evidence the plan stands on.
6. **Work / defect inventory** — concrete items (`file:line` for remediation; requirements for greenfield),
   each mapped to a milestone.
7. **Program architecture** — the milestone→feature model + per-feature close-out loop + per-milestone gate,
   **by reference to `/milestone`** (no duplication).
8. **Milestones** — table: `# | milestone | objective | features (per-feature implement/validate) | gate`.
   Up front; **no execution detail**.
9. **★ Terminal Acceptance (definition-of-done, up front)** —
   - **Pass bar:** "the mission shall be X" (measurable).
   - **What to validate:** the concrete checklist.
   - **How to validate:** commands / method / re-audit shape.
   - **Who validates:** the independent `mission-validator` (charter linked).
   - **On miss:** bounded remediation micro-milestone (HS-5), re-run, operator decides continue vs replan.
10. **Hard-stop catalog** — HS-1..n, generalized per mission (reuse milestone's catalog + mission-specific).
11. **Constraints respected** — house rules, ADRs, invariants (e.g. H-PRE-1), no-merge.
12. **Execution model** — one `mission.md` governs; per-milestone plans via `/milestone`; fresh session per
    milestone; token discipline + model policy.
13. **End-state / reconciliation** — close-out checklist (every feature has evidence; zero unplanned scope;
    every defer has a written trigger; terminal acceptance passed; operator sign-off).

## 7. `discovery-brief.md` template — section spec
- **Header** — mission slug · type · date · agents/models used · scope swept.
- **Method** — what each agent did; what was verified vs assumed (skeptic-pass result).
- **Findings** — the cited inventory (`file:line` / requirement / site census), each tagged with proposed
  milestone home.
- **Constraints & risks surfaced** — house rules, invariants, blast-radius hotspots.
- **Open questions for the operator** — what Phase 2 must lock.
- **Coverage statement** — what was *not* swept (no silent caps).

## 8. `mission-validator` agent — spec
- Lives at `.claude/agents/mission-validator.md`; **program-scale sibling** of `milestone-validator`.
- Runs **once**, at the very end, after the last milestone's `milestone-validator` returns PASS.
- Reads `mission.md` §9 Terminal Acceptance, **executes the declared validation** (re-audit / E2E /
  acceptance suite — whatever the mission declared), writes `qa/mission-validation.md` with a PASS/FAIL
  verdict and per-criterion evidence.
- **Separation of powers:** judges and writes the verdict **only** — never edits code, never fixes
  findings, never flips program status. The main session + operator own the final Grade/Done declaration.
- On **FAIL** → HS-5: bounded remediation micro-milestone, then re-dispatch. Loop with operator decision.

## 9. The seam to `/milestone`
- `/mission` produces precisely `/milestone` Phase-1 inputs: a governing spec (`mission.md`) + a program
  README. `/milestone` then runs Phase 2→5 per milestone **unchanged**.
- The **only** addition beyond milestone's machinery is the program-scale `mission-validator` terminal gate
  tied to `mission.md` §9. Per-feature/per-milestone discipline is referenced, not re-implemented.

## 10. Token efficiency & model policy
- Discovery sized to the mission (3–6 agents default); **haiku** mechanical, **sonnet** analysis,
  **never fable** workers, **≤15 concurrent** (per the workflow-model-balancing standing rule).
- Reuse milestone templates/agents; `mission.md` points to `/milestone` rather than duplicating it.
- Parallel section-drafting only where independent. Skeptic passes scoped to risk.

## 11. Adaptive playbooks (D4) — summary
| Type | Discovery shape | Extra `mission.md` emphasis | Typical terminal validation |
|------|-----------------|-----------------------------|-----------------------------|
| remediation | audit current state, `file:line` defect inventory, dimension grades | defect inventory §6 maps every defect to a milestone | independent re-audit hits the pass bar |
| greenfield-build | requirements + prior-art/library research + integration map | requirements + architecture in §6/§7 | acceptance suite / E2E proves the spec |
| enhancement | targeted impact scan + consumer/blast-radius map | non-regression of existing consumers | regression + new-capability acceptance |
| migration | site census + from→to map + risk hotspots | exhaustive census in §6; CI guards | 0 old-pattern sites + behavior parity |

Detail lives in `references/mission-discovery.md`.

## 12. Hard-stop catalog (the skill ships these defaults; missions generalize)
| ID | Trigger | Action |
|----|---------|--------|
| HS-1 | Every milestone boundary (inherited from `/milestone`) | Operator review gate; no next milestone / no merge without approval |
| HS-5 | Terminal acceptance (mission-validator) misses the pass bar | Bounded remediation micro-milestone, re-run; operator decides continue vs replan |
| HS-6 | Scope drift discovered during discovery/decomposition | Stop; surface the deviation; re-interview before authoring `mission.md` |
| (inherit) | HS-2/HS-3/HS-4 from `/milestone` apply during execution | as defined by `/milestone` |

## 13. Build & acceptance (how we know the skill itself is done)
Built via `skill-creator` (D6). The skill is accepted when:
- [ ] `.claude/skills/mission/SKILL.md` encodes the 6 phases, principles, seam, and a triggering
      description (skill-creator description-eval run).
- [ ] All five runtime artifacts exist and are coherent: `templates/mission.md`, `templates/discovery-brief.md`,
      `references/mission-discovery.md`, `references/mission-decomposition.md`,
      `references/mission-terminal-acceptance.md`.
- [ ] `.claude/agents/mission-validator.md` exists with strict separation-of-powers wording.
- [ ] **Dry-run proof:** running `/mission` on a small intent produces a valid `mission.md` +
      `discovery-brief.md` + scaffolded tree that `/milestone` Phase 1 accepts without modification.
- [ ] `mission.md` template includes a non-optional §9 Terminal Acceptance with all five fields.

## 14. Open risks / non-goals restated
- Discovery rigor is a token/quality trade-off; default lean, operator can opt into a heavier sweep.
- `/mission` must not drift into executing milestones — the seam (D1) is the guardrail; if a mission is tiny
  enough that milestones are overkill, `/mission` should say so and recommend a single `writing-plans` plan
  instead of forcing the structure.
- The skill composes brainstorming/skill-creator/milestone; it must not re-implement them.
