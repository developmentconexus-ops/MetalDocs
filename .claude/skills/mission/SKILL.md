---
name: mission
description: >-
  Plan a large body of work as a MISSION — research the problem, lock the strategy with the operator,
  and decompose intent into an evidence-backed governing spec (mission.md) plus a scaffolded program tree,
  then hand off to the `milestone` skill to execute. This is the planning layer ABOVE `milestone`: it
  produces exactly what `milestone` Phase 1 assumes already exists (a governing spec + a program README),
  and adds a program-scale terminal acceptance gate defined up front. Use whenever the operator wants to
  BUILD, REMEDIATE, ENHANCE, or MIGRATE something substantial and needs it turned into milestones and
  features — e.g. "let's plan a mission to…", "I want to implement X", "turn this into a mission",
  "decompose this into milestones", "scope out this initiative", "what's our plan to reach <goal>". Trigger
  even without the word "mission" when the intent is a multi-milestone program that needs research +
  strategy + decomposition before any milestone runs. Prefer this over jumping straight into `milestone`
  (which starts mid-stream, assuming a spec already exists) and over a single `writing-plans` plan whenever
  the work is big enough to need staged delivery with a defined definition-of-done.
---

# Mission Workflow

Turn a raw intent into a **mission**: an evidence-backed governing spec (`mission.md`) whose own
definition-of-done is written **up front**, plus a scaffolded program tree, then hand off to the
`milestone` skill to execute. A mission decomposes into **milestones**; milestones into **features**.

This skill owns the layer *above* execution. The `milestone` skill's Phase 1 ("Program init") says:
*"identify the governing spec; if none exists, stop and create one first."* **This skill is that missing
front-end.** It researches, locks the strategy with the operator, decomposes, and produces precisely the
inputs `milestone` expects — then gets out of the way. It does **not** execute milestones.

**Canonical worked example of the output:** `docs/superpowers/milestones/grade-a-architecture-remediation/`
— a mission built by hand (governing spec + program README + M0–M5 + an independent re-audit as terminal
acceptance). `/mission` automates the construction of that front matter. Read it to see a finished mission.

## Core principles

- **Plan, then hand off.** A mission does research → decision-gate → `mission.md` → scaffold, then invokes
  `milestone`. It never runs a milestone itself. The seam is the guardrail: if the mission starts executing,
  it has overstepped. Each milestone still runs in its own fresh session with its own gates.
- **Evidence, not vibes.** Decomposition stands on a written, cited `discovery-brief.md`. Every claim in
  `mission.md` — every defect, every requirement, every milestone — traces back to a discovery finding.
  Decomposing on assumptions is the failure mode this skill exists to kill.
- **Definition-of-done, up front.** `mission.md` states what the end-state *shall be*, what to validate, and
  how — **before any milestone runs**. This is consumer-contract-first discipline lifted to program scale:
  you cannot judge "done" against a bar you write afterward.
- **Strategy is the operator's, and it is locked.** The load-bearing strategic choices (scope, sequencing,
  proof-of-done) are decided **with** the operator through a fail-closed interview and recorded as a
  **Locked Decisions** table. Never guess strategy to keep moving.
- **Separation of powers at the terminal gate.** The mission's terminal acceptance is judged by an
  independent `mission-validator` subagent that **judges and writes the verdict only** — never edits code,
  never flips status. A mission is never "done" by self-assertion.
- **Adapt to the kind of mission.** One skill, four playbooks — `remediation | greenfield-build |
  enhancement | migration`. The discovery shape, the spec emphasis, and the validation strategy adapt; the
  `mission.md` template is shared.
- **Token-efficient by construction.** Discovery is sized to the mission (a few agents, not a fleet).
  Per-feature and per-milestone machinery is **referenced** from `milestone`, never duplicated here.
- **Don't force the structure.** If the work is small enough that milestones are overkill, say so and
  recommend a single `writing-plans` plan instead. A mission is for programs that need more than one gate.

## Output layout (reuses the `milestone` tree — zero-friction seam)

Scaffold one tree per mission under `docs/superpowers/milestones/<mission-slug>/`:

```
docs/superpowers/milestones/<mission-slug>/
  mission.md            # ★ the governing spec + Terminal Acceptance (definition-of-done, up front)
  discovery-brief.md    # the cited evidence base every claim in mission.md traces to
  README.md             # program index (from milestone's templates/program-README.md)
  milestone-<n>-<slug>/ # authored by the `milestone` skill (its Phase 2), NOT by `/mission`
    milestone.md
  qa/
    mission-validation.md  # the mission-validator's terminal PASS/FAIL verdict (written at the very end)
```

`README.md`'s "Governing spec:" line points to `./mission.md`. `/mission` writes only `mission.md`,
`discovery-brief.md`, and `README.md`; the `milestone` skill owns every `milestone-<n>/` folder and operates
on this tree **unchanged** — `mission.md` *is* the governing spec it looks for.

## Workflow

```dot
digraph mission {
  "Phase 0: Intake & framing" [shape=box];
  "Phase 1: Discovery (adaptive, lean) -> discovery-brief.md" [shape=box];
  "Phase 2: Decision interview (fail-closed) -> Locked Decisions" [shape=box];
  "Phase 3: Decompose & author mission.md" [shape=box];
  "Phase 4: Self-review + operator gate" [shape=diamond];
  "Phase 5: Scaffold tree + commit" [shape=box];
  "Hand off to milestone skill (executes M0..)" [shape=box];
  "Last milestone passed its milestone-validator" [shape=diamond];
  "Dispatch mission-validator (terminal acceptance)" [shape=box];
  "Verdict?" [shape=diamond];
  "Operator sign-off / program close-out" [shape=doublecircle];

  "Phase 0: Intake & framing" -> "Phase 1: Discovery (adaptive, lean) -> discovery-brief.md";
  "Phase 1: Discovery (adaptive, lean) -> discovery-brief.md" -> "Phase 2: Decision interview (fail-closed) -> Locked Decisions";
  "Phase 2: Decision interview (fail-closed) -> Locked Decisions" -> "Phase 3: Decompose & author mission.md";
  "Phase 3: Decompose & author mission.md" -> "Phase 4: Self-review + operator gate";
  "Phase 4: Self-review + operator gate" -> "Phase 3: Decompose & author mission.md" [label="changes"];
  "Phase 4: Self-review + operator gate" -> "Phase 5: Scaffold tree + commit" [label="approved"];
  "Phase 5: Scaffold tree + commit" -> "Hand off to milestone skill (executes M0..)";
  "Hand off to milestone skill (executes M0..)" -> "Last milestone passed its milestone-validator";
  "Last milestone passed its milestone-validator" -> "Dispatch mission-validator (terminal acceptance)" [label="yes"];
  "Dispatch mission-validator (terminal acceptance)" -> "Verdict?";
  "Verdict?" -> "Phase 3: Decompose & author mission.md" [label="FAIL -> HS-5 micro-milestone"];
  "Verdict?" -> "Operator sign-off / program close-out" [label="PASS"];
}
```

### Phase 0 — Intake & framing (cheap; no fan-out)
1. Capture the mission statement — the intent in the operator's words.
2. **Classify the mission type:** `remediation | greenfield-build | enhancement | migration`. If it's
   ambiguous, ask **one** question. The type selects the Phase-1 discovery playbook.
3. **Load house context cheaply** so the plan respects the rules: `CLAUDE.md`, `wiki/README.md`,
   `wiki/references/current-agent-handoff.md`, the skill-routing table, and any relevant memories. A mission
   that violates house rules (skill routing, DB/FE/BE boundaries, hard-stops) is dead on arrival.
4. Pick a `<mission-slug>` (short kebab-case); verify it doesn't collide with an existing program tree.
5. Output a **one-screen framing**: intent · type · slug · success-in-one-sentence. No files yet.

### Phase 1 — Discovery (always produces a brief; effort proportional to risk)
Read `references/mission-discovery.md` for the per-type playbooks and the model/token policy. The
`discovery-brief.md` artifact is non-negotiable — decomposing on un-cited assumptions is the failure this
phase exists to prevent. The *effort*, though, scales to the unknowns: a well-understood small mission may
warrant a single confirm-and-cite agent; an unfamiliar audit warrants real fan-out.

1. Fan out a set of parallel agents **sized to the mission** — default **3–6**, down to **1** for a small
   well-scoped mission, and a large fleet only when the mission is genuinely huge **and** the operator opts
   in (say what it costs first). Model policy: **sonnet** for analysis/judgement, **haiku** for mechanical
   sweeps, **never fable** for workers, **≤15 concurrent**.
2. Run the playbook for the mission type (audit / requirements+prior-art / impact scan / site census).
3. Run **one cheap skeptic pass** over the findings so the evidence base isn't hallucinated. Scale rigor to
   risk; if you skip verification for a trivially-checkable finding, say so.
4. Write **`discovery-brief.md`** (template in `templates/discovery-brief.md`) — the cited evidence base.
   Include a **coverage statement**: what was *not* swept. No silent caps.

### Phase 2 — Decision interview (fail-closed)
This phase **is** the brainstorming discipline applied to strategy — so compose, don't re-implement: where
`superpowers:brainstorming` is installed, use it to run the interview; this loop is the fallback when it
isn't. If a `brainstorming` session already produced this mission's intent, **carry those decisions forward
into the Locked Decisions table — do not re-interview** the operator on choices already made.
1. Present **2–3 strategic approaches** with trade-offs and a clear recommendation (scope: full vs bounded;
   sequencing; what counts as proof-of-done).
2. Interview the operator **one question at a time** on the load-bearing choices. Record each answer in the
   **Locked Decisions (D1..Dn)** table of `mission.md`.
3. **Fail closed.** If scope, approach, or the terminal bar is ambiguous, interview — never guess.

### Phase 3 — Decompose & author `mission.md`
Read `references/mission-decomposition.md` for the decomposition heuristics and anti-patterns, and
`references/mission-terminal-acceptance.md` for how to write §8.

1. Decompose intent → **milestones → features**. Each milestone is a bounded slice **validatable in one
   pass**; each feature is the atomic unit with *what to implement* + *what to validate* (objective
   acceptance). Order dependencies first and **risk-isolating work last** (the grade-a mission put the
   systemic ports last "so it cannot regress the grade").
2. Author `mission.md` from `templates/mission.md`. You may draft genuinely-independent sections in parallel
   (milestone table / hard-stop catalog / terminal acceptance) and reconcile in the main session — only
   where it actually saves time.
3. **No execution detail in `mission.md`.** The milestones say *what* and *what-to-validate*, never *how*.
   The "how" lives downstream in each `milestone.md` and feature `spec.md`/`plan.md`. `mission.md` is a
   stable contract you validate against — it must not drift into "whatever we ended up doing".

### Phase 4 — Self-review + operator gate
1. Self-review `mission.md` with fresh eyes:
   - **Placeholders** — no TBD/TODO/empty sections.
   - **Consistency** — milestones, decisions, and terminal acceptance don't contradict each other.
   - **Scope** — each milestone is validatable in one pass; none is secretly two milestones.
   - **Ambiguity** — every acceptance criterion is objectively checkable.
   - **Traceability** — every milestone/feature maps to a discovery finding, and every discovery finding
     maps to a milestone **or** an explicit out-of-scope note. Orphans on either side are a defect.
2. Present `mission.md` to the operator for review. Iterate until approved. **This is a hard gate** — do not
   scaffold or hand off without approval.

### Phase 5 — Scaffold + hand off to `milestone`
`/mission` writes the **program-level** files only — it does **not** author any milestone's `milestone.md`
and creates no `milestone-<n>/` folder. Owning the milestone files is `milestone`'s job (its Phase 2);
authoring them in both skills would create two sources of truth for the same milestone spec.
1. Scaffold the program index: copy `milestone`'s `templates/program-README.md` → `README.md`; fill the
   milestone table (titles + one-line objectives + status `planned`); and **add the terminal-gate line** to
   the close-out checklist (a literal "dispatch `mission-validator`" step — see "Terminal acceptance" below)
   so the gate can't be silently skipped once this session is gone.
2. **Commit** the mission artifacts — `mission.md` + `discovery-brief.md` + `README.md` (standing
   authorization, `CLAUDE.md §5.0`). **Never push.**
3. **Hand off as a baton, not a call.** A mission is large; executing every milestone in this same context
   would blow it out, and §11/D2 commit to a **fresh session per milestone**. So **stop here.** Tell the
   operator: the mission is specced and scaffolded; start M0 by invoking `/milestone` **in a fresh session**
   (it reads `mission.md` as its governing spec and owns all milestone files from its Phase 2). The next
   operator gate is HS-1 at the M0 boundary. Do **not** begin executing M0 in this session.

### Terminal acceptance (after the LAST milestone)
The one thing the mission adds beyond `milestone`'s machinery — and the easiest to lose. **Trigger
ownership:** by the time the last milestone closes, this `/mission` session is long gone and `/milestone`
knows nothing about the terminal gate. So the trigger is made durable during Phase 5 in two places — the
program `README.md` close-out checklist carries an explicit **"dispatch `mission-validator`"** line, and
`mission.md` §8 names who invokes it (the operator, or whichever session closes the last milestone). Without
that, the gate is silently skipped.

After the last milestone passes its own `milestone-validator` **and** clears its HS-1 operator gate:
1. **Run the validation according to its shape (a single subagent cannot fan out — it has no `Agent` tool):**
   - **Fan-out validation** (e.g. a remediation re-audit across many dimensions) → the **main session** runs
     it with `Workflow`/`Agent` and captures the artifact, then dispatches `mission-validator` to **judge
     that artifact** against §8.
   - **Deterministic validation** (test suites, CI greps, single-pass review) → dispatch `mission-validator`
     directly; it runs those itself.
2. `mission-validator` (`.claude/agents/mission-validator.md`) reads `mission.md` §8, verifies each criterion
   against real evidence, and writes `qa/mission-validation.md` (PASS/FAIL + per-criterion). Separation of
   powers: it judges and writes the verdict only — never edits code, never flips status.
3. On **FAIL** → **HS-5**: bounded remediation micro-milestone for the missed criteria via `milestone`, then
   re-dispatch. The operator decides continue vs replan at each loop.
4. On **PASS** → present to the operator for final sign-off; write the §12 program close-out in `README.md`.

## Hard-stop catalog (defaults; each mission generalizes them in `mission.md` §9)

| ID | Trigger | Action |
|----|---------|--------|
| HS-1 | Every milestone boundary (inherited from `milestone`) | Operator review gate; no next milestone and no merge without approval |
| HS-5 | The terminal `mission-validator` misses the §8 pass bar | Bounded remediation micro-milestone, then re-dispatch; operator decides continue vs replan |
| HS-6 | Scope drift surfaces during discovery or decomposition | Stop; surface the deviation; re-interview before authoring/continuing `mission.md` |
| HS-2 / HS-3 / HS-4 | During execution, inherited from `milestone` | As defined by the `milestone` skill |

Always make the in-force hard-stops explicit in `mission.md`.

## Templates & references

In `templates/` (copy and fill at runtime):
- `mission.md` — the governing spec + up-front Terminal Acceptance (the mission's definition-of-done)
- `discovery-brief.md` — the cited evidence base discovery produces

In `references/` (read when you reach the relevant phase):
- `mission-discovery.md` — the adaptive discovery playbooks (by type) + model/token policy
- `mission-decomposition.md` — intent → milestones → features heuristics + anti-patterns
- `mission-terminal-acceptance.md` — how to author §8 + the `mission-validator` charter

Reused from the `milestone` skill (referenced, not copied):
- `.claude/skills/milestone/templates/program-README.md` — the program index (the only milestone template
  `/mission` fills; every `milestone.md` is authored later by the `milestone` skill, not here)

Agent: `.claude/agents/mission-validator.md` — the independent, program-scale terminal-acceptance judge.
